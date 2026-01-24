import { describe, it, expect } from 'vitest';
import { navItems } from '../lib/nav';

describe('navItems', () => {
  it('contains dashboard as first item', () => {
    expect(navItems[0].path).toBe('/');
    expect(navItems[0].name).toBe('Dashboard');
  });

  it('all items have required properties', () => {
    navItems.forEach((item, index) => {
      expect(item.name, `item ${index} should have name`).toBeTruthy();
      expect(item.path, `item ${index} should have path`).toBeTruthy();
      expect(item.icon, `item ${index} should have icon`).toBeTruthy();
      expect(item.description, `item ${index} should have description`).toBeTruthy();
    });
  });

  it('paths are unique', () => {
    const paths = navItems.map(item => item.path);
    const uniquePaths = new Set(paths);
    expect(uniquePaths.size).toBe(paths.length);
  });

  it('contains expected tools', () => {
    const names = navItems.map(item => item.name);
    expect(names).toContain('Diff Utility');
    expect(names).toContain('Base64 Tool');
    expect(names).toContain('JSON Formatter');
    expect(names).toContain('Color Converter');
  });
});
