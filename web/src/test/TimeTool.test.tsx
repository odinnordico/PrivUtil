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
    expect(screen.getByText('Now')).toBeInTheDocument();
  });

  it('handles time conversion and displays all format rows', async () => {
    const mockResponse = TimeResponse.create({
      iso: '2023-01-01T00:00:00Z',
      unix: 1672531200 as number,
      utc: '2023-01-01T00:00:00Z',
      local: '2023-01-01 00:00:00 +0000 UTC',
    });
    vi.mocked(client.timeConvert).mockResolvedValue(mockResponse);

    render(<TimeTool />);

    // Wait for initial "now" conversion
    await waitFor(() => {
      expect(client.timeConvert).toHaveBeenCalled();
    });

    // All format rows should be visible
    expect(screen.getByLabelText('Unix (sec)')).toBeInTheDocument();
    expect(screen.getByLabelText('Unix (ms)')).toBeInTheDocument();
    expect(screen.getByLabelText('ISO 8601')).toBeInTheDocument();
    expect(screen.getByLabelText('UTC')).toBeInTheDocument();
    expect(screen.getByLabelText('Local')).toBeInTheDocument();
  });

  it('converts when user edits a row and presses Enter', async () => {
    const mockResponse = TimeResponse.create({
      iso: '2023-01-01T00:00:00Z',
      unix: 1672531200 as number,
      utc: '2023-01-01T00:00:00Z',
      local: '2023-01-01 00:00:00 +0000 UTC',
    });
    vi.mocked(client.timeConvert).mockResolvedValue(mockResponse);

    render(<TimeTool />);

    await waitFor(() => expect(client.timeConvert).toHaveBeenCalledTimes(1));

    const isoInput = screen.getByLabelText('ISO 8601');
    fireEvent.change(isoInput, { target: { value: '2023-01-01T00:00:00Z' } });
    fireEvent.keyDown(isoInput, { key: 'Enter' });

    await waitFor(() => {
      expect(client.timeConvert).toHaveBeenCalledWith(
        expect.objectContaining({ input: '2023-01-01T00:00:00Z' }),
      );
    });
  });

  it('converts when user clicks Use button', async () => {
    const mockResponse = TimeResponse.create({
      iso: '2023-01-01T00:00:00Z',
      unix: 1672531200 as number,
      utc: '2023-01-01T00:00:00Z',
      local: '2023-01-01 00:00:00 +0000 UTC',
    });
    vi.mocked(client.timeConvert).mockResolvedValue(mockResponse);

    render(<TimeTool />);

    await waitFor(() => expect(client.timeConvert).toHaveBeenCalledTimes(1));

    const unixInput = screen.getByLabelText('Unix (sec)');
    fireEvent.change(unixInput, { target: { value: '1672531200' } });

    const useButtons = screen.getAllByRole('button', { name: 'Use' });
    fireEvent.click(useButtons[0]);

    await waitFor(() => {
      expect(client.timeConvert).toHaveBeenCalledWith(
        expect.objectContaining({ input: '1672531200' }),
      );
    });
  });

  it('shows error message for invalid input format', async () => {
    const invalidResponse = TimeResponse.create({
      iso: 'Invalid input format',
      unix: 0 as number,
      utc: '',
      local: '',
    });
    vi.mocked(client.timeConvert).mockResolvedValueOnce(
      TimeResponse.create({ iso: '2023-01-01T00:00:00Z', unix: 1672531200 as number, utc: '', local: '' }),
    );
    vi.mocked(client.timeConvert).mockResolvedValueOnce(invalidResponse);

    render(<TimeTool />);

    await waitFor(() => expect(client.timeConvert).toHaveBeenCalledTimes(1));

    const isoInput = screen.getByLabelText('ISO 8601');
    fireEvent.change(isoInput, { target: { value: 'not-a-date' } });
    fireEvent.keyDown(isoInput, { key: 'Enter' });

    await waitFor(() => {
      expect(screen.getByText('Invalid input format')).toBeInTheDocument();
    });
  });

  it('converts to now when Now button is clicked', async () => {
    const mockResponse = TimeResponse.create({
      iso: '2023-01-01T00:00:00Z',
      unix: 1672531200 as number,
      utc: '2023-01-01T00:00:00Z',
      local: '2023-01-01 00:00:00 +0000 UTC',
    });
    vi.mocked(client.timeConvert).mockResolvedValue(mockResponse);

    render(<TimeTool />);

    const nowButton = screen.getByText('Now');
    fireEvent.click(nowButton);

    await waitFor(() => {
      expect(client.timeConvert).toHaveBeenCalledWith(
        expect.objectContaining({ input: 'now' }),
      );
    });
  });
});
