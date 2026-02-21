import { createEffect, createSignal, onCleanup, onMount, Show, createMemo, lazy, Suspense } from 'solid-js';
import Icon from './components/Icon';
import DashboardView from './components/DashboardView';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import DuplicateModal from './components/DuplicateModal';
import { useAppStore } from './store/appStore';
import { downloadStore, setDownloadStore } from './store/downloadStore';
import { normalizeDownloadStatus } from './utils/downloadStatus';
import logo from './assets/logo.png';



import { detectMediaType } from './utils/mediaType';
import { useSavedPlaylists } from './hooks/useSavedPlaylists';
import { useDownloadManager } from './hooks/useDownloadManager';
import { buildLibraryModel } from './utils/libraryModel';
import wsService from './services/websocket';

// Lazily load non-critical views

const DownloadView = lazy(() => import('./components/DownloadView'));
const LibraryView = lazy(() => import('./components/LibraryView'));
const SettingsView = lazy(() => import('./components/SettingsView'));
const Player = lazy(() => import('./components/Player'));


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

function App() {
  const { state, setState } = useAppStore();
  const [playerQueue, setPlayerQueue] = createSignal([]);

  /* Hook for Saved Playlists */
  const {
    initError: savedPlaylistInitError,
    initialize: initializeSavedPlaylists,
    createPlaylist: createSavedPlaylist,
    renamePlaylist: renameSavedPlaylist,
    deletePlaylist: deleteSavedPlaylist,
    assignPlaylist: assignSavedPlaylist
  } = useSavedPlaylists();

  /* Hook for Download Manager */
  const { listenForProgress, startDownload } = useDownloadManager();

  let mediaListAbortController = null;
  let mediaListRequestToken = 0;
  let lastSyncedLibraryJobId = '';
  let pendingLibrarySyncJobId = '';
  let librarySyncRetryTimer = null;
  let isDisposed = false;

    const activeTab = () => state.ui.activeTab || 'dashboard'; // Default to dashboard if not set
  const isAdvanced = () => state.ui.isAdvanced;

  // Memoized Library Model for Dashboard

  const libraryModel = createMemo(() => buildLibraryModel({
    downloads: state.library.downloads,
    savedPlaylists: state.library.savedPlaylists,
    playlistAssignments: state.library.playlistAssignments,
    typeFilter: state.library.typeFilter,
    sortKey: state.library.sortKey,
    filters: state.library.filters,
  }));

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

    const retryLibraryMetadataScan = async () => {
    const synced = await fetchMediaFiles();
    if (!synced) {
      return {
        ok: false,
        message: 'Unable to refresh metadata right now. Please retry in a few seconds.',
      };
    }
    return {
      ok: true,
      message: 'Library refreshed. Legacy files without sidecars may still need re-download for complete metadata and thumbnails.',
    };
  };

  const removeDuplicatePrompt = (promptId) => {
    if (!promptId) return;
    setState('download', 'duplicateQueue', (prev) => prev.filter((item) => item.promptId !== promptId));
  };

  const submitDuplicateChoice = async (choice) => {
    const prompt = state.download.duplicateQueue[0];
    if (!prompt) return;
    setState('download', 'duplicateError', '');
    try {
      const res = await fetch('/api/download/duplicate-response', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jobId: prompt.jobId,
          promptId: prompt.promptId,
          choice,
        }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || data.error) {
        if (res.status === 404 || res.status === 409) {
          removeDuplicatePrompt(prompt.promptId);
          setState('download', 'duplicateError', '');
          return;
        }
        setState('download', 'duplicateError', data.error || 'Failed to submit duplicate choice');
        return;
      }
      removeDuplicatePrompt(prompt.promptId);
      setState('download', 'duplicateError', '');
    } catch (e) {
      setState('download', 'duplicateError', e.message || 'Failed to submit duplicate choice');
    }
  };

    const handleDuplicateShortcut = (e) => {
    if (state.download.duplicateQueue.length === 0) return;
    
    // Check if user is typing in an input or textarea
    const activeElement = document.activeElement;
    const isTyping = activeElement && (
      activeElement.tagName === 'INPUT' || 
      activeElement.tagName === 'TEXTAREA' || 
      activeElement.isContentEditable
    );
    
    if (isTyping) return;
    if (e.metaKey || e.ctrlKey || e.altKey) return;

    switch (e.key) {

      case 'o': e.preventDefault(); submitDuplicateChoice('overwrite'); break;
      case 'O': e.preventDefault(); submitDuplicateChoice('overwrite_all'); break;
      case 's': e.preventDefault(); submitDuplicateChoice('skip'); break;
      case 'S': e.preventDefault(); submitDuplicateChoice('skip_all'); break;
      case 'r': e.preventDefault(); submitDuplicateChoice('rename'); break;
      case 'R': e.preventDefault(); submitDuplicateChoice('rename_all'); break;
      case 'q': e.preventDefault(); submitDuplicateChoice('cancel'); break;
      default: break;
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

      // Load media files when component mounts
  onMount(() => {
    void fetchMediaFiles();
    void initializeSavedPlaylists();
    wsService.connect();

    if (typeof window !== 'undefined') {

      window.addEventListener('keydown', handleDuplicateShortcut);
    }

    // Ensure activeTab is set to dashboard if empty (first load)
    if (!state.ui.activeTab) {
      setActiveTab('dashboard');
    }
  });


  // Refresh media files when switching to library tab (or dashboard?)
  createEffect(() => {
    const tab = activeTab();
    if (tab === 'library' || tab === 'dashboard') {
      void fetchMediaFiles();
    }
  });

    // Keep library data fresh after terminal download outcomes
  createEffect(() => {
    const statuses = downloadStore.jobStatuses;
    for (const jobId in statuses) {
      const job = statuses[jobId];
      if (typeof jobId !== 'string' || jobId === '') {
        continue;
      }
      if (jobId === lastSyncedLibraryJobId || jobId === pendingLibrarySyncJobId) {
        continue;
      }

      if (!shouldSyncLibraryForTerminalDownload(job?.status, job?.stats)) {
        continue;
      }
      void syncLibraryForJob(jobId, 1);
    }
  });


    onCleanup(() => {
    isDisposed = true;
    clearLibrarySyncRetryTimer();
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleDuplicateShortcut);
    }
    if (mediaListAbortController) {
      mediaListAbortController.abort();
      mediaListAbortController = null;
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
      // 'audio' and 'video' are mapped to 'Music' and 'YouTube Video' in UI but verify filter compatibility
      // For now, setting typeFilter to 'all' or specific type might depend on LibraryView implementation
      // Use 'all' safely or map if needed. LibraryView handles type mapping.
      setState('library', 'typeFilter', 'all');
    }
    setState('library', 'section', 'all_media');
    setState('library', 'navPath', {
      creatorType: '',
      creatorName: '',
      albumName: '',
      playlistKey: '',
      playlistKind: '',
    });
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

    return (
    <div class="flex h-screen bg-[radial-gradient(circle_at_12%_8%,rgba(56,189,248,0.16),transparent_35%),radial-gradient(circle_at_88%_2%,rgba(20,184,166,0.14),transparent_30%),linear-gradient(180deg,#05070a,#070b12_45%,#05070a)] text-gray-200 overflow-hidden font-sans select-none">

      <Sidebar activeTab={activeTab()} onTabChange={setActiveTab} />

      {/* Main Content */}
      <main class="flex-1 flex flex-col bg-transparent relative min-w-0">
        <Header
          activeTab={activeTab()}
          isAdvanced={isAdvanced()}
          onToggleAdvanced={toggleAdvanced}
        />

        <div class="flex-1 overflow-y-auto p-6 md:p-10 custom-scrollbar">
          <Suspense fallback={
            <div class="flex flex-col items-center justify-center h-full text-gray-500 gap-4 animate-pulse">
              <Icon name="loader" class="w-10 h-10 animate-spin text-accent-primary" />
              <span class="text-xs font-bold uppercase tracking-widest">Loading View...</span>
            </div>
          }>
            <Show when={activeTab() === 'dashboard'}>
              <div class="max-w-7xl mx-auto">
                                <DashboardView
                  libraryModel={libraryModel}
                  onTabChange={setActiveTab}
                  onDownload={startDownload}
                  onPlay={(item) => openPlayer(item, state.library.downloads)}
                />

              </div>
            </Show>

            <Show when={activeTab() === 'download'}>
              <div class="max-w-4xl mx-auto">
                <DownloadView
                  onOpenLibrary={openLibrary}
                  onStartDownload={(jobId) => listenForProgress(jobId)}
                />
              </div>
            </Show>

            <Show when={activeTab() === 'library'}>
              <div class="max-w-dynamic mx-auto h-full">
                <LibraryView
                  downloads={() => state.library.downloads}
                  section={() => state.library.section}
                  viewMode={() => state.library.viewMode}
                  typeFilter={() => state.library.typeFilter}
                  navPath={() => state.library.navPath}
                  filters={() => state.library.filters}
                  sortKey={() => state.library.sortKey}
                  uiState={() => state.library.ui}
                  savedPlaylists={() => state.library.savedPlaylists}
                  playlistAssignments={() => state.library.playlistAssignments}
                  onSectionChange={(nextSection) => setState('library', 'section', nextSection)}
                  onViewModeChange={(nextViewMode) => setState('library', 'viewMode', nextViewMode)}
                  onTypeFilterChange={(nextTypeFilter) => setState('library', 'typeFilter', nextTypeFilter)}
                  onNavPathChange={(nextNavPath) => setState('library', 'navPath', nextNavPath)}
                  onFilterChange={(filterKey, value) => setState('library', 'filters', filterKey, value)}
                  onClearFilters={() => setState('library', 'filters', {
                    query: '',
                    savedPlaylistId: '',
                  })}
                  onSortKeyChange={(nextSortKey) => setState('library', 'sortKey', nextSortKey)}
                                    onUiStateChange={(key, value) => setState('library', 'ui', key, value)}
                  openPlayer={(item, queue) => openPlayer(item, queue)}
                  onOpenQueue={openQueue}

                  onCreatePlaylist={createSavedPlaylist}
                  onRenamePlaylist={renameSavedPlaylist}
                  onDeletePlaylist={deleteSavedPlaylist}
                  onAssignPlaylist={assignSavedPlaylist}
                  onRetryMetadataScan={retryLibraryMetadataScan}
                />
              </div>
            </Show>

            <Show when={activeTab() === 'settings'}>
              <div class="max-w-4xl mx-auto">
                <SettingsView />
              </div>
            </Show>
          </Suspense>

          <Show when={state.download.duplicateQueue.length > 0}>
            <DuplicateModal
              prompt={state.download.duplicateQueue[0]}
              onSelect={submitDuplicateChoice}
              error={state.download.duplicateError}
            />
          </Show>

          {/* Global Notification Toast */}
          <Show when={downloadStore.notification}>
            <div class="fixed bottom-10 left-1/2 -translate-x-1/2 z-[100] animate-in fade-in slide-in-from-bottom-4 duration-500">
              <div class={`px-6 py-4 rounded-2xl border backdrop-blur-xl shadow-2xl flex items-center gap-4 min-w-[340px] max-w-md ${
                downloadStore.notification.type === 'success' 
                  ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' 
                  : 'bg-red-500/10 border-red-500/20 text-red-400'
              }`}>
                <div class="relative">
                  <div class={`p-2.5 rounded-xl ${
                    downloadStore.notification.type === 'success' ? 'bg-emerald-500/20' : 'bg-red-500/20'
                  }`}>
                    <Icon name={downloadStore.notification.type === 'success' ? 'check-circle-2' : 'alert-circle'} class="w-5 h-5 relative z-10" />
                  </div>
                  <Show when={downloadStore.notification.type === 'success'}>
                    <img src={logo} class="absolute -top-10 -left-4 w-12 h-12 object-contain animate-bounce -rotate-12 pointer-events-none drop-shadow-lg" alt="Gopher Success" />
                  </Show>
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-bold leading-tight line-clamp-2">{downloadStore.notification.message}</p>
                </div>
                <button 
                  onClick={() => setDownloadStore('notification', null)}
                  class="p-1 hover:bg-white/5 rounded-lg transition-colors"
                >
                  <Icon name="x" class="w-4 h-4 opacity-50" />
                </button>
              </div>
            </div>
          </Show>
        </div>
      </main>

      <Suspense>
        <Show when={state.player.active}>
          <Player
            active={state.player.active}
            minimized={state.player.minimized}
            media={state.player.selectedMedia}
            onClose={closePlayer}
            onMinimize={() => setState('player', 'minimized', true)}
            onRestore={() => setState('player', 'minimized', false)}
            onNext={playNextInQueue}
            onPrevious={playNextInQueue}
          />
        </Show>
      </Suspense>
    </div>
  );
}

export default App;




