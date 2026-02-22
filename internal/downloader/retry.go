package downloader

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// retryConfig controls retry behavior for HTTP requests.
type retryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

var defaultRetryConfig = retryConfig{
	MaxRetries:   3,
	InitialDelay: 500 * time.Millisecond,
	MaxDelay:     8 * time.Second,
}

// retryTransport wraps an http.RoundTripper and retries transient failures
// with exponential backoff and jitter.
type retryTransport struct {
	base   http.RoundTripper
	config retryConfig
}

func newRetryTransport(base http.RoundTripper, config retryConfig) *retryTransport {
	return &retryTransport{base: base, config: config}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= t.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := t.backoffDelay(attempt)
			if err := sleepWithContext(req.Context(), delay); err != nil {
				if lastResp != nil {
					lastResp.Body.Close()
				}
				return nil, err
			}
		}

		// Clone the request for retry safety (body may have been consumed).
		cloned := req
		if attempt > 0 {
			var err error
			cloned, err = cloneRequest(req)
			if err != nil {
				// Can't clone (e.g. streaming body) — return last result
				if lastResp != nil {
					return lastResp, nil
				}
				return nil, lastErr
			}
		}

		resp, err := t.base.RoundTrip(cloned)
		if err != nil {
			if !isRetryableError(err) {
				return nil, err
			}
			lastErr = err
			continue
		}

		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Close the body before retrying to free the connection
		if lastResp != nil {
			lastResp.Body.Close()
		}
		lastResp = resp
		lastErr = nil
	}

	// Exhausted retries — return whatever we got last
	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}

// backoffDelay calculates delay with exponential backoff and jitter.
func (t *retryTransport) backoffDelay(attempt int) time.Duration {
	base := float64(t.config.InitialDelay) * math.Pow(2, float64(attempt-1))
	if base > float64(t.config.MaxDelay) {
		base = float64(t.config.MaxDelay)
	}
	// Add jitter: ±25%
	jitter := base * 0.25 * (rand.Float64()*2 - 1) //nolint:gosec
	return time.Duration(base + jitter)
}

// isRetryableStatus returns true for HTTP status codes that indicate transient failures.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// isRetryableError returns true for network errors that are typically transient.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Timeout errors (connection timeout, TLS handshake timeout, etc.)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	// Connection reset, refused, etc.
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	return false
}

// cloneRequest creates a copy of the request suitable for retry.
func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		clone.Body = body
	}
	return clone, nil
}

// sleepWithContext sleeps for the given duration, returning early if the context is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
