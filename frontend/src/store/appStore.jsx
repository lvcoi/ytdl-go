import { createContext, createEffect, useContext } from 'solid-js';
import { createStore } from 'solid-js/store';

const APP_STATE_STORAGE_KEY = 'ytdl-go:app-state:v1';
const VALID_TABS = new Set(['download', 'library', 'settings']);
const VALID_DUPLICATE_POLICIES = new Set(['prompt', 'overwrite', 'skip', 'rename']);
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
  },
  player: {
    active: false,
    selectedMedia: null,
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
  download: {
    urlInput: state.download.urlInput,
  },
});

const toString = (value, fallback) => (typeof value === 'string' ? value : fallback);
const toBoolean = (value, fallback) => (typeof value === 'boolean' ? value : fallback);
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
  const persistedUrlInput = typeof persisted?.download?.urlInput === 'string'
    ? persisted.download.urlInput
    : baseState.download.urlInput;

  return {
    ...baseState,
    ui: {
      activeTab,
      isAdvanced,
    },
    settings: {
      ...persistedSettings,
    },
    download: {
      ...baseState.download,
      urlInput: persistedUrlInput,
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
