import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@solidjs/testing-library';
import DirectDownload from './DirectDownload';

describe('DirectDownload component', () => {
    it('renders the input and download button', () => {
        const { getByPlaceholderText, getByRole } = render(() => (
            <DirectDownload onDownload={() => {}} />
        ));
        
        expect(getByPlaceholderText(/Paste URL/i)).toBeInTheDocument();
        expect(getByRole('button', { name: /Download Now/i })).toBeInTheDocument();
    });

    it('updates input value on change', () => {
        const { getByPlaceholderText } = render(() => (
            <DirectDownload onDownload={() => {}} />
        ));
        
        const input = getByPlaceholderText(/Paste URL/i);
        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=123' } });
        expect(input.value).toBe('https://youtube.com/watch?v=123');
    });

    it('calls onDownload with input value when button is clicked', () => {
        const onDownload = vi.fn();
        const { getByPlaceholderText, getByRole } = render(() => (
            <DirectDownload onDownload={onDownload} />
        ));
        
        const input = getByPlaceholderText(/Paste URL/i);
        const button = getByRole('button', { name: /Download Now/i });
        
        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=456' } });
        fireEvent.click(button);
        
        expect(onDownload).toHaveBeenCalledWith('https://youtube.com/watch?v=456');
    });

    it('clears input after successful download trigger', () => {
        const { getByPlaceholderText, getByRole } = render(() => (
            <DirectDownload onDownload={() => {}} />
        ));
        
        const input = getByPlaceholderText(/Paste URL/i);
        const button = getByRole('button', { name: /Download Now/i });
        
        fireEvent.input(input, { target: { value: 'test-url' } });
        fireEvent.click(button);
        
        expect(input.value).toBe('');
    });
});
