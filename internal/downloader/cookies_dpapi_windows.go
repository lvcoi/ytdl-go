//go:build windows

package downloader

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// osUnprotectData decrypts DPAPI protected data on Windows.
func osUnprotectData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty DPAPI blob")
	}
	in := windows.DataBlob{
		Size: uint32(len(data)),
		Data: &data[0],
	}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return nil, fmt.Errorf("CryptUnprotectData: %w", err)
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	if out.Size == 0 || out.Data == nil {
		return nil, fmt.Errorf("CryptUnprotectData returned empty data")
	}
	plaintext := make([]byte, out.Size)
	copy(plaintext, unsafe.Slice(out.Data, out.Size))
	return plaintext, nil
}
