package downloader

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kkdai/youtube/v2"
)

type HLSManifest struct {
	Variants  []HLSVariant
	Segments  []HLSSegment
	Encrypted bool
	KeyMethod string
	KeyURI    string
}

type HLSVariant struct {
	URI        string
	Bandwidth  int
	Resolution string
	Codecs     string
}

type HLSSegment struct {
	URI      string
	Duration float64
}

func ParseHLSManifest(data []byte) (HLSManifest, error) {
	if !bytes.Contains(data, []byte("#EXTM3U")) {
		return HLSManifest{}, errors.New("not an HLS manifest")
	}

	var manifest HLSManifest
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var pendingVariant *HLSVariant
	var lastDuration float64
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			attrs := parseHLSAttributes(strings.TrimPrefix(line, "#EXT-X-STREAM-INF:"))
			variant := &HLSVariant{
				Bandwidth:  parseInt(attrs["BANDWIDTH"]),
				Resolution: attrs["RESOLUTION"],
				Codecs:     attrs["CODECS"],
			}
			pendingVariant = variant
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			durationText := strings.TrimPrefix(line, "#EXTINF:")
			durationText = strings.TrimSuffix(durationText, ",")
			if duration, err := strconv.ParseFloat(durationText, 64); err == nil {
				lastDuration = duration
			}
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-KEY:") {
			attrs := parseHLSAttributes(strings.TrimPrefix(line, "#EXT-X-KEY:"))
			method := strings.ToUpper(attrs["METHOD"])
			if method != "" && method != "NONE" {
				manifest.Encrypted = true
				manifest.KeyMethod = method
				manifest.KeyURI = attrs["URI"]
			}
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if pendingVariant != nil {
			pendingVariant.URI = line
			manifest.Variants = append(manifest.Variants, *pendingVariant)
			pendingVariant = nil
			continue
		}

		manifest.Segments = append(manifest.Segments, HLSSegment{
			URI:      line,
			Duration: lastDuration,
		})
		lastDuration = 0
	}

	if err := scanner.Err(); err != nil {
		return HLSManifest{}, err
	}

	return manifest, nil
}

func DetectHLSDrm(manifest HLSManifest) (bool, string) {
	if !manifest.Encrypted {
		return false, ""
	}
	if manifest.KeyMethod != "" {
		return true, manifest.KeyMethod
	}
	return true, "encrypted"
}

func DetectDASHDrm(data []byte) (bool, string) {
	lower := strings.ToLower(string(data))
	for _, marker := range dashDRMMarkers {
		if strings.Contains(lower, marker) {
			return true, marker
		}
	}
	return false, ""
}

func parseHLSAttributes(raw string) map[string]string {
	attrs := map[string]string{}
	parts := splitHLSAttributes(raw)
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToUpper(kv[0]))
		value := strings.TrimSpace(kv[1])
		value = strings.Trim(value, "\"")
		if key != "" {
			attrs[key] = value
		}
	}
	return attrs
}

func splitHLSAttributes(raw string) []string {
	var parts []string
	var b strings.Builder
	inQuotes := false
	for _, r := range raw {
		switch r {
		case '"':
			inQuotes = !inQuotes
			b.WriteRune(r)
		case ',':
			if inQuotes {
				b.WriteRune(r)
				continue
			}
			parts = append(parts, b.String())
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func parseInt(value string) int {
	if value == "" {
		return 0
	}
	num, _ := strconv.Atoi(value)
	return num
}

var dashDRMMarkers = []string{
	"widevine",
	"playready",
	"cenc",
	"urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed",
	"urn:uuid:9a04f079-9840-4286-ab92-e65be0885f95",
	"urn:mpeg:dash:mp4protection:2011",
}

type segmentDownloadPlan struct {
	URLs        []string
	TempDir     string
	Prefix      string
	Concurrency int
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
