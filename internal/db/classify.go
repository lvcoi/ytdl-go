package db

import (
	"strings"
)

// ClassifyMediaType determines the media type (music, podcast, movie, video)
// based on YouTube metadata signals.
func ClassifyMediaType(url, channelName, category string, hasArtist, hasAlbum, audioOnly bool) string {
	// Music detection
	if strings.Contains(url, "music.youtube.com") {
		return "music"
	}
	if hasArtist && hasAlbum {
		return "music"
	}
	if strings.HasSuffix(channelName, " - Topic") {
		return "music"
	}
	if audioOnly {
		return "music"
	}

	// Podcast detection
	categoryLower := strings.ToLower(category)
	if categoryLower == "podcasts" {
		return "podcast"
	}

	// Movie detection
	if categoryLower == "movies" {
		return "movie"
	}
	if channelName == "YouTube Movies & TV" {
		return "movie"
	}

	// Default fallback
	return "video"
}
