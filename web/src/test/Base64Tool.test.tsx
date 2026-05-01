import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Base64Tool } from '../components/Base64Tool';
import { client } from '../lib/client';
import { Base64Response } from '../proto/proto/privutil';

vi.mock('../lib/client', () => ({
  client: {
    base64Encode: vi.fn(),
    base64Decode: vi.fn(),
  },
}));

// Returns the action button (last match when tab + action share the same label)
function getActionButton(name: RegExp) {
  const all = screen.getAllByRole('button', { name });
  return all[all.length - 1];
}

describe('Base64Tool', () => {
  it('renders correctly with encode tab active', () => {
    render(<Base64Tool />);
    expect(screen.getByText('Base64 Encoder/Decoder')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter text to encode…')).toBeInTheDocument();
  });

  it('handles text encoding', async () => {
    const mockResponse = Base64Response.create({ text: 'aGVsbG8=' });
    vi.mocked(client.base64Encode).mockResolvedValue(mockResponse);

    render(<Base64Tool />);
    fireEvent.change(screen.getByPlaceholderText('Enter text to encode…'), {
      target: { value: 'hello' },
    });
    fireEvent.click(getActionButton(/encode/i));

    await waitFor(() => {
      expect(client.base64Encode).toHaveBeenCalled();
      expect(screen.getByDisplayValue('aGVsbG8=')).toBeInTheDocument();
    });
  });

  it('handles text decoding and displays as text', async () => {
    const helloBytes = new TextEncoder().encode('hello');
    const mockResponse = Base64Response.create({ data: helloBytes, mimeType: 'text/plain; charset=utf-8' });
    vi.mocked(client.base64Decode).mockResolvedValue(mockResponse);

    render(<Base64Tool />);
    // Switch to decode tab (first button named "decode")
    fireEvent.click(screen.getAllByRole('button', { name: /decode/i })[0]);

    fireEvent.change(screen.getByPlaceholderText(/Paste Base64/i), {
      target: { value: 'aGVsbG8=' },
    });
    fireEvent.click(getActionButton(/decode/i));

    await waitFor(() => {
      expect(client.base64Decode).toHaveBeenCalled();
      expect(screen.getByDisplayValue('hello')).toBeInTheDocument();
    });
  });
});
