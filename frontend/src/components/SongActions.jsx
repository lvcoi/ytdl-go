import { createSignal, Show, For, onCleanup } from 'solid-js';
import Icon from './Icon';
import Tooltip from './Tooltip';
import { useAppStore } from '../store/appStore';
import { useQueueManager } from '../hooks/useQueueManager';

export default function SongActions(props) {
    const { state } = useAppStore();
    const { addToQueue, addToSavedPlaylist } = useQueueManager();
    const [showPlaylistDropdown, setShowPlaylistDropdown] = createSignal(false);
    let dropdownRef;

    const handleClickOutside = (e) => {
        if (dropdownRef && !dropdownRef.contains(e.target)) {
            setShowPlaylistDropdown(false);
        }
    };

    const openDropdown = () => {
        setShowPlaylistDropdown(true);
        document.addEventListener('click', handleClickOutside, { once: true });
    };

    onCleanup(() => {
        document.removeEventListener('click', handleClickOutside);
    });

    const handleAddToPlaylist = (playlistId) => {
        const mediaKey = props.media?.filepath || props.media?.id || '';
        addToSavedPlaylist(mediaKey, playlistId);
        setShowPlaylistDropdown(false);
    };

    const handleAddToQueue = (e) => {
        e.stopPropagation();
        addToQueue(props.media);
    };

    return (
        <div class="flex items-center gap-1">
            {/* Add to Playlist */}
            <div class="relative" ref={dropdownRef}>
                <Tooltip text="Add to Playlist">
                    <button
                        onClick={(e) => { e.stopPropagation(); openDropdown(); }}
                        aria-label="Add to Playlist"
                        class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-white/10
                                transition-colors duration-150 focus:outline-none focus:ring-1
                                focus:ring-white/20"
                    >
                        <Icon name="list-plus" class="w-4 h-4" />
                    </button>
                </Tooltip>

                <Show when={showPlaylistDropdown()}>
                    <div class="absolute right-0 top-full mt-1 z-50 min-w-[200px] max-h-[240px]
                                 overflow-y-auto rounded-xl border border-white/10 bg-gray-900/95
                                 backdrop-blur-xl shadow-2xl animate-in fade-in zoom-in-95 duration-150
                                custom-scrollbar">
                        <div class="p-1.5">
                            <Show
                                when={state.library.savedPlaylists.length > 0}
                                fallback={
                                    <p class="px-3 py-2 text-xs text-gray-500 text-center">
                                        No playlists yet.
                                    </p>
                                }
                            >
                                <For each={state.library.savedPlaylists}>
                                    {(playlist) => (
                                        <button
                                            onClick={(e) => { e.stopPropagation(); handleAddToPlaylist(playlist.id); }}
                                            class="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg
                                                    text-left text-sm text-gray-300 hover:bg-white/10
                                                    hover:text-white transition-colors duration-100"
                                        >
                                            <Icon name="list-music" class="w-3.5 h-3.5 text-gray-500 shrink-0" />
                                            <span class="truncate">{playlist.name}</span>
                                        </button>
                                    )}
                                </For>
                            </Show>
                        </div>
                    </div>
                </Show>
            </div>

            {/* Add to Queue */}
            <Tooltip text="Add to Queue">
                <button
                    onClick={handleAddToQueue}
                    aria-label="Add to Queue"
                    class="p-1.5 rounded-lg text-gray-400 hover:text-white hover:bg-white/10
                            transition-colors duration-150 focus:outline-none focus:ring-1
                            focus:ring-white/20"
                >
                    <Icon name="circle-plus" class="w-4 h-4" />
                </button>
            </Tooltip>
        </div>
    );
}
