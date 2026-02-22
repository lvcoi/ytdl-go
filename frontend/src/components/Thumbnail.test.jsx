import { describe, it, expect } from 'vitest';
import { render } from '@solidjs/testing-library';
import Thumbnail from './Thumbnail';

describe('Thumbnail component', () => {
  it('renders fallback icon when no src is provided', () => {
    const { container } = render(() => <Thumbnail />);
    expect(container.querySelector('svg')).toBeInTheDocument();
  });

  it('renders image when src is provided', () => {
    const testSrc = 'https://example.com/thumb.jpg';
    const { getByRole } = render(() => <Thumbnail src={testSrc} />);
    const img = getByRole('img');
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src', testSrc);
  });

  it('applies size classes', () => {
    const { container: smContainer } = render(() => <Thumbnail size="sm" />);
    expect(smContainer.firstChild).toHaveClass('w-24');

    const { container: mdContainer } = render(() => <Thumbnail size="md" />);
    expect(mdContainer.firstChild).toHaveClass('w-full');
  });
});
