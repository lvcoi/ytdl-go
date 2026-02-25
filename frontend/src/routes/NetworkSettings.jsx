import { useAppStore } from '../store/appStore';
import Icon from '../components/Icon';

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
                </div>
            </section>
        </div>
    );
}
