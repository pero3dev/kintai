import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { LoginPage } from '../LoginPage';
import { HomeDashboardPage } from '../HomeDashboardPage';
import { DashboardPage } from '../DashboardPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'employee' as 'admin' | 'manager' | 'employee',
    is_active: true,
  } as {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: 'admin' | 'manager' | 'employee';
    is_active: boolean;
  } | null,
  language: 'ja',
  availableApps: [] as Array<{
    id: string;
    nameKey: string;
    descriptionKey: string;
    icon: string;
    color: string;
    basePath: string;
    enabled: boolean;
    comingSoon?: boolean;
  }>,
  navigate: vi.fn(),
  setAuth: vi.fn(),
}));

const apiMocks = vi.hoisted(() => ({
  auth: { login: vi.fn() },
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
  useNavigate: () => state.navigate,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, values?: Record<string, unknown>) =>
      values && 'name' in values ? `${key}:${String(values.name)}` : key,
    i18n: {
      language: state.language,
      changeLanguage: vi.fn(),
    },
  }),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: (options: { queryKey: unknown[]; enabled?: boolean; queryFn?: () => unknown }) => {
    if (options.enabled === false) return { data: undefined };
    options.queryFn?.();
    return { data: state.queryData.get(JSON.stringify(options.queryKey)) };
  },
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    user: state.user,
    setAuth: state.setAuth,
  }),
}));

vi.mock('@/config/apps', () => ({
  getAvailableApps: () => state.availableApps,
}));

vi.mock('@/api/client', () => ({
  api: {
    auth: { login: apiMocks.auth.login },
    attendance: { getToday: vi.fn() },
    notifications: { getList: vi.fn(), getUnreadCount: vi.fn() },
    leaves: { getPending: vi.fn() },
    dashboard: { getStats: vi.fn() },
  },
}));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

beforeEach(() => {
  state.queryData.clear();
  state.user = {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'employee',
    is_active: true,
  };
  state.language = 'ja';
  state.availableApps = [];
  state.navigate.mockReset();
  state.setAuth.mockReset();
  apiMocks.auth.login.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe('LoginPage', () => {
  it('handles success and error branches', async () => {
    apiMocks.auth.login.mockResolvedValueOnce({
      user: { id: 'u1' },
      access_token: 'access',
      refresh_token: 'refresh',
    });
    apiMocks.auth.login.mockRejectedValueOnce(new Error('login failed'));
    apiMocks.auth.login.mockRejectedValueOnce('failed');
    render(<LoginPage />);

    fireEvent.change(screen.getByPlaceholderText('user@example.com'), {
      target: { value: 'admin@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('••••••••'), {
      target: { value: 'password123' },
    });
    fireEvent.click(screen.getByRole('button', { name: /auth\.loginButton/i }));
    await waitFor(() => {
      expect(state.setAuth).toHaveBeenCalled();
      expect(state.navigate).toHaveBeenCalledWith({ to: '/' });
    });

    fireEvent.click(screen.getByRole('button', { name: /auth\.loginButton/i }));
    expect(await screen.findByText('login failed')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /auth\.loginButton/i }));
    expect(await screen.findByText('common.error')).toBeInTheDocument();
  });

  it('shows validation errors', async () => {
    render(<LoginPage />);
    fireEvent.click(screen.getByRole('button', { name: /auth\.loginButton/i }));
    expect(await screen.findByText('auth.validation.invalidEmail')).toBeInTheDocument();
    expect(screen.getByText('auth.validation.passwordMin')).toBeInTheDocument();
  });
});

describe('HomeDashboardPage', () => {
  it('covers attendance/app/notification/pending branches', () => {
    state.user = null;
    state.availableApps = [
      {
        id: 'attendance',
        nameKey: 'apps.attendance.name',
        descriptionKey: 'apps.attendance.description',
        icon: 'schedule',
        color: 'bg-blue-500',
        basePath: '/',
        enabled: true,
      },
      {
        id: 'wiki',
        nameKey: 'apps.wiki.name',
        descriptionKey: 'apps.wiki.description',
        icon: 'menu_book',
        color: 'bg-amber-500',
        basePath: '/wiki',
        enabled: false,
        comingSoon: true,
      },
    ];
    setQueryData(['notifications', 'home'], { data: [] });
    setQueryData(['notifications', 'unread-count'], { count: 0 });
    render(<HomeDashboardPage />);
    expect(screen.getByText('home.notClockedIn')).toBeInTheDocument();
    expect(screen.getByText('notifications.empty')).toBeInTheDocument();
    expect(screen.getByText('appSwitcher.comingSoon')).toBeInTheDocument();

    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['attendance', 'today'], { clock_in_time: '2026-02-10T09:00:00Z' });
    setQueryData(['leaves', 'pending'], { data: [{}, {}] });
    const { unmount } = render(<HomeDashboardPage />);
    expect(screen.getByText('home.pendingApprovals')).toBeInTheDocument();
    expect(screen.getByText('attendance.clockOut')).toBeInTheDocument();
    unmount();

    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-02-10T20:00:00'));
    state.language = 'en';
    render(<HomeDashboardPage />);
    expect(screen.getByText(/home\.greetingEvening/)).toBeInTheDocument();
  });

  it('covers clock-out display and recent notification branches', () => {
    state.user = { ...state.user!, role: 'employee' };
    state.availableApps = [
      {
        id: 'attendance',
        nameKey: 'apps.attendance.name',
        descriptionKey: 'apps.attendance.description',
        icon: 'schedule',
        color: 'bg-blue-500',
        basePath: '/',
        enabled: true,
      },
    ];
    setQueryData(['attendance', 'today'], {
      clock_in_time: '2026-02-10T09:00:00Z',
      clock_out_time: '2026-02-10T18:00:00Z',
    });
    setQueryData(['notifications', 'home'], {
      data: [
        { id: 'n1', title: 'Unread', created_at: '2026-02-10T10:00:00Z', is_read: false },
        { id: 'n2', title: 'Read', created_at: '2026-02-10T11:00:00Z', is_read: true },
      ],
    });
    setQueryData(['notifications', 'unread-count'], { count: 2 });

    render(<HomeDashboardPage />);
    expect(screen.queryByRole('link', { name: 'attendance.clockOut' })).not.toBeInTheDocument();
    expect(screen.getByText('Unread')).toBeInTheDocument();
    expect(screen.getByText('Read')).toBeInTheDocument();
  });

  it('covers morning greeting branch', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-02-10T08:00:00'));
    state.user = { ...state.user!, role: 'employee' };
    state.availableApps = [];
    setQueryData(['notifications', 'home'], { data: [] });
    setQueryData(['notifications', 'unread-count'], { count: 0 });
    render(<HomeDashboardPage />);
    expect(screen.getByText(/home\.greetingMorning/)).toBeInTheDocument();
  });
});

describe('DashboardPage', () => {
  it('covers no-stats and detailed-stats branches', () => {
    setQueryData(['attendance', 'today'], {});
    setQueryData(['dashboard', 'stats'], undefined);
    const { unmount } = render(<DashboardPage />);
    expect(screen.getByText('attendance.notClockedIn')).toBeInTheDocument();
    expect(screen.queryByText('dashboard.departmentStatus')).not.toBeInTheDocument();
    unmount();

    setQueryData(['attendance', 'today'], {
      clock_in: '2026-02-10T09:00:00Z',
      clock_out: '2026-02-10T18:00:00Z',
    });
    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['dashboard', 'stats'], {
      today_present_count: 8,
      today_absent_count: 2,
      pending_leaves: 3,
      monthly_overtime: 130,
      weekly_trend: [
        { date: '2026-02-10', present_count: 8, absent_count: 0, attendance_rate: 95 },
        { date: '2026-02-11', present_count: 7, absent_count: 1, attendance_rate: 75 },
        { date: '2026-02-12', present_count: 5, absent_count: 3, attendance_rate: 55 },
        { date: '2026-02-13', present_count: 2, absent_count: 6, attendance_rate: 30 },
      ],
      department_stats: [
        { department_name: 'A', total_employees: 10, present_today: 10, attendance_rate: 0.95 },
        { department_name: 'B', total_employees: 10, present_today: 7, attendance_rate: 0.75 },
        { department_name: 'C', total_employees: 10, present_today: 4, attendance_rate: 0.4 },
      ],
    });
    render(<DashboardPage />);
    expect(screen.getByText('dashboard.departmentStatus')).toBeInTheDocument();
    expect(screen.getByText('attendance.clockedOut')).toBeInTheDocument();
  });

  it('covers clocked-in branch', () => {
    setQueryData(['attendance', 'today'], {
      clock_in: '2026-02-10T09:00:00Z',
    });
    state.user = { ...state.user!, role: 'manager' };
    setQueryData(['dashboard', 'stats'], {
      today_present_count: 5,
      today_absent_count: 0,
      pending_leaves: 1,
      monthly_overtime: 0,
      weekly_trend: [],
      department_stats: [],
    });
    render(<DashboardPage />);
    expect(screen.getByText('attendance.clockedIn')).toBeInTheDocument();
  });

  it('covers monthly goal zero-rate branch', () => {
    setQueryData(['attendance', 'today'], {});
    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['dashboard', 'stats'], {
      today_present_count: 0,
      today_absent_count: 0,
      pending_leaves: 0,
      monthly_overtime: 0,
      weekly_trend: [],
      department_stats: [],
    });
    render(<DashboardPage />);
    expect(screen.getByText('0% dashboard.achieved')).toBeInTheDocument();
  });
});
