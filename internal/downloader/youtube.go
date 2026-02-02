package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const (
	minChunkSize     int64 = 256 * 1024      // 256KB keeps progress responsive on small files
	maxChunkSize     int64 = 2 * 1024 * 1024 // cap to avoid excessive requests on large files
	targetChunkCount int64 = 64
)

// adjustChunkSize picks a smaller chunk size for the YouTube client to keep
// progress updates frequent without spawning thousands of requests.
func adjustChunkSize(client *youtube.Client, contentLength int64) {
	if client == nil || contentLength <= 0 {
		return
	}
	chunk := contentLength / targetChunkCount
	if chunk < minChunkSize {
		chunk = minChunkSize
	} else if chunk > maxChunkSize {
		chunk = maxChunkSize
	}
	client.ChunkSize = chunk
}

// ffmpegAvailable checks if ffmpeg is installed and accessible
func ffmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// downloadWithFFmpegFallback downloads a progressive format and extracts audio using ffmpeg
func downloadWithFFmpegFallback(ctx context.Context, client *youtube.Client, video *youtube.Video, opts Options, printer *Printer, prefix string, audioOutputPath string, progress *progressWriter) (downloadResult, error) {
	result := downloadResult{}

	// Find the best progressive format with high-quality audio
	var progressiveFormat *youtube.Format
	preferredItags := []int{22, 18} // 720p first, then 360p
	for _, targetItag := range preferredItags {
		for i := range video.Formats {
			f := &video.Formats[i]
			if f.ItagNo == targetItag && f.AudioChannels > 0 {
				progressiveFormat = f
				break
			}
		}
		if progressiveFormat != nil {
			break
		}
	}
	if progressiveFormat == nil {
		return result, wrapCategory(CategoryUnsupported, errors.New("no progressive format available for ffmpeg fallback"))
	}

	// Create temp file for video download
	tempVideoPath := audioOutputPath + ".tmp.mp4"
	defer os.Remove(tempVideoPath)

	if progressiveFormat.ItagNo == 22 {
		printer.Log(LogInfo, "step 1/3: downloading 720p video (itag 22, 192kbps AAC source)")
	} else {
		printer.Log(LogInfo, "step 1/3: downloading 360p video (itag 18, 96kbps AAC source)")
		printer.Log(LogWarn, "note: 720p source unavailable, using 360p video for audio extraction")
	}

	adjustChunkSize(client, progressiveFormat.ContentLength)
	stream, size, err := client.GetStreamContext(ctx, video, progressiveFormat)
	if err != nil {
		return result, wrapCategory(CategoryNetwork, fmt.Errorf("downloading progressive format: %w", err))
	}
	if size <= 0 && progressiveFormat.ContentLength > 0 {
		size = progressiveFormat.ContentLength
	}
	defer stream.Close()

	tempFile, err := os.Create(tempVideoPath)
	if err != nil {
		return result, wrapCategory(CategoryFilesystem, err)
	}

	var writer io.Writer = tempFile
	if !opts.Quiet {
		if progress != nil {
			progress.Reset(size)
		} else {
			progress = newProgressWriter(size, printer, prefix)
		}
		writer = io.MultiWriter(tempFile, progress)
	}

	_, err = copyWithContext(ctx, writer, stream)
	tempFile.Close()
	if err != nil {
		return result, wrapCategory(CategoryNetwork, fmt.Errorf("downloading progressive format: %w", err))
	}
	if progress != nil {
		progress.Finish()
	}

	// Extract audio using ffmpeg
	printer.Log(LogInfo, "step 2/3: extracting audio track")
	printer.Log(LogInfo, "step 3/3: encoding to Opus @ 160kbps")
	if err := extractAudio(tempVideoPath, audioOutputPath); err != nil {
		return result, wrapCategory(CategoryFilesystem, fmt.Errorf("ffmpeg extraction failed: %w", err))
	}

	// Get file size
	if fi, err := os.Stat(audioOutputPath); err == nil {
		result.bytes = fi.Size()
	}
	result.outputPath = audioOutputPath
	printer.Log(LogInfo, fmt.Sprintf("complete: %s (Opus audio)", filepath.Base(audioOutputPath)))
	return result, nil
}

// extractAudio extracts audio from a video file using ffmpeg
func extractAudio(inputPath, outputPath string) error {
	// Determine output format from extension
	ext := strings.ToLower(filepath.Ext(outputPath))
	kwargs := ffmpeg.KwArgs{"vn": ""}

	switch ext {
	case ".mp3":
		kwargs["acodec"] = "libmp3lame"
		kwargs["q:a"] = "2"
	case ".m4a", ".aac":
		kwargs["acodec"] = "aac"
		kwargs["b:a"] = "192k"
	case ".opus":
		kwargs["acodec"] = "libopus"
		kwargs["b:a"] = "160k" // Match itag 251 quality
	case ".webm":
		kwargs["acodec"] = "libopus"
		kwargs["b:a"] = "160k" // Match itag 251 quality
	default:
		// Copy audio codec if possible
		kwargs["acodec"] = "copy"
	}

	return ffmpeg.Input(inputPath).
		Output(outputPath, kwargs).
		OverWriteOutput().
		Silent(true).
		Run()
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
	outputPath, skip, err := handleExistingPath(outputPath, opts, printer)
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

	adjustChunkSize(client, format.ContentLength)
	stream, size, err := client.GetStreamContext(ctx, video, format)
	if err != nil {
		return result, wrapCategory(CategoryNetwork, fmt.Errorf("starting stream: %w", err))
	}
	if size <= 0 && format.ContentLength > 0 {
		size = format.ContentLength
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
			stream = nil
			stream, size, err = client.GetStreamContext(ctx, video, &formatSingle)
			if err != nil {
				return result, wrapCategory(CategoryNetwork, fmt.Errorf("retry failed: %w", err))
			}
			if size <= 0 && format.ContentLength > 0 {
				size = format.ContentLength
			}

			writer = file
			if !opts.Quiet {
				if progress != nil {
					progress.Reset(size)
				} else {
					progress = newProgressWriter(size, printer, prefix)
				}
				writer = io.MultiWriter(file, progress)
			} else {
				progress = nil
			}
			result.hadProgress = progress != nil

			written, err = copyWithContext(ctx, writer, stream)
			result.retried = true
		}
		if err != nil {
			// If audio-only format fails with 403, try ffmpeg fallback
			isAudioOnlyFormat := format.AudioChannels > 0 && format.Width == 0 && format.Height == 0
			audioOnlyItags := map[int]bool{251: true, 140: true, 250: true, 249: true, 139: true, 171: true}
			isAudioOnlyItag := audioOnlyItags[opts.Itag]
			if (opts.AudioOnly || isAudioOnlyFormat || isAudioOnlyItag) && isUnexpectedStatus(err, http.StatusForbidden) && ffmpegAvailable() {
				printer.Log(LogWarn, "YouTube blocked chunked audio-only download (403); switching to ffmpeg fallback to preserve Opus quality")
				printer.Log(LogInfo, "ffmpeg fallback: download video → extract audio → encode Opus @ 160kbps")
				file.Close()
				os.Remove(outputPath)
				return downloadWithFFmpegFallback(ctx, client, video, opts, printer, prefix, outputPath, progress)
			}
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

func isUnexpectedStatus(err error, code int) bool {
	var statusErr youtube.ErrUnexpectedStatusCode
	if errors.As(err, &statusErr) {
		return int(statusErr) == code
	}
	return false
}
