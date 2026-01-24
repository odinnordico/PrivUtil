import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CertTool } from '../components/CertTool';
import { client } from '../lib/client';
import { CertResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    certParse: vi.fn(),
  },
}));

describe('CertTool', () => {
  it('renders correctly', () => {
    render(<CertTool />);
    expect(screen.getByText(/Certificate Inspector/i)).toBeInTheDocument();
  });

  it('handles certificate parsing on change', async () => {
    const mockResponse = CertResponse.create({ subject: 'CN=example.com', issuer: 'CN=Root CA' });
    (client.certParse as any).mockResolvedValue(mockResponse);

    const { container } = render(<CertTool />);
    const input = screen.getByPlaceholderText(/BEGIN CERTIFICATE/);
    fireEvent.change(input, { target: { value: 'PEM-DATA' } });
    
    await waitFor(() => {
      expect(client.certParse).toHaveBeenCalled();
      expect(container.textContent).toContain('CN=example.com');
      expect(container.textContent).toContain('CN=Root CA');
    });
  });
});
