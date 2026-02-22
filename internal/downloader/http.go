package downloader

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/kkdai/youtube/v2"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
const musicUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

// sharedTransport is a global HTTP transport with proper connection pooling settings
// to avoid resource exhaustion from creating new connections for each request.
// This transport is immutable after initialization and safe for concurrent use.
var sharedTransport = &http.Transport{
	MaxIdleConns:        100,              // Maximum idle connections across all hosts
	MaxIdleConnsPerHost: 10,               // Maximum idle connections per host
	MaxConnsPerHost:     0,                // Unlimited concurrent connections per host
	IdleConnTimeout:     90 * time.Second, // How long idle connections are kept alive
}

// CloseIdleConnections closes any idle connections in the shared transport.
// Call this at program exit to ensure clean shutdown.
func CloseIdleConnections() {
	sharedTransport.CloseIdleConnections()
}

type consistentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *consistentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request before modifying headers. RoundTrip must not mutate
	// the original request per the http.RoundTripper contract; doing so can
	// cause data races when the HTTP client retries or follows redirects.
	clone := req.Clone(req.Context())

	if clone.Header.Get("User-Agent") == "" {
		clone.Header.Set("User-Agent", t.userAgent)
	}
	if clone.Header.Get("Accept-Language") == "" {
		clone.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}
	if clone.Header.Get("Accept") == "" {
		clone.Header.Set("Accept", "*/*")
	}
	return t.base.RoundTrip(clone)
}

// newHTTPClient creates an HTTP client with a consistent transport and given timeout
func newHTTPClient(timeout time.Duration) *http.Client {
	transport := &consistentTransport{
		base:      sharedTransport,
		userAgent: defaultUserAgent,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func newClient(opts Options) *youtube.Client {
	// Create cookie jar for persistent cookies across requests
	jar, _ := cookiejar.New(nil)

	// Wrap with consistent headers
	transport := &consistentTransport{
		base:      sharedTransport,
		userAgent: defaultUserAgent,
	}

	httpClient := &http.Client{
		Timeout:   opts.Timeout,
		Jar:       jar,
		Transport: transport,
	}

	return &youtube.Client{HTTPClient: httpClient}
}
