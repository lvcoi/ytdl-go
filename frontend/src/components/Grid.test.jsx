import { describe, it, expect, afterEach } from 'vitest';
import { render, screen, cleanup } from '@solidjs/testing-library';
import { Grid, GridItem } from './Grid';

describe('Grid component', () => {
    afterEach(cleanup);

    it('renders as a 16-column grid', () => {
        const { container } = render(() => <Grid />);
        const grid = container.firstChild;
        expect(grid).toHaveStyle({ display: 'grid' });
        expect(grid).toHaveStyle({ 'grid-template-columns': 'repeat(16, 1fr)' });
    });

            it('renders grid lines only in edit mode', () => {
        let { unmount } = render(() => <Grid isEditMode={false} />);
        expect(document.querySelector('.dashboard-grid-lines')).toBeNull();
        unmount();

        ({ unmount } = render(() => <Grid isEditMode={true} />));
        expect(document.querySelector('.dashboard-grid-lines')).toBeInTheDocument();
        unmount();
    });
});

describe('GridItem component', () => {
    afterEach(cleanup);

    it('applies inline styles for placement', () => {
        const { container } = render(() => (
            <GridItem x={2} y={3} width={4} height={5} />
        ));
        const item = container.firstChild;
        
        expect(item.style.gridColumn).toBe('3 / span 4');
        expect(item.style.gridRow).toBe('4 / span 5');
    });

    it('shows resize handles only in edit mode', () => {
        let { container, unmount } = render(() => <GridItem isEditMode={false} />);
        expect(container.querySelector('.cursor-nwse-resize')).toBeNull();
        unmount();

        ({ container, unmount } = render(() => <GridItem isEditMode={true} />));
        expect(container.querySelector('.cursor-nwse-resize')).toBeInTheDocument();
        unmount();
    });
});
