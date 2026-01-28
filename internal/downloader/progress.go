package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

type progressWriter struct {
	size     int64
	total    int64
	start    time.Time
	finished bool
	prefix   string
	printer  *Printer
	taskID   string
}

func newProgressWriter(size int64, printer *Printer, prefix string) *progressWriter {
	taskID := ""
	// Note: printer.renderer is not currently implemented
	return &progressWriter{
		size:    size,
		start:   time.Now(),
		prefix:  prefix,
		printer: printer,
		taskID:  taskID,
	}
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.total += int64(n)

	p.print()
	return n, nil
}

func (p *progressWriter) print() {
	if p.finished {
		return
	}
	if !p.printer.progressEnabled {
		return
	}
	// Note: printer.renderer is not currently implemented
	line := p.printer.progressLine(p.prefix, p.total, p.size, time.Since(p.start))
	p.printer.writeProgressLine(line)
}

func (p *progressWriter) Finish() {
	if p.finished {
		return
	}
	p.finished = true
	if !p.printer.progressEnabled {
		line := p.printer.progressLine(p.prefix, p.total, p.size, time.Since(p.start))
		fmt.Fprintf(os.Stderr, "%s\n", line)
		return
	}
	p.print()
	// Note: printer.renderer is not currently implemented
	p.printer.writeProgressLine("\n")
}

func (p *progressWriter) NewLine() {
	if p.finished {
		return
	}
	// Note: printer.renderer is not currently implemented
	if !p.printer.progressEnabled {
		return
	}
	p.printer.writeProgressLine("\n")
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
