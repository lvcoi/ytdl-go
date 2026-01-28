package downloader

import (
	"context"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type Extractor interface {
	Name() string
	Match(url string) bool
	Process(ctx context.Context, url string, opts Options, printer *Printer) error
}

type YouTubeExtractor struct{}

func (e YouTubeExtractor) Name() string {
	return "youtube"
}

func (e YouTubeExtractor) Match(raw string) bool {
	if looksLikePlaylist(raw) {
		return true
	}
	return isYouTubeURL(raw)
}

func (e YouTubeExtractor) Process(ctx context.Context, raw string, opts Options, printer *Printer) error {
	url := ConvertMusicURL(raw)
	isMusicURL := strings.Contains(raw, "music.youtube.com")

	if looksLikePlaylist(url) {
		return processPlaylist(ctx, url, opts, printer, isMusicURL)
	}

	youtube.DefaultClient = youtube.AndroidClient
	client := newClient(opts)
	video, err := client.GetVideoContext(ctx, url)
	if err != nil {
		return wrapFetchError(err, "fetching metadata")
	}
	if opts.InfoOnly {
		return printInfo(video)
	}
	if opts.ListFormats {
		return listFormats(video, opts, printer)
	}

	prefix := printer.Prefix(1, 1, video.Title)
	result, err := downloadVideo(ctx, client, video, opts, outputContext{
		SourceURL:     url,
		MetaOverrides: opts.MetaOverrides,
	}, printer, prefix)
	if result.skipped {
		printer.ItemSkipped(prefix, "exists")
	} else {
		printer.ItemResult(prefix, result, err)
	}

	if opts.JSON {
		status := "ok"
		errMsg := ""
		if result.skipped {
			status = "skip"
			errMsg = "exists"
		} else if err != nil {
			status = "error"
			errMsg = err.Error()
		}
		emitJSONResult(jsonResult{
			Type:    "item",
			Status:  status,
			URL:     url,
			ID:      video.ID,
			Title:   video.Title,
			Output:  result.outputPath,
			Bytes:   result.bytes,
			Retries: result.retried,
			Error:   errMsg,
		})
	}

	if err != nil {
		return markReported(err)
	}
	okCount := 1
	skipped := 0
	if result.skipped {
		okCount = 0
		skipped = 1
	}
	printer.Summary(1, okCount, 0, skipped, result.bytes)
	return nil
}

type DirectExtractor struct{}

func (e DirectExtractor) Name() string {
	return "direct"
}

func (e DirectExtractor) Match(raw string) bool {
	return !isYouTubeURL(raw)
}

func (e DirectExtractor) Process(ctx context.Context, raw string, opts Options, printer *Printer) error {
	result, err := processDirect(ctx, raw, opts, printer)
	if opts.JSON {
		status := "ok"
		errMsg := ""
		if err != nil {
			status = "error"
			errMsg = err.Error()
		}
		emitJSONResult(jsonResult{
			Type:    "item",
			Status:  status,
			URL:     raw,
			Output:  result.outputPath,
			Bytes:   result.bytes,
			Retries: result.retried,
			Error:   errMsg,
		})
	}
	return err
}
