import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Dashboard } from '../components/Dashboard';

const renderWithSearch = (searchQuery: string = '') => {
  return render(
    <MemoryRouter initialEntries={[`/?q=${searchQuery}`]}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
      </Routes>
    </MemoryRouter>
  );
};

describe('Dashboard', () => {
  it('renders dashboard title', () => {
    renderWithSearch('');
    expect(screen.getByText('PrivUtil')).toBeInTheDocument();
  });

  it('renders tool cards', () => {
    renderWithSearch('');
    expect(screen.getByText('Diff Utility')).toBeInTheDocument();
    expect(screen.getByText('Base64 Tool')).toBeInTheDocument();
  });

  it('filters tools on search', () => {
    renderWithSearch('json');
    expect(screen.getByText('JSON Formatter')).toBeInTheDocument();
    expect(screen.queryByText('Diff Utility')).not.toBeInTheDocument();
  });

  it('shows no results message when no match', () => {
    renderWithSearch('xyz123nonexistent');
    expect(screen.getByText(/No tools found/)).toBeInTheDocument();
  });
});
