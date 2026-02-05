package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kkdai/youtube/v2"
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
