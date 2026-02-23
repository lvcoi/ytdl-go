package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	id3v2 "github.com/bogem/id3v2/v2"
)

// embedAudioTags attempts to embed metadata tags into the given audio file.
// For .mp3 files, ID3v2 tags are used directly.
// For other formats (.m4a, .mp4, .webm, .opus, .ogg, .mkv), FFmpeg is attempted.
func embedAudioTags(metadata ItemMetadata, outputPath string, printer *Printer) {
	if outputPath == "" || metadata.Status != "ok" {
		return
	}
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".mp3":
		if err := embedID3Tags(metadata, outputPath); err != nil {
			if printer != nil {
				printer.Log(LogWarn, fmt.Sprintf("warning: metadata tag embedding failed: %v", err))
			}
		}
	case ".m4a", ".mp4", ".webm", ".opus", ".ogg", ".mkv":
		if err := embedFFmpegTags(metadata, outputPath); err != nil {
			if printer != nil {
				printer.Log(LogWarn, fmt.Sprintf("warning: ffmpeg metadata embedding failed for %s: %v", ext, err))
			}
		}
	default:
		// Silently skip unsupported formats
	}
}

func embedID3Tags(metadata ItemMetadata, outputPath string) error {
	tag, err := id3v2.Open(outputPath, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	if metadata.Title != "" {
		tag.SetTitle(metadata.Title)
	}
	if metadata.Artist != "" {
		tag.SetArtist(metadata.Artist)
	}
	if metadata.Album != "" {
		tag.SetAlbum(metadata.Album)
	}
	if metadata.ReleaseYear != 0 {
		tag.AddTextFrame(tag.CommonID("Year"), tag.DefaultEncoding(), strconv.Itoa(metadata.ReleaseYear))
	}
	if metadata.Track != 0 {
		tag.AddTextFrame(tag.CommonID("Track number/Position in set"), tag.DefaultEncoding(), strconv.Itoa(metadata.Track))
	}
	return tag.Save()
}

// embedFFmpegTags uses ffmpeg to embed metadata into formats that don't support ID3.
func embedFFmpegTags(metadata ItemMetadata, outputPath string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Build metadata arguments
	args := []string{
		"-i", outputPath,
		"-y",
		"-c", "copy",
	}
	if metadata.Title != "" {
		args = append(args, "-metadata", "title="+metadata.Title)
	}
	if metadata.Artist != "" {
		args = append(args, "-metadata", "artist="+metadata.Artist)
	}
	if metadata.Album != "" {
		args = append(args, "-metadata", "album="+metadata.Album)
	}
	if metadata.ReleaseYear != 0 {
		args = append(args, "-metadata", "date="+strconv.Itoa(metadata.ReleaseYear))
	}
	if metadata.Track != 0 {
		args = append(args, "-metadata", "track="+strconv.Itoa(metadata.Track))
	}

	// Write to a temp file then rename
	dir := filepath.Dir(outputPath)
	tmpFile := filepath.Join(dir, ".tmp_tagged_"+filepath.Base(outputPath))
	args = append(args, tmpFile)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp file on failure
		os.Remove(tmpFile)
		stderr := strings.TrimSpace(string(output))
		if stderr != "" {
			return fmt.Errorf("failed to embed metadata for output format %s: %s: %w", filepath.Ext(outputPath), stderr, err)
		}
		return fmt.Errorf("failed to embed metadata for output format %s: %w", filepath.Ext(outputPath), err)
	}

	// Replace original with tagged version
	if err := os.Rename(tmpFile, outputPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace original file with tagged version: %w", err)
	}

	return nil
}
