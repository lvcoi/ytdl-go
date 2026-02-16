const acceptedDownloadStatuses = new Set(['queued', 'running', 'reconnecting', 'complete', 'error']);
const activeDownloadStreamStatuses = new Set(['queued', 'running', 'reconnecting']);
const terminalDownloadStatuses = new Set(['complete', 'error']);

export const normalizeDownloadStatus = (value) => {
  if (typeof value !== 'string') return '';
  let normalized = value.trim().toLowerCase();
  if (normalized === 'done') normalized = 'complete';
  return acceptedDownloadStatuses.has(normalized) ? normalized : '';
};

export const isActiveDownloadStreamStatus = (status) => activeDownloadStreamStatuses.has(status);

export const isTerminalDownloadStatus = (status) => terminalDownloadStatuses.has(status);
