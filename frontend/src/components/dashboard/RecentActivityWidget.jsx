import { For, Show } from 'solid-js';
import Thumbnail from '../Thumbnail';
import Icon from '../Icon';
import logo from '../../assets/logo.png';

export default function RecentActivityWidget(props) {
    return (
        <div class="space-y-4">
            <div class="flex items-center justify-between px-2">
                <div class="flex items-center gap-2">
                    <h2 class="text-xl font-bold text-white">Recent Additions</h2>
                    <Icon name="history" class="w-4 h-4 text-accent-primary/60" />
                </div>
                <button onClick={() => props.onTabChange('library')} class="text-xs font-bold text-accent-primary hover:text-white transition-colors flex items-center gap-1">
                    See All
                    <Icon name="chevron-right" class="w-3 h-3" />
                </button>
            </div>

            <div class="relative group/carousel">
                <div class="flex gap-4 overflow-x-auto pb-4 pt-2 px-2 custom-scrollbar snap-x scroll-smooth">
                    <For each={props.stats?.recentItems || []}>
                        {(item) => (
                            <div 
                                onClick={() => props.onPlay?.(item)}
                                class="flex-shrink-0 w-72 flex flex-col gap-3 p-4 rounded-2xl border border-white/5 bg-black/20 hover:border-accent-primary/40 hover:bg-black/40 transition-all group cursor-pointer snap-start"
                            >
                                <div class="relative aspect-video rounded-xl overflow-hidden shadow-lg group-hover:scale-[1.02] transition-transform">
                                    <Thumbnail src={item.thumbnailUrl} alt={item.title} size="md" class="w-full h-full" />
                                    <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                                        <div class="w-10 h-10 rounded-full bg-white/20 backdrop-blur-md flex items-center justify-center text-white border border-white/30">
                                            <Icon name="play" class="w-5 h-5 fill-current" />
                                        </div>
                                    </div>
                                    <div class="absolute top-2 right-2 px-2 py-0.5 rounded-full bg-black/60 backdrop-blur-md border border-white/10 text-[9px] font-black uppercase tracking-widest text-white/90">
                                        {item.type}
                                    </div>
                                </div>
                                <div class="min-w-0 space-y-1">
                                    <h4 class="font-bold text-white truncate text-sm leading-snug">{item.title}</h4>
                                    <div class="flex items-center justify-between gap-2">
                                        <p class="text-xs text-accent-primary/80 truncate font-semibold">{item.creator}</p>
                                        <span class="text-[10px] text-gray-600 font-medium whitespace-nowrap">{item.date || 'Just now'}</span>
                                    </div>
                                </div>
                            </div>
                        )}
                    </For>
                    <Show when={!props.stats?.recentItems?.length}>
                        <div class="w-full py-16 flex flex-col items-center justify-center text-center gap-4 bg-black/20 rounded-2xl border border-dashed border-white/10">
                            <img src={logo} alt="Empty Library" class="w-24 h-24 opacity-20 grayscale brightness-50" />
                            <div class="space-y-1">
                                <p class="text-white/40 font-black uppercase tracking-widest text-xs">Nothing here yet!</p>
                                <p class="text-gray-600 font-bold text-[10px] uppercase tracking-tighter italic">Feed me URLs!</p>
                            </div>
                        </div>
                    </Show>
                </div>
            </div>
        </div>
    );
}

