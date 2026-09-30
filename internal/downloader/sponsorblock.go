package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lvcoi/ytdl-lib/v2"
)

// sponsorSegment is a single SponsorBlock skip segment in milliseconds.
type sponsorSegment struct {
	Start      float64  `json:"start"` // seconds
	End        float64  `json:"end"`   // seconds
	Category   string   `json:"category"`
	ActionType string   `json:"actionType"`
	Description string  `json:"description"`
}

var sponsorblockCategories = []string{
	"sponsor", "selfpromo", "interaction", "intro", "outro",
	"preview", "music_offtopic", "filler", "poi_highlight", "chapter",
}

func parseSponsorblockCategories(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") && !strings.HasPrefix(part, "music_off") {
			// Allow "selfpromo-partial" style aliases by stripping suffix.
			part = strings.SplitN(part, "-", 2)[0]
		}
		valid := false
		for _, cat := range sponsorblockCategories {
			if part == cat {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("unknown SponsorBlock category %q (valid: %s)", part, strings.Join(sponsorblockCategories, ", "))
		}
		out = append(out, part)
	}
	return out, nil
}

// fetchSponsorblockSegments queries the SponsorBlock API for a video.
func fetchSponsorblockSegments(ctx context.Context, videoID string, categories []string, timeout time.Duration) ([]sponsorSegment, error) {
	if len(categories) == 0 {
		return nil, nil
	}
	apiURL := fmt.Sprintf("https://sponsor.ajay.app/api/skipSegments?videoID=%s&categories=%s",
		urlQueryEscape(videoID), urlQueryEscape(stringOrJoin(categories)))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, wrapCategory(CategoryNetwork, err)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, wrapCategory(CategoryNetwork, fmt.Errorf("querying SponsorBlock: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // no segments known
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, wrapCategory(CategoryNetwork, fmt.Errorf("querying SponsorBlock: unexpected status %d", resp.StatusCode))
	}
	var segments []struct {
		Segment  []float64 `json:"segment"`
		Category string    `json:"category"`
		ActionType string  `json:"actionType"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024)).Decode(&segments); err != nil {
		return nil, wrapCategory(CategoryNetwork, fmt.Errorf("parsing SponsorBlock response: %w", err))
	}
	out := make([]sponsorSegment, 0, len(segments))
	for _, s := range segments {
		if len(s.Segment) != 2 {
			continue
		}
		out = append(out, sponsorSegment{
			Start:      s.Segment[0],
			End:        s.Segment[1],
			Category:   s.Category,
			ActionType: s.ActionType,
			Description: s.Description,
		})
	}
	return out, nil
}

func stringOrJoin(parts []string) string {
	return "[\"" + strings.Join(parts, "\",\"") + "\"]"
}

func urlQueryEscape(value string) string {
	return strings.ReplaceAll(value, "\"", "%22")
}

// chapter describes a media chapter in milliseconds.
type chapter struct {
	StartMs     int64
	EndMs       int64
	Title       string
}

// complementSegments converts skip segments into the kept time ranges.
func complementSegments(segments []sponsorSegment, durationSec float64) []chapter {
	sortSegments(segments)
	var keeps []chapter
	cursor := 0.0 // seconds
	for _, s := range segments {
		if s.Start > cursor {
			keeps = append(keeps, chapter{
				StartMs: int64(cursor * 1000),
				EndMs:   int64(s.Start * 1000),
				Title:   "",
			})
		}
		if s.End > cursor {
			cursor = s.End
		}
	}
	if durationSec <= 0 || cursor < durationSec {
		keeps = append(keeps, chapter{
			StartMs: int64(cursor * 1000),
			EndMs:   int64(durationSec * 1000),
		})
	}
	return keeps
}

func sortSegments(segments []sponsorSegment) {
	for i := 1; i < len(segments); i++ {
		for j := i; j > 0 && segments[j].Start < segments[j-1].Start; j-- {
			segments[j], segments[j-1] = segments[j-1], segments[j]
		}
	}
}

// segmentsToChapters converts SponsorBlock segments into chapter markers.
func segmentsToChapters(segments []sponsorSegment, durationSec float64) []chapter {
	sortSegments(segments)
	chapters := make([]chapter, 0, len(segments))
	for _, s := range segments {
		end := s.End
		if durationSec > 0 && end > durationSec {
			end = durationSec
		}
		title := s.Category
		if s.Description != "" {
			title = s.Category + ": " + s.Description
		}
		chapters = append(chapters, chapter{
			StartMs: int64(s.Start * 1000),
			EndMs:   int64(end * 1000),
			Title:   title,
		})
	}
	return chapters
}

// writeFFMetadataFile writes an ffmetadata file with the given chapters and
// returns its path.
func writeFFMetadataFile(outputPath, baseDir string, chapters []chapter) (string, error) {
	path, err := artifactPath(outputPath, ".meta.txt", baseDir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(";FFMETADATA1\n")
	for i, ch := range chapters {
		title := ch.Title
		if title == "" {
			title = "Chapter " + strconv.Itoa(i+1)
		}
		fmt.Fprintf(&b, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=%s\n", ch.StartMs, ch.EndMs, strings.ReplaceAll(title, "\n", " "))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("writing chapter metadata: %w", err))
	}
	return path, nil
}

// embedChapters embeds chapter markers into the media file via ffmpeg
// (stream copy; works for mp4, m4a, mkv, webm, mp3).
func embedChapters(outputPath, baseDir string, chapters []chapter) error {
	if len(chapters) == 0 {
		return nil
	}
	metaPath, err := writeFFMetadataFile(outputPath, baseDir, chapters)
	if err != nil {
		return err
	}
	defer os.Remove(metaPath)

	dir := filepath.Dir(outputPath)
	tmpFile := filepath.Join(dir, ".tmp_chapters_"+filepath.Base(outputPath))
	args := []string{
		"-y",
		"-i", outputPath,
		"-i", metaPath,
		"-map", "0",
		"-map_metadata", "1",
		"-map_chapters", "1",
		"-c", "copy",
		tmpFile,
	}
	cmd := ffmpegCommand(args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to embed chapters: %s: %w", strings.TrimSpace(string(output)), err)
	}
	if err := os.Rename(tmpFile, outputPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace original file with chaptered version: %w", err)
	}
	return nil
}

// removeSponsorSegments cuts the marked segments out of the finished file
// using ffmpeg stream-copy cuts and concatenation.
func removeSponsorSegments(outputPath, baseDir string, skipSegments []sponsorSegment, durationSec float64) error {
	keeps := complementSegments(skipSegments, durationSec)
	if len(keeps) == 0 {
		return nil
	}
	// If nothing is actually removed (single keep spanning whole file), bail out.
	if len(keeps) == 1 && keeps[0].StartMs <= 0 && (durationSec <= 0 || int64(durationSec*1000) <= keeps[0].EndMs+1000) {
		return nil
	}

	dir := filepath.Dir(outputPath)
	partsDir := filepath.Join(dir, ".tmp_cuts_"+filepath.Base(outputPath))
	if err := os.MkdirAll(partsDir, 0o755); err != nil {
		return wrapCategory(CategoryFilesystem, err)
	}
	defer os.RemoveAll(partsDir)

	var listFile strings.Builder
	for i, keep := range keeps {
		part := filepath.Join(partsDir, fmt.Sprintf("part%03d%s", i, filepath.Ext(outputPath)))
		args := []string{
			"-y",
			"-ss", fmt.Sprintf("%.3f", float64(keep.StartMs)/1000),
			"-i", outputPath,
		}
		if keep.EndMs > keep.StartMs {
			args = append(args, "-to", fmt.Sprintf("%.3f", float64(keep.EndMs)/1000))
		}
		args = append(args, "-c", "copy", part)
		cmd := ffmpegCommand(args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ffmpeg cut segment %d failed: %s: %w", i+1, strings.TrimSpace(string(output)), err)
		}
		listFile.WriteString(fmt.Sprintf("file '%s'\n", strings.ReplaceAll(part, "'", "'\\''")))
	}

	listPath := filepath.Join(partsDir, "concat.txt")
	if err := os.WriteFile(listPath, []byte(listFile.String()), 0o644); err != nil {
		return wrapCategory(CategoryFilesystem, err)
	}
	tmpFile := filepath.Join(dir, ".tmp_removed_"+filepath.Base(outputPath))
	args := []string{
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		tmpFile,
	}
	cmd := ffmpegCommand(args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("ffmpeg concat failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	if err := os.Rename(tmpFile, outputPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace original file with edited version: %w", err)
	}
	return nil
}

// applySponsorblock runs the SponsorBlock mark/remove pipeline on a finished
// download and returns any warning-worthy error.
func applySponsorblock(ctx context.Context, video *youtube.Video, outputPath, baseDir string, opts Options, printer *Printer) error {
	markCats, err := parseSponsorblockCategories(opts.SponsorblockMark)
	if err != nil {
		return err
	}
	removeCats, err := parseSponsorblockCategories(opts.SponsorblockRemove)
	if err != nil {
		return err
	}
	if len(markCats) == 0 && len(removeCats) == 0 {
		return nil
	}
	durationSec := video.Duration.Seconds()

	var removeSegments []sponsorSegment
	if len(removeCats) > 0 {
		printer.Log(LogInfo, "SponsorBlock: fetching skip segments")
		segments, err := fetchSponsorblockSegments(ctx, video.ID, removeCats, opts.Timeout)
		if err != nil {
			return err
		}
		removeSegments = segments
	}
	var markSegments []sponsorSegment
	if len(markCats) > 0 {
		segments, err := fetchSponsorblockSegments(ctx, video.ID, markCats, opts.Timeout)
		if err != nil {
			return err
		}
		markSegments = segments
	}

	if len(removeSegments) > 0 {
		printer.Log(LogInfo, fmt.Sprintf("SponsorBlock: removing %d segment(s)", len(removeSegments)))
		if err := removeSponsorSegments(outputPath, baseDir, removeSegments, durationSec); err != nil {
			return err
		}
	}

	if len(markSegments) > 0 {
		printer.Log(LogInfo, fmt.Sprintf("SponsorBlock: adding %d chapter marker(s)", len(markSegments)))
		if err := embedChapters(outputPath, baseDir, segmentsToChapters(markSegments, durationSec)); err != nil {
			return err
		}
	}
	return nil
}
