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
		printer := newPrinter(Options{})
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
		printer := newPrinter(Options{})
		progress := newProgressWriter(10, printer, "[1/1] demo")
		_, _ = progress.Write([]byte("12345"))
		printer.Log("log message")
		_, _ = progress.Write([]byte("12345"))
		progress.Finish()
	})

	firstIdx := strings.Index(output, "[1/1] demo")
	logIdx := strings.Index(output, "log message")
	lastIdx := strings.LastIndex(output, "[1/1] demo")

	if firstIdx == -1 || logIdx == -1 || lastIdx == -1 {
		t.Fatalf("expected progress + log output, got %q", output)
	}
	if !(firstIdx < logIdx && logIdx < lastIdx) {
		t.Fatalf("expected log line between progress lines, got %q", output)
	}
	if !strings.Contains(output, "\nlog message\n") {
		t.Fatalf("expected log message on its own line, got %q", output)
	}
}
