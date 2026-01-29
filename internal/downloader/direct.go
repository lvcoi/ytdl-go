package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

type directInfo struct {
	URL         string
	Kind        string // "file", "hls", "dash"
	ContentType string
	Size        int64
	Title       string
	Author      string
	ID          string
	Ext         string
}

func processDirect(ctx context.Context, rawURL string, opts Options, printer *Printer) (downloadResult, error) {
	info, err := probeDirectURL(ctx, rawURL, opts.Timeout)
	if err != nil {
		return downloadResult{}, err
	}

	video := &youtube.Video{
		ID:     info.ID,
		Title:  info.Title,
		Author: info.Author,
	}

	switch info.Kind {
	case "hls":
		video.HLSManifestURL = info.URL
		ctxInfo := outputContext{SourceURL: info.URL, MetaOverrides: opts.MetaOverrides}
		return downloadHLS(ctx, newClient(opts), video, opts, ctxInfo, printer, printer.Prefix(1, 1, info.Title))
	case "dash":
		video.DASHManifestURL = info.URL
		ctxInfo := outputContext{SourceURL: info.URL, MetaOverrides: opts.MetaOverrides}
		return downloadDASH(ctx, newClient(opts), video, opts, ctxInfo, printer, printer.Prefix(1, 1, info.Title))
	default:
		return downloadDirectFile(ctx, info, opts, printer)
	}
}

func probeDirectURL(ctx context.Context, rawURL string, timeout time.Duration) (directInfo, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return directInfo{}, wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid URL: %w", err))
	}
	ext := strings.ToLower(filepath.Ext(parsed.Path))
	ext = strings.TrimPrefix(ext, ".")
	info := directInfo{
		URL:  rawURL,
		Kind: kindFromExt(ext),
		Ext:  ext,
	}
	info.Title, info.ID = titleFromURL(parsed)

	if info.Kind == "" {
		// Probe content type to decide.
		resp, err := headOrGet(ctx, rawURL, timeout)
		if err != nil {
			return directInfo{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return directInfo{}, wrapCategory(CategoryRestricted, fmt.Errorf("restricted content (status %d)", resp.StatusCode))
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return directInfo{}, wrapCategory(CategoryNetwork, fmt.Errorf("unexpected status %d", resp.StatusCode))
		}
		contentType := resp.Header.Get("Content-Type")
		info.ContentType = contentType
		if resp.ContentLength > 0 {
			info.Size = resp.ContentLength
		}
		if guessKindFromContentType(contentType) != "" {
			info.Kind = guessKindFromContentType(contentType)
		}
		if info.Ext == "" {
			if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
				info.Ext = strings.TrimPrefix(exts[0], ".")
			}
		}
	}

	if info.ContentType == "" || strings.Contains(strings.ToLower(info.ContentType), "text/html") {
		if title, author, err := fetchPageMetadata(ctx, rawURL, timeout); err == nil {
			if title != "" {
				info.Title = sanitize(title)
				info.ID = sanitize(title)
			}
			if author != "" {
				info.Author = author
			}
		}
	}

	if info.Kind == "" {
		return directInfo{}, wrapCategory(CategoryUnsupported, fmt.Errorf("unsupported URL or content type (use --list-formats if applicable)"))
	}
	if info.Ext == "" && info.Kind == "file" {
		info.Ext = "bin"
	}
	return info, nil
}

func headOrGet(ctx context.Context, rawURL string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode != http.StatusMethodNotAllowed {
		return resp, nil
	}
	if resp != nil {
		resp.Body.Close()
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func kindFromExt(ext string) string {
	switch ext {
	case "m3u8":
		return "hls"
	case "mpd":
		return "dash"
	case "mp4", "webm", "mov", "m4v":
		return "file"
	default:
		return ""
	}
}

func guessKindFromContentType(contentType string) string {
	ctype := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ctype {
	case "application/vnd.apple.mpegurl", "application/x-mpegurl":
		return "hls"
	case "application/dash+xml":
		return "dash"
	}
	if strings.HasPrefix(ctype, "video/") || strings.HasPrefix(ctype, "audio/") {
		return "file"
	}
	return ""
}

func titleFromURL(parsed *url.URL) (string, string) {
	base := path.Base(parsed.Path)
	base = strings.TrimSpace(base)
	if base == "" || base == "/" || base == "." {
		base = parsed.Host
	}
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		base = "download"
	}
	safe := sanitize(base)
	return safe, safe
}

func downloadDirectFile(ctx context.Context, info directInfo, opts Options, printer *Printer) (downloadResult, error) {
	format := &youtube.Format{
		MimeType: fmt.Sprintf("video/%s", info.Ext),
	}
	video := &youtube.Video{
		ID:     info.ID,
		Title:  info.Title,
		Author: info.Author,
	}
	outputPath, err := resolveOutputPath(opts.OutputTemplate, video, format, outputContext{SourceURL: info.URL, MetaOverrides: opts.MetaOverrides})
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, err)
	}
	outputPath, skip, err := handleExistingPath(outputPath, opts, printer)
	if err != nil {
		return downloadResult{}, err
	}
	if skip {
		return downloadResult{skipped: true, outputPath: outputPath}, nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("creating output directory: %w", err))
	}
	partPath := outputPath + partSuffix
	resumePath := outputPath + resumeSuffix

	state := fileResumeState{URL: info.URL, BytesWritten: 0}
	if loaded, err := loadFileResume(resumePath); err == nil && loaded.URL == info.URL {
		state = loaded
	}

	file, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("opening temp file: %w", err))
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.URL, nil)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryNetwork, err)
	}
	if state.BytesWritten > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", state.BytesWritten))
	}
	resp, err := doWithRetry(req, opts.Timeout, maxSegmentRetries)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryNetwork, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return downloadResult{}, wrapCategory(CategoryRestricted, fmt.Errorf("restricted content (status %d)", resp.StatusCode))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("unexpected status %d", resp.StatusCode))
	}

	var writer io.Writer = file
	var progress *progressWriter
	if !opts.Quiet {
		progress = newProgressWriter(resp.ContentLength, printer, printer.Prefix(1, 1, info.Title))
		progress.SetCurrent(state.BytesWritten)
		writer = io.MultiWriter(file, progress)
	}

	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("download failed: %w", err))
	}
	state.BytesWritten += written
	if progress != nil {
		progress.Finish()
	}
	if err := saveFileResume(resumePath, state); err != nil {
		return downloadResult{}, err
	}
	if err := file.Close(); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("closing temp file: %w", err))
	}
	if err := os.Rename(partPath, outputPath); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("renaming output: %w", err))
	}
	if err := validateOutputFile(outputPath, format); err != nil {
		return downloadResult{}, err
	}
	_ = os.Remove(resumePath)
	metadata := buildItemMetadata(video, format, outputContext{SourceURL: info.URL, MetaOverrides: opts.MetaOverrides}, outputPath, "ok", nil)
	if err := writeSidecar(outputPath, metadata); err != nil {
		return downloadResult{}, err
	}
	return downloadResult{bytes: state.BytesWritten, outputPath: outputPath}, nil
}

type fileResumeState struct {
	URL          string `json:"url"`
	BytesWritten int64  `json:"bytes_written"`
}

func loadFileResume(path string) (fileResumeState, error) {
	file, err := os.Open(path)
	if err != nil {
		return fileResumeState{}, err
	}
	defer file.Close()
	var state fileResumeState
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return fileResumeState{}, err
	}
	return state, nil
}

func saveFileResume(path string, state fileResumeState) error {
	file, err := os.Create(path)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("saving resume state: %w", err))
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(state); err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("saving resume state: %w", err))
	}
	return nil
}

func doWithRetry(req *http.Request, timeout time.Duration, maxAttempts int) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	}
	return nil, lastErr
}
