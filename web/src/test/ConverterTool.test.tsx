import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ConverterTool } from '../components/ConverterTool';
import { client } from '../lib/client';
import { ConvertResponse, DataFormat } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    convert: vi.fn(),
  },
}));

describe('ConverterTool', () => {
  it('renders correctly', () => {
    render(<ConverterTool />);
    expect(screen.getByText('Universal Converter')).toBeInTheDocument();
    const input = screen.getByPlaceholderText(/Paste.*here.../);
    expect(input).toBeInTheDocument();
  });

  it('handles conversion', async () => {
    const mockResponse = ConvertResponse.create({ data: 'key: value' });
    (client.convert as any).mockResolvedValue(mockResponse);

    render(<ConverterTool />);
    const input = screen.getByPlaceholderText(/Paste.*here.../);
    fireEvent.change(input, { target: { value: '{"key": "value"}' } });
    
    // Select YAML as target (the second select is "To")
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: DataFormat.YAML.toString() } });
    
    const convertButton = screen.getByText('Convert');
    fireEvent.click(convertButton);

    await waitFor(() => {
      expect(client.convert).toHaveBeenCalled();
      expect(screen.getByDisplayValue('key: value')).toBeInTheDocument();
    });
  });
});
