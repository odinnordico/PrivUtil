import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SimilarityTool } from '../components/SimilarityTool';
import { client } from '../lib/client';
import { SimilarityResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    textSimilarity: vi.fn(),
  },
}));

describe('SimilarityTool', () => {
  it('renders correctly', () => {
    render(<SimilarityTool />);
    // Header text check
    expect(screen.getByText(/Similarity Counter/i)).toBeInTheDocument();
  });

  it('handles similarity calculation', async () => {
    const mockResponse = SimilarityResponse.create({ distance: 2, similarity: 0.8 });
    (client.textSimilarity as any).mockResolvedValue(mockResponse);

    const { container } = render(<SimilarityTool />);
    fireEvent.change(screen.getByPlaceholderText(/First text/i), { target: { value: 'apple' } });
    fireEvent.change(screen.getByPlaceholderText(/Second text/i), { target: { value: 'apply' } });
    
    fireEvent.click(screen.getByRole('button', { name: /Calculate/i }));

    await waitFor(() => {
      expect(client.textSimilarity).toHaveBeenCalled();
      expect(container.textContent).toMatch(/2/);
      expect(container.textContent).toMatch(/80/);
    });
  });
});
