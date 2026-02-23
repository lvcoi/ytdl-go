package db

import (
	"strings"
)

// ClassifyMediaType determines the media type (music, podcast, movie, video)
// based on YouTube metadata signals scraped during the GetVideoContext phase.
//
// Logic matrix:
//   - Music:   URL contains music.youtube.com, OR has Artist+Album tags,
//     OR channel name ends in " - Topic"
//   - Podcast: category is "Podcasts", OR playlist is a designated podcast playlist
//   - Movie:   category is "Movies", OR channel is "YouTube Movies & TV"
//   - Video:   fallback for everything else
func ClassifyMediaType(url, channelName, category, artist, album string, audioOnly bool) string {
	// Music detection
	// Check 1: Does the URL contain music.youtube.com?
	if strings.Contains(url, "music.youtube.com") {
		return "music"
	}
	// Check 2: Does the metadata contain explicit Artist and Album tags?
	if strings.TrimSpace(artist) != "" && strings.TrimSpace(album) != "" {
		return "music"
	}
	// Check 3: Does the YouTube Channel name end in " - Topic"?
	if strings.HasSuffix(channelName, " - Topic") {
		return "music"
	}
	// Audio-only downloads without other classification signals default to music
	if audioOnly {
		return "music"
	}

	// Podcast detection
	// Check 1: Category explicitly labeled "Podcasts"
	categoryLower := strings.ToLower(strings.TrimSpace(category))
	if categoryLower == "podcasts" {
		return "podcast"
	}

	// Movie detection
	// Check 1: Category is "Movies"
	if categoryLower == "movies" {
		return "movie"
	}
	// Check 2: Channel name is "YouTube Movies & TV"
	if channelName == "YouTube Movies & TV" {
		return "movie"
	}

	// Default fallback: standard video content
	return "video"
}
