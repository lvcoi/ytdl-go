import { createMemo, lazy, Suspense, createSignal, For, Show, onMount, createEffect, onCleanup } from 'solid-js';
import ActiveDownloads from './ActiveDownloads';
import { Grid, GridItem } from './Grid';
import WelcomeWidget from './dashboard/WelcomeWidget';
import StatsWidget from './dashboard/StatsWidget';
import RecentActivityWidget from './dashboard/RecentActivityWidget';
import Icon from './Icon';

const QuickDownload = lazy(() => import('./QuickDownload'));

const DASHBOARD_LAYOUT_KEY = 'ytdl-go:dashboard-layout:v2';
const DASHBOARD_LAYOUT_LEGACY_KEY = 'ytdl-go:dashboard-layout:v1';

const DEFAULT_WIDGETS = [
    { id: 'welcome', span: 4, enabled: true, x: 0, y: 0, width: 4, height: 2 },
    { id: 'quick-download', span: 3, enabled: true, x: 0, y: 2, width: 3, height: 2 },
    { id: 'active-downloads', span: 1, enabled: true, x: 3, y: 2, width: 1, height: 2 },
    { id: 'recent-activity', span: 3, enabled: true, x: 0, y: 4, width: 3, height: 2 },
    { id: 'stats', span: 1, enabled: true, x: 3, y: 4, width: 1, height: 2 },
];

export default function DashboardView(props) {
    let loaded = false;
    const [isEditMode, setIsEditMode] = createSignal(false);
    const [widgets, setWidgets] = createSignal([...DEFAULT_WIDGETS]);
    const [dragState, setDragState] = createSignal({
        isDragging: false,
        widgetId: null,
        startX: 0,
        startY: 0,
        originalX: 0,
        originalY: 0,
        currentX: 0,
        currentY: 0
    });
    const [resizeState, setResizeState] = createSignal({
        isResizing: false,
        widgetId: null,
        direction: null,
        startX: 0,
        startY: 0,
        originalWidth: 0,
        originalHeight: 0,
        originalX: 0,
        originalY: 0,
        currentWidth: 0,
        currentHeight: 0,
        currentX: 0,
        currentY: 0
    });

    // Helper function to migrate legacy positions with span awareness
    const migratePositions = (legacyWidgets) => {
        const cols = 4;
        let cursorX = 0;
        let cursorY = 0;
        return legacyWidgets.map((widget) => {
            const span = widget.span || 1;
            if (cursorX + span > cols) {
                cursorX = 0;
                cursorY += 2;
            }
            const result = {
                ...widget,
                x: cursorX,
                y: cursorY,
                width: span,
                height: 2
            };
            cursorX += span;
            if (cursorX >= cols) {
                cursorX = 0;
                cursorY += 2;
            }
            return result;
        });
    };

    // Load layout from localStorage with migration support
    onMount(() => {
        if (typeof localStorage === 'undefined' || typeof localStorage.getItem !== 'function') return;
        
        // Try new format first
        let saved = localStorage.getItem(DASHBOARD_LAYOUT_KEY);
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                if (Array.isArray(parsed)) {
                    setWidgets(parsed);
                    queueMicrotask(() => { loaded = true; });
                    return;
                }
            } catch (e) {
                console.warn('Failed to load new dashboard layout:', e);
            }
        }
        
        // Fallback to legacy format and migrate
        saved = localStorage.getItem(DASHBOARD_LAYOUT_LEGACY_KEY);
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                if (Array.isArray(parsed)) {
                    // Migrate legacy format to new format
                    const migrated = migratePositions(parsed);
                    setWidgets(migrated);
                    // Save in new format
                    localStorage.setItem(DASHBOARD_LAYOUT_KEY, JSON.stringify(migrated));
                }
            } catch (e) {
                console.warn('Failed to load legacy dashboard layout:', e);
            }
        }
        
        queueMicrotask(() => { loaded = true; });
    });

    // Auto-save layout changes
    createEffect(() => {
        const w = widgets();
        if (!loaded) return;
        if (typeof localStorage !== 'undefined' && typeof localStorage.setItem === 'function') {
            localStorage.setItem(DASHBOARD_LAYOUT_KEY, JSON.stringify(w));
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

    const resetLayout = () => {
        setWidgets([...DEFAULT_WIDGETS]);
    };

    // Drag functionality
    const handleDragStart = (widgetId, e) => {
        if (!isEditMode()) return;
        
        const widget = widgets().find(w => w.id === widgetId);
        if (!widget) return;

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = 'none';
        document.body.style.cursor = 'grabbing';

        setDragState({
            isDragging: true,
            widgetId,
            startX: e.clientX,
            startY: e.clientY,
            originalX: widget.x,
            originalY: widget.y,
            currentX: widget.x,
            currentY: widget.y
        });

        e.preventDefault();
    };

    const handleDragMove = (e) => {
        const drag = dragState();
        if (!drag.isDragging) return;

        const deltaX = Math.round((e.clientX - drag.startX) / 100); // Rough grid cell size
        const deltaY = Math.round((e.clientY - drag.startY) / 100);

        const widget = widgets().find(w => w.id === drag.widgetId);
        const widgetWidth = widget?.width || 1;
        const newX = Math.max(0, Math.min(4 - widgetWidth, drag.originalX + deltaX));
        const newY = Math.max(0, drag.originalY + deltaY);

        setDragState(prev => ({
            ...prev,
            currentX: newX,
            currentY: newY
        }));
    };

    const handleDragEnd = () => {
        const drag = dragState();
        if (!drag.isDragging) return;

        setWidgets(prev => prev.map(w => 
            w.id === drag.widgetId 
                ? { ...w, x: drag.currentX, y: drag.currentY }
                : w
        ));

        setDragState({
            isDragging: false,
            widgetId: null,
            startX: 0,
            startY: 0,
            originalX: 0,
            originalY: 0,
            currentX: 0,
            currentY: 0
        });

        if (!resizeState().isResizing) {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.userSelect = '';
            document.body.style.cursor = '';
        }
    };

    // Resize functionality
    const handleResizeStart = (widgetId, direction, e) => {
        if (!isEditMode()) return;
        
        const widget = widgets().find(w => w.id === widgetId);
        if (!widget) return;

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = 'none';
        document.body.style.cursor = getResizeCssCursor(direction);

        setResizeState({
            isResizing: true,
            widgetId,
            direction,
            startX: e.clientX,
            startY: e.clientY,
            originalWidth: widget.width,
            originalHeight: widget.height,
            originalX: widget.x,
            originalY: widget.y,
            currentWidth: widget.width,
            currentHeight: widget.height,
            currentX: widget.x,
            currentY: widget.y
        });

        e.preventDefault();
    };

    const handleResizeMove = (e) => {
        const resize = resizeState();
        if (!resize.isResizing) return;

        const deltaX = Math.round((e.clientX - resize.startX) / 100);
        const deltaY = Math.round((e.clientY - resize.startY) / 100);

        let newWidth = resize.originalWidth;
        let newHeight = resize.originalHeight;
        let newX = resize.originalX;
        let newY = resize.originalY;

        // Handle different resize directions
        if (resize.direction.includes('e')) {
            newWidth = Math.max(1, Math.min(4 - resize.originalX, resize.originalWidth + deltaX));
        }
        if (resize.direction.includes('w')) {
            newWidth = Math.max(1, Math.min(resize.originalX + resize.originalWidth, resize.originalWidth - deltaX));
            newX = resize.originalX + resize.originalWidth - newWidth;
        }
        if (resize.direction.includes('s')) {
            newHeight = Math.max(1, resize.originalHeight + deltaY);
        }
        if (resize.direction.includes('n')) {
            newHeight = Math.max(1, resize.originalHeight - deltaY);
            newY = resize.originalY + resize.originalHeight - newHeight;
        }

        setResizeState(prev => ({
            ...prev,
            currentWidth: newWidth,
            currentHeight: newHeight,
            currentX: newX,
            currentY: newY
        }));
    };

    const handleResizeEnd = () => {
        const resize = resizeState();
        if (!resize.isResizing) return;

        setWidgets(prev => prev.map(w => 
            w.id === resize.widgetId 
                ? { 
                    ...w, 
                    width: resize.currentWidth, 
                    height: resize.currentHeight,
                    x: resize.currentX,
                    y: resize.currentY,
                    span: resize.currentWidth // Update span for compatibility
                }
                : w
        ));

        setResizeState({
            isResizing: false,
            widgetId: null,
            direction: null,
            startX: 0,
            startY: 0,
            originalWidth: 0,
            originalHeight: 0,
            originalX: 0,
            originalY: 0,
            currentWidth: 0,
            currentHeight: 0,
            currentX: 0,
            currentY: 0
        });

        if (!dragState().isDragging) {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.userSelect = '';
            document.body.style.cursor = '';
        }
    };

    // Global mouse event handlers
    const handleMouseMove = (e) => {
        handleDragMove(e);
        handleResizeMove(e);
    };

    const handleMouseUp = () => {
        handleDragEnd();
        handleResizeEnd();
    };

    // Helper to map resize direction to CSS cursor
    const getResizeCssCursor = (direction) => {
        const map = {
            'n': 'ns-resize', 's': 'ns-resize',
            'e': 'ew-resize', 'w': 'ew-resize',
            'ne': 'nesw-resize', 'sw': 'nesw-resize',
            'nw': 'nwse-resize', 'se': 'nwse-resize'
        };
        return map[direction] || 'nwse-resize';
    };

    // Component cleanup
    onCleanup(() => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = '';
        document.body.style.cursor = '';
    });

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
                <div class="flex items-center gap-2">
                    <Show when={isEditMode()}>
                        <button
                            onClick={resetLayout}
                            class="flex items-center gap-2 px-3 py-2 rounded-xl border border-white/10 bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white transition-all text-xs font-bold"
                        >
                            <Icon name="refresh-cw" class="w-3.5 h-3.5" />
                            Reset
                        </button>
                    </Show>
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
            </div>

            <Grid>
                <For each={widgets()}>
                    {(widget) => (
                        <Show when={widget.enabled || isEditMode()}>
                            <GridItem 
                                span={widget.width || widget.span}
                                widgetId={widget.id}
                                isEditMode={isEditMode()}
                                onResizeStart={handleResizeStart}
                                class={`relative group/widget transition-all duration-500 ${isEditMode() ? 'scale-[0.98] ring-2 ring-dashed ring-white/10 rounded-[2.2rem] p-1' : ''} ${dragState().isDragging && dragState().widgetId === widget.id ? 'opacity-50' : ''} ${resizeState().isResizing && resizeState().widgetId === widget.id ? 'ring-2 ring-accent-primary' : ''}`}
                                style={
                                    dragState().isDragging && dragState().widgetId === widget.id 
                                        ? { 
                                            transform: `translate(${(dragState().currentX - dragState().originalX) * 100}px, ${(dragState().currentY - dragState().originalY) * 100}px)`,
                                            'z-index': 1000
                                        }
                                        : resizeState().isResizing && resizeState().widgetId === widget.id
                                        ? {
                                            width: `${resizeState().currentWidth * 25}%`,
                                            height: `${resizeState().currentHeight * 100}px`,
                                            'z-index': 999
                                        }
                                        : {}
                                }
                            >
                                <div 
                                    class={`${isEditMode() && !widget.enabled ? 'opacity-30' : ''} h-full ${isEditMode() ? 'cursor-move' : ''}`}
                                    onMouseDown={(e) => handleDragStart(widget.id, e)}
                                >
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

