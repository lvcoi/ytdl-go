import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@solidjs/testing-library';
import DashboardView from './DashboardView';

// Mock child components to isolate DashboardView logic
vi.mock('./QuickDownload', () => ({ default: () => <div data-testid="quick-download">QuickDownload</div> }));
vi.mock('./ActiveDownloads', () => ({ default: () => <div data-testid="active-downloads">ActiveDownloads</div> }));
vi.mock('./dashboard/WelcomeWidget', () => ({ default: () => <div data-testid="welcome-widget">WelcomeWidget</div> }));
vi.mock('./dashboard/StatsWidget', () => ({ default: () => <div data-testid="stats-widget">StatsWidget</div> }));
vi.mock('./dashboard/RecentActivityWidget', () => ({ default: () => <div data-testid="recent-activity-widget">RecentActivityWidget</div> }));
vi.mock('./ConcurrencyWidget', () => ({ default: () => <div data-testid="concurrency-widget">ConcurrencyWidget</div> }));

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
        
        // Find by role and name (more robust than exact text match which might be split by icons)
        const editButton = screen.getByRole('button', { name: /Edit Layout/i });
        expect(editButton).toBeInTheDocument();
        
        fireEvent.click(editButton);
        
        expect(screen.getByRole('button', { name: /Done/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Reset/i })).toBeInTheDocument();
        expect(screen.getByText(/Dashboard: Edit Mode/i)).toBeInTheDocument();
    });

    it('shows reset button only in edit mode', async () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Initially should not show reset button
        expect(screen.queryByRole('button', { name: /Reset/i })).not.toBeInTheDocument();
        
        // Enter edit mode
        const editButton = screen.getByRole('button', { name: /Edit Layout/i });
        fireEvent.click(editButton);
        
        // Should show reset button
        expect(screen.getByRole('button', { name: /Reset/i })).toBeInTheDocument();
        
        // Exit edit mode
        const doneButton = screen.getByRole('button', { name: /Done/i });
        fireEvent.click(doneButton);
        
        // Should hide reset button again
        expect(screen.queryByRole('button', { name: /Reset/i })).not.toBeInTheDocument();
    });

        it('loads layout from localStorage on mount', () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const savedLayout = [
            { id: 'welcome', enabled: true, x: 0, y: 0, width: 16, height: 3 },
            { id: 'quick-download', enabled: false, x: 0, y: 3, width: 8, height: 2 }
        ];
        
        global.localStorage.getItem.mockReturnValue(JSON.stringify(savedLayout));
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        expect(global.localStorage.getItem).toHaveBeenCalledWith('ytdl-go:dashboard-layout:v3');
    });

    it('migrates legacy layout format', () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const legacyLayout = [
            { id: 'welcome', span: 4, enabled: true },
            { id: 'quick-download', span: 3, enabled: false }
        ];
        
        global.localStorage.getItem.mockImplementation((key) => {
            if (key === 'ytdl-go:dashboard-layout:v3') return null;
            if (key === 'ytdl-go:dashboard-layout:v2') return JSON.stringify(legacyLayout);
            return null;
        });
        
        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Should try to load both formats
        expect(global.localStorage.getItem).toHaveBeenCalledWith('ytdl-go:dashboard-layout:v3');
        expect(global.localStorage.getItem).toHaveBeenCalledWith('ytdl-go:dashboard-layout:v2');
        
        // Should save migrated format
        expect(global.localStorage.setItem).toHaveBeenCalledWith(
            'ytdl-go:dashboard-layout:v3',
            expect.stringContaining('"width"')
        );
        
        // Verify migrated positions account for span (x*4)
        const savedCall = global.localStorage.setItem.mock.calls.find(
            ([key]) => key === 'ytdl-go:dashboard-layout:v3'
        );
        const migrated = JSON.parse(savedCall[1]);
        expect(migrated[0].x).toBe(0);
        expect(migrated[0].y).toBe(0);
        expect(migrated[0].width).toBe(16); // 4 * 4
        expect(migrated[1].x).toBe(0);
        expect(migrated[1].y).toBe(2);
        expect(migrated[1].width).toBe(12); // 3 * 4
    });

    it('does not overwrite saved layout with defaults on mount', () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const savedLayout = [
            { id: 'welcome', enabled: true, x: 0, y: 0, width: 16, height: 3 },
            { id: 'quick-download', enabled: false, x: 0, y: 3, width: 8, height: 2 }
        ];
        
        global.localStorage.getItem.mockImplementation((key) => {
            if (key === 'ytdl-go:dashboard-layout:v3') return JSON.stringify(savedLayout);
            return null;
        });

        render(() => <DashboardView libraryModel={mockModel} />);

        // Should not call setItem when loading existing layout (no changes made)
        const setItemCalls = global.localStorage.setItem.mock.calls.filter(
            ([key]) => key === 'ytdl-go:dashboard-layout:v3'
        );
        
        expect(setItemCalls.length).toBe(0);
    });

        it('persists layout changes after loading an existing v2 layout', async () => {
        const mockModel = { items: [], artists: [], videos: [], podcasts: [] };
        const savedLayout = [
            { id: 'welcome', enabled: true, x: 0, y: 0, width: 16, height: 3 },
            { id: 'quick-download', enabled: true, x: 0, y: 3, width: 8, height: 2 },
            { id: 'active-downloads', enabled: true, x: 12, y: 3, width: 4, height: 4 },
            { id: 'recent-activity', enabled: true, x: 0, y: 5, width: 12, height: 4 },
            { id: 'stats', enabled: true, x: 12, y: 7, width: 4, height: 3 },
        ];

        global.localStorage.getItem.mockImplementation((key) => {
            if (key === 'ytdl-go:dashboard-layout:v3') return JSON.stringify(savedLayout);
            return null;
        });

        render(() => <DashboardView libraryModel={mockModel} />);
        
        // Wait for queueMicrotask to complete
        await new Promise(resolve => setTimeout(resolve, 0));
        
        global.localStorage.setItem.mockClear(); // Clear any mount-time calls

        // Enter edit mode and reset layout
        fireEvent.click(screen.getByRole('button', { name: /Edit Layout/i }));
        fireEvent.click(screen.getByRole('button', { name: /Reset/i }));

        const setItemCalls = global.localStorage.setItem.mock.calls.filter(
            ([key]) => key === 'ytdl-go:dashboard-layout:v3'
        );
        expect(setItemCalls.length).toBeGreaterThan(0);
    });
});
