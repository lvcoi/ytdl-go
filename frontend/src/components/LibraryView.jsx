import { For } from 'solid-js';
import Icon from './Icon';

export default function LibraryView({ downloads, openPlayer }) {
  return (
    <div class="space-y-6 animate-in fade-in slide-in-from-right-4 duration-500">
        <div class="flex items-center justify-between mb-8">
            <div>
                <h1 class="text-3xl font-black text-white">Your Library</h1>
                <p class="text-gray-500">Access your downloaded media instantly.</p>
            </div>
            <div class="flex gap-2">
                <button class="p-3 bg-white/5 rounded-xl hover:text-white transition-all"><Icon name="search" class="w-5 h-5" /></button>
                <button class="p-3 bg-white/5 rounded-xl hover:text-white transition-all"><Icon name="filter" class="w-5 h-5" /></button>
            </div>
        </div>

        <div class="grid gap-3">
            <For each={downloads()}>
                {(item) => (
                    <div class="group flex items-center justify-between p-4 bg-[#0a0c14] border border-white/5 rounded-2xl hover:border-blue-500/30 transition-all cursor-default">
                        <div class="flex items-center gap-5">
                            <div class="w-16 h-16 bg-white/5 rounded-xl flex items-center justify-center relative overflow-hidden group-hover:bg-blue-600/20 transition-all">
                                <Icon name={item.type === 'video' ? 'film' : 'music'} class="w-6 h-6 text-gray-600 group-hover:text-blue-400" />
                            </div>
                            <div>
                                <div class="font-bold text-white group-hover:text-blue-400 transition-colors">{item.title}</div>
                                <div class="text-xs text-gray-500 font-medium">{item.artist} • {item.size} • {item.date}</div>
                            </div>
                        </div>
                        <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all">
                            <button onClick={() => openPlayer(item)} class="p-3 bg-blue-600 rounded-xl text-white shadow-lg shadow-blue-600/20 hover:scale-105 active:scale-95 transition-all">
                                <Icon name="play" class="w-5 h-5 fill-white" />
                            </button>
                            <button class="p-3 bg-white/5 rounded-xl text-gray-400 hover:text-white transition-all">
                                <Icon name="external-link" class="w-5 h-5" />
                            </button>
                        </div>
                    </div>
                )}
            </For>
        </div>
    </div>
  );
}