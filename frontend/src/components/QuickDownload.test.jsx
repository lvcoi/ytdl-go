import { describe, it, expect, vi } from 'vitest';
import { screen, fireEvent } from '@solidjs/testing-library';
import { renderWithRouter } from '../test/renderWithRouter';
import QuickDownload from './QuickDownload';

describe('QuickDownload', () => {
    it('renders the input and button', () => {
        renderWithRouter(() => <QuickDownload />);
        expect(screen.getByPlaceholderText(/Paste YouTube URL/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Download/i })).toBeInTheDocument();
    });

    it('updates input value on change', () => {
        renderWithRouter(() => <QuickDownload />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL/i);
        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=123' } });
        expect(input).toHaveValue('https://youtube.com/watch?v=123');
    });

    it('calls onDownload with trimmed URL when form is submitted', () => {
        const mockDownload = vi.fn();
        renderWithRouter(() => <QuickDownload onDownload={mockDownload} />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL/i);
        const button = screen.getByRole('button', { name: /Download/i });

        fireEvent.input(input, { target: { value: '  https://youtube.com/watch?v=123  ' } });
        fireEvent.click(button);

        expect(mockDownload).toHaveBeenCalledWith('https://youtube.com/watch?v=123');
    });

    it('calls onDownload with comma-separated URLs', () => {
        const mockDownload = vi.fn();
        renderWithRouter(() => <QuickDownload onDownload={mockDownload} />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL/i);
        const button = screen.getByRole('button', { name: /Download/i });

        const multiUrl = 'https://youtu.be/1, https://youtu.be/2';
        fireEvent.input(input, { target: { value: multiUrl } });
        fireEvent.click(button);

        expect(mockDownload).toHaveBeenCalledWith(multiUrl);
    });

    it('clears input after successful download', () => {
        const mockDownload = vi.fn();
        renderWithRouter(() => <QuickDownload onDownload={mockDownload} />);
        const input = screen.getByPlaceholderText(/Paste YouTube URL/i);
        const button = screen.getByRole('button', { name: /Download/i });

        fireEvent.input(input, { target: { value: 'https://youtube.com/watch?v=123' } });
        fireEvent.click(button);

        expect(input).toHaveValue('');
    });

    it('disables button when input is empty', () => {
        renderWithRouter(() => <QuickDownload />);
        const button = screen.getByRole('button', { name: /Download/i });
        expect(button).toBeDisabled();
    });
});
