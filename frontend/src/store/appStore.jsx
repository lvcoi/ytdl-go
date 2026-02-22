import { createContext, createEffect, useContext } from 'solid-js';
import { createStore } from 'solid-js/store';

const APP_STATE_STORAGE_KEY = 'ytdl-go:app-state:v1';
const VALID_TABS = new Set(['download', 'library', 'settings', 'dashboard']);
const VALID_DUPLICATE_POLICIES = new Set(['prompt', 'overwrite', 'skip', 'rename']);
const VALID_LIBRARY_MEDIA_TYPES = new Set(['video', 'audio']);
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

const createDefaultState = () => ({
  ui: {
    activeTab: 'download',
    isAdvanced: false,
  },
  settings: { ...defaultSettings },
  library: {
    downloads: [],
    activeMediaType: 'video',
    filters: {
      creator: '',
      collection: '',
      playlist: '',
      savedPlaylistId: '',
    },
    sortKey: 'newest',
    savedPlaylists: [],
    playlistAssignments: {},
  },
  player: {
    active: false,
    selectedMedia: null,
  },
  dashboard: {
    widgets: [
      { id: 'quick-download', type: 'quick-download', col: 1, row: 1, colSpan: 4, rowSpan: 2 },
      { id: 'recent-downloads', type: 'recent-downloads', col: 5, row: 1, colSpan: 8, rowSpan: 4 },
      { id: 'system-stats', type: 'system-stats', col: 1, row: 3, colSpan: 4, rowSpan: 2 },
      { id: 'storage', type: 'storage', col: 1, row: 5, colSpan: 4, rowSpan: 2 },
    ]
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
  // Keep this allowlist intentionally narrow. Runtime download state
  // (`isDownloading`, `jobStatus`, task progress, logs, duplicate queue)
  // is transient and should not survive reloads.
  ui: {
    activeTab: state.ui.activeTab,
    isAdvanced: state.ui.isAdvanced,
  },
  settings: {
    ...state.settings,
  },
  library: {
    activeMediaType: state.library.activeMediaType,
    filters: {
      creator: state.library.filters.creator,
      collection: state.library.filters.collection,
      playlist: state.library.filters.playlist,
      savedPlaylistId: state.library.filters.savedPlaylistId,
    },
    sortKey: state.library.sortKey,
    savedPlaylists: state.library.savedPlaylists,
    playlistAssignments: state.library.playlistAssignments,
  },
  download: {
    urlInput: state.download.urlInput,
  },
  dashboard: {
    widgets: state.dashboard.widgets,
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

const sanitizeLibrary = (rawLibrary) => {
  const raw = rawLibrary && typeof rawLibrary === 'object' ? rawLibrary : {};

  const activeMediaType = toString(raw.activeMediaType, 'video');
  const sortKey = toString(raw.sortKey, 'newest');
  const savedPlaylists = sanitizeSavedPlaylists(raw.savedPlaylists);
  const validSavedPlaylistIds = new Set(savedPlaylists.map((playlist) => playlist.id));
  return {
    activeMediaType: VALID_LIBRARY_MEDIA_TYPES.has(activeMediaType) ? activeMediaType : 'video',
    filters: sanitizeLibraryFilters(raw.filters, validSavedPlaylistIds),
    sortKey: VALID_LIBRARY_SORT_KEYS.has(sortKey) ? sortKey : 'newest',
    savedPlaylists,
    playlistAssignments: sanitizePlaylistAssignments(raw.playlistAssignments, validSavedPlaylistIds),
  };
};

const sanitizeDashboardWidgets = (rawWidgets) => {
  if (!Array.isArray(rawWidgets)) {
    return null;
  }
  const out = [];
  for (const entry of rawWidgets) {
    const w = entry && typeof entry === 'object' ? entry : {};
    const id = toString(w.id, '').trim();
    const type = toString(w.type, '').trim();
    const col = typeof w.col === 'number' ? Math.trunc(w.col) : Number(w.col);
    const row = typeof w.row === 'number' ? Math.trunc(w.row) : Number(w.row);
    const colSpan = typeof w.colSpan === 'number' ? Math.trunc(w.colSpan) : Number(w.colSpan);
    const rowSpan = typeof w.rowSpan === 'number' ? Math.trunc(w.rowSpan) : Number(w.rowSpan);
    if (
      id === '' || type === '' ||
      !Number.isFinite(col) || col < 1 ||
      !Number.isFinite(row) || row < 1 ||
      !Number.isFinite(colSpan) || colSpan < 1 ||
      !Number.isFinite(rowSpan) || rowSpan < 1
    ) {
      continue;
    }
    out.push({ id, type, col, row, colSpan, rowSpan });
  }
  return out.length > 0 ? out : null;
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

  const persistedDashboardWidgets = sanitizeDashboardWidgets(persisted?.dashboard?.widgets)
    ?? baseState.dashboard.widgets;

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
      activeMediaType: persistedLibrary.activeMediaType,
      filters: persistedLibrary.filters,
      sortKey: persistedLibrary.sortKey,
      savedPlaylists: persistedLibrary.savedPlaylists,
      playlistAssignments: persistedLibrary.playlistAssignments,
    },
    download: {
      ...baseState.download,
      urlInput: persistedUrlInput,
    },
    dashboard: {
      widgets: persistedDashboardWidgets,
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
