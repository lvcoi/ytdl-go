import { useAppStore } from '../store/appStore';
import Icon from '../components/Icon';

function TextField({ label, description, placeholder, value, onInput }) {
    return (
        <div class="space-y-2 p-4 bg-white/5 rounded-2xl border border-white/5 hover:border-white/10 transition-colors">
            <span class="font-bold text-gray-200">{label}</span>
            <input
                type="text"
                value={value}
                placeholder={placeholder}
                onInput={(e) => onInput(e.target.value)}
                class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm font-mono text-gray-300 placeholder-gray-600 focus:outline-none focus:border-blue-500/50"
            />
            <span class="text-xs text-gray-500">{description}</span>
        </div>
    );
}

export default function NetworkSettings() {
    const { state, setState } = useAppStore();

    return (
        <div class="max-w-4xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <header>
                <h1 class="text-3xl font-black tracking-tight text-transparent bg-clip-text bg-vibrant-gradient mb-2">
                    Network & APIs
                </h1>
                <p class="text-gray-400 font-medium">
                    Manage proxy configurations, rate limits, and external API tokens.
                </p>
            </header>

            <section class="bg-bg-surface border border-white/5 rounded-3xl p-6 shadow-2xl overflow-hidden relative">
                <div class="absolute top-0 right-0 w-64 h-64 bg-blue-500/5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 pointer-events-none"></div>

                <h2 class="text-xl font-bold text-gray-200 mb-6 flex items-center gap-2">
                    <Icon name="wifi" class="w-5 h-5 text-blue-400" />
                    Connection Preferences
                </h2>

                <div class="space-y-6 relative z-10">
                    <div class="flex items-center justify-between p-4 bg-white/5 rounded-2xl border border-white/5 hover:border-white/10 transition-colors">
                        <div class="flex flex-col">
                            <span class="font-bold text-gray-200">Enforce Cookie Usage</span>
                            <span class="text-sm text-gray-400 mt-1">
                                Utilize browser cookies to bypass standard rate limits on supported sites.
                            </span>
                        </div>
                        <label class="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                class="sr-only peer"
                                checked={state.settings.useCookies}
                                onChange={(e) => setState('settings', 'useCookies', e.target.checked)}
                            />
                            <div class="w-11 h-6 bg-white/10 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-primary"></div>
                        </label>
                    </div>

                    <div class="flex items-center justify-between p-4 bg-white/5 rounded-2xl border border-white/5 hover:border-white/10 transition-colors">
                        <div class="flex flex-col">
                            <span class="font-bold text-gray-200">PO Token Extension</span>
                            <span class="text-sm text-gray-400 mt-1">
                                Enable experimental Proof of Origin tokens for enhanced age-gated media access.
                            </span>
                        </div>
                        <label class="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                class="sr-only peer"
                                checked={state.settings.poTokenExtension}
                                onChange={(e) => setState('settings', 'poTokenExtension', e.target.checked)}
                            />
                            <div class="w-11 h-6 bg-white/10 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent-primary"></div>
                        </label>
                    </div>

                    <TextField
                        label="Proxy"
                        description="HTTP/HTTPS/SOCKS5 proxy used for all downloads, e.g. socks5://127.0.0.1:1080."
                        placeholder="socks5://127.0.0.1:1080"
                        value={state.settings.proxy}
                        onInput={(v) => setState('settings', 'proxy', v)}
                    />

                    <TextField
                        label="Download Rate Limit"
                        description="Cap the download rate, e.g. 500K or 2M. Leave empty for unlimited."
                        placeholder="2M"
                        value={state.settings.limitRate}
                        onInput={(v) => setState('settings', 'limitRate', v)}
                    />

                    <TextField
                        label="Sleep Between API Requests"
                        description="Seconds to wait between YouTube API calls — protects against 429 rate limits on long playlists."
                        placeholder="1"
                        value={String(state.settings.sleepRequests || '')}
                        onInput={(v) => {
                            const parsed = Number(v);
                            setState('settings', 'sleepRequests', Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : 0);
                        }}
                    />
                </div>
            </section>
        </div>
    );
}
