package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

const (
	maxSegmentRetries = 3
	resumeSuffix      = ".resume.json"
	partSuffix        = ".part"
)

func sanitizeOutputPath(path string) (string, string, error) {
	if path == "" {
		return "", "", fmt.Errorf("empty output path")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolving output path: %w", err)
	}
	baseDir := filepath.Dir(absPath)
	if baseDir == "" || baseDir == string(filepath.Separator) {
		return "", "", fmt.Errorf("invalid base directory for output path: %q", baseDir)
	}
	return absPath, baseDir, nil
}

func downloadAdaptive(ctx context.Context, client *youtube.Client, video *youtube.Video, opts Options, ctxInfo outputContext, printer *Printer, prefix string, formatErr error) (downloadResult, error) {
	if opts.AudioOnly {
		return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("audio-only adaptive downloads are not supported yet (use --list-formats): %w", formatErr))
	}

	if video.HLSManifestURL != "" {
		return downloadHLS(ctx, client, video, opts, ctxInfo, printer, prefix)
	}
	if video.DASHManifestURL != "" {
		return downloadDASH(ctx, client, video, opts, ctxInfo, printer, prefix)
	}
	return downloadResult{}, formatErr
}

func downloadHLS(ctx context.Context, client *youtube.Client, video *youtube.Video, opts Options, ctxInfo outputContext, printer *Printer, prefix string) (downloadResult, error) {
	manifestURL := video.HLSManifestURL
	data, err := fetchManifest(ctx, client, manifestURL)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("fetching HLS manifest: %w", err))
	}

	manifest, err := ParseHLSManifest(data)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("parsing HLS manifest: %w", err))
	}

	if ok, method := DetectHLSDrm(manifest); ok {
		message := "encrypted HLS manifest"
		if method != "" {
			message = fmt.Sprintf("encrypted HLS manifest (%s)", method)
		}
		return downloadResult{}, wrapCategory(CategoryRestricted, fmt.Errorf("%s", message))
	}

	playlistURL := manifestURL
	selectedVariant := HLSVariant{}
	if len(manifest.Variants) > 0 {
		selected, err := selectHLSVariant(manifest.Variants, opts.Quality)
		if err != nil {
			return downloadResult{}, wrapCategory(CategoryUnsupported, err)
		}
		selectedVariant = selected
		playlistURL = resolveManifestURL(manifestURL, selected.URI)
		data, err = fetchManifest(ctx, client, playlistURL)
		if err != nil {
			return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("fetching HLS variant: %w", err))
		}
		manifest, err = ParseHLSManifest(data)
		if err != nil {
			return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("parsing HLS variant: %w", err))
		}
		if ok, method := DetectHLSDrm(manifest); ok {
			message := "encrypted HLS manifest"
			if method != "" {
				message = fmt.Sprintf("encrypted HLS manifest (%s)", method)
			}
			return downloadResult{}, wrapCategory(CategoryRestricted, fmt.Errorf("%s", message))
		}
	}

	if len(manifest.Segments) == 0 {
		return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("no HLS segments found"))
	}

	format := hlsFormatFromSegments(manifest.Segments, opts.Quality, selectedVariant)
	outputPath, err := resolveOutputPath(opts.OutputTemplate, video, format, ctxInfo, opts.OutputDir)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, err)
	}
	// Normalize and derive a safe base directory for subsequent temp files.
	outputPath, baseDir, err := sanitizeOutputPath(outputPath)
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

	result, err := downloadHLSSegments(ctx, client, playlistURL, manifest.Segments, outputPath, baseDir, opts, printer, prefix)
	if err == nil {
		result.outputPath = outputPath
	}
	return result, err
}

func selectHLSVariant(variants []HLSVariant, quality string) (HLSVariant, error) {
	targetHeight, preferLowest, err := parseVideoQuality(quality)
	if err != nil {
		return HLSVariant{}, err
	}

	best := HLSVariant{}
	hasBest := false
	for _, variant := range variants {
		height := parseResolutionHeight(variant.Resolution)
		if targetHeight > 0 && height > targetHeight {
			continue
		}
		if !hasBest || hlsVariantBetter(variant, best, height) {
			best = variant
			hasBest = true
		}
	}

	if !hasBest && targetHeight > 0 {
		for _, variant := range variants {
			height := parseResolutionHeight(variant.Resolution)
			if !hasBest || height < parseResolutionHeight(best.Resolution) || (height == parseResolutionHeight(best.Resolution) && variant.Bandwidth > best.Bandwidth) {
				best = variant
				hasBest = true
			}
		}
	}

	if !hasBest && preferLowest {
		for _, variant := range variants {
			height := parseResolutionHeight(variant.Resolution)
			if !hasBest || height < parseResolutionHeight(best.Resolution) || (height == parseResolutionHeight(best.Resolution) && variant.Bandwidth > best.Bandwidth) {
				best = variant
				hasBest = true
			}
		}
	}

	if !hasBest {
		for _, variant := range variants {
			if !hasBest || hlsVariantBetter(variant, best, parseResolutionHeight(variant.Resolution)) {
				best = variant
				hasBest = true
			}
		}
	}

	if !hasBest {
		return HLSVariant{}, fmt.Errorf("no HLS variants available")
	}
	return best, nil
}

func hlsVariantBetter(candidate, current HLSVariant, candidateHeight int) bool {
	currentHeight := parseResolutionHeight(current.Resolution)
	if candidateHeight != currentHeight {
		return candidateHeight > currentHeight
	}
	return candidate.Bandwidth > current.Bandwidth
}

func parseResolutionHeight(res string) int {
	if res == "" {
		return 0
	}
	parts := strings.Split(res, "x")
	if len(parts) != 2 {
		return 0
	}
	height, _ := strconv.Atoi(parts[1])
	return height
}

func hlsFormatFromSegments(segments []HLSSegment, quality string, variant HLSVariant) *youtube.Format {
	ext := "ts"
	if len(segments) > 0 {
		ext = strings.TrimPrefix(filepath.Ext(segments[0].URI), ".")
		if ext == "" {
			ext = "ts"
		}
	}
	mime := "video/" + ext
	if quality == "" && variant.Resolution != "" {
		quality = variant.Resolution
	}
	format := &youtube.Format{
		MimeType:     mime,
		QualityLabel: quality,
	}
	return format
}

func fetchManifest(ctx context.Context, client *youtube.Client, manifestURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func resolveManifestURL(baseURL, ref string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ref
	}
	rel, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(rel).String()
}

type hlsResumeState struct {
	ManifestURL  string `json:"manifest_url"`
	SegmentCount int    `json:"segment_count"`
	NextIndex    int    `json:"next_index"`
	BytesWritten int64  `json:"bytes_written"`
}

func downloadHLSSegments(ctx context.Context, client *youtube.Client, playlistURL string, segments []HLSSegment, outputPath string, baseDir string, opts Options, printer *Printer, prefix string) (downloadResult, error) {
	partPath, err := artifactPath(outputPath, partSuffix)
	if err != nil {
		return downloadResult{}, err
	}
	resumePath, err := artifactPath(outputPath, resumeSuffix)
	if err != nil {
		return downloadResult{}, err
	}
	segmentDir, err := artifactPath(outputPath, ".segments")
	if err != nil {
		return downloadResult{}, err
	}

	state := hlsResumeState{
		ManifestURL:  playlistURL,
		SegmentCount: len(segments),
		NextIndex:    0,
		BytesWritten: 0,
	}
	if loaded, err := loadHLSResume(resumePath); err == nil && loaded.ManifestURL == playlistURL && loaded.SegmentCount == len(segments) {
		state = loaded
	}

	useParallel := opts.SegmentConcurrency != 1 && state.NextIndex == 0 && state.BytesWritten == 0
	if useParallel {
		tempDir, err := validateSegmentTempDir(segmentDir)
		if err != nil {
			return downloadResult{}, err
		}
		file, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("opening temp file: %w", err))
		}
		defer file.Close()

		urls := make([]string, len(segments))
		for i, seg := range segments {
			urls[i] = resolveManifestURL(playlistURL, seg.URI)
		}
		plan := segmentDownloadPlan{
			URLs:        urls,
			TempDir:     tempDir,
			Prefix:      prefix,
			Concurrency: opts.SegmentConcurrency,
			BaseDir:     baseDir,
		}
		total, err := downloadSegmentsParallel(ctx, client, plan, file, printer)
		if err != nil {
			return downloadResult{}, err
		}
		if err := file.Close(); err != nil {
			return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("closing temp file: %w", err))
		}
		if err := os.Rename(partPath, outputPath); err != nil {
			return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("renaming output: %w", err))
		}
		if err := validateOutputFile(outputPath, nil); err != nil {
			return downloadResult{}, err
		}
		_ = os.RemoveAll(segmentDir)
		return downloadResult{bytes: total, outputPath: outputPath}, nil
	}

	file, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("opening temp file: %w", err))
	}
	defer file.Close()

	var writer io.Writer = file
	var progress *progressWriter
	if !opts.Quiet {
		progress = newProgressWriter(0, printer, prefix)
		progress.total.Store(state.BytesWritten)
		writer = io.MultiWriter(file, progress)
	}

	for idx := state.NextIndex; idx < len(segments); idx++ {
		segmentURL := resolveManifestURL(playlistURL, segments[idx].URI)
		if err := downloadSegmentWithRetry(ctx, client, segmentURL, writer); err != nil {
			if progress != nil {
				progress.NewLine()
			}
			return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("segment %d failed: %w", idx+1, err))
		}

		state.NextIndex = idx + 1
		if progress != nil {
			state.BytesWritten = progress.total.Load()
		}
		if err := saveHLSResume(resumePath, state); err != nil {
			return downloadResult{}, err
		}
	}

	if progress != nil {
		progress.Finish()
	}

	if err := file.Close(); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("closing temp file: %w", err))
	}
	if err := os.Rename(partPath, outputPath); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("renaming output: %w", err))
	}
	if err := validateOutputFile(outputPath, nil); err != nil {
		return downloadResult{}, err
	}
	_ = os.Remove(resumePath)

	return downloadResult{bytes: state.BytesWritten, outputPath: outputPath}, nil
}

func downloadSegmentWithRetry(ctx context.Context, client *youtube.Client, segmentURL string, writer io.Writer) error {
	var lastErr error
	for attempt := 1; attempt <= maxSegmentRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, segmentURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.HTTPClient.Do(req)
		if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			_, copyErr := io.Copy(writer, resp.Body)
			resp.Body.Close()
			if copyErr == nil {
				return nil
			}
			lastErr = copyErr
		} else {
			if resp != nil {
				resp.Body.Close()
				lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
			} else {
				lastErr = err
			}
		}
		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	}
	return lastErr
}

func loadHLSResume(path string) (hlsResumeState, error) {
	file, err := os.Open(path)
	if err != nil {
		return hlsResumeState{}, err
	}
	defer file.Close()
	var state hlsResumeState
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return hlsResumeState{}, err
	}
	return state, nil
}

func saveHLSResume(path string, state hlsResumeState) error {
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
