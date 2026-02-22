import { createSignal, onCleanup } from 'solid-js';
import { resolveCollisions, compactLayout } from '../components/dashboard/gridCollision';
import { WIDGET_REGISTRY, GRID_COLS } from '../components/dashboard/widgetRegistry';

export function useDashboardDnD(props) {
    const {
        isEditMode,
        widgets,
        setWidgets,
        cellSize,
        gridRef,
        onPushUndo
    } = props;

    const [ghostPos, setGhostPos] = createSignal(null);
    let preInteractionWidgets = null; // snapshot

    const [dragState, setDragState] = createSignal({
        isDragging: false,
        widgetId: null,
        startX: 0,
        startY: 0,
        gridOffsetX: 0,
        gridOffsetY: 0,
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

    const handleDragStart = (widgetId, e) => {
        if (!isEditMode()) return;

        const widget = widgets().find(w => w.id === widgetId);
        if (!widget) return;

        onPushUndo?.(widgets());
        preInteractionWidgets = structuredClone(widgets());

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = 'none';
        document.body.style.cursor = 'grabbing';

        setGhostPos({ x: widget.x, y: widget.y, width: widget.width, height: widget.height });

        // Support both direct element and ref object
        const el = gridRef?.current || gridRef;
        const gridRect = el?.getBoundingClientRect();
        
        setDragState({
            isDragging: true,
            widgetId,
            startX: e.clientX,
            startY: e.clientY,
            gridOffsetX: gridRect?.left || 0,
            gridOffsetY: gridRect?.top || 0,
            originalX: widget.x,
            originalY: widget.y,
            currentX: widget.x,
            currentY: widget.y
        });

        e.preventDefault();
    };

    const handleDragMove = (e) => {
        const drag = dragState();
        if (!drag.isDragging || !preInteractionWidgets) return;

        const cell = cellSize();
        if (!cell.width) return;

        const mouseGridX = e.clientX - drag.gridOffsetX;
        const mouseGridY = e.clientY - drag.gridOffsetY;
        const startGridX = drag.startX - drag.gridOffsetX;
        const startGridY = drag.startY - drag.gridOffsetY;

        const deltaX = Math.round((mouseGridX - startGridX) / cell.width);
        const deltaY = Math.round((mouseGridY - startGridY) / cell.height);

        const snapshot = preInteractionWidgets;
        const widget = snapshot.find(w => w.id === drag.widgetId);
        if (!widget) return;

        const newX = Math.max(0, Math.min(GRID_COLS - widget.width, drag.originalX + deltaX));
        const newY = Math.max(0, drag.originalY + deltaY);

        setGhostPos({ x: newX, y: newY, width: widget.width, height: widget.height });

        const updated = { ...widget, x: newX, y: newY };
        const withMove = snapshot.map(w => w.id === drag.widgetId ? updated : w);
        const resolved = resolveCollisions(withMove, updated);
        setWidgets(resolved);

        setDragState(prev => ({
            ...prev,
            currentX: newX,
            currentY: newY
        }));
    };

    const handleDragEnd = () => {
        const drag = dragState();
        if (!drag.isDragging) return;

        setWidgets(prev => compactLayout(prev));

        preInteractionWidgets = null;
        setGhostPos(null);
        setDragState({
            isDragging: false,
            widgetId: null,
            startX: 0,
            startY: 0,
            gridOffsetX: 0,
            gridOffsetY: 0,
            originalX: 0,
            originalY: 0,
            currentX: 0,
            currentY: 0
        });

        if (!resizeState().isResizing) {
            cleanupListeners();
        }
    };

    const handleResizeStart = (widgetId, direction, e) => {
        if (!isEditMode()) return;

        const widget = widgets().find(w => w.id === widgetId);
        if (!widget) return;

        onPushUndo?.(widgets());
        preInteractionWidgets = structuredClone(widgets());

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = 'none';
        document.body.style.cursor = getResizeCssCursor(direction);

        setGhostPos({ x: widget.x, y: widget.y, width: widget.width, height: widget.height });

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
        if (!resize.isResizing || !preInteractionWidgets) return;

        const cell = cellSize();
        if (!cell.width) return;

        const deltaX = Math.round((e.clientX - resize.startX) / cell.width);
        const deltaY = Math.round((e.clientY - resize.startY) / cell.height);

        const reg = WIDGET_REGISTRY[resize.widgetId];
        const minW = reg?.minW || 1;
        const minH = reg?.minH || 1;

        let newWidth = resize.originalWidth;
        let newHeight = resize.originalHeight;
        let newX = resize.originalX;
        let newY = resize.originalY;

        if (resize.direction.includes('e')) {
            newWidth = Math.max(minW, Math.min(GRID_COLS - resize.originalX, resize.originalWidth + deltaX));
        }
        if (resize.direction.includes('w')) {
            newWidth = Math.max(minW, Math.min(resize.originalX + resize.originalWidth, resize.originalWidth - deltaX));
            newX = resize.originalX + resize.originalWidth - newWidth;
        }
        if (resize.direction.includes('s')) {
            newHeight = Math.max(minH, resize.originalHeight + deltaY);
        }
        if (resize.direction.includes('n')) {
            newHeight = Math.max(minH, resize.originalHeight - deltaY);
            newY = resize.originalY + resize.originalHeight - newHeight;
        }

        newX = Math.max(0, newX);
        newY = Math.max(0, newY);

        setGhostPos({ x: newX, y: newY, width: newWidth, height: newHeight });

        const snapshot = preInteractionWidgets;
        const widget = snapshot.find(w => w.id === resize.widgetId);
        if (widget) {
            const updated = { ...widget, x: newX, y: newY, width: newWidth, height: newHeight };
            const withResize = snapshot.map(w => w.id === resize.widgetId ? updated : w);
            const resolved = resolveCollisions(withResize, updated);
            setWidgets(resolved);
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

        setWidgets(prev => compactLayout(prev));

        preInteractionWidgets = null;
        setGhostPos(null);
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
            cleanupListeners();
        }
    };

    const handleMouseMove = (e) => {
        handleDragMove(e);
        handleResizeMove(e);
    };

    const handleMouseUp = () => {
        handleDragEnd();
        handleResizeEnd();
    };

    const cleanupListeners = () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
        document.body.style.userSelect = '';
        document.body.style.cursor = '';
    };

    onCleanup(cleanupListeners);

    return {
        dragState,
        resizeState,
        ghostPos,
        handleDragStart,
        handleResizeStart
    };
}
