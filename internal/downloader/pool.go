package downloader

import (
	"context"
	"fmt"
	"sync"

	"github.com/lvcoi/ytdl-go/internal/ws"
)

// Task represents a download unit.
type Task struct {
	ID       string
	URLs     []string
	Options  Options
	Jobs     int
	Execute  func(ctx context.Context, urls []string, opts Options, jobs int) ([]any, int)
	OnFinish func(id string, err error)
}

// WSBroadcaster is an interface to decouple the pool from the WebSocket hub.
type WSBroadcaster interface {
	Broadcast(msg ws.WSMessage)
}

// Pool manages a fixed number of workers to process download tasks.
type Pool struct {
	TaskQueue   chan Task // Strict Unbuffered Channel
	WorkerQueue chan int  // Job distribution via worker IDs
	Workers     int
	Hub         WSBroadcaster
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewPool(workers int, hub WSBroadcaster) *Pool {
	return &Pool{
		TaskQueue:   make(chan Task), // Strict Unbuffered
		WorkerQueue: make(chan int, workers),
		Workers:     workers,
		Hub:         hub,
	}
}

func (p *Pool) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	for i := 0; i < p.Workers; i++ {
		// Initialize the worker queue with worker IDs
		p.WorkerQueue <- i
		go p.worker(i)
	}
}

func (p *Pool) AddTask(t Task) {
	go func() {
		select {
		case p.TaskQueue <- t:
		case <-p.ctx.Done():
		}
	}()
}

func (p *Pool) worker(id int) {
	for {
		select {
		case <-p.ctx.Done():
			return
		case workerID := <-p.WorkerQueue:
			select {
			case <-p.ctx.Done():
				return
			case task := <-p.TaskQueue:
				p.wg.Add(1)
				p.processTask(task)
				p.wg.Done()
			}
			// Return worker to queue
			select {
			case p.WorkerQueue <- workerID:
			case <-p.ctx.Done():
			}
		}
	}
}

func (p *Pool) processTask(t Task) {
	// Emit a "starting" message -> mapped to "progress" with status "starting"
	p.Hub.Broadcast(ws.WSMessage{
		Type: "progress",
		Payload: ws.ProgressPayload{
			ID:     t.ID,
			Status: "starting",
		},
	})

	// Inject our renderer that broadcasts via the Hub
	t.Options.Renderer = &poolRenderer{
		id:  t.ID,
		hub: p.Hub,
	}

	// Use provided context for cancellation
	_, exitCode := t.Execute(p.ctx, t.URLs, t.Options, t.Jobs)
	var err error
	if exitCode != 0 {
		err = fmt.Errorf("exit code %d", exitCode)
		p.Hub.Broadcast(ws.WSMessage{
			Type: "error",
			Payload: ws.ErrorPayload{
				ID:      t.ID,
				Message: err.Error(),
				Code:    exitCode,
			},
		})
	}

	if t.OnFinish != nil {
		t.OnFinish(t.ID, err)
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

type poolRenderer struct {
	id  string
	hub WSBroadcaster
}

func (r *poolRenderer) Register(prefix string, size int64) string {
	// Emit a progress update with the filename (prefix)
	r.hub.Broadcast(ws.WSMessage{
		Type: "progress",
		Payload: ws.ProgressPayload{
			ID:       r.id,
			Filename: prefix,
			Status:   "downloading",
			Percent:  0,
		},
	})
	return r.id
}

func (r *poolRenderer) Update(id string, current, total int64) {
	percent := 0.0
	if total > 0 {
		percent = float64(current) * 100 / float64(total)
	}
	r.hub.Broadcast(ws.WSMessage{
		Type: "progress",
		Payload: ws.ProgressPayload{
			ID:      r.id,
			Percent: percent,
			Status:  "downloading",
		},
	})
}

func (r *poolRenderer) Finish(id string) {
	r.hub.Broadcast(ws.WSMessage{
		Type: "progress",
		Payload: ws.ProgressPayload{
			ID:      r.id,
			Percent: 100,
			Status:  "complete",
		},
	})
}

func (r *poolRenderer) Log(level LogLevel, msg string) {}
