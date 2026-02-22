import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@solidjs/testing-library';
import WidgetDrawer from './WidgetDrawer';

// Mock registry
vi.mock('./widgetRegistry', () => ({
    WIDGET_REGISTRY: {
        'test-widget': {
            id: 'test-widget',
            label: 'Test Widget',
            icon: 'box',
            defaultW: 4,
            defaultH: 4
        }
    }
}));

describe('WidgetDrawer', () => {
    afterEach(cleanup);

    it('renders when open', () => {
        render(() => <WidgetDrawer isOpen={true} widgets={[]} />);
        expect(screen.getByText('Add Widgets')).toBeInTheDocument();
        expect(screen.getByText('Test Widget')).toBeInTheDocument();
    });

    it('calls onAdd when adding a new widget', () => {
        const onAdd = vi.fn();
        render(() => <WidgetDrawer isOpen={true} widgets={[]} onAdd={onAdd} />);
        
        fireEvent.click(screen.getByText('Add to Dashboard'));
        expect(onAdd).toHaveBeenCalledWith('test-widget');
    });

    it('calls onRemove when removing an existing widget', () => {
        const onRemove = vi.fn();
        const widgets = [{ id: 'test-widget', enabled: true }];
        render(() => <WidgetDrawer isOpen={true} widgets={widgets} onRemove={onRemove} />);
        
        fireEvent.click(screen.getByText('Remove Widget'));
        expect(onRemove).toHaveBeenCalledWith('test-widget');
    });
});
