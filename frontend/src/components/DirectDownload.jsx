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
        <div class="rounded-[2rem] border-2 border-accent-primary/30 bg-accent-primary/5 p-1 transition-all duration-500 focus-within:border-accent-primary focus-within:bg-accent-primary/10 group shadow-2xl shadow-accent-primary/10">
            <form onSubmit={handleDownload} class="flex items-center gap-2">
                <div class="flex-1 flex items-center gap-3 px-6 py-4">
                    <Icon name="link-2" class="w-5 h-5 text-accent-primary animate-pulse" />
                    <input
                        type="text"
                        value={url()}
                        onInput={(e) => setUrl(e.target.value)}
                        placeholder="Paste URL to download immediately..."
                        class="w-full bg-transparent border-none outline-none text-white font-medium placeholder:text-gray-500 placeholder:font-normal"
                    />
                </div>
                <button
                    type="submit"
                    disabled={!url().trim()}
                    class={`px-8 py-4 rounded-[1.7rem] font-black uppercase tracking-widest transition-all duration-300 flex items-center gap-2 m-1 ${
                        url().trim() 
                        ? 'bg-accent-primary text-white hover:scale-[1.02] active:scale-95 shadow-lg shadow-accent-primary/40' 
                        : 'bg-white/5 text-gray-500 cursor-not-allowed'
                    }`}
                >
                    <Icon name="download" class={`w-4 h-4 ${url().trim() ? 'animate-bounce' : ''}`} />
                    Download Now
                </button>
            </form>
        </div>
    );
}
