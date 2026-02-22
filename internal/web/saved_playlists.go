package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	maxSavedPlaylistNameLength = 80
	savedPlaylistsFileName     = "saved_playlists.json"
)

type savedPlaylist struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type savedPlaylistState struct {
	Playlists   []savedPlaylist   `json:"playlists"`
	Assignments map[string]string `json:"assignments"`
}

type savedPlaylistMigrationResponse struct {
	Playlists   []savedPlaylist   `json:"playlists"`
	Assignments map[string]string `json:"assignments"`
	Migrated    bool              `json:"migrated"`
}

type savedPlaylistStore struct {
	path string
	mu   sync.Mutex
}

func newSavedPlaylistStore(path string) *savedPlaylistStore {
	return &savedPlaylistStore{path: path}
}

func (s *savedPlaylistStore) Load() (savedPlaylistState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loadLocked()
}

func (s *savedPlaylistStore) Replace(next savedPlaylistState) (savedPlaylistState, error) {
	normalized := normalizeSavedPlaylistState(next)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.saveLocked(normalized); err != nil {
		return emptySavedPlaylistState(), err
	}
	return cloneSavedPlaylistState(normalized), nil
}

func (s *savedPlaylistStore) MigrateFromLegacy(next savedPlaylistState) (savedPlaylistState, bool, error) {
	normalizedIncoming := normalizeSavedPlaylistState(next)

	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.loadLocked()
	if err != nil {
		return emptySavedPlaylistState(), false, err
	}
	if hasSavedPlaylistData(current) || !hasSavedPlaylistData(normalizedIncoming) {
		return current, false, nil
	}
	if err := s.saveLocked(normalizedIncoming); err != nil {
		return emptySavedPlaylistState(), false, err
	}
	return cloneSavedPlaylistState(normalizedIncoming), true, nil
}

func (s *savedPlaylistStore) loadLocked() (savedPlaylistState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return emptySavedPlaylistState(), nil
		}
		return emptySavedPlaylistState(), fmt.Errorf("reading saved playlists: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return emptySavedPlaylistState(), fmt.Errorf("saved playlists file is empty")
	}

	var decoded savedPlaylistState
	if err := json.Unmarshal(data, &decoded); err != nil {
		return emptySavedPlaylistState(), fmt.Errorf("parsing saved playlists: %w", err)
	}
	return normalizeSavedPlaylistState(decoded), nil
}

func (s *savedPlaylistStore) saveLocked(next savedPlaylistState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("creating playlist data directory: %w", err)
	}
	encoded, err := json.MarshalIndent(normalizeSavedPlaylistState(next), "", "  ")
	if err != nil {
		return fmt.Errorf("encoding saved playlists: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("writing saved playlists temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("committing saved playlists file: %w", err)
	}
	return nil
}

func emptySavedPlaylistState() savedPlaylistState {
	return savedPlaylistState{
		Playlists:   []savedPlaylist{},
		Assignments: map[string]string{},
	}
}

func hasSavedPlaylistData(state savedPlaylistState) bool {
	return len(state.Playlists) > 0 || len(state.Assignments) > 0
}

func normalizeSavedPlaylistState(raw savedPlaylistState) savedPlaylistState {
	out := emptySavedPlaylistState()
	seenIDs := make(map[string]struct{}, len(raw.Playlists))
	seenNames := make(map[string]struct{}, len(raw.Playlists))

	for _, entry := range raw.Playlists {
		id := strings.TrimSpace(entry.ID)
		name := normalizeSavedPlaylistName(entry.Name)
		if id == "" || name == "" {
			continue
		}
		if _, exists := seenIDs[id]; exists {
			continue
		}
		nameKey := strings.ToLower(name)
		if _, exists := seenNames[nameKey]; exists {
			continue
		}
		seenIDs[id] = struct{}{}
		seenNames[nameKey] = struct{}{}
		out.Playlists = append(out.Playlists, savedPlaylist{
			ID:        id,
			Name:      name,
			CreatedAt: strings.TrimSpace(entry.CreatedAt),
			UpdatedAt: strings.TrimSpace(entry.UpdatedAt),
		})
	}

	validIDs := make(map[string]struct{}, len(out.Playlists))
	for _, playlist := range out.Playlists {
		validIDs[playlist.ID] = struct{}{}
	}

	for mediaKey, playlistID := range raw.Assignments {
		normalizedMediaKey := strings.TrimSpace(mediaKey)
		normalizedPlaylistID := strings.TrimSpace(playlistID)
		if normalizedMediaKey == "" || normalizedPlaylistID == "" {
			continue
		}
		if _, ok := validIDs[normalizedPlaylistID]; !ok {
			continue
		}
		out.Assignments[normalizedMediaKey] = normalizedPlaylistID
	}

	return out
}

func normalizeSavedPlaylistName(value string) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if len(normalized) > maxSavedPlaylistNameLength {
		return normalized[:maxSavedPlaylistNameLength]
	}
	return normalized
}

func cloneSavedPlaylistState(source savedPlaylistState) savedPlaylistState {
	out := savedPlaylistState{
		Playlists:   append([]savedPlaylist(nil), source.Playlists...),
		Assignments: make(map[string]string, len(source.Assignments)),
	}
	for mediaKey, playlistID := range source.Assignments {
		out.Assignments[mediaKey] = playlistID
	}
	return out
}
