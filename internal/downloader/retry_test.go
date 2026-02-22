package downloader

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryTransport_NoRetryOnSuccess(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}), defaultRetryConfig)

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("expected 1 call, got %d", c)
	}
}

func TestRetryTransport_RetriesOn5xx(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 2 {
			return &http.Response{StatusCode: 502, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 3 {
		t.Fatalf("expected 3 calls, got %d", c)
	}
}

func TestRetryTransport_RetriesOn429(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return &http.Response{StatusCode: 429, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_NoRetryOn403(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: 403, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("expected 1 call (no retry for 403), got %d", c)
	}
}

func TestRetryTransport_NoRetryOn400(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: 400, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("expected 1 call (no retry for 400), got %d", c)
	}
}

func TestRetryTransport_ExhaustedRetries(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: 503, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 2, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return last response after exhausting retries
	if resp.StatusCode != 503 {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 3 { // 1 initial + 2 retries
		t.Fatalf("expected 3 calls, got %d", c)
	}
}

func TestRetryTransport_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			cancel() // Cancel after first call
			return &http.Response{StatusCode: 502, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)
	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Fatalf("expected 1 call before cancellation, got %d", c)
	}
}

func TestRetryTransport_RetriesOnTimeout(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return nil, &net.OpError{Op: "dial", Err: &timeoutError{}}
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Fatalf("expected 2 calls, got %d", c)
	}
}

func TestRetryTransport_RetriesWithBody(t *testing.T) {
	var calls int32
	transport := newRetryTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		// Read body to verify it's present on retry
		if req.Body != nil {
			body, _ := io.ReadAll(req.Body)
			if string(body) != "test-body" {
				t.Fatalf("attempt %d: unexpected body: %q", n, body)
			}
		}
		if n == 1 {
			return &http.Response{StatusCode: 500, Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	}), retryConfig{MaxRetries: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})

	body := "test-body"
	req, _ := http.NewRequest("POST", "https://example.com", strings.NewReader(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Fatalf("expected 2 calls, got %d", c)
	}
}

func TestBackoffDelay(t *testing.T) {
	rt := newRetryTransport(nil, retryConfig{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
	})

	// Attempt 1: ~100ms ± 25%
	d1 := rt.backoffDelay(1)
	if d1 < 75*time.Millisecond || d1 > 125*time.Millisecond {
		t.Fatalf("attempt 1 delay out of range: %v", d1)
	}

	// Attempt 2: ~200ms ± 25%
	d2 := rt.backoffDelay(2)
	if d2 < 150*time.Millisecond || d2 > 250*time.Millisecond {
		t.Fatalf("attempt 2 delay out of range: %v", d2)
	}

	// Attempt 3: ~400ms ± 25%
	d3 := rt.backoffDelay(3)
	if d3 < 300*time.Millisecond || d3 > 500*time.Millisecond {
		t.Fatalf("attempt 3 delay out of range: %v", d3)
	}
}

func TestIsRetryableStatus(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	nonRetryable := []int{200, 201, 301, 400, 401, 403, 404}

	for _, code := range retryable {
		if !isRetryableStatus(code) {
			t.Errorf("expected %d to be retryable", code)
		}
	}
	for _, code := range nonRetryable {
		if isRetryableStatus(code) {
			t.Errorf("expected %d to NOT be retryable", code)
		}
	}
}

// --- Test helpers ---

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// timeoutError is a mock error that satisfies net.Error with Timeout() = true.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true } //nolint:staticcheck
