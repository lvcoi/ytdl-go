package web

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
)

// ProgressEvent is a structured event sent over SSE.
type ProgressEvent struct {
	Type     string  `json:"type"`
	ID       string  `json:"id,omitempty"`
	Label    string  `json:"label,omitempty"`
	Current  int64   `json:"current,omitempty"`
	Total    int64   `json:"total,omitempty"`
	Percent  float64 `json:"percent,omitempty"`
	Level    string  `json:"level,omitempty"`
	Message  string  `json:"message,omitempty"`
	PromptID string  `json:"promptId,omitempty"`
	Path     string  `json:"path,omitempty"`
	Filename string  `json:"filename,omitempty"`
}

// Job represents an async download job.
type Job struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
	URLs        []string           `json:"urls"`
	CreatedAt   time.Time          `json:"created_at"`
	Events      chan ProgressEvent `json:"-"`
	Results     []app.Result       `json:"results,omitempty"`
	ExitCode    int                `json:"exit_code,omitempty"`
	Error       string             `json:"error,omitempty"`
	CompletedAt time.Time          `json:"completed_at,omitempty"`

	promptCounter atomic.Int64                                 `json:"-"`
	promptMu      sync.Mutex                                   `json:"-"`
	promptClosed  bool                                         `json:"-"`
	pendingPrompt map[string]chan downloader.DuplicateDecision `json:"-"`
	mu            sync.RWMutex                                 `json:"-"`
}

// jobTracker manages active download jobs.
type jobTracker struct {
	jobs    sync.Map
	counter atomic.Int64
}

var tracker = &jobTracker{}

var (
	errDuplicatePromptNotFound = errors.New("duplicate prompt not found")
	errDuplicatePromptClosed   = errors.New("duplicate prompt subsystem closed")
)

const (
	criticalEventTimeout   = 2 * time.Second
	duplicatePromptTimeout = 120 * time.Second
)

func (jt *jobTracker) Create(urls []string) *Job {
	id := fmt.Sprintf("job_%d", jt.counter.Add(1))
	urlCopy := append([]string(nil), urls...)
	job := &Job{
		ID:            id,
		Status:        "queued",
		URLs:          urlCopy,
		CreatedAt:     time.Now(),
		Events:        make(chan ProgressEvent, 256),
		pendingPrompt: make(map[string]chan downloader.DuplicateDecision),
	}
	jt.jobs.Store(id, job)
	return job
}

func (jt *jobTracker) Get(id string) (*Job, bool) {
	v, ok := jt.jobs.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Job), true
}

func (jt *jobTracker) ActiveCount() int {
	count := 0
	jt.jobs.Range(func(_, v any) bool {
		if j, ok := v.(*Job); ok && j.isActive() {
			count++
		}
		return true
	})
	return count
}

func (jt *jobTracker) Delete(id string) {
	jt.jobs.Delete(id)
}

func (jt *jobTracker) RemoveExpired(now time.Time, completedTTL, erroredTTL time.Duration) int {
	removed := 0
	jt.jobs.Range(func(key, value any) bool {
		id, ok := key.(string)
		if !ok {
			return true
		}
		job, ok := value.(*Job)
		if !ok {
			return true
		}
		if job.isExpired(now, completedTTL, erroredTTL) {
			jt.jobs.Delete(id)
			removed++
		}
		return true
	})
	return removed
}

func (jt *jobTracker) StartCleanup(ctx context.Context, interval, completedTTL, erroredTTL time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				jt.RemoveExpired(now, completedTTL, erroredTTL)
			}
		}
	}()
}

func (j *Job) isActive() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status == "queued" || j.Status == "running"
}

func (j *Job) StatusValue() string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status
}

func (j *Job) setTerminalStatusLocked(status string) {
	j.Status = status
	if status == "complete" || status == "error" {
		j.CompletedAt = time.Now()
		return
	}
	j.CompletedAt = time.Time{}
}

func (j *Job) SetStatus(status string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.setTerminalStatusLocked(status)
}

func (j *Job) SetOutcome(results []app.Result, exitCode int) string {
	resultsCopy := append([]app.Result(nil), results...)

	j.mu.Lock()
	defer j.mu.Unlock()

	j.Results = resultsCopy
	j.ExitCode = exitCode
	j.Error = ""

	if exitCode != 0 {
		j.setTerminalStatusLocked("error")
		for _, result := range resultsCopy {
			if result.Error != "" {
				j.Error = result.Error
				break
			}
		}
		return j.Status
	}

	j.setTerminalStatusLocked("complete")
	return j.Status
}

func (j *Job) isExpired(now time.Time, completedTTL, erroredTTL time.Duration) bool {
	j.mu.RLock()
	status := j.Status
	completedAt := j.CompletedAt
	j.mu.RUnlock()

	if completedAt.IsZero() {
		return false
	}
	switch status {
	case "complete":
		if completedTTL <= 0 {
			return false
		}
		return now.Sub(completedAt) > completedTTL
	case "error":
		if erroredTTL <= 0 {
			return false
		}
		return now.Sub(completedAt) > erroredTTL
	default:
		return false
	}
}

func (j *Job) enqueueCriticalEvent(evt ProgressEvent, maxWait time.Duration) (ok bool) {
	if j == nil || j.Events == nil {
		return false
	}
	if maxWait <= 0 {
		maxWait = criticalEventTimeout
	}
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	timer := time.NewTimer(maxWait)
	defer timer.Stop()
	select {
	case j.Events <- evt:
		return true
	case <-timer.C:
		return false
	}
}

func (j *Job) registerDuplicatePrompt() (string, chan downloader.DuplicateDecision, error) {
	if j == nil {
		return "", nil, errDuplicatePromptClosed
	}
	j.promptMu.Lock()
	defer j.promptMu.Unlock()

	if j.promptClosed {
		return "", nil, errDuplicatePromptClosed
	}
	if j.pendingPrompt == nil {
		j.pendingPrompt = make(map[string]chan downloader.DuplicateDecision)
	}
	promptID := fmt.Sprintf("dup_%d", j.promptCounter.Add(1))
	resp := make(chan downloader.DuplicateDecision, 1)
	j.pendingPrompt[promptID] = resp
	return promptID, resp, nil
}

func (j *Job) popDuplicatePrompt(promptID string) (chan downloader.DuplicateDecision, bool) {
	if j == nil {
		return nil, false
	}
	j.promptMu.Lock()
	defer j.promptMu.Unlock()

	resp, ok := j.pendingPrompt[promptID]
	if !ok {
		return nil, false
	}
	delete(j.pendingPrompt, promptID)
	return resp, true
}

func (j *Job) resolveDuplicatePrompt(promptID string, decision downloader.DuplicateDecision) error {
	if j == nil {
		return errDuplicatePromptClosed
	}
	j.promptMu.Lock()
	resp, ok := j.pendingPrompt[promptID]
	if ok {
		delete(j.pendingPrompt, promptID)
	}
	closed := j.promptClosed
	j.promptMu.Unlock()
	if !ok {
		if closed {
			return errDuplicatePromptClosed
		}
		return errDuplicatePromptNotFound
	}
	select {
	case resp <- decision:
	default:
	}
	close(resp)
	return nil
}

func (j *Job) expireDuplicatePrompt(promptID string) {
	resp, ok := j.popDuplicatePrompt(promptID)
	if !ok {
		return
	}
	close(resp)
}

func (j *Job) closeDuplicatePrompts(defaultDecision downloader.DuplicateDecision) {
	if j == nil {
		return
	}
	j.promptMu.Lock()
	if j.promptClosed {
		j.promptMu.Unlock()
		return
	}
	j.promptClosed = true
	pending := j.pendingPrompt
	j.pendingPrompt = make(map[string]chan downloader.DuplicateDecision)
	j.promptMu.Unlock()

	for _, resp := range pending {
		if defaultDecision != "" {
			select {
			case resp <- defaultDecision:
			default:
			}
		}
		close(resp)
	}
}

type webDuplicatePrompter struct {
	ctx     context.Context
	job     *Job
	timeout time.Duration
}

func newWebDuplicatePrompter(ctx context.Context, job *Job) *webDuplicatePrompter {
	return &webDuplicatePrompter{
		ctx:     ctx,
		job:     job,
		timeout: duplicatePromptTimeout,
	}
}

func (p *webDuplicatePrompter) PromptDuplicate(path string) (downloader.DuplicateDecision, error) {
	if p == nil || p.job == nil {
		return downloader.DuplicateDecisionSkip, nil
	}

	promptID, response, err := p.job.registerDuplicatePrompt()
	if err != nil {
		return downloader.DuplicateDecisionSkip, nil
	}

	event := ProgressEvent{
		Type:     "duplicate",
		PromptID: promptID,
		Path:     path,
		Filename: filepath.Base(path),
	}
	if !p.job.enqueueCriticalEvent(event, criticalEventTimeout) {
		p.job.expireDuplicatePrompt(promptID)
		return downloader.DuplicateDecisionSkip, nil
	}

	timeout := p.timeout
	if timeout <= 0 {
		timeout = duplicatePromptTimeout
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case decision, ok := <-response:
		if !ok || decision == "" {
			return downloader.DuplicateDecisionSkip, nil
		}
		return decision, nil
	case <-p.ctx.Done():
		p.job.expireDuplicatePrompt(promptID)
		return downloader.DuplicateDecisionSkip, nil
	case <-timer.C:
		p.job.expireDuplicatePrompt(promptID)
		return downloader.DuplicateDecisionSkip, nil
	}
}

var _ downloader.DuplicatePrompter = (*webDuplicatePrompter)(nil)

// webRenderer implements downloader.ProgressRenderer for SSE streaming.
type webRenderer struct {
	events chan<- ProgressEvent
}

func (w *webRenderer) Register(prefix string, size int64) string {
	id := fmt.Sprintf("%s@%d", prefix, time.Now().UnixNano())
	select {
	case w.events <- ProgressEvent{Type: "register", ID: id, Label: prefix, Total: size}:
	default:
	}
	return id
}

func (w *webRenderer) Update(id string, current, total int64) {
	percent := 0.0
	if total > 0 {
		percent = float64(current) * 100 / float64(total)
	}
	select {
	case w.events <- ProgressEvent{Type: "progress", ID: id, Current: current, Total: total, Percent: percent}:
	default:
	}
}

func (w *webRenderer) Finish(id string) {
	select {
	case w.events <- ProgressEvent{Type: "finish", ID: id}:
	default:
	}
}

func (w *webRenderer) Log(level downloader.LogLevel, msg string) {
	levelStr := "info"
	switch level {
	case downloader.LogDebug:
		levelStr = "debug"
	case downloader.LogWarn:
		levelStr = "warn"
	case downloader.LogError:
		levelStr = "error"
	}
	select {
	case w.events <- ProgressEvent{Type: "log", Level: levelStr, Message: msg}:
	default:
	}
}

// Compile-time check that webRenderer satisfies the interface.
var _ downloader.ProgressRenderer = (*webRenderer)(nil)
