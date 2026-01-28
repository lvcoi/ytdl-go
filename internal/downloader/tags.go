package downloader

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	id3v2 "github.com/bogem/id3v2/v2"
)

func embedAudioTags(metadata ItemMetadata, outputPath string, printer *Printer) {
	if outputPath == "" || metadata.Status != "ok" {
		return
	}
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".mp3":
		if err := embedID3Tags(metadata, outputPath); err != nil {
			if printer != nil {
				printer.Log(LogWarn, fmt.Sprintf("metadata tag embedding failed: %v", err))
			}
		}
	default:
		// Silently skip unsupported formats (.webm, .opus, etc.)
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
