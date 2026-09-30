package downloader

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lvcoi/ytdl-lib/v2"
)

var fetchMusicPlaylistEntriesFn = fetchMusicPlaylistEntries

func processPlaylist(ctx context.Context, url string, opts Options, printer *Printer, isMusicURL bool) error {
	playlistClient := newClientTypeForType("web", opts)
	playlist, err := playlistClient.GetPlaylistContext(ctx, url)
	if err != nil {
		return wrapAccessError(fmt.Errorf("fetching playlist: %w", err))
	}

	// Fetch playlist title from YouTube Music (the library often returns empty/generic titles)
	// Do this before ListFormats/InfoOnly so they get the proper title
	if isMusicURL {
		if title, err := fetchMusicPlaylistTitle(ctx, playlist.ID, opts.Timeout); err == nil && title != "" {
			playlist.Title = title
		}
	}
	if playlist.Title == "" {
		playlist.Title = "Playlist"
	}

	// Check InfoOnly first, then ListFormats (consistent with single-video path)
	if opts.InfoOnly {
		return printPlaylistInfo(playlist)
	}
	if opts.ListFormats {
		return listPlaylistFormats(ctx, playlist, opts, printer)
	}

	// Only check for empty videos when actually downloading
	if len(playlist.Videos) == 0 {
		return wrapCategory(CategoryUnsupported, errors.New("playlist has no videos"))
	}

	// --playlist-items selection
	itemSelector, err := parsePlaylistItems(opts.PlaylistItems)
	if err != nil {
		return wrapCategory(CategoryInvalidURL, err)
	}
	if itemSelector != nil {
		filtered := make([]*youtube.PlaylistEntry, 0, len(playlist.Videos))
		for i, entry := range playlist.Videos {
			if itemSelector.includes(i + 1) {
				filtered = append(filtered, entry)
			}
		}
		printer.Log(LogInfo, fmt.Sprintf("playlist-items: %d of %d entries selected", len(filtered), len(playlist.Videos)))
		playlist.Videos = filtered
		if len(filtered) == 0 {
			return wrapCategory(CategoryUnsupported, errors.New("no playlist entries match --playlist-items"))
		}
	}

	// --match-filter
	filter, err := parseMatchFilter(opts.MatchFilter)
	if err != nil {
		return wrapCategory(CategoryInvalidURL, err)
	}

	// --download-archive session
	var archive *downloadArchive
	if opts.DownloadArchive != "" {
		archive, err = newDownloadArchive(opts.DownloadArchive)
		if err != nil {
			return err
		}
		opts.ArchiveSession = archive
	}

	albumMeta := resolveMusicPlaylistAlbumMeta(ctx, playlist.ID, opts, isMusicURL, printer)

	printer.Log(LogInfo, fmt.Sprintf("playlist: %s (%d videos)", playlist.Title, len(playlist.Videos)))

	videoClient := newClientTypeForType("android", opts)
	type playlistOutcome struct {
		ok      bool
		failed  bool
		skipped bool
		bytes   int64
		broken  bool
	}

	applyMatchFilter := func(i int, entry *youtube.PlaylistEntry, video *youtube.Video) bool {
		if filter == nil {
			return true
		}
		ok, reason := filter.matches(filterVideoInfo{
			ID:        video.ID,
			Title:     video.Title,
			Author:    video.Author,
			Duration:  video.Duration.Seconds(),
			ViewCount: video.Views,
		})
		if !ok {
			printer.ItemSkipped(printer.Prefix(i+1, len(playlist.Videos), entryTitle(entry)), reason)
		}
		return ok
	}

	applyArchiveSkip := func(i int, entry *youtube.PlaylistEntry) bool {
		if archive == nil || entry == nil || entry.ID == "" {
			return false
		}
		if archive.contains(entry.ID) {
			prefix := printer.Prefix(i+1, len(playlist.Videos), entryTitle(entry))
			if opts.BreakOnExisting {
				printer.ItemResult(prefix, downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("entry %s already in download archive (--break-on-existing)", entry.ID)))
				return true // treat as "break"
			}
			printer.ItemSkipped(prefix, "already in download archive")
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "skip",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            entry.ID,
					Title:         entryTitle(entry),
					Error:         "already in download archive",
				})
			}
			return true
		}
		return false
	}

	handleEntry := func(i int, entry *youtube.PlaylistEntry) playlistOutcome {
		total := len(playlist.Videos)
		prefix := printer.Prefix(i+1, total, entryTitle(entry))
		if entry == nil || entry.ID == "" {
			printer.ItemSkipped(prefix, "missing playlist entry")
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "skip",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            "",
					Title:         "",
					Error:         "missing playlist entry",
				})
			}
			return playlistOutcome{skipped: true}
		}

		video, err := videoClient.VideoFromPlaylistEntryContext(ctx, entry)
		if err != nil {
			err = wrapFetchError(err, "fetching video metadata")
			printer.ItemResult(prefix, downloadResult{}, err)
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "error",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            entry.ID,
					Title:         entryTitle(entry),
					Error:         err.Error(),
				})
			}
			return playlistOutcome{failed: true}
		}

		if applyArchiveSkip(i, entry) {
			if opts.BreakOnExisting && archive.contains(entry.ID) {
				return playlistOutcome{skipped: true, broken: true}
			}
			return playlistOutcome{skipped: true}
		}
		if !applyMatchFilter(i, entry, video) {
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "skip",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            entry.ID,
					Title:         entryTitle(entry),
					Error:         "filtered",
				})
			}
			return playlistOutcome{skipped: true}
		}

		meta := albumMeta[entry.ID]
		entryTitle := entry.Title
		entryAuthor := entry.Author
		if meta.Title != "" {
			entryTitle = meta.Title
		}
		if meta.Artist != "" {
			entryAuthor = meta.Artist
		}

		result, err := downloadVideo(ctx, videoClient, video, opts, outputContext{
			Playlist:      playlist,
			Index:         i + 1,
			Total:         total,
			EntryTitle:    entryTitle,
			EntryAuthor:   entryAuthor,
			EntryAlbum:    meta.Album,
			SourceURL:     watchURLForID(entry.ID),
			PlaylistURL:   url,
			MetaOverrides: opts.MetaOverrides,
		}, printer, prefix)
		if result.skipped {
			printer.ItemSkipped(prefix, "exists")
			if opts.JSON {
				emitJSONResult(jsonResult{
					Type:          "item",
					Status:        "skip",
					PlaylistID:    playlist.ID,
					PlaylistTitle: playlist.Title,
					Index:         i + 1,
					ID:            entry.ID,
					Title:         entryTitle,
					Output:        result.outputPath,
					Bytes:         result.bytes,
					Retries:       result.retried,
					Error:         "exists",
				})
			}
			return playlistOutcome{skipped: true}
		}
		printer.ItemResult(prefix, result, err)
		if opts.JSON {
			status := "ok"
			errMsg := ""
			if err != nil {
				status = "error"
				errMsg = err.Error()
			}
			emitJSONResult(jsonResult{
				Type:          "item",
				Status:        status,
				PlaylistID:    playlist.ID,
				PlaylistTitle: playlist.Title,
				Index:         i + 1,
				ID:            entry.ID,
				Title:         entryTitle,
				Output:        result.outputPath,
				Bytes:         result.bytes,
				Retries:       result.retried,
				Error:         errMsg,
			})
		}

		if err != nil {
			return playlistOutcome{failed: true}
		}

		return playlistOutcome{ok: true, bytes: result.bytes}
	}

	successes := 0
	failures := 0
	skipped := 0
	var totalBytes int64

	// Always download sequentially to:
	// 1. Avoid bandwidth contention between concurrent downloads
	// 2. Properly clean up connections after each download
	// 3. Prevent zombie processes from accumulating
	for i, entry := range playlist.Videos {
		if i > 0 && opts.SleepInterval > 0 {
			if sleepErr := opts.sleepBeforePlaylistEntry(ctx); sleepErr != nil {
				break
			}
		}
		outcome := handleEntry(i, entry)
		if outcome.skipped {
			skipped++
			if outcome.broken {
				printer.Log(LogInfo, "stopping playlist: --break-on-existing triggered")
				break
			}
			continue
		}
		if outcome.failed {
			failures++
			continue
		}
		if outcome.ok {
			successes++
			totalBytes += outcome.bytes
		}
	}

	printer.Summary(len(playlist.Videos), successes, failures, skipped, totalBytes)
	if successes == 0 {
		return markReported(wrapCategory(CategoryUnsupported, errors.New("no playlist entries downloaded successfully")))
	}
	return nil
}

func resolveMusicPlaylistAlbumMeta(ctx context.Context, playlistID string, opts Options, isMusicURL bool, printer *Printer) map[string]musicEntryMeta {
	if !isMusicURL {
		return map[string]musicEntryMeta{}
	}

	albumMeta, err := fetchMusicPlaylistEntriesFn(ctx, playlistID, opts)
	if err != nil {
		if printer != nil {
			printer.Log(LogWarn, fmt.Sprintf("warning: album metadata enrichment unavailable for playlist %s: %v", playlistID, err))
		}
		return map[string]musicEntryMeta{}
	}
	return albumMeta
}

func listPlaylistFormats(ctx context.Context, playlist *youtube.Playlist, opts Options, _ *Printer) error {
	if len(playlist.Videos) == 0 {
		return nil
	}

	client := newClientTypeForType("android", opts)

	for i, entry := range playlist.Videos {
		if entry == nil || entry.ID == "" {
			continue
		}
		video, err := client.VideoFromPlaylistEntryContext(ctx, entry)
		if err != nil {
			return wrapFetchError(err, "fetching video metadata")
		}

		if err := renderFormats(video, opts, playlist.ID, playlist.Title, i+1, len(playlist.Videos)); err != nil {
			return err
		}
	}
	return nil
}

func entryTitle(entry *youtube.PlaylistEntry) string {
	if entry == nil {
		return ""
	}
	if entry.Title != "" {
		return entry.Title
	}
	return entry.ID
}

func wrapAccessError(err error) error {
	if err == nil {
		return nil
	}
	if isRestrictedAccess(err) {
		return fmt.Errorf("restricted access: %w", err)
	}
	return err
}

func isRestrictedAccess(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	restrictedMarkers := []string{
		"private",
		"sign in",
		"login",
		"members only",
		"premium",
		"copyright",
		"video unavailable",
		"content unavailable",
		"age-restricted",
		"age restricted",
		"not available",
	}
	for _, marker := range restrictedMarkers {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}
