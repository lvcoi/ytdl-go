import { createSignal, createEffect, onMount, onCleanup, Show } from 'solid-js';

// Base layout constants
const GRID_COLS = 16;
const ROW_HEIGHT = 80; // px
const GAP = 12; // px

export function Grid(props) {
    let gridRef;

    return (
        <div 
            ref={(el) => {
                gridRef = el;
                if (props.ref) props.ref(el);
            }}
            class="relative w-full"
            style={{
                display: 'grid',
                'grid-template-columns': `repeat(${GRID_COLS}, 1fr)`,
                'grid-auto-rows': `${ROW_HEIGHT}px`,
                gap: `${GAP}px`,
                'min-height': props.totalRows ? `${props.totalRows * (ROW_HEIGHT + GAP)}px` : 'auto'
            }}
        >
                        {/* Grid Lines Overlay */}
            <Show when={props.isEditMode}>
                <div 
                    class="dashboard-grid-lines absolute inset-0 pointer-events-none z-0 rounded-3xl"
                    style={{
                        'background-size': `calc((100% + ${GAP}px) / ${GRID_COLS}) ${ROW_HEIGHT + GAP}px`,
                        'background-position': `-${GAP/2}px -${GAP/2}px` // Offset for gap centering
                    }}
                />
            </Show>
            
            {/* Ghost Preview */}
            <Show when={props.ghost}>
                <div 
                    class="absolute z-0 bg-accent-primary/20 border-2 border-accent-primary/50 rounded-[2rem] transition-all duration-200"
                    style={{
                        'grid-column': `${props.ghost.x + 1} / span ${props.ghost.width}`,
                        'grid-row': `${props.ghost.y + 1} / span ${props.ghost.height}`
                    }}
                />
            </Show>
            
            {/* Grid Content */}
            {props.children}
        </div>
    );
}

export function GridItem(props) {
    const handleResizeStart = (direction, e) => {
        e.preventDefault();
        e.stopPropagation();
        if (props.onResizeStart) {
            props.onResizeStart(props.widgetId, direction, e);
        }
    };

    return (
        <div
            class={`${props.class || ''} relative z-10`}
            style={{
                'grid-column': `${(props.x || 0) + 1} / span ${props.width || props.span || 1}`,
                'grid-row': `${(props.y || 0) + 1} / span ${props.height || 1}`,
                transition: 'grid-column 0.25s ease, grid-row 0.25s ease, grid-column-end 0.25s ease, grid-row-end 0.25s ease, transform 0.1s ease',
                ...(props.style || {})
            }}
            onMouseDown={props.onMouseDown}
        >
            {props.children}
            
            {/* Resize Handles - Only visible in edit mode */}
            <Show when={props.isEditMode}>
                {/* Corners */}
                <div class="absolute -top-1.5 -left-1.5 w-4 h-4 cursor-nwse-resize z-20 hover:bg-white/20 rounded-full" onMouseDown={(e) => handleResizeStart('nw', e)} />
                <div class="absolute -top-1.5 -right-1.5 w-4 h-4 cursor-nesw-resize z-20 hover:bg-white/20 rounded-full" onMouseDown={(e) => handleResizeStart('ne', e)} />
                <div class="absolute -bottom-1.5 -left-1.5 w-4 h-4 cursor-nesw-resize z-20 hover:bg-white/20 rounded-full" onMouseDown={(e) => handleResizeStart('sw', e)} />
                <div class="absolute -bottom-1.5 -right-1.5 w-4 h-4 cursor-nwse-resize z-20 hover:bg-white/20 rounded-full" onMouseDown={(e) => handleResizeStart('se', e)} />
                
                {/* Edges */}
                <div class="absolute top-0 left-2 right-2 h-1.5 cursor-ns-resize z-20 hover:bg-white/20" onMouseDown={(e) => handleResizeStart('n', e)} />
                <div class="absolute bottom-0 left-2 right-2 h-1.5 cursor-ns-resize z-20 hover:bg-white/20" onMouseDown={(e) => handleResizeStart('s', e)} />
                <div class="absolute left-0 top-2 bottom-2 w-1.5 cursor-ew-resize z-20 hover:bg-white/20" onMouseDown={(e) => handleResizeStart('w', e)} />
                <div class="absolute right-0 top-2 bottom-2 w-1.5 cursor-ew-resize z-20 hover:bg-white/20" onMouseDown={(e) => handleResizeStart('e', e)} />
            </Show>
        </div>
    );
}

