package downloader

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

type progressEvent struct {
	kind   string
	id     string
	prefix string
	total  int64
	delta  int64
	value  int64
	msg    string
	ack    chan struct{}
}

type ProgressRenderer struct {
	out         io.Writer
	printer     *Printer
	interactive bool
	events      chan progressEvent
	bars        map[string]*progress.Tracker
	renderer    progress.Writer
	renderOnce  sync.Once
	seq         uint64
	orderSeq    uint64
	mu          sync.Mutex
}

func newProgressRenderer(out io.Writer, printer *Printer) *ProgressRenderer {
	progressWriter := progress.NewWriter()
	progressWriter.SetOutputWriter(out)
	progressWriter.SetAutoStop(false)
	progressWriter.SetSortBy(progress.SortByIndex)
	progressWriter.SetUpdateFrequency(150 * time.Millisecond)
	if printer != nil {
		progressWriter.SetTerminalWidth(printer.columns)
		progressWriter.SetMessageLength(printer.titleWidth + 10)
		if printer.columns > 100 {
			progressWriter.SetTrackerLength(40)
		} else {
			progressWriter.SetTrackerLength(30)
		}
	}

	renderer := &ProgressRenderer{
		out:         out,
		printer:     printer,
		interactive: printer != nil && printer.interactive,
		events:      make(chan progressEvent, 256),
		bars:        map[string]*progress.Tracker{},
		renderer:    progressWriter,
	}
	go renderer.loop()
	return renderer
}

func (r *ProgressRenderer) Register(prefix string, total int64) string {
	id := fmt.Sprintf("bar-%d", atomic.AddUint64(&r.seq, 1))
	r.events <- progressEvent{kind: "register", id: id, prefix: prefix, total: total}
	return id
}

func (r *ProgressRenderer) Update(id string, delta, value, total int64) {
	select {
	case r.events <- progressEvent{kind: "update", id: id, delta: delta, value: value, total: total}:
	default:
	}
}

func (r *ProgressRenderer) Finish(id string) {
	r.events <- progressEvent{kind: "finish", id: id}
}

func (r *ProgressRenderer) Log(msg string) {
	r.events <- progressEvent{kind: "log", msg: msg}
}

func (r *ProgressRenderer) Flush() {
	ack := make(chan struct{})
	r.events <- progressEvent{kind: "flush", ack: ack}
	<-ack
}

func (r *ProgressRenderer) loop() {
	for {
		select {
		case event := <-r.events:
			switch event.kind {
			case "register":
				r.handleRegister(event)
			case "update":
				r.handleUpdate(event)
			case "finish":
				r.handleFinish(event)
			case "log":
				r.handleLog(event)
			case "flush":
				r.handleFlush(event)
				if event.ack != nil {
					close(event.ack)
				}
			}
		}
	}
}

func (r *ProgressRenderer) handleRegister(event progressEvent) {
	if _, exists := r.bars[event.id]; exists {
		return
	}
	tracker := &progress.Tracker{
		Message:            event.prefix,
		Total:              event.total,
		Units:              progress.UnitsBytes,
		RemoveOnCompletion: true,
		Index:              atomic.AddUint64(&r.orderSeq, 1),
	}
	r.bars[event.id] = tracker
	r.startRender()
	r.renderer.AppendTracker(tracker)
}

func (r *ProgressRenderer) handleUpdate(event progressEvent) {
	bar := r.bars[event.id]
	if bar == nil {
		return
	}
	if event.total > 0 {
		bar.UpdateTotal(event.total)
	}
	if event.value > 0 {
		bar.SetValue(event.value)
		return
	}
	if event.delta != 0 {
		bar.Increment(event.delta)
	}
}

func (r *ProgressRenderer) handleFinish(event progressEvent) {
	bar := r.bars[event.id]
	if bar == nil {
		return
	}
	bar.MarkAsDone()
}

func (r *ProgressRenderer) handleLog(event progressEvent) {
	r.startRender()
	r.renderer.Log(event.msg)
}

func (r *ProgressRenderer) handleFlush(event progressEvent) {
	r.startRender()
	if !r.renderer.IsRenderInProgress() {
		go r.renderer.Render()
	}
}

func (r *ProgressRenderer) startRender() {
	r.renderOnce.Do(func() {
		go r.renderer.Render()
	})
}
