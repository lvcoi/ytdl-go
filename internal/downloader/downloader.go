package downloader

import (
	"bufio"
	"bytes"
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
	"text/tabwriter"
	"time"

	"github.com/kkdai/youtube/v2"
)

// duplicateAction represents the user's choice for handling existing files
type duplicateAction int

const (
	duplicateAskEach duplicateAction = iota
	duplicateOverwriteAll
	duplicateSkipAll
	duplicateRenameAll
)

var globalDuplicateAction = duplicateAskEach

// ResetDuplicateAction resets the global duplicate handling preference (for testing or new sessions)
func ResetDuplicateAction() {
	globalDuplicateAction = duplicateAskEach
}

// Options describes CLI behavior for a download run.
type Options struct {
	OutputTemplate     string
	AudioOnly          bool
	InfoOnly           bool
	ListFormats        bool
	Quiet              bool
	JSON               bool
	Quality            string
	Format             string
	Timeout            time.Duration
	ProgressLayout     string
	LogLevel           string
	SegmentConcurrency int
	MetaOverrides      map[string]string
}

type outputContext struct {
	Playlist      *youtube.Playlist
	Index         int
	Total         int
	EntryTitle    string
	EntryAuthor   string
	EntryAlbum    string
	SourceURL     string
	PlaylistURL   string
	MetaOverrides map[string]string
}

type downloadResult struct {
	bytes       int64
	outputPath  string
	retried     bool
	hadProgress bool
	skipped     bool
}

type jsonResult struct {
	Type          string `json:"type"`
	Status        string `json:"status"`
	URL           string `json:"url,omitempty"`
	ID            string `json:"id,omitempty"`
	Title         string `json:"title,omitempty"`
	Output        string `json:"output,omitempty"`
	Bytes         int64  `json:"bytes,omitempty"`
	Retries       bool   `json:"retried,omitempty"`
	Skipped       bool   `json:"skipped,omitempty"`
	Error         string `json:"error,omitempty"`
	PlaylistID    string `json:"playlist_id,omitempty"`
	PlaylistTitle string `json:"playlist_title,omitempty"`
	Index         int    `json:"index,omitempty"`
	Total         int    `json:"total,omitempty"`
}

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

func emitJSONResult(res jsonResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(res)
}

func validateInputURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid URL: %w", err))
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid URL: missing scheme or host"))
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme))
	}
	return parsed.String(), nil
}

func categoryForYouTubeError(err error) ErrorCategory {
	switch {
	case errors.Is(err, youtube.ErrLoginRequired),
		errors.Is(err, youtube.ErrVideoPrivate),
		errors.Is(err, youtube.ErrNotPlayableInEmbed):
		return CategoryRestricted
	case errors.Is(err, youtube.ErrInvalidPlaylist),
		errors.Is(err, youtube.ErrInvalidCharactersInVideoID),
		errors.Is(err, youtube.ErrVideoIDMinLength):
		return CategoryInvalidURL
	}

	var statusErr *youtube.ErrPlayabiltyStatus
	if errors.As(err, &statusErr) {
		return CategoryRestricted
	}

	return CategoryNetwork
}

func wrapFetchError(err error, context string) error {
	category := categoryForYouTubeError(err)
	wrapped := fmt.Errorf("%s: %w", context, err)
	switch category {
	case CategoryRestricted:
		wrapped = fmt.Errorf("restricted content (login/paywall/age/private): %w", wrapped)
	case CategoryInvalidURL:
		wrapped = fmt.Errorf("invalid URL or playlist ID: %w", wrapped)
	}
	return wrapCategory(category, wrapped)
}

// Process fetches metadata, selects the best matching format, and downloads it.
func Process(ctx context.Context, url string, opts Options) error {
	progressManager := newProgressManager(opts)
	if progressManager != nil {
		progressManager.Start(ctx)
		defer progressManager.Stop()
	}
	printer := newPrinter(opts, progressManager)

	normalizedURL, err := validateInputURL(url)
	if err != nil {
		return err
	}

	// Check if it's a music URL before converting
	isMusicURL := strings.Contains(url, "music.youtube.com")

	// Convert YouTube Music URLs to regular YouTube URLs
	normalizedURL = ConvertMusicURL(normalizedURL)

	if looksLikePlaylist(normalizedURL) {
		if playlistIDRegex.MatchString(normalizedURL) {
			return processPlaylist(ctx, normalizedURL, opts, printer, isMusicURL)
		}
		if err := validateURL(normalizedURL); err != nil {
			return err
		}
		return processPlaylist(ctx, normalizedURL, opts, printer, isMusicURL)
	}
	if err := validateURL(normalizedURL); err != nil {
		return err
	}

	extractor, err := selectExtractor(normalizedURL)
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

func renderFormats(video *youtube.Video, header string, opts Options, playlistID, playlistTitle string, index, total int) error {
	if opts.JSON {
		payload := struct {
			Type          string       `json:"type"`
			PlaylistID    string       `json:"playlist_id,omitempty"`
			PlaylistTitle string       `json:"playlist_title,omitempty"`
			Index         int          `json:"index,omitempty"`
			Total         int          `json:"total,omitempty"`
			ID            string       `json:"id"`
			Title         string       `json:"title"`
			Formats       []formatInfo `json:"formats"`
		}{
			Type:          "formats",
			PlaylistID:    playlistID,
			PlaylistTitle: playlistTitle,
			Index:         index,
			Total:         total,
			ID:            video.ID,
			Title:         video.Title,
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
		enc.SetEscapeHTML(false)
		return enc.Encode(payload)
	}

	fmt.Fprintln(os.Stdout, header)
	fmt.Fprintln(os.Stdout, "itag  ext   quality    size     audio  video")
	for _, f := range video.Formats {
		size := "-"
		if f.ContentLength > 0 {
			size = humanBytes(int64(f.ContentLength))
		}
		audio := "-"
		if f.AudioChannels > 0 {
			audio = fmt.Sprintf("%dch", f.AudioChannels)
		}
		videoRes := "-"
		if f.Width > 0 || f.Height > 0 {
			videoRes = fmt.Sprintf("%dx%d", f.Width, f.Height)
		}
		qual := f.QualityLabel
		if qual == "" {
			qual = f.Quality
		}
		fmt.Fprintf(os.Stdout, "%4d  %-4s %-10s %-8s %-5s %-8s\n",
			f.ItagNo,
			mimeToExt(f.MimeType),
			qual,
			size,
			audio,
			videoRes,
		)
	}
	return nil
}

func listFormats(video *youtube.Video, opts Options, _ *Printer) error {
	header := fmt.Sprintf("Formats for %s (%s)", video.Title, video.ID)
	return renderFormats(video, header, opts, "", "", 0, 0)
}

func listPlaylistFormats(ctx context.Context, playlist *youtube.Playlist, opts Options, _ *Printer) error {
	if len(playlist.Videos) == 0 {
		return wrapCategory(CategoryUnsupported, errors.New("playlist has no videos"))
	}

	youtube.DefaultClient = youtube.AndroidClient
	client := newClient(opts)
	for i, entry := range playlist.Videos {
		if entry == nil || entry.ID == "" {
			continue
		}
		video, err := client.VideoFromPlaylistEntryContext(ctx, entry)
		if err != nil {
			return wrapFetchError(err, "fetching video metadata")
		}
		header := fmt.Sprintf("[%d/%d] %s (%s)", i+1, len(playlist.Videos), entryTitle(entry), entry.ID)
		if err := renderFormats(video, header, opts, playlist.ID, playlist.Title, i+1, len(playlist.Videos)); err != nil {
			return err
		}
		if !opts.JSON {
			fmt.Fprintln(os.Stdout)
		}
	}
	return nil
}

func newClient(opts Options) *youtube.Client {
	return &youtube.Client{
		HTTPClient: &http.Client{Timeout: opts.Timeout},
	}
}

func validateURL(input string) error {
	if playlistIDRegex.MatchString(input) {
		return nil
	}
	parsed, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid URL: unsupported scheme %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return errors.New("invalid URL: missing host")
	}
	return nil
}

func wrapAccessError(err error) error {
	if err == nil {
		return nil
	}
	if isRestrictedAccess(err) {
		return fmt.Errorf("restricted access: %w", err)
	}
	return err
}

func isRestrictedAccess(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	restrictedMarkers := []string{
		"private",
		"sign in",
		"login",
		"members only",
		"premium",
		"copyright",
		"video unavailable",
		"content unavailable",
		"age-restricted",
		"age restricted",
		"not available",
	}
	for _, marker := range restrictedMarkers {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

func processPlaylist(ctx context.Context, url string, opts Options, printer *Printer, isMusicURL bool) error {
	savedClient := youtube.DefaultClient
	defer func() {
		youtube.DefaultClient = savedClient
	}()

	youtube.DefaultClient = youtube.WebClient
	playlistClient := newClient(opts)
	playlist, err := playlistClient.GetPlaylistContext(ctx, url)
	if err != nil {
		return wrapAccessError(fmt.Errorf("fetching playlist: %w", err))
	}

	if opts.InfoOnly {
		return printPlaylistInfo(playlist)
	}
	if opts.ListFormats {
		return listPlaylistFormats(ctx, playlist, opts, printer)
	}

	if len(playlist.Videos) == 0 {
		return wrapCategory(CategoryUnsupported, errors.New("playlist has no videos"))
	}

	// Fetch playlist title from YouTube Music (the library often returns empty/generic titles)
	if isMusicURL {
		if title, err := fetchMusicPlaylistTitle(ctx, playlist.ID, opts.Timeout); err == nil && title != "" {
			playlist.Title = title
		}
	}
	if playlist.Title == "" || playlist.Title == "Playlist" {
		playlist.Title = "Playlist"
	}

	albumMeta := map[string]musicEntryMeta{}
	if strings.Contains(opts.OutputTemplate, "{album}") {
		var err error
		albumMeta, err = fetchMusicPlaylistEntries(ctx, playlist.ID, opts)
		if err != nil {
			return wrapCategory(CategoryNetwork, fmt.Errorf("fetching album metadata: %w", err))
		}
	}

	printer.Log(LogInfo, fmt.Sprintf("playlist: %s (%d videos)", playlist.Title, len(playlist.Videos)))

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
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "skip",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            "",
					Title:         "",
					Error:         "missing playlist entry",
				})
			}
			continue
		}

		video, err := videoClient.VideoFromPlaylistEntryContext(ctx, entry)
		if err != nil {
			err = wrapFetchError(err, "fetching video metadata")
			failures++
			printer.ItemResult(prefix, downloadResult{}, err)
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "error",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            entry.ID,
					Title:         entryTitle(entry),
					Error:         err.Error(),
				})
			}
			continue
		}

		meta := albumMeta[entry.ID]
		entryTitle := entry.Title
		entryAuthor := entry.Author
		if meta.Title != "" {
			entryTitle = meta.Title
		}
		if meta.Artist != "" {
			entryAuthor = meta.Artist
		}

		result, err := downloadVideo(ctx, videoClient, video, opts, outputContext{
			Playlist:      playlist,
			Index:         i + 1,
			Total:         len(playlist.Videos),
			EntryTitle:    entryTitle,
			EntryAuthor:   entryAuthor,
			EntryAlbum:    meta.Album,
			SourceURL:     watchURLForID(entry.ID),
			PlaylistURL:   url,
			MetaOverrides: opts.MetaOverrides,
		}, printer, prefix)
		if result.skipped {
			skipped++
			printer.ItemSkipped(prefix, "exists")
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "skip",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            entry.ID,
					Title:         entryTitle,
					Output:        result.outputPath,
					Bytes:         result.bytes,
					Retries:       result.retried,
					Error:         "exists",
				})
			}
			continue
		}
		printer.ItemResult(prefix, result, err)
		if opts.JSON {
			status := "ok"
			errMsg := ""
			if err != nil {
				status = "error"
				errMsg = err.Error()
			}
			emitJSONResult(jsonResult{
				Type:          "item",
				Status:        status,
				PlaylistID:    playlist.ID,
				PlaylistTitle: playlist.Title,
				Index:         i + 1,
				ID:            entry.ID,
				Title:         entryTitle,
				Output:        result.outputPath,
				Bytes:         result.bytes,
				Retries:       result.retried,
				Error:         errMsg,
			})
		}

		if err != nil {
			failures++
			continue
		}

		successes++
		totalBytes += result.bytes
	}

	printer.Summary(len(playlist.Videos), successes, failures, skipped, totalBytes)
	if successes == 0 {
		return markReported(wrapCategory(CategoryUnsupported, errors.New("no playlist entries downloaded successfully")))
	}
	return nil
}

func downloadVideo(ctx context.Context, client *youtube.Client, video *youtube.Video, opts Options, ctxInfo outputContext, printer *Printer, prefix string) (result downloadResult, err error) {
	var (
		format     *youtube.Format
		outputPath string
	)
	defer func() {
		if outputPath == "" {
			return
		}
		status := "ok"
		if result.skipped {
			status = "skipped"
		} else if err != nil {
			status = "error"
		}
		metadata := buildItemMetadata(video, format, ctxInfo, outputPath, status, err)
		if opts.AudioOnly {
			embedAudioTags(metadata, outputPath, printer)
		}
	}()

	result = downloadResult{}
	format, err = selectFormat(video, opts)
	if err != nil {
		if errorCategory(err) == CategoryUnsupported {
			if video.HLSManifestURL != "" || video.DASHManifestURL != "" {
				return downloadAdaptive(ctx, client, video, opts, ctxInfo, printer, prefix, err)
			}
		}
		return result, err
	}

	outputPath, err = resolveOutputPath(opts.OutputTemplate, video, format, ctxInfo)
	if err != nil {
		return result, wrapCategory(CategoryFilesystem, err)
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
		return result, wrapCategory(CategoryFilesystem, fmt.Errorf("creating output directory: %w", err))
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return result, wrapCategory(CategoryFilesystem, fmt.Errorf("opening output file: %w", err))
	}
	defer file.Close()

	stream, size, err := client.GetStreamContext(ctx, video, format)
	if err != nil {
		return result, wrapCategory(CategoryNetwork, fmt.Errorf("starting stream: %w", err))
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
			printer.Log(LogWarn, "warning: 403 from chunked download, retrying with single request")
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
				return result, wrapCategory(CategoryNetwork, fmt.Errorf("retry failed: %w", err))
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
			return result, wrapCategory(CategoryNetwork, fmt.Errorf("download failed: %w", err))
		}
	}

	if progress != nil {
		progress.Finish()
	}

	if err := validateOutputFile(outputPath, format); err != nil {
		return result, err
	}

	result.bytes = written
	return result, nil
}

func selectFormat(video *youtube.Video, opts Options) (*youtube.Format, error) {
	candidates := make([]*youtube.Format, 0, len(video.Formats))
	for i := range video.Formats {
		format := &video.Formats[i]

		if opts.AudioOnly {
			if format.AudioChannels == 0 {
				continue
			}
			if format.Width != 0 || format.Height != 0 {
				continue
			}
		} else {
			if format.AudioChannels == 0 || format.Width == 0 || format.Height == 0 {
				continue
			}
		}

		if opts.Format != "" && !formatMatches(format, opts.Format) {
			continue
		}

		candidates = append(candidates, format)
	}

	if len(candidates) == 0 {
		// If format was specified but not found, try again without format filter (fallback)
		if opts.Format != "" {
			fallbackOpts := opts
			fallbackOpts.Format = ""
			fallbackFormat, err := selectFormat(video, fallbackOpts)
			if err == nil && fallbackFormat != nil {
				// Found a fallback format - return it (caller should warn user)
				return fallbackFormat, nil
			}
		}

		reason := "no progressive (audio+video) formats available"
		if opts.AudioOnly {
			reason = "no audio-only formats available (try without --audio or use --list-formats)"
		}
		if opts.Format != "" {
			reason = fmt.Sprintf("no formats available for requested format %s (use --list-formats)", opts.Format)
		} else if !opts.AudioOnly {
			reason = "no progressive (audio+video) formats available (try --audio or use --list-formats)"
		}
		return nil, wrapCategory(CategoryUnsupported, errors.New(reason))
	}

	if opts.AudioOnly {
		return pickAudioFormat(candidates, opts)
	}
	return pickVideoFormat(candidates, opts)
}

func pickVideoFormat(candidates []*youtube.Format, opts Options) (*youtube.Format, error) {
	targetHeight, preferLowest, err := parseVideoQuality(opts.Quality)
	if err != nil {
		return nil, wrapCategory(CategoryUnsupported, err)
	}

	// Prefer highest available when no explicit target.
	if targetHeight == 0 && !preferLowest {
		var best *youtube.Format
		for _, f := range candidates {
			if best == nil || betterVideoFormat(f, best) {
				best = f
			}
		}
		return best, nil
	}

	var best *youtube.Format
	if targetHeight > 0 {
		for _, f := range candidates {
			if f.Height == 0 || f.Height > targetHeight {
				continue
			}
			if best == nil || f.Height > best.Height || (f.Height == best.Height && bitrateForFormat(f) > bitrateForFormat(best)) {
				best = f
			}
		}
	}

	if best == nil && targetHeight > 0 {
		// No option under target: pick the closest above target.
		for _, f := range candidates {
			if best == nil || f.Height < best.Height || (f.Height == best.Height && bitrateForFormat(f) > bitrateForFormat(best)) {
				best = f
			}
		}
	}

	if best == nil && preferLowest {
		for _, f := range candidates {
			if best == nil || f.Height < best.Height || (f.Height == best.Height && bitrateForFormat(f) > bitrateForFormat(best)) {
				best = f
			}
		}
	}

	if best == nil {
		return nil, wrapCategory(CategoryUnsupported, errors.New("no matching video formats available (use --list-formats)"))
	}

	return best, nil
}

func pickAudioFormat(candidates []*youtube.Format, opts Options) (*youtube.Format, error) {
	targetBitrate, preferLowest, err := parseAudioQuality(opts.Quality)
	if err != nil {
		return nil, wrapCategory(CategoryUnsupported, err)
	}

	var best *youtube.Format
	if targetBitrate > 0 {
		for _, f := range candidates {
			br := bitrateForFormat(f)
			if br == 0 || br > targetBitrate {
				continue
			}
			if best == nil || br > bitrateForFormat(best) {
				best = f
			}
		}
		// Nothing under target? fall back to the lowest above target.
		if best == nil {
			for _, f := range candidates {
				br := bitrateForFormat(f)
				if br == 0 {
					continue
				}
				if best == nil || br < bitrateForFormat(best) {
					best = f
				}
			}
		}
	}

	if best == nil && preferLowest {
		for _, f := range candidates {
			br := bitrateForFormat(f)
			if br == 0 {
				continue
			}
			if best == nil || br < bitrateForFormat(best) {
				best = f
			}
		}
	}

	if best == nil {
		for _, f := range candidates {
			if best == nil || bitrateForFormat(f) > bitrateForFormat(best) {
				best = f
			}
		}
	}

	if best == nil {
		return nil, wrapCategory(CategoryUnsupported, errors.New("no matching audio formats available (use --list-formats)"))
	}
	return best, nil
}

func parseVideoQuality(q string) (target int, preferLowest bool, err error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" || q == "best" {
		return 0, false, nil
	}
	if q == "worst" {
		return 0, true, nil
	}
	q = strings.TrimSuffix(q, "p")
	value, convErr := strconv.Atoi(q)
	if convErr != nil {
		return 0, false, fmt.Errorf("invalid quality value %q (expected like 720p)", q)
	}
	return value, false, nil
}

func parseAudioQuality(q string) (bitrate int, preferLowest bool, err error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" || q == "best" {
		return 0, false, nil
	}
	if q == "worst" {
		return 0, true, nil
	}

	q = strings.TrimSuffix(q, "bps")
	q = strings.TrimSuffix(q, "kbps")
	hasK := strings.HasSuffix(q, "k")
	q = strings.TrimSuffix(q, "k")
	value, convErr := strconv.Atoi(q)
	if convErr != nil {
		return 0, false, fmt.Errorf("invalid audio quality %q (expected like 128k)", q)
	}
	if hasK || value < 1000 {
		value = value * 1000
	}
	return value, false, nil
}

func formatMatches(format *youtube.Format, desired string) bool {
	return strings.EqualFold(mimeToExt(format.MimeType), strings.TrimSpace(strings.ToLower(desired)))
}

func isUnexpectedStatus(err error, code int) bool {
	var statusErr youtube.ErrUnexpectedStatusCode
	if errors.As(err, &statusErr) {
		return int(statusErr) == code
	}
	return false
}

func betterVideoFormat(candidate, current *youtube.Format) bool {
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

func watchURLForID(id string) string {
	if id == "" {
		return ""
	}
	return "https://www.youtube.com/watch?v=" + id
}

func handleExistingPath(path string, opts Options) (string, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, false, nil
		}
		return "", false, wrapCategory(CategoryFilesystem, err)
	}
	if info.IsDir() {
		return "", false, wrapCategory(CategoryFilesystem, fmt.Errorf("output path is a directory: %s", path))
	}

	// Check if we have a global "apply to all" action
	switch globalDuplicateAction {
	case duplicateOverwriteAll:
		return path, false, nil
	case duplicateSkipAll:
		return path, true, nil
	case duplicateRenameAll:
		newPath, err := nextAvailablePath(path)
		return newPath, false, err
	}

	if !isTerminal(os.Stdin) {
		if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "warning: %s exists; overwriting (stdin not a TTY)\n", path)
		}
		return path, false, nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s exists.\n", path)
		fmt.Fprint(os.Stderr, "  [o]verwrite, [s]kip, [r]ename, [q]uit\n")
		fmt.Fprint(os.Stderr, "  [O]verwrite all, [S]kip all, [R]ename all: ")
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return "", false, wrapCategory(CategoryFilesystem, readErr)
		}
		switch strings.TrimSpace(line) {
		case "o", "overwrite":
			return path, false, nil
		case "O", "Overwrite":
			globalDuplicateAction = duplicateOverwriteAll
			return path, false, nil
		case "s", "skip":
			return path, true, nil
		case "S", "Skip":
			globalDuplicateAction = duplicateSkipAll
			return path, true, nil
		case "r", "rename":
			newPath, err := nextAvailablePath(path)
			return newPath, false, err
		case "R", "Rename":
			globalDuplicateAction = duplicateRenameAll
			newPath, err := nextAvailablePath(path)
			return newPath, false, err
		case "q", "quit":
			return "", false, errors.New("aborted by user")
		default:
			fmt.Fprintln(os.Stderr, "please enter o, s, r, q (or uppercase for 'all')")
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
			return "", wrapCategory(CategoryFilesystem, err)
		}
	}
	return "", wrapCategory(CategoryFilesystem, fmt.Errorf("unable to find available filename for %s", path))
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
	album := ctxInfo.EntryAlbum
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
	album = sanitizeOptional(album)

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

func sanitizeOptional(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	return sanitize(name)
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
	enc.SetEscapeHTML(false)
	return enc.Encode(payload)
}

func writeFormats(output io.Writer, video *youtube.Video) error {
	writer := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "itag\tquality\tquality_label\tmime_type\text\tbitrate\tchannels\tresolution\tsize")
	for _, f := range video.Formats {
		resolution := ""
		if f.Width > 0 && f.Height > 0 {
			resolution = fmt.Sprintf("%dx%d", f.Width, f.Height)
		}
		fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%s\t%d\t%d\t%s\t%d\n",
			f.ItagNo,
			f.Quality,
			f.QualityLabel,
			f.MimeType,
			mimeToExt(f.MimeType),
			bitrateForFormat(&f),
			f.AudioChannels,
			resolution,
			f.ContentLength,
		)
	}
	return writer.Flush()
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

func isYouTubeURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	host = strings.TrimPrefix(host, "www.")
	return host == "youtube.com" || host == "youtu.be" || host == "music.youtube.com"
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

const musicUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

type musicEntryMeta struct {
	Title  string
	Artist string
	Album  string
}

type musicConfig struct {
	apiKey  string
	context map[string]any
}

func fetchMusicPlaylistEntries(ctx context.Context, playlistID string, opts Options) (map[string]musicEntryMeta, error) {
	cfg, err := fetchMusicConfig(ctx, playlistID, opts.Timeout)
	if err != nil {
		return nil, err
	}

	browseID := playlistBrowseID(playlistID)
	payload := map[string]any{
		"context":  cfg.context,
		"browseId": browseID,
	}

	result := map[string]musicEntryMeta{}
	response, err := fetchMusicBrowse(ctx, cfg.apiKey, payload, opts.Timeout)
	if err != nil {
		return nil, err
	}

	items, continuation, err := extractMusicPlaylistItems(response)
	if err != nil {
		return nil, err
	}
	appendMusicEntries(result, items)

	for continuation != "" {
		payload = map[string]any{
			"context":      cfg.context,
			"continuation": continuation,
		}
		response, err = fetchMusicBrowse(ctx, cfg.apiKey, payload, opts.Timeout)
		if err != nil {
			return nil, err
		}
		items, continuation, err = extractMusicPlaylistItems(response)
		if err != nil {
			return nil, err
		}
		appendMusicEntries(result, items)
	}

	if len(result) == 0 {
		return nil, errors.New("no playlist entries found in YouTube Music response")
	}
	return result, nil
}

func fetchMusicConfig(ctx context.Context, playlistID string, timeout time.Duration) (musicConfig, error) {
	playlistURL := "https://music.youtube.com/playlist?list=" + url.QueryEscape(playlistID)
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
	if err != nil {
		return musicConfig{}, err
	}
	req.Header.Set("User-Agent", musicUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return musicConfig{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return musicConfig{}, fmt.Errorf("unexpected response %d from YouTube Music", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return musicConfig{}, err
	}

	re := regexp.MustCompile(`(?s)ytcfg\.set\((\{.*?\})\);`)
	match := re.FindSubmatch(body)
	if match == nil {
		return musicConfig{}, errors.New("ytcfg.set data not found in YouTube Music page")
	}

	var cfg struct {
		APIKey  string         `json:"INNERTUBE_API_KEY"`
		Context map[string]any `json:"INNERTUBE_CONTEXT"`
	}
	if err := json.Unmarshal(match[1], &cfg); err != nil {
		return musicConfig{}, err
	}
	if cfg.APIKey == "" || len(cfg.Context) == 0 {
		return musicConfig{}, errors.New("missing innertube config in YouTube Music page")
	}

	return musicConfig{
		apiKey:  cfg.APIKey,
		context: cfg.Context,
	}, nil
}

func fetchMusicPlaylistTitle(ctx context.Context, playlistID string, timeout time.Duration) (string, error) {
	// Fetch the HTML page and extract og:title meta tag
	playlistURL := "https://music.youtube.com/playlist?list=" + url.QueryEscape(playlistID)
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", musicUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response %d from YouTube Music", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try og:title meta tag first (most reliable) - handle attribute order variations
	ogTitleRe := regexp.MustCompile(`(?i)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	if match := ogTitleRe.FindSubmatch(body); match != nil {
		return string(match[1]), nil
	}
	// Try reverse order (content before property)
	ogTitleRe2 := regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:title["']`)
	if match := ogTitleRe2.FindSubmatch(body); match != nil {
		return string(match[1]), nil
	}

	// Try <title> tag as fallback
	titleRe := regexp.MustCompile(`<title>([^<]+)</title>`)
	if match := titleRe.FindSubmatch(body); match != nil {
		title := strings.TrimSuffix(string(match[1]), " - YouTube Music")
		title = strings.TrimSuffix(title, " - YouTube")
		if title != "" {
			return title, nil
		}
	}

	return "", errors.New("playlist title not found in YouTube Music page")
}

func fetchMusicBrowse(ctx context.Context, apiKey string, payload map[string]any, timeout time.Duration) (map[string]any, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: timeout}
	endpoint := "https://music.youtube.com/youtubei/v1/browse?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", musicUserAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response %d from YouTube Music browse", resp.StatusCode)
	}

	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func playlistBrowseID(playlistID string) string {
	if strings.HasPrefix(playlistID, "VL") {
		return playlistID
	}
	return "VL" + playlistID
}

func extractMusicPlaylistItems(payload map[string]any) ([]any, string, error) {
	if shelf := findMusicPlaylistShelf(payload); shelf != nil {
		items := asSlice(shelf["contents"])
		if len(items) == 0 {
			return nil, "", errors.New("no playlist entries in YouTube Music shelf")
		}
		return items, findContinuationToken(items), nil
	}

	if cont := asMap(getPath(payload, "continuationContents", "musicPlaylistShelfContinuation")); cont != nil {
		items := asSlice(cont["contents"])
		if len(items) > 0 {
			return items, findContinuationToken(items), nil
		}
	}

	actions := asSlice(getPath(payload, "onResponseReceivedActions"))
	for _, action := range actions {
		appendAction := asMap(asMap(action)["appendContinuationItemsAction"])
		if appendAction == nil {
			continue
		}
		items := asSlice(appendAction["continuationItems"])
		if len(items) == 0 {
			continue
		}
		return items, findContinuationToken(items), nil
	}

	return nil, "", errors.New("no playlist items found in YouTube Music response")
}

func findMusicPlaylistShelf(payload map[string]any) map[string]any {
	queue := []any{payload}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		switch value := node.(type) {
		case map[string]any:
			if shelf := asMap(value["musicPlaylistShelfRenderer"]); shelf != nil {
				return shelf
			}
			for _, child := range value {
				switch child.(type) {
				case map[string]any, []any:
					queue = append(queue, child)
				}
			}
		case []any:
			for _, child := range value {
				switch child.(type) {
				case map[string]any, []any:
					queue = append(queue, child)
				}
			}
		}
	}
	return nil
}

func findContinuationToken(items []any) string {
	for _, item := range items {
		cont := asMap(asMap(item)["continuationItemRenderer"])
		if cont == nil {
			continue
		}
		token := getString(getPath(cont, "continuationEndpoint", "continuationCommand", "token"))
		if token != "" {
			return token
		}
	}
	return ""
}

func appendMusicEntries(dest map[string]musicEntryMeta, items []any) {
	for _, item := range items {
		renderer := asMap(asMap(item)["musicResponsiveListItemRenderer"])
		if renderer == nil {
			continue
		}
		videoID, meta := parseMusicEntry(renderer)
		if videoID == "" {
			continue
		}
		dest[videoID] = meta
	}
}

func parseMusicEntry(renderer map[string]any) (string, musicEntryMeta) {
	videoID := getString(getPath(renderer, "playlistItemData", "videoId"))
	if videoID == "" {
		videoID = findVideoIDFromRuns(renderer)
	}

	meta := musicEntryMeta{
		Title:  columnText(renderer, 0),
		Artist: findRunTextByPageType(renderer, "MUSIC_PAGE_TYPE_ARTIST"),
		Album:  findRunTextByPageType(renderer, "MUSIC_PAGE_TYPE_ALBUM"),
	}
	if meta.Title == "" {
		meta.Title = findRunTextWithWatchEndpoint(renderer)
	}
	if meta.Artist == "" {
		meta.Artist = columnText(renderer, 1)
	}
	if meta.Album == "" {
		meta.Album = columnText(renderer, 2)
	}

	return videoID, meta
}

func findRunTextByPageType(renderer map[string]any, pageType string) string {
	for _, group := range []string{"flexColumns", "secondaryFlexColumns"} {
		for _, col := range asSlice(renderer[group]) {
			colRenderer := asMap(asMap(col)["musicResponsiveListItemFlexColumnRenderer"])
			if colRenderer == nil {
				continue
			}
			runs := asSlice(getPath(colRenderer, "text", "runs"))
			for _, run := range runs {
				runMap := asMap(run)
				if getString(getPath(runMap, "navigationEndpoint", "browseEndpoint", "browseEndpointContextSupportedConfigs", "browseEndpointContextMusicConfig", "pageType")) == pageType {
					if text := getString(runMap["text"]); text != "" {
						return text
					}
				}
			}
		}
	}
	return ""
}

func findRunTextWithWatchEndpoint(renderer map[string]any) string {
	for _, group := range []string{"flexColumns", "secondaryFlexColumns"} {
		for _, col := range asSlice(renderer[group]) {
			colRenderer := asMap(asMap(col)["musicResponsiveListItemFlexColumnRenderer"])
			if colRenderer == nil {
				continue
			}
			runs := asSlice(getPath(colRenderer, "text", "runs"))
			for _, run := range runs {
				runMap := asMap(run)
				if getString(getPath(runMap, "navigationEndpoint", "watchEndpoint", "videoId")) != "" {
					if text := getString(runMap["text"]); text != "" {
						return text
					}
				}
			}
		}
	}
	return ""
}

func findVideoIDFromRuns(renderer map[string]any) string {
	for _, group := range []string{"flexColumns", "secondaryFlexColumns"} {
		for _, col := range asSlice(renderer[group]) {
			colRenderer := asMap(asMap(col)["musicResponsiveListItemFlexColumnRenderer"])
			if colRenderer == nil {
				continue
			}
			runs := asSlice(getPath(colRenderer, "text", "runs"))
			for _, run := range runs {
				runMap := asMap(run)
				if videoID := getString(getPath(runMap, "navigationEndpoint", "watchEndpoint", "videoId")); videoID != "" {
					return videoID
				}
			}
		}
	}
	return ""
}

func columnText(renderer map[string]any, index int) string {
	cols := asSlice(renderer["flexColumns"])
	if index < 0 || index >= len(cols) {
		return ""
	}
	colRenderer := asMap(asMap(cols[index])["musicResponsiveListItemFlexColumnRenderer"])
	if colRenderer == nil {
		return ""
	}
	return extractText(colRenderer["text"])
}

func extractText(value any) string {
	textMap := asMap(value)
	if textMap == nil {
		if s, ok := value.(string); ok {
			return s
		}
		return ""
	}
	if runs := asSlice(textMap["runs"]); len(runs) > 0 {
		return runsText(runs)
	}
	if s, ok := textMap["simpleText"].(string); ok {
		return s
	}
	return ""
}

func runsText(runs []any) string {
	var b strings.Builder
	for _, run := range runs {
		runMap := asMap(run)
		if runMap == nil {
			continue
		}
		if text, ok := runMap["text"].(string); ok {
			b.WriteString(text)
		}
	}
	return b.String()
}

func asMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	m, _ := value.(map[string]any)
	return m
}

func asSlice(value any) []any {
	if value == nil {
		return nil
	}
	s, _ := value.([]any)
	return s
}

func getPath(value map[string]any, keys ...string) any {
	var current any = value
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[key]
	}
	return current
}

func getString(value any) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}
