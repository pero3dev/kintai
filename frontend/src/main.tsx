import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider, createRouter } from '@tanstack/react-router';
import { routeTree } from './routes';
import { useThemeStore, applyTheme, watchSystemTheme } from './stores/themeStore';
import './i18n';
import './index.css';

// 初期テーマ適用
const initialTheme = useThemeStore.getState().theme;
applyTheme(initialTheme);

// OSのテーマ変更を監視
watchSystemTheme(() => {
  const theme = useThemeStore.getState().theme;
  if (theme === 'system') {
    applyTheme('system');
  }
});

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
    },
  },
});

const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </React.StrictMode>,
);
