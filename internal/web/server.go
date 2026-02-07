package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
)

//go:embed assets/*
var embeddedAssets embed.FS

const maxRequestBodyBytes = 1 << 20 // 1 MiB

const (
	defaultMediaListLimit = 200
	maxMediaListLimit     = 500
	jobCompletedTTL       = 15 * time.Minute
	jobErroredTTL         = 30 * time.Minute
	jobCleanupInterval    = time.Minute
)

var audioMediaExtensions = map[string]struct{}{
	".aac":  {},
	".alac": {},
	".flac": {},
	".m4a":  {},
	".mp3":  {},
	".oga":  {},
	".ogg":  {},
	".opus": {},
	".wav":  {},
	".wma":  {},
}

var videoMediaExtensions = map[string]struct{}{
	".3gp":  {},
	".avi":  {},
	".m4v":  {},
	".mkv":  {},
	".mov":  {},
	".mp4":  {},
	".mpeg": {},
	".mpg":  {},
	".ogv":  {},
	".ts":   {},
	".webm": {},
	".wmv":  {},
}

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
	OnDuplicate         string            `json:"on-duplicate"`
}

type DuplicateResponseRequest struct {
	JobID    string `json:"jobId"`
	PromptID string `json:"promptId"`
	Choice   string `json:"choice"`
}

type mediaItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Size     string `json:"size"`
	Date     string `json:"date"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
}

type mediaListResponse struct {
	Items      []mediaItem `json:"items"`
	NextOffset *int        `json:"next_offset"`
}

// formatBytes formats a byte size into a human-readable string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
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

type requestError struct {
	status  int
	message string
}

func (e *requestError) Error() string { return e.message }

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) *requestError {
	ct := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil || mediaType != "application/json" {
		return &requestError{http.StatusUnsupportedMediaType, "content type must be application/json"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return &requestError{http.StatusRequestEntityTooLarge, "request body too large"}
		}
		return &requestError{http.StatusBadRequest, "invalid JSON payload"}
	}
	if err := dec.Decode(new(struct{})); err != io.EOF {
		return &requestError{http.StatusBadRequest, "invalid JSON payload"}
	}
	return nil
}

func parseDownloadRequest(w http.ResponseWriter, r *http.Request) (*DownloadRequest, downloader.Options, int, *requestError) {
	var req DownloadRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		return nil, downloader.Options{}, 0, err
	}
	if len(req.URLs) == 0 {
		return nil, downloader.Options{}, 0, &requestError{http.StatusBadRequest, "no urls provided"}
	}
	onDuplicate, err := downloader.ParseDuplicatePolicy(req.Options.OnDuplicate)
	if err != nil {
		return nil, downloader.Options{}, 0, &requestError{http.StatusBadRequest, err.Error()}
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
		OnDuplicate:         onDuplicate,
	}
	if err := validateWebOutputTemplate(opts.OutputTemplate); err != nil {
		return nil, downloader.Options{}, 0, &requestError{http.StatusBadRequest, err.Error()}
	}
	if opts.OutputTemplate == "" {
		opts.OutputTemplate = "{title}.{ext}"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 3 * time.Minute
	}

	return &req, opts, req.Options.Jobs, nil
}

func ListenAndServe(ctx context.Context, addr string) error {
	startedAt := time.Now()

	// Create a dedicated media directory for downloads.
	mediaDir, err := filepath.Abs("media")
	if err != nil {
		return fmt.Errorf("resolving media directory: %w", err)
	}
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		return fmt.Errorf("creating media directory: %w", err)
	}
	log.Printf("Media directory: %s", mediaDir)
	tracker.StartCleanup(ctx, jobCleanupInterval, jobCompletedTTL, jobErroredTTL)

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
		req, opts, jobs, err := parseDownloadRequest(w, r)
		if err != nil {
			writeJSONError(w, err.status, err.message)
			return
		}
		opts.OutputDir = mediaDir
		opts.Quiet = true

		job := tracker.Create(req.URLs)
		renderer := &webRenderer{events: job.Events}
		opts.Renderer = renderer
		opts.DuplicatePrompter = newWebDuplicatePrompter(ctx, job)

		go func() {
			job.SetStatus("running")
			results, exitCode := app.Run(ctx, req.URLs, opts, jobs)
			job.closeDuplicatePrompts(downloader.DuplicateDecisionSkip)
			status := job.SetOutcome(results, exitCode)
			snapshot := job.progressSnapshot()
			doneEvt := ProgressEvent{
				Type:     "done",
				Status:   status,
				Message:  status,
				ExitCode: snapshot.ExitCode,
				Error:    snapshot.Error,
			}
			if snapshot.Stats.Total > 0 {
				stats := snapshot.Stats
				doneEvt.Stats = &stats
			}
			_ = job.enqueueCriticalEvent(doneEvt, criticalEventTimeout)
			job.CloseEvents()
		}()

		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "queued",
			"jobId":   job.ID,
			"message": fmt.Sprintf("Download started for %d item(s).", len(req.URLs)),
		})
	})

	mux.HandleFunc("/api/download/duplicate-response", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req DuplicateResponseRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			writeJSONError(w, err.status, err.message)
			return
		}
		if req.JobID == "" || req.PromptID == "" || req.Choice == "" {
			writeJSONError(w, http.StatusBadRequest, "jobId, promptId, and choice are required")
			return
		}
		job, ok := tracker.Get(req.JobID)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "job not found")
			return
		}
		decision, err := downloader.ParseDuplicateDecision(req.Choice)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := job.resolveDuplicatePrompt(req.PromptID, decision); err != nil {
			if errors.Is(err, errDuplicatePromptNotFound) {
				writeJSONError(w, http.StatusNotFound, "duplicate prompt not found")
				return
			}
			if errors.Is(err, errDuplicatePromptClosed) {
				writeJSONError(w, http.StatusConflict, "duplicate prompt has closed")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, "failed to resolve duplicate prompt")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/download/progress", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		jobID := r.URL.Query().Get("id")
		if jobID == "" {
			writeJSONError(w, http.StatusBadRequest, "missing id parameter")
			return
		}
		job, ok := tracker.Get(jobID)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "job not found")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSONError(w, http.StatusInternalServerError, "streaming not supported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher.Flush()

		afterSeq := int64(0)
		if seq, ok := parseProgressSeq(r.URL.Query().Get("since")); ok {
			afterSeq = seq
		} else if seq, ok := parseProgressSeq(r.Header.Get("Last-Event-ID")); ok {
			afterSeq = seq
		}

		stream, cancel := job.Subscribe(afterSeq)
		defer cancel()

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)

		for {
			select {
			case <-r.Context().Done():
				return
			case evt, ok := <-stream:
				if !ok {
					doneEvt := ProgressEvent{
						Type:    "done",
						Status:  job.StatusValue(),
						Message: job.StatusValue(),
					}
					snapshot := job.progressSnapshot()
					doneEvt.ExitCode = snapshot.ExitCode
					doneEvt.Error = snapshot.Error
					if snapshot.Stats.Total > 0 {
						stats := snapshot.Stats
						doneEvt.Stats = &stats
					}
					writeSSEEvent(w, flusher, enc, doneEvt)
					return
				}
				writeSSEEvent(w, flusher, enc, evt)
			}
		}
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		uptime := time.Since(startedAt).Truncate(time.Second).String()
		writeJSON(w, http.StatusOK, map[string]any{
			"active_downloads": tracker.ActiveCount(),
			"uptime":           uptime,
		})
	})

	mux.HandleFunc("/api/media/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		reqPath := strings.TrimPrefix(r.URL.Path, "/api/media/")

		// If no filename provided, list all media files
		if reqPath == "" {
			offset, limit, err := parseMediaListPagination(r)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, err.Error())
				return
			}

			allItems, err := listMediaFiles(mediaDir)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to read media directory")
				return
			}

			items, nextOffset := paginateMediaItems(allItems, offset, limit)
			writeJSON(w, http.StatusOK, mediaListResponse{
				Items:      items,
				NextOffset: nextOffset,
			})
			return
		}

		fullPath, status, err := resolveMediaPath(mediaDir, reqPath)
		if err != nil {
			if status == 0 {
				status = http.StatusBadRequest
			}
			writeJSONError(w, status, err.Error())
			return
		}
		http.ServeFile(w, r, fullPath)
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
		Handler:           withSecurityHeaders(mux),
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

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, enc *json.Encoder, evt ProgressEvent) {
	if evt.Seq > 0 || evt.Type == "snapshot" {
		fmt.Fprintf(w, "id: %d\n", evt.Seq)
	}
	fmt.Fprintf(w, "data: ")
	_ = enc.Encode(evt)
	fmt.Fprintf(w, "\n")
	flusher.Flush()
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

func withSecurityHeaders(next http.Handler) http.Handler {
	const cspValue = "default-src 'self'; base-uri 'self'; frame-ancestors 'none'; object-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; media-src 'self'"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", cspValue)
		next.ServeHTTP(w, r)
	})
}

func parseMediaListPagination(r *http.Request) (offset int, limit int, err error) {
	offset = 0
	limit = defaultMediaListLimit

	q := r.URL.Query()
	if rawOffset := q.Get("offset"); rawOffset != "" {
		parsed, parseErr := strconv.Atoi(rawOffset)
		if parseErr != nil || parsed < 0 {
			return 0, 0, fmt.Errorf("invalid offset parameter")
		}
		offset = parsed
	}
	if rawLimit := q.Get("limit"); rawLimit != "" {
		parsed, parseErr := strconv.Atoi(rawLimit)
		if parseErr != nil || parsed <= 0 {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
		if parsed > maxMediaListLimit {
			parsed = maxMediaListLimit
		}
		limit = parsed
	}
	return offset, limit, nil
}

func listMediaFiles(mediaDir string) ([]mediaItem, error) {
	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		return nil, err
	}

	type enrichedMediaItem struct {
		item    mediaItem
		modTime time.Time
	}

	items := make([]enrichedMediaItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			log.Printf("skipping media entry %q: %v", entry.Name(), infoErr)
			continue
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		mediaType := mediaTypeForExtension(ext)

		title := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		items = append(items, enrichedMediaItem{
			item: mediaItem{
				ID:       info.Name(),
				Title:    title,
				Artist:   "Unknown Artist",
				Size:     formatBytes(info.Size()),
				Date:     info.ModTime().Format("2006-01-02"),
				Type:     mediaType,
				Filename: info.Name(),
			},
			modTime: info.ModTime(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].modTime.Equal(items[j].modTime) {
			return items[i].item.Filename < items[j].item.Filename
		}
		return items[i].modTime.After(items[j].modTime)
	})

	out := make([]mediaItem, 0, len(items))
	for _, item := range items {
		out = append(out, item.item)
	}
	return out, nil
}

func mediaTypeForExtension(ext string) string {
	ext = strings.ToLower(ext)
	if _, ok := audioMediaExtensions[ext]; ok {
		return "audio"
	}
	if _, ok := videoMediaExtensions[ext]; ok {
		return "video"
	}
	return "video"
}

func paginateMediaItems(items []mediaItem, offset int, limit int) ([]mediaItem, *int) {
	total := len(items)
	if offset >= total {
		return []mediaItem{}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	page := append([]mediaItem(nil), items[offset:end]...)
	if end >= total {
		return page, nil
	}

	next := end
	return page, &next
}

func resolveMediaPath(mediaDir, reqPath string) (string, int, error) {
	cleaned := filepath.Clean(reqPath)
	if cleaned == "." || cleaned == "" {
		return "", http.StatusBadRequest, fmt.Errorf("invalid path")
	}
	if strings.Contains(cleaned, "..") || filepath.IsAbs(cleaned) {
		return "", http.StatusBadRequest, fmt.Errorf("invalid path")
	}

	fullPath := filepath.Join(mediaDir, cleaned)
	realMediaDir, err := resolveRealPath(mediaDir)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to resolve media directory")
	}
	realTargetPath, err := resolveRealPath(fullPath)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("invalid path")
	}
	rel, err := filepath.Rel(realMediaDir, realTargetPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", http.StatusForbidden, fmt.Errorf("access denied")
	}
	return fullPath, 0, nil
}

func resolveRealPath(path string) (string, error) {
	cleaned := filepath.Clean(path)
	realPath, err := filepath.EvalSymlinks(cleaned)
	if err == nil {
		return realPath, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	parent := filepath.Dir(cleaned)
	if parent == cleaned {
		return "", err
	}

	realParent, parentErr := resolveRealPath(parent)
	if parentErr != nil {
		return "", parentErr
	}
	return filepath.Join(realParent, filepath.Base(cleaned)), nil
}
