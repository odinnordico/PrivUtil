import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CronTool } from '../components/CronTool';
import { client } from '../lib/client';
import { CronResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    cronExplain: vi.fn(),
  },
}));

describe('CronTool', () => {
  it('renders correctly', () => {
    render(<CronTool />);
    expect(screen.getByText(/Cron Expression Tester/i)).toBeInTheDocument();
  });

  it('handles cron explanation', async () => {
    const mockResponse = CronResponse.create({ 
      description: 'Every minute',
      nextRuns: '2023-01-01 00:00:00\n2023-01-01 00:01:00'
    });
    vi.mocked(client.cronExplain).mockResolvedValue(mockResponse);

    render(<CronTool />);
    // Placeholder from output
    const input = screen.getByPlaceholderText('* * * * *');
    fireEvent.change(input, { target: { value: '*/5 * * * *' } });
    
    await waitFor(() => {
      expect(client.cronExplain).toHaveBeenCalled();
      expect(screen.getByText('Every minute')).toBeInTheDocument();
    });
  });
});
