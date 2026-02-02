package downloader

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func selectFormat(video *youtube.Video, opts Options) (*youtube.Format, error) {
	// If itag is specified, search for it directly.
	// Itag takes precedence over quality, format, and audio-only options.
	if opts.Itag > 0 {
		for i := range video.Formats {
			format := &video.Formats[i]
			if format.ItagNo == opts.Itag {
				return format, nil
			}
		}
		return nil, wrapCategory(CategoryUnsupported, fmt.Errorf("itag %d not found (use --list-formats to see available itags)", opts.Itag))
	}

	candidates := make([]*youtube.Format, 0, len(video.Formats))
	for i := range video.Formats {
		format := &video.Formats[i]

		if opts.AudioOnly {
			if format.AudioChannels == 0 {
				continue
			}
			if format.Width != 0 || format.Height != 0 {
				continue
			}
		} else {
			if format.AudioChannels == 0 || format.Width == 0 || format.Height == 0 {
				continue
			}
		}

		if opts.Format != "" && !formatMatches(format, opts.Format) {
			continue
		}

		candidates = append(candidates, format)
	}

	if len(candidates) == 0 {
		// If format was specified but not found, try again without format filter (fallback)
		if opts.Format != "" {
			fallbackOpts := opts
			fallbackOpts.Format = ""
			fallbackFormat, err := selectFormat(video, fallbackOpts)
			if err == nil && fallbackFormat != nil {
				// Found a fallback format - return it (caller should warn user)
				return fallbackFormat, nil
			}
		}

		reason := "no progressive (audio+video) formats available"
		if opts.AudioOnly {
			reason = "no audio-only formats available (try without --audio or use --list-formats)"
		}
		if opts.Format != "" {
			reason = fmt.Sprintf("no formats available for requested format %s (use --list-formats)", opts.Format)
		} else if !opts.AudioOnly {
			reason = "no progressive (audio+video) formats available (try --audio or use --list-formats)"
		}
		return nil, wrapCategory(CategoryUnsupported, errors.New(reason))
	}

	if opts.AudioOnly {
		return pickAudioFormat(candidates, opts)
	}
	return pickVideoFormat(candidates, opts)
}

func pickVideoFormat(candidates []*youtube.Format, opts Options) (*youtube.Format, error) {
	targetHeight, preferLowest, err := parseVideoQuality(opts.Quality)
	if err != nil {
		return nil, wrapCategory(CategoryUnsupported, err)
	}

	// Prefer highest available when no explicit target.
	if targetHeight == 0 && !preferLowest {
		var best *youtube.Format
		for _, f := range candidates {
			if best == nil || betterVideoFormat(f, best) {
				best = f
			}
		}
		return best, nil
	}

	var best *youtube.Format
	if targetHeight > 0 {
		for _, f := range candidates {
			if f.Height == 0 || f.Height > targetHeight {
				continue
			}
			if best == nil || f.Height > best.Height || (f.Height == best.Height && bitrateForFormat(f) > bitrateForFormat(best)) {
				best = f
			}
		}
	}

	if best == nil && targetHeight > 0 {
		// No option under target: pick the closest above target.
		for _, f := range candidates {
			if best == nil || f.Height < best.Height || (f.Height == best.Height && bitrateForFormat(f) > bitrateForFormat(best)) {
				best = f
			}
		}
	}

	if best == nil && preferLowest {
		for _, f := range candidates {
			if best == nil || f.Height < best.Height || (f.Height == best.Height && bitrateForFormat(f) > bitrateForFormat(best)) {
				best = f
			}
		}
	}

	if best == nil {
		return nil, wrapCategory(CategoryUnsupported, errors.New("no matching video formats available (use --list-formats)"))
	}

	return best, nil
}

func pickAudioFormat(candidates []*youtube.Format, opts Options) (*youtube.Format, error) {
	targetBitrate, preferLowest, err := parseAudioQuality(opts.Quality)
	if err != nil {
		return nil, wrapCategory(CategoryUnsupported, err)
	}

	var best *youtube.Format
	if targetBitrate > 0 {
		for _, f := range candidates {
			br := bitrateForFormat(f)
			if br == 0 || br > targetBitrate {
				continue
			}
			if best == nil || br > bitrateForFormat(best) {
				best = f
			}
		}
		// Nothing under target? fall back to the lowest above target.
		if best == nil {
			for _, f := range candidates {
				br := bitrateForFormat(f)
				if br == 0 {
					continue
				}
				if best == nil || br < bitrateForFormat(best) {
					best = f
				}
			}
		}
	}

	if best == nil && preferLowest {
		for _, f := range candidates {
			br := bitrateForFormat(f)
			if br == 0 {
				continue
			}
			if best == nil || br < bitrateForFormat(best) {
				best = f
			}
		}
	}

	if best == nil {
		for _, f := range candidates {
			if best == nil || bitrateForFormat(f) > bitrateForFormat(best) {
				best = f
			}
		}
	}

	if best == nil {
		return nil, wrapCategory(CategoryUnsupported, errors.New("no matching audio formats available (use --list-formats)"))
	}
	return best, nil
}

func parseVideoQuality(q string) (target int, preferLowest bool, err error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" || q == "best" {
		return 0, false, nil
	}
	if q == "worst" {
		return 0, true, nil
	}
	q = strings.TrimSuffix(q, "p")
	value, convErr := strconv.Atoi(q)
	if convErr != nil {
		return 0, false, fmt.Errorf("invalid quality value %q (expected like 720p)", q)
	}
	return value, false, nil
}

func parseAudioQuality(q string) (bitrate int, preferLowest bool, err error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" || q == "best" {
		return 0, false, nil
	}
	if q == "worst" {
		return 0, true, nil
	}

	q = strings.TrimSuffix(q, "bps")
	q = strings.TrimSuffix(q, "kbps")
	hasK := strings.HasSuffix(q, "k")
	q = strings.TrimSuffix(q, "k")
	value, convErr := strconv.Atoi(q)
	if convErr != nil {
		return 0, false, fmt.Errorf("invalid audio quality %q (expected like 128k)", q)
	}
	if hasK || value < 1000 {
		value = value * 1000
	}
	return value, false, nil
}

func formatMatches(format *youtube.Format, desired string) bool {
	return strings.EqualFold(mimeToExt(format.MimeType), strings.TrimSpace(strings.ToLower(desired)))
}

func betterVideoFormat(candidate, current *youtube.Format) bool {
	if candidate.Height != current.Height {
		return candidate.Height > current.Height
	}
	return bitrateForFormat(candidate) > bitrateForFormat(current)
}
