import { For, Show } from 'solid-js';
import Icon from '../Icon';
import { WIDGET_REGISTRY } from './widgetRegistry';

export default function WidgetDrawer(props) {
    return (
        <div 
            class={`fixed inset-y-0 right-0 w-80 bg-surface-primary border-l border-white/10 shadow-2xl transform transition-transform duration-300 z-50 flex flex-col ${props.isOpen ? 'translate-x-0' : 'translate-x-full'}`}
        >
            <div class="p-6 border-b border-white/10 flex items-center justify-between">
                <h2 class="text-lg font-bold text-white">Add Widgets</h2>
                <button 
                    onClick={props.onClose}
                    class="p-2 rounded-lg hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                >
                    <Icon name="x" class="w-5 h-5" />
                </button>
            </div>
            
            <div class="flex-1 overflow-y-auto p-6 space-y-4">
                <For each={Object.values(WIDGET_REGISTRY)}>
                    {(item) => {
                        const widget = props.widgets.find(w => w.id === item.id);
                        const isEnabled = widget?.enabled;
                        
                        return (
                            <div class="p-4 rounded-xl bg-white/5 border border-white/10 hover:border-accent-primary/50 transition-colors group">
                                <div class="flex items-center gap-3 mb-3">
                                    <div class="p-2 rounded-lg bg-accent-primary/20 text-accent-primary">
                                        <Icon name={item.icon} class="w-5 h-5" />
                                    </div>
                                    <div>
                                        <div class="font-bold text-white">{item.label}</div>
                                        <div class="text-xs text-gray-400">
                                            {item.defaultW}x{item.defaultH} units
                                        </div>
                                    </div>
                                </div>
                                
                                <button
                                    onClick={() => isEnabled ? props.onRemove(item.id) : props.onAdd(item.id)}
                                    class={`w-full py-2 rounded-lg text-sm font-bold transition-all ${isEnabled 
                                        ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30' 
                                        : 'bg-accent-primary/20 text-accent-primary hover:bg-accent-primary/30'}`}
                                >
                                    {isEnabled ? 'Remove Widget' : 'Add to Dashboard'}
                                </button>
                            </div>
                        );
                    }}
                </For>
            </div>
        </div>
    );
}
