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

// duplicateAction represents the user's choice for handling existing files
type duplicateAction int

const (
	duplicateAskEach duplicateAction = iota
	duplicateOverwriteAll
	duplicateSkipAll
	duplicateRenameAll
)

var globalDuplicateAction = duplicateAskEach

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

func handleExistingPath(path string, opts Options, printer *Printer) (string, bool, error) {
	if isTerminal(os.Stdin) {
		waitForPromptClear()
	}
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

	// Check if we have a global "apply to all" action
	switch globalDuplicateAction {
	case duplicateOverwriteAll:
		return path, false, nil
	case duplicateSkipAll:
		return path, true, nil
	case duplicateRenameAll:
		newPath, err := nextAvailablePath(path)
		return newPath, false, err
	}

	if !isTerminal(os.Stdin) {
		if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "warning: %s exists; overwriting (stdin not a TTY)\n", path)
		}
		return path, false, nil
	}

	if printer != nil && printer.progressEnabled && printer.manager != nil {
		choice, err := printer.manager.PromptDuplicate(path)
		if err != nil {
			return "", false, wrapCategory(CategoryFilesystem, err)
		}
		switch choice {
		case promptOverwrite:
			return path, false, nil
		case promptOverwriteAll:
			globalDuplicateAction = duplicateOverwriteAll
			return path, false, nil
		case promptSkip:
			return path, true, nil
		case promptSkipAll:
			globalDuplicateAction = duplicateSkipAll
			return path, true, nil
		case promptRename:
			newPath, err := nextAvailablePath(path)
			return newPath, false, err
		case promptRenameAll:
			globalDuplicateAction = duplicateRenameAll
			newPath, err := nextAvailablePath(path)
			return newPath, false, err
		case promptQuit:
			return "", false, errors.New("aborted by user")
		default:
			return "", false, errors.New("duplicate prompt failed")
		}
	}

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
			return path, false, nil
		case "O", "Overwrite":
			globalDuplicateAction = duplicateOverwriteAll
			return path, false, nil
		case "s", "skip":
			return path, true, nil
		case "S", "Skip":
			globalDuplicateAction = duplicateSkipAll
			return path, true, nil
		case "r", "rename":
			newPath, err := nextAvailablePath(path)
			return newPath, false, err
		case "R", "Rename":
			globalDuplicateAction = duplicateRenameAll
			newPath, err := nextAvailablePath(path)
			return newPath, false, err
		case "q", "quit":
			return "", false, errors.New("aborted by user")
		default:
			fmt.Fprintln(os.Stderr, "please enter o, s, r, q (or uppercase for 'all')")
		}
	}
}

func nextAvailablePath(path string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if _, err := os.Stat(candidate); err != nil {
			if os.IsNotExist(err) {
				return candidate, nil
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
