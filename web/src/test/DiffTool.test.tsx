import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DiffTool } from '../components/DiffTool';
import { client } from '../lib/client';
import { DiffResponse, DiffFilesResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    diff: vi.fn(),
    diffFiles: vi.fn(),
  },
}));

// jsdom's Blob may lack arrayBuffer(); provide a deterministic stub.
function binFile(name: string, bytes: number[]): File {
  const f = new File([new Uint8Array(bytes)], name);
  Object.defineProperty(f, 'arrayBuffer', { value: () => Promise.resolve(new Uint8Array(bytes).buffer) });
  return f;
}

describe('DiffTool', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders correctly', () => {
    render(<DiffTool />);
    expect(screen.getByText('Diff Viewer')).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Paste original text/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Paste modified text/i)).toBeInTheDocument();
  });

  it('handles diff generation', async () => {
    const mockResponse = DiffResponse.create({ diffHtml: '<ins>added</ins>' });
    vi.mocked(client.diff).mockResolvedValue(mockResponse);

    render(<DiffTool />);
    fireEvent.change(screen.getByPlaceholderText(/Paste original text/i), { target: { value: 'old' } });
    fireEvent.change(screen.getByPlaceholderText(/Paste modified text/i), { target: { value: 'new' } });
    
    const diffButton = screen.getByText('Compare');
    fireEvent.click(diffButton);

    await waitFor(() => {
      expect(client.diff).toHaveBeenCalled();
      const result = screen.getByTestId('diff-output');
      expect(result).toHaveTextContent('added');
    });
  });

  it('compares uploaded binary files by checksum', async () => {
    const mockResponse = DiffFilesResponse.create({
      isText: false,
      checksum1: 'aaaa',
      checksum2: 'bbbb',
      checksumsMatch: false,
      checksumAlgo: 'SHA-256',
      message: 'One or both files are not readable as text (binary). Compared by SHA-256 checksum instead.',
    });
    vi.mocked(client.diffFiles).mockResolvedValue(mockResponse);

    const { container } = render(<DiffTool />);
    const inputs = container.querySelectorAll('input[type="file"]');
    expect(inputs.length).toBe(2);
    fireEvent.change(inputs[0], { target: { files: [binFile('a.bin', [0, 1, 2])] } });
    fireEvent.change(inputs[1], { target: { files: [binFile('b.bin', [3, 4, 5])] } });

    fireEvent.click(screen.getByText('Compare'));

    await waitFor(() => {
      expect(client.diffFiles).toHaveBeenCalled();
      expect(screen.getByText(/checksums do not match/i)).toBeInTheDocument();
      expect(screen.getByText('aaaa')).toBeInTheDocument();
    });
    expect(client.diff).not.toHaveBeenCalled();
  });
});
