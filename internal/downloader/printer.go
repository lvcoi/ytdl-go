package downloader

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

func parseLogLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "debug":
		return LogDebug
	case "warn", "warning":
		return LogWarn
	case "error":
		return LogError
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

type Printer struct {
	quiet           bool
	color           bool
	columns         int
	titleWidth      int
	logLevel        LogLevel
	progressEnabled bool
	interactive     bool
	layout          string
	renderer        *progressRenderer
	manager         *ProgressManager
	mu              sync.RWMutex
}

func newPrinter(opts Options, manager *ProgressManager) *Printer {
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

	var renderer *progressRenderer
	if manager != nil {
		renderer = &progressRenderer{manager: manager}
	}

	interactive := !opts.Quiet && manager != nil

	printer := &Printer{
		quiet:           opts.Quiet,
		color:           supportsColor(),
		columns:         columns,
		titleWidth:      titleWidth,
		logLevel:        parseLogLevel(opts.LogLevel),
		progressEnabled: interactive,
		interactive:     interactive,
		layout:          opts.ProgressLayout,
		renderer:        renderer,
		manager:         manager,
	}
	return printer
}

func (p *Printer) Prefix(index, total int, title string) string {
	if total <= 0 {
		total = 1
	}
	p.mu.RLock()
	titleWidth := p.titleWidth
	p.mu.RUnlock()
	width := len(strconv.Itoa(total))
	idx := fmt.Sprintf("%*d/%d", width, index, total)
	return fmt.Sprintf("[%s] %-*s", idx, titleWidth, truncateText(title, titleWidth))
}

func (p *Printer) progressLine(prefix string, current, total int64, elapsed time.Duration) string {
	if p.layout != "" {
		return formatProgressLayout(p.layout, prefix, current, total, elapsed)
	}
	speed := ""
	if elapsed > 0 {
		speed = humanBytes(int64(float64(current)/elapsed.Seconds())) + "/s"
	}

	if total > 0 {
		percent := float64(current) * 100 / float64(total)
		return fmt.Sprintf("%s %6.2f%% %s / %s %s",
			prefix,
			percent,
			padLeft(humanBytes(current), 9),
			padLeft(humanBytes(total), 9),
			padLeft(speed, 10),
		)
	}

	return fmt.Sprintf("%s %s %s",
		prefix,
		padLeft(humanBytes(current), 9),
		padLeft(speed, 10),
	)
}

func (p *Printer) ItemResult(prefix string, result downloadResult, err error) {
	if err == nil && p.quiet {
		return
	}

	statusText := "OK"
	statusColor := colorGreen
	detail := fmt.Sprintf("%s %s", padLeft(humanBytes(result.bytes), 9), result.outputPath)
	if result.retried {
		detail = fmt.Sprintf("%s (retry)", detail)
		statusColor = colorYellow
	}

	if err != nil {
		statusText = "FAIL"
		statusColor = colorRed
		detail = err.Error()
	}

	plainStatus := statusText
	if result.retried && err == nil {
		plainStatus = "OK"
	}

	p.mu.RLock()
	columns := p.columns
	p.mu.RUnlock()
	maxDetail := columns - len(prefix) - len(plainStatus) - 3
	if maxDetail < 0 {
		maxDetail = 0
	}
	detail = truncateText(detail, maxDetail)

	if p.renderer != nil && p.progressEnabled {
		level := LogInfo
		if err != nil {
			level = LogError
		} else if result.retried {
			level = LogWarn
		}
		message := fmt.Sprintf("%s %s %s", prefix, statusText, detail)
		p.renderer.Log(level, message)
		return
	}

	status := p.colorize(statusText, statusColor)
	message := fmt.Sprintf("%s %s %s", prefix, status, detail)
	if result.hadProgress {
		p.clearLine()
	}
	fmt.Fprintln(os.Stderr, message)
}

func (p *Printer) ItemSkipped(prefix, reason string) {
	if p.quiet {
		return
	}
	p.mu.RLock()
	columns := p.columns
	p.mu.RUnlock()
	maxDetail := columns - len(prefix) - len("SKIP") - 3
	if maxDetail < 0 {
		maxDetail = 0
	}
	reason = truncateText(reason, maxDetail)
	if p.renderer != nil && p.progressEnabled {
		message := fmt.Sprintf("%s SKIP %s", prefix, reason)
		p.renderer.Log(LogWarn, message)
		return
	}

	status := p.colorize("SKIP", colorYellow)
	message := fmt.Sprintf("%s %s %s", prefix, status, reason)
	fmt.Fprintln(os.Stderr, message)
}

func (p *Printer) Summary(total, ok, failed, skipped int, bytes int64) {
	if p.quiet {
		return
	}
	if p.renderer != nil && p.progressEnabled {
		line := fmt.Sprintf("Summary: OK %d | FAIL %d | SKIP %d | TOTAL %d | SIZE %s",
			ok, failed, skipped, total, humanBytes(bytes))
		level := LogInfo
		if failed > 0 {
			level = LogError
		} else if skipped > 0 {
			level = LogWarn
		}
		p.renderer.Log(level, line)
		return
	}

	okLabel := p.colorize("OK", colorGreen)
	failLabel := p.colorize("FAIL", colorRed)
	skipLabel := p.colorize("SKIP", colorYellow)
	line := fmt.Sprintf("Summary: %s %d | %s %d | %s %d | TOTAL %d | SIZE %s",
		okLabel, ok, failLabel, failed, skipLabel, skipped, total, humanBytes(bytes))
	fmt.Fprintln(os.Stderr, line)
}

func (p *Printer) Log(level LogLevel, message string) {
	if p.quiet {
		return
	}
	if level < p.logLevel {
		return
	}
	if p.renderer != nil && p.progressEnabled {
		p.renderer.Log(level, message)
		return
	}

	label := levelLabel(level)
	fmt.Fprintf(os.Stderr, "%s %s\n", label, message)
}

func (p *Printer) colorize(text, color string) string {
	if !p.color || color == "" {
		return text
	}
	return color + text + colorReset
}

func (p *Printer) clearLine() {
	if !p.interactive {
		return
	}
	width := p.columns
	if width <= 0 {
		width = 100
	}
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", width))
}

func (p *Printer) writeProgressLine(line string) {
	if line == "\n" {
		fmt.Fprint(os.Stderr, "\n")
		return
	}
	fmt.Fprintf(os.Stderr, "\r%s", line)
}

func padLeft(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return strings.Repeat(" ", width-len(value)) + value
}

func truncateText(text string, max int) string {
	if max <= 0 || len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}

func formatProgressLayout(layout, prefix string, current, total int64, elapsed time.Duration) string {
	percent := ""
	eta := ""
	rate := ""
	if elapsed > 0 {
		rate = humanBytes(int64(float64(current)/elapsed.Seconds())) + "/s"
	}
	if total > 0 {
		percent = fmt.Sprintf("%.2f%%", float64(current)*100/float64(total))
		if current > 0 {
			remaining := time.Duration(float64(elapsed) * (float64(total-current) / float64(current)))
			eta = formatETACompact(remaining)
		}
	}

	line := layout
	line = strings.ReplaceAll(line, "{label}", prefix)
	line = strings.ReplaceAll(line, "{prefix}", prefix)
	line = strings.ReplaceAll(line, "{percent}", percent)
	line = strings.ReplaceAll(line, "{current}", humanBytes(current))
	line = strings.ReplaceAll(line, "{total}", humanBytes(total))
	line = strings.ReplaceAll(line, "{rate}", rate)
	line = strings.ReplaceAll(line, "{eta}", eta)
	line = strings.ReplaceAll(line, "{bytes}", humanBytes(current))
	return strings.TrimSpace(line)
}

func formatETACompact(duration time.Duration) string {
	if duration <= 0 {
		return ""
	}
	seconds := int(duration.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, seconds%60)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

func terminalColumns() int {
	if columns := os.Getenv("COLUMNS"); columns != "" {
		if val, err := strconv.Atoi(columns); err == nil && val > 0 {
			return val
		}
	}
	return 0
}

func supportsColor() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
		return true
	}
	if os.Getenv("CLICOLOR") == "0" {
		return false
	}
	info, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

const (
	colorReset  = "\x1b[0m"
	colorGreen  = "\x1b[32m"
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
)
