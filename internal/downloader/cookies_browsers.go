package downloader

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// loadChromiumCookies reads cookies from a Chromium based browser profile
// (Chrome, Edge, Brave). Values encrypted by the browser are decrypted with
// the OS level key found in Local State.
func loadChromiumCookies(browser, profile string) ([]cookieEntry, error) {
	userDataDir := chromiumBaseDir(browser)
	if userDataDir == "" {
		return nil, fmt.Errorf("browser %q is not supported on this platform", browser)
	}
	if profile == "" {
		profile = "Default"
	}
	profileDir := filepath.Join(userDataDir, profile)
	if _, err := os.Stat(profileDir); err != nil {
		return nil, fmt.Errorf("profile directory not found: %s", profileDir)
	}

	crypt := chromiumCryptResult{}
	if key, err := localStateKey(userDataDir); err == nil {
		crypt.key = key
	}

	cookieDB := filepath.Join(profileDir, "Network", "Cookies")
	if _, err := os.Stat(cookieDB); err != nil {
		cookieDB = filepath.Join(profileDir, "Cookies")
		if _, err := os.Stat(cookieDB); err != nil {
			return nil, fmt.Errorf("cookies database not found in %s", profileDir)
		}
	}

	tmpDB, cleanup, err := copySQLiteForRead(cookieDB)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	db, err := sql.Open("sqlite", tmpDB)
	if err != nil {
		return nil, fmt.Errorf("opening copied cookie database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT host_key, name, value, encrypted_value, path, is_secure, expires_utc FROM cookies`)
	if err != nil {
		return nil, fmt.Errorf("querying cookies: %w", err)
	}
	defer rows.Close()

	var entries []cookieEntry
	for rows.Next() {
		var host, name, value, encValue, path string
		var secure int
		var expiresUTC int64 // microseconds since 1601-01-01
		if err := rows.Scan(&host, &name, &value, &encValue, &path, &secure, &expiresUTC); err != nil {
			continue
		}
		entry := cookieEntry{
			Domain:            host,
			IncludeSubdomains: strings.HasPrefix(host, "."),
			Path:              path,
			Secure:            secure != 0,
			Name:              name,
		}
		if expiresUTC > 0 {
			entry.Expires = time.Unix(expiresUTC/1_000_000-chromiumEpochOffset, 0)
		}
		if value == "" && encValue != "" {
			plain, err := decryptChromiumCookie(crypt, []byte(encValue))
			if err != nil {
				return nil, err
			}
			value = string(plain)
		}
		entry.Value = value
		if entry.Name == "" {
			continue
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no cookies found for browser %q (is the browser installed and signed in?)", browser)
	}
	return entries, nil
}

const chromiumEpochOffset = 11644473600 // seconds between 1601-01-01 and 1970-01-01

// firefoxProfileDir resolves the Firefox profile directory on any platform.
func firefoxProfileDir(profile string) (string, error) {
	var base string
	switch {
	case os.Getenv("APPDATA") != "":
		base = filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")
	case os.Getenv("HOME") != "":
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")); err == nil {
			base = filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox")
		} else {
			base = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles")
		}
	default:
		return "", fmt.Errorf("cannot locate Firefox profile directory")
	}
	dirs, err := os.ReadDir(base)
	if err != nil {
		return "", fmt.Errorf("reading Firefox profiles directory: %w", err)
	}
	if profile != "" {
		for _, d := range dirs {
			if d.IsDir() && strings.EqualFold(d.Name(), profile) {
				return filepath.Join(base, d.Name()), nil
			}
		}
		return "", fmt.Errorf("Firefox profile %q not found", profile)
	}
	// Prefer the default-release profile, then any *.default*, then any dir.
	var fallback string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".default-release") {
			return filepath.Join(base, d.Name()), nil
		}
		if strings.Contains(name, ".default") && fallback == "" {
			fallback = d.Name()
		}
	}
	if fallback != "" {
		return filepath.Join(base, fallback), nil
	}
	for _, d := range dirs {
		if d.IsDir() {
			if _, err := os.Stat(filepath.Join(base, d.Name(), "cookies.sqlite")); err == nil {
				return filepath.Join(base, d.Name()), nil
			}
		}
	}
	return "", fmt.Errorf("no Firefox profile with cookies.sqlite found in %s", base)
}

// loadFirefoxCookies reads cookies from a Firefox profile. Firefox stores
// cookie values in plaintext, so no OS key is required.
func loadFirefoxCookies(browser, profile string) ([]cookieEntry, error) {
	profileDir, err := firefoxProfileDir(profile)
	if err != nil {
		return nil, err
	}
	tmpDB, cleanup, err := copySQLiteForRead(filepath.Join(profileDir, "cookies.sqlite"))
	if err != nil {
		return nil, err
	}
	defer cleanup()

	db, err := sql.Open("sqlite", tmpDB)
	if err != nil {
		return nil, fmt.Errorf("opening copied cookie database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT host, name, value, path, isSecure, expiry FROM moz_cookies`)
	if err != nil {
		return nil, fmt.Errorf("querying cookies: %w", err)
	}
	defer rows.Close()

	var entries []cookieEntry
	for rows.Next() {
		var host, name, value, path string
		var secure int
		var expiry int64
		if err := rows.Scan(&host, &name, &value, &path, &secure, &expiry); err != nil {
			continue
		}
		entry := cookieEntry{
			Domain:            host,
			IncludeSubdomains: strings.HasPrefix(host, "."),
			Path:              path,
			Secure:            secure != 0,
			Name:              name,
			Value:             value,
		}
		if expiry > 0 {
			entry.Expires = time.Unix(expiry, 0)
		}
		if entry.Name == "" {
			continue
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no cookies found for browser %q (is the browser installed and signed in?)", browser)
	}
	return entries, nil
}
