package downloader

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	pagerTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0B0B0B")).
			Background(lipgloss.Color("#7FDBFF")).
			Bold(true).
			Padding(0, 1)

	pagerHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6ADC8")).
			Faint(true)
)

type pagerModel struct {
	viewport viewport.Model
	title    string
	ready    bool
	width    int
	height   int
}

func newPagerModel(title, content string) *pagerModel {
	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7FDBFF"))
	vp.SetContent(content)
	return &pagerModel{
		viewport: vp,
		title:    title,
		width:    80,
		height:   24,
	}
}

func (m *pagerModel) Init() tea.Cmd {
	return nil
}

func (m *pagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 2
		borderHeight := 2
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - headerHeight - borderHeight
		m.viewport, cmd = m.viewport.Update(msg)
		m.ready = true
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.viewport.SetYOffset(m.viewport.YOffset - 1)
		case "down", "j":
			m.viewport.SetYOffset(m.viewport.YOffset + 1)
		case "pgup", "b":
			m.viewport.HalfViewUp()
		case "pgdown", "f", " ":
			m.viewport.HalfViewDown()
		case "home", "g":
			m.viewport.GotoTop()
		case "end", "G":
			m.viewport.GotoBottom()
		}
		return m, nil
	case tea.MouseMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *pagerModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString(pagerTitleStyle.Render(m.title))
	b.WriteString(" ")
	scrollPercent := m.viewport.ScrollPercent() * 100
	b.WriteString(pagerHelpStyle.Render(fmt.Sprintf("%.0f%% · ↑/↓/pgup/pgdn scroll · q quit", scrollPercent)))
	b.WriteString("\n")
	b.WriteString(m.viewport.View())
	return b.String()
}

// RunPager displays content in a scrollable pager.
func RunPager(title, content string) error {
	model := newPagerModel(title, content)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	_, err := p.Run()
	return err
}
