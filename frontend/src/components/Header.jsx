import Icon from './Icon';

export default function Header(props) {
    const isAdvanced = () => props.isAdvanced;
    const title = () => {
        switch (props.activeTab) {
            case 'dashboard': return 'Dashboard';
            case 'download': return 'New Download';
            case 'library': return 'Media Library';
            case 'settings': return 'Settings';
            default: return 'ytdl-go';
        }
    };

    return (
        <header class="h-24 px-10 flex items-center justify-between border-b border-white/5 bg-transparent backdrop-blur-sm z-20">
            <div class="flex items-center gap-4 h-9">
                <h2 class="text-2xl font-black text-white tracking-tighter uppercase leading-none">{title()}</h2>
                <div class="h-4 w-px bg-white/10 mx-2" />
                <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20">
                    <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                    <span class="text-[10px] font-bold text-emerald-400 uppercase tracking-widest">YT_AUTH_OK</span>
                </div>
            </div>

            <div class="flex items-center gap-3 h-9">
                <button
                    onClick={props.onToggleAdvanced}
                    class={`flex items-center gap-2 px-4 py-2 rounded-xl border transition-all duration-300 ${isAdvanced()
                        ? 'bg-accent-primary/20 border-accent-primary/30 text-accent-secondary shadow-lg shadow-accent-primary/10'
                        : 'bg-white/5 border-white/10 text-gray-400 hover:bg-white/10'
                        }`}
                >
                    <Icon name="settings-2" class={`w-3.5 h-3.5 ${isAdvanced() ? 'animate-pulse-slow' : ''}`} />
                    <span class="text-[10px] font-bold uppercase tracking-widest">Advanced Mode</span>
                </button>

                <div class="w-10 h-10 rounded-2xl bg-gradient-to-br from-gray-800 to-gray-900 border border-white/10 flex items-center justify-center shadow-inner group cursor-pointer hover:border-accent-primary/50 transition-colors">
                    <Icon name="user" class="w-5 h-5 text-gray-400 group-hover:text-white transition-colors" />
                </div>
            </div>
        </header>
    );
}

