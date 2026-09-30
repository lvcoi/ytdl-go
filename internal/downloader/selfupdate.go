package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// AppVersion is the current release version of ytdl-go. It is compared
// against the latest GitHub release tag by -U.
const AppVersion = "0.2.0-beta"

const updateRepoAPI = "https://api.github.com/repos/lvcoi/ytdl-go/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+`)

// normalizeVersion strips a leading "v"/"V" and any pre-release suffix so
// "v0.2.0-beta" and "0.2.0" compare numerically.
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	if m := semverRe.FindString(v); m != "" {
		return m
	}
	return v
}

// isNewerVersion reports whether candidate is newer than current.
func isNewerVersion(current, candidate string) bool {
	cur := parseSemver(normalizeVersion(current))
	cand := parseSemver(normalizeVersion(candidate))
	if isInvalidSemver(cand) {
		return false
	}
	if isInvalidSemver(cur) {
		return true
	}
	for i := 0; i < 3; i++ {
		if cand[i] != cur[i] {
			return cand[i] > cur[i]
		}
	}
	return false
}

func isInvalidSemver(v [3]int) bool {
	return v[0] < 0 || v[1] < 0 || v[2] < 0
}

func parseSemver(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{-1, -1, -1}
	}
	for i, p := range parts {
		n := 0
		if _, err := fmt.Sscanf(p, "%d", &n); err != nil {
			return [3]int{-1, -1, -1}
		}
		out[i] = n
	}
	return out
}

// fetchLatestRelease queries the GitHub API for the latest release.
func fetchLatestRelease(ctx context.Context, timeout time.Duration) (*githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateRepoAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, wrapCategory(CategoryNetwork, fmt.Errorf("checking latest release: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, wrapCategory(CategoryNetwork, fmt.Errorf("checking latest release: unexpected status %d", resp.StatusCode))
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16*1024*1024)).Decode(&release); err != nil {
		return nil, wrapCategory(CategoryNetwork, fmt.Errorf("parsing release info: %w", err))
	}
	return &release, nil
}

// updateAssetName computes the expected release asset name for this platform.
func updateAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("ytdl-go-windows-%s.exe", runtime.GOARCH)
	case "linux":
		return fmt.Sprintf("ytdl-go-linux-%s", runtime.GOARCH)
	case "darwin":
		return fmt.Sprintf("ytdl-go-darwin-%s", runtime.GOARCH)
	}
	return ""
}

// SelfUpdate checks GitHub for a newer release and, if one exists, downloads
// the platform asset and replaces the running binary. The old binary is kept
// next to the new one as "<exe>.old" until the replace succeeds.
func SelfUpdate(ctx context.Context, timeout time.Duration) (string, error) {
	release, err := fetchLatestRelease(ctx, timeout)
	if err != nil {
		return "", err
	}
	if !isNewerVersion(AppVersion, release.TagName) {
		return fmt.Sprintf("ytdl-go %s is up to date (latest release: %s)", AppVersion, release.TagName), nil
	}

	assetName := updateAssetName()
	if assetName == "" {
		return "", wrapCategory(CategoryUnsupported, fmt.Errorf("self-update is not supported on %s/%s", runtime.GOOS, runtime.GOARCH))
	}
	var assetURL string
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, assetName) {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return "", wrapCategory(CategoryUnsupported, fmt.Errorf("no release asset %q found; download it manually from https://github.com/lvcoi/ytdl-go/releases", assetName))
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("resolving executable path: %w", err))
	}
	exeReal, err := filepath.EvalSymlinks(exePath)
	if err == nil {
		exePath = exeReal
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(exePath), ".ytdl-update-*.tmp")
	if err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("creating update temp file: %w", err))
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading update: %w", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading update: unexpected status %d", resp.StatusCode))
	}
	if _, err := io.Copy(tmpFile, io.LimitReader(resp.Body, 1<<30)); err != nil {
		tmpFile.Close()
		return "", wrapCategory(CategoryNetwork, fmt.Errorf("downloading update: %w", err))
	}
	if err := tmpFile.Close(); err != nil {
		return "", wrapCategory(CategoryFilesystem, err)
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return "", wrapCategory(CategoryFilesystem, err)
	}

	// Replace via rename: rename old out of the way, move new in, restore on failure.
	backup := exePath + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(exePath, backup); err != nil {
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("backing up current binary: %w", err))
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		_ = os.Rename(backup, exePath) // restore
		return "", wrapCategory(CategoryFilesystem, fmt.Errorf("installing update: %w", err))
	}
	_ = os.Remove(backup)

	return fmt.Sprintf("updated to %s (was %s). Restart to use the new version.", release.TagName, AppVersion), nil
}
