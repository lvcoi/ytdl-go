package downloader

import (
	"context"
	"io"
	"net/http"

	youtube "github.com/lvcoi/ytdl-lib/v2"
)

// HTTPDoer executes raw HTTP requests. *http.Client satisfies this interface.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// YouTubeClient defines the interface for interacting with the YouTube
// data/streaming API. This decouples the downloader from the concrete
// youtube.Client type, enabling testing with mocks and a future migration
// to a forked or custom client implementation.
type YouTubeClient interface {
	GetVideoContext(ctx context.Context, url string) (*youtube.Video, error)
	GetPlaylistContext(ctx context.Context, url string) (*youtube.Playlist, error)
	GetStreamContext(ctx context.Context, video *youtube.Video, format *youtube.Format) (io.ReadCloser, int64, error)
	VideoFromPlaylistEntryContext(ctx context.Context, entry *youtube.PlaylistEntry) (*youtube.Video, error)
	// HTTP returns the underlying HTTP client for raw requests (manifests, segments).
	HTTP() HTTPDoer
	// SetChunkSize configures the chunk size for stream downloads.
	SetChunkSize(size int64)
	// SetClientInfo configures the client identity (e.g. Android, Web, iOS).
	SetClientInfo(info youtube.ClientInfo)
}

// youtubeClientAdapter wraps *youtube.Client to satisfy the YouTubeClient interface.
// It delegates all YouTube API methods to the underlying client and provides
// accessor methods for HTTPClient and ChunkSize fields.
type youtubeClientAdapter struct {
	*youtube.Client
}

func (a *youtubeClientAdapter) HTTP() HTTPDoer                      { return a.Client.HTTPClient }
func (a *youtubeClientAdapter) SetChunkSize(s int64)                { a.Client.ChunkSize = s }
func (a *youtubeClientAdapter) SetClientInfo(ci youtube.ClientInfo) { a.Client.ClientType = &ci }

// Compile-time check: *youtubeClientAdapter must implement YouTubeClient.
var _ YouTubeClient = (*youtubeClientAdapter)(nil)
