import Icon from './Icon';

export default function SettingsView() {
  return (
    <div class="space-y-10 transition-smooth animate-in fade-in slide-in-from-right-4 duration-500 glass-vibrant p-10 rounded-[2.5rem] border border-accent-primary/20 shadow-2xl relative overflow-hidden">
      <div class="absolute top-0 right-0 p-12 opacity-5 pointer-events-none">
        <Icon name="settings" class="w-48 h-48 rotate-12" />
      </div>

      <div class="relative space-y-8">
        <div class="space-y-2">
          <h1 class="text-5xl font-black tracking-tight text-white">System Settings</h1>
          <p class="text-gray-400 font-medium text-lg">Configure your environment and authentication providers.</p>
        </div>

        <div class="space-y-4">
          <div class="p-8 glass rounded-[2rem] border border-white/5 space-y-6 group hover:border-accent-primary/30 transition-smooth">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-5">
                <div class="p-4 bg-accent-primary/10 text-accent-primary rounded-2xl">
                  <Icon name="key" class="w-6 h-6" />
                </div>
                <div>
                  <div class="text-xl font-black text-white">YouTube Auth Cookies</div>
                  <div class="text-sm font-medium text-gray-500">Currently synced from local browser profile</div>
                </div>
              </div>
              <div class="relative has-tooltip">
                <span class="tooltip glass border border-white/10 text-[10px] px-4 py-2 rounded-xl shadow-2xl mb-4 w-56 text-center leading-relaxed z-50">
                  Coming Soon: Manual cookie upload and session management.
                </span>
                <button
                  type="button"
                  disabled
                  class="px-6 py-3 bg-white/5 rounded-xl text-xs font-black uppercase tracking-widest text-gray-700 border border-white/5 cursor-not-allowed opacity-50"
                >
                  Re-sync
                </button>
              </div>
            </div>

            <div class="h-[1px] bg-white/5"></div>

            <div class="flex items-center justify-between">
              <div class="flex items-center gap-5">
                <div class="p-4 bg-accent-secondary/10 text-accent-secondary rounded-2xl">
                  <Icon name="shield-check" class="w-6 h-6" />
                </div>
                <div>
                  <div class="text-xl font-black text-white">PO Token Extension</div>
                  <div class="text-sm font-medium text-gray-500">Enable third-party providers for token generation</div>
                </div>
              </div>
              <div class="relative has-tooltip">
                <span class="tooltip glass border border-white/10 text-[10px] px-4 py-2 rounded-xl shadow-2xl mb-4 w-60 text-center leading-relaxed z-50">
                  Coming Soon: Toggle between local and remote token providers.
                </span>
                <button
                  type="button"
                  disabled
                  class="w-14 h-8 bg-white/5 rounded-full flex items-center px-1.5 border border-white/5 cursor-not-allowed opacity-50"
                >
                  <span class="w-5 h-5 bg-gray-700 rounded-full shadow-lg"></span>
                </button>
              </div>
            </div>

            <div class="h-[1px] bg-white/5"></div>

            <div class="flex items-center justify-between">
              <div class="flex items-center gap-5">
                <div class="p-4 bg-emerald-500/10 text-emerald-400 rounded-2xl">
                  <Icon name="refresh-cw" class="w-6 h-6" />
                </div>
                <div>
                  <div class="text-xl font-black text-white">Auto-Update</div>
                  <div class="text-sm font-medium text-gray-500">Keep ytdl-go binary and UI up to date</div>
                </div>
              </div>
              <div class="w-14 h-8 bg-emerald-500/20 rounded-full flex items-center justify-end px-1.5 border border-emerald-500/30 cursor-pointer">
                <span class="w-5 h-5 bg-emerald-400 rounded-full shadow-[0_0_15px_rgba(52,211,153,0.5)]"></span>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2 p-4 glass rounded-2xl border-white/5 text-[10px] font-black uppercase tracking-widest text-gray-600 justify-center">
          <Icon name="info" class="w-3 h-3" />
          Settings are saved locally to your browser profile.
        </div>
      </div>
    </div>
  );
}
