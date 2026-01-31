package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/kkdai/youtube/v2"
)

type segmentDownloadPlan struct {
	URLs        []string
	TempDir     string
	Prefix      string
	Concurrency int
}

const (
	// minConcurrentDownloads is the minimum number of concurrent downloads
	// for I/O-bound operations, even on single-core systems
	minConcurrentDownloads = 8
	
	// ioMultiplier determines how many concurrent downloads per CPU core
	// Downloads are I/O-bound (waiting for network), not CPU-bound,
	// so we can run 2x more workers than CPU cores for better throughput
	ioMultiplier = 2
)

func defaultSegmentConcurrency(value int) int {
	if value > 0 {
		return value
	}
	// Use more workers than CPUs for I/O-bound playlist downloads
	// Most time is spent waiting for network I/O, not CPU
	cpu := runtime.NumCPU()
	if cpu < 2 {
		return minConcurrentDownloads
	}
	// Use 2x CPU count for better throughput on I/O-bound operations
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
