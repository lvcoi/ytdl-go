package downloader

import "testing"

func TestValidateInputURL(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid http", input: "http://example.com/video.mp4", wantErr: false},
		{name: "valid https", input: "https://example.com/watch?v=123", wantErr: false},
		{name: "missing scheme", input: "example.com/video.mp4", wantErr: true},
		{name: "empty", input: " ", wantErr: true},
		{name: "unsupported scheme", input: "ftp://example.com/video.mp4", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInputURL(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
		})
	}
}
