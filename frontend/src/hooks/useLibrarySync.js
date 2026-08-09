import { createEffect, onMount, onCleanup } from 'solid-js';
import { useAppStore } from '../store/appStore';
import { downloadStore } from '../store/downloadStore';
import { normalizeDownloadStatus } from '../utils/downloadStatus';

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

export function useLibrarySync() {
    const { setState } = useAppStore();

    let mediaListAbortController = null;
    let mediaListRequestToken = 0;
    let lastSyncedLibraryJobId = '';
    let pendingLibrarySyncJobId = '';
    let librarySyncRetryTimer = null;
    let isDisposed = false;

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

    const clearLibrarySyncRetryTimer = () => {
        if (librarySyncRetryTimer) {
            clearTimeout(librarySyncRetryTimer);
            librarySyncRetryTimer = null;
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
        if (mediaListAbortController) {
            mediaListAbortController.abort();
            mediaListAbortController = null;
        }
    });

    return { fetchMediaFiles, retryLibraryMetadataScan: async () => {
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
        }
    };
}
