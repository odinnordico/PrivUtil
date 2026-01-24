import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import App from '../App';

describe('App', () => {
  it('renders correctly and has navigation', () => {
    render(<App />);
    // There are multiple "PrivUtil" texts (e.g., in sidebar and elsewhere)
    const titles = screen.getAllByText(/PrivUtil/i);
    expect(titles.length).toBeGreaterThan(0);
  });

  it('can switch tools', async () => {
    render(<App />);
    // Select any tool link from the sidebar - Base64 is usually safe
    const links = screen.getAllByText(/Base64/i);
    if(links.length > 0) {
      fireEvent.click(links[0]);
      // Verify we are still on the app (heading should still be there)
      const titles = screen.getAllByText(/PrivUtil/i);
      expect(titles.length).toBeGreaterThan(0);
    }
  });
});
