package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lvcoi/ytdl-lib/v2"
)

// downloadLiveHLS records an ongoing YouTube live stream by polling the HLS
// playlist and appending newly published segments. Recording stops when the
// stream ends (playlist gains #EXT-X-ENDLIST) or the context is cancelled.
func downloadLiveHLS(ctx context.Context, client YouTubeClient, video *youtube.Video, opts Options, ctxInfo outputContext, printer *Printer, prefix string) (downloadResult, error) {
	manifestURL := video.HLSManifestURL
	if manifestURL == "" {
		return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("no HLS manifest available for live stream"))
	}

	printer.Log(LogInfo, "live stream: fetching master playlist")
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
	}

	format := &youtube.Format{MimeType: "video/mp2t", QualityLabel: selectedVariant.Resolution}
	outputPath, err := resolveOutputPath(opts.OutputTemplate, video, format, ctxInfo, opts.OutputDir)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("creating output directory: %w", err))
	}

	partPath, err := artifactPath(outputPath, partSuffix, opts.OutputDir)
	if err != nil {
		return downloadResult{}, err
	}
	file, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("opening temp file: %w", err))
	}
	defer file.Close()

	var writer io.Writer = file
	if !opts.Quiet {
		if limited, err := applyRateLimit(ctx, writer, opts); err != nil {
			return downloadResult{}, err
		} else if limited != nil {
			writer = limited
		}
	}

	seen := map[string]struct{}{}
	pollInterval := opts.LiveSegmentInterval
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	var totalBytes int64
	var segmentCount int
	printer.Log(LogInfo, "live stream: recording (Ctrl+C to stop; finalizes automatically when the stream ends)")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	refresh := func() (bool, error) {
		data, err := fetchManifest(ctx, client, playlistURL)
		if err != nil {
			return false, err
		}
		media, err := ParseHLSManifest(data)
		if err != nil {
			return false, err
		}
		for _, seg := range media.Segments {
			uri := resolveManifestURL(playlistURL, seg.URI)
			if _, ok := seen[uri]; ok {
				continue
			}
			seen[uri] = struct{}{}
			if err := downloadSegmentWithRetry(ctx, client, uri, writer); err != nil {
				return media.HasEndlist, wrapCategory(CategoryNetwork, fmt.Errorf("live segment %d failed: %w", segmentCount+1, err))
			}
			segmentCount++
		}
		return media.HasEndlist, nil
	}

	ended := false
	for !ended {
		if end, err := refresh(); err != nil {
			if ctx.Err() != nil {
				break
			}
			printer.Log(LogWarn, fmt.Sprintf("live: manifest refresh failed: %v (retrying)", err))
		} else {
			ended = end
		}
		if ended || ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}

	if err := file.Close(); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("closing temp file: %w", err))
	}
	if fi, statErr := os.Stat(partPath); statErr == nil {
		totalBytes = fi.Size()
	}
	if segmentCount == 0 {
		os.Remove(partPath)
		return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("no live segments could be recorded"))
	}
	if err := os.Rename(partPath, outputPath); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("renaming output: %w", err))
	}
	if !ended {
		printer.Log(LogWarn, "live: recording stopped before the stream ended (output may be truncated)")
	}
	printer.Log(LogInfo, fmt.Sprintf("live: recorded %d segments (%s)", segmentCount, humanBytes(totalBytes)))
	return downloadResult{bytes: totalBytes, outputPath: outputPath, format: format}, nil
}
