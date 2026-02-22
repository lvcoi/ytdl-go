import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@solidjs/testing-library';
import LayoutPresets from './LayoutPresets';

describe('LayoutPresets', () => {
    afterEach(cleanup);

    const mockLayouts = [
        { id: 'default', name: 'Default', isFactory: true },
        { id: 'custom-1', name: 'My Layout', isFactory: false, isPrimary: true }
    ];

    it('toggles dropdown on click', () => {
        render(() => <LayoutPresets layouts={mockLayouts} />);
        
        expect(screen.queryByText('My Layout')).not.toBeInTheDocument();
        
        fireEvent.click(screen.getByRole('button')); // The trigger button
        
        expect(screen.getByText('My Layout')).toBeInTheDocument();
    });

    it('calls onLoad when layout clicked', () => {
        const onLoad = vi.fn();
        render(() => <LayoutPresets layouts={mockLayouts} onLoad={onLoad} />);
        
        fireEvent.click(screen.getByRole('button'));
        fireEvent.click(screen.getByText('My Layout'));
        
        expect(onLoad).toHaveBeenCalledWith('custom-1');
    });

    it('shows save input when "Save Current" clicked', () => {
        render(() => <LayoutPresets layouts={mockLayouts} />);
        
        fireEvent.click(screen.getByRole('button'));
        fireEvent.click(screen.getByText('Save Current Layout'));
        
        expect(screen.getByPlaceholderText('Layout Name')).toBeInTheDocument();
    });
});
