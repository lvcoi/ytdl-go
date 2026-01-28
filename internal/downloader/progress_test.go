package downloader

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = writer
	fn()
	time.Sleep(10 * time.Millisecond)
	_ = writer.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()
	return buf.String()
}

func TestProgressNonTTYUsesNewlines(t *testing.T) {
	output := captureStderr(t, func() {
		printer := newPrinter(Options{}, nil)
		progress := newProgressWriter(10, printer, "[1/1] demo")
		_, _ = progress.Write([]byte("12345"))
		progress.Finish()
	})

	if strings.Contains(output, "\r") {
		t.Fatalf("expected no carriage returns in non-TTY output, got %q", output)
	}
	if !strings.Contains(output, "[1/1] demo") {
		t.Fatalf("expected progress prefix in output, got %q", output)
	}
}

func TestProgressInterleavesLogsNonTTY(t *testing.T) {
	output := captureStderr(t, func() {
		printer := newPrinter(Options{}, nil)
		progress := newProgressWriter(10, printer, "[1/1] demo")
		_, _ = progress.Write([]byte("12345"))
		printer.Log(LogInfo, "log message")
		_, _ = progress.Write([]byte("12345"))
		progress.Finish()
	})

	// For non-TTY output, progress is only printed at the end
	logIdx := strings.Index(output, "log message")
	progressIdx := strings.Index(output, "[1/1] demo")
	
	if logIdx == -1 || progressIdx == -1 {
		t.Fatalf("expected progress + log output, got %q", output)
	}
	
	// Log should appear before the final progress line in non-TTY mode
	if logIdx >= progressIdx {
		t.Fatalf("expected log message before progress line in non-TTY mode, got %q", output)
	}
	
	if !strings.Contains(output, "[INFO] log message") {
		t.Fatalf("expected [INFO] log message in output, got %q", output)
	}
}
