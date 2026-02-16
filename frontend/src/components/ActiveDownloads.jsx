import { For, Show, createMemo } from 'solid-js';
import Icon from './Icon';

import { useAppStore } from '../store/appStore';

export default function ActiveDownloads() {
    const { state } = useAppStore();

    const jobStatus = () => state.download.jobStatus;
    const progressTasks = () => state.download.progressTasks;

    const sortedTasks = createMemo(() => {
        const tasks = progressTasks();
        if (!tasks) return [];
        return Object.entries(tasks).sort(([, a], [, b]) => {
            // Sort by done status (running first), then by label
            if (a.done !== b.done) return a.done ? 1 : -1;
            return (a.label || '').localeCompare(b.label || '');
        });
    });

    const hasActiveDownloads = createMemo(() => sortedTasks().length > 0);

    const overallStatus = createMemo(() => jobStatus()?.status || 'idle');
    const overallMessage = createMemo(() => jobStatus()?.message || '');

    const statusColor = createMemo(() => {
        const status = overallStatus();
        if (status === 'error') return 'text-red-400';
        if (status === 'complete') return 'text-emerald-400';
        if (status === 'running') return 'text-accent-primary';
        return 'text-gray-400';
    });

    return (
        <div class="rounded-[2rem] border border-white/5 bg-black/20 p-6 flex flex-col gap-4 h-full">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-bold text-white flex items-center gap-2">
                    <Icon name="download-cloud" class="w-5 h-5 text-accent-primary" />
                    Active Downloads
                </h3>
                <span class={`text-xs font-bold uppercase tracking-widest ${statusColor()}`}>
                    {overallStatus()}
                </span>
            </div>

            <Show when={hasActiveDownloads()} fallback={
                <div class="flex flex-col items-center justify-center flex-1 py-12 text-gray-500 gap-3">
                    <div class="w-12 h-12 rounded-2xl bg-white/5 flex items-center justify-center mb-1">
                        <Icon 
                            name={overallStatus() === 'running' ? "loader-2" : "check-circle-2"} 
                            class={`w-6 h-6 ${overallStatus() === 'running' ? 'text-accent-primary animate-spin' : 'opacity-20'}`} 
                        />
                    </div>
                    <span class="text-xs font-bold uppercase tracking-widest text-center px-4">
                        {overallStatus() === 'running' ? 'Initializing Download...' : 'No active tasks'}
                    </span>
                    
                    <Show when={jobStatus()?.error}>
                        <div class="mt-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-[10px] text-red-400 font-mono text-left w-full break-all">
                            <span class="font-bold uppercase mr-1">[Error]</span>
                            {jobStatus().error}
                        </div>
                    </Show>
                </div>
            }>
                <div class="space-y-3 overflow-y-auto custom-scrollbar flex-1 pr-2 max-h-[300px]">
                    <Show when={overallMessage()}>
                        <div class={`text-xs font-medium px-1 truncate mb-2 ${overallStatus() === 'error' ? 'text-red-400' : 'text-gray-400'}`}>
                            {overallMessage()}
                        </div>
                    </Show>
                    <For each={sortedTasks()}>
                        {([id, task]) => (
                            <div class="rounded-xl border border-white/5 bg-white/5 p-3 space-y-2">
                                <div class="flex justify-between items-center text-xs">
                                    <span class="font-bold text-slate-200 truncate max-w-[70%]">{task.label || id}</span>
                                    <span class={`font-mono ${task.done ? 'text-emerald-400' : 'text-accent-primary'}`}>
                                        {task.percent?.toFixed(1)}%
                                    </span>
                                </div>
                                <div class="h-1.5 w-full rounded-full bg-black/40 overflow-hidden">
                                    <div
                                        class={`h-full transition-all duration-300 ${task.done ? 'bg-emerald-500' : 'bg-vibrant-gradient'}`}
                                        style={{ width: `${task.percent}%` }}
                                    />
                                </div>
                            </div>
                        )}
                    </For>
                </div>
            </Show>
        </div>
    );
}

