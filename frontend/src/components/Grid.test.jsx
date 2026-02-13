import { describe, it, expect } from 'vitest';
import { render } from '@solidjs/testing-library';
import { Grid, GridItem } from './Grid';

describe('Grid component', () => {
  it('renders children', () => {
    const { getByText } = render(() => <Grid>Test Child</Grid>);
    expect(getByText('Test Child')).toBeInTheDocument();
  });

  it('applies default grid classes', () => {
    const { container } = render(() => <Grid>Test</Grid>);
    const gridDiv = container.firstChild;
    expect(gridDiv).toHaveClass('grid');
    expect(gridDiv).toHaveClass('grid-cols-1');
    expect(gridDiv).toHaveClass('sm:grid-cols-2');
  });

  it('applies custom grid classes', () => {
    const { container } = render(() => <Grid cols="grid-cols-10">Test</Grid>);
    const gridDiv = container.firstChild;
    expect(gridDiv).toHaveClass('grid-cols-10');
    expect(gridDiv).not.toHaveClass('grid-cols-1');
  });
});

describe('GridItem component', () => {
  it('renders children and applies glass classes', () => {
    const { getByText, container } = render(() => <GridItem>Item Content</GridItem>);
    expect(getByText('Item Content')).toBeInTheDocument();
    expect(container.firstChild).toHaveClass('glass');
    expect(container.firstChild).toHaveClass('rounded-2xl');
  });
});
