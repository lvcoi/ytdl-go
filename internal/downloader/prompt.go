package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var promptMu sync.Mutex
var promptCond = sync.NewCond(&promptMu)
var promptActive bool

func beginPrompt() {
	promptMu.Lock()
	for promptActive {
		promptCond.Wait()
	}
	promptActive = true
	promptMu.Unlock()
}

func endPrompt() {
	promptMu.Lock()
	promptActive = false
	promptCond.Broadcast()
	promptMu.Unlock()
}

func waitForPromptClear() {
	promptMu.Lock()
	for promptActive {
		promptCond.Wait()
	}
	promptMu.Unlock()
}

func applyDuplicatePolicy(policy DuplicatePolicy, path, baseDir string) (string, bool, error) {
	switch policy {
	case DuplicatePolicyOverwrite:
		return path, false, nil
	case DuplicatePolicySkip:
		return path, true, nil
	case DuplicatePolicyRename:
		newPath, err := nextAvailablePath(path, baseDir)
		return newPath, false, err
	default:
		return "", false, errors.New("duplicate prompt failed")
	}
}

func applyDuplicateDecision(decision DuplicateDecision, path, baseDir string, session *DuplicateSession) (string, bool, error) {
	switch decision {
	case DuplicateDecisionOverwrite:
		return path, false, nil
	case DuplicateDecisionOverwriteAll:
		if session != nil {
			session.SetApplyAllDecision(decision)
		}
		return path, false, nil
	case DuplicateDecisionSkip:
		return path, true, nil
	case DuplicateDecisionSkipAll:
		if session != nil {
			session.SetApplyAllDecision(decision)
		}
		return path, true, nil
	case DuplicateDecisionRename:
		newPath, err := nextAvailablePath(path, baseDir)
		return newPath, false, err
	case DuplicateDecisionRenameAll:
		if session != nil {
			session.SetApplyAllDecision(decision)
		}
		newPath, err := nextAvailablePath(path, baseDir)
		return newPath, false, err
	case DuplicateDecisionCancel:
		return "", false, errors.New("aborted by user")
	default:
		return "", false, errors.New("duplicate prompt failed")
	}
}

func promptChoiceToDecision(choice promptChoice) DuplicateDecision {
	switch choice {
	case promptOverwrite:
		return DuplicateDecisionOverwrite
	case promptOverwriteAll:
		return DuplicateDecisionOverwriteAll
	case promptSkip:
		return DuplicateDecisionSkip
	case promptSkipAll:
		return DuplicateDecisionSkipAll
	case promptRename:
		return DuplicateDecisionRename
	case promptRenameAll:
		return DuplicateDecisionRenameAll
	case promptQuit:
		return DuplicateDecisionCancel
	default:
		return ""
	}
}

func handleExistingPath(path, baseDir string, opts Options, printer *Printer) (string, bool, error) {
	stdinTTY := isTerminal(os.Stdin)
	if stdinTTY {
		waitForPromptClear()
	}
	base := baseDir
	if base == "" {
		base = "."
	}
	relPath, err := filepath.Rel(base, path)
	if err != nil {
		return "", false, wrapCategory(CategoryFilesystem, err)
	}
	validatedPath, err := validatedOutputPath(relPath, baseDir)
	if err != nil {
		return "", false, wrapCategory(CategoryFilesystem, err)
	}
	path = validatedPath
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, false, nil
		}
		return "", false, wrapCategory(CategoryFilesystem, err)
	}
	if info.IsDir() {
		return "", false, wrapCategory(CategoryFilesystem, fmt.Errorf("output path is a directory: %s", path))
	}

	if policy, ok := opts.DuplicateSession.ApplyAllPolicy(); ok {
		return applyDuplicatePolicy(policy, path, baseDir)
	}

	policy := opts.OnDuplicate.OrDefault()
	if policy != DuplicatePolicyPrompt {
		return applyDuplicatePolicy(policy, path, baseDir)
	}

	if opts.DuplicatePrompter != nil {
		decision, err := opts.DuplicatePrompter.PromptDuplicate(path)
		if err != nil {
			return "", false, wrapCategory(CategoryFilesystem, err)
		}
		return applyDuplicateDecision(decision, path, baseDir, opts.DuplicateSession)
	}

	// Use TUI-based prompt if available (either ProgressManager or SeamlessTUI)
	if stdinTTY && printer != nil && printer.progressEnabled {
		var choice promptChoice
		var err error

		if printer.seamlessTUI != nil {
			choice, err = printer.seamlessTUI.PromptDuplicate(path)
		} else if printer.manager != nil {
			choice, err = printer.manager.PromptDuplicate(path)
		} else {
			// Fall through to stdin prompt below
			goto stdinPrompt
		}

		if err != nil {
			return "", false, wrapCategory(CategoryFilesystem, err)
		}
		return applyDuplicateDecision(promptChoiceToDecision(choice), path, baseDir, opts.DuplicateSession)
	}

	if !stdinTTY {
		if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "warning: %s exists; overwriting (stdin not a TTY)\n", path)
		}
		return path, false, nil
	}

stdinPrompt:

	reader := bufio.NewReader(os.Stdin)
	beginPrompt()
	defer endPrompt()
	for {
		fmt.Fprintf(os.Stderr, "%s exists.\n", path)
		fmt.Fprint(os.Stderr, "  [o]verwrite, [s]kip, [r]ename, [q]uit\n")
		fmt.Fprint(os.Stderr, "  [O]verwrite all, [S]kip all, [R]ename all: ")
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return "", false, wrapCategory(CategoryFilesystem, readErr)
		}
		switch strings.TrimSpace(line) {
		case "o", "overwrite":
			return applyDuplicateDecision(DuplicateDecisionOverwrite, path, baseDir, opts.DuplicateSession)
		case "O", "Overwrite":
			return applyDuplicateDecision(DuplicateDecisionOverwriteAll, path, baseDir, opts.DuplicateSession)
		case "s", "skip":
			return applyDuplicateDecision(DuplicateDecisionSkip, path, baseDir, opts.DuplicateSession)
		case "S", "Skip":
			return applyDuplicateDecision(DuplicateDecisionSkipAll, path, baseDir, opts.DuplicateSession)
		case "r", "rename":
			return applyDuplicateDecision(DuplicateDecisionRename, path, baseDir, opts.DuplicateSession)
		case "R", "Rename":
			return applyDuplicateDecision(DuplicateDecisionRenameAll, path, baseDir, opts.DuplicateSession)
		case "q", "quit":
			return applyDuplicateDecision(DuplicateDecisionCancel, path, baseDir, opts.DuplicateSession)
		default:
			fmt.Fprintln(os.Stderr, "please enter o, s, r, q (or uppercase for 'all')")
		}
	}
}

func nextAvailablePath(path, baseDir string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		baseDirectory := baseDir
		if baseDirectory == "" {
			baseDirectory = "."
		}
		relCandidate, err := filepath.Rel(baseDirectory, candidate)
		if err != nil {
			return "", wrapCategory(CategoryFilesystem, err)
		}
		validated, err := validatedOutputPath(relCandidate, baseDir)
		if err != nil {
			return "", wrapCategory(CategoryFilesystem, err)
		}
		if _, err := os.Stat(validated); err != nil {
			if os.IsNotExist(err) {
				return validated, nil
			}
			return "", wrapCategory(CategoryFilesystem, err)
		}
	}
	return "", wrapCategory(CategoryFilesystem, fmt.Errorf("unable to find available filename for %s", path))
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
