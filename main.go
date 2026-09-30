package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
	webserver "github.com/lvcoi/ytdl-go/internal/web"
)

func main() {
	// Ensure HTTP connections are properly closed at exit
	defer downloader.CloseIdleConnections()

	var opts downloader.Options
	var meta metaFlags
	var jobs int
	var web bool
	var webAddr string
	var serverHost string
	var serverPort int
	var updateSelf bool
	var printVersion bool

	flag.StringVar(&opts.OutputTemplate, "o", "{title}.{ext}", "output path or template (supports {title}, {artist}, {album}, {id}, {ext}, {quality}, {playlist_title}, {playlist_id}, {index}, {count})")
	flag.StringVar(&opts.OutputDir, "output-dir", "", "base output directory for security enforcement (prevents directory traversal)")
	flag.BoolVar(&opts.AudioOnly, "audio", false, "download best available audio only")
	flag.BoolVar(&opts.InfoOnly, "info", false, "print video metadata as JSON without downloading")
	flag.BoolVar(&opts.ListFormats, "list-formats", false, "list available formats and exit")
	flag.StringVar(&opts.Quality, "quality", "", "preferred quality (e.g. 1080p, 720p, 128k, best, worst)")
	flag.StringVar(&opts.Format, "format", "", "preferred container/extension (e.g. mp4, webm, m4a)")
	flag.IntVar(&opts.Itag, "itag", 0, "download specific format by itag number (use --list-formats to see available itags)")
	flag.Var(&meta, "meta", "metadata override key=value (repeatable)")
	flag.StringVar(&opts.ProgressLayout, "progress-layout", "", "progress layout template (e.g. \"{label} {percent} {current}/{total} {rate} {eta}\")")
	flag.IntVar(&opts.SegmentConcurrency, "segment-concurrency", 0, "parallel segment downloads (0=auto)")
	flag.IntVar(&opts.PlaylistConcurrency, "playlist-concurrency", 0, "parallel playlist entry downloads (0=auto)")
	flag.IntVar(&jobs, "jobs", 1, "number of concurrent downloads")
	flag.BoolVar(&opts.JSON, "json", false, "emit JSON output (suppresses human-readable progress)")
	flag.DurationVar(&opts.Timeout, "timeout", 3*time.Minute, "per-request timeout")
	flag.BoolVar(&opts.Quiet, "quiet", false, "suppress progress output (errors still shown)")
	flag.StringVar(&opts.LogLevel, "log-level", "info", "log level: debug, info, warn, error")
	flag.StringVar(&opts.CookiesFile, "cookies", "", "Netscape format cookies.txt file (for age-restricted/login content)")
	flag.StringVar(&opts.CookiesFromBrowser, "cookies-from-browser", "", "load cookies from browser: chrome, edge, brave, firefox (optionally ':profile')")
	flag.BoolVar(&opts.WriteThumbnail, "write-thumbnail", false, "write the thumbnail image next to the output file")
	flag.BoolVar(&opts.EmbedThumbnail, "embed-thumbnail", false, "embed the thumbnail as cover art (mp3/m4a/mp4/mkv)")
	flag.StringVar(&opts.AudioFormat, "audio-format", "", "target audio format for -audio: mp3, opus, m4a, aac, flac, wav, best (default best)")
	flag.StringVar(&opts.AudioQuality, "audio-quality", "", "audio quality for --audio-format (e.g. 160k, or 0-9 VBR)")
	flag.BoolVar(&opts.WriteSubs, "write-subs", false, "write subtitle files (.vtt)")
	flag.BoolVar(&opts.WriteAutoSubs, "write-auto-subs", false, "write auto-generated subtitle files")
	flag.StringVar(&opts.SubLangs, "sub-langs", "", "subtitle languages to download (comma separated, \"all\" for everything; default en)")
	flag.BoolVar(&opts.EmbedSubs, "embed-subs", false, "embed subtitles into the media container")
	flag.StringVar(&opts.SponsorblockMark, "sponsorblock-mark", "", "SponsorBlock categories to mark as chapters (e.g. sponsor,selfpromo)")
	flag.StringVar(&opts.SponsorblockRemove, "sponsorblock-remove", "", "SponsorBlock categories to cut out (e.g. sponsor,selfpromo,interaction)")
	flag.StringVar(&opts.PlaylistItems, "playlist-items", "", "playlist entries to download, e.g. 1:5,8,10:12 (1-based, ~ negates)")
	flag.StringVar(&opts.MatchFilter, "match-filter", "", "only download matching entries, e.g. duration<600&view_count>1000")
	flag.StringVar(&opts.DownloadArchive, "download-archive", "", "file recording downloaded video IDs; existing entries are skipped")
	flag.BoolVar(&opts.BreakOnExisting, "break-on-existing", false, "stop when an entry is already in the download archive")
	flag.StringVar(&opts.LimitRate, "limit-rate", "", "maximum download rate, e.g. 500K or 2M")
	flag.DurationVar(&opts.SleepInterval, "sleep-interval", 0, "sleep at least this long between playlist downloads")
	flag.DurationVar(&opts.MaxSleepInterval, "max-sleep-interval", 0, "sleep a random time up to this long between playlist downloads")
	flag.DurationVar(&opts.SleepRequests, "sleep-requests", 0, "sleep between YouTube API requests (429 rate-limit protection)")
	flag.StringVar(&opts.Proxy, "proxy", "", "use the specified HTTP/HTTPS/SOCKS5 proxy, e.g. socks5://127.0.0.1:1080")
	flag.BoolVar(&opts.Live, "live", false, "record a live stream from the HLS manifest (auto-detected when the video is live)")
	flag.BoolVar(&updateSelf, "update", false, "check GitHub for a newer release and self-update")
	flag.BoolVar(&printVersion, "version", false, "print the version and exit")
	flag.BoolVar(&web, "web", false, "launch the web UI server")
	flag.StringVar(&webAddr, "web-addr", "", "web server address (overrides host/port)")
	flag.StringVar(&serverHost, "host", "0.0.0.0", "web server host")
	flag.IntVar(&serverPort, "port", 8888, "web server port")
	flag.Parse()

	if webAddr == "" {
		webAddr = fmt.Sprintf("%s:%d", serverHost, serverPort)
	}

	opts.MetaOverrides = meta.Values()

	if printVersion {
		fmt.Printf("ytdl-go %s\n", downloader.AppVersion)
		return
	}

	// Configure process-wide proxy before any network activity.
	if opts.Proxy != "" {
		if err := downloader.ConfigureProxy(opts.Proxy); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(downloader.ExitCode(err))
		}
	}

	// Self-update: check GitHub releases and replace the binary if newer.
	if updateSelf {
		fmt.Printf("ytdl-go %s: checking for updates...\n", downloader.AppVersion)
		message, err := downloader.SelfUpdate(context.Background(), 5*time.Minute)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update failed: %v\n", err)
			os.Exit(downloader.ExitCode(err))
		}
		fmt.Println(message)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if web {
		if err := webserver.ListenAndServe(ctx, webAddr, jobs); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "web server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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

	resultsList, exitCode := app.Run(ctx, urls, opts, jobs)
	for _, res := range resultsList {
		if res.Err != nil {
			if opts.JSON {
				if downloader.IsReported(res.Err) {
					continue
				}
				writeJSONError(res.URL, res.Err)
			} else if !downloader.IsReported(res.Err) {
				fmt.Fprintf(os.Stderr, "error: %v\n", res.Err)
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
