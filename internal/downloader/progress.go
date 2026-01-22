package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type progressWriter struct {
	size      int64
	total     int64
	start     time.Time
	lastPrint time.Time
	finished  bool
	lastLen   int
	prefix    string
	printer   *Printer
}

func newProgressWriter(size int64, printer *Printer, prefix string) *progressWriter {
	return &progressWriter{
		size:    size,
		start:   time.Now(),
		prefix:  prefix,
		printer: printer,
	}
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.total += int64(n)

	now := time.Now()
	shouldPrint := p.total == p.size
	if now.Sub(p.lastPrint) > 200*time.Millisecond {
		shouldPrint = true
	}

	if shouldPrint {
		p.print()
		p.lastPrint = now
	}
	return n, nil
}

func (p *progressWriter) print() {
	if p.finished {
		return
	}

	line := p.printer.progressLine(p.prefix, p.total, p.size, time.Since(p.start))
	if p.lastLen > len(line) {
		line += strings.Repeat(" ", p.lastLen-len(line))
	}
	p.lastLen = len(line)
	fmt.Fprintf(os.Stderr, "\r%s", line)
}

func (p *progressWriter) Finish() {
	if p.finished {
		return
	}
	p.finished = true
	p.print()
	// Move to next line after finishing
	fmt.Fprint(os.Stderr, "\n")
}

func (p *progressWriter) NewLine() {
	if p.finished {
		return
	}
	p.lastLen = 0
	fmt.Fprint(os.Stderr, "\n")
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.r.Read(p)
	}
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	reader := &contextReader{ctx: ctx, r: src}
	return io.Copy(dst, reader)
}
