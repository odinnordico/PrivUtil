import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GeneratorTool } from '../components/GeneratorTool';
import { client } from '../lib/client';
import { UuidResponse, LoremResponse, HashResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    generateUuid: vi.fn(),
    generateLorem: vi.fn(),
    calculateHash: vi.fn(),
  },
}));

describe('GeneratorTool', () => {
  it('renders correctly', () => {
    render(<GeneratorTool />);
    expect(screen.getByText('Generators')).toBeInTheDocument();
  });

  it('handles UUID generation', async () => {
    const mockResponse = UuidResponse.create({ uuids: ['mock-uuid'] });
    (client.generateUuid as any).mockResolvedValue(mockResponse);

    render(<GeneratorTool />);
    // Select UUIDs tab if not active, but from DOM it seems active by default or has label
    const uuidTab = screen.getByText('UUIDs');
    fireEvent.click(uuidTab);
    
    const generateButton = screen.getByRole('button', { name: /Generate/i });
    fireEvent.click(generateButton);

    await waitFor(() => {
      expect(client.generateUuid).toHaveBeenCalled();
      expect(screen.getByText('mock-uuid')).toBeInTheDocument();
    });
  });

  it('handles Lorem Ipsum generation', async () => {
    const mockResponse = LoremResponse.create({ text: 'mock-lorem' });
    (client.generateLorem as any).mockResolvedValue(mockResponse);

    render(<GeneratorTool />);
    // Switch to Lorem tab
    const loremTab = screen.getByText('Lorem Ipsum');
    fireEvent.click(loremTab);
    
    const generateButton = screen.getByRole('button', { name: /Generate/i });
    fireEvent.click(generateButton);

    await waitFor(() => {
      expect(client.generateLorem).toHaveBeenCalled();
      expect(screen.getByDisplayValue('mock-lorem')).toBeInTheDocument();
    });
  });

  it('handles Hash calculation', async () => {
    const mockResponse = HashResponse.create({ hash: 'mock-hash' });
    (client.calculateHash as any).mockResolvedValue(mockResponse);

    const { container } = render(<GeneratorTool />);
    // Switch to Hash tab
    const hashTab = screen.getByText('Hash Calculator');
    fireEvent.click(hashTab);
    
    const input = await screen.findByPlaceholderText(/Text to hash/i);
    fireEvent.change(input, { target: { value: 'test' } });
    
    // Find precise button from output: "Calculate Hash"
    const calculateButton = screen.getByText('Calculate Hash');
    fireEvent.click(calculateButton);

    await waitFor(() => {
      expect(client.calculateHash).toHaveBeenCalled();
      expect(container.textContent).toContain('mock-hash');
    });
  });
});
