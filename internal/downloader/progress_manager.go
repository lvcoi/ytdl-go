package downloader

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

// ProgressManager wraps go-pretty's progress.Writer to coordinate
// multiple concurrent download progress trackers.
type ProgressManager struct {
	writer    progress.Writer
	trackers  map[string]*progress.Tracker
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	started   bool
}

// NewProgressManager creates a new progress manager if progress output is enabled.
// Returns nil if progress should not be displayed (quiet mode, non-TTY, etc).
func NewProgressManager(opts Options) *ProgressManager {
	if opts.Quiet || opts.JSON {
		return nil
	}
	if !isTerminal(os.Stderr) || !supportsANSI() {
		return nil
	}

	writer := progress.NewWriter()
	writer.SetOutputWriter(os.Stderr)
	writer.SetAutoStop(false)
	writer.SetTrackerLength(25)
	writer.SetMessageLength(40)
	writer.SetUpdateFrequency(time.Millisecond * 100)
	
	style := progress.StyleDefault
	style.Visibility.ETA = true
	style.Visibility.Percentage = true
	style.Visibility.Speed = true
	style.Visibility.Value = true
	writer.SetStyle(style)

	return &ProgressManager{
		writer:   writer,
		trackers: make(map[string]*progress.Tracker),
	}
}

// Start begins the progress rendering loop in a background goroutine.
func (pm *ProgressManager) Start(ctx context.Context) {
	if pm == nil {
		return
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.started {
		return
	}

	pm.ctx, pm.cancel = context.WithCancel(ctx)
	pm.started = true

	go func() {
		for {
			select {
			case <-pm.ctx.Done():
				return
			default:
				pm.writer.Render()
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
}

// Stop stops the progress rendering and cleans up.
func (pm *ProgressManager) Stop() {
	if pm == nil {
		return
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.started {
		return
	}

	if pm.cancel != nil {
		pm.cancel()
	}
	pm.writer.Stop()
	pm.started = false
}

// Log writes a message to the progress output without interfering with progress bars.
func (pm *ProgressManager) Log(level LogLevel, msg string) {
	if pm == nil {
		return
	}
	pm.writer.Log(msg)
}

// progressRenderer adapts the ProgressManager to provide the renderer interface
// expected by progressWriter.
type progressRenderer struct {
	manager *ProgressManager
}

// Register creates a new progress tracker for the given task.
// Returns a taskID that can be used to update or finish the tracker.
func (pr *progressRenderer) Register(prefix string, size int64) string {
	if pr == nil || pr.manager == nil {
		return ""
	}

	pr.manager.mu.Lock()
	defer pr.manager.mu.Unlock()

	tracker := &progress.Tracker{
		Message: prefix,
		Total:   size,
		Units:   progress.UnitsBytes,
	}
	
	pr.manager.writer.AppendTracker(tracker)
	
	// Use prefix as taskID for simplicity
	taskID := prefix
	pr.manager.trackers[taskID] = tracker
	
	return taskID
}

// Update updates the progress of a task.
func (pr *progressRenderer) Update(taskID string, _ int64, current, total int64) {
	if pr == nil || pr.manager == nil {
		return
	}

	pr.manager.mu.Lock()
	defer pr.manager.mu.Unlock()

	tracker, ok := pr.manager.trackers[taskID]
	if !ok {
		return
	}

	tracker.SetValue(current)
	if total > 0 && tracker.Total != total {
		tracker.Total = total
	}
}

// Finish marks a task as complete.
func (pr *progressRenderer) Finish(taskID string) {
	if pr == nil || pr.manager == nil {
		return
	}

	pr.manager.mu.Lock()
	defer pr.manager.mu.Unlock()

	tracker, ok := pr.manager.trackers[taskID]
	if !ok {
		return
	}

	tracker.MarkAsDone()
	delete(pr.manager.trackers, taskID)
}

// Log writes a message without interfering with progress bars.
func (pr *progressRenderer) Log(msg string) {
	if pr == nil || pr.manager == nil {
		return
	}
	pr.manager.writer.Log(msg)
}
