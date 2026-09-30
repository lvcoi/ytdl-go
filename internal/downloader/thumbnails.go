package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/lvcoi/ytdl-lib/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// writeThumbnailFile downloads the best thumbnail next to the output file
// and returns the artifact path.
func writeThumbnailFile(ctx context.Context, video *youtube.Video, outputPath, baseDir string, client HTTPDoer, printer *Printer) (string, error) {
	thumbURL := bestThumbnailURL(video.Thumbnails)
	if thumbURL == "" {
		return "", wrapCategory(CategoryUnsupported, fmt.Errorf("no thumbnail available"))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, thumbURL, nil)
	if err != nil {
		return "", wrapCategory(CategoryNetwork, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading thumbnail: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading thumbnail: unexpected status %d", resp.StatusCode))
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024*1024))
	if err != nil {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading thumbnail: %w", err))
	}

	ext := ".jpg"
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	switch {
	case strings.Contains(contentType, "png"), strings.HasSuffix(strings.ToLower(thumbURL), ".png"):
		ext = ".png"
	case strings.Contains(contentType, "webp"), strings.HasSuffix(strings.ToLower(thumbURL), ".webp"):
		ext = ".webp"
	}

	path, err := artifactPath(outputPath, ext, baseDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", wrapCategory(CategoryFilesystem, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("writing thumbnail: %w", err))
	}
	return path, nil
}

// embedThumbnailFile embeds a thumbnail image into the media file:
//   - mp3: ID3v2 APIC frame
//   - m4a/mp4 family: ffmpeg attached_pic
//   - mkv/opus/ogg: ffmpeg attachment
//   - webm: not supported by the muxer (warned by the caller)
func embedThumbnailFile(outputPath, thumbPath string, printer *Printer) error {
	if outputPath == "" || thumbPath == "" {
		return nil
	}
	ext := strings.ToLower(filepath.Ext(outputPath))
	thumbExt := strings.ToLower(filepath.Ext(thumbPath))
	if thumbExt == ".webp" && ext != ".mp3" {
		// The mp4/mjpeg path cannot use webp covers; convert to jpg first.
		converted := strings.TrimSuffix(thumbPath, thumbExt) + ".jpg"
		if err := ffmpeg.Input(thumbPath).Output(converted, ffmpeg.KwArgs{"q:v": "2"}).OverWriteOutput().Silent(true).Run(); err != nil {
			return fmt.Errorf("converting webp thumbnail: %w", err)
		}
		thumbPath = converted
	}

	switch ext {
	case ".mp3":
		return embedID3Picture(outputPath, thumbPath)
	case ".m4a", ".mp4", ".m4v", ".mov":
		return ffmpegAttachPicture(outputPath, thumbPath, "attached_pic")
	case ".mkv", ".opus", ".ogg":
		return ffmpegAttachPicture(outputPath, thumbPath, "attachment")
	default:
		return fmt.Errorf("thumbnail embedding not supported for %s files", ext)
	}
}

func embedID3Picture(outputPath, thumbPath string) error {
	data, err := os.ReadFile(thumbPath)
	if err != nil {
		return err
	}
	mime := "image/jpeg"
	switch strings.ToLower(filepath.Ext(thumbPath)) {
	case ".png":
		mime = "image/png"
	case ".webp":
		mime = "image/webp"
	}
	tag, err := id3v2.Open(outputPath, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    mime,
		PictureType: id3v2.PTFrontCover,
		Picture:     data,
	}
	tag.AddAttachedPicture(pic)
	return tag.Save()
}

// ffmpegAttachPicture remuxes the media with an attached cover image using
// stream copy. mode is "attached_pic" (mp4 family) or "attachment" (mkv).
func ffmpegAttachPicture(outputPath, thumbPath, mode string) error {
	dir := filepath.Dir(outputPath)
	tmpFile := filepath.Join(dir, ".tmp_cover_"+filepath.Base(outputPath))

	args := []string{
		"-y",
		"-i", outputPath,
		"-i", thumbPath,
		"-map", "0",
		"-map", "1",
		"-c", "copy",
		"-c:v:1", "mjpeg",
	}
	switch mode {
	case "attached_pic":
		args = append(args, "-disposition:v:1", "attached_pic")
	default:
		args = append(args, "-metadata:s:v:1", "title=Album cover", "-disposition:v:1", "attached_pic")
	}
	args = append(args, tmpFile)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpFile)
		stderr := strings.TrimSpace(string(output))
		if stderr != "" {
			return fmt.Errorf("failed to embed thumbnail: %s: %w", stderr, err)
		}
		return fmt.Errorf("failed to embed thumbnail: %w", err)
	}
	if err := os.Rename(tmpFile, outputPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace original file with cover-embedded version: %w", err)
	}
	return nil
}
