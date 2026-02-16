import { Show } from 'solid-js';
import Icon from './Icon';

export default function Header(props) {
    const isAdvanced = () => props.isAdvanced;
    const title = () => props.title || '';

    return (
        <header class="h-20 border-b border-white/5 flex items-center justify-between px-10 glass sticky top-0 z-20 backdrop-blur-md bg-bg-surface/50">
            <h2 class="text-lg font-bold text-white capitalize tracking-tight">{title()}</h2>
            <div class="flex items-center gap-6">
                <div class="relative has-tooltip group">
                    <span class="tooltip bg-bg-surface-soft text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-56 text-center leading-relaxed opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none">
                        Coming Soon: Live auth cookie status and diagnostics.
                    </span>
                    <button
                        type="button"
                        disabled
                        aria-disabled="true"
                        aria-label="YouTube auth status details (Coming Soon)"
                        class="flex items-center gap-3 px-4 py-2 bg-white/5 rounded-full border border-white/5 cursor-not-allowed opacity-70 hover:bg-white/10 transition-colors"
                    >
                        <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse shadow-[0_0_8px_rgba(34,197,94,0.6)]"></div>
                        <span class="text-xs font-bold text-gray-400">YT_AUTH_OK</span>
                        <Icon name="chevron-down" class="w-3 h-3 text-gray-500" />
                    </button>
                </div>

                <button
                    onClick={props.onToggleAdvanced}
                    class={`px-4 py-2 rounded-full text-xs font-bold transition-all duration-300 border ${isAdvanced()
                        ? 'bg-accent-primary border-accent-primary text-white shadow-lg shadow-accent-primary/20'
                        : 'bg-white/5 border-white/5 text-gray-500 hover:text-gray-300 hover:bg-white/10'
                        }`}
                >
                    Advanced Mode
                </button>
            </div>
        </header>
    );
}
