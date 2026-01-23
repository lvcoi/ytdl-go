package downloader

import (
	"bytes"
	"errors"
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
		{name: "empty scheme", input: "//www.youtube.com/watch?v=dQw4w9WgXcQ", wantErr: true},
		{name: "missing host", input: "https:///video", wantErr: true},
		{name: "unsupported scheme", input: "ftp://example.com/video", wantErr: true},
		{name: "playlist id", input: "PL59FEE129ADFF2B12", wantErr: false},
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

func TestIsRestrictedAccess(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantHit bool
	}{
		{name: "private", err: errors.New("video is private"), wantHit: true},
		{name: "sign in", err: errors.New("sign in to confirm your age"), wantHit: true},
		{name: "members only", err: errors.New("members only video"), wantHit: true},
		{name: "premium", err: errors.New("premium content"), wantHit: true},
		{name: "copyright", err: errors.New("copyright claim"), wantHit: true},
		{name: "video unavailable", err: errors.New("video unavailable"), wantHit: true},
		{name: "content unavailable", err: errors.New("content unavailable"), wantHit: true},
		{name: "age restricted", err: errors.New("age restricted"), wantHit: true},
		{name: "not available", err: errors.New("not available in your country"), wantHit: true},
		{name: "nil error", err: nil, wantHit: false},
		{name: "unrelated", err: errors.New("temporary network error"), wantHit: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isRestrictedAccess(test.err); got != test.wantHit {
				t.Fatalf("expected %v for %v, got %v", test.wantHit, test.err, got)
			}
		})
	}
}

func TestWrapAccessError(t *testing.T) {
	restrictedErr := errors.New("video is private")
	wrapped := wrapAccessError(restrictedErr)
	if wrapped == restrictedErr {
		t.Fatalf("expected restricted error to be wrapped")
	}
	if wrapped == nil || !strings.Contains(wrapped.Error(), "restricted access") {
		t.Fatalf("expected restricted access prefix, got %v", wrapped)
	}

	plainErr := errors.New("plain error")
	if got := wrapAccessError(plainErr); got != plainErr {
		t.Fatalf("expected plain error to be unchanged, got %v", got)
	}
	if got := wrapAccessError(nil); got != nil {
		t.Fatalf("expected nil to remain nil, got %v", got)
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
