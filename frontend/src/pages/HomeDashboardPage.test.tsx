import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { HomeDashboardPage } from './HomeDashboardPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'manager' as 'admin' | 'manager' | 'employee',
    is_active: true,
  },
  availableApps: [
    {
      id: 'attendance',
      nameKey: 'nav.attendance',
      descriptionKey: 'home.welcomeMessage',
      icon: 'schedule',
      color: 'bg-indigo-500',
      basePath: '/attendance',
      enabled: true,
    },
    {
      id: 'expense',
      nameKey: 'nav.expenses',
      descriptionKey: 'home.welcomeMessage',
      icon: 'receipt_long',
      color: 'bg-emerald-500',
      basePath: '/expenses',
      enabled: true,
      comingSoon: true,
    },
  ],
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    to,
    children,
    ...props
  }: {
    to: string;
    children: ReactNode;
    [key: string]: unknown;
  }) => (
    <a href={to} data-to={to} {...props}>
      {children}
    </a>
  ),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: {
      language: 'ja',
      changeLanguage: vi.fn(),
    },
  }),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: (options: { queryKey: unknown[]; enabled?: boolean }) => {
    if (options.enabled === false) return { data: undefined };
    return { data: state.queryData.get(JSON.stringify(options.queryKey)) };
  },
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({ user: state.user }),
}));

vi.mock('@/config/apps', () => ({
  getAvailableApps: () => state.availableApps,
}));

vi.mock('@/api/client', () => ({
  api: {
    attendance: { getToday: vi.fn() },
    notifications: { getList: vi.fn(), getUnreadCount: vi.fn() },
    leaves: { getPending: vi.fn() },
  },
}));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

describe('HomeDashboardPage', () => {
  beforeEach(() => {
    state.queryData.clear();
    state.user = {
      id: 'u1',
      email: 'user@example.com',
      first_name: 'Taro',
      last_name: 'Yamada',
      role: 'manager',
      is_active: true,
    };
  });

  it('renders unread and pending summary with app shortcuts', () => {
    setQueryData(['attendance', 'today'], { clock_in_time: '2026-02-11T09:00:00Z' });
    setQueryData(['notifications', 'home'], { data: [] });
    setQueryData(['notifications', 'unread-count'], { count: 3 });
    setQueryData(['leaves', 'pending'], { data: [{ id: 'l1' }, { id: 'l2' }] });

    render(<HomeDashboardPage />);

    expect(screen.getByText('home.apps')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('appSwitcher.comingSoon')).toBeInTheDocument();
  });
});
