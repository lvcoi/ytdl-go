package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var cwdMu sync.Mutex

func withTempCWD(t *testing.T, fn func(tmpDir string)) {
	t.Helper()
	cwdMu.Lock()
	defer cwdMu.Unlock()

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWD)
	}()

	fn(tmpDir)
}

func startWebServerForTest(t *testing.T, ctx context.Context) (baseURL string, wait func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- ListenAndServe(ctx, addr)
	}()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	statusURL := fmt.Sprintf("http://%s/api/status", addr)
	deadline := time.Now().Add(5 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready in time")
		}
		resp, err := client.Get(statusURL)
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	waitFn := func() {
		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("server error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for server shutdown")
		}
	}

	return fmt.Sprintf("http://%s", addr), waitFn
}

func TestMediaListPaginationEndpoint(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		tracker = &jobTracker{}

		mediaDir := filepath.Join(tmpDir, "media")
		if err := os.MkdirAll(mediaDir, 0o755); err != nil {
			t.Fatalf("mkdir media: %v", err)
		}

		createMedia := func(name string, modTime time.Time) {
			path := filepath.Join(mediaDir, name)
			if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
				t.Fatalf("write media %s: %v", name, err)
			}
			if err := os.Chtimes(path, modTime, modTime); err != nil {
				t.Fatalf("chtimes %s: %v", name, err)
			}
		}

		now := time.Now()
		createMedia("old.mp4", now.Add(-3*time.Hour))
		createMedia("mid.mp4", now.Add(-2*time.Hour))
		createMedia("new.mp3", now.Add(-1*time.Hour))

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}

		resp, err := client.Get(baseURL + "/api/media/?limit=2&offset=0")
		if err != nil {
			t.Fatalf("request first page: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for first page, got %d", resp.StatusCode)
		}

		var page1 mediaListResponse
		if err := json.NewDecoder(resp.Body).Decode(&page1); err != nil {
			t.Fatalf("decode first page: %v", err)
		}
		if len(page1.Items) != 2 {
			t.Fatalf("expected 2 items on first page, got %d", len(page1.Items))
		}
		if page1.Items[0].Filename != "new.mp3" {
			t.Fatalf("expected newest file first, got %q", page1.Items[0].Filename)
		}
		if page1.NextOffset == nil || *page1.NextOffset != 2 {
			t.Fatalf("expected next_offset=2, got %v", page1.NextOffset)
		}

		resp2, err := client.Get(baseURL + "/api/media/?limit=2&offset=2")
		if err != nil {
			t.Fatalf("request second page: %v", err)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for second page, got %d", resp2.StatusCode)
		}

		var page2 mediaListResponse
		if err := json.NewDecoder(resp2.Body).Decode(&page2); err != nil {
			t.Fatalf("decode second page: %v", err)
		}
		if len(page2.Items) != 1 {
			t.Fatalf("expected 1 item on second page, got %d", len(page2.Items))
		}
		if page2.NextOffset != nil {
			t.Fatalf("expected no next_offset on final page, got %v", *page2.NextOffset)
		}

		badResp, err := client.Get(baseURL + "/api/media/?offset=-1")
		if err != nil {
			t.Fatalf("request invalid offset: %v", err)
		}
		defer badResp.Body.Close()
		if badResp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid offset, got %d", badResp.StatusCode)
		}
	})
}

func TestMediaFileServePathValidation(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		tracker = &jobTracker{}

		mediaDir := filepath.Join(tmpDir, "media")
		if err := os.MkdirAll(mediaDir, 0o755); err != nil {
			t.Fatalf("mkdir media: %v", err)
		}

		insidePath := filepath.Join(mediaDir, "inside.mp4")
		if err := os.WriteFile(insidePath, []byte("ok"), 0o644); err != nil {
			t.Fatalf("write inside media: %v", err)
		}

		outsidePath := filepath.Join(tmpDir, "outside.txt")
		if err := os.WriteFile(outsidePath, []byte("secret"), 0o644); err != nil {
			t.Fatalf("write outside file: %v", err)
		}

		symlinkPath := filepath.Join(mediaDir, "escape.txt")
		symlinkErr := os.Symlink(outsidePath, symlinkPath)

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}

		validResp, err := client.Get(baseURL + "/api/media/inside.mp4")
		if err != nil {
			t.Fatalf("request valid file: %v", err)
		}
		defer validResp.Body.Close()
		if validResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for valid media file, got %d", validResp.StatusCode)
		}

		traversalResp, err := client.Get(baseURL + "/api/media/%2e%2e%2foutside.txt")
		if err != nil {
			t.Fatalf("request traversal path: %v", err)
		}
		defer traversalResp.Body.Close()
		if traversalResp.StatusCode != http.StatusBadRequest && traversalResp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 400/403 for traversal path, got %d", traversalResp.StatusCode)
		}

		if symlinkErr == nil {
			symlinkResp, err := client.Get(baseURL + "/api/media/escape.txt")
			if err != nil {
				t.Fatalf("request symlink path: %v", err)
			}
			defer symlinkResp.Body.Close()
			if symlinkResp.StatusCode != http.StatusForbidden {
				t.Fatalf("expected 403 for symlink escape, got %d", symlinkResp.StatusCode)
			}
		}
	})
}

func TestResolveMediaPathRejectsInvalidAndSymlinkEscape(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatalf("mkdir media: %v", err)
	}

	validFile := filepath.Join(mediaDir, "song.mp3")
	if err := os.WriteFile(validFile, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write valid file: %v", err)
	}

	resolved, status, err := resolveMediaPath(mediaDir, "song.mp3")
	if err != nil {
		t.Fatalf("expected valid media path, got error: %v", err)
	}
	if status != 0 {
		t.Fatalf("expected status 0, got %d", status)
	}
	if resolved != validFile {
		t.Fatalf("expected resolved path %q, got %q", validFile, resolved)
	}

	if _, status, err := resolveMediaPath(mediaDir, "../outside.txt"); err == nil || status != http.StatusBadRequest {
		t.Fatalf("expected bad request for traversal path")
	}

	outsidePath := filepath.Join(tmpDir, "outside.txt")
	if err := os.WriteFile(outsidePath, []byte("secret"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	symlinkPath := filepath.Join(mediaDir, "escape.txt")
	if err := os.Symlink(outsidePath, symlinkPath); err == nil {
		if _, status, err := resolveMediaPath(mediaDir, "escape.txt"); err == nil || status != http.StatusForbidden {
			t.Fatalf("expected forbidden for symlink escape")
		}
	}
}

func TestDecodeJSONBodyRejectsOversizedPayload(t *testing.T) {
	withTempCWD(t, func(_ string) {
		tracker = &jobTracker{}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}
		oversizedField := strings.Repeat("a", maxRequestBodyBytes+1)

		downloadPayload := []byte(fmt.Sprintf(`{"urls":["%s"],"options":{"output":"{title}.{ext}"}}`, oversizedField))
		reqDownload, err := http.NewRequest(http.MethodPost, baseURL+"/api/download", bytes.NewReader(downloadPayload))
		if err != nil {
			t.Fatalf("new request for /api/download: %v", err)
		}
		reqDownload.Header.Set("Content-Type", "application/json")

		respDownload, err := client.Do(reqDownload)
		if err != nil {
			t.Fatalf("request /api/download: %v", err)
		}
		defer respDownload.Body.Close()
		if respDownload.StatusCode != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected 413 for /api/download, got %d", respDownload.StatusCode)
		}

		dupPayload := []byte(fmt.Sprintf(`{"jobId":"job_1","promptId":"dup_1","choice":"%s"}`, oversizedField))
		reqDuplicate, err := http.NewRequest(http.MethodPost, baseURL+"/api/download/duplicate-response", bytes.NewReader(dupPayload))
		if err != nil {
			t.Fatalf("new request for /api/download/duplicate-response: %v", err)
		}
		reqDuplicate.Header.Set("Content-Type", "application/json")

		respDuplicate, err := client.Do(reqDuplicate)
		if err != nil {
			t.Fatalf("request /api/download/duplicate-response: %v", err)
		}
		defer respDuplicate.Body.Close()
		if respDuplicate.StatusCode != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected 413 for /api/download/duplicate-response, got %d", respDuplicate.StatusCode)
		}
	})
}

func TestSecurityHeadersPresentOnResponses(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		tracker = &jobTracker{}

		mediaDir := filepath.Join(tmpDir, "media")
		if err := os.MkdirAll(mediaDir, 0o755); err != nil {
			t.Fatalf("mkdir media: %v", err)
		}
		if err := os.WriteFile(filepath.Join(mediaDir, "sample.mp4"), []byte("ok"), 0o644); err != nil {
			t.Fatalf("write media file: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}
		paths := []string{"/api/status", "/api/media/sample.mp4", "/"}
		for _, path := range paths {
			resp, err := client.Get(baseURL + path)
			if err != nil {
				t.Fatalf("request %s: %v", path, err)
			}
			resp.Body.Close()

			if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
				t.Fatalf("%s: expected X-Content-Type-Options=nosniff, got %q", path, got)
			}
			if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
				t.Fatalf("%s: expected X-Frame-Options=DENY, got %q", path, got)
			}
			if got := resp.Header.Get("Referrer-Policy"); got != "no-referrer" {
				t.Fatalf("%s: expected Referrer-Policy=no-referrer, got %q", path, got)
			}
			csp := resp.Header.Get("Content-Security-Policy")
			if !strings.Contains(csp, "default-src 'self'") || !strings.Contains(csp, "media-src 'self'") {
				t.Fatalf("%s: expected strict CSP, got %q", path, csp)
			}
		}
	})
}

func TestMediaTypeForExtension(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{ext: ".mp3", want: "audio"},
		{ext: ".opus", want: "audio"},
		{ext: ".wma", want: "audio"},
		{ext: ".m4a", want: "audio"},
		{ext: ".mp4", want: "video"},
		{ext: ".m4v", want: "video"},
		{ext: ".ts", want: "video"},
		{ext: ".ogv", want: "video"},
		{ext: ".unknown", want: "video"},
	}

	for _, tt := range tests {
		if got := mediaTypeForExtension(tt.ext); got != tt.want {
			t.Fatalf("mediaTypeForExtension(%q): expected %q, got %q", tt.ext, tt.want, got)
		}
	}
}
