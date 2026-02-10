import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { AttendancePage } from '../AttendancePage';
import { LeavesPage } from '../LeavesPage';
import { OvertimePage } from '../OvertimePage';
import { CorrectionsPage } from '../CorrectionsPage';
import { NotificationsPage } from '../NotificationsPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'employee' as 'admin' | 'manager' | 'employee',
    is_active: true,
  },
  mutationPending: false,
  invalidateQueries: vi.fn(),
}));

const apiMocks = vi.hoisted(() => ({
  attendance: { clockIn: vi.fn(), clockOut: vi.fn() },
  leaves: { create: vi.fn(), approve: vi.fn() },
  overtime: { create: vi.fn(), approve: vi.fn() },
  corrections: { create: vi.fn(), approve: vi.fn() },
  notifications: {
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => (key === 'common.unknown' ? '' : key),
    i18n: { language: 'ja', changeLanguage: vi.fn() },
  }),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: (options: { queryKey: unknown[]; enabled?: boolean; queryFn?: () => unknown }) => {
    if (options.enabled === false) return { data: undefined };
    options.queryFn?.();
    return { data: state.queryData.get(JSON.stringify(options.queryKey)) };
  },
  useMutation: (options: {
    mutationFn: (vars?: unknown) => unknown;
    onSuccess?: () => void;
  }) => ({
    mutate: (vars?: unknown) => {
      Promise.resolve(options.mutationFn(vars)).then(() => options.onSuccess?.());
    },
    isPending: state.mutationPending,
  }),
  useQueryClient: () => ({ invalidateQueries: state.invalidateQueries }),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({ user: state.user }),
}));

vi.mock('@/components/ui/Pagination', () => ({
  Pagination: () => <div data-testid="pagination" />,
}));

vi.mock('@/api/client', () => ({
  api: {
    attendance: {
      getToday: vi.fn(),
      getList: vi.fn(),
      getSummary: vi.fn(),
      clockIn: apiMocks.attendance.clockIn,
      clockOut: apiMocks.attendance.clockOut,
    },
    leaves: {
      getList: vi.fn(),
      getPending: vi.fn(),
      create: apiMocks.leaves.create,
      approve: apiMocks.leaves.approve,
    },
    overtime: {
      getList: vi.fn(),
      getPending: vi.fn(),
      getAlerts: vi.fn(),
      create: apiMocks.overtime.create,
      approve: apiMocks.overtime.approve,
    },
    corrections: {
      getList: vi.fn(),
      getPending: vi.fn(),
      create: apiMocks.corrections.create,
      approve: apiMocks.corrections.approve,
    },
    notifications: {
      getList: vi.fn(),
      getUnreadCount: vi.fn(),
      markAsRead: apiMocks.notifications.markAsRead,
      markAllAsRead: apiMocks.notifications.markAllAsRead,
      delete: apiMocks.notifications.delete,
    },
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
  state.invalidateQueries.mockReset();
  apiMocks.attendance.clockIn.mockReset();
  apiMocks.attendance.clockOut.mockReset();
  apiMocks.leaves.create.mockReset();
  apiMocks.leaves.approve.mockReset();
  apiMocks.overtime.create.mockReset();
  apiMocks.overtime.approve.mockReset();
  apiMocks.corrections.create.mockReset();
  apiMocks.corrections.approve.mockReset();
  apiMocks.notifications.markAsRead.mockReset();
  apiMocks.notifications.markAllAsRead.mockReset();
  apiMocks.notifications.delete.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe('AttendancePage', () => {
  it('covers status/summary/pagination and mutation branches', async () => {
    setQueryData(['attendance', 'today'], { clock_in: '2026-02-10T09:00:00Z' });
    setQueryData(['attendance', 'summary'], {
      total_work_days: 10,
      total_work_minutes: 600,
      total_overtime_minutes: 120,
      average_work_minutes: 300,
    });
    setQueryData(['attendance', 'list', 1, 20], {
      data: [{ id: 'r1', date: '2026-02-10', status: 'present', work_minutes: 60, overtime_minutes: 30 }],
      total_pages: 2,
      total: 1,
    });
    apiMocks.attendance.clockOut.mockResolvedValue({});

    render(<AttendancePage />);
    expect(screen.getByText('attendance.totalWorkDays')).toBeInTheDocument();
    expect(screen.getByTestId('pagination')).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText('attendance.note'), {
      target: { value: 'note' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'attendance.clockOut' }));
    await waitFor(() => {
      expect(apiMocks.attendance.clockOut).toHaveBeenCalledWith({ note: 'note' });
    });
  });

  it('covers timer update effect branch', () => {
    vi.useFakeTimers();
    setQueryData(['attendance', 'today'], {});
    setQueryData(['attendance', 'summary'], undefined);
    setQueryData(['attendance', 'list', 1, 20], { data: [], total_pages: 0, total: 0 });
    render(<AttendancePage />);
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText('attendance.todayStatus')).toBeInTheDocument();
  });

  it('covers clock-in and status badge variants', async () => {
    setQueryData(['attendance', 'today'], {
      clock_in: '2026-02-10T09:00:00Z',
      clock_out: '2026-02-10T18:00:00Z',
    });
    setQueryData(['attendance', 'summary'], undefined);
    setQueryData(['attendance', 'list', 1, 20], {
      data: [
        { id: 'r2', date: '2026-02-11', status: 'absent', work_minutes: 0, overtime_minutes: 0 },
        { id: 'r3', date: '2026-02-12', status: 'leave', work_minutes: 30, overtime_minutes: 0 },
        { id: 'r4', date: '2026-02-13', status: 'other', work_minutes: 15, overtime_minutes: 15 },
      ],
      total_pages: 0,
      total: 3,
    });
    const { unmount } = render(<AttendancePage />);
    expect(screen.getByRole('button', { name: 'attendance.clockOut' })).toBeInTheDocument();
    expect(screen.getAllByText('absent').length).toBeGreaterThan(0);
    expect(screen.getAllByText('leave').length).toBeGreaterThan(0);
    expect(screen.getAllByText('other').length).toBeGreaterThan(0);
    unmount();

    setQueryData(['attendance', 'today'], {});
    apiMocks.attendance.clockIn.mockResolvedValue({});
    render(<AttendancePage />);
    fireEvent.change(screen.getByPlaceholderText('attendance.note'), {
      target: { value: 'start work' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'attendance.clockIn' }));
    await waitFor(() => {
      expect(apiMocks.attendance.clockIn).toHaveBeenCalledWith({ note: 'start work' });
    });
  });
});

describe('LeavesPage', () => {
  it('covers employee and admin branches', async () => {
    setQueryData(['leaves', 'my', 1, 20], { data: [], total_pages: 0, total: 0 });
    render(<LeavesPage />);
    fireEvent.click(screen.getByRole('button', { name: 'leaves.newRequest' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    expect(await screen.findByText('leaves.validation.startDateRequired')).toBeInTheDocument();
    expect(screen.queryByText('leaves.pending')).not.toBeInTheDocument();

    state.user = { ...state.user, role: 'admin' };
    setQueryData(['leaves', 'my', 1, 20], { data: [{ id: 'm1', leave_type: 'paid', start_date: 'x', end_date: 'x', status: 'unknown' }], total_pages: 1, total: 1 });
    setQueryData(['leaves', 'pending', 1], {
      data: [{ id: 'p1', user: { last_name: 'A', first_name: 'B' }, start_date: 'x', end_date: 'x', leave_type: 'paid', reason: 'urgent' }],
      total_pages: 2,
      total: 1,
    });
    apiMocks.leaves.approve.mockResolvedValue({});
    const { unmount } = render(<LeavesPage />);
    fireEvent.click(screen.getByText('leaves.approve'));
    fireEvent.click(screen.getByText('leaves.reject'));
    await waitFor(() => {
      expect(apiMocks.leaves.approve).toHaveBeenCalled();
    });
    unmount();
  });

  it('covers create success and status/reason display branches', async () => {
    setQueryData(['leaves', 'my', 1, 20], {
      data: [
        { id: 'm1', leave_type: 'paid', start_date: '2026-02-01', end_date: '2026-02-01', status: 'pending', reason: 'Vacation' },
        { id: 'm2', leave_type: 'sick', start_date: '2026-02-02', end_date: '2026-02-02', status: 'approved', reason: '' },
        { id: 'm3', leave_type: 'special', start_date: '2026-02-03', end_date: '2026-02-03', status: 'rejected' },
      ],
      total_pages: 1,
      total: 3,
    });
    apiMocks.leaves.create.mockResolvedValue({});

    render(<LeavesPage />);
    fireEvent.click(screen.getByRole('button', { name: 'leaves.newRequest' }));
    fireEvent.change(document.querySelector('input[name="start_date"]') as HTMLInputElement, {
      target: { value: '2026-02-10' },
    });
    fireEvent.change(document.querySelector('input[name="end_date"]') as HTMLInputElement, {
      target: { value: '2026-02-10' },
    });
    fireEvent.change(document.querySelector('input[name="reason"]') as HTMLInputElement, {
      target: { value: 'Family event' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    await waitFor(() => {
      expect(apiMocks.leaves.create).toHaveBeenCalled();
    });
    expect(screen.getAllByText('Vacation').length).toBeGreaterThan(0);
    expect(screen.getAllByText('leaves.pending').length).toBeGreaterThan(0);
    expect(screen.getAllByText('leaves.approved').length).toBeGreaterThan(0);
    expect(screen.getAllByText('leaves.rejected').length).toBeGreaterThan(0);
  });
});

describe('OvertimePage', () => {
  it('covers alerts/admin/pending/no-data branches', async () => {
    setQueryData(['overtime', 'my', 1, 20], { data: [], total_pages: 0, total: 0 });
    render(<OvertimePage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();

    state.user = { ...state.user, role: 'admin' };
    setQueryData(['overtime', 'my', 1, 20], {
      data: [{ id: 'm1', date: 'x', planned_minutes: 60, reason: 'x', status: 'unknown' }],
      total_pages: 1,
      total: 1,
    });
    setQueryData(['overtime', 'pending', 1], {
      data: [{ id: 'p1', user: { last_name: 'A', first_name: 'B' }, date: 'x', planned_minutes: 120, reason: 'r' }],
      total_pages: 2,
      total: 1,
    });
    setQueryData(['overtime', 'alerts'], [{
      user_id: 'u1',
      user_name: 'A B',
      monthly_overtime_hours: 46,
      monthly_limit_hours: 45,
      yearly_overtime_hours: 361,
      yearly_limit_hours: 360,
      is_monthly_exceeded: true,
      is_yearly_exceeded: true,
    }]);
    apiMocks.overtime.approve.mockResolvedValue({});
    const { unmount } = render(<OvertimePage />);
    fireEvent.click(screen.getByText('common.approve'));
    fireEvent.click(screen.getByText('common.reject'));
    await waitFor(() => {
      expect(apiMocks.overtime.approve).toHaveBeenCalled();
    });
    unmount();
  });

  it('covers create form and overtime status branches', async () => {
    state.user = { ...state.user, role: 'admin' };
    setQueryData(['overtime', 'my', 1, 20], {
      data: [
        { id: 'm1', date: '2026-02-10', planned_minutes: 60, reason: 'r1', status: 'pending' },
        { id: 'm2', date: '2026-02-11', planned_minutes: 60, reason: 'r2', status: 'approved' },
        { id: 'm3', date: '2026-02-12', planned_minutes: 60, reason: 'r3', status: 'rejected' },
        { id: 'm4', date: '2026-02-13', planned_minutes: 60, reason: 'r4', status: 'unknown' },
      ],
      total_pages: 1,
      total: 4,
    });
    setQueryData(['overtime', 'pending', 1], { data: [], total_pages: 1, total: 0 });
    setQueryData(['overtime', 'alerts'], [{
      user_id: 'u2',
      user_name: 'No Alert',
      monthly_overtime_hours: 10,
      monthly_limit_hours: 45,
      yearly_overtime_hours: 100,
      yearly_limit_hours: 360,
      is_monthly_exceeded: false,
      is_yearly_exceeded: false,
    }]);
    apiMocks.overtime.create.mockResolvedValue({});

    render(<OvertimePage />);
    fireEvent.click(screen.getByRole('button', { name: 'overtime.newRequest' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    expect(await screen.findByText('overtime.validation.dateRequired')).toBeInTheDocument();
    expect(screen.getByText('overtime.validation.minutesRequired')).toBeInTheDocument();
    expect(screen.getByText('overtime.validation.reasonRequired')).toBeInTheDocument();
    fireEvent.change(document.querySelector('input[name="date"]') as HTMLInputElement, {
      target: { value: '2026-02-15' },
    });
    fireEvent.change(document.querySelector('input[name="planned_minutes"]') as HTMLInputElement, {
      target: { value: '90' },
    });
    fireEvent.change(document.querySelector('input[name="reason"]') as HTMLInputElement, {
      target: { value: 'Release task' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    await waitFor(() => {
      expect(apiMocks.overtime.create).toHaveBeenCalled();
    });
    expect(screen.getAllByText('common.pending').length).toBeGreaterThan(0);
    expect(screen.getAllByText('common.approved').length).toBeGreaterThan(0);
    expect(screen.getAllByText('common.rejected').length).toBeGreaterThan(0);
    expect(screen.getAllByText('unknown').length).toBeGreaterThan(0);
    expect(screen.queryByText('common.monthlyExceeded')).not.toBeInTheDocument();
    expect(screen.queryByText('common.yearlyExceeded')).not.toBeInTheDocument();
  });
});

describe('CorrectionsPage', () => {
  it('covers form/admin/pending/no-data branches', async () => {
    setQueryData(['corrections', 'my', 1, 20], { data: [], total_pages: 0, total: 0 });
    render(<CorrectionsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'corrections.newRequest' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    expect(await screen.findByText('corrections.validation.dateRequired')).toBeInTheDocument();

    state.user = { ...state.user, role: 'manager' };
    setQueryData(['corrections', 'my', 1, 20], { data: [{ id: 'm1', date: 'x', reason: 'r', status: 'unknown' }], total_pages: 1, total: 1 });
    setQueryData(['corrections', 'pending', 1], {
      data: [{ id: 'p1', user: { last_name: 'A', first_name: 'B' }, date: 'x', reason: 'r' }],
      total_pages: 2,
      total: 1,
    });
    apiMocks.corrections.approve.mockResolvedValue({});
    const { unmount } = render(<CorrectionsPage />);
    fireEvent.click(screen.getByText('common.approve'));
    fireEvent.click(screen.getByText('common.reject'));
    await waitFor(() => {
      expect(apiMocks.corrections.approve).toHaveBeenCalled();
    });
    unmount();
  });

  it('covers create success and corrected-time/status branches', async () => {
    state.user = { ...state.user, role: 'manager' };
    setQueryData(['corrections', 'my', 1, 20], {
      data: [
        { id: 'm1', date: '2026-02-10', reason: 'r1', status: 'pending', corrected_clock_in: '2026-02-10T09:00:00Z', corrected_clock_out: '2026-02-10T18:00:00Z' },
        { id: 'm2', date: '2026-02-11', reason: 'r2', status: 'approved', corrected_clock_in: null, corrected_clock_out: null },
        { id: 'm3', date: '2026-02-12', reason: 'r3', status: 'rejected', corrected_clock_in: null, corrected_clock_out: null },
        { id: 'm4', date: '2026-02-13', reason: 'r4', status: 'unknown', corrected_clock_in: null, corrected_clock_out: null },
      ],
      total_pages: 1,
      total: 4,
    });
    setQueryData(['corrections', 'pending', 1], {
      data: [{
        id: 'p1',
        user: { last_name: 'A', first_name: 'B' },
        date: '2026-02-12',
        reason: 'adjust',
        corrected_clock_in: '2026-02-12T10:00:00Z',
        corrected_clock_out: '2026-02-12T19:00:00Z',
      }],
      total_pages: 1,
      total: 1,
    });
    apiMocks.corrections.create.mockResolvedValue({});

    render(<CorrectionsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'corrections.newRequest' }));
    fireEvent.change(document.querySelector('input[name="date"]') as HTMLInputElement, {
      target: { value: '2026-02-20' },
    });
    fireEvent.change(document.querySelector('input[name="reason"]') as HTMLInputElement, {
      target: { value: 'Missed punch' },
    });
    fireEvent.change(document.querySelector('input[name="corrected_clock_in"]') as HTMLInputElement, {
      target: { value: '09:30' },
    });
    fireEvent.change(document.querySelector('input[name="corrected_clock_out"]') as HTMLInputElement, {
      target: { value: '18:30' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    await waitFor(() => {
      expect(apiMocks.corrections.create).toHaveBeenCalled();
    });
    expect(screen.getAllByText('common.pending').length).toBeGreaterThan(0);
    expect(screen.getAllByText('common.approved').length).toBeGreaterThan(0);
    expect(screen.getAllByText('common.rejected').length).toBeGreaterThan(0);
    expect(screen.getAllByText('unknown').length).toBeGreaterThan(0);
  });
});

describe('NotificationsPage', () => {
  it('covers unread/read/empty branches and actions', async () => {
    setQueryData(['notifications', 1, 20], {
      data: [
        { id: 'n1', type: 'leave_approved', title: 'Leave', message: 'ok', created_at: '2026-02-10T10:00:00Z', is_read: false },
        { id: 'n2', type: 'unknown', title: 'Unknown', message: 'msg', created_at: '2026-02-10T11:00:00Z', is_read: true },
      ],
      total_pages: 1,
      total: 2,
    });
    setQueryData(['notifications', 'unread-count'], { unread: 1 });
    apiMocks.notifications.markAllAsRead.mockResolvedValue({});
    apiMocks.notifications.markAsRead.mockResolvedValue({});
    apiMocks.notifications.delete.mockResolvedValue({});
    const { unmount } = render(<NotificationsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'notifications.markAllRead' }));
    fireEvent.click(screen.getByTitle('common.markAsRead'));
    fireEvent.click(screen.getAllByTitle('common.delete')[0]);
    await waitFor(() => {
      expect(apiMocks.notifications.markAllAsRead).toHaveBeenCalled();
      expect(apiMocks.notifications.markAsRead).toHaveBeenCalled();
      expect(apiMocks.notifications.delete).toHaveBeenCalled();
    });
    unmount();

    setQueryData(['notifications', 1, 20], { data: [], total_pages: 0, total: 0 });
    setQueryData(['notifications', 'unread-count'], { unread: 0 });
    const { unmount: unmountEmpty } = render(<NotificationsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
    unmountEmpty();

    setQueryData(['notifications', 1, 20], {
      data: [{ id: 'n3', type: 'correction_result', title: 'Correction', message: 'done', created_at: '2026-02-11T10:00:00Z', is_read: false }],
      total_pages: 1,
      total: 1,
    });
    state.queryData.delete(JSON.stringify(['notifications', 'unread-count']));
    render(<NotificationsPage />);
    expect(screen.queryByRole('button', { name: 'notifications.markAllRead' })).not.toBeInTheDocument();
  });
});
