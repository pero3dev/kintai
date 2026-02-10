import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { ShiftsPage } from '../ShiftsPage';
import { UsersPage } from '../UsersPage';
import { ProjectsPage } from '../ProjectsPage';
import { HolidaysPage } from '../HolidaysPage';
import { ExportPage } from '../ExportPage';
import { ApprovalFlowsPage } from '../ApprovalFlowsPage';

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
  invalidateQueries: vi.fn(),
}));

const apiMocks = vi.hoisted(() => ({
  shifts: { create: vi.fn(), delete: vi.fn() },
  users: { create: vi.fn(), update: vi.fn(), delete: vi.fn() },
  projects: { create: vi.fn() },
  timeEntries: { create: vi.fn(), delete: vi.fn() },
  holidays: { create: vi.fn(), delete: vi.fn() },
  export: {
    attendance: vi.fn(),
    leaves: vi.fn(),
    overtime: vi.fn(),
    projects: vi.fn(),
  },
  approvalFlows: { create: vi.fn(), update: vi.fn(), delete: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
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
    onError?: (error: Error) => void;
  }) => ({
    mutate: (vars?: unknown) => {
      Promise.resolve(options.mutationFn(vars))
        .then(() => options.onSuccess?.())
        .catch((error: Error) => options.onError?.(error));
    },
    isPending: false,
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
    shifts: { getList: vi.fn(), create: apiMocks.shifts.create, delete: apiMocks.shifts.delete },
    users: {
      getAll: vi.fn(),
      create: apiMocks.users.create,
      update: apiMocks.users.update,
      delete: apiMocks.users.delete,
    },
    departments: { getAll: vi.fn() },
    projects: { getAll: vi.fn(), create: apiMocks.projects.create },
    timeEntries: {
      getList: vi.fn(),
      getSummary: vi.fn(),
      create: apiMocks.timeEntries.create,
      delete: apiMocks.timeEntries.delete,
    },
    holidays: {
      getByYear: vi.fn(),
      getCalendar: vi.fn(),
      getWorkingDays: vi.fn(),
      create: apiMocks.holidays.create,
      delete: apiMocks.holidays.delete,
    },
    export: {
      attendance: apiMocks.export.attendance,
      leaves: apiMocks.export.leaves,
      overtime: apiMocks.export.overtime,
      projects: apiMocks.export.projects,
    },
    approvalFlows: {
      getAll: vi.fn(),
      create: apiMocks.approvalFlows.create,
      update: apiMocks.approvalFlows.update,
      delete: apiMocks.approvalFlows.delete,
    },
  },
}));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

function getWeekRange(baseDate: Date) {
  const start = new Date(baseDate);
  start.setDate(start.getDate() - start.getDay());
  const end = new Date(start);
  end.setDate(end.getDate() + 6);
  return {
    startDate: start.toISOString().split('T')[0],
    endDate: end.toISOString().split('T')[0],
  };
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
  apiMocks.shifts.create.mockReset();
  apiMocks.shifts.delete.mockReset();
  apiMocks.users.create.mockReset();
  apiMocks.users.update.mockReset();
  apiMocks.users.delete.mockReset();
  apiMocks.projects.create.mockReset();
  apiMocks.timeEntries.create.mockReset();
  apiMocks.timeEntries.delete.mockReset();
  apiMocks.holidays.create.mockReset();
  apiMocks.holidays.delete.mockReset();
  apiMocks.export.attendance.mockReset();
  apiMocks.export.leaves.mockReset();
  apiMocks.export.overtime.mockReset();
  apiMocks.export.projects.mockReset();
  apiMocks.approvalFlows.create.mockReset();
  apiMocks.approvalFlows.update.mockReset();
  apiMocks.approvalFlows.delete.mockReset();
  vi.stubGlobal('confirm', vi.fn(() => true));
  vi.stubGlobal('URL', {
    createObjectURL: vi.fn(() => 'blob:mock'),
    revokeObjectURL: vi.fn(),
  });
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ShiftsPage', () => {
  it('covers admin and non-admin branches', async () => {
    state.user = { ...state.user!, role: 'admin' };
    const { startDate, endDate } = getWeekRange(new Date());
    setQueryData(['shifts', startDate, endDate], [
      { id: 's1', user_id: 'u1', date: `${startDate}T00:00:00Z`, shift_type: 'morning' },
    ]);
    setQueryData(['users'], {
      data: [
        { id: 'u1', first_name: 'Taro', last_name: 'Yamada', role: 'employee' },
        { id: 'u2', first_name: 'Hanako', last_name: 'Suzuki', role: 'employee' },
      ],
    });
    apiMocks.shifts.create.mockResolvedValue({});
    apiMocks.shifts.delete.mockResolvedValue({});

    const { unmount } = render(<ShiftsPage />);
    const morningShiftCell = screen
      .getAllByText('shifts.types.morning')
      .find((node) => node.closest('td') || node.closest('div.cursor-pointer'));
    expect(morningShiftCell).toBeTruthy();
    fireEvent.click(morningShiftCell!);
    await waitFor(() => {
      expect(apiMocks.shifts.delete).toHaveBeenCalled();
    });
    (globalThis.confirm as ReturnType<typeof vi.fn>).mockReturnValueOnce(false);
    fireEvent.click(morningShiftCell!);
    await waitFor(() => {
      expect(apiMocks.shifts.delete).toHaveBeenCalledTimes(1);
    });
    const suzukiShiftCell =
      screen
        .getAllByText('Suzuki Hanako')
        .map((node) => node.closest('tr')?.querySelectorAll('td')?.[1] ?? null)
        .find((node) => node !== null) ??
      screen
        .getAllByText('Suzuki Hanako')
        .map((node) => node.closest('div.cursor-pointer'))
        .find((node) => node !== null);
    expect(suzukiShiftCell).toBeTruthy();
    fireEvent.click(suzukiShiftCell!);
    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));
    await waitFor(() => {
      expect(apiMocks.shifts.create).toHaveBeenCalled();
    });
    const headerButtons = screen.getAllByRole('button');
    fireEvent.click(headerButtons[0]);
    fireEvent.click(headerButtons[1]);
    unmount();

    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['shifts', startDate, endDate], []);
    setQueryData(['users'], { data: [] });
    const { unmount: unmountNoUsers } = render(<ShiftsPage />);
    expect(screen.getByText('shifts.noUsers')).toBeInTheDocument();
    unmountNoUsers();

    state.user = { ...state.user!, role: 'employee' };
    setQueryData(['shifts', startDate, endDate], []);
    const { unmount: unmountEmployee } = render(<ShiftsPage />);
    expect(screen.getAllByText('Yamada Taro').length).toBeGreaterThan(0);
    unmountEmployee();

    state.user = null;
    setQueryData(['shifts', startDate, endDate], []);
    render(<ShiftsPage />);
    expect(screen.getByText('shifts.noShiftData')).toBeInTheDocument();
  });
});

describe('UsersPage', () => {
  it('covers edit/delete/create branches', async () => {
    setQueryData(['users', 1, 20], {
      data: [
        { id: 'u1', email: 'a@example.com', first_name: 'A', last_name: 'B', role: 'admin', is_active: true },
        { id: 'u2', email: 'b@example.com', first_name: 'C', last_name: 'D', role: 'employee', is_active: false },
        { id: 'u3', email: 'c@example.com', first_name: 'E', last_name: 'F', role: 'manager', is_active: true },
      ],
      total_pages: 2,
      total: 3,
    });
    setQueryData(['departments'], [{ id: 'd1', name: 'Dev' }]);
    apiMocks.users.create.mockResolvedValue({});
    apiMocks.users.create.mockRejectedValueOnce(new Error('create failed'));
    apiMocks.users.update.mockResolvedValue({});
    apiMocks.users.delete.mockResolvedValue({});

    render(<UsersPage />);
    fireEvent.click(screen.getAllByText('common.edit')[0]);
    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));
    await waitFor(() => {
      expect(apiMocks.users.update).toHaveBeenCalled();
    });
    fireEvent.click(screen.getAllByTitle('common.delete')[0]);
    await waitFor(() => {
      expect(apiMocks.users.delete).toHaveBeenCalled();
    });
    (globalThis.confirm as ReturnType<typeof vi.fn>).mockReturnValueOnce(false);
    fireEvent.click(screen.getAllByTitle('common.delete')[0]);
    await waitFor(() => {
      expect(apiMocks.users.delete).toHaveBeenCalledTimes(1);
    });
    apiMocks.users.update.mockRejectedValueOnce(new Error('update failed'));
    fireEvent.click(screen.getAllByText('common.edit')[0]);
    fireEvent.click(screen.getByRole('button', { name: 'common.save' }));
    expect(await screen.findByText('update failed')).toBeInTheDocument();
    fireEvent.click(screen.getAllByRole('button', { name: 'common.cancel' })[0]);
    fireEvent.click(screen.getByRole('button', { name: 'users.addNew' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    expect((await screen.findAllByText('users.requiredFieldsError')).length).toBeGreaterThan(0);
    fireEvent.change(screen.getByPlaceholderText('user@example.com'), {
      target: { value: 'new@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('users.passwordRequirement'), {
      target: { value: 'Password123' },
    });
    const createTextInputs = document.querySelectorAll('.fixed input[type="text"]');
    fireEvent.change(createTextInputs[0] as HTMLInputElement, { target: { value: 'Yamada' } });
    fireEvent.change(createTextInputs[1] as HTMLInputElement, { target: { value: 'Ichiro' } });
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    expect(await screen.findByText('create failed')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    await waitFor(() => {
      expect(apiMocks.users.create).toHaveBeenCalledTimes(2);
    });
  });
});

describe('ProjectsPage', () => {
  it('covers tabs/forms/summary and non-admin branch', async () => {
    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['projects', 1, 12], {
      data: [
        { id: 'p1', code: 'PRJ-1', name: 'Project A', description: 'desc', budget_hours: 100, status: 'active' },
        { id: 'p2', code: 'PRJ-2', name: 'Project B', status: 'unknown' },
      ],
      total_pages: 1,
      total: 2,
    });
    setQueryData(['timeEntries', 'my'], [
      { id: 'te1', date: '2026-02-10', project: { name: 'Project A' }, minutes: 60, description: 'task' },
    ]);
    setQueryData(['timeEntries', 'summary'], [
      { project_code: 'A', project_name: 'Alpha', total_hours: 95, budget_hours: 100, member_count: 2 },
      { project_code: 'B', project_name: 'Beta', total_hours: 80, budget_hours: 100, member_count: 2 },
      { project_code: 'C', project_name: 'Gamma', total_hours: 40, budget_hours: 100, member_count: 2 },
      { project_code: 'D', project_name: 'Delta', total_hours: 20, budget_hours: null, member_count: 1 },
    ]);
    apiMocks.projects.create.mockResolvedValue({});
    apiMocks.timeEntries.create.mockResolvedValue({});
    apiMocks.timeEntries.delete.mockResolvedValue({});

    const { unmount } = render(<ProjectsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'projects.newProject' }));
    fireEvent.change(document.querySelector('input[name="name"]') as HTMLInputElement, {
      target: { value: 'New Project' },
    });
    fireEvent.change(document.querySelector('input[name="code"]') as HTMLInputElement, {
      target: { value: 'PRJ-NEW' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    await waitFor(() => {
      expect(apiMocks.projects.create).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByRole('button', { name: 'projects.logTime' }));
    fireEvent.change(document.querySelector('select[name="project_id"]') as HTMLSelectElement, {
      target: { value: 'p1' },
    });
    fireEvent.change(document.querySelector('input[name="date"]') as HTMLInputElement, {
      target: { value: '2026-02-10' },
    });
    fireEvent.change(document.querySelector('input[name="minutes"]') as HTMLInputElement, {
      target: { value: '30' },
    });
    fireEvent.change(document.querySelector('input[name="description"]') as HTMLInputElement, {
      target: { value: 'coding' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.submit' }));
    await waitFor(() => {
      expect(apiMocks.timeEntries.create).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByRole('button', { name: 'projects.myTimeEntries' }));
    fireEvent.click(screen.getAllByText('common.delete')[0]);
    await waitFor(() => {
      expect(apiMocks.timeEntries.delete).toHaveBeenCalled();
    });
    fireEvent.click(screen.getByRole('button', { name: 'projects.summary' }));
    expect(screen.getByText('95%')).toBeInTheDocument();
    unmount();

    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['projects', 1, 12], { data: [], total_pages: 0, total: 0 });
    setQueryData(['timeEntries', 'my'], []);
    setQueryData(['timeEntries', 'summary'], []);
    const { unmount: unmountSummaryEmpty } = render(<ProjectsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'projects.myTimeEntries' }));
    expect(screen.getAllByText('common.noData').length).toBeGreaterThan(0);
    fireEvent.click(screen.getByRole('button', { name: 'projects.summary' }));
    expect(screen.getByText('common.noData')).toBeInTheDocument();
    unmountSummaryEmpty();

    state.user = { ...state.user!, role: 'employee' };
    setQueryData(['projects', 1, 12], { data: [], total_pages: 0, total: 0 });
    setQueryData(['timeEntries', 'my'], []);
    render(<ProjectsPage />);
    expect(screen.queryByRole('button', { name: 'projects.summary' })).not.toBeInTheDocument();
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });
});

describe('HolidaysPage', () => {
  it('covers admin and non-admin branches', async () => {
    const now = new Date();
    const currentYear = now.getFullYear();
    const currentMonth = now.getMonth() + 1;
    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['holidays', currentYear], [
      { id: 'h1', name: 'Holiday', date: `${currentYear}-01-01`, holiday_type: 'national' },
      { id: 'h2', name: 'Company', date: `${currentYear}-02-01`, holiday_type: 'company' },
      { id: 'h3', name: 'Optional', date: `${currentYear}-03-01`, holiday_type: 'optional' },
      { id: 'h4', name: 'Other', date: `${currentYear}-04-01`, holiday_type: 'other' },
    ]);
    setQueryData(['holidays', 'calendar', currentYear, currentMonth], [
      { date: `${currentYear}-${String(currentMonth).padStart(2, '0')}-01`, is_holiday: true, is_weekend: false, holiday_name: 'Holiday' },
      { date: `${currentYear}-${String(currentMonth).padStart(2, '0')}-02`, is_holiday: false, is_weekend: true },
      { date: `${currentYear}-${String(currentMonth).padStart(2, '0')}-03`, is_holiday: false, is_weekend: false },
    ]);
    setQueryData(['holidays', 'working-days', currentYear, currentMonth], { working_days: 20, holidays: 2, weekends: 9, total_days: 31 });
    apiMocks.holidays.create.mockResolvedValue({});
    apiMocks.holidays.delete.mockResolvedValue({});

    const { unmount } = render(<HolidaysPage />);
    fireEvent.click(screen.getByRole('button', { name: 'holidays.add' }));
    fireEvent.change(document.querySelector('input[name="date"]') as HTMLInputElement, {
      target: { value: '2026-01-20' },
    });
    fireEvent.change(document.querySelector('input[name="name"]') as HTMLInputElement, {
      target: { value: 'My holiday' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    await waitFor(() => {
      expect(apiMocks.holidays.create).toHaveBeenCalled();
    });
    fireEvent.click(screen.getAllByText('common.delete')[0]);
    await waitFor(() => {
      expect(apiMocks.holidays.delete).toHaveBeenCalled();
    });
    expect(screen.getByText('holidays.types.company')).toBeInTheDocument();
    expect(screen.getByText('holidays.types.optional')).toBeInTheDocument();
    expect(screen.getByText('other')).toBeInTheDocument();
    unmount();

    state.user = { ...state.user!, role: 'employee' };
    setQueryData(['holidays', currentYear], []);
    render(<HolidaysPage />);
    expect(screen.queryByRole('button', { name: 'holidays.add' })).not.toBeInTheDocument();
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers month navigation edge branches', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-15T09:00:00Z'));
    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['holidays', 2026], []);
    setQueryData(['holidays', 'calendar', 2026, 1], []);
    setQueryData(['holidays', 'working-days', 2026, 1], { working_days: 20, holidays: 0, weekends: 8, total_days: 31 });
    const { unmount } = render(<HolidaysPage />);
    const iconButtons = screen
      .getAllByRole('button')
      .filter((button) => button.textContent?.trim() === '');
    fireEvent.click(iconButtons[0]);
    fireEvent.click(iconButtons[1]);
    unmount();

    vi.setSystemTime(new Date('2026-03-15T09:00:00Z'));
    setQueryData(['holidays', 2026], []);
    setQueryData(['holidays', 'calendar', 2026, 3], []);
    setQueryData(['holidays', 'working-days', 2026, 3], { working_days: 20, holidays: 0, weekends: 8, total_days: 31 });
    render(<HolidaysPage />);
    const nonEdgeButtons = screen
      .getAllByRole('button')
      .filter((button) => button.textContent?.trim() === '');
    fireEvent.click(nonEdgeButtons[0]);
    fireEvent.click(nonEdgeButtons[1]);
  });
});

describe('ExportPage', () => {
  it('covers admin and non-admin export branches', async () => {
    state.user = { ...state.user!, role: 'employee' };
    const { unmount } = render(<ExportPage />);
    expect(screen.getByText('export.adminOnly')).toBeInTheDocument();
    unmount();

    state.user = { ...state.user!, role: 'admin' };
    const blob = new Blob(['a,b\n1,2'], { type: 'text/csv' });
    apiMocks.export.attendance.mockResolvedValue(blob);
    apiMocks.export.leaves.mockResolvedValue(blob);
    apiMocks.export.overtime.mockResolvedValue(blob);
    apiMocks.export.projects.mockResolvedValue(blob);
    vi.spyOn(console, 'error').mockImplementation(() => {});

    render(<ExportPage />);
    const csvButtons = screen.getAllByText('CSV');
    csvButtons.forEach((label) => {
      fireEvent.click(label.closest('button') as HTMLButtonElement);
    });
    await waitFor(() => {
      expect(apiMocks.export.attendance).toHaveBeenCalled();
      expect(apiMocks.export.leaves).toHaveBeenCalled();
      expect(apiMocks.export.overtime).toHaveBeenCalled();
      expect(apiMocks.export.projects).toHaveBeenCalled();
    });
    apiMocks.export.projects.mockRejectedValueOnce(new Error('fail'));
    fireEvent.click(screen.getAllByText('CSV')[3].closest('button') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.export.projects).toHaveBeenCalledTimes(2);
    });
  });
});

describe('ApprovalFlowsPage', () => {
  it('covers empty/list/form/action branches', async () => {
    setQueryData(['approval-flows'], []);
    const { unmount } = render(<ApprovalFlowsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
    unmount();

    setQueryData(['approval-flows'], [
      {
        id: 'f1',
        name: 'Flow 1',
        flow_type: 'leave',
        is_active: true,
        steps: [
          { id: 's1', step_order: 2, step_type: 'specific_user', approver_role: '' },
          { id: 's2', step_order: 1, step_type: 'role', approver_role: 'manager' },
        ],
      },
      { id: 'f2', name: 'Flow 2', flow_type: 'unknown', is_active: false, steps: [] },
      {
        id: 'f3',
        name: 'Flow 3',
        flow_type: 'overtime',
        is_active: true,
        steps: [{ id: 's3', step_order: 1, step_type: 'role', approver_role: 'admin' }],
      },
      { id: 'f4', name: 'Flow 4', flow_type: 'correction', is_active: true },
    ]);
    apiMocks.approvalFlows.create.mockResolvedValue({});
    apiMocks.approvalFlows.update.mockResolvedValue({});
    apiMocks.approvalFlows.delete.mockResolvedValue({});

    render(<ApprovalFlowsPage />);
    fireEvent.click(screen.getByRole('button', { name: 'approvalFlows.create' }));
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    expect(await screen.findByText('approvalFlows.validation.nameRequired')).toBeInTheDocument();
    fireEvent.click(screen.getByText('approvalFlows.addStep'));
    fireEvent.change(screen.getByPlaceholderText('approvalFlows.placeholder'), { target: { value: 'New Flow' } });
    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));
    await waitFor(() => {
      expect(apiMocks.approvalFlows.create).toHaveBeenCalled();
    });
    fireEvent.click(screen.getAllByText('approvalFlows.disable')[0]);
    fireEvent.click(screen.getAllByText('common.delete')[0]);
    await waitFor(() => {
      expect(apiMocks.approvalFlows.update).toHaveBeenCalled();
      expect(apiMocks.approvalFlows.delete).toHaveBeenCalled();
    });
    expect(screen.getByText('→')).toBeInTheDocument();
  });
});
