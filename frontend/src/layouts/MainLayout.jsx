import Sidebar from '../components/Sidebar';
import Header from '../components/Header';
import { Show, Suspense, lazy } from 'solid-js';
import DuplicateModal from '../components/DuplicateModal';
import Icon from '../components/Icon';
import logo from '../assets/logo.png';
import { useAppStore } from '../store/appStore';
import { downloadStore, setDownloadStore } from '../store/downloadStore';
import { usePlayerController } from '../hooks/usePlayerController';
import { useGlobalShortcuts } from '../hooks/useGlobalShortcuts';

const Player = lazy(() => import('../components/Player'));

export default function MainLayout(props) {
    const { state, setState } = useAppStore();
    const { 
        playerQueue, 
        closePlayer, 
        playNextInQueue 
    } = usePlayerController();
    const { submitDuplicateChoice } = useGlobalShortcuts();

    const isAdvanced = () => state.ui.isAdvanced;
    const toggleAdvanced = () => setState('ui', 'isAdvanced', (prev) => !prev);

    return (
        <div class="flex h-screen bg-[radial-gradient(circle_at_12%_8%,rgba(56,189,248,0.16),transparent_35%),radial-gradient(circle_at_88%_2%,rgba(20,184,166,0.14),transparent_30%),linear-gradient(180deg,#05070a,#070b12_45%,#05070a)] text-gray-200 overflow-hidden font-sans select-none">
            <Sidebar />
            <main class="flex-1 flex flex-col bg-transparent relative min-w-0">
                <Header isAdvanced={isAdvanced()} onToggleAdvanced={toggleAdvanced} />
                <div class="flex-1 overflow-y-auto p-6 md:p-10 custom-scrollbar">
                    {props.children}

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
                            <div class={`px-6 py-4 rounded-2xl border backdrop-blur-xl shadow-2xl flex items-center gap-4 min-w-[340px] max-w-md ${downloadStore.notification.type === 'success'
                                    ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400'
                                    : 'bg-red-500/10 border-red-500/20 text-red-400'
                                }`}>
                                <div class="relative">
                                    <div class={`p-2.5 rounded-xl ${downloadStore.notification.type === 'success' ? 'bg-emerald-500/20' : 'bg-red-500/20'
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
