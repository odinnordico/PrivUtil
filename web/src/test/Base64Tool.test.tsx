import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Base64Tool } from '../components/Base64Tool';
import { client } from '../lib/client';
import { Base64Response } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    base64Encode: vi.fn(),
    base64Decode: vi.fn(),
  },
}));

describe('Base64Tool', () => {
  it('renders correctly', () => {
    render(<Base64Tool />);
    expect(screen.getByText('Base64 Encoder/Decoder')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter text to encode or decode...')).toBeInTheDocument();
  });

  it('handles encoding', async () => {
    const mockResponse = Base64Response.create({ text: 'aGVsbG8=' });
    vi.mocked(client.base64Encode).mockResolvedValue(mockResponse);

    render(<Base64Tool />);
    const input = screen.getByPlaceholderText('Enter text to encode or decode...');
    fireEvent.change(input, { target: { value: 'hello' } });
    
    const encodeButton = screen.getByRole('button', { name: 'Encode' });
    fireEvent.click(encodeButton);

    await waitFor(() => {
      expect(client.base64Encode).toHaveBeenCalled();
      expect(screen.getByDisplayValue('aGVsbG8=')).toBeInTheDocument();
    });
  });

  it('handles decoding', async () => {
    const mockResponse = Base64Response.create({ text: 'hello' });
    vi.mocked(client.base64Decode).mockResolvedValue(mockResponse);

    render(<Base64Tool />);
    const input = screen.getByPlaceholderText('Enter text to encode or decode...');
    fireEvent.change(input, { target: { value: 'aGVsbG8=' } });
    
    const decodeButton = screen.getByRole('button', { name: 'Decode' });
    fireEvent.click(decodeButton);

    await waitFor(() => {
      expect(client.base64Decode).toHaveBeenCalled();
      expect(screen.getByDisplayValue('hello')).toBeInTheDocument();
    });
  });
});
