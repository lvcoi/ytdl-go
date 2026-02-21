import { createMemo, lazy, Suspense, createSignal, For, Show, onMount, createEffect, onCleanup, untrack } from 'solid-js';
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
    GRID_ROW_HEIGHT_PX,
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
    let gridRef;
    const [hasLoaded, setHasLoaded] = createSignal(false);
    const [isEditMode, setIsEditMode] = createSignal(false);
    const [isDrawerOpen, setIsDrawerOpen] = createSignal(false);
    const [widgets, setWidgets] = createSignal([...DEFAULT_LAYOUT_WIDGETS]);
    const [layoutState, setLayoutState] = createSignal({
        version: 3,
        activeLayoutId: 'default',
        layouts: {
            'default': { id: 'default', name: 'Default', widgets: [...DEFAULT_LAYOUT_WIDGETS], isFactory: true }
        }
    });
    const [ghostPos, setGhostPos] = createSignal(null);
    const [cellSize, setCellSize] = createSignal({ width: 0, height: GRID_ROW_HEIGHT_PX + GRID_GAP_PX });
    
    // Undo/Redo Stacks
    const [undoStack, setUndoStack] = createSignal([]);
    const [redoStack, setRedoStack] = createSignal([]);

        const pushUndo = (currentWidgets) => {
        const stack = undoStack();
        if (stack.length >= 50) stack.shift();
        setUndoStack([...stack, JSON.parse(JSON.stringify(currentWidgets))]);
        setRedoStack([]);
    };

    const undo = () => {
        const stack = undoStack();
        if (stack.length === 0) return;
        
        const previous = stack[stack.length - 1];
        const newStack = stack.slice(0, -1);
        
        setRedoStack([...redoStack(), JSON.parse(JSON.stringify(widgets()))]);
        setUndoStack(newStack);
        setWidgets(previous);
    };

    const redo = () => {
        const stack = redoStack();
        if (stack.length === 0) return;
        
        const next = stack[stack.length - 1];
        const newStack = stack.slice(0, -1);
        
        setUndoStack([...undoStack(), JSON.parse(JSON.stringify(widgets()))]);
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

    const [dragState, setDragState] = createSignal({
        isDragging: false,
        widgetId: null,
        startX: 0,
        startY: 0,
        originalX: 0,
        originalY: 0
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

    // Measure grid on resize
    onMount(() => {
        if (!gridRef) return;
        const observer = new ResizeObserver(entries => {
            for (const entry of entries) {
                const width = entry.contentRect.width;
                const colWidth = (width + GRID_GAP_PX) / GRID_COLS;
                setCellSize(prev => ({ ...prev, width: colWidth }));
            }
        });
        observer.observe(gridRef);
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
                    setWidgets(loadedWidgets);
                    setTimeout(() => setHasLoaded(true), 0);
                    return;
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
        
        setTimeout(() => setHasLoaded(true), 0);
    });

    // Auto-save layout changes
    createEffect(() => {
        const w = widgets();
        if (!untrack(hasLoaded)) return;
        if (typeof localStorage !== 'undefined' && typeof localStorage.setItem === 'function') {
            localStorage.setItem(DASHBOARD_LAYOUT_KEY_V3, JSON.stringify(w));
        }
    });

                // Layout Management Functions
    const handleSaveLayout = (name) => {
        const id = crypto.randomUUID();
        const newLayout = { id, name, widgets: JSON.parse(JSON.stringify(widgets())), isFactory: false };
        
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
            setWidgets(JSON.parse(JSON.stringify(layout.widgets)));
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

    const handleAddWidget = (id) => {
        pushUndo(widgets());
        
        const template = WIDGET_REGISTRY[id];
        const existing = widgets().find(w => w.id === id);
        
        if (existing) {
            // Re-enable existing
            setWidgets(prev => prev.map(w => w.id === id ? { ...w, enabled: true } : w));
        } else {
            // Find spot for new
            const pos = findOpenPosition(widgets(), template.defaultW, template.defaultH, GRID_COLS);
            setWidgets(prev => [...prev, {
                id,
                enabled: true,
                x: pos.x,
                y: pos.y,
                width: template.defaultW,
                height: template.defaultH
            }]);
        }
    };

    const handleRemoveWidget = (id) => {
        pushUndo(widgets());
        setWidgets(prev => prev.map(w => w.id === id ? { ...w, enabled: false } : w));
    };

    const toggleWidget = (id) => {
        pushUndo(widgets());
        setWidgets(prev => prev.map(w => w.id === id ? { ...w, enabled: !w.enabled } : w));
    };

    const resetLayout = () => {
        setWidgets([...DEFAULT_LAYOUT_WIDGETS]);
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

        // Use measured cell size from registry constants for now (Task 6 will add precise measurement)
        const cellWidthPx = window.innerWidth / GRID_COLS; // Approx
        const cellHeightPx = 80 + 12; // Row + Gap

        const deltaX = Math.round((e.clientX - drag.startX) / cellWidthPx);
        const deltaY = Math.round((e.clientY - drag.startY) / cellHeightPx);

        const widget = widgets().find(w => w.id === drag.widgetId);
        const widgetWidth = widget?.width || 1;
        const newX = Math.max(0, Math.min(GRID_COLS - widgetWidth, drag.originalX + deltaX));
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

        const cellWidthPx = window.innerWidth / GRID_COLS; // Approx
        const cellHeightPx = 80 + 12; // Row + Gap

        const deltaX = Math.round((e.clientX - resize.startX) / cellWidthPx);
        const deltaY = Math.round((e.clientY - resize.startY) / cellHeightPx);

        let newWidth = resize.originalWidth;
        let newHeight = resize.originalHeight;
        let newX = resize.originalX;
        let newY = resize.originalY;

        // Handle different resize directions
        if (resize.direction.includes('e')) {
            newWidth = Math.max(1, Math.min(GRID_COLS - resize.originalX, resize.originalWidth + deltaX));
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
                    y: resize.currentY
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
                        >
                            <Icon name="rotate-ccw" class="w-3.5 h-3.5" />
                        </button>
                                                                        <button
                            onClick={redo}
                            disabled={redoStack().length === 0}
                            class="p-2 rounded-xl border border-white/10 bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                            title="Redo (Ctrl+Shift+Z)"
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
                </div>
            </div>

                        <Grid 
                isEditMode={isEditMode()} 
                totalRows={totalRows()} 
                ghost={ghostPos()} 
                ref={(el) => gridRef = el}
            >
                <For each={widgets()}>
                    {(widget) => (
                        <Show when={widget.enabled || isEditMode()}>
                                                        <GridItem 
                                x={widget.x}
                                y={widget.y}
                                width={widget.width}
                                height={widget.height}
                                widgetId={widget.id}
                                isEditMode={isEditMode()}
                                onResizeStart={handleResizeStart}
                                // Hide dragged widget (it's represented by ghost)
                                class={`relative group/widget transition-all duration-500 ${isEditMode() ? 'scale-[0.98] ring-2 ring-dashed ring-white/10 rounded-[2.2rem] p-1' : ''} ${dragState().isDragging && dragState().widgetId === widget.id ? 'opacity-0' : ''} ${resizeState().isResizing && resizeState().widgetId === widget.id ? 'opacity-0' : ''}`}
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


