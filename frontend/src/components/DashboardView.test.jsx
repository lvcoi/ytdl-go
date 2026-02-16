import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@solidjs/testing-library';
import DashboardView from './DashboardView';
import { AppStoreProvider } from '../store/appStore';

describe('DashboardView component', () => {
    const mockLibraryModel = () => ({
        items: [],
        artists: [],
        videos: [],
        podcasts: []
    });

    it('renders welcome message and DirectDownload', () => {
        const { getByText, getByPlaceholderText } = render(() => (
            <AppStoreProvider>
                <DashboardView 
                    libraryModel={mockLibraryModel} 
                    onTabChange={() => {}} 
                    onDirectDownload={() => {}}
                />
            </AppStoreProvider>
        ));
        
        expect(getByText(/Welcome Back!/i)).toBeInTheDocument();
        expect(getByPlaceholderText(/Paste URL to download immediately/i)).toBeInTheDocument();
    });

    it('triggers onDirectDownload when DirectDownload is used', () => {
        const onDirectDownload = vi.fn();
        const { getByPlaceholderText, getByRole } = render(() => (
            <AppStoreProvider>
                <DashboardView 
                    libraryModel={mockLibraryModel} 
                    onTabChange={() => {}} 
                    onDirectDownload={onDirectDownload}
                />
            </AppStoreProvider>
        ));
        
        const input = getByPlaceholderText(/Paste URL to download immediately/i);
        const button = getByRole('button', { name: /Download Now/i });
        
        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=123' } });
        fireEvent.click(button);
        
        expect(onDirectDownload).toHaveBeenCalledWith('https://youtube.com/watch?v=123');
    });
});
