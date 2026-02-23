package downloader

import (
	"reflect"
	"testing"
)

func TestSplitHLSAttributes(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: `BANDWIDTH=1280000`,
			want:  []string{`BANDWIDTH=1280000`},
		},
		{
			input: `BANDWIDTH=1280000,RESOLUTION=1280x720`,
			want:  []string{`BANDWIDTH=1280000`, `RESOLUTION=1280x720`},
		},
		{
			input: `CODECS="avc1.4d401e,mp4a.40.2",BANDWIDTH=1280000`,
			want:  []string{`CODECS="avc1.4d401e,mp4a.40.2"`, `BANDWIDTH=1280000`},
		},
		{
			input: `URI="https://example.com/key",METHOD=AES-128`,
			want:  []string{`URI="https://example.com/key"`, `METHOD=AES-128`},
		},
		{
			input: ``,
			want:  nil,
		},
	}

	for _, tt := range tests {
		got := splitHLSAttributes(tt.input)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("splitHLSAttributes(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseHLSAttributes(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]string
	}{
		{
			input: `BANDWIDTH=1280000,RESOLUTION=1280x720,CODECS="avc1.4d401e,mp4a.40.2"`,
			want: map[string]string{
				"BANDWIDTH":  "1280000",
				"RESOLUTION": "1280x720",
				"CODECS":     "avc1.4d401e,mp4a.40.2",
			},
		},
		{
			input: `bandwidth=1280000,resolution=1280x720`,
			want: map[string]string{
				"BANDWIDTH":  "1280000",
				"RESOLUTION": "1280x720",
			},
		},
		{
			input: ` URI="https://example.com/key" , METHOD=AES-128 `,
			want: map[string]string{
				"URI":    "https://example.com/key",
				"METHOD": "AES-128",
			},
		},
		{
			input: `INVALID`,
			want:  map[string]string{},
		},
		{
			input: `KEY=`,
			want: map[string]string{
				"KEY": "",
			},
		},
		{
			input: ` =VALUE`,
			want:  map[string]string{},
		},
		{
			input: ``,
			want:  map[string]string{},
		},
	}

	for _, tt := range tests {
		got := parseHLSAttributes(tt.input)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseHLSAttributes(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
