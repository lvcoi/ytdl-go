import { createSignal, createMemo, Show } from 'solid-js';
import Icon from './Icon';

export default function QuickDownload(props) {
    const [url, setUrl] = createSignal('');

    const isValid = createMemo(() => {
        const value = url().trim();
        if (value.length === 0) return false;
        
        // Split by comma and check if at least one valid URL exists
        const urls = value.split(',').map(u => u.trim());
        return urls.some(u => u.startsWith('http://') || u.startsWith('https://'));
    });

    const isInvalid = createMemo(() => {
        return url().trim().length > 0 && !isValid();
    });

    const handleDownload = (e) => {
        e.preventDefault();
        if (isValid()) {
            props.onDownload(url().trim());
            setUrl('');
        }
    };

    return (
        <div class="rounded-[2rem] border border-white/5 bg-black/20 p-6 flex flex-col gap-4">
            <div class="flex items-center justify-between">
                <h3 class="text-sm font-bold text-gray-400 uppercase tracking-widest flex items-center gap-2">
                    <Icon name="zap" class="w-4 h-4 text-accent-primary" />
                    Quick Download
                </h3>
            </div>
            
            <form onSubmit={handleDownload} class="flex flex-col gap-1">
                <div class="flex gap-2">
                    <div class="relative flex-1 group">
                        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                            <Icon name="search" class={`w-4 h-4 transition-colors ${isInvalid() ? 'text-red-400' : 'text-gray-500 group-focus-within:text-accent-primary'}`} />
                        </div>
                        <input
                            type="text"
                            value={url()}
                            onInput={(e) => setUrl(e.target.value)}
                            placeholder="Paste YouTube URL..."
                            aria-label="YouTube URL"
                            aria-invalid={isInvalid()}
                            class={`w-full bg-white/5 border rounded-xl py-3 pl-11 pr-4 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 transition-all ${
                                isInvalid()
                                ? 'border-red-500/50 focus:ring-red-500/50 focus:border-red-500/50'
                                : 'border-white/10 focus:ring-accent-primary/50 focus:border-accent-primary/50'
                            }`}
                        />
                    </div>
                    <button
                        type="submit"
                        disabled={!isValid()}
                        class={`px-6 py-3 rounded-xl font-black uppercase tracking-widest flex items-center gap-2 transition-all ${
                            isValid()
                            ? 'bg-white text-black hover:scale-105 active:scale-95 shadow-lg'
                            : 'bg-white/5 text-gray-600 cursor-not-allowed border border-white/5'
                        }`}
                    >
                        <Icon name="download-cloud" class="w-4 h-4" />
                        Download
                    </button>
                </div>
                <Show when={isInvalid()}>
                    <p class="text-[10px] text-red-400 font-medium px-1">Please enter a valid YouTube URL</p>
                </Show>
            </form>
            <p class="text-[10px] text-gray-500 font-medium">
                Downloads will use your default settings. Go to the <span class="text-gray-400 cursor-pointer hover:underline" onClick={() => props.onTabChange?.('download')}>Download</span> tab for more options.
            </p>
        </div>
    );
}
