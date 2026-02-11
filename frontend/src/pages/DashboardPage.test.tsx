import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { DashboardPage } from './DashboardPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: {
    id: 'u1',
    email: 'manager@example.com',
    first_name: 'Hanako',
    last_name: 'Suzuki',
    role: 'manager' as 'admin' | 'manager' | 'employee',
    is_active: true,
  },
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

vi.mock('@/api/client', () => ({
  api: {
    attendance: { getToday: vi.fn() },
    dashboard: { getStats: vi.fn() },
  },
}));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

describe('DashboardPage', () => {
  beforeEach(() => {
    state.queryData.clear();
    state.user = {
      id: 'u1',
      email: 'manager@example.com',
      first_name: 'Hanako',
      last_name: 'Suzuki',
      role: 'manager',
      is_active: true,
    };
  });

  it('renders stat cards and department table when stats exist', () => {
    setQueryData(['attendance', 'today'], {
      clock_in: '2026-02-11T09:00:00Z',
      clock_out: null,
    });
    setQueryData(['dashboard', 'stats'], {
      today_present_count: 8,
      today_absent_count: 2,
      pending_leaves: 1,
      monthly_overtime: 420,
      weekly_trend: [
        {
          date: '2026-02-09',
          present_count: 8,
          absent_count: 2,
          attendance_rate: 80,
        },
      ],
      department_stats: [
        {
          department_name: 'Engineering',
          total_employees: 10,
          present_today: 8,
          attendance_rate: 0.8,
        },
      ],
    });

    render(<DashboardPage />);

    expect(screen.getByText('dashboard.todayPresent')).toBeInTheDocument();
    expect(screen.getAllByText('Engineering').length).toBeGreaterThan(0);
    expect(screen.getAllByText('80.0%').length).toBeGreaterThan(0);
  });
});
