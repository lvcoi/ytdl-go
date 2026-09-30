package downloader

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lvcoi/ytdl-lib/v2"
)

// subtitleTrack pairs a caption track with its resolved language label.
type subtitleTrack struct {
	Lang  string
	Name  string
	Auto  bool // auto-generated (ASR)
	Track youtube.CaptionTrack
}

// selectSubtitleTracks filters the video's caption tracks according to
// --sub-langs, --write-subs and --write-auto-subs.
func selectSubtitleTracks(video *youtube.Video, opts Options) []subtitleTrack {
	langs := strings.TrimSpace(opts.SubLangs)
	if langs == "" {
		langs = "en"
	}
	wantAll := strings.EqualFold(langs, "all")
	wanted := map[string]struct{}{}
	for _, l := range strings.Split(langs, ",") {
		l = strings.TrimSpace(l)
		if l != "" {
			wanted[strings.ToLower(l)] = struct{}{}
		}
	}

	var tracks []subtitleTrack
	for _, ct := range video.CaptionTracks {
		auto := strings.EqualFold(ct.Kind, "asr")
		if auto && !opts.WriteAutoSubs && opts.WriteSubs {
			continue
		}
		if !auto && !opts.WriteSubs && !opts.WriteAutoSubs {
			// neither requested: skip entirely
			continue
		}
		lang := strings.ToLower(ct.LanguageCode)
		if !wantAll {
			if _, ok := wanted[lang]; !ok {
				// allow base-language matches ("en" picks "en-US")
				base := strings.SplitN(lang, "-", 2)[0]
				if _, ok := wanted[base]; !ok {
					continue
				}
			}
		}
		name := ct.Name.SimpleText
		if name == "" {
			name = ct.LanguageCode
		}
		tracks = append(tracks, subtitleTrack{
			Lang:  ct.LanguageCode,
			Name:  name,
			Auto:  auto,
			Track: ct,
		})
	}
	sort.Slice(tracks, func(i, j int) bool {
		if tracks[i].Auto != tracks[j].Auto {
			return !tracks[i].Auto // manual first
		}
		return tracks[i].Lang < tracks[j].Lang
	})
	return tracks
}

// fetchSubtitleContent downloads a caption track as WebVTT. The "&fmt=vtt"
// parameter makes YouTube's timedtext endpoint return VTT for both manual and
// auto-generated tracks.
func fetchSubtitleContent(ctx context.Context, track subtitleTrack, client HTTPDoer) (string, error) {
	base, err := url.Parse(track.Track.BaseURL)
	if err != nil || base.Host == "" {
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid caption track URL"))
	}
	q := base.Query()
	if q.Get("fmt") == "" {
		q.Set("fmt", "vtt")
	}
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return "", wrapCategory(CategoryNetwork, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading subtitles: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading subtitles: unexpected status %d", resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024*1024))
	if err != nil {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading subtitles: %w", err))
	}
	content := string(body)
	if strings.Contains(content, "<transcript>") || strings.Contains(content, "<timedtext") {
		// Endpoint ignored fmt=vtt; fall back to XML conversion.
		vtt, convErr := timedtextXMLToVTT(body)
		if convErr == nil {
			content = vtt
		}
	}
	return content, nil
}

type timedtextDoc struct {
	Body []struct {
		P []struct {
			Text string `xml:",chardata"`
			T    int64  `xml:"t,attr"`
			D    int64  `xml:"d,attr"`
		} `xml:"p"`
	} `xml:"body"`
}

// timedtextXMLToVTT converts YouTube's timedtext XML to WebVTT.
func timedtextXMLToVTT(data []byte) (string, error) {
	var doc timedtextDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("parsing timedtext XML: %w", err)
	}
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	idx := 0
	for _, p := range doc.Body {
		for _, seg := range p.P {
			idx++
			text := strings.TrimSpace(seg.Text)
			if text == "" {
				continue
			}
			start := vttTimestamp(seg.T)
			end := vttTimestamp(seg.T + seg.D)
			fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n", idx, start, end, text)
		}
	}
	return b.String(), nil
}

func vttTimestamp(ms int64) string {
	hours := ms / 3_600_000
	minutes := (ms % 3_600_000) / 60_000
	seconds := (ms % 60_000) / 1000
	millis := ms % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}

// subtitleArtifactPath returns the artifact path for a subtitle track file.
func subtitleArtifactPath(outputPath, baseDir, lang string) (string, error) {
	lang = strings.ReplaceAll(lang, "/", "-")
	return artifactPath(outputPath, "."+lang+".vtt", baseDir)
}

// writeSubtitleTracks downloads and writes the selected caption tracks.
func writeSubtitleTracks(ctx context.Context, video *youtube.Video, outputPath, baseDir string, opts Options, client HTTPDoer, printer *Printer) ([]string, error) {
	tracks := selectSubtitleTracks(video, opts)
	if len(tracks) == 0 {
		if opts.SubLangs == "" {
			printer.Log(LogWarn, "no subtitle tracks available")
		}
		return nil, nil
	}
	var written []string
	for _, track := range tracks {
		content, err := fetchSubtitleContent(ctx, track, client)
		if err != nil {
			printer.Log(LogWarn, fmt.Sprintf("warning: subtitles for %s unavailable: %v", track.Lang, err))
			continue
		}
		path, err := subtitleArtifactPath(outputPath, baseDir, track.Lang)
		if err != nil {
			return written, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return written, wrapCategory(CategoryFilesystem, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return written, wrapCategory(CategoryFilesystem, fmt.Errorf("writing subtitles: %w", err))
		}
		written = append(written, path)
		printer.Log(LogInfo, fmt.Sprintf("subtitles: %s (%s) -> %s", track.Lang, track.Name, filepath.Base(path)))
	}
	return written, nil
}

// embedSubtitleTrack embeds a WebVTT subtitle into the media container using
// ffmpeg (mov_text for mp4/m4a, webvtt for webm, srt for mkv).
func embedSubtitleTrack(outputPath, subPath string) error {
	ext := strings.ToLower(filepath.Ext(outputPath))
	dir := filepath.Dir(outputPath)
	tmpFile := filepath.Join(dir, ".tmp_subs_"+filepath.Base(outputPath))

	subCodec := "mov_text"
	switch ext {
	case ".webm":
		subCodec = "webvtt"
	case ".mkv", ".opus", ".ogg":
		subCodec = "srt"
	}

	args := []string{
		"-y",
		"-i", outputPath,
		"-i", subPath,
		"-map", "0",
		"-map", "1:0",
		"-c:v", "copy",
		"-c:a", "copy",
		"-c:s", subCodec,
		"-metadata:s:s:0", "language=" + subtitleLanguageCode(subPath),
	}
	if ext == ".mp4" || ext == ".m4a" || ext == ".m4v" || ext == ".mov" {
		args = append(args, "-disposition:s:0", "subtitles")
	}
	args = append(args, tmpFile)

	cmd := ffmpegCommand(args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpFile)
		stderr := strings.TrimSpace(string(output))
		if stderr != "" {
			return fmt.Errorf("failed to embed subtitles: %s: %w", stderr, err)
		}
		return fmt.Errorf("failed to embed subtitles: %w", err)
	}
	if err := os.Rename(tmpFile, outputPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace original file with subtitled version: %w", err)
	}
	return nil
}

// subtitleLanguageCode extracts a BCP-47-ish language code from a file named
// "<output>.<lang>.vtt" (defaults to "und").
func subtitleLanguageCode(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base)) // drop .vtt
	idx := strings.LastIndex(base, ".")
	if idx <= 0 {
		return "und"
	}
	lang := base[idx+1:]
	if len(lang) < 2 || len(lang) > 10 || strings.Contains(lang, "_") {
		return "und"
	}
	for _, r := range lang {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '-' || r >= '0' && r <= '9') {
			return "und"
		}
	}
	return lang
}
