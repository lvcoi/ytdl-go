const UNKNOWN_ARTIST = 'Unknown Artist';
const UNKNOWN_CHANNEL = 'Unknown Channel';
const UNKNOWN_ALBUM = 'Unknown Album';
const UNKNOWN_PLAYLIST = 'Standalone';

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

const normalizeMediaType = (value) => {
  const lowered = String(value || '').toLowerCase();
  if (lowered === 'audio') return 'audio';
  if (lowered === 'video') return 'video';
  return 'video';
};

const metadataFor = (item) => (item?.metadata && typeof item.metadata === 'object' ? item.metadata : {});

const toTimestamp = (item) => {
  if (typeof item?.modified_at === 'string') {
    const parsed = Date.parse(item.modified_at);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  if (typeof item?.modifiedAt === 'string') {
    const parsed = Date.parse(item.modifiedAt);
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

const compareText = (left, right) => String(left || '').localeCompare(String(right || ''), undefined, {
  sensitivity: 'base',
  numeric: true,
});

const withLatestThumb = (current, candidate) => {
  if (current !== '') {
    return current;
  }
  return candidate || '';
};

const normalizeItem = (rawItem, savedPlaylistById, playlistAssignments) => {
  const metadata = metadataFor(rawItem);
  const type = normalizeMediaType(rawItem?.type);
  const title = firstNonEmpty(rawItem?.title, metadata.title, 'Untitled');
  const creator = firstNonEmpty(
    rawItem?.artist,
    metadata.artist,
    metadata.author,
    type === 'audio' ? UNKNOWN_ARTIST : UNKNOWN_CHANNEL,
  );
  const album = firstNonEmpty(
    rawItem?.album,
    metadata.album,
    type === 'audio' ? UNKNOWN_ALBUM : creator,
  );
  const sourcePlaylist = firstNonEmpty(
    rawItem?.playlist?.title,
    metadata?.playlist?.title,
    UNKNOWN_PLAYLIST,
  );
  const thumbnailUrl = firstNonEmpty(
    rawItem?.thumbnail_url,
    rawItem?.thumbnailUrl,
    rawItem?.thumbnailURL,
    metadata.thumbnail_url,
    metadata.thumbnailUrl,
    metadata.thumbnailURL,
  );
  const mediaKey = firstNonEmpty(rawItem?.relative_path, rawItem?.filename, rawItem?.id);
  const savedPlaylistId = firstNonEmpty(playlistAssignments[mediaKey], '');
  const savedPlaylist = savedPlaylistById.get(savedPlaylistId);
  const timestamp = toTimestamp(rawItem);
  const hasSidecar = Boolean(rawItem?.has_sidecar ?? rawItem?.hasSidecar);

  const anomalies = [];
  if (!hasSidecar) {
    anomalies.push('missing_sidecar');
  }
  if (creator === UNKNOWN_ARTIST || creator === UNKNOWN_CHANNEL) {
    anomalies.push('missing_creator');
  }
  if (title === 'Untitled') {
    anomalies.push('missing_title');
  }

  return {
    raw: rawItem,
    mediaKey,
    filename: String(rawItem?.filename || ''),
    relativePath: firstNonEmpty(rawItem?.relative_path, rawItem?.filename, ''),
    folder: String(rawItem?.folder || ''),
    id: String(rawItem?.id || rawItem?.filename || mediaKey),
    type,
    title,
    creator,
    album,
    sourcePlaylist,
    savedPlaylistId: savedPlaylist?.id || '',
    savedPlaylistName: savedPlaylist?.name || 'Unassigned',
    date: firstNonEmpty(rawItem?.date, ''),
    modifiedAt: firstNonEmpty(rawItem?.modified_at, rawItem?.modifiedAt, ''),
    size: firstNonEmpty(rawItem?.size, ''),
    sizeBytes: Number.isFinite(Number(rawItem?.size_bytes)) ? Number(rawItem.size_bytes) : 0,
    sourceURL: firstNonEmpty(rawItem?.source_url, metadata.source_url, ''),
    thumbnailUrl,
    metadata,
    timestamp,
    hasSidecar,
    anomalies,
    hasAnomaly: anomalies.length > 0,
  };
};

export const sortMediaItems = (items, sortKey) => {
  const sorted = [...items];
  sorted.sort((left, right) => {
    switch (sortKey) {
      case 'oldest':
        return left.timestamp - right.timestamp || compareText(left.title, right.title);
      case 'creator_asc':
        return compareText(left.creator, right.creator) || compareText(left.title, right.title);
      case 'creator_desc':
        return compareText(right.creator, left.creator) || compareText(left.title, right.title);
      case 'collection_asc':
        return compareText(left.album, right.album) || compareText(left.title, right.title);
      case 'collection_desc':
        return compareText(right.album, left.album) || compareText(left.title, right.title);
      case 'playlist_asc':
        return compareText(left.sourcePlaylist, right.sourcePlaylist) || compareText(left.title, right.title);
      case 'playlist_desc':
        return compareText(right.sourcePlaylist, left.sourcePlaylist) || compareText(left.title, right.title);
      case 'newest':
      default:
        return right.timestamp - left.timestamp || compareText(left.title, right.title);
    }
  });
  return sorted;
};

const itemMatchesFilters = (item, filters) => {
  const query = String(filters.query || '').trim().toLowerCase();
  if (query) {
    const haystack = [
      item.title,
      item.creator,
      item.album,
      item.sourcePlaylist,
      item.savedPlaylistName,
      item.filename,
    ].join(' ').toLowerCase();
    if (!haystack.includes(query)) {
      return false;
    }
  }

  if (filters.creator && item.creator !== filters.creator) {
    return false;
  }
  if (filters.collection && item.album !== filters.collection) {
    return false;
  }
  if (filters.playlist && item.sourcePlaylist !== filters.playlist) {
    return false;
  }
  if (filters.savedPlaylistId && item.savedPlaylistId !== filters.savedPlaylistId) {
    return false;
  }
  return true;
};

const pushByLatest = (items) => [...items].sort((left, right) => (
  right.latestTimestamp - left.latestTimestamp || compareText(left.name, right.name)
));

const buildArtistGroups = (items) => {
  const byArtist = new Map();
  for (const item of items.filter((entry) => entry.type === 'audio')) {
    if (!byArtist.has(item.creator)) {
      byArtist.set(item.creator, {
        name: item.creator,
        type: 'artist',
        items: [],
        albumsByName: new Map(),
        latestTimestamp: item.timestamp,
        thumbnailUrl: item.thumbnailUrl || '',
      });
    }
    const group = byArtist.get(item.creator);
    group.items.push(item);
    group.latestTimestamp = Math.max(group.latestTimestamp, item.timestamp);
    group.thumbnailUrl = withLatestThumb(group.thumbnailUrl, item.thumbnailUrl);

    if (!group.albumsByName.has(item.album)) {
      group.albumsByName.set(item.album, {
        name: item.album,
        items: [],
        latestTimestamp: item.timestamp,
        thumbnailUrl: item.thumbnailUrl || '',
      });
    }
    const album = group.albumsByName.get(item.album);
    album.items.push(item);
    album.latestTimestamp = Math.max(album.latestTimestamp, item.timestamp);
    album.thumbnailUrl = withLatestThumb(album.thumbnailUrl, item.thumbnailUrl);
  }

  const groups = [];
  const byName = new Map();
  for (const [name, value] of byArtist.entries()) {
    const albums = pushByLatest(Array.from(value.albumsByName.values()))
      .map((album) => ({
        ...album,
        items: sortMediaItems(album.items, 'newest'),
        count: album.items.length,
      }));

    const normalized = {
      name,
      type: 'artist',
      items: sortMediaItems(value.items, 'newest'),
      count: value.items.length,
      latestTimestamp: value.latestTimestamp,
      thumbnailUrl: value.thumbnailUrl,
      albums,
    };
    groups.push(normalized);
    byName.set(name, normalized);
  }
  return { groups: pushByLatest(groups), byName };
};

const buildChannelGroups = (items) => {
  const byChannel = new Map();
  for (const item of items.filter((entry) => entry.type === 'video')) {
    if (!byChannel.has(item.creator)) {
      byChannel.set(item.creator, {
        name: item.creator,
        type: 'channel',
        items: [],
        latestTimestamp: item.timestamp,
        thumbnailUrl: item.thumbnailUrl || '',
      });
    }
    const group = byChannel.get(item.creator);
    group.items.push(item);
    group.latestTimestamp = Math.max(group.latestTimestamp, item.timestamp);
    group.thumbnailUrl = withLatestThumb(group.thumbnailUrl, item.thumbnailUrl);
  }

  const groups = [];
  const byName = new Map();
  for (const [name, value] of byChannel.entries()) {
    const normalized = {
      name,
      type: 'channel',
      items: sortMediaItems(value.items, 'newest'),
      count: value.items.length,
      latestTimestamp: value.latestTimestamp,
      thumbnailUrl: value.thumbnailUrl,
    };
    groups.push(normalized);
    byName.set(name, normalized);
  }
  return { groups: pushByLatest(groups), byName };
};

const buildPlaylistGroups = (items, savedPlaylists) => {
  const sourceMap = new Map();
  for (const item of items) {
    const label = firstNonEmpty(item.sourcePlaylist, UNKNOWN_PLAYLIST);
    const key = `source:${label.toLowerCase()}`;
    if (!sourceMap.has(key)) {
      sourceMap.set(key, {
        key,
        kind: 'source',
        name: label,
        items: [],
        latestTimestamp: item.timestamp,
        thumbnailUrl: item.thumbnailUrl || '',
      });
    }
    const entry = sourceMap.get(key);
    entry.items.push(item);
    entry.latestTimestamp = Math.max(entry.latestTimestamp, item.timestamp);
    entry.thumbnailUrl = withLatestThumb(entry.thumbnailUrl, item.thumbnailUrl);
  }

  const savedMap = new Map();
  for (const saved of savedPlaylists) {
    savedMap.set(saved.id, {
      key: `saved:${saved.id}`,
      kind: 'saved',
      name: saved.name,
      playlistId: saved.id,
      items: [],
      latestTimestamp: 0,
      thumbnailUrl: '',
    });
  }

  for (const item of items) {
    if (!item.savedPlaylistId || !savedMap.has(item.savedPlaylistId)) {
      continue;
    }
    const entry = savedMap.get(item.savedPlaylistId);
    entry.items.push(item);
    entry.latestTimestamp = Math.max(entry.latestTimestamp, item.timestamp);
    entry.thumbnailUrl = withLatestThumb(entry.thumbnailUrl, item.thumbnailUrl);
  }

  const sourceGroups = pushByLatest(Array.from(sourceMap.values()))
    .map((entry) => ({ ...entry, count: entry.items.length, items: sortMediaItems(entry.items, 'newest') }));
  const savedGroups = pushByLatest(Array.from(savedMap.values()))
    .map((entry) => ({ ...entry, count: entry.items.length, items: sortMediaItems(entry.items, 'newest') }));

  const combined = pushByLatest([...sourceGroups, ...savedGroups]);
  const byKey = new Map(combined.map((entry) => [entry.key, entry]));
  return { sourceGroups, savedGroups, combined, byKey };
};

const uniqueSortedValues = (values) => (
  Array.from(new Set(values.filter((value) => typeof value === 'string' && value.trim() !== '').map((value) => value.trim())))
    .sort(compareText)
);

export const buildLibraryModel = ({
  downloads,
  savedPlaylists,
  playlistAssignments,
  typeFilter,
  sortKey,
  filters,
}) => {
  const normalizedDownloads = Array.isArray(downloads) ? downloads : [];
  const normalizedSavedPlaylists = Array.isArray(savedPlaylists) ? savedPlaylists : [];
  const assignments = playlistAssignments && typeof playlistAssignments === 'object' ? playlistAssignments : {};

  const savedPlaylistById = new Map(normalizedSavedPlaylists.map((playlist) => [String(playlist.id || '').trim(), playlist]));
  const normalizedItems = normalizedDownloads.map((entry) => normalizeItem(entry, savedPlaylistById, assignments));
  const anomalyItems = normalizedItems.filter((item) => item.hasAnomaly);

  const typeScopedItems = normalizedItems.filter((item) => (
    typeFilter === 'all' ? true : item.type === typeFilter
  ));
  const filteredItems = typeScopedItems.filter((item) => itemMatchesFilters(item, filters || {}));
  const sortedItems = sortMediaItems(filteredItems, sortKey);

  const artistGroups = buildArtistGroups(filteredItems);
  const channelGroups = buildChannelGroups(filteredItems);
  const playlistGroups = buildPlaylistGroups(filteredItems, normalizedSavedPlaylists);

  const creatorCards = pushByLatest([
    ...artistGroups.groups.map((entry) => ({ ...entry, key: `artist:${entry.name}`, creatorType: 'artist' })),
    ...channelGroups.groups.map((entry) => ({ ...entry, key: `channel:${entry.name}`, creatorType: 'channel' })),
  ]);

  const filterOptions = {
    creators: uniqueSortedValues(filteredItems.map((item) => item.creator)),
    collections: uniqueSortedValues(filteredItems.map((item) => item.album)),
    sourcePlaylists: uniqueSortedValues(filteredItems.map((item) => item.sourcePlaylist)),
  };

  return {
    items: normalizedItems,
    filteredItems: sortedItems,
    anomalyItems,
    anomalyCount: anomalyItems.length,
    artists: artistGroups.groups,
    artistsByName: artistGroups.byName,
    channels: channelGroups.groups,
    channelsByName: channelGroups.byName,
    creatorCards,
    playlists: playlistGroups.combined,
    playlistsByKey: playlistGroups.byKey,
    sourcePlaylists: playlistGroups.sourceGroups,
    savedPlaylists: playlistGroups.savedGroups,
    filterOptions,
  };
};
