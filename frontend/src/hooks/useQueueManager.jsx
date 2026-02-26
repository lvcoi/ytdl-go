import { useAppStore } from '../store/appStore';
import { produce } from 'solid-js/store';
import { setDownloadStore } from '../store/downloadStore';

let notificationTimer;
const notify = (message, type = 'success') => {
    clearTimeout(notificationTimer);
    setDownloadStore('notification', { message, type });
    notificationTimer = setTimeout(() => setDownloadStore('notification', null), 3000);
};

export function useQueueManager() {
    const { state, setState } = useAppStore();

    /**
     * Add a single media item to the end of the player queue.
     * Activates the player if not already active.
     * Skips if the item is already enqueued (matched by filepath).
     */
    const addToQueue = (media) => {
        if (!media?.filepath) return;

        setState(produce((s) => {
            const alreadyQueued = s.player.queue.some(
                (item) => item.filepath === media.filepath
            );
            if (alreadyQueued) return;

            s.player.queue.push({ ...media });

            if (!s.player.active) {
                s.player.active = true;
                s.player.minimized = false;
                s.player.selectedMedia = { ...media };
            }
        }));
        notify('Added to Queue');
    };

    /**
     * Append an array of media items to the queue, deduplicating
     * against items already present.
     */
    const addPlaylistToQueue = (mediaItems) => {
        if (!Array.isArray(mediaItems) || mediaItems.length === 0) return;

        setState(produce((s) => {
            const existingPaths = new Set(s.player.queue.map((item) => item.filepath));

            for (const media of mediaItems) {
                if (media?.filepath && !existingPaths.has(media.filepath)) {
                    s.player.queue.push({ ...media });
                    existingPaths.add(media.filepath);
                }
            }

            if (!s.player.active && s.player.queue.length > 0) {
                s.player.active = true;
                s.player.minimized = false;
                s.player.selectedMedia = { ...s.player.queue[0] };
            }
        }));
    };

    /**
     * Replace the current queue with the provided playlist items,
     * immediately start playing the first track.
     */
    const playPlaylist = (mediaItems) => {
        if (!Array.isArray(mediaItems) || mediaItems.length === 0) return;

        const deduplicated = [];
        const seenPaths = new Set();
        for (const media of mediaItems) {
            if (media?.filepath && !seenPaths.has(media.filepath)) {
                deduplicated.push({ ...media });
                seenPaths.add(media.filepath);
            }
        }

        if (deduplicated.length === 0) return;

        setState('player', {
            active: true,
            minimized: false,
            selectedMedia: deduplicated[0],
            queue: deduplicated,
        });
    };

    /**
     * Assign a media item to a saved playlist via the playlistAssignments map.
     * Each media key maps to an array of playlist IDs (multi-playlist support).
     */
    const addToSavedPlaylist = (mediaKey, playlistId) => {
        if (!mediaKey || !playlistId) return;

        const playlistExists = state.library.savedPlaylists.some(
            (pl) => pl.id === playlistId
        );
        if (!playlistExists) return;

        setState('library', 'playlistAssignments', mediaKey, (prev) => {
            const current = Array.isArray(prev) ? prev : (prev ? [prev] : []);
            if (current.includes(playlistId)) return current;
            return [...current, playlistId];
        });
        notify('Added to Playlist');
    };

    /**
     * Remove a media item from a saved playlist.
     */
    const removeFromSavedPlaylist = (mediaKey, playlistId) => {
        if (!mediaKey || !playlistId) return;

        setState('library', 'playlistAssignments', mediaKey, (prev) => {
            const current = Array.isArray(prev) ? prev : (prev ? [prev] : []);
            return current.filter((id) => id !== playlistId);
        });
        notify('Removed from Playlist');
    };

    /**
     * Remove a single item from the queue by index.
     * If the removed item was currently selected, advance to the next track.
     */
    const removeFromQueue = (indexToRemove) => {
        setState(produce((s) => {
            const removed = s.player.queue[indexToRemove];
            s.player.queue.splice(indexToRemove, 1);

            if (s.player.queue.length === 0) {
                s.player.active = false;
                s.player.selectedMedia = null;
            } else if (removed && s.player.selectedMedia?.filepath === removed.filepath) {
                const nextIndex = Math.min(indexToRemove, s.player.queue.length - 1);
                s.player.selectedMedia = { ...s.player.queue[nextIndex] };
            }
        }));
    };

    /**
     * Clear the entire queue and deactivate the player.
     */
    const clearQueue = () => {
        setState('player', {
            queue: [],
            selectedMedia: null,
            active: false,
        });
    };

    /**
     * Move a queue item from one index to another.
     */
    const reorderQueue = (fromIndex, toIndex) => {
        setState(produce((s) => {
            const [removed] = s.player.queue.splice(fromIndex, 1);
            if (removed) {
                s.player.queue.splice(toIndex, 0, removed);
            }
        }));
    };

    return {
        addToQueue,
        addPlaylistToQueue,
        playPlaylist,
        addToSavedPlaylist,
        removeFromSavedPlaylist,
        removeFromQueue,
        clearQueue,
        reorderQueue,
    };
}
