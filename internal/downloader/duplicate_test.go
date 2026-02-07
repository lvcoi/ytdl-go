package downloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDuplicatePolicy(t *testing.T) {
	tests := []struct {
		input   string
		want    DuplicatePolicy
		wantErr bool
	}{
		{input: "", want: DuplicatePolicyPrompt},
		{input: "prompt", want: DuplicatePolicyPrompt},
		{input: "overwrite", want: DuplicatePolicyOverwrite},
		{input: "skip", want: DuplicatePolicySkip},
		{input: "rename", want: DuplicatePolicyRename},
		{input: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		got, err := ParseDuplicatePolicy(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("ParseDuplicatePolicy(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseDuplicatePolicy(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParseDuplicatePolicy(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDuplicateDecision(t *testing.T) {
	got, err := ParseDuplicateDecision("overwrite-all")
	if err != nil {
		t.Fatalf("ParseDuplicateDecision returned error: %v", err)
	}
	if got != DuplicateDecisionOverwriteAll {
		t.Fatalf("ParseDuplicateDecision returned %q, want %q", got, DuplicateDecisionOverwriteAll)
	}
}

func TestDuplicateSessionIsolation(t *testing.T) {
	sessionA := NewDuplicateSession()
	sessionB := NewDuplicateSession()
	if !sessionA.SetApplyAllDecision(DuplicateDecisionSkipAll) {
		t.Fatalf("failed to set session A apply-all decision")
	}
	if _, ok := sessionB.ApplyAllPolicy(); ok {
		t.Fatalf("session B unexpectedly inherited apply-all policy from session A")
	}
}

func TestHandleExistingPathPolicySkip(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "video.mp4")
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	out, skip, err := handleExistingPath(path, base, Options{
		OnDuplicate:      DuplicatePolicySkip,
		DuplicateSession: NewDuplicateSession(),
		Quiet:            true,
	}, nil)
	if err != nil {
		t.Fatalf("handleExistingPath returned error: %v", err)
	}
	if !skip {
		t.Fatalf("expected skip=true")
	}
	if out != path {
		t.Fatalf("expected original path, got %q", out)
	}
}

func TestHandleExistingPathPolicyRename(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "video.mp4")
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	out, skip, err := handleExistingPath(path, base, Options{
		OnDuplicate:      DuplicatePolicyRename,
		DuplicateSession: NewDuplicateSession(),
		Quiet:            true,
	}, nil)
	if err != nil {
		t.Fatalf("handleExistingPath returned error: %v", err)
	}
	if skip {
		t.Fatalf("expected skip=false")
	}
	if out == path {
		t.Fatalf("expected renamed output path")
	}
	if !strings.Contains(filepath.Base(out), " (1)") {
		t.Fatalf("expected renamed suffix in %q", out)
	}
}

func TestHandleExistingPathApplyAllSession(t *testing.T) {
	base := t.TempDir()
	first := filepath.Join(base, "video1.mp4")
	second := filepath.Join(base, "video2.mp4")
	if err := os.WriteFile(first, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile first failed: %v", err)
	}
	if err := os.WriteFile(second, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile second failed: %v", err)
	}

	session := NewDuplicateSession()
	if !session.SetApplyAllDecision(DuplicateDecisionRenameAll) {
		t.Fatalf("failed to set apply-all decision")
	}
	opts := Options{
		OnDuplicate:      DuplicatePolicyPrompt,
		DuplicateSession: session,
		Quiet:            true,
	}

	out1, skip1, err := handleExistingPath(first, base, opts, nil)
	if err != nil {
		t.Fatalf("handleExistingPath first returned error: %v", err)
	}
	if skip1 {
		t.Fatalf("expected first skip=false")
	}
	if out1 == first {
		t.Fatalf("expected first path to be renamed")
	}

	out2, skip2, err := handleExistingPath(second, base, opts, nil)
	if err != nil {
		t.Fatalf("handleExistingPath second returned error: %v", err)
	}
	if skip2 {
		t.Fatalf("expected second skip=false")
	}
	if out2 == second {
		t.Fatalf("expected second path to be renamed")
	}
}
