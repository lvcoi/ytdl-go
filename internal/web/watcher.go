package web

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// mediaWatcher monitors the media directory for external file changes
// (create, remove, rename) and broadcasts library_update WebSocket events
// so connected UI clients can refresh without polling.
type mediaWatcher struct {
	mediaDir string
	watcher  *fsnotify.Watcher

	// debounce coalesces rapid bursts of filesystem events into a single
	// library_update broadcast.
	debounceMu    sync.Mutex
	debounceTimer *time.Timer
}

const (
	watcherDebounceInterval = 500 * time.Millisecond
)

// newMediaWatcher creates a watcher that recursively monitors mediaDir.
// Call Run to start processing events; cancel ctx to stop.
func newMediaWatcher(mediaDir string) (*mediaWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	mw := &mediaWatcher{
		mediaDir: mediaDir,
		watcher:  w,
	}

	// Add the root and all existing subdirectories.
	if err := mw.addRecursive(mediaDir); err != nil {
		w.Close()
		return nil, err
	}

	return mw, nil
}

// addRecursive walks dir and adds every directory to the watcher.
func (mw *mediaWatcher) addRecursive(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if d.IsDir() {
			if watchErr := mw.watcher.Add(path); watchErr != nil {
				log.Printf("watcher: failed to watch %q: %v", path, watchErr)
			}
		}
		return nil
	})
}

// Run processes filesystem events until ctx is cancelled.
func (mw *mediaWatcher) Run(ctx context.Context) {
	defer mw.watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-mw.watcher.Events:
			if !ok {
				return
			}
			mw.handleEvent(event)

		case err, ok := <-mw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher: error: %v", err)
		}
	}
}

func (mw *mediaWatcher) handleEvent(event fsnotify.Event) {
	// Ignore sidecar JSON, DB files, and the data folder internals.
	ext := strings.ToLower(filepath.Ext(event.Name))
	if ext == ".json" || ext == ".db" || ext == ".db-shm" || ext == ".db-wal" || ext == ".db-journal" {
		return
	}

	// If a new directory was created, start watching it too.
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			_ = mw.addRecursive(event.Name)
		}
	}

	// Only act on media-relevant operations.
	if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) {
		return
	}

	relPath, err := filepath.Rel(mw.mediaDir, event.Name)
	if err != nil {
		return
	}
	relPath = filepath.ToSlash(relPath)

	action := "changed"
	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		action = "deleted"

		// Clean up the DB record for removed files.
		if globalDB != nil {
			if _, delErr := globalDB.DeleteMediaByPath(relPath); delErr != nil {
				log.Printf("watcher: failed to delete DB record for %q: %v", relPath, delErr)
			}
		}
	} else if event.Has(fsnotify.Create) {
		action = "created"
	}

	// Debounce: coalesce rapid events into one broadcast.
	mw.debounceMu.Lock()
	if mw.debounceTimer != nil {
		mw.debounceTimer.Stop()
	}
	mw.debounceTimer = time.AfterFunc(watcherDebounceInterval, func() {
		broadcastLibraryUpdate(action, relPath)
	})
	mw.debounceMu.Unlock()
}
