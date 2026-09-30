//go:build !windows

package downloader

import "fmt"

// osUnprotectData is only implemented on Windows. Chromium browsers on other
// platforms use libsecret/keychain which is not supported here; Firefox works.
func osUnprotectData(data []byte) ([]byte, error) {
	return nil, fmt.Errorf("chromium cookie decryption is only supported on Windows; use --cookies-file with an exported cookies.txt instead")
}
