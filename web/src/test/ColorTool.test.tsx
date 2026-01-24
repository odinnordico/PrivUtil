import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ColorTool } from '../components/ColorTool';
import { client } from '../lib/client';
import { ColorResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    colorConvert: vi.fn(),
  },
}));

describe('ColorTool', () => {
  it('renders correctly', () => {
    render(<ColorTool />);
    expect(screen.getByText('Color Converter')).toBeInTheDocument();
    // Exact placeholder from output
    expect(screen.getByPlaceholderText('#RRGGBB or rgb(r,g,b)')).toBeInTheDocument();
  });

  it('handles color conversion', async () => {
    const mockResponse = ColorResponse.create({ hex: '#FF0000', rgb: 'rgb(255, 0, 0)', hsl: 'hsl(0, 100%, 50%)' });
    vi.mocked(client.colorConvert).mockResolvedValue(mockResponse);

    const { container } = render(<ColorTool />);
    fireEvent.change(screen.getByPlaceholderText('#RRGGBB or rgb(r,g,b)'), { target: { value: '#ff0000' } });
    
    await waitFor(() => {
      expect(client.colorConvert).toHaveBeenCalled();
      // Values are uppercase in the component display for hex
      expect(container.textContent).toContain('FF0000');
    });
  });
});
