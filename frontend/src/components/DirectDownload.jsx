import { createSignal } from 'solid-js';
import Icon from './Icon';

export default function DirectDownload(props) {
    const [url, setUrl] = createSignal('');

    const handleSubmit = (e) => {
        e.preventDefault();
        const currentUrl = url().trim();
        if (currentUrl) {
            props.onDownload(currentUrl);
            setUrl('');
        }
    };

    return (
        <form 
            onSubmit={handleSubmit}
            class="flex items-center gap-2 p-1.5 rounded-2xl bg-black/40 border border-white/5 focus-within:border-accent-primary/30 focus-within:bg-black/60 transition-all duration-300 group"
        >
            <div class="flex-1 flex items-center gap-3 px-4">
                <Icon name="link" class="w-4 h-4 text-gray-500 group-focus-within:text-accent-primary transition-colors" />
                <input
                    type="text"
                    value={url()}
                    onInput={(e) => setUrl(e.currentTarget.value)}
                    placeholder="Paste YouTube URL here for immediate download..."
                    class="w-full bg-transparent border-none outline-none text-sm text-white placeholder:text-gray-600 font-medium"
                />
            </div>
            <button
                type="submit"
                disabled={!url().trim()}
                class="px-6 py-2.5 rounded-xl bg-white text-black font-black uppercase tracking-widest text-[10px] hover:scale-[1.02] active:scale-[0.98] disabled:opacity-30 disabled:grayscale disabled:hover:scale-100 transition-all shadow-lg"
            >
                Download
            </button>
        </form>
    );
}
