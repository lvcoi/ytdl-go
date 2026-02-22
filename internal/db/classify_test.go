package db

import "testing"

func TestClassifyMediaType(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		channelName string
		category    string
		artist      string
		album       string
		audioOnly   bool
		want        string
	}{
		// Music checks
		{
			name: "music.youtube.com URL",
			url:  "https://music.youtube.com/watch?v=abc",
			want: "music",
		},
		{
			name:   "has artist and album tags",
			url:    "https://youtube.com/watch?v=abc",
			artist: "Taylor Swift",
			album:  "1989",
			want:   "music",
		},
		{
			name:        "Topic channel suffix",
			url:         "https://youtube.com/watch?v=abc",
			channelName: "Taylor Swift - Topic",
			want:        "music",
		},
		{
			name:      "audio only download",
			url:       "https://youtube.com/watch?v=abc",
			audioOnly: true,
			want:      "music",
		},
		{
			name:   "artist without album is not music",
			url:    "https://youtube.com/watch?v=abc",
			artist: "Some Creator",
			want:   "video",
		},
		// Podcast checks
		{
			name:     "podcast category",
			url:      "https://youtube.com/watch?v=abc",
			category: "Podcasts",
			want:     "podcast",
		},
		// Movie checks
		{
			name:     "movie category",
			url:      "https://youtube.com/watch?v=abc",
			category: "Movies",
			want:     "movie",
		},
		{
			name:        "YouTube Movies and TV channel",
			url:         "https://youtube.com/watch?v=abc",
			channelName: "YouTube Movies & TV",
			want:        "movie",
		},
		// Default fallback
		{
			name: "default video",
			url:  "https://youtube.com/watch?v=abc",
			want: "video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyMediaType(tt.url, tt.channelName, tt.category, tt.artist, tt.album, tt.audioOnly)
			if got != tt.want {
				t.Errorf("ClassifyMediaType() = %q, want %q", got, tt.want)
			}
		})
	}
}
