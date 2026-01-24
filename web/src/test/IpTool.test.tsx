import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { IpTool } from '../components/IpTool';
import { client } from '../lib/client';
import { IpResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    ipCalc: vi.fn(),
  },
}));

describe('IpTool', () => {
  it('renders correctly', () => {
    render(<IpTool />);
    expect(screen.getByText(/IP.*Calculator/i)).toBeInTheDocument();
  });

  it('handles IP calculation', async () => {
    const mockResponse = IpResponse.create({ 
      network: '192.168.1.0',
      broadcast: '192.168.1.255',
      netmask: '255.255.255.0'
    });
    vi.mocked(client.ipCalc).mockResolvedValue(mockResponse);

    const { container } = render(<IpTool />);
    // Exact placeholder from DOM output
    const input = screen.getByPlaceholderText('192.168.1.0/24');
    fireEvent.change(input, { target: { value: '192.168.1.0/24' } });
    
    // Exact button text
    const calculateButton = screen.getByText('Calculate');
    fireEvent.click(calculateButton);

    await waitFor(() => {
      expect(client.ipCalc).toHaveBeenCalled();
      expect(container.textContent).toContain('192.168.1.0');
      expect(container.textContent).toContain('255.255.255.0');
    });
  });
});
