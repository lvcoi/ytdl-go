import Icon from './Icon';

export default function Sidebar(props) {
    const activeTab = () => props.activeTab;

    return (
        <aside class="w-72 bg-bg-surface/85 border-r border-white/10 flex flex-col p-6 backdrop-blur-xl transition-all duration-300">
            <div class="flex items-center gap-3 mb-10 px-2">
                <div class="w-10 h-10 bg-accent-primary rounded-2xl flex items-center justify-center shadow-lg shadow-accent-primary/20">
                    <Icon name="zap" class="w-6 h-6 text-white fill-white" />
                </div>
                <span class="text-xl font-bold tracking-tight text-white">ytdl-go</span>
            </div>

            <nav class="flex-1 space-y-2">
                <button
                    onClick={() => props.onTabChange('dashboard')}
                    class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${activeTab() === 'dashboard'
                        ? 'bg-accent-primary/10 text-accent-secondary shadow-sm shadow-accent-primary/5'
                        : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'
                        }`}
                >
                    <Icon name="layout-dashboard" class="w-5 h-5" />
                    <span class="font-bold text-sm">Dashboard</span>
                </button>

                <div class="pt-4 pb-2 px-4 text-xs font-bold text-gray-600 uppercase tracking-widest">
                    Media
                </div>

                <button
                    onClick={() => props.onTabChange('download')}
                    class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${activeTab() === 'download'
                        ? 'bg-accent-primary/10 text-accent-secondary shadow-sm shadow-accent-primary/5'
                        : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'
                        }`}
                >
                    <Icon name="plus-circle" class="w-5 h-5" />
                    <span class="font-bold text-sm">New Download</span>
                </button>
                <button
                    onClick={() => props.onTabChange('library')}
                    class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${activeTab() === 'library'
                        ? 'bg-accent-primary/10 text-accent-secondary shadow-sm shadow-accent-primary/5'
                        : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'
                        }`}
                >
                    <Icon name="layers" class="w-5 h-5" />
                    <span class="font-bold text-sm">Library</span>
                </button>

                <div class="pt-4 pb-2 px-4 text-xs font-bold text-gray-600 uppercase tracking-widest">
                    System
                </div>

                <button
                    onClick={() => props.onTabChange('settings')}
                    class={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 ${activeTab() === 'settings'
                        ? 'bg-accent-primary/10 text-accent-secondary shadow-sm shadow-accent-primary/5'
                        : 'text-gray-500 hover:bg-white/5 hover:text-gray-300'
                        }`}
                >
                    <Icon name="sliders" class="w-5 h-5" />
                    <span class="font-bold text-sm">Settings</span>
                </button>
            </nav>

            <div class="mt-auto relative has-tooltip">
                <span class="tooltip bg-bg-surface-soft text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-56 text-center leading-relaxed">
                    Coming Soon: extension provider management and health status.
                </span>
                <button
                    type="button"
                    disabled
                    aria-disabled="true"
                    aria-label="Extensions panel (Coming Soon)"
                    class="w-full p-4 bg-white/5 rounded-2xl border border-white/5 opacity-70 cursor-not-allowed text-left hover:bg-white/10 transition-colors"
                >
                    <div class="flex items-center gap-2 mb-2 text-xs font-bold text-gray-500 uppercase tracking-widest">
                        <Icon name="puzzle" class="w-3 h-3" />
                        Extensions
                    </div>
                    <div class="flex items-center justify-between text-xs">
                        <span class="text-gray-500">PO Token Provider</span>
                        <span class="px-2 py-0.5 bg-white/10 text-gray-500 rounded-full font-bold border border-white/10">
                            Coming Soon
                        </span>
                    </div>
                </button>
            </div>
        </aside>
    );
}
