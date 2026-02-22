import { Show } from 'solid-js';
import Icon from '../Icon';

const WIDGET_TITLES = {
  'quick-download': 'Quick Download',
  'recent-downloads': 'Recent Downloads',
  'system-stats': 'System Stats',
  'storage': 'Storage',
};

export default function WidgetContainer(props) {
  // props: widget, isEditing, isDragging, isResizing, onDragStart, onResizeStart, children

  const widget = () => props.widget;
  const isEditing = () => props.isEditing;
  const isDragging = () => props.isDragging;
  const isResizing = () => props.isResizing;

  const gridStyle = () => ({
    'grid-column': `${widget().col} / span ${widget().colSpan}`,
    'grid-row': `${widget().row} / span ${widget().rowSpan}`,
    opacity: isDragging() ? '0.4' : '1',
    'z-index': isDragging() || isResizing() ? '100' : '1',
    transition: isDragging() || isResizing() ? 'none' : 'opacity 0.2s',
  });

  const contentBlocked = () => isDragging() || isResizing();

  return (
    <div
      style={gridStyle()}
      class="relative flex flex-col bg-[#0a0c14] border border-white/10 rounded-2xl overflow-hidden"
    >
      {/* Header */}
      <div class="flex items-center justify-between px-4 py-3 border-b border-white/5 flex-shrink-0">
        <span class="text-sm font-semibold text-gray-300">{WIDGET_TITLES[widget().type] || widget().type}</span>
        <Show when={isEditing()}>
          <div
            class="cursor-grab active:cursor-grabbing p-1 rounded-lg hover:bg-white/10 text-gray-500 hover:text-gray-300 transition-colors"
            onMouseDown={(e) => {
              e.preventDefault();
              props.onDragStart && props.onDragStart(e, widget().id);
            }}
          >
            <Icon name="grip" class="w-4 h-4" />
          </div>
        </Show>
      </div>

      {/* Content */}
      <div
        class="flex-1 overflow-hidden p-3"
        style={{ 'pointer-events': contentBlocked() ? 'none' : 'auto' }}
      >
        {props.children}
      </div>

      {/* Resize handle */}
      <Show when={isEditing()}>
        <div
          class="absolute bottom-0 right-0 w-6 h-6 flex items-end justify-end pr-1 pb-1 cursor-se-resize text-gray-600 hover:text-gray-300 transition-colors"
          onMouseDown={(e) => {
            e.preventDefault();
            e.stopPropagation();
            props.onResizeStart && props.onResizeStart(e, widget().id);
          }}
        >
          <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor">
            <circle cx="8" cy="8" r="1.5"/>
            <circle cx="4" cy="8" r="1.5"/>
            <circle cx="8" cy="4" r="1.5"/>
          </svg>
        </div>
      </Show>
    </div>
  );
}
