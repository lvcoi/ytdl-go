import { For, Show, onCleanup, onMount, createMemo } from 'solid-js';
import Icon from './Icon';
import Thumbnail from './Thumbnail';
import { Grid, GridItem } from './Grid';
import DuplicateModal from './DuplicateModal';
import { MAX_JOBS, MAX_TIMEOUT_SECONDS, useAppStore } from '../store/appStore';
import { downloadStore, setDownloadStore } from '../store/downloadStore';
import { useDownloadManager } from '../hooks/useDownloadManager';
import {
  normalizeDownloadStatus,
} from '../utils/downloadStatus';
import ActiveDownloads from './ActiveDownloads';

const normalizeStatus = normalizeDownloadStatus;
const statusTone = (status) => {
  if (status === 'error') {
    return {
      card: 'bg-red-500/5 border-red-500/20',
      icon: 'bg-red-500/10 text-red-400',
      accent: 'text-red-400',
      bar: 'bg-red-500',
    };
  }
  if (status === 'complete') {
    return {
      card: 'bg-green-500/5 border-green-500/20',
      icon: 'bg-green-500/10 text-green-400',
      accent: 'text-green-400',
      bar: 'bg-green-500',
    };
  }
  if (status === 'reconnecting') {
    return {
      card: 'bg-amber-500/5 border-amber-500/20',
      icon: 'bg-amber-500/10 text-amber-400',
      accent: 'text-amber-400',
      bar: 'bg-amber-500',
    };
  }
  return {
    card: 'glass-vibrant',
    icon: 'bg-accent-primary/10 text-accent-primary',
    accent: 'text-accent-primary',
    bar: 'bg-vibrant-gradient',
  };
};

const toNonNegativeInteger = (value, fallback = 0) => {
  const parsed = typeof value === 'number' ? value : Number(value);
  if (!Number.isFinite(parsed)) return fallback;
  return Math.max(0, Math.trunc(parsed));
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

const statusIconName = (status) => {
  const s = normalizeStatus(status);
  if (s === 'error') return 'alert-circle';
  if (s === 'complete') return 'check-circle-2';
  if (s === 'reconnecting') return 'loader';
  if (s === 'running') return 'loader';
  if (s === 'queued') return 'loader';
  return 'download-cloud';
};

const statusTitle = (status) => {
  const s = normalizeStatus(status);
  if (s === 'error') return 'Download Failed';
  if (s === 'complete') return 'Download Complete';
  if (s === 'reconnecting') return 'Reconnecting...';
  if (s === 'running') return 'Downloading...';
  if (s === 'queued') return 'Queued';
  return 'Ready';
};

const statusDefaultMessage = (status) => {
  const s = normalizeStatus(status);
  if (s === 'error') return 'An unexpected error occurred.';
  if (s === 'complete') return 'All tasks finished successfully.';
  if (s === 'reconnecting') return 'Connection interrupted, attempting to resume...';
  if (s === 'running') return 'Processing download tasks...';
  if (s === 'queued') return 'Waiting for available worker slot...';
  return 'Ready to start.';
};

export default function DownloadView(props = {}) {
  const { state, setState } = useAppStore();
  const { startDownload, cancelDownload } = useDownloadManager();

  const urlInput = () => state.download.urlInput;

  const setUrlInput = (nextValue) => {
    setState('download', 'urlInput', nextValue);
  };

  const currentJobId = () => downloadStore.lastJobId;
  const jobStatus = () => currentJobId() ? downloadStore.jobStatuses[currentJobId()] : null;
  
  const isDownloading = () => {
      const status = jobStatus()?.status;
      return status === 'queued' || status === 'running' || status === 'reconnecting';
  };

    const logMessages = () => downloadStore.logs;

  const settings = () => state.settings;
  const isAdvanced = () => state.ui.isAdvanced;

  const setSettings = (nextSettings) => {
    if (typeof nextSettings === 'function') {
      setState('settings', (prev) => {
        const resolved = nextSettings(prev);
        if (!resolved || typeof resolved !== 'object') {
          return prev;
        }
        return { ...prev, ...resolved };
      });
      return;
    }
    setState('settings', (prev) => ({ ...prev, ...nextSettings }));
  };

  const currentStats = () => normalizeStats(jobStatus()?.stats);

  const currentStatus = () => normalizeStatus(jobStatus()?.status);
  const currentStatusTone = () => statusTone(currentStatus());

    const openLibrary = () => {
    if (typeof props.onOpenLibrary === 'function') {
      props.onOpenLibrary();
    }
  };

  const handleDownload = async () => {

    if (isDownloading()) return;
    await startDownload(urlInput());
  };

  return (
    <div class="space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-500">

      {/* Hero Section / Input */}
      <div class="space-y-6">
        <div class="flex flex-col gap-2">
          <h1 class="text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-emerald-400">
            Download Media
          </h1>
          <p class="text-gray-400 text-sm">
            Paste one or more URLs below to start downloading. Supports YouTube videos, playlists, and channels.
          </p>
        </div>

        <div class="relative group">
          <div class="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-emerald-500/20 rounded-2xl blur-xl transition-opacity opacity-0 group-focus-within:opacity-100" />
          <div class="relative glass rounded-2xl border border-white/10 p-2 flex flex-col md:flex-row gap-2 transition-colors focus-within:border-blue-500/50 focus-within:bg-black/40">
            <textarea
              value={urlInput()}
              onInput={(e) => setUrlInput(e.target.value)}
              placeholder="https://www.youtube.com/watch?v=..."
              rows={3}
              class="flex-1 bg-transparent text-white placeholder-gray-500 p-3 text-sm focus:outline-none resize-none custom-scrollbar font-mono leading-relaxed"
              spellcheck={false}
            />
            <div class="flex md:flex-col justify-end gap-2 p-1">
              <button
                onClick={() => setUrlInput('')}
                disabled={!urlInput()}
                class="p-2 rounded-xl text-gray-500 hover:text-white hover:bg-white/10 transition-all disabled:opacity-0"
                title="Clear input"
              >
                <Icon name="x" class="w-5 h-5" />
              </button>
              <button
                onClick={handleDownload}
                disabled={isDownloading() || !urlInput().trim()}
                class={`p-3 rounded-xl flex items-center justify-center gap-2 font-bold shadow-lg transition-all ${isDownloading()
                    ? 'bg-white/5 text-gray-500 cursor-not-allowed'
                    : 'bg-blue-600 hover:bg-blue-500 text-white hover:shadow-blue-500/25 active:scale-95'
                  }`}
                title="Start Download"
              >
                <Show when={isDownloading()} fallback={<Icon name="download" class="w-5 h-5" />}>
                  <Icon name="loader" class="w-5 h-5 animate-spin" />
                </Show>
              </button>
            </div>
          </div>
        </div>
      </div>

      <Grid class="!p-0 !gap-8">
        <div class="lg:col-span-3 space-y-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-3">
              <label class="text-xs font-bold text-gray-500 uppercase tracking-wider flex items-center gap-2">
                <Icon name="file-audio" class="w-3.5 h-3.5" />
                Format
              </label>
              <div class="grid grid-cols-2 gap-2">
                <button
                  onClick={() => setSettings({ audioOnly: false })}
                  class={`p-3 rounded-xl border text-sm font-semibold transition-all flex items-center justify-center gap-2 ${!settings().audioOnly
                      ? 'border-blue-500/50 bg-blue-500/10 text-blue-400'
                      : 'border-white/5 bg-white/5 text-gray-400 hover:bg-white/10'
                    }`}
                >
                  <Icon name="video" class="w-4 h-4" />
                  Video + Audio
                </button>
                <button
                  onClick={() => setSettings({ audioOnly: true })}
                  class={`p-3 rounded-xl border text-sm font-semibold transition-all flex items-center justify-center gap-2 ${settings().audioOnly
                      ? 'border-purple-500/50 bg-purple-500/10 text-purple-400'
                      : 'border-white/5 bg-white/5 text-gray-400 hover:bg-white/10'
                    }`}
                >
                  <Icon name="music" class="w-4 h-4" />
                  Audio Only
                </button>
              </div>
            </div>

            <div class="space-y-3">
              <label class="text-xs font-bold text-gray-500 uppercase tracking-wider flex items-center gap-2">
                <Icon name="sliders" class="w-3.5 h-3.5" />
                Quality
              </label>
              <div class="relative">
                <select
                  value={settings().quality}
                  onChange={(e) => setSettings({ quality: e.target.value })}
                  class="w-full appearance-none bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-medium text-gray-200 focus:outline-none focus:border-blue-500/50 transition-colors cursor-pointer"
                >
                  <option value="best" class="bg-[#0f172a]">Best Available</option>
                  <option value="1080p" class="bg-[#0f172a]">1080p (HD)</option>
                  <option value="720p" class="bg-[#0f172a]">720p (HD)</option>
                  <option value="480p" class="bg-[#0f172a]">480p</option>
                  <option value="worst" class="bg-[#0f172a]">Worst / Smallest</option>
                </select>
                <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-gray-500">
                  <Icon name="chevron-down" class="w-4 h-4" />
                </div>
              </div>
            </div>
          </div>

          <Show when={isAdvanced()}>
            <div class="pt-4 border-t border-white/5 grid grid-cols-1 md:grid-cols-2 gap-4 animate-in slide-in-from-top-2">
              <div class="space-y-2">
                <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Filename Template</label>
                <input
                  type="text"
                  value={settings().output}
                  onInput={(e) => setSettings({ output: e.target.value })}
                  class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm font-mono text-gray-300 focus:outline-none focus:border-blue-500/50"
                />
                <p class="text-[10px] text-gray-500">Variables: {'{title}'}, {'{id}'}, {'{upload_date}'}, {'{channel}'}</p>
              </div>

              <div class="space-y-2">
                <label class="text-xs font-bold text-gray-500 uppercase tracking-wider">Duplicate Handling</label>
                <select
                  value={settings().onDuplicate}
                  onChange={(e) => setSettings({ onDuplicate: e.target.value })}
                  class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm text-gray-300 focus:outline-none focus:border-blue-500/50 appearance-none"
                >
                  <option value="prompt" class="bg-[#0f172a]">Ask me</option>
                  <option value="overwrite" class="bg-[#0f172a]">Overwrite</option>
                  <option value="skip" class="bg-[#0f172a]">Skip</option>
                  <option value="rename" class="bg-[#0f172a]">Rename</option>
                </select>
              </div>
            </div>
          </Show>
        </div>

        <GridItem class="lg:col-span-2">
          <Show when={jobStatus()}>
            <div class={`p-8 rounded-[2rem] border min-h-[400px] flex flex-col gap-6 transition-all duration-500 ${currentStatusTone().card}`}>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-4">
                                    <div class={`p-3 rounded-xl ${currentStatusTone().icon}`}>
                    <Show when={!downloadStore.isDisconnected} fallback={<Icon name="wifi-off" class="w-6 h-6 text-red-400 animate-pulse" />}>
                      <Icon name={statusIconName(currentStatus())} class="w-6 h-6 animate-pulse" />
                    </Show>
                  </div>

                                    <div>
                    <div class={`text-lg font-bold ${currentStatusTone().accent}`}>
                      {statusTitle(currentStatus())}
                    </div>
                    <div class="flex items-center gap-2">
                      <div class="text-sm text-gray-400 max-w-md truncate">
                        {jobStatus()?.message || statusDefaultMessage(currentStatus())}
                      </div>
                      <Show when={isDownloading()}>
                        <button 
                          onClick={() => cancelDownload(currentJobId())}
                          class="p-1 hover:bg-red-500/10 rounded-lg text-red-400/60 hover:text-red-400 transition-colors"
                          title="Stop Download"
                        >
                          <Icon name="x" class="w-3.5 h-3.5" />
                        </button>
                      </Show>
                    </div>
                  </div>

                </div>

                <Show when={currentStats()}>
                  <div class="flex items-center gap-6 text-sm">
                    <For each={[
                      { label: 'Total', value: currentStats()?.total, color: 'text-gray-300' },
                      { label: 'Success', value: currentStats()?.succeeded, color: 'text-emerald-400' },
                      { label: 'Failed', value: currentStats()?.failed, color: 'text-red-400' },
                    ]}>
                      {(stat) => (
                        <div class="flex flex-col items-center">
                          <span class="text-[10px] font-bold uppercase tracking-widest text-gray-500">{stat.label}</span>
                          <span class={`font-mono font-bold ${stat.color}`}>{stat.value}</span>
                        </div>
                      )}
                    </For>
                  </div>
                </Show>
              </div>

              <div class="flex-1 min-h-0 border-t border-white/5 pt-6">
                <ActiveDownloads />
              </div>

              <Show when={jobStatus()?.error}>
                <div class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-200 text-sm flex items-start gap-3">
                  <Icon name="alert-triangle" class="w-5 h-5 shrink-0 text-red-400" />
                  <p class="leading-relaxed">{jobStatus()?.error}</p>
                </div>
              </Show>

              <Show when={logMessages().length > 0}>
                <div class="p-4 rounded-xl bg-black/40 border border-white/5 font-mono text-xs text-gray-400 space-y-1 max-h-32 overflow-y-auto custom-scrollbar">
                  <For each={logMessages()}>
                    {(log, i) => (
                      <div class={`flex gap-2 ${log.level === 'error' ? 'text-red-400' : ''}`}>
                        <span class="opacity-50 select-none">{String(i()).padStart(2, '0')}</span>
                        <span>{log.message}</span>
                      </div>
                    )}
                  </For>
                </div>
              </Show>

              <Show when={currentStatus() === 'complete' || currentStatus() === 'error'}>
                <div class="flex justify-center pt-2">
                  <button
                    onClick={openLibrary}
                    class="px-6 py-2 rounded-full border border-white/10 bg-white/5 hover:bg-white/10 text-sm font-bold text-gray-300 transition-all hover:scale-105"
                  >
                    View in Library
                  </button>
                </div>
              </Show>
            </div>
          </Show>

          <Show when={!jobStatus()}>
            <div class="h-64 rounded-[2rem] border border-white/5 bg-white/5 flex flex-col items-center justify-center text-gray-500 gap-4">
              <div class="w-16 h-16 rounded-3xl bg-black/40 flex items-center justify-center">
                <Icon name="download-cloud" class="w-8 h-8 opacity-50" />
              </div>
              <p class="font-medium">Ready to download</p>
            </div>
          </Show>
        </GridItem>
            </Grid>
    </div>
  );
}

