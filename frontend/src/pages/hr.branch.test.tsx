import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { HRDashboardPage } from './hr/HRDashboardPage';
import { HREmployeesPage } from './hr/HREmployeesPage';
import { HREmployeeDetailPage } from './hr/HREmployeeDetailPage';
import { HRDepartmentsPage } from './hr/HRDepartmentsPage';
import { HREvaluationsPage } from './hr/HREvaluationsPage';
import { HRGoalsPage } from './hr/HRGoalsPage';
import { HRTrainingPage } from './hr/HRTrainingPage';
import { HRRecruitmentPage } from './hr/HRRecruitmentPage';
import { HRDocumentsPage } from './hr/HRDocumentsPage';
import { HRAnnouncementsPage } from './hr/HRAnnouncementsPage';
import { HRAttendanceIntegrationPage } from './hr/HRAttendanceIntegrationPage';
import { HROrgChartPage } from './hr/HROrgChartPage';
import { HROneOnOnePage } from './hr/HROneOnOnePage';
import { HRSkillMapPage } from './hr/HRSkillMapPage';
import { HRSalarySimulatorPage } from './hr/HRSalarySimulatorPage';
import { HROnboardingPage } from './hr/HROnboardingPage';
import { HROffboardingPage } from './hr/HROffboardingPage';
import { HRSurveyPage } from './hr/HRSurveyPage';

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  loadingKeys: new Set<string>(),
  params: { employeeId: 'emp-1' },
  invalidateQueries: vi.fn(),
  mutationPending: false,
}));

const apiMocks = vi.hoisted(() => ({
  hr: {
    getStats: vi.fn(),
    getRecentActivities: vi.fn(),

    getEmployees: vi.fn(),
    createEmployee: vi.fn(),

    getDepartments: vi.fn(),
    createDepartment: vi.fn(),
    updateDepartment: vi.fn(),
    deleteDepartment: vi.fn(),

    getEmployee: vi.fn(),
    updateEmployee: vi.fn(),
    getSalaryHistory: vi.fn(),

    getGoals: vi.fn(),
    createGoal: vi.fn(),
    updateGoalProgress: vi.fn(),
    deleteGoal: vi.fn(),

    getDocuments: vi.fn(),
    uploadDocument: vi.fn(),
    deleteDocument: vi.fn(),
    downloadDocument: vi.fn(),

    getEvaluations: vi.fn(),
    getEvaluationCycles: vi.fn(),
    createEvaluation: vi.fn(),
    submitEvaluation: vi.fn(),
    createEvaluationCycle: vi.fn(),

    getTrainingPrograms: vi.fn(),
    createTrainingProgram: vi.fn(),
    enrollTraining: vi.fn(),
    completeTraining: vi.fn(),

    getPositions: vi.fn(),
    createPosition: vi.fn(),
    getApplicants: vi.fn(),
    createApplicant: vi.fn(),
    updateApplicantStage: vi.fn(),

    getAnnouncements: vi.fn(),
    createAnnouncement: vi.fn(),
    updateAnnouncement: vi.fn(),
    deleteAnnouncement: vi.fn(),

    getAttendanceIntegration: vi.fn(),
    getAttendanceAlerts: vi.fn(),
    getAttendanceTrend: vi.fn(),

    getOrgChart: vi.fn(),

    getOneOnOnes: vi.fn(),
    createOneOnOne: vi.fn(),
    deleteOneOnOne: vi.fn(),
    toggleActionItem: vi.fn(),

    getSkillMap: vi.fn(),
    getSkillGapAnalysis: vi.fn(),
    addEmployeeSkill: vi.fn(),

    getSalaryOverview: vi.fn(),
    getBudgetOverview: vi.fn(),
    simulateSalary: vi.fn(),

    getOnboardings: vi.fn(),
    getOnboardingTemplates: vi.fn(),
    createOnboarding: vi.fn(),
    toggleOnboardingTask: vi.fn(),
    getOnboarding: vi.fn(),

    getOffboardings: vi.fn(),
    getTurnoverAnalytics: vi.fn(),
    createOffboarding: vi.fn(),
    toggleOffboardingChecklist: vi.fn(),
    getOffboarding: vi.fn(),

    getSurveys: vi.fn(),
    getSurveyResults: vi.fn(),
    createSurvey: vi.fn(),
    publishSurvey: vi.fn(),
    closeSurvey: vi.fn(),
    deleteSurvey: vi.fn(),
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
  useParams: () => state.params,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, values?: Record<string, unknown>) => {
      if (key.startsWith('hr.offboarding.checklistItems.')) {
        return `translated.${key.split('.').pop()}`;
      }
      if (!values || typeof values !== 'object') return key;
      if ('count' in values) return `${key}:${String(values.count)}`;
      return `${key}:${JSON.stringify(values)}`;
    },
    i18n: { language: 'ja', changeLanguage: vi.fn() },
  }),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: (options: {
    queryKey: unknown[];
    enabled?: boolean;
    queryFn?: () => unknown;
  }) => {
    if (options.enabled === false) return { data: undefined, isLoading: false };
    const key = JSON.stringify(options.queryKey);
    let data = state.queryData.get(key);
    try {
      const result = options.queryFn?.();
      if (
        data === undefined &&
        result !== undefined &&
        typeof (result as { then?: unknown }).then !== 'function'
      ) {
        data = result;
      }
    } catch {
      // ignore in tests
    }
    return { data, isLoading: state.loadingKeys.has(key) };
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
  useQueryClient: () => ({ invalidateQueries: state.invalidateQueries }),
}));

vi.mock('@/api/client', () => ({ api: apiMocks }));

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(JSON.stringify(key), value);
}

function setLoading(key: unknown[], loading = true) {
  const target = JSON.stringify(key);
  if (loading) {
    state.loadingKeys.add(target);
    return;
  }
  state.loadingKeys.delete(target);
}

function clickButtonByText(text: string, index = 0) {
  const buttons = screen
    .getAllByText(text)
    .map((node) => node.closest('button'))
    .filter((button): button is HTMLButtonElement => button instanceof HTMLButtonElement);
  if (!buttons[index]) throw new Error(`button not found for ${text}`);
  fireEvent.click(buttons[index]);
}

function clickButtonByIcon(icon: string, index = 0) {
  const buttons = screen
    .getAllByText(icon)
    .map((node) => node.closest('button'))
    .filter((button): button is HTMLButtonElement => button instanceof HTMLButtonElement);
  if (!buttons[index]) throw new Error(`icon button not found for ${icon}`);
  fireEvent.click(buttons[index]);
}

function resetHarness() {
  state.queryData.clear();
  state.loadingKeys.clear();
  state.params = { employeeId: 'emp-1' };
  state.invalidateQueries.mockReset();
  state.mutationPending = false;

  Object.values(apiMocks.hr).forEach((fn) => {
    (fn as ReturnType<typeof vi.fn>).mockReset();
  });

  vi.stubGlobal('confirm', vi.fn(() => true));
  vi.stubGlobal('URL', {
    createObjectURL: vi.fn(() => 'blob:mock'),
    revokeObjectURL: vi.fn(),
  });
  vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
}

beforeEach(() => {
  resetHarness();
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('HRDashboardPage', () => {
  it('covers empty and populated branches', () => {
    setQueryData(['hr-stats'], undefined);
    setQueryData(['hr-activities'], { data: [] });

    const { unmount } = render(<HRDashboardPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
    unmount();

    setQueryData(['hr-stats'], {
      total_employees: 10,
      active_employees: 8,
      new_hires_this_month: 1,
      turnover_rate: 12.34,
      open_positions: 2,
      upcoming_reviews: 3,
      training_completion: 87,
      pending_documents: 4,
    });
    setQueryData(['hr-activities'], [
      { id: 'a1', message: 'New employee', timestamp: '2026-02-10 10:00' },
      { id: 'a2', icon: 'campaign', message: 'Announcement', timestamp: '2026-02-10 11:00' },
    ]);

    render(<HRDashboardPage />);
    expect(screen.getByText('12.3%')).toBeInTheDocument();
    expect(screen.getByText('87%')).toBeInTheDocument();
    expect(screen.getByText('New employee')).toBeInTheDocument();
  });

  it('covers missing activities fallback branch', () => {
    setQueryData(['hr-stats'], { total_employees: 1 });
    setQueryData(['hr-activities'], undefined);

    render(<HRDashboardPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });
});

describe('HREmployeesPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-employees', 1, '', '', '', '']);
    setQueryData(['hr-departments'], { data: [] });

    const { unmount } = render(<HREmployeesPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-employees', 1, '', '', '', ''], false);
    setQueryData(['hr-employees', 1, '', '', '', ''], { data: [], total: 0 });

    render(<HREmployeesPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers list filters pagination and create branches', async () => {
    apiMocks.hr.createEmployee.mockResolvedValue({});
    setQueryData(['hr-departments'], { data: [{ id: 'd1', name: 'Dev' }] });
    setQueryData(['hr-employees', 1, '', '', '', ''], {
      data: Array.from({ length: 20 }).map((_, i) => ({
        id: `e${i + 1}`,
        employee_id: i === 0 ? '' : `EMP-${i + 1}`,
        first_name: `F${i + 1}`,
        last_name: `L${i + 1}`,
        department_name: i === 0 ? '' : 'Dev',
        position: i === 0 ? '' : 'Engineer',
        status: i === 0 ? 'mystery' : 'active',
        employment_type: i === 0 ? 'mystery' : 'fullTime',
        hire_date: i === 0 ? '' : '2026-01-01',
      })),
      total: 25,
    });
    setQueryData(['hr-employees', 2, '', '', '', ''], { data: [], total: 25 });
    setQueryData(['hr-employees', 1, 'abc', '', '', ''], { data: [], total: 0 });
    setQueryData(['hr-employees', 1, 'abc', 'd1', '', ''], { data: [], total: 0 });
    setQueryData(['hr-employees', 1, 'abc', 'd1', 'active', ''], { data: [], total: 0 });
    setQueryData(['hr-employees', 1, 'abc', 'd1', 'active', 'fullTime'], { data: [], total: 0 });

    render(<HREmployeesPage />);
    expect(screen.getAllByText('hr.employees.statuses.mystery').length).toBeGreaterThan(0);

    clickButtonByIcon('chevron_right');
    await waitFor(() => {
      expect(apiMocks.hr.getEmployees).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2, page_size: 20 })
      );
    });

    fireEvent.change(screen.getByPlaceholderText('common.search'), { target: { value: 'abc' } });
    const filters = screen.getAllByRole('combobox');
    fireEvent.change(filters[0], { target: { value: 'd1' } });
    fireEvent.change(filters[1], { target: { value: 'active' } });
    fireEvent.change(filters[2], { target: { value: 'fullTime' } });

    clickButtonByText('hr.employees.addEmployee', 0);
    fireEvent.change(screen.getByPlaceholderText('EMP-001'), { target: { value: 'EMP-101' } });
    fireEvent.change(document.querySelector('.fixed input[type="email"]') as HTMLInputElement, {
      target: { value: 'new@example.com' },
    });
    const modalInputs = document.querySelectorAll('.fixed input');
    fireEvent.change(modalInputs[2] as HTMLInputElement, { target: { value: 'Yamada' } });
    fireEvent.change(modalInputs[3] as HTMLInputElement, { target: { value: 'Taro' } });
    clickButtonByText('common.create');

    await waitFor(() => {
      expect(apiMocks.hr.createEmployee).toHaveBeenCalled();
    });
  });

  it('covers alternate payload shapes and initials fallback', () => {
    setQueryData(['hr-departments'], [{ id: 'd1', name: 'Dev' }]);
    setQueryData(['hr-employees', 1, '', '', '', ''], {
      employees: [
        {
          id: 'e-fallback',
          employee_id: '',
          first_name: 'NoLast',
          last_name: '',
          department_name: '',
          position: '',
          status: 'active',
          employment_type: 'fullTime',
          hire_date: '',
        },
      ],
      total: 1,
    });

    render(<HREmployeesPage />);
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);
  });
});

describe('HREmployeeDetailPage', () => {
  it('covers loading and history empty/loading branches', () => {
    state.params = { employeeId: 'emp-1' };
    setLoading(['hr-employee', 'emp-1']);
    const { unmount } = render(<HREmployeeDetailPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-employee', 'emp-1'], false);
    setQueryData(['hr-employee', 'emp-1'], { id: 'emp-1', first_name: 'Taro', last_name: 'Yamada' });
    setLoading(['hr-salary-history', 'emp-1']);
    const { unmount: unmount2 } = render(<HREmployeeDetailPage />);
    clickButtonByText('hr.detail.tabs.history');
    expect(screen.getAllByText('common.loading').length).toBeGreaterThan(0);
    unmount2();

    setLoading(['hr-salary-history', 'emp-1'], false);
    setQueryData(['hr-salary-history', 'emp-1'], { data: [] });
    render(<HREmployeeDetailPage />);
    clickButtonByText('hr.detail.tabs.history');
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers profile goals docs history and edit branches', async () => {
    apiMocks.hr.updateEmployee.mockResolvedValue({});
    state.params = { employeeId: 'emp-1' };
    setQueryData(['hr-employee', 'emp-1'], {
      id: 'emp-1',
      employee_id: 'EMP-1',
      first_name: 'Taro',
      last_name: 'Yamada',
      email: 'taro@example.com',
      position: 'Engineer',
      department_name: 'Dev',
      status: 'active',
      employment_type: 'fullTime',
      skills: ['Go', 'React'],
      address: 'Tokyo',
    });
    setQueryData(['hr-goals', 'emp-1'], {
      data: [{ id: 'g1', title: 'Goal A', due_date: '2026-03-01', progress: 40 }],
    });
    setQueryData(['hr-documents', 'emp-1'], {
      data: [{ id: 'd1', name: 'Contract', type: 'contract', upload_date: '2026-02-01' }],
    });
    setQueryData(['hr-salary-history', 'emp-1'], {
      data: [
        { id: 'h1', effective_date: '2026-01-01', net_salary: 300000, reason: 'raise' },
        { id: 'h2', effective_date: '2025-01-01', net_salary: 200000, reason: 'base up' },
        { id: 'h3', effective_date: '2024-01-01', net_salary: 250000, reason: 'adjust' },
      ],
    });

    render(<HREmployeeDetailPage />);
    expect(screen.getAllByText('Yamada Taro').length).toBeGreaterThan(0);
    expect(screen.getByText('Go')).toBeInTheDocument();

    clickButtonByText('hr.detail.editProfile');
    const emailInput = document.querySelector('.fixed input[type="email"]') as HTMLInputElement;
    fireEvent.change(emailInput, { target: { value: 'updated@example.com' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.updateEmployee).toHaveBeenCalledWith(
        'emp-1',
        expect.objectContaining({ email: 'updated@example.com' })
      );
    });

    clickButtonByText('hr.detail.tabs.goals');
    expect(screen.getByText('Goal A')).toBeInTheDocument();

    clickButtonByText('hr.detail.tabs.documents');
    expect(screen.getByText('Contract')).toBeInTheDocument();

    clickButtonByText('hr.detail.tabs.history');
    expect(screen.getByText('raise')).toBeInTheDocument();
    expect(screen.getByText('adjust')).toBeInTheDocument();
  });

  it('covers detail fallback values for tabs and form defaults', async () => {
    apiMocks.hr.updateEmployee.mockResolvedValue({});
    state.params = { employeeId: 'emp-fallback' };
    setQueryData(['hr-employee', 'emp-fallback'], {
      id: 'emp-fallback',
      first_name: '',
      last_name: '',
      email: '',
      gender: '',
      employment_type: '',
      status: '',
      skills: [],
    });
    setQueryData(['hr-goals', 'emp-fallback'], { data: [] });
    setQueryData(['hr-documents', 'emp-fallback'], { data: [] });
    setQueryData(['hr-salary-history', 'emp-fallback'], {
      data: [
        { id: 'h1', effective_date: '', reason: '', net_salary: 0, allowances: 0, deductions: 0 },
        { id: 'h2' },
      ],
    });

    const { unmount } = render(<HREmployeeDetailPage />);
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);

    clickButtonByText('hr.detail.tabs.goals');
    expect(screen.getByText('common.noData')).toBeInTheDocument();

    clickButtonByText('hr.detail.tabs.documents');
    expect(screen.getByText('common.noData')).toBeInTheDocument();

    clickButtonByText('hr.detail.tabs.history');
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);

    clickButtonByText('hr.detail.editProfile');
    const modalInputs = document.querySelectorAll('.fixed input');
    expect((modalInputs[0] as HTMLInputElement).value).toBe('');
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.updateEmployee).toHaveBeenCalledWith(
        'emp-fallback',
        expect.objectContaining({ last_name: '', first_name: '' })
      );
    });
    unmount();

    setQueryData(['hr-goals', 'emp-fallback'], {
      data: [{ id: 'g-zero', title: 'Zero Goal', due_date: '', progress: undefined }],
    });
    render(<HREmployeeDetailPage />);
    clickButtonByText('hr.detail.tabs.goals');
    expect(screen.getByText('0%')).toBeInTheDocument();
  });
});

describe('HRDepartmentsPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-departments']);
    const { unmount } = render(<HRDepartmentsPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-departments'], false);
    setQueryData(['hr-departments'], { data: [] });
    render(<HRDepartmentsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers list chart create edit delete branches', async () => {
    apiMocks.hr.createDepartment.mockResolvedValue({});
    apiMocks.hr.updateDepartment.mockResolvedValue({});
    apiMocks.hr.deleteDepartment.mockResolvedValue({});
    setQueryData(['hr-departments'], {
      data: [
        {
          id: 'd1',
          name: 'Engineering',
          code: 'ENG',
          description: 'core',
          manager_name: 'Alice',
          member_count: 10,
        },
        {
          id: 'd2',
          name: 'Platform',
          code: 'PLT',
          parent_id: 'd1',
          manager_name: '',
          member_count: 4,
        },
      ],
    });

    render(<HRDepartmentsPage />);
    expect(screen.getByText('Engineering')).toBeInTheDocument();

    clickButtonByIcon('edit', 0);
    fireEvent.change(screen.getByDisplayValue('Engineering'), { target: { value: 'Engineering 2' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.updateDepartment).toHaveBeenCalled();
    });

    (globalThis.confirm as ReturnType<typeof vi.fn>).mockReturnValueOnce(false).mockReturnValueOnce(true);
    clickButtonByIcon('delete', 0);
    clickButtonByIcon('delete', 0);
    await waitFor(() => {
      expect(apiMocks.hr.deleteDepartment).toHaveBeenCalledTimes(1);
    });

    clickButtonByText('hr.departments.addDepartment');
    const nameInput = document.querySelector('.fixed input') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'New Dept' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createDepartment).toHaveBeenCalled();
    });

    clickButtonByIcon('account_tree');
    expect(screen.getByText('Platform')).toBeInTheDocument();
  });

  it('covers edit defaults, budget conversion and org fallback labels', async () => {
    apiMocks.hr.updateDepartment.mockResolvedValue({});
    apiMocks.hr.createDepartment.mockResolvedValue({});
    setQueryData(['hr-departments'], {
      data: [
        {
          id: 'd-root',
          name: 'Root',
          code: '',
          manager_name: '',
          member_count: 0,
        },
        {
          id: 'd-child',
          parent_id: 'd-root',
          name: 'Child',
        },
      ],
    });

    render(<HRDepartmentsPage />);
    clickButtonByIcon('edit', 1);
    const formInputs = document.querySelectorAll('.fixed input');
    fireEvent.change(formInputs[0] as HTMLInputElement, { target: { value: 'Child Updated' } });
    fireEvent.change(formInputs[2] as HTMLInputElement, { target: { value: '1000' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.updateDepartment).toHaveBeenCalledWith(
        'd-child',
        expect.objectContaining({ budget: 1000 })
      );
    });

    clickButtonByText('hr.departments.addDepartment');
    const createInputs = document.querySelectorAll('.fixed input');
    fireEvent.change(createInputs[0] as HTMLInputElement, { target: { value: 'Budget Dept' } });
    fireEvent.change(createInputs[2] as HTMLInputElement, { target: { value: '3000' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createDepartment).toHaveBeenCalledWith(
        expect.objectContaining({ budget: 3000 })
      );
    });

    clickButtonByIcon('account_tree');
    expect(
      screen.getAllByText((_, node) => node?.textContent?.includes('hr.departments.manager') ?? false).length
    ).toBeGreaterThan(0);
  });
});

describe('HREvaluationsPage', () => {
  it('covers loading and no-data branches', () => {
    setLoading(['hr-evaluations', { cycle_id: '', status: '' }]);
    setQueryData(['hr-evaluation-cycles'], { data: [] });
    const { unmount } = render(<HREvaluationsPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-evaluations', { cycle_id: '', status: '' }], false);
    setQueryData(['hr-evaluations', { cycle_id: '', status: '' }], { data: [] });
    render(<HREvaluationsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers evaluation cycle forms and actions', async () => {
    apiMocks.hr.submitEvaluation.mockResolvedValue({});
    apiMocks.hr.createEvaluation.mockResolvedValue({});
    apiMocks.hr.createEvaluationCycle.mockResolvedValue({});
    setQueryData(['hr-evaluation-cycles'], {
      data: [
        { id: 'c1', name: '2026 H1', status: 'active', start_date: '2026-01-01', end_date: '2026-06-30' },
        { id: 'c2', name: '2026 H2', status: 'draft', start_date: '2026-07-01', end_date: '2026-12-31', description: 'next' },
      ],
    });
    setQueryData(['hr-evaluations', { cycle_id: '', status: '' }], {
      data: [
        {
          id: 'e1',
          employee_name: 'Taro',
          cycle_name: '2026 H1',
          status: 'draft',
          criteria: [{ name: 'Quality', score: 4 }],
        },
        {
          id: 'e2',
          employee_name: 'Hanako',
          cycle_name: '2026 H1',
          status: 'completed',
          final_score: 'A',
        },
      ],
    });

    render(<HREvaluationsPage />);
    clickButtonByText('hr.evaluations.submit');
    await waitFor(() => {
      expect(apiMocks.hr.submitEvaluation).toHaveBeenCalledWith('e1');
    });

    clickButtonByText('hr.evaluations.startEvaluation');
    const modalSelect = document.querySelector('.fixed select') as HTMLSelectElement;
    fireEvent.change(modalSelect, { target: { value: 'c1' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createEvaluation).toHaveBeenCalled();
    });

    clickButtonByText('hr.evaluations.evaluationCycle');
    clickButtonByText('hr.evaluations.addCycle');
    const cycleNameInput = document.querySelector('.fixed input') as HTMLInputElement;
    fireEvent.change(cycleNameInput, { target: { value: '2027 H1' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createEvaluationCycle).toHaveBeenCalled();
    });
  });

  it('covers evaluation fallback fields and unknown status mappings', () => {
    setQueryData(['hr-evaluation-cycles'], [
      { id: 'c-fallback', name: 'Fallback Cycle', start_date: '', end_date: '' },
    ]);
    setQueryData(['hr-evaluations', { cycle_id: '', status: '' }], [
      {
        id: 'e-fallback-1',
        employee_name: '',
        cycle_name: '',
        department: '',
        final_score: 'Z',
        criteria: [{ name: 'Impact', score: 0 }],
      },
      {
        id: 'e-fallback-2',
        employee_name: 'Unknown',
        cycle_name: 'Fallback Cycle',
        status: 'mystery',
      },
    ]);

    render(<HREvaluationsPage />);
    expect(screen.getAllByText('hr.evaluations.statuses.draft').length).toBeGreaterThan(0);
    expect(screen.getAllByText('hr.evaluations.statuses.mystery').length).toBeGreaterThan(0);
    expect(screen.getByText((_, node) => node?.textContent === '-/5' || false)).toBeInTheDocument();
  });

  it('covers cycle tab empty branch', () => {
    setQueryData(['hr-evaluation-cycles'], { data: [] });
    setQueryData(['hr-evaluations', { cycle_id: '', status: '' }], { data: [] });

    render(<HREvaluationsPage />);
    clickButtonByText('hr.evaluations.evaluationCycle');
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });
});

describe('HRGoalsPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-goals', { category: '', priority: '', status: '' }]);
    const { unmount } = render(<HRGoalsPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-goals', { category: '', priority: '', status: '' }], false);
    setQueryData(['hr-goals', { category: '', priority: '', status: '' }], { data: [] });
    render(<HRGoalsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers list progress delete and create branches', async () => {
    apiMocks.hr.updateGoalProgress.mockResolvedValue({});
    apiMocks.hr.deleteGoal.mockResolvedValue({});
    apiMocks.hr.createGoal.mockResolvedValue({});
    setQueryData(['hr-goals', { category: '', priority: '', status: '' }], {
      data: [
        {
          id: 'g1',
          title: 'Improve tests',
          status: 'in_progress',
          progress: 40,
          priority: 'high',
          category: 'performance',
          key_results: [{ title: 'KR1', completed: false }],
          target_date: '2026-03-01',
        },
        { id: 'g2', title: 'Finish training', status: 'completed', progress: 100, priority: 'low' },
        { id: 'g3', title: 'Cancelled goal', status: 'cancelled', progress: 20, priority: 'medium' },
      ],
    });

    render(<HRGoalsPage />);
    fireEvent.change(screen.getByDisplayValue('40') as HTMLInputElement, { target: { value: '60' } });
    await waitFor(() => {
      expect(apiMocks.hr.updateGoalProgress).toHaveBeenCalledWith('g1', 60);
    });

    (globalThis.confirm as ReturnType<typeof vi.fn>).mockReturnValueOnce(false).mockReturnValueOnce(true);
    clickButtonByIcon('delete', 0);
    clickButtonByIcon('delete', 0);
    await waitFor(() => {
      expect(apiMocks.hr.deleteGoal).toHaveBeenCalledTimes(1);
    });

    clickButtonByText('hr.goals.addGoal');
    const modalInputs = document.querySelectorAll('.fixed input');
    fireEvent.change(modalInputs[0] as HTMLInputElement, { target: { value: 'New Goal' } });
    const goalTextareas = document.querySelectorAll('.fixed textarea');
    fireEvent.change(goalTextareas[1] as HTMLTextAreaElement, {
      target: { value: 'KR-A\nKR-B' },
    });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createGoal).toHaveBeenCalledWith(
        expect.objectContaining({
          key_results: [{ title: 'KR-A', completed: false }, { title: 'KR-B', completed: false }],
        })
      );
    });
  });

  it('covers progress/color and fallback metadata branches', () => {
    setQueryData(['hr-goals', { category: '', priority: '', status: '' }], {
      data: [
        {
          id: 'g-mid',
          title: 'Mid Progress Goal',
          status: 'in_progress',
          progress: 70,
          priority: 'high',
          description: 'Need visible description',
          key_results: [{ title: 'KR done', completed: true }],
        },
        {
          id: 'g-fallback',
          title: 'Fallback Goal',
          category: '',
          progress: undefined,
        },
      ],
    });

    render(<HRGoalsPage />);
    expect(screen.getByText('Need visible description')).toBeInTheDocument();
    expect(screen.getByText('KR done')).toBeInTheDocument();
    expect(screen.getAllByText('hr.goals.statuses.not_started').length).toBeGreaterThan(0);
    expect(screen.getAllByText('70%').length).toBeGreaterThan(0);
  });
});

describe('HRTrainingPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-training', { category: '', status: '' }]);
    const { unmount } = render(<HRTrainingPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-training', { category: '', status: '' }], false);
    setQueryData(['hr-training', { category: '', status: '' }], { data: [] });
    render(<HRTrainingPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers enroll complete full and create branches', async () => {
    apiMocks.hr.enrollTraining.mockResolvedValue({});
    apiMocks.hr.completeTraining.mockResolvedValue({});
    apiMocks.hr.createTrainingProgram.mockResolvedValue({});
    setQueryData(['hr-training', { category: '', status: '' }], {
      data: [
        {
          id: 't1',
          title: 'Security',
          category: 'compliance',
          status: 'upcoming',
          start_date: '2026-02-10',
          end_date: '2026-02-11',
          max_participants: 5,
          enrolled_count: 2,
          is_enrolled: false,
        },
        {
          id: 't2',
          title: 'Leadership',
          category: 'leadership',
          status: 'in_progress',
          start_date: '2026-02-10',
          end_date: '2026-02-11',
          max_participants: 10,
          enrolled_count: 3,
          is_enrolled: true,
        },
        {
          id: 't3',
          title: 'Completed',
          category: 'technical',
          status: 'completed',
          start_date: '2026-02-10',
          end_date: '2026-02-11',
          is_enrolled: true,
        },
        {
          id: 't4',
          title: 'Full',
          category: 'technical',
          status: 'upcoming',
          start_date: '2026-02-10',
          end_date: '2026-02-11',
          max_participants: 1,
          enrolled_count: 1,
          is_enrolled: false,
        },
      ],
    });

    render(<HRTrainingPage />);
    clickButtonByText('hr.training.enroll', 0);
    clickButtonByText('hr.training.complete', 0);
    await waitFor(() => {
      expect(apiMocks.hr.enrollTraining).toHaveBeenCalledWith('t1');
      expect(apiMocks.hr.completeTraining).toHaveBeenCalledWith('t2');
    });

    clickButtonByText('hr.training.addProgram');
    fireEvent.change(document.querySelector('.fixed input') as HTMLInputElement, { target: { value: 'New Program' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createTrainingProgram).toHaveBeenCalled();
    });
  });

  it('covers category/status fallbacks and create max_participants conversion', async () => {
    apiMocks.hr.createTrainingProgram.mockResolvedValue({});
    setQueryData(['hr-training', { category: '', status: '' }], {
      data: [
        {
          id: 'tf1',
          title: 'Unknown Category',
          category: 'mystery',
          start_date: '2026-02-10',
          end_date: '2026-02-11',
          description: 'desc',
          instructor: 'Sensei',
          location: 'Room A',
          is_enrolled: false,
        },
        {
          id: 'tf2',
          title: 'Default Category',
          start_date: '2026-02-10',
          end_date: '2026-02-11',
          is_enrolled: false,
        },
      ],
    });

    render(<HRTrainingPage />);
    expect(screen.getByText('desc')).toBeInTheDocument();
    expect(screen.getByText('hr.training.instructor: Sensei')).toBeInTheDocument();
    expect(screen.getByText('Room A')).toBeInTheDocument();

    clickButtonByText('hr.training.addProgram');
    const inputs = document.querySelectorAll('.fixed input');
    fireEvent.change(inputs[0] as HTMLInputElement, { target: { value: 'Program With Cap' } });
    fireEvent.change(inputs[4] as HTMLInputElement, { target: { value: '15' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createTrainingProgram).toHaveBeenCalledWith(
        expect.objectContaining({ max_participants: 15 })
      );
    });
  });
});

describe('HRRecruitmentPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-positions']);
    setQueryData(['hr-applicants'], { data: [] });
    const { unmount } = render(<HRRecruitmentPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-positions'], false);
    setQueryData(['hr-positions'], { data: [] });
    render(<HRRecruitmentPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers position/applicant actions and stage branches', async () => {
    apiMocks.hr.createPosition.mockResolvedValue({});
    apiMocks.hr.createApplicant.mockResolvedValue({});
    apiMocks.hr.updateApplicantStage.mockResolvedValue({});

    setQueryData(['hr-positions'], {
      data: [
        { id: 'p1', title: 'Backend Engineer', department: 'Dev', status: 'open', applicant_count: 2 },
        { id: 'p2', title: 'QA Engineer', department: 'QA', status: 'closed', applicant_count: 1 },
      ],
    });
    setQueryData(['hr-applicants'], {
      data: [
        { id: 'a1', name: 'Alice', position_id: 'p1', position_title: 'Backend Engineer', email: 'a@example.com', stage: 'applied' },
        { id: 'a2', name: 'Bob', position_id: 'p1', position_title: 'Backend Engineer', email: 'b@example.com', stage: 'hired' },
        { id: 'a3', name: 'Carol', position_id: 'p2', position_title: 'QA Engineer', email: 'c@example.com', stage: 'rejected' },
      ],
    });

    render(<HRRecruitmentPage />);
    clickButtonByText('hr.recruitment.addPosition');
    fireEvent.change(document.querySelector('.fixed input') as HTMLInputElement, { target: { value: 'New Role' } });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createPosition).toHaveBeenCalled();
    });

    clickButtonByText('hr.recruitment.applicants');
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'p1' } });
    clickButtonByIcon('arrow_forward', 0);
    clickButtonByText('hr.recruitment.stages.rejected');
    await waitFor(() => {
      expect(apiMocks.hr.updateApplicantStage).toHaveBeenCalledWith('a1', 'screening');
      expect(apiMocks.hr.updateApplicantStage).toHaveBeenCalledWith('a1', 'rejected');
    });

    clickButtonByText('hr.recruitment.addApplicant');
    const selects = document.querySelectorAll('.fixed select');
    fireEvent.change(selects[0] as HTMLSelectElement, { target: { value: 'p1' } });
    fireEvent.change(document.querySelector('.fixed input') as HTMLInputElement, {
      target: { value: 'New Applicant' },
    });
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createApplicant).toHaveBeenCalled();
    });
  });

  it('covers applicant loading and empty list branches', () => {
    setQueryData(['hr-positions'], { data: [{ id: 'p1', title: 'Backend' }] });
    setLoading(['hr-applicants']);
    const { unmount } = render(<HRRecruitmentPage />);
    clickButtonByText('hr.recruitment.applicants');
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-applicants'], false);
    setQueryData(['hr-applicants'], { data: [] });
    render(<HRRecruitmentPage />);
    clickButtonByText('hr.recruitment.applicants');
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers recruitment fallback values for cards and applicant stages', () => {
    setQueryData(['hr-positions'], {
      data: [
        { id: 'pf1', title: 'Fallback Position', status: 'mystery', applicant_count: 0 },
        { id: 'pf2', title: 'Detailed Position', status: undefined, department: 'Dev', location: 'Tokyo', description: 'Hiring now', employment_type: 'fullTime', salary_range: '5m-7m', applicant_count: 2 },
      ],
    });
    setQueryData(['hr-applicants'], [
      { id: 'af1', name: '', position_id: 'pf1', position_title: '', email: '', stage: '' },
      { id: 'af2', name: 'Bob', position_id: 'pf2', position_title: 'Detailed Position', email: 'bob@example.com', stage: 'screening' },
    ]);

    render(<HRRecruitmentPage />);
    expect(screen.getByText('Hiring now')).toBeInTheDocument();
    expect(screen.getAllByText('hr.recruitment.statuses.open').length).toBeGreaterThan(0);

    clickButtonByText('hr.recruitment.applicants');
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'pf1' } });
    expect(screen.getAllByText('hr.recruitment.stages.applied').length).toBeGreaterThan(0);
  });
});

describe('HRDocumentsPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-documents', { type: '', search: '' }]);
    const { unmount } = render(<HRDocumentsPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-documents', { type: '', search: '' }], false);
    setQueryData(['hr-documents', { type: '', search: '' }], { data: [] });
    render(<HRDocumentsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers download delete filters and upload branches', async () => {
    apiMocks.hr.downloadDocument.mockResolvedValueOnce(new Blob(['a'])).mockRejectedValueOnce(new Error('x'));
    apiMocks.hr.deleteDocument.mockResolvedValue({});
    apiMocks.hr.uploadDocument.mockResolvedValue({});

    setQueryData(['hr-documents', { type: '', search: '' }], {
      data: [
        { id: 'd1', title: 'Contract A', type: 'contract', file_size: 100, uploaded_at: '2026-02-10', file_name: 'a.pdf' },
        { id: 'd2', title: 'Policy B', type: 'policy', file_size: 2048, created_at: '2026-02-11', file_name: 'b.pdf' },
        { id: 'd3', title: 'Tax C', type: 'tax', file_size: 3 * 1024 * 1024, uploaded_at: '2026-02-12', file_name: 'c.pdf' },
      ],
    });
    setQueryData(['hr-documents', { type: '', search: 'contract' }], { data: [] });
    setQueryData(['hr-documents', { type: 'contract', search: 'contract' }], { data: [] });

    render(<HRDocumentsPage />);
    clickButtonByIcon('download', 0);
    clickButtonByIcon('download', 1);

    (globalThis.confirm as ReturnType<typeof vi.fn>).mockReturnValueOnce(false).mockReturnValueOnce(true);
    clickButtonByIcon('delete', 0);
    clickButtonByIcon('delete', 0);

    fireEvent.change(screen.getByPlaceholderText('common.search'), { target: { value: 'contract' } });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'contract' } });

    clickButtonByText('hr.documents.upload', 0);
    const fileInput = document.querySelector('.fixed input[type="file"]') as HTMLInputElement;
    const file = new File(['abc'], 'upload.pdf', { type: 'application/pdf' });
    fireEvent.change(fileInput, { target: { files: [file] } });
    clickButtonByText('hr.documents.upload', 1);

    await waitFor(() => {
      expect(apiMocks.hr.downloadDocument).toHaveBeenCalledTimes(2);
      expect(apiMocks.hr.deleteDocument).toHaveBeenCalledTimes(1);
      expect(apiMocks.hr.uploadDocument).toHaveBeenCalled();
    });
  });

  it('covers document fallback fields and no-file upload guard', async () => {
    apiMocks.hr.downloadDocument.mockResolvedValue(new Blob(['x']));
    setQueryData(['hr-documents', { type: '', search: '' }], {
      data: [
        {
          id: 'df1',
          title: 'No Type File',
          type: '',
          description: 'Doc description',
          file_name: '',
          uploaded_at: '',
          created_at: '',
        },
        {
          id: 'df2',
          title: 'Unknown Type',
          type: 'mystery',
          file_size: 0,
        },
      ],
    });
    setQueryData(['hr-documents', { type: 'contract', search: 'zzz' }], { data: [] });

    render(<HRDocumentsPage />);
    expect(screen.getAllByText('Doc description').length).toBeGreaterThan(0);
    clickButtonByIcon('download', 0);
    await waitFor(() => {
      expect(apiMocks.hr.downloadDocument).toHaveBeenCalledWith('df1');
    });

    clickButtonByText('hr.documents.upload', 0);
    clickButtonByText('hr.documents.upload', 1);
    expect(apiMocks.hr.uploadDocument).not.toHaveBeenCalled();

    fireEvent.change(screen.getByPlaceholderText('common.search'), { target: { value: 'zzz' } });
    fireEvent.change(screen.getAllByRole('combobox')[0] as HTMLSelectElement, { target: { value: 'contract' } });
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });
});

describe('HRAnnouncementsPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-announcements', { priority: '' }]);
    const { unmount } = render(<HRAnnouncementsPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-announcements', { priority: '' }], false);
    setQueryData(['hr-announcements', { priority: '' }], { data: [] });
    render(<HRAnnouncementsPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers create edit delete target branches', async () => {
    apiMocks.hr.createAnnouncement.mockResolvedValue({});
    apiMocks.hr.updateAnnouncement.mockResolvedValue({});
    apiMocks.hr.deleteAnnouncement.mockResolvedValue({});
    setQueryData(['hr-announcements', { priority: '' }], {
      data: [
        {
          id: 'a1',
          title: 'Pinned',
          content: 'Important',
          priority: 'high',
          target: 'department',
          target_value: 'Dev',
          is_pinned: true,
          author_name: 'Admin',
          created_at: '2026-02-10',
        },
        {
          id: 'a2',
          title: 'Normal',
          content: 'FYI',
          priority: 'normal',
          target: 'all',
          is_pinned: false,
          author_name: 'HR',
          created_at: '2026-02-09',
        },
      ],
    });

    render(<HRAnnouncementsPage />);
    expect(screen.getByText('Pinned')).toBeInTheDocument();

    clickButtonByText('hr.announcements.addAnnouncement');
    const modalInputs = document.querySelectorAll('.fixed input');
    fireEvent.change(modalInputs[0] as HTMLInputElement, { target: { value: 'Created' } });
    fireEvent.change(document.querySelector('.fixed textarea') as HTMLTextAreaElement, { target: { value: 'Body' } });
    const modalSelects = document.querySelectorAll('.fixed select');
    fireEvent.change(modalSelects[1] as HTMLSelectElement, { target: { value: 'department' } });
    fireEvent.change(modalInputs[1] as HTMLInputElement, { target: { value: 'Platform' } });
    fireEvent.click(document.querySelector('.fixed input[type="checkbox"]') as HTMLInputElement);
    clickButtonByText('common.create');
    await waitFor(() => {
      expect(apiMocks.hr.createAnnouncement).toHaveBeenCalled();
    });

    clickButtonByIcon('edit', 0);
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.updateAnnouncement).toHaveBeenCalled();
    });

    (globalThis.confirm as ReturnType<typeof vi.fn>).mockReturnValueOnce(false).mockReturnValueOnce(true);
    clickButtonByIcon('delete', 0);
    clickButtonByIcon('delete', 0);
    await waitFor(() => {
      expect(apiMocks.hr.deleteAnnouncement).toHaveBeenCalledTimes(1);
    });
  });

  it('covers announcement sort and fallback field branches', async () => {
    apiMocks.hr.updateAnnouncement.mockResolvedValue({});
    setQueryData(['hr-announcements', { priority: '' }], {
      data: [
        { id: 'u1', title: 'Unpinned', content: 'Regular', is_pinned: false },
        { id: 'p1', title: 'Pinned Top', content: 'Priority', is_pinned: true, priority: 'urgent', target: 'department', target_value: 'Dev' },
        { id: 'f1', title: '', content: '', priority: '', target: '', target_value: '', is_pinned: undefined, author_name: '', created_at: '' },
      ],
    });

    render(<HRAnnouncementsPage />);
    expect(screen.getByText('Pinned Top')).toBeInTheDocument();
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);

    clickButtonByIcon('edit', 2);
    const modalInputs = document.querySelectorAll('.fixed input');
    fireEvent.change(modalInputs[0] as HTMLInputElement, { target: { value: 'Fallback Edited' } });
    fireEvent.change(document.querySelector('.fixed textarea') as HTMLTextAreaElement, { target: { value: 'Edited body' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.updateAnnouncement).toHaveBeenCalledWith(
        'f1',
        expect.objectContaining({
          title: 'Fallback Edited',
          content: 'Edited body',
          priority: 'normal',
          target: 'all',
          target_value: '',
          is_pinned: false,
        })
      );
    });
  });
});

describe('HRAttendanceIntegrationPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-attendance-integration', 'thisMonth', '']);
    setQueryData(['hr-attendance-alerts'], { data: [] });
    setQueryData(['hr-attendance-trend', 'thisMonth'], { data: [] });
    const { unmount } = render(<HRAttendanceIntegrationPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-attendance-integration', 'thisMonth', ''], false);
    setQueryData(['hr-attendance-integration', 'thisMonth', ''], { data: { employees: [] } });
    render(<HRAttendanceIntegrationPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers alerts trend table and filters branches', () => {
    setQueryData(['hr-attendance-integration', 'thisMonth', ''], {
      data: {
        avg_work_hours: 8.1,
        total_overtime: 123,
        late_rate: 4.5,
        leave_usage: 11.2,
        employees: [
          {
            id: 'e1',
            name: 'Alice',
            department: 'Dev',
            overtime_hours: 10,
            late_count: 1,
            leave_usage: 20,
            absent_days: 0,
            risk_level: 'high',
          },
        ],
      },
    });
    setQueryData(['hr-attendance-alerts'], {
      data: [
        { type: 'overtime', severity: 'high', employee_name: 'Alice', message: 'Over threshold' },
        { type: 'late', severity: 'low', employee_name: 'Bob', message: 'Late arrival' },
      ],
    });
    setQueryData(['hr-attendance-trend', 'thisMonth'], {
      data: [
        { label: 'W1', overtime_hours: 5 },
        { label: 'W2', overtime_hours: 10 },
      ],
    });
    setQueryData(['hr-attendance-integration', 'lastMonth', 'Dev'], { data: { employees: [] } });

    render(<HRAttendanceIntegrationPage />);
    expect(
      screen.getAllByText(
        (_, node) => node?.textContent?.includes('hr.attendanceIntegration.alerts') ?? false
      ).length
    ).toBeGreaterThan(0);
    expect(screen.getByText('hr.attendanceIntegration.trend')).toBeInTheDocument();
    expect(screen.getAllByText('Alice').length).toBeGreaterThan(0);

    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'lastMonth' } });
    fireEvent.change(screen.getByPlaceholderText('hr.attendanceIntegration.department'), {
      target: { value: 'Dev' },
    });
    expect(apiMocks.hr.getAttendanceIntegration).toHaveBeenCalledWith({ period: 'lastMonth', department: 'Dev' });
  });

  it('covers attendance fallback payload and default values', () => {
    setQueryData(['hr-attendance-integration', 'thisMonth', ''], {
      employees: [
        {
          id: 'emp-fallback',
          name: '',
          department: '',
          overtime_hours: 0,
          late_count: 0,
          leave_usage: 0,
        },
      ],
    });
    setQueryData(['hr-attendance-alerts'], [
      { type: 'other', severity: undefined, employee_name: '', message: '' },
    ]);
    setQueryData(['hr-attendance-trend', 'thisMonth'], [
      { label: '', overtime_hours: 0 },
      {},
    ]);

    render(<HRAttendanceIntegrationPage />);
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);
    expect(screen.getAllByText('0h').length).toBeGreaterThan(0);
    expect(screen.getAllByText('hr.attendanceIntegration.riskLevels.medium').length).toBeGreaterThan(0);
    expect(screen.getAllByText('hr.attendanceIntegration.riskLevels.none').length).toBeGreaterThan(0);
  });
});

describe('HROrgChartPage', () => {
  it('covers loading and no-data by search branches', () => {
    setLoading(['hr-org-chart']);
    const { unmount } = render(<HROrgChartPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-org-chart'], false);
    setQueryData(['hr-org-chart'], {
      data: { id: 'root', name: 'CEO', position: 'CEO', department: 'HQ', children: [] },
    });
    render(<HROrgChartPage />);
    fireEvent.change(screen.getByPlaceholderText('hr.orgChart.searchPlaceholder'), {
      target: { value: 'not-found' },
    });
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers array-to-tree conversion, flat mode and expand toggle branches', () => {
    setQueryData(['hr-org-chart'], {
      data: [
        {
          id: 'd1',
          name: 'Engineering',
          parent_id: null,
          employees: [{ id: 'u1', name: 'Alice', position: 'Mgr' }],
        },
        {
          id: 'd2',
          name: 'Platform',
          parent_id: 'd1',
          employees: [{ id: 'u2', name: 'Bob', position: 'Lead' }],
        },
        {
          id: 'd3',
          name: 'HR',
          parent_id: null,
          employees: [],
        },
      ],
    });

    render(<HROrgChartPage />);
    expect(screen.getAllByText('Engineering').length).toBeGreaterThan(0);
    clickButtonByText('hr.orgChart.flatView');
    expect(screen.getAllByText('Platform').length).toBeGreaterThan(0);
    fireEvent.click(screen.getAllByText('Engineering')[0]);
  });

  it('covers single-root conversion and filtered-children branch', () => {
    setQueryData(['hr-org-chart'], {
      data: [
        {
          id: 'root-only',
          parent_id: null,
          employees: [{ id: 'u1' }],
        },
        {
          id: 'child-only',
          parent_id: 'root-only',
          name: 'Child Dept',
          employees: [{ id: 'u2', name: '', position: '' }],
        },
      ],
    });

    render(<HROrgChartPage />);
    fireEvent.change(screen.getByPlaceholderText('hr.orgChart.searchPlaceholder'), {
      target: { value: 'Child' },
    });
    expect(screen.getAllByText('Child Dept').length).toBeGreaterThan(0);
    clickButtonByText('hr.orgChart.flatView');
    expect(screen.getAllByText('Child Dept').length).toBeGreaterThan(0);
  });
});

describe('HROneOnOnePage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-one-on-ones', '']);
    const { unmount } = render(<HROneOnOnePage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-one-on-ones', ''], false);
    setQueryData(['hr-one-on-ones', ''], { data: [] });
    render(<HROneOnOnePage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers create detail toggle-action and delete branches', async () => {
    apiMocks.hr.createOneOnOne.mockResolvedValue({});
    apiMocks.hr.toggleActionItem.mockResolvedValue({});
    apiMocks.hr.deleteOneOnOne.mockResolvedValue({});
    setQueryData(['hr-one-on-ones', ''], {
      data: [
        {
          id: 'm1',
          employee_name: 'Alice',
          scheduled_date: '2026-02-10T10:00:00',
          status: 'scheduled',
          frequency: 'weekly',
          mood: 'good',
          agenda: 'Agenda',
          notes: 'Notes',
          action_items: [{ id: 'a1', title: 'Follow up', completed: false }],
        },
      ],
    });

    render(<HROneOnOnePage />);
    clickButtonByText('hr.oneOnOne.schedule');
    const participantInput = screen.getByPlaceholderText('hr.oneOnOne.participant');
    fireEvent.change(participantInput, { target: { value: 'emp-2' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.createOneOnOne).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByText('Alice'));
    fireEvent.click(document.querySelector('.fixed button.size-5') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.hr.toggleActionItem).toHaveBeenCalledWith('m1', 'a1');
    });

    clickButtonByText('common.delete');
    await waitFor(() => {
      expect(apiMocks.hr.deleteOneOnOne).toHaveBeenCalledWith('m1');
    });
  });

  it('covers one-on-one fallback fields and unknown status/mood branches', async () => {
    apiMocks.hr.toggleActionItem.mockResolvedValue({});
    setQueryData(['hr-one-on-ones', ''], {
      data: [
        {
          id: 'm-fallback',
          employee_name: 'Modal User',
          scheduled_date: '',
          status: 'mystery',
          frequency: '',
          mood: 'odd',
          action_items: [
            { id: 'ia', title: '', completed: true },
            { id: 'ib', title: 'Todo', completed: false },
          ],
        },
        {
          id: 'm-empty',
          employee_name: '',
          status: 'mystery',
          frequency: '',
          mood: 'odd',
        },
      ],
    });

    render(<HROneOnOnePage />);
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);
    expect(
      screen.getAllByText((_, node) => node?.textContent?.includes('hr.oneOnOne.frequencies.biweekly') ?? false).length
    ).toBeGreaterThan(0);
    fireEvent.click(screen.getByText('Modal User'));
    fireEvent.click(document.querySelector('.fixed button.size-5') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.hr.toggleActionItem).toHaveBeenCalledWith('m-fallback', 'ia');
    });
  });
});

describe('HRSkillMapPage', () => {
  it('covers loading and no-data gap branches', () => {
    setLoading(['hr-skill-map', '']);
    setQueryData(['hr-skill-gap', ''], { data: [] });
    const { unmount } = render(<HRSkillMapPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-skill-map', ''], false);
    setQueryData(['hr-skill-map', ''], { data: [] });
    render(<HRSkillMapPage />);
    clickButtonByText('hr.skillMap.gapAnalysis');
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers map add-skill and gap-analysis branches', async () => {
    apiMocks.hr.addEmployeeSkill.mockResolvedValue({});
    setQueryData(['hr-skill-map', ''], {
      data: [
        {
          id: 'e1',
          name: 'Alice',
          position: 'Engineer',
          department: 'Dev',
          skills: [
            { skill_name: 'Go', category: 'technical', level: 5 },
            { skill_name: 'Review', category: 'soft', level: 0 },
          ],
        },
        { id: 'e2', name: 'Bob', position: 'Engineer', department: 'Dev', skills: [] },
      ],
    });
    setQueryData(['hr-skill-gap', ''], {
      data: [{ skill_name: 'Leadership', category: 'management', current_avg: 2.3, required_level: 4, gap: 1.7 }],
    });

    render(<HRSkillMapPage />);
    clickButtonByText('hr.skillMap.addSkill', 0);
    fireEvent.change(screen.getByPlaceholderText('hr.skillMap.skillName'), {
      target: { value: 'Design' },
    });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.addEmployeeSkill).toHaveBeenCalledWith(
        'e1',
        expect.objectContaining({ skill_name: 'Design' })
      );
    });

    clickButtonByText('hr.skillMap.gapAnalysis');
    expect(screen.getByText('Leadership')).toBeInTheDocument();
  });

  it('covers skill and gap fallback values', async () => {
    apiMocks.hr.addEmployeeSkill.mockResolvedValue({});
    setQueryData(['hr-skill-map', ''], [
      {
        id: 'sm1',
        name: '',
        position: '',
        department: '',
        skills: [{ category: '', skill_name: '', level: 10 }],
      },
      {
        id: 'sm2',
        name: 'NoSkills',
        position: 'Engineer',
        department: 'Dev',
        skills: { not: 'array' },
      },
    ]);
    setQueryData(['hr-skill-gap', ''], [
      { gap: 0 },
      { skill_name: 'Need More Leadership', category: 'domain', current_avg: 1, required_level: 3, gap: 2 },
    ]);

    render(<HRSkillMapPage />);
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);
    clickButtonByText('hr.skillMap.addSkill', 0);
    fireEvent.change(screen.getByPlaceholderText('hr.skillMap.skillName'), {
      target: { value: 'Communication' },
    });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.addEmployeeSkill).toHaveBeenCalledWith(
        'sm1',
        expect.objectContaining({ skill_name: 'Communication' })
      );
    });

    clickButtonByText('hr.skillMap.gapAnalysis');
    expect(screen.getByText('Need More Leadership')).toBeInTheDocument();
  });
});

describe('HRSalarySimulatorPage', () => {
  it('covers fallback and full data simulation branches', async () => {
    const { unmount } = render(<HRSalarySimulatorPage />);
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);
    unmount();

    apiMocks.hr.simulateSalary.mockResolvedValue({
      data: {
        proposed_salary: 5500000,
        percentile: 72,
        explanation: 'Based on score and years',
      },
    });
    setQueryData(['hr-salary-overview', ''], {
      data: {
        avg_salary: 4000000,
        median_salary: 3800000,
        total_payroll: 80000000,
        headcount: 20,
        department_breakdown: [{ name: 'Dev', avg_salary: 4500000, headcount: 10 }],
      },
    });
    setQueryData(['hr-budget-overview', ''], {
      data: {
        total_budget: 100000000,
        used_budget: 70000000,
        departments: [{ name: 'Dev', usage_rate: 70 }],
      },
    });

    render(<HRSalarySimulatorPage />);
    fireEvent.change(screen.getByPlaceholderText('hr.salary.grade'), { target: { value: 'G5' } });
    fireEvent.change(screen.getByPlaceholderText('hr.salary.position'), { target: { value: 'Senior' } });
    fireEvent.change(screen.getByPlaceholderText('hr.salary.evaluationScore'), { target: { value: '4.5' } });
    fireEvent.change(screen.getByPlaceholderText('hr.salary.yearsOfService'), { target: { value: '6' } });
    clickButtonByText('hr.salary.runSimulation');

    await waitFor(() => {
      expect(apiMocks.hr.simulateSalary).toHaveBeenCalled();
    });
    expect(screen.getByText('Based on score and years')).toBeInTheDocument();
    expect(screen.getAllByText('Dev').length).toBeGreaterThan(0);
  });

  it('covers simulator/budget fallback values and response shapes', async () => {
    apiMocks.hr.simulateSalary
      .mockResolvedValueOnce({ proposed_salary: 0, percentile: 0 })
      .mockResolvedValueOnce(null);
    setQueryData(['hr-salary-overview', ''], {
      data: {
        department_breakdown: [{}, { name: '', avg_salary: 0, headcount: 0 }],
      },
    });
    setQueryData(['hr-budget-overview', ''], {
      data: {
        total_budget: 5000000,
        used_budget: 0,
        departments: [{}, { name: 'Ops', usage_rate: 130 }],
      },
    });

    render(<HRSalarySimulatorPage />);
    clickButtonByText('hr.salary.runSimulation');
    await waitFor(() => {
      expect(apiMocks.hr.simulateSalary).toHaveBeenCalledTimes(1);
    });
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);

    clickButtonByText('hr.salary.runSimulation');
    await waitFor(() => {
      expect(apiMocks.hr.simulateSalary).toHaveBeenCalledTimes(2);
    });
    expect(screen.getAllByText('¥0').length).toBeGreaterThan(0);
  });
});

describe('HROnboardingPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-onboardings', '']);
    setQueryData(['hr-onboarding-templates'], { data: [] });
    const { unmount } = render(<HROnboardingPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-onboardings', ''], false);
    setQueryData(['hr-onboardings', ''], { data: [] });
    render(<HROnboardingPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers create and toggle-task branches', async () => {
    apiMocks.hr.createOnboarding.mockResolvedValue({});
    apiMocks.hr.toggleOnboardingTask.mockResolvedValue({});
    apiMocks.hr.getOnboarding.mockResolvedValue({
      data: {
        id: 'o1',
        employee_name: 'Alice',
        start_date: '2026-02-10',
        tasks: [{ id: 't1', title: 'Set account', category: 'it', completed: true }],
      },
    });
    setQueryData(['hr-onboarding-templates'], { data: [{ id: 'tmpl1', name: 'Default' }] });
    setQueryData(['hr-onboardings', ''], {
      data: [
        {
          id: 'o1',
          employee_name: 'Alice',
          start_date: '2026-02-10',
          status: 'in_progress',
          tasks: [{ id: 't1', title: 'Set account', category: 'it', completed: false }],
        },
      ],
    });

    render(<HROnboardingPage />);
    clickButtonByText('hr.onboarding.startOnboarding');
    fireEvent.change(screen.getByPlaceholderText('hr.onboarding.employee'), { target: { value: 'emp-3' } });
    fireEvent.change(screen.getByPlaceholderText('hr.onboarding.mentor'), { target: { value: 'mentor-1' } });
    fireEvent.change(document.querySelector('.fixed input[type="date"]') as HTMLInputElement, { target: { value: '2026-02-12' } });
    fireEvent.change(document.querySelector('.fixed select') as HTMLSelectElement, { target: { value: 'tmpl1' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.createOnboarding).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByText('Alice'));
    fireEvent.click(document.querySelector('.fixed button.size-5') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.hr.toggleOnboardingTask).toHaveBeenCalledWith('o1', 't1');
      expect(apiMocks.hr.getOnboarding).toHaveBeenCalledWith('o1');
    });
  });

  it('covers onboarding fallback fields and task aggregates', async () => {
    apiMocks.hr.toggleOnboardingTask.mockResolvedValue({});
    apiMocks.hr.getOnboarding.mockResolvedValue({
      id: 'ofb',
      employee_name: '',
      start_date: '',
      mentor_name: 'Mentor A',
      tasks: { invalid: true },
    });
    setQueryData(['hr-onboarding-templates'], [{ id: 'tmpl-empty' }, { id: 'tmpl-ok', name: 'Template B' }]);
    setQueryData(['hr-onboardings', ''], {
      data: [
        {
          id: 'ofb',
          employee_name: '',
          start_date: '',
          status: 'mystery',
          department: 'Dev',
          tasks: [
            { id: 'tk1', title: '', category: '', completed: true, due_date: '2026-03-01' },
            { id: 'tk2', title: 'Todo', completed: false },
          ],
        },
        {
          id: 'ob2',
          employee_name: 'Bob',
          status: undefined,
          tasks: { invalid: true },
        },
      ],
    });

    render(<HROnboardingPage />);
    expect(screen.getAllByText('?').length).toBeGreaterThan(0);
    fireEvent.click(document.querySelectorAll('.glass-card.rounded-2xl.p-4.cursor-pointer')[0] as Element);
    fireEvent.click(document.querySelector('.fixed button.size-5') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.hr.toggleOnboardingTask).toHaveBeenCalledWith('ofb', 'tk1');
      expect(apiMocks.hr.getOnboarding).toHaveBeenCalledWith('ofb');
    });
    expect(screen.getAllByText((_, node) => node?.textContent?.includes('Mentor A') ?? false).length).toBeGreaterThan(0);
  });
});

describe('HROffboardingPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-offboardings', '']);
    const { unmount } = render(<HROffboardingPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-offboardings', ''], false);
    setQueryData(['hr-offboardings', ''], { data: [] });
    render(<HROffboardingPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers analytics create and checklist branches', async () => {
    apiMocks.hr.createOffboarding.mockResolvedValue({});
    apiMocks.hr.toggleOffboardingChecklist.mockResolvedValue({});
    apiMocks.hr.getOffboarding.mockResolvedValue({
      data: {
        id: 'f1',
        employee_name: 'Alice',
        reason: 'resignation',
        last_working_date: '2026-02-20',
        checklist: [{ key: 'asset_return', completed: true }],
        notes: 'Exit interview notes',
      },
    });

    setQueryData(['hr-offboardings', ''], {
      data: [
        {
          id: 'f1',
          employee_name: 'Alice',
          reason: 'resignation',
          last_working_date: '2026-02-20',
          status: 'pending',
          checklist: [{ key: 'asset_return', completed: false }],
        },
      ],
    });
    setQueryData(['hr-turnover-analytics'], {
      data: {
        turnover_rate: 12.5,
        total_departures: 5,
        avg_tenure: '3.2y',
        reason_breakdown: [{ reason: 'resignation', count: 3 }],
        department_breakdown: [{ name: 'Dev', count: 2, rate: 10 }],
      },
    });

    render(<HROffboardingPage />);
    clickButtonByText('hr.offboarding.analytics');
    expect(screen.getByText('12.5%')).toBeInTheDocument();

    clickButtonByText('hr.offboarding.startOffboarding');
    fireEvent.change(screen.getByPlaceholderText('hr.offboarding.employee'), {
      target: { value: 'emp-5' },
    });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.createOffboarding).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByText('Alice'));
    fireEvent.click(document.querySelector('.fixed button.size-5') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.hr.toggleOffboardingChecklist).toHaveBeenCalledWith('f1', 'asset_return');
      expect(apiMocks.hr.getOffboarding).toHaveBeenCalledWith('f1');
    });
  });

  it('covers offboarding fallback analytics and detail branches', async () => {
    apiMocks.hr.toggleOffboardingChecklist.mockResolvedValue({});
    apiMocks.hr.getOffboarding.mockResolvedValue({
      id: 'of1',
      employee_name: '',
      reason: '',
      last_working_date: '',
      checklist: { invalid: true },
      notes: '',
    });
    setQueryData(['hr-offboardings', ''], {
      data: [
        {
          id: 'of1',
          employee_name: '',
          reason: '',
          last_working_date: '',
          status: 'mystery',
          checklist: [{ key: 'custom_item', label: 'Custom Label', completed: true }],
        },
        {
          id: 'of2',
          employee_name: 'Bob',
          checklist: { invalid: true },
        },
      ],
    });
    setQueryData(['hr-turnover-analytics'], {
      data: {
        turnover_rate: 0,
        total_departures: 0,
        avg_tenure: '',
        reason_breakdown: [{ reason: '', count: 0 }],
        department_breakdown: [{ name: '', count: 0, rate: 0 }],
      },
    });

    render(<HROffboardingPage />);
    clickButtonByText('hr.offboarding.analytics');
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);
    fireEvent.click(document.querySelectorAll('.glass-card.rounded-2xl.p-4.cursor-pointer')[0] as Element);
    fireEvent.click(document.querySelector('.fixed button.size-5') as HTMLButtonElement);
    await waitFor(() => {
      expect(apiMocks.hr.toggleOffboardingChecklist).toHaveBeenCalledWith('of1', 'custom_item');
      expect(apiMocks.hr.getOffboarding).toHaveBeenCalledWith('of1');
    });
  });
});

describe('HRSurveyPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['hr-surveys', '']);
    const { unmount } = render(<HRSurveyPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();

    setLoading(['hr-surveys', ''], false);
    setQueryData(['hr-surveys', ''], { data: [] });
    render(<HRSurveyPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers create publish close results and delete branches', async () => {
    apiMocks.hr.createSurvey.mockResolvedValue({});
    apiMocks.hr.publishSurvey.mockResolvedValue({});
    apiMocks.hr.closeSurvey.mockResolvedValue({});
    apiMocks.hr.deleteSurvey.mockResolvedValue({});

    setQueryData(['hr-surveys', ''], {
      data: [
        {
          id: 's1',
          title: 'Draft Survey',
          type: 'engagement',
          status: 'draft',
          questions: [{ id: 'q1' }],
          response_count: 0,
        },
        {
          id: 's2',
          title: 'Active Survey',
          type: 'pulse',
          status: 'active',
          questions: [{ id: 'q1' }, { id: 'q2' }],
          response_count: 10,
        },
        {
          id: 's3',
          title: 'Closed Survey',
          type: 'satisfaction',
          status: 'closed',
          questions: [{ id: 'q1' }],
          response_count: 5,
        },
      ],
    });
    setQueryData(['hr-survey-results', 's2'], {
      data: {
        response_rate: 70,
        avg_score: 4.2,
        enps: 20,
        questions: [{ text: 'Q1', avg_score: 4.5, responses: 10 }],
      },
    });

    render(<HRSurveyPage />);
    clickButtonByText('hr.survey.createSurvey');
    fireEvent.change(screen.getByPlaceholderText('hr.survey.surveyTitle'), { target: { value: 'Created Survey' } });
    fireEvent.change(screen.getByPlaceholderText('hr.survey.questionsPlaceholder'), { target: { value: 'Question 1\nQuestion 2' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.hr.createSurvey).toHaveBeenCalledWith(
        expect.objectContaining({ questions: [{ id: 'q1', text: 'Question 1', type: 'rating' }, { id: 'q2', text: 'Question 2', type: 'rating' }] })
      );
    });

    clickButtonByText('hr.survey.publish');
    clickButtonByText('hr.survey.close');
    clickButtonByText('hr.survey.results', 0);
    await waitFor(() => {
      expect(apiMocks.hr.publishSurvey).toHaveBeenCalledWith('s1');
      expect(apiMocks.hr.closeSurvey).toHaveBeenCalledWith('s2');
    });
    expect(screen.getByText('Q1')).toBeInTheDocument();

    clickButtonByIcon('delete', 0);
    await waitFor(() => {
      expect(apiMocks.hr.deleteSurvey).toHaveBeenCalledWith('s1');
    });
  });

  it('covers survey/result fallback fields and unknown card styles', () => {
    setQueryData(['hr-surveys', ''], {
      data: [
        {
          id: 'sv-unknown',
          title: '',
          type: 'unknown',
          status: 'mystery',
          description: 'Unknown survey description',
          questions: 'not-array',
        },
        {
          id: 'sv-active',
          title: '',
          type: '',
          status: 'active',
          questions: 'not-array',
          response_count: 0,
        },
      ],
    });
    setQueryData(['hr-survey-results', 'sv-active'], {
      data: {
        response_rate: 0,
        avg_score: 0,
        enps: null,
        questions: [{ avg_score: 0, responses: 0 }],
      },
    });

    render(<HRSurveyPage />);
    expect(screen.getByText('Unknown survey description')).toBeInTheDocument();
    expect(screen.getAllByText('hr.survey.statuses.mystery').length).toBeGreaterThan(0);
    clickButtonByText('hr.survey.results');
    expect(screen.getByText('Q1')).toBeInTheDocument();
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);
  });
});
