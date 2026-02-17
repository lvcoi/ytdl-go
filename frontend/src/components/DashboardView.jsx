import { createMemo, lazy, Suspense } from 'solid-js';
import ActiveDownloads from './ActiveDownloads';
import { Grid, GridItem } from './Grid';
import WelcomeWidget from './dashboard/WelcomeWidget';
import StatsWidget from './dashboard/StatsWidget';
import RecentActivityWidget from './dashboard/RecentActivityWidget';

const QuickDownload = lazy(() => import('./QuickDownload'));

export default function DashboardView(props) {
    const libraryModel = createMemo(() => (typeof props.libraryModel === 'function' ? props.libraryModel() : props.libraryModel));

    const stats = createMemo(() => {
        const model = libraryModel();
        return {
            totalItems: model?.items?.length || 0,
            totalCreators: (model?.artists?.length || 0) + (model?.videos?.length || 0) + (model?.podcasts?.length || 0),
            recentItems: model?.items?.slice(0, 4) || [], // Latest 4 items
        };
    });

    const handleQuickDownload = (url) => {
        if (props.onDownload) {
            props.onDownload(url);
        }
        // Optionally, switch to the active downloads view
        // This could be a user setting in the future
    };

    return (
        <div class="transition-smooth animate-in fade-in slide-in-from-right-4 duration-500">
            <Grid>
                {/* Welcome & Quick Actions Section - Full Width */}
                <GridItem span={4}>
                    <WelcomeWidget stats={stats()} onTabChange={props.onTabChange} />
                </GridItem>

                {/* Quick Download */}
                <GridItem span={3}>
                    <Suspense fallback={
                         <div class="rounded-[2rem] border border-dashed border-white/10 bg-black/20 p-6 h-[148px] animate-pulse" />
                    }>
                        <QuickDownload onDownload={handleQuickDownload} onTabChange={props.onTabChange} />
                    </Suspense>
                </GridItem>

                {/* Active Downloads Widget */}
                <GridItem span={1}>
                    <div class="h-full">
                        <ActiveDownloads />
                    </div>
                </GridItem>

                {/* Recent Activity */}
                <GridItem span={3}>
                    <RecentActivityWidget stats={stats()} onTabChange={props.onTabChange} />
                </GridItem>

                {/* Library Stats */}
                <GridItem span={1}>
                    <StatsWidget stats={stats()} />
                </GridItem>
            </Grid>
        </div>
    );
}
