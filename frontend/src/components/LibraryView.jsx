import { For, Show, createEffect, createMemo, createSignal } from 'solid-js';
import Icon from './Icon';
import Thumbnail from './Thumbnail';
import { Grid, GridItem } from './Grid';
import { buildLibraryModel } from '../utils/libraryModel';

const SECTION_OPTIONS = [
  { value: 'artists', label: 'Music' },
  { value: 'videos', label: 'YouTube Videos' },
  { value: 'podcasts', label: 'Podcasts' },
  { value: 'playlists', label: 'Playlists' },
  { value: 'all_media', label: 'All Media' },
];

const VIEW_MODE_OPTIONS = [
  { value: 'gallery', label: 'Gallery' },
  { value: 'columns', label: 'Columns' },
  { value: 'list', label: 'List' },
];

const TYPE_FILTER_OPTIONS = [
  { value: 'all', label: 'All Types' },
  { value: 'Music', label: 'Music' },
  { value: 'YouTube Video', label: 'YouTube Videos' },
  { value: 'Podcast', label: 'Podcast' },
];

const SORT_OPTIONS = [
  { value: 'newest', label: 'Newest first' },
  { value: 'oldest', label: 'Oldest first' },
  { value: 'creator_asc', label: 'Creator (A-Z)' },
  { value: 'creator_desc', label: 'Creator (Z-A)' },
  { value: 'collection_asc', label: 'Collection (A-Z)' },
  { value: 'collection_desc', label: 'Collection (Z-A)' },
  { value: 'playlist_asc', label: 'Playlist (A-Z)' },
  { value: 'playlist_desc', label: 'Playlist (Z-A)' },
];

const MAX_SAVED_PLAYLIST_NAME_LENGTH = 80;

const EMPTY_NAV_PATH = {
  creatorType: '',
  creatorName: '',
  albumName: '',
  playlistKey: '',
  playlistKind: '',
};

const normalizeSavedPlaylistName = (value) => (
  String(value || '')
    .trim()
    .replace(/\s+/g, ' ')
    .slice(0, MAX_SAVED_PLAYLIST_NAME_LENGTH)
);

const normalizeNavPath = (raw) => {
  const value = raw && typeof raw === 'object' ? raw : {};
  return {
    creatorType: typeof value.creatorType === 'string' ? value.creatorType : '',
    creatorName: typeof value.creatorName === 'string' ? value.creatorName : '',
    albumName: typeof value.albumName === 'string' ? value.albumName : '',
    playlistKey: typeof value.playlistKey === 'string' ? value.playlistKey : '',
    playlistKind: typeof value.playlistKind === 'string' ? value.playlistKind : '',
  };
};

const normalizeFilters = (raw) => {
  const value = raw && typeof raw === 'object' ? raw : {};
  return {
    query: typeof value.query === 'string' ? value.query : '',
    creator: typeof value.creator === 'string' ? value.creator : '',
    collection: typeof value.collection === 'string' ? value.collection : '',
    playlist: typeof value.playlist === 'string' ? value.playlist : '',
    savedPlaylistId: typeof value.savedPlaylistId === 'string' ? value.savedPlaylistId : '',
  };
};

const normalizeUiState = (raw) => {
  const value = raw && typeof raw === 'object' ? raw : {};
  return {
    advancedFiltersOpen: Boolean(value.advancedFiltersOpen),
    metadataBannerDismissed: Boolean(value.metadataBannerDismissed),
  };
};

const firstNonEmpty = (...values) => {
  for (const value of values) {
    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (trimmed !== '') {
        return trimmed;
      }
    }
  }
  return '';
};

export default function LibraryView(props) {
  const [newSavedPlaylistName, setNewSavedPlaylistName] = createSignal('');
  const [playlistMessage, setPlaylistMessage] = createSignal('');
  const [playlistMessageTone, setPlaylistMessageTone] = createSignal('neutral');
  const [showSavedPlaylistManager, setShowSavedPlaylistManager] = createSignal(false);
  const [selectedDetailKey, setSelectedDetailKey] = createSignal('');
  const [metadataRetryMessage, setMetadataRetryMessage] = createSignal('');

  const downloads = createMemo(() => {
    const source = typeof props.downloads === 'function' ? props.downloads() : props.downloads;
    return Array.isArray(source) ? source : [];
  });

  const section = createMemo(() => {
    const source = typeof props.section === 'function' ? props.section() : props.section;
    return SECTION_OPTIONS.some((option) => option.value === source) ? source : 'artists';
  });

  const viewMode = createMemo(() => {
    const source = typeof props.viewMode === 'function' ? props.viewMode() : props.viewMode;
    return VIEW_MODE_OPTIONS.some((option) => option.value === source) ? source : 'gallery';
  });

  const typeFilter = createMemo(() => {
    const source = typeof props.typeFilter === 'function' ? props.typeFilter() : props.typeFilter;
    return TYPE_FILTER_OPTIONS.some((option) => option.value === source) ? source : 'all';
  });

  const navPath = createMemo(() => {
    const source = typeof props.navPath === 'function' ? props.navPath() : props.navPath;
    return normalizeNavPath(source);
  });

  const filters = createMemo(() => {
    const source = typeof props.filters === 'function' ? props.filters() : props.filters;
    return normalizeFilters(source);
  });

  const sortKey = createMemo(() => {
    const source = typeof props.sortKey === 'function' ? props.sortKey() : props.sortKey;
    return SORT_OPTIONS.some((option) => option.value === source) ? source : 'newest';
  });

  const savedPlaylists = createMemo(() => {
    const source = typeof props.savedPlaylists === 'function' ? props.savedPlaylists() : props.savedPlaylists;
    return Array.isArray(source) ? source : [];
  });

  const playlistAssignments = createMemo(() => {
    const source = typeof props.playlistAssignments === 'function' ? props.playlistAssignments() : props.playlistAssignments;
    return source && typeof source === 'object' ? source : {};
  });

  const savedPlaylistSyncError = createMemo(() => {
    const source = typeof props.savedPlaylistSyncError === 'function'
      ? props.savedPlaylistSyncError()
      : props.savedPlaylistSyncError;
    return typeof source === 'string' ? source.trim() : '';
  });

  const uiState = createMemo(() => {
    const source = typeof props.uiState === 'function' ? props.uiState() : props.uiState;
    return normalizeUiState(source);
  });

  const model = createMemo(() => buildLibraryModel({
    downloads: downloads(),
    savedPlaylists: savedPlaylists(),
    playlistAssignments: playlistAssignments(),
    typeFilter: typeFilter(),
    sortKey: sortKey(),
    filters: filters(),
  }));

  const setPlaylistFeedback = (tone, text) => {
    setPlaylistMessageTone(tone);
    setPlaylistMessage(text);
  };

  const clearNavPath = () => {
    if (typeof props.onNavPathChange === 'function') {
      props.onNavPathChange({ ...EMPTY_NAV_PATH });
    }
  };

  const setNavPath = (next) => {
    if (typeof props.onNavPathChange === 'function') {
      props.onNavPathChange({ ...EMPTY_NAV_PATH, ...next });
    }
  };

  const updateFilter = (filterKey, value) => {
    if (typeof props.onFilterChange === 'function') {
      props.onFilterChange(filterKey, value);
    }
  };

  const hasActiveFilters = createMemo(() => {
    const value = filters();
    return value.query !== ''
      || value.creator !== ''
      || value.collection !== ''
      || value.playlist !== ''
      || value.savedPlaylistId !== '';
  });

  const clearFilters = () => {
    if (typeof props.onClearFilters === 'function') {
      props.onClearFilters();
      return;
    }
    updateFilter('query', '');
    updateFilter('creator', '');
    updateFilter('collection', '');
    updateFilter('playlist', '');
    updateFilter('savedPlaylistId', '');
  };

  const setSection = (nextSection) => {
    if (typeof props.onSectionChange === 'function') {
      props.onSectionChange(nextSection);
    }
    clearNavPath();
    setSelectedDetailKey('');
  };

  const setViewMode = (nextViewMode) => {
    if (typeof props.onViewModeChange === 'function') {
      props.onViewModeChange(nextViewMode);
    }
    setSelectedDetailKey('');
  };

  const setTypeFilter = (nextTypeFilter) => {
    if (typeof props.onTypeFilterChange === 'function') {
      props.onTypeFilterChange(nextTypeFilter);
    }
    clearNavPath();
    setSelectedDetailKey('');
  };

  const handleRetryMetadataScan = async () => {
    if (typeof props.onRetryMetadataScan !== 'function') {
      return;
    }
    try {
      const result = await props.onRetryMetadataScan();
      const message = typeof result?.message === 'string' && result.message.trim() !== ''
        ? result.message.trim()
        : 'Library refreshed. Legacy files without sidecar metadata may require re-download for full metadata and thumbnails.';
      setMetadataRetryMessage(message);
    } catch (error) {
      setMetadataRetryMessage('Retry failed. Please check backend logs and try again.');
    }
  };

  const handleDismissMetadataBanner = () => {
    if (typeof props.onDismissMetadataBanner === 'function') {
      props.onDismissMetadataBanner();
    }
  };

  const handleCreateSavedPlaylist = async () => {
    const nextName = normalizeSavedPlaylistName(newSavedPlaylistName());
    if (nextName === '') {
      setPlaylistFeedback('error', 'Enter a playlist name.');
      return;
    }
    if (typeof props.onCreateSavedPlaylist !== 'function') {
      return;
    }

    const result = await props.onCreateSavedPlaylist(nextName);
    if (!result || result.ok === false) {
      setPlaylistFeedback('error', result?.error || 'Unable to create saved playlist.');
      return;
    }
    setNewSavedPlaylistName('');
    setPlaylistFeedback('success', `Saved playlist "${nextName}" created.`);
  };

  const handleRenameSavedPlaylist = async (playlist) => {
    if (typeof window === 'undefined' || typeof props.onRenameSavedPlaylist !== 'function') {
      return;
    }
    const proposedName = window.prompt('Rename saved playlist', playlist.name);
    if (proposedName === null) {
      return;
    }

    const result = await props.onRenameSavedPlaylist(playlist.id, proposedName);
    if (!result || result.ok === false) {
      setPlaylistFeedback('error', result?.error || 'Unable to rename saved playlist.');
      return;
    }

    const normalizedName = normalizeSavedPlaylistName(proposedName);
    setPlaylistFeedback('success', `Saved playlist renamed to "${normalizedName || playlist.name}".`);
  };

  const handleDeleteSavedPlaylist = async (playlist) => {
    if (typeof window === 'undefined' || typeof props.onDeleteSavedPlaylist !== 'function') {
      return;
    }

    const confirmed = window.confirm(`Delete "${playlist.name}" and remove its assignments?`);
    if (!confirmed) {
      return;
    }

    const result = await props.onDeleteSavedPlaylist(playlist.id);
    if (!result || result.ok === false) {
      setPlaylistFeedback('error', result?.error || 'Unable to delete saved playlist.');
      return;
    }

    setPlaylistFeedback('success', `Saved playlist "${playlist.name}" deleted.`);
  };

  const handleAssignSavedPlaylist = (item, nextSavedPlaylistId) => {
    if (typeof props.onAssignSavedPlaylist !== 'function') {
      return;
    }
    if (item.mediaKey === '') {
      return;
    }
    void props.onAssignSavedPlaylist(item.mediaKey, nextSavedPlaylistId);
  };

  const handlePlayItem = (item, queueItems) => {
    if (typeof props.openPlayer !== 'function') {
      return;
    }
    const queue = Array.isArray(queueItems) ? queueItems : [];
    props.openPlayer(item.raw, queue.map((entry) => entry.raw));
  };

  const explorer = createMemo(() => {
    const libraryModel = model();
    const currentSection = section();
    const currentNav = navPath();

    if (currentSection === 'artists') {
      if (currentNav.creatorType === 'artist' && currentNav.creatorName !== '') {
        const creator = libraryModel.artistsByName.get(currentNav.creatorName);
        if (!creator) {
          return {
            kind: 'landing',
            title: 'Music',
            subtitle: 'Browse music creators.',
            artists: libraryModel.artists,
            videos: libraryModel.videos,
            breadcrumbs: [],
          };
        }

        if (currentNav.albumName !== '') {
          const album = creator.albums.find((entry) => entry.name === currentNav.albumName);
          if (album) {
            return {
              kind: 'items',
              title: album.name,
              subtitle: `${creator.name} • ${album.count} track${album.count === 1 ? '' : 's'}`,
              items: album.items,
              queueItems: album.items,
              breadcrumbs: [
                { label: 'Music', nav: { ...EMPTY_NAV_PATH } },
                { label: creator.name, nav: { creatorType: 'artist', creatorName: creator.name } },
                { label: album.name, nav: { creatorType: 'artist', creatorName: creator.name, albumName: album.name } },
              ],
            };
          }
        }

        return {
          kind: 'albums',
          title: creator.name,
          subtitle: `${creator.count} track${creator.count === 1 ? '' : 's'} across ${creator.albums.length} album${creator.albums.length === 1 ? '' : 's'}`,
          albums: creator.albums,
          creator,
          breadcrumbs: [
            { label: 'Music', nav: { ...EMPTY_NAV_PATH } },
            { label: creator.name, nav: { creatorType: 'artist', creatorName: creator.name } },
          ],
        };
      }

      return {
        kind: 'landing',
        title: 'Music',
        subtitle: 'Music artists from your collection.',
        artists: libraryModel.artists,
        videos: libraryModel.videos,
        breadcrumbs: [],
      };
    }

    if (currentSection === 'videos') {
      if (currentNav.creatorType === 'video_creator' && currentNav.creatorName !== '') {
        const creator = libraryModel.videosByName.get(currentNav.creatorName);
        if (creator) {
          return {
            kind: 'items',
            title: creator.name,
            subtitle: `${creator.count} video${creator.count === 1 ? '' : 's'}`,
            items: creator.items,
            queueItems: creator.items,
            breadcrumbs: [
              { label: 'YouTube Videos', nav: { ...EMPTY_NAV_PATH } },
              { label: creator.name, nav: { creatorType: 'video_creator', creatorName: creator.name } },
            ],
          };
        }
      }

      return {
        kind: 'creators',
        title: 'YouTube Videos',
        subtitle: 'YouTube channels from your collection.',
        creators: libraryModel.videos,
        creatorType: 'video_creator',
        breadcrumbs: [],
      };
    }

    if (currentSection === 'podcasts') {
      if (currentNav.creatorType === 'podcast' && currentNav.creatorName !== '') {
        const podcast = libraryModel.podcastsByName.get(currentNav.creatorName);
        if (podcast) {
          return {
            kind: 'items',
            title: podcast.name,
            subtitle: `${podcast.count} episode${podcast.count === 1 ? '' : 's'}`,
            items: podcast.items,
            queueItems: podcast.items,
            breadcrumbs: [
              { label: 'Podcasts', nav: { ...EMPTY_NAV_PATH } },
              { label: podcast.name, nav: { creatorType: 'podcast', creatorName: podcast.name } },
            ],
          };
        }
      }

      return {
        kind: 'creators',
        title: 'Podcasts',
        subtitle: 'Podcasts from your collection.',
        creators: libraryModel.podcasts,
        creatorType: 'podcast',
        breadcrumbs: [],
      };
    }

    if (currentSection === 'playlists') {
      if (currentNav.playlistKey !== '') {
        const playlist = libraryModel.playlistsByKey.get(currentNav.playlistKey);
        if (playlist) {
          return {
            kind: 'items',
            title: playlist.name,
            subtitle: `${playlist.count} item${playlist.count === 1 ? '' : 's'} • ${playlist.kind === 'source' ? 'Source playlist' : 'Saved playlist'}`,
            items: playlist.items,
            queueItems: playlist.items,
            breadcrumbs: [
              { label: 'Playlists', nav: { ...EMPTY_NAV_PATH } },
              { label: playlist.name, nav: { playlistKey: playlist.key, playlistKind: playlist.kind } },
            ],
          };
        }
      }

      return {
        kind: 'playlists',
        title: 'Playlists',
        subtitle: 'Combined source and saved playlists.',
        playlists: libraryModel.playlists,
        sourcePlaylists: libraryModel.sourcePlaylists,
        savedPlaylists: libraryModel.savedPlaylists,
        breadcrumbs: [],
      };
    }

    return {
      kind: 'items',
      title: 'All Media',
      subtitle: `${libraryModel.filteredItems.length} filtered item${libraryModel.filteredItems.length === 1 ? '' : 's'}`,
      items: libraryModel.filteredItems,
      queueItems: libraryModel.filteredItems,
      breadcrumbs: [],
    };
  });

  const creatorDetailRows = createMemo(() => {
    const context = explorer();
    if (context.kind === 'landing') {
      const artists = context.artists.map((entry) => ({
        key: `artist:${entry.name}`,
        name: entry.name,
        type: 'Music',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: `${entry.albums.length} album${entry.albums.length === 1 ? '' : 's'}`,
        onOpen: () => setNavPath({ creatorType: 'artist', creatorName: entry.name }),
      }));
      const videos = context.videos.map((entry) => ({
        key: `video_creator:${entry.name}`,
        name: entry.name,
        type: 'YouTube Video',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: 'Video library',
        onOpen: () => {
          setNavPath({ creatorType: 'video_creator', creatorName: entry.name });
        },
      }));
      const podcasts = context.podcasts?.map((entry) => ({
        key: `podcast:${entry.name}`,
        name: entry.name,
        type: 'Podcast',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: 'Podcast library',
        onOpen: () => {
          setSection('podcasts');
          setNavPath({ creatorType: 'podcast', creatorName: entry.name });
        },
      })) || [];
      return [...artists, ...videos, ...podcasts].sort((left, right) => right.latestTimestamp - left.latestTimestamp);
    }

    if (context.kind === 'creators') {
      return context.creators.map((entry) => ({
        key: `${context.creatorType}:${entry.name}`,
        name: entry.name,
        type: context.creatorType === 'video_creator' ? 'YouTube Video' : 'Music',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: context.creatorType === 'video_creator'
          ? 'Video library'
          : `${entry.albums?.length || 0} album${(entry.albums?.length || 0) === 1 ? '' : 's'}`,
        onOpen: () => setNavPath({ creatorType: context.creatorType, creatorName: entry.name }),
      }));
    }

    if (context.kind === 'albums') {
      return context.albums.map((entry) => ({
        key: `album:${entry.name}`,
        name: entry.name,
        type: 'Album',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: context.creator.name,
        onOpen: () => setNavPath({ creatorType: 'artist', creatorName: context.creator.name, albumName: entry.name }),
      }));
    }

    if (context.kind === 'playlists') {
      return context.playlists.map((entry) => ({
        key: entry.key,
        name: entry.name,
        type: entry.kind === 'source' ? 'Source Playlist' : 'Saved Playlist',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: entry.kind === 'saved' && entry.count === 0 ? 'No assignments yet' : '',
        onOpen: () => setNavPath({ playlistKey: entry.key, playlistKind: entry.kind }),
      }));
    }

    return [];
  });

  const explorerItems = createMemo(() => (explorer().kind === 'items' ? explorer().items : []));
  const explorerQueueItems = createMemo(() => (explorer().kind === 'items' ? explorer().queueItems : []));
  const explorerCreators = createMemo(() => (explorer().kind === 'creators' ? explorer().creators : []));
  const explorerAlbums = createMemo(() => (explorer().kind === 'albums' ? explorer().albums : []));
  const explorerPlaylists = createMemo(() => (explorer().kind === 'playlists' ? explorer().playlists : []));
  const explorerLandingArtists = createMemo(() => (explorer().kind === 'landing' ? explorer().artists : []));
  const explorerLandingVideos = createMemo(() => (explorer().kind === 'landing' ? explorer().videos : []));

  const explorerHeader = createMemo(() => {
    const context = explorer();
    if (context.kind === 'items' && navPath().creatorType === 'artist' && navPath().albumName !== '') {
      return (
        <div class="space-y-1">
          <h1 class="text-4xl font-black tracking-tight text-white">{navPath().creatorName}</h1>
          <h2 class="text-2xl font-bold text-accent-primary/80">{navPath().albumName}</h2>
        </div>
      );
    }
    return (
      <div>
        <h1 class="text-4xl font-black tracking-tight text-white">{context.title || 'Library'}</h1>
        <p class="text-base text-gray-400 font-medium">{context.subtitle}</p>
      </div>
    );
  });

  createEffect(() => {
    const context = explorer();
    const currentSelected = selectedDetailKey();

    if (viewMode() === 'gallery') {
      return;
    }

    if (context.kind === 'items') {
      const firstKey = context.items[0]?.mediaKey || '';
      if (!context.items.some((entry) => entry.mediaKey === currentSelected)) {
        setSelectedDetailKey(firstKey);
      }
      return;
    }

    const rows = creatorDetailRows();
    const firstKey = rows[0]?.key || '';
    if (!rows.some((entry) => entry.key === currentSelected)) {
      setSelectedDetailKey(firstKey);
    }
  });

  const selectedDetailItem = createMemo(() => {
    const context = explorer();
    if (context.kind !== 'items') {
      return null;
    }
    return context.items.find((entry) => entry.mediaKey === selectedDetailKey()) || null;
  });

  const selectedDetailRow = createMemo(() => {
    const rows = creatorDetailRows();
    return rows.find((entry) => entry.key === selectedDetailKey()) || null;
  });

  const sectionCountLabel = createMemo(() => {
    const context = explorer();
    if (context.kind === 'items') {
      return `${context.items.length} items`;
    }
    if (context.kind === 'albums') {
      return `${context.albums.length} albums`;
    }
    if (context.kind === 'playlists') {
      return `${context.playlists.length} playlists`;
    }
    if (context.kind === 'creators') {
      return `${context.creators.length} creators`;
    }
    if (context.kind === 'landing') {
      return `${context.artists.length + context.videos.length} creators`;
    }
    return '';
  });

  return (
    <div class="space-y-6 transition-smooth animate-in fade-in slide-in-from-right-4 duration-500">
      <div class="relative overflow-hidden rounded-[2rem] border border-accent-primary/20 glass-vibrant p-8 shadow-2xl">
        <div class="flex flex-col gap-8 lg:flex-row lg:items-start">
          <aside class="w-full lg:w-64 shrink-0 rounded-2xl border border-white/5 bg-black/40 p-4">
            <div class="text-[10px] font-black uppercase tracking-[0.2em] text-accent-primary/80 ml-2 mb-3">Library</div>
            <div class="space-y-1">
              <For each={SECTION_OPTIONS}>
                {(option) => (
                  <button
                    type="button"
                    onClick={() => setSection(option.value)}
                    class={`w-full rounded-xl px-4 py-3 text-left text-sm font-bold transition-smooth ${section() === option.value
                      ? 'bg-accent-primary/20 text-white border border-accent-primary/30 shadow-vibrant'
                      : 'text-gray-400 hover:text-white hover:bg-white/5 border border-transparent'
                      }`}
                  >
                    {option.label}
                  </button>
                )}
              </For>
            </div>
          </aside>

          <div class="min-w-0 flex-1 space-y-6">
            <div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
              {explorerHeader()}
              <div class="rounded-full border border-accent-secondary/30 bg-accent-secondary/10 px-4 py-1.5 text-xs font-black uppercase tracking-[0.15em] text-accent-secondary">
                {sectionCountLabel()}
              </div>
            </div>

            <Show when={model().anomalyCount > 0 && !uiState().metadataBannerDismissed}>
              <div class="rounded-2xl border border-amber-300/30 bg-amber-500/10 p-4 text-sm text-amber-100">
                <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <div class="font-semibold">
                      {model().anomalyCount} item{model().anomalyCount === 1 ? '' : 's'} missing expected metadata coverage.
                    </div>
                    <div class="text-xs text-amber-100/80">
                      Unknown buckets remain visible for transparency. Retry scan, then re-download legacy files without sidecars if needed.
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => { void handleRetryMetadataScan(); }}
                      class="rounded-lg border border-amber-200/40 bg-amber-300/10 px-3 py-1.5 text-xs font-semibold text-amber-100 hover:bg-amber-300/20 transition-colors"
                    >
                      Retry metadata
                    </button>
                    <button
                      type="button"
                      onClick={handleDismissMetadataBanner}
                      class="rounded-lg border border-white/20 px-3 py-1.5 text-xs font-semibold text-slate-200 hover:bg-white/10 transition-colors"
                    >
                      Dismiss
                    </button>
                  </div>
                </div>
                <Show when={metadataRetryMessage() !== ''}>
                  <div class="mt-2 text-xs text-amber-100/80">{metadataRetryMessage()}</div>
                </Show>
              </div>
            </Show>

            <div class="flex flex-wrap items-center gap-3 rounded-[1.5rem] border border-white/5 bg-black/40 p-3">
              <div class="inline-flex rounded-xl border border-white/10 bg-white/5 p-1">
                <For each={VIEW_MODE_OPTIONS}>
                  {(option) => (
                    <button
                      type="button"
                      onClick={() => setViewMode(option.value)}
                      class={`rounded-lg px-4 py-2 text-xs font-black uppercase tracking-widest transition-smooth ${viewMode() === option.value
                        ? 'bg-accent-primary text-white shadow-vibrant'
                        : 'text-gray-500 hover:text-white hover:bg-white/5'
                        }`}
                    >
                      {option.label}
                    </button>
                  )}
                </For>
              </div>

              <select
                value={typeFilter()}
                onChange={(event) => setTypeFilter(event.currentTarget.value)}
                class="rounded-xl border border-white/10 bg-black/40 px-4 py-2 text-xs font-bold text-white outline-none focus:border-accent-primary/50 transition-smooth"
              >
                <For each={TYPE_FILTER_OPTIONS}>
                  {(option) => <option value={option.value}>{option.label}</option>}
                </For>
              </select>

              <select
                value={sortKey()}
                onChange={(event) => {
                  if (typeof props.onSortKeyChange === 'function') {
                    props.onSortKeyChange(event.currentTarget.value);
                  }
                }}
                class="rounded-xl border border-white/10 bg-black/40 px-4 py-2 text-xs font-bold text-white outline-none focus:border-accent-primary/50 transition-smooth"
              >
                <For each={SORT_OPTIONS}>
                  {(option) => <option value={option.value}>{option.label}</option>}
                </For>
              </select>

              <button
                type="button"
                onClick={() => {
                  if (typeof props.onToggleAdvancedFilters === 'function') {
                    props.onToggleAdvancedFilters();
                  }
                }}
                class={`rounded-xl border px-4 py-2 text-xs font-bold transition-smooth ${uiState().advancedFiltersOpen
                  ? 'border-accent-primary/40 bg-accent-primary/20 text-white'
                  : 'border-white/10 bg-white/5 text-gray-400 hover:bg-white/10'
                  }`}
              >
                Filters
              </button>

              <Show when={section() === 'playlists'}>
                <button
                  type="button"
                  onClick={() => setShowSavedPlaylistManager(true)}
                  class="rounded-xl border border-accent-secondary/40 bg-accent-secondary/20 px-4 py-2 text-xs font-bold text-white hover:bg-accent-secondary/30 transition-smooth"
                >
                  Manage Playlists
                </button>
              </Show>

              <button
                type="button"
                onClick={clearFilters}
                disabled={!hasActiveFilters()}
                class={`ml-auto rounded-xl border px-4 py-2 text-xs font-bold transition-smooth ${hasActiveFilters()
                  ? 'border-white/20 bg-white/5 text-white hover:bg-white/10'
                  : 'border-white/5 bg-white/2 text-gray-700 cursor-not-allowed'
                  }`}
              >
                Reset
              </button>
            </div>

            <Show when={uiState().advancedFiltersOpen}>
              <div class="grid gap-3 rounded-2xl border border-white/10 bg-black/20 p-4 md:grid-cols-2 xl:grid-cols-5">
                <label class="space-y-1 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
                  Search
                  <input
                    type="text"
                    value={filters().query}
                    onInput={(event) => updateFilter('query', event.currentTarget.value)}
                    placeholder="Title, creator, album"
                    class="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                  />
                </label>

                <label class="space-y-1 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
                  Creator
                  <select
                    value={filters().creator}
                    onChange={(event) => updateFilter('creator', event.currentTarget.value)}
                    class="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                  >
                    <option value="">All creators</option>
                    <For each={model().filterOptions.creators}>
                      {(value) => <option value={value}>{value}</option>}
                    </For>
                  </select>
                </label>

                <label class="space-y-1 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
                  Collection
                  <select
                    value={filters().collection}
                    onChange={(event) => updateFilter('collection', event.currentTarget.value)}
                    class="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                  >
                    <option value="">All collections</option>
                    <For each={model().filterOptions.collections}>
                      {(value) => <option value={value}>{value}</option>}
                    </For>
                  </select>
                </label>

                <label class="space-y-1 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
                  Source Playlist
                  <select
                    value={filters().playlist}
                    onChange={(event) => updateFilter('playlist', event.currentTarget.value)}
                    class="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                  >
                    <option value="">All source playlists</option>
                    <For each={model().filterOptions.sourcePlaylists}>
                      {(value) => <option value={value}>{value}</option>}
                    </For>
                  </select>
                </label>

                <label class="space-y-1 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
                  Saved Playlist
                  <select
                    value={filters().savedPlaylistId}
                    onChange={(event) => updateFilter('savedPlaylistId', event.currentTarget.value)}
                    class="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                  >
                    <option value="">All saved playlists</option>
                    <For each={savedPlaylists()}>
                      {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                    </For>
                  </select>
                </label>
              </div>
            </Show>

            <Show when={savedPlaylistSyncError() !== ''}>
              <div class="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-red-400/30 bg-red-500/10 px-4 py-3 text-xs text-red-100">
                <span>{savedPlaylistSyncError()}</span>
                <Show when={typeof props.onRetrySavedPlaylistSync === 'function'}>
                  <button
                    type="button"
                    onClick={() => { void props.onRetrySavedPlaylistSync(); }}
                    class="rounded-lg border border-red-300/40 px-2 py-1 font-semibold text-red-100 hover:bg-red-500/20 transition-colors"
                  >
                    Retry
                  </button>
                </Show>
              </div>
            </Show>

            <Show when={explorer().breadcrumbs.length > 0}>
              <div class="flex flex-wrap items-center gap-2 text-xs">
                <For each={explorer().breadcrumbs}>
                  {(crumb, index) => (
                    <>
                      <button
                        type="button"
                        onClick={() => setNavPath(crumb.nav)}
                        class="rounded-md border border-white/15 bg-white/5 px-2 py-1 text-slate-200 hover:bg-white/10 transition-colors"
                      >
                        {crumb.label}
                      </button>
                      <Show when={index() < explorer().breadcrumbs.length - 1}>
                        <span class="text-slate-500">/</span>
                      </Show>
                    </>
                  )}
                </For>
              </div>
            </Show>

            <h2 class="text-xl font-bold text-white">{explorer().title}</h2>
          </div>
        </div>
      </div>

      <Show when={viewMode() === 'gallery'}>
        <Show
          when={explorer().kind === 'landing'}
          fallback={(
            <Show
              when={explorer().kind !== 'items'}
              fallback={(
                <Show
                  when={explorerItems().length > 0}
                  fallback={<div class="rounded-3xl border border-white/5 bg-black/20 p-12 text-center text-sm text-gray-500 font-medium">No media items match this view yet.</div>}
                >
                  <Grid class="!p-0 !gap-6 sm:grid-cols-2 xl:grid-cols-3">
                    <For each={explorerItems()}>
                      {(item) => (
                        <div class="group relative overflow-hidden rounded-[2rem] border border-white/5 bg-black/40 transition-smooth hover:border-accent-primary/50 hover:shadow-vibrant">
                          <div class="relative">
                            <Thumbnail
                              src={item.thumbnailUrl}
                              alt={item.title}
                              size="md"
                              class="!rounded-none"
                            />
                            <div class="absolute left-4 top-4 rounded-full border border-black/30 bg-black/60 backdrop-blur-md px-3 py-1 text-[10px] font-black uppercase tracking-widest text-white/90">
                              {item.type}
                            </div>
                          </div>
                          <div class="space-y-4 p-6">
                            <div class="space-y-1">
                              <div class="truncate text-lg font-black text-white">{item.title}</div>
                              <div class="truncate text-sm font-bold text-accent-primary/80">{item.creator}</div>
                              <div class="truncate text-xs font-medium text-gray-500">{item.album} • {item.date || item.modifiedAt}</div>
                            </div>
                            <div class="flex items-center gap-4">
                              <button
                                type="button"
                                onClick={() => handlePlayItem(item, explorerQueueItems())}
                                class="rounded-xl bg-vibrant-gradient px-6 py-2.5 text-xs font-black uppercase tracking-widest text-white hover:scale-105 transition-smooth shadow-vibrant"
                              >
                                Play
                              </button>
                              <div class="min-w-0 flex-1">
                                <select
                                  value={item.savedPlaylistId}
                                  onChange={(event) => handleAssignSavedPlaylist(item, event.currentTarget.value)}
                                  class="w-full truncate rounded-xl border border-white/10 bg-black/40 px-3 py-2 text-[10px] font-black uppercase tracking-widest text-white outline-none focus:border-accent-primary/50 transition-smooth"
                                >
                                  <option value="">Unassigned</option>
                                  <For each={savedPlaylists()}>
                                    {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                                  </For>
                                </select>
                              </div>
                            </div>
                          </div>
                        </div>
                      )}
                    </For>
                  </Grid>
                </Show>
              )}
            >
              <Show when={explorer().kind === 'creators'}>
                <Show
                  when={explorerCreators().length > 0}
                  fallback={<div class="rounded-3xl border border-white/5 bg-black/20 p-12 text-center text-sm text-gray-500 font-medium">No creators available for this filter set.</div>}
                >
                  <Grid class="!p-0 !gap-6 sm:grid-cols-2 xl:grid-cols-4">
                    <For each={explorerCreators()}>
                      {(entry) => (
                        <button
                          type="button"
                          onClick={() => setNavPath({ creatorType: explorer().creatorType, creatorName: entry.name })}
                          class="group relative overflow-hidden rounded-[2rem] border border-white/5 bg-black/40 text-left transition-smooth hover:border-accent-primary/50 hover:shadow-vibrant"
                        >
                          <Thumbnail
                            src={entry.thumbnailUrl}
                            alt={entry.name}
                            size="md"
                            class="!rounded-none"
                          />
                          <div class="space-y-1 p-5">
                            <div class="truncate text-base font-black text-white">{entry.name}</div>
                            <div class="text-xs font-bold text-gray-500">{entry.count} item{entry.count === 1 ? '' : 's'}</div>
                          </div>
                        </button>
                      )}
                    </For>
                  </Grid>
                </Show>
              </Show>

              <Show when={explorer().kind === 'albums'}>
                <Show
                  when={explorerAlbums().length > 0}
                  fallback={<div class="rounded-3xl border border-white/5 bg-black/20 p-12 text-center text-sm text-gray-500 font-medium">No albums available for this artist.</div>}
                >
                  <Grid class="!p-0 !gap-6 sm:grid-cols-2 xl:grid-cols-4">
                    <For each={explorerAlbums()}>
                      {(album) => (
                        <button
                          type="button"
                          onClick={() => setNavPath({ creatorType: 'artist', creatorName: explorer().creator.name, albumName: album.name })}
                          class="group relative overflow-hidden rounded-[2rem] border border-white/5 bg-black/40 text-left transition-smooth hover:border-accent-primary/50 hover:shadow-vibrant"
                        >
                          <Thumbnail
                            src={album.thumbnailUrl}
                            alt={album.name}
                            size="md"
                            class="!aspect-square !rounded-none"
                          />
                          <div class="space-y-1 p-5">
                            <div class="truncate text-base font-black text-white">{album.name}</div>
                            <div class="text-xs font-bold text-gray-500">{album.count} track{album.count === 1 ? '' : 's'}</div>
                          </div>
                        </button>
                      )}
                    </For>
                  </Grid>
                </Show>
              </Show>

              <Show when={explorer().kind === 'playlists'}>
                <Show
                  when={explorerPlaylists().length > 0}
                  fallback={<div class="rounded-3xl border border-white/5 bg-black/20 p-12 text-center text-sm text-gray-500 font-medium">No playlists available for this filter set.</div>}
                >
                  <Grid class="!p-0 !gap-6 sm:grid-cols-2 xl:grid-cols-3">
                    <For each={explorerPlaylists()}>
                      {(playlist) => (
                        <button
                          type="button"
                          onClick={() => setNavPath({ playlistKey: playlist.key, playlistKind: playlist.kind })}
                          class="group relative overflow-hidden rounded-[2rem] border border-white/5 bg-black/40 text-left transition-smooth hover:border-accent-primary/50 hover:shadow-vibrant"
                        >
                          <div class="relative">
                            <Thumbnail
                              src={playlist.thumbnailUrl}
                              alt={playlist.name}
                              size="md"
                              class="!rounded-none"
                            />
                            <div class={`absolute left-4 top-4 rounded-full border px-3 py-1 text-[10px] font-black uppercase tracking-widest backdrop-blur-md ${playlist.kind === 'source'
                              ? 'border-emerald-500/30 bg-emerald-500/20 text-emerald-400'
                              : 'border-accent-secondary/30 bg-accent-secondary/20 text-accent-secondary'
                              }`}
                            >
                              {playlist.kind}
                            </div>
                          </div>
                          <div class="space-y-1 p-5">
                            <div class="truncate text-base font-black text-white">{playlist.name}</div>
                            <div class="text-xs font-bold text-gray-500">{playlist.count} item{playlist.count === 1 ? '' : 's'}</div>
                          </div>
                        </button>
                      )}
                    </For>
                  </Grid>
                </Show>
              </Show>
            </Show>
          )}
        >
          <div class="space-y-10">
            <Show when={explorerLandingArtists().length > 0}>
              <div class="space-y-6">
                <div class="text-xs font-black uppercase tracking-[0.25em] text-accent-primary/80 ml-2">Artists</div>
                <Grid class="!p-0 !gap-6 sm:grid-cols-2 xl:grid-cols-4">
                  <For each={explorerLandingArtists()}>
                    {(artist) => (
                      <button
                        type="button"
                        onClick={() => setNavPath({ creatorType: 'artist', creatorName: artist.name })}
                        class="group relative overflow-hidden rounded-[2rem] border border-white/5 bg-black/40 text-left transition-smooth hover:border-accent-primary/50 hover:shadow-vibrant"
                      >
                        <Thumbnail
                          src={artist.thumbnailUrl}
                          alt={artist.name}
                          size="md"
                          class="!rounded-none"
                        />
                        <div class="space-y-1 p-5">
                          <div class="truncate text-base font-black text-white">{artist.name}</div>
                          <div class="text-xs font-bold text-gray-500">{artist.count} track{artist.count === 1 ? '' : 's'}</div>
                        </div>
                      </button>
                    )}
                  </For>
                </Grid>
              </div>
            </Show>



            <Show when={model().podcasts.length > 0}>
              <div class="space-y-6">
                <div class="text-xs font-black uppercase tracking-[0.25em] text-emerald-500/80 ml-2">Podcasts</div>
                <Grid class="!p-0 !gap-6 sm:grid-cols-2 xl:grid-cols-4">
                  <For each={model().podcasts}>
                    {(podcast) => (
                      <button
                        type="button"
                        onClick={() => {
                          setSection('podcasts');
                          setNavPath({ creatorType: 'podcast', creatorName: podcast.name });
                        }}
                        class="group relative overflow-hidden rounded-[2rem] border border-white/5 bg-black/40 text-left transition-smooth hover:border-emerald-500/50 hover:shadow-vibrant"
                      >
                        <Thumbnail
                          src={podcast.thumbnailUrl}
                          alt={podcast.name}
                          size="md"
                          class="!rounded-none"
                        />
                        <div class="space-y-1 p-5">
                          <div class="truncate text-base font-black text-white">{podcast.name}</div>
                          <div class="text-xs font-bold text-gray-500">{podcast.count} episode{podcast.count === 1 ? '' : 's'}</div>
                        </div>
                      </button>
                    )}
                  </For>
                </Grid>
              </div>
            </Show>
          </div>
        </Show>
      </Show>

      <Show when={viewMode() === 'columns'}>
        <Show
          when={explorer().kind === 'items'}
          fallback={(
            <div class="overflow-hidden rounded-3xl border border-white/5 bg-black/20">
              <table class="w-full text-sm">
                <thead class="bg-white/5 text-[10px] font-black uppercase tracking-[0.2em] text-gray-500">
                  <tr>
                    <th class="px-6 py-4 text-left">Name</th>
                    <th class="px-6 py-4 text-left">Type</th>
                    <th class="px-6 py-4 text-left">Count</th>
                    <th class="px-6 py-4 text-left">Latest</th>
                    <th class="px-6 py-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                  <For each={creatorDetailRows()}>
                    {(row) => (
                      <tr class="text-white hover:bg-white/2 transition-colors">
                        <td class="px-6 py-4">
                          <div class="font-bold">{row.name}</div>
                          <Show when={row.subtitle !== ''}>
                            <div class="text-xs font-medium text-gray-500">{row.subtitle}</div>
                          </Show>
                        </td>
                        <td class="px-6 py-4 text-xs font-bold text-accent-primary/80 uppercase tracking-widest">{row.type}</td>
                        <td class="px-6 py-4 text-xs font-bold text-gray-400">{row.count}</td>
                        <td class="px-6 py-4 text-xs font-medium text-gray-500">{row.latestTimestamp > 0 ? new Date(row.latestTimestamp).toLocaleDateString() : '-'}</td>
                        <td class="px-6 py-4 text-right">
                          <button
                            type="button"
                            onClick={row.onOpen}
                            class="rounded-xl border border-accent-primary/40 bg-accent-primary/20 px-4 py-2 text-xs font-black uppercase tracking-widest text-white hover:bg-accent-primary/30 transition-smooth"
                          >
                            Open
                          </button>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          )}
        >
          <Show
            when={explorerItems().length > 0}
            fallback={<div class="rounded-3xl border border-white/5 bg-black/20 p-12 text-center text-sm text-gray-500 font-medium">No media items match this view yet.</div>}
          >
            <div class="overflow-x-auto rounded-3xl border border-white/5 bg-black/20">
              <table class="w-full min-w-[980px] text-sm">
                <thead class="bg-white/5 text-[10px] font-black uppercase tracking-[0.2em] text-gray-500">
                  <tr>
                    <th class="px-6 py-4 text-left">Media</th>
                    <th class="px-6 py-4 text-left">Creator</th>
                    <th class="px-6 py-4 text-left">Collection</th>
                    <th class="px-6 py-4 text-left">Playlists</th>
                    <th class="px-6 py-4 text-left">Type</th>
                    <th class="px-6 py-4 text-left">Date</th>
                    <th class="px-6 py-4 text-left">Assignment</th>
                    <th class="px-6 py-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                  <For each={explorerItems()}>
                    {(item) => (
                      <tr class="align-top text-white hover:bg-white/2 transition-colors">
                        <td class="px-6 py-4">
                          <div class="flex items-start gap-4">
                            <Thumbnail src={item.thumbnailUrl} alt={item.title} size="sm" class="flex-shrink-0" />
                            <div class="min-w-0">
                              <div class="truncate font-bold text-base">{item.title}</div>
                              <div class="truncate text-xs font-medium text-gray-500">{item.size || 'Unknown size'} • {item.filename}</div>
                            </div>
                          </div>
                        </td>
                        <td class="px-6 py-4 text-xs font-bold text-accent-primary/80">{item.creator}</td>
                        <td class="px-6 py-4 text-xs font-bold text-gray-400">{item.album}</td>
                        <td class="px-6 py-4 text-xs font-medium text-gray-500">
                          <div>{item.sourcePlaylist}</div>
                          <div class="text-accent-secondary/60">Saved: {item.savedPlaylistName}</div>
                        </td>
                        <td class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-gray-500">{item.type}</td>
                        <td class="px-6 py-4 text-xs font-medium text-gray-500">{firstNonEmpty(item.date, item.modifiedAt)}</td>
                        <td class="px-6 py-4">
                          <select
                            value={item.savedPlaylistId}
                            onChange={(event) => handleAssignSavedPlaylist(item, event.currentTarget.value)}
                            class="w-full rounded-xl border border-white/10 bg-black/40 px-3 py-2 text-[10px] font-black uppercase tracking-widest text-white outline-none focus:border-accent-primary/50 transition-smooth"
                          >
                            <option value="">Unassigned</option>
                            <For each={savedPlaylists()}>
                              {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                            </For>
                          </select>
                        </td>
                        <td class="px-6 py-4 text-right">
                          <button
                            type="button"
                            onClick={() => handlePlayItem(item, explorerQueueItems())}
                            class="rounded-xl bg-vibrant-gradient px-4 py-2 text-xs font-black uppercase tracking-widest text-white hover:scale-105 transition-smooth shadow-vibrant"
                          >
                            Play
                          </button>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>
        </Show>
      </Show>

      <Show when={viewMode() === 'columns'}>
        <Show
          when={explorer().kind === 'items'}
          fallback={(
            <div class="overflow-hidden rounded-3xl border border-white/5 bg-black/20">
              <table class="w-full text-sm">
                <thead class="bg-white/5 text-[10px] font-black uppercase tracking-[0.2em] text-gray-500">
                  <tr>
                    <th class="px-6 py-4 text-left">Name</th>
                    <th class="px-6 py-4 text-left">Type</th>
                    <th class="px-6 py-4 text-left">Items</th>
                    <th class="px-6 py-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                  <For each={creatorDetailRows()}>
                    {(row) => (
                      <tr class="text-white hover:bg-white/2 transition-colors">
                        <td class="px-6 py-4">
                          <div class="flex items-center gap-4">
                            <Thumbnail src={row.thumbnailUrl} alt={row.name} size="sm" class="h-10 w-10 !rounded-full flex-shrink-0" />
                            <div>
                              <div class="font-bold">{row.name}</div>
                              <Show when={row.subtitle !== ''}>
                                <div class="text-xs font-medium text-gray-500">{row.subtitle}</div>
                              </Show>
                            </div>
                          </div>
                        </td>
                        <td class="px-6 py-4 text-xs font-bold text-accent-primary/80 uppercase tracking-widest">{row.type}</td>
                        <td class="px-6 py-4 text-xs font-bold text-gray-400">{row.count}</td>
                        <td class="px-6 py-4 text-right">
                          <button
                            type="button"
                            onClick={row.onOpen}
                            class="rounded-xl border border-accent-primary/40 bg-accent-primary/20 px-4 py-2 text-xs font-black uppercase tracking-widest text-white hover:bg-accent-primary/30 transition-smooth"
                          >
                            Open
                          </button>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          )}
        >
          <Show
            when={explorerItems().length > 0}
            fallback={<div class="rounded-3xl border border-white/5 bg-black/20 p-12 text-center text-sm text-gray-500 font-medium">No media items match this view yet.</div>}
          >
            <div class="overflow-x-auto rounded-3xl border border-white/5 bg-black/20">
              <table class="w-full min-w-[980px] text-sm">
                <thead class="bg-white/5 text-[10px] font-black uppercase tracking-[0.2em] text-gray-500">
                  <tr>
                    <th class="px-6 py-4 text-left">Media</th>
                    <th class="px-6 py-4 text-left">Creator</th>
                    <th class="px-6 py-4 text-left">Collection</th>
                    <th class="px-6 py-4 text-left">Type</th>
                    <th class="px-6 py-4 text-left">Date</th>
                    <th class="px-6 py-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                  <For each={explorerItems()}>
                    {(item) => (
                      <tr class="align-top text-white hover:bg-white/2 transition-colors">
                        <td class="px-6 py-4">
                          <div class="flex items-start gap-4">
                            <Thumbnail src={item.thumbnailUrl} alt={item.title} size="sm" class="flex-shrink-0 !w-20" />
                            <div class="min-w-0">
                              <div class="truncate font-bold text-base">{item.title}</div>
                              <div class="truncate text-xs font-medium text-gray-500">{item.size || 'Unknown size'}</div>
                            </div>
                          </div>
                        </td>
                        <td class="px-6 py-4 text-xs font-bold text-accent-primary/80">{item.creator}</td>
                        <td class="px-6 py-4 text-xs font-bold text-gray-400">{item.album}</td>
                        <td class="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-gray-500">{item.type}</td>
                        <td class="px-6 py-4 text-xs font-medium text-gray-500">{firstNonEmpty(item.date, item.modifiedAt)}</td>
                        <td class="px-6 py-4 text-right">
                          <button
                            type="button"
                            onClick={() => handlePlayItem(item, explorerQueueItems())}
                            class="rounded-xl bg-accent-primary px-4 py-2 text-xs font-black uppercase tracking-widest text-white hover:scale-105 transition-smooth"
                          >
                            Play
                          </button>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>
        </Show>
      </Show>

      <Show when={viewMode() === 'list'}>
        <Show
          when={explorer().kind === 'items'}
          fallback={(
            <div class="space-y-2">
              <For each={creatorDetailRows()}>
                {(row) => (
                  <button
                    type="button"
                    onClick={row.onOpen}
                    class="w-full flex items-center justify-between gap-6 p-4 rounded-2xl border border-white/5 bg-black/20 hover:border-accent-primary/30 hover:bg-black/40 transition-smooth group"
                  >
                    <div class="flex items-center gap-4 min-w-0">
                      <Thumbnail src={row.thumbnailUrl} alt={row.name} size="sm" class="h-12 w-12 !rounded-full flex-shrink-0" />
                      <div class="text-left min-w-0">
                        <div class="font-black text-white truncate">{row.name}</div>
                        <div class="text-[10px] font-bold text-accent-primary/60 uppercase tracking-widest">{row.type}</div>
                      </div>
                    </div>
                    <div class="flex items-center gap-8 shrink-0">
                      <div class="text-right">
                        <div class="text-xs font-bold text-white">{row.count}</div>
                        <div class="text-[10px] font-bold text-gray-600 uppercase tracking-widest">Items</div>
                      </div>
                      <Icon name="chevron-right" class="w-5 h-5 text-gray-700 group-hover:text-white transition-smooth" />
                    </div>
                  </button>
                )}
              </For>
            </div>
          )}
        >
          <div class="overflow-hidden rounded-3xl border border-white/5 bg-black/20">
            <table class="w-full text-sm">
              <thead class="bg-white/5 text-[10px] font-black uppercase tracking-[0.2em] text-gray-500">
                <tr>
                  <th class="px-6 py-4 text-left">Track</th>
                  <th class="px-6 py-4 text-left">Artist</th>
                  <th class="px-6 py-4 text-left">Album</th>
                  <th class="px-6 py-4 text-right">Length</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5">
                <For each={explorerItems()}>
                  {(item) => (
                    <tr
                      class="text-white hover:bg-accent-primary/10 cursor-pointer transition-colors group"
                      onClick={() => handlePlayItem(item, explorerQueueItems())}
                    >
                      <td class="px-6 py-4">
                        <div class="flex items-center gap-4">
                          <div class="w-8 text-[10px] font-black text-gray-600 group-hover:text-accent-primary transition-colors">
                            <Icon name="play" class="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                            <span class="group-hover:hidden">01</span>
                          </div>
                          <div class="font-bold truncate">{item.title}</div>
                        </div>
                      </td>
                      <td class="px-6 py-4 text-xs font-bold text-gray-400">{item.creator}</td>
                      <td class="px-6 py-4 text-xs font-bold text-gray-500">{item.album}</td>
                      <td class="px-6 py-4 text-right text-xs font-mono text-gray-500">{item.size || '--:--'}</td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </Show>
      </Show>

      <Show when={showSavedPlaylistManager()}>
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4 animate-in fade-in duration-300">
          <div class="w-full max-w-xl rounded-[2rem] border border-white/10 bg-bg-surface p-8 shadow-2xl relative overflow-hidden">
            <div class="absolute top-0 right-0 p-12 opacity-5 pointer-events-none">
              <Icon name="layers" class="w-32 h-32 rotate-12" />
            </div>

            <div class="relative space-y-6">
              <div class="flex items-center justify-between">
                <h3 class="text-2xl font-black text-white">Saved Playlists</h3>
                <button
                  type="button"
                  onClick={() => setShowSavedPlaylistManager(false)}
                  class="rounded-xl border border-white/10 bg-white/5 p-2 text-gray-400 hover:text-white hover:bg-white/10 transition-smooth"
                >
                  <Icon name="x" class="h-6 w-6" />
                </button>
              </div>

              <div class="flex gap-3">
                <input
                  type="text"
                  value={newSavedPlaylistName()}
                  onInput={(event) => setNewSavedPlaylistName(event.currentTarget.value)}
                  onKeyDown={(event) => {
                    if (event.key === 'Enter') {
                      event.preventDefault();
                      void handleCreateSavedPlaylist();
                    }
                  }}
                  placeholder="Playlist name..."
                  class="w-full rounded-xl border border-white/10 bg-black/40 px-4 py-3 text-sm text-white outline-none focus:border-accent-primary/50 transition-smooth"
                />
                <button
                  type="button"
                  onClick={() => { void handleCreateSavedPlaylist(); }}
                  class="rounded-xl bg-accent-primary px-6 py-3 text-sm font-black uppercase tracking-widest text-white hover:scale-105 transition-smooth shadow-vibrant"
                >
                  Create
                </button>
              </div>

              <Show when={playlistMessage() !== ''}>
                <div class={`text-xs font-bold px-4 py-2 rounded-lg ${playlistMessageTone() === 'error'
                  ? 'bg-red-500/10 text-red-400 border border-red-500/20'
                  : playlistMessageTone() === 'success'
                    ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                    : 'bg-white/5 text-gray-400 border border-white/10'
                  }`}
                >
                  {playlistMessage()}
                </div>
              </Show>

              <div class="max-h-80 space-y-3 overflow-y-auto pr-2 custom-scrollbar">
                <Show
                  when={savedPlaylists().length > 0}
                  fallback={<div class="rounded-2xl border border-white/5 bg-white/2 p-6 text-center text-xs font-medium text-gray-500">No saved playlists yet.</div>}
                >
                  <For each={savedPlaylists()}>
                    {(playlist) => (
                      <div class="flex items-center justify-between gap-4 rounded-2xl border border-white/5 bg-white/5 px-4 py-3 group/item hover:border-white/10 transition-smooth">
                        <div class="min-w-0">
                          <div class="truncate text-sm font-bold text-white">{playlist.name}</div>
                          <div class="text-[10px] font-bold text-gray-600 uppercase tracking-widest">ID: {playlist.id.slice(0, 8)}...</div>
                        </div>
                        <div class="flex items-center gap-2 opacity-0 group-hover/item:opacity-100 transition-smooth">
                          <button
                            type="button"
                            onClick={() => { void handleRenameSavedPlaylist(playlist); }}
                            class="rounded-lg border border-white/10 px-3 py-1.5 text-[10px] font-black uppercase tracking-widest text-gray-400 hover:text-white hover:bg-white/10 transition-smooth"
                          >
                            Rename
                          </button>
                          <button
                            type="button"
                            onClick={() => { void handleDeleteSavedPlaylist(playlist); }}
                            class="rounded-lg border border-red-500/20 px-3 py-1.5 text-[10px] font-black uppercase tracking-widest text-red-400 hover:bg-red-500/10 transition-smooth"
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    )}
                  </For>
                </Show>
              </div>
            </div>
          </div>
        </div>
      </Show>
    </div>
  );
}
