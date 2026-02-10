import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { ExpenseDashboardPage } from '../expenses/ExpenseDashboardPage';
import { ExpenseNewPage } from '../expenses/ExpenseNewPage';
import { ExpenseHistoryPage } from '../expenses/ExpenseHistoryPage';
import { ExpenseApprovePage } from '../expenses/ExpenseApprovePage';
import { ExpenseDetailPage } from '../expenses/ExpenseDetailPage';
import { ExpenseReportPage } from '../expenses/ExpenseReportPage';
import { ExpenseTemplatesPage } from '../expenses/ExpenseTemplatesPage';
import { ExpensePolicyPage } from '../expenses/ExpensePolicyPage';
import { ExpenseNotificationsPage } from '../expenses/ExpenseNotificationsPage';
import { ExpenseAdvancedApprovePage } from '../expenses/ExpenseAdvancedApprovePage';

type Role = 'admin' | 'manager' | 'employee';
type MockUser = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: Role;
  is_active: boolean;
};

const state = vi.hoisted(() => ({
  queryData: new Map<string, unknown>(),
  loadingKeys: new Set<string>(),
  user: {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'employee' as Role,
    is_active: true,
  } as MockUser | null,
  language: 'ja',
  params: { expenseId: 'exp-1' },
  navigate: vi.fn(),
  invalidateQueries: vi.fn(),
  mutationPending: false,
  mutationError: false,
  mutationErrorValue: new Error('mutation failed'),
}));

const apiMocks = vi.hoisted(() => ({
  expenses: {
    getStats: vi.fn(),
    getList: vi.fn(),
    getPending: vi.fn(),
    getTemplates: vi.fn(),
    create: vi.fn(),
    uploadReceipt: vi.fn(),
    approve: vi.fn(),
    advancedApprove: vi.fn(),
    getApprovalFlowConfig: vi.fn(),
    getDelegates: vi.fn(),
    setDelegate: vi.fn(),
    removeDelegate: vi.fn(),
    getByID: vi.fn(),
    getComments: vi.fn(),
    getHistory: vi.fn(),
    addComment: vi.fn(),
    delete: vi.fn(),
    update: vi.fn(),
    getReport: vi.fn(),
    getMonthlyTrend: vi.fn(),
    exportCSV: vi.fn(),
    exportPDF: vi.fn(),
    getNotifications: vi.fn(),
    getReminders: vi.fn(),
    markNotificationRead: vi.fn(),
    markAllNotificationsRead: vi.fn(),
    dismissReminder: vi.fn(),
    getNotificationSettings: vi.fn(),
    updateNotificationSettings: vi.fn(),
    getPolicies: vi.fn(),
    getBudgets: vi.fn(),
    createPolicy: vi.fn(),
    updatePolicy: vi.fn(),
    deletePolicy: vi.fn(),
    getPolicyViolations: vi.fn(),
    createTemplate: vi.fn(),
    updateTemplate: vi.fn(),
    deleteTemplate: vi.fn(),
    useTemplate: vi.fn(),
  },
  users: {
    getAll: vi.fn(),
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
      if (!values) return key;
      if ('count' in values) return `${key}:${String(values.count)}`;
      if ('num' in values) return `${key}:${String(values.num)}`;
      if ('day' in values) return `${key}:${String(values.day)}`;
      if ('number' in values) return `${key}:${String(values.number)}`;
      if ('name' in values) return `${key}:${String(values.name)}`;
      return `${key}:${JSON.stringify(values)}`;
    },
    i18n: {
      language: state.language,
      changeLanguage: vi.fn(),
    },
  }),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: (options: {
    queryKey: unknown[];
    enabled?: boolean;
    queryFn?: () => unknown;
  }) => {
    if (options.enabled === false) {
      return { data: undefined, isLoading: false };
    }
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
    isError: state.mutationError,
    error: state.mutationErrorValue,
  }),
  useQueryClient: () => ({
    invalidateQueries: state.invalidateQueries,
  }),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    user: state.user,
  }),
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
    onPageChange: (page: number) => void;
    onPageSizeChange?: (size: number) => void;
  }) => (
    <div data-testid="pagination">
      <button onClick={() => onPageChange(Math.min(totalPages, currentPage + 1))}>pagination-next</button>
      <button onClick={() => onPageSizeChange?.(50)}>pagination-size</button>
    </div>
  ),
}));

vi.mock('@/api/client', () => ({
  api: apiMocks,
}));

function queryKey(key: unknown[]) {
  return JSON.stringify(key);
}

function setQueryData(key: unknown[], value: unknown) {
  state.queryData.set(queryKey(key), value);
}

function setLoading(key: unknown[], loading = true) {
  const target = queryKey(key);
  if (loading) {
    state.loadingKeys.add(target);
    return;
  }
  state.loadingKeys.delete(target);
}

function getButtonByText(text: string, index = 0) {
  const target = screen.getAllByText(text)[index];
  const button = target.closest('button');
  if (!button) {
    throw new Error(`button not found for text: ${text}`);
  }
  return button as HTMLButtonElement;
}

function clickButtonByText(text: string, index = 0) {
  fireEvent.click(getButtonByText(text, index));
}

function resetHarness() {
  state.queryData.clear();
  state.loadingKeys.clear();
  state.user = {
    id: 'u1',
    email: 'user@example.com',
    first_name: 'Taro',
    last_name: 'Yamada',
    role: 'employee',
    is_active: true,
  };
  state.language = 'ja';
  state.params = { expenseId: 'exp-1' };
  state.mutationPending = false;
  state.mutationError = false;
  state.mutationErrorValue = new Error('mutation failed');
  state.navigate.mockReset();
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
  vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
  class MockFileReader {
    onload: ((event: { target: { result: string } }) => void) | null = null;

    readAsDataURL() {
      this.onload?.({ target: { result: 'data:image/png;base64,mock' } });
    }
  }
  vi.stubGlobal('FileReader', MockFileReader as unknown as typeof FileReader);
}

beforeEach(() => {
  resetHarness();
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ExpenseDashboardPage', () => {
  it('covers empty and manager branches', () => {
    setQueryData(['expense-stats'], undefined);
    setQueryData(['expenses', 'recent'], { data: [] });

    const { unmount } = render(<ExpenseDashboardPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
    expect(screen.queryByText('expenses.approve.title')).not.toBeInTheDocument();
    unmount();

    state.user = { ...state.user!, role: 'admin' };
    state.language = 'en';
    setQueryData(['expense-stats'], {
      total_this_month: 12345,
      pending_count: 2,
      approved_this_month: 6789,
      reimbursed_total: 9999,
    });
    setQueryData(['expenses', 'recent'], {
      expenses: [
        {
          id: 'e1',
          title: 'Taxi',
          category: 'transportation',
          amount: 1200,
          status: 'pending',
          expense_date: '2026-02-01',
        },
        {
          id: 'e2',
          title: 'Unknown',
          category: 'unknown',
          amount: 300,
          status: 'mystery',
        },
      ],
    });

    render(<ExpenseDashboardPage />);
    expect(screen.getByText('expenses.approve.title')).toBeInTheDocument();
    expect(screen.getByText('expenses.status.pending')).toBeInTheDocument();
    expect(screen.getByText('expenses.status.mystery')).toBeInTheDocument();
  });
});

describe('ExpenseApprovePage', () => {
  it('covers no-permission branch', () => {
    state.user = { ...state.user!, role: 'employee' };
    render(<ExpenseApprovePage />);
    expect(screen.getByText('common.noPermission')).toBeInTheDocument();
  });

  it('covers loading and empty states', () => {
    state.user = { ...state.user!, role: 'manager' };
    setLoading(['expenses', 'pending', 1, 10]);
    const { unmount } = render(<ExpenseApprovePage />);
    expect(screen.getAllByText('common.loading').length).toBeGreaterThan(0);
    unmount();
    setLoading(['expenses', 'pending', 1, 10], false);

    setQueryData(['expenses', 'pending', 1, 10], { data: [], total: 0 });
    render(<ExpenseApprovePage />);
    expect(screen.getAllByText('expenses.approve.noPending').length).toBeGreaterThan(0);
  });

  it('covers approve/reject/pagination branches', async () => {
    state.user = { ...state.user!, role: 'admin' };
    apiMocks.expenses.approve.mockResolvedValue({});
    setQueryData(['expenses', 'pending', 1, 10], {
      data: [
        {
          id: 'e1',
          user_name: 'Taro',
          title: '交通費',
          category: 'transportation',
          expense_date: '2026-02-01',
          amount: 1500,
        },
      ],
      total: 30,
    });
    setQueryData(['expenses', 'pending', 2, 10], { data: [], total: 30 });
    setQueryData(['expenses', 'pending', 1, 50], { data: [], total: 30 });

    render(<ExpenseApprovePage />);

    clickButtonByText('common.approve', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.approve).toHaveBeenCalledWith('e1', { status: 'approved' });
    });

    clickButtonByText('common.reject', 0);
    fireEvent.change(screen.getAllByPlaceholderText('expenses.approve.rejectReason')[0], {
      target: { value: '理由' },
    });
    clickButtonByText('common.confirm', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.approve).toHaveBeenCalledWith('e1', {
        status: 'rejected',
        rejected_reason: '理由',
      });
    });

    fireEvent.click(screen.getByText('pagination-next'));
    fireEvent.click(screen.getByText('pagination-size'));
    expect(apiMocks.expenses.getPending).toHaveBeenCalledWith({ page: 2, page_size: 10 });
    expect(apiMocks.expenses.getPending).toHaveBeenCalledWith({ page: 1, page_size: 50 });
  });

  it('covers optional value and reject-cancel branches', () => {
    state.user = { ...state.user!, role: 'admin' };
    setQueryData(['expenses', 'pending', 1, 10], {
      data: [
        {
          id: 'e2',
          title: 'No Optional',
          amount: 0,
          status: 'pending',
        },
      ],
      total: 1,
    });

    render(<ExpenseApprovePage />);
    expect(screen.getAllByText('-').length).toBeGreaterThan(0);
    clickButtonByText('common.reject', 0);
    clickButtonByText('common.cancel', 0);
    expect(screen.queryAllByPlaceholderText('expenses.approve.rejectReason').length).toBe(0);
  });
});

describe('ExpenseHistoryPage', () => {
  it('covers loading and empty branches', () => {
    setLoading(['expenses', 'history', 1, 10, '', '']);
    const { unmount } = render(<ExpenseHistoryPage />);
    expect(screen.getAllByText('common.loading').length).toBeGreaterThan(0);
    unmount();
    setLoading(['expenses', 'history', 1, 10, '', ''], false);

    setQueryData(['expenses', 'history', 1, 10, '', ''], { data: [], total: 0 });
    render(<ExpenseHistoryPage />);
    expect(screen.getAllByText('common.noData').length).toBeGreaterThan(0);
  });

  it('covers list, filters and pagination branches', () => {
    setQueryData(['expenses', 'history', 1, 10, '', ''], {
      data: [
        {
          id: 'e1',
          title: 'Draft Expense',
          expense_date: '2026-02-10',
          category: 'transportation',
          amount: 1000,
          status: 'draft',
        },
        {
          id: 'e2',
          title: 'Unknown Expense',
          amount: 2000,
          status: 'mystery',
          category: 'unknown',
        },
      ],
      total: 21,
    });
    setQueryData(['expenses', 'history', 1, 10, 'approved', ''], { data: [], total: 0 });
    setQueryData(['expenses', 'history', 1, 10, 'approved', 'meals'], { data: [], total: 11 });
    setQueryData(['expenses', 'history', 2, 10, 'approved', 'meals'], { data: [], total: 11 });
    setQueryData(['expenses', 'history', 1, 50, 'approved', 'meals'], { data: [], total: 11 });

    render(<ExpenseHistoryPage />);
    expect(screen.getByText('common.edit')).toBeInTheDocument();
    expect(screen.getAllByText('expenses.status.mystery').length).toBeGreaterThan(0);

    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: 'approved' } });
    fireEvent.change(selects[1], { target: { value: 'meals' } });

    fireEvent.click(screen.getByText('pagination-next'));
    fireEvent.click(screen.getByText('pagination-size'));
    expect(apiMocks.expenses.getList).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'approved', category: 'meals' })
    );
  });
});

describe('ExpenseNewPage', () => {
  it('covers validation and submit-type branches', async () => {
    apiMocks.expenses.create.mockResolvedValue({});
    setQueryData(['expense-templates'], { data: [] });

    render(<ExpenseNewPage />);
    clickButtonByText('expenses.actions.submit');
    expect(await screen.findByText('expenses.validation.titleRequired')).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText('expenses.placeholders.title'), {
      target: { value: '経費タイトル' },
    });

    const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement;
    const amountInput = document.querySelector('input[type="number"]') as HTMLInputElement;
    const categorySelect = screen.getByRole('combobox');
    fireEvent.change(dateInput, { target: { value: '2026-02-10' } });
    fireEvent.change(categorySelect, { target: { value: 'transportation' } });
    fireEvent.change(screen.getByPlaceholderText('expenses.placeholders.description'), {
      target: { value: 'タクシー代' },
    });
    fireEvent.change(amountInput, { target: { value: '1200' } });

    clickButtonByText('expenses.actions.saveDraft');
    await waitFor(() => {
      expect(apiMocks.expenses.create).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'pending' })
      );
    });

    clickButtonByText('expenses.actions.submit');
    await waitFor(() => {
      expect(apiMocks.expenses.create).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'pending' })
      );
    });
  });

  it('covers template and receipt upload branches', async () => {
    apiMocks.expenses.uploadReceipt.mockResolvedValue({ url: 'https://example.com/r.png' });
    setQueryData(['expense-templates'], {
      data: [
        {
          id: 't1',
          name: '交通費テンプレート',
          title: '交通費',
          category: 'transportation',
          description: '移動',
          amount: 900,
        },
        {
          id: 't2',
          name: '無効テンプレート',
          category: 'other',
          description: 'invalid',
          amount: 100,
        },
      ],
    });

    render(<ExpenseNewPage />);
    clickButtonByText('expenses.receipt.fromTemplate');
    fireEvent.click(screen.getByText('交通費テンプレート'));
    expect(screen.getByDisplayValue('移動')).toBeInTheDocument();

    clickButtonByText('expenses.receipt.fromTemplate');
    fireEvent.click(screen.getByText('無効テンプレート'));
    expect(screen.getByText('無効テンプレート')).toBeInTheDocument();

    clickButtonByText('expenses.new.addItem');
    clickButtonByText('close', 0);

    const dropzone = screen.getAllByText('expenses.receipt.dragOrTap')[0].closest('div')!;
    const invalidFile = new File(['x'], 'x.txt', { type: 'text/plain' });
    fireEvent.drop(dropzone, { dataTransfer: { files: [invalidFile] } });
    expect(apiMocks.expenses.uploadReceipt).not.toHaveBeenCalled();

    const imageFile = new File(['img'], 'receipt.png', { type: 'image/png' });
    fireEvent.dragOver(dropzone);
    fireEvent.dragLeave(dropzone);
    fireEvent.drop(dropzone, { dataTransfer: { files: [imageFile] } });
    await waitFor(() => {
      expect(apiMocks.expenses.uploadReceipt).toHaveBeenCalledWith(imageFile);
    });
    expect(await screen.findByAltText('Receipt')).toBeInTheDocument();

    const pdfFile = new File(['pdf'], 'receipt.pdf', { type: 'application/pdf' });
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(fileInput, { target: { files: [pdfFile] } });
    await waitFor(() => {
      expect(apiMocks.expenses.uploadReceipt).toHaveBeenCalledWith(pdfFile);
    });

    apiMocks.expenses.uploadReceipt.mockRejectedValueOnce(new Error('upload failed'));
    fireEvent.drop(dropzone, { dataTransfer: { files: [imageFile] } });
    await waitFor(() => {
      expect(apiMocks.expenses.uploadReceipt).toHaveBeenCalledTimes(3);
    });
  });

  it('covers mutation error/pending view branches', () => {
    setQueryData(['expense-templates'], { data: [] });
    state.mutationError = true;
    state.mutationErrorValue = new Error('create failed');
    const { unmount } = render(<ExpenseNewPage />);
    expect(screen.getByText('create failed')).toBeInTheDocument();
    unmount();

    state.mutationError = false;
    state.mutationPending = true;
    render(<ExpenseNewPage />);
    expect(screen.getByText('common.submitting')).toBeInTheDocument();
  });
});

describe('ExpenseDetailPage', () => {
  it('covers loading and not-found branches', () => {
    setLoading(['expense', 'exp-1']);
    const { unmount } = render(<ExpenseDetailPage />);
    expect(screen.getByText('hourglass_empty')).toBeInTheDocument();
    unmount();
    setLoading(['expense', 'exp-1'], false);

    render(<ExpenseDetailPage />);
    expect(screen.getByText('common.noData')).toBeInTheDocument();
  });

  it('covers draft detail tabs and draft actions', async () => {
    apiMocks.expenses.update.mockResolvedValue({});
    apiMocks.expenses.delete.mockResolvedValue({});
    apiMocks.expenses.addComment.mockResolvedValue({});
    setQueryData(['expense', 'exp-1'], {
      id: 'exp-1',
      title: 'Draft Expense',
      status: 'draft',
      user_name: 'Taro',
      created_at: '2026-02-10T10:00:00Z',
      items: [
        {
          expense_date: '2026-02-10',
          category: 'transportation',
          description: 'Taxi',
          amount: 1000,
          receipt_url: 'https://example.com/r1.png',
        },
        {
          category: 'unknown',
          amount: 0,
        },
      ],
      notes: 'memo',
    });
    setQueryData(['expense-comments', 'exp-1'], {
      data: [{ id: 'c1', user_name: 'Manager', content: 'ok', created_at: '2026-02-10T11:00:00Z' }],
    });
    setQueryData(['expense-history', 'exp-1'], {
      data: [
        { action: 'created', user_name: 'Taro', details: 'first', created_at: '2026-02-10T09:00:00Z' },
        { action: 'updated', created_at: '2026-02-10T12:00:00Z' },
      ],
    });

    render(<ExpenseDetailPage />);
    clickButtonByText('expenses.actions.submit');
    await waitFor(() => {
      expect(apiMocks.expenses.update).toHaveBeenCalledWith('exp-1', { status: 'pending' });
    });

    (globalThis.confirm as ReturnType<typeof vi.fn>)
      .mockReturnValueOnce(false)
      .mockReturnValueOnce(true);
    clickButtonByText('common.delete');
    clickButtonByText('common.delete');
    await waitFor(() => {
      expect(apiMocks.expenses.delete).toHaveBeenCalledWith('exp-1');
    });

    clickButtonByText('expenses.detail.tabs.comments');
    fireEvent.change(screen.getByPlaceholderText('expenses.detail.commentPlaceholder'), {
      target: { value: '  コメント  ' },
    });
    clickButtonByText('expenses.detail.addComment');
    await waitFor(() => {
      expect(apiMocks.expenses.addComment).toHaveBeenCalledWith('exp-1', { content: 'コメント' });
    });

    clickButtonByText('expenses.detail.tabs.history');
    expect(screen.getByText('created')).toBeInTheDocument();
    clickButtonByText('expenses.detail.tabs.details');
    expect(screen.getByText('memo')).toBeInTheDocument();
  });

  it('covers manager-approval branches', async () => {
    apiMocks.expenses.approve.mockResolvedValue({});
    state.user = { ...state.user!, role: 'manager' };
    setQueryData(['expense', 'exp-1'], {
      id: 'exp-1',
      title: 'Pending Expense',
      status: 'pending',
      user_name: 'User',
      items: [],
    });
    setQueryData(['expense-comments', 'exp-1'], { data: [] });
    setQueryData(['expense-history', 'exp-1'], { data: [] });

    render(<ExpenseDetailPage />);
    expect(screen.getByText('expenses.detail.approvalAction')).toBeInTheDocument();
    clickButtonByText('common.approve', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.approve).toHaveBeenCalledWith('exp-1', { status: 'approved' });
    });

    clickButtonByText('common.reject', 0);
    const rejectTextarea = screen.getByPlaceholderText('expenses.approve.rejectReason');
    fireEvent.change(rejectTextarea, { target: { value: '  却下理由  ' } });
    clickButtonByText('expenses.detail.confirmReject');
    await waitFor(() => {
      expect(apiMocks.expenses.approve).toHaveBeenCalledWith('exp-1', {
        status: 'rejected',
        rejected_reason: '却下理由',
      });
    });

    clickButtonByText('common.reject', 0);
    clickButtonByText('common.cancel');
  });
});

describe('ExpenseReportPage', () => {
  it('covers fallback/no-data and export branches', async () => {
    apiMocks.expenses.exportCSV.mockResolvedValue(new Blob(['a,b']));
    apiMocks.expenses.exportPDF.mockRejectedValueOnce(new Error('pdf failed'));
    render(<ExpenseReportPage />);

    expect(screen.getAllByText('common.noData').length).toBeGreaterThan(0);
    clickButtonByText('expenses.report.period.quarter');
    clickButtonByText('expenses.report.period.year');

    const monthInput = document.querySelector('input[type="month"]') as HTMLInputElement;
    fireEvent.change(monthInput, { target: { value: '2025-11' } });

    clickButtonByText('CSV');
    clickButtonByText('PDF');
    await waitFor(() => {
      expect(apiMocks.expenses.exportCSV).toHaveBeenCalled();
      expect(apiMocks.expenses.exportPDF).toHaveBeenCalled();
    });
  });

  it('covers category/trend/department/status summary branches', () => {
    apiMocks.expenses.getReport.mockReturnValue({
      total_amount: 3000,
      total_count: 3,
      category_breakdown: [
        { category: 'meals', amount: 1000 },
        { category: 'unknown', amount: 500 },
      ],
      department_breakdown: [
        { department: 'Dev', amount: 2000, count: 2 },
        { department: 'HR', amount: 1000, count: 1 },
      ],
      status_summary: {
        draft: 1,
        pending: 1,
        approved: 1,
        rejected: 0,
        reimbursed: 0,
        approved_amount: 1000,
        pending_amount: 500,
      },
    });
    apiMocks.expenses.getMonthlyTrend.mockReturnValue({
      data: [
        { month: '2026-01', amount: 1000 },
        { month: '2026-02', amount: 2000 },
      ],
    });

    render(<ExpenseReportPage />);
    expect(screen.getByText('expenses.report.departmentBreakdown')).toBeInTheDocument();
    expect(screen.getByText('expenses.categories.meals')).toBeInTheDocument();
    expect(screen.getByText('2026-01')).toBeInTheDocument();
  });

  it('covers export catch/success opposite branches', async () => {
    apiMocks.expenses.exportCSV.mockRejectedValueOnce(new Error('csv failed'));
    apiMocks.expenses.exportPDF.mockResolvedValueOnce(new Blob(['pdf']));

    render(<ExpenseReportPage />);
    clickButtonByText('CSV');
    clickButtonByText('PDF');
    await waitFor(() => {
      expect(apiMocks.expenses.exportCSV).toHaveBeenCalled();
      expect(apiMocks.expenses.exportPDF).toHaveBeenCalled();
    });
  });

  it('covers pie-chart total zero branch', () => {
    apiMocks.expenses.getReport.mockReturnValue({
      total_amount: 0,
      total_count: 0,
      category_breakdown: [{ category: 'meals', amount: 0 }],
      department_breakdown: [],
      status_summary: {
        draft: 0,
        pending: 0,
        approved: 0,
        rejected: 0,
        reimbursed: 0,
        approved_amount: 0,
        pending_amount: 0,
      },
    });
    apiMocks.expenses.getMonthlyTrend.mockReturnValue({ months: [] });

    render(<ExpenseReportPage />);
    expect(screen.queryByText('expenses.fields.totalAmount')).not.toBeInTheDocument();
  });
});

describe('ExpenseTemplatesPage', () => {
  it('covers loading/no-templates branches', () => {
    setLoading(['expense-templates']);
    const { unmount } = render(<ExpenseTemplatesPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();
    setLoading(['expense-templates'], false);

    setQueryData(['expense-templates'], { data: [] });
    render(<ExpenseTemplatesPage />);
    expect(screen.getByText('expenses.templates.noTemplates')).toBeInTheDocument();
  });

  it('covers create/edit/delete/use branches', async () => {
    apiMocks.expenses.createTemplate.mockResolvedValue({});
    apiMocks.expenses.updateTemplate.mockResolvedValue({});
    apiMocks.expenses.deleteTemplate.mockResolvedValue({});
    apiMocks.expenses.useTemplate
      .mockResolvedValueOnce({ id: 'exp-generated' })
      .mockResolvedValueOnce({});
    setQueryData(['expense-templates'], {
      data: [
        {
          id: 'r1',
          name: '通常テンプレ',
          title: 'ランチ',
          category: 'meals',
          amount: 1500,
          description: '昼食',
          is_recurring: false,
        },
        {
          id: 'rec1',
          name: '定期テンプレ',
          title: '交通費',
          category: '',
          amount: 0,
          is_recurring: true,
          recurring_day: 10,
        },
      ],
    });

    render(<ExpenseTemplatesPage />);
    expect(screen.getByText('expenses.templates.recurringTemplates')).toBeInTheDocument();

    clickButtonByText('expenses.templates.create');
    fireEvent.change(screen.getByPlaceholderText('expenses.templates.templateNamePlaceholder'), {
      target: { value: '新規テンプレ' },
    });
    fireEvent.change(screen.getByPlaceholderText('expenses.placeholders.title'), {
      target: { value: '会議費' },
    });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.expenses.createTemplate).toHaveBeenCalled();
    });

    clickButtonByText('edit', 0);
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.expenses.updateTemplate).toHaveBeenCalled();
    });

    clickButtonByText('expenses.templates.useTemplate', 0);
    await waitFor(() => {
      expect(state.navigate).toHaveBeenCalledWith({
        to: '/expenses/$expenseId',
        params: { expenseId: 'exp-generated' },
      });
    });

    clickButtonByText('expenses.templates.useTemplate', 1);
    await waitFor(() => {
      expect(state.navigate).toHaveBeenCalledWith({ to: '/expenses/new' });
    });

    (globalThis.confirm as ReturnType<typeof vi.fn>)
      .mockReturnValueOnce(false)
      .mockReturnValueOnce(true);
    clickButtonByText('delete', 0);
    clickButtonByText('delete', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.deleteTemplate).toHaveBeenCalledTimes(1);
    });
  });
});

describe('ExpensePolicyPage', () => {
  it('covers non-admin branches', () => {
    state.user = { ...state.user!, role: 'employee' };
    setQueryData(['expense-policies'], { data: [] });
    const { unmount } = render(<ExpensePolicyPage />);
    expect(screen.getByText('expenses.policy.noPolicies')).toBeInTheDocument();
    unmount();

    setQueryData(['expense-policies'], {
      data: [
        { id: 'p1', category: 'meals', is_active: true, description: 'active' },
        { id: 'p2', category: 'other', is_active: false, description: 'inactive' },
      ],
    });
    render(<ExpensePolicyPage />);
    expect(screen.getByText('expenses.categories.meals')).toBeInTheDocument();
    expect(screen.queryByText('expenses.categories.other')).not.toBeInTheDocument();
  });

  it('covers admin budgets/form/edit/delete/violation branches', async () => {
    state.user = { ...state.user!, role: 'admin' };
    apiMocks.expenses.createPolicy.mockResolvedValue({});
    apiMocks.expenses.updatePolicy.mockResolvedValue({});
    apiMocks.expenses.deletePolicy.mockResolvedValue({});
    setQueryData(['expense-policies'], {
      data: [
        {
          id: 'p1',
          category: 'transportation',
          monthly_limit: 10000,
          per_claim_limit: 5000,
          auto_approve_limit: 2000,
          requires_receipt_above: 1000,
          description: 'policy',
          is_active: true,
        },
      ],
    });
    setQueryData(['expense-budgets'], {
      data: [
        { department: 'Dev', used_amount: 95000, budget_amount: 100000 },
        { department: 'HR', used_amount: 20000, budget_amount: 100000 },
      ],
    });
    setQueryData(['expense-policy-violations'], {
      data: [
        {
          user_name: 'Taro',
          expense_title: '超過',
          violation_message: 'limit exceeded',
          amount: 9999,
        },
      ],
    });

    render(<ExpensePolicyPage />);
    expect(screen.getByText('expenses.policy.budgetOverview')).toBeInTheDocument();
    expect(screen.getByText('expenses.policy.violations')).toBeInTheDocument();

    clickButtonByText('expenses.policy.addPolicy');
    const categorySelect = screen.getByRole('combobox');
    fireEvent.change(categorySelect, { target: { value: 'meals' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.expenses.createPolicy).toHaveBeenCalled();
    });

    clickButtonByText('edit', 0);
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.expenses.updatePolicy).toHaveBeenCalled();
    });

    (globalThis.confirm as ReturnType<typeof vi.fn>)
      .mockReturnValueOnce(false)
      .mockReturnValueOnce(true);
    clickButtonByText('delete', 0);
    clickButtonByText('delete', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.deletePolicy).toHaveBeenCalledTimes(1);
    });
  });

  it('covers admin loading/no-policy/inactive branches', () => {
    state.user = { ...state.user!, role: 'admin' };
    setLoading(['expense-policies']);
    const { unmount } = render(<ExpensePolicyPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();
    setLoading(['expense-policies'], false);

    setQueryData(['expense-policies'], { data: [] });
    const { unmount: unmountEmpty } = render(<ExpensePolicyPage />);
    expect(screen.getByText('expenses.policy.noPolicies')).toBeInTheDocument();
    unmountEmpty();

    setQueryData(['expense-policies'], {
      data: [
        {
          id: 'p2',
          category: 'unknown',
          monthly_limit: 0,
          per_claim_limit: 0,
          auto_approve_limit: 0,
          requires_receipt_above: 0,
          is_active: false,
        },
      ],
    });
    render(<ExpensePolicyPage />);
    expect(screen.getByText('expenses.policy.inactive')).toBeInTheDocument();
  });
});

describe('ExpenseNotificationsPage', () => {
  it('covers loading/empty and settings toggle branches', async () => {
    setLoading(['expense-notifications', 'all']);
    const { unmount } = render(<ExpenseNotificationsPage />);
    expect(screen.getByText('common.loading')).toBeInTheDocument();
    unmount();
    setLoading(['expense-notifications', 'all'], false);

    setQueryData(['expense-notifications', 'all'], { data: [] });
    setQueryData(['expense-reminders'], { data: [] });
    render(<ExpenseNotificationsPage />);
    expect(screen.getByText('expenses.notifications.allRead')).toBeInTheDocument();
    expect(screen.getAllByText('expenses.notifications.empty').length).toBeGreaterThan(0);

    const toggles = document.querySelectorAll('.relative.w-11.h-6');
    fireEvent.click(toggles[0]);
    await waitFor(() => {
      expect(apiMocks.expenses.updateNotificationSettings).toHaveBeenCalled();
    });
  });

  it('covers unread/reminder/filter/list branches', async () => {
    apiMocks.expenses.markNotificationRead.mockResolvedValue({});
    apiMocks.expenses.markAllNotificationsRead.mockResolvedValue({});
    apiMocks.expenses.dismissReminder.mockResolvedValue({});
    setQueryData(['expense-notification-settings'], {
      on_approved: false,
      on_rejected: true,
      on_comment: true,
      on_reimbursed: false,
      month_end_reminder: true,
      overdue_reminder: false,
      reminder_days_before: 3,
    });
    setQueryData(['expense-notifications', 'all'], {
      data: [
        {
          id: 'n1',
          type: 'approved',
          title: 'Approved',
          message: 'ok',
          is_read: false,
          created_at: '2026-02-10T10:00:00Z',
          expense_id: 'e1',
        },
        {
          id: 'n2',
          type: 'unknown',
          title: 'Unknown',
          message: 'x',
          is_read: true,
        },
      ],
    });
    setQueryData(['expense-notifications', 'unread'], { data: [] });
    setQueryData(['expense-notifications', 'action_required'], { data: [] });
    setQueryData(['expense-reminders'], {
      data: [
        { id: 'r1', type: 'month_end', title: 'Month End', message: 'm1', action_url: '/expenses/new' },
        { id: 'r2', type: 'overdue', title: 'Overdue', message: 'm2' },
        { id: 'r3', type: 'other', title: 'Other', message: 'm3' },
      ],
    });

    render(<ExpenseNotificationsPage />);
    expect(screen.getByText('expenses.notifications.unreadCount:1')).toBeInTheDocument();
    clickButtonByText('expenses.notifications.markAllRead');
    await waitFor(() => {
      expect(apiMocks.expenses.markAllNotificationsRead).toHaveBeenCalled();
    });

    fireEvent.click(screen.getByText('Approved'));
    await waitFor(() => {
      expect(apiMocks.expenses.markNotificationRead).toHaveBeenCalledWith('n1');
    });
    fireEvent.click(screen.getByText('Unknown'));
    expect(apiMocks.expenses.markNotificationRead).toHaveBeenCalledTimes(1);

    clickButtonByText('close', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.dismissReminder).toHaveBeenCalled();
    });

    clickButtonByText('expenses.notifications.filter.unread');
    clickButtonByText('expenses.notifications.filter.action_required');
    expect(apiMocks.expenses.getNotifications).toHaveBeenCalledWith({ filter: 'unread' });
    expect(apiMocks.expenses.getNotifications).toHaveBeenCalledWith({ filter: 'action_required' });
  });
});

describe('ExpenseAdvancedApprovePage', () => {
  it('covers no-permission/loading/empty branches', () => {
    state.user = { ...state.user!, role: 'employee' };
    const { unmount } = render(<ExpenseAdvancedApprovePage />);
    expect(screen.getByText('common.noPermission')).toBeInTheDocument();
    unmount();

    state.user = { ...state.user!, role: 'manager' };
    setLoading(['expenses', 'pending-advanced', 1, 10]);
    const { unmount: unmountLoading } = render(<ExpenseAdvancedApprovePage />);
    expect(screen.getAllByText('common.loading').length).toBeGreaterThan(0);
    unmountLoading();
    setLoading(['expenses', 'pending-advanced', 1, 10], false);

    setQueryData(['expenses', 'pending-advanced', 1, 10], { data: [], total: 0 });
    render(<ExpenseAdvancedApprovePage />);
    expect(screen.getAllByText('expenses.approve.noPending').length).toBeGreaterThan(0);
  });

  it('covers flow/delegate/actions/pagination branches', async () => {
    state.user = { ...state.user!, role: 'admin' };
    apiMocks.expenses.advancedApprove.mockResolvedValue({});
    apiMocks.expenses.setDelegate.mockResolvedValue({});
    apiMocks.expenses.removeDelegate.mockResolvedValue({});
    setQueryData(['expenses', 'pending-advanced', 1, 10], {
      data: [
        {
          id: 'e1',
          title: 'Advanced Expense',
          user_name: 'Taro',
          amount: 5000,
          status: 'pending',
          current_step: 1,
          created_at: '2026-02-10T10:00:00Z',
        },
      ],
      total: 25,
    });
    setQueryData(['expenses', 'pending-advanced', 2, 10], { data: [], total: 25 });
    setQueryData(['expenses', 'pending-advanced', 1, 50], { data: [], total: 25 });
    setQueryData(['expense-approval-flow-config'], {
      data: {
        steps: [
          { name: '一次承認', approver_role: 'manager', auto_approve_below: 1000 },
          { name: '最終承認', approver_role: 'admin', auto_approve_below: 0 },
        ],
      },
    });
    setQueryData(['expense-delegates'], {
      data: [{ id: 'd1', delegate_name: 'Hanako', start_date: '2026-02-10', end_date: '2026-02-20' }],
    });
    setQueryData(['users-for-delegate'], {
      data: [
        { id: 'u1', role: 'admin', first_name: 'Me', last_name: 'Self' },
        { id: 'u2', role: 'manager', first_name: 'Hanako', last_name: 'Suzuki' },
        { id: 'u3', role: 'employee', first_name: 'Emp', last_name: 'User' },
      ],
    });

    render(<ExpenseAdvancedApprovePage />);
    expect(screen.getByText('expenses.advancedApprove.flowSteps')).toBeInTheDocument();
    expect(
      screen.getAllByText((_, node) => node?.textContent?.includes('expenses.advancedApprove.autoApprove') ?? false)
        .length
    ).toBeGreaterThan(0);

    clickButtonByText('expenses.advancedApprove.delegate');
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: 'u2' } });
    const dateInputs = document.querySelectorAll('input[type="date"]');
    fireEvent.change(dateInputs[0], { target: { value: '2026-02-10' } });
    fireEvent.change(dateInputs[1], { target: { value: '2026-02-20' } });
    clickButtonByText('common.save');
    await waitFor(() => {
      expect(apiMocks.expenses.setDelegate).toHaveBeenCalledWith({
        delegate_to: 'u2',
        start_date: '2026-02-10',
        end_date: '2026-02-20',
      });
    });

    clickButtonByText('expenses.advancedApprove.delegate');
    clickButtonByText('delete', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.removeDelegate).toHaveBeenCalledWith('d1');
    });

    clickButtonByText('common.approve', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.advancedApprove).toHaveBeenCalledWith('e1', { action: 'approve' });
    });

    clickButtonByText('expenses.advancedApprove.return', 0);
    fireEvent.change(screen.getAllByPlaceholderText('expenses.advancedApprove.returnReason')[0], {
      target: { value: '  差し戻し  ' },
    });
    clickButtonByText('common.confirm', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.advancedApprove).toHaveBeenCalledWith('e1', {
        action: 'return',
        reason: '差し戻し',
      });
    });

    clickButtonByText('common.reject', 0);
    fireEvent.change(screen.getAllByPlaceholderText('expenses.approve.rejectReason')[0], {
      target: { value: '却下理由' },
    });
    clickButtonByText('common.confirm', 0);
    await waitFor(() => {
      expect(apiMocks.expenses.advancedApprove).toHaveBeenCalledWith('e1', {
        action: 'reject',
        reason: '却下理由',
      });
    });

    fireEvent.click(screen.getByText('pagination-next'));
    fireEvent.click(screen.getByText('pagination-size'));
    expect(apiMocks.expenses.getPending).toHaveBeenCalledWith({ page: 2, page_size: 10 });
    expect(apiMocks.expenses.getPending).toHaveBeenCalledWith({ page: 1, page_size: 50 });
  });
});
