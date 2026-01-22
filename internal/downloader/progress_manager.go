package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

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
	writer        io.Writer
	layout        string
	logLevel      LogLevel
	width         int
	widthFunc     func() int
	events        chan progressEvent
	stop          chan struct{}
	done          chan struct{}
	renderedLines int
	tasks         map[string]*progressTask
	order         []string
	counter       uint64
	ticker        *time.Ticker
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
	manager := &progressManager{
		writer:    os.Stderr,
		layout:    layout,
		logLevel:  parseLogLevel(opts.LogLevel),
		widthFunc: terminalWidth,
		events:    make(chan progressEvent, 128),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		tasks:     make(map[string]*progressTask),
	}
	return manager
}

func newProgressManagerForTest(writer io.Writer, widthFunc func() int, layout string) *progressManager {
	if strings.TrimSpace(layout) == "" {
		layout = "{label} {percent} {bar} {bytes} {rate} {eta}"
	}
	return &progressManager{
		writer:    writer,
		layout:    layout,
		logLevel:  LogInfo,
		widthFunc: widthFunc,
		events:    make(chan progressEvent, 128),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
		tasks:     make(map[string]*progressTask),
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
	m.ticker = time.NewTicker(200 * time.Millisecond)
	go m.loop(ctx)
}

func (m *progressManager) Stop() {
	if m == nil {
		return
	}
	close(m.stop)
	<-m.done
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
			m.finalRender()
			return
		case <-m.stop:
			m.finalRender()
			return
		case <-m.ticker.C:
			m.render()
		case <-sigch:
			m.sendEvent(progressEvent{kind: eventResize, width: m.widthFunc()})
		case event := <-m.events:
			switch event.kind {
			case eventAddTask:
				m.tasks[event.taskID] = &progressTask{
					id:    event.taskID,
					label: event.label,
					total: event.total,
				}
				m.order = append(m.order, event.taskID)
				m.render()
			case eventUpdateTask:
				task := m.tasks[event.taskID]
				if task == nil {
					break
				}
				task.current = event.current
				if event.total > 0 {
					task.total = event.total
				}
				task.elapsed = event.elapsed
			case eventFinishTask:
				task := m.tasks[event.taskID]
				if task != nil {
					task.finished = true
				}
				m.render()
			case eventLog:
				m.logLine(event.level, event.message)
			case eventResize:
				if event.width > 0 {
					m.width = event.width
					m.render()
				}
			}
		}
	}
}

func (m *progressManager) finalRender() {
	m.render()
	if m.ticker != nil {
		m.ticker.Stop()
	}
}

func (m *progressManager) render() {
	lines := make([]string, 0, len(m.order))
	for _, id := range m.order {
		task := m.tasks[id]
		if task == nil {
			continue
		}
		lines = append(lines, m.renderTask(task))
	}
	m.renderLines(lines)
}

func (m *progressManager) logLine(level LogLevel, message string) {
	m.clearLines()
	label := levelLabel(level)
	if message == "" {
		fmt.Fprintln(m.writer, label)
	} else {
		fmt.Fprintf(m.writer, "%s %s\n", label, message)
	}
	m.render()
}

func (m *progressManager) renderLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	builder := &strings.Builder{}
	if m.renderedLines > 0 {
		fmt.Fprintf(builder, "\x1b[%dA", m.renderedLines)
	}
	for i, line := range lines {
		builder.WriteString("\x1b[2K")
		builder.WriteString(line)
		if i < len(lines)-1 {
			builder.WriteString("\n")
		}
	}
	if extra := m.renderedLines - len(lines); extra > 0 {
		for i := 0; i < extra; i++ {
			builder.WriteString("\n\x1b[2K")
		}
	}
	m.renderedLines = len(lines)
	fmt.Fprint(m.writer, builder.String())
}

func (m *progressManager) clearLines() {
	if m.renderedLines == 0 {
		return
	}
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "\x1b[%dA", m.renderedLines)
	for i := 0; i < m.renderedLines; i++ {
		builder.WriteString("\x1b[2K")
		if i < m.renderedLines-1 {
			builder.WriteString("\n")
		}
	}
	fmt.Fprint(m.writer, builder.String())
	m.renderedLines = 0
}

func (m *progressManager) renderTask(task *progressTask) string {
	width := m.width
	if width <= 0 {
		width = 100
	}

	percentText := "--.--%"
	if task.total > 0 {
		percent := float64(task.current) * 100 / float64(task.total)
		percentText = fmt.Sprintf("%6.2f%%", percent)
	}

	rate := ""
	if task.elapsed > 0 {
		bytesPerSec := int64(float64(task.current) / task.elapsed.Seconds())
		rate = humanBytes(bytesPerSec) + "/s"
	}

	eta := ""
	if task.total > 0 && task.elapsed > 0 {
		remaining := task.total - task.current
		if remaining < 0 {
			remaining = 0
		}
		bytesPerSec := float64(task.current) / task.elapsed.Seconds()
		if bytesPerSec > 0 {
			eta = formatETA(time.Duration(float64(remaining)/bytesPerSec) * time.Second)
		}
	}

	bytesText := humanBytes(task.current)
	if task.total > 0 {
		bytesText = fmt.Sprintf("%s/%s", bytesText, humanBytes(task.total))
	}

	fields := map[string]string{
		"{label}":   task.label,
		"{percent}": percentText,
		"{rate}":    padLeft(rate, 9),
		"{eta}":     padLeft(eta, 6),
		"{bytes}":   padLeft(bytesText, 17),
		"{current}": humanBytes(task.current),
		"{total}":   humanBytes(task.total),
	}

	layout := m.layout
	barWidth := 0
	if strings.Contains(layout, "{bar}") {
		base := layout
		for key, value := range fields {
			base = strings.ReplaceAll(base, key, value)
		}
		base = strings.ReplaceAll(base, "{bar}", "")
		barWidth = width - visibleLength(base) - 1
		if barWidth < 10 {
			barWidth = 10
		}
		bar := renderBar(task.current, task.total, barWidth)
		layout = strings.ReplaceAll(layout, "{bar}", bar)
	}

	for key, value := range fields {
		layout = strings.ReplaceAll(layout, key, value)
	}

	if overflow := visibleLength(layout) - width; overflow > 0 {
		label := fields["{label}"]
		maxLabel := len(label) - overflow
		if maxLabel < 0 {
			maxLabel = 0
		}
		fields["{label}"] = truncateText(label, maxLabel)
		layout = m.layout
		if barWidth > 0 {
			layout = strings.ReplaceAll(layout, "{bar}", renderBar(task.current, task.total, barWidth))
		}
		for key, value := range fields {
			layout = strings.ReplaceAll(layout, key, value)
		}
	}

	return layout
}

func renderBar(current, total int64, width int) string {
	if width <= 0 {
		return ""
	}
	fill := 0
	if total > 0 {
		fill = int(float64(current) / float64(total) * float64(width))
		if fill > width {
			fill = width
		}
	}
	return "[" + strings.Repeat("=", fill) + strings.Repeat(" ", width-fill) + "]"
}

func formatETA(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	seconds := int(d.Seconds())
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
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
