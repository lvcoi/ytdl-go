package downloader

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type rendererEvent struct {
	kind   string
	id     string
	prefix string
	total  int64
	delta  int64
	value  int64
	msg    string
	ack    chan struct{}
}

type progressBar struct {
	prefix   string
	total    int64
	current  int64
	start    time.Time
	finished bool
}

type ProgressRenderer struct {
	out         io.Writer
	printer     *Printer
	interactive bool
	events      chan rendererEvent
	bars        map[string]*progressBar
	order       []string
	lastBars    int
	seq         uint64
	mu          sync.Mutex
}

func newProgressRenderer(out io.Writer, printer *Printer) *ProgressRenderer {
	renderer := &ProgressRenderer{
		out:         out,
		printer:     printer,
		interactive: printer != nil && printer.interactive,
		events:      make(chan rendererEvent, 256),
		bars:        map[string]*progressBar{},
		order:       []string{},
	}
	go renderer.loop()
	return renderer
}

func (r *ProgressRenderer) Register(prefix string, total int64) string {
	id := fmt.Sprintf("bar-%d", atomic.AddUint64(&r.seq, 1))
	r.events <- rendererEvent{kind: "register", id: id, prefix: prefix, total: total}
	return id
}

func (r *ProgressRenderer) Update(id string, delta, value, total int64) {
	select {
	case r.events <- rendererEvent{kind: "update", id: id, delta: delta, value: value, total: total}:
	default:
	}
}

func (r *ProgressRenderer) Finish(id string) {
	r.events <- rendererEvent{kind: "finish", id: id}
}

func (r *ProgressRenderer) Log(msg string) {
	r.events <- rendererEvent{kind: "log", msg: msg}
}

func (r *ProgressRenderer) Flush() {
	ack := make(chan struct{})
	r.events <- rendererEvent{kind: "flush", ack: ack}
	<-ack
}

func (r *ProgressRenderer) loop() {
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	dirty := false
	for {
		select {
		case event := <-r.events:
			switch event.kind {
			case "register":
				r.handleRegister(event)
				dirty = true
			case "update":
				r.handleUpdate(event)
				dirty = true
			case "finish":
				r.handleFinish(event)
				dirty = true
			case "log":
				r.handleLog(event)
				dirty = false
			case "flush":
				r.render()
				if event.ack != nil {
					close(event.ack)
				}
				dirty = false
			}
		case <-ticker.C:
			if dirty {
				r.render()
				dirty = false
			}
		}
	}
}

func (r *ProgressRenderer) handleRegister(event rendererEvent) {
	if _, exists := r.bars[event.id]; exists {
		return
	}
	r.bars[event.id] = &progressBar{
		prefix:  event.prefix,
		total:   event.total,
		start:   time.Now(),
		current: 0,
	}
	r.order = append(r.order, event.id)
}

func (r *ProgressRenderer) handleUpdate(event rendererEvent) {
	bar := r.bars[event.id]
	if bar == nil {
		return
	}
	bar.current += event.delta
	if event.value > 0 {
		bar.current = event.value
	}
	if event.total > 0 {
		bar.total = event.total
	}
}

func (r *ProgressRenderer) handleFinish(event rendererEvent) {
	bar := r.bars[event.id]
	if bar == nil {
		return
	}
	bar.finished = true
	if bar.total > 0 && bar.current < bar.total {
		bar.current = bar.total
	}
	// Remove finished bar from display
	delete(r.bars, event.id)
	newOrder := make([]string, 0, len(r.order))
	for _, id := range r.order {
		if id != event.id {
			newOrder = append(newOrder, id)
		}
	}
	r.order = newOrder
}

func (r *ProgressRenderer) handleLog(event rendererEvent) {
	if r.interactive {
		r.clearBars()
		fmt.Fprintln(r.out, event.msg)
		r.render()
		return
	}
	if len(r.bars) > 0 {
		r.render()
	}
	fmt.Fprintln(r.out, event.msg)
	if len(r.bars) > 0 {
		r.render()
	}
}

func (r *ProgressRenderer) render() {
	if r.printer != nil {
		r.printer.refreshLayout()
	}

	lines := []string{}
	for _, id := range r.order {
		bar := r.bars[id]
		if bar == nil || bar.finished {
			continue
		}
		line := r.renderBar(bar)
		lines = append(lines, line)
	}

	if r.interactive {
		r.clearBars()
		for _, line := range lines {
			fmt.Fprintf(r.out, "\r\x1b[2K%s\n", line)
		}
		r.lastBars = len(lines)
		return
	}

	for _, line := range lines {
		fmt.Fprintln(r.out, line)
	}
	r.lastBars = len(lines)
}

func (r *ProgressRenderer) renderBar(bar *progressBar) string {
	elapsed := time.Since(bar.start)

	// Calculate percentage
	percent := 0.0
	if bar.total > 0 {
		percent = float64(bar.current) * 100 / float64(bar.total)
	}

	// Calculate speed
	speed := ""
	if elapsed.Seconds() > 0.5 {
		bytesPerSec := float64(bar.current) / elapsed.Seconds()
		speed = humanBytes(int64(bytesPerSec)) + "/s"
	}

	// Calculate ETA
	eta := ""
	if bar.current > 0 && bar.total > 0 && bar.current < bar.total {
		remaining := time.Duration(float64(elapsed) * (float64(bar.total-bar.current) / float64(bar.current)))
		eta = formatTimeCompact(remaining)
	}

	// Determine bar width based on terminal size
	columns := 100
	if r.printer != nil {
		r.printer.mu.RLock()
		columns = r.printer.columns
		r.printer.mu.RUnlock()
	}

	// Calculate available space for bar after prefix and stats
	// Format: "prefix ━━━━━ XX% SIZE SPEED ETA"
	prefixLen := len(bar.prefix)
	statsLen := 30 // approximate space for stats
	barWidth := columns - prefixLen - statsLen - 4
	if barWidth < 15 {
		barWidth = 15
	}
	if barWidth > 50 {
		barWidth = 50
	}

	filled := 0
	if bar.total > 0 {
		filled = int(float64(barWidth) * float64(bar.current) / float64(bar.total))
	}
	if filled > barWidth {
		filled = barWidth
	}

	// Colors for Rich-style gradient bar
	barColor := "\x1b[38;5;199m" // Bright magenta/pink
	dimColor := "\x1b[38;5;239m" // Dark gray for unfilled
	reset := "\x1b[0m"
	cyan := "\x1b[36m"
	green := "\x1b[32m"
	yellow := "\x1b[33m"

	barStr := barColor + strings.Repeat("━", filled) + reset + dimColor + strings.Repeat("━", barWidth-filled) + reset

	// Build stats string with colors
	var stats strings.Builder
	stats.WriteString(fmt.Sprintf("%s%3.0f%%%s", green, percent, reset))

	if bar.total > 0 {
		stats.WriteString(fmt.Sprintf(" %s%s%s", cyan, humanBytes(bar.total), reset))
	}
	if speed != "" {
		stats.WriteString(fmt.Sprintf(" %s%s%s", yellow, speed, reset))
	}
	if eta != "" {
		stats.WriteString(fmt.Sprintf(" %s", eta))
	}

	return fmt.Sprintf("%s %s %s", bar.prefix, barStr, stats.String())
}

func formatTimeCompact(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	s := int(d.Seconds())
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	m := s / 60
	s = s % 60
	if m < 60 {
		return fmt.Sprintf("%d:%02d", m, s)
	}
	h := m / 60
	m = m % 60
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}

func (r *ProgressRenderer) clearBars() {
	if r.lastBars == 0 {
		return
	}
	fmt.Fprintf(r.out, "\x1b[%dA", r.lastBars)
	for i := 0; i < r.lastBars; i++ {
		fmt.Fprint(r.out, "\r\x1b[2K")
		if i < r.lastBars-1 {
			fmt.Fprint(r.out, "\n")
		}
	}
	fmt.Fprintf(r.out, "\x1b[%dA", r.lastBars)
}
