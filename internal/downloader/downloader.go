package downloader

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

// Options describes CLI behavior for a download run.
type Options struct {
	OutputTemplate string
	AudioOnly      bool
	InfoOnly       bool
	Quiet          bool
	Timeout        time.Duration
}

type outputContext struct {
	Playlist    *youtube.Playlist
	Index       int
	Total       int
	EntryTitle  string
	EntryAuthor string
}

type downloadResult struct {
	bytes       int64
	outputPath  string
	retried     bool
	hadProgress bool
	skipped     bool
}

type reportedError struct {
	err error
}

func (e reportedError) Error() string {
	return e.err.Error()
}

func (e reportedError) Unwrap() error {
	return e.err
}

func markReported(err error) error {
	if err == nil {
		return nil
	}
	return reportedError{err: err}
}

// IsReported returns true if the error has already been printed to stderr.
func IsReported(err error) bool {
	var re reportedError
	return errors.As(err, &re)
}

// Process fetches metadata, selects the best matching format, and downloads it.
func Process(ctx context.Context, url string, opts Options) error {
	printer := newPrinter(opts)

	// Convert YouTube Music URLs to regular YouTube URLs
	url = ConvertMusicURL(url)

	if looksLikePlaylist(url) {
		return processPlaylist(ctx, url, opts, printer)
	}

	youtube.DefaultClient = youtube.AndroidClient
	client := newClient(opts)
	video, err := client.GetVideoContext(ctx, url)
	if err != nil {
		return fmt.Errorf("fetching metadata: %w", err)
	}
	if opts.InfoOnly {
		return printInfo(video)
	}
	prefix := printer.Prefix(1, 1, video.Title)
	result, err := downloadVideo(ctx, client, video, opts, outputContext{}, printer, prefix)
	if result.skipped {
		printer.ItemSkipped(prefix, "exists")
	} else {
		printer.ItemResult(prefix, result, err)
	}
	if err != nil {
		return markReported(err)
	}
	okCount := 1
	skipped := 0
	if result.skipped {
		okCount = 0
		skipped = 1
	}
	printer.Summary(1, okCount, 0, skipped, result.bytes)
	return nil
}

func newClient(opts Options) *youtube.Client {
	return &youtube.Client{
		HTTPClient: &http.Client{Timeout: opts.Timeout},
	}
}

func processPlaylist(ctx context.Context, url string, opts Options, printer *Printer) error {
	savedClient := youtube.DefaultClient
	defer func() {
		youtube.DefaultClient = savedClient
	}()

	youtube.DefaultClient = youtube.WebClient
	playlistClient := newClient(opts)
	playlist, err := playlistClient.GetPlaylistContext(ctx, url)
	if err != nil {
		return fmt.Errorf("fetching playlist: %w", err)
	}

	if opts.InfoOnly {
		return printPlaylistInfo(playlist)
	}

	if len(playlist.Videos) == 0 {
		return errors.New("playlist has no videos")
	}

	if !opts.Quiet {
		fmt.Fprintf(os.Stderr, "playlist: %s (%d videos)\n", playlist.Title, len(playlist.Videos))
	}

	youtube.DefaultClient = youtube.AndroidClient
	videoClient := newClient(opts)
	successes := 0
	failures := 0
	skipped := 0
	var totalBytes int64
	for i, entry := range playlist.Videos {
		prefix := printer.Prefix(i+1, len(playlist.Videos), entryTitle(entry))
		if entry == nil || entry.ID == "" {
			skipped++
			printer.ItemSkipped(prefix, "missing playlist entry")
			continue
		}

		video, err := videoClient.VideoFromPlaylistEntryContext(ctx, entry)
		if err != nil {
			failures++
			printer.ItemResult(prefix, downloadResult{}, err)
			continue
		}

		result, err := downloadVideo(ctx, videoClient, video, opts, outputContext{
			Playlist:    playlist,
			Index:       i + 1,
			Total:       len(playlist.Videos),
			EntryTitle:  entry.Title,
			EntryAuthor: entry.Author,
		}, printer, prefix)
		if result.skipped {
			skipped++
			printer.ItemSkipped(prefix, "exists")
			continue
		}
		printer.ItemResult(prefix, result, err)

		if err != nil {
			failures++
			continue
		}

		successes++
		totalBytes += result.bytes
	}

	printer.Summary(len(playlist.Videos), successes, failures, skipped, totalBytes)
	if successes == 0 {
		return markReported(errors.New("no playlist entries downloaded successfully"))
	}
	return nil
}

func downloadVideo(ctx context.Context, client *youtube.Client, video *youtube.Video, opts Options, ctxInfo outputContext, printer *Printer, prefix string) (downloadResult, error) {
	result := downloadResult{}
	format, err := selectFormat(video, opts.AudioOnly)
	if err != nil {
		return result, err
	}

	outputPath, err := resolveOutputPath(opts.OutputTemplate, video, format, ctxInfo)
	if err != nil {
		return result, err
	}
	outputPath, skip, err := handleExistingPath(outputPath, opts)
	if err != nil {
		return result, err
	}
	if skip {
		result.skipped = true
		result.outputPath = outputPath
		return result, nil
	}
	result.outputPath = outputPath

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return result, fmt.Errorf("creating output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return result, fmt.Errorf("opening output file: %w", err)
	}
	defer file.Close()

	stream, size, err := client.GetStreamContext(ctx, video, format)
	if err != nil {
		return result, fmt.Errorf("starting stream: %w", err)
	}
	defer func() {
		if stream != nil {
			stream.Close()
		}
	}()

	var writer io.Writer = file
	var progress *progressWriter
	if !opts.Quiet {
		progress = newProgressWriter(size, printer, prefix)
		writer = io.MultiWriter(file, progress)
	}
	result.hadProgress = progress != nil

	written, err := copyWithContext(ctx, writer, stream)
	if err != nil {
		if isUnexpectedStatus(err, http.StatusForbidden) {
			if progress != nil {
				progress.NewLine()
			}
			if !opts.Quiet {
				fmt.Fprintln(os.Stderr, "warning: 403 from chunked download, retrying with single request")
			}
			if _, seekErr := file.Seek(0, 0); seekErr != nil {
				return result, fmt.Errorf("retry failed: %w", seekErr)
			}
			if truncErr := file.Truncate(0); truncErr != nil {
				return result, fmt.Errorf("retry failed: %w", truncErr)
			}

			formatSingle := *format
			formatSingle.ContentLength = 0

			stream.Close()
			stream, size, err = client.GetStreamContext(ctx, video, &formatSingle)
			if err != nil {
				return result, fmt.Errorf("retry failed: %w", err)
			}

			writer = file
			if !opts.Quiet {
				progress = newProgressWriter(size, printer, prefix)
				writer = io.MultiWriter(file, progress)
			} else {
				progress = nil
			}
			result.hadProgress = progress != nil

			written, err = copyWithContext(ctx, writer, stream)
			result.retried = true
		}
		if err != nil {
			return result, fmt.Errorf("download failed: %w", err)
		}
	}

	if progress != nil {
		progress.Finish()
	}

	result.bytes = written
	return result, nil
}

func selectFormat(video *youtube.Video, audioOnly bool) (*youtube.Format, error) {
	var best *youtube.Format

	for i := range video.Formats {
		format := &video.Formats[i]

		if audioOnly {
			if format.AudioChannels == 0 {
				continue
			}
			if format.Width != 0 || format.Height != 0 {
				continue
			}
		} else {
			if format.AudioChannels == 0 {
				continue
			}
			if format.Width == 0 || format.Height == 0 {
				continue
			}
		}

		if best == nil || betterFormat(format, best, audioOnly) {
			best = format
		}
	}

	if best == nil {
		if audioOnly {
			return nil, errors.New("no audio-only formats available")
		}
		return nil, errors.New("no progressive (audio+video) formats available")
	}
	return best, nil
}

func isUnexpectedStatus(err error, code int) bool {
	var statusErr youtube.ErrUnexpectedStatusCode
	if errors.As(err, &statusErr) {
		return int(statusErr) == code
	}
	return false
}

func betterFormat(candidate, current *youtube.Format, audioOnly bool) bool {
	if audioOnly {
		return bitrateForFormat(candidate) > bitrateForFormat(current)
	}

	if candidate.Height != current.Height {
		return candidate.Height > current.Height
	}
	return bitrateForFormat(candidate) > bitrateForFormat(current)
}

func entryTitle(entry *youtube.PlaylistEntry) string {
	if entry == nil {
		return ""
	}
	if entry.Title != "" {
		return entry.Title
	}
	return entry.ID
}

func handleExistingPath(path string, opts Options) (string, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, false, nil
		}
		return "", false, err
	}
	if info.IsDir() {
		return "", false, fmt.Errorf("output path is a directory: %s", path)
	}

	if !isTerminal(os.Stdin) {
		if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "warning: %s exists; overwriting (stdin not a TTY)\n", path)
		}
		return path, false, nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s exists. [o]verwrite, [s]kip, [r]ename, [q]uit? ", path)
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return "", false, readErr
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "o", "overwrite":
			return path, false, nil
		case "s", "skip":
			return path, true, nil
		case "r", "rename":
			newPath, err := nextAvailablePath(path)
			if err != nil {
				return "", false, err
			}
			return newPath, false, nil
		case "q", "quit":
			return "", false, errors.New("aborted by user")
		default:
			fmt.Fprintln(os.Stderr, "please enter o, s, r, or q")
		}
	}
}

func nextAvailablePath(path string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if _, err := os.Stat(candidate); err != nil {
			if os.IsNotExist(err) {
				return candidate, nil
			}
			return "", err
		}
	}
	return "", fmt.Errorf("unable to find available filename for %s", path)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func bitrateForFormat(f *youtube.Format) int {
	if f.Bitrate > 0 {
		return f.Bitrate
	}
	if f.AverageBitrate > 0 {
		return f.AverageBitrate
	}
	return 0
}

func resolveOutputPath(template string, video *youtube.Video, format *youtube.Format, ctxInfo outputContext) (string, error) {
	if template == "" {
		template = "{title}.{ext}"
	}

	title := sanitize(video.Title)
	ext := mimeToExt(format.MimeType)
	artist := video.Author
	album := ""
	quality := format.QualityLabel
	if quality == "" {
		if b := bitrateForFormat(format); b > 0 {
			quality = fmt.Sprintf("%dk", b/1000)
		}
	}
	playlistTitle := ""
	playlistID := ""
	index := ""
	total := ""
	if ctxInfo.Playlist != nil {
		playlistTitle = sanitize(ctxInfo.Playlist.Title)
		playlistID = ctxInfo.Playlist.ID
		album = playlistTitle
		if ctxInfo.Index > 0 {
			index = strconv.Itoa(ctxInfo.Index)
		}
		if ctxInfo.Total > 0 {
			total = strconv.Itoa(ctxInfo.Total)
		}
		if ctxInfo.EntryTitle != "" {
			title = sanitize(ctxInfo.EntryTitle)
		}
		if ctxInfo.EntryAuthor != "" {
			artist = ctxInfo.EntryAuthor
		}
	}
	artist = sanitize(artist)

	replacer := strings.NewReplacer(
		"{title}", title,
		"{artist}", artist,
		"{album}", album,
		"{id}", video.ID,
		"{ext}", ext,
		"{quality}", quality,
		"{playlist_title}", playlistTitle,
		"{playlist_id}", playlistID,
		"{index}", index,
		"{count}", total,
	)
	path := replacer.Replace(template)

	// Treat existing directory or explicit trailing slash as "put file inside".
	if strings.HasSuffix(template, "/") {
		path = filepath.Join(path, fmt.Sprintf("%s.%s", title, ext))
	} else if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, fmt.Sprintf("%s.%s", title, ext))
	}

	if filepath.Ext(path) == "" {
		path = path + "." + ext
	}
	return path, nil
}

func sanitize(name string) string {
	invalid := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	clean := invalid.ReplaceAllString(name, "-")
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return "video"
	}
	return clean
}

func mimeToExt(mime string) string {
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	parts := strings.Split(mime, "/")
	if len(parts) == 2 {
		switch parts[1] {
		case "3gpp":
			return "3gp"
		default:
			return parts[1]
		}
	}
	return "bin"
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for n >= unit*div && exp < 4 {
		div *= unit
		exp++
	}
	value := float64(n) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f%s", value, suffix[exp])
}

func printInfo(video *youtube.Video) error {
	type formatInfo struct {
		Itag         int    `json:"itag"`
		MimeType     string `json:"mime_type"`
		Quality      string `json:"quality"`
		QualityLabel string `json:"quality_label"`
		Bitrate      int    `json:"bitrate"`
		AvgBitrate   int    `json:"average_bitrate"`
		AudioQuality string `json:"audio_quality"`
		SampleRate   string `json:"audio_sample_rate"`
		Channels     int    `json:"audio_channels"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Size         int64  `json:"content_length"`
		Ext          string `json:"ext"`
	}

	payload := struct {
		ID       string       `json:"id"`
		Title    string       `json:"title"`
		Artist   string       `json:"artist"`
		Author   string       `json:"author"`
		Duration int          `json:"duration_seconds"`
		Formats  []formatInfo `json:"formats"`
	}{
		ID:       video.ID,
		Title:    video.Title,
		Artist:   video.Author,
		Author:   video.Author,
		Duration: int(video.Duration.Seconds()),
	}

	for _, f := range video.Formats {
		payload.Formats = append(payload.Formats, formatInfo{
			Itag:         f.ItagNo,
			MimeType:     f.MimeType,
			Quality:      f.Quality,
			QualityLabel: f.QualityLabel,
			Bitrate:      f.Bitrate,
			AvgBitrate:   f.AverageBitrate,
			AudioQuality: f.AudioQuality,
			SampleRate:   f.AudioSampleRate,
			Channels:     f.AudioChannels,
			Width:        f.Width,
			Height:       f.Height,
			Size:         int64(f.ContentLength),
			Ext:          mimeToExt(f.MimeType),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func printPlaylistInfo(playlist *youtube.Playlist) error {
	type entryInfo struct {
		Index    int    `json:"index"`
		ID       string `json:"id"`
		Title    string `json:"title"`
		Author   string `json:"author"`
		Duration int    `json:"duration_seconds"`
	}

	payload := struct {
		ID          string      `json:"id"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Author      string      `json:"author"`
		VideoCount  int         `json:"video_count"`
		Videos      []entryInfo `json:"videos"`
	}{
		ID:          playlist.ID,
		Title:       playlist.Title,
		Description: playlist.Description,
		Author:      playlist.Author,
		VideoCount:  len(playlist.Videos),
	}

	for i, entry := range playlist.Videos {
		if entry == nil {
			continue
		}
		payload.Videos = append(payload.Videos, entryInfo{
			Index:    i + 1,
			ID:       entry.ID,
			Title:    entry.Title,
			Author:   entry.Author,
			Duration: int(entry.Duration.Seconds()),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

var (
	playlistIDRegex  = regexp.MustCompile(`^[A-Za-z0-9_-]{13,42}$`)
	playlistURLRegex = regexp.MustCompile(`[?&]list=([A-Za-z0-9_-]{13,42})`)
)

func looksLikePlaylist(url string) bool {
	return playlistIDRegex.MatchString(url) || playlistURLRegex.MatchString(url)
}

// ConvertMusicURL converts YouTube Music URLs to regular YouTube URLs
func ConvertMusicURL(u string) string {
	// Parse the URL
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}

	// If it's not a music.youtube.com URL, return as-is
	if parsed.Host != "music.youtube.com" {
		return u
	}

	// Replace the host with www.youtube.com
	parsed.Host = "www.youtube.com"

	// Remove any music-specific parameters
	query := parsed.Query()
	delete(query, "si")
	parsed.RawQuery = query.Encode()

	return parsed.String()
}
