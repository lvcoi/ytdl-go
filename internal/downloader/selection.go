package downloader

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// downloadArchive tracks downloaded video IDs in a plain text file (one ID per
// line, yt-dlp compatible). It is safe for concurrent use across jobs.
type downloadArchive struct {
	mu   sync.Mutex
	path string
	ids  map[string]struct{}
}

func newDownloadArchive(path string) (*downloadArchive, error) {
	a := &downloadArchive{path: path, ids: map[string]struct{}{}}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return a, nil
		}
		return nil, wrapCategory(CategoryFilesystem, fmt.Errorf("opening download archive: %w", err))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// yt-dlp lines look like "youtube <id>"; accept bare IDs too.
		fields := strings.Fields(line)
		a.ids[fields[len(fields)-1]] = struct{}{}
	}
	return a, nil
}

func (a *downloadArchive) contains(id string) bool {
	if a == nil || id == "" {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.ids[id]
	return ok
}

func (a *downloadArchive) record(id string) error {
	if a == nil || id == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.ids[id]; ok {
		return nil
	}
	file, err := os.OpenFile(a.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("opening download archive: %w", err))
	}
	defer file.Close()
	if _, err := fmt.Fprintf(file, "youtube %s\n", id); err != nil {
		return wrapCategory(CategoryFilesystem, fmt.Errorf("writing download archive: %w", err))
	}
	a.ids[id] = struct{}{}
	return nil
}

// playlistItemSelector filters playlist entries by index, supporting yt-dlp
// style expressions such as "1:5,8,10:12", "3", or "1,3:5" plus "~7" (negation).
type playlistItemSelector struct {
	ranges []playlistRange
}

type playlistRange struct {
	start, end int // inclusive; end == 0 means open-ended ("3:")
	negate     bool
}

// parsePlaylistItems parses a yt-dlp style --playlist-items expression.
// A "~" prefix negates a range (skip matching entries).
func parsePlaylistItems(expr string) (*playlistItemSelector, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}
	sel := &playlistItemSelector{}
	for _, part := range strings.Split(expr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rng := playlistRange{}
		if strings.HasPrefix(part, "~") {
			rng.negate = true
			part = part[1:]
		}
		if strings.Contains(part, ":") {
			bounds := strings.SplitN(part, ":", 2)
			var start int
			var err error
			if bounds[0] != "" {
				start, err = strconv.Atoi(strings.TrimSpace(bounds[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid playlist-items expression %q", expr)
				}
			} else {
				start = 1
			}
			end := 0
			if bounds[1] != "" {
				end, err = strconv.Atoi(strings.TrimSpace(bounds[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid playlist-items expression %q", expr)
				}
			}
			if start < 1 {
				return nil, fmt.Errorf("invalid playlist-items expression %q (indices are 1-based)", expr)
			}
			rng.start, rng.end = start, end
		} else {
			n, err := strconv.Atoi(part)
			if err != nil || n < 1 {
				return nil, fmt.Errorf("invalid playlist-items expression %q", expr)
			}
			rng.start, rng.end = n, n
		}
		sel.ranges = append(sel.ranges, rng)
	}
	if len(sel.ranges) == 0 {
		return nil, nil
	}
	return sel, nil
}

// includes reports whether the 1-based index is selected.
func (s *playlistItemSelector) includes(index int) bool {
	if s == nil || len(s.ranges) == 0 {
		return true
	}
	anyPositive := false
	for _, r := range s.ranges {
		if !r.negate {
			anyPositive = true
			break
		}
	}
	if !anyPositive {
		// all negations: include everything except the negated entries
		return !s.excluded(index)
	}
	return !s.excluded(index) && s.selected(index)
}

func (s *playlistItemSelector) selected(index int) bool {
	for _, r := range s.ranges {
		if r.negate {
			continue
		}
		if r.start <= index && (r.end == 0 || index <= r.end) {
			return true
		}
	}
	return false
}

func (s *playlistItemSelector) excluded(index int) bool {
	for _, r := range s.ranges {
		if !r.negate {
			continue
		}
		if r.start <= index && (r.end == 0 || index <= r.end) {
			return true
		}
	}
	return false
}

// matchFilter evaluates yt-dlp style filter expressions against a video.
// Supported: numeric comparisons with <, <=, >, >=, =, != on fields
// duration, view_count; string matching with = (exact) and *= (contains) on
// title, author, id. Conditions can be chained with & (all must match).
type matchFilter struct {
	conditions []matchCondition
}

type matchCondition struct {
	field string
	op    string
	value string
}

func parseMatchFilter(expr string) (*matchFilter, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}
	f := &matchFilter{}
	for _, part := range strings.Split(expr, "&") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cond := matchCondition{}
		// Longest operators first so "*=" wins over "=".
		for _, op := range []string{"*=", "!=", ">=", "<=", ">", "<", "="} {
			if idx := strings.Index(part, op); idx > 0 {
				cond.field = strings.TrimSpace(part[:idx])
				cond.op = op
				cond.value = strings.TrimSpace(part[idx+len(op):])
				break
			}
		}
		if cond.field == "" || cond.op == "" {
			return nil, fmt.Errorf("invalid match-filter condition %q", part)
		}
		f.conditions = append(f.conditions, cond)
	}
	if len(f.conditions) == 0 {
		return nil, nil
	}
	return f, nil
}

type filterVideoInfo struct {
	ID        string
	Title     string
	Author    string
	Duration  float64
	ViewCount int
}

// matches returns true when every condition passes (an empty filter passes).
func (f *matchFilter) matches(info filterVideoInfo) (bool, string) {
	if f == nil {
		return true, ""
	}
	for _, cond := range f.conditions {
		if !cond.matches(info) {
			return false, fmt.Sprintf("does not pass filter %s%s%s", cond.field, cond.op, cond.value)
		}
	}
	return true, ""
}

func (c *matchCondition) matches(info filterVideoInfo) bool {
	var num float64
	var str string
	isNumeric := false
	switch c.field {
	case "duration", "duration_seconds":
		num, isNumeric = info.Duration, true
	case "view_count", "views":
		num, isNumeric = float64(info.ViewCount), true
	case "title":
		str = info.Title
	case "author", "channel":
		str = info.Author
	case "id":
		str = info.ID
	default:
		// Unknown fields are ignored, mirroring yt-dlp's behavior.
		return true
	}

	switch c.op {
	case "*=":
		return strings.Contains(strings.ToLower(str), strings.ToLower(strings.Trim(c.value, `"'`)))
	case "=":
		if isNumeric {
			target, err := strconv.ParseFloat(strings.Trim(c.value, `"'`), 64)
			return err == nil && num == target
		}
		return strings.EqualFold(str, strings.Trim(c.value, `"'`))
	case "!=":
		if isNumeric {
			target, err := strconv.ParseFloat(strings.Trim(c.value, `"'`), 64)
			return err == nil && num != target
		}
		return !strings.EqualFold(str, strings.Trim(c.value, `"'`))
	}

	target, err := strconv.ParseFloat(strings.Trim(c.value, `"'`), 64)
	if err != nil {
		return true // non-numeric value for a numeric comparison: ignore condition
	}
	switch c.op {
	case "<":
		return num < target
	case "<=":
		return num <= target
	case ">":
		return num > target
	case ">=":
		return num >= target
	}
	return true
}
