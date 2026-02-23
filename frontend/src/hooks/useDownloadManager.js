import { onMount, onCleanup, batch } from 'solid-js';
import { useAppStore } from '../store/appStore';
import { downloadStore, setDownloadStore } from '../store/downloadStore';
import wsService from '../services/websocket';
import {
    normalizeDownloadStatus,
} from '../utils/downloadStatus';

const reconnectDelaysMs = [1000, 2000, 4000, 8000, 10000];

const maxReconnectAttempts = 5;
const maxVisibleLogs = 80;
const acceptedLogLevels = new Set(['debug', 'info', 'warn', 'error']);

const statusDefaultMessage = (status) => {
    switch (status) {
        case 'queued': return 'Waiting for a worker to start the job.';
        case 'running': return 'Download in progress...';
        case 'reconnecting': return 'Reconnecting to the progress stream...';
        case 'complete': return 'All downloads in this job are finished.';
        case 'error': return 'Download failed.';
        default: return '';
    }
};

const toNonNegativeInteger = (value, fallback = 0) => {
    const parsed = typeof value === 'number' ? value : Number(value);
    if (!Number.isFinite(parsed)) return fallback;
    return Math.max(0, Math.trunc(parsed));
};

const toFinitePercent = (value) => {
    const parsed = typeof value === 'number' ? value : Number(value);
    if (!Number.isFinite(parsed)) return 0;
    if (parsed < 0) return 0;
    if (parsed > 100) return 100;
    return parsed;
};

const normalizeStats = (value) => {
    if (!value || typeof value !== 'object') return null;
    const stats = {
        total: toNonNegativeInteger(value.total, 0),
        succeeded: toNonNegativeInteger(value.succeeded, 0),
        failed: toNonNegativeInteger(value.failed, 0),
    };
    if (stats.total <= 0 && stats.succeeded <= 0 && stats.failed <= 0) return null;
    return stats;
};

const getStatusMessage = (status, rawMessage, rawError) => {
    const message = typeof rawMessage === 'string' ? rawMessage.trim() : '';
    if (message !== '' && message !== status) return message;
    if (status === 'error') {
        const error = typeof rawError === 'string' ? rawError.trim() : '';
        if (error) return error;
    }
    return statusDefaultMessage(status);
};

const normalizeLogLevel = (value) => {
    if (typeof value !== 'string') return 'info';
    const normalized = value.trim().toLowerCase();
    return acceptedLogLevels.has(normalized) ? normalized : 'info';
};

const normalizeStatus = normalizeDownloadStatus;

const notifyJobOutcome = (job) => {

    const title = job.status === 'complete' ? 'Download Complete' : 'Download Failed';
    const body = job.status === 'complete' 
        ? `Successfully downloaded ${job.stats?.succeeded || 0} items.`
        : job.error || job.message || 'An error occurred during download.';

    // In-app notification
    setDownloadStore('notification', {
        type: job.status === 'complete' ? 'success' : 'error',
        message: `${title}: ${body}`
    });
    
    // Auto-clear notification after 10 seconds
    setTimeout(() => {
        setDownloadStore('notification', null);
    }, 10000);

    if (typeof window === 'undefined' || !('Notification' in window)) return;
    
    if (Notification.permission === 'granted') {
        new Notification(title, {
            body,
            icon: '/favicon.ico'
        });
    }
};


export function useDownloadManager() {

    const { state, setState } = useAppStore();

        const handleWsEvent = (evt) => {
        if (!evt || typeof evt !== 'object' || typeof evt.type !== 'string') return;
        
        // Unwrap payload if it exists (legacy support vs new structure)
        // The new service sends { type, payload }
        // Some parts of this function seem to expect properties directly on evt
        const data = evt.payload || evt;

        batch(() => {
            try {
            switch (evt.type) {
                case 'snapshot':
                    // snapshot is usually in the payload or top level?
                    // Let's assume payload for now based on service structure
                    if (data.snapshot) {
                        applySnapshot(data.snapshot);
                    } else if (evt.snapshot) {
                         applySnapshot(evt.snapshot);
                    }
                    break;
                case 'status':
                    try {
                     // status events might be top level or in payload
                    setDownloadStore('jobStatuses', data.jobId, (prev) => {
                        const nextStatus = normalizeDownloadStatus(data.status);
                        const next = {
                            ...(prev || {}),
                            jobId: data.jobId,
                            status: nextStatus,
                            message: getStatusMessage(data.status, data.message, data.error),
                            error: data.error,
                            stats: normalizeStats(data.stats) || prev?.stats,
                            exitCode: data.exitCode
                        };
                        
                        // Notify on terminal state change
                        if (prev?.status !== nextStatus && (nextStatus === 'complete' || nextStatus === 'error')) {
                            notifyJobOutcome(next);

                            // Clear only tasks belonging to this job 300ms after completion for UI transition
                            const completedJobId = data.jobId;
                            setTimeout(() => {
                                try {
                                    const current = downloadStore.activeDownloads;
                                    const keysToRemove = Object.keys(current).filter(
                                        (key) => current[key]?.jobId === completedJobId || current[key]?.done
                                    );
                                    for (const key of keysToRemove) {
                                        setDownloadStore('activeDownloads', key, undefined);
                                    }
                                } catch (err) {
                                    console.error("Failed to clear completed job tasks:", err);
                                }
                            }, 300);
                        }
                        
                        return next;
                    });
                    } catch (err) {
                        console.error("Critical error processing status payload:", err);
                    }
                    break;

                                case 'register':
                case 'progress':
                    // If it's the new payload format, id might be missing from top level but present in payload
                    // But our wsService already merged them: { ...payload, type }
                    // Actually, wsService dispatches { type, payload }
                    // So we must use data.id
                    if (!data.id) return;

                    setDownloadStore('activeDownloads', data.id, (prev) => ({
                        id: data.id,
                        jobId: data.jobId,
                        // Map filename correctly and persist it
                        filename: data.filename || prev?.filename,
                        label: data.label || data.filename || prev?.label || data.id,
                        
                        total: toNonNegativeInteger(data.total, prev?.total || 0),
                        current: toNonNegativeInteger(data.current, prev?.current || 0),
                        percent: toFinitePercent(data.percent),
                        status: data.status || prev?.status || 'downloading',
                        eta: data.eta || prev?.eta,
                        
                        done: (data.status === 'complete') || (data.percent >= 100)
                    }));
                    break;

                case 'finish':
                    if (data.id) {
                        setDownloadStore('activeDownloads', data.id, 'done', true);
                        setDownloadStore('activeDownloads', data.id, 'percent', 100);
                    }
                    break;
                case 'duplicate':
                    if (typeof data.promptId === 'string' && data.promptId !== '') {
                        setState('download', 'duplicateQueue', (prev) => {
                            if (prev.some((item) => item.promptId === data.promptId)) return prev;
                            return [
                                ...prev,
                                {
                                    jobId: data.jobId,
                                    promptId: data.promptId,
                                    path: data.path || '',
                                    filename: data.filename || '', 
                                },
                            ];
                        });
                        setState('download', 'duplicateError', '');
                        if (import.meta.env?.DEV) console.debug('[ws] Duplicate detected:', data.filename);
                    }
                    break;
                case 'duplicate-resolved':
                    if (typeof data.promptId === 'string' && data.promptId !== '') {
                        setState('download', 'duplicateQueue', (prev) => prev.filter((item) => item.promptId !== data.promptId));
                        setState('download', 'duplicateError', '');
                        if (import.meta.env?.DEV) console.debug('[ws] Duplicate resolved:', data.promptId);
                    }
                    break;

                case 'log':
                    setDownloadStore('logs', (prev) => [
                        ...prev, 
                        { level: normalizeLogLevel(data.level), message: data.message }
                    ].slice(-maxVisibleLogs));
                    break;
                case 'done':
                    setDownloadStore('jobStatuses', data.jobId, 'status', normalizeStatus(data.status));
                    break;
                default:
                    break;
            }
            } catch (err) {
                console.error("Error processing WebSocket event:", err);
            }
        });
    };

        const applySnapshot = (snapshot) => {
        const jobId = snapshot.jobId;
        
        setDownloadStore('jobStatuses', jobId, {
            jobId: jobId,
            status: normalizeStatus(snapshot.status),
            message: getStatusMessage(snapshot.status, '', snapshot.error),
            error: snapshot.error,
            stats: normalizeStats(snapshot.stats),
            exitCode: snapshot.exitCode
        });

        if (Array.isArray(snapshot.tasks)) {
            for (const task of snapshot.tasks) {
                setDownloadStore('activeDownloads', task.id, {
                    id: task.id,
                    jobId: jobId,
                    label: task.label || task.id,
                    total: toNonNegativeInteger(task.total, 0),
                    current: toNonNegativeInteger(task.current, 0),
                    percent: toFinitePercent(task.percent),
                    done: Boolean(task.done) || task.percent >= 100
                });
            }
        }
        
        if (Array.isArray(snapshot.duplicates) && snapshot.duplicates.length > 0) {
            setState('download', 'duplicateQueue', (prev) => {
                const next = [...prev];
                for (const dup of snapshot.duplicates) {
                    if (!next.some(item => item.promptId === dup.promptId)) {
                        next.push({
                            jobId: jobId,
                            promptId: dup.promptId,
                            path: dup.path,
                            filename: dup.filename
                        });
                    }
                }
                return next;
            });
        }
        
        if (Array.isArray(snapshot.logs)) {

            // We could merge logs, but for now just take them
            const logs = snapshot.logs.map(log => ({
                level: normalizeLogLevel(log.level),
                message: log.message
            }));
            setDownloadStore('logs', logs.slice(-maxVisibleLogs));
        }
    };

            onMount(() => {
        const cleanup = wsService.addListener(handleWsEvent);
        
        if (typeof window !== 'undefined' && 'Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }

        onCleanup(cleanup);
    });


        onCleanup(() => {
        // Cleanup happens in the onMount closure
    });


        const startDownload = async (url) => {
        const urls = url.split(',').map(u => u.trim()).filter(u => u.length > 0);
        if (urls.length === 0) return;

        try {
            const response = await fetch('/api/download', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    urls: urls,
                    options: {
                        output: state.settings.output,
                        quality: state.settings.quality,
                        jobs: state.settings.jobs,
                        timeout: state.settings.timeout,
                                                audio: state.settings.audioOnly,
                        "on-duplicate": state.settings.onDuplicate,
                        "use-cookies": state.settings.useCookies,
                        "po-token": state.settings.poTokenExtension ? "true" : "", // Placeholder logic as actual token may come from extension
                    }

                })
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to start download');
            }
            setDownloadStore('lastJobId', data.jobId);

            // The resulting events will come back over the WebSocket

        } catch (error) {
            console.error('Download start failed:', error);
            // Update store with global error
            setDownloadStore('logs', (prev) => [
                ...prev,
                { level: 'error', message: error.message || 'Failed to start download' }
            ].slice(-maxVisibleLogs));
        }
    };

    const cancelDownload = async (jobId) => {
        if (!jobId) return;
        try {
            const response = await fetch('/api/download/cancel', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ jobId })
            });
            if (!response.ok) {
                const data = await response.json().catch(() => ({}));
                throw new Error(data.error || 'Failed to cancel download');
            }
        } catch (error) {
            console.error('Download cancel failed:', error);
        }
    };

    return {
        startDownload,
        cancelDownload,
        // listenForProgress is now implicit via the global WebSocket and store

        listenForProgress: (jobId) => {
            // No-op, kept for compatibility if needed
            console.log('Now listening for job:', jobId);
        }
    };
}
