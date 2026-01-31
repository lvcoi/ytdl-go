package downloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type progressWriter struct {
	size       int64
	total      atomic.Int64
	start      time.Time
	lastUpdate atomic.Int64 // Unix nanoseconds
	finished   atomic.Bool
	prefix     string
	printer    *Printer
	taskID     string
	renderer   *progressRenderer
	mu         sync.Mutex // Protects fields not using atomics
}

func newProgressWriter(size int64, printer *Printer, prefix string) *progressWriter {
	taskID := ""
	var renderer *progressRenderer
	if printer != nil && printer.renderer != nil {
		renderer = printer.renderer
		taskID = renderer.Register(prefix, size)
	}
	now := time.Now()
	pw := &progressWriter{
		size:     size,
		start:    now,
		prefix:   prefix,
		printer:  printer,
		taskID:   taskID,
		renderer: renderer,
	}
	pw.lastUpdate.Store(now.UnixNano())
	return pw
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.total.Add(int64(n))

	// Throttle progress updates to avoid performance overhead
	// Update at most once per 100ms (10 times per second)
	now := time.Now()
	lastUpdateNano := p.lastUpdate.Load()
	if now.UnixNano()-lastUpdateNano >= 100*time.Millisecond.Nanoseconds() {
		if p.lastUpdate.CompareAndSwap(lastUpdateNano, now.UnixNano()) {
			p.print()
		}
	}
	return n, nil
}

func (p *progressWriter) print() {
	if p.finished.Load() {
		return
	}
	if !p.printer.progressEnabled {
		return
	}
	total := p.total.Load()
	if p.renderer != nil && p.taskID != "" {
		p.renderer.Update(p.taskID, total, p.size)
		return
	}
	line := p.printer.progressLine(p.prefix, total, p.size, time.Since(p.start))
	p.printer.writeProgressLine(line)
}

func (p *progressWriter) Finish() {
	if p.finished.Swap(true) {
		return
	}
	total := p.total.Load()
	if !p.printer.progressEnabled {
		line := p.printer.progressLine(p.prefix, total, p.size, time.Since(p.start))
		fmt.Fprintf(os.Stderr, "%s\n", line)
		return
	}
	if p.renderer != nil && p.taskID != "" {
		// Force a final update before finishing
		p.renderer.Update(p.taskID, total, p.size)
		p.renderer.Finish(p.taskID)
		return
	}
	// Force a final update before finishing
	p.print()
	p.printer.writeProgressLine("\n")
}

func (p *progressWriter) NewLine() {
	if p.finished.Load() {
		return
	}
	if !p.printer.progressEnabled {
		return
	}
	if p.renderer != nil && p.taskID != "" {
		return
	}
	p.printer.writeProgressLine("\n")
}

func (p *progressWriter) Reset(size int64) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	p.size = size
	p.total.Store(0)
	p.start = now
	p.lastUpdate.Store(now.UnixNano())
	p.finished.Store(false)
	if p.renderer != nil && p.taskID != "" {
		p.renderer.Update(p.taskID, 0, p.size)
	}
}

func (p *progressWriter) SetCurrent(current int64) {
	if p == nil {
		return
	}
	p.total.Store(current)
	if p.renderer != nil && p.taskID != "" {
		p.renderer.Update(p.taskID, current, p.size)
	}
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
