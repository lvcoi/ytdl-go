package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func resolveOutputPath(template string, video *youtube.Video, format *youtube.Format, ctxInfo outputContext) (string, error) {
	if template == "" {
		template = "{title}.{ext}"
	}

	title := sanitize(video.Title)
	ext := mimeToExt(format.MimeType)
	artist := video.Author
	album := ctxInfo.EntryAlbum
	quality := format.QualityLabel
	if quality == "" {
		if b := bitrateForFormat(format); b > 0 {
			quality = fmt.Sprintf("%dk", b/1000)
		}
	}
	playlistTitle := ""
	playlistID := ""
	index := ""
	total := ""
	if ctxInfo.Playlist != nil {
		playlistTitle = sanitize(ctxInfo.Playlist.Title)
		playlistID = ctxInfo.Playlist.ID
		if ctxInfo.Index > 0 {
			index = strconv.Itoa(ctxInfo.Index)
		}
		if ctxInfo.Total > 0 {
			total = strconv.Itoa(ctxInfo.Total)
		}
		if ctxInfo.EntryTitle != "" {
			title = sanitize(ctxInfo.EntryTitle)
		}
		if ctxInfo.EntryAuthor != "" {
			artist = ctxInfo.EntryAuthor
		}
	}
	artist = sanitize(artist)
	album = sanitizeOptional(album)

	replacer := strings.NewReplacer(
		"{title}", title,
		"{artist}", artist,
		"{album}", album,
		"{id}", video.ID,
		"{ext}", ext,
		"{quality}", quality,
		"{playlist_title}", playlistTitle,
		"{playlist_id}", playlistID,
		"{index}", index,
		"{count}", total,
	)
	path := replacer.Replace(template)

	// Treat existing directory or explicit trailing slash as "put file inside".
	if strings.HasSuffix(template, "/") {
		path = filepath.Join(path, fmt.Sprintf("%s.%s", title, ext))
	} else if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, fmt.Sprintf("%s.%s", title, ext))
	}

	if filepath.Ext(path) == "" {
		path = path + "." + ext
	}
	return path, nil
}

func sanitize(name string) string {
	invalid := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	clean := invalid.ReplaceAllString(name, "-")
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return "video"
	}
	return clean
}

func sanitizeOptional(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	return sanitize(name)
}

func mimeToExt(mime string) string {
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	parts := strings.Split(mime, "/")
	if len(parts) == 2 {
		switch parts[1] {
		case "3gpp":
			return "3gp"
		default:
			return parts[1]
		}
	}
	return "bin"
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for n >= unit*div && exp < 4 {
		div *= unit
		exp++
	}
	value := float64(n) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f%s", value, suffix[exp])
}

func bitrateForFormat(f *youtube.Format) int {
	if f.Bitrate > 0 {
		return f.Bitrate
	}
	if f.AverageBitrate > 0 {
		return f.AverageBitrate
	}
	return 0
}
