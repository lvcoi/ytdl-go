import { createSignal } from 'solid-js';
import Icon from './Icon';

export default function DirectDownload(props) {
    const [url, setUrl] = createSignal('');

    const handleDownload = (e) => {
        e.preventDefault();
        const currentUrl = url().trim();
        if (currentUrl) {
            props.onDownload(currentUrl);
            setUrl('');
        }
    };

    return (
        <div class="rounded-2xl border border-accent-primary/20 bg-accent-primary/5 p-0.5 transition-all duration-500 focus-within:border-accent-primary/40 focus-within:bg-accent-primary/10 group shadow-lg shadow-accent-primary/5 h-12 flex items-center">
            <form onSubmit={handleDownload} class="flex items-center gap-2 w-full h-full">
                <div class="flex-1 flex items-center gap-2 px-4 h-full">
                    <Icon name="link-2" class="w-4 h-4 text-accent-primary opacity-60 group-focus-within:opacity-100 transition-opacity" />
                    <input
                        type="text"
                        value={url()}
                        onInput={(e) => setUrl(e.target.value)}
                        placeholder="Quick Download: Paste URL..."
                        class="w-full bg-transparent border-none outline-none text-white text-sm placeholder:text-gray-500 placeholder:font-normal"
                    />
                </div>
                <button
                    type="submit"
                    disabled={!url().trim()}
                    class={`px-4 h-[calc(100%-4px)] rounded-xl font-black text-[10px] uppercase tracking-widest transition-all duration-300 flex items-center gap-2 mr-0.5 ${
                        url().trim() 
                        ? 'bg-accent-primary text-white hover:scale-[1.02] active:scale-95 shadow-md shadow-accent-primary/30' 
                        : 'bg-white/5 text-gray-500 cursor-not-allowed'
                    }`}
                >
                    <Icon name="download" class="w-3 h-3" />
                    Download
                </button>
            </form>
        </div>
    );
}
