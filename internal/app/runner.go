package app

import (
	"context"

	"github.com/lvcoi/ytdl-go/internal/downloader"
)

type Result struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
	Err   error  `json:"-"`
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
		// Note: ProgressManager uses Bubble Tea TUI which renders to os.Stderr.
		// In a web server context, this output goes to the server's stderr logs
		// rather than being visible to web clients. The Quiet option in opts
		// should be set to true by web clients to minimize this output.
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

	for _, url := range urls {
		select {
		case <-ctx.Done():
			close(tasks)
			goto done
		case tasks <- task{url: url}:
		}
	}
	close(tasks)

done:
	output := make([]Result, 0, len(urls))
	exitCode := 0
	for i := 0; i < len(urls); i++ {
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
