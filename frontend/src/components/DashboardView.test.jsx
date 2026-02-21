import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@solidjs/testing-library';
import DashboardView from './DashboardView';

// Mock child components to isolate DashboardView logic
vi.mock('./QuickDownload', () => ({ default: () => <div data-testid="quick-download">QuickDownload</div> }));
vi.mock('./ActiveDownloads', () => ({ default: () => <div data-testid="active-downloads">ActiveDownloads</div> }));
vi.mock('./dashboard/WelcomeWidget', () => ({ default: () => <div data-testid="welcome-widget">WelcomeWidget</div> }));
vi.mock('./dashboard/StatsWidget', () => ({ default: () => <div data-testid="stats-widget">StatsWidget</div> }));
vi.mock('./dashboard/RecentActivityWidget', () => ({ default: () => <div data-testid="recent-activity-widget">RecentActivityWidget</div> }));

describe('DashboardView', () => {
    beforeEach(() => {
        global.localStorage.getItem.mockClear();
        global.localStorage.setItem.mockClear();
    });

    afterEach(() => {
        cleanup();
    });

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

    it('enters edit mode when edit button is clicked', async () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        const editButton = screen.getByText('Edit Layout');
        expect(editButton).toBeInTheDocument();
        
        fireEvent.click(editButton);
        
        expect(screen.getByText('Done')).toBeInTheDocument();
        expect(screen.getByText('Reset')).toBeInTheDocument();
        expect(screen.getByText('Dashboard: Edit Mode')).toBeInTheDocument();
    });

    it('shows reset button only in edit mode', async () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Initially should not show reset button
        expect(screen.queryByText('Reset')).not.toBeInTheDocument();
        
        // Enter edit mode
        const editButton = screen.getByText('Edit Layout');
        fireEvent.click(editButton);
        
        // Should show reset button
        expect(screen.getByText('Reset')).toBeInTheDocument();
        
        // Exit edit mode
        const doneButton = screen.getByText('Done');
        fireEvent.click(doneButton);
        
        // Should hide reset button again
        expect(screen.queryByText('Reset')).not.toBeInTheDocument();
    });

    it('loads layout from localStorage on mount', () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const savedLayout = [
            { id: 'welcome', span: 4, enabled: true, x: 0, y: 0, width: 4, height: 2 },
            { id: 'quick-download', span: 3, enabled: false, x: 0, y: 2, width: 3, height: 2 }
        ];
        
        global.localStorage.getItem.mockReturnValue(JSON.stringify(savedLayout));
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        expect(global.localStorage.getItem).toHaveBeenCalledWith('ytdl-go:dashboard-layout:v2');
    });

    it('migrates legacy layout format', () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const legacyLayout = [
            { id: 'welcome', span: 4, enabled: true },
            { id: 'quick-download', span: 3, enabled: false }
        ];
        
        global.localStorage.getItem.mockImplementation((key) => {
            if (key === 'ytdl-go:dashboard-layout:v2') return null;
            if (key === 'ytdl-go:dashboard-layout:v1') return JSON.stringify(legacyLayout);
            return null;
        });
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Should try to load both formats
        expect(global.localStorage.getItem).toHaveBeenCalledWith('ytdl-go:dashboard-layout:v2');
        expect(global.localStorage.getItem).toHaveBeenCalledWith('ytdl-go:dashboard-layout:v1');
        
        // Should save migrated format
        expect(global.localStorage.setItem).toHaveBeenCalledWith(
            'ytdl-go:dashboard-layout:v2',
            expect.stringContaining('"width"')
        );
        
        // Verify migrated positions account for span
        const savedCall = global.localStorage.setItem.mock.calls.find(
            ([key]) => key === 'ytdl-go:dashboard-layout:v2'
        );
        const migrated = JSON.parse(savedCall[1]);
        expect(migrated[0].x).toBe(0);
        expect(migrated[0].y).toBe(0);
        expect(migrated[1].x).toBe(0);
        expect(migrated[1].y).toBe(2);
    });

    it('does not overwrite saved layout with defaults on mount', () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const savedLayout = [
            { id: 'welcome', span: 4, enabled: true, x: 0, y: 0, width: 4, height: 2 },
            { id: 'quick-download', span: 3, enabled: false, x: 0, y: 2, width: 3, height: 2 }
        ];
        
        global.localStorage.getItem.mockImplementation((key) => {
            if (key === 'ytdl-go:dashboard-layout:v2') return JSON.stringify(savedLayout);
            return null;
        });

        render(() => <DashboardView libraryModel={mockModel} />);

        // Should not call setItem when loading existing layout (no changes made)
        // The only setItem calls might be from migration, but since v2 exists, there should be none
        const setItemCalls = global.localStorage.setItem.mock.calls.filter(
            ([key]) => key === 'ytdl-go:dashboard-layout:v2'
        );
        
        // With the hasLoaded guard, loading an existing layout shouldn't trigger setItem
        expect(setItemCalls.length).toBe(0);
    });

    it('persists layout changes after loading an existing v2 layout', async () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const savedLayout = [
            { id: 'welcome', span: 4, enabled: true, x: 0, y: 0, width: 4, height: 2 },
            { id: 'quick-download', span: 3, enabled: true, x: 0, y: 2, width: 3, height: 2 },
            { id: 'active-downloads', span: 1, enabled: true, x: 3, y: 2, width: 1, height: 2 },
            { id: 'recent-activity', span: 3, enabled: true, x: 0, y: 4, width: 3, height: 2 },
            { id: 'stats', span: 1, enabled: true, x: 3, y: 4, width: 1, height: 2 },
        ];

        global.localStorage.getItem.mockImplementation((key) => {
            if (key === 'ytdl-go:dashboard-layout:v2') return JSON.stringify(savedLayout);
            return null;
        });

        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Wait for queueMicrotask to complete
        await new Promise(resolve => setTimeout(resolve, 0));
        
        global.localStorage.setItem.mockClear(); // Clear any mount-time calls

        // Enter edit mode and reset layout
        fireEvent.click(screen.getByText('Edit Layout'));
        fireEvent.click(screen.getByText('Reset'));

        const setItemCalls = global.localStorage.setItem.mock.calls.filter(
            ([key]) => key === 'ytdl-go:dashboard-layout:v2'
        );
        expect(setItemCalls.length).toBeGreaterThan(0);
    });
});
