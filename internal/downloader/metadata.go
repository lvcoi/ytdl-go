package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

const extractorName = "kkdai/youtube"

type ItemMetadata struct {
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	Artist           string       `json:"artist,omitempty"`
	Author           string       `json:"author,omitempty"`
	Album            string       `json:"album,omitempty"`
	Track            int          `json:"track,omitempty"`
	Disc             int          `json:"disc,omitempty"`
	ReleaseDate      string       `json:"release_date,omitempty"`
	ReleaseYear      int          `json:"release_year,omitempty"`
	DurationSeconds  int          `json:"duration_seconds,omitempty"`
	ThumbnailURL     string       `json:"thumbnail_url,omitempty"`
	SourceURL        string       `json:"source_url"`
	Extractor        string       `json:"extractor"`
	ExtractorVersion string       `json:"extractor_version,omitempty"`
	Output           string       `json:"output,omitempty"`
	Format           string       `json:"format,omitempty"`
	Quality          string       `json:"quality,omitempty"`
	Status           string       `json:"status"`
	Error            string       `json:"error,omitempty"`
	Playlist         *PlaylistRef `json:"playlist,omitempty"`
	Warnings         []string     `json:"warnings,omitempty"`
}

type PlaylistRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
	Index int    `json:"index,omitempty"`
	Count int    `json:"count,omitempty"`
}

func buildItemMetadata(video *youtube.Video, format *youtube.Format, ctxInfo outputContext, outputPath string, status string, err error) ItemMetadata {
	title := stringsOrFallback(ctxInfo.EntryTitle, video.Title, video.ID)
	artist := stringsOrFallback(ctxInfo.EntryAuthor, video.Author)
	album := ctxInfo.EntryAlbum

	warnings := []string{}
	if title == "" {
		title = video.ID
		warnings = append(warnings, "missing title")
	}
	if artist == "" {
		warnings = append(warnings, "missing artist")
	}

	sourceURL := ctxInfo.SourceURL
	if sourceURL == "" && video.ID != "" {
		sourceURL = watchURLForID(video.ID)
	}

	metadata := ItemMetadata{
		ID:               video.ID,
		Title:            title,
		Artist:           artist,
		Author:           video.Author,
		Album:            album,
		Track:            ctxInfo.Index,
		ReleaseDate:      formatDate(video.PublishDate),
		ReleaseYear:      formatYear(video.PublishDate),
		DurationSeconds:  int(video.Duration.Seconds()),
		ThumbnailURL:     bestThumbnailURL(video.Thumbnails),
		SourceURL:        sourceURL,
		Extractor:        extractorName,
		ExtractorVersion: extractorVersion(),
		Output:           outputPath,
		Status:           status,
		Warnings:         warnings,
	}

	if format != nil {
		metadata.Format = mimeToExt(format.MimeType)
		if format.QualityLabel != "" {
			metadata.Quality = format.QualityLabel
		} else if b := bitrateForFormat(format); b > 0 {
			metadata.Quality = fmt.Sprintf("%dk", b/1000)
		}
	}

	if err != nil {
		metadata.Error = err.Error()
	}

	if ctxInfo.Playlist != nil {
		metadata.Playlist = &PlaylistRef{
			ID:    ctxInfo.Playlist.ID,
			Title: ctxInfo.Playlist.Title,
			URL:   ctxInfo.PlaylistURL,
			Index: ctxInfo.Index,
			Count: ctxInfo.Total,
		}
	}

	if len(ctxInfo.MetaOverrides) > 0 {
		applyMetaOverrides(&metadata, ctxInfo.MetaOverrides)
	}

	return metadata
}

func writeSidecar(outputPath string, metadata ItemMetadata) error {
	if outputPath == "" {
		return nil
	}
	path := sidecarPath(outputPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("creating sidecar directory: %w", err))
	}

	file, err := os.Create(path)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("creating sidecar: %w", err))
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(metadata); err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("writing sidecar: %w", err))
	}
	return nil
}

func sidecarPath(outputPath string) string {
	return outputPath + ".json"
}

func bestThumbnailURL(thumbnails youtube.Thumbnails) string {
	bestURL := ""
	var bestArea uint
	for _, thumb := range thumbnails {
		area := thumb.Width * thumb.Height
		if area >= bestArea {
			bestArea = area
			bestURL = thumb.URL
		}
	}
	return bestURL
}

func extractorVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/kkdai/youtube/v2" {
				return dep.Version
			}
		}
	}
	return ""
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatYear(value time.Time) int {
	if value.IsZero() {
		return 0
	}
	return value.Year()
}

func stringsOrFallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func applyMetaOverrides(metadata *ItemMetadata, overrides map[string]string) {
	if metadata == nil || len(overrides) == 0 {
		return
	}
	for key, value := range overrides {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "title":
			metadata.Title = value
		case "artist":
			metadata.Artist = value
		case "author":
			metadata.Author = value
		case "album":
			metadata.Album = value
		case "track":
			if parsed, err := strconv.Atoi(value); err == nil {
				metadata.Track = parsed
			} else {
				metadata.Warnings = append(metadata.Warnings, "invalid track override")
			}
		case "disc":
			if parsed, err := strconv.Atoi(value); err == nil {
				metadata.Disc = parsed
			} else {
				metadata.Warnings = append(metadata.Warnings, "invalid disc override")
			}
		case "release_date":
			metadata.ReleaseDate = value
		case "release_year":
			if parsed, err := strconv.Atoi(value); err == nil {
				metadata.ReleaseYear = parsed
			} else {
				metadata.Warnings = append(metadata.Warnings, "invalid release_year override")
			}
		case "source_url":
			metadata.SourceURL = value
		}
	}
}
