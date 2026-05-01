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
    const textareas = screen.getAllByRole('textbox');
    expect(textareas.length).toBeGreaterThan(0);
  });

  it('handles conversion on input change', async () => {
    const mockResponse = ConvertResponse.create({ data: 'key: value' });
    vi.mocked(client.convert).mockResolvedValue(mockResponse);

    render(<ConverterTool />);
    const textareas = screen.getAllByRole('textbox');
    const inputArea = textareas[0];
    fireEvent.change(inputArea, { target: { value: '{"key": "value"}' } });

    // Select YAML as target (the second select is "To")
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: DataFormat.YAML.toString() } });

    await waitFor(() => {
      expect(client.convert).toHaveBeenCalled();
    });
  });
});
