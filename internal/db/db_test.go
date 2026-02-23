package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("database file was not created")
	}
}

func TestInsertAndListMedia(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()

	record := MediaRecord{
		Title:     "Test Song",
		Artist:    "Test Artist",
		Album:     "Test Album",
		Duration:  240,
		MediaType: "music",
		FilePath:  "/tmp/test/song.mp3",
		SourceURL: "https://youtube.com/watch?v=abc",
		VideoID:   "abc",
		Format:    "mp3",
	}

	id, err := d.InsertMedia(record)
	if err != nil {
		t.Fatalf("InsertMedia failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	records, err := d.ListMedia(10, 0)
	if err != nil {
		t.Fatalf("ListMedia failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Title != "Test Song" {
		t.Fatalf("expected title 'Test Song', got %q", records[0].Title)
	}
	if records[0].MediaType != "music" {
		t.Fatalf("expected media_type 'music', got %q", records[0].MediaType)
	}
}

func TestUpsertMedia(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()

	record := MediaRecord{
		Title:     "Original Title",
		Artist:    "Artist",
		MediaType: "video",
		FilePath:  "/tmp/test/video.mp4",
		VideoID:   "xyz",
	}

	_, err = d.UpsertMedia(record)
	if err != nil {
		t.Fatalf("first UpsertMedia failed: %v", err)
	}

	record.Title = "Updated Title"
	record.TagsEmbedded = true
	_, err = d.UpsertMedia(record)
	if err != nil {
		t.Fatalf("second UpsertMedia failed: %v", err)
	}

	records, err := d.ListMedia(10, 0)
	if err != nil {
		t.Fatalf("ListMedia failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record after upsert, got %d", len(records))
	}
	if records[0].Title != "Updated Title" {
		t.Fatalf("expected updated title, got %q", records[0].Title)
	}
	if !records[0].TagsEmbedded {
		t.Fatalf("expected tags_embedded=true after upsert")
	}
}

func TestCount(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()

	count, err := d.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	for i := 0; i < 3; i++ {
		_, err := d.InsertMedia(MediaRecord{
			Title:     "Song",
			MediaType: "music",
			FilePath:  filepath.Join("/tmp/test", "song"+string(rune('A'+i))+".mp3"),
		})
		if err != nil {
			t.Fatalf("InsertMedia failed: %v", err)
		}
	}

	count, err = d.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestDuplicateFilePathReturnsError(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer d.Close()

	record := MediaRecord{
		Title:    "Song",
		FilePath: "/tmp/test/song.mp3",
	}

	_, err = d.InsertMedia(record)
	if err != nil {
		t.Fatalf("first InsertMedia failed: %v", err)
	}

	_, err = d.InsertMedia(record)
	if err == nil {
		t.Fatalf("expected error on duplicate file_path insert")
	}
}
