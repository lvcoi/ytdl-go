import { createSignal, onMount, createEffect } from 'solid-js';
import Icon from './components/Icon';
import DownloadView from './components/DownloadView';
import LibraryView from './components/LibraryView';
import SettingsView from './components/SettingsView';
import Player from './components/Player';

function App() {
  const [activeTab, setActiveTab] = createSignal('download');
  const [isAdvanced, setIsAdvanced] = createSignal(false);
  const [downloads, setDownloads] = createSignal([]);
  const [playerActive, setPlayerActive] = createSignal(false);
  const [selectedMedia, setSelectedMedia] = createSignal(null);
  const [settings, setSettings] = createSignal({
    output: '{title}.{ext}',
    quality: 'best',
    jobs: 1,
    timeout: 180,
    format: '',
    audioOnly: false,
    onDuplicate: 'prompt',
    useCookies: true,
    poTokenExtension: false
  });

  // Fetch media files from the API
  const fetchMediaFiles = async () => {
    try {
      const response = await fetch('/api/media/');
      if (response.ok) {
        const payload = await response.json();
        const files = Array.isArray(payload) ? payload : (payload.items || []);
        setDownloads(files);
      }
    } catch (error) {
      console.error('Failed to fetch media files:', error);
    }
  };

  // Load media files when component mounts and when switching to library tab
  onMount(() => {
    fetchMediaFiles();
  });

  // Refresh media files when switching to library tab
  createEffect(() => {
    if (activeTab() === 'library') {
      fetchMediaFiles();
    }
  });

  const openPlayer = (item) => {
    // Add the full media URL for the player
    const mediaItem = {
      ...item,
      url: `/api/media/${item.filename}`
    };
    setSelectedMedia(mediaItem);
    setPlayerActive(true);
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
                <button onClick={() => setIsAdvanced(!isAdvanced())} class={`px-4 py-2 rounded-full text-xs font-bold transition-all ${isAdvanced() ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/20' : 'bg-white/5 text-gray-500 hover:text-gray-300'}`}>
                    Advanced Mode
                </button>
            </div>
        </header>

        <div class="flex-1 overflow-y-auto p-10 custom-scrollbar">
            <div class="max-w-4xl mx-auto">
                {activeTab() === 'download' && (
                    <DownloadView 
                        settings={settings} 
                        setSettings={setSettings} 
                        isAdvanced={isAdvanced}
                    />
                )}
                {activeTab() === 'library' && (
                    <LibraryView 
                        downloads={downloads} 
                        openPlayer={openPlayer} 
                    />
                )}
                {activeTab() === 'settings' && <SettingsView />}
            </div>
        </div>
        
        {playerActive() && (
          <Player 
             media={selectedMedia} 
             onClose={() => setPlayerActive(false)} 
          />
        )}
      </main>
    </div>
  );
}

export default App;
