import { beforeEach, describe, expect, it, vi } from 'vitest';
import { applyTheme, useThemeStore, watchSystemTheme } from './themeStore';

type MatchMediaMocks = {
  matchMedia: ReturnType<typeof vi.fn>;
  addEventListener: ReturnType<typeof vi.fn>;
  removeEventListener: ReturnType<typeof vi.fn>;
};

function setupMatchMedia(matches: boolean): MatchMediaMocks {
  const addEventListener = vi.fn();
  const removeEventListener = vi.fn();

  const matchMedia = vi.fn().mockImplementation(() => {
    return {
      matches,
      media: '(prefers-color-scheme: dark)',
      onchange: null,
      addEventListener,
      removeEventListener,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as unknown as MediaQueryList;
  });

  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    configurable: true,
    value: matchMedia,
  });

  return { matchMedia, addEventListener, removeEventListener };
}

describe('themeStore', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.className = '';
    document.documentElement.style.colorScheme = '';
    useThemeStore.setState({ theme: 'system' });
    setupMatchMedia(false);
  });

  describe('store state', () => {
    it('has system as the initial theme', () => {
      expect(useThemeStore.getState().theme).toBe('system');
    });

    it('updates theme via setTheme', () => {
      useThemeStore.getState().setTheme('light');
      expect(useThemeStore.getState().theme).toBe('light');

      useThemeStore.getState().setTheme('dark');
      expect(useThemeStore.getState().theme).toBe('dark');
    });
  });

  describe('applyTheme', () => {
    it('applies dark mode when theme is dark', () => {
      setupMatchMedia(false);

      applyTheme('dark');

      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(document.documentElement.style.colorScheme).toBe('dark');
    });

    it('applies dark mode when theme is system and OS prefers dark', () => {
      setupMatchMedia(true);

      applyTheme('system');

      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(document.documentElement.style.colorScheme).toBe('dark');
    });

    it('applies light mode when theme is system and OS does not prefer dark', () => {
      document.documentElement.classList.add('dark');
      setupMatchMedia(false);

      applyTheme('system');

      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(document.documentElement.style.colorScheme).toBe('light');
    });

    it('applies light mode when theme is light', () => {
      document.documentElement.classList.add('dark');
      setupMatchMedia(true);

      applyTheme('light');

      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(document.documentElement.style.colorScheme).toBe('light');
    });
  });

  describe('watchSystemTheme', () => {
    it('registers and unregisters change listener', () => {
      const { addEventListener, removeEventListener } = setupMatchMedia(false);
      const callback = vi.fn();

      const unwatch = watchSystemTheme(callback);
      expect(addEventListener).toHaveBeenCalledWith('change', callback);

      unwatch();
      expect(removeEventListener).toHaveBeenCalledWith('change', callback);
    });
  });
});
