import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { ExpenseDashboardPage } from './ExpenseDashboardPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: {
    id: 'u1',
    email: 'admin@example.com',
    first_name: 'Admin',
    last_name: 'User',
    role: 'admin' as 'admin' | 'manager' | 'employee',
    is_active: true,
  },
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
  useQuery: (options: { queryKey: unknown[] }) => ({
    data: state.queryData.get(JSON.stringify(options.queryKey)),
  }),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({ user: state.user }),
}));

vi.mock('@/api/client', () => ({
  api: {
    expenses: {
      getStats: vi.fn(),
      getList: vi.fn(),
    },
  },
}));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

describe('ExpenseDashboardPage', () => {
  beforeEach(() => {
    state.queryData.clear();
    state.user = {
      id: 'u1',
      email: 'admin@example.com',
      first_name: 'Admin',
      last_name: 'User',
      role: 'admin',
      is_active: true,
    };
  });

  it('renders stats, recent expense and manager quick action', () => {
    setQueryData(['expense-stats'], {
      total_this_month: 120000,
      pending_count: 3,
      approved_this_month: 60000,
      reimbursed_total: 50000,
    });
    setQueryData(['expenses', 'recent'], {
      data: [
        {
          id: 'e1',
          title: 'Taxi',
          category: 'transportation',
          amount: 1800,
          status: 'pending',
          expense_date: '2026-02-10',
        },
      ],
    });

    render(<ExpenseDashboardPage />);

    expect(screen.getByText('expenses.dashboard.title')).toBeInTheDocument();
    expect(screen.getByText('Taxi')).toBeInTheDocument();
    expect(screen.getByText('expenses.approve.title')).toBeInTheDocument();
  });
});
