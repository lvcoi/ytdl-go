import { createMemo, For } from 'solid-js';
import Icon from './Icon';
import ActiveDownloads from './ActiveDownloads';
import Thumbnail from './Thumbnail';
import DirectDownload from './DirectDownload';

export default function DashboardView(props) {
    const libraryModel = createMemo(() => (typeof props.libraryModel === 'function' ? props.libraryModel() : props.libraryModel()));

    const stats = createMemo(() => {
        const model = libraryModel();
        return {
            totalItems: model?.items?.length || 0,
            totalCreators: (model?.artists?.length || 0) + (model?.videos?.length || 0) + (model?.podcasts?.length || 0),
            recentItems: model?.items?.slice(0, 4) || [], // Latest 4 items
        };
    });

    return (
        <div class="space-y-10 transition-smooth animate-in fade-in slide-in-from-right-4 duration-500 pb-20">

            {/* Top Section: Welcome & Stats */}
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div class="lg:col-span-2 space-y-6">
                    {/* Welcome Card */}
                    <div class="rounded-[2rem] border border-accent-primary/20 glass-vibrant p-8 relative overflow-hidden group flex flex-col justify-center min-h-[220px]">
                        <div class="absolute top-0 right-0 p-8 opacity-10 group-hover:opacity-20 transition-opacity duration-500">
                            <Icon name="layout-dashboard" class="w-48 h-48 rotate-12" />
                        </div>
                        <div class="relative z-10 space-y-4">
                            <h1 class="text-4xl font-black text-white tracking-tight">
                                Welcome Back!
                            </h1>
                            <p class="text-lg text-gray-300 max-w-lg">
                                Your media library is ready. You have <span class="text-white font-bold">{stats().totalItems} items</span> across <span class="text-white font-bold">{stats().totalCreators} creators</span>.
                            </p>
                        </div>
                    </div>

                    {/* Direct Download row - aligned with card width */}
                    <DirectDownload onDownload={props.onDirectDownload} />

                    {/* Quick Actions row */}
                    <div class="flex flex-wrap gap-3">
                        <button
                            onClick={() => props.onTabChange('download')}
                            class="px-6 py-3 rounded-xl bg-white text-black font-black uppercase tracking-widest hover:scale-105 transition-transform shadow-lg flex items-center gap-2"
                        >
                            <Icon name="plus-circle" class="w-4 h-4" />
                            New Download
                        </button>
                        <button
                            onClick={() => props.onTabChange('library')}
                            class="px-6 py-3 rounded-xl bg-black/40 text-white border border-white/10 font-black uppercase tracking-widest hover:bg-black/60 transition-colors flex items-center gap-2"
                        >
                            <Icon name="layers" class="w-4 h-4" />
                            Browse Library
                        </button>
                    </div>
                </div>

                {/* Storage / Quick Stats (Placeholder for real storage data if available later) */}
                <div class="rounded-[2rem] border border-white/5 bg-black/20 p-6 flex flex-col justify-center gap-4">
                    <h3 class="text-sm font-bold text-gray-400 uppercase tracking-widest">Library Stats</h3>
                    <div class="grid grid-cols-2 gap-4">
                        <div class="p-4 rounded-2xl bg-white/5 border border-white/5">
                            <div class="text-3xl font-black text-accent-primary">{stats().totalCreators}</div>
                            <div class="text-[10px] font-bold text-gray-500 uppercase">Creators</div>
                        </div>
                        <div class="p-4 rounded-2xl bg-white/5 border border-white/5">
                            <div class="text-3xl font-black text-accent-secondary">{stats().totalItems}</div>
                            <div class="text-[10px] font-bold text-gray-500 uppercase">Media Files</div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Recent Activity */}
                <div class="lg:col-span-2 space-y-4">
                    <div class="flex items-center justify-between px-2">
                        <h2 class="text-xl font-bold text-white">Recent Additions</h2>
                        <button onClick={() => props.onTabChange('library')} class="text-xs font-bold text-accent-primary hover:text-white transition-colors">See All</button>
                    </div>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <For each={stats().recentItems}>
                            {(item) => (
                                <div class="flex gap-4 p-4 rounded-2xl border border-white/5 bg-black/20 hover:bg-black/40 transition-colors group">
                                    <Thumbnail src={item.thumbnailUrl} alt={item.title} size="sm" class="w-16 h-16 rounded-lg flex-shrink-0 shadow-lg" />
                                    <div class="min-w-0 flex-1 flex flex-col justify-center">
                                        <h4 class="font-bold text-white truncate text-sm">{item.title}</h4>
                                        <p class="text-xs text-gray-400 truncate">{item.creator}</p>
                                        <div class="mt-2 flex items-center gap-2">
                                            <span class="text-[10px] font-bold px-2 py-0.5 rounded bg-white/10 text-gray-300 uppercase tracking-wider">{item.type}</span>
                                            <span class="text-[10px] text-gray-600">{item.date || 'Just now'}</span>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </For>
                        <Show when={stats().recentItems.length === 0}>
                            <div class="col-span-full py-12 text-center text-gray-500 text-sm font-medium bg-black/20 rounded-2xl border border-dashed border-white/10">
                                No Recent Items
                            </div>
                        </Show>
                    </div>
                </div>

                {/* Active Downloads Widget */}
                <div class="h-full">
                    <ActiveDownloads />
                </div>
            </div>
        </div>
    );
}
