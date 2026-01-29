package downloader

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	progressbar "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressManager manages download progress rendering using Bubble Tea.
type ProgressManager struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	program *tea.Program
	started bool
}

// NewProgressManager creates a new progress manager.
func NewProgressManager(opts Options) *ProgressManager {
	return &ProgressManager{}
}

// Start begins the progress rendering in a separate goroutine.
func (pm *ProgressManager) Start(ctx context.Context) {
	if pm == nil {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.started {
		return
	}

	model := newProgressModel()
	opts := []tea.ProgramOption{
		tea.WithOutput(os.Stderr),
		tea.WithoutSignalHandler(),
	}
	program := tea.NewProgram(model, opts...)

	pm.ctx, pm.cancel = context.WithCancel(ctx)
	pm.program = program
	pm.started = true

	go func() {
		_, _ = program.Run()
	}()

	go func() {
		<-pm.ctx.Done()
		pm.send(stopMsg{})
	}()
}

// Stop stops the progress rendering and waits for it to finish.
func (pm *ProgressManager) Stop() {
	if pm == nil {
		return
	}

	pm.mu.Lock()
	program := pm.program
	pm.mu.Unlock()

	if program != nil {
		program.Send(stopMsg{})
	}
}

// Log enqueues a message to be rendered above the progress bars.
func (pm *ProgressManager) Log(level LogLevel, msg string) {
	if pm == nil || msg == "" {
		return
	}
	// Send logs through the model
	pm.send(logMsg{level: level, text: msg})
}

func (pm *ProgressManager) send(msg tea.Msg) {
	if pm == nil {
		return
	}
	pm.mu.Lock()
	program := pm.program
	pm.mu.Unlock()
	if program != nil {
		program.Send(msg)
	}
}

type progressRenderer struct {
	manager *ProgressManager
}

func NewProgressRenderer(manager *ProgressManager) *progressRenderer {
	return &progressRenderer{manager: manager}
}

func (pr *progressRenderer) Register(prefix string, size int64) string {
	if pr == nil || pr.manager == nil {
		return ""
	}
	id := fmt.Sprintf("%s@%d", prefix, time.Now().UnixNano())
	pr.manager.send(registerMsg{
		id:    id,
		label: prefix,
		total: size,
		start: time.Now(),
	})
	return id
}

func (pr *progressRenderer) Update(id string, current, total int64) {
	if pr == nil || pr.manager == nil {
		return
	}
	pr.manager.send(updateMsg{id: id, current: current, total: total})
}

func (pr *progressRenderer) Finish(id string) {
	if pr == nil || pr.manager == nil {
		return
	}
	pr.manager.send(finishMsg{id: id})
}

func (pr *progressRenderer) Log(level LogLevel, msg string) {
	if pr == nil || pr.manager == nil {
		return
	}
	pr.manager.Log(level, msg)
}

type registerMsg struct {
	id    string
	label string
	total int64
	start time.Time
}

type updateMsg struct {
	id      string
	current int64
	total   int64
}

type finishMsg struct {
	id string
}

type logMsg struct {
	level LogLevel
	text  string
}

type stopMsg struct{}

// Styles following Bubble Tea examples
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0B0B0B")).
			Background(lipgloss.Color("#FFE66D")).
			Bold(true).
			Padding(0, 1)

	percentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00F5D4")).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Bold(true)

	progressBarStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Bold(true)

	etaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6ADC8")).
			Faint(true)

	logInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7FDBFF")).
			Bold(true)

	logWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD166")).
			Bold(true)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)
)

type progressModel struct {
	tasks map[string]*progressTask
	order []string
	width int
	quit  bool
	log   string
}

type progressTask struct {
	id      string
	label   string
	total   int64
	current int64
	started time.Time
	percent float64
	bar     progressbar.Model
}

func newProgressModel() *progressModel {
	return &progressModel{
		tasks: make(map[string]*progressTask),
		order: make([]string, 0),
		width: 80,
	}
}

func barWidth(total int) int {
	width := total - 10
	if width < 10 {
		return 10
	}
	return width
}

func truncateLine(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}

func (m *progressModel) Init() tea.Cmd {
	return nil
}

func (m *progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		for _, task := range m.tasks {
			task.bar.Width = barWidth(m.width)
		}
	case registerMsg:
		if _, exists := m.tasks[msg.id]; exists {
			return m, nil
		}
		m.order = append(m.order, msg.id)
		bar := progressbar.New(
			progressbar.WithGradient("#FF006E", "#00F5FF"),
			progressbar.WithWidth(barWidth(m.width)),
			progressbar.WithoutPercentage(),
		)
		task := &progressTask{
			id:      msg.id,
			label:   msg.label,
			total:   msg.total,
			started: msg.start,
			bar:     bar,
		}
		m.tasks[msg.id] = task
		return m, task.bar.SetPercent(0)
	case updateMsg:
		if task, ok := m.tasks[msg.id]; ok {
			task.current = msg.current
			if msg.total > 0 {
				task.total = msg.total
			}
			if task.total > 0 {
				task.percent = math.Min(1, math.Max(0, float64(task.current)/float64(task.total)))
				return m, task.bar.SetPercent(task.percent)
			}
		}
	case finishMsg:
		if task, ok := m.tasks[msg.id]; ok {
			task.percent = 1
			return m, task.bar.SetPercent(1)
		}
	case logMsg:
		// Update log message (single line)
		var style lipgloss.Style
		switch msg.level {
		case LogError:
			style = logErrorStyle
		case LogWarn:
			style = logWarnStyle
		default:
			style = logInfoStyle
		}
		m.log = style.Render(truncateLine(msg.text, m.width))
		return m, nil
	case progressbar.FrameMsg:
		cmds := make([]tea.Cmd, 0, len(m.tasks))
		for _, task := range m.tasks {
			if task == nil {
				continue
			}
			model, cmd := task.bar.Update(msg)
			if updated, ok := model.(progressbar.Model); ok {
				task.bar = updated
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
	case stopMsg:
		m.quit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *progressModel) View() string {
	if m.quit {
		return ""
	}

	var b strings.Builder

	// Show log message if any
	if m.log != "" {
		b.WriteString(m.log)
		b.WriteString("\n")
	}

	// Render downloads
	if len(m.order) > 0 {
		b.WriteString(titleStyle.Render(" Downloads"))
		b.WriteString("\n")

		for _, id := range m.order {
			task, ok := m.tasks[id]
			if !ok {
				continue
			}

			elapsed := time.Since(task.started)
			eta := estimateETA(task.current, task.total, elapsed)

			percentText := percentStyle.Render(fmt.Sprintf("%5.1f%%", task.percent*100))
			labelText := labelStyle.Render(task.label)

			// First line: percentage and label
			b.WriteString(fmt.Sprintf("%s %s\n", percentText, labelText))

			// Second line: progress bar
			bar := task.bar.View()
			b.WriteString(progressBarStyle.Render(bar))
			b.WriteString("\n")

			// Third line: ETA
			etaText := etaStyle.Render(fmt.Sprintf("elapsed %s · eta %s",
				formatDurationShort(elapsed),
				formatDurationShort(eta)))
			b.WriteString(fmt.Sprintf("        %s\n", etaText))
		}
	}

	return b.String()
}

func estimateETA(current, total int64, elapsed time.Duration) time.Duration {
	if total <= 0 || current <= 0 {
		return 0
	}
	remaining := total - current
	if remaining <= 0 {
		return 0
	}
	rate := float64(current) / elapsed.Seconds()
	if rate <= 0 {
		return 0
	}
	return time.Duration(float64(remaining)/rate) * time.Second
}

func formatDurationShort(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), math.Mod(d.Seconds(), 60))
	} else {
		return fmt.Sprintf("%.0fh%.0fm", d.Hours(), math.Mod(d.Minutes(), 60))
	}
}
