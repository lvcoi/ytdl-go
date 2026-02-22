import { createContext, createEffect, useContext } from 'solid-js';
import { createStore } from 'solid-js/store';

const APP_STATE_STORAGE_KEY = 'ytdl-go:app-state:v1';
const VALID_TABS = new Set(['download', 'library', 'settings']);
const VALID_DUPLICATE_POLICIES = new Set(['prompt', 'overwrite', 'skip', 'rename']);
const VALID_LIBRARY_SECTIONS = new Set(['artists', 'channels', 'playlists', 'all_media']);
const VALID_LIBRARY_VIEW_MODES = new Set(['gallery', 'list', 'detail']);
const VALID_LIBRARY_TYPE_FILTERS = new Set(['all', 'audio', 'video']);
const VALID_LIBRARY_SORT_KEYS = new Set([
  'newest',
  'oldest',
  'creator_asc',
  'creator_desc',
  'collection_asc',
  'collection_desc',
  'playlist_asc',
  'playlist_desc',
]);
const MAX_SAVED_PLAYLIST_NAME_LENGTH = 80;

export const MAX_JOBS = 32;
export const MAX_TIMEOUT_SECONDS = 24 * 60 * 60;

const defaultSettings = {
  output: '{title}.{ext}',
  quality: 'best',
  jobs: 1,
  timeout: 180,
  format: '',
  audioOnly: false,
  onDuplicate: 'prompt',
  useCookies: true,
  poTokenExtension: false,
};

const emptyLibraryNavPath = {
  creatorType: '',
  creatorName: '',
  albumName: '',
  playlistKey: '',
  playlistKind: '',
};

const emptyLibraryFilters = {
  query: '',
  creator: '',
  collection: '',
  playlist: '',
  savedPlaylistId: '',
};

const createDefaultState = () => ({
  ui: {
    activeTab: 'dashboard',
    isAdvanced: false,
  },
  settings: { ...defaultSettings },
  library: {
    downloads: [],
    section: 'artists',
    viewMode: 'gallery',
    typeFilter: 'all',
    navPath: { ...emptyLibraryNavPath },
    filters: { ...emptyLibraryFilters },
    sortKey: 'newest',
    savedPlaylists: [],
    playlistAssignments: {},
    ui: {
      advancedFiltersOpen: false,
      metadataBannerDismissed: false,
    },
  },
  player: {
    active: false,
    selectedMedia: null,
    minimized: false,
  },
  download: {
    urlInput: '',
    isDownloading: false,
    jobStatus: null,
    progressTasks: {},
    logMessages: [],
    duplicateQueue: [],
    duplicateError: '',
  },
});

const AppStoreContext = createContext();
let hasLoggedStorageReadError = false;

const getPersistedState = (state) => ({
  ui: {
    activeTab: state.ui.activeTab,
    isAdvanced: state.ui.isAdvanced,
  },
  settings: {
    ...state.settings,
  },
  library: {
    section: state.library.section,
    viewMode: state.library.viewMode,
    typeFilter: state.library.typeFilter,
    navPath: {
      ...state.library.navPath,
    },
    filters: {
      query: state.library.filters.query,
      creator: state.library.filters.creator,
      collection: state.library.filters.collection,
      playlist: state.library.filters.playlist,
      savedPlaylistId: state.library.filters.savedPlaylistId,
    },
    sortKey: state.library.sortKey,
    savedPlaylists: state.library.savedPlaylists,
    playlistAssignments: state.library.playlistAssignments,
    ui: {
      advancedFiltersOpen: state.library.ui.advancedFiltersOpen,
      metadataBannerDismissed: state.library.ui.metadataBannerDismissed,
    },
  },
  download: {
    urlInput: state.download.urlInput,
    activeJobId: state.download.activeJobId,
  },
});

const toString = (value, fallback) => (typeof value === 'string' ? value : fallback);
const toBoolean = (value, fallback) => (typeof value === 'boolean' ? value : fallback);

const normalizeSavedPlaylistName = (value) => (
  toString(value, '')
    .trim()
    .replace(/\s+/g, ' ')
    .slice(0, MAX_SAVED_PLAYLIST_NAME_LENGTH)
);

const toBoundedPositiveInteger = (value, fallback, max) => {
  const parsed = typeof value === 'number' ? value : Number(value);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  const normalized = Math.trunc(parsed);
  if (normalized <= 0) {
    return fallback;
  }
  return Math.min(normalized, max);
};

const sanitizeSettings = (rawSettings) => {
  const raw = rawSettings && typeof rawSettings === 'object' ? rawSettings : {};

  const onDuplicate = toString(raw.onDuplicate, defaultSettings.onDuplicate);
  return {
    output: toString(raw.output, defaultSettings.output),
    quality: toString(raw.quality, defaultSettings.quality),
    jobs: toBoundedPositiveInteger(raw.jobs, defaultSettings.jobs, MAX_JOBS),
    timeout: toBoundedPositiveInteger(raw.timeout, defaultSettings.timeout, MAX_TIMEOUT_SECONDS),
    format: toString(raw.format, defaultSettings.format),
    audioOnly: toBoolean(raw.audioOnly, defaultSettings.audioOnly),
    onDuplicate: VALID_DUPLICATE_POLICIES.has(onDuplicate) ? onDuplicate : defaultSettings.onDuplicate,
    useCookies: toBoolean(raw.useCookies, defaultSettings.useCookies),
    poTokenExtension: toBoolean(raw.poTokenExtension, defaultSettings.poTokenExtension),
  };
};

const sanitizeLibraryFilters = (rawFilters, validSavedPlaylistIds = new Set()) => {
  const raw = rawFilters && typeof rawFilters === 'object' ? rawFilters : {};
  const savedPlaylistId = toString(raw.savedPlaylistId, '').trim();
  return {
    query: toString(raw.query, ''),
    creator: toString(raw.creator, ''),
    collection: toString(raw.collection, ''),
    playlist: toString(raw.playlist, ''),
    savedPlaylistId: validSavedPlaylistIds.has(savedPlaylistId) ? savedPlaylistId : '',
  };
};

const sanitizeSavedPlaylists = (rawPlaylists) => {
  if (!Array.isArray(rawPlaylists)) {
    return [];
  }
  const seenIds = new Set();
  const out = [];
  for (const entry of rawPlaylists) {
    const value = entry && typeof entry === 'object' ? entry : {};
    const id = toString(value.id, '').trim();
    const name = normalizeSavedPlaylistName(value.name);
    if (id === '' || name === '' || seenIds.has(id)) {
      continue;
    }
    seenIds.add(id);
    out.push({
      id,
      name,
      createdAt: toString(value.createdAt, ''),
      updatedAt: toString(value.updatedAt, ''),
    });
  }
  return out;
};

const sanitizePlaylistAssignments = (rawAssignments, validSavedPlaylistIds) => {
  const raw = rawAssignments && typeof rawAssignments === 'object' ? rawAssignments : {};
  const out = {};
  for (const [mediaKey, value] of Object.entries(raw)) {
    const normalizedMediaKey = String(mediaKey || '').trim();
    const savedPlaylistId = toString(value, '').trim();
    if (normalizedMediaKey === '' || savedPlaylistId === '' || !validSavedPlaylistIds.has(savedPlaylistId)) {
      continue;
    }
    out[normalizedMediaKey] = savedPlaylistId;
  }
  return out;
};

const sanitizeLibraryNavPath = (rawNavPath) => {
  const raw = rawNavPath && typeof rawNavPath === 'object' ? rawNavPath : {};
  const creatorType = toString(raw.creatorType, '').trim().toLowerCase();
  const playlistKind = toString(raw.playlistKind, '').trim().toLowerCase();

  return {
    creatorType: creatorType === 'artist' || creatorType === 'channel' ? creatorType : '',
    creatorName: toString(raw.creatorName, ''),
    albumName: toString(raw.albumName, ''),
    playlistKey: toString(raw.playlistKey, ''),
    playlistKind: playlistKind === 'source' || playlistKind === 'saved' ? playlistKind : '',
  };
};

const sanitizeLibraryUiState = (rawUiState) => {
  const raw = rawUiState && typeof rawUiState === 'object' ? rawUiState : {};
  return {
    advancedFiltersOpen: toBoolean(raw.advancedFiltersOpen, false),
    metadataBannerDismissed: toBoolean(raw.metadataBannerDismissed, false),
  };
};

const sanitizeLibrary = (rawLibrary) => {
  const raw = rawLibrary && typeof rawLibrary === 'object' ? rawLibrary : {};

  const section = toString(raw.section, 'artists');
  const viewMode = toString(raw.viewMode, 'gallery');
  const sortKey = toString(raw.sortKey, 'newest');
  const savedPlaylists = sanitizeSavedPlaylists(raw.savedPlaylists);
  const validSavedPlaylistIds = new Set(savedPlaylists.map((playlist) => playlist.id));

  const requestedTypeFilter = toString(raw.typeFilter, '').trim().toLowerCase();
  const legacyActiveMediaType = toString(raw.activeMediaType, '').trim().toLowerCase();
  let typeFilter = 'all';
  if (VALID_LIBRARY_TYPE_FILTERS.has(requestedTypeFilter)) {
    typeFilter = requestedTypeFilter;
  } else if (VALID_LIBRARY_TYPE_FILTERS.has(legacyActiveMediaType)) {
    typeFilter = legacyActiveMediaType;
  }

  return {
    section: VALID_LIBRARY_SECTIONS.has(section) ? section : 'artists',
    viewMode: VALID_LIBRARY_VIEW_MODES.has(viewMode) ? viewMode : 'gallery',
    typeFilter,
    navPath: sanitizeLibraryNavPath(raw.navPath),
    filters: sanitizeLibraryFilters(raw.filters, validSavedPlaylistIds),
    sortKey: VALID_LIBRARY_SORT_KEYS.has(sortKey) ? sortKey : 'newest',
    savedPlaylists,
    playlistAssignments: sanitizePlaylistAssignments(raw.playlistAssignments, validSavedPlaylistIds),
    ui: sanitizeLibraryUiState(raw.ui),
  };
};

const readPersistedState = () => {
  if (typeof window === 'undefined') {
    return null;
  }

  try {
    const raw = window.localStorage.getItem(APP_STATE_STORAGE_KEY);
    if (!raw) {
      return null;
    }
    return JSON.parse(raw);
  } catch (error) {
    if (!hasLoggedStorageReadError) {
      console.warn('Failed to read persisted app state from localStorage:', error);
      hasLoggedStorageReadError = true;
    }
    return null;
  }
};

const getInitialState = () => {
  const baseState = createDefaultState();
  const persisted = readPersistedState();
  if (!persisted || typeof persisted !== 'object') {
    return baseState;
  }

  const activeTab = VALID_TABS.has(persisted?.ui?.activeTab)
    ? persisted.ui.activeTab
    : baseState.ui.activeTab;
  const isAdvanced = typeof persisted?.ui?.isAdvanced === 'boolean'
    ? persisted.ui.isAdvanced
    : baseState.ui.isAdvanced;

  const persistedSettings = sanitizeSettings(persisted.settings);
  const persistedLibrary = sanitizeLibrary(persisted.library);
  const persistedUrlInput = typeof persisted?.download?.urlInput === 'string'
    ? persisted.download.urlInput
    : baseState.download.urlInput;
  const persistedActiveJobId = typeof persisted?.download?.activeJobId === 'string'
    ? persisted.download.activeJobId
    : '';

  return {
    ...baseState,
    ui: {
      activeTab,
      isAdvanced,
    },
    settings: {
      ...persistedSettings,
    },
    library: {
      ...baseState.library,
      section: persistedLibrary.section,
      viewMode: persistedLibrary.viewMode,
      typeFilter: persistedLibrary.typeFilter,
      navPath: persistedLibrary.navPath,
      filters: persistedLibrary.filters,
      sortKey: persistedLibrary.sortKey,
      savedPlaylists: persistedLibrary.savedPlaylists,
      playlistAssignments: persistedLibrary.playlistAssignments,
      ui: persistedLibrary.ui,
    },
    download: {
      ...baseState.download,
      urlInput: persistedUrlInput,
      activeJobId: persistedActiveJobId,
    },
  };
};

export function AppStoreProvider(props) {
  const [state, setState] = createStore(getInitialState());
  let hasLoggedStorageWriteError = false;

  createEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    try {
      window.localStorage.setItem(APP_STATE_STORAGE_KEY, JSON.stringify(getPersistedState(state)));
      hasLoggedStorageWriteError = false;
    } catch (error) {
      if (!hasLoggedStorageWriteError) {
        console.warn('Failed to persist app state to localStorage:', error);
        hasLoggedStorageWriteError = true;
      }
    }
  });

  return (
    <AppStoreContext.Provider value={{ state, setState }}>
      {props.children}
    </AppStoreContext.Provider>
  );
}

export function useAppStore() {
  const context = useContext(AppStoreContext);
  if (!context) {
    throw new Error('useAppStore must be used within an AppStoreProvider');
  }
  return context;
}
