import { createEffect, onCleanup } from 'solid-js';
import { useAppStore } from '../store/appStore';
import {
    isActiveDownloadStreamStatus,
    isTerminalDownloadStatus,
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

const toBoundedPositiveInteger = (value, fallback, max) => {
    const parsed = typeof value === 'number' ? value : Number(value);
    if (!Number.isFinite(parsed)) return fallback;
    const normalized = Math.trunc(parsed);
    if (normalized <= 0) return fallback;
    return Math.min(normalized, max);
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
const isActiveStreamStatus = isActiveDownloadStreamStatus;
const isTerminalStatus = isTerminalDownloadStatus;

const resolveExitCode = (status, explicitExitCode, previousExitCode) => {
    if (explicitExitCode != null) return explicitExitCode;
    if (status === 'complete') return 0;
    if (status === 'error') return previousExitCode ?? null;
    return null;
};

const reportSseClientError = (message, error) => {
    if (!import.meta.env?.DEV) return;
    if (error) {
        console.warn(`[download-sse] ${message}`, error);
        return;
    }
    console.warn(`[download-sse] ${message}`);
};

export function useDownloadManager() {
    const { state, setState } = useAppStore();

    let eventSource = null;
    let reconnectTimer = null;
    let reconnectAttempts = 0;
    let activeJobId = '';
    let lastEventSeq = 0;

    const jobStatus = () => state.download.jobStatus;
    const duplicateQueue = () => state.download.duplicateQueue;

    const setJobStatus = (nextValue) => {
        setState('download', 'jobStatus', (prev) => (
            typeof nextValue === 'function' ? nextValue(prev) : nextValue
        ));
    };

    const setProgressTasks = (nextValue) => {
        setState('download', 'progressTasks', (prev) => (
            typeof nextValue === 'function' ? nextValue(prev) : nextValue
        ));
    };

    const setLogMessages = (nextValue) => {
        setState('download', 'logMessages', (prev) => (
            typeof nextValue === 'function' ? nextValue(prev) : nextValue
        ));
    };

    const setDuplicateQueue = (nextValue) => {
        setState('download', 'duplicateQueue', (prev) => (
            typeof nextValue === 'function' ? nextValue(prev) : nextValue
        ));
    };

    const setDuplicateError = (nextValue) => {
        setState('download', 'duplicateError', nextValue);
    };

    const setIsDownloading = (nextValue) => {
        setState('download', 'isDownloading', nextValue);
    };

    const closeProgressStream = () => {
        if (eventSource) {
            eventSource.close();
            eventSource = null;
        }
    };

    const clearReconnectTimer = () => {
        if (reconnectTimer) {
            clearTimeout(reconnectTimer);
            reconnectTimer = null;
        }
    };

    const resetProgressStreamState = () => {
        clearReconnectTimer();
        closeProgressStream();
        reconnectAttempts = 0;
        activeJobId = '';
        lastEventSeq = 0;
    };

    const markStreamConnected = () => {
        reconnectAttempts = 0;
        clearReconnectTimer();
        setJobStatus((prev) => {
            if (!prev || prev.status !== 'reconnecting') return prev;
            return {
                ...prev,
                status: 'running',
                message: statusDefaultMessage('running'),
                error: '',
            };
        });
    };

    const applySnapshot = (snapshot, expectedJobId) => {
        if (!snapshot || typeof snapshot !== 'object') return;

        const snapshotJobId = typeof snapshot.jobId === 'string' && snapshot.jobId
            ? snapshot.jobId
            : expectedJobId;

        if (snapshotJobId !== expectedJobId) return;

        const tasks = {};
        if (Array.isArray(snapshot.tasks)) {
            for (const task of snapshot.tasks) {
                if (!task || typeof task !== 'object' || typeof task.id !== 'string' || task.id === '') continue;
                const total = toNonNegativeInteger(task.total, 0);
                const current = toNonNegativeInteger(task.current, 0);
                const percent = toFinitePercent(task.percent);
                const done = Boolean(task.done) || percent >= 100;
                tasks[task.id] = {
                    label: task.label || task.id,
                    total,
                    current: total > 0 && current > total ? total : current,
                    percent: done ? 100 : percent,
                    done,
                };
            }
        }
        setProgressTasks(tasks);

        const logs = [];
        if (Array.isArray(snapshot.logs)) {
            for (const log of snapshot.logs.slice(-maxVisibleLogs)) {
                if (!log || typeof log !== 'object') continue;
                if (typeof log.message !== 'string' || log.message === '') continue;
                logs.push({
                    level: normalizeLogLevel(log.level),
                    message: log.message,
                });
            }
        }
        setLogMessages(logs);

        const duplicates = [];
        if (Array.isArray(snapshot.duplicates)) {
            for (const duplicate of snapshot.duplicates) {
                if (!duplicate || typeof duplicate !== 'object') continue;
                if (typeof duplicate.promptId !== 'string' || duplicate.promptId === '') continue;
                duplicates.push({
                    jobId: snapshotJobId,
                    promptId: duplicate.promptId,
                    path: typeof duplicate.path === 'string' ? duplicate.path : '',
                    filename: typeof duplicate.filename === 'string' ? duplicate.filename : '',
                });
            }
        }
        setDuplicateQueue(duplicates);
        setDuplicateError('');

        const snapshotStatus = normalizeStatus(snapshot.status);
        const snapshotStats = normalizeStats(snapshot.stats);
        const snapshotError = typeof snapshot.error === 'string' ? snapshot.error : '';
        const snapshotExitCode = Number.isFinite(Number(snapshot.exitCode))
            ? Math.trunc(Number(snapshot.exitCode))
            : null;

        const snapshotLastSeq = Number(snapshot.lastSeq);
        if (Number.isFinite(snapshotLastSeq) && snapshotLastSeq > lastEventSeq) {
            lastEventSeq = Math.trunc(snapshotLastSeq);
        }

        setJobStatus((prev) => {
            const nextStatus = snapshotStatus || normalizeStatus(prev?.status) || 'running';
            return {
                ...(prev || {}),
                jobId: snapshotJobId,
                status: nextStatus,
                message: getStatusMessage(nextStatus, '', snapshotError),
                error: nextStatus === 'error' ? (snapshotError || prev?.error || statusDefaultMessage('error')) : '',
                exitCode: resolveExitCode(nextStatus, snapshotExitCode, prev?.exitCode),
                stats: snapshotStats || prev?.stats || null,
            };
        });

        if (snapshotStatus && isTerminalStatus(snapshotStatus)) {
            setIsDownloading(false);
            return;
        }
        setIsDownloading(true);
    };

    const handleStatusEvent = (evt, expectedJobId) => {
        const status = normalizeStatus(evt?.status || evt?.message);
        if (!status) return;

        const eventError = typeof evt.error === 'string' ? evt.error : '';
        const eventStats = normalizeStats(evt.stats);
        const eventExitCode = Number.isFinite(Number(evt.exitCode))
            ? Math.trunc(Number(evt.exitCode))
            : null;

        setJobStatus((prev) => ({
            ...(prev || {}),
            jobId: expectedJobId,
            status,
            message: getStatusMessage(status, evt.message, eventError),
            error: status === 'error' ? (eventError || prev?.error || statusDefaultMessage('error')) : '',
            exitCode: resolveExitCode(status, eventExitCode, prev?.exitCode),
            stats: eventStats || prev?.stats || null,
        }));

        if (isTerminalStatus(status)) {
            setIsDownloading(false);
            return;
        }
        setIsDownloading(true);
    };

    const handleDoneEvent = (evt, expectedJobId) => {
        const status = normalizeStatus(evt?.status || evt?.message) || 'complete';
        const eventError = typeof evt.error === 'string' ? evt.error : '';
        const eventStats = normalizeStats(evt.stats);
        const eventExitCode = Number.isFinite(Number(evt.exitCode))
            ? Math.trunc(Number(evt.exitCode))
            : null;

        setJobStatus((prev) => ({
            ...(prev || {}),
            jobId: expectedJobId,
            status,
            message: getStatusMessage(status, evt.message, eventError),
            error: status === 'error' ? (eventError || prev?.error || statusDefaultMessage('error')) : '',
            exitCode: resolveExitCode(status, eventExitCode, prev?.exitCode),
            stats: eventStats || prev?.stats || null,
        }));

        setIsDownloading(false);
        setDuplicateQueue([]);
        setDuplicateError('');
        
        // Stop the stream but don't clear the task/log state yet so user can see final results
        closeProgressStream();
        clearReconnectTimer();
        activeJobId = '';
    };

    const listenForProgress = (jobId) => {
        resetProgressStreamState();
        activeJobId = jobId;

        const connect = () => {
            if (activeJobId !== jobId) return;

            clearReconnectTimer();
            closeProgressStream();

            const query = new URLSearchParams({ id: jobId });
            if (lastEventSeq > 0) {
                query.set('since', String(lastEventSeq));
            }

            eventSource = new EventSource(`/api/download/progress?${query.toString()}`);

            eventSource.onmessage = (e) => {
                if (activeJobId !== jobId) return;
                try {
                    const evt = JSON.parse(e.data);
                    if (!evt || typeof evt !== 'object' || typeof evt.type !== 'string') return;

                    const eventJobId = typeof evt.jobId === 'string' && evt.jobId ? evt.jobId : jobId;
                    if (eventJobId !== jobId) return;

                    if (evt.type !== 'snapshot') {
                        const seq = Number(evt.seq);
                        if (Number.isFinite(seq) && seq > 0) {
                            const normalizedSeq = Math.trunc(seq);
                            if (normalizedSeq <= lastEventSeq) return;
                            lastEventSeq = normalizedSeq;
                        }
                    }

                    markStreamConnected();

                    switch (evt.type) {
                        case 'snapshot':
                            applySnapshot(evt.snapshot, jobId);
                            break;
                        case 'status':
                            handleStatusEvent(evt, jobId);
                            break;
                        case 'register':
                            if (typeof evt.id === 'string' && evt.id !== '') {
                                setProgressTasks((prev) => ({
                                    ...prev,
                                    [evt.id]: {
                                        ...(prev[evt.id] || {}),
                                        label: evt.label || prev[evt.id]?.label || evt.id,
                                        total: toNonNegativeInteger(evt.total, prev[evt.id]?.total || 0),
                                        current: toNonNegativeInteger(evt.current, 0),
                                        percent: toFinitePercent(evt.percent),
                                        done: false,
                                    },
                                }));
                            }
                            break;
                        case 'progress':
                            if (typeof evt.id === 'string' && evt.id !== '') {
                                setProgressTasks((prev) => {
                                    const existing = prev[evt.id] || {};
                                    const total = toNonNegativeInteger(evt.total, existing.total || 0);
                                    let current = toNonNegativeInteger(evt.current, existing.current || 0);
                                    if (total > 0 && current > total) current = total;
                                    const percent = toFinitePercent(
                                        typeof evt.percent === 'number' && Number.isFinite(evt.percent)
                                            ? evt.percent
                                            : (total > 0 ? (current * 100) / total : existing.percent || 0),
                                    );
                                    return {
                                        ...prev,
                                        [evt.id]: {
                                            ...existing,
                                            label: evt.label || existing.label || evt.id,
                                            total,
                                            current,
                                            percent,
                                            done: percent >= 100,
                                        },
                                    };
                                });
                            }
                            break;
                        case 'finish':
                            if (typeof evt.id === 'string' && evt.id !== '') {
                                setProgressTasks((prev) => ({
                                    ...prev,
                                    [evt.id]: {
                                        ...(prev[evt.id] || {}),
                                        label: prev[evt.id]?.label || evt.id,
                                        current: Math.max(
                                            toNonNegativeInteger(prev[evt.id]?.current, 0),
                                            toNonNegativeInteger(prev[evt.id]?.total, 0),
                                        ),
                                        total: toNonNegativeInteger(prev[evt.id]?.total, 0),
                                        percent: 100,
                                        done: true,
                                    },
                                }));
                            }
                            break;
                        case 'log':
                            if (typeof evt.message === 'string' && evt.message !== '') {
                                setLogMessages((prev) => [
                                    ...prev,
                                    {
                                        level: normalizeLogLevel(evt.level),
                                        message: evt.message,
                                    },
                                ].slice(-maxVisibleLogs));
                            }
                            break;
                        case 'duplicate':
                            if (typeof evt.promptId === 'string' && evt.promptId !== '') {
                                setDuplicateQueue((prev) => {
                                    if (prev.some((item) => item.promptId === evt.promptId)) return prev;
                                    return [
                                        ...prev,
                                        {
                                            jobId,
                                            promptId: evt.promptId,
                                            path: typeof evt.path === 'string' ? evt.path : '',
                                            filename: typeof evt.filename === 'string' ? evt.filename : '', // Fix: use evt.filename
                                        },
                                    ];
                                });
                                setDuplicateError('');
                            }
                            break;
                        case 'duplicate-resolved':
                            if (typeof evt.promptId === 'string' && evt.promptId !== '') {
                                setDuplicateQueue((prev) => prev.filter((item) => item.promptId !== evt.promptId));
                                setDuplicateError('');
                            }
                            break;
                        case 'done':
                            handleDoneEvent(evt, jobId);
                            break;
                        default:
                            break;
                    }
                } catch (error) {
                    reportSseClientError('Failed to parse/process SSE payload', error);
                }
            };

            eventSource.onerror = () => {
                if (activeJobId !== jobId) return;
                closeProgressStream();
                clearReconnectTimer();

                const status = normalizeStatus(jobStatus()?.status);
                if (!isActiveStreamStatus(status)) return;

                if (reconnectAttempts >= maxReconnectAttempts) {
                    resetProgressStreamState();
                    setJobStatus((prev) => ({
                        ...(prev || { jobId }),
                        jobId,
                        status: 'error',
                        message: 'Progress stream disconnected.',
                        error: 'Connection lost. Progress updates stopped before completion.',
                    }));
                    setIsDownloading(false);
                    setDuplicateQueue([]);
                    setDuplicateError('');
                    return;
                }

                reconnectAttempts += 1;
                const delay = reconnectDelaysMs[Math.min(reconnectAttempts - 1, reconnectDelaysMs.length - 1)];
                setJobStatus((prev) => ({
                    ...(prev || { jobId }),
                    jobId,
                    status: 'reconnecting',
                    message: `Reconnecting... (${reconnectAttempts}/${maxReconnectAttempts})`,
                    error: '',
                }));

                reconnectTimer = setTimeout(() => {
                    reconnectTimer = null;
                    if (activeJobId === jobId) {
                        connect();
                    }
                }, delay);
            };
        };

        connect();
    };

    onCleanup(() => {
        resetProgressStreamState();
    });

    return {
        listenForProgress,
    };
}
