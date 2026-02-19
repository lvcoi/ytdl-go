package db

import "testing"

func TestClassifyMediaType(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		channelName string
		category    string
		hasArtist   bool
		hasAlbum    bool
		audioOnly   bool
		want        string
	}{
		{
			name: "music.youtube.com URL",
			url:  "https://music.youtube.com/watch?v=abc",
			want: "music",
		},
		{
			name:      "has artist and album",
			url:       "https://youtube.com/watch?v=abc",
			hasArtist: true,
			hasAlbum:  true,
			want:      "music",
		},
		{
			name:        "Topic channel",
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
			name:     "podcast category",
			url:      "https://youtube.com/watch?v=abc",
			category: "Podcasts",
			want:     "podcast",
		},
		{
			name:     "movie category",
			url:      "https://youtube.com/watch?v=abc",
			category: "Movies",
			want:     "movie",
		},
		{
			name:        "YouTube Movies channel",
			url:         "https://youtube.com/watch?v=abc",
			channelName: "YouTube Movies & TV",
			want:        "movie",
		},
		{
			name: "default video",
			url:  "https://youtube.com/watch?v=abc",
			want: "video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyMediaType(tt.url, tt.channelName, tt.category, tt.hasArtist, tt.hasAlbum, tt.audioOnly)
			if got != tt.want {
				t.Errorf("ClassifyMediaType() = %q, want %q", got, tt.want)
			}
		})
	}
}
