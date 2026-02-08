export default function SettingsView() {
  return (
    <div class="space-y-6">
      <h1 class="text-3xl font-black text-white">System Settings</h1>
      <div class="p-8 bg-[#0a0c14] border border-white/5 rounded-[2rem] space-y-6">
        <div class="flex items-center justify-between">
          <div>
            <div class="font-bold text-white">YouTube Auth Cookies</div>
            <div class="text-xs text-gray-500">Currently synced from local browser</div>
          </div>
          <div class="relative has-tooltip">
            <span class="tooltip bg-gray-800 text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-44 text-center leading-relaxed">
              Coming Soon: manual cookie re-sync from this panel.
            </span>
            <button
              type="button"
              disabled
              aria-disabled="true"
              aria-label="Re-sync cookies (Coming Soon)"
              class="px-4 py-2 bg-white/5 rounded-xl text-xs font-bold text-gray-500 border border-white/10 cursor-not-allowed opacity-70"
            >
              Re-sync
            </button>
          </div>
        </div>
        <div class="h-[1px] bg-white/5"></div>
        <div class="flex items-center justify-between">
          <div>
            <div class="font-bold text-white">PO Token Extension</div>
            <div class="text-xs text-gray-500">Enable third-party providers for token generation</div>
          </div>
          <div class="relative has-tooltip">
            <span class="tooltip bg-gray-800 text-[10px] px-2 py-1 rounded shadow-xl mb-4 border border-white/10 w-48 text-center leading-relaxed">
              Coming Soon: runtime PO token provider controls.
            </span>
            <button
              type="button"
              disabled
              aria-disabled="true"
              aria-label="PO Token Extension toggle (Coming Soon)"
              class="w-12 h-6 bg-white/10 rounded-full flex items-center px-1 border border-white/10 cursor-not-allowed opacity-70"
            >
              <span class="w-4 h-4 bg-white rounded-full"></span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
