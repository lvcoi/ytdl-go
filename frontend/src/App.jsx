import { createEffect, createSignal, onCleanup, onMount, Show } from 'solid-js';
import Icon from './components/Icon';
import DownloadView from './components/DownloadView';
import LibraryView from './components/LibraryView';
import SettingsView from './components/SettingsView';
import Player from './components/Player';
import { useAppStore } from './store/appStore';
import { normalizeDownloadStatus } from './utils/downloadStatus';
import { detectMediaType } from './utils/mediaType';

const toSucceededCount = (stats) => {
  if (!stats || typeof stats !== 'object') return 0;
  const parsed = Number(stats.succeeded);
  if (!Number.isFinite(parsed)) return 0;
  return Math.max(0, Math.trunc(parsed));
};

const shouldSyncLibraryForTerminalDownload = (status, stats) => {
  const normalized = normalizeDownloadStatus(status);
  if (normalized === 'complete') return true;
  if (normalized !== 'error') return false;
  return toSucceededCount(stats) > 0;
};

const encodeMediaPath = (relativePath) => (
  String(relativePath || '')
    .split('/')
    .map((segment) => encodeURIComponent(segment))
    .join('/')
);
const normalizeQueueKey = (value) => String(value || '').trim();

const MAX_SAVED_PLAYLIST_NAME_LENGTH = 80;
const SAVED_PLAYLISTS_ENDPOINT = '/api/library/playlists';
const SAVED_PLAYLISTS_MIGRATION_ENDPOINT = '/api/library/playlists/migrate';
const SAVED_PLAYLISTS_MIGRATION_KEY = 'ytdl-go:saved-playlists:backend-migration:v1';

const normalizeSavedPlaylistName = (value) => (
  String(value || '')
    .trim()
    .replace(/\s+/g, ' ')
    .slice(0, MAX_SAVED_PLAYLIST_NAME_LENGTH)
);
const normalizeSavedPlaylistId = (value) => String(value || '').trim();
const normalizeMediaKey = (value) => String(value || '').trim();
const normalizeSavedPlaylistEntry = (value) => {
  const source = value && typeof value === 'object' ? value : {};
  const id = normalizeSavedPlaylistId(source.id);
  const name = normalizeSavedPlaylistName(source.name);
  if (id === '' || name === '') {
    return null;
  }
  return {
    id,
    name,
    createdAt: typeof source.createdAt === 'string' ? source.createdAt.trim() : '',
    updatedAt: typeof source.updatedAt === 'string' ? source.updatedAt.trim() : '',
  };
};
const normalizeSavedPlaylists = (value) => {
  if (!Array.isArray(value)) {
    return [];
  }
  const out = [];
  const seenIds = new Set();
  const seenNames = new Set();
  for (const entry of value) {
    const normalized = normalizeSavedPlaylistEntry(entry);
    if (!normalized) {
      continue;
    }
    if (seenIds.has(normalized.id)) {
      continue;
    }
    const nameKey = normalized.name.toLowerCase();
    if (seenNames.has(nameKey)) {
      continue;
    }
    seenIds.add(normalized.id);
    seenNames.add(nameKey);
    out.push(normalized);
  }
  return out;
};
const normalizePlaylistAssignments = (value, validPlaylistIds) => {
  const source = value && typeof value === 'object' ? value : {};
  const out = {};
  for (const [mediaKey, playlistId] of Object.entries(source)) {
    const normalizedMediaKey = normalizeMediaKey(mediaKey);
    const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
    if (normalizedMediaKey === '' || normalizedPlaylistId === '' || !validPlaylistIds.has(normalizedPlaylistId)) {
      continue;
    }
    out[normalizedMediaKey] = normalizedPlaylistId;
  }
  return out;
};
const normalizeSavedPlaylistStatePayload = (value) => {
  const source = value && typeof value === 'object' ? value : {};
  const playlists = normalizeSavedPlaylists(source.playlists);
  const validPlaylistIds = new Set(playlists.map((playlist) => playlist.id));
  return {
    playlists,
    assignments: normalizePlaylistAssignments(source.assignments, validPlaylistIds),
  };
};
const hasSavedPlaylistStateData = (value) => (
  Array.isArray(value?.playlists) && value.playlists.length > 0
) || (
  value?.assignments && typeof value.assignments === 'object' && Object.keys(value.assignments).length > 0
);
const hasSavedPlaylistNameConflict = (playlists, playlistName, excludedId = '') => (
  playlists.some((playlist) => (
    playlist.id !== excludedId &&
    playlist.name.localeCompare(playlistName, undefined, { sensitivity: 'base' }) === 0
  ))
);
const createSavedPlaylistId = () => {
  if (typeof globalThis !== 'undefined' && globalThis.crypto && typeof globalThis.crypto.randomUUID === 'function') {
    return globalThis.crypto.randomUUID();
  }
  return `saved-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
};
const responseErrorMessage = async (response, fallbackMessage) => {
  try {
    const payload = await response.json();
    if (payload && typeof payload.error === 'string' && payload.error.trim() !== '') {
      return payload.error.trim();
    }
  } catch (error) {
    // Ignore JSON parse issues and use fallback.
  }
  return fallbackMessage;
};
const isAbortError = (error) => (
  error && typeof error === 'object' && error.name === 'AbortError'
);

function App() {
  const { state, setState } = useAppStore();
  const [savedPlaylistInitError, setSavedPlaylistInitError] = createSignal('');
  const [playerQueue, setPlayerQueue] = createSignal([]);
  let mediaListAbortController = null;
  let savedPlaylistInitAbortController = null;
  let savedPlaylistPersistAbortController = null;
  let mediaListRequestToken = 0;
  let savedPlaylistMutationVersion = 0;
  let savedPlaylistMutationQueue = Promise.resolve();
  let lastSyncedLibraryJobId = '';
  let pendingLibrarySyncJobId = '';
  let librarySyncRetryTimer = null;
  let isDisposed = false;

  const activeTab = () => state.ui.activeTab;
  const isAdvanced = () => state.ui.isAdvanced;
  const downloadJobStatus = () => state.download.jobStatus;

  const setActiveTab = (tab) => {
    setState('ui', 'activeTab', tab);
  };

  const toggleAdvanced = () => {
    setState('ui', 'isAdvanced', (prev) => !prev);
  };

  const openLibrary = () => {
    setActiveTab('library');
  };

  const clearLibrarySyncRetryTimer = () => {
    if (librarySyncRetryTimer) {
      clearTimeout(librarySyncRetryTimer);
      librarySyncRetryTimer = null;
    }
  };

  const enqueueSavedPlaylistMutation = (mutateFn) => {
    const next = savedPlaylistMutationQueue.then(mutateFn, mutateFn);
    savedPlaylistMutationQueue = next.catch(() => {});
    return next;
  };

  const hasCompletedSavedPlaylistMigration = () => {
    if (typeof window === 'undefined') {
      return false;
    }
    try {
      return window.localStorage.getItem(SAVED_PLAYLISTS_MIGRATION_KEY) === '1';
    } catch (error) {
      return false;
    }
  };

  const markSavedPlaylistMigrationComplete = () => {
    if (typeof window === 'undefined') {
      return;
    }
    try {
      window.localStorage.setItem(SAVED_PLAYLISTS_MIGRATION_KEY, '1');
    } catch (error) {
      // Ignore storage write failures; backend remains source of truth.
    }
  };

  const getSavedPlaylistStateSnapshot = () => normalizeSavedPlaylistStatePayload({
    playlists: state.library.savedPlaylists,
    assignments: state.library.playlistAssignments,
  });

  const applySavedPlaylistState = (nextValue) => {
    const normalized = normalizeSavedPlaylistStatePayload(nextValue);
    const validIds = new Set(normalized.playlists.map((playlist) => playlist.id));
    setState('library', 'savedPlaylists', normalized.playlists);
    setState('library', 'playlistAssignments', normalized.assignments);
    setState('library', 'filters', 'savedPlaylistId', (current) => (
      validIds.has(normalizeSavedPlaylistId(current)) ? current : ''
    ));
  };

  const fetchSavedPlaylistState = async (signal) => {
    const response = await fetch(SAVED_PLAYLISTS_ENDPOINT, { signal });
    if (!response.ok) {
      throw new Error(await responseErrorMessage(response, 'Unable to load saved playlists from backend.'));
    }
    const payload = await response.json();
    return normalizeSavedPlaylistStatePayload(payload);
  };

  const persistSavedPlaylistState = async (nextState, signal) => {
    const normalized = normalizeSavedPlaylistStatePayload(nextState);
    const response = await fetch(SAVED_PLAYLISTS_ENDPOINT, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(normalized),
      signal,
    });
    if (!response.ok) {
      throw new Error(await responseErrorMessage(response, 'Unable to persist saved playlists.'));
    }
    const payload = await response.json();
    return normalizeSavedPlaylistStatePayload(payload);
  };

  const migrateSavedPlaylistState = async (legacyState, signal) => {
    const normalizedLegacy = normalizeSavedPlaylistStatePayload(legacyState);
    const response = await fetch(SAVED_PLAYLISTS_MIGRATION_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(normalizedLegacy),
      signal,
    });
    if (!response.ok) {
      throw new Error(await responseErrorMessage(response, 'Unable to migrate saved playlists.'));
    }
    const payload = await response.json();
    return {
      state: normalizeSavedPlaylistStatePayload(payload),
      migrated: Boolean(payload && payload.migrated),
    };
  };

  const persistAndApplySavedPlaylistState = async (nextState) => {
    savedPlaylistMutationVersion += 1;
    const mutationVersion = savedPlaylistMutationVersion;
    if (savedPlaylistPersistAbortController) {
      savedPlaylistPersistAbortController.abort();
    }
    const controller = new AbortController();
    savedPlaylistPersistAbortController = controller;
    try {
      const persisted = await persistSavedPlaylistState(nextState, controller.signal);
      if (!isDisposed && mutationVersion === savedPlaylistMutationVersion) {
        applySavedPlaylistState(persisted);
        setSavedPlaylistInitError('');
      }
      return persisted;
    } finally {
      if (savedPlaylistPersistAbortController === controller) {
        savedPlaylistPersistAbortController = null;
      }
    }
  };

  const initializeSavedPlaylists = async () => {
    if (savedPlaylistInitAbortController) {
      savedPlaylistInitAbortController.abort();
    }
    const controller = new AbortController();
    savedPlaylistInitAbortController = controller;
    const startMutationVersion = savedPlaylistMutationVersion;
    const legacySnapshot = getSavedPlaylistStateSnapshot();
    try {
      const backendState = await fetchSavedPlaylistState(controller.signal);
      if (isDisposed || controller.signal.aborted || savedPlaylistMutationVersion !== startMutationVersion) {
        return;
      }
      applySavedPlaylistState(backendState);
      setSavedPlaylistInitError('');

      if (hasSavedPlaylistStateData(backendState)) {
        markSavedPlaylistMigrationComplete();
        return;
      }
      if (!hasSavedPlaylistStateData(legacySnapshot)) {
        if (!hasCompletedSavedPlaylistMigration()) {
          markSavedPlaylistMigrationComplete();
        }
        return;
      }

      const migration = await migrateSavedPlaylistState(legacySnapshot, controller.signal);
      if (isDisposed || controller.signal.aborted || savedPlaylistMutationVersion !== startMutationVersion) {
        return;
      }
      applySavedPlaylistState(migration.state);
      if (migration.migrated || hasSavedPlaylistStateData(migration.state)) {
        markSavedPlaylistMigrationComplete();
      }
    } catch (error) {
      if (isAbortError(error) || (controller.signal && controller.signal.aborted)) {
        return;
      }
      const message = error instanceof Error ? error.message : 'Failed to initialize saved playlists.';
      setSavedPlaylistInitError(message);
      console.error('Failed to initialize saved playlists:', error);
    } finally {
      if (savedPlaylistInitAbortController === controller) {
        savedPlaylistInitAbortController = null;
      }
    }
  };

  // Fetch media files from the API
  const fetchMediaFiles = async () => {
    if (isDisposed) {
      return false;
    }
    if (mediaListAbortController) {
      mediaListAbortController.abort();
    }
    mediaListAbortController = new AbortController();
    const requestToken = ++mediaListRequestToken;

    try {
      const response = await fetch('/api/media/', { signal: mediaListAbortController.signal });
      if (response.ok) {
        const payload = await response.json();
        if (isDisposed || requestToken !== mediaListRequestToken) {
          return false;
        }
        const files = Array.isArray(payload) ? payload : (payload.items || []);
        setState('library', 'downloads', files);
        return true;
      }
      return false;
    } catch (error) {
      if (error && typeof error === 'object' && error.name === 'AbortError') {
        return false;
      }
      console.error('Failed to fetch media files:', error);
      return false;
    } finally {
      if (requestToken === mediaListRequestToken) {
        mediaListAbortController = null;
      }
    }
  };

  const syncLibraryForJob = async (jobId, attempt = 1) => {
    if (!jobId || isDisposed) {
      return;
    }
    pendingLibrarySyncJobId = jobId;
    const synced = await fetchMediaFiles();

    if (isDisposed || pendingLibrarySyncJobId !== jobId) {
      return;
    }

    if (synced) {
      lastSyncedLibraryJobId = jobId;
      pendingLibrarySyncJobId = '';
      clearLibrarySyncRetryTimer();
      return;
    }

    if (attempt >= 3) {
      pendingLibrarySyncJobId = '';
      return;
    }

    clearLibrarySyncRetryTimer();
    librarySyncRetryTimer = setTimeout(() => {
      librarySyncRetryTimer = null;
      void syncLibraryForJob(jobId, attempt + 1);
    }, attempt * 1000);
  };

  // Load media files when component mounts and when switching to library tab
  onMount(() => {
    void fetchMediaFiles();
    void initializeSavedPlaylists();
  });

  // Refresh media files when switching to library tab
  createEffect(() => {
    if (activeTab() === 'library') {
      void fetchMediaFiles();
    }
  });

  // Keep library data fresh after terminal download outcomes, even when
  // the user stays on the download tab.
  createEffect(() => {
    const currentJobStatus = downloadJobStatus();
    const jobId = currentJobStatus?.jobId;
    if (typeof jobId !== 'string' || jobId === '') {
      return;
    }
    if (jobId === lastSyncedLibraryJobId || jobId === pendingLibrarySyncJobId) {
      return;
    }

    if (!shouldSyncLibraryForTerminalDownload(currentJobStatus?.status, currentJobStatus?.stats)) {
      return;
    }
    void syncLibraryForJob(jobId, 1);
  });

  onCleanup(() => {
    isDisposed = true;
    clearLibrarySyncRetryTimer();
    if (mediaListAbortController) {
      mediaListAbortController.abort();
      mediaListAbortController = null;
    }
    if (savedPlaylistInitAbortController) {
      savedPlaylistInitAbortController.abort();
      savedPlaylistInitAbortController = null;
    }
    if (savedPlaylistPersistAbortController) {
      savedPlaylistPersistAbortController.abort();
      savedPlaylistPersistAbortController = null;
    }
  });

  const toPlayerMediaItem = (item) => ({
    ...item,
    url: `/api/media/${encodeMediaPath(item.filename)}`,
  });

  const toQueueItems = (candidateItems, anchorItem) => {
    const fallback = Array.isArray(state.library.downloads) ? state.library.downloads : [];
    const source = Array.isArray(candidateItems) && candidateItems.length > 0 ? candidateItems : fallback;
    const uniqueItems = [];
    const seen = new Set();
    for (const entry of source) {
      if (!entry || typeof entry !== 'object') {
        continue;
      }
      const queueKey = normalizeQueueKey(entry.filename);
      if (queueKey === '' || seen.has(queueKey)) {
        continue;
      }
      seen.add(queueKey);
      uniqueItems.push(entry);
    }
    const anchorKey = normalizeQueueKey(anchorItem?.filename);
    if (anchorKey !== '' && !seen.has(anchorKey)) {
      uniqueItems.unshift(anchorItem);
    }
    return uniqueItems;
  };

  const openPlayer = (item, queueItems) => {
    if (!item || typeof item !== 'object' || String(item.filename || '').trim() === '') {
      return;
    }
    const preparedQueue = toQueueItems(queueItems, item);
    setPlayerQueue(preparedQueue);
    setState('player', 'selectedMedia', toPlayerMediaItem(item));
    setState('player', 'minimized', false);
    setState('player', 'active', true);
  };

  const closePlayer = () => {
    setState('player', 'active', false);
    setState('player', 'selectedMedia', null);
    setState('player', 'minimized', false);
    setPlayerQueue([]);
  };

  const openQueue = () => {
    const selected = state.player.selectedMedia;
    if (selected) {
      const mediaType = detectMediaType(selected);
      if (mediaType === 'audio' || mediaType === 'video') {
        setState('library', 'activeMediaType', mediaType);
      }
    }
    setActiveTab('library');
  };

  const playNextInQueue = () => {
    const queue = playerQueue();
    if (queue.length === 0) {
      return;
    }
    const activeFilename = normalizeQueueKey(state.player.selectedMedia?.filename);
    const currentIndex = queue.findIndex((entry) => normalizeQueueKey(entry.filename) === activeFilename);
    const nextIndex = currentIndex < 0 ? 0 : (currentIndex + 1) % queue.length;
    const nextItem = queue[nextIndex];
    if (!nextItem) {
      return;
    }
    setState('player', 'selectedMedia', toPlayerMediaItem(nextItem));
    setState('player', 'active', true);
  };

  const createSavedPlaylist = (rawName) => enqueueSavedPlaylistMutation(async () => {
    const normalizedName = normalizeSavedPlaylistName(rawName);
    if (normalizedName === '') {
      return { ok: false, error: 'Playlist name is required.' };
    }

    const current = getSavedPlaylistStateSnapshot();
    if (hasSavedPlaylistNameConflict(current.playlists, normalizedName)) {
      return { ok: false, error: 'A playlist with that name already exists.' };
    }

    const timestamp = new Date().toISOString();
    const createdPlaylist = {
      id: createSavedPlaylistId(),
      name: normalizedName,
      createdAt: timestamp,
      updatedAt: timestamp,
    };

    try {
      const persisted = await persistAndApplySavedPlaylistState({
        playlists: [...current.playlists, createdPlaylist],
        assignments: current.assignments,
      });
      const persistedPlaylist = persisted.playlists.find((playlist) => playlist.id === createdPlaylist.id);
      if (!persistedPlaylist) {
        return { ok: false, error: 'A playlist with that name already exists.' };
      }
      return { ok: true, playlist: persistedPlaylist };
    } catch (error) {
      if (isAbortError(error)) {
        return { ok: false, error: 'Request canceled.' };
      }
      return { ok: false, error: error instanceof Error ? error.message : 'Unable to create saved playlist.' };
    }
  });

  const renameSavedPlaylist = (playlistId, rawName) => enqueueSavedPlaylistMutation(async () => {
    const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
    const normalizedName = normalizeSavedPlaylistName(rawName);
    if (normalizedPlaylistId === '') {
      return { ok: false, error: 'Playlist not found.' };
    }
    if (normalizedName === '') {
      return { ok: false, error: 'Playlist name is required.' };
    }

    const current = getSavedPlaylistStateSnapshot();
    const exists = current.playlists.some((playlist) => playlist.id === normalizedPlaylistId);
    if (!exists) {
      return { ok: false, error: 'Playlist not found.' };
    }
    if (hasSavedPlaylistNameConflict(current.playlists, normalizedName, normalizedPlaylistId)) {
      return { ok: false, error: 'A playlist with that name already exists.' };
    }

    const updatedAt = new Date().toISOString();
    const nextPlaylists = current.playlists.map((playlist) => (
      playlist.id === normalizedPlaylistId
        ? { ...playlist, name: normalizedName, updatedAt }
        : playlist
    ));

    try {
      const persisted = await persistAndApplySavedPlaylistState({
        playlists: nextPlaylists,
        assignments: current.assignments,
      });
      const persistedPlaylist = persisted.playlists.find((playlist) => playlist.id === normalizedPlaylistId);
      if (!persistedPlaylist) {
        return { ok: false, error: 'Playlist not found.' };
      }
      if (persistedPlaylist.name.localeCompare(normalizedName, undefined, { sensitivity: 'base' }) !== 0) {
        return { ok: false, error: 'A playlist with that name already exists.' };
      }
      return { ok: true };
    } catch (error) {
      if (isAbortError(error)) {
        return { ok: false, error: 'Request canceled.' };
      }
      return { ok: false, error: error instanceof Error ? error.message : 'Unable to rename saved playlist.' };
    }
  });

  const deleteSavedPlaylist = (playlistId) => enqueueSavedPlaylistMutation(async () => {
    const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
    if (normalizedPlaylistId === '') {
      return { ok: false, error: 'Playlist not found.' };
    }

    const current = getSavedPlaylistStateSnapshot();
    const nextPlaylists = current.playlists.filter((playlist) => playlist.id !== normalizedPlaylistId);
    if (nextPlaylists.length === current.playlists.length) {
      return { ok: false, error: 'Playlist not found.' };
    }

    const nextAssignments = {};
    for (const [mediaKey, assignedPlaylistId] of Object.entries(current.assignments)) {
      if (assignedPlaylistId === normalizedPlaylistId) {
        continue;
      }
      nextAssignments[mediaKey] = assignedPlaylistId;
    }

    try {
      await persistAndApplySavedPlaylistState({
        playlists: nextPlaylists,
        assignments: nextAssignments,
      });
      return { ok: true };
    } catch (error) {
      if (isAbortError(error)) {
        return { ok: false, error: 'Request canceled.' };
      }
      return { ok: false, error: error instanceof Error ? error.message : 'Unable to delete saved playlist.' };
    }
  });

  const assignSavedPlaylist = (mediaKey, playlistId) => enqueueSavedPlaylistMutation(async () => {
    const normalizedMediaKey = normalizeMediaKey(mediaKey);
    const normalizedPlaylistId = normalizeSavedPlaylistId(playlistId);
    if (normalizedMediaKey === '') {
      return;
    }

    const current = getSavedPlaylistStateSnapshot();
    const playlistExists = normalizedPlaylistId === '' || current.playlists.some(
      (playlist) => playlist.id === normalizedPlaylistId,
    );
    if (!playlistExists) {
      return;
    }

    const nextAssignments = { ...current.assignments };
    if (normalizedPlaylistId === '') {
      if (!(normalizedMediaKey in nextAssignments)) {
        return;
      }
      delete nextAssignments[normalizedMediaKey];
    } else {
      if (nextAssignments[normalizedMediaKey] === normalizedPlaylistId) {
        return;
      }
      nextAssignments[normalizedMediaKey] = normalizedPlaylistId;
    }

    try {
      await persistAndApplySavedPlaylistState({
        playlists: current.playlists,
        assignments: nextAssignments,
      });
    } catch (error) {
      if (isAbortError(error)) {
        return;
      }
      console.error('Failed to assign saved playlist:', error);
    }
  });

  return (
    <div class="flex h-screen bg-[#05070a] text-gray-200 overflow-hidden font-sans select-none">
      {/* Sidebar */}
      <aside class="w-72 bg-[#0a0c14] border-r border-white/5 flex flex-col p-6">
        <div class="flex items-center gap-3 mb-10 px-2">
            <div class="w-10 h-10 bg-blue-600 rounded-2xl flex items-center justify-center shadow-lg shadow-blue-600/20">
                <Icon name="zap" class="w-6 h-6 text-white fill-white" />
            </div>
            <span class="text-xl font-bold tracking-tight text-white">ytdl-go</span>
        </div>

        <nav class="flex-1 space-y-1">
            <button onClick={() => setActiveTab('download')} class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${activeTab() === 'download' ? 'bg-blue-600/10 text-blue-400' : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'}`}>
                <Icon name="plus-circle" class="w-5 h-5" />
                <span class="font-semibold text-sm">New Download</span>
            </button>
            <button onClick={() => setActiveTab('library')} class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${activeTab() === 'library' ? 'bg-blue-600/10 text-blue-400' : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'}`}>
                <Icon name="layers" class="w-5 h-5" />
                <span class="font-semibold text-sm">Library</span>
            </button>
            <button onClick={() => setActiveTab('settings')} class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${activeTab() === 'settings' ? 'bg-blue-600/10 text-blue-400' : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'}`}>
                <Icon name="sliders" class="w-5 h-5" />
                <span class="font-semibold text-sm">Configurations</span>
            </button>
        </nav>

        <div class="mt-auto relative has-tooltip">
            <span class="tooltip bg-gray-800 text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-56 text-center leading-relaxed">
                Coming Soon: extension provider management and health status.
            </span>
            <button
              type="button"
              disabled
              aria-disabled="true"
              aria-label="Extensions panel (Coming Soon)"
              class="w-full p-4 bg-white/5 rounded-2xl border border-white/5 opacity-70 cursor-not-allowed text-left"
            >
              <div class="flex items-center gap-2 mb-2 text-xs font-bold text-gray-500 uppercase tracking-widest">
                  <Icon name="puzzle" class="w-3 h-3" />
                  Extensions
              </div>
              <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-500">PO Token Provider</span>
                  <span class="px-2 py-0.5 bg-white/10 text-gray-500 rounded-full font-bold border border-white/10">Coming Soon</span>
              </div>
            </button>
        </div>
      </aside>

      {/* Main Content */}
      <main class="flex-1 flex flex-col bg-[#05070a] relative">
        <header class="h-20 border-b border-white/5 flex items-center justify-between px-10 glass sticky top-0 z-20">
            <h2 class="text-lg font-bold text-white capitalize">{activeTab()}</h2>
            <div class="flex items-center gap-6">
                <div class="relative has-tooltip">
                  <span class="tooltip bg-gray-800 text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-56 text-center leading-relaxed">
                    Coming Soon: live auth cookie status and diagnostics.
                  </span>
                  <button
                    type="button"
                    disabled
                    aria-disabled="true"
                    aria-label="YouTube auth status details (Coming Soon)"
                    class="flex items-center gap-3 px-4 py-2 bg-white/5 rounded-full border border-white/5 cursor-not-allowed opacity-70"
                  >
                    <div class="w-2 h-2 bg-gray-500 rounded-full"></div>
                    <span class="text-xs font-bold text-gray-400 italic">YT_AUTH_OK</span>
                    <Icon name="chevron-down" class="w-3 h-3 text-gray-500" />
                  </button>
                </div>
                <button onClick={toggleAdvanced} class={`px-4 py-2 rounded-full text-xs font-bold transition-all ${isAdvanced() ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/20' : 'bg-white/5 text-gray-500 hover:text-gray-300'}`}>
                    Advanced Mode
                </button>
            </div>
        </header>

        <div class="flex-1 overflow-y-auto p-10 custom-scrollbar">
            <div class={`max-w-4xl mx-auto ${activeTab() === 'download' ? '' : 'hidden'}`}>
                <DownloadView onOpenLibrary={openLibrary} />
            </div>
            <Show when={activeTab() === 'library'}>
              <div class="max-w-4xl mx-auto">
                  <LibraryView
                      downloads={() => state.library.downloads}
                      activeMediaType={() => state.library.activeMediaType}
                      filters={() => state.library.filters}
                      sortKey={() => state.library.sortKey}
                      savedPlaylists={() => state.library.savedPlaylists}
                      playlistAssignments={() => state.library.playlistAssignments}
                      onMediaTypeChange={(nextType) => setState('library', 'activeMediaType', nextType)}
                      onFilterChange={(filterKey, value) => setState('library', 'filters', filterKey, value)}
                      onClearFilters={() => setState('library', 'filters', {
                        creator: '',
                        collection: '',
                        playlist: '',
                        savedPlaylistId: '',
                      })}
                      onSortKeyChange={(nextSortKey) => setState('library', 'sortKey', nextSortKey)}
                      onCreateSavedPlaylist={createSavedPlaylist}
                      onRenameSavedPlaylist={renameSavedPlaylist}
                      onDeleteSavedPlaylist={deleteSavedPlaylist}
                      onAssignSavedPlaylist={assignSavedPlaylist}
                      savedPlaylistSyncError={() => savedPlaylistInitError()}
                      onRetrySavedPlaylistSync={() => { void initializeSavedPlaylists(); }}
                      openPlayer={openPlayer}
                  />
              </div>
            </Show>
            <Show when={activeTab() === 'settings'}>
              <div class="max-w-4xl mx-auto">
                  <SettingsView />
              </div>
            </Show>
        </div>
        
        {state.player.active && (
          <Player 
             media={state.player.selectedMedia} 
             minimized={() => state.player.minimized}
             queueCount={() => playerQueue().length}
             canGoNext={() => playerQueue().length > 1}
             onMinimize={() => setState('player', 'minimized', true)}
             onRestore={() => setState('player', 'minimized', false)}
             onNext={playNextInQueue}
             onQueue={openQueue}
             onClose={closePlayer}
          />
        )}
      </main>
    </div>
  );
}

export default App;
