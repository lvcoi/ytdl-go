import { render, screen, fireEvent } from '@solidjs/testing-library';
import { describe, it, expect, vi } from 'vitest';
import DirectDownload from './DirectDownload';

describe('DirectDownload component', () => {
    it('renders the input and download button', () => {
        render(() => <DirectDownload onDownload={() => {}} />);
        expect(screen.getByPlaceholderText(/Paste YouTube URL here/i)).toBeInTheDocument();
        expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('updates input value on change', () => {
        render(() => <DirectDownload onDownload={() => {}} />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL here/i);
        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=123' } });
        expect(input.value).toBe('https://youtube.com/watch?v=123');
    });

    it('calls onDownload with input value when button is clicked', () => {
        const onDownload = vi.fn();
        render(() => <DirectDownload onDownload={onDownload} />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL here/i);
        const button = screen.getByRole('button');

        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=abc' } });
        fireEvent.click(button);

        expect(onDownload).toHaveBeenCalledWith('https://youtube.com/watch?v=abc');
    });

    it('clears input after successful download trigger', () => {
        render(() => <DirectDownload onDownload={() => {}} />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL here/i);
        const button = screen.getByRole('button');

        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=xyz' } });
        fireEvent.click(button);

        expect(input.value).toBe('');
    });
});
