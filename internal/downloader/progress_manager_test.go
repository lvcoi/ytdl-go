package downloader

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProgressManagerMultipleBars(t *testing.T) {
	var buf bytes.Buffer
	width := 80
	manager := newProgressManagerForTest(&buf, func() int { return width }, "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.Start(ctx)

	taskA := manager.AddTask("a", 100)
	taskB := manager.AddTask("b", 200)
	manager.UpdateTask(taskA, 50, 100, 500*time.Millisecond)
	manager.UpdateTask(taskB, 10, 200, 500*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	manager.Stop()

	output := buf.String()
	alphaIdx := strings.LastIndex(output, "a")
	betaIdx := strings.LastIndex(output, "b")
	if alphaIdx == -1 || betaIdx == -1 {
		t.Fatalf("expected both progress labels in output, got: %q", output)
	}
	if alphaIdx > betaIdx {
		t.Fatalf("expected stable ordering (alpha before beta), got output: %q", output)
	}
}

func TestProgressManagerResizeEvent(t *testing.T) {
	var buf bytes.Buffer
	width := 80
	manager := newProgressManagerForTest(&buf, func() int { return width }, "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.Start(ctx)

	task := manager.AddTask("r", 100)
	manager.UpdateTask(task, 50, 100, 500*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	// Send resize event and wait for it to be processed
	width = 40
	manager.sendEvent(progressEvent{kind: eventResize, width: width})
	time.Sleep(50 * time.Millisecond)
	
	// Stop the manager before reading the buffer to avoid race condition
	manager.Stop()

	// Now safely read the output after the goroutine has stopped
	output := buf.String()
	if !strings.Contains(output, "r") {
		t.Fatalf("expected output to include label, got: %q", output)
	}
	if strings.Count(output, "[") == 0 {
		t.Fatalf("expected output to include progress bar, got: %q", output)
	}
}

func TestProgressManagerLogging(t *testing.T) {
	var buf bytes.Buffer
	manager := newProgressManagerForTest(&buf, func() int { return 80 }, "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager.Start(ctx)

	task := manager.AddTask("l", 100)
	manager.UpdateTask(task, 25, 100, 500*time.Millisecond)
	manager.Log(LogInfo, "hello")
	time.Sleep(50 * time.Millisecond)
	manager.Stop()

	output := buf.String()
	logIdx := strings.Index(output, "[INFO] hello")
	progressIdx := strings.LastIndex(output, "l")
	if logIdx == -1 {
		t.Fatalf("expected log output, got: %q", output)
	}
	if progressIdx == -1 || progressIdx < logIdx {
		t.Fatalf("expected progress to render after log output, got: %q", output)
	}
}

func TestProgressWriterNonTTYOutput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	defer reader.Close()
	defer writer.Close()

	originalStderr := os.Stderr
	os.Stderr = writer
	defer func() {
		os.Stderr = originalStderr
	}()

	printer := newPrinter(Options{}, nil)
	printer.progressEnabled = false
	progress := newProgressWriter(100, printer, "[1/1] test")
	_, _ = progress.Write(make([]byte, 100))
	progress.Finish()
	writer.Close()

	out, _ := io.ReadAll(reader)
	output := string(out)
	if !strings.Contains(output, "100.00%") {
		t.Fatalf("expected non-tty output to include progress summary, got: %q", output)
	}
}
