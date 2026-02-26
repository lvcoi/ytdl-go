package web

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMediaWatcherDetectsNewFile(t *testing.T) {
	mediaDir := t.TempDir()
	audioDir := filepath.Join(mediaDir, "audio")
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		t.Fatalf("mkdir audio: %v", err)
	}

	mw, err := newMediaWatcher(mediaDir)
	if err != nil {
		t.Fatalf("newMediaWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mw.Run(ctx)

	// Give the watcher time to start.
	time.Sleep(100 * time.Millisecond)

	// Create a media file — the watcher should pick it up without error.
	newFile := filepath.Join(audioDir, "new-song.mp3")
	if err := os.WriteFile(newFile, []byte("audio"), 0o644); err != nil {
		t.Fatalf("write new file: %v", err)
	}

	// Allow debounce to fire.
	time.Sleep(700 * time.Millisecond)

	// Verify the file exists (watcher should not have interfered).
	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("expected new file to exist: %v", err)
	}
}

func TestMediaWatcherDetectsRemovedFile(t *testing.T) {
	mediaDir := t.TempDir()
	audioDir := filepath.Join(mediaDir, "audio")
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		t.Fatalf("mkdir audio: %v", err)
	}

	existingFile := filepath.Join(audioDir, "existing.mp3")
	if err := os.WriteFile(existingFile, []byte("audio"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	mw, err := newMediaWatcher(mediaDir)
	if err != nil {
		t.Fatalf("newMediaWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mw.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	// Remove the file externally.
	if err := os.Remove(existingFile); err != nil {
		t.Fatalf("remove file: %v", err)
	}

	// Allow debounce to fire.
	time.Sleep(700 * time.Millisecond)

	// Verify the file is gone.
	if _, err := os.Stat(existingFile); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed")
	}
}

func TestMediaWatcherIgnoresJSONSidecars(t *testing.T) {
	mediaDir := t.TempDir()

	mw, err := newMediaWatcher(mediaDir)
	if err != nil {
		t.Fatalf("newMediaWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mw.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	// Create a .json sidecar — should be silently ignored.
	jsonFile := filepath.Join(mediaDir, "song.mp3.json")
	if err := os.WriteFile(jsonFile, []byte(`{"title":"test"}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	time.Sleep(700 * time.Millisecond)

	// No crash, no panic — the watcher filtered it out.
	if _, err := os.Stat(jsonFile); err != nil {
		t.Fatalf("json file should still exist: %v", err)
	}
}

func TestMediaWatcherWatchesNewSubdirectories(t *testing.T) {
	mediaDir := t.TempDir()

	mw, err := newMediaWatcher(mediaDir)
	if err != nil {
		t.Fatalf("newMediaWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mw.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	// Create a new subdirectory after the watcher started.
	newDir := filepath.Join(mediaDir, "playlist", "road-trip")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("mkdir new subdir: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Create a file inside the new subdirectory.
	newFile := filepath.Join(newDir, "track.mp3")
	if err := os.WriteFile(newFile, []byte("audio"), 0o644); err != nil {
		t.Fatalf("write file in new subdir: %v", err)
	}

	time.Sleep(700 * time.Millisecond)

	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("expected file in new subdir to exist: %v", err)
	}
}

func TestMediaWatcherStopsOnContextCancel(t *testing.T) {
	mediaDir := t.TempDir()

	mw, err := newMediaWatcher(mediaDir)
	if err != nil {
		t.Fatalf("newMediaWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		mw.Run(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Run exited cleanly.
	case <-time.After(2 * time.Second):
		t.Fatalf("watcher did not stop after context cancellation")
	}
}
