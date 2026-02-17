import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@solidjs/testing-library';
import DashboardView from './DashboardView';

// Mock child components to isolate DashboardView logic
vi.mock('./QuickDownload', () => ({ default: () => <div data-testid="quick-download">QuickDownload</div> }));
vi.mock('./ActiveDownloads', () => ({ default: () => <div data-testid="active-downloads">ActiveDownloads</div> }));
vi.mock('./dashboard/WelcomeWidget', () => ({ default: () => <div data-testid="welcome-widget">WelcomeWidget</div> }));
vi.mock('./dashboard/StatsWidget', () => ({ default: () => <div data-testid="stats-widget">StatsWidget</div> }));
vi.mock('./dashboard/RecentActivityWidget', () => ({ default: () => <div data-testid="recent-activity-widget">RecentActivityWidget</div> }));

describe('DashboardView', () => {
    it('renders all widgets within the grid', async () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Check for all widgets
        expect(await screen.findByTestId('quick-download')).toBeInTheDocument();
        expect(screen.getByTestId('welcome-widget')).toBeInTheDocument();
        expect(screen.getByTestId('stats-widget')).toBeInTheDocument();
        expect(screen.getByTestId('recent-activity-widget')).toBeInTheDocument();
        expect(screen.getByTestId('active-downloads')).toBeInTheDocument();
    });
});
