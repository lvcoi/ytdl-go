package app

import (
	"context"
	"sync"

	"github.com/lvcoi/ytdl-go/internal/downloader"
)

type Result struct {
	URL   string `json:"url"`
	Err   error  `json:"-"`
	Error string `json:"error,omitempty"`
}

func Run(ctx context.Context, urls []string, opts downloader.Options, jobs int) ([]Result, int) {
	if jobs < 1 {
		jobs = 1
	}

	type task struct {
		url string
	}
	tasks := make(chan task)
	results := make(chan Result, len(urls))

	var sharedManager *downloader.ProgressManager
	if jobs > 1 {
		sharedManager = downloader.NewProgressManager(opts)
		if sharedManager != nil {
			sharedManager.Start(ctx)
			defer sharedManager.Stop()
		}
	}

	// Use WaitGroup to track worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-tasks:
					if !ok {
						return
					}
					var err error
					if jobs > 1 && sharedManager != nil {
						err = downloader.ProcessWithManager(ctx, t.url, opts, sharedManager)
					} else {
						err = downloader.Process(ctx, t.url, opts)
					}
					result := Result{URL: t.url, Err: err}
					if err != nil {
						result.Error = err.Error()
					}
					select {
					case results <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Track number of tasks actually submitted
	submitted := 0
	for _, url := range urls {
		select {
		case <-ctx.Done():
			close(tasks)
			goto done
		case tasks <- task{url: url}:
			submitted++
		}
	}
	close(tasks)

done:
	// Close results channel after all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	output := make([]Result, 0, submitted)
	exitCode := 0
	contextCancelled := false
	
	// Collect results from submitted tasks only.
	// The range loop will exit when the results channel is closed
	// (which happens after all workers finish via the WaitGroup).
	for res := range results {
		output = append(output, res)
		if res.Err != nil {
			if code := downloader.ExitCode(res.Err); code > exitCode {
				exitCode = code
			}
		}
		// Track if context was cancelled during collection
		select {
		case <-ctx.Done():
			contextCancelled = true
		default:
		}
	}
	
	// If context was cancelled at any point, use exit code 130 (interrupted)
	if contextCancelled && exitCode == 0 {
		exitCode = 130
	}

	return output, exitCode
}
