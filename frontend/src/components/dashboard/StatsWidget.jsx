export default function StatsWidget(props) {
    return (
        <div class="rounded-[2rem] border border-white/5 bg-black/20 p-6 flex flex-col justify-center gap-4 h-full">
            <h3 class="text-sm font-bold text-gray-400 uppercase tracking-widest">Library Stats</h3>
            <div class="grid grid-cols-2 gap-4">
                <div class="p-4 rounded-2xl bg-white/5 border border-white/5">
                    <div class="text-3xl font-black text-accent-primary">{props.stats?.totalCreators || 0}</div>
                    <div class="text-[10px] font-bold text-gray-500 uppercase">Creators</div>
                </div>
                <div class="p-4 rounded-2xl bg-white/5 border border-white/5">
                    <div class="text-3xl font-black text-accent-secondary">{props.stats?.totalItems || 0}</div>
                    <div class="text-[10px] font-bold text-gray-500 uppercase">Media Files</div>
                </div>
            </div>
        </div>
    );
}
