//go:build !windows

package downloader

import (
	"os"
	"os/signal"
	"syscall"
)

func resizeSignalChannel() (<-chan os.Signal, func()) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGWINCH)
	return sigch, func() {
		signal.Stop(sigch)
	}
}
