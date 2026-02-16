import { render, screen, fireEvent } from '@solidjs/testing-library';
import { describe, it, expect, vi } from 'vitest';
import DashboardView from './DashboardView';
import { AppStoreProvider } from '../store/appStore';

describe('DashboardView component', () => {
    const mockLibraryModel = () => ({
        items: [],
        artists: [],
        videos: [],
        podcasts: [],
    });

    const renderDashboard = (props = {}) => {
        return render(() => (
            <AppStoreProvider>
                <DashboardView 
                    libraryModel={mockLibraryModel}
                    onTabChange={() => {}}
                    onDirectDownload={() => {}}
                    {...props}
                />
            </AppStoreProvider>
        ));
    };

    it('renders welcome message and DirectDownload', () => {
        renderDashboard();
        expect(screen.getByText(/Welcome Back!/i)).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/Paste YouTube URL here/i)).toBeInTheDocument();
    });

            it('triggers onDirectDownload when DirectDownload is used', () => {
        const onDirectDownload = vi.fn();
        renderDashboard({ onDirectDownload });
        
        const input = screen.getByPlaceholderText(/Paste YouTube URL here/i);
        // Be extremely specific to the form's submit button
        const button = screen.getByRole('button', { name: /^Download$/i });

        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=test' } });
        fireEvent.click(button);

        expect(onDirectDownload).toHaveBeenCalledWith('https://youtube.com/watch?v=test');
    });


});
