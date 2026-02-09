import { For, Show, createEffect, createMemo, createSignal } from 'solid-js';
import Icon from './Icon';
import { buildLibraryModel } from '../utils/libraryModel';

const SECTION_OPTIONS = [
  { value: 'artists', label: 'Artists' },
  { value: 'channels', label: 'Channels' },
  { value: 'playlists', label: 'Playlists' },
  { value: 'all_media', label: 'All Media' },
];

const VIEW_MODE_OPTIONS = [
  { value: 'gallery', label: 'Gallery' },
  { value: 'list', label: 'List' },
  { value: 'detail', label: 'Detail' },
];

const TYPE_FILTER_OPTIONS = [
  { value: 'all', label: 'All Types' },
  { value: 'audio', label: 'Audio' },
  { value: 'video', label: 'Video' },
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
  const [failedThumbByKey, setFailedThumbByKey] = createSignal({});
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

  const markThumbFailed = (key, src) => {
    setFailedThumbByKey((current) => ({ ...current, [key]: src }));
  };

  const canRenderThumb = (key, src) => (
    typeof src === 'string'
    && src.trim() !== ''
    && failedThumbByKey()[key] !== src
  );

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
            title: 'Artists & Channels',
            subtitle: 'Browse creators by recent activity.',
            artists: libraryModel.artists,
            channels: libraryModel.channels,
            breadcrumbs: [],
          };
        }

        if (currentNav.albumName !== '') {
          const album = creator.albums.find((entry) => entry.name === currentNav.albumName);
          if (album) {
            return {
              kind: 'items',
              title: album.name,
              subtitle: `${creator.name} • ${album.count} item${album.count === 1 ? '' : 's'}`,
              items: album.items,
              queueItems: album.items,
              breadcrumbs: [
                { label: 'Artists', nav: { ...EMPTY_NAV_PATH } },
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
            { label: 'Artists', nav: { ...EMPTY_NAV_PATH } },
            { label: creator.name, nav: { creatorType: 'artist', creatorName: creator.name } },
          ],
        };
      }

      return {
        kind: 'landing',
        title: 'Artists & Channels',
        subtitle: 'Rows of creator cards sorted by recent downloads.',
        artists: libraryModel.artists,
        channels: libraryModel.channels,
        breadcrumbs: [],
      };
    }

    if (currentSection === 'channels') {
      if (currentNav.creatorType === 'channel' && currentNav.creatorName !== '') {
        const channel = libraryModel.channelsByName.get(currentNav.creatorName);
        if (channel) {
          return {
            kind: 'items',
            title: channel.name,
            subtitle: `${channel.count} video${channel.count === 1 ? '' : 's'}`,
            items: channel.items,
            queueItems: channel.items,
            breadcrumbs: [
              { label: 'Channels', nav: { ...EMPTY_NAV_PATH } },
              { label: channel.name, nav: { creatorType: 'channel', creatorName: channel.name } },
            ],
          };
        }
      }

      return {
        kind: 'creators',
        title: 'Channels',
        subtitle: 'Browse downloaded videos by channel.',
        creators: libraryModel.channels,
        creatorType: 'channel',
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
        type: 'Artist',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: `${entry.albums.length} album${entry.albums.length === 1 ? '' : 's'}`,
        onOpen: () => setNavPath({ creatorType: 'artist', creatorName: entry.name }),
      }));
      const channels = context.channels.map((entry) => ({
        key: `channel:${entry.name}`,
        name: entry.name,
        type: 'Channel',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: 'Video library',
        onOpen: () => {
          setSection('channels');
          setNavPath({ creatorType: 'channel', creatorName: entry.name });
        },
      }));
      return [...artists, ...channels].sort((left, right) => right.latestTimestamp - left.latestTimestamp);
    }

    if (context.kind === 'creators') {
      return context.creators.map((entry) => ({
        key: `${context.creatorType}:${entry.name}`,
        name: entry.name,
        type: context.creatorType === 'channel' ? 'Channel' : 'Artist',
        count: entry.count,
        latestTimestamp: entry.latestTimestamp,
        thumbnailUrl: entry.thumbnailUrl,
        subtitle: context.creatorType === 'channel'
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
  const explorerLandingChannels = createMemo(() => (explorer().kind === 'landing' ? explorer().channels : []));

  createEffect(() => {
    const context = explorer();
    const currentSelected = selectedDetailKey();

    if (viewMode() !== 'detail') {
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
      return `${context.artists.length + context.channels.length} creators`;
    }
    return '';
  });

  return (
    <div class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-500">
      <div class="relative overflow-hidden rounded-3xl border border-cyan-500/10 bg-[radial-gradient(circle_at_20%_20%,rgba(56,189,248,0.12),transparent_52%),radial-gradient(circle_at_80%_0%,rgba(20,184,166,0.1),transparent_40%),linear-gradient(180deg,#090c12,#070a10)] p-6 shadow-[0_0_0_1px_rgba(15,23,42,0.45),0_20px_80px_rgba(6,10,18,0.6)]">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start">
          <aside class="w-full lg:w-56 shrink-0 rounded-2xl border border-white/10 bg-black/20 p-3">
            <div class="text-[10px] font-semibold uppercase tracking-[0.14em] text-cyan-200/80">Library Sections</div>
            <div class="mt-2 space-y-1">
              <For each={SECTION_OPTIONS}>
                {(option) => (
                  <button
                    type="button"
                    onClick={() => setSection(option.value)}
                    class={`w-full rounded-xl px-3 py-2 text-left text-sm font-semibold transition-colors ${
                      section() === option.value
                        ? 'bg-cyan-500/20 text-cyan-100 border border-cyan-400/40'
                        : 'text-slate-300 hover:text-white hover:bg-white/10 border border-transparent'
                    }`}
                  >
                    {option.label}
                  </button>
                )}
              </For>
            </div>
          </aside>

          <div class="min-w-0 flex-1 space-y-4">
            <div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
              <div>
                <h1 class="text-3xl font-black tracking-tight text-white">Your Library</h1>
                <p class="text-sm text-slate-400">{explorer().subtitle}</p>
              </div>
              <div class="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-3 py-1 text-xs font-semibold uppercase tracking-[0.12em] text-cyan-100">
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

            <div class="flex flex-wrap items-center gap-2 rounded-2xl border border-white/10 bg-black/20 p-3">
              <div class="inline-flex rounded-xl border border-white/10 bg-white/5 p-1">
                <For each={VIEW_MODE_OPTIONS}>
                  {(option) => (
                    <button
                      type="button"
                      onClick={() => setViewMode(option.value)}
                      class={`rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors ${
                        viewMode() === option.value
                          ? 'bg-cyan-500/30 text-cyan-100'
                          : 'text-slate-300 hover:text-white hover:bg-white/10'
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
                class="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40"
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
                class="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40"
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
                class={`rounded-xl border px-3 py-2 text-xs font-semibold transition-colors ${
                  uiState().advancedFiltersOpen
                    ? 'border-cyan-400/40 bg-cyan-500/20 text-cyan-100'
                    : 'border-white/10 bg-white/5 text-slate-200 hover:bg-white/10'
                }`}
              >
                Advanced Filters
              </button>

              <Show when={section() === 'playlists'}>
                <button
                  type="button"
                  onClick={() => setShowSavedPlaylistManager(true)}
                  class="rounded-xl border border-cyan-400/30 bg-cyan-500/10 px-3 py-2 text-xs font-semibold text-cyan-100 hover:bg-cyan-500/20 transition-colors"
                >
                  Manage Saved Playlists
                </button>
              </Show>

              <button
                type="button"
                onClick={clearFilters}
                disabled={!hasActiveFilters()}
                class={`ml-auto rounded-xl border px-3 py-2 text-xs font-semibold transition-colors ${
                  hasActiveFilters()
                    ? 'border-white/20 bg-white/5 text-slate-100 hover:bg-white/10'
                    : 'border-white/10 bg-white/5 text-slate-500 cursor-not-allowed'
                }`}
              >
                Clear Filters
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
                  fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-8 text-center text-sm text-slate-400">No media items match this view yet.</div>}
                >
                  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                    <For each={explorerItems()}>
                      {(item) => {
                        const thumbKey = `media:${item.mediaKey}`;
                        return (
                          <div class="group overflow-hidden rounded-2xl border border-white/10 bg-black/25 transition-colors hover:border-cyan-400/40">
                            <div class="relative aspect-video overflow-hidden bg-slate-900">
                              <Show
                                when={canRenderThumb(thumbKey, item.thumbnailUrl)}
                                fallback={<div class="flex h-full items-center justify-center"><Icon name={item.type === 'audio' ? 'music' : 'film'} class="h-8 w-8 text-slate-500" /></div>}
                              >
                                <img
                                  src={item.thumbnailUrl}
                                  alt={item.title}
                                  loading="lazy"
                                  class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
                                  onError={() => markThumbFailed(thumbKey, item.thumbnailUrl)}
                                />
                              </Show>
                              <div class="absolute left-2 top-2 rounded-full border border-black/30 bg-black/45 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-white/90">
                                {item.type}
                              </div>
                            </div>
                            <div class="space-y-2 p-3">
                              <div class="truncate text-sm font-bold text-white">{item.title}</div>
                              <div class="truncate text-xs text-slate-400">{item.creator}</div>
                              <div class="truncate text-[11px] text-slate-500">{item.album} • {item.date || item.modifiedAt}</div>
                              <div class="flex items-center gap-2">
                                <button
                                  type="button"
                                  onClick={() => handlePlayItem(item, explorerQueueItems())}
                                  class="rounded-lg bg-cyan-500 px-2 py-1 text-xs font-semibold text-slate-950 hover:bg-cyan-400 transition-colors"
                                >
                                  Play
                                </button>
                                <label class="min-w-0 flex-1 text-[10px] font-semibold uppercase tracking-wide text-slate-400">
                                  Saved
                                  <select
                                    value={item.savedPlaylistId}
                                    onChange={(event) => handleAssignSavedPlaylist(item, event.currentTarget.value)}
                                    class="mt-1 w-full truncate rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-xs text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                                  >
                                    <option value="">Unassigned</option>
                                    <For each={savedPlaylists()}>
                                      {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                                    </For>
                                  </select>
                                </label>
                              </div>
                            </div>
                          </div>
                        );
                      }}
                    </For>
                  </div>
                </Show>
              )}
            >
              <Show when={explorer().kind === 'creators'}>
                <Show
                  when={explorerCreators().length > 0}
                  fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-8 text-center text-sm text-slate-400">No creators available for this filter set.</div>}
                >
                  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                    <For each={explorerCreators()}>
                      {(entry) => {
                        const thumbKey = `creator:${explorer().creatorType}:${entry.name}`;
                        return (
                          <button
                            type="button"
                            onClick={() => setNavPath({ creatorType: explorer().creatorType, creatorName: entry.name })}
                            class="group overflow-hidden rounded-2xl border border-white/10 bg-black/25 text-left transition-colors hover:border-cyan-400/40"
                          >
                            <div class="relative aspect-video bg-slate-900">
                              <Show
                                when={canRenderThumb(thumbKey, entry.thumbnailUrl)}
                                fallback={<div class="flex h-full items-center justify-center"><Icon name={explorer().creatorType === 'channel' ? 'film' : 'music'} class="h-8 w-8 text-slate-500" /></div>}
                              >
                                <img
                                  src={entry.thumbnailUrl}
                                  alt={entry.name}
                                  loading="lazy"
                                  class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
                                  onError={() => markThumbFailed(thumbKey, entry.thumbnailUrl)}
                                />
                              </Show>
                            </div>
                            <div class="space-y-1 p-3">
                              <div class="truncate text-sm font-bold text-white">{entry.name}</div>
                              <div class="text-xs text-slate-400">{entry.count} item{entry.count === 1 ? '' : 's'}</div>
                            </div>
                          </button>
                        );
                      }}
                    </For>
                  </div>
                </Show>
              </Show>

              <Show when={explorer().kind === 'albums'}>
                <Show
                  when={explorerAlbums().length > 0}
                  fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-8 text-center text-sm text-slate-400">No albums available for this artist.</div>}
                >
                  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                    <For each={explorerAlbums()}>
                      {(album) => {
                        const thumbKey = `album:${explorer().creator.name}:${album.name}`;
                        return (
                          <button
                            type="button"
                            onClick={() => setNavPath({ creatorType: 'artist', creatorName: explorer().creator.name, albumName: album.name })}
                            class="group overflow-hidden rounded-2xl border border-white/10 bg-black/25 text-left transition-colors hover:border-cyan-400/40"
                          >
                            <div class="relative aspect-square bg-slate-900">
                              <Show
                                when={canRenderThumb(thumbKey, album.thumbnailUrl)}
                                fallback={<div class="flex h-full items-center justify-center"><Icon name="music" class="h-8 w-8 text-slate-500" /></div>}
                              >
                                <img
                                  src={album.thumbnailUrl}
                                  alt={album.name}
                                  loading="lazy"
                                  class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
                                  onError={() => markThumbFailed(thumbKey, album.thumbnailUrl)}
                                />
                              </Show>
                            </div>
                            <div class="space-y-1 p-3">
                              <div class="truncate text-sm font-bold text-white">{album.name}</div>
                              <div class="text-xs text-slate-400">{album.count} track{album.count === 1 ? '' : 's'}</div>
                            </div>
                          </button>
                        );
                      }}
                    </For>
                  </div>
                </Show>
              </Show>

              <Show when={explorer().kind === 'playlists'}>
                <Show
                  when={explorerPlaylists().length > 0}
                  fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-8 text-center text-sm text-slate-400">No playlists available for this filter set.</div>}
                >
                  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                    <For each={explorerPlaylists()}>
                      {(playlist) => {
                        const thumbKey = `playlist:${playlist.key}`;
                        return (
                          <button
                            type="button"
                            onClick={() => setNavPath({ playlistKey: playlist.key, playlistKind: playlist.kind })}
                            class="group overflow-hidden rounded-2xl border border-white/10 bg-black/25 text-left transition-colors hover:border-cyan-400/40"
                          >
                            <div class="relative aspect-video bg-slate-900">
                              <Show
                                when={canRenderThumb(thumbKey, playlist.thumbnailUrl)}
                                fallback={<div class="flex h-full items-center justify-center"><Icon name="layers" class="h-8 w-8 text-slate-500" /></div>}
                              >
                                <img
                                  src={playlist.thumbnailUrl}
                                  alt={playlist.name}
                                  loading="lazy"
                                  class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
                                  onError={() => markThumbFailed(thumbKey, playlist.thumbnailUrl)}
                                />
                              </Show>
                              <div class={`absolute left-2 top-2 rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${
                                playlist.kind === 'source'
                                  ? 'border-emerald-300/50 bg-emerald-500/20 text-emerald-100'
                                  : 'border-cyan-300/50 bg-cyan-500/20 text-cyan-100'
                              }`}
                              >
                                {playlist.kind}
                              </div>
                            </div>
                            <div class="space-y-1 p-3">
                              <div class="truncate text-sm font-bold text-white">{playlist.name}</div>
                              <div class="text-xs text-slate-400">{playlist.count} item{playlist.count === 1 ? '' : 's'}</div>
                            </div>
                          </button>
                        );
                      }}
                    </For>
                  </div>
                </Show>
              </Show>
            </Show>
          )}
        >
          <div class="space-y-6">
            <Show
              when={explorerLandingArtists().length > 0}
              fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-6 text-sm text-slate-400">No artist groups available for this filter set.</div>}
            >
              <div>
                <div class="mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-cyan-200/80">Artists</div>
                <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                  <For each={explorerLandingArtists()}>
                    {(artist) => {
                      const thumbKey = `artist:${artist.name}`;
                      return (
                        <button
                          type="button"
                          onClick={() => setNavPath({ creatorType: 'artist', creatorName: artist.name })}
                          class="group overflow-hidden rounded-2xl border border-white/10 bg-black/25 text-left transition-colors hover:border-cyan-400/40"
                        >
                          <div class="relative aspect-video bg-slate-900">
                            <Show
                              when={canRenderThumb(thumbKey, artist.thumbnailUrl)}
                              fallback={<div class="flex h-full items-center justify-center"><Icon name="music" class="h-8 w-8 text-slate-500" /></div>}
                            >
                              <img
                                src={artist.thumbnailUrl}
                                alt={artist.name}
                                loading="lazy"
                                class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
                                onError={() => markThumbFailed(thumbKey, artist.thumbnailUrl)}
                              />
                            </Show>
                          </div>
                          <div class="space-y-1 p-3">
                            <div class="truncate text-sm font-bold text-white">{artist.name}</div>
                            <div class="text-xs text-slate-400">{artist.count} track{artist.count === 1 ? '' : 's'}</div>
                          </div>
                        </button>
                      );
                    }}
                  </For>
                </div>
              </div>
            </Show>

            <Show
              when={explorerLandingChannels().length > 0}
              fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-6 text-sm text-slate-400">No channel groups available for this filter set.</div>}
            >
              <div>
                <div class="mb-3 text-xs font-semibold uppercase tracking-[0.12em] text-cyan-200/80">Channels</div>
                <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                  <For each={explorerLandingChannels()}>
                    {(channel) => {
                      const thumbKey = `channel:${channel.name}`;
                      return (
                        <button
                          type="button"
                          onClick={() => {
                            setSection('channels');
                            setNavPath({ creatorType: 'channel', creatorName: channel.name });
                          }}
                          class="group overflow-hidden rounded-2xl border border-white/10 bg-black/25 text-left transition-colors hover:border-cyan-400/40"
                        >
                          <div class="relative aspect-video bg-slate-900">
                            <Show
                              when={canRenderThumb(thumbKey, channel.thumbnailUrl)}
                              fallback={<div class="flex h-full items-center justify-center"><Icon name="film" class="h-8 w-8 text-slate-500" /></div>}
                            >
                              <img
                                src={channel.thumbnailUrl}
                                alt={channel.name}
                                loading="lazy"
                                class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-[1.03]"
                                onError={() => markThumbFailed(thumbKey, channel.thumbnailUrl)}
                              />
                            </Show>
                          </div>
                          <div class="space-y-1 p-3">
                            <div class="truncate text-sm font-bold text-white">{channel.name}</div>
                            <div class="text-xs text-slate-400">{channel.count} video{channel.count === 1 ? '' : 's'}</div>
                          </div>
                        </button>
                      );
                    }}
                  </For>
                </div>
              </div>
            </Show>
          </div>
        </Show>
      </Show>

      <Show when={viewMode() === 'list'}>
        <Show
          when={explorer().kind === 'items'}
          fallback={(
            <div class="overflow-hidden rounded-2xl border border-white/10 bg-black/20">
              <table class="w-full text-sm">
                <thead class="bg-white/5 text-xs uppercase tracking-wide text-slate-400">
                  <tr>
                    <th class="px-4 py-3 text-left">Name</th>
                    <th class="px-4 py-3 text-left">Type</th>
                    <th class="px-4 py-3 text-left">Count</th>
                    <th class="px-4 py-3 text-left">Latest</th>
                    <th class="px-4 py-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={creatorDetailRows()}>
                    {(row) => (
                      <tr class="border-t border-white/5 text-slate-200">
                        <td class="px-4 py-3">
                          <div class="font-semibold text-white">{row.name}</div>
                          <Show when={row.subtitle !== ''}>
                            <div class="text-xs text-slate-500">{row.subtitle}</div>
                          </Show>
                        </td>
                        <td class="px-4 py-3 text-xs text-slate-400">{row.type}</td>
                        <td class="px-4 py-3 text-xs text-slate-400">{row.count}</td>
                        <td class="px-4 py-3 text-xs text-slate-400">{row.latestTimestamp > 0 ? new Date(row.latestTimestamp).toLocaleDateString() : '-'}</td>
                        <td class="px-4 py-3 text-right">
                          <button
                            type="button"
                            onClick={row.onOpen}
                            class="rounded-lg border border-cyan-400/40 bg-cyan-500/20 px-2 py-1 text-xs font-semibold text-cyan-100 hover:bg-cyan-500/30 transition-colors"
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
            fallback={<div class="rounded-2xl border border-white/10 bg-black/20 p-8 text-center text-sm text-slate-400">No media items match this view yet.</div>}
          >
            <div class="overflow-x-auto rounded-2xl border border-white/10 bg-black/20">
              <table class="w-full min-w-[980px] text-sm">
                <thead class="bg-white/5 text-xs uppercase tracking-wide text-slate-400">
                  <tr>
                    <th class="px-4 py-3 text-left">Media</th>
                    <th class="px-4 py-3 text-left">Creator</th>
                    <th class="px-4 py-3 text-left">Collection</th>
                    <th class="px-4 py-3 text-left">Playlists</th>
                    <th class="px-4 py-3 text-left">Type</th>
                    <th class="px-4 py-3 text-left">Date</th>
                    <th class="px-4 py-3 text-left">Saved Assignment</th>
                    <th class="px-4 py-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={explorerItems()}>
                    {(item) => {
                      const thumbKey = `list:${item.mediaKey}`;
                      return (
                        <tr class="border-t border-white/5 align-top text-slate-200">
                          <td class="px-4 py-3">
                            <div class="flex items-start gap-3">
                              <div class="h-14 w-24 overflow-hidden rounded-lg border border-white/10 bg-slate-900">
                                <Show
                                  when={canRenderThumb(thumbKey, item.thumbnailUrl)}
                                  fallback={<div class="flex h-full items-center justify-center"><Icon name={item.type === 'audio' ? 'music' : 'film'} class="h-5 w-5 text-slate-500" /></div>}
                                >
                                  <img
                                    src={item.thumbnailUrl}
                                    alt={item.title}
                                    loading="lazy"
                                    class="h-full w-full object-cover"
                                    onError={() => markThumbFailed(thumbKey, item.thumbnailUrl)}
                                  />
                                </Show>
                              </div>
                              <div class="min-w-0">
                                <div class="truncate font-semibold text-white">{item.title}</div>
                                <div class="truncate text-xs text-slate-500">{item.size || 'Unknown size'} • {item.filename}</div>
                              </div>
                            </div>
                          </td>
                          <td class="px-4 py-3 text-xs text-slate-300">{item.creator}</td>
                          <td class="px-4 py-3 text-xs text-slate-300">{item.album}</td>
                          <td class="px-4 py-3 text-xs text-slate-300">
                            <div>{item.sourcePlaylist}</div>
                            <div class="text-slate-500">Saved: {item.savedPlaylistName}</div>
                          </td>
                          <td class="px-4 py-3 text-xs uppercase text-slate-400">{item.type}</td>
                          <td class="px-4 py-3 text-xs text-slate-400">{firstNonEmpty(item.date, item.modifiedAt)}</td>
                          <td class="px-4 py-3">
                            <select
                              value={item.savedPlaylistId}
                              onChange={(event) => handleAssignSavedPlaylist(item, event.currentTarget.value)}
                              class="w-full rounded-lg border border-white/10 bg-white/5 px-2 py-1 text-xs text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40"
                            >
                              <option value="">Unassigned</option>
                              <For each={savedPlaylists()}>
                                {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                              </For>
                            </select>
                          </td>
                          <td class="px-4 py-3 text-right">
                            <button
                              type="button"
                              onClick={() => handlePlayItem(item, explorerQueueItems())}
                              class="rounded-lg bg-cyan-500 px-2 py-1 text-xs font-semibold text-slate-950 hover:bg-cyan-400 transition-colors"
                            >
                              Play
                            </button>
                          </td>
                        </tr>
                      );
                    }}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>
        </Show>
      </Show>

      <Show when={viewMode() === 'detail'}>
        <div class="grid gap-4 lg:grid-cols-[1fr_320px]">
          <Show
            when={explorer().kind === 'items'}
            fallback={(
              <div class="overflow-hidden rounded-2xl border border-white/10 bg-black/20">
                <table class="w-full text-sm">
                  <thead class="bg-white/5 text-xs uppercase tracking-wide text-slate-400">
                    <tr>
                      <th class="px-4 py-3 text-left">Name</th>
                      <th class="px-4 py-3 text-left">Type</th>
                      <th class="px-4 py-3 text-left">Count</th>
                    </tr>
                  </thead>
                  <tbody>
                    <For each={creatorDetailRows()}>
                      {(row) => (
                        <tr
                          class={`cursor-pointer border-t border-white/5 text-slate-200 transition-colors ${
                            selectedDetailKey() === row.key ? 'bg-cyan-500/10' : 'hover:bg-white/5'
                          }`}
                          onClick={() => setSelectedDetailKey(row.key)}
                        >
                          <td class="px-4 py-3">
                            <div class="font-semibold text-white">{row.name}</div>
                            <Show when={row.subtitle !== ''}>
                              <div class="text-xs text-slate-500">{row.subtitle}</div>
                            </Show>
                          </td>
                          <td class="px-4 py-3 text-xs text-slate-400">{row.type}</td>
                          <td class="px-4 py-3 text-xs text-slate-400">{row.count}</td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
            )}
          >
            <div class="overflow-hidden rounded-2xl border border-white/10 bg-black/20">
              <table class="w-full text-sm">
                <thead class="bg-white/5 text-xs uppercase tracking-wide text-slate-400">
                  <tr>
                    <th class="px-4 py-3 text-left">Title</th>
                    <th class="px-4 py-3 text-left">Creator</th>
                    <th class="px-4 py-3 text-left">Collection</th>
                    <th class="px-4 py-3 text-left">Type</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={explorerItems()}>
                    {(item) => (
                      <tr
                        class={`cursor-pointer border-t border-white/5 text-slate-200 transition-colors ${
                          selectedDetailKey() === item.mediaKey ? 'bg-cyan-500/10' : 'hover:bg-white/5'
                        }`}
                        onClick={() => setSelectedDetailKey(item.mediaKey)}
                      >
                        <td class="px-4 py-3">
                          <div class="font-semibold text-white">{item.title}</div>
                          <div class="text-xs text-slate-500">{item.filename}</div>
                        </td>
                        <td class="px-4 py-3 text-xs text-slate-300">{item.creator}</td>
                        <td class="px-4 py-3 text-xs text-slate-300">{item.album}</td>
                        <td class="px-4 py-3 text-xs uppercase text-slate-400">{item.type}</td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>

          <aside class="rounded-2xl border border-white/10 bg-black/20 p-4">
            <Show
              when={explorer().kind === 'items'}
              fallback={(
                <Show
                  when={selectedDetailRow()}
                  fallback={<div class="text-sm text-slate-500">Select a row to inspect details.</div>}
                >
                  <div class="space-y-3">
                    <div>
                      <div class="text-[11px] uppercase tracking-wide text-slate-500">Name</div>
                      <div class="text-sm font-semibold text-white">{selectedDetailRow()?.name}</div>
                    </div>
                    <div>
                      <div class="text-[11px] uppercase tracking-wide text-slate-500">Type</div>
                      <div class="text-xs text-slate-300">{selectedDetailRow()?.type}</div>
                    </div>
                    <div>
                      <div class="text-[11px] uppercase tracking-wide text-slate-500">Count</div>
                      <div class="text-xs text-slate-300">{selectedDetailRow()?.count}</div>
                    </div>
                    <button
                      type="button"
                      onClick={() => selectedDetailRow()?.onOpen?.()}
                      class="rounded-lg bg-cyan-500 px-3 py-1.5 text-xs font-semibold text-slate-950 hover:bg-cyan-400 transition-colors"
                    >
                      Open
                    </button>
                  </div>
                </Show>
              )}
            >
              <Show
                when={selectedDetailItem()}
                fallback={<div class="text-sm text-slate-500">Select an item to inspect metadata.</div>}
              >
                {(itemAccessor) => {
                  const item = itemAccessor();
                  const thumbKey = `detail:${item.mediaKey}`;
                  return (
                    <div class="space-y-3">
                      <div class="overflow-hidden rounded-xl border border-white/10 bg-slate-900">
                        <div class="aspect-video">
                          <Show
                            when={canRenderThumb(thumbKey, item.thumbnailUrl)}
                            fallback={<div class="flex h-full items-center justify-center"><Icon name={item.type === 'audio' ? 'music' : 'film'} class="h-8 w-8 text-slate-500" /></div>}
                          >
                            <img
                              src={item.thumbnailUrl}
                              alt={item.title}
                              loading="lazy"
                              class="h-full w-full object-cover"
                              onError={() => markThumbFailed(thumbKey, item.thumbnailUrl)}
                            />
                          </Show>
                        </div>
                      </div>
                      <div>
                        <div class="text-[11px] uppercase tracking-wide text-slate-500">Title</div>
                        <div class="text-sm font-semibold text-white">{item.title}</div>
                      </div>
                      <div class="grid grid-cols-2 gap-2 text-xs text-slate-300">
                        <div>
                          <div class="text-[11px] uppercase tracking-wide text-slate-500">Creator</div>
                          <div>{item.creator}</div>
                        </div>
                        <div>
                          <div class="text-[11px] uppercase tracking-wide text-slate-500">Collection</div>
                          <div>{item.album}</div>
                        </div>
                        <div>
                          <div class="text-[11px] uppercase tracking-wide text-slate-500">Type</div>
                          <div class="uppercase">{item.type}</div>
                        </div>
                        <div>
                          <div class="text-[11px] uppercase tracking-wide text-slate-500">Date</div>
                          <div>{firstNonEmpty(item.date, item.modifiedAt, '-')}</div>
                        </div>
                        <div class="col-span-2">
                          <div class="text-[11px] uppercase tracking-wide text-slate-500">Source Playlist</div>
                          <div>{item.sourcePlaylist}</div>
                        </div>
                      </div>

                      <label class="block text-[11px] font-semibold uppercase tracking-wide text-slate-500">
                        Saved Playlist
                        <select
                          value={item.savedPlaylistId}
                          onChange={(event) => handleAssignSavedPlaylist(item, event.currentTarget.value)}
                          class="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-2 py-1.5 text-xs text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40 normal-case"
                        >
                          <option value="">Unassigned</option>
                          <For each={savedPlaylists()}>
                            {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                          </For>
                        </select>
                      </label>

                      <div class="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handlePlayItem(item, explorerQueueItems())}
                          class="rounded-lg bg-cyan-500 px-3 py-1.5 text-xs font-semibold text-slate-950 hover:bg-cyan-400 transition-colors"
                        >
                          Play
                        </button>
                        <Show when={item.sourceURL !== ''}>
                          <a
                            href={item.sourceURL}
                            target="_blank"
                            rel="noreferrer"
                            class="rounded-lg border border-white/20 px-3 py-1.5 text-xs font-semibold text-slate-200 hover:bg-white/10 transition-colors"
                          >
                            Source
                          </a>
                        </Show>
                      </div>
                    </div>
                  );
                }}
              </Show>
            </Show>
          </aside>
        </div>
      </Show>

      <Show when={showSavedPlaylistManager()}>
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div class="w-full max-w-xl rounded-2xl border border-white/10 bg-[#0b0f16] p-5 shadow-2xl">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-white">Manage Saved Playlists</h3>
              <button
                type="button"
                onClick={() => setShowSavedPlaylistManager(false)}
                class="rounded-lg border border-white/15 bg-white/5 p-2 text-slate-300 hover:bg-white/10 transition-colors"
              >
                <Icon name="x" class="h-4 w-4" />
              </button>
            </div>

            <div class="mt-4 flex gap-2">
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
                placeholder="New saved playlist name"
                class="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/40"
              />
              <button
                type="button"
                onClick={() => { void handleCreateSavedPlaylist(); }}
                class="rounded-xl bg-cyan-500 px-4 py-2 text-sm font-semibold text-slate-950 hover:bg-cyan-400 transition-colors"
              >
                Create
              </button>
            </div>

            <Show when={playlistMessage() !== ''}>
              <div class={`mt-2 text-xs ${
                playlistMessageTone() === 'error'
                  ? 'text-red-300'
                  : playlistMessageTone() === 'success'
                    ? 'text-emerald-300'
                    : 'text-slate-400'
              }`}
              >
                {playlistMessage()}
              </div>
            </Show>

            <div class="mt-4 max-h-80 space-y-2 overflow-y-auto pr-1">
              <Show
                when={savedPlaylists().length > 0}
                fallback={<div class="rounded-xl border border-white/10 bg-white/5 p-3 text-xs text-slate-400">No saved playlists yet.</div>}
              >
                <For each={savedPlaylists()}>
                  {(playlist) => (
                    <div class="flex items-center justify-between gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2">
                      <div class="min-w-0">
                        <div class="truncate text-sm font-semibold text-white">{playlist.name}</div>
                        <div class="text-[11px] text-slate-500">ID: {playlist.id}</div>
                      </div>
                      <div class="flex items-center gap-1">
                        <button
                          type="button"
                          onClick={() => { void handleRenameSavedPlaylist(playlist); }}
                          class="rounded-lg border border-white/15 px-2 py-1 text-xs font-semibold text-slate-200 hover:bg-white/10 transition-colors"
                        >
                          Rename
                        </button>
                        <button
                          type="button"
                          onClick={() => { void handleDeleteSavedPlaylist(playlist); }}
                          class="rounded-lg border border-red-400/30 px-2 py-1 text-xs font-semibold text-red-200 hover:bg-red-500/20 transition-colors"
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
      </Show>
    </div>
  );
}
