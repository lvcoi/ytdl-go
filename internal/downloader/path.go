package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lvcoi/ytdl-lib/v2"
)

func resolveOutputPath(template string, video *youtube.Video, format *youtube.Format, ctxInfo outputContext, baseDir string) (string, error) {
	if template == "" {
		template = "{title}.{ext}"
	}

	title := sanitize(video.Title)
	videoID := sanitize(video.ID)
	ext := mimeToExt(format.MimeType)
	artist := video.Author
	album := ctxInfo.EntryAlbum
	quality := format.QualityLabel
	if quality == "" {
		if b := bitrateForFormat(format); b > 0 {
			quality = fmt.Sprintf("%dk", b/1000)
		}
	} else {
		quality = sanitize(quality)
	}
	playlistTitle := ""
	playlistID := ""
	index := ""
	total := ""
	if ctxInfo.Playlist != nil {
		playlistTitle = sanitize(ctxInfo.Playlist.Title)
		playlistID = sanitize(ctxInfo.Playlist.ID)
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

	// yt-dlp style extended fields
	uploadDate := ""
	uploadYear := ""
	if !video.PublishDate.IsZero() {
		uploadDate = video.PublishDate.Format("20060102")
		uploadYear = strconv.Itoa(video.PublishDate.Year())
	}
	viewCount := ""
	if video.Views > 0 {
		viewCount = strconv.Itoa(video.Views)
	}
	duration := ""
	durationString := ""
	if video.Duration > 0 {
		duration = strconv.FormatInt(int64(video.Duration.Seconds()), 10)
		durationString = formatDurationString(video.Duration)
	}
	resolution := ""
	if format.Width > 0 && format.Height > 0 {
		resolution = fmt.Sprintf("%dx%d", format.Width, format.Height)
	}
	fps := ""
	if format.FPS > 0 {
		fps = strconv.Itoa(format.FPS)
	}
	vcodec := sanitizeOptional(formatVideoCodec(format))
	acodec := sanitizeOptional(formatAudioCodec(format))
	bitrate := ""
	if b := bitrateForFormat(format); b > 0 {
		bitrate = strconv.Itoa(b)
	}
	autonumber := ""
	if ctxInfo.Autonumber > 0 {
		autonumber = strconv.Itoa(ctxInfo.Autonumber)
	}
	channelID := sanitizeOptional(video.ChannelID)
	channelHandle := sanitizeOptional(video.ChannelHandle)

	replacer := strings.NewReplacer(
		"{title}", title,
		"{artist}", artist,
		"{album}", album,
		"{id}", videoID,
		"{ext}", ext,
		"{quality}", quality,
		"{playlist_title}", playlistTitle,
		"{playlist-title}", playlistTitle,
		"{playlist_id}", playlistID,
		"{playlist-id}", playlistID,
		"{index}", index,
		"{count}", total,
		// extended yt-dlp style fields
		"{upload_date}", uploadDate,
		"{upload_year}", uploadYear,
		"{channel}", artist,
		"{channel_id}", channelID,
		"{channel_handle}", channelHandle,
		"{view_count}", viewCount,
		"{views}", viewCount,
		"{duration}", duration,
		"{duration_string}", durationString,
		"{resolution}", resolution,
		"{fps}", fps,
		"{vcodec}", vcodec,
		"{acodec}", acodec,
		"{bitrate}", bitrate,
		"{autonumber}", autonumber,
	)
	path := replacer.Replace(template)
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("absolute output paths are not allowed in template %q", template)
	}

	// Treat existing directory or explicit trailing slash as "put file inside".
	if strings.HasSuffix(template, "/") {
		// Interpret the template (after replacement) as a directory, and construct
		// the final output path inside that directory in a traversal-safe way.
		dir := path
		filename := fmt.Sprintf("%s.%s", title, ext)
		path = filepath.Join(dir, filename)
	} else {
		// Check if the template refers to an existing directory under baseDir.
		dirCandidate := validatedOutputDirCandidate(path, baseDir)
		if dirCandidate != "" {
			if info, err := os.Stat(dirCandidate); err == nil && info.IsDir() {
				filename := fmt.Sprintf("%s.%s", title, ext)
				path = filepath.Join(dirCandidate, filename)
			}
		}
	}

	if filepath.Ext(path) == "" {
		path = path + "." + ext
	}
	return validatedOutputPath(path, baseDir)
}

func validatedOutputPath(resolved string, baseDir string) (string, error) {
	if resolved == "" {
		return "", fmt.Errorf("output path is empty")
	}
	if filepath.IsAbs(resolved) {
		return "", fmt.Errorf("absolute output paths are not allowed")
	}
	// Check for path traversal on the uncleaned path first. This catches ".." segments
	// before filepath.Clean() collapses them with preceding directory names.
	// For example, "a/b/../c" would become "a/c" after Clean(), hiding the traversal attempt.
	if hasPathTraversal(resolved) {
		return "", fmt.Errorf("output paths cannot contain '..' segments")
	}
	cleaned := filepath.Clean(resolved)
	// Always resolve output paths relative to a base directory.
	// If no baseDir is provided, use the current working directory (".") as the base.
	if baseDir == "" {
		baseDir = "."
	}
	baseClean := filepath.Clean(baseDir)

	combined := cleaned
	if rel, err := filepath.Rel(baseClean, cleaned); err != nil || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		combined = filepath.Join(baseClean, cleaned)
	}

	// Resolve symlinks to prevent symlink-based directory traversal attacks.
	baseReal, err := filepath.EvalSymlinks(baseClean)
	if err != nil {
		// If baseDir doesn't exist or can't be resolved, use the cleaned path.
		baseReal = baseClean
	}
	combinedReal, err := filepath.EvalSymlinks(combined)
	if err != nil {
		// If combined path doesn't exist yet (which is normal for new files),
		// resolve the directory part and append the filename.
		dir := filepath.Dir(combined)
		dirReal, dirErr := filepath.EvalSymlinks(dir)
		if dirErr != nil {
			// Directory doesn't exist yet, use the original combined path.
			combinedReal = combined
		} else {
			combinedReal = filepath.Join(dirReal, filepath.Base(combined))
		}
	}

	rel, err := filepath.Rel(baseReal, combinedReal)
	if err != nil {
		return "", fmt.Errorf("resolve output path relative to %q: %w", baseReal, err)
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("output path escapes base directory %q", baseReal)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("output path escapes base directory %q", baseClean)
	}
	return combined, nil
}

func validatedOutputDirCandidate(path string, baseDir string) string {
	safe, err := validatedOutputPath(path, baseDir)
	if err != nil {
		// If the path is unsafe relative to baseDir, return a value that will cause os.Stat to fail.
		// This avoids probing arbitrary filesystem locations.
		return ""
	}
	// Reconstruct the path from the validated relative path to break the data-flow
	// from user input to the returned value (addresses CodeQL security alert).
	if baseDir == "" {
		baseDir = "."
	}
	baseClean := filepath.Clean(baseDir)
	rel, err := filepath.Rel(baseClean, safe)
	if err != nil {
		return ""
	}
	// Sanitize by reconstructing the path: join the base directory with the
	// validated relative path. This ensures the returned path is constrained
	// to the base directory.
	return filepath.Join(baseClean, rel)
}

func artifactPath(outputPath, suffix, baseDir string) (string, error) {
	if outputPath == "" {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("output path is empty"))
	}
	if strings.Contains(suffix, "/") || strings.Contains(suffix, "\\") {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("invalid artifact suffix"))
	}
	base := baseDir
	if base == "" {
		base = "."
	}
	relOutput, err := filepath.Rel(base, outputPath)
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("resolve artifact path: %w", err))
	}
	dir := filepath.Dir(relOutput)
	baseName := filepath.Base(relOutput)
	artifact := filepath.Join(dir, baseName+suffix)
	validated, err := validatedOutputPath(artifact, baseDir)
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, err)
	}
	return validated, nil
}

// hasPathTraversal checks if a path contains ".." components.
// This function is designed to work on uncleaned paths to detect path traversal
// attempts before filepath.Clean() collapses them. For example, "a/../../../etc/passwd"
// contains ".." segments that could escape the base directory, which this function
// will detect even before the path is cleaned.
func hasPathTraversal(path string) bool {
	for _, part := range strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	}) {
		if part == ".." {
			return true
		}
	}
	return false
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

func bitrateForFormat(f *youtube.Format) int {
	if f.Bitrate > 0 {
		return f.Bitrate
	}
	if f.AverageBitrate > 0 {
		return f.AverageBitrate
	}
	return 0
}

// formatDurationString renders a duration as H:MM:SS or M:SS.
func formatDurationString(d time.Duration) string {
	total := int64(d.Seconds())
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// mimeCodecs extracts the "codecs=" parameter from a mime type string such
// as "video/mp4; codecs=avc1.640028, opus".
func mimeCodecs(mime string) []string {
	idx := strings.Index(strings.ToLower(mime), "codecs=")
	if idx < 0 {
		return nil
	}
	rest := mime[idx+len("codecs="):]
	if strings.HasPrefix(rest, "\"") {
		end := strings.Index(rest[1:], "\"")
		if end < 0 {
			return nil
		}
		rest = rest[1 : end+1]
	} else {
		if end := strings.IndexAny(rest, "; \""); end >= 0 {
			rest = rest[:end]
		}
	}
	var out []string
	for _, c := range strings.Split(rest, ",") {
		if c = strings.TrimSpace(c); c != "" {
			out = append(out, c)
		}
	}
	return out
}

// formatVideoCodec returns the video codec from a format's mime type.
func formatVideoCodec(f *youtube.Format) string {
	if f == nil || strings.HasPrefix(strings.ToLower(f.MimeType), "audio/") {
		return ""
	}
	for _, c := range mimeCodecs(f.MimeType) {
		lower := strings.ToLower(c)
		if strings.HasPrefix(lower, "avc") || strings.HasPrefix(lower, "vp") ||
			strings.HasPrefix(lower, "h264") || strings.HasPrefix(lower, "hevc") ||
			strings.HasPrefix(lower, "av01") || strings.HasPrefix(lower, "h26") {
			return strings.SplitN(c, ".", 2)[0]
		}
	}
	return ""
}

// formatAudioCodec returns the audio codec from a format's mime type.
func formatAudioCodec(f *youtube.Format) string {
	if f == nil {
		return ""
	}
	for _, c := range mimeCodecs(f.MimeType) {
		lower := strings.ToLower(c)
		if strings.HasPrefix(lower, "mp4a") || strings.HasPrefix(lower, "opus") ||
			strings.HasPrefix(lower, "vorbis") || strings.HasPrefix(lower, "ec-3") ||
			strings.HasPrefix(lower, "ac-3") || strings.HasPrefix(lower, "flac") {
			return strings.SplitN(c, ".", 2)[0]
		}
	}
	return ""
}
