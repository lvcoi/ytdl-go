package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kkdai/youtube/v2"
)

type segmentDownloadPlan struct {
	URLs        []string
	TempDir     string
	Prefix      string
	Concurrency int
	BaseDir     string
}

const (
	// minConcurrentDownloads is the minimum number of concurrent downloads
	// for I/O-bound operations, even on single-core systems
	minConcurrentDownloads = 4

	// ioMultiplier determines how many concurrent downloads per CPU core
	// Set to 1x to match CPU cores (users can adjust with -playlist-concurrency flag)
	ioMultiplier = 1
)

func defaultSegmentConcurrency(value int) int {
	if value > 0 {
		return value
	}
	// Use workers based on CPU count for playlist downloads
	// Users can adjust concurrency with the -playlist-concurrency flag
	cpu := runtime.NumCPU()
	if cpu < 2 {
		return minConcurrentDownloads
	}
	return cpu * ioMultiplier
}

func downloadSegmentsParallel(ctx context.Context, client *youtube.Client, plan segmentDownloadPlan, writer io.Writer, printer *Printer) (int64, error) {
	concurrency := defaultSegmentConcurrency(plan.Concurrency)
	if concurrency < 2 || len(plan.URLs) < 2 {
		return downloadSegmentsSequential(ctx, client, plan, writer, printer)
	}

	if err := os.MkdirAll(plan.TempDir, 0o755); err != nil {
		return 0, wrapCategory(CategoryFilesystem, fmt.Errorf("creating temp dir: %w", err))
	}

	var totalBytes int64
	progress := (*progressWriter)(nil)
	if printer != nil {
		progress = newProgressWriter(0, printer, plan.Prefix)
	}
	var progressMu sync.Mutex
	counter := &progressCounter{total: &totalBytes, progress: progress, mu: &progressMu}

	type job struct {
		Index int
		URL   string
	}
	jobs := make(chan job)
	var wg sync.WaitGroup
	var firstErr atomic.Value

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			if firstErr.Load() != nil {
				return
			}
			dest := filepath.Join(plan.TempDir, fmt.Sprintf("segment-%06d.part", j.Index))
			if info, err := os.Stat(dest); err == nil && info.Size() > 0 {
				continue
			}
			file, err := os.Create(dest)
			if err != nil {
				firstErr.Store(wrapCategory(CategoryFilesystem, fmt.Errorf("creating segment file: %w", err)))
				return
			}
			segmentWriter := io.Writer(file)
			if progress != nil {
				segmentWriter = io.MultiWriter(file, counter)
			}
			err = downloadSegmentWithRetry(ctx, client, j.URL, segmentWriter)
			file.Close()
			if err != nil {
				firstErr.Store(wrapCategory(CategoryNetwork, fmt.Errorf("segment %d failed: %w", j.Index+1, err)))
				return
			}
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker()
	}

	for i, url := range plan.URLs {
		jobs <- job{Index: i, URL: url}
	}
	close(jobs)
	wg.Wait()

	if err := firstErr.Load(); err != nil {
		if progress != nil {
			progress.NewLine()
		}
		return atomic.LoadInt64(&totalBytes), err.(error)
	}

	if progress != nil {
		progress.Finish()
	}

	for i := range plan.URLs {
		path := filepath.Join(plan.TempDir, fmt.Sprintf("segment-%06d.part", i))
		file, err := os.Open(path)
		if err != nil {
			return atomic.LoadInt64(&totalBytes), wrapCategory(CategoryFilesystem, fmt.Errorf("opening segment: %w", err))
		}
		if _, err := io.Copy(writer, file); err != nil {
			file.Close()
			return atomic.LoadInt64(&totalBytes), wrapCategory(CategoryFilesystem, fmt.Errorf("assembling segments: %w", err))
		}
		file.Close()
	}

	for i := range plan.URLs {
		path := filepath.Join(plan.TempDir, fmt.Sprintf("segment-%06d.part", i))
		_ = os.Remove(path)
	}

	return atomic.LoadInt64(&totalBytes), nil
}

func validateSegmentTempDir(tempDir, baseDir string) (string, error) {
	// Always root segment temp directories under a safe, application-controlled
	// base directory inside the system temp directory, regardless of user-
	// controlled output paths. This prevents arbitrary filesystem writes.
	safeBase := filepath.Join(os.TempDir(), "ytdl-go-segments")
	absBase, err := filepath.Abs(safeBase)
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("resolving base directory for temp segments: %w", err))
	}
	// Disallow using filesystem root directly as the base for temp segments.
	if absBase == string(filepath.Separator) {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("refusing to use filesystem root as base directory for temp segments"))
	}
	// Ensure the base directory exists and is a directory before using it for temp files.
	if info, err := os.Stat(absBase); err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(absBase, 0o755); mkErr != nil {
				return "", wrapCategory(CategoryFilesystem, fmt.Errorf("creating base directory for temp segments: %w", mkErr))
			}
		} else {
			return "", wrapCategory(CategoryFilesystem, fmt.Errorf("stat base directory for temp segments: %w", err))
		}
	} else if !info.IsDir() {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("base path for temp segments is not a directory: %q", absBase))

	// If no specific tempDir is requested, create a new directory under the safe base.
	}
	if tempDir == "" {
		created, err := os.MkdirTemp(absBase, "segments-")
		if err != nil {
			return "", wrapCategory(CategoryFilesystem, fmt.Errorf("creating temp segments dir: %w", err))
		}
		return created, nil
	}

	// If a tempDir is provided, treat it as a subpath of the safe base and ensure
	// it cannot escape that base.
	candidate := filepath.Join(absBase, tempDir)
	absTemp, err := filepath.Abs(candidate)
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("resolving temp directory: %w", err))
	}
	// Evaluate symlinks to prevent symlink-based directory traversal attacks.
	// This ensures that symlinks cannot be used to escape the base directory.
	evalBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("evaluating base directory symlinks: %w", err))
	}
	// Disallow using filesystem root directly as the base after resolving symlinks.
	if evalBase == string(filepath.Separator) {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("refusing to use filesystem root as base directory for temp segments after symlink evaluation"))
	}
	// For the temp directory, we need to handle the case where it doesn't exist yet.
	// EvalSymlinks requires the path to exist, so we evaluate the existing parent path.
	evalTemp := absTemp
	var statErr error
	if _, statErr = os.Lstat(absTemp); statErr == nil {
		// Path exists, evaluate it directly
		var evalErr error
		evalTemp, evalErr = filepath.EvalSymlinks(absTemp)
		if evalErr != nil {
			return "", wrapCategory(CategoryFilesystem, fmt.Errorf("evaluating temp directory symlinks: %w", evalErr))
		}
	} else if os.IsNotExist(statErr) {
		// Path doesn't exist, evaluate the closest existing parent
		originalAbsTemp := absTemp
		parent := filepath.Dir(absTemp)
		found := false
		// Walk up the directory tree until we find an existing parent
		for {
			if _, parentStatErr := os.Lstat(parent); parentStatErr == nil {
				evalParent, evalErr := filepath.EvalSymlinks(parent)
				if evalErr != nil {
					return "", wrapCategory(CategoryFilesystem, fmt.Errorf("evaluating parent directory symlinks: %w", evalErr))
				}
				// Compute the relative path from parent to the original temp path.
				relPath, relErr := filepath.Rel(parent, originalAbsTemp)
				if relErr != nil {
					return "", wrapCategory(CategoryFilesystem, fmt.Errorf("computing relative path: %w", relErr))
				}
				// NOTE SECURITY / RACE CONDITION:
				// At this point we have:
				//   - Resolved symlinks only for the *existing* parent directory (evalParent).
				//   - A relative suffix (relPath) that may refer to path components which do not yet exist.
				//
				// When we reconstruct evalTemp below, any symlinks that might later appear in the
				// non-existent portion of relPath will not have been validated here. This means the
				// security guarantee provided by this check only applies to path components that
				// exist at validation time.
				//
				// Callers that rely on strict confinement (e.g. ensuring the final path remains
				// under a specific base directory) SHOULD perform a second validation after
				// creating the intermediate directories, to ensure no symlinks were introduced
				// between validation time and creation time.
				// Reconstruct the full path using the evaluated parent and the (currently) unchecked suffix.
				evalTemp = filepath.Join(evalParent, relPath)
				found = true
				break
			} else if !os.IsNotExist(parentStatErr) {
				return "", wrapCategory(CategoryFilesystem, fmt.Errorf("stat parent directory: %w", parentStatErr))
			}
			// Move up one level; stop if we've reached the root
			nextParent := filepath.Dir(parent)
			if nextParent == parent {
				// We've reached the filesystem root without finding an existing directory
				break
			}
			parent = nextParent
		}
		if !found {
			// This shouldn't happen in normal operation since at least "/" or "C:\" should exist.
			// Return an error rather than proceeding with an unevaluated path.
			return "", wrapCategory(CategoryFilesystem, fmt.Errorf("could not find existing parent directory for temp path validation"))
		}
	} else {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("stat temp directory: %w", statErr))
	}
	rel, err := filepath.Rel(evalBase, evalTemp)
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("relating temp directory: %w", err))
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("temp dir %q escapes base directory %q", absTemp, absBase))
	if info, err := os.Stat(absTemp); err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(absTemp, 0o755); mkErr != nil {
				return "", wrapCategory(CategoryFilesystem, fmt.Errorf("creating temp segments dir: %w", mkErr))
			}
		} else {
			return "", wrapCategory(CategoryFilesystem, fmt.Errorf("stat temp directory for segments: %w", err))
		}
	} else if !info.IsDir() {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("temp path for segments is not a directory: %q", absTemp))
	}
	}
	return evalTemp, nil
}

func downloadSegmentsSequential(ctx context.Context, client *youtube.Client, plan segmentDownloadPlan, writer io.Writer, printer *Printer) (int64, error) {
	progress := (*progressWriter)(nil)
	if printer != nil {
		progress = newProgressWriter(0, printer, plan.Prefix)
	}
	var totalBytes int64
	var progressMu sync.Mutex
	counter := &progressCounter{total: &totalBytes, progress: progress, mu: &progressMu}
	for i, url := range plan.URLs {
		segmentWriter := writer
		if progress != nil {
			segmentWriter = io.MultiWriter(writer, counter)
		}
		if err := downloadSegmentWithRetry(ctx, client, url, segmentWriter); err != nil {
			if progress != nil {
				progress.NewLine()
			}
			return totalBytes, wrapCategory(CategoryNetwork, fmt.Errorf("segment %d failed: %w", i+1, err))
		}
	}
	if progress != nil {
		progress.Finish()
	}
	return totalBytes, nil
}

type progressCounter struct {
	total    *int64
	progress *progressWriter
	mu       *sync.Mutex
}

func (pc *progressCounter) Write(p []byte) (int, error) {
	if pc == nil || pc.total == nil {
		return len(p), nil
	}
	n := len(p)
	updated := atomic.AddInt64(pc.total, int64(n))
	if pc.progress != nil && pc.mu != nil {
		pc.mu.Lock()
		pc.progress.SetCurrent(updated)
		pc.mu.Unlock()
	}
	return n, nil
}
