package web

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lvcoi/ytdl-go/internal/downloader"
)

func setCompletedAtForTest(job *Job, completedAt time.Time) {
	job.mu.Lock()
	job.CompletedAt = completedAt
	job.mu.Unlock()
}

func TestJobTrackerRemoveExpired(t *testing.T) {
	jt := &jobTracker{}

	completeJob := jt.Create([]string{"https://example.com/1"})
	completeJob.SetOutcome(nil, 0)
	setCompletedAtForTest(completeJob, time.Now().Add(-16*time.Minute))

	errorJob := jt.Create([]string{"https://example.com/2"})
	errorJob.SetOutcome(nil, 1)
	setCompletedAtForTest(errorJob, time.Now().Add(-31*time.Minute))

	activeJob := jt.Create([]string{"https://example.com/3"})
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
	job := jt.Create([]string{"https://example.com"})

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
	job := jt.Create([]string{"https://example.com"})

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
