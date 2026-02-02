package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type musicEntryMeta struct {
	Title  string
	Artist string
	Album  string
}

type musicConfig struct {
	apiKey  string
	context map[string]any
}

func fetchMusicPlaylistEntries(ctx context.Context, playlistID string, opts Options) (map[string]musicEntryMeta, error) {
	cfg, err := fetchMusicConfig(ctx, playlistID, opts.Timeout)
	if err != nil {
		return nil, err
	}

	browseID := playlistBrowseID(playlistID)
	payload := map[string]any{
		"context":  cfg.context,
		"browseId": browseID,
	}

	result := map[string]musicEntryMeta{}
	response, err := fetchMusicBrowse(ctx, cfg.apiKey, payload, opts.Timeout)
	if err != nil {
		return nil, err
	}

	items, continuation, err := extractMusicPlaylistItems(response)
	if err != nil {
		return nil, err
	}
	appendMusicEntries(result, items)

	for continuation != "" {
		payload = map[string]any{
			"context":      cfg.context,
			"continuation": continuation,
		}
		response, err = fetchMusicBrowse(ctx, cfg.apiKey, payload, opts.Timeout)
		if err != nil {
			return nil, err
		}
		items, continuation, err = extractMusicPlaylistItems(response)
		if err != nil {
			return nil, err
		}
		appendMusicEntries(result, items)
	}

	if len(result) == 0 {
		return nil, errors.New("no playlist entries found in YouTube Music response")
	}
	return result, nil
}

func fetchMusicConfig(ctx context.Context, playlistID string, timeout time.Duration) (musicConfig, error) {
	playlistURL := "https://music.youtube.com/playlist?list=" + url.QueryEscape(playlistID)
	client := newHTTPClient(timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
	if err != nil {
		return musicConfig{}, err
	}
	req.Header.Set("User-Agent", musicUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return musicConfig{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return musicConfig{}, fmt.Errorf("unexpected response %d from YouTube Music", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return musicConfig{}, err
	}

	re := regexp.MustCompile(`(?s)ytcfg\.set\((\{.*?\})\);`)
	match := re.FindSubmatch(body)
	if match == nil {
		return musicConfig{}, errors.New("ytcfg.set data not found in YouTube Music page")
	}

	var cfg struct {
		APIKey  string         `json:"INNERTUBE_API_KEY"`
		Context map[string]any `json:"INNERTUBE_CONTEXT"`
	}
	if err := json.Unmarshal(match[1], &cfg); err != nil {
		return musicConfig{}, err
	}
	if cfg.APIKey == "" || len(cfg.Context) == 0 {
		return musicConfig{}, errors.New("missing innertube config in YouTube Music page")
	}

	return musicConfig{
		apiKey:  cfg.APIKey,
		context: cfg.Context,
	}, nil
}

func fetchMusicPlaylistTitle(ctx context.Context, playlistID string, timeout time.Duration) (string, error) {
	// Fetch the HTML page and extract og:title meta tag
	playlistURL := "https://music.youtube.com/playlist?list=" + url.QueryEscape(playlistID)
	client := newHTTPClient(timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", musicUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response %d from YouTube Music", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try og:title meta tag first (most reliable) - handle attribute order variations
	ogTitleRe := regexp.MustCompile(`(?i)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	if match := ogTitleRe.FindSubmatch(body); match != nil {
		return string(match[1]), nil
	}
	// Try reverse order (content before property)
	ogTitleRe2 := regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:title["']`)
	if match := ogTitleRe2.FindSubmatch(body); match != nil {
		return string(match[1]), nil
	}

	// Try <title> tag as fallback
	titleRe := regexp.MustCompile(`<title>([^<]+)</title>`)
	if match := titleRe.FindSubmatch(body); match != nil {
		title := strings.TrimSuffix(string(match[1]), " - YouTube Music")
		title = strings.TrimSuffix(title, " - YouTube")
		if title != "" {
			return title, nil
		}
	}

	return "", errors.New("playlist title not found in YouTube Music page")
}

func fetchMusicBrowse(ctx context.Context, apiKey string, payload map[string]any, timeout time.Duration) (map[string]any, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := newHTTPClient(timeout)
	endpoint := "https://music.youtube.com/youtubei/v1/browse?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", musicUserAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response %d from YouTube Music browse", resp.StatusCode)
	}

	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func playlistBrowseID(playlistID string) string {
	if strings.HasPrefix(playlistID, "VL") {
		return playlistID
	}
	return "VL" + playlistID
}

func extractMusicPlaylistItems(payload map[string]any) ([]any, string, error) {
	if shelf := findMusicPlaylistShelf(payload); shelf != nil {
		items := asSlice(shelf["contents"])
		if len(items) == 0 {
			return nil, "", errors.New("no playlist entries in YouTube Music shelf")
		}
		return items, findContinuationToken(items), nil
	}

	if cont := asMap(getPath(payload, "continuationContents", "musicPlaylistShelfContinuation")); cont != nil {
		items := asSlice(cont["contents"])
		if len(items) > 0 {
			return items, findContinuationToken(items), nil
		}
	}

	actions := asSlice(getPath(payload, "onResponseReceivedActions"))
	for _, action := range actions {
		appendAction := asMap(asMap(action)["appendContinuationItemsAction"])
		if appendAction == nil {
			continue
		}
		items := asSlice(appendAction["continuationItems"])
		if len(items) == 0 {
			continue
		}
		return items, findContinuationToken(items), nil
	}

	return nil, "", errors.New("no playlist items found in YouTube Music response")
}

func findMusicPlaylistShelf(payload map[string]any) map[string]any {
	queue := []any{payload}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		switch value := node.(type) {
		case map[string]any:
			if shelf := asMap(value["musicPlaylistShelfRenderer"]); shelf != nil {
				return shelf
			}
			for _, child := range value {
				switch child.(type) {
				case map[string]any, []any:
					queue = append(queue, child)
				}
			}
		case []any:
			for _, child := range value {
				switch child.(type) {
				case map[string]any, []any:
					queue = append(queue, child)
				}
			}
		}
	}
	return nil
}

func findContinuationToken(items []any) string {
	for _, item := range items {
		cont := asMap(asMap(item)["continuationItemRenderer"])
		if cont == nil {
			continue
		}
		token := getString(getPath(cont, "continuationEndpoint", "continuationCommand", "token"))
		if token != "" {
			return token
		}
	}
	return ""
}

func appendMusicEntries(dest map[string]musicEntryMeta, items []any) {
	for _, item := range items {
		renderer := asMap(asMap(item)["musicResponsiveListItemRenderer"])
		if renderer == nil {
			continue
		}
		videoID, meta := parseMusicEntry(renderer)
		if videoID == "" {
			continue
		}
		dest[videoID] = meta
	}
}

func parseMusicEntry(renderer map[string]any) (string, musicEntryMeta) {
	videoID := getString(getPath(renderer, "playlistItemData", "videoId"))
	if videoID == "" {
		videoID = findVideoIDFromRuns(renderer)
	}

	meta := musicEntryMeta{
		Title:  columnText(renderer, 0),
		Artist: findRunTextByPageType(renderer, "MUSIC_PAGE_TYPE_ARTIST"),
		Album:  findRunTextByPageType(renderer, "MUSIC_PAGE_TYPE_ALBUM"),
	}
	if meta.Title == "" {
		meta.Title = findRunTextWithWatchEndpoint(renderer)
	}
	if meta.Artist == "" {
		meta.Artist = columnText(renderer, 1)
	}
	if meta.Album == "" {
		meta.Album = columnText(renderer, 2)
	}

	return videoID, meta
}

func findRunTextByPageType(renderer map[string]any, pageType string) string {
	for _, group := range []string{"flexColumns", "secondaryFlexColumns"} {
		for _, col := range asSlice(renderer[group]) {
			colRenderer := asMap(asMap(col)["musicResponsiveListItemFlexColumnRenderer"])
			if colRenderer == nil {
				continue
			}
			runs := asSlice(getPath(colRenderer, "text", "runs"))
			for _, run := range runs {
				runMap := asMap(run)
				if getString(getPath(runMap, "navigationEndpoint", "browseEndpoint", "browseEndpointContextSupportedConfigs", "browseEndpointContextMusicConfig", "pageType")) == pageType {
					if text := getString(runMap["text"]); text != "" {
						return text
					}
				}
			}
		}
	}
	return ""
}

func findRunTextWithWatchEndpoint(renderer map[string]any) string {
	for _, group := range []string{"flexColumns", "secondaryFlexColumns"} {
		for _, col := range asSlice(renderer[group]) {
			colRenderer := asMap(asMap(col)["musicResponsiveListItemFlexColumnRenderer"])
			if colRenderer == nil {
				continue
			}
			runs := asSlice(getPath(colRenderer, "text", "runs"))
			for _, run := range runs {
				runMap := asMap(run)
				if getString(getPath(runMap, "navigationEndpoint", "watchEndpoint", "videoId")) != "" {
					if text := getString(runMap["text"]); text != "" {
						return text
					}
				}
			}
		}
	}
	return ""
}

func findVideoIDFromRuns(renderer map[string]any) string {
	for _, group := range []string{"flexColumns", "secondaryFlexColumns"} {
		for _, col := range asSlice(renderer[group]) {
			colRenderer := asMap(asMap(col)["musicResponsiveListItemFlexColumnRenderer"])
			if colRenderer == nil {
				continue
			}
			runs := asSlice(getPath(colRenderer, "text", "runs"))
			for _, run := range runs {
				runMap := asMap(run)
				if videoID := getString(getPath(runMap, "navigationEndpoint", "watchEndpoint", "videoId")); videoID != "" {
					return videoID
				}
			}
		}
	}
	return ""
}

func columnText(renderer map[string]any, index int) string {
	cols := asSlice(renderer["flexColumns"])
	if index < 0 || index >= len(cols) {
		return ""
	}
	colRenderer := asMap(asMap(cols[index])["musicResponsiveListItemFlexColumnRenderer"])
	if colRenderer == nil {
		return ""
	}
	return extractText(colRenderer["text"])
}

func extractText(value any) string {
	textMap := asMap(value)
	if textMap == nil {
		if s, ok := value.(string); ok {
			return s
		}
		return ""
	}
	if runs := asSlice(textMap["runs"]); len(runs) > 0 {
		return runsText(runs)
	}
	if s, ok := textMap["simpleText"].(string); ok {
		return s
	}
	return ""
}

func runsText(runs []any) string {
	var b strings.Builder
	for _, run := range runs {
		runMap := asMap(run)
		if runMap == nil {
			continue
		}
		if text, ok := runMap["text"].(string); ok {
			b.WriteString(text)
		}
	}
	return b.String()
}

func asMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	m, _ := value.(map[string]any)
	return m
}

func asSlice(value any) []any {
	if value == nil {
		return nil
	}
	s, _ := value.([]any)
	return s
}

func getPath(value map[string]any, keys ...string) any {
	var current any = value
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[key]
	}
	return current
}

func getString(value any) string {
	if value == nil {
		return ""
	}
	s, _ := value.(string)
	return s
}
