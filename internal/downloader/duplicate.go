package downloader

import (
	"fmt"
	"strings"
	"sync"
)

type DuplicatePolicy string

const (
	DuplicatePolicyPrompt    DuplicatePolicy = "prompt"
	DuplicatePolicyOverwrite DuplicatePolicy = "overwrite"
	DuplicatePolicySkip      DuplicatePolicy = "skip"
	DuplicatePolicyRename    DuplicatePolicy = "rename"
)

type DuplicateDecision string

const (
	DuplicateDecisionOverwrite    DuplicateDecision = "overwrite"
	DuplicateDecisionOverwriteAll DuplicateDecision = "overwrite_all"
	DuplicateDecisionSkip         DuplicateDecision = "skip"
	DuplicateDecisionSkipAll      DuplicateDecision = "skip_all"
	DuplicateDecisionRename       DuplicateDecision = "rename"
	DuplicateDecisionRenameAll    DuplicateDecision = "rename_all"
	DuplicateDecisionCancel       DuplicateDecision = "cancel"
)

// DuplicatePrompter resolves duplicate file decisions for non-TTY flows.
type DuplicatePrompter interface {
	PromptDuplicate(path string) (DuplicateDecision, error)
}

// DuplicateSession stores per-run duplicate state like "apply to all" choices.
type DuplicateSession struct {
	mu       sync.RWMutex
	applyAll DuplicatePolicy
}

func NewDuplicateSession() *DuplicateSession {
	return &DuplicateSession{}
}

func normalizeDuplicateToken(v string) string {
	s := strings.TrimSpace(strings.ToLower(v))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func ParseDuplicatePolicy(raw string) (DuplicatePolicy, error) {
	switch normalizeDuplicateToken(raw) {
	case "", string(DuplicatePolicyPrompt):
		return DuplicatePolicyPrompt, nil
	case string(DuplicatePolicyOverwrite):
		return DuplicatePolicyOverwrite, nil
	case string(DuplicatePolicySkip):
		return DuplicatePolicySkip, nil
	case string(DuplicatePolicyRename):
		return DuplicatePolicyRename, nil
	default:
		return "", fmt.Errorf("invalid on-duplicate policy: %q", raw)
	}
}

func (p DuplicatePolicy) OrDefault() DuplicatePolicy {
	normalized, err := ParseDuplicatePolicy(string(p))
	if err != nil {
		return DuplicatePolicyPrompt
	}
	return normalized
}

func ParseDuplicateDecision(raw string) (DuplicateDecision, error) {
	switch normalizeDuplicateToken(raw) {
	case string(DuplicateDecisionOverwrite):
		return DuplicateDecisionOverwrite, nil
	case string(DuplicateDecisionOverwriteAll):
		return DuplicateDecisionOverwriteAll, nil
	case string(DuplicateDecisionSkip):
		return DuplicateDecisionSkip, nil
	case string(DuplicateDecisionSkipAll):
		return DuplicateDecisionSkipAll, nil
	case string(DuplicateDecisionRename):
		return DuplicateDecisionRename, nil
	case string(DuplicateDecisionRenameAll):
		return DuplicateDecisionRenameAll, nil
	case string(DuplicateDecisionCancel):
		return DuplicateDecisionCancel, nil
	default:
		return "", fmt.Errorf("invalid duplicate choice: %q", raw)
	}
}

func (d DuplicateDecision) IsApplyAll() bool {
	return d == DuplicateDecisionOverwriteAll || d == DuplicateDecisionSkipAll || d == DuplicateDecisionRenameAll
}

func (s *DuplicateSession) ApplyAllPolicy() (DuplicatePolicy, bool) {
	if s == nil {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.applyAll == "" {
		return "", false
	}
	return s.applyAll, true
}

func (s *DuplicateSession) SetApplyAllPolicy(policy DuplicatePolicy) bool {
	if s == nil {
		return false
	}
	normalized, err := ParseDuplicatePolicy(string(policy))
	if err != nil || normalized == DuplicatePolicyPrompt {
		return false
	}
	s.mu.Lock()
	s.applyAll = normalized
	s.mu.Unlock()
	return true
}

func (s *DuplicateSession) SetApplyAllDecision(decision DuplicateDecision) bool {
	switch decision {
	case DuplicateDecisionOverwriteAll:
		return s.SetApplyAllPolicy(DuplicatePolicyOverwrite)
	case DuplicateDecisionSkipAll:
		return s.SetApplyAllPolicy(DuplicatePolicySkip)
	case DuplicateDecisionRenameAll:
		return s.SetApplyAllPolicy(DuplicatePolicyRename)
	default:
		return false
	}
}

func (s *DuplicateSession) ClearApplyAll() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.applyAll = ""
	s.mu.Unlock()
}
