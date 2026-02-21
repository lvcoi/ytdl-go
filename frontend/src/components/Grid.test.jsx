import { render, screen } from '@solidjs/testing-library';
import { describe, it, expect } from 'vitest';
import { Grid, GridItem } from './Grid';

describe('Grid component', () => {
    it('renders children', () => {
        render(() => <Grid><div>Child</div></Grid>);
        expect(screen.getByText('Child')).toBeInTheDocument();
    });

    it('applies default grid classes', () => {
        const { container } = render(() => <Grid />);
        expect(container.querySelector('.grid')).toHaveClass('grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6');
    });

    it('applies custom grid classes', () => {
        const { container } = render(() => <Grid class="custom-class" />);
        expect(container.firstChild).toHaveClass('custom-class');
    });
});

describe('GridItem component', () => {
    it('renders children and applies default span', () => {
        const { container } = render(() => <GridItem><div>Item</div></GridItem>);
        expect(screen.getByText('Item')).toBeInTheDocument();
        expect(container.firstChild).toHaveClass('lg:col-span-1');
    });

    it('applies correct span class for span=2', () => {
        const { container } = render(() => <GridItem span={2} />);
        expect(container.firstChild).toHaveClass('lg:col-span-2');
    });

    it('applies correct span class for span=4', () => {
        const { container } = render(() => <GridItem span={4} />);
        expect(container.firstChild).toHaveClass('lg:col-span-4');
    });

    it('applies custom classes', () => {
        const { container } = render(() => <GridItem class="custom-item" />);
        expect(container.firstChild).toHaveClass('custom-item');
    });
});
