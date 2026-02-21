import { For, Show } from 'solid-js';
import Thumbnail from '../Thumbnail';

export default function RecentActivityWidget(props) {
    return (
        <div class="space-y-4">
            <div class="flex items-center justify-between px-2">
                <h2 class="text-xl font-bold text-white">Recent Additions</h2>
                <button onClick={() => props.onTabChange('library')} class="text-xs font-bold text-accent-primary hover:text-white transition-colors">See All</button>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <For each={props.stats?.recentItems || []}>
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
                <Show when={!props.stats?.recentItems?.length}>
                    <div class="col-span-full py-12 text-center text-gray-500 text-sm font-medium bg-black/20 rounded-2xl border border-dashed border-white/10">
                        No Recent Items
                    </div>
                </Show>
            </div>
        </div>
    );
}
