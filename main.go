package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lvcoi/ytdl-go/internal/downloader"
)

func main() {
	var opts downloader.Options
	var meta metaFlags
	var jobs int

	flag.StringVar(&opts.OutputTemplate, "o", "{title}.{ext}", "output path or template (supports {title}, {artist}, {album}, {id}, {ext}, {quality}, {playlist_title}, {playlist_id}, {index}, {count})")
	flag.BoolVar(&opts.AudioOnly, "audio", false, "download best available audio only")
	flag.BoolVar(&opts.InfoOnly, "info", false, "print video metadata as JSON without downloading")
	flag.BoolVar(&opts.ListFormats, "list-formats", false, "list available formats and exit")
	flag.StringVar(&opts.Quality, "quality", "", "preferred quality (e.g. 1080p, 720p, 128k, best, worst)")
	flag.StringVar(&opts.Format, "format", "", "preferred container/extension (e.g. mp4, webm, m4a)")
	flag.Var(&meta, "meta", "metadata override key=value (repeatable)")
	flag.StringVar(&opts.ProgressLayout, "progress-layout", "", "progress layout template (e.g. \"{label} {percent} {bar} {bytes} {rate} {eta}\")")
	flag.IntVar(&opts.SegmentConcurrency, "segment-concurrency", 0, "parallel segment downloads (0=auto)")
	flag.IntVar(&jobs, "jobs", 1, "number of concurrent downloads")
	flag.BoolVar(&opts.JSON, "json", false, "emit JSON output (suppresses human-readable progress)")
	flag.DurationVar(&opts.Timeout, "timeout", 3*time.Minute, "per-request timeout")
	flag.BoolVar(&opts.Quiet, "quiet", false, "suppress progress output (errors still shown)")
	flag.StringVar(&opts.LogLevel, "log-level", "info", "log level: debug, info, warn, error")
	flag.Parse()

	opts.MetaOverrides = meta.Values()

	urls := flag.Args()
	if len(urls) == 0 {
		err := downloader.CategorizedError{Category: downloader.CategoryInvalidURL, Err: errors.New("no url provided")}
		if opts.JSON {
			writeJSONError("", err)
		} else {
			fmt.Fprintf(os.Stderr, "usage: %s [options] <url> [url...]\n", os.Args[0])
			flag.PrintDefaults()
		}
		os.Exit(downloader.ExitCode(err))
	}

	if jobs < 1 {
		jobs = 1
	}
	if opts.JSON {
		opts.Quiet = true
	}

	type task struct {
		index int
		url   string
	}
	type result struct {
		url string
		err error
	}
	tasks := make(chan task)
	results := make(chan result, len(urls))
	ctx := context.Background()

	// Create shared progress manager for concurrent jobs
	var sharedManager *downloader.ProgressManager
	if jobs > 1 {
		sharedManager = downloader.NewProgressManager(opts)
		if sharedManager != nil {
			sharedManager.Start(ctx)
			defer sharedManager.Stop()
		}
	}

	for i := 0; i < jobs; i++ {
		go func() {
			for t := range tasks {
				var err error
				if jobs > 1 && sharedManager != nil {
					// Use shared progress manager for parallel downloads
					err = downloader.ProcessWithManager(ctx, t.url, opts, sharedManager)
				} else {
					// Single job uses its own progress manager
					err = downloader.Process(ctx, t.url, opts)
				}
				results <- result{url: t.url, err: err}
			}
		}()
	}

	for i, url := range urls {
		tasks <- task{index: i, url: url}
	}
	close(tasks)

	exitCode := 0
	for range urls {
		res := <-results
		if res.err != nil {
			if code := downloader.ExitCode(res.err); code > exitCode {
				exitCode = code
			}
			if opts.JSON {
				writeJSONError(res.url, res.err)
			} else if !downloader.IsReported(res.err) {
				fmt.Fprintf(os.Stderr, "error: %v\n", res.err)
			}
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func writeJSONError(url string, err error) {
	payload := struct {
		Type     string `json:"type"`
		URL      string `json:"url,omitempty"`
		Category string `json:"category"`
		Error    string `json:"error"`
	}{
		Type:     "error",
		URL:      url,
		Category: string(downloader.CategoryOf(err)),
		Error:    err.Error(),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

type metaFlags struct {
	values map[string]string
}

func (m *metaFlags) String() string {
	if m == nil || len(m.values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.values))
	for key, value := range m.values {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, ",")
}

func (m *metaFlags) Set(value string) error {
	if m.values == nil {
		m.values = map[string]string{}
	}
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid meta override %q (expected key=value)", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return fmt.Errorf("invalid meta override %q (empty key)", value)
	}
	m.values[strings.ToLower(key)] = val
	return nil
}

func (m *metaFlags) Values() map[string]string {
	if m == nil || len(m.values) == 0 {
		return nil
	}
	out := make(map[string]string, len(m.values))
	for k, v := range m.values {
		out[k] = v
	}
	return out
}
