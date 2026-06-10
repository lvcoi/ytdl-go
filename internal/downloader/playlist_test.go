package downloader

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestResolveMusicPlaylistAlbumMetaSkipsWhenNotMusicURL(t *testing.T) {
	called := false
	restore := fetchMusicPlaylistEntriesFn
	fetchMusicPlaylistEntriesFn = func(context.Context, string, Options) (map[string]musicEntryMeta, error) {
		called = true
		return nil, errors.New("should not be called")
	}
	defer func() { fetchMusicPlaylistEntriesFn = restore }()

	out := resolveMusicPlaylistAlbumMeta(context.Background(), "PL123", Options{Timeout: time.Second}, false, nil)
	if called {
		t.Fatalf("fetchMusicPlaylistEntriesFn should not be called when isMusicURL=false")
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map for non-music playlist, got %d entries", len(out))
	}
}

func TestResolveMusicPlaylistAlbumMetaFallsBackOnFetchError(t *testing.T) {
	restore := fetchMusicPlaylistEntriesFn
	fetchMusicPlaylistEntriesFn = func(context.Context, string, Options) (map[string]musicEntryMeta, error) {
		return nil, errors.New("network down")
	}
	defer func() { fetchMusicPlaylistEntriesFn = restore }()

	out := resolveMusicPlaylistAlbumMeta(context.Background(), "PL123", Options{Timeout: time.Second}, true, nil)
	if len(out) != 0 {
		t.Fatalf("expected empty map on fetch error, got %d entries", len(out))
	}
}

func TestResolveMusicPlaylistAlbumMetaReturnsFetchedData(t *testing.T) {
	restore := fetchMusicPlaylistEntriesFn
	expected := map[string]musicEntryMeta{
		"abc123": {
			Title:  "Track A",
			Artist: "Artist A",
			Album:  "Album A",
		},
	}
	fetchMusicPlaylistEntriesFn = func(context.Context, string, Options) (map[string]musicEntryMeta, error) {
		return expected, nil
	}
	defer func() { fetchMusicPlaylistEntriesFn = restore }()

	out := resolveMusicPlaylistAlbumMeta(context.Background(), "PL123", Options{Timeout: time.Second}, true, nil)
	if len(out) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(out))
	}
	if out["abc123"].Album != "Album A" {
		t.Fatalf("expected album metadata passthrough, got %+v", out["abc123"])
	}
}
