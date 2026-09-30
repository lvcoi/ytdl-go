package downloader

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	youtube "github.com/lvcoi/ytdl-lib/v2"
)

// activeProxy holds an optional proxy URL (http/https/socks5) configured via
// --proxy. It is read by proxyBaseTransport when building new transports.
var activeProxy atomic.Pointer[url.URL]

// ConfigureProxy sets the process-wide proxy used for all outgoing requests.
func ConfigureProxy(proxyURL string) error {
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		activeProxy.Store(nil)
		return nil
	}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return wrapCategory(CategoryInvalidURL, fmt.Errorf("invalid proxy URL: %w", err))
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return wrapCategory(CategoryInvalidURL, fmt.Errorf("unsupported proxy scheme %q (use http, https or socks5)", parsed.Scheme))
	}
	activeProxy.Store(parsed)
	return nil
}

// proxyBaseTransport returns the shared transport, or a proxy-configured clone.
func proxyBaseTransport() *http.Transport {
	proxy := activeProxy.Load()
	if proxy == nil {
		return sharedTransport
	}
	clone := sharedTransport.Clone()
	clone.Proxy = func(*http.Request) (*url.URL, error) {
		return proxy, nil
	}
	return clone
}

// sleepTransport adds --sleep-requests throttling between YouTube API calls
// (POST requests to /youtubei/ endpoints) without delaying media GETs.
type sleepTransport struct {
	base          http.RoundTripper
	sleepDuration time.Duration
}

func (t *sleepTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if t.sleepDuration > 0 && req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/youtubei/") {
		time.Sleep(t.sleepDuration)
	}
	return resp, err
}

// newClientWithOptions builds a YouTubeClient honoring proxy, cookies and
// sleep-requests options. It mirrors newClient from http.go and should be
// used in its place everywhere so the options take effect.
func newClientWithOptions(opts Options) YouTubeClient {
	jar, _ := cookiejar.New(nil)
	var transport http.RoundTripper = &consistentTransport{
		base:      proxyBaseTransport(),
		userAgent: defaultUserAgent,
	}
	if bg, err := NewBgUtils(); err == nil {
		if token, err := bg.GeneratePlaceholder("WEB"); err == nil {
			transport = &bgTransport{
				base:    transport,
				poToken: token,
			}
		}
	}
	if opts.SleepRequests > 0 {
		transport = &sleepTransport{
			base:          transport,
			sleepDuration: opts.SleepRequests,
		}
	}
	transport = newRetryTransport(transport, defaultRetryConfig)
	httpClient := &http.Client{
		Timeout:   opts.Timeout,
		Jar:       jar,
		Transport: transport,
	}
	if staticCookies, err := loadStaticCookies(opts); err == nil && len(staticCookies) > 0 {
		httpClient.Jar = &staticCookieJar{primary: jar, static: staticCookies}
	}
	return &youtubeClientAdapter{&youtube.Client{HTTPClient: httpClient}}
}

// newClientTypeForType creates a YouTubeClient for the given client type,
// honoring proxy/cookie/sleep options (options-aware newClientForType).
func newClientTypeForType(clientType string, opts Options) YouTubeClient {
	client := newClientWithOptions(opts)
	switch clientType {
	case "web":
		client.SetClientInfo(youtube.WebClient)
	case "android":
		client.SetClientInfo(youtube.AndroidClient)
	case "ios":
		client.SetClientInfo(youtube.IOSClient)
	case "embedded":
		client.SetClientInfo(youtube.EmbeddedClient)
	default:
		client.SetClientInfo(youtube.AndroidClient)
	}
	return client
}
