package downloader

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	playlistIDRegex  = regexp.MustCompile(`^[A-Za-z0-9_-]{13,42}$`)
	playlistURLRegex = regexp.MustCompile(`[?&]list=([A-Za-z0-9_-]{13,42})`)
)

func validateInputURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid URL: %w", err))
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid URL: missing scheme or host"))
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return "", wrapCategory(CategoryInvalidURL, fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme))
	}
	return parsed.String(), nil
}

func looksLikePlaylist(url string) bool {
	return playlistIDRegex.MatchString(url) || playlistURLRegex.MatchString(url)
}

func isYouTubeURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	host = strings.TrimPrefix(host, "www.")
	return host == "youtube.com" || host == "youtu.be" || host == "music.youtube.com"
}

// normalizeHostname returns the normalized hostname from a URL:
// lowercase, with "www." prefix removed, and port stripped.
func normalizeHostname(parsed *url.URL) string {
	host := strings.ToLower(parsed.Hostname())
	return strings.TrimPrefix(host, "www.")
}

// ConvertMusicURL converts YouTube Music URLs to regular YouTube URLs
func ConvertMusicURL(u string) string {
	// Parse the URL
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}

	// If it's not a music.youtube.com URL, return as-is
	if normalizeHostname(parsed) != "music.youtube.com" {
		return u
	}

	// Replace the host with www.youtube.com (do not preserve port)
	// so that downstream YouTube URL checks that rely on Host match as expected.
	parsed.Host = "www.youtube.com"

	// Remove any music-specific parameters
	query := parsed.Query()
	delete(query, "si")
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

// NormalizeYouTubeURL converts alternate YouTube URL forms (live/shorts/youtu.be) to watch?v=.
func NormalizeYouTubeURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	host := normalizeHostname(parsed)
	if host != "youtube.com" && host != "youtu.be" {
		return u
	}
	query := parsed.Query()
	if host == "youtu.be" {
		id := strings.TrimPrefix(parsed.Path, "/")
		if id != "" {
			query.Set("v", id)
			parsed.Path = "/watch"
			parsed.RawQuery = query.Encode()
		}
		return parsed.String()
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) >= 2 && (parts[0] == "live" || parts[0] == "shorts") {
		if query.Get("v") == "" && parts[1] != "" {
			query.Set("v", parts[1])
		}
		parsed.Path = "/watch"
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	return u
}

// isMusicYouTubeURL checks if a URL is a YouTube Music URL by parsing and normalizing the hostname
func isMusicYouTubeURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}

	return normalizeHostname(parsed) == "music.youtube.com"
}

func watchURLForID(id string) string {
	if id == "" {
		return ""
	}
	return "https://www.youtube.com/watch?v=" + id
}
