package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
)

//go:embed assets/*
var embeddedAssets embed.FS

const maxRequestBodyBytes = 1 << 20 // 1 MiB

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

// validateWebOutputTemplate ensures that the output template provided via the web API
// cannot be used to write files outside the intended output area. It is intentionally
// conservative and rejects absolute paths, parent directory references, and directory
// components in the literal prefix before the first placeholder.
func validateWebOutputTemplate(tmpl string) error {
	// Empty template is allowed; it will be defaulted later.
	if strings.TrimSpace(tmpl) == "" {
		return nil
	}

	// Reject simple parent-directory traversal patterns.
	if strings.Contains(tmpl, "..") {
		return fmt.Errorf("invalid output template: parent directory references are not allowed")
	}

	// Basic absolute path checks for Unix-like and Windows-style paths.
	if strings.HasPrefix(tmpl, "/") || strings.HasPrefix(tmpl, `\\`) {
		return fmt.Errorf("invalid output template: absolute paths are not allowed")
	}
	if len(tmpl) >= 2 && ((tmpl[0] >= 'A' && tmpl[0] <= 'Z') || (tmpl[0] >= 'a' && tmpl[0] <= 'z')) && tmpl[1] == ':' {
		return fmt.Errorf("invalid output template: absolute paths are not allowed")
	}

	// Disallow explicit directory components in the literal prefix before the first placeholder.
	prefixEnd := strings.Index(tmpl, "{")
	if prefixEnd == -1 {
		prefixEnd = len(tmpl)
	}
	literalPrefix := tmpl[:prefixEnd]
	if strings.Contains(literalPrefix, "/") || strings.Contains(literalPrefix, `\`) {
		return fmt.Errorf("invalid output template: directory components are not allowed in the literal prefix")
	}

	return nil
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
		ct := r.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "application/json" {
			writeJSONError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		var req DownloadRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		// Ensure no trailing data beyond the first JSON object
		if err := dec.Decode(new(struct{})); err != io.EOF {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}
		if len(req.URLs) == 0 {
			writeJSONError(w, http.StatusBadRequest, "no urls provided")
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
		// Validate web-provided output template to avoid writing files to arbitrary locations.
		if err := validateWebOutputTemplate(opts.OutputTemplate); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if opts.OutputTemplate == "" {
			opts.OutputTemplate = "{title}.{ext}"
		}
		if opts.Timeout == 0 {
			opts.Timeout = 3 * time.Minute
		}
		opts.Quiet = true

		results, exitCode := app.Run(ctx, req.URLs, opts, req.Options.Jobs)
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
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
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
