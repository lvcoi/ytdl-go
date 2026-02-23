import { describe, it, expect } from 'vitest';
import { normalizeLibrary, filterLibrary } from './libraryModel';

// Mock data
const mockDownloads = [
  {
    filename: 'song1.mp3',
    title: 'Song 1',
    artist: 'Artist A',
    album: 'Album X',
    type: 'Music',
    date: '2023-01-01',
    size: '5MB',
  },
  {
    filename: 'video1.mp4',
    title: 'Video 1',
    artist: 'Creator B',
    type: 'YouTube Video',
    date: '2023-01-02',
    size: '50MB',
  },
  {
    filename: 'podcast1.mp3',
    title: 'Podcast 1',
    artist: 'Podcaster C',
    type: 'Podcast',
    date: '2023-01-03',
    size: '20MB',
  }
];

const mockSavedPlaylists = [
  { id: 'playlist1', name: 'My Playlist' }
];

const mockAssignments = {
  'song1.mp3': 'playlist1'
};

describe('libraryModel', () => {
  describe('normalizeLibrary', () => {
    it('normalizes raw downloads into items', () => {
      const result = normalizeLibrary(mockDownloads, mockSavedPlaylists, mockAssignments);

      expect(result.items).toHaveLength(3);
      expect(result.anomalyCount).toBeDefined();

      const song1 = result.items.find(i => i.filename === 'song1.mp3');
      expect(song1).toBeDefined();
      expect(song1.title).toBe('Song 1');
      expect(song1.creator).toBe('Artist A');
      expect(song1.savedPlaylistId).toBe('playlist1');
      expect(song1.savedPlaylistName).toBe('My Playlist');
    });

    it('handles missing assignments gracefully', () => {
      const result = normalizeLibrary(mockDownloads, mockSavedPlaylists, {});
      const song1 = result.items.find(i => i.filename === 'song1.mp3');
      expect(song1.savedPlaylistId).toBe('');
      expect(song1.savedPlaylistName).toBe('Unassigned');
    });
  });

  describe('filterLibrary', () => {
    // Helper to get normalized data first
    const normalizedData = normalizeLibrary(mockDownloads, mockSavedPlaylists, mockAssignments);

    it('filters items by type', () => {
      const result = filterLibrary(
        normalizedData,
        mockSavedPlaylists,
        'Music', // typeFilter
        'newest', // sortKey
        {} // filters
      );

      expect(result.filteredItems).toHaveLength(1);
      expect(result.filteredItems[0].title).toBe('Song 1');
    });

    it('filters items by query', () => {
      const result = filterLibrary(
        normalizedData,
        mockSavedPlaylists,
        'all',
        'newest',
        { query: 'Video' }
      );

      expect(result.filteredItems).toHaveLength(1);
      expect(result.filteredItems[0].title).toBe('Video 1');
    });

    it('groups items by artist', () => {
      const result = filterLibrary(
        normalizedData,
        mockSavedPlaylists,
        'all',
        'newest',
        {}
      );

      expect(result.artists).toHaveLength(1);
      expect(result.artists[0].name).toBe('Artist A');
      expect(result.artists[0].items).toHaveLength(1);
    });

    it('groups items by video creator', () => {
        const result = filterLibrary(
          normalizedData,
          mockSavedPlaylists,
          'all',
          'newest',
          {}
        );

        expect(result.videos).toHaveLength(1);
        expect(result.videos[0].name).toBe('Creator B');
      });

    it('includes saved playlists in playlist groups', () => {
      const result = filterLibrary(
        normalizedData,
        mockSavedPlaylists,
        'all',
        'newest',
        {}
      );

      expect(result.savedPlaylists).toHaveLength(1);
      expect(result.savedPlaylists[0].name).toBe('My Playlist');
      expect(result.savedPlaylists[0].items).toHaveLength(1); // Song 1 is assigned
    });
  });
});
