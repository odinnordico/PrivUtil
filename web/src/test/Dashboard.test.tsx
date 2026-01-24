import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { Dashboard } from '../components/Dashboard';

// Wrap component with router
const renderWithRouter = (component: React.ReactElement) => {
  return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('Dashboard', () => {
  it('renders dashboard title', () => {
    renderWithRouter(<Dashboard />);
    expect(screen.getByText('PrivUtil')).toBeInTheDocument();
  });

  it('renders search input', () => {
    renderWithRouter(<Dashboard />);
    expect(screen.getByPlaceholderText('Search tools...')).toBeInTheDocument();
  });

  it('renders tool cards', () => {
    renderWithRouter(<Dashboard />);
    expect(screen.getByText('Diff Utility')).toBeInTheDocument();
    expect(screen.getByText('Base64 Tool')).toBeInTheDocument();
  });

  it('filters tools on search', () => {
    renderWithRouter(<Dashboard />);
    const searchInput = screen.getByPlaceholderText('Search tools...');
    
    fireEvent.change(searchInput, { target: { value: 'json' } });
    
    expect(screen.getByText('JSON Formatter')).toBeInTheDocument();
    expect(screen.queryByText('Diff Utility')).not.toBeInTheDocument();
  });

  it('shows no results message when no match', () => {
    renderWithRouter(<Dashboard />);
    const searchInput = screen.getByPlaceholderText('Search tools...');
    
    fireEvent.change(searchInput, { target: { value: 'xyz123nonexistent' } });
    
    expect(screen.getByText(/No tools found/)).toBeInTheDocument();
  });
});
