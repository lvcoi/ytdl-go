import { For, Show, createEffect, createMemo, createSignal } from 'solid-js';
import Icon from './Icon';

const MEDIA_TYPE_OPTIONS = [
  { value: 'video', label: 'Video' },
  { value: 'audio', label: 'Audio' },
];

const SORT_OPTIONS = [
  { value: 'newest', label: 'Newest first' },
  { value: 'oldest', label: 'Oldest first' },
  { value: 'creator_asc', label: 'Artist/Creator (A-Z)' },
  { value: 'creator_desc', label: 'Artist/Creator (Z-A)' },
  { value: 'collection_asc', label: 'Album/Channel (A-Z)' },
  { value: 'collection_desc', label: 'Album/Channel (Z-A)' },
  { value: 'playlist_asc', label: 'Playlist (A-Z)' },
  { value: 'playlist_desc', label: 'Playlist (Z-A)' },
];

const MAX_SAVED_PLAYLIST_NAME_LENGTH = 80;
const VALID_SORT_KEYS = new Set(SORT_OPTIONS.map((option) => option.value));

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

const normalizeMediaType = (value) => (String(value || '').toLowerCase() === 'audio' ? 'audio' : 'video');
const normalizeSavedPlaylistName = (value) => (
  String(value || '')
    .trim()
    .replace(/\s+/g, ' ')
    .slice(0, MAX_SAVED_PLAYLIST_NAME_LENGTH)
);
const mediaKeyForItem = (item) => (
  firstNonEmpty(item?.relative_path, item?.filename, item?.id)
);
const metadataFor = (item) => (item?.metadata && typeof item.metadata === 'object' ? item.metadata : {});
const creatorLabel = (item) => {
  const metadata = metadataFor(item);
  return firstNonEmpty(item?.artist, metadata.artist, metadata.author, 'Unknown Artist');
};
const albumOrChannelLabel = (item) => {
  const metadata = metadataFor(item);
  return firstNonEmpty(item?.album, metadata.album, creatorLabel(item), 'Unknown Collection');
};
const sourcePlaylistLabel = (item) => {
  const metadata = metadataFor(item);
  return firstNonEmpty(item?.playlist?.title, metadata?.playlist?.title, 'Standalone');
};

const compareText = (left, right) => String(left || '').localeCompare(String(right || ''), undefined, {
  sensitivity: 'base',
  numeric: true,
});

const toTimestamp = (item) => {
  if (typeof item?.modified_at === 'string') {
    const parsed = Date.parse(item.modified_at);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  if (typeof item?.date === 'string') {
    const parsed = Date.parse(item.date);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return 0;
};

const toUniqueSortedValues = (values) => (
  Array.from(new Set(values.filter((value) => typeof value === 'string' && value.trim() !== '').map((value) => value.trim())))
    .sort(compareText)
);

const normalizeSavedPlaylists = (rawSavedPlaylists) => {
  if (!Array.isArray(rawSavedPlaylists)) {
    return [];
  }

  const seen = new Set();
  const playlists = [];
  for (const entry of rawSavedPlaylists) {
    const value = entry && typeof entry === 'object' ? entry : {};
    const id = String(value.id || '').trim();
    const name = normalizeSavedPlaylistName(value.name);
    if (id === '' || name === '' || seen.has(id)) {
      continue;
    }
    seen.add(id);
    playlists.push({
      id,
      name,
    });
  }
  return playlists;
};

const normalizePlaylistAssignments = (rawAssignments) => {
  const value = rawAssignments && typeof rawAssignments === 'object' ? rawAssignments : {};
  const assignments = {};
  for (const [mediaKey, playlistId] of Object.entries(value)) {
    const normalizedMediaKey = String(mediaKey || '').trim();
    const normalizedPlaylistId = String(playlistId || '').trim();
    if (normalizedMediaKey === '' || normalizedPlaylistId === '') {
      continue;
    }
    assignments[normalizedMediaKey] = normalizedPlaylistId;
  }
  return assignments;
};

const sortMediaItems = (items, sortKey) => {
  const nextItems = [...items];
  nextItems.sort((left, right) => {
    switch (sortKey) {
      case 'oldest':
        return toTimestamp(left) - toTimestamp(right) || compareText(left.filename, right.filename);
      case 'creator_asc':
        return compareText(creatorLabel(left), creatorLabel(right)) || compareText(left.title, right.title);
      case 'creator_desc':
        return compareText(creatorLabel(right), creatorLabel(left)) || compareText(left.title, right.title);
      case 'collection_asc':
        return compareText(albumOrChannelLabel(left), albumOrChannelLabel(right)) || compareText(left.title, right.title);
      case 'collection_desc':
        return compareText(albumOrChannelLabel(right), albumOrChannelLabel(left)) || compareText(left.title, right.title);
      case 'playlist_asc':
        return compareText(sourcePlaylistLabel(left), sourcePlaylistLabel(right)) || compareText(left.title, right.title);
      case 'playlist_desc':
        return compareText(sourcePlaylistLabel(right), sourcePlaylistLabel(left)) || compareText(left.title, right.title);
      case 'newest':
      default:
        return toTimestamp(right) - toTimestamp(left) || compareText(left.filename, right.filename);
    }
  });
  return nextItems;
};

export default function LibraryView(props) {
  const [newSavedPlaylistName, setNewSavedPlaylistName] = createSignal('');
  const [playlistMessage, setPlaylistMessage] = createSignal('');
  const [playlistMessageTone, setPlaylistMessageTone] = createSignal('neutral');

  const downloads = createMemo(() => {
    const source = typeof props.downloads === 'function' ? props.downloads() : props.downloads;
    return Array.isArray(source) ? source : [];
  });

  const activeMediaType = createMemo(() => {
    const source = typeof props.activeMediaType === 'function' ? props.activeMediaType() : props.activeMediaType;
    return normalizeMediaType(source);
  });

  const filters = createMemo(() => {
    const source = typeof props.filters === 'function' ? props.filters() : props.filters;
    const value = source && typeof source === 'object' ? source : {};
    return {
      creator: typeof value.creator === 'string' ? value.creator : '',
      collection: typeof value.collection === 'string' ? value.collection : '',
      playlist: typeof value.playlist === 'string' ? value.playlist : '',
      savedPlaylistId: typeof value.savedPlaylistId === 'string' ? value.savedPlaylistId : '',
    };
  });

  const sortKey = createMemo(() => {
    const source = typeof props.sortKey === 'function' ? props.sortKey() : props.sortKey;
    return VALID_SORT_KEYS.has(source) ? source : 'newest';
  });

  const savedPlaylists = createMemo(() => {
    const source = typeof props.savedPlaylists === 'function' ? props.savedPlaylists() : props.savedPlaylists;
    return normalizeSavedPlaylists(source);
  });

  const playlistAssignments = createMemo(() => {
    const source = typeof props.playlistAssignments === 'function' ? props.playlistAssignments() : props.playlistAssignments;
    return normalizePlaylistAssignments(source);
  });

  const savedPlaylistById = createMemo(() => new Map(savedPlaylists().map((playlist) => [playlist.id, playlist])));
  const savedPlaylistOptions = createMemo(() => savedPlaylists());

  const savedPlaylistIdForItem = (item) => {
    const mediaKey = mediaKeyForItem(item);
    if (mediaKey === '') {
      return '';
    }
    const assignedPlaylistId = playlistAssignments()[mediaKey] || '';
    return savedPlaylistById().has(assignedPlaylistId) ? assignedPlaylistId : '';
  };

  const savedPlaylistLabelForItem = (item) => {
    const playlist = savedPlaylistById().get(savedPlaylistIdForItem(item));
    return playlist ? playlist.name : 'Unassigned';
  };

  const mediaCounts = createMemo(() => {
    const counts = { video: 0, audio: 0 };
    for (const item of downloads()) {
      counts[normalizeMediaType(item?.type)] += 1;
    }
    return counts;
  });

  const typedItems = createMemo(() => downloads().filter((item) => normalizeMediaType(item?.type) === activeMediaType()));

  const filterOptions = createMemo(() => {
    const creators = new Set();
    const collections = new Set();
    const sourcePlaylists = new Set();

    for (const item of typedItems()) {
      const creator = creatorLabel(item);
      if (typeof creator === 'string') {
        const trimmed = creator.trim();
        if (trimmed !== '') {
          creators.add(trimmed);
        }
      }

      const collection = albumOrChannelLabel(item);
      if (typeof collection === 'string') {
        const trimmed = collection.trim();
        if (trimmed !== '') {
          collections.add(trimmed);
        }
      }

      const sourcePlaylist = sourcePlaylistLabel(item);
      if (typeof sourcePlaylist === 'string') {
        const trimmed = sourcePlaylist.trim();
        if (trimmed !== '') {
          sourcePlaylists.add(trimmed);
        }
      }
    }

    return {
      creator: Array.from(creators).sort(compareText),
      collection: Array.from(collections).sort(compareText),
      playlist: Array.from(sourcePlaylists).sort(compareText),
    };
  });

  const creatorOptions = createMemo(() => filterOptions().creator);
  const collectionOptions = createMemo(() => filterOptions().collection);
  const sourcePlaylistOptions = createMemo(() => filterOptions().playlist);

  const filteredItems = createMemo(() => {
    const activeFilters = filters();
    return typedItems().filter((item) => (
      (activeFilters.creator === '' || creatorLabel(item) === activeFilters.creator) &&
      (activeFilters.collection === '' || albumOrChannelLabel(item) === activeFilters.collection) &&
      (activeFilters.playlist === '' || sourcePlaylistLabel(item) === activeFilters.playlist) &&
      (activeFilters.savedPlaylistId === '' || savedPlaylistIdForItem(item) === activeFilters.savedPlaylistId)
    ));
  });

  const visibleItems = createMemo(() => sortMediaItems(filteredItems(), sortKey()));

  const hasActiveFilters = createMemo(() => {
    const value = filters();
    return value.creator !== '' || value.collection !== '' || value.playlist !== '' || value.savedPlaylistId !== '';
  });

  const handleMediaTypeChange = (nextType) => {
    if (typeof props.onMediaTypeChange === 'function') {
      props.onMediaTypeChange(normalizeMediaType(nextType));
    }
  };

  const handleFilterChange = (filterKey, value) => {
    if (typeof props.onFilterChange === 'function') {
      props.onFilterChange(filterKey, value);
    }
  };

  const handleClearFilters = () => {
    if (typeof props.onClearFilters === 'function') {
      props.onClearFilters();
      return;
    }
    if (typeof props.onFilterChange === 'function') {
      props.onFilterChange('creator', '');
      props.onFilterChange('collection', '');
      props.onFilterChange('playlist', '');
      props.onFilterChange('savedPlaylistId', '');
    }
  };

  const handleSortKeyChange = (nextSortKey) => {
    if (typeof props.onSortKeyChange === 'function') {
      props.onSortKeyChange(VALID_SORT_KEYS.has(nextSortKey) ? nextSortKey : 'newest');
    }
  };

  const setPlaylistFeedback = (tone, text) => {
    setPlaylistMessageTone(tone);
    setPlaylistMessage(text);
  };

  const handleCreateSavedPlaylist = () => {
    const nextName = normalizeSavedPlaylistName(newSavedPlaylistName());
    if (nextName === '') {
      setPlaylistFeedback('error', 'Enter a playlist name.');
      return;
    }
    if (typeof props.onCreateSavedPlaylist !== 'function') {
      return;
    }

    const result = props.onCreateSavedPlaylist(nextName);
    if (!result || result.ok === false) {
      setPlaylistFeedback('error', result?.error || 'Unable to create saved playlist.');
      return;
    }
    setNewSavedPlaylistName('');
    setPlaylistFeedback('success', `Saved playlist "${nextName}" created.`);
  };

  const handleRenameSavedPlaylist = (playlist) => {
    if (typeof window === 'undefined' || typeof props.onRenameSavedPlaylist !== 'function') {
      return;
    }
    const proposedName = window.prompt('Rename saved playlist', playlist.name);
    if (proposedName === null) {
      return;
    }
    const result = props.onRenameSavedPlaylist(playlist.id, proposedName);
    if (!result || result.ok === false) {
      setPlaylistFeedback('error', result?.error || 'Unable to rename saved playlist.');
      return;
    }
    const normalizedName = normalizeSavedPlaylistName(proposedName);
    setPlaylistFeedback('success', `Saved playlist renamed to "${normalizedName || playlist.name}".`);
  };

  const handleDeleteSavedPlaylist = (playlist) => {
    if (typeof window === 'undefined' || typeof props.onDeleteSavedPlaylist !== 'function') {
      return;
    }
    const confirmed = window.confirm(`Delete "${playlist.name}" and remove its assignments?`);
    if (!confirmed) {
      return;
    }
    const result = props.onDeleteSavedPlaylist(playlist.id);
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
    const mediaKey = mediaKeyForItem(item);
    if (mediaKey === '') {
      return;
    }
    props.onAssignSavedPlaylist(mediaKey, nextSavedPlaylistId);
  };

  const activeMediaLabel = createMemo(() => (activeMediaType() === 'audio' ? 'audio' : 'video'));
  const hasMediaForTab = createMemo(() => typedItems().length > 0);

  createEffect(() => {
    const activeFilters = filters();
    const creators = new Set(creatorOptions());
    const collections = new Set(collectionOptions());
    const sourcePlaylists = new Set(sourcePlaylistOptions());
    const savedPlaylistIds = new Set(savedPlaylistOptions().map((playlist) => playlist.id));

    if (activeFilters.creator !== '' && !creators.has(activeFilters.creator)) {
      handleFilterChange('creator', '');
    }
    if (activeFilters.collection !== '' && !collections.has(activeFilters.collection)) {
      handleFilterChange('collection', '');
    }
    if (activeFilters.playlist !== '' && !sourcePlaylists.has(activeFilters.playlist)) {
      handleFilterChange('playlist', '');
    }
    if (activeFilters.savedPlaylistId !== '' && !savedPlaylistIds.has(activeFilters.savedPlaylistId)) {
      handleFilterChange('savedPlaylistId', '');
    }
  });

  return (
    <div class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-500">
      <div class="space-y-4">
        <div class="flex items-center justify-between gap-4">
          <div>
            <h1 class="text-3xl font-black text-white">Your Library</h1>
            <p class="text-gray-500">Access your downloaded media instantly.</p>
          </div>
          <div class="text-right text-xs text-gray-500 font-semibold tracking-wide uppercase">
            Showing {visibleItems().length} of {typedItems().length} {activeMediaLabel()} items
          </div>
        </div>

        <div class="inline-flex items-center gap-2 p-1 bg-white/5 border border-white/5 rounded-2xl">
          <For each={MEDIA_TYPE_OPTIONS}>
            {(option) => (
              <button
                onClick={() => handleMediaTypeChange(option.value)}
                class={`px-4 py-2 rounded-xl text-sm font-semibold transition-all ${
                  activeMediaType() === option.value
                    ? 'bg-blue-600/20 text-blue-300 border border-blue-500/40'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-white/5 border border-transparent'
                }`}
              >
                <span>{option.label}</span>
                <span class={`ml-2 text-xs ${activeMediaType() === option.value ? 'text-blue-200' : 'text-gray-500'}`}>
                  {mediaCounts()[option.value]}
                </span>
              </button>
            )}
          </For>
        </div>

        <div class="p-4 bg-[#0a0c14] border border-white/5 rounded-2xl space-y-3">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="text-xs font-semibold tracking-wide text-gray-400 uppercase">Saved Playlists</div>
              <div class="text-xs text-gray-500">Create custom collections and assign items instantly.</div>
            </div>
            <div class="text-xs text-gray-500">
              {savedPlaylistOptions().length} saved {savedPlaylistOptions().length === 1 ? 'playlist' : 'playlists'}
            </div>
          </div>

          <div class="flex flex-col gap-2 sm:flex-row">
            <input
              type="text"
              value={newSavedPlaylistName()}
              onInput={(event) => setNewSavedPlaylistName(event.currentTarget.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  event.preventDefault();
                  handleCreateSavedPlaylist();
                }
              }}
              placeholder="New saved playlist name"
              class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
            />
            <button
              type="button"
              onClick={handleCreateSavedPlaylist}
              class="px-4 py-2 rounded-xl bg-blue-600 text-white text-sm font-semibold hover:bg-blue-500 transition-colors"
            >
              Create
            </button>
          </div>

          <Show when={playlistMessage() !== ''}>
            <div class={`text-xs ${
              playlistMessageTone() === 'error'
                ? 'text-red-400'
                : playlistMessageTone() === 'success'
                  ? 'text-green-400'
                  : 'text-gray-400'
            }`}
            >
              {playlistMessage()}
            </div>
          </Show>

          <Show
            when={savedPlaylistOptions().length > 0}
            fallback={<div class="text-xs text-gray-500">No saved playlists yet.</div>}
          >
            <div class="flex flex-wrap gap-2">
              <For each={savedPlaylistOptions()}>
                {(playlist) => (
                  <div class={`inline-flex items-center gap-1 rounded-full border px-3 py-1 text-xs transition-colors ${
                    filters().savedPlaylistId === playlist.id
                      ? 'border-blue-500/40 bg-blue-600/10 text-blue-200'
                      : 'border-white/10 bg-white/5 text-gray-300'
                  }`}
                  >
                    <button
                      type="button"
                      onClick={() => handleFilterChange('savedPlaylistId', filters().savedPlaylistId === playlist.id ? '' : playlist.id)}
                      class="font-semibold hover:text-white transition-colors"
                    >
                      {playlist.name}
                    </button>
                    <button
                      type="button"
                      onClick={() => handleRenameSavedPlaylist(playlist)}
                      class="px-1 text-gray-400 hover:text-gray-200 transition-colors"
                      title={`Rename ${playlist.name}`}
                    >
                      Rename
                    </button>
                    <button
                      type="button"
                      onClick={() => handleDeleteSavedPlaylist(playlist)}
                      class="px-1 text-red-400/80 hover:text-red-300 transition-colors"
                      title={`Delete ${playlist.name}`}
                    >
                      Delete
                    </button>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </div>

        <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 p-4 bg-[#0a0c14] border border-white/5 rounded-2xl">
          <label class="space-y-2 text-xs font-semibold tracking-wide text-gray-400 uppercase">
            Artist / Creator
            <select
              value={filters().creator}
              onChange={(event) => handleFilterChange('creator', event.currentTarget.value)}
              class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
            >
              <option value="">All Artists/Creators</option>
              <For each={creatorOptions()}>
                {(value) => <option value={value}>{value}</option>}
              </For>
            </select>
          </label>

          <label class="space-y-2 text-xs font-semibold tracking-wide text-gray-400 uppercase">
            Album / Channel
            <select
              value={filters().collection}
              onChange={(event) => handleFilterChange('collection', event.currentTarget.value)}
              class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
            >
              <option value="">All Albums/Channels</option>
              <For each={collectionOptions()}>
                {(value) => <option value={value}>{value}</option>}
              </For>
            </select>
          </label>

          <label class="space-y-2 text-xs font-semibold tracking-wide text-gray-400 uppercase">
            Source Playlist
            <select
              value={filters().playlist}
              onChange={(event) => handleFilterChange('playlist', event.currentTarget.value)}
              class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
            >
              <option value="">All Source Playlists</option>
              <For each={sourcePlaylistOptions()}>
                {(value) => <option value={value}>{value}</option>}
              </For>
            </select>
          </label>

          <label class="space-y-2 text-xs font-semibold tracking-wide text-gray-400 uppercase">
            Saved Playlist
            <select
              value={filters().savedPlaylistId}
              onChange={(event) => handleFilterChange('savedPlaylistId', event.currentTarget.value)}
              class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
            >
              <option value="">All Saved Playlists</option>
              <For each={savedPlaylistOptions()}>
                {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
              </For>
            </select>
          </label>

          <div class="space-y-2 text-xs font-semibold tracking-wide text-gray-400 uppercase">
            Sort
            <div class="flex gap-2">
              <select
                value={sortKey()}
                onChange={(event) => handleSortKeyChange(event.currentTarget.value)}
                class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
              >
                <For each={SORT_OPTIONS}>
                  {(option) => <option value={option.value}>{option.label}</option>}
                </For>
              </select>
              <button
                type="button"
                onClick={handleClearFilters}
                disabled={!hasActiveFilters()}
                class={`px-3 py-2 rounded-xl border transition-all ${
                  hasActiveFilters()
                    ? 'bg-white/5 border-white/10 text-gray-200 hover:text-white hover:border-white/20'
                    : 'bg-white/5 border-white/5 text-gray-500 cursor-not-allowed'
                }`}
                title="Clear filters"
              >
                <Icon name="x" class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <Show
        when={visibleItems().length > 0}
        fallback={(
          <div class="p-8 bg-[#0a0c14] border border-white/5 rounded-2xl text-center space-y-2">
            <div class="text-sm font-semibold text-white">
              {hasMediaForTab()
                ? 'No items match the current filters.'
                : `No ${activeMediaLabel()} downloads yet.`}
            </div>
            <div class="text-xs text-gray-500">
              {hasMediaForTab()
                ? 'Adjust filters or clear them to show more media.'
                : 'Download new media to populate this section.'}
            </div>
          </div>
        )}
      >
        <div class="grid gap-3">
          <For each={visibleItems()}>
            {(item) => (
              <div class="group flex items-center justify-between p-4 bg-[#0a0c14] border border-white/5 rounded-2xl hover:border-blue-500/30 transition-all cursor-default">
                <div class="flex items-center gap-5 min-w-0">
                  <div class="w-16 h-16 bg-white/5 rounded-xl flex items-center justify-center relative overflow-hidden group-hover:bg-blue-600/20 transition-all shrink-0">
                    <Icon name={normalizeMediaType(item?.type) === 'video' ? 'film' : 'music'} class="w-6 h-6 text-gray-600 group-hover:text-blue-400" />
                  </div>
                  <div class="min-w-0">
                    <div class="font-bold text-white group-hover:text-blue-400 transition-colors truncate">{item.title}</div>
                    <div class="text-xs text-gray-500 font-medium truncate">{creatorLabel(item)} • {item.size} • {item.date}</div>
                    <div class="mt-2 flex flex-wrap gap-2 text-[10px] font-semibold uppercase tracking-wide">
                      <span class="px-2 py-1 rounded-full bg-white/5 border border-white/10 text-gray-400">
                        {albumOrChannelLabel(item)}
                      </span>
                      <span class="px-2 py-1 rounded-full bg-white/5 border border-white/10 text-gray-400">
                        {sourcePlaylistLabel(item)}
                      </span>
                      <span class={`px-2 py-1 rounded-full border ${
                        savedPlaylistIdForItem(item) === ''
                          ? 'bg-white/5 border-white/10 text-gray-500'
                          : 'bg-blue-600/10 border-blue-500/30 text-blue-200'
                      }`}
                      >
                        Saved: {savedPlaylistLabelForItem(item)}
                      </span>
                    </div>
                    <label class="mt-3 block text-[10px] font-semibold uppercase tracking-wide text-gray-400">
                      Saved playlist
                      <select
                        value={savedPlaylistIdForItem(item)}
                        onChange={(event) => handleAssignSavedPlaylist(item, event.currentTarget.value)}
                        class="mt-1 w-full max-w-xs bg-white/5 border border-white/10 rounded-lg px-2 py-1 text-xs text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40 normal-case tracking-normal"
                      >
                        <option value="">Unassigned</option>
                        <For each={savedPlaylistOptions()}>
                          {(playlist) => <option value={playlist.id}>{playlist.name}</option>}
                        </For>
                      </select>
                    </label>
                  </div>
                </div>
                <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all shrink-0">
                  <button
                    onClick={() => {
                      if (typeof props.openPlayer === 'function') {
                        props.openPlayer(item);
                      }
                    }}
                    class="p-3 bg-blue-600 rounded-xl text-white shadow-lg shadow-blue-600/20 hover:scale-105 active:scale-95 transition-all"
                  >
                    <Icon name="play" class="w-5 h-5 fill-white" />
                  </button>
                  <button class="p-3 bg-white/5 rounded-xl text-gray-400 hover:text-white transition-all">
                    <Icon name="external-link" class="w-5 h-5" />
                  </button>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  );
}
