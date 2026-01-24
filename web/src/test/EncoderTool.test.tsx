import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { EncoderTool } from '../components/EncoderTool';
import { client } from '../lib/client';
import { TextResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    urlEncode: vi.fn(),
    urlDecode: vi.fn(),
    htmlEncode: vi.fn(),
    htmlDecode: vi.fn(),
  },
}));

describe('EncoderTool', () => {
  it('renders correctly', () => {
    render(<EncoderTool />);
    expect(screen.getByText('Encoders / Decoders')).toBeInTheDocument();
  });

  it('handles URL encoding', async () => {
    const mockResponse = TextResponse.create({ text: 'hello%20world' });
    (client.urlEncode as any).mockResolvedValue(mockResponse);

    render(<EncoderTool />);
    // Select input textarea mapping
    const textareas = screen.getAllByRole('textbox');
    fireEvent.change(textareas[0], { target: { value: 'hello world' } });
    
    // Exact button text match
    fireEvent.click(screen.getByText('Encode'));

    await waitFor(() => {
      expect(client.urlEncode).toHaveBeenCalled();
      expect(screen.getByDisplayValue('hello%20world')).toBeInTheDocument();
    });
  });

  it('handles HTML decoding', async () => {
    const mockResponse = TextResponse.create({ text: '<script>' });
    (client.htmlDecode as any).mockResolvedValue(mockResponse);

    render(<EncoderTool />);
    // Switch to HTML mode
    fireEvent.click(screen.getByText('HTML'));
    
    const textareas = screen.getAllByRole('textbox');
    fireEvent.change(textareas[0], { target: { value: '&lt;script&gt;' } });
    
    fireEvent.click(screen.getByText('Decode'));

    await waitFor(() => {
      expect(client.htmlDecode).toHaveBeenCalled();
      expect(screen.getByDisplayValue('<script>')).toBeInTheDocument();
    });
  });
});
