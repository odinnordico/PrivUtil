import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PasswordTool } from '../components/PasswordTool';
import { client } from '../lib/client';
import { PasswordResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    generatePassword: vi.fn(),
  },
}));

describe('PasswordTool', () => {
  it('renders correctly', () => {
    render(<PasswordTool />);
    expect(screen.getByText('Password Generator')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /generate/i })).toBeInTheDocument();
  });

  it('handles password generation', async () => {
    const mockResponse = PasswordResponse.create({ passwords: ['safe-password-123'] });
    (client.generatePassword as any).mockResolvedValue(mockResponse);

    render(<PasswordTool />);
    
    const generateButton = screen.getByRole('button', { name: /generate/i });
    fireEvent.click(generateButton);

    await waitFor(() => {
      expect(client.generatePassword).toHaveBeenCalled();
      expect(screen.getByText('safe-password-123')).toBeInTheDocument();
    });
  });
});
