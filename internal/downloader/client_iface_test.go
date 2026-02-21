package downloader

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	youtube "github.com/lvcoi/ytdl-lib/v2"
)

// mockYouTubeClient is a test double that satisfies YouTubeClient.
type mockYouTubeClient struct {
	getVideoFn    func(ctx context.Context, url string) (*youtube.Video, error)
	getPlaylistFn func(ctx context.Context, url string) (*youtube.Playlist, error)
	getStreamFn   func(ctx context.Context, video *youtube.Video, format *youtube.Format) (io.ReadCloser, int64, error)
	videoFromFn   func(ctx context.Context, entry *youtube.PlaylistEntry) (*youtube.Video, error)
	httpDoer      HTTPDoer
	chunkSize     int64
}

func (m *mockYouTubeClient) GetVideoContext(ctx context.Context, url string) (*youtube.Video, error) {
	if m.getVideoFn != nil {
		return m.getVideoFn(ctx, url)
	}
	return &youtube.Video{}, nil
}

func (m *mockYouTubeClient) GetPlaylistContext(ctx context.Context, url string) (*youtube.Playlist, error) {
	if m.getPlaylistFn != nil {
		return m.getPlaylistFn(ctx, url)
	}
	return &youtube.Playlist{}, nil
}

func (m *mockYouTubeClient) GetStreamContext(ctx context.Context, video *youtube.Video, format *youtube.Format) (io.ReadCloser, int64, error) {
	if m.getStreamFn != nil {
		return m.getStreamFn(ctx, video, format)
	}
	return io.NopCloser(&io.LimitedReader{}), 0, nil
}

func (m *mockYouTubeClient) VideoFromPlaylistEntryContext(ctx context.Context, entry *youtube.PlaylistEntry) (*youtube.Video, error) {
	if m.videoFromFn != nil {
		return m.videoFromFn(ctx, entry)
	}
	return &youtube.Video{}, nil
}

func (m *mockYouTubeClient) HTTP() HTTPDoer                        { return m.httpDoer }
func (m *mockYouTubeClient) SetChunkSize(s int64)                  { m.chunkSize = s }
func (m *mockYouTubeClient) SetClientInfo(info youtube.ClientInfo) {}

// Compile-time checks.
var _ YouTubeClient = (*youtubeClientAdapter)(nil)
var _ YouTubeClient = (*mockYouTubeClient)(nil)

func TestYouTubeClientAdapter_HTTP(t *testing.T) {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	adapter := &youtubeClientAdapter{&youtube.Client{HTTPClient: httpClient}}

	doer := adapter.HTTP()
	if doer == nil {
		t.Fatal("HTTP() returned nil")
	}
	if doer != httpClient {
		t.Fatal("HTTP() did not return the underlying HTTPClient")
	}
}

func TestYouTubeClientAdapter_SetChunkSize(t *testing.T) {
	adapter := &youtubeClientAdapter{&youtube.Client{}}

	adapter.SetChunkSize(1024 * 1024)
	if adapter.Client.ChunkSize != 1024*1024 {
		t.Fatalf("expected ChunkSize 1048576, got %d", adapter.Client.ChunkSize)
	}
}

func TestMockYouTubeClient_SatisfiesInterface(t *testing.T) {
	mock := &mockYouTubeClient{
		httpDoer: &http.Client{},
	}

	// Use the mock through the interface to verify it works.
	var client YouTubeClient = mock

	if client.HTTP() == nil {
		t.Fatal("HTTP() returned nil")
	}

	client.SetChunkSize(512)
	if mock.chunkSize != 512 {
		t.Fatalf("expected chunkSize 512, got %d", mock.chunkSize)
	}

	video, err := client.GetVideoContext(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video == nil {
		t.Fatal("expected non-nil video")
	}
}
