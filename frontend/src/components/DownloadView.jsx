import { createSignal, For, Show, onCleanup } from 'solid-js';
import Icon from './Icon';
import DuplicateModal from './DuplicateModal';

export default function DownloadView({ settings, setSettings, isAdvanced }) {
  const [urlInput, setUrlInput] = createSignal('');
  const [isDownloading, setIsDownloading] = createSignal(false);
  const [jobStatus, setJobStatus] = createSignal(null);
  const [progressTasks, setProgressTasks] = createSignal({});
  const [logMessages, setLogMessages] = createSignal([]);
  const [duplicateQueue, setDuplicateQueue] = createSignal([]);
  const [duplicateError, setDuplicateError] = createSignal('');

  const reconnectDelaysMs = [1000, 2000, 4000, 8000, 10000];
  const maxReconnectAttempts = 5;

  let eventSource = null;
  let reconnectTimer = null;
  let reconnectAttempts = 0;
  let activeJobId = '';

  const activeDuplicate = () => duplicateQueue()[0];

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
          // Prompt may have expired or closed server-side; drop stale entry and continue.
          setDuplicateQueue((prev) => prev.slice(1));
          setDuplicateError('');
          return;
        }
        setDuplicateError(data.error || 'Failed to submit duplicate choice');
        return;
      }
      setDuplicateQueue((prev) => prev.slice(1));
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

  if (typeof window !== 'undefined') {
    window.addEventListener('keydown', handleDuplicateShortcut);
  }

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

  const markStreamConnected = () => {
    reconnectAttempts = 0;
    clearReconnectTimer();
    setJobStatus((prev) => {
      if (!prev || prev.status !== 'reconnecting') return prev;
      return {
        ...prev,
        status: 'running',
        message: 'Download in progress...',
        error: '',
      };
    });
  };

  onCleanup(() => {
    resetProgressStreamState();
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleDuplicateShortcut);
    }
  });

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
      setJobStatus({ status: 'error', error: `${label}: ${preview}${suffix}` });
      setIsDownloading(false);
      return;
    }
    if (urls.length === 0) {
      setJobStatus({ status: 'error', error: 'No valid URLs provided.' });
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
        jobs: parseInt(s.jobs, 10) || 1,
        timeout: parseInt(s.timeout, 10) || 180,
        'on-duplicate': s.onDuplicate || 'prompt',
      }
    };

    try {
      const res = await fetch('/api/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (data.error) {
        setJobStatus({ status: 'error', error: data.error });
        setIsDownloading(false);
        return;
      }
      setJobStatus({ status: 'running', jobId: data.jobId, message: data.message });
      listenForProgress(data.jobId);
    } catch (e) {
      setJobStatus({ status: 'error', error: e.message });
      setIsDownloading(false);
    }
  };

  const listenForProgress = (jobId) => {
    resetProgressStreamState();
    activeJobId = jobId;

    const connect = () => {
      if (activeJobId !== jobId) return;
      closeProgressStream();
      eventSource = new EventSource(`/api/download/progress?id=${encodeURIComponent(jobId)}`);

      eventSource.onmessage = (e) => {
        if (activeJobId !== jobId) return;
        try {
          const evt = JSON.parse(e.data);
          markStreamConnected();
          switch (evt.type) {
            case 'register':
              setProgressTasks((prev) => ({ ...prev, [evt.id]: { label: evt.label, total: evt.total, current: 0, percent: 0 } }));
              break;
            case 'progress':
              setProgressTasks((prev) => ({ ...prev, [evt.id]: { ...prev[evt.id], current: evt.current, total: evt.total, percent: evt.percent } }));
              break;
            case 'finish':
              setProgressTasks((prev) => ({ ...prev, [evt.id]: { ...prev[evt.id], percent: 100, done: true } }));
              break;
            case 'log':
              setLogMessages((prev) => [...prev, { level: evt.level, message: evt.message }].slice(-50));
              break;
            case 'duplicate':
              setDuplicateQueue((prev) => [...prev, {
                jobId,
                promptId: evt.promptId,
                path: evt.path,
                filename: evt.filename,
              }]);
              setDuplicateError('');
              break;
            case 'done':
              resetProgressStreamState();
              setJobStatus((prev) => ({
                ...(prev || {}),
                status: evt.message === 'complete' ? 'complete' : 'error',
                error: evt.message === 'complete' ? '' : (prev?.error || 'Download failed'),
              }));
              setIsDownloading(false);
              setDuplicateQueue([]);
              setDuplicateError('');
              break;
          }
        } catch (_) {}
      };

      eventSource.onerror = () => {
        if (activeJobId !== jobId) return;
        closeProgressStream();
        clearReconnectTimer();

        const state = jobStatus()?.status;
        if (state !== 'running' && state !== 'reconnecting') {
          return;
        }

        if (reconnectAttempts >= maxReconnectAttempts) {
          activeJobId = '';
          reconnectAttempts = 0;
          setJobStatus((prev) => ({
            ...(prev || {}),
            status: 'error',
            error: 'Connection lost. Progress updates stopped.',
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

  const humanBytes = (bytes) => {
    if (!bytes || bytes <= 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i];
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
            onClick={() => setSettings({...settings(), audioOnly: !settings().audioOnly})} 
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
        <div class={`p-6 rounded-3xl border-2 space-y-4 animate-in fade-in duration-300 ${
          jobStatus()?.status === 'error' ? 'bg-red-500/5 border-red-500/20'
            : jobStatus()?.status === 'complete' ? 'bg-green-500/5 border-green-500/20'
            : jobStatus()?.status === 'reconnecting' ? 'bg-amber-500/5 border-amber-500/20'
            : 'bg-blue-500/5 border-blue-500/20'
        }`}>
          <div class="flex items-center gap-3">
            <div class={`p-2 rounded-xl ${
              jobStatus()?.status === 'error' ? 'bg-red-500/10 text-red-400'
                : (jobStatus()?.status === 'complete') ? 'bg-green-500/10 text-green-400'
                : (jobStatus()?.status === 'reconnecting') ? 'bg-amber-500/10 text-amber-400'
                : 'bg-blue-500/10 text-blue-400'
            }`}>
              <Icon
                name={jobStatus()?.status === 'error' ? 'alert-circle' : (jobStatus()?.status === 'complete') ? 'check-circle-2' : 'loader'}
                class={`w-5 h-5 ${jobStatus()?.status !== 'complete' && jobStatus()?.status !== 'error' ? 'animate-spin' : ''}`}
              />
            </div>
            <div>
              <div class="font-bold text-white">
                {jobStatus()?.status === 'error' ? 'Download Failed'
                  : jobStatus()?.status === 'complete' ? 'Download Complete'
                  : jobStatus()?.status === 'reconnecting' ? 'Reconnecting...'
                  : 'Downloading...'}
              </div>
              <Show when={jobStatus()?.message}>
                <div class="text-xs text-gray-500">{jobStatus().message}</div>
              </Show>
              <Show when={jobStatus()?.status === 'error' && jobStatus()?.error}>
                <div class="text-xs text-red-400">{jobStatus().error}</div>
              </Show>
            </div>
          </div>

          <Show when={Object.keys(progressTasks()).length > 0}>
            <div class="space-y-3">
              <For each={Object.entries(progressTasks())}>
                {([id, task]) => (
                  <div class="space-y-1">
                    <div class="flex items-center justify-between text-xs">
                      <span class="text-gray-400 truncate flex-1">{task.label}</span>
                      <span class="text-gray-500 ml-2">
                        {task.done ? '100%' : `${(task.percent || 0).toFixed(1)}%`}
                        {task.total > 0 ? ` · ${humanBytes(task.current)} / ${humanBytes(task.total)}` : ''}
                      </span>
                    </div>
                    <div class="w-full h-1.5 bg-white/5 rounded-full overflow-hidden">
                      <div
                        class={`h-full rounded-full transition-all duration-300 ${task.done ? 'bg-green-500' : 'bg-blue-500'}`}
                        style={{ width: `${Math.min(100, task.percent || 0)}%` }}
                      ></div>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>

          <Show when={logMessages().length > 0}>
            <div class="max-h-32 overflow-y-auto custom-scrollbar space-y-0.5">
              <For each={logMessages()}>
                {(log) => (
                  <div class={`text-[11px] font-mono px-2 py-0.5 rounded ${
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
                        onInput={(e) => setSettings({...settings(), output: e.target.value})} 
                        class="w-full bg-[#05070a] border border-white/10 rounded-xl p-3 outline-none focus:border-blue-500 text-gray-300" 
                    />
                </div>
                <div class="space-y-2">
                    <label class="text-xs font-bold text-gray-500">Concurrent Jobs</label>
                    <input 
                        type="number" 
                        value={settings().jobs} 
                        onInput={(e) => setSettings({...settings(), jobs: e.target.value})} 
                        class="w-full bg-[#05070a] border border-white/10 rounded-xl p-3 outline-none focus:border-blue-500 text-gray-300" 
                    />
                </div>
                <div class="space-y-2 col-span-2">
                    <label class="text-xs font-bold text-gray-500">Duplicate Policy</label>
                    <select
                        value={settings().onDuplicate || 'prompt'}
                        onChange={(e) => setSettings({...settings(), onDuplicate: e.target.value})}
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
