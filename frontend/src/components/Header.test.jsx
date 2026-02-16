import { render, screen } from '@solidjs/testing-library';
import { describe, it, expect } from 'vitest';
import Header from './Header';
import { AppStoreProvider } from '../store/appStore';

describe('Header component', () => {
    const renderHeader = (props = {}) => {
        return render(() => (
            <AppStoreProvider>
                <Header 
                    activeTab="dashboard" 
                    isAdvanced={false} 
                    onToggleAdvanced={() => {}} 
                    {...props} 
                />
            </AppStoreProvider>
        ));
    };

    it('renders title and right-side actions correctly', () => {
        renderHeader();
        expect(screen.getByText('Dashboard')).toBeInTheDocument();
        expect(screen.getByText('Advanced Mode')).toBeInTheDocument();
    });

            it('action buttons have consistent text styling', () => {
        renderHeader();
        const advancedText = screen.getByText('Advanced Mode');
        expect(advancedText).toHaveClass('tracking-widest');
    });


});
