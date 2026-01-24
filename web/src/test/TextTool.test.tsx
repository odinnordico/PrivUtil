import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { TextTool } from '../components/TextTool';
import { client } from '../lib/client';
import { TextInspectResponse, TextManipulateResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    textInspect: vi.fn(),
    textManipulate: vi.fn(),
  },
}));

describe('TextTool', () => {
  it('renders correctly', () => {
    render(<TextTool />);
    expect(screen.getByText('Text Tools')).toBeInTheDocument();
  });

  it('handles text inspection', async () => {
    const mockResponse = TextInspectResponse.create({ 
      charCount: 10, 
      wordCount: 2, 
      lineCount: 1 
    });
    vi.mocked(client.textInspect).mockResolvedValue(mockResponse);

    const { container } = render(<TextTool />);
    const input = screen.getByPlaceholderText(/Paste text here/i);
    fireEvent.change(input, { target: { value: 'hello world' } });
    
    await waitFor(() => {
      expect(client.textInspect).toHaveBeenCalled();
      // Check for the values anywhere in the text content
      expect(container.textContent).toMatch(/Chars:.*10/);
      expect(container.textContent).toMatch(/Words:.*2/);
    });
  });

  it('handles text manipulation', async () => {
    const mockResponse = TextManipulateResponse.create({ text: 'WORLD\nHELLO' });
    vi.mocked(client.textManipulate).mockResolvedValue(mockResponse);

    render(<TextTool />);
    const input = screen.getByPlaceholderText(/Paste text here/i);
    fireEvent.change(input, { target: { value: 'hello\nworld' } });
    
    const sortButton = screen.getByRole('button', { name: /Sort Z-A/i });
    fireEvent.click(sortButton);

    await waitFor(() => {
      expect(client.textManipulate).toHaveBeenCalled();
      expect(input).toHaveValue('WORLD\nHELLO');
    });
  });
});
