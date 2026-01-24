import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DiffTool } from '../components/DiffTool';
import { client } from '../lib/client';
import { DiffResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    diff: vi.fn(),
  },
}));

describe('DiffTool', () => {
  it('renders correctly', () => {
    render(<DiffTool />);
    expect(screen.getByText('Diff Viewer')).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Paste original text/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Paste modified text/i)).toBeInTheDocument();
  });

  it('handles diff generation', async () => {
    const mockResponse = DiffResponse.create({ diffHtml: '<ins>added</ins>' });
    vi.mocked(client.diff).mockResolvedValue(mockResponse);

    render(<DiffTool />);
    fireEvent.change(screen.getByPlaceholderText(/Paste original text/i), { target: { value: 'old' } });
    fireEvent.change(screen.getByPlaceholderText(/Paste modified text/i), { target: { value: 'new' } });
    
    const diffButton = screen.getByText('Compare');
    fireEvent.click(diffButton);

    await waitFor(() => {
      expect(client.diff).toHaveBeenCalled();
      const result = screen.getByTestId('diff-output');
      expect(result).toHaveTextContent('added');
    });
  });
});
