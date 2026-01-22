//go:build windows

package downloader

import "os"

func resizeSignalChannel() (<-chan os.Signal, func()) {
	return nil, nil
}
