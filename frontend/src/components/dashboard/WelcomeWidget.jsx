import { createMemo } from 'solid-js';
import Icon from '../Icon';
import logo from '../../assets/logo.png';

export default function WelcomeWidget(props) {
    return (
        <div class="rounded-[2rem] border border-accent-primary/20 glass-vibrant p-8 relative overflow-hidden group h-full">
            <div class="absolute top-0 right-0 p-8 opacity-10 group-hover:opacity-20 transition-opacity duration-500">
                <img src={logo} alt="ytdl-go logo" class="w-48 h-48 rotate-12 object-contain" />
            </div>

            <div class="relative z-10 space-y-4">
                <h1 class="text-4xl font-black text-white tracking-tight">
                    Welcome Back!
                </h1>
                <p class="text-lg text-gray-300 max-w-lg">
                    Your media library is ready. You have <span class="text-white font-bold">{props.stats?.totalItems || 0} items</span> across <span class="text-white font-bold">{props.stats?.totalCreators || 0} creators</span>.
                </p>

                <div class="flex flex-wrap gap-3 pt-4">
                    <button
                        onClick={() => props.onTabChange('download')}
                        class="px-6 py-3 rounded-xl bg-white text-black font-black uppercase tracking-widest hover:scale-105 transition-transform shadow-lg flex items-center gap-2"
                    >
                        <Icon name="plus-circle" class="w-4 h-4" />
                        New Download
                    </button>
                    <button
                        onClick={() => props.onTabChange('library')}
                        class="px-6 py-3 rounded-xl bg-black/40 text-white border border-white/10 font-black uppercase tracking-widest hover:bg-black/60 transition-colors flex items-center gap-2"
                    >
                        <Icon name="layers" class="w-4 h-4" />
                        Browse Library
                    </button>
                </div>
            </div>
        </div>
    );
}
