import { createSignal, createResource, Show, splitProps } from 'solid-js';
import Icon from './Icon';

const fetchSystemInfo = async () => {
    try {
        const res = await fetch('/api/system/info');
        if (!res.ok) throw new Error('Failed to fetch system info');
        return await res.json();
    } catch (err) {
        console.error('Failed to fetch system info:', err);
        return { cpuCores: 4 }; // Fallback default
    }
};

/**
 * Smart concurrency heuristics:
 * - Playlist: Job Concurrency = max(1, CPU - 1)
 * - Single large file (>500MB): Segment Concurrency = floor(CPU * 1.5)
 */
const computeSmartDefaults = (cpuCores, downloadType) => {
    if (downloadType === 'playlist') {
        return {
            jobConcurrency: Math.max(1, cpuCores - 1),
            segmentConcurrency: Math.max(1, Math.min(cpuCores, 4)),
        };
    }
    // Single large file
    return {
        jobConcurrency: 1,
        segmentConcurrency: Math.max(1, Math.floor(cpuCores * 1.5)),
    };
};

export default function ConcurrencyWidget(props) {
    const [local] = splitProps(props, ['onSettingsChange']);
    const [expanded, setExpanded] = createSignal(false);
    const [autoMode, setAutoMode] = createSignal(true);
    const [jobConcurrency, setJobConcurrency] = createSignal(1);
    const [segmentConcurrency, setSegmentConcurrency] = createSignal(4);
    const [downloadType, setDownloadType] = createSignal('single');

    const [systemInfo] = createResource(fetchSystemInfo);
    const cpuCores = () => systemInfo()?.cpuCores ?? 4;

    const effectiveJobConcurrency = () => {
        if (!autoMode()) return jobConcurrency();
        return computeSmartDefaults(cpuCores(), downloadType()).jobConcurrency;
    };

    const effectiveSegmentConcurrency = () => {
        if (!autoMode()) return segmentConcurrency();
        return computeSmartDefaults(cpuCores(), downloadType()).segmentConcurrency;
    };

    const handleAutoToggle = () => {
        const newAuto = !autoMode();
        setAutoMode(newAuto);
        if (newAuto) {
            const defaults = computeSmartDefaults(cpuCores(), downloadType());
            setJobConcurrency(defaults.jobConcurrency);
            setSegmentConcurrency(defaults.segmentConcurrency);
        }
        notifyChange();
    };

    const notifyChange = () => {
        if (typeof local.onSettingsChange === 'function') {
            local.onSettingsChange({
                jobConcurrency: effectiveJobConcurrency(),
                segmentConcurrency: effectiveSegmentConcurrency(),
                autoMode: autoMode(),
            });
        }
    };

    return (
        <div class="rounded-2xl border border-white/5 bg-black/20 overflow-hidden">
            {/* Toggle Header */}
            <button
                class="w-full flex items-center justify-between px-5 py-3 text-sm font-bold text-slate-300 hover:text-white transition-colors"
                onClick={() => setExpanded(!expanded())}
            >
                <span class="flex items-center gap-2">
                    <Icon name="settings" class="w-4 h-4 text-accent-primary" />
                    Advanced Concurrency
                </span>
                <Icon
                    name="chevron-down"
                    class={`w-4 h-4 transition-transform duration-200 ${expanded() ? 'rotate-180' : ''}`}
                />
            </button>

            <Show when={expanded()}>
                <div class="px-5 pb-5 space-y-4 border-t border-white/5 pt-4">
                    {/* Auto/Smart Toggle */}
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2">
                            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider">
                                Auto / Smart Mode
                            </span>
                            <div class="group relative">
                                <Icon name="info" class="w-3.5 h-3.5 text-slate-500 cursor-help" />
                                <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block w-56 p-2 bg-slate-800 rounded-lg text-[10px] text-slate-300 shadow-xl z-10">
                                    Auto mode dynamically adjusts concurrency based on your CPU cores ({cpuCores()}) and the download type.
                                </div>
                            </div>
                        </div>
                        <button
                            class={`relative w-10 h-5 rounded-full transition-colors duration-200 ${
                                autoMode() ? 'bg-emerald-500' : 'bg-slate-600'
                            }`}
                            onClick={handleAutoToggle}
                        >
                            <span
                                class={`absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow transition-transform duration-200 ${
                                    autoMode() ? 'translate-x-5' : 'translate-x-0'
                                }`}
                            />
                        </button>
                    </div>

                    {/* Download Type (only visible in Auto mode) */}
                    <Show when={autoMode()}>
                        <div class="space-y-1.5">
                            <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                                Download Type Detection
                            </label>
                            <div class="flex gap-2">
                                <button
                                    class={`flex-1 px-3 py-1.5 rounded-lg text-xs font-bold transition-colors ${
                                        downloadType() === 'playlist'
                                            ? 'bg-accent-primary/20 text-accent-primary border border-accent-primary/30'
                                            : 'bg-white/5 text-slate-400 border border-white/5 hover:text-white'
                                    }`}
                                    onClick={() => { setDownloadType('playlist'); notifyChange(); }}
                                >
                                    Playlist
                                </button>
                                <button
                                    class={`flex-1 px-3 py-1.5 rounded-lg text-xs font-bold transition-colors ${
                                        downloadType() === 'single'
                                            ? 'bg-accent-primary/20 text-accent-primary border border-accent-primary/30'
                                            : 'bg-white/5 text-slate-400 border border-white/5 hover:text-white'
                                    }`}
                                    onClick={() => { setDownloadType('single'); notifyChange(); }}
                                >
                                    Large File
                                </button>
                            </div>
                        </div>
                    </Show>

                    {/* Job Concurrency Slider */}
                    <div class="space-y-1.5">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-1.5">
                                <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                                    Job Concurrency
                                </label>
                                <div class="group relative">
                                    <Icon name="info" class="w-3 h-3 text-slate-600 cursor-help" />
                                    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block w-48 p-2 bg-slate-800 rounded-lg text-[10px] text-slate-300 shadow-xl z-10">
                                        Simultaneous post-processing tracks (CPU-bound). Higher values use more CPU.
                                    </div>
                                </div>
                            </div>
                            <span class="text-xs font-mono text-accent-primary font-bold">
                                {effectiveJobConcurrency()}
                            </span>
                        </div>
                        <input
                            type="range"
                            min="1"
                            max={Math.max(1, cpuCores() * 2)}
                            step="1"
                            value={effectiveJobConcurrency()}
                            disabled={autoMode()}
                            onInput={(e) => {
                                setJobConcurrency(parseInt(e.target.value));
                                notifyChange();
                            }}
                            class="w-full h-1 bg-white/10 rounded-full appearance-none cursor-pointer accent-accent-primary disabled:opacity-40"
                        />
                    </div>

                    {/* Segment Concurrency Slider */}
                    <div class="space-y-1.5">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-1.5">
                                <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                                    Segment Concurrency
                                </label>
                                <div class="group relative">
                                    <Icon name="info" class="w-3 h-3 text-slate-600 cursor-help" />
                                    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block w-48 p-2 bg-slate-800 rounded-lg text-[10px] text-slate-300 shadow-xl z-10">
                                        Parallel network stream downloads (I/O-bound). Higher values use more bandwidth.
                                    </div>
                                </div>
                            </div>
                            <span class="text-xs font-mono text-accent-primary font-bold">
                                {effectiveSegmentConcurrency()}
                            </span>
                        </div>
                        <input
                            type="range"
                            min="1"
                            max={Math.max(1, cpuCores() * 3)}
                            step="1"
                            value={effectiveSegmentConcurrency()}
                            disabled={autoMode()}
                            onInput={(e) => {
                                setSegmentConcurrency(parseInt(e.target.value));
                                notifyChange();
                            }}
                            class="w-full h-1 bg-white/10 rounded-full appearance-none cursor-pointer accent-accent-primary disabled:opacity-40"
                        />
                    </div>

                    {/* CPU Info */}
                    <div class="text-[10px] text-slate-600 text-center pt-1">
                        Detected {cpuCores()} CPU cores
                    </div>
                </div>
            </Show>
        </div>
    );
}
