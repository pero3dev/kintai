import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { HRDashboardPage } from './HRDashboardPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
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

vi.mock('@/api/client', () => ({
  api: {
    hr: {
      getStats: vi.fn(),
      getRecentActivities: vi.fn(),
    },
  },
}));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

describe('HRDashboardPage', () => {
  beforeEach(() => {
    state.queryData.clear();
  });

  it('renders HR stats and recent activity cards', () => {
    setQueryData(['hr-stats'], {
      total_employees: 42,
      active_employees: 40,
      new_hires_this_month: 3,
      turnover_rate: 4.2,
      open_positions: 5,
      upcoming_reviews: 8,
      training_completion: 76,
      pending_documents: 2,
    });
    setQueryData(['hr-activities'], {
      data: [
        {
          id: 'a1',
          icon: 'person_add',
          message: 'New employee onboarded',
          timestamp: '2026-02-11 09:00',
        },
      ],
    });

    render(<HRDashboardPage />);

    expect(screen.getByText('hr.dashboard.title')).toBeInTheDocument();
    expect(screen.getByText('New employee onboarded')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
  });

  it('shows no-data message when activity feed is empty', () => {
    setQueryData(['hr-stats'], undefined);
    setQueryData(['hr-activities'], { data: [] });

    render(<HRDashboardPage />);

    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });
});
