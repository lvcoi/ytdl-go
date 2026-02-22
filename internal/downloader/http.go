package downloader

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	youtube "github.com/lvcoi/ytdl-lib/v2"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
const musicUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

var sharedTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout:   10 * time.Second,
	ResponseHeaderTimeout: 15 * time.Second,
	IdleConnTimeout:       90 * time.Second,
}

func CloseIdleConnections() {
	sharedTransport.CloseIdleConnections()
}

type consistentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *consistentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.userAgent)
	}
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}
	return t.base.RoundTrip(req)
}

type bgTransport struct {
	base    http.RoundTripper
	poToken string
}

func (t *bgTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "/youtubei/v1/player") && t.poToken != "" && req.Method == "POST" {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			var data map[string]interface{}
			if json.Unmarshal(body, &data) == nil {
				if playbackContext, ok := data["playbackContext"].(map[string]interface{}); ok {
					if contentPlaybackContext, ok := playbackContext["contentPlaybackContext"].(map[string]interface{}); ok {
						contentPlaybackContext["serviceIntegrityDimensions"] = map[string]interface{}{
							"poToken": t.poToken,
						}
					}
				}
				newBody, _ := json.Marshal(data)
				req.Body = io.NopCloser(bytes.NewReader(newBody))
				req.ContentLength = int64(len(newBody))
			} else {
				req.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
	}
	return t.base.RoundTrip(req)
}

func newHTTPClient(timeout time.Duration) *http.Client {
	var transport http.RoundTripper = &consistentTransport{
		base:      sharedTransport,
		userAgent: defaultUserAgent,
	}
	transport = newRetryTransport(transport, defaultRetryConfig)
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func newClient(opts Options) YouTubeClient {
	jar, _ := cookiejar.New(nil)
	var transport http.RoundTripper = &consistentTransport{
		base:      sharedTransport,
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
	transport = newRetryTransport(transport, defaultRetryConfig)
	httpClient := &http.Client{
		Timeout:   opts.Timeout,
		Jar:       jar,
		Transport: transport,
	}
	return &youtubeClientAdapter{&youtube.Client{HTTPClient: httpClient}}
}

// newClientForType creates a YouTubeClient configured for the given
// client type ("android", "web", "ios", "embedded").
func newClientForType(clientType string, opts Options) YouTubeClient {
	client := newClient(opts)
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
