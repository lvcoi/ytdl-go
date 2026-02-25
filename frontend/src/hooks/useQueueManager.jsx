import { useAppStore } from '../store/appStore';
import { produce } from 'solid-js/store';

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
     */
    const addToSavedPlaylist = (mediaKey, playlistId) => {
        if (!mediaKey || !playlistId) return;

        const playlistExists = state.library.savedPlaylists.some(
            (pl) => pl.id === playlistId
        );
        if (!playlistExists) return;

        setState('library', 'playlistAssignments', mediaKey, playlistId);
    };

    return {
        addToQueue,
        addPlaylistToQueue,
        playPlaylist,
        addToSavedPlaylist,
    };
}
