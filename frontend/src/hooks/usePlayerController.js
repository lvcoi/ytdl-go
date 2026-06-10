import { useAppStore } from '../store/appStore';

const encodeMediaPath = (relativePath) => (
    String(relativePath || '')
        .split('/')
        .map((segment) => encodeURIComponent(segment))
        .join('/')
);
const normalizeQueueKey = (value) => String(value || '').trim();

export function usePlayerController() {
    const { state, setState } = useAppStore();

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
        setState('player', 'queue', preparedQueue);
        setState('player', 'selectedMedia', toPlayerMediaItem(item));
        setState('player', 'minimized', false);
        setState('player', 'active', true);
    };

    const closePlayer = () => {
        setState('player', 'active', false);
        setState('player', 'selectedMedia', null);
        setState('player', 'minimized', false);
        setState('player', 'queue', []);
    };

    const playNextInQueue = () => {
        const queue = state.player.queue || [];
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

    return {
        playerQueue: () => state.player.queue,
        openPlayer,
        closePlayer,
        playNextInQueue,
    };
}
