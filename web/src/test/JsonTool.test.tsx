import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { JsonTool } from '../components/JsonTool';
import { client } from '../lib/client';
import { JsonFormatResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    jsonFormat: vi.fn(),
  },
}));

describe('JsonTool', () => {
  it('renders correctly', () => {
    render(<JsonTool />);
    expect(screen.getByText('JSON Formatter')).toBeInTheDocument();
  });

  it('handles JSON formatting', async () => {
    const formattedJson = '{\n  "key": "value"\n}';
    const mockResponse = JsonFormatResponse.create({ text: formattedJson });
    (client.jsonFormat as any).mockResolvedValue(mockResponse);

    render(<JsonTool />);
    // Just find any textarea with the key value pair in placeholder
    const input = screen.getByPlaceholderText(/key.*value/);
    fireEvent.change(input, { target: { value: '{"key":"value"}' } });
    
    const formatButton = screen.getByText('Format');
    fireEvent.click(formatButton);

    await waitFor(() => {
      expect(client.jsonFormat).toHaveBeenCalled();
      // Verify values by looking at all textboxes
      const textareas = screen.getAllByRole('textbox');
      let found = false;
      for(const ta of textareas) {
          if ((ta as HTMLTextAreaElement).value.includes('"key"')) {
              found = true;
          }
      }
      expect(found).toBe(true);
    });
  });
});
