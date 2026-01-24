import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SqlTool } from '../components/SqlTool';
import { client } from '../lib/client';
import { SqlResponse } from '../proto/proto/privutil';

// Mock the grpc client
vi.mock('../lib/client', () => ({
  client: {
    sqlFormat: vi.fn(),
  },
}));

describe('SqlTool', () => {
  it('renders correctly', () => {
    render(<SqlTool />);
    expect(screen.getByText('SQL Formatter')).toBeInTheDocument();
  });

  it('handles SQL formatting', async () => {
    const formattedSql = 'SELECT *\nFROM users';
    const mockResponse = SqlResponse.create({ formatted: formattedSql });
    vi.mocked(client.sqlFormat).mockResolvedValue(mockResponse);

    render(<SqlTool />);
    const input = screen.getByPlaceholderText(/SELECT.*FROM/i);
    fireEvent.change(input, { target: { value: 'select * from users' } });
    
    const formatButton = screen.getByRole('button', { name: /Format SQL/i });
    fireEvent.click(formatButton);

    await waitFor(() => {
      expect(client.sqlFormat).toHaveBeenCalled();
      const textareas = screen.getAllByRole('textbox');
      const val = (textareas[1] as HTMLTextAreaElement).value;
      expect(val).toMatch(/SELECT/i);
      expect(val).toMatch(/FROM users/i);
    });
  });
});
