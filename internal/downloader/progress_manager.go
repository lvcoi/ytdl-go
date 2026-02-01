package downloader

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	progressbar "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressManager manages download progress rendering using Bubble Tea.
type ProgressManager struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	program *tea.Program
	started bool
	done    chan struct{}
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
		tea.WithAltScreen(),
		tea.WithoutSignalHandler(),
	}
	program := tea.NewProgram(model, opts...)

	pm.ctx, pm.cancel = context.WithCancel(ctx)
	pm.program = program
	pm.started = true
	pm.done = make(chan struct{})

	go func() {
		defer close(pm.done)
		_, _ = program.Run()
		if pm.cancel != nil {
			pm.cancel()
		}
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
	done := pm.done
	pm.mu.Unlock()

	if program != nil {
		program.Send(stopMsg{})
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
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

func (pm *ProgressManager) PromptDuplicate(path string) (promptChoice, error) {
	if pm == nil {
		return promptQuit, errors.New("no progress manager")
	}
	resp := make(chan promptChoice, 1)
	pm.send(promptMsg{path: path, resp: resp})
	if pm.ctx == nil {
		choice := <-resp
		return choice, nil
	}
	select {
	case choice := <-resp:
		return choice, nil
	case <-pm.ctx.Done():
		return promptQuit, pm.ctx.Err()
	}
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

type promptMsg struct {
	path string
	resp chan promptChoice
}

type promptChoice int

const (
	promptOverwrite promptChoice = iota
	promptSkip
	promptRename
	promptOverwriteAll
	promptSkipAll
	promptRenameAll
	promptQuit
)

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

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FDE68A")).
			Bold(true)

	promptTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0B0B0B")).
				Background(lipgloss.Color("#FF006E")).
				Bold(true).
				Padding(0, 1)

	promptOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EAEAEA")).
				Bold(true)

	promptSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0B0B0B")).
				Background(lipgloss.Color("#00F5D4")).
				Bold(true).
				Padding(0, 1)

	promptOverwriteStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B9A")).Bold(true)
	promptOverwriteAllStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3B30")).Bold(true)
	promptSkipStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#00F5D4")).Bold(true)
	promptSkipAllStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D27A")).Bold(true)
	promptRenameStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD166")).Bold(true)
	promptRenameAllStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB703")).Bold(true)
	promptQuitStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#C0C0C0")).Bold(true)

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7FDBFF"))
)

type progressModel struct {
	tasks        map[string]*progressTask
	order        []string
	width        int
	height       int
	quit         bool
	log          string
	promptActive bool
	promptPath   string
	promptResp   chan promptChoice
	promptQueue  []promptMsg
	promptIndex  int
	vp           viewport.Model
	vpReady      bool
}

type progressTask struct {
	id       string
	label    string
	total    int64
	current  int64
	started  time.Time
	finished time.Time
	percent  float64
	bar      progressbar.Model
	spin     spinner.Model
	done     bool
}

func newProgressModel() *progressModel {
	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7FDBFF"))
	return &progressModel{
		tasks:  make(map[string]*progressTask),
		order:  make([]string, 0),
		width:  80,
		height: 24,
		vp:     vp,
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
		m.height = msg.Height
		headerHeight := 2
		borderHeight := 2
		m.vp.Width = msg.Width - 2
		m.vp.Height = msg.Height - headerHeight - borderHeight
		m.vp, _ = m.vp.Update(msg)
		m.vpReady = true
		for _, task := range m.tasks {
			task.bar.Width = barWidth(m.width)
		}
	case promptMsg:
		if m.promptActive {
			m.promptQueue = append(m.promptQueue, msg)
			return m, nil
		}
		m.promptActive = true
		m.promptPath = msg.path
		m.promptResp = msg.resp
		m.promptIndex = 0
		return m, nil
	case registerMsg:
		if _, exists := m.tasks[msg.id]; exists {
			return m, nil
		}
		m.order = append(m.order, msg.id)
		spin := spinner.New()
		spin.Spinner = spinner.MiniDot
		spin.Style = spinnerStyle
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
			spin:    spin,
		}
		m.tasks[msg.id] = task
		return m, tea.Batch(task.bar.SetPercent(0), task.spin.Tick)
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
			task.done = true
			task.finished = time.Now()
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
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.promptActive {
			switch msg.String() {
			case "up", "k":
				m.vp.SetYOffset(m.vp.YOffset - 1)
			case "down", "j":
				m.vp.SetYOffset(m.vp.YOffset + 1)
			case "pgup":
				m.vp.HalfViewUp()
			case "pgdown", "f", " ":
				m.vp.HalfViewDown()
			case "home", "g":
				m.vp.GotoTop()
			case "end", "G":
				m.vp.GotoBottom()
			}
			return m, nil
		}
		choice := promptChoice(-1)
		switch msg.String() {
		case "up", "k":
			if m.promptIndex > 0 {
				m.promptIndex--
			} else {
				m.promptIndex = 6
			}
			return m, nil
		case "down", "j", "tab":
			if m.promptIndex < 6 {
				m.promptIndex++
			} else {
				m.promptIndex = 0
			}
			return m, nil
		case "enter":
			choice = promptChoiceForIndex(m.promptIndex)
		case "o":
			choice = promptOverwrite
		case "O":
			choice = promptOverwriteAll
		case "s":
			choice = promptSkip
		case "S":
			choice = promptSkipAll
		case "r":
			choice = promptRename
		case "R":
			choice = promptRenameAll
		case "q", "esc":
			choice = promptQuit
		}
		if choice >= 0 {
			if m.promptResp != nil {
				select {
				case m.promptResp <- choice:
				default:
				}
			}
			if choice == promptOverwriteAll || choice == promptSkipAll || choice == promptRenameAll || choice == promptQuit {
				for _, queued := range m.promptQueue {
					if queued.resp == nil {
						continue
					}
					select {
					case queued.resp <- choice:
					default:
					}
				}
				m.promptQueue = nil
			}
			m.promptActive = false
			m.promptPath = ""
			m.promptResp = nil
			m.promptIndex = 0
			if len(m.promptQueue) > 0 {
				next := m.promptQueue[0]
				m.promptQueue = m.promptQueue[1:]
				m.promptActive = true
				m.promptPath = next.path
				m.promptResp = next.resp
				m.promptIndex = 0
			}
			return m, tea.ClearScreen
		}
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
	case spinner.TickMsg:
		cmds := make([]tea.Cmd, 0, len(m.tasks))
		for _, task := range m.tasks {
			if task == nil {
				continue
			}
			updated, cmd := task.spin.Update(msg)
			task.spin = updated
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

	if m.promptActive {
		var dialog strings.Builder
		dialog.WriteString(promptTitleStyle.Render(" Duplicate File "))
		dialog.WriteString("\n\n")
		dialog.WriteString(promptStyle.Render(truncateLine(m.promptPath, m.width-8)))
		dialog.WriteString("\n\n")

		options := []struct {
			label string
			key   string
			style lipgloss.Style
			all   bool
		}{
			{label: "Overwrite", key: "o", style: promptOverwriteStyle},
			{label: "Overwrite All", key: "O", style: promptOverwriteAllStyle, all: true},
			{label: "Skip", key: "s", style: promptSkipStyle},
			{label: "Skip All", key: "S", style: promptSkipAllStyle, all: true},
			{label: "Rename", key: "r", style: promptRenameStyle},
			{label: "Rename All", key: "R", style: promptRenameAllStyle, all: true},
			{label: "Quit", key: "q", style: promptQuitStyle},
		}

		for i, opt := range options {
			text := fmt.Sprintf("[%s] %s", opt.key, opt.label)
			if opt.all {
				text = fmt.Sprintf("[%s] %s (all)", opt.key, opt.label)
			}
			if i == m.promptIndex {
				dialog.WriteString(promptSelectedStyle.Render(text))
			} else {
				dialog.WriteString(promptOptionStyle.Render(opt.style.Render(text)))
			}
			dialog.WriteString("\n")
		}
		dialog.WriteString("\n")
		dialog.WriteString(promptStyle.Render("Use ↑/↓ + Enter or press the hotkey."))

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF006E")).
			Padding(1, 2).
			Render(dialog.String())

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}

	// Render downloads through viewport for scrolling
	if len(m.order) > 0 {
		var taskContent strings.Builder
		for _, id := range m.order {
			task, ok := m.tasks[id]
			if !ok {
				continue
			}

			var elapsed time.Duration
			var eta time.Duration
			var rate string

			if task.done {
				elapsed = task.finished.Sub(task.started)
				eta = 0
				rate = formatRate(task.current, elapsed)
			} else {
				elapsed = time.Since(task.started)
				eta = estimateETA(task.current, task.total, elapsed)
				rate = formatRate(task.current, elapsed)
			}

			percentText := percentStyle.Render(fmt.Sprintf("%5.1f%%", task.percent*100))
			labelText := labelStyle.Render(task.label)

			spinText := ""
			if !task.done {
				spinText = task.spin.View()
				if spinText != "" {
					spinText = spinnerStyle.Render(spinText)
				}
			}
			taskContent.WriteString(fmt.Sprintf("%s %s %s\n", spinText, percentText, labelText))

			bar := task.bar.View()
			taskContent.WriteString(progressBarStyle.Render(bar))
			taskContent.WriteString("\n")

			bytesLine := fmt.Sprintf("%s / %s · %s",
				humanBytes(task.current),
				humanBytes(task.total),
				rate,
			)
			taskContent.WriteString(fmt.Sprintf("        %s\n", etaStyle.Render(bytesLine)))

			var etaText string
			if task.done {
				etaText = etaStyle.Render(fmt.Sprintf("completed in %s", formatDurationShort(elapsed)))
			} else {
				etaText = etaStyle.Render(fmt.Sprintf("elapsed %s · eta %s",
					formatDurationShort(elapsed),
					formatDurationShort(eta)))
			}
			taskContent.WriteString(fmt.Sprintf("        %s\n", etaText))
		}

		m.vp.SetContent(taskContent.String())
		b.WriteString(titleStyle.Render(" Downloads"))
		b.WriteString(" ")
		b.WriteString(etaStyle.Render(fmt.Sprintf("(↑/↓ scroll, %d items)", len(m.order))))
		b.WriteString("\n")
		b.WriteString(m.vp.View())
	}

	return b.String()
}

func formatRate(current int64, elapsed time.Duration) string {
	if elapsed <= 0 {
		return "--/s"
	}
	rate := int64(float64(current) / elapsed.Seconds())
	if rate <= 0 {
		return "--/s"
	}
	return humanBytes(rate) + "/s"
}

func promptChoiceForIndex(index int) promptChoice {
	switch index {
	case 0:
		return promptOverwrite
	case 1:
		return promptOverwriteAll
	case 2:
		return promptSkip
	case 3:
		return promptSkipAll
	case 4:
		return promptRename
	case 5:
		return promptRenameAll
	default:
		return promptQuit
	}
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
