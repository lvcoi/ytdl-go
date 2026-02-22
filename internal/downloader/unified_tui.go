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
	"github.com/lvcoi/ytdl-lib/v2"
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
	mu            sync.Mutex
	program       *tea.Program
	model         *seamlessModel
	selectionChan chan int      // Receives selection when user picks a format (doesn't quit TUI)
	doneChan      chan struct{} // Closed when TUI program exits
	ctx           context.Context
	cancel        context.CancelFunc
}

type seamlessModel struct {
	view           SeamlessView
	formatSelector *formatSelectorModel
	progress       *seamlessProgressModel
	selectedItag   int
	width          int
	height         int
	quitting       bool
	selectionChan  chan int // Channel to signal selection without quitting TUI
}

type seamlessProgressModel struct {
	tasks        map[string]*seamlessTask
	order        []string
	log          string
	vp           viewport.Model
	vpReady      bool
	width        int
	height       int
	promptActive bool
	promptPath   string
	promptResp   chan promptChoice
	promptQueue  []seamlessPromptMsg
	promptIndex  int
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

type seamlessPromptMsg struct {
	path string
	resp chan promptChoice
}

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
		model:         model,
		selectionChan: make(chan int, 1),
		doneChan:      make(chan struct{}),
	}
}

// Start begins the seamless TUI
func (st *SeamlessTUI) Start(ctx context.Context) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.ctx, st.cancel = context.WithCancel(ctx)

	// Pass selectionChan to model so it can signal selection without quitting
	st.model.selectionChan = st.selectionChan

	st.program = tea.NewProgram(st.model,
		tea.WithAltScreen(),
		tea.WithOutput(os.Stderr),
	)

	go func() {
		st.program.Run()
		close(st.doneChan)
	}()

	go func() {
		<-st.ctx.Done()
		st.Send(seamlessStopMsg{})
	}()
}

// WaitForSelection blocks until user selects a format, returns the itag (0 if cancelled)
func (st *SeamlessTUI) WaitForSelection() int {
	select {
	case itag := <-st.selectionChan:
		return itag
	case <-st.doneChan:
		// TUI quit without selection (user cancelled)
		return 0
	}
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

// PromptDuplicate shows a duplicate file prompt and returns the user's choice
func (st *SeamlessTUI) PromptDuplicate(path string) (promptChoice, error) {
	if st == nil {
		return promptQuit, errors.New("no seamless TUI")
	}
	resp := make(chan promptChoice, 1)
	st.Send(seamlessPromptMsg{path: path, resp: resp})
	if st.ctx == nil {
		choice := <-resp
		return choice, nil
	}
	select {
	case choice := <-resp:
		return choice, nil
	case <-st.ctx.Done():
		return promptQuit, st.ctx.Err()
	}
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
			// Signal selection through channel (non-blocking)
			if m.selectionChan != nil {
				select {
				case m.selectionChan <- m.selectedItag:
				default:
				}
			}
			// Transition to progress view instead of quitting
			m.view = SeamlessViewProgress
			m.progress.width = m.width
			m.progress.height = m.height
			// Return a WindowSizeMsg to initialize the progress viewport
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
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
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.progress.promptActive {
			return m.handlePromptKey(msg)
		}
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "up", "k":
			m.progress.vp.SetYOffset(m.progress.vp.YOffset - 1)
		case "down", "j":
			m.progress.vp.SetYOffset(m.progress.vp.YOffset + 1)
		}
	case seamlessPromptMsg:
		if m.progress.promptActive {
			m.progress.promptQueue = append(m.progress.promptQueue, msg)
			return m, nil
		}
		m.progress.promptActive = true
		m.progress.promptPath = msg.path
		m.progress.promptResp = msg.resp
		m.progress.promptIndex = 0
		return m, nil
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

func (m *seamlessModel) handlePromptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	choice := promptChoice(-1)
	switch msg.String() {
	case "up", "k":
		if m.progress.promptIndex > 0 {
			m.progress.promptIndex--
		} else {
			m.progress.promptIndex = 6
		}
		return m, nil
	case "down", "j", "tab":
		if m.progress.promptIndex < 6 {
			m.progress.promptIndex++
		} else {
			m.progress.promptIndex = 0
		}
		return m, nil
	case "enter":
		choice = promptChoiceForIndex(m.progress.promptIndex)
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
		if m.progress.promptResp != nil {
			select {
			case m.progress.promptResp <- choice:
			default:
			}
		}
		// Handle "all" choices for queued prompts
		if choice == promptOverwriteAll || choice == promptSkipAll || choice == promptRenameAll || choice == promptQuit {
			for _, queued := range m.progress.promptQueue {
				if queued.resp == nil {
					continue
				}
				select {
				case queued.resp <- choice:
				default:
				}
			}
			m.progress.promptQueue = nil
		}
		m.progress.promptActive = false
		m.progress.promptPath = ""
		m.progress.promptResp = nil
		m.progress.promptIndex = 0
		// Process next queued prompt if any
		if len(m.progress.promptQueue) > 0 {
			next := m.progress.promptQueue[0]
			m.progress.promptQueue = m.progress.promptQueue[1:]
			m.progress.promptActive = true
			m.progress.promptPath = next.path
			m.progress.promptResp = next.resp
			m.progress.promptIndex = 0
		}
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
	// If prompt is active, render only the centered modal
	if m.progress.promptActive {
		return m.renderPrompt()
	}

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
				etaText = etaStyle.Render(fmt.Sprintf("completed in %s", formatDurationShort(elapsed)))
			} else {
				etaText = etaStyle.Render(fmt.Sprintf("elapsed %s · eta %s",
					formatDurationShort(elapsed),
					formatDurationShort(eta)))
			}
			taskContent.WriteString(fmt.Sprintf("        %s\n", etaText))
		}

		m.progress.vp.SetContent(taskContent.String())
		b.WriteString(m.progress.vp.View())
	}

	return b.String()
}

func (m *seamlessModel) renderPrompt() string {
	// Build the modal content
	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF006E")).
		Render("File Already Exists")
	content.WriteString(title)
	content.WriteString("\n\n")

	// File path (truncated if needed)
	maxPathWidth := 40
	path := m.progress.promptPath
	if len(path) > maxPathWidth {
		path = "..." + path[len(path)-maxPathWidth+3:]
	}
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7FDBFF")).
		Render(path))
	content.WriteString("\n\n")

	options := []string{
		"[o] Overwrite",
		"[O] Overwrite all",
		"[s] Skip",
		"[S] Skip all",
		"[r] Rename",
		"[R] Rename all",
		"[q] Quit",
	}

	for i, opt := range options {
		if i == m.progress.promptIndex {
			content.WriteString(selectorSelectedStyle.Render(opt))
		} else {
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A6ADC8")).
				Render(opt))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("↑/↓ select · Enter confirm"))

	// Create modal box style
	modalWidth := 46
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF006E")).
		Padding(1, 2).
		Width(modalWidth)

	modal := modalStyle.Render(content.String())

	// Center the modal horizontally and vertically
	modalHeight := lipgloss.Height(modal)
	modalRenderedWidth := lipgloss.Width(modal)

	// Calculate centering
	horizontalPadding := 0
	if m.progress.width > modalRenderedWidth {
		horizontalPadding = (m.progress.width - modalRenderedWidth) / 2
	}
	verticalPadding := 0
	if m.progress.height > modalHeight {
		verticalPadding = (m.progress.height - modalHeight) / 2
	}

	// Build centered output
	var out strings.Builder

	// Add vertical padding (blank lines)
	for i := 0; i < verticalPadding; i++ {
		out.WriteString(strings.Repeat(" ", m.progress.width))
		out.WriteString("\n")
	}

	// Add horizontal padding to each line of the modal
	lines := strings.Split(modal, "\n")
	padding := strings.Repeat(" ", horizontalPadding)
	for _, line := range lines {
		out.WriteString(padding)
		out.WriteString(line)
		out.WriteString("\n")
	}

	return out.String()
}

func seamlessBarWidth(total int) int {
	return barWidth(total)
}
func seamlessFormatRate(current int64, elapsed time.Duration) string {
	return formatRate(current, elapsed)
}

func seamlessEstimateETA(current, total int64, elapsed time.Duration) time.Duration {
	return estimateETA(current, total, elapsed)
}

// seamlessProgressRenderer implements the same interface as progressRenderer
// but routes updates to SeamlessTUI instead of ProgressManager
type seamlessProgressRenderer struct {
	tui *SeamlessTUI
}

func (spr *seamlessProgressRenderer) Register(prefix string, size int64) string {
	if spr == nil || spr.tui == nil {
		return ""
	}
	return spr.tui.Register(prefix, size)
}

func (spr *seamlessProgressRenderer) Update(id string, current, total int64) {
	if spr == nil || spr.tui == nil {
		return
	}
	spr.tui.Update(id, current, total)
}

func (spr *seamlessProgressRenderer) Finish(id string) {
	if spr == nil || spr.tui == nil {
		return
	}
	spr.tui.Finish(id)
}

func (spr *seamlessProgressRenderer) Log(level LogLevel, msg string) {
	if spr == nil || spr.tui == nil {
		return
	}
	spr.tui.Log(level, msg)
}

// RunSeamlessFormatSelector displays formats in an interactive selector and returns
// the selected itag along with the SeamlessTUI for continued progress display.
// The caller should call tui.TransitionToProgress() after selection, then use
// the returned printer for download progress, and finally call tui.Stop() when done.
func RunSeamlessFormatSelector(ctx context.Context, video *youtube.Video, title string, playlistID, playlistTitle string, index, total int) (selectedItag int, tui *SeamlessTUI, err error) {
	tui = NewSeamlessTUI(video, title)

	// Update format selector with playlist context
	tui.model.formatSelector.playlistID = playlistID
	tui.model.formatSelector.playlistTitle = playlistTitle
	tui.model.formatSelector.index = index
	tui.model.formatSelector.total = total

	tui.Start(ctx)
	selectedItag = tui.WaitForSelection()

	if selectedItag == 0 {
		// User cancelled - stop TUI and return
		tui.Stop()
		return 0, nil, nil
	}

	return selectedItag, tui, nil
}

// NewSeamlessPrinter creates a Printer that routes progress to the SeamlessTUI
func NewSeamlessPrinter(opts Options, tui *SeamlessTUI) *Printer {
	columns := terminalColumns()
	if columns <= 0 {
		columns = 100
	}

	titleWidth := columns - 44
	if titleWidth < 20 {
		titleWidth = 20
	}
	if titleWidth > 60 {
		titleWidth = 60
	}

	renderer := &seamlessProgressRenderer{tui: tui}

	return &Printer{
		quiet:           opts.Quiet,
		color:           supportsColor(),
		columns:         columns,
		titleWidth:      titleWidth,
		logLevel:        parseLogLevel(opts.LogLevel),
		progressEnabled: !opts.Quiet,
		interactive:     !opts.Quiet,
		layout:          opts.ProgressLayout,
		renderer:        renderer,
		seamlessTUI:     tui,
	}
}
