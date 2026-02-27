import { createSignal, Show, For, onCleanup } from 'solid-js';
import Icon from './Icon';
import Tooltip from './Tooltip';
import { useAppStore } from '../store/appStore';
import { useQueueManager } from '../hooks/useQueueManager';

export default function SongActions(props) {
    const { state } = useAppStore();
    const { addToQueue, addToSavedPlaylist, removeFromSavedPlaylist } = useQueueManager();
    const [showPlaylistDropdown, setShowPlaylistDropdown] = createSignal(false);
    let dropdownRef;

    const mediaKey = () => props.media?.filepath || props.media?.id || '';

    const handleClickOutside = (e) => {
        if (dropdownRef && !dropdownRef.contains(e.target)) {
            closeDropdown();
        }
    };

    const openDropdown = (e) => {
        e.stopPropagation();
        e.preventDefault();
        setShowPlaylistDropdown(true);
        document.addEventListener('pointerdown', handleClickOutside);
    };

    const closeDropdown = () => {
        setShowPlaylistDropdown(false);
        document.removeEventListener('pointerdown', handleClickOutside);
    };

    onCleanup(() => {
        document.removeEventListener('pointerdown', handleClickOutside);
    });

    const togglePlaylist = (playlistId) => {
        const key = mediaKey();
        if (!key) return;
        const assignments = state.library.playlistAssignments[key] || [];
        if (assignments.includes(playlistId)) {
            removeFromSavedPlaylist(key, playlistId);
        } else {
            addToSavedPlaylist(key, playlistId);
        }
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
                        onClick={openDropdown}
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
                            <h4 class="px-3 py-1.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">Add to Playlist</h4>
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
                                        <label
                                            class="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg
                                                    cursor-pointer text-sm text-gray-300 hover:bg-white/10
                                                    hover:text-white transition-colors duration-100"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <input
                                                type="checkbox"
                                                class="accent-blue-500 shrink-0"
                                                checked={(state.library.playlistAssignments[mediaKey()] || []).includes(playlist.id)}
                                                onChange={() => togglePlaylist(playlist.id)}
                                            />
                                            <Icon name="list-music" class="w-3.5 h-3.5 text-gray-500 shrink-0" />
                                            <span class="truncate">{playlist.name}</span>
                                        </label>
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
