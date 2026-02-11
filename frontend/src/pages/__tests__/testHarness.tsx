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

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  user: null as MockUser | null,
  params: {} as Record<string, string>,
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

const apiMocks = vi.hoisted(() => ({
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
  expenses: {
    getList: vi.fn(),
    getByID: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    getPending: vi.fn(),
    approve: vi.fn(),
    getStats: vi.fn(),
    getComments: vi.fn(),
    addComment: vi.fn(),
    getHistory: vi.fn(),
    getReport: vi.fn(),
    getMonthlyTrend: vi.fn(),
    exportCSV: vi.fn(),
    exportPDF: vi.fn(),
    uploadReceipt: vi.fn(),
    getTemplates: vi.fn(),
    createTemplate: vi.fn(),
    updateTemplate: vi.fn(),
    deleteTemplate: vi.fn(),
    useTemplate: vi.fn(),
    getPolicies: vi.fn(),
    createPolicy: vi.fn(),
    updatePolicy: vi.fn(),
    deletePolicy: vi.fn(),
    getBudgets: vi.fn(),
    getPolicyViolations: vi.fn(),
    getNotifications: vi.fn(),
    markNotificationRead: vi.fn(),
    markAllNotificationsRead: vi.fn(),
    getReminders: vi.fn(),
    dismissReminder: vi.fn(),
    getNotificationSettings: vi.fn(),
    updateNotificationSettings: vi.fn(),
    advancedApprove: vi.fn(),
    getApprovalFlowConfig: vi.fn(),
    getDelegates: vi.fn(),
    setDelegate: vi.fn(),
    removeDelegate: vi.fn(),
  },
  hr: {
    getStats: vi.fn(),
    getRecentActivities: vi.fn(),
    getEmployees: vi.fn(),
    getEmployee: vi.fn(),
    createEmployee: vi.fn(),
    updateEmployee: vi.fn(),
    deleteEmployee: vi.fn(),
    getDepartments: vi.fn(),
    getDepartment: vi.fn(),
    createDepartment: vi.fn(),
    updateDepartment: vi.fn(),
    deleteDepartment: vi.fn(),
    getEvaluations: vi.fn(),
    getEvaluation: vi.fn(),
    createEvaluation: vi.fn(),
    updateEvaluation: vi.fn(),
    submitEvaluation: vi.fn(),
    getEvaluationCycles: vi.fn(),
    createEvaluationCycle: vi.fn(),
    getGoals: vi.fn(),
    getGoal: vi.fn(),
    createGoal: vi.fn(),
    updateGoal: vi.fn(),
    deleteGoal: vi.fn(),
    updateGoalProgress: vi.fn(),
    getTrainingPrograms: vi.fn(),
    getTrainingProgram: vi.fn(),
    createTrainingProgram: vi.fn(),
    updateTrainingProgram: vi.fn(),
    deleteTrainingProgram: vi.fn(),
    enrollTraining: vi.fn(),
    completeTraining: vi.fn(),
    getPositions: vi.fn(),
    getPosition: vi.fn(),
    createPosition: vi.fn(),
    updatePosition: vi.fn(),
    getApplicants: vi.fn(),
    createApplicant: vi.fn(),
    updateApplicantStage: vi.fn(),
    getDocuments: vi.fn(),
    uploadDocument: vi.fn(),
    deleteDocument: vi.fn(),
    downloadDocument: vi.fn(),
    getAnnouncements: vi.fn(),
    getAnnouncement: vi.fn(),
    createAnnouncement: vi.fn(),
    updateAnnouncement: vi.fn(),
    deleteAnnouncement: vi.fn(),
    getAttendanceIntegration: vi.fn(),
    getAttendanceAlerts: vi.fn(),
    getAttendanceTrend: vi.fn(),
    getOrgChart: vi.fn(),
    simulateOrgChange: vi.fn(),
    getOneOnOnes: vi.fn(),
    getOneOnOne: vi.fn(),
    createOneOnOne: vi.fn(),
    updateOneOnOne: vi.fn(),
    deleteOneOnOne: vi.fn(),
    addActionItem: vi.fn(),
    toggleActionItem: vi.fn(),
    getSkillMap: vi.fn(),
    getSkillGapAnalysis: vi.fn(),
    addEmployeeSkill: vi.fn(),
    updateEmployeeSkill: vi.fn(),
    getSalaryOverview: vi.fn(),
    simulateSalary: vi.fn(),
    getSalaryHistory: vi.fn(),
    getBudgetOverview: vi.fn(),
    getOnboardings: vi.fn(),
    getOnboarding: vi.fn(),
    createOnboarding: vi.fn(),
    updateOnboarding: vi.fn(),
    toggleOnboardingTask: vi.fn(),
    getOnboardingTemplates: vi.fn(),
    createOnboardingTemplate: vi.fn(),
    getOffboardings: vi.fn(),
    getOffboarding: vi.fn(),
    createOffboarding: vi.fn(),
    updateOffboarding: vi.fn(),
    toggleOffboardingChecklist: vi.fn(),
    getTurnoverAnalytics: vi.fn(),
    getSurveys: vi.fn(),
    getSurvey: vi.fn(),
    createSurvey: vi.fn(),
    updateSurvey: vi.fn(),
    deleteSurvey: vi.fn(),
    publishSurvey: vi.fn(),
    closeSurvey: vi.fn(),
    getSurveyResults: vi.fn(),
    submitSurveyResponse: vi.fn(),
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
  useNavigate: () => state.navigate,
  useParams: () => state.params,
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
  useQuery: (options: { queryKey: unknown[]; enabled?: boolean; queryFn?: () => unknown }) => {
    if (options.enabled === false) return { data: undefined, isLoading: false };
    const key = JSON.stringify(options.queryKey);
    let data = state.queryData.get(key);
    if (data === undefined) {
      try {
        const result = options.queryFn?.();
        if (result !== undefined && typeof (result as { then?: unknown }).then !== 'function') {
          data = result;
        }
      } catch {
        // ignore query fallback errors in tests
      }
    }
    return { data, isLoading: false };
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

export function setUserRole(role: Role) {
  if (!state.user) return;
  state.user = { ...state.user, role };
}

export function setUser(user: MockUser | null) {
  state.user = user;
}

export function getApiMocks() {
  return apiMocks;
}

export function resetHarness() {
  state.queryData.clear();
  state.mutationPending = false;
  state.language = 'ja';
  state.params = {};
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

export function setRouteParams(params: Record<string, string>) {
  state.params = params;
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
