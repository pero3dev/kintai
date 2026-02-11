import { beforeEach, describe, expect, it, vi } from 'vitest';

type Theme = 'light' | 'dark' | 'system';

async function loadMain(options?: { initialTheme?: Theme }) {
  const initialTheme = options?.initialTheme ?? 'light';
  let currentTheme: Theme = initialTheme;
  let systemThemeWatcher: (() => void) | undefined;

  const applyTheme = vi.fn();
  const watchSystemTheme = vi.fn((watcher: () => void) => {
    systemThemeWatcher = watcher;
  });
  const createRoot = vi.fn(() => ({ render: vi.fn() }));
  const createRouter = vi.fn(() => ({ mockedRouter: true }));

  vi.doMock('./stores/themeStore', () => ({
    useThemeStore: {
      getState: () => ({ theme: currentTheme }),
    },
    applyTheme,
    watchSystemTheme,
  }));

  vi.doMock('./routes', () => ({
    routeTree: { mockedRouteTree: true },
  }));

  vi.doMock('@tanstack/react-router', () => ({
    createRouter,
    RouterProvider: () => null,
  }));

  vi.doMock('react-dom/client', () => ({
    default: { createRoot },
  }));

  vi.doMock('./i18n', () => ({}));
  vi.doMock('./index.css', () => ({}));

  await import('./main');

  return {
    applyTheme,
    createRoot,
    createRouter,
    watchSystemTheme,
    setTheme: (theme: Theme) => {
      currentTheme = theme;
    },
    runSystemThemeWatcher: () => {
      systemThemeWatcher?.();
    },
  };
}

describe('main.tsx', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    document.body.innerHTML = '<div id="root"></div>';
  });

  it('boots the app and applies the initial theme', async () => {
    const app = await loadMain({ initialTheme: 'dark' });

    expect(app.applyTheme).toHaveBeenCalledWith('dark');
    expect(app.watchSystemTheme).toHaveBeenCalledTimes(1);
    expect(app.createRouter).toHaveBeenCalledWith({ routeTree: { mockedRouteTree: true } });
    expect(app.createRoot).toHaveBeenCalledWith(document.getElementById('root'));
  });

  it('re-applies system theme only when current theme is system', async () => {
    const app = await loadMain({ initialTheme: 'light' });

    app.runSystemThemeWatcher();
    expect(app.applyTheme).toHaveBeenCalledTimes(1);

    app.setTheme('system');
    app.runSystemThemeWatcher();
    expect(app.applyTheme).toHaveBeenNthCalledWith(2, 'system');
  });
});
