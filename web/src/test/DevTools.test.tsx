import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DevTools } from '../components/DevTools';
import { client } from '../lib/client';
import { JwtResponse, RegexResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    jwtDecode: vi.fn(),
    regexTest: vi.fn(),
    jsonToGo: vi.fn(),
  },
}));

describe('DevTools', () => {
  it('renders correctly', () => {
    render(<DevTools />);
    expect(screen.getByText('Developer Utilities')).toBeInTheDocument();
  });

  it('handles JWT decoding', async () => {
    const mockResponse = JwtResponse.create({ header: '{"alg":"HS256"}', payload: '{"sub":"123"}' });
    (client.jwtDecode as any).mockResolvedValue(mockResponse);

    render(<DevTools />);
    // Tab switching if needed, but JWT is default. Select input by exact placeholder.
    const input = screen.getByPlaceholderText('Paste JWT here...');
    fireEvent.change(input, { target: { value: 'header.payload.sig' } });
    
    await waitFor(() => {
      expect(client.jwtDecode).toHaveBeenCalled();
      // Look for text in the pre/code overflow areas
      const headerLabel = screen.getByText('Header');
      expect(headerLabel.parentElement?.textContent).toContain('alg');
    });
  });

  it('handles Regex testing', async () => {
    const mockResponse = RegexResponse.create({ match: true, matches: ['hello'] });
    (client.regexTest as any).mockResolvedValue(mockResponse);

    render(<DevTools />);
    // Switch to Regex Tester tab
    fireEvent.click(screen.getByText('Regex Tester'));
    
    const patternInput = await screen.findByPlaceholderText(/e\.g\./i);
    fireEvent.change(patternInput, { target: { value: 'h.*o' } });
    
    const testInput = screen.getByPlaceholderText(/Test string/i);
    fireEvent.change(testInput, { target: { value: 'hello' } });

    fireEvent.click(screen.getByText('Test Regex'));

    await waitFor(() => {
      expect(client.regexTest).toHaveBeenCalled();
      // Match found text
      expect(screen.getByText(/1.*found/)).toBeInTheDocument();
    });
  });
});
