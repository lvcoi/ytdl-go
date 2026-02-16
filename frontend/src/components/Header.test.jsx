import { describe, it, expect } from 'vitest';
import { render } from '@solidjs/testing-library';
import Header from './Header';

describe('Header component', () => {
    it('renders title and right-side actions correctly', () => {
        const { getByText } = render(() => (
            <Header title="Dashboard" />
        ));
        
        expect(getByText('Dashboard')).toBeInTheDocument();
        expect(getByText('YT_AUTH_OK')).toBeInTheDocument();
        expect(getByText('Advanced Mode')).toBeInTheDocument();
    });

    it('action buttons have consistent text styling', () => {
        const { getByText } = render(() => (
            <Header title="Dashboard" />
        ));
        
        const ytAuthText = getByText('YT_AUTH_OK');
        const advancedModeButton = getByText('Advanced Mode');
        
        // Let's assume we want them both non-italic for balance
        expect(ytAuthText).not.toHaveClass('italic');
    });
});
