package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestConsistentTransportDoesNotMutateOriginalRequest(t *testing.T) {
	transport := &consistentTransport{
		base:      http.DefaultTransport,
		userAgent: "TestAgent/1.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	// Verify original request has no User-Agent, Accept-Language, or Accept headers.
	if got := req.Header.Get("User-Agent"); got != "" {
		t.Fatalf("expected empty User-Agent before RoundTrip, got %q", got)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()

	// After RoundTrip, the original request must remain unmodified.
	if got := req.Header.Get("User-Agent"); got != "" {
		t.Fatalf("RoundTrip mutated original request User-Agent to %q", got)
	}
	if got := req.Header.Get("Accept-Language"); got != "" {
		t.Fatalf("RoundTrip mutated original request Accept-Language to %q", got)
	}
	if got := req.Header.Get("Accept"); got != "" {
		t.Fatalf("RoundTrip mutated original request Accept to %q", got)
	}
}

func TestConsistentTransportSetsDefaultHeaders(t *testing.T) {
	var receivedUA, receivedLang, receivedAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		receivedLang = r.Header.Get("Accept-Language")
		receivedAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &consistentTransport{
		base:      http.DefaultTransport,
		userAgent: "TestAgent/1.0",
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()

	if receivedUA != "TestAgent/1.0" {
		t.Fatalf("expected User-Agent %q, got %q", "TestAgent/1.0", receivedUA)
	}
	if receivedLang != "en-US,en;q=0.9" {
		t.Fatalf("expected Accept-Language %q, got %q", "en-US,en;q=0.9", receivedLang)
	}
	if receivedAccept != "*/*" {
		t.Fatalf("expected Accept %q, got %q", "*/*", receivedAccept)
	}
}

func TestConsistentTransportPreservesExistingHeaders(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &consistentTransport{
		base:      http.DefaultTransport,
		userAgent: "TestAgent/1.0",
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("User-Agent", "CustomAgent/2.0")

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()

	if receivedUA != "CustomAgent/2.0" {
		t.Fatalf("expected preserved User-Agent %q, got %q", "CustomAgent/2.0", receivedUA)
	}
}

func TestConsistentTransportConcurrentSafety(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &consistentTransport{
		base:      http.DefaultTransport,
		userAgent: "TestAgent/1.0",
	}

	// Create a single request and use it concurrently to detect data races.
	// With -race flag, this would catch mutations to the original request.
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Errorf("RoundTrip: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	// Original request must still have no headers set.
	if got := req.Header.Get("User-Agent"); got != "" {
		t.Fatalf("concurrent RoundTrip mutated original request User-Agent to %q", got)
	}
}

func TestDoWithRetryRetriesOn5xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "temporarily unavailable")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "success")
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := doWithRetry(req, 5*time.Second, 3)
	if err != nil {
		t.Fatalf("doWithRetry returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoWithRetryExhaustsAttempts(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := doWithRetry(req, 10*time.Second, 3)
	if err == nil {
		if resp != nil {
			resp.Body.Close()
		}
		t.Fatalf("expected error after exhausting retries")
	}
	// retryTransport makes 1 initial attempt plus defaultRetryConfig.MaxRetries retries
	// before returning the final 5xx, which doWithRetry then converts to an error.
	expectedAttempts := 1 + defaultRetryConfig.MaxRetries
	if attempts != expectedAttempts {
		t.Fatalf("expected %d server hits (retryTransport: 1 initial + %d retries), got %d",
			expectedAttempts, defaultRetryConfig.MaxRetries, attempts)
	}
}

func TestDoWithRetryPassesThrough4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := doWithRetry(req, 5*time.Second, 3)
	if err != nil {
		t.Fatalf("doWithRetry returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
	// 4xx should not be retried
	if attempts != 1 {
		t.Fatalf("expected 1 attempt for 4xx, got %d", attempts)
	}
}

func TestNewClientForType_Android(t *testing.T) {
	opts := Options{Timeout: 5 * time.Second}
	client := newClientForType("android", opts)
	if client == nil {
		t.Fatal("expected non-nil client for android type")
	}
	adapter, ok := client.(*youtubeClientAdapter)
	if !ok {
		t.Fatal("expected *youtubeClientAdapter underlying type")
	}
	if adapter.Client.ClientType == nil {
		t.Fatal("expected ClientType to be set")
	}
	if adapter.Client.ClientType.Name != "ANDROID" {
		t.Fatalf("expected ClientType.Name %q, got %q", "ANDROID", adapter.Client.ClientType.Name)
	}
}

func TestNewClientForType_Web(t *testing.T) {
	opts := Options{Timeout: 5 * time.Second}
	client := newClientForType("web", opts)
	if client == nil {
		t.Fatal("expected non-nil client for web type")
	}
	adapter, ok := client.(*youtubeClientAdapter)
	if !ok {
		t.Fatal("expected *youtubeClientAdapter underlying type")
	}
	if adapter.Client.ClientType == nil {
		t.Fatal("expected ClientType to be set")
	}
	if adapter.Client.ClientType.Name != "WEB" {
		t.Fatalf("expected ClientType.Name %q, got %q", "WEB", adapter.Client.ClientType.Name)
	}
}

func TestNewClientForType_ConcurrentSafety(t *testing.T) {
	opts := Options{Timeout: 5 * time.Second}
	clients := make([]YouTubeClient, 10)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			clients[i] = newClientForType("web", opts)
		}()
	}
	wg.Wait()

	// Verify each goroutine got its own distinct client with correct state.
	seen := make(map[*http.Client]bool)
	for i, c := range clients {
		if c == nil {
			t.Errorf("client[%d] is nil", i)
			continue
		}
		adapter, ok := c.(*youtubeClientAdapter)
		if !ok {
			t.Errorf("client[%d] is not *youtubeClientAdapter", i)
			continue
		}
		if adapter.Client.ClientType == nil || adapter.Client.ClientType.Name != "WEB" {
			t.Errorf("client[%d] has wrong ClientType", i)
		}
		// Each client must own a distinct *http.Client — no shared state.
		if seen[adapter.Client.HTTPClient] {
			t.Errorf("client[%d] shares an *http.Client with another goroutine's client", i)
		}
		seen[adapter.Client.HTTPClient] = true
	}
}
