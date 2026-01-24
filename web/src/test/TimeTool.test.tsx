import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { TimeTool } from '../components/TimeTool';
import { client } from '../lib/client';
import { TimeResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    timeConvert: vi.fn(),
  },
}));

describe('TimeTool', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders correctly', () => {
    render(<TimeTool />);
    expect(screen.getByText('Time Converter')).toBeInTheDocument();
  });

  it('handles time conversion', async () => {
    const mockResponse = TimeResponse.create({ 
      iso: '2023-01-01T00:00:00Z',
      unix: 1672531200 as any,
      utc: '2023-01-01 00:00:00 UTC',
      local: '2023-01-01 00:00:00'
    });
    (client.timeConvert as any).mockResolvedValue(mockResponse);

    const { container } = render(<TimeTool />);
    const input = screen.getByPlaceholderText(/Unix timestamp/i);
    fireEvent.change(input, { target: { value: 'now' } });
    
    const convertButton = screen.getByText('Convert');
    fireEvent.click(convertButton);

    await waitFor(() => {
      expect(client.timeConvert).toHaveBeenCalled();
      // Values are displayed in the list
      expect(container.textContent).toContain('2023-01-01T00:00:00Z');
    });
  });
});
