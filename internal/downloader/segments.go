package downloader

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
	"strings"
)

type HLSManifest struct {
	Variants  []HLSVariant
	Segments  []HLSSegment
	Encrypted bool
	KeyMethod string
	KeyURI    string
}

type HLSVariant struct {
	URI        string
	Bandwidth  int
	Resolution string
	Codecs     string
}

type HLSSegment struct {
	URI      string
	Duration float64
}

func ParseHLSManifest(data []byte) (HLSManifest, error) {
	if !bytes.Contains(data, []byte("#EXTM3U")) {
		return HLSManifest{}, errors.New("not an HLS manifest")
	}

	var manifest HLSManifest
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var pendingVariant *HLSVariant
	var lastDuration float64
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			attrs := parseHLSAttributes(strings.TrimPrefix(line, "#EXT-X-STREAM-INF:"))
			variant := &HLSVariant{
				Bandwidth:  parseInt(attrs["BANDWIDTH"]),
				Resolution: attrs["RESOLUTION"],
				Codecs:     attrs["CODECS"],
			}
			pendingVariant = variant
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			durationText := strings.TrimPrefix(line, "#EXTINF:")
			durationText = strings.TrimSuffix(durationText, ",")
			if duration, err := strconv.ParseFloat(durationText, 64); err == nil {
				lastDuration = duration
			}
			continue
		}

		if strings.HasPrefix(line, "#EXT-X-KEY:") {
			attrs := parseHLSAttributes(strings.TrimPrefix(line, "#EXT-X-KEY:"))
			method := strings.ToUpper(attrs["METHOD"])
			if method != "" && method != "NONE" {
				manifest.Encrypted = true
				manifest.KeyMethod = method
				manifest.KeyURI = attrs["URI"]
			}
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if pendingVariant != nil {
			pendingVariant.URI = line
			manifest.Variants = append(manifest.Variants, *pendingVariant)
			pendingVariant = nil
			continue
		}

		manifest.Segments = append(manifest.Segments, HLSSegment{
			URI:      line,
			Duration: lastDuration,
		})
		lastDuration = 0
	}

	if err := scanner.Err(); err != nil {
		return HLSManifest{}, err
	}

	return manifest, nil
}

func DetectHLSDrm(manifest HLSManifest) (bool, string) {
	if !manifest.Encrypted {
		return false, ""
	}
	if manifest.KeyMethod != "" {
		return true, manifest.KeyMethod
	}
	return true, "encrypted"
}

func DetectDASHDrm(data []byte) (bool, string) {
	lower := strings.ToLower(string(data))
	for _, marker := range dashDRMMarkers {
		if strings.Contains(lower, marker) {
			return true, marker
		}
	}
	return false, ""
}

func parseHLSAttributes(raw string) map[string]string {
	attrs := map[string]string{}
	parts := splitHLSAttributes(raw)
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToUpper(kv[0]))
		value := strings.TrimSpace(kv[1])
		value = strings.Trim(value, "\"")
		if key != "" {
			attrs[key] = value
		}
	}
	return attrs
}

func splitHLSAttributes(raw string) []string {
	var parts []string
	var b strings.Builder
	inQuotes := false
	for _, r := range raw {
		switch r {
		case '"':
			inQuotes = !inQuotes
			b.WriteRune(r)
		case ',':
			if inQuotes {
				b.WriteRune(r)
				continue
			}
			parts = append(parts, b.String())
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func parseInt(value string) int {
	if value == "" {
		return 0
	}
	num, _ := strconv.Atoi(value)
	return num
}

var dashDRMMarkers = []string{
	"widevine",
	"playready",
	"cenc",
	"urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed",
	"urn:uuid:9a04f079-9840-4286-ab92-e65be0885f95",
	"urn:mpeg:dash:mp4protection:2011",
}
