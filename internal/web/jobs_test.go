package web

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lvcoi/ytdl-go/internal/app"
	"github.com/lvcoi/ytdl-go/internal/downloader"
)

func createTestJob(t *testing.T, jt *jobTracker, urls []string) *Job {
	t.Helper()
	job := jt.Create(urls)
	t.Cleanup(func() {
		job.CloseEvents()
	})
	return job
}

func waitForEventSeq(t *testing.T, job *Job, minSeq int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if job.eventSeq.Load() >= minSeq {
			job.eventMu.Lock()
			lastSeq := int64(0)
			if n := len(job.eventHistory); n > 0 {
				lastSeq = job.eventHistory[n-1].Seq
			}
			job.eventMu.Unlock()
			if lastSeq >= minSeq {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	job.eventMu.Lock()
	lastSeq := int64(0)
	if n := len(job.eventHistory); n > 0 {
		lastSeq = job.eventHistory[n-1].Seq
	}
	job.eventMu.Unlock()
	t.Fatalf(
		"timed out waiting for event seq >= %d (eventSeq=%d, lastHistorySeq=%d)",
		minSeq,
		job.eventSeq.Load(),
		lastSeq,
	)
}

func waitForCondition(t *testing.T, timeout time.Duration, check func() bool, onTimeout func() string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s", onTimeout())
}

func readEventWithTimeout(t *testing.T, ch <-chan ProgressEvent, timeout time.Duration) ProgressEvent {
	t.Helper()
	select {
	case evt, ok := <-ch:
		if !ok {
			t.Fatalf("expected event, stream closed")
		}
		return evt
	case <-time.After(timeout):
		t.Fatalf("timed out waiting for event")
		return ProgressEvent{}
	}
}

func setCompletedAtForTest(job *Job, completedAt time.Time) {
	job.mu.Lock()
	job.CompletedAt = completedAt
	job.mu.Unlock()
}

func TestJobTrackerRemoveExpired(t *testing.T) {
	jt := &jobTracker{}

	completeJob := createTestJob(t, jt, []string{"https://example.com/1"})
	completeJob.SetOutcome(nil, 0)
	setCompletedAtForTest(completeJob, time.Now().Add(-16*time.Minute))

	errorJob := createTestJob(t, jt, []string{"https://example.com/2"})
	errorJob.SetOutcome(nil, 1)
	setCompletedAtForTest(errorJob, time.Now().Add(-31*time.Minute))

	activeJob := createTestJob(t, jt, []string{"https://example.com/3"})
	activeJob.SetStatus("running")

	removed := jt.RemoveExpired(time.Now(), 15*time.Minute, 30*time.Minute)
	if removed != 2 {
		t.Fatalf("expected 2 jobs removed, got %d", removed)
	}

	if _, ok := jt.Get(activeJob.ID); !ok {
		t.Fatalf("expected active job to remain")
	}
	if _, ok := jt.Get(completeJob.ID); ok {
		t.Fatalf("expected completed job to be removed")
	}
	if _, ok := jt.Get(errorJob.ID); ok {
		t.Fatalf("expected errored job to be removed")
	}
}

func TestJobConcurrentStateAccess(t *testing.T) {
	jt := &jobTracker{}
	job := createTestJob(t, jt, []string{"https://example.com"})

	const loops = 500
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < loops; j++ {
				job.SetStatus("running")
				job.StatusValue()
				job.isActive()
			}
		}()
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < loops; j++ {
				job.SetOutcome(nil, 0)
				job.SetOutcome(nil, 1)
			}
		}()
	}

	wg.Wait()
	status := job.StatusValue()
	if status == "" {
		t.Fatalf("expected non-empty status")
	}
}

func TestWebDuplicatePrompterReturnsOnContextCancel(t *testing.T) {
	jt := &jobTracker{}
	job := createTestJob(t, jt, []string{"https://example.com"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	prompter := newWebDuplicatePrompter(ctx, job)
	prompter.timeout = time.Minute

	start := time.Now()
	decision, err := prompter.PromptDuplicate("/tmp/example.mp4")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("PromptDuplicate returned unexpected error: %v", err)
	}
	if decision != downloader.DuplicateDecisionSkip {
		t.Fatalf("expected skip decision on canceled context, got %q", decision)
	}
	if elapsed > time.Second {
		t.Fatalf("PromptDuplicate should return quickly on canceled context; took %v", elapsed)
	}

	job.promptMu.Lock()
	defer job.promptMu.Unlock()
	if len(job.pendingPrompt) != 0 {
		t.Fatalf("expected no pending prompts after canceled context, got %d", len(job.pendingPrompt))
	}
}

func TestJobSubscribeSnapshotReplayAndLiveEvent(t *testing.T) {
	jt := &jobTracker{}
	job := createTestJob(t, jt, []string{"https://example.com"})

	startSeq := job.eventSeq.Load()
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:  "register",
		ID:    "task_1",
		Label: "Track 1",
		Total: 100,
	}, time.Second) {
		t.Fatalf("failed to enqueue register event")
	}
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:    "progress",
		ID:      "task_1",
		Current: 55,
		Total:   100,
	}, time.Second) {
		t.Fatalf("failed to enqueue progress event")
	}
	waitForEventSeq(t, job, startSeq+2)

	stream, cancel := job.Subscribe(0)
	defer cancel()

	first := readEventWithTimeout(t, stream, time.Second)
	if first.Type != "snapshot" {
		t.Fatalf("expected first event type snapshot, got %q", first.Type)
	}
	if first.Snapshot == nil {
		t.Fatalf("expected snapshot payload in snapshot event")
	}
	if first.Snapshot.JobID != job.ID {
		t.Fatalf("expected snapshot job id %q, got %q", job.ID, first.Snapshot.JobID)
	}

	seenRegister := false
	seenProgress := false
	for i := 0; i < 8; i++ {
		evt := readEventWithTimeout(t, stream, time.Second)
		if evt.Type == "register" {
			seenRegister = true
		}
		if evt.Type == "progress" {
			seenProgress = true
		}
		if evt.Type != "snapshot" {
			if evt.Seq <= 0 {
				t.Fatalf("expected replay/live event to include positive seq, got %d for type %q", evt.Seq, evt.Type)
			}
			if evt.JobID != job.ID {
				t.Fatalf("expected event job id %q, got %q", job.ID, evt.JobID)
			}
		}
		if seenRegister && seenProgress {
			break
		}
	}
	if !seenRegister || !seenProgress {
		t.Fatalf("expected replay stream to include register and progress events")
	}

	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:    "log",
		Level:   "info",
		Message: "live-message",
	}, time.Second) {
		t.Fatalf("failed to enqueue live log event")
	}

	foundLiveLog := false
	for i := 0; i < 8; i++ {
		evt := readEventWithTimeout(t, stream, time.Second)
		if evt.Type == "log" && evt.Message == "live-message" {
			foundLiveLog = true
			break
		}
	}
	if !foundLiveLog {
		t.Fatalf("expected to receive live log event")
	}
}

func TestJobProgressSnapshotIncludesTaskDuplicateAndStats(t *testing.T) {
	jt := &jobTracker{}
	job := createTestJob(t, jt, []string{"https://example.com"})

	startSeq := job.eventSeq.Load()
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:     "duplicate",
		PromptID: "dup_1",
		Path:     "/tmp/song.mp3",
		Filename: "song.mp3",
	}, time.Second) {
		t.Fatalf("failed to enqueue duplicate event")
	}
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:  "register",
		ID:    "task_1",
		Label: "Track 1",
		Total: 100,
	}, time.Second) {
		t.Fatalf("failed to enqueue register event")
	}
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:    "progress",
		ID:      "task_1",
		Current: 100,
		Total:   100,
		Percent: 100,
	}, time.Second) {
		t.Fatalf("failed to enqueue progress event")
	}
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type: "finish",
		ID:   "task_1",
	}, time.Second) {
		t.Fatalf("failed to enqueue finish event")
	}
	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:    "log",
		Level:   "info",
		Message: "hello",
	}, time.Second) {
		t.Fatalf("failed to enqueue log event")
	}

	waitForEventSeq(t, job, startSeq+5)
	snapshot := job.progressSnapshot()
	if len(snapshot.Duplicates) != 1 {
		t.Fatalf("expected 1 duplicate prompt, got %d", len(snapshot.Duplicates))
	}
	if len(snapshot.Tasks) != 1 {
		t.Fatalf("expected 1 task in snapshot, got %d", len(snapshot.Tasks))
	}
	if !snapshot.Tasks[0].Done {
		t.Fatalf("expected task to be marked done in snapshot")
	}
	if len(snapshot.Logs) == 0 || snapshot.Logs[len(snapshot.Logs)-1].Message != "hello" {
		t.Fatalf("expected snapshot logs to include emitted message")
	}

	if !job.enqueueCriticalEvent(ProgressEvent{
		Type:     "duplicate-resolved",
		PromptID: "dup_1",
		Message:  string(downloader.DuplicateDecisionSkip),
	}, time.Second) {
		t.Fatalf("failed to enqueue duplicate-resolved event")
	}
	waitForEventSeq(t, job, startSeq+6)
	waitForCondition(t, time.Second, func() bool {
		return len(job.progressSnapshot().Duplicates) == 0
	}, func() string {
		return "expected duplicate prompt list to be empty after resolve"
	})
	snapshot = job.progressSnapshot()

	status := job.SetOutcome([]app.Result{
		{URL: "u1"},
		{URL: "u2", Error: "boom"},
	}, 1)
	if status != "error" {
		t.Fatalf("expected terminal status error, got %q", status)
	}
	waitForEventSeq(t, job, startSeq+7)
	snapshot = job.progressSnapshot()
	if snapshot.Status != "error" {
		t.Fatalf("expected snapshot status error, got %q", snapshot.Status)
	}
	if snapshot.Stats.Total != 2 || snapshot.Stats.Succeeded != 1 || snapshot.Stats.Failed != 1 {
		t.Fatalf("unexpected stats: %+v", snapshot.Stats)
	}
}
