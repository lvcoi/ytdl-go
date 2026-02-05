package downloader

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type mpd struct {
	XMLName xml.Name `xml:"MPD"`
	BaseURL []string `xml:"BaseURL"`
	Period  []period `xml:"Period"`
}

type period struct {
	AdaptationSet []adaptationSet `xml:"AdaptationSet"`
}

type adaptationSet struct {
	MimeType       string           `xml:"mimeType,attr"`
	ContentType    string           `xml:"contentType,attr"`
	Lang           string           `xml:"lang,attr"`
	BaseURL        []string         `xml:"BaseURL"`
	SegmentList    *segmentList     `xml:"SegmentList"`
	Representation []representation `xml:"Representation"`
}

type representation struct {
	ID          string       `xml:"id,attr"`
	Bandwidth   int          `xml:"bandwidth,attr"`
	Width       int          `xml:"width,attr"`
	Height      int          `xml:"height,attr"`
	MimeType    string       `xml:"mimeType,attr"`
	BaseURL     []string     `xml:"BaseURL"`
	SegmentList *segmentList `xml:"SegmentList"`
}

type segmentList struct {
	Initialization *segmentURL  `xml:"Initialization"`
	SegmentURL     []segmentURL `xml:"SegmentURL"`
}

type segmentURL struct {
	Media     string `xml:"media,attr"`
	SourceURL string `xml:"sourceURL,attr"`
	Range     string `xml:"mediaRange,attr"`
}

type dashRepresentation struct {
	Rep        representation
	Adaptation adaptationSet
	BaseURL    string
	Segments   []string
	InitURL    string
}

func downloadDASH(ctx context.Context, client *youtube.Client, video *youtube.Video, opts Options, ctxInfo outputContext, printer *Printer, prefix string) (downloadResult, error) {
	data, err := fetchManifest(ctx, client, video.DASHManifestURL)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("fetching DASH manifest: %w", err))
	}

	if ok, marker := DetectDASHDrm(data); ok {
		message := "encrypted DASH manifest"
		if marker != "" {
			message = fmt.Sprintf("encrypted DASH manifest (%s)", marker)
		}
		return downloadResult{}, wrapCategory(CategoryRestricted, fmt.Errorf("%s", message))
	}

	var manifest mpd
	if err := xml.Unmarshal(data, &manifest); err != nil {
		return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("parsing DASH manifest: %w", err))
	}

	representations := collectDASHRepresentations(manifest, video.DASHManifestURL)
	if len(representations) == 0 {
		return downloadResult{}, wrapCategory(CategoryUnsupported, fmt.Errorf("no DASH representations with segment lists"))
	}

	selected, err := selectDASHRepresentation(representations, opts)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryUnsupported, err)
	}

	format := dashFormatFromRepresentation(selected, opts.Quality)
	outputPath, err := resolveOutputPath(opts.OutputTemplate, video, format, ctxInfo, opts.OutputDir)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, err)
	}
	outputPath, skip, err := handleExistingPath(outputPath, opts.OutputDir, opts, printer)
	if err != nil {
		return downloadResult{}, err
	}
	if skip {
		return downloadResult{skipped: true, outputPath: outputPath}, nil
	}

	result, err := downloadDASHSegments(ctx, client, selected, outputPath, opts.OutputDir, opts, printer, prefix)
	if err == nil {
		result.outputPath = outputPath
	}
	return result, err
}

func collectDASHRepresentations(manifest mpd, manifestURL string) []dashRepresentation {
	base := firstString(manifest.BaseURL)
	if base == "" {
		base = manifestURL
	}
	var reps []dashRepresentation
	for _, p := range manifest.Period {
		for _, adapt := range p.AdaptationSet {
			adaptBase := resolveManifestURL(base, firstString(adapt.BaseURL))
			for _, rep := range adapt.Representation {
				repBase := resolveManifestURL(adaptBase, firstString(rep.BaseURL))
				segList := rep.SegmentList
				if segList == nil {
					segList = adapt.SegmentList
				}
				if segList == nil || len(segList.SegmentURL) == 0 {
					continue
				}
				segments := []string{}
				for _, seg := range segList.SegmentURL {
					url := seg.Media
					if url == "" {
						url = seg.SourceURL
					}
					if url == "" {
						continue
					}
					segments = append(segments, resolveManifestURL(repBase, url))
				}
				if len(segments) == 0 {
					continue
				}
				initURL := ""
				if segList.Initialization != nil {
					initURL = segList.Initialization.SourceURL
				}
				if initURL != "" {
					initURL = resolveManifestURL(repBase, initURL)
				}
				reps = append(reps, dashRepresentation{
					Rep:        rep,
					Adaptation: adapt,
					BaseURL:    repBase,
					Segments:   segments,
					InitURL:    initURL,
				})
			}
		}
	}
	return reps
}

func selectDASHRepresentation(reps []dashRepresentation, opts Options) (dashRepresentation, error) {
	if opts.AudioOnly {
		audio := filterDASHReps(reps, true)
		if len(audio) == 0 {
			return dashRepresentation{}, fmt.Errorf("no audio DASH representations available (use --list-formats)")
		}
		return pickDASHAudio(audio, opts.Quality), nil
	}

	video := filterDASHReps(reps, false)
	if len(video) == 0 {
		video = reps
	}
	return pickDASHVideo(video, opts.Quality), nil
}

func filterDASHReps(reps []dashRepresentation, audio bool) []dashRepresentation {
	filtered := []dashRepresentation{}
	for _, rep := range reps {
		mime := strings.ToLower(rep.Rep.MimeType)
		if mime == "" {
			mime = strings.ToLower(rep.Adaptation.MimeType)
		}
		content := strings.ToLower(rep.Adaptation.ContentType)
		if audio {
			if strings.Contains(mime, "audio") || content == "audio" {
				filtered = append(filtered, rep)
			}
		} else {
			if strings.Contains(mime, "video") || content == "video" {
				filtered = append(filtered, rep)
			}
		}
	}
	return filtered
}

func pickDASHVideo(reps []dashRepresentation, quality string) dashRepresentation {
	targetHeight, preferLowest, err := parseVideoQuality(quality)
	if err != nil {
		targetHeight = 0
		preferLowest = false
	}
	best := reps[0]
	if targetHeight == 0 && !preferLowest {
		for _, rep := range reps[1:] {
			if rep.Rep.Height > best.Rep.Height || (rep.Rep.Height == best.Rep.Height && rep.Rep.Bandwidth > best.Rep.Bandwidth) {
				best = rep
			}
		}
		return best
	}

	if targetHeight > 0 {
		found := false
		for _, rep := range reps {
			if rep.Rep.Height == 0 || rep.Rep.Height > targetHeight {
				continue
			}
			if !found || rep.Rep.Height > best.Rep.Height || (rep.Rep.Height == best.Rep.Height && rep.Rep.Bandwidth > best.Rep.Bandwidth) {
				best = rep
				found = true
			}
		}
		if found {
			return best
		}
	}

	if preferLowest {
		for _, rep := range reps[1:] {
			if rep.Rep.Height < best.Rep.Height || (rep.Rep.Height == best.Rep.Height && rep.Rep.Bandwidth > best.Rep.Bandwidth) {
				best = rep
			}
		}
		return best
	}

	for _, rep := range reps[1:] {
		if rep.Rep.Height < best.Rep.Height || (rep.Rep.Height == best.Rep.Height && rep.Rep.Bandwidth > best.Rep.Bandwidth) {
			best = rep
		}
	}
	return best
}

func pickDASHAudio(reps []dashRepresentation, quality string) dashRepresentation {
	targetBitrate, preferLowest, err := parseAudioQuality(quality)
	if err != nil {
		targetBitrate = 0
		preferLowest = false
	}
	best := reps[0]
	if targetBitrate == 0 && !preferLowest {
		for _, rep := range reps[1:] {
			if rep.Rep.Bandwidth > best.Rep.Bandwidth {
				best = rep
			}
		}
		return best
	}
	if targetBitrate > 0 {
		found := false
		for _, rep := range reps {
			if rep.Rep.Bandwidth == 0 || rep.Rep.Bandwidth > targetBitrate {
				continue
			}
			if !found || rep.Rep.Bandwidth > best.Rep.Bandwidth {
				best = rep
				found = true
			}
		}
		if found {
			return best
		}
	}
	if preferLowest {
		for _, rep := range reps[1:] {
			if rep.Rep.Bandwidth < best.Rep.Bandwidth {
				best = rep
			}
		}
		return best
	}
	for _, rep := range reps[1:] {
		if rep.Rep.Bandwidth < best.Rep.Bandwidth {
			best = rep
		}
	}
	return best
}

func dashFormatFromRepresentation(rep dashRepresentation, quality string) *youtube.Format {
	mime := rep.Rep.MimeType
	if mime == "" {
		mime = rep.Adaptation.MimeType
	}
	ext := strings.TrimPrefix(filepath.Ext(rep.Segments[0]), ".")
	if ext != "" {
		mime = "video/" + ext
	}
	return &youtube.Format{
		MimeType:     mime,
		QualityLabel: quality,
		Bitrate:      rep.Rep.Bandwidth,
		Width:        rep.Rep.Width,
		Height:       rep.Rep.Height,
	}
}

type dashResumeState struct {
	ManifestURL  string `json:"manifest_url"`
	SegmentCount int    `json:"segment_count"`
	NextIndex    int    `json:"next_index"`
	BytesWritten int64  `json:"bytes_written"`
	InitDone     bool   `json:"init_done"`
}

func downloadDASHSegments(ctx context.Context, client *youtube.Client, rep dashRepresentation, outputPath, baseDir string, opts Options, printer *Printer, prefix string) (downloadResult, error) {
	partPath, err := artifactPath(outputPath, partSuffix, baseDir)
	if err != nil {
		return downloadResult{}, err
	}
	resumePath, err := artifactPath(outputPath, resumeSuffix, baseDir)
	if err != nil {
		return downloadResult{}, err
	}
	segmentDir, err := artifactPath(outputPath, ".segments", baseDir)
	if err != nil {
		return downloadResult{}, err
	}

	state := dashResumeState{
		ManifestURL:  rep.BaseURL,
		SegmentCount: len(rep.Segments),
		NextIndex:    0,
		BytesWritten: 0,
		InitDone:     false,
	}
	if loaded, err := loadDASHResume(resumePath); err == nil && loaded.SegmentCount == len(rep.Segments) {
		state = loaded
	}

	useParallel := opts.SegmentConcurrency != 1 && state.NextIndex == 0 && state.BytesWritten == 0 && !state.InitDone
	if useParallel {
		tempDir, err := validateSegmentTempDir(segmentDir, baseDir)
		if err != nil {
			return downloadResult{}, err
		}
		file, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("opening temp file: %w", err))
		}
		defer file.Close()

		if rep.InitURL != "" {
			if err := downloadSegmentWithRetry(ctx, client, rep.InitURL, file); err != nil {
				return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("initialization segment failed: %w", err))
			}
		}
		plan := segmentDownloadPlan{
			URLs:        rep.Segments,
			TempDir:     tempDir,
			Prefix:      prefix,
			Concurrency: opts.SegmentConcurrency,
			BaseDir:     baseDir,
		}
		total, err := downloadSegmentsParallel(ctx, client, plan, file, printer)
		if err != nil {
			return downloadResult{}, err
		}
		if err := file.Close(); err != nil {
			return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("closing temp file: %w", err))
		}
		if err := os.Rename(partPath, outputPath); err != nil {
			return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("renaming output: %w", err))
		}
		if err := validateOutputFile(outputPath, nil); err != nil {
			return downloadResult{}, err
		}
		_ = os.RemoveAll(segmentDir)
		return downloadResult{bytes: total, outputPath: outputPath}, nil
	}

	file, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("opening temp file: %w", err))
	}
	defer file.Close()

	var writer io.Writer = file
	var progress *progressWriter
	if !opts.Quiet {
		progress = newProgressWriter(0, printer, prefix)
		progress.total.Store(state.BytesWritten)
		writer = io.MultiWriter(file, progress)
	}

	if !state.InitDone && rep.InitURL != "" {
		if err := downloadSegmentWithRetry(ctx, client, rep.InitURL, writer); err != nil {
			return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("initialization segment failed: %w", err))
		}
		state.InitDone = true
		if progress != nil {
			state.BytesWritten = progress.total.Load()
		}
		if err := saveDASHResume(resumePath, state); err != nil {
			return downloadResult{}, err
		}
	}

	for idx := state.NextIndex; idx < len(rep.Segments); idx++ {
		if err := downloadSegmentWithRetry(ctx, client, rep.Segments[idx], writer); err != nil {
			if progress != nil {
				progress.NewLine()
			}
			return downloadResult{}, wrapCategory(CategoryNetwork, fmt.Errorf("segment %d failed: %w", idx+1, err))
		}
		state.NextIndex = idx + 1
		if progress != nil {
			state.BytesWritten = progress.total.Load()
		}
		if err := saveDASHResume(resumePath, state); err != nil {
			return downloadResult{}, err
		}
	}

	if progress != nil {
		progress.Finish()
	}

	if err := file.Close(); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("closing temp file: %w", err))
	}
	if err := os.Rename(partPath, outputPath); err != nil {
		return downloadResult{}, wrapCategory(CategoryFilesystem, fmt.Errorf("renaming output: %w", err))
	}
	if err := validateOutputFile(outputPath, nil); err != nil {
		return downloadResult{}, err
	}
	_ = os.Remove(resumePath)

	return downloadResult{bytes: state.BytesWritten, outputPath: outputPath}, nil
}

func loadDASHResume(path string) (dashResumeState, error) {
	file, err := os.Open(path)
	if err != nil {
		return dashResumeState{}, err
	}
	defer file.Close()
	var state dashResumeState
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return dashResumeState{}, err
	}
	return state, nil
}

func saveDASHResume(path string, state dashResumeState) error {
	file, err := os.Create(path)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("saving resume state: %w", err))
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(state); err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("saving resume state: %w", err))
	}
	return nil
}

func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}
