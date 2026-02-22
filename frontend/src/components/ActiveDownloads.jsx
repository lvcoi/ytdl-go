import { For, Show, createMemo } from 'solid-js';
import Icon from './Icon';
import { downloadStore } from '../store/downloadStore';

export default function ActiveDownloads() {
    // Sorted tasks from global store
    const sortedTasks = createMemo(() => {
        const tasks = Object.values(downloadStore.activeDownloads);
        // Sort: active first, then by ID or name
        return tasks.sort((a, b) => {
             // Error/Running first, complete last
             if (a.done !== b.done) return a.done ? 1 : -1;
             return (a.filename || a.id).localeCompare(b.filename || b.id);
        });
    });

    const hasTasks = createMemo(() => sortedTasks().length > 0);

    return (
        <div class="rounded-[2rem] border border-white/5 bg-black/20 p-6 flex flex-col gap-4 h-full">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-bold text-white flex items-center gap-2">
                    <Icon name="download-cloud" class="w-5 h-5 text-accent-primary" />
                    Active Downloads
                </h3>
            </div>

            <Show when={hasTasks()} fallback={
                <div class="flex flex-col items-center justify-center flex-1 py-12 text-gray-500 gap-3">
                    <Icon name="check-circle-2" class="w-10 h-10 opacity-20" />
                    <span class="text-xs font-bold uppercase tracking-widest">No active tasks</span>
                </div>
            }>
                <div
                    class="space-y-3 overflow-y-auto custom-scrollbar flex-1 pr-2 max-h-[300px]"
                    style={{
                        "mask-image": "linear-gradient(to bottom, black 0%, black 85%, transparent 100%)",
                        "-webkit-mask-image": "linear-gradient(to bottom, black 0%, black 85%, transparent 100%)"
                    }}
                >
                    <For each={sortedTasks()}>
                        {(task) => (
                            <div
                                class="rounded-xl border border-white/5 bg-white/5 p-3 space-y-2 transition-all duration-300"
                                classList={{
                                    "opacity-0 scale-95": task.done,
                                    "opacity-100 scale-100": !task.done,
                                }}
                            >
                                <div class="flex justify-between items-center text-xs">
                                    <span class="font-bold text-slate-200 truncate max-w-[70%]">
                                        {task.filename || task.id}
                                    </span>
                                    <span class={`font-mono ${
                                        task.status === 'error' ? 'text-red-400' : 
                                        task.done ? 'text-emerald-400' : 'text-accent-primary'
                                    }`}>
                                        {task.status === 'error' ? 'Error' : 
                                         task.percent !== undefined ? `${task.percent.toFixed(1)}%` : '0.0%'}
                                    </span>
                                </div>
                                
                                <Show when={task.status === 'error'}>
                                    <div class="text-[10px] text-red-400 font-mono break-all">
                                        {task.error}
                                    </div>
                                </Show>

                                <div class="h-1.5 w-full rounded-full bg-black/40 overflow-hidden">
                                    <div
                                        class={`h-full transition-all duration-300 ${
                                            task.status === 'error' ? 'bg-red-500' :
                                            task.done ? 'bg-emerald-500' : 'bg-vibrant-gradient'
                                        }`}
                                        style={{ width: `${task.percent || 0}%` }}
                                    />
                                </div>
                                
                                <Show when={task.eta && !task.done && task.status !== 'error'}>
                                    <div class="text-[10px] text-slate-500 text-right">
                                        ETA: {task.eta}
                                    </div>
                                </Show>
                            </div>
                        )}
                    </For>
                </div>
            </Show>
        </div>
    );
}
