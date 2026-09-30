package downloader

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// cookieEntry is a fully resolved cookie ready to be injected into requests.
type cookieEntry struct {
	Domain   string `json:"domain"` // e.g. ".youtube.com"
	IncludeSubdomains bool
	Path     string
	Secure   bool
	Expires  time.Time // zero = session cookie
	Name     string
	Value    string
}

func (c cookieEntry) matchesURL(u *url.URL) bool {
	host := strings.ToLower(u.Hostname())
	domain := strings.ToLower(strings.TrimPrefix(c.Domain, "."))
	if domain == "" {
		return false
	}
	if c.IncludeSubdomains || strings.HasPrefix(c.Domain, ".") {
		if host != domain && !strings.HasSuffix(host, "."+domain) {
			return false
		}
	} else if host != domain {
		return false
	}
	if c.Path != "" && !strings.HasPrefix(u.Path, c.Path) {
		return false
	}
	if c.Secure && u.Scheme != "https" {
		return false
	}
	if !c.Expires.IsZero() && c.Expires.Before(time.Now()) {
		return false
	}
	return true
}

// staticCookieJar implements http.CookieJar on top of a fixed cookie set.
// It delegates storage of dynamic cookies to an optional primary jar.
type staticCookieJar struct {
	primary http.CookieJar
	static  []cookieEntry
}

func (j *staticCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if j.primary != nil {
		j.primary.SetCookies(u, cookies)
	}
}

func (j *staticCookieJar) Cookies(u *url.URL) []*http.Cookie {
	var out []*http.Cookie
	if j.primary != nil {
		out = j.primary.Cookies(u)
	}
	for _, entry := range j.static {
		if entry.matchesURL(u) {
			out = append(out, &http.Cookie{Name: entry.Name, Value: entry.Value})
		}
	}
	return out
}

// loadStaticCookies resolves cookies from --cookies-file and/or
// --cookies-from-browser into a single sorted list.
func loadStaticCookies(opts Options) ([]cookieEntry, error) {
	var entries []cookieEntry
	if strings.TrimSpace(opts.CookiesFile) != "" {
		parsed, err := parseNetscapeCookiesFile(opts.CookiesFile)
		if err != nil {
			return nil, wrapCategory(CategoryFilesystem, err)
		}
		entries = append(entries, parsed...)
	}
	if browserSpec := strings.TrimSpace(opts.CookiesFromBrowser); browserSpec != "" {
		parsed, err := loadBrowserCookies(browserSpec)
		if err != nil {
			return nil, wrapCategory(CategoryRestricted, err)
		}
		entries = append(entries, parsed...)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Domain != entries[j].Domain {
			return entries[i].Domain < entries[j].Domain
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

// parseNetscapeCookiesFile reads a Netscape format cookies.txt file.
func parseNetscapeCookiesFile(path string) ([]cookieEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cookies file: %w", err)
	}
	var entries []cookieEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "# ") || line == "# Netscape HTTP Cookie File" {
			// "#HttpOnly_" prefixed lines are real cookies, keep them.
			continue
		}
		if strings.HasPrefix(line, "#HttpOnly_") {
			line = strings.TrimPrefix(line, "#HttpOnly_")
		} else if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 7 {
			continue
		}
		entry := cookieEntry{
			Domain:            fields[0],
			IncludeSubdomains: strings.EqualFold(fields[1], "TRUE"),
			Path:              fields[2],
			Secure:            strings.EqualFold(fields[3], "TRUE"),
			Name:              fields[5],
			Value:             fields[6],
		}
		if expiry, err := strconv.ParseInt(strings.TrimSpace(fields[4]), 10, 64); err == nil && expiry > 0 {
			entry.Expires = time.Unix(expiry, 0)
		}
		if entry.Name == "" {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// loadBrowserCookies extracts cookies from a local browser profile.
// The spec accepts "chrome", "edge", "brave", "firefox" with an optional
// ":profile" suffix (e.g. "chrome:Profile 2").
func loadBrowserCookies(spec string) ([]cookieEntry, error) {
	parts := strings.SplitN(spec, ":", 2)
	browser := strings.ToLower(strings.TrimSpace(parts[0]))
	profile := ""
	if len(parts) == 2 {
		profile = strings.TrimSpace(parts[1])
	}
	switch browser {
	case "chrome", "edge", "brave":
		return loadChromiumCookies(browser, profile)
	case "firefox", "waterfox", "librewolf":
		return loadFirefoxCookies(browser, profile)
	default:
		return nil, fmt.Errorf("unsupported browser %q (supported: chrome, edge, brave, firefox)", browser)
	}
}

func chromiumUserDataURL(browser, profile string) (string, error) {
	base := chromiumBaseDir(browser)
	if base == "" {
		return "", fmt.Errorf("browser %q is not supported on this platform", browser)
	}
	if profile == "" {
		profile = "Default"
	}
	return filepath.Join(base, profile), nil
}

func chromiumBaseDir(browser string) string {
	switch browser {
	case "chrome":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data")
	case "edge":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data")
	case "brave":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "BraveSoftware", "Brave-Browser", "User Data")
	}
	return ""
}

// chromiumCryptResult carries the resolved OS-level key material.
type chromiumCryptResult struct {
	key []byte
}

// decryptChromiumCookie decrypts a v10/v11 encrypted cookie value. v20
// (Chrome 127+ app-bound encryption) is not supported and returns a clear error.
func decryptChromiumCookie(crypt chromiumCryptResult, value []byte) ([]byte, error) {
	if len(value) < 3 {
		return value, nil
	}
	version := string(value[:3])
	payload := value[3:]
	switch version {
	case "v10", "v11":
		if crypt.key == nil {
			return nil, fmt.Errorf("no decryption key available (failed to read Local State)")
		}
		return aesGCMDecryptImpl(crypt.key, payload)
	case "v20":
		return nil, fmt.Errorf("cookie uses Chrome 127+ app-bound encryption (v20) which is not supported; export cookies to a Netscape file instead")
	default:
		// Older unencrypted or unknown format: return raw value.
		return value, nil
	}
}

// aesGCMDecryptImpl decrypts Chromium v10/v11 cookie payloads: a 12-byte
// nonce followed by ciphertext with a 16-byte GCM tag.
func aesGCMDecryptImpl(key, payload []byte) ([]byte, error) {
	if len(payload) < 12+16 {
		return nil, fmt.Errorf("encrypted cookie payload too short (%d bytes)", len(payload))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}
	plain, err := gcm.Open(nil, payload[:12], payload[12:], nil)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decrypt: %w", err)
	}
	return plain, nil
}

// copySQLiteForRead copies a SQLite database (plus WAL/SHM sidecars) to a
// temp file so it can be read while the browser is running.
func copySQLiteForRead(dbPath string) (string, func(), error) {
	tmp, err := os.CreateTemp("", "ytdl-cookies-*.sqlite")
	if err != nil {
		return "", nil, wrapCategory(CategoryFilesystem, err)
	}
	cleanup := func() {
		_ = os.Remove(tmp.Name())
		_ = os.Remove(tmp.Name() + "-wal")
		_ = os.Remove(tmp.Name() + "-shm")
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	if err := copyFileContents(dbPath, tmp.Name()); err != nil {
		cleanup()
		return "", nil, wrapCategory(CategoryFilesystem, fmt.Errorf("copying %s: %w", dbPath, err))
	}
	_ = copyFileContents(dbPath+"-wal", tmp.Name()+"-wal")
	_ = copyFileContents(dbPath+"-shm", tmp.Name()+"-shm")
	return tmp.Name(), cleanup, nil
}

func copyFileContents(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

// localStateKey reads and OS-decrypts the AES key from Chromium's Local State.
func localStateKey(userDataDir string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(userDataDir, "Local State"))
	if err != nil {
		return nil, fmt.Errorf("reading Local State: %w", err)
	}
	var state struct {
		OsCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing Local State: %w", err)
	}
	if state.OsCrypt.EncryptedKey == "" {
		return nil, fmt.Errorf("Local State has no encrypted_key")
	}
	blob, err := base64.StdEncoding.DecodeString(state.OsCrypt.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("decoding encrypted_key: %w", err)
	}
	if len(blob) > 5 && string(blob[:5]) == "DPAPI" {
		blob = blob[5:]
	}
	key, err := osUnprotectData(blob)
	if err != nil {
		return nil, fmt.Errorf("decrypting cookie key (DPAPI): %w", err)
	}
	return key, nil
}

// unpadPKCS7 removes PKCS#7 padding after CBC decryption.
func unpadPKCS7(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > len(data) {
		return data
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return data
		}
	}
	return data[:len(data)-pad]
}
