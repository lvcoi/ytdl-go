package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// MediaRecord represents a row in the media table.
type MediaRecord struct {
	ID              int64
	Title           string
	Artist          string
	Album           string
	Duration        int
	MediaType       string
	FilePath        string
	SourceURL       string
	ThumbnailURL    string
	Format          string
	Quality         string
	FileSize        int64
	VideoID         string
	TagsEmbedded    bool
	TagError        string
	TrackNumber     int
	PlaylistID      string
	PlaylistTitle   string
	PlaylistIndex   int
	CreatedAt       time.Time
}

const createTableSQL = `
CREATE TABLE IF NOT EXISTS media (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT NOT NULL DEFAULT '',
    artist          TEXT NOT NULL DEFAULT '',
    album           TEXT NOT NULL DEFAULT '',
    duration        INTEGER NOT NULL DEFAULT 0,
    media_type      TEXT NOT NULL DEFAULT 'video',
    file_path       TEXT NOT NULL UNIQUE,
    source_url      TEXT NOT NULL DEFAULT '',
    thumbnail_url   TEXT NOT NULL DEFAULT '',
    format          TEXT NOT NULL DEFAULT '',
    quality         TEXT NOT NULL DEFAULT '',
    file_size       INTEGER NOT NULL DEFAULT 0,
    video_id        TEXT NOT NULL DEFAULT '',
    tags_embedded   INTEGER NOT NULL DEFAULT 0,
    tag_error       TEXT NOT NULL DEFAULT '',
    track_number    INTEGER NOT NULL DEFAULT 0,
    playlist_id     TEXT NOT NULL DEFAULT '',
    playlist_title  TEXT NOT NULL DEFAULT '',
    playlist_index  INTEGER NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_media_media_type ON media(media_type);
CREATE INDEX IF NOT EXISTS idx_media_artist ON media(artist);
CREATE INDEX IF NOT EXISTS idx_media_video_id ON media(video_id);
CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at);
`

// DB wraps an SQLite connection for the media catalog.
type DB struct {
	db *sql.DB
	mu sync.Mutex
}

// Open opens or creates the SQLite database at the given path.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database at %s: %w", path, err)
	}

	// SQLite pragmas for performance and reliability
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("setting pragma %q: %w", pragma, err)
		}
	}

	if _, err := sqlDB.Exec(createTableSQL); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return &DB{db: sqlDB}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

// InsertMedia inserts a new media record and returns the inserted ID.
func (d *DB) InsertMedia(record MediaRecord) (int64, error) {
	if d == nil || d.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tagsEmbedded := 0
	if record.TagsEmbedded {
		tagsEmbedded = 1
	}

	result, err := d.db.Exec(`
		INSERT INTO media (
			title, artist, album, duration, media_type,
			file_path, source_url, thumbnail_url, format, quality,
			file_size, video_id, tags_embedded, tag_error,
			track_number, playlist_id, playlist_title, playlist_index
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.Title, record.Artist, record.Album, record.Duration, record.MediaType,
		record.FilePath, record.SourceURL, record.ThumbnailURL, record.Format, record.Quality,
		record.FileSize, record.VideoID, tagsEmbedded, record.TagError,
		record.TrackNumber, record.PlaylistID, record.PlaylistTitle, record.PlaylistIndex,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting media record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// UpsertMedia inserts or updates a media record by file_path.
func (d *DB) UpsertMedia(record MediaRecord) (int64, error) {
	if d == nil || d.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tagsEmbedded := 0
	if record.TagsEmbedded {
		tagsEmbedded = 1
	}

	_, err := d.db.Exec(`
		INSERT INTO media (
			title, artist, album, duration, media_type,
			file_path, source_url, thumbnail_url, format, quality,
			file_size, video_id, tags_embedded, tag_error,
			track_number, playlist_id, playlist_title, playlist_index
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			title=excluded.title, artist=excluded.artist, album=excluded.album,
			duration=excluded.duration, media_type=excluded.media_type,
			source_url=excluded.source_url, thumbnail_url=excluded.thumbnail_url,
			format=excluded.format, quality=excluded.quality,
			file_size=excluded.file_size, video_id=excluded.video_id,
			tags_embedded=excluded.tags_embedded, tag_error=excluded.tag_error,
			track_number=excluded.track_number,
			playlist_id=excluded.playlist_id, playlist_title=excluded.playlist_title,
			playlist_index=excluded.playlist_index
	`,
		record.Title, record.Artist, record.Album, record.Duration, record.MediaType,
		record.FilePath, record.SourceURL, record.ThumbnailURL, record.Format, record.Quality,
		record.FileSize, record.VideoID, tagsEmbedded, record.TagError,
		record.TrackNumber, record.PlaylistID, record.PlaylistTitle, record.PlaylistIndex,
	)
	if err != nil {
		return 0, fmt.Errorf("upserting media record: %w", err)
	}

	// LastInsertId is unreliable for ON CONFLICT DO UPDATE; query the actual row ID.
	var id int64
	if err := d.db.QueryRow("SELECT id FROM media WHERE file_path = ?", record.FilePath).Scan(&id); err != nil {
		return 0, fmt.Errorf("querying upserted media id: %w", err)
	}
	return id, nil
}

// ListMedia returns all media records, ordered by created_at descending.
func (d *DB) ListMedia(limit, offset int) ([]MediaRecord, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := d.db.Query(`
		SELECT id, title, artist, album, duration, media_type,
			file_path, source_url, thumbnail_url, format, quality,
			file_size, video_id, tags_embedded, tag_error,
			track_number, playlist_id, playlist_title, playlist_index, created_at
		FROM media
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("querying media: %w", err)
	}
	defer rows.Close()

	var records []MediaRecord
	for rows.Next() {
		var r MediaRecord
		var tagsEmbedded int
		if err := rows.Scan(
			&r.ID, &r.Title, &r.Artist, &r.Album, &r.Duration, &r.MediaType,
			&r.FilePath, &r.SourceURL, &r.ThumbnailURL, &r.Format, &r.Quality,
			&r.FileSize, &r.VideoID, &tagsEmbedded, &r.TagError,
			&r.TrackNumber, &r.PlaylistID, &r.PlaylistTitle, &r.PlaylistIndex, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning media row: %w", err)
		}
		r.TagsEmbedded = tagsEmbedded != 0
		records = append(records, r)
	}
	return records, rows.Err()
}

// Count returns the total number of media records.
func (d *DB) Count() (int, error) {
	if d == nil || d.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var count int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM media").Scan(&count); err != nil {
		return 0, fmt.Errorf("counting media: %w", err)
	}
	return count, nil
}
