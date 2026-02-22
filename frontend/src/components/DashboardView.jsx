import { createMemo, lazy, Suspense, createSignal, For, Show, onMount, createEffect, onCleanup } from 'solid-js';
import { useNavigate, useSearchParams } from '@solidjs/router';
import { useAppStore } from '../store/appStore';
import { useLibraryModel } from '../hooks/useLibraryModel';
import { usePlayerController } from '../hooks/usePlayerController';
import { useDownloadManager } from '../hooks/useDownloadManager';
import { useDashboardDnD } from '../hooks/useDashboardDnD';
import ActiveDownloads from './ActiveDownloads';
import { Grid, GridItem } from './Grid';
import WelcomeWidget from './dashboard/WelcomeWidget';
import StatsWidget from './dashboard/StatsWidget';
import RecentActivityWidget from './dashboard/RecentActivityWidget';
import Icon from './Icon';
import ConcurrencyWidget from './ConcurrencyWidget';
import { 
    WIDGET_REGISTRY, 
    DEFAULT_LAYOUT_WIDGETS, 
    GRID_COLS,
    GRID_GAP_PX
} from './dashboard/widgetRegistry';
import { resolveCollisions, compactLayout, findOpenPosition } from './dashboard/gridCollision';
import WidgetDrawer from './dashboard/WidgetDrawer';
import LayoutPresets from './dashboard/LayoutPresets';

const QuickDownload = lazy(() => import('./QuickDownload'));

const DASHBOARD_LAYOUT_KEY_V3 = 'ytdl-go:dashboard-layout:v3';
const DASHBOARD_LAYOUT_LEGACY_KEY_V2 = 'ytdl-go:dashboard-layout:v2';
const DASHBOARD_LAYOUT_LEGACY_KEY_V1 = 'ytdl-go:dashboard-layout:v1';

export default function DashboardView(props) {
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const { state } = useAppStore();
    const libraryModelHook = useLibraryModel();
    const { openPlayer } = usePlayerController();
    const { startDownload } = useDownloadManager();

    // Derived state for edit mode from URL
    const isEditMode = () => searchParams.edit === 'true';
    const setIsEditMode = (enabled) => {
        setSearchParams({ edit: enabled ? 'true' : undefined });
    };

    const libraryModel = createMemo(() => {
        if (props.libraryModel) {
            return typeof props.libraryModel === 'function' ? props.libraryModel() : props.libraryModel;
        }
        return libraryModelHook();
    });

    const handleTabChange = (tab) => {
        if (props.onTabChange) {
            props.onTabChange(tab);
            return;
        }
        const routes = {
            'dashboard': '/',
            'download': '/download',
            'library': '/library',
            'settings': '/settings'
        };
        if (routes[tab]) navigate(routes[tab]);
    };

    const handleDownload = (url) => {
        if (props.onDownload) {
            props.onDownload(url);
        } else {
            startDownload(url);
        }
    };

    const handlePlay = (item) => {
        if (props.onPlay) {
            props.onPlay(item);
        } else {
            // Default to playing from all downloads if no specific queue provided via props
            openPlayer(item, state.library.downloads);
        }
    };

    const [hasLoaded, setHasLoaded] = createSignal(false);
    const [isDrawerOpen, setIsDrawerOpen] = createSignal(false);
    const [widgets, setWidgets] = createSignal([...DEFAULT_LAYOUT_WIDGETS]);
    const [layoutState, setLayoutState] = createSignal({
        version: 3,
        activeLayoutId: 'default',
        layouts: {
            'default': { id: 'default', name: 'Default', widgets: [...DEFAULT_LAYOUT_WIDGETS], isFactory: true }
        }
    });
    const [cellSize, setCellSize] = createSignal({ width: 0, height: 0, trackWidth: 0 });
    
    // Undo/Redo Stacks
    const [undoStack, setUndoStack] = createSignal([]);
    const [redoStack, setRedoStack] = createSignal([]);

    const pushUndo = (currentWidgets) => {
        const stack = undoStack();
        const trimmed = stack.length >= 50 ? stack.slice(1) : stack;
        setUndoStack([...trimmed, structuredClone(currentWidgets)]);
        setRedoStack([]);
    };

    const undo = () => {
        const stack = undoStack();
        if (stack.length === 0) return;
        
        const previous = stack[stack.length - 1];
        const newStack = stack.slice(0, -1);
        
        setRedoStack([...redoStack(), structuredClone(widgets())]);
        setUndoStack(newStack);
        setWidgets(previous);
    };

    const redo = () => {
        const stack = redoStack();
        if (stack.length === 0) return;
        
        const next = stack[stack.length - 1];
        const newStack = stack.slice(0, -1);
        
        setUndoStack([...undoStack(), structuredClone(widgets())]);
        setRedoStack(newStack);
        setWidgets(next);
    };

    // Keyboard shortcuts
    onMount(() => {
        const handleKeyDown = (e) => {
            if (!isEditMode()) return;
            if ((e.ctrlKey || e.metaKey) && e.key === 'z') {
                e.preventDefault();
                if (e.shiftKey) {
                    redo();
                } else {
                    undo();
                }
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        onCleanup(() => window.removeEventListener('keydown', handleKeyDown));
    });

    // We wrap gridRef in an object so we can pass it by reference to the hook.
    // gridRefContainer.current is populated by the Grid ref callback after mount.
    const gridRefContainer = { current: null };

    // Use Custom Hook for Drag and Drop
    const {
        dragState,
        resizeState,
        ghostPos,
        handleDragStart,
        handleResizeStart
    } = useDashboardDnD({
        isEditMode,
        widgets,
        setWidgets,
        cellSize,
        gridRef: gridRefContainer,
        onPushUndo: pushUndo
    });

    // Measure grid on resize — compute square cell dimensions
    onMount(() => {
        const el = gridRefContainer.current;
        if (!el) return;
        const observer = new ResizeObserver(entries => {
            for (const entry of entries) {
                const containerWidth = entry.contentRect.width;
                // Track width = actual cell width (excluding gaps)
                const trackWidth = (containerWidth - (GRID_COLS - 1) * GRID_GAP_PX) / GRID_COLS;
                // Tile size = track + gap, used for drag/resize snapping
                const tileSize = trackWidth + GRID_GAP_PX;
                setCellSize({ width: tileSize, height: tileSize, trackWidth: Math.round(trackWidth) });
            }
        });
        observer.observe(el);
        onCleanup(() => observer.disconnect());
    });

    // Helper function to migrate legacy positions (v1/v2 4-col -> v3 16-col)
    const migratePositions = (legacyWidgets) => {
        const cols = 4;
        let cursorX = 0;
        let cursorY = 0;
        
        // First pass: resolve 4-col positions if they don't exist
        const resolved4Col = legacyWidgets.map((widget) => {
            // If x/y exist (v2), keep them. If not (v1), compute flow layout.
            let x = widget.x;
            let y = widget.y;
            const span = widget.span || widget.width || 1;
            
            if (x === undefined || y === undefined) {
                if (cursorX + span > cols) {
                    cursorX = 0;
                    cursorY += 2;
                }
                x = cursorX;
                y = cursorY;
                
                cursorX += span;
                if (cursorX >= cols) {
                    cursorX = 0;
                    cursorY += 2;
                }
            }
            
            return { ...widget, x, y, span };
        });

        // Second pass: convert to 16-col
        return resolved4Col.map(widget => ({
            id: widget.id,
            enabled: widget.enabled !== false,
            // 4-col -> 16-col: multiply X and Width by 4
            x: widget.x * 4,
            y: widget.y, // Rows stay same height (roughly) or mapped 1:1 if row height matches
            width: (widget.span || 1) * 4,
            height: widget.height || 2 // Default height from v2 was usually 2
        }));
    };

    // Load layout from localStorage with migration support
    onMount(() => {
        if (typeof localStorage === 'undefined' || typeof localStorage.getItem !== 'function') return;
        
        // Try v3 format first
        let saved = localStorage.getItem(DASHBOARD_LAYOUT_KEY_V3);
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                // Basic validation for v3 structure
                if (Array.isArray(parsed) || (parsed.version === 3 && Array.isArray(parsed.layouts?.default?.widgets))) {
                    const loadedWidgets = Array.isArray(parsed) ? parsed : parsed.layouts[parsed.activeLayoutId || 'default'].widgets;
                    // Validate every widget has required numeric grid fields
                    const isValid = loadedWidgets.every(w =>
                        w.id && typeof w.x === 'number' && typeof w.y === 'number' &&
                        typeof w.width === 'number' && w.width > 0 &&
                        typeof w.height === 'number' && w.height > 0
                    );
                    if (isValid) {
                        setWidgets(compactLayout(loadedWidgets));
                        setHasLoaded(true);
                        return;
                    }
                    console.warn('Invalid v3 layout data — missing grid fields, using defaults');
                }
            } catch (e) {
                console.warn('Failed to load v3 dashboard layout:', e);
            }
        }
        
        // Fallback to legacy formats and migrate
        saved = localStorage.getItem(DASHBOARD_LAYOUT_LEGACY_KEY_V2) || localStorage.getItem(DASHBOARD_LAYOUT_LEGACY_KEY_V1);
        if (saved) {
            try {
                const parsed = JSON.parse(saved);
                if (Array.isArray(parsed)) {
                    // Migrate legacy format to new format
                    const migrated = migratePositions(parsed);
                    setWidgets(migrated);
                    // Save in new format
                    const v3Data = {
                        version: 3,
                        activeLayoutId: 'default',
                        layouts: {
                            'default': { name: 'Default', widgets: migrated, isFactory: false }
                        }
                    };
                    // For now, just save array to keep simple until full preset support
                    localStorage.setItem(DASHBOARD_LAYOUT_KEY_V3, JSON.stringify(migrated));
                }
            } catch (e) {
                console.warn('Failed to load legacy dashboard layout:', e);
            }
        }
        
        setHasLoaded(true);
    });

    // Auto-save layout changes
    let skipNextSave = true;
    createEffect(() => {
        const w = widgets();
        const loaded = hasLoaded();
        if (!loaded) return;
        if (skipNextSave) {
            skipNextSave = false;
            return;
        }
        if (typeof localStorage !== 'undefined' && typeof localStorage.setItem === 'function') {
            localStorage.setItem(DASHBOARD_LAYOUT_KEY_V3, JSON.stringify(w));
        }
    });

    // Layout Management Functions
    const handleSaveLayout = (name) => {
        const id = crypto.randomUUID();
        const newLayout = { id, name, widgets: structuredClone(widgets()), isFactory: false };
        
        setLayoutState(prev => ({
            ...prev,
            activeLayoutId: id,
            layouts: { ...prev.layouts, [id]: newLayout }
        }));
        setWidgets(newLayout.widgets);
    };

    const handleLoadLayout = (id) => {
        const layout = layoutState().layouts[id];
        if (layout) {
            pushUndo(widgets());
            setLayoutState(prev => ({ ...prev, activeLayoutId: id }));
            setWidgets(structuredClone(layout.widgets));
        }
    };

    const handleDeleteLayout = (id) => {
        if (id === 'default') return; // Cannot delete default
        
        const state = layoutState();
        const newLayouts = { ...state.layouts };
        delete newLayouts[id];
        
        let newActiveId = state.activeLayoutId;
        if (state.activeLayoutId === id) {
            newActiveId = 'default';
            setWidgets(newLayouts['default'].widgets);
        }
        
        setLayoutState(prev => ({
            ...prev,
            activeLayoutId: newActiveId,
            layouts: newLayouts
        }));
    };

    const handleSetPrimary = (id) => {
        setLayoutState(prev => {
            const newLayouts = {};
            for (const key in prev.layouts) {
                newLayouts[key] = { ...prev.layouts[key], isPrimary: key === id };
            }
            return { ...prev, layouts: newLayouts };
        });
    };

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

    const handleAddWidget = (id) => {
        pushUndo(widgets());
        
        const template = WIDGET_REGISTRY[id];
        const existing = widgets().find(w => w.id === id);
        
        if (existing) {
            // Re-enable existing and resolve any collisions
            const reEnabled = { ...existing, enabled: true };
            const withEnabled = widgets().map(w => w.id === id ? reEnabled : w);
            const resolved = resolveCollisions(withEnabled, reEnabled);
            setWidgets(compactLayout(resolved));
        } else {
            // Find non-overlapping spot for new widget
            const enabledWidgets = widgets().filter(w => w.enabled);
            const pos = findOpenPosition(enabledWidgets, template.defaultW, template.defaultH, GRID_COLS);
            const newWidget = {
                id,
                enabled: true,
                x: pos.x,
                y: pos.y,
                width: template.defaultW,
                height: template.defaultH
            };
            const withNew = [...widgets(), newWidget];
            const resolved = resolveCollisions(withNew, newWidget);
            setWidgets(compactLayout(resolved));
        }
    };

    const handleRemoveWidget = (id) => {
        pushUndo(widgets());
        setWidgets(prev => prev.map(w => w.id === id ? { ...w, enabled: false } : w));
    };

    const resetLayout = () => {
        pushUndo(widgets());
        setWidgets([...DEFAULT_LAYOUT_WIDGETS]);
    };

    const renderWidget = (widget) => {
        switch (widget.id) {
            case 'welcome':
                return <WelcomeWidget stats={stats()} />;
            case 'quick-download':
                return (
                    <Suspense fallback={<div class="rounded-[2rem] border border-dashed border-white/10 bg-black/20 p-6 h-[148px] animate-pulse" />}>
                        <QuickDownload onDownload={handleQuickDownload} />
                    </Suspense>
                );
            case 'active-downloads':
                return <div class="h-full"><ActiveDownloads /></div>;
            case 'recent-activity':
                return <RecentActivityWidget stats={stats()} onPlay={props.onPlay} />;
            case 'stats':
                return <StatsWidget stats={stats()} />;
            case 'concurrency':
                return <ConcurrencyWidget />;
            default:
                return null;
        }
    };

    // Calculate total rows needed
    const totalRows = createMemo(() => {
        let maxRow = 0;
        widgets().forEach(w => {
            if (w.enabled) {
                maxRow = Math.max(maxRow, w.y + w.height);
            }
        });
        return Math.max(maxRow, 6); // Min 6 rows
    });

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
                        <div class="h-4 w-px bg-white/10 mx-1" />
                        <button
                            onClick={undo}
                            disabled={undoStack().length === 0}
                            class="p-2 rounded-xl border border-white/10 bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                            title="Undo (Ctrl+Z)"
                            aria-label="Undo"
                        >
                            <Icon name="rotate-ccw" class="w-3.5 h-3.5" />
                        </button>
                        <button
                            onClick={redo}
                            disabled={redoStack().length === 0}
                            class="p-2 rounded-xl border border-white/10 bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                            title="Redo (Ctrl+Shift+Z)"
                            aria-label="Redo"
                        >
                            <Icon name="rotate-cw" class="w-3.5 h-3.5" />
                        </button>
                        <div class="h-4 w-px bg-white/10 mx-1" />
                        <LayoutPresets 
                            activeLayoutId={layoutState().activeLayoutId}
                            activeLayoutName={layoutState().layouts[layoutState().activeLayoutId]?.name}
                            layouts={Object.values(layoutState().layouts)}
                            onSave={handleSaveLayout}
                            onLoad={handleLoadLayout}
                            onDelete={handleDeleteLayout}
                            onSetPrimary={handleSetPrimary}
                        />
                        <div class="h-4 w-px bg-white/10 mx-1" />
                        <button
                            onClick={() => setIsDrawerOpen(!isDrawerOpen())}
                            class={`p-2 rounded-xl border transition-all ${isDrawerOpen()
                                ? 'bg-accent-primary/20 border-accent-primary/30 text-accent-primary'
                                : 'bg-white/5 border-white/10 text-gray-400 hover:bg-white/10 hover:text-white'}`}
                            title="Add Widgets"
                            aria-label="Add Widgets"
                            aria-expanded={isDrawerOpen()}
                        >
                            <Icon name="plus" class="w-3.5 h-3.5" />
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
                    <div class="relative group">
                        <button
                            class="p-2 rounded-xl border border-white/10 bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white transition-all"
                            title="Dashboard Settings Info"
                            aria-label="Dashboard settings info"
                        >
                            <Icon name="info" class="w-3.5 h-3.5" />
                        </button>
                        <div class="absolute right-0 top-full mt-2 w-80 p-4 rounded-xl bg-[#0b111a] border border-white/10 shadow-2xl opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
                            <h3 class="text-sm font-bold text-white mb-2">Default Settings</h3>
                            <p class="text-xs text-gray-400 leading-relaxed">
                                Downloads use your default settings. Go to the <span class="text-accent-primary">Download</span> tab for more options.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <Grid 
                isEditMode={isEditMode()} 
                totalRows={totalRows()} 
                ghost={ghostPos()} 
                rowHeight={cellSize().trackWidth}
                ref={(el) => gridRefContainer.current = el}
            >
                <For each={widgets()}>
                    {(widget) => (
                        <Show when={widget.enabled}>
                            <GridItem 
                                x={widget.x}
                                y={widget.y}
                                width={widget.width}
                                height={widget.height}
                                widgetId={widget.id}
                                isEditMode={isEditMode()}
                                onResizeStart={handleResizeStart}
                                onMouseDown={(e) => handleDragStart(widget.id, e)}
                                class={`relative group/widget transition-all duration-500 ${isEditMode() ? 'border-2 border-dashed border-white/20 rounded-[2rem]' : ''} ${dragState().isDragging && dragState().widgetId === widget.id ? 'opacity-0' : ''} ${resizeState().isResizing && resizeState().widgetId === widget.id ? 'opacity-0' : ''}`}
                                style={{ 'z-index': 10 + widget.y * GRID_COLS + widget.x }}
                            >
                                <div 
                                    class={`${isEditMode() && !widget.enabled ? 'opacity-30' : ''} h-full ${isEditMode() ? 'cursor-move' : ''}`}
                                >
                                    {renderWidget(widget)}
                                </div>
                            </GridItem>
                        </Show>
                    )}
                </For>
            </Grid>

            {/* Widget Drawer */}
            <Show when={isEditMode()}>
                <WidgetDrawer 
                    isOpen={isDrawerOpen()} 
                    onClose={() => setIsDrawerOpen(false)}
                    widgets={widgets()}
                    onAdd={handleAddWidget}
                    onRemove={handleRemoveWidget}
                />
            </Show>
        </div>
    );
}
