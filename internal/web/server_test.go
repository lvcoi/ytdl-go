package web

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
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
	origTracker := tracker

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	tracker = &jobTracker{}
	defer func() {
		tracker = origTracker
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
		select {
		case serveErr := <-errCh:
			if serveErr != nil && !errors.Is(serveErr, context.Canceled) {
				t.Fatalf("server failed before readiness: %v", serveErr)
			}
			t.Fatalf("server exited before readiness")
		default:
		}

		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready in time")
		}
		resp, err := client.Get(statusURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
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

func TestEnsureMediaLayoutCreatesRequiredSubdirs(t *testing.T) {
	mediaDir := filepath.Join(t.TempDir(), "media")
	if err := ensureMediaLayout(mediaDir); err != nil {
		t.Fatalf("ensureMediaLayout: %v", err)
	}

	required := []string{mediaFolderAudio, mediaFolderVideo, mediaFolderPlaylist, mediaFolderData}
	for _, folder := range required {
		path := filepath.Join(mediaDir, folder)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", path)
		}
	}
}

func TestResolveWebMediaDirDefaultsToMediaFolder(t *testing.T) {
	withTempCWD(t, func(_ string) {
		resolved, err := resolveWebMediaDir()
		if err != nil {
			t.Fatalf("resolveWebMediaDir: %v", err)
		}
		want, err := filepath.Abs(defaultMediaDirName)
		if err != nil {
			t.Fatalf("abs default media dir: %v", err)
		}
		if resolved != want {
			t.Fatalf("expected default media directory %q, got %q", want, resolved)
		}
	})
}

func TestResolveWebMediaDirPrefersFolderWithExistingMedia(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		defaultMediaDir := filepath.Join(tmpDir, defaultMediaDirName)
		legacyMediaDir := filepath.Join(tmpDir, legacyMediaDirName)
		if err := os.MkdirAll(defaultMediaDir, 0o755); err != nil {
			t.Fatalf("mkdir default media dir: %v", err)
		}
		if err := os.MkdirAll(legacyMediaDir, 0o755); err != nil {
			t.Fatalf("mkdir legacy media dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(defaultMediaDir, "saved-playlists.json"), []byte("{}"), 0o644); err != nil {
			t.Fatalf("write default metadata file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(legacyMediaDir, "existing-track.mp3"), []byte("audio"), 0o644); err != nil {
			t.Fatalf("write legacy media file: %v", err)
		}

		resolved, err := resolveWebMediaDir()
		if err != nil {
			t.Fatalf("resolveWebMediaDir: %v", err)
		}
		expectedLegacyDir, err := filepath.Abs(legacyMediaDirName)
		if err != nil {
			t.Fatalf("abs legacy media dir: %v", err)
		}
		if resolved != expectedLegacyDir {
			t.Fatalf("expected legacy media directory %q, got %q", expectedLegacyDir, resolved)
		}
	})
}

func TestResolveWebMediaDirRespectsEnvironmentOverride(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		overrideDir := filepath.Join(tmpDir, "custom-media")
		t.Setenv(mediaDirEnvVar, overrideDir)

		resolved, err := resolveWebMediaDir()
		if err != nil {
			t.Fatalf("resolveWebMediaDir: %v", err)
		}
		if resolved != overrideDir {
			t.Fatalf("expected override media directory %q, got %q", overrideDir, resolved)
		}
	})
}

func TestListenWithPortFallbackUsesAlternatePortWhenBusy(t *testing.T) {
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen blocker: %v", err)
	}
	defer blocker.Close()

	blockedAddr := blocker.Addr().String()
	_, blockedPortText, err := net.SplitHostPort(blockedAddr)
	if err != nil {
		t.Fatalf("split blocked addr: %v", err)
	}
	blockedPort, err := strconv.Atoi(blockedPortText)
	if err != nil {
		t.Fatalf("parse blocked port: %v", err)
	}

	ln, fallbackCount, err := listenWithPortFallback(blockedAddr, 200)
	if err != nil {
		t.Fatalf("listenWithPortFallback: %v", err)
	}
	defer ln.Close()

	if fallbackCount < 1 {
		t.Fatalf("expected at least one fallback attempt, got %d", fallbackCount)
	}

	_, actualPortText, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split actual addr: %v", err)
	}
	actualPort, err := strconv.Atoi(actualPortText)
	if err != nil {
		t.Fatalf("parse actual port: %v", err)
	}

	expectedPort := blockedPort + fallbackCount
	if actualPort != expectedPort {
		t.Fatalf("expected fallback to bind port %d, got %d", expectedPort, actualPort)
	}
}

func TestListenWithPortFallbackFailsWithoutRetries(t *testing.T) {
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen blocker: %v", err)
	}
	defer blocker.Close()

	_, _, err = listenWithPortFallback(blocker.Addr().String(), 0)
	if err == nil {
		t.Fatalf("expected bind failure when fallback retries are disabled")
	}
}

func TestListenWithPortFallbackStopsAtPortUpperBound(t *testing.T) {
	const topPortAddr = "127.0.0.1:65535"

	blocker, err := net.Listen("tcp", topPortAddr)
	if err != nil {
		t.Skipf("unable to reserve %s for overflow test: %v", topPortAddr, err)
	}
	defer blocker.Close()

	_, _, err = listenWithPortFallback(topPortAddr, 10)
	if err == nil {
		t.Fatalf("expected bind failure at port upper bound")
	}
	if !strings.Contains(err.Error(), "after 1 attempt(s)") {
		t.Fatalf("expected a single bind attempt at upper bound, got: %v", err)
	}
}

func TestFormatWebURL(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{name: "wildcard ipv4", addr: "0.0.0.0:8080", want: "http://127.0.0.1:8080"},
		{name: "wildcard ipv6", addr: "[::]:8080", want: "http://127.0.0.1:8080"},
		{name: "empty host", addr: ":8080", want: "http://127.0.0.1:8080"},
		{name: "localhost", addr: "127.0.0.1:8080", want: "http://127.0.0.1:8080"},
		{name: "ipv6 zone id", addr: "[fe80::1%en0]:8080", want: "http://[fe80::1%25en0]:8080"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatWebURL(tc.addr); got != tc.want {
				t.Fatalf("formatWebURL(%q) = %q, want %q", tc.addr, got, tc.want)
			}
		})
	}
}

func TestSavedPlaylistsAPIReplaceAndLoad(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		tracker = &jobTracker{}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}

		initialResp, err := client.Get(baseURL + "/api/library/playlists")
		if err != nil {
			t.Fatalf("request initial playlists: %v", err)
		}
		defer initialResp.Body.Close()
		if initialResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for initial playlists, got %d", initialResp.StatusCode)
		}
		var initialState savedPlaylistState
		if err := json.NewDecoder(initialResp.Body).Decode(&initialState); err != nil {
			t.Fatalf("decode initial playlists: %v", err)
		}
		if len(initialState.Playlists) != 0 {
			t.Fatalf("expected no saved playlists initially, got %d", len(initialState.Playlists))
		}
		if len(initialState.Assignments) != 0 {
			t.Fatalf("expected no assignments initially, got %d", len(initialState.Assignments))
		}

		putPayload := `{
			"playlists": [
				{"id": "pl-1", "name": "  Road   Trip  ", "createdAt": "2026-02-07T10:00:00Z", "updatedAt": "2026-02-07T10:00:00Z"},
				{"id": "pl-2", "name": "Chill Mix"},
				{"id": "pl-2", "name": "Duplicate Id"},
				{"id": "pl-3", "name": "Road Trip"}
			],
			"assignments": {
				"video/road.mp4": "pl-1",
				"video/ghost.mp4": "missing",
				"": "pl-1"
			}
		}`
		putReq, err := http.NewRequest(http.MethodPut, baseURL+"/api/library/playlists", strings.NewReader(putPayload))
		if err != nil {
			t.Fatalf("new put request: %v", err)
		}
		putReq.Header.Set("Content-Type", "application/json")

		putResp, err := client.Do(putReq)
		if err != nil {
			t.Fatalf("put playlists: %v", err)
		}
		defer putResp.Body.Close()
		if putResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for put playlists, got %d", putResp.StatusCode)
		}

		var savedState savedPlaylistState
		if err := json.NewDecoder(putResp.Body).Decode(&savedState); err != nil {
			t.Fatalf("decode put response: %v", err)
		}
		if len(savedState.Playlists) != 2 {
			t.Fatalf("expected 2 normalized playlists, got %d", len(savedState.Playlists))
		}
		if savedState.Playlists[0].Name != "Road Trip" {
			t.Fatalf("expected normalized first playlist name, got %q", savedState.Playlists[0].Name)
		}
		if savedState.Playlists[1].ID != "pl-2" {
			t.Fatalf("expected second playlist id pl-2, got %q", savedState.Playlists[1].ID)
		}
		if len(savedState.Assignments) != 1 || savedState.Assignments["video/road.mp4"] != "pl-1" {
			t.Fatalf("expected filtered assignments to keep only valid mapping, got %+v", savedState.Assignments)
		}

		playlistFile := filepath.Join(tmpDir, "media", mediaFolderData, savedPlaylistsFileName)
		fileData, err := os.ReadFile(playlistFile)
		if err != nil {
			t.Fatalf("read saved playlist file: %v", err)
		}
		if len(fileData) == 0 {
			t.Fatalf("expected saved playlist file to contain data")
		}

		reloadResp, err := client.Get(baseURL + "/api/library/playlists")
		if err != nil {
			t.Fatalf("request reloaded playlists: %v", err)
		}
		defer reloadResp.Body.Close()
		if reloadResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for reloaded playlists, got %d", reloadResp.StatusCode)
		}

		var reloadedState savedPlaylistState
		if err := json.NewDecoder(reloadResp.Body).Decode(&reloadedState); err != nil {
			t.Fatalf("decode reloaded playlists: %v", err)
		}
		if len(reloadedState.Playlists) != 2 {
			t.Fatalf("expected 2 playlists after reload, got %d", len(reloadedState.Playlists))
		}
		if reloadedState.Assignments["video/road.mp4"] != "pl-1" {
			t.Fatalf("expected assignment to persist after reload, got %+v", reloadedState.Assignments)
		}
	})
}

func TestSavedPlaylistsMigrationEndpointSeedsOnlyWhenEmpty(t *testing.T) {
	withTempCWD(t, func(_ string) {
		tracker = &jobTracker{}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}
		migratePayload := `{
			"playlists": [
				{"id": "legacy-1", "name": "Legacy Mix"}
			],
			"assignments": {
				"audio/legacy.mp3": "legacy-1"
			}
		}`

		migrateReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/library/playlists/migrate", strings.NewReader(migratePayload))
		if err != nil {
			t.Fatalf("new migrate request: %v", err)
		}
		migrateReq.Header.Set("Content-Type", "application/json")

		migrateResp, err := client.Do(migrateReq)
		if err != nil {
			t.Fatalf("migrate request: %v", err)
		}
		defer migrateResp.Body.Close()
		if migrateResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for first migrate request, got %d", migrateResp.StatusCode)
		}

		var firstMigration savedPlaylistMigrationResponse
		if err := json.NewDecoder(migrateResp.Body).Decode(&firstMigration); err != nil {
			t.Fatalf("decode first migration response: %v", err)
		}
		if !firstMigration.Migrated {
			t.Fatalf("expected first migration request to migrate data")
		}
		if len(firstMigration.Playlists) != 1 || firstMigration.Playlists[0].ID != "legacy-1" {
			t.Fatalf("expected migrated playlist to be returned, got %+v", firstMigration.Playlists)
		}

		secondPayload := `{
			"playlists": [
				{"id": "legacy-2", "name": "Should Not Replace"}
			],
			"assignments": {
				"audio/new.mp3": "legacy-2"
			}
		}`
		secondReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/library/playlists/migrate", strings.NewReader(secondPayload))
		if err != nil {
			t.Fatalf("new second migrate request: %v", err)
		}
		secondReq.Header.Set("Content-Type", "application/json")

		secondResp, err := client.Do(secondReq)
		if err != nil {
			t.Fatalf("second migrate request: %v", err)
		}
		defer secondResp.Body.Close()
		if secondResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for second migrate request, got %d", secondResp.StatusCode)
		}

		var secondMigration savedPlaylistMigrationResponse
		if err := json.NewDecoder(secondResp.Body).Decode(&secondMigration); err != nil {
			t.Fatalf("decode second migration response: %v", err)
		}
		if secondMigration.Migrated {
			t.Fatalf("expected second migration request to be ignored")
		}
		if len(secondMigration.Playlists) != 1 || secondMigration.Playlists[0].ID != "legacy-1" {
			t.Fatalf("expected existing migrated data to remain unchanged, got %+v", secondMigration.Playlists)
		}
		if secondMigration.Assignments["audio/legacy.mp3"] != "legacy-1" {
			t.Fatalf("expected existing assignment to remain unchanged, got %+v", secondMigration.Assignments)
		}
	})
}

func TestSavedPlaylistsEndpointMethodAndPayloadValidation(t *testing.T) {
	withTempCWD(t, func(_ string) {
		tracker = &jobTracker{}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}

		postReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/library/playlists", strings.NewReader(`{}`))
		if err != nil {
			t.Fatalf("new invalid method request: %v", err)
		}
		postReq.Header.Set("Content-Type", "application/json")
		postResp, err := client.Do(postReq)
		if err != nil {
			t.Fatalf("invalid method request failed: %v", err)
		}
		postResp.Body.Close()
		if postResp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 for POST /api/library/playlists, got %d", postResp.StatusCode)
		}

		getMigrateResp, err := client.Get(baseURL + "/api/library/playlists/migrate")
		if err != nil {
			t.Fatalf("invalid method migrate request failed: %v", err)
		}
		getMigrateResp.Body.Close()
		if getMigrateResp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405 for GET /api/library/playlists/migrate, got %d", getMigrateResp.StatusCode)
		}

		invalidJSONReq, err := http.NewRequest(http.MethodPut, baseURL+"/api/library/playlists", strings.NewReader(`{"playlists":`))
		if err != nil {
			t.Fatalf("new invalid json request: %v", err)
		}
		invalidJSONReq.Header.Set("Content-Type", "application/json")
		invalidJSONResp, err := client.Do(invalidJSONReq)
		if err != nil {
			t.Fatalf("invalid json request failed: %v", err)
		}
		invalidJSONResp.Body.Close()
		if invalidJSONResp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid JSON payload, got %d", invalidJSONResp.StatusCode)
		}
	})
}

func TestSavedPlaylistsEndpointReturnsServerErrorForEmptyDataFile(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		tracker = &jobTracker{}

		dataDir := filepath.Join(tmpDir, "media", mediaFolderData)
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			t.Fatalf("mkdir data dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dataDir, savedPlaylistsFileName), []byte(" \n\t"), 0o644); err != nil {
			t.Fatalf("write empty playlist file: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(baseURL + "/api/library/playlists")
		if err != nil {
			t.Fatalf("request playlists: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500 for empty playlist file, got %d", resp.StatusCode)
		}
	})
}

func TestParseDownloadRequestNormalizesOutputTemplateForMediaLayout(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantOutput string
	}{
		{
			name:       "default video template when empty output",
			body:       `{"urls":["https://example.com/watch?v=abc"],"options":{"audio":false}}`,
			wantOutput: "video/{title}.{ext}",
		},
		{
			name:       "default audio template when empty output",
			body:       `{"urls":["https://example.com/watch?v=abc"],"options":{"audio":true}}`,
			wantOutput: "audio/{title}.{ext}",
		},
		{
			name:       "prefix plain template with media root",
			body:       `{"urls":["https://example.com/watch?v=abc"],"options":{"audio":false,"output":"{artist}/{title}.{ext}"}}`,
			wantOutput: "video/{artist}/{title}.{ext}",
		},
		{
			name:       "keep known media root prefix",
			body:       `{"urls":["https://example.com/watch?v=abc"],"options":{"audio":false,"output":"audio/{title}.{ext}"}}`,
			wantOutput: "audio/{title}.{ext}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/download", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			_, opts, _, reqErr := parseDownloadRequest(rec, req)
			if reqErr != nil {
				t.Fatalf("parseDownloadRequest returned error: %v", reqErr)
			}
			if opts.OutputTemplate != tt.wantOutput {
				t.Fatalf("expected output template %q, got %q", tt.wantOutput, opts.OutputTemplate)
			}
		})
	}
}

func TestMediaListIncludesSidecarMetadataAndNestedPaths(t *testing.T) {
	withTempCWD(t, func(tmpDir string) {
		tracker = &jobTracker{}

		mediaDir := filepath.Join(tmpDir, "media")
		nestedDir := filepath.Join(mediaDir, "Playlists", "Road Trip")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("mkdir nested media dir: %v", err)
		}

		nestedMediaPath := filepath.Join(nestedDir, "01 - Song.mp3")
		if err := os.WriteFile(nestedMediaPath, []byte("audio"), 0o644); err != nil {
			t.Fatalf("write nested media: %v", err)
		}
		nestedModTime := time.Now().Add(-30 * time.Minute)
		if err := os.Chtimes(nestedMediaPath, nestedModTime, nestedModTime); err != nil {
			t.Fatalf("chtimes nested media: %v", err)
		}

		nestedSidecar := `{
			"id": "vid123",
			"title": "Road Song",
			"artist": "The Drivers",
			"album": "Night Miles",
			"duration_seconds": 245,
			"release_date": "2024-01-15",
			"source_url": "https://www.youtube.com/watch?v=vid123",
			"extractor": "kkdai/youtube",
			"output": "/absolute/path/should/not/leak.mp3",
			"format": "mp3",
			"status": "ok",
			"playlist": {
				"id": "PLROAD",
				"title": "Road Trip",
				"url": "https://www.youtube.com/playlist?list=PLROAD",
				"index": 1,
				"count": 10
			}
		}`
		if err := os.WriteFile(nestedMediaPath+".json", []byte(nestedSidecar), 0o644); err != nil {
			t.Fatalf("write nested sidecar: %v", err)
		}

		rootMediaPath := filepath.Join(mediaDir, "single.mp4")
		if err := os.WriteFile(rootMediaPath, []byte("video"), 0o644); err != nil {
			t.Fatalf("write root media: %v", err)
		}
		rootModTime := time.Now().Add(-2 * time.Hour)
		if err := os.Chtimes(rootMediaPath, rootModTime, rootModTime); err != nil {
			t.Fatalf("chtimes root media: %v", err)
		}

		// Should be ignored by media listing.
		if err := os.WriteFile(filepath.Join(mediaDir, "playlist.json"), []byte(`{"id":"demo"}`), 0o644); err != nil {
			t.Fatalf("write playlist metadata artifact: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(baseURL + "/api/media/?limit=10")
		if err != nil {
			t.Fatalf("request media list: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for media list, got %d", resp.StatusCode)
		}

		var payload mediaListResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode media list: %v", err)
		}

		if len(payload.Items) != 2 {
			t.Fatalf("expected 2 media items (excluding json artifacts), got %d", len(payload.Items))
		}
		if payload.NextOffset != nil {
			t.Fatalf("expected no next_offset for complete page, got %v", *payload.NextOffset)
		}

		nestedRelPath := filepath.ToSlash(filepath.Join("Playlists", "Road Trip", "01 - Song.mp3"))
		itemsByFilename := make(map[string]mediaItem, len(payload.Items))
		for _, item := range payload.Items {
			itemsByFilename[item.Filename] = item
		}

		nestedItem, ok := itemsByFilename[nestedRelPath]
		if !ok {
			t.Fatalf("expected nested media item %q in response", nestedRelPath)
		}
		if !nestedItem.HasSidecar {
			t.Fatalf("expected nested item to report sidecar metadata")
		}
		if nestedItem.Title != "Road Song" {
			t.Fatalf("expected sidecar title, got %q", nestedItem.Title)
		}
		if nestedItem.Artist != "The Drivers" {
			t.Fatalf("expected sidecar artist, got %q", nestedItem.Artist)
		}
		if nestedItem.Album != "Night Miles" {
			t.Fatalf("expected sidecar album, got %q", nestedItem.Album)
		}
		if nestedItem.DurationSeconds != 245 {
			t.Fatalf("expected sidecar duration, got %d", nestedItem.DurationSeconds)
		}
		if nestedItem.Date != "2024-01-15" {
			t.Fatalf("expected release date from sidecar, got %q", nestedItem.Date)
		}
		if nestedItem.Type != "audio" {
			t.Fatalf("expected audio type for mp3, got %q", nestedItem.Type)
		}
		if nestedItem.RelativePath != nestedRelPath {
			t.Fatalf("expected relative path %q, got %q", nestedRelPath, nestedItem.RelativePath)
		}
		if nestedItem.Folder != filepath.ToSlash(filepath.Join("Playlists", "Road Trip")) {
			t.Fatalf("expected folder path to be preserved, got %q", nestedItem.Folder)
		}
		if nestedItem.Metadata.Output != nestedRelPath {
			t.Fatalf("expected metadata output to be normalized to relative path, got %q", nestedItem.Metadata.Output)
		}
		if nestedItem.Playlist == nil || nestedItem.Playlist.ID != "PLROAD" {
			t.Fatalf("expected playlist metadata from sidecar, got %+v", nestedItem.Playlist)
		}

		rootItem, ok := itemsByFilename["single.mp4"]
		if !ok {
			t.Fatalf("expected root media item in response")
		}
		if rootItem.HasSidecar {
			t.Fatalf("expected root item without sidecar to report has_sidecar=false")
		}
		if rootItem.Artist != "Unknown Artist" {
			t.Fatalf("expected fallback artist for root item, got %q", rootItem.Artist)
		}
		if rootItem.Metadata.Extractor != "library" {
			t.Fatalf("expected fallback extractor for root item, got %q", rootItem.Metadata.Extractor)
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

func TestDownloadProgressStreamIncludesSnapshotAndSequencedEvents(t *testing.T) {
	withTempCWD(t, func(_ string) {
		tracker = &jobTracker{}
		job := tracker.Create([]string{"https://example.com/watch?v=abc"})
		defer job.CloseEvents()

		if !job.enqueueCriticalEvent(ProgressEvent{
			Type:  "register",
			ID:    "task_1",
			Label: "Track 1",
			Total: 100,
		}, time.Second) {
			t.Fatalf("failed to enqueue register event")
		}
		if !job.enqueueCriticalEvent(ProgressEvent{
			Type:    "progress",
			ID:      "task_1",
			Current: 10,
			Total:   100,
		}, time.Second) {
			t.Fatalf("failed to enqueue progress event")
		}

		ctx, cancel := context.WithCancel(context.Background())
		baseURL, wait := startWebServerForTest(t, ctx)
		defer func() {
			cancel()
			wait()
		}()

		req, err := http.NewRequest(http.MethodGet, baseURL+"/api/download/progress?id="+job.ID, nil)
		if err != nil {
			t.Fatalf("new request for progress stream: %v", err)
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request progress stream: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from progress stream, got %d", resp.StatusCode)
		}

		reader := bufio.NewReader(resp.Body)
		readSSE := func() (idLine string, evt ProgressEvent, err error) {
			var dataLine string
			for {
				line, readErr := reader.ReadString('\n')
				if readErr != nil {
					return "", ProgressEvent{}, readErr
				}
				trimmed := strings.TrimRight(line, "\r\n")
				if trimmed == "" {
					break
				}
				if strings.HasPrefix(trimmed, "id: ") {
					idLine = strings.TrimPrefix(trimmed, "id: ")
				}
				if strings.HasPrefix(trimmed, "data: ") {
					dataLine = strings.TrimPrefix(trimmed, "data: ")
				}
			}
			if dataLine == "" {
				return idLine, ProgressEvent{}, nil
			}
			if unmarshalErr := json.Unmarshal([]byte(dataLine), &evt); unmarshalErr != nil {
				return idLine, ProgressEvent{}, unmarshalErr
			}
			return idLine, evt, nil
		}

		sawSnapshot := false
		sawSnapshotID := false
		sawSequenced := false
		for i := 0; i < 8 && (!sawSnapshot || !sawSequenced); i++ {
			idLine, evt, err := readSSE()
			if err != nil {
				t.Fatalf("read SSE event: %v", err)
			}
			if evt.Type == "" {
				continue
			}
			if evt.Type == "snapshot" {
				sawSnapshot = true
				if idLine == "0" {
					sawSnapshotID = true
				}
				if evt.Snapshot == nil {
					t.Fatalf("snapshot event missing snapshot payload")
				}
				if evt.Snapshot.JobID != job.ID {
					t.Fatalf("expected snapshot job id %q, got %q", job.ID, evt.Snapshot.JobID)
				}
				continue
			}
			if idLine != "" {
				sawSequenced = true
				if evt.Seq <= 0 {
					t.Fatalf("expected positive sequence for event %q, got %d", evt.Type, evt.Seq)
				}
				if evt.JobID != job.ID {
					t.Fatalf("expected event job id %q, got %q", job.ID, evt.JobID)
				}
			}
		}

		if !sawSnapshot {
			t.Fatalf("expected snapshot event in SSE stream")
		}
		if !sawSnapshotID {
			t.Fatalf("expected snapshot event to include SSE id")
		}
		if !sawSequenced {
			t.Fatalf("expected at least one sequenced replay/live event with SSE id")
		}
	})
}
