import { describe, it, expect } from 'vitest';
import { render } from '@solidjs/testing-library';
import SettingsView from './SettingsView';

describe('SettingsView component', () => {
  it('renders correctly and uses modern theme classes', () => {
    const { getByText, container } = render(() => <SettingsView />);
    expect(getByText('System Settings')).toBeInTheDocument();
    
    // Check for modern layout container
    const mainWrapper = container.firstChild;
    expect(mainWrapper).toHaveClass('glass-vibrant');
  });
});
