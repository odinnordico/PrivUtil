import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StringTool } from '../components/StringTool';
import { client } from '../lib/client';
import { CaseResponse, EscapeResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    caseConvert: vi.fn(),
    stringEscape: vi.fn(),
  },
}));

describe('StringTool', () => {
  it('renders correctly', () => {
    render(<StringTool />);
    expect(screen.getByText('String Utilities')).toBeInTheDocument();
  });

  it('handles case conversion', async () => {
    const mockResponse = CaseResponse.create({ camel: 'helloWorld', snake: 'hello_world' });
    vi.mocked(client.caseConvert).mockResolvedValue(mockResponse);

    render(<StringTool />);
    const input = screen.getByPlaceholderText(/Type variable name/i);
    fireEvent.change(input, { target: { value: 'hello world' } });
    
    await waitFor(() => {
      expect(client.caseConvert).toHaveBeenCalled();
      expect(screen.getByDisplayValue('helloWorld')).toBeInTheDocument();
    });
  });

  it('handles string escaping', async () => {
    const mockResponse = EscapeResponse.create({ result: 'escaped-text' });
    vi.mocked(client.stringEscape).mockResolvedValue(mockResponse);

    render(<StringTool />);
    // Switch tab by name
    const escaperTab = screen.getByRole('button', { name: /String Escaper/i });
    fireEvent.click(escaperTab);
    
    const input = await screen.findByPlaceholderText(/Input text/i);
    fireEvent.change(input, { target: { value: 'text' } });
    
    const escapeButton = screen.getByRole('button', { name: 'Escape' });
    fireEvent.click(escapeButton);

    await waitFor(() => {
      expect(client.stringEscape).toHaveBeenCalled();
      expect(screen.getByDisplayValue('escaped-text')).toBeInTheDocument();
    });
  });
});
