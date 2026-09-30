package downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lvcoi/ytdl-lib/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var audioFormats = map[string]bool{
	"mp3": true, "opus": true, "m4a": true, "aac": true,
	"flac": true, "wav": true, "best": true,
}

// normalizeAudioFormat validates --audio-format input.
func normalizeAudioFormat(format string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" || format == "best" {
		return "", nil
	}
	if !audioFormats[format] {
		return "", wrapCategory(CategoryUnsupported, fmt.Errorf("unsupported audio format %q (supported: mp3, opus, m4a, aac, flac, wav, best)", format))
	}
	if format == "aac" {
		format = "m4a" // aac ADTS vs m4a: use m4a container
	}
	return format, nil
}

// convertAudioFile converts an audio file to the target format using ffmpeg,
// optionally applying --audio-quality ("160k" CBR or "0"-"9" VBR).
func convertAudioFile(outputPath, targetFormat, quality string) error {
	targetExt := "." + targetFormat
	if strings.EqualFold(filepath.Ext(outputPath), targetExt) {
		return nil // already in the target format
	}
	tmpPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".tmp_convert" + targetExt

	kwargs := ffmpegAudioArgs(targetFormat, quality)
	if err := ffmpeg.Input(outputPath).Output(tmpPath, kwargs).OverWriteOutput().Silent(true).Run(); err != nil {
		os.Remove(tmpPath)
		return wrapCategory(CategoryFilesystem, fmt.Errorf("converting audio to %s: %w", targetFormat, err))
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		os.Remove(tmpPath)
		return wrapCategory(CategoryFilesystem, fmt.Errorf("replacing audio file: %w", err))
	}
	return nil
}

// ffmpegAudioArgs maps format + quality to ffmpeg kwargs (mirrors extractAudio).
func ffmpegAudioArgs(format, quality string) ffmpeg.KwArgs {
	kwargs := ffmpeg.KwArgs{"vn": ""}
	switch format {
	case "mp3":
		kwargs["acodec"] = "libmp3lame"
		applyAudioQuality(kwargs, quality, "q:a", "2", "b:a", "192k")
	case "m4a":
		kwargs["acodec"] = "aac"
		applyAudioQuality(kwargs, quality, "", "", "b:a", "192k")
	case "opus", "webm":
		kwargs["acodec"] = "libopus"
		applyAudioQuality(kwargs, quality, "", "", "b:a", "160k")
	case "flac":
		kwargs["acodec"] = "flac"
	case "wav":
		kwargs["acodec"] = "pcm_s16le"
	default:
		kwargs["acodec"] = "copy"
	}
	return kwargs
}

func applyAudioQuality(kwargs ffmpeg.KwArgs, quality, vbrKey, vbrDefault, cbrKey, cbrDefault string) {
	quality = strings.TrimSpace(quality)
	if quality == "" {
		if vbrKey != "" {
			kwargs[vbrKey] = vbrDefault
		} else if cbrKey != "" {
			kwargs[cbrKey] = cbrDefault
		}
		return
	}
	// "0"-"9" means VBR quality scale; anything else is a bitrate like "160k".
	if len(quality) == 1 && quality[0] >= '0' && quality[0] <= '9' {
		if vbrKey != "" {
			kwargs[vbrKey] = quality
		} else {
			// Map VBR scale to a rough bitrate for CBR-only codecs.
			scales := map[string]string{"0": "320k", "1": "256k", "2": "224k", "3": "192k", "4": "160k", "5": "128k", "6": "112k", "7": "96k", "8": "80k", "9": "64k"}
			kwargs[cbrKey] = scales[quality]
		}
		return
	}
	kwargs[cbrKey] = quality
}

// isLiveVideo heuristically detects a currently-live YouTube stream: no fixed
// duration and an HLS manifest present.
func isLiveVideo(video *youtube.Video) bool {
	return video != nil && video.Duration <= 0 && video.HLSManifestURL != ""
}

// postProcessVideo runs every configured post-download step in order:
// audio conversion, SponsorBlock removal/marking, subtitle embedding,
// thumbnail writing/embedding, and archive recording.
func postProcessVideo(ctx context.Context, client YouTubeClient, video *youtube.Video, outputPath string, opts Options, printer *Printer) error {
	if outputPath == "" {
		return nil
	}

	// 1. Audio conversion (--audio-format / --audio-quality)
	if opts.AudioOnly {
		if target, err := normalizeAudioFormat(opts.AudioFormat); err != nil {
			return err
		} else if target != "" {
			if err := convertAudioFile(outputPath, target, opts.AudioQuality); err != nil {
				return err
			}
			printer.Log(LogInfo, fmt.Sprintf("audio converted to %s", target))
		}
	}

	// 2. SponsorBlock (remove segments, add chapter markers)
	if opts.SponsorblockRemove != "" || opts.SponsorblockMark != "" {
		if err := applySponsorblock(ctx, video, outputPath, opts.OutputDir, opts, printer); err != nil {
			return err
		}
	}

	// 3. Subtitles: write sidecar .vtt files and/or embed into the container
	if opts.WriteSubs || opts.WriteAutoSubs || opts.EmbedSubs {
		var written []string
		var err error
		written, err = writeSubtitleTracks(ctx, video, outputPath, opts.OutputDir, opts, client.HTTP(), printer)
		if err != nil {
			return err
		}
		if opts.EmbedSubs && len(written) > 0 {
			// Embed the first (best) track; the rest remain as sidecar files.
			if embedErr := embedSubtitleTrack(outputPath, written[0]); embedErr != nil {
				printer.Log(LogWarn, fmt.Sprintf("warning: subtitle embedding failed: %v", embedErr))
			} else {
				printer.Log(LogInfo, fmt.Sprintf("subtitles embedded: %s", filepath.Base(written[0])))
			}
		}
	}

	// 4. Thumbnail: write sidecar image and/or embed as cover art
	if opts.WriteThumbnail || opts.EmbedThumbnail {
		thumbPath, err := writeThumbnailFile(ctx, video, outputPath, opts.OutputDir, client.HTTP(), printer)
		if err != nil {
			if opts.EmbedThumbnail {
				return err
			}
			printer.Log(LogWarn, fmt.Sprintf("warning: thumbnail download failed: %v", err))
		} else {
			if opts.WriteThumbnail {
				printer.Log(LogInfo, fmt.Sprintf("thumbnail: %s", filepath.Base(thumbPath)))
			}
			if opts.EmbedThumbnail {
				if strings.EqualFold(filepath.Ext(outputPath), ".webm") {
					printer.Log(LogWarn, "warning: the WebM muxer does not support cover art; skipping thumbnail embed")
				} else if embedErr := embedThumbnailFile(outputPath, thumbPath, printer); embedErr != nil {
					printer.Log(LogWarn, fmt.Sprintf("warning: thumbnail embedding failed: %v", embedErr))
				} else {
					printer.Log(LogInfo, "thumbnail embedded as cover art")
				}
			}
			if !opts.WriteThumbnail {
				os.Remove(thumbPath)
			}
		}
	}

	return nil
}

// ffmpegCommand builds an *exec.Cmd for ffmpeg (helper shared by post processors).
func ffmpegCommand(args ...string) *exec.Cmd {
	return exec.Command("ffmpeg", args...)
}
