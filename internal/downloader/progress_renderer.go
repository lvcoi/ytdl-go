package downloader

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"fortio.org/progressbar"
)

type barState struct {
	bar     *progressbar.Bar
	prefix  string
	total   int64
	current int64
}

type ProgressRenderer struct {
	out         io.Writer
	printer     *Printer
	interactive bool
	bars        map[string]*barState
	seq         uint64
	mu          sync.Mutex
}

func newProgressRenderer(out io.Writer, printer *Printer) *ProgressRenderer {
	interactive := printer != nil && printer.interactive

	return &ProgressRenderer{
		out:         out,
		printer:     printer,
		interactive: interactive,
		bars:        make(map[string]*barState),
	}
}

func (r *ProgressRenderer) Register(prefix string, total int64) string {
	id := fmt.Sprintf("bar-%d", atomic.AddUint64(&r.seq, 1))

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.interactive {
		r.bars[id] = &barState{prefix: prefix, total: total}
		return id
	}

	cfg := progressbar.Config{
		Width:        30,
		UseColors:    true,
		Color:        progressbar.RedBar, // Pink/magenta-ish
		Prefix:       prefix + " ",
		Suffix:       fmt.Sprintf(" %s", humanBytes(total)),
		ScreenWriter: os.Stderr,
	}

	bar := cfg.NewBar()
	r.bars[id] = &barState{bar: bar, prefix: prefix, total: total}

	return id
}

func (r *ProgressRenderer) Update(id string, delta, value, total int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.bars[id]
	if b == nil {
		return
	}

	if total > 0 {
		b.total = total
	}

	if b.bar == nil {
		return
	}

	// Track current value
	if value > 0 {
		b.current = value
	} else if delta > 0 {
		b.current += delta
	}

	var pct float64
	if b.total > 0 {
		pct = float64(b.current) / float64(b.total) * 100
	}

	// Update suffix with current/total
	b.bar.UpdateSuffix(fmt.Sprintf(" %s/%s", humanBytes(b.current), humanBytes(b.total)))
	b.bar.Progress(pct)
}

func (r *ProgressRenderer) Finish(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.bars[id]
	if b == nil {
		return
	}

	if b.bar != nil {
		b.bar.UpdateSuffix(fmt.Sprintf(" %s done", humanBytes(b.total)))
		b.bar.Progress(100)
		b.bar.End()
	}

	delete(r.bars, id)
}

func (r *ProgressRenderer) Log(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Fprintln(r.out, msg)
}

func (r *ProgressRenderer) Flush() {
	// Nothing to do
}

func (r *ProgressRenderer) Wait() {
	// Nothing to do
}
