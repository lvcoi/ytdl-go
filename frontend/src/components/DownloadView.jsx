import { For, Show, onCleanup, onMount } from 'solid-js';
import Icon from './Icon';
import DuplicateModal from './DuplicateModal';
import { MAX_JOBS, MAX_TIMEOUT_SECONDS, useAppStore } from '../store/appStore';

const reconnectDelaysMs = [1000, 2000, 4000, 8000, 10000];
const maxReconnectAttempts = 5;
const maxVisibleLogs = 80;

const activeStreamStatuses = new Set(['queued', 'running', 'reconnecting']);
const terminalStatuses = new Set(['complete', 'error']);
const acceptedStatuses = new Set(['queued', 'running', 'reconnecting', 'complete', 'error']);
const acceptedLogLevels = new Set(['debug', 'info', 'warn', 'error']);

const normalizeStatus = (value) => {
  if (typeof value !== 'string') return '';
  const normalized = value.trim().toLowerCase();
  return acceptedStatuses.has(normalized) ? normalized : '';
};

const isActiveStreamStatus = (status) => activeStreamStatuses.has(status);
const isTerminalStatus = (status) => terminalStatuses.has(status);

const statusTitle = (status) => {
  switch (status) {
    case 'queued':
      return 'Queued';
    case 'running':
      return 'Downloading...';
    case 'reconnecting':
      return 'Reconnecting...';
    case 'complete':
      return 'Download Complete';
    case 'error':
      return 'Download Failed';
    default:
      return 'Download Status';
  }
};

const statusDefaultMessage = (status) => {
  switch (status) {
    case 'queued':
      return 'Waiting for a worker to start the job.';
    case 'running':
      return 'Download in progress...';
    case 'reconnecting':
      return 'Reconnecting to the progress stream...';
    case 'complete':
      return 'All downloads in this job are finished.';
    case 'error':
      return 'Download failed.';
    default:
      return '';
  }
};

const statusIconName = (status) => {
  if (status === 'complete') return 'check-circle-2';
  if (status === 'error') return 'alert-circle';
  return 'loader';
};

const statusTone = (status) => {
  if (status === 'error') {
    return {
      card: 'bg-red-500/5 border-red-500/20',
      icon: 'bg-red-500/10 text-red-400',
      accent: 'text-red-400',
    };
  }
  if (status === 'complete') {
    return {
      card: 'bg-green-500/5 border-green-500/20',
      icon: 'bg-green-500/10 text-green-400',
      accent: 'text-green-400',
    };
  }
  if (status === 'reconnecting') {
    return {
      card: 'bg-amber-500/5 border-amber-500/20',
      icon: 'bg-amber-500/10 text-amber-400',
      accent: 'text-amber-400',
    };
  }
  return {
    card: 'bg-blue-500/5 border-blue-500/20',
    icon: 'bg-blue-500/10 text-blue-400',
    accent: 'text-blue-400',
  };
};

const toBoundedPositiveInteger = (value, fallback, max) => {
  const parsed = typeof value === 'number' ? value : Number(value);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  const normalized = Math.trunc(parsed);
  if (normalized <= 0) {
    return fallback;
  }
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
  if (stats.total <= 0 && stats.succeeded <= 0 && stats.failed <= 0) {
    return null;
  }
  return stats;
};

const humanBytes = (bytes) => {
  const normalized = toNonNegativeInteger(bytes, 0);
  if (normalized <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.min(Math.floor(Math.log(normalized) / Math.log(1024)), units.length - 1);
  const value = normalized / (1024 ** index);
  const precision = index === 0 ? 0 : 1;
  return `${value.toFixed(precision)} ${units[index]}`;
};

const formatPercent = (value, isDone) => {
  if (isDone) return '100%';
  return `${toFinitePercent(value).toFixed(1)}%`;
};

const getStatusMessage = (status, rawMessage, rawError) => {
  const message = typeof rawMessage === 'string' ? rawMessage.trim() : '';
  if (message !== '' && message !== status) {
    return message;
  }
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

const reportSseClientError = (message, error) => {
  if (!import.meta.env?.DEV) return;
  if (error) {
    console.warn(`[download-sse] ${message}`, error);
    return;
  }
  console.warn(`[download-sse] ${message}`);
};

const resolveExitCode = (status, explicitExitCode, previousExitCode) => {
  if (explicitExitCode != null) return explicitExitCode;
  if (status === 'complete') return 0;
  if (status === 'error') return previousExitCode ?? null;
  return null;
};

export default function DownloadView() {
  const { state, setState } = useAppStore();

  const urlInput = () => state.download.urlInput;
  const setUrlInput = (nextValue) => {
    setState('download', 'urlInput', nextValue);
  };

  const isDownloading = () => state.download.isDownloading;
  const setIsDownloading = (nextValue) => {
    setState('download', 'isDownloading', nextValue);
  };

  const jobStatus = () => state.download.jobStatus;
  const setJobStatus = (nextValue) => {
    setState('download', 'jobStatus', (prev) => (
      typeof nextValue === 'function' ? nextValue(prev) : nextValue
    ));
  };

  const progressTasks = () => state.download.progressTasks;
  const setProgressTasks = (nextValue) => {
    setState('download', 'progressTasks', (prev) => (
      typeof nextValue === 'function' ? nextValue(prev) : nextValue
    ));
  };

  const logMessages = () => state.download.logMessages;
  const setLogMessages = (nextValue) => {
    setState('download', 'logMessages', (prev) => (
      typeof nextValue === 'function' ? nextValue(prev) : nextValue
    ));
  };

  const duplicateQueue = () => state.download.duplicateQueue;
  const setDuplicateQueue = (nextValue) => {
    setState('download', 'duplicateQueue', (prev) => (
      typeof nextValue === 'function' ? nextValue(prev) : nextValue
    ));
  };

  const duplicateError = () => state.download.duplicateError;
  const setDuplicateError = (nextValue) => {
    setState('download', 'duplicateError', nextValue);
  };

  const settings = () => state.settings;
  const isAdvanced = () => state.ui.isAdvanced;
  const setSettings = (nextSettings) => {
    if (typeof nextSettings === 'function') {
      setState('settings', (prev) => {
        const resolved = nextSettings(prev);
        if (!resolved || typeof resolved !== 'object') {
          return prev;
        }
        return {
          ...prev,
          ...resolved,
        };
      });
      return;
    }

    if (!nextSettings || typeof nextSettings !== 'object') {
      return;
    }

    setState('settings', (prev) => ({
      ...prev,
      ...nextSettings,
    }));
  };

  let eventSource = null;
  let reconnectTimer = null;
  let reconnectAttempts = 0;
  let activeJobId = '';
  let lastEventSeq = 0;

  const activeDuplicate = () => duplicateQueue()[0];

  const sortedTaskEntries = () => (
    Object.entries(progressTasks()).sort(([, a], [, b]) => {
      const aLabel = (a?.label || '').toLowerCase();
      const bLabel = (b?.label || '').toLowerCase();
      if (aLabel === bLabel) return 0;
      return aLabel < bLabel ? -1 : 1;
    })
  );

  const taskSummary = () => {
    const entries = sortedTaskEntries();
    let done = 0;
    for (const [, task] of entries) {
      if (task?.done || toFinitePercent(task?.percent) >= 100) {
        done += 1;
      }
    }
    return { total: entries.length, done };
  };

  const currentStats = () => normalizeStats(jobStatus()?.stats);
  const currentStatusTone = () => statusTone(normalizeStatus(jobStatus()?.status));

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

  const parseInputUrls = (rawInput) => {
    const lines = rawInput
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean);

    const validUrls = [];
    const invalidUrls = [];

    for (const value of lines) {
      try {
        const parsed = new URL(value);
        if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
          validUrls.push(value);
          continue;
        }
      } catch (_) {}
      invalidUrls.push(value);
    }

    return { validUrls, invalidUrls };
  };

  const removeDuplicatePrompt = (promptId) => {
    if (!promptId) return;
    setDuplicateQueue((prev) => prev.filter((item) => item.promptId !== promptId));
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

    if (snapshotJobId !== expectedJobId) {
      return;
    }

    const tasks = {};
    if (Array.isArray(snapshot.tasks)) {
      for (const task of snapshot.tasks) {
        if (!task || typeof task !== 'object' || typeof task.id !== 'string' || task.id === '') {
          continue;
        }
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
      setDuplicateQueue([]);
      setDuplicateError('');
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
    resetProgressStreamState();
  };

  const submitDuplicateChoice = async (choice) => {
    const prompt = activeDuplicate();
    if (!prompt) return;
    setDuplicateError('');
    try {
      const res = await fetch('/api/download/duplicate-response', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jobId: prompt.jobId,
          promptId: prompt.promptId,
          choice,
        }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || data.error) {
        if (res.status === 404 || res.status === 409) {
          removeDuplicatePrompt(prompt.promptId);
          setDuplicateError('');
          return;
        }
        setDuplicateError(data.error || 'Failed to submit duplicate choice');
        return;
      }
      removeDuplicatePrompt(prompt.promptId);
      setDuplicateError('');
    } catch (e) {
      setDuplicateError(e.message || 'Failed to submit duplicate choice');
    }
  };

  const handleDuplicateShortcut = (e) => {
    if (!activeDuplicate()) return;
    if (e.metaKey || e.ctrlKey || e.altKey) return;

    switch (e.key) {
      case 'o':
        e.preventDefault();
        submitDuplicateChoice('overwrite');
        break;
      case 'O':
        e.preventDefault();
        submitDuplicateChoice('overwrite_all');
        break;
      case 's':
        e.preventDefault();
        submitDuplicateChoice('skip');
        break;
      case 'S':
        e.preventDefault();
        submitDuplicateChoice('skip_all');
        break;
      case 'r':
        e.preventDefault();
        submitDuplicateChoice('rename');
        break;
      case 'R':
        e.preventDefault();
        submitDuplicateChoice('rename_all');
        break;
      case 'q':
        e.preventDefault();
        submitDuplicateChoice('cancel');
        break;
      default:
        break;
    }
  };

  onMount(() => {
    if (typeof window !== 'undefined') {
      window.addEventListener('keydown', handleDuplicateShortcut);
    }
  });

  onCleanup(() => {
    resetProgressStreamState();
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleDuplicateShortcut);
    }
  });

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
          if (!evt || typeof evt !== 'object' || typeof evt.type !== 'string') {
            return;
          }

          const eventJobId = typeof evt.jobId === 'string' && evt.jobId
            ? evt.jobId
            : jobId;
          if (eventJobId !== jobId) {
            return;
          }

          if (evt.type !== 'snapshot') {
            const seq = Number(evt.seq);
            if (Number.isFinite(seq) && seq > 0) {
              const normalizedSeq = Math.trunc(seq);
              if (normalizedSeq <= lastEventSeq) {
                return;
              }
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
                  if (total > 0 && current > total) {
                    current = total;
                  }
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
                  if (prev.some((item) => item.promptId === evt.promptId)) {
                    return prev;
                  }
                  return [
                    ...prev,
                    {
                      jobId,
                      promptId: evt.promptId,
                      path: typeof evt.path === 'string' ? evt.path : '',
                      filename: typeof evt.filename === 'string' ? evt.filename : '',
                    },
                  ];
                });
                setDuplicateError('');
              }
              break;
            case 'duplicate-resolved':
              if (typeof evt.promptId === 'string' && evt.promptId !== '') {
                removeDuplicatePrompt(evt.promptId);
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
        if (!isActiveStreamStatus(status)) {
          return;
        }

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

  const handleDownload = async () => {
    if (!urlInput().trim()) return;

    resetProgressStreamState();
    setIsDownloading(true);
    setJobStatus(null);
    setProgressTasks({});
    setLogMessages([]);
    setDuplicateQueue([]);
    setDuplicateError('');

    const { validUrls: urls, invalidUrls } = parseInputUrls(urlInput());
    if (invalidUrls.length > 0) {
      const preview = invalidUrls.slice(0, 3).join(', ');
      const suffix = invalidUrls.length > 3 ? ', ...' : '';
      const label = invalidUrls.length === 1 ? 'Invalid URL' : `Invalid URLs (${invalidUrls.length})`;
      setJobStatus({ status: 'error', message: statusDefaultMessage('error'), error: `${label}: ${preview}${suffix}` });
      setIsDownloading(false);
      return;
    }
    if (urls.length === 0) {
      setJobStatus({ status: 'error', message: statusDefaultMessage('error'), error: 'No valid URLs provided.' });
      setIsDownloading(false);
      return;
    }

    const s = settings();
    const payload = {
      urls,
      options: {
        output: s.output,
        audio: s.audioOnly,
        quality: s.quality,
        format: s.format,
        jobs: toBoundedPositiveInteger(s.jobs, 1, MAX_JOBS),
        timeout: toBoundedPositiveInteger(s.timeout, 180, MAX_TIMEOUT_SECONDS),
        'on-duplicate': s.onDuplicate || 'prompt',
      },
    };

    try {
      const res = await fetch('/api/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || data.error || typeof data.jobId !== 'string' || data.jobId === '') {
        setJobStatus({
          status: 'error',
          message: statusDefaultMessage('error'),
          error: data.error || 'Failed to start download job.',
        });
        setIsDownloading(false);
        return;
      }

      setJobStatus({
        status: 'queued',
        jobId: data.jobId,
        message: getStatusMessage('queued', data.message, ''),
        error: '',
        exitCode: null,
        stats: null,
      });
      listenForProgress(data.jobId);
    } catch (e) {
      setJobStatus({
        status: 'error',
        message: statusDefaultMessage('error'),
        error: e.message || 'Failed to start download job.',
      });
      setIsDownloading(false);
    }
  };

  return (
    <div class="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <div class="space-y-2">
        <h1 class="text-4xl font-black text-white">Unlock Content.</h1>
        <p class="text-gray-500 font-medium">Paste your YouTube URLs below to begin high-speed extraction.</p>
      </div>

      <div class="relative group">
        <textarea
          value={urlInput()}
          onInput={(e) => setUrlInput(e.target.value)}
          class="w-full h-64 bg-[#0a0c14] border-2 border-white/5 rounded-[2rem] p-8 outline-none focus:border-blue-500/50 focus:ring-4 focus:ring-blue-500/10 transition-all text-xl font-medium placeholder:text-gray-800 custom-scrollbar shadow-2xl"
          placeholder="Enter URLs (one per line)..."
        ></textarea>
        <div class="absolute bottom-6 right-6 flex gap-3">
          <button onClick={() => setUrlInput('')} class="p-4 bg-white/5 rounded-2xl hover:bg-red-500/10 hover:text-red-400 transition-all">
            <Icon name="trash-2" class="w-6 h-6" />
          </button>
          <button
            disabled={isDownloading()}
            onClick={handleDownload}
            class="px-8 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl font-bold flex items-center gap-3 transition-all shadow-xl shadow-blue-600/30 disabled:opacity-50"
          >
            <Icon name="download-cloud" class="w-6 h-6" />
            {isDownloading() ? 'Processing...' : 'Start Extraction'}
          </button>
        </div>
      </div>

      <div class="grid grid-cols-3 gap-4">
        <button
          onClick={() => setSettings({ ...settings(), audioOnly: !settings().audioOnly })}
          class={`p-6 rounded-3xl border-2 transition-all flex flex-col gap-3 ${settings().audioOnly ? 'bg-blue-600/10 border-blue-500/50' : 'bg-white/5 border-transparent hover:border-white/10'}`}
        >
          <div class="p-3 bg-purple-500/10 text-purple-400 rounded-xl w-fit"><Icon name="music" class="w-5 h-5" /></div>
          <div class="text-left">
            <div class="font-bold text-white">Audio Only</div>
            <div class="text-xs text-gray-500">Extract high-quality MP3/Opus</div>
          </div>
        </button>
        <div class="p-6 rounded-3xl bg-white/5 border-2 border-transparent flex flex-col gap-3 group has-tooltip relative cursor-help">
          <span class="tooltip bg-gray-800 text-[10px] px-3 py-2 rounded-xl shadow-2xl mb-4 border border-white/10 w-56 text-center leading-relaxed">YouTube prevents downloads without valid Proof-of-Origin (PO) tokens. Automated bypass is active.</span>
          <div class="p-3 bg-amber-500/10 text-amber-400 rounded-xl w-fit"><Icon name="shield-check" class="w-5 h-5" /></div>
          <div class="text-left">
            <div class="font-bold text-white">PO Token Guard</div>
            <div class="text-xs text-gray-500">Automated Bot Detection Bypass</div>
          </div>
        </div>
        <div class="p-6 rounded-3xl bg-white/5 border-2 border-transparent flex flex-col gap-3">
          <div class="p-3 bg-green-500/10 text-green-400 rounded-xl w-fit"><Icon name="database" class="w-5 h-5" /></div>
          <div class="text-left">
            <div class="font-bold text-white">Smart Meta</div>
            <div class="text-xs text-gray-500">Auto-tagging & Organization</div>
          </div>
        </div>
      </div>

      <Show when={jobStatus()}>
        <div class={`p-6 rounded-3xl border-2 space-y-4 animate-in fade-in duration-300 ${currentStatusTone().card}`}>
          <div class="flex items-start gap-3">
            <div class={`p-2 rounded-xl ${currentStatusTone().icon}`}>
              <Icon
                name={statusIconName(normalizeStatus(jobStatus()?.status))}
                class={`w-5 h-5 ${!isTerminalStatus(normalizeStatus(jobStatus()?.status)) ? 'animate-spin' : ''}`}
              />
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-bold text-white">{statusTitle(normalizeStatus(jobStatus()?.status))}</div>
              <Show when={jobStatus()?.message}>
                <div class="text-xs text-gray-500 break-words">{jobStatus().message}</div>
              </Show>
              <Show when={jobStatus()?.status === 'error' && jobStatus()?.error}>
                <div class={`text-xs break-words ${currentStatusTone().accent}`}>{jobStatus().error}</div>
              </Show>
              <div class="mt-2 flex flex-wrap gap-3 text-[11px] text-gray-500">
                <Show when={jobStatus()?.jobId}>
                  <span>Job: <span class="text-gray-400 font-mono">{jobStatus().jobId}</span></span>
                </Show>
                <Show when={taskSummary().total > 0}>
                  <span>Tasks: <span class="text-gray-400">{taskSummary().done}/{taskSummary().total}</span></span>
                </Show>
                <Show when={isTerminalStatus(normalizeStatus(jobStatus()?.status)) && jobStatus()?.exitCode != null}>
                  <span>Exit Code: <span class="text-gray-400">{jobStatus().exitCode}</span></span>
                </Show>
              </div>
            </div>
          </div>

          <Show when={currentStats()}>
            <div class="grid grid-cols-3 gap-2">
              <div class="rounded-xl border border-white/10 bg-white/5 p-3">
                <div class="text-[10px] uppercase tracking-wider text-gray-500">Total</div>
                <div class="text-sm font-bold text-white">{currentStats().total}</div>
              </div>
              <div class="rounded-xl border border-white/10 bg-white/5 p-3">
                <div class="text-[10px] uppercase tracking-wider text-gray-500">Succeeded</div>
                <div class="text-sm font-bold text-green-400">{currentStats().succeeded}</div>
              </div>
              <div class="rounded-xl border border-white/10 bg-white/5 p-3">
                <div class="text-[10px] uppercase tracking-wider text-gray-500">Failed</div>
                <div class="text-sm font-bold text-red-400">{currentStats().failed}</div>
              </div>
            </div>
          </Show>

          <Show when={sortedTaskEntries().length > 0}>
            <div class="space-y-3">
              <For each={sortedTaskEntries()}>
                {([id, task]) => {
                  const percent = toFinitePercent(task?.percent);
                  return (
                    <div class="space-y-1">
                      <div class="flex items-center justify-between text-xs">
                        <span class="text-gray-400 truncate flex-1">{task?.label || id}</span>
                        <span class="text-gray-500 ml-2">
                          {formatPercent(percent, Boolean(task?.done))}
                          {task?.total > 0 ? ` · ${humanBytes(task.current)} / ${humanBytes(task.total)}` : ''}
                        </span>
                      </div>
                      <div class="w-full h-1.5 bg-white/5 rounded-full overflow-hidden">
                        <div
                          class={`h-full rounded-full transition-all duration-300 ${task?.done ? 'bg-green-500' : 'bg-blue-500'}`}
                          style={{ width: `${Math.min(100, percent)}%` }}
                        ></div>
                      </div>
                    </div>
                  );
                }}
              </For>
            </div>
          </Show>

          <Show when={logMessages().length > 0}>
            <div class="max-h-32 overflow-y-auto custom-scrollbar space-y-0.5">
              <For each={logMessages()}>
                {(log) => (
                  <div class={`text-[11px] font-mono px-2 py-0.5 rounded break-words ${
                    log.level === 'error' ? 'text-red-400' : log.level === 'warn' ? 'text-amber-400' : 'text-gray-500'
                  }`}>
                    {log.message}
                  </div>
                )}
              </For>
            </div>
          </Show>
        </div>
      </Show>

      {isAdvanced() && (
        <div class="p-8 bg-[#0a0c14] border border-white/5 rounded-[2rem] space-y-6 animate-in zoom-in-95 duration-300">
          <h3 class="font-bold text-white flex items-center gap-2">
            <Icon name="terminal" class="w-4 h-4 text-blue-400" />
            Power User Options
          </h3>
          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-xs font-bold text-gray-500">Output Template</label>
              <input
                value={settings().output}
                onInput={(e) => setSettings({ ...settings(), output: e.target.value })}
                class="w-full bg-[#05070a] border border-white/10 rounded-xl p-3 outline-none focus:border-blue-500 text-gray-300"
              />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-bold text-gray-500">Concurrent Jobs</label>
              <input
                type="number"
                value={settings().jobs}
                onInput={(e) => setSettings({ ...settings(), jobs: e.target.value })}
                class="w-full bg-[#05070a] border border-white/10 rounded-xl p-3 outline-none focus:border-blue-500 text-gray-300"
              />
            </div>
            <div class="space-y-2 col-span-2">
              <label class="text-xs font-bold text-gray-500">Duplicate Policy</label>
              <select
                value={settings().onDuplicate || 'prompt'}
                onChange={(e) => setSettings({ ...settings(), onDuplicate: e.target.value })}
                class="w-full bg-[#05070a] border border-white/10 rounded-xl p-3 outline-none focus:border-blue-500 text-gray-300"
              >
                <option value="prompt">Prompt (default)</option>
                <option value="overwrite">Overwrite</option>
                <option value="skip">Skip</option>
                <option value="rename">Rename</option>
              </select>
            </div>
          </div>
        </div>
      )}

      <Show when={activeDuplicate()}>
        <DuplicateModal
          prompt={activeDuplicate()}
          error={duplicateError()}
          onSelect={submitDuplicateChoice}
        />
      </Show>
    </div>
  );
}
