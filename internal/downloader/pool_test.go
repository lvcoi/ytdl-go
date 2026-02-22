package downloader

import (
	"context"
	"testing"
	"time"

	"github.com/lvcoi/ytdl-go/internal/ws"
)

// MockHub implements WSBroadcaster for testing.
type MockHub struct {
	messages []ws.WSMessage
}

func (m *MockHub) Broadcast(msg ws.WSMessage) {
	m.messages = append(m.messages, msg)
}

func TestPool_ProcessTask(t *testing.T) {
	mockHub := &MockHub{}
	// Create a pool with 1 worker
	pool := NewPool(1, mockHub)

	// Context for test
	pool.Start(context.Background())
	defer pool.Stop()

	// Define a task that succeeds
	taskID := "test_task_1"
	task := Task{
		ID:   taskID,
		URLs: []string{"http://example.com"},
		Execute: func(ctx context.Context, urls []string, opts Options, jobs int) ([]any, int) {
			// Simulate download progress
			if opts.Renderer != nil {
				opts.Renderer.Update(taskID, 50, 100)
				opts.Renderer.Finish(taskID)
			}
			return nil, 0 // Success
		},
	}

	// Add task to pool
	pool.AddTask(task)

	// Wait for processing (since it's async)
	// In a real test we might use a channel or waitgroup, but for this simple verification:
	time.Sleep(100 * time.Millisecond)

	// Assertions
	// Expected messages: "starting" -> "downloading" (50%) -> "complete" (100%)
	// Note: "complete" is sent by Finish(), which also sends "progress" with 100%.

	if len(mockHub.messages) == 0 {
		t.Fatal("expected messages, got none")
	}

	// Verify "starting"
	if mockHub.messages[0].Type != "progress" {
		t.Errorf("expected first message type 'progress', got %s", mockHub.messages[0].Type)
	}
	payload, ok := mockHub.messages[0].Payload.(ws.ProgressPayload)
	if !ok || payload.Status != "starting" {
		t.Errorf("expected first message status 'starting', got %v", payload)
	}

	// Verify at least one progress update
	foundProgress := false
	foundComplete := false
	for _, msg := range mockHub.messages {
		if p, ok := msg.Payload.(ws.ProgressPayload); ok {
			if p.Status == "downloading" && p.Percent == 50 {
				foundProgress = true
			}
			if p.Status == "complete" && p.Percent == 100 {
				foundComplete = true
			}
		}
	}

	if !foundProgress {
		t.Error("did not find expected progress update (50%)")
	}
	if !foundComplete {
		t.Error("did not find expected completion message")
	}
}
