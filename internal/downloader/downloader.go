package downloader

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	youtube "github.com/lvcoi/ytdl-lib/v2"
)

// Options describes CLI behavior for a download run.
type Options struct {
	OutputTemplate      string
	OutputDir           string
	AudioOnly           bool
	InfoOnly            bool
	ListFormats         bool
	Quiet               bool
	JSON                bool
	Quality             string
	Format              string
	Itag                int
	MetaOverrides       map[string]string
	SegmentConcurrency  int
	PlaylistConcurrency int
	Timeout             time.Duration
	ProgressLayout      string
	LogLevel            string
	Renderer            ProgressRenderer  `json:"-"`
	OnDuplicate         DuplicatePolicy   `json:"on-duplicate,omitempty"`
	DuplicatePrompter   DuplicatePrompter `json:"-"`
	DuplicateSession    *DuplicateSession `json:"-"`
	UseCookies          bool
	PoToken             string

	// yt-dlp inspired feature flags
	Proxy               string        `json:"proxy,omitempty"`              // http/https/socks5 proxy URL
	WriteThumbnail      bool          `json:"write-thumbnail,omitempty"`
	EmbedThumbnail      bool          `json:"embed-thumbnail,omitempty"`
	AudioFormat         string        `json:"audio-format,omitempty"`       // mp3, opus, m4a, aac, flac, wav, best
	AudioQuality        string        `json:"audio-quality,omitempty"`      // e.g. "160k", "0" (VBR q)
	WriteSubs           bool          `json:"write-subs,omitempty"`
	WriteAutoSubs       bool          `json:"write-auto-subs,omitempty"`
	SubLangs            string        `json:"sub-langs,omitempty"`          // comma separated, "all" for everything
	EmbedSubs           bool          `json:"embed-subs,omitempty"`
	SponsorblockMark    string        `json:"sponsorblock-mark,omitempty"`   // comma separated categories
	SponsorblockRemove  string        `json:"sponsorblock-remove,omitempty"` // comma separated categories
	PlaylistItems       string        `json:"playlist-items,omitempty"`      // e.g. "1:5,8,10:12"
	MatchFilter         string        `json:"match-filter,omitempty"`        // e.g. "duration<600&view_count>1000"
	DownloadArchive     string        `json:"download-archive,omitempty"`
	BreakOnExisting     bool          `json:"break-on-existing,omitempty"`
	LimitRate           string        `json:"limit-rate,omitempty"`          // e.g. "500K", "2M"
	SleepInterval       time.Duration `json:"sleep-interval,omitempty"`
	MaxSleepInterval    time.Duration `json:"max-sleep-interval,omitempty"`
	SleepRequests       time.Duration `json:"sleep-requests,omitempty"`
	CookiesFile         string        `json:"cookies-file,omitempty"`
	CookiesFromBrowser  string        `json:"cookies-from-browser,omitempty"` // chrome, edge, brave, firefox
	Live                bool          `json:"live,omitempty"`
	LiveSegmentInterval time.Duration `json:"-"` // poll interval for live manifests (0=auto)

	// Session-scoped helpers (not serialized)
	ArchiveSession *downloadArchive `json:"-"`
}

type outputContext struct {
	Playlist      *youtube.Playlist
	Index         int
	Total         int
	EntryTitle    string
	EntryAuthor   string
	EntryAlbum    string
	SourceURL     string
	PlaylistURL   string
	MetaOverrides map[string]string
	Autonumber    int // per-run sequential number for {autonumber}
}

type downloadResult struct {
	bytes       int64
	outputPath  string
	format      *youtube.Format
	retried     bool
	hadProgress bool
	skipped     bool
}

type reportedError struct {
	err error
}

func (e reportedError) Error() string {
	return e.err.Error()
}

func (e reportedError) Unwrap() error {
	return e.err
}

func markReported(err error) error {
	if err == nil {
		return nil
	}
	return reportedError{err: err}
}

// IsReported returns true if the error has already been printed to stderr.
func IsReported(err error) bool {
	var re reportedError
	return stderrors.As(err, &re)
}

func categoryForYouTubeError(err error) ErrorCategory {
	switch {
	case stderrors.Is(err, youtube.ErrLoginRequired),
		stderrors.Is(err, youtube.ErrVideoPrivate),
		stderrors.Is(err, youtube.ErrNotPlayableInEmbed):
		return CategoryRestricted
	case stderrors.Is(err, youtube.ErrInvalidPlaylist),
		stderrors.Is(err, youtube.ErrInvalidCharactersInVideoID),
		stderrors.Is(err, youtube.ErrVideoIDMinLength):
		return CategoryInvalidURL
	}

	var statusErr *youtube.ErrPlayabiltyStatus
	if stderrors.As(err, &statusErr) {
		return CategoryRestricted
	}

	return CategoryNetwork
}

func wrapFetchError(err error, context string) error {
	category := categoryForYouTubeError(err)
	wrapped := fmt.Errorf("%s: %w", context, err)
	switch category {
	case CategoryRestricted:
		wrapped = fmt.Errorf("restricted content (login/paywall/age/private): %w", wrapped)
	case CategoryInvalidURL:
		wrapped = fmt.Errorf("invalid URL or playlist ID: %w", wrapped)
	}
	return wrapCategory(category, wrapped)
}

// Process fetches metadata, selects the best matching format, and downloads it.
func Process(ctx context.Context, url string, opts Options) error {
	return ProcessWithManager(ctx, url, opts, nil)
}

// ProcessWithManager is like Process but allows sharing a progress manager across
// multiple concurrent downloads. If manager is nil, a new one is created.
func ProcessWithManager(ctx context.Context, url string, opts Options, manager *ProgressManager) error {
	var ownedManager *ProgressManager
	if manager == nil && opts.Renderer == nil {
		ownedManager = NewProgressManager(opts)
		if ownedManager != nil {
			ownedManager.Start(ctx)
			defer ownedManager.Stop()
		}
		manager = ownedManager
	}
	printer := newPrinter(opts, manager)

	normalizedURL, err := validateInputURL(url)
	if err != nil {
		return err
	}
	originalURL := normalizedURL
	url = ConvertMusicURL(normalizedURL)
	url = NormalizeYouTubeURL(url)

	// Detect YouTube Music URLs by parsing and normalizing the hostname
	isMusicURL := isMusicYouTubeURL(originalURL)

	if looksLikePlaylist(url) {
		return processPlaylist(ctx, url, opts, printer, isMusicURL)
	}

	if !isYouTubeURL(url) {
		result, err := processDirect(ctx, url, opts, printer)
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
				URL:     url,
				Output:  result.outputPath,
				Bytes:   result.bytes,
				Retries: result.retried,
				Error:   errMsg,
			})
		}
		return err
	}

	client := newClientTypeForType("android", opts)
	video, err := client.GetVideoContext(ctx, url)
	if err != nil {
		return wrapFetchError(err, "fetching video metadata")
	}

	if opts.InfoOnly {
		return printVideoInfo(video)
	}
	if opts.ListFormats {
		return renderFormats(video, opts, "", "", 0, 0)
	}

	ctxInfo := outputContext{}
	ctxInfo.Autonumber = nextAutonumber()
	prefix := printer.Prefix(1, 1, video.Title)
	result, err := downloadVideo(ctx, client, video, opts, ctxInfo, printer, prefix)
	if err != nil {
		printer.ItemResult(prefix, result, err)
		if opts.JSON {
			emitJSONResult(jsonResult{
				Type:    "item",
				Status:  "error",
				URL:     url,
				ID:      video.ID,
				Title:   video.Title,
				Output:  result.outputPath,
				Bytes:   result.bytes,
				Retries: result.retried,
				Error:   err.Error(),
			})
		}
		return markReported(err)
	}
	okCount := 1
	skipped := 0
	if result.skipped {
		okCount = 0
		skipped = 1
	}
	printer.ItemResult(prefix, result, nil)
	if opts.JSON {
		status := "ok"
		if result.skipped {
			status = "skip"
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
			Skipped: result.skipped,
		})
	}
	printer.Summary(1, okCount, 0, skipped, result.bytes)
	return nil
}

func renderFormats(video *youtube.Video, opts Options, playlistID, playlistTitle string, index, total int) error {
	if opts.JSON {
		return renderFormatsJSON(video, playlistID, playlistTitle, index, total)
	}
	title := fmt.Sprintf(" Formats: %s ", video.Title)

	ctx := context.Background()
	selectedItag, tui, err := RunSeamlessFormatSelector(ctx, video, title, playlistID, playlistTitle, index, total)
	if err != nil {
		return err
	}
	if selectedItag == 0 || tui == nil {
		// User cancelled
		return nil
	}

	// User selected a format - transition to progress view and download
	opts.Itag = selectedItag
	opts.ListFormats = false

	tui.TransitionToProgress()
	defer tui.Stop()

	client := newClientTypeForType("android", opts)
	printer := NewSeamlessPrinter(opts, tui)

	ctxInfo := outputContext{}
	if playlistID != "" {
		ctxInfo.Index = index
		ctxInfo.Total = total
	}
	prefix := printer.Prefix(1, 1, video.Title)
	result, err := downloadVideo(ctx, client, video, opts, ctxInfo, printer, prefix)
	if err != nil {
		printer.ItemResult(prefix, result, err)
	}
	return err
}
