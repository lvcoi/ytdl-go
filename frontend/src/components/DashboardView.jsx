import { createMemo, lazy, Suspense, createSignal, For, Show, onMount, createEffect } from 'solid-js';
import ActiveDownloads from './ActiveDownloads';
import { Grid, GridItem } from './Grid';
import WelcomeWidget from './dashboard/WelcomeWidget';
import StatsWidget from './dashboard/StatsWidget';
import RecentActivityWidget from './dashboard/RecentActivityWidget';
import Icon from './Icon';

const QuickDownload = lazy(() => import('./QuickDownload'));

const DASHBOARD_LAYOUT_KEY = 'ytdl-go:dashboard-layout:v1';

const DEFAULT_WIDGETS = [
    { id: 'welcome', span: 4, enabled: true },
    { id: 'quick-download', span: 3, enabled: true },
    { id: 'active-downloads', span: 1, enabled: true },
    { id: 'recent-activity', span: 3, enabled: true },
    { id: 'stats', span: 1, enabled: true },
];

export default function DashboardView(props) {
    const [isEditMode, setIsEditMode] = createSignal(false);
    const [widgets, setWidgets] = createSignal([...DEFAULT_WIDGETS]);

    onMount(() => {
        if (typeof localStorage === 'undefined' || typeof localStorage.getItem !== 'function') return;
        const saved = localStorage.getItem(DASHBOARD_LAYOUT_KEY);
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                if (Array.isArray(parsed)) {
                    setWidgets(parsed);
                }
            } catch (e) {
                console.warn('Failed to load dashboard layout:', e);
            }
        }
    });

    createEffect(() => {
        if (typeof localStorage !== 'undefined' && typeof localStorage.setItem === 'function') {
            localStorage.setItem(DASHBOARD_LAYOUT_KEY, JSON.stringify(widgets()));
        }
    });

    const libraryModel = createMemo(() => (typeof props.libraryModel === 'function' ? props.libraryModel() : props.libraryModel));

    const stats = createMemo(() => {
        const model = libraryModel();
        return {
            totalItems: model?.items?.length || 0,
            totalCreators: (model?.artists?.length || 0) + (model?.videos?.length || 0) + (model?.podcasts?.length || 0),
            recentItems: model?.items?.slice(0, 4) || [], // Latest 4 items
        };
    });

    const handleQuickDownload = (url) => {
        if (props.onDownload) {
            props.onDownload(url);
        }
    };

    const toggleWidget = (id) => {
        setWidgets(prev => prev.map(w => w.id === id ? { ...w, enabled: !w.enabled } : w));
    };

    const renderWidget = (widget) => {
        switch (widget.id) {
            case 'welcome':
                return <WelcomeWidget stats={stats()} onTabChange={props.onTabChange} />;
            case 'quick-download':
                return (
                    <Suspense fallback={<div class="rounded-[2rem] border border-dashed border-white/10 bg-black/20 p-6 h-[148px] animate-pulse" />}>
                        <QuickDownload onDownload={handleQuickDownload} onTabChange={props.onTabChange} />
                    </Suspense>
                );
            case 'active-downloads':
                return <div class="h-full"><ActiveDownloads /></div>;
            case 'recent-activity':
                return <RecentActivityWidget stats={stats()} onTabChange={props.onTabChange} onPlay={props.onPlay} />;
            case 'stats':
                return <StatsWidget stats={stats()} />;
            default:
                return null;
        }
    };

    return (
        <div class="transition-smooth animate-in fade-in slide-in-from-right-4 duration-500 space-y-6">
            <div class="flex items-center justify-between px-2">
                <div class="flex items-center gap-2">
                    <div class={`w-2 h-2 rounded-full ${isEditMode() ? 'bg-amber-500 animate-pulse' : 'bg-accent-primary'}`} />
                    <span class="text-xs font-black uppercase tracking-widest text-gray-500">
                        {isEditMode() ? 'Dashboard: Edit Mode' : 'Dashboard'}
                    </span>
                </div>
                <button
                    onClick={() => setIsEditMode(!isEditMode())}
                    class={`flex items-center gap-2 px-4 py-2 rounded-xl border transition-all duration-300 text-xs font-bold ${isEditMode()
                        ? 'bg-amber-500/20 border-amber-500/30 text-amber-400'
                        : 'bg-white/5 border-white/10 text-gray-400 hover:bg-white/10 hover:text-white'
                        }`}
                >
                    <Icon name={isEditMode() ? 'check' : 'pencil'} class="w-3.5 h-3.5" />
                    {isEditMode() ? 'Done' : 'Edit Layout'}
                </button>
            </div>

            <Grid>
                <For each={widgets()}>
                    {(widget) => (
                        <Show when={widget.enabled || isEditMode()}>
                            <GridItem 
                                span={widget.span} 
                                class={`relative group/widget transition-all duration-500 ${isEditMode() ? 'scale-[0.98] ring-2 ring-dashed ring-white/10 rounded-[2.2rem] p-1' : ''}`}
                            >
                                <div class={`${isEditMode() && !widget.enabled ? 'opacity-30' : ''} h-full`}>
                                    {renderWidget(widget)}
                                </div>
                            </GridItem>
                        </Show>
                    )}
                </For>
            </Grid>
        </div>
    );
}

