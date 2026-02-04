package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
)

//go:embed assets/*
var embeddedAssets embed.FS

type DownloadRequest struct {
	URLs    []string  `json:"urls"`
	Options WebOption `json:"options"`
}

type WebOption struct {
	Output              string            `json:"output"`
	Audio               bool              `json:"audio"`
	Info                bool              `json:"info"`
	ListFormats         bool              `json:"list-formats"`
	Quality             string            `json:"quality"`
	Format              string            `json:"format"`
	Itag                int               `json:"itag"`
	Meta                map[string]string `json:"meta"`
	ProgressLayout      string            `json:"progress-layout"`
	SegmentConcurrency  int               `json:"segment-concurrency"`
	PlaylistConcurrency int               `json:"playlist-concurrency"`
	Jobs                int               `json:"jobs"`
	JSON                bool              `json:"json"`
	TimeoutSeconds      int               `json:"timeout"`
	Quiet               bool              `json:"quiet"`
	LogLevel            string            `json:"log-level"`
}

type DownloadResponse struct {
	Type     string             `json:"type"`
	Status   string             `json:"status"`
	Results  []app.Result       `json:"results,omitempty"`
	ExitCode int                `json:"exit_code,omitempty"`
	Error    string             `json:"error,omitempty"`
	Options  downloader.Options `json:"options,omitempty"`
}

func ListenAndServe(ctx context.Context, addr string) error {
	assets, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(assets))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		// Limit request body size to 1MB to prevent resource exhaustion
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
		var req DownloadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		if len(req.URLs) == 0 {
			writeJSONError(w, http.StatusBadRequest, "no urls provided")
			return
		}

		// Validate URLs for format and length
		for _, url := range req.URLs {
			if len(url) > 2048 {
				writeJSONError(w, http.StatusBadRequest, "url exceeds maximum length of 2048 characters")
				return
			}
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				writeJSONError(w, http.StatusBadRequest, "invalid url scheme: must be http:// or https://")
				return
			}
		}

		// Validate integer parameters to prevent negative or extremely large values
		if !validateIntRange(w, req.Options.SegmentConcurrency, 0, 100, "segment-concurrency") {
			return
		}
		if !validateIntRange(w, req.Options.PlaylistConcurrency, 0, 100, "playlist-concurrency") {
			return
		}
		if !validateIntRange(w, req.Options.Itag, 0, 1000000, "itag") {
			return
		}
		if !validateIntRange(w, req.Options.TimeoutSeconds, 0, 86400, "timeout-seconds") {
			return
		}
		if !validateIntRange(w, req.Options.Jobs, 0, 100, "jobs") {
			return
		}

		// Validate string parameters (Quality and Format)
		if !validateStringParam(w, req.Options.Quality, "quality") {
			return
		}
		if !validateStringParam(w, req.Options.Format, "format") {
			return
		}

		// Validate log-level parameter
		validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
		if req.Options.LogLevel != "" && !validLogLevels[req.Options.LogLevel] {
			writeJSONError(w, http.StatusBadRequest, "invalid log-level: must be debug, info, warn, or error")
			return
		}

		opts := downloader.Options{
			OutputTemplate:      req.Options.Output,
			AudioOnly:           req.Options.Audio,
			InfoOnly:            req.Options.Info,
			ListFormats:         req.Options.ListFormats,
			Quality:             req.Options.Quality,
			Format:              req.Options.Format,
			Itag:                req.Options.Itag,
			MetaOverrides:       req.Options.Meta,
			ProgressLayout:      req.Options.ProgressLayout,
			SegmentConcurrency:  req.Options.SegmentConcurrency,
			PlaylistConcurrency: req.Options.PlaylistConcurrency,
			JSON:                req.Options.JSON,
			Timeout:             time.Duration(req.Options.TimeoutSeconds) * time.Second,
			Quiet:               req.Options.Quiet,
			LogLevel:            req.Options.LogLevel,
		}
		if opts.OutputTemplate == "" {
			opts.OutputTemplate = "{title}.{ext}"
		}
		if opts.Timeout == 0 {
			opts.Timeout = 3 * time.Minute
		}

		// Create a separate context for the download operation. Use the client's
		// request context as the parent so downloads are cancelled if the client
		// disconnects, but set a timeout based on the user's specified timeout
		// (capped at 30 minutes for long-running downloads).
		downloadTimeout := 30 * time.Minute
		if opts.Timeout > 0 && opts.Timeout < downloadTimeout {
			downloadTimeout = opts.Timeout
		}
		downloadCtx, cancel := context.WithTimeout(r.Context(), downloadTimeout)
		defer cancel()

		// Note: This endpoint blocks until all downloads complete before sending
		// a response. For large files or multiple URLs, this can take significant
		// time. The WriteTimeout is set to 30 minutes to accommodate long-running
		// downloads. Consider this a synchronous, long-running operation.
		results, exitCode := app.Run(downloadCtx, req.URLs, opts, req.Options.Jobs)
		payload := DownloadResponse{
			Type:     "download",
			Status:   "ok",
			Results:  results,
			ExitCode: exitCode,
			Options:  opts,
		}
		writeJSON(w, http.StatusOK, payload)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			serveIndex(w, assets)
			return
		}
		if fileExists(assets, strings.TrimPrefix(r.URL.Path, "/")) {
			fileServer.ServeHTTP(w, r)
			return
		}
		serveIndex(w, assets)
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Minute,
	}

	// Create a listener to ensure we can log after successfully binding
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Web server listening on http://%s\n", addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	payload := DownloadResponse{
		Type:   "error",
		Status: "error",
		Error:  message,
	}
	writeJSON(w, status, payload)
}

func validateIntRange(w http.ResponseWriter, value int, min, max int, paramName string) bool {
	if value < min || value > max {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("%s must be between %d and %d", paramName, min, max))
		return false
	}
	return true
}

func validateStringParam(w http.ResponseWriter, value string, paramName string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		// Empty is allowed and signals "use default" behavior downstream
		return true
	}
	if len(trimmed) > 256 {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("%s exceeds maximum length of 256 characters", paramName))
		return false
	}
	// Reject control characters (non-printable ASCII)
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid characters in %s", paramName))
			return false
		}
	}
	return true
}

func serveIndex(w http.ResponseWriter, assets fs.FS) {
	data, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		http.Error(w, "missing index", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func fileExists(assets fs.FS, name string) bool {
	if name == "" {
		return false
	}
	f, err := assets.Open(name)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
