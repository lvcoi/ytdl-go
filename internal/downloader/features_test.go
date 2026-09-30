package downloader

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lvcoi/ytdl-lib/v2"
)

func TestParseRateLimit(t *testing.T) {
	cases := []struct {
		input    string
		expected float64
		wantErr  bool
	}{
		{"500K", 500_000, false},
		{"2M", 2_000_000, false},
		{"1G", 1_000_000_000, false},
		{"1500000", 1_500_000, false},
		{"512KiB", 512 * 1024, false},
		{"100b", 100, false},
		{"", 0, true},
		{"K", 0, true},
		{"-5M", 0, true},
		{"10X", 0, true},
	}
	for _, tc := range cases {
		got, err := parseRateLimit(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseRateLimit(%q): expected error, got %v", tc.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseRateLimit(%q): unexpected error %v", tc.input, err)
			continue
		}
		if got != tc.expected {
			t.Errorf("parseRateLimit(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestRateLimiterThrottles(t *testing.T) {
	limiter := newRateLimiter(1000) // 1000 bytes/sec
	if limiter == nil {
		t.Fatal("expected limiter")
	}
	// Manual clock: sleeps advance the fake clock instead of real time.
	base := time.Now()
	limiter.clock = func() time.Time { return base }
	limiter.sleep = func(d time.Duration) { base = base.Add(d) }

	// First burst allowed immediately (bucket full).
	if err := limiter.wait(t.Context(), 1000); err != nil {
		t.Fatalf("first wait failed: %v", err)
	}
	// Bucket empty: waiting for 2000 bytes requires ~2 more seconds of tokens.
	if err := limiter.wait(t.Context(), 2000); err != nil {
		t.Fatalf("second wait failed: %v", err)
	}
	if base.Sub(time.Now()) > 10*time.Second {
		t.Errorf("clock advanced too far: %v", base.Sub(time.Now()))
	}
}

func TestParsePlaylistItems(t *testing.T) {
	cases := []struct {
		expr     string
		includes map[int]bool
		wantErr  bool
	}{
		{"", map[int]bool{1: true, 2: true}, false},
		{"3", map[int]bool{1: false, 2: false, 3: true, 4: false}, false},
		{"1:3", map[int]bool{1: true, 2: true, 3: true, 4: false}, false},
		{"2:", map[int]bool{1: false, 2: true, 5: true}, false},
		{":3", map[int]bool{1: true, 3: true, 4: false}, false},
		{"1:5,8,10:12", map[int]bool{1: true, 5: true, 6: false, 8: true, 10: true, 12: true, 13: false}, false},
		{"~2", map[int]bool{1: true, 2: false, 3: true}, false},
		{"0", nil, true},
		{"a", nil, true},
		{"1:x", nil, true},
	}
	for _, tc := range cases {
		sel, err := parsePlaylistItems(tc.expr)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parsePlaylistItems(%q): expected error", tc.expr)
			}
			continue
		}
		if err != nil {
			t.Errorf("parsePlaylistItems(%q): unexpected error %v", tc.expr, err)
			continue
		}
		for index, want := range tc.includes {
			if got := sel.includes(index); got != want {
				t.Errorf("parsePlaylistItems(%q).includes(%d) = %v, want %v", tc.expr, index, got, want)
			}
		}
	}
}

func TestParseMatchFilter(t *testing.T) {
	f, err := parseMatchFilter("duration<600&view_count>1000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(f.conditions))
	}

	info := filterVideoInfo{ID: "abc", Title: "Hello World", Author: "Channel", Duration: 300, ViewCount: 5000}
	if ok, reason := f.matches(info); !ok {
		t.Errorf("expected match, got rejection: %s", reason)
	}
	long := info
	long.Duration = 700
	if ok, _ := f.matches(long); ok {
		t.Error("expected rejection for duration>600")
	}

	f2, err := parseMatchFilter("title*=hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok, _ := f2.matches(info); !ok {
		t.Error("expected contains match")
	}

	if _, err := parseMatchFilter("nosuchoperator"); err == nil {
		t.Error("expected error for malformed condition")
	}
}

func TestDownloadArchiveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "archive.txt")

	archive, err := newDownloadArchive(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archive.contains("abc123") {
		t.Error("fresh archive should be empty")
	}
	if err := archive.record("abc123"); err != nil {
		t.Fatalf("record failed: %v", err)
	}
	if err := archive.record("abc123"); err != nil {
		t.Fatalf("duplicate record failed: %v", err)
	}
	if !archive.contains("abc123") {
		t.Error("recorded ID should be present")
	}

	// Reload from disk: yt-dlp style line format should be honored.
	if err := os.WriteFile(path, []byte("youtube abc123\nyoutube def456\n# comment\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reloaded, err := newDownloadArchive(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	for _, id := range []string{"abc123", "def456"} {
		if !reloaded.contains(id) {
			t.Errorf("reloaded archive missing %s", id)
		}
	}
}

func TestParseNetscapeCookiesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.txt")
	content := "# Netscape HTTP Cookie File\n" +
		".youtube.com\tTRUE\t/\tTRUE\t1893456000\tSID\tsomevalue\n" +
		"#HttpOnly_.youtube.com\tTRUE\t/\tTRUE\t1893456000\tHSID\tvalue2\n" +
		"# comment line\n" +
		"broken line without tabs\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cookies, err := parseNetscapeCookiesFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
	if cookies[0].Name != "SID" || cookies[0].Value != "somevalue" {
		t.Errorf("unexpected first cookie: %+v", cookies[0])
	}
	if cookies[0].Domain != ".youtube.com" || !cookies[0].IncludeSubdomains || !cookies[0].Secure {
		t.Errorf("unexpected cookie attributes: %+v", cookies[0])
	}
	if !cookies[1].matchesURL(mustParseURL(t, "https://www.youtube.com/watch")) {
		t.Error("HttpOnly cookie should match youtube URL")
	}
}

func TestStaticCookieJar(t *testing.T) {
	static := []cookieEntry{
		{Domain: ".youtube.com", IncludeSubdomains: true, Path: "/", Name: "SID", Value: "v1", Expires: time.Now().Add(time.Hour)},
		{Domain: ".google.com", IncludeSubdomains: true, Path: "/", Name: "LSID", Value: "v2", Expires: time.Now().Add(time.Hour)},
		{Domain: ".expired.com", Path: "/", Name: "OLD", Value: "v3", Expires: time.Now().Add(-time.Hour)},
	}
	jar := &staticCookieJar{static: static}
	got := jar.Cookies(mustParseURL(t, "https://www.youtube.com/watch?v=x"))
	if len(got) != 1 || got[0].Name != "SID" {
		t.Errorf("expected only SID cookie, got %+v", got)
	}
	if got := jar.Cookies(mustParseURL(t, "https://accounts.google.com/o/oauth2")); len(got) != 1 || got[0].Name != "LSID" {
		t.Errorf("expected LSID cookie, got %+v", got)
	}
}

func TestNormalizeAudioFormat(t *testing.T) {
	if got, err := normalizeAudioFormat("MP3"); err != nil || got != "mp3" {
		t.Errorf("normalizeAudioFormat(MP3) = %q, %v", got, err)
	}
	if got, err := normalizeAudioFormat("aac"); err != nil || got != "m4a" {
		t.Errorf("normalizeAudioFormat(aac) = %q, %v", got, err)
	}
	if got, err := normalizeAudioFormat("best"); err != nil || got != "" {
		t.Errorf("normalizeAudioFormat(best) = %q, %v", got, err)
	}
	if _, err := normalizeAudioFormat("wma"); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		current, candidate string
		want               bool
	}{
		{"0.2.0-beta", "0.2.0", false},
		{"0.2.0-beta", "v0.2.1-beta", true},
		{"0.2.0-beta", "0.1.9", false},
		{"0.2.0-beta", "1.0.0", true},
		{"0.2.0-beta", "not-semver", false},
	}
	for _, tc := range cases {
		if got := isNewerVersion(tc.current, tc.candidate); got != tc.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tc.current, tc.candidate, got, tc.want)
		}
	}
}

func TestTimedtextXMLToVTT(t *testing.T) {
	xmlData := []byte(`<timedtext><body><p t="1000" d="2000">Hello</p><p t="4000" d="1500">World</p></body></timedtext>`)
	vtt, err := timedtextXMLToVTT(xmlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(vtt, "WEBVTT") || !contains(vtt, "00:00:01.000 --> 00:00:03.000") || !contains(vtt, "Hello") {
		t.Errorf("unexpected VTT output:\n%s", vtt)
	}
}

func TestVCodecAndACodecFromMime(t *testing.T) {
	format := &youtube.Format{MimeType: "video/mp4; codecs=\"avc1.640028, mp4a.40.2\""}
	if got := formatVideoCodec(format); got != "avc1" {
		t.Errorf("formatVideoCodec = %q, want avc1", got)
	}
	if got := formatAudioCodec(format); got != "mp4a" {
		t.Errorf("formatAudioCodec = %q, want mp4a", got)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return u
}
