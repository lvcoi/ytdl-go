package downloader

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/kkdai/youtube/v2"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid https", input: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", wantErr: false},
		{name: "valid http", input: "http://example.com/video.mp4", wantErr: false},
		{name: "missing scheme", input: "www.youtube.com/watch?v=dQw4w9WgXcQ", wantErr: true},
		{name: "missing host", input: "https:///video", wantErr: true},
		{name: "unsupported scheme", input: "ftp://example.com/video", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateURL(test.input)
			if test.wantErr && err == nil {
				t.Fatalf("expected error for %q", test.input)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", test.input, err)
			}
		})
	}
}

func TestWriteFormats(t *testing.T) {
	video := &youtube.Video{
		Formats: []youtube.Format{
			{
				ItagNo:        22,
				Quality:       "hd720",
				QualityLabel:  "720p",
				MimeType:      "video/mp4",
				Bitrate:       2000000,
				AudioChannels: 2,
				Width:         1280,
				Height:        720,
				ContentLength: 1234,
			},
		},
	}

	var buf bytes.Buffer
	if err := writeFormats(&buf, video); err != nil {
		t.Fatalf("writeFormats returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "itag") {
		t.Fatalf("expected header in output, got: %q", output)
	}
	if !strings.Contains(output, "720p") {
		t.Fatalf("expected format row in output, got: %q", output)
	}
}

func TestIsRestrictedAccess(t *testing.T) {
	tests := []struct {
		name       string
		errMsg     string
		wantResult bool
	}{
		{name: "nil error", errMsg: "", wantResult: false},
		{name: "private video", errMsg: "This video is private", wantResult: true},
		{name: "sign in required", errMsg: "Please sign in to view", wantResult: true},
		{name: "login required", errMsg: "Login required", wantResult: true},
		{name: "members only", errMsg: "This content is members only", wantResult: true},
		{name: "premium content", errMsg: "Premium subscription required", wantResult: true},
		{name: "copyright restriction", errMsg: "Copyright claim by owner", wantResult: true},
		{name: "unavailable", errMsg: "This video is unavailable", wantResult: true},
		{name: "age-restricted with hyphen", errMsg: "This video is age-restricted", wantResult: true},
		{name: "age restricted without hyphen", errMsg: "This video is age restricted", wantResult: true},
		{name: "not available", errMsg: "Video not available in your country", wantResult: true},
		{name: "case insensitive - PRIVATE", errMsg: "PRIVATE VIDEO", wantResult: true},
		{name: "case insensitive - MiXeD", errMsg: "MeMbErS OnLy", wantResult: true},
		{name: "network error", errMsg: "network timeout", wantResult: false},
		{name: "generic error", errMsg: "something went wrong", wantResult: false},
		{name: "false positive - availability", errMsg: "check availability", wantResult: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error
			if test.errMsg != "" {
				err = fmt.Errorf("%s", test.errMsg)
			}
			result := isRestrictedAccess(err)
			if result != test.wantResult {
				t.Fatalf("isRestrictedAccess(%v) = %v, want %v", err, result, test.wantResult)
			}
		})
	}
}
