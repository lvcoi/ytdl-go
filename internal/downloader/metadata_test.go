package downloader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFinalizeDownloadMetadataWritesSidecar(t *testing.T) {
	baseDir := t.TempDir()
	outputPath := filepath.Join(baseDir, "video", "sample.mp4")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte("video"), 0o644); err != nil {
		t.Fatalf("write output file: %v", err)
	}

	metadata := ItemMetadata{
		ID:              "vid123",
		Title:           "Sample Title",
		Artist:          "Sample Artist",
		Album:           "Sample Album",
		ReleaseDate:     "2026-02-08",
		DurationSeconds: 215,
		ThumbnailURL:    "https://i.ytimg.com/vi/vid123/hqdefault.jpg",
		SourceURL:       "https://www.youtube.com/watch?v=vid123",
		Extractor:       extractorName,
		Output:          outputPath,
		Format:          "mp4",
		Status:          "ok",
		Playlist: &PlaylistRef{
			ID:    "PL123",
			Title: "Playlist",
			URL:   "https://www.youtube.com/playlist?list=PL123",
			Index: 1,
			Count: 10,
		},
	}

	if err := finalizeDownloadMetadata(outputPath, baseDir, metadata, false, nil); err != nil {
		t.Fatalf("finalizeDownloadMetadata: %v", err)
	}

	sidecarPath := outputPath + ".json"
	raw, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatalf("read sidecar: %v", err)
	}

	var parsed ItemMetadata
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse sidecar json: %v", err)
	}
	if parsed.Artist != metadata.Artist {
		t.Fatalf("expected artist %q, got %q", metadata.Artist, parsed.Artist)
	}
	if parsed.Album != metadata.Album {
		t.Fatalf("expected album %q, got %q", metadata.Album, parsed.Album)
	}
	if parsed.ThumbnailURL != metadata.ThumbnailURL {
		t.Fatalf("expected thumbnail_url %q, got %q", metadata.ThumbnailURL, parsed.ThumbnailURL)
	}
	if parsed.SourceURL != metadata.SourceURL {
		t.Fatalf("expected source_url %q, got %q", metadata.SourceURL, parsed.SourceURL)
	}
	if parsed.ReleaseDate != metadata.ReleaseDate {
		t.Fatalf("expected release_date %q, got %q", metadata.ReleaseDate, parsed.ReleaseDate)
	}
	if parsed.DurationSeconds != metadata.DurationSeconds {
		t.Fatalf("expected duration_seconds %d, got %d", metadata.DurationSeconds, parsed.DurationSeconds)
	}
	if parsed.Playlist == nil || parsed.Playlist.ID != "PL123" {
		t.Fatalf("expected playlist metadata in sidecar, got %+v", parsed.Playlist)
	}
}
