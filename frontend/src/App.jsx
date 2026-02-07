import { createEffect, onCleanup, onMount, Show } from 'solid-js';
import Icon from './components/Icon';
import DownloadView from './components/DownloadView';
import LibraryView from './components/LibraryView';
import SettingsView from './components/SettingsView';
import Player from './components/Player';
import { useAppStore } from './store/appStore';
import { normalizeDownloadStatus } from './utils/downloadStatus';

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

function App() {
  const { state, setState } = useAppStore();
  let mediaListAbortController = null;
  let mediaListRequestToken = 0;
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
  });

  const openPlayer = (item) => {
    // Add the full media URL for the player
    const mediaItem = {
      ...item,
      url: `/api/media/${encodeMediaPath(item.filename)}`
    };
    setState('player', 'selectedMedia', mediaItem);
    setState('player', 'active', true);
  };

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

        <div class="mt-auto p-4 bg-white/5 rounded-2xl border border-white/5">
            <div class="flex items-center gap-2 mb-2 text-xs font-bold text-gray-500 uppercase tracking-widest">
                <Icon name="puzzle" class="w-3 h-3" />
                Extensions
            </div>
            <div class="flex items-center justify-between text-xs">
                <span class="text-gray-400">PO Token Provider</span>
                <span class="px-2 py-0.5 bg-green-500/10 text-green-500 rounded-full font-bold">Active</span>
            </div>
        </div>
      </aside>

      {/* Main Content */}
      <main class="flex-1 flex flex-col bg-[#05070a] relative">
        <header class="h-20 border-b border-white/5 flex items-center justify-between px-10 glass sticky top-0 z-20">
            <h2 class="text-lg font-bold text-white capitalize">{activeTab()}</h2>
            <div class="flex items-center gap-6">
                <div class="flex items-center gap-3 px-4 py-2 bg-white/5 rounded-full border border-white/5 has-tooltip cursor-pointer">
                    <span class="tooltip bg-gray-800 text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-48 text-center leading-relaxed">Logged in. Cookies synced for age-restricted content.</span>
                    <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                    <span class="text-xs font-bold text-gray-300 italic">YT_AUTH_OK</span>
                    <Icon name="chevron-down" class="w-3 h-3 text-gray-500" />
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
                      onMediaTypeChange={(nextType) => setState('library', 'activeMediaType', nextType)}
                      onFilterChange={(filterKey, value) => setState('library', 'filters', filterKey, value)}
                      onClearFilters={() => setState('library', 'filters', {
                        creator: '',
                        collection: '',
                        playlist: '',
                      })}
                      onSortKeyChange={(nextSortKey) => setState('library', 'sortKey', nextSortKey)}
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
             onClose={() => setState('player', 'active', false)} 
          />
        )}
      </main>
    </div>
  );
}

export default App;
