import { createSignal, For, Show } from 'solid-js';
import Icon from '../Icon';

export default function LayoutPresets(props) {
    const [isOpen, setIsOpen] = createSignal(false);
    const [isNaming, setIsNaming] = createSignal(false);
    const [newName, setNewName] = createSignal('');

    const handleSave = (e) => {
        e.preventDefault();
        if (newName().trim()) {
            props.onSave(newName().trim());
            setNewName('');
            setIsNaming(false);
        }
    };

    return (
        <div class="relative">
            <button
                onClick={() => setIsOpen(!isOpen())}
                class="flex items-center gap-2 px-3 py-2 rounded-xl border border-white/10 bg-white/5 text-gray-400 hover:bg-white/10 hover:text-white transition-all text-xs font-bold"
            >
                <Icon name="layout" class="w-3.5 h-3.5" />
                <span class="max-w-[100px] truncate">{props.activeLayoutName || 'Layouts'}</span>
                <Icon name="chevron-down" class="w-3 h-3" />
            </button>

            <Show when={isOpen()}>
                <div class="absolute top-full left-0 mt-2 w-64 bg-surface-primary border border-white/10 rounded-xl shadow-xl z-50 overflow-hidden">
                    {/* Header / Save New */}
                    <div class="p-3 border-b border-white/10">
                        <Show when={!isNaming()} fallback={
                            <form onSubmit={handleSave} class="flex gap-2">
                                <input
                                    type="text"
                                    value={newName()}
                                    onInput={(e) => setNewName(e.target.value)}
                                    placeholder="Layout Name"
                                    class="flex-1 bg-black/20 border border-white/10 rounded-lg px-2 py-1 text-xs text-white focus:outline-none focus:border-accent-primary/50"
                                    autoFocus
                                />
                                <button type="submit" class="p-1 rounded-lg bg-accent-primary/20 text-accent-primary hover:bg-accent-primary/30">
                                    <Icon name="check" class="w-3.5 h-3.5" />
                                </button>
                                <button type="button" onClick={() => setIsNaming(false)} class="p-1 rounded-lg hover:bg-white/10 text-gray-400">
                                    <Icon name="x" class="w-3.5 h-3.5" />
                                </button>
                            </form>
                        }>
                            <button 
                                onClick={() => setIsNaming(true)}
                                class="w-full flex items-center justify-center gap-2 py-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-xs font-bold text-gray-300 transition-colors"
                            >
                                <Icon name="plus" class="w-3.5 h-3.5" />
                                Save Current Layout
                            </button>
                        </Show>
                    </div>

                    {/* Layout List */}
                    <div class="max-h-60 overflow-y-auto p-2 space-y-1">
                        <For each={props.layouts}>
                            {(layout) => (
                                <div class={`group flex items-center justify-between p-2 rounded-lg transition-colors ${props.activeLayoutId === layout.id ? 'bg-accent-primary/10' : 'hover:bg-white/5'}`}>
                                    <button 
                                        onClick={() => { props.onLoad(layout.id); setIsOpen(false); }}
                                        class="flex-1 text-left flex items-center gap-2 overflow-hidden"
                                    >
                                        <span class={`text-xs font-medium truncate ${props.activeLayoutId === layout.id ? 'text-accent-primary' : 'text-gray-300'}`}>
                                            {layout.name}
                                        </span>
                                        <Show when={layout.isPrimary}>
                                            <Icon name="star" class="w-3 h-3 text-amber-400 fill-amber-400" />
                                        </Show>
                                    </button>
                                    
                                    <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <Show when={!layout.isFactory}>
                                            <button 
                                                onClick={() => props.onSetPrimary(layout.id)}
                                                class="p-1 rounded hover:bg-white/10 text-gray-500 hover:text-amber-400"
                                                title="Set as Startup Default"
                                            >
                                                <Icon name="star" class="w-3 h-3" />
                                            </button>
                                            <button 
                                                onClick={() => props.onDelete(layout.id)}
                                                class="p-1 rounded hover:bg-white/10 text-gray-500 hover:text-red-400"
                                                title="Delete Layout"
                                            >
                                                <Icon name="trash-2" class="w-3 h-3" />
                                            </button>
                                        </Show>
                                    </div>
                                </div>
                            )}
                        </For>
                    </div>
                    
                    {/* Reset to Default */}
                    <div class="p-2 border-t border-white/10 bg-black/20">
                        <button 
                            onClick={() => { props.onLoad('default'); setIsOpen(false); }}
                            class="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-xs text-gray-500 hover:text-gray-300 hover:bg-white/5 transition-colors"
                        >
                            <Icon name="refresh-cw" class="w-3 h-3" />
                            Restore Factory Default
                        </button>
                    </div>
                </div>
            </Show>
            
            {/* Backdrop to close */}
            <Show when={isOpen()}>
                <div class="fixed inset-0 z-40" onClick={() => setIsOpen(false)} />
            </Show>
        </div>
    );
}
