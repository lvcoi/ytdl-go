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
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkdai/youtube/v2"
)

// SeamlessView represents the current view in the seamless TUI
type SeamlessView int

const (
	SeamlessViewFormatSelector SeamlessView = iota
	SeamlessViewProgress
)

// SeamlessTUI provides a unified TUI that handles format selection and progress
// display in a single program, enabling seamless transitions without alt screen flicker.
type SeamlessTUI struct {
	mu         sync.Mutex
	program    *tea.Program
	model      *seamlessModel
	resultChan chan int
	ctx        context.Context
	cancel     context.CancelFunc
}

type seamlessModel struct {
	view           SeamlessView
	formatSelector *formatSelectorModel
	progress       *seamlessProgressModel
	selectedItag   int
	width          int
	height         int
	quitting       bool
}

type seamlessProgressModel struct {
	tasks   map[string]*seamlessTask
	order   []string
	log     string
	vp      viewport.Model
	vpReady bool
	width   int
	height  int
}

type seamlessTask struct {
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

// Messages for seamless TUI
type seamlessRegisterMsg struct {
	id    string
	label string
	total int64
	start time.Time
}

type seamlessUpdateMsg struct {
	id      string
	current int64
	total   int64
}

type seamlessFinishMsg struct {
	id string
}

type seamlessLogMsg struct {
	level LogLevel
	text  string
}

type seamlessStopMsg struct{}

type seamlessTransitionMsg struct{}

// NewSeamlessTUI creates a new seamless TUI for format selection with progress display
func NewSeamlessTUI(video *youtube.Video, title string) *SeamlessTUI {
	formatSelector := newFormatSelectorModel(video, title, "", "", 0, 0)

	progressVP := viewport.New(80, 20)
	progressVP.MouseWheelEnabled = true
	progressVP.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7FDBFF"))

	model := &seamlessModel{
		view:           SeamlessViewFormatSelector,
		formatSelector: formatSelector,
		progress: &seamlessProgressModel{
			tasks: make(map[string]*seamlessTask),
			order: make([]string, 0),
			vp:    progressVP,
		},
	}

	return &SeamlessTUI{
		model:      model,
		resultChan: make(chan int, 1),
	}
}

// Start begins the seamless TUI
func (st *SeamlessTUI) Start(ctx context.Context) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.ctx, st.cancel = context.WithCancel(ctx)
	st.program = tea.NewProgram(st.model,
		tea.WithAltScreen(),
		tea.WithOutput(os.Stderr),
	)

	go func() {
		result, _ := st.program.Run()
		if m, ok := result.(*seamlessModel); ok && m.selectedItag > 0 {
			st.resultChan <- m.selectedItag
		} else {
			st.resultChan <- 0
		}
		close(st.resultChan)
	}()

	go func() {
		<-st.ctx.Done()
		st.Send(seamlessStopMsg{})
	}()
}

// WaitForSelection blocks until user selects a format, returns the itag (0 if cancelled)
func (st *SeamlessTUI) WaitForSelection() int {
	return <-st.resultChan
}

// TransitionToProgress switches from format selector to progress view
func (st *SeamlessTUI) TransitionToProgress() {
	st.Send(seamlessTransitionMsg{})
}

// Send sends a message to the TUI
func (st *SeamlessTUI) Send(msg tea.Msg) {
	st.mu.Lock()
	p := st.program
	st.mu.Unlock()
	if p != nil {
		p.Send(msg)
	}
}

// Register registers a new download task
func (st *SeamlessTUI) Register(prefix string, size int64) string {
	id := fmt.Sprintf("%s@%d", prefix, time.Now().UnixNano())
	st.Send(seamlessRegisterMsg{
		id:    id,
		label: prefix,
		total: size,
		start: time.Now(),
	})
	return id
}

// Update updates a download task's progress
func (st *SeamlessTUI) Update(id string, current, total int64) {
	st.Send(seamlessUpdateMsg{id: id, current: current, total: total})
}

// Finish marks a download task as complete
func (st *SeamlessTUI) Finish(id string) {
	st.Send(seamlessFinishMsg{id: id})
}

// Log displays a log message
func (st *SeamlessTUI) Log(level LogLevel, msg string) {
	st.Send(seamlessLogMsg{level: level, text: msg})
}

// Stop stops the TUI
func (st *SeamlessTUI) Stop() {
	st.mu.Lock()
	cancel := st.cancel
	st.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Model implementation

func (m *seamlessModel) Init() tea.Cmd {
	return nil
}

func (m *seamlessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size for all views
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
	}

	// Handle global messages
	switch msg.(type) {
	case seamlessStopMsg:
		m.quitting = true
		return m, tea.Quit
	case seamlessTransitionMsg:
		if m.view == SeamlessViewFormatSelector && m.selectedItag > 0 {
			m.view = SeamlessViewProgress
			m.progress.width = m.width
			m.progress.height = m.height
		}
		return m, nil
	}

	// Route based on current view
	switch m.view {
	case SeamlessViewFormatSelector:
		return m.updateFormatSelector(msg)
	case SeamlessViewProgress:
		return m.updateProgress(msg)
	}

	return m, nil
}

func (m *seamlessModel) updateFormatSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle format selector's quit message - capture selection and transition
	if _, ok := msg.(quitMsg); ok {
		if m.formatSelector.selected >= 0 && m.formatSelector.selected < len(m.formatSelector.formats) {
			m.selectedItag = m.formatSelector.formats[m.formatSelector.selected].ItagNo
			// Transition to progress view instead of quitting
			m.view = SeamlessViewProgress
			m.progress.width = m.width
			m.progress.height = m.height
			return m, nil
		}
		// User cancelled
		return m, tea.Quit
	}

	// Forward to format selector
	updated, cmd := m.formatSelector.Update(msg)
	if fs, ok := updated.(*formatSelectorModel); ok {
		m.formatSelector = fs
	}
	return m, cmd
}

func (m *seamlessModel) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.progress.vp.SetYOffset(m.progress.vp.YOffset - 1)
		case "down", "j":
			m.progress.vp.SetYOffset(m.progress.vp.YOffset + 1)
		}
	case tea.WindowSizeMsg:
		m.progress.width = msg.Width
		m.progress.height = msg.Height
		headerHeight := 2
		borderHeight := 2
		m.progress.vp.Width = msg.Width - 2
		m.progress.vp.Height = msg.Height - headerHeight - borderHeight
		m.progress.vp, _ = m.progress.vp.Update(msg)
		m.progress.vpReady = true
	case seamlessRegisterMsg:
		if _, exists := m.progress.tasks[msg.id]; exists {
			return m, nil
		}
		m.progress.order = append(m.progress.order, msg.id)
		spin := spinner.New()
		spin.Spinner = spinner.MiniDot
		spin.Style = spinnerStyle
		bar := progressbar.New(
			progressbar.WithGradient("#FF006E", "#00F5FF"),
			progressbar.WithWidth(seamlessBarWidth(m.progress.width)),
			progressbar.WithoutPercentage(),
		)
		task := &seamlessTask{
			id:      msg.id,
			label:   msg.label,
			total:   msg.total,
			started: msg.start,
			bar:     bar,
			spin:    spin,
		}
		m.progress.tasks[msg.id] = task
		cmds = append(cmds, task.bar.SetPercent(0), task.spin.Tick)
	case seamlessUpdateMsg:
		if task, ok := m.progress.tasks[msg.id]; ok {
			task.current = msg.current
			if msg.total > 0 {
				task.total = msg.total
			}
			if task.total > 0 {
				task.percent = math.Min(1, math.Max(0, float64(task.current)/float64(task.total)))
				cmds = append(cmds, task.bar.SetPercent(task.percent))
			}
		}
	case seamlessFinishMsg:
		if task, ok := m.progress.tasks[msg.id]; ok {
			task.percent = 1
			task.done = true
			task.finished = time.Now()
			cmds = append(cmds, task.bar.SetPercent(1))
		}
	case seamlessLogMsg:
		var style lipgloss.Style
		switch msg.level {
		case LogError:
			style = logErrorStyle
		case LogWarn:
			style = logWarnStyle
		default:
			style = logInfoStyle
		}
		m.progress.log = style.Render(truncateLine(msg.text, m.progress.width))
	case progressbar.FrameMsg:
		for _, task := range m.progress.tasks {
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
	case spinner.TickMsg:
		for _, task := range m.progress.tasks {
			if task == nil {
				continue
			}
			updated, cmd := task.spin.Update(msg)
			task.spin = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *seamlessModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.view {
	case SeamlessViewFormatSelector:
		return m.formatSelector.View()
	case SeamlessViewProgress:
		return m.viewProgress()
	}
	return ""
}

func (m *seamlessModel) viewProgress() string {
	var b strings.Builder

	// Log message
	if m.progress.log != "" {
		b.WriteString(m.progress.log)
		b.WriteString("\n")
	}

	// Title with selected itag
	b.WriteString(titleStyle.Render(fmt.Sprintf(" Downloading itag %d ", m.selectedItag)))
	b.WriteString(" ")
	b.WriteString(etaStyle.Render(fmt.Sprintf("(%d items)", len(m.progress.order))))
	b.WriteString("\n")

	// Render tasks
	if len(m.progress.order) == 0 {
		b.WriteString(labelStyle.Render("Waiting for download to start..."))
		b.WriteString("\n")
	} else {
		var taskContent strings.Builder
		for _, id := range m.progress.order {
			task, ok := m.progress.tasks[id]
			if !ok {
				continue
			}

			var elapsed time.Duration
			var eta time.Duration
			var rate string

			if task.done {
				elapsed = task.finished.Sub(task.started)
				eta = 0
				rate = seamlessFormatRate(task.current, elapsed)
			} else {
				elapsed = time.Since(task.started)
				eta = seamlessEstimateETA(task.current, task.total, elapsed)
				rate = seamlessFormatRate(task.current, elapsed)
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
			taskContent.WriteString(progressBarStyle.Render(task.bar.View()))
			taskContent.WriteString("\n")

			bytesLine := fmt.Sprintf("%s / %s · %s",
				humanBytes(task.current),
				humanBytes(task.total),
				rate,
			)
			taskContent.WriteString(fmt.Sprintf("        %s\n", etaStyle.Render(bytesLine)))

			var etaText string
			if task.done {
				etaText = etaStyle.Render(fmt.Sprintf("completed in %s", seamlessFormatDuration(elapsed)))
			} else {
				etaText = etaStyle.Render(fmt.Sprintf("elapsed %s · eta %s",
					seamlessFormatDuration(elapsed),
					seamlessFormatDuration(eta)))
			}
			taskContent.WriteString(fmt.Sprintf("        %s\n", etaText))
		}

		m.progress.vp.SetContent(taskContent.String())
		b.WriteString(m.progress.vp.View())
	}

	return b.String()
}

func seamlessBarWidth(total int) int {
	width := total - 10
	if width < 10 {
		return 10
	}
	return width
}

func seamlessFormatRate(current int64, elapsed time.Duration) string {
	if elapsed <= 0 {
		return "--/s"
	}
	rate := int64(float64(current) / elapsed.Seconds())
	if rate <= 0 {
		return "--/s"
	}
	return humanBytes(rate) + "/s"
}

func seamlessEstimateETA(current, total int64, elapsed time.Duration) time.Duration {
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

func seamlessFormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), math.Mod(d.Seconds(), 60))
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), math.Mod(d.Minutes(), 60))
}

// RunSeamlessFormatSelector runs the format selector with seamless transition to progress.
// It returns the selected itag and a SeamlessTUI that can be used for progress updates.
// If the user cancels, itag will be 0 and tui will be nil.
func RunSeamlessFormatSelector(ctx context.Context, video *youtube.Video, title string) (int, *SeamlessTUI) {
	tui := NewSeamlessTUI(video, title)
	tui.Start(ctx)
	itag := tui.WaitForSelection()
	if itag == 0 {
		return 0, nil
	}
	return itag, tui
}
