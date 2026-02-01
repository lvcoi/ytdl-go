package downloader

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkdai/youtube/v2"
)

var (
	selectorTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0B0B0B")).
				Background(lipgloss.Color("#7FDBFF")).
				Bold(true).
				Padding(0, 1)

	selectorHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A6ADC8")).
				Faint(true)

	selectorHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")).
				Bold(true)

	selectorSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0B0B0B")).
				Background(lipgloss.Color("#00F5D4")).
				Bold(true)

	selectorFormatStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EAEAEA"))
)

const (
	digitBufferTimeout = 1500 * time.Millisecond
)

type formatSelectorModel struct {
	viewport      viewport.Model
	title         string
	ready         bool
	width         int
	height        int
	formats       []youtube.Format
	selected      int
	video         *youtube.Video
	playlistID    string
	playlistTitle string
	index         int
	total         int
	quitting      bool
	digitBuffer   string
	lastDigitTime time.Time
	lastDigit     string
}

type quitMsg struct{}

type digitBufferExpireMsg struct {
	expireTime time.Time
}

func newFormatSelectorModel(video *youtube.Video, title string, playlistID, playlistTitle string, index, total int) *formatSelectorModel {
	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7FDBFF"))

	// Sort formats by itag
	formats := make([]youtube.Format, len(video.Formats))
	copy(formats, video.Formats)
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].ItagNo < formats[j].ItagNo
	})

	content := buildFormatContent(formats, -1)
	vp.SetContent(content)

	return &formatSelectorModel{
		viewport:      vp,
		title:         title,
		width:         80,
		height:        24,
		formats:       formats,
		selected:      -1,
		video:         video,
		playlistID:    playlistID,
		playlistTitle: playlistTitle,
		index:         index,
		total:         total,
		quitting:      false,
		digitBuffer:   "",
		lastDigitTime: time.Time{},
		lastDigit:     "",
	}
}

func buildFormatContent(formats []youtube.Format, selected int) string {
	var b strings.Builder

	// Header
	b.WriteString(selectorHeaderStyle.Render("itag   ext    quality      size       audio   video"))
	b.WriteString("\n")

	for i, f := range formats {
		size := "-"
		if f.ContentLength > 0 {
			size = humanBytes(int64(f.ContentLength))
		}
		audio := "-"
		if f.AudioChannels > 0 {
			audio = fmt.Sprintf("%dch", f.AudioChannels)
		}
		videoRes := "-"
		if f.Width > 0 || f.Height > 0 {
			videoRes = fmt.Sprintf("%dx%d", f.Width, f.Height)
		}
		qual := f.QualityLabel
		if qual == "" {
			qual = f.Quality
		}

		line := fmt.Sprintf("%5d   %-5s  %-12s %-10s %-7s %s",
			f.ItagNo,
			mimeToExt(f.MimeType),
			qual,
			size,
			audio,
			videoRes,
		)

		if i == selected {
			line = selectorSelectedStyle.Render(line)
		} else {
			line = selectorFormatStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *formatSelectorModel) Init() tea.Cmd {
	return nil
}

// resetDigitBufferIfExpired clears the digit buffer if enough time has passed since the last digit
func (m *formatSelectorModel) resetDigitBufferIfExpired() {
	if !m.lastDigitTime.IsZero() && time.Since(m.lastDigitTime) > digitBufferTimeout {
		m.digitBuffer = ""
		m.lastDigit = ""
	}
}

func scheduleDigitBufferExpiry(expireTime time.Time) tea.Cmd {
	return tea.Tick(digitBufferTimeout, func(time.Time) tea.Msg {
		return digitBufferExpireMsg{expireTime: expireTime}
	})
}

func quitAfterDelay() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
		return quitMsg{}
	})
}

func (m *formatSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If quitting, ignore all messages except window size and quitMsg
	if m.quitting {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			return m, nil
		case quitMsg:
			return m, tea.Quit
		default:
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 2
		borderHeight := 2
		helpHeight := 2
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - headerHeight - borderHeight - helpHeight
		m.viewport, cmd = m.viewport.Update(msg)
		m.ready = true
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			m.selected = -1
			return m, quitAfterDelay()
		case "b", "back":
			m.quitting = true
			m.selected = -1
			return m, quitAfterDelay()
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			} else if len(m.formats) > 0 {
				m.selected = len(m.formats) - 1
			}
			m.updateContent()
		case "down", "j":
			if m.selected < len(m.formats)-1 {
				m.selected++
			} else if m.selected >= len(m.formats)-1 && len(m.formats) > 0 {
				m.selected = 0
			}
			m.updateContent()
		case "pgup":
			// Move selection up by 10
			if m.selected >= 10 {
				m.selected -= 10
			} else if len(m.formats) > 0 {
				m.selected = 0
			}
			m.updateContent()
		case "pgdown", "f":
			// Move selection down by 10
			if m.selected < len(m.formats)-10 {
				m.selected += 10
			} else if len(m.formats) > 0 {
				m.selected = len(m.formats) - 1
			}
			m.updateContent()
		case "home", "g":
			if len(m.formats) > 0 {
				m.selected = 0
			}
			m.updateContent()
		case "end", "G":
			if len(m.formats) > 0 {
				m.selected = len(m.formats) - 1
			}
			m.updateContent()
		case "enter":
			if m.selected >= 0 && m.selected < len(m.formats) {
				m.quitting = true
				return m, quitAfterDelay()
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Multi-digit itag selection with cycling support
			m.resetDigitBufferIfExpired()
			
			digit := msg.String()
			now := time.Now()
			
			// Check if this is a repeated single digit press (for cycling)
			if m.digitBuffer == digit && m.lastDigit == digit && len(digit) == 1 {
				// Cycle to next matching format
				matches := []int{}
				for i, f := range m.formats {
					itagStr := strconv.Itoa(f.ItagNo)
					if strings.HasPrefix(itagStr, digit) {
						matches = append(matches, i)
					}
				}
				
				if len(matches) > 0 {
					// Find current position in matches and move to next
					currentPos := -1
					for i, idx := range matches {
						if idx == m.selected {
							currentPos = i
							break
						}
					}
					
					// If not found in matches, start from first; otherwise move to next
					var nextPos int
					if currentPos == -1 {
						nextPos = 0
					} else {
						nextPos = (currentPos + 1) % len(matches)
					}
					m.selected = matches[nextPos]
					m.lastDigit = digit
					m.lastDigitTime = now
					m.updateContent()
					return m, scheduleDigitBufferExpiry(now)
				}
			} else {
				// New digit or different digit - reset cycling
				m.digitBuffer += digit
				m.lastDigit = digit
				m.lastDigitTime = now
				
				// Try to find exact match first
				targetItag, err := strconv.Atoi(m.digitBuffer)
				exactMatch := -1
				if err == nil {
					for i, f := range m.formats {
						if f.ItagNo == targetItag {
							exactMatch = i
							break
						}
					}
				}
				
				if exactMatch >= 0 {
					// Found exact match - select it and clear buffer
					m.selected = exactMatch
					m.digitBuffer = ""
					m.lastDigit = ""
					m.updateContent()
					return m, nil
				} else {
					// No exact match yet - find first format with itag starting with buffer
					found := false
					for i, f := range m.formats {
						itagStr := strconv.Itoa(f.ItagNo)
						if strings.HasPrefix(itagStr, m.digitBuffer) {
							m.selected = i
							found = true
							break
						}
					}
					if found {
						m.updateContent()
						return m, scheduleDigitBufferExpiry(now)
					} else {
						// No match - reset buffer
						m.digitBuffer = ""
						m.lastDigit = ""
						return m, nil
					}
				}
			}
		}
		return m, nil
	case digitBufferExpireMsg:
		// Only clear if this expiry matches the current lastDigitTime
		if !m.lastDigitTime.IsZero() && msg.expireTime.Equal(m.lastDigitTime) {
			m.digitBuffer = ""
			m.lastDigit = ""
		}
		return m, nil
	case tea.MouseMsg:
		// Allow mouse scrolling even when selection is active
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case quitMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m *formatSelectorModel) updateContent() {
	content := buildFormatContent(m.formats, m.selected)
	m.viewport.SetContent(content)

	// Auto-scroll to keep selection in view
	if m.selected >= 0 {
		// Each line is approximately 1 line height in the viewport
		// Account for header line (1) + some padding
		lineHeight := 1
		headerOffset := 1
		targetLine := headerOffset + (m.selected * lineHeight)

		// Get current viewport position
		viewportTop := m.viewport.YOffset
		viewportBottom := viewportTop + m.viewport.Height - 2 // -2 for borders/padding

		// Scroll up if selection is above viewport
		if targetLine < viewportTop {
			m.viewport.YOffset = targetLine
		} else if targetLine >= viewportBottom {
			// Scroll down if selection is below viewport
			m.viewport.YOffset = targetLine - m.viewport.Height + 3
		}
	}
}

func (m *formatSelectorModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	var b strings.Builder
	b.WriteString(selectorTitleStyle.Render(m.title))
	b.WriteString(" ")

	if m.quitting {
		if m.selected >= 0 && m.selected < len(m.formats) {
			selectedFormat := m.formats[m.selected]
			b.WriteString(selectorHelpStyle.Render(fmt.Sprintf("Selected: itag %d ✓", selectedFormat.ItagNo)))
		} else {
			b.WriteString(selectorHelpStyle.Render("Cancelled"))
		}
	} else if m.digitBuffer != "" {
		// Show digit buffer when typing
		b.WriteString(selectorHelpStyle.Render(fmt.Sprintf("Typing itag: %s_", m.digitBuffer)))
	} else if m.selected >= 0 && m.selected < len(m.formats) {
		selectedFormat := m.formats[m.selected]
		b.WriteString(selectorHelpStyle.Render(fmt.Sprintf("Selected: itag %d · Enter to download", selectedFormat.ItagNo)))
	} else {
		b.WriteString(selectorHelpStyle.Render("↑/↓ select · Enter download · q quit · b back"))
	}
	b.WriteString("\n")
	b.WriteString(m.viewport.View())

	// Help line at bottom
	b.WriteString("\n")
	if !m.quitting {
		b.WriteString(selectorHelpStyle.Render("Type digits to select itag (e.g., 101), Home/End for first/last, b to go back"))
	}

	return b.String()
}

func (m *formatSelectorModel) SelectedItag() int {
	if m.selected >= 0 && m.selected < len(m.formats) {
		return m.formats[m.selected].ItagNo
	}
	return 0
}

// RunFormatSelector displays formats in an interactive selector and returns the selected itag.
func RunFormatSelector(video *youtube.Video, title string, playlistID, playlistTitle string, index, total int) (int, error) {
	model := newFormatSelectorModel(video, title, playlistID, playlistTitle, index, total)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return 0, err
	}

	if m, ok := result.(*formatSelectorModel); ok {
		return m.SelectedItag(), nil
	}

	return 0, nil
}
