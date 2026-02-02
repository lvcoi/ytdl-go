package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type oEmbedResponse struct {
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
}

func fetchPageMetadata(ctx context.Context, pageURL string, timeout time.Duration) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", "", err
	}
	client := newHTTPClient(timeout)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", "", err
	}

	content := string(body)
	title := stringsOrFallback(
		findMeta(content, "og:title"),
		findMeta(content, "twitter:title"),
		findTitleTag(content),
	)
	author := stringsOrFallback(
		findMeta(content, "author"),
		findMeta(content, "og:site_name"),
	)

	if oembedURL := findOEmbedURL(content); oembedURL != "" {
		if resolved := resolveManifestURL(pageURL, oembedURL); resolved != "" {
			oembed, err := fetchOEmbed(ctx, resolved, timeout)
			if err == nil {
				if title == "" {
					title = oembed.Title
				}
				if author == "" {
					author = oembed.AuthorName
				}
			}
		}
	}

	return strings.TrimSpace(title), strings.TrimSpace(author), nil
}

func fetchOEmbed(ctx context.Context, oembedURL string, timeout time.Duration) (oEmbedResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, oembedURL, nil)
	if err != nil {
		return oEmbedResponse{}, err
	}
	client := newHTTPClient(timeout)
	resp, err := client.Do(req)
	if err != nil {
		return oEmbedResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oEmbedResponse{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var payload oEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return oEmbedResponse{}, err
	}
	return payload, nil
}

func findTitleTag(html string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		return htmlUnescape(strings.TrimSpace(match[1]))
	}
	return ""
}

func findMeta(html, key string) string {
	pattern := fmt.Sprintf(`(?is)<meta[^>]+(?:property|name)=["']%s["'][^>]+content=["'](.*?)["'][^>]*>`, regexp.QuoteMeta(key))
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		return htmlUnescape(strings.TrimSpace(match[1]))
	}
	return ""
}

func findOEmbedURL(html string) string {
	re := regexp.MustCompile(`(?is)<link[^>]+rel=["']alternate["'][^>]+type=["']application/json\+oembed["'][^>]+href=["']([^"']+)["'][^>]*>`)
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func htmlUnescape(value string) string {
	replacer := strings.NewReplacer("&amp;", "&", "&quot;", "\"", "&#39;", "'", "&lt;", "<", "&gt;", ">")
	return replacer.Replace(value)
}
