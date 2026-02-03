package downloader

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type jsonResult struct {
	Type          string `json:"type"`
	Status        string `json:"status"`
	URL           string `json:"url,omitempty"`
	ID            string `json:"id,omitempty"`
	Title         string `json:"title,omitempty"`
	Output        string `json:"output,omitempty"`
	Bytes         int64  `json:"bytes,omitempty"`
	Retries       bool   `json:"retried,omitempty"`
	Skipped       bool   `json:"skipped,omitempty"`
	Error         string `json:"error,omitempty"`
	PlaylistID    string `json:"playlist_id,omitempty"`
	PlaylistTitle string `json:"playlist_title,omitempty"`
	Index         int    `json:"index,omitempty"`
	Total         int    `json:"total,omitempty"`
}

type formatInfo struct {
	Itag         int    `json:"itag"`
	MimeType     string `json:"mime_type"`
	Quality      string `json:"quality"`
	QualityLabel string `json:"quality_label"`
	Bitrate      int    `json:"bitrate"`
	AvgBitrate   int    `json:"average_bitrate"`
	AudioQuality string `json:"audio_quality"`
	SampleRate   string `json:"audio_sample_rate"`
	Channels     int    `json:"audio_channels"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Size         int64  `json:"content_length"`
	Ext          string `json:"ext"`
}

func emitJSONResult(res jsonResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(res)
}

func renderFormatsJSON(video *youtube.Video, playlistID, playlistTitle string, index, total int) error {
	payload := struct {
		Type          string       `json:"type"`
		PlaylistID    string       `json:"playlist_id,omitempty"`
		PlaylistTitle string       `json:"playlist_title,omitempty"`
		Index         int          `json:"index,omitempty"`
		Total         int          `json:"total,omitempty"`
		ID            string       `json:"id"`
		Title         string       `json:"title"`
		Formats       []formatInfo `json:"formats"`
	}{
		Type:          "formats",
		PlaylistID:    playlistID,
		PlaylistTitle: playlistTitle,
		Index:         index,
		Total:         total,
		ID:            video.ID,
		Title:         video.Title,
	}
	for _, f := range video.Formats {
		payload.Formats = append(payload.Formats, formatInfo{
			Itag:         f.ItagNo,
			MimeType:     f.MimeType,
			Quality:      f.Quality,
			QualityLabel: f.QualityLabel,
			Bitrate:      f.Bitrate,
			AvgBitrate:   f.AverageBitrate,
			AudioQuality: f.AudioQuality,
			SampleRate:   f.AudioSampleRate,
			Channels:     f.AudioChannels,
			Width:        f.Width,
			Height:       f.Height,
			Size:         int64(f.ContentLength),
			Ext:          mimeToExt(f.MimeType),
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(payload)
}

func printVideoInfo(video *youtube.Video) error {
	payload := struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Author      string `json:"author"`
		Duration    int    `json:"duration_seconds"`
	}{
		ID:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		Author:      video.Author,
		Duration:    int(video.Duration.Seconds()),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func printPlaylistInfo(playlist *youtube.Playlist) error {
	type entryInfo struct {
		Index    int    `json:"index"`
		ID       string `json:"id"`
		Title    string `json:"title"`
		Author   string `json:"author"`
		Duration int    `json:"duration_seconds"`
	}

	payload := struct {
		ID          string      `json:"id"`
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Author      string      `json:"author"`
		VideoCount  int         `json:"video_count"`
		Videos      []entryInfo `json:"videos"`
	}{
		ID:          playlist.ID,
		Title:       playlist.Title,
		Description: playlist.Description,
		Author:      playlist.Author,
		VideoCount:  len(playlist.Videos),
	}

	for i, entry := range playlist.Videos {
		if entry == nil {
			continue
		}
		payload.Videos = append(payload.Videos, entryInfo{
			Index:    i + 1,
			ID:       entry.ID,
			Title:    entry.Title,
			Author:   entry.Author,
			Duration: int(entry.Duration.Seconds()),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func validateOutputFile(path string, format *youtube.Format) error {
	info, err := os.Stat(path)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("stat output: %w", err))
	}
	if info.Size() == 0 {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("output file is empty"))
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4", ".m4v", ".mov", ".m4a", ".m4s":
		return validateMP4(path)
	case ".webm", ".mkv":
		return validateEBML(path)
	case ".ts":
		return validateMPEGTS(path)
	case ".mp3":
		return validateMP3(path)
	default:
		if format != nil && strings.Contains(strings.ToLower(format.MimeType), "mp4") {
			return validateMP4(path)
		}
		return nil
	}
}

func readHeader(path string, size int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func validateMP4(path string) error {
	header, err := readHeader(path, 12)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("read mp4 header: %w", err))
	}
	if len(header) < 8 || string(header[4:8]) != "ftyp" {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("invalid mp4 header"))
	}
	body, err := readHeader(path, 1024*1024)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("read mp4 body: %w", err))
	}
	if !bytes.Contains(body, []byte("moov")) && !bytes.Contains(body, []byte("moof")) {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("missing moov/moof atom"))
	}
	return nil
}

func validateEBML(path string) error {
	header, err := readHeader(path, 4)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("read ebml header: %w", err))
	}
	if len(header) < 4 || binary.BigEndian.Uint32(header) != 0x1A45DFA3 {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("invalid webm header"))
	}
	return nil
}

func validateMPEGTS(path string) error {
	header, err := readHeader(path, 189)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("read ts header: %w", err))
	}
	if len(header) < 1 || header[0] != 0x47 {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("invalid transport stream header"))
	}
	if len(header) >= 189 && header[188] != 0x47 {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("invalid transport stream sync"))
	}
	return nil
}

func validateMP3(path string) error {
	header, err := readHeader(path, 3)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("read mp3 header: %w", err))
	}
	if len(header) < 3 {
		return wrapCategory(CategoryUnsupported, fmt.Errorf("invalid mp3 header"))
	}
	if string(header) == "ID3" || header[0] == 0xFF {
		return nil
	}
	return wrapCategory(CategoryUnsupported, fmt.Errorf("invalid mp3 header"))
}
