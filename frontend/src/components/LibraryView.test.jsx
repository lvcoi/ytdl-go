import { describe, it, expect } from 'vitest';
import { render } from '@solidjs/testing-library';
import LibraryView from './LibraryView';

describe('LibraryView component', () => {
  const mockProps = {
    downloads: [],
    savedPlaylists: [],
    playlistAssignments: {},
  };

        it('renders initial state correctly', () => {
        const { getAllByText } = render(() => <LibraryView {...mockProps} />);
    // Initial section is 'music', which has title 'Music'
    expect(getAllByText('Music').length).toBeGreaterThan(0);

  });




  it('uses the Media-First vibrant theme classes', () => {
    const { container } = render(() => <LibraryView {...mockProps} />);
    // The main header container should use glass-vibrant or similar modern classes
    // In current implementation it has a long style string but not the 'glass-vibrant' class.
    const h1 = container.querySelector('h1');
    const headerWrapper = h1.closest('.relative');
    expect(headerWrapper).toHaveClass('glass-vibrant');
  });
});
