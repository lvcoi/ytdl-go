package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/term"
)

type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

type progressEventType int

const (
	eventAddTask progressEventType = iota
	eventUpdateTask
	eventFinishTask
	eventLog
	eventResize
)

type progressEvent struct {
	kind    progressEventType
	taskID  string
	label   string
	current int64
	total   int64
	elapsed time.Duration
	level   LogLevel
	message string
	width   int
}

type progressTask struct {
	id       string
	label    string
	current  int64
	total    int64
	elapsed  time.Duration
	finished bool
}

type progressManager struct {
	writer   io.Writer
	layout   string
	logLevel LogLevel
	width    int
	widthFunc func() int
	events   chan progressEvent
	stop     chan struct{}
	done     chan struct{}
	tasks    map[string]*progressTask
	trackers map[string]*progress.Tracker
	order    []string
	counter  uint64
	pw       progress.Writer
}

func newProgressManager(opts Options) *progressManager {
	if opts.Quiet {
		return nil
	}
	if !isTerminal(os.Stderr) || !supportsANSI() {
		return nil
	}
	layout := opts.ProgressLayout
	if strings.TrimSpace(layout) == "" {
		layout = "{label} {percent} {bar} {bytes} {rate} {eta}"
	}
	
	// Create go-pretty progress writer
	pw := progress.NewWriter()
	pw.SetAutoStop(false)
	pw.SetOutputWriter(os.Stderr)
	pw.SetUpdateFrequency(200 * time.Millisecond)
	pw.Style().Visibility.ETA = true
	pw.Style().Visibility.Speed = true
	pw.Style().Visibility.Value = true
	pw.Style().Visibility.Percentage = true
	pw.Style().Options.SpeedOverallFormatter = progress.FormatBytes
	pw.Style().Colors = progress.StyleColorsExample
	pw.SetStyle(progress.StyleBlocks)
	
	manager := &progressManager{
		writer:    os.Stderr,
		layout:    layout,
		logLevel:  parseLogLevel(opts.LogLevel),
		widthFunc: terminalWidth,
		events:    make(chan progressEvent, 128),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		tasks:     make(map[string]*progressTask),
		trackers:  make(map[string]*progress.Tracker),
		pw:        pw,
	}
	return manager
}

func newProgressManagerForTest(writer io.Writer, widthFunc func() int, layout string) *progressManager {
	if strings.TrimSpace(layout) == "" {
		layout = "{label} {percent} {bar} {bytes} {rate} {eta}"
	}
	
	// Create go-pretty progress writer for test
	pw := progress.NewWriter()
	pw.SetAutoStop(false)
	pw.SetOutputWriter(writer)
	pw.SetUpdateFrequency(200 * time.Millisecond)
	pw.Style().Visibility.ETA = true
	pw.Style().Visibility.Speed = true
	pw.Style().Visibility.Value = true
	pw.Style().Visibility.Percentage = true
	pw.SetStyle(progress.StyleBlocks)
	
	return &progressManager{
		writer:    writer,
		layout:    layout,
		logLevel:  LogInfo,
		widthFunc: widthFunc,
		events:    make(chan progressEvent, 128),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		tasks:     make(map[string]*progressTask),
		trackers:  make(map[string]*progress.Tracker),
		pw:        pw,
	}
}

func (m *progressManager) Start(ctx context.Context) {
	if m == nil {
		return
	}
	m.width = m.widthFunc()
	if m.width <= 0 {
		m.width = 100
	}
	
	// Start go-pretty progress writer rendering
	go m.pw.Render()
	
	go m.loop(ctx)
}

func (m *progressManager) Stop() {
	if m == nil {
		return
	}
	close(m.stop)
	<-m.done
	
	// Stop the progress writer (it will render one final time)
	m.pw.Stop()
}

func (m *progressManager) AddTask(label string, total int64) string {
	if m == nil {
		return ""
	}
	id := fmt.Sprintf("task-%d", atomic.AddUint64(&m.counter, 1))
	m.sendEvent(progressEvent{kind: eventAddTask, taskID: id, label: label, total: total})
	return id
}

func (m *progressManager) UpdateTask(id string, current, total int64, elapsed time.Duration) {
	if m == nil || id == "" {
		return
	}
	m.sendEvent(progressEvent{kind: eventUpdateTask, taskID: id, current: current, total: total, elapsed: elapsed})
}

func (m *progressManager) FinishTask(id string) {
	if m == nil || id == "" {
		return
	}
	m.sendEvent(progressEvent{kind: eventFinishTask, taskID: id})
}

func (m *progressManager) Log(level LogLevel, message string) {
	if m == nil {
		return
	}
	if level < m.logLevel {
		return
	}
	m.sendEvent(progressEvent{kind: eventLog, level: level, message: message})
}

func (m *progressManager) sendEvent(event progressEvent) {
	select {
	case m.events <- event:
	default:
	}
}

func (m *progressManager) loop(ctx context.Context) {
	sigch, stopSignals := resizeSignalChannel()
	if stopSignals != nil {
		defer stopSignals()
	}
	defer close(m.done)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stop:
			return
		case <-sigch:
			m.sendEvent(progressEvent{kind: eventResize, width: m.widthFunc()})
		case event := <-m.events:
			switch event.kind {
			case eventAddTask:
				// Create go-pretty tracker
				tracker := &progress.Tracker{
					Message: event.label,
					Total:   event.total,
					Units:   progress.UnitsBytes,
				}
				m.pw.AppendTracker(tracker)
				
				m.tasks[event.taskID] = &progressTask{
					id:    event.taskID,
					label: event.label,
					total: event.total,
				}
				m.trackers[event.taskID] = tracker
				m.order = append(m.order, event.taskID)
			case eventUpdateTask:
				task := m.tasks[event.taskID]
				tracker := m.trackers[event.taskID]
				if task == nil || tracker == nil {
					break
				}
				task.current = event.current
				if event.total > 0 {
					task.total = event.total
					tracker.UpdateTotal(event.total)
				}
				task.elapsed = event.elapsed
				tracker.SetValue(event.current)
			case eventFinishTask:
				task := m.tasks[event.taskID]
				tracker := m.trackers[event.taskID]
				if task != nil {
					task.finished = true
				}
				if tracker != nil {
					tracker.MarkAsDone()
				}
			case eventLog:
				m.logLine(event.level, event.message)
			case eventResize:
				if event.width > 0 {
					m.width = event.width
					m.pw.SetTerminalWidth(event.width)
				}
			}
		}
	}
}

func (m *progressManager) logLine(level LogLevel, message string) {
	label := levelLabel(level)
	if message == "" {
		m.pw.Log("%s", label)
	} else {
		m.pw.Log("%s %s", label, message)
	}
}

func visibleLength(text string) int {
	return len(text)
}

func parseLogLevel(value string) LogLevel {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return LogDebug
	case "warn", "warning":
		return LogWarn
	case "error":
		return LogError
	case "info", "":
		return LogInfo
	default:
		return LogInfo
	}
}

func levelLabel(level LogLevel) string {
	switch level {
	case LogDebug:
		return "[DEBUG]"
	case LogWarn:
		return "[WARN]"
	case LogError:
		return "[ERROR]"
	default:
		return "[INFO]"
	}
}

func supportsANSI() bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stderr.Fd()))
	if err != nil {
		return 0
	}
	return width
}
