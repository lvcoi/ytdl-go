package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
)

// ProgressStats summarizes final job outcome counts.
type ProgressStats struct {
	Total     int `json:"total,omitempty"`
	Succeeded int `json:"succeeded,omitempty"`
	Failed    int `json:"failed,omitempty"`
}

// ProgressTaskSnapshot captures the latest known state of a progress task.
type ProgressTaskSnapshot struct {
	ID      string  `json:"id"`
	Label   string  `json:"label,omitempty"`
	Current int64   `json:"current,omitempty"`
	Total   int64   `json:"total,omitempty"`
	Percent float64 `json:"percent,omitempty"`
	Done    bool    `json:"done,omitempty"`
}

// ProgressLogSnapshot captures a log line emitted during a job.
type ProgressLogSnapshot struct {
	Seq     int64  `json:"seq,omitempty"`
	At      string `json:"at,omitempty"`
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
}

// DuplicatePromptSnapshot captures a pending duplicate-file prompt.
type DuplicatePromptSnapshot struct {
	PromptID string `json:"promptId"`
	Path     string `json:"path,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// ProgressSnapshot captures a full replayable view of job state.
type ProgressSnapshot struct {
	JobID       string                    `json:"jobId"`
	Status      string                    `json:"status"`
	CreatedAt   time.Time                 `json:"createdAt"`
	CompletedAt *time.Time                `json:"completedAt,omitempty"`
	ExitCode    int                       `json:"exitCode,omitempty"`
	Error       string                    `json:"error,omitempty"`
	LastSeq     int64                     `json:"lastSeq,omitempty"`
	Stats       ProgressStats             `json:"stats,omitempty"`
	Tasks       []ProgressTaskSnapshot    `json:"tasks,omitempty"`
	Logs        []ProgressLogSnapshot     `json:"logs,omitempty"`
	Duplicates  []DuplicatePromptSnapshot `json:"duplicates,omitempty"`
}

// ProgressEvent is a structured event sent over SSE.
type ProgressEvent struct {
	Type     string            `json:"type"`
	JobID    string            `json:"jobId,omitempty"`
	Seq      int64             `json:"seq,omitempty"`
	At       string            `json:"at,omitempty"`
	Status   string            `json:"status,omitempty"`
	ID       string            `json:"id,omitempty"`
	Label    string            `json:"label,omitempty"`
	Current  int64             `json:"current,omitempty"`
	Total    int64             `json:"total,omitempty"`
	Percent  float64           `json:"percent,omitempty"`
	Level    string            `json:"level,omitempty"`
	Message  string            `json:"message,omitempty"`
	Error    string            `json:"error,omitempty"`
	PromptID string            `json:"promptId,omitempty"`
	Path     string            `json:"path,omitempty"`
	Filename string            `json:"filename,omitempty"`
	ExitCode int               `json:"exitCode,omitempty"`
	Stats    *ProgressStats    `json:"stats,omitempty"`
	Snapshot *ProgressSnapshot `json:"snapshot,omitempty"`
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
	Stats       ProgressStats      `json:"stats,omitempty"`

	promptCounter atomic.Int64                                 `json:"-"`
	promptMu      sync.Mutex                                   `json:"-"`
	promptClosed  bool                                         `json:"-"`
	pendingPrompt map[string]chan downloader.DuplicateDecision `json:"-"`

	mu sync.RWMutex `json:"-"`

	eventMu            sync.Mutex                      `json:"-"`
	eventClosed        bool                            `json:"-"`
	eventSeq           atomic.Int64                    `json:"-"`
	eventHistory       []ProgressEvent                 `json:"-"`
	subscribers        map[int64]chan ProgressEvent    `json:"-"`
	subscriberCounter  atomic.Int64                    `json:"-"`
	taskState          map[string]ProgressTaskSnapshot `json:"-"`
	logState           []ProgressLogSnapshot           `json:"-"`
	duplicatePromptMap map[string]DuplicatePromptSnapshot
	closeOnce          sync.Once `json:"-"`
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
	criticalEventTimeout    = 2 * time.Second
	duplicatePromptTimeout  = 120 * time.Second
	maxJobEventHistory      = 4096
	maxJobLogHistory        = 200
	baseSubscriberBufferLen = 128
)

func (jt *jobTracker) Create(urls []string) *Job {
	id := fmt.Sprintf("job_%d", jt.counter.Add(1))
	urlCopy := append([]string(nil), urls...)
	job := &Job{
		ID:                 id,
		Status:             "queued",
		URLs:               urlCopy,
		CreatedAt:          time.Now(),
		Events:             make(chan ProgressEvent, 256),
		pendingPrompt:      make(map[string]chan downloader.DuplicateDecision),
		subscribers:        make(map[int64]chan ProgressEvent),
		taskState:          make(map[string]ProgressTaskSnapshot),
		duplicatePromptMap: make(map[string]DuplicatePromptSnapshot),
	}
	go job.runEventBroker()
	job.emitStatusEvent("queued", "queued")
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
	if v, ok := jt.jobs.Load(id); ok {
		if job, jobOK := v.(*Job); jobOK {
			job.CloseEvents()
		}
	}
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
			job.CloseEvents()
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

func (j *Job) CloseEvents() {
	if j == nil {
		return
	}
	j.closeOnce.Do(func() {
		close(j.Events)
	})
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
	j.setTerminalStatusLocked(status)
	j.mu.Unlock()
	j.emitStatusEvent(status, status)
}

func (j *Job) SetOutcome(results []app.Result, exitCode int) string {
	resultsCopy := append([]app.Result(nil), results...)
	stats := computeProgressStats(resultsCopy)

	j.mu.Lock()
	j.Results = resultsCopy
	j.ExitCode = exitCode
	j.Error = ""
	j.Stats = stats

	if exitCode != 0 {
		j.setTerminalStatusLocked("error")
		for _, result := range resultsCopy {
			if result.Error != "" {
				j.Error = result.Error
				break
			}
		}
		status := j.Status
		errMsg := j.Error
		statsCopy := j.Stats
		j.mu.Unlock()
		j.emitTerminalStatusEvent(status, errMsg, exitCode, statsCopy)
		return status
	}

	j.setTerminalStatusLocked("complete")
	status := j.Status
	statsCopy := j.Stats
	j.mu.Unlock()
	j.emitTerminalStatusEvent(status, "", exitCode, statsCopy)
	return status
}

func (j *Job) emitStatusEvent(status, message string) {
	if j == nil {
		return
	}
	evt := ProgressEvent{
		Type:    "status",
		Status:  status,
		Message: message,
	}
	_ = j.enqueueCriticalEvent(evt, criticalEventTimeout)
}

func (j *Job) emitTerminalStatusEvent(status, errMsg string, exitCode int, stats ProgressStats) {
	if j == nil {
		return
	}
	evt := ProgressEvent{
		Type:     "status",
		Status:   status,
		Message:  status,
		Error:    errMsg,
		ExitCode: exitCode,
	}
	if stats.Total > 0 {
		statsCopy := stats
		evt.Stats = &statsCopy
	}
	_ = j.enqueueCriticalEvent(evt, criticalEventTimeout)
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
		if recovered := recover(); recovered != nil {
			log.Printf("web: dropping critical event for job %q after panic on send: %v", j.ID, recovered)
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
	_ = j.enqueueCriticalEvent(ProgressEvent{
		Type:     "duplicate-resolved",
		PromptID: promptID,
		Message:  string(decision),
	}, criticalEventTimeout)
	return nil
}

func (j *Job) expireDuplicatePrompt(promptID string) {
	resp, ok := j.popDuplicatePrompt(promptID)
	if !ok {
		return
	}
	close(resp)
	_ = j.enqueueCriticalEvent(ProgressEvent{
		Type:     "duplicate-resolved",
		PromptID: promptID,
		Message:  string(downloader.DuplicateDecisionSkip),
	}, criticalEventTimeout)
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

	for promptID, resp := range pending {
		if defaultDecision != "" {
			select {
			case resp <- defaultDecision:
			default:
			}
		}
		close(resp)
		_ = j.enqueueCriticalEvent(ProgressEvent{
			Type:     "duplicate-resolved",
			PromptID: promptID,
			Message:  string(defaultDecision),
		}, criticalEventTimeout)
	}
}

func (j *Job) runEventBroker() {
	for evt := range j.Events {
		normalized := j.normalizeEvent(evt)
		j.recordEvent(normalized)
	}

	j.eventMu.Lock()
	if j.eventClosed {
		j.eventMu.Unlock()
		return
	}
	j.eventClosed = true
	subs := j.subscribers
	j.subscribers = make(map[int64]chan ProgressEvent)
	j.eventMu.Unlock()

	for _, sub := range subs {
		close(sub)
	}
}

func (j *Job) normalizeEvent(evt ProgressEvent) ProgressEvent {
	if evt.JobID == "" {
		evt.JobID = j.ID
	}
	evt.Seq = j.eventSeq.Add(1)
	if evt.At == "" {
		evt.At = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if evt.Type == "progress" && evt.Total > 0 && evt.Percent == 0 && evt.Current > 0 {
		evt.Percent = float64(evt.Current) * 100 / float64(evt.Total)
	}
	if evt.Type == "finish" && evt.Percent <= 0 {
		evt.Percent = 100
	}
	if evt.Type == "status" && evt.Status == "" {
		evt.Status = evt.Message
	}
	if evt.Type == "done" {
		if evt.Status == "" {
			evt.Status = evt.Message
		}
		if evt.Message == "" {
			evt.Message = evt.Status
		}
	}
	return evt
}

func (j *Job) recordEvent(evt ProgressEvent) {
	j.eventMu.Lock()

	j.eventHistory = append(j.eventHistory, evt)
	if len(j.eventHistory) > maxJobEventHistory {
		over := len(j.eventHistory) - maxJobEventHistory
		j.eventHistory = append([]ProgressEvent(nil), j.eventHistory[over:]...)
	}
	j.applyEventLocked(evt)

	subs := make([]chan ProgressEvent, 0, len(j.subscribers))
	for _, sub := range j.subscribers {
		subs = append(subs, sub)
	}
	j.eventMu.Unlock()

	critical := isCriticalEventType(evt.Type)
	for _, sub := range subs {
		if critical {
			timer := time.NewTimer(200 * time.Millisecond)
			select {
			case sub <- evt:
			case <-timer.C:
			}
			timer.Stop()
			continue
		}
		select {
		case sub <- evt:
		default:
		}
	}
}

func isCriticalEventType(eventType string) bool {
	switch eventType {
	case "status", "done", "duplicate", "duplicate-resolved":
		return true
	default:
		return false
	}
}

func (j *Job) applyEventLocked(evt ProgressEvent) {
	switch evt.Type {
	case "register":
		task := j.taskState[evt.ID]
		task.ID = evt.ID
		if evt.Label != "" {
			task.Label = evt.Label
		}
		task.Total = evt.Total
		task.Current = evt.Current
		task.Percent = evt.Percent
		task.Done = false
		j.taskState[evt.ID] = task
	case "progress":
		task := j.taskState[evt.ID]
		task.ID = evt.ID
		if evt.Label != "" {
			task.Label = evt.Label
		}
		if evt.Total > 0 || task.Total == 0 {
			task.Total = evt.Total
		}
		task.Current = evt.Current
		task.Percent = evt.Percent
		task.Done = false
		j.taskState[evt.ID] = task
	case "finish":
		task := j.taskState[evt.ID]
		task.ID = evt.ID
		if task.Percent < 100 {
			task.Percent = 100
		}
		if task.Total > 0 && task.Current < task.Total {
			task.Current = task.Total
		}
		task.Done = true
		j.taskState[evt.ID] = task
	case "log":
		j.logState = append(j.logState, ProgressLogSnapshot{
			Seq:     evt.Seq,
			At:      evt.At,
			Level:   evt.Level,
			Message: evt.Message,
		})
		if len(j.logState) > maxJobLogHistory {
			over := len(j.logState) - maxJobLogHistory
			j.logState = append([]ProgressLogSnapshot(nil), j.logState[over:]...)
		}
	case "duplicate":
		if evt.PromptID != "" {
			j.duplicatePromptMap[evt.PromptID] = DuplicatePromptSnapshot{
				PromptID: evt.PromptID,
				Path:     evt.Path,
				Filename: evt.Filename,
			}
		}
	case "duplicate-resolved":
		if evt.PromptID != "" {
			delete(j.duplicatePromptMap, evt.PromptID)
		}
	}
}

// Subscribe creates a replay-capable event stream for a job.
// A snapshot event is emitted first, followed by historical events newer than afterSeq,
// then live events.
func (j *Job) Subscribe(afterSeq int64) (<-chan ProgressEvent, func()) {
	if j == nil {
		ch := make(chan ProgressEvent)
		close(ch)
		return ch, func() {}
	}
	if afterSeq < 0 {
		afterSeq = 0
	}

	snapshotEvt := j.snapshotEvent()

	j.eventMu.Lock()
	replay := make([]ProgressEvent, 0, len(j.eventHistory))
	for _, evt := range j.eventHistory {
		if evt.Seq > afterSeq {
			replay = append(replay, evt)
		}
	}
	bufferSize := baseSubscriberBufferLen + len(replay) + 1
	if bufferSize < baseSubscriberBufferLen {
		bufferSize = baseSubscriberBufferLen
	}
	ch := make(chan ProgressEvent, bufferSize)
	ch <- snapshotEvt
	for _, evt := range replay {
		ch <- evt
	}
	if j.eventClosed {
		close(ch)
		j.eventMu.Unlock()
		return ch, func() {}
	}
	subID := j.subscriberCounter.Add(1)
	j.subscribers[subID] = ch
	j.eventMu.Unlock()

	cancel := func() {
		j.eventMu.Lock()
		if _, ok := j.subscribers[subID]; ok {
			delete(j.subscribers, subID)
		}
		j.eventMu.Unlock()
	}

	return ch, cancel
}

func (j *Job) snapshotEvent() ProgressEvent {
	snapshot := j.progressSnapshot()
	return ProgressEvent{
		Type:     "snapshot",
		JobID:    j.ID,
		At:       time.Now().UTC().Format(time.RFC3339Nano),
		Status:   snapshot.Status,
		Snapshot: &snapshot,
	}
}

func (j *Job) progressSnapshot() ProgressSnapshot {
	status, createdAt, completedAt, exitCode, errMsg, stats := j.stateSnapshot()

	j.eventMu.Lock()
	tasks := make([]ProgressTaskSnapshot, 0, len(j.taskState))
	for _, task := range j.taskState {
		tasks = append(tasks, task)
	}
	logs := append([]ProgressLogSnapshot(nil), j.logState...)
	duplicates := make([]DuplicatePromptSnapshot, 0, len(j.duplicatePromptMap))
	for _, dup := range j.duplicatePromptMap {
		duplicates = append(duplicates, dup)
	}
	lastSeq := j.eventSeq.Load()
	j.eventMu.Unlock()

	sort.Slice(tasks, func(i, k int) bool {
		if tasks[i].Label == tasks[k].Label {
			return tasks[i].ID < tasks[k].ID
		}
		return tasks[i].Label < tasks[k].Label
	})
	sort.Slice(duplicates, func(i, k int) bool {
		if duplicates[i].Filename == duplicates[k].Filename {
			return duplicates[i].PromptID < duplicates[k].PromptID
		}
		return duplicates[i].Filename < duplicates[k].Filename
	})

	snapshot := ProgressSnapshot{
		JobID:      j.ID,
		Status:     status,
		CreatedAt:  createdAt,
		LastSeq:    lastSeq,
		Stats:      stats,
		Tasks:      tasks,
		Logs:       logs,
		Duplicates: duplicates,
	}
	if !completedAt.IsZero() {
		t := completedAt
		snapshot.CompletedAt = &t
	}
	if exitCode != 0 {
		snapshot.ExitCode = exitCode
	}
	if errMsg != "" {
		snapshot.Error = errMsg
	}
	return snapshot
}

func (j *Job) stateSnapshot() (status string, createdAt, completedAt time.Time, exitCode int, errMsg string, stats ProgressStats) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status, j.CreatedAt, j.CompletedAt, j.ExitCode, j.Error, j.Stats
}

func computeProgressStats(results []app.Result) ProgressStats {
	stats := ProgressStats{Total: len(results)}
	for _, result := range results {
		if result.Error == "" {
			stats.Succeeded++
			continue
		}
		stats.Failed++
	}
	return stats
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

func safeEnqueueEvent(events chan<- ProgressEvent, evt ProgressEvent) {
	if events == nil {
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("web: dropped renderer event %q after panic on send: %v", evt.Type, recovered)
		}
	}()
	select {
	case events <- evt:
	default:
	}
}

func (w *webRenderer) Register(prefix string, size int64) string {
	id := fmt.Sprintf("%s@%d", prefix, time.Now().UnixNano())
	safeEnqueueEvent(w.events, ProgressEvent{Type: "register", ID: id, Label: prefix, Total: size})
	return id
}

func (w *webRenderer) Update(id string, current, total int64) {
	percent := 0.0
	if total > 0 {
		percent = float64(current) * 100 / float64(total)
	}
	safeEnqueueEvent(w.events, ProgressEvent{Type: "progress", ID: id, Current: current, Total: total, Percent: percent})
}

func (w *webRenderer) Finish(id string) {
	safeEnqueueEvent(w.events, ProgressEvent{Type: "finish", ID: id})
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
	safeEnqueueEvent(w.events, ProgressEvent{Type: "log", Level: levelStr, Message: msg})
}

func parseProgressSeq(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	seq, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seq < 0 {
		return 0, false
	}
	return seq, true
}

// Compile-time check that webRenderer satisfies the interface.
var _ downloader.ProgressRenderer = (*webRenderer)(nil)
