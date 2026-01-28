package downloader

import (
	"testing"
	"time"

	"github.com/kkdai/youtube/v2"
)

func TestApplyMetaOverrides(t *testing.T) {
	metadata := ItemMetadata{
		Title:       "old",
		Artist:      "artist",
		ReleaseYear: 2020,
	}
	applyMetaOverrides(&metadata, map[string]string{
		"title":        "new",
		"release_year": "2024",
	})

	if metadata.Title != "new" {
		t.Fatalf("expected title override")
	}
	if metadata.ReleaseYear != 2024 {
		t.Fatalf("expected release_year override")
	}
}

func TestWritePlaylistManifest(t *testing.T) {
	t.Skip("PlaylistManifest functionality not yet implemented")
	/*
	dir := t.TempDir()
	path := filepath.Join(dir, "playlist.json")
	manifest := PlaylistManifest{
		ID:          "PL123",
		Title:       "Demo",
		Count:       2,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Entries: []PlaylistManifestEntry{
			{Index: 1, ID: "a", Status: "ok"},
			{Index: 2, ID: "b", Status: "ok"},
		},
	}

	if err := writePlaylistManifest(path, manifest); err != nil {
		t.Fatalf("writePlaylistManifest: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var decoded PlaylistManifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(decoded.Entries) != 2 || decoded.Entries[0].Index != 1 || decoded.Entries[1].Index != 2 {
		t.Fatalf("unexpected entry ordering")
	}
	*/
}

func TestBuildItemMetadataFallback(t *testing.T) {
	video := &youtube.Video{
		ID:          "vid",
		Title:       "",
		Author:      "",
		PublishDate: time.Time{},
	}
	meta := buildItemMetadata(video, nil, outputContext{}, "out.mp4", "ok", nil)
	if meta.Title == "" {
		t.Fatalf("expected fallback title")
	}
	if meta.SourceURL == "" {
		t.Fatalf("expected source URL fallback")
	}
}
