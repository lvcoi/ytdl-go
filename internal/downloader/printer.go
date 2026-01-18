package downloader

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Printer struct {
	quiet      bool
	color      bool
	columns    int
	titleWidth int
}

func newPrinter(opts Options) *Printer {
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

	return &Printer{
		quiet:      opts.Quiet,
		color:      supportsColor(),
		columns:    columns,
		titleWidth: titleWidth,
	}
}

func (p *Printer) Prefix(index, total int, title string) string {
	if total <= 0 {
		total = 1
	}
	width := len(strconv.Itoa(total))
	idx := fmt.Sprintf("%*d/%d", width, index, total)
	return fmt.Sprintf("[%s] %-*s", idx, p.titleWidth, truncateText(title, p.titleWidth))
}

func (p *Printer) progressLine(prefix string, current, total int64, elapsed time.Duration) string {
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

	if result.hadProgress {
		p.clearLine()
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

	status := p.colorize(statusText, statusColor)
	plainStatus := statusText
	if result.retried && err == nil {
		plainStatus = "OK"
	}

	maxDetail := p.columns - len(prefix) - len(plainStatus) - 3
	if maxDetail < 0 {
		maxDetail = 0
	}
	detail = truncateText(detail, maxDetail)

	fmt.Fprintf(os.Stderr, "%s %s %s\n", prefix, status, detail)
}

func (p *Printer) ItemSkipped(prefix, reason string) {
	if p.quiet {
		return
	}
	status := p.colorize("SKIP", colorYellow)
	maxDetail := p.columns - len(prefix) - len("SKIP") - 3
	if maxDetail < 0 {
		maxDetail = 0
	}
	fmt.Fprintf(os.Stderr, "%s %s %s\n", prefix, status, truncateText(reason, maxDetail))
}

func (p *Printer) Summary(total, ok, failed, skipped int, bytes int64) {
	if p.quiet {
		return
	}
	okLabel := p.colorize("OK", colorGreen)
	failLabel := p.colorize("FAIL", colorRed)
	skipLabel := p.colorize("SKIP", colorYellow)
	fmt.Fprintf(os.Stderr, "Summary: %s %d | %s %d | %s %d | TOTAL %d | SIZE %s\n",
		okLabel, ok, failLabel, failed, skipLabel, skipped, total, humanBytes(bytes))
}

func (p *Printer) colorize(text, color string) string {
	if !p.color || color == "" {
		return text
	}
	return color + text + colorReset
}

func (p *Printer) clearLine() {
	width := p.columns
	if width <= 0 {
		width = 100
	}
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", width))
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
