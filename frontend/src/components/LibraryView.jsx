import { For, Show, createEffect, createMemo } from 'solid-js';
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
const metadataFor = (item) => (item?.metadata && typeof item.metadata === 'object' ? item.metadata : {});
const creatorLabel = (item) => {
  const metadata = metadataFor(item);
  return firstNonEmpty(item?.artist, metadata.artist, metadata.author, 'Unknown Artist');
};
const albumOrChannelLabel = (item) => {
  const metadata = metadataFor(item);
  return firstNonEmpty(item?.album, metadata.album, creatorLabel(item), 'Unknown Collection');
};
const playlistLabel = (item) => {
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
        return compareText(playlistLabel(left), playlistLabel(right)) || compareText(left.title, right.title);
      case 'playlist_desc':
        return compareText(playlistLabel(right), playlistLabel(left)) || compareText(left.title, right.title);
      case 'newest':
      default:
        return toTimestamp(right) - toTimestamp(left) || compareText(left.filename, right.filename);
    }
  });
  return nextItems;
};

export default function LibraryView(props) {
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
    };
  });

  const sortKey = createMemo(() => {
    const source = typeof props.sortKey === 'function' ? props.sortKey() : props.sortKey;
    return VALID_SORT_KEYS.has(source) ? source : 'newest';
  });

  const mediaCounts = createMemo(() => {
    const counts = { video: 0, audio: 0 };
    for (const item of downloads()) {
      counts[normalizeMediaType(item?.type)] += 1;
    }
    return counts;
  });

  const typedItems = createMemo(() => downloads().filter((item) => normalizeMediaType(item?.type) === activeMediaType()));
  const creatorOptions = createMemo(() => toUniqueSortedValues(typedItems().map((item) => creatorLabel(item))));
  const collectionOptions = createMemo(() => toUniqueSortedValues(typedItems().map((item) => albumOrChannelLabel(item))));
  const playlistOptions = createMemo(() => toUniqueSortedValues(typedItems().map((item) => playlistLabel(item))));

  const filteredItems = createMemo(() => {
    const activeFilters = filters();
    return typedItems().filter((item) => (
      (activeFilters.creator === '' || creatorLabel(item) === activeFilters.creator) &&
      (activeFilters.collection === '' || albumOrChannelLabel(item) === activeFilters.collection) &&
      (activeFilters.playlist === '' || playlistLabel(item) === activeFilters.playlist)
    ));
  });

  const visibleItems = createMemo(() => sortMediaItems(filteredItems(), sortKey()));

  const hasActiveFilters = createMemo(() => {
    const value = filters();
    return value.creator !== '' || value.collection !== '' || value.playlist !== '';
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
    }
  };

  const handleSortKeyChange = (nextSortKey) => {
    if (typeof props.onSortKeyChange === 'function') {
      props.onSortKeyChange(VALID_SORT_KEYS.has(nextSortKey) ? nextSortKey : 'newest');
    }
  };

  const activeMediaLabel = createMemo(() => (activeMediaType() === 'audio' ? 'audio' : 'video'));
  const hasMediaForTab = createMemo(() => typedItems().length > 0);

  createEffect(() => {
    const activeFilters = filters();
    const creators = new Set(creatorOptions());
    const collections = new Set(collectionOptions());
    const playlists = new Set(playlistOptions());

    if (activeFilters.creator !== '' && !creators.has(activeFilters.creator)) {
      handleFilterChange('creator', '');
    }
    if (activeFilters.collection !== '' && !collections.has(activeFilters.collection)) {
      handleFilterChange('collection', '');
    }
    if (activeFilters.playlist !== '' && !playlists.has(activeFilters.playlist)) {
      handleFilterChange('playlist', '');
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

        <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-4 p-4 bg-[#0a0c14] border border-white/5 rounded-2xl">
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
            Playlist
            <select
              value={filters().playlist}
              onChange={(event) => handleFilterChange('playlist', event.currentTarget.value)}
              class="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
            >
              <option value="">All Playlists</option>
              <For each={playlistOptions()}>
                {(value) => <option value={value}>{value}</option>}
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
                        {playlistLabel(item)}
                      </span>
                    </div>
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
