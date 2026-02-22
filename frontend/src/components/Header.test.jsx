import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@solidjs/testing-library';
import Header from './Header';

describe('Header', () => {
    it('renders the title based on activeTab', () => {
        render(() => <Header activeTab="dashboard" />);
        expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    it('renders the YT_AUTH_OK badge', () => {
        render(() => <Header activeTab="dashboard" />);
        expect(screen.getByText('YT_AUTH_OK')).toBeInTheDocument();
    });

    it('toggles Advanced Mode style based on prop', () => {
        const { unmount } = render(() => <Header activeTab="dashboard" isAdvanced={true} />);
        const button = screen.getByText('Advanced Mode').closest('button');
        expect(button).toHaveClass('bg-accent-primary/20');
        unmount();

        render(() => <Header activeTab="dashboard" isAdvanced={false} />);
        const inactiveButton = screen.getByText('Advanced Mode').closest('button');
        expect(inactiveButton).toHaveClass('bg-white/5');
    });

    it('calls onToggleAdvanced when button is clicked', () => {
        const toggleMock = vi.fn();
        render(() => <Header activeTab="dashboard" onToggleAdvanced={toggleMock} />);
        fireEvent.click(screen.getByText('Advanced Mode'));
        expect(toggleMock).toHaveBeenCalled();
    });
});
