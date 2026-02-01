package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/refraction-networking/utls"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	poTokenQueryKey  = "pot"
	visitorHeader    = "X-Goog-Visitor-Id"
	poTokenHeader    = "X-YouTube-PO-Token"
	visitorQueryKey  = "visitorData"
)

// YTSession holds the dynamic requirements needed to satisfy the server.
type YTSession struct {
	Client      *http.Client
	PoToken     string
	VisitorData string
	Cookies     []*http.Cookie
	UserAgent   string
	VideoID     string
	NTransform  func(string) (string, error)

	mu sync.RWMutex
}

// NewSession initializes a client with a spoofed TLS fingerprint (uTLS).
func NewSession() *YTSession {
	jar, _ := cookiejar.New(nil)
	transport := newUTLSTransport(utls.HelloChrome_120)
	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   45 * time.Second,
	}
	return &YTSession{
		Client:    client,
		UserAgent: defaultUserAgent,
	}
}

// RefreshSessionRequirements uses a headless browser to get fresh tokens.
// This is the "Solver" - it fixes the "Why" of the 403.
func (s *YTSession) RefreshSessionRequirements(videoID string) error {
	if strings.TrimSpace(videoID) == "" {
		return errors.New("missing videoID for session refresh")
	}
	log.Println("[Solver] 403 detected. Refreshing PoToken and VisitorData via Headless Chrome...")

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	watchURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	var rawPayload string
	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(watchURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Evaluate(`(function () {
			function getCfg(key) {
				try {
					if (window.ytcfg && window.ytcfg.get) {
						return window.ytcfg.get(key);
					}
				} catch (e) {}
				if (window.yt && window.yt.config_) {
					return window.yt.config_[key];
				}
				return null;
			}
			return JSON.stringify({
				visitorData: getCfg("VISITOR_DATA"),
				poToken: getCfg("PO_TOKEN") || getCfg("PO_TOKEN_2")
			});
		})();`, &rawPayload),
	); err != nil {
		return fmt.Errorf("chromedp navigation failed: %w", err)
	}

	var payload struct {
		VisitorData string `json:"visitorData"`
		PoToken     string `json:"poToken"`
	}
	if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
		return fmt.Errorf("parsing session payload: %w", err)
	}
	if payload.VisitorData == "" || payload.PoToken == "" {
		return fmt.Errorf("missing visitorData or poToken (visitor=%t, po=%t)", payload.VisitorData != "", payload.PoToken != "")
	}

	cdpCookies, err := network.GetAllCookies().Do(ctx)
	if err != nil {
		return fmt.Errorf("reading browser cookies: %w", err)
	}
	converted := make([]*http.Cookie, 0, len(cdpCookies))
	for _, cookie := range cdpCookies {
		converted = append(converted, &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HTTPOnly,
			Expires:  time.Unix(int64(cookie.Expires), 0),
		})
	}

	s.mu.Lock()
	s.PoToken = payload.PoToken
	s.VisitorData = payload.VisitorData
	s.Cookies = converted
	s.VideoID = videoID
	s.mu.Unlock()

	if s.Client != nil && s.Client.Jar != nil {
		if parsed, err := url.Parse("https://www.youtube.com"); err == nil {
			s.Client.Jar.SetCookies(parsed, converted)
		}
	}

	return nil
}

// DownloadFragment attempts to fetch a specific byte range.
// It enforces the "Fail Fast, Solve, Retry" loop.
func (s *YTSession) DownloadFragment(ctx context.Context, rawURL string, start, end int64) ([]byte, error) {
	if s == nil || s.Client == nil {
		return nil, errors.New("nil session client")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for attempt := 1; attempt <= 2; attempt++ {
		req, err := s.newRequest(ctx, rawURL, start, end)
		if err != nil {
			return nil, err
		}
		resp, err := s.Client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
			defer resp.Body.Close()
			return io.ReadAll(resp.Body)
		}

		resp.Body.Close()
		if resp.StatusCode == http.StatusForbidden {
			log.Printf("[Attempt %d] 403 Forbidden. Session invalid.", attempt)
			s.mu.RLock()
			videoID := s.VideoID
			s.mu.RUnlock()
			if err := s.RefreshSessionRequirements(videoID); err != nil {
				return nil, fmt.Errorf("failed to solve requirements: %w", err)
			}
			continue
		}

		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return nil, errors.New("failed to download fragment after solving requirements")
}

func downloadSegmentWithSession(ctx context.Context, session *YTSession, segmentURL string, writer io.Writer) error {
	if session == nil {
		return errors.New("nil session for segment download")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for attempt := 1; attempt <= 2; attempt++ {
		req, err := session.newRequest(ctx, segmentURL, -1, -1)
		if err != nil {
			return err
		}
		resp, err := session.Client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
			_, copyErr := io.Copy(writer, resp.Body)
			resp.Body.Close()
			if copyErr == nil {
				return nil
			}
			return copyErr
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusForbidden {
			log.Printf("[Attempt %d] 403 Forbidden. Session invalid.", attempt)
			session.mu.RLock()
			videoID := session.VideoID
			session.mu.RUnlock()
			if err := session.RefreshSessionRequirements(videoID); err != nil {
				return fmt.Errorf("failed to solve requirements: %w", err)
			}
			continue
		}
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return errors.New("failed to download segment after solving requirements")
}

func (s *YTSession) newRequest(ctx context.Context, rawURL string, start, end int64) (*http.Request, error) {
	prepared, err := s.prepareURL(rawURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, prepared, nil)
	if err != nil {
		return nil, err
	}

	if start >= 0 && end >= start {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	}

	s.mu.RLock()
	userAgent := s.UserAgent
	visitor := s.VisitorData
	poToken := s.PoToken
	cookies := append([]*http.Cookie(nil), s.Cookies...)
	s.mu.RUnlock()

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if visitor != "" {
		req.Header.Set(visitorHeader, visitor)
	}
	if poToken != "" {
		req.Header.Set(poTokenHeader, poToken)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://www.youtube.com")
	req.Header.Set("Referer", "https://www.youtube.com/")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

func (s *YTSession) prepareURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()

	s.mu.RLock()
	poToken := s.PoToken
	visitor := s.VisitorData
	transform := s.NTransform
	s.mu.RUnlock()

	if poToken != "" && query.Get(poTokenQueryKey) == "" {
		query.Set(poTokenQueryKey, poToken)
	}
	if visitor != "" && query.Get(visitorQueryKey) == "" {
		query.Set(visitorQueryKey, visitor)
	}

	if transform != nil {
		if nVal := query.Get("n"); nVal != "" {
			newVal, err := transform(nVal)
			if err != nil {
				return "", fmt.Errorf("transforming n param: %w", err)
			}
			query.Set("n", newVal)
		}
	}

	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func newUTLSTransport(hello utls.ClientHelloID) *http.Transport {
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			rawConn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				rawConn.Close()
				return nil, err
			}
			config := &utls.Config{ServerName: host, NextProtos: []string{"h2", "http/1.1"}}
			conn := utls.UClient(rawConn, config, hello)
			if err := conn.Handshake(); err != nil {
				rawConn.Close()
				return nil, err
			}
			return conn, nil
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
