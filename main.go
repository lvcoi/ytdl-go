package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lvcoi/ytdl-go/internal/downloader"
)

func main() {
	var opts downloader.Options

	flag.StringVar(&opts.OutputTemplate, "o", "{title}.{ext}", "output path or template (supports {title}, {artist}, {album}, {id}, {ext}, {quality}, {playlist_title}, {playlist_id}, {index}, {count})")
	flag.BoolVar(&opts.AudioOnly, "audio", false, "download best available audio only")
	flag.BoolVar(&opts.InfoOnly, "info", false, "print video metadata as JSON without downloading")
	flag.BoolVar(&opts.ListFormats, "list-formats", false, "list available formats without downloading")
	flag.DurationVar(&opts.Timeout, "timeout", 3*time.Minute, "per-request timeout")
	flag.BoolVar(&opts.Quiet, "quiet", false, "suppress progress output (errors still shown)")
	flag.StringVar(&opts.ProgressLayout, "progress-layout", "", "progress layout template (e.g. \"{label} {percent} {bar} {bytes} {rate} {eta}\")")
	flag.StringVar(&opts.LogLevel, "log-level", "info", "log level: debug, info, warn, error")
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <url> [url...]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	for i, url := range urls {
		fmt.Fprintf(os.Stderr, "\n[%d/%d] %s\n", i+1, len(urls), url)

		err := downloader.Process(context.Background(), url, opts)
		if err != nil {
			if !downloader.IsReported(err) {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			continue
		}
	}
}
