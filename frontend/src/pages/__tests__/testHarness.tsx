import { vi } from 'vitest';
import type { ReactNode } from 'react';

export type Role = 'admin' | 'manager' | 'employee';
export type MockUser = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: Role;
  is_active: boolean;
};

export const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: null as MockUser | null,
  language: 'ja',
  mutationPending: false,
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
  invalidateQueries: vi.fn(),
}));

export const apiMocks = vi.hoisted(() => ({
  auth: { login: vi.fn() },
  attendance: {
    getToday: vi.fn(),
    getList: vi.fn(),
    getSummary: vi.fn(),
    clockIn: vi.fn(),
    clockOut: vi.fn(),
  },
  dashboard: { getStats: vi.fn() },
  notifications: {
    getList: vi.fn(),
    getUnreadCount: vi.fn(),
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
    delete: vi.fn(),
  },
  leaves: {
    getList: vi.fn(),
    getPending: vi.fn(),
    create: vi.fn(),
    approve: vi.fn(),
  },
  shifts: { getList: vi.fn(), create: vi.fn(), delete: vi.fn() },
  users: { getAll: vi.fn(), create: vi.fn(), update: vi.fn(), delete: vi.fn() },
  departments: { getAll: vi.fn() },
  overtime: {
    getList: vi.fn(),
    getPending: vi.fn(),
    getAlerts: vi.fn(),
    create: vi.fn(),
    approve: vi.fn(),
  },
  corrections: {
    getList: vi.fn(),
    getPending: vi.fn(),
    create: vi.fn(),
    approve: vi.fn(),
  },
  projects: { getAll: vi.fn(), create: vi.fn() },
  timeEntries: {
    getList: vi.fn(),
    getSummary: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
  },
  holidays: {
    getByYear: vi.fn(),
    getCalendar: vi.fn(),
    getWorkingDays: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
  },
  export: {
    attendance: vi.fn(),
    leaves: vi.fn(),
    overtime: vi.fn(),
    projects: vi.fn(),
  },
  approvalFlows: { getAll: vi.fn(), create: vi.fn(), update: vi.fn(), delete: vi.fn() },
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
    t: (key: string, values?: Record<string, unknown>) => {
      if (values && 'name' in values) return `${key}:${String(values.name)}`;
      if (values && 'num' in values) return `${key}:${String(values.num)}`;
      return key;
    },
    i18n: {
      language: state.language,
      changeLanguage: vi.fn(),
    },
  }),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: (options: { queryKey: unknown[]; enabled?: boolean }) => {
    if (options.enabled === false) return { data: undefined };
    return { data: state.queryData.get(JSON.stringify(options.queryKey)) };
  },
  useMutation: (options: {
    mutationFn: (vars?: unknown) => unknown;
    onSuccess?: (data?: unknown, vars?: unknown) => void;
    onError?: (error: Error, vars?: unknown) => void;
  }) => ({
    mutate: (vars?: unknown) => {
      try {
        const result = options.mutationFn(vars);
        Promise.resolve(result)
          .then((data) => options.onSuccess?.(data, vars))
          .catch((error) => options.onError?.(error as Error, vars));
      } catch (error) {
        options.onError?.(error as Error, vars);
      }
    },
    isPending: state.mutationPending,
  }),
  useQueryClient: () => ({
    invalidateQueries: state.invalidateQueries,
  }),
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

vi.mock('@/components/ui/Pagination', () => ({
  Pagination: ({
    currentPage,
    totalPages,
    onPageChange,
    onPageSizeChange,
  }: {
    currentPage: number;
    totalPages: number;
    onPageChange: (p: number) => void;
    onPageSizeChange?: (s: number) => void;
  }) => (
    <div data-testid="pagination">
      <button onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}>
        pagination-next
      </button>
      <button onClick={() => onPageSizeChange?.(50)}>pagination-size</button>
    </div>
  ),
}));

vi.mock('@/api/client', () => ({
  api: apiMocks,
}));

export function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

export function resetHarness() {
  state.queryData.clear();
  state.mutationPending = false;
  state.language = 'ja';
  state.availableApps = [];
  state.user = {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'employee',
    is_active: true,
  };
  state.navigate.mockReset();
  state.setAuth.mockReset();
  state.invalidateQueries.mockReset();
  Object.values(apiMocks).forEach((group) => {
    Object.values(group).forEach((fn) => {
      (fn as ReturnType<typeof vi.fn>).mockReset();
    });
  });
  vi.stubGlobal('confirm', vi.fn(() => true));
  vi.stubGlobal('URL', {
    createObjectURL: vi.fn(() => 'blob:mock'),
    revokeObjectURL: vi.fn(),
  });
}

export function getWeekRange(baseDate: Date) {
  const start = new Date(baseDate);
  start.setDate(start.getDate() - start.getDay());
  const end = new Date(start);
  end.setDate(end.getDate() + 6);
  return {
    startDate: start.toISOString().split('T')[0],
    endDate: end.toISOString().split('T')[0],
  };
}
