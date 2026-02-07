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
                <button class="px-4 py-2 bg-white/5 rounded-xl text-xs font-bold hover:bg-white/10 transition-all">Re-sync</button>
            </div>
            <div class="h-[1px] bg-white/5"></div>
            <div class="flex items-center justify-between">
                <div>
                    <div class="font-bold text-white">PO Token Extension</div>
                    <div class="text-xs text-gray-500">Enable third-party providers for token generation</div>
                </div>
                <div class="w-12 h-6 bg-blue-600 rounded-full flex items-center px-1">
                    <div class="w-4 h-4 bg-white rounded-full ml-auto"></div>
                </div>
            </div>
        </div>
    </div>
  );
}