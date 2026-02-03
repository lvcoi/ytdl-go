package app

import (
	"context"

	"github.com/lvcoi/ytdl-go/internal/downloader"
)

type Result struct {
	URL string
	Err error
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

	for i := 0; i < jobs; i++ {
		go func() {
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
					select {
					case results <- Result{URL: t.url, Err: err}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

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
	output := make([]Result, 0, submitted)
	exitCode := 0
	for i := 0; i < submitted; i++ {
		select {
		case <-ctx.Done():
			if sharedManager != nil {
				sharedManager.Stop()
			}
			return output, 130
		case res := <-results:
			output = append(output, res)
			if res.Err != nil {
				if code := downloader.ExitCode(res.Err); code > exitCode {
					exitCode = code
				}
			}
		}
	}

	return output, exitCode
}
