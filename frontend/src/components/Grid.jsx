import { Show } from 'solid-js';

export function Grid(props) {
    return (
        <div class={`relative ${props.class || ''}`}>
            {/* Grid container */}
            <div class={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6`}>
                {props.children}
            </div>
        </div>
    );
}

export function GridItem(props) {
    const spanClass = () => {
        switch(props.span) {
            case 2: return 'lg:col-span-2';
            case 3: return 'lg:col-span-3';
            case 4: return 'lg:col-span-4';
            default: return 'lg:col-span-1';
        }
    };

    const handleResizeStart = (e, direction) => {
        e.preventDefault();
        e.stopPropagation();
        
        if (props.onResizeStart) {
            props.onResizeStart(props.widgetId, direction, e);
        }
    };

    const getResizeCursor = (direction) => {
        const cursors = {
            'n': 'cursor-n-resize',
            's': 'cursor-s-resize',
            'e': 'cursor-e-resize',
            'w': 'cursor-w-resize',
            'ne': 'cursor-ne-resize',
            'nw': 'cursor-nw-resize',
            'se': 'cursor-se-resize',
            'sw': 'cursor-sw-resize'
        };
        return cursors[direction] || '';
    };

    return (
        <div 
            class={`${spanClass()} ${props.class || ''} relative group`}
            data-widget-id={props.widgetId}
        >
            {/* Resize handles - only show in edit mode */}
            <Show when={props.isEditMode}>
                {/* Edge handles */}
                <div 
                    class={`absolute top-0 left-2 right-2 h-2 ${getResizeCursor('n')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-t-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'n')}
                />
                <div 
                    class={`absolute bottom-0 left-2 right-2 h-2 ${getResizeCursor('s')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-b-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 's')}
                />
                <div 
                    class={`absolute top-2 left-0 bottom-2 w-2 ${getResizeCursor('w')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-l-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'w')}
                />
                <div 
                    class={`absolute top-2 right-0 bottom-2 w-2 ${getResizeCursor('e')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-r-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'e')}
                />
                
                {/* Corner handles */}
                <div 
                    class={`absolute top-0 left-0 w-4 h-4 ${getResizeCursor('nw')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-tl-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'nw')}
                />
                <div 
                    class={`absolute top-0 right-0 w-4 h-4 ${getResizeCursor('ne')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-tr-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'ne')}
                />
                <div 
                    class={`absolute bottom-0 left-0 w-4 h-4 ${getResizeCursor('sw')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-bl-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'sw')}
                />
                <div 
                    class={`absolute bottom-0 right-0 w-4 h-4 ${getResizeCursor('se')} resize-handle opacity-0 group-hover:opacity-100 transition-opacity rounded-br-lg`}
                    onMouseDown={(e) => handleResizeStart(e, 'se')}
                />
            </Show>
            
            {/* Widget content */}
            <div class="h-full">
                {props.children}
            </div>
        </div>
    );
}
