import { createSignal, createMemo, For, Show } from 'solid-js';
import { useAppStore } from '../store/appStore';
import Icon from './Icon';
import WidgetContainer from './widgets/WidgetContainer';
import QuickDownloadWidget from './widgets/QuickDownloadWidget';
import RecentDownloadsWidget from './widgets/RecentDownloadsWidget';
import SystemStatsWidget from './widgets/SystemStatsWidget';
import StorageWidget from './widgets/StorageWidget';

const ROW_HEIGHT = 80;
const GAP = 8;
const COLS = 12;

const WIDGET_CATALOG = [
  { type: 'quick-download', label: 'Quick Download', icon: 'download', description: 'Paste a URL and start downloading', defaultColSpan: 4, defaultRowSpan: 2 },
  { type: 'recent-downloads', label: 'Recent Downloads', icon: 'clock', description: 'See your most recent media', defaultColSpan: 6, defaultRowSpan: 3 },
  { type: 'system-stats', label: 'System Stats', icon: 'bar-chart', description: 'Active and completed downloads', defaultColSpan: 4, defaultRowSpan: 2 },
  { type: 'storage', label: 'Storage', icon: 'hard-drive', description: 'Media library statistics', defaultColSpan: 4, defaultRowSpan: 2 },
];

function renderWidget(type, rowSpan, colSpan) {
  switch (type) {
    case 'quick-download': return <QuickDownloadWidget rowSpan={rowSpan} colSpan={colSpan} />;
    case 'recent-downloads': return <RecentDownloadsWidget rowSpan={rowSpan} colSpan={colSpan} />;
    case 'system-stats': return <SystemStatsWidget rowSpan={rowSpan} colSpan={colSpan} />;
    case 'storage': return <StorageWidget rowSpan={rowSpan} colSpan={colSpan} />;
    default: return <div class="text-gray-500 text-sm">Unknown widget</div>;
  }
}

export default function DashboardView() {
  const { state, setState } = useAppStore();
  const [isEditing, setIsEditing] = createSignal(false);
  const [showDrawer, setShowDrawer] = createSignal(false);
  const [dragging, setDragging] = createSignal(null);
  const [resizing, setResizing] = createSignal(null);
  const [dragPreview, setDragPreview] = createSignal(null);
  const [resizePreview, setResizePreview] = createSignal(null);

  let gridRef;

  const widgets = createMemo(() => state.dashboard?.widgets || []);

  const setWidgets = (updater) => {
    setState('dashboard', 'widgets', updater);
  };

  const getCellDimensions = () => {
    if (!gridRef) return { cellW: 0, cellH: ROW_HEIGHT + GAP };
    const rect = gridRef.getBoundingClientRect();
    const totalGapWidth = GAP * (COLS - 1);
    const cellW = (rect.width - totalGapWidth) / COLS;
    const cellH = ROW_HEIGHT + GAP;
    return { cellW, cellH, rect };
  };

  const getGridCell = (mouseX, mouseY) => {
    const { cellW, cellH, rect } = getCellDimensions();
    if (!rect || cellW <= 0) return { col: 1, row: 1 };
    const relX = mouseX - rect.left;
    const relY = mouseY - rect.top;
    const col = Math.max(1, Math.min(COLS, Math.floor(relX / (cellW + GAP)) + 1));
    const row = Math.max(1, Math.floor(relY / cellH) + 1);
    return { col, row };
  };

  const handleGridMouseMove = (e) => {
    const drag = dragging();
    const resize = resizing();

    if (drag) {
      const { col, row } = getGridCell(e.clientX, e.clientY);
      const widget = widgets().find(w => w.id === drag.widgetId);
      if (!widget) return;
      const newCol = Math.max(1, Math.min(COLS - widget.colSpan + 1, col));
      const newRow = Math.max(1, row);
      setDragPreview({ col: newCol, row: newRow });
    }

    if (resize) {
      const widget = widgets().find(w => w.id === resize.widgetId);
      if (!widget) return;
      const { cellW, cellH, rect } = getCellDimensions();
      if (!rect || cellW <= 0) return;
      const dx = e.clientX - resize.startMouseX;
      const dy = e.clientY - resize.startMouseY;
      const dCols = Math.round(dx / (cellW + GAP));
      const dRows = Math.round(dy / cellH);
      const newColSpan = Math.max(2, Math.min(COLS - widget.col + 1, resize.startColSpan + dCols));
      const newRowSpan = Math.max(1, resize.startRowSpan + dRows);
      setResizePreview({ colSpan: newColSpan, rowSpan: newRowSpan });
    }
  };

  const handleGridMouseUp = () => {
    const drag = dragging();
    const resize = resizing();

    if (drag) {
      const preview = dragPreview();
      if (preview) {
        setWidgets((prev) =>
          prev.map(w => w.id === drag.widgetId ? { ...w, col: preview.col, row: preview.row } : w)
        );
      }
      setDragging(null);
      setDragPreview(null);
    }

    if (resize) {
      const preview = resizePreview();
      if (preview) {
        setWidgets((prev) =>
          prev.map(w => w.id === resize.widgetId ? { ...w, colSpan: preview.colSpan, rowSpan: preview.rowSpan } : w)
        );
      }
      setResizing(null);
      setResizePreview(null);
    }
  };

  const handleDragStart = (e, widgetId) => {
    const widget = widgets().find(w => w.id === widgetId);
    if (!widget) return;
    setDragging({
      widgetId,
      startMouseX: e.clientX,
      startMouseY: e.clientY,
      startCol: widget.col,
      startRow: widget.row,
    });
    setDragPreview({ col: widget.col, row: widget.row });
  };

  const handleResizeStart = (e, widgetId) => {
    const widget = widgets().find(w => w.id === widgetId);
    if (!widget) return;
    setResizing({
      widgetId,
      startMouseX: e.clientX,
      startMouseY: e.clientY,
      startColSpan: widget.colSpan,
      startRowSpan: widget.rowSpan,
    });
    setResizePreview({ colSpan: widget.colSpan, rowSpan: widget.rowSpan });
  };

  const addWidget = (catalogItem) => {
    const id = `${catalogItem.type}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const newWidget = {
      id,
      type: catalogItem.type,
      col: 1,
      row: 1,
      colSpan: catalogItem.defaultColSpan,
      rowSpan: catalogItem.defaultRowSpan,
    };
    setWidgets((prev) => [...prev, newWidget]);
    setShowDrawer(false);
  };

  const getEffectiveWidget = (widget) => {
    const drag = dragging();
    const resize = resizing();
    let result = { ...widget };

    if (drag && drag.widgetId === widget.id) {
      const preview = dragPreview();
      if (preview) {
        result = { ...result, col: preview.col, row: preview.row };
      }
    }
    if (resize && resize.widgetId === widget.id) {
      const preview = resizePreview();
      if (preview) {
        result = { ...result, colSpan: preview.colSpan, rowSpan: preview.rowSpan };
      }
    }
    return result;
  };

  return (
    <div class="relative w-full min-h-full">
      {/* Toolbar */}
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <h1 class="text-xl font-bold text-white">Dashboard</h1>
          <Show when={isEditing()}>
            <span class="px-2 py-0.5 bg-orange-500/15 text-orange-400 text-xs font-bold rounded-full border border-orange-500/20">
              Editing
            </span>
          </Show>
        </div>
        <div class="flex items-center gap-2">
          <Show when={isEditing()}>
            <button
              onClick={() => setShowDrawer(v => !v)}
              class="flex items-center gap-2 px-3 py-2 bg-blue-600/15 hover:bg-blue-600/25 text-blue-400 text-sm font-semibold rounded-xl border border-blue-500/20 transition-all"
            >
              <Icon name="plus" class="w-4 h-4" />
              Add Widget
            </button>
          </Show>
          <button
            onClick={() => {
              const wasEditing = isEditing();
              setIsEditing(v => !v);
              if (wasEditing) {
                setShowDrawer(false);
                setDragging(null);
                setResizing(null);
                setDragPreview(null);
                setResizePreview(null);
              }
            }}
            class={`flex items-center gap-2 px-3 py-2 text-sm font-semibold rounded-xl border transition-all ${
              isEditing()
                ? 'bg-green-600/15 hover:bg-green-600/25 text-green-400 border-green-500/20'
                : 'bg-white/5 hover:bg-white/10 text-gray-400 border-white/10'
            }`}
          >
            <Icon name={isEditing() ? 'check' : 'edit-2'} class="w-4 h-4" />
            {isEditing() ? 'Done' : 'Edit'}
          </button>
        </div>
      </div>

      {/* Grid area */}
      <div class="relative">
        {/* Edit mode grid overlay */}
        <Show when={isEditing()}>
          <div
            class="absolute inset-0 pointer-events-none z-0 rounded-xl overflow-hidden"
            style={{
              'background-image': 'linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px)',
              'background-size': `calc((100% + ${GAP}px) / ${COLS}) ${ROW_HEIGHT + GAP}px`,
              'background-position': '0 0',
            }}
          />
        </Show>

        {/* Main grid */}
        <div
          ref={gridRef}
          class="relative"
          style={{
            display: 'grid',
            'grid-template-columns': `repeat(${COLS}, 1fr)`,
            'grid-auto-rows': `${ROW_HEIGHT}px`,
            gap: `${GAP}px`,
            'min-height': `${ROW_HEIGHT * 7 + GAP * 6}px`,
            cursor: dragging() ? 'grabbing' : 'default',
          }}
          onMouseMove={handleGridMouseMove}
          onMouseUp={handleGridMouseUp}
          onMouseLeave={handleGridMouseUp}
        >
          <For each={widgets()}>
            {(widget) => {
              const effective = () => getEffectiveWidget(widget);
              const isDraggingThis = () => dragging()?.widgetId === widget.id;
              const isResizingThis = () => resizing()?.widgetId === widget.id;

              return (
                <WidgetContainer
                  widget={effective()}
                  isEditing={isEditing()}
                  isDragging={isDraggingThis()}
                  isResizing={isResizingThis()}
                  onDragStart={handleDragStart}
                  onResizeStart={handleResizeStart}
                >
                  {renderWidget(widget.type, effective().rowSpan, effective().colSpan)}
                </WidgetContainer>
              );
            }}
          </For>
        </div>
      </div>

      {/* Widget Drawer */}
      <Show when={showDrawer() && isEditing()}>
        <div
          class="fixed inset-y-0 right-0 w-72 bg-[#0a0c14] border-l border-white/10 z-50 flex flex-col shadow-2xl"
          style={{ top: 0, bottom: 0 }}
        >
          <div class="flex items-center justify-between px-5 py-4 border-b border-white/10">
            <span class="font-bold text-white text-sm">Add Widget</span>
            <button
              onClick={() => setShowDrawer(false)}
              class="p-1.5 rounded-lg hover:bg-white/10 text-gray-500 hover:text-gray-300 transition-colors"
            >
              <Icon name="x" class="w-4 h-4" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4 space-y-3">
            <For each={WIDGET_CATALOG}>
              {(item) => (
                <button
                  onClick={() => addWidget(item)}
                  class="w-full text-left p-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl transition-all"
                >
                  <div class="flex items-center gap-3 mb-1">
                    <div class="w-8 h-8 bg-blue-600/20 rounded-lg flex items-center justify-center">
                      <Icon name={item.icon} class="w-4 h-4 text-blue-400" />
                    </div>
                    <span class="text-sm font-semibold text-white">{item.label}</span>
                  </div>
                  <p class="text-xs text-gray-500 ml-11">{item.description}</p>
                </button>
              )}
            </For>
          </div>
        </div>
      </Show>
    </div>
  );
}
