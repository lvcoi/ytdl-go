package downloader

import (
	"testing"

	youtube "github.com/lvcoi/ytdl-lib/v2"
)

// testVideo returns a Video with a mix of progressive, audio-only, and
// video-only formats for exercising selectFormat.
func testVideo() *youtube.Video {
	return &youtube.Video{
		ID:    "test123",
		Title: "Test Video",
		Formats: youtube.FormatList{
			// Progressive (audio+video) formats
			{ItagNo: 18, MimeType: "video/mp4", Width: 640, Height: 360, AudioChannels: 2, Bitrate: 500_000},
			{ItagNo: 22, MimeType: "video/mp4", Width: 1280, Height: 720, AudioChannels: 2, Bitrate: 2_000_000},
			{ItagNo: 37, MimeType: "video/mp4", Width: 1920, Height: 1080, AudioChannels: 2, Bitrate: 4_000_000},
			// Audio-only formats
			{ItagNo: 140, MimeType: "audio/mp4", AudioChannels: 2, Bitrate: 128_000},
			{ItagNo: 251, MimeType: "audio/webm", AudioChannels: 2, Bitrate: 160_000},
			// Video-only format (no audio)
			{ItagNo: 137, MimeType: "video/mp4", Width: 1920, Height: 1080, Bitrate: 4_000_000},
		},
	}
}

func TestSelectFormat(t *testing.T) {
	tests := []struct {
		name      string
		opts      Options
		wantItag  int
		wantErr   bool
	}{
		{
			name:     "default selects best progressive format",
			opts:     Options{},
			wantItag: 37, // 1080p progressive, highest resolution
		},
		{
			name:     "itag selection returns matching format",
			opts:     Options{Itag: 22},
			wantItag: 22,
		},
		{
			name:    "itag not found returns error",
			opts:    Options{Itag: 9999},
			wantErr: true,
		},
		{
			name:     "audio-only selects highest bitrate audio",
			opts:     Options{AudioOnly: true},
			wantItag: 251, // 160kbps > 128kbps
		},
		{
			name:     "quality filter selects format at or below 720p",
			opts:     Options{Quality: "720p"},
			wantItag: 22, // 720p progressive
		},
	}

	video := testVideo()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectFormat(video, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got format itag %d", got.ItagNo)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ItagNo != tt.wantItag {
				t.Errorf("got itag %d, want %d", got.ItagNo, tt.wantItag)
			}
		})
	}
}
