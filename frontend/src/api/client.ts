import createClient from 'openapi-fetch';
import { useAuthStore } from '@/stores/authStore';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

// トークンリフレッシュの排他制御
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

async function tryRefreshToken(): Promise<boolean> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }
  isRefreshing = true;
  refreshPromise = (async () => {
    let refreshed = false;
    try {
      const refreshToken = useAuthStore.getState().refreshToken;
      if (refreshToken) {
        const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (response.ok) {
          const data = await response.json();
          if (data.access_token) {
            useAuthStore.getState().setAuth(
              data.user || useAuthStore.getState().user,
              data.access_token,
              data.refresh_token || refreshToken,
            );
            refreshed = true;
          }
        }
      }
    } catch {
      refreshed = false;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
    return refreshed;
  })();
  return refreshPromise;
}

// OpenAPI クライアント（型安全なAPIクライアント）
export const apiClient = createClient({
  baseUrl: API_BASE_URL,
});

// リクエストインターセプター: 認証トークンを自動付与
apiClient.use({
  async onRequest({ request }) {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`);
    }
    return request;
  },
  async onResponse({ request, response }) {
    if (response.status === 401) {
      const refreshed = await tryRefreshToken();
      if (refreshed) {
        // リフレッシュ成功 → リトライ
        const token = useAuthStore.getState().accessToken;
        const retryResponse = await fetch(request.url, {
          ...request,
          headers: {
            ...Object.fromEntries(request.headers.entries()),
            Authorization: `Bearer ${token}`,
          },
        });
        return retryResponse;
      }
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return response;
  },
});

// 汎用APIヘルパー (OpenAPIスキーマ生成前の仮実装)
const BASE_URL = API_BASE_URL;

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = useAuthStore.getState().accessToken;
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    // リフレッシュトークンでリトライ
    const refreshed = await tryRefreshToken();
    if (refreshed) {
      const newToken = useAuthStore.getState().accessToken;
      const retryResponse = await fetch(`${BASE_URL}${url}`, {
        ...options,
        headers: {
          ...headers,
          Authorization: `Bearer ${newToken}`,
        },
      });
      if (!retryResponse.ok) {
        if (retryResponse.status === 401) {
          useAuthStore.getState().logout();
          window.location.href = '/login';
          throw new Error('Unauthorized');
        }
        const error = await retryResponse.json();
        throw new Error(error.message || 'APIエラーが発生しました');
      }
      if (retryResponse.status === 204) return null;
      return retryResponse.json();
    }
    useAuthStore.getState().logout();
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'APIエラーが発生しました');
  }

  if (response.status === 204) return null;
  return response.json();
}

async function fetchWithAuthBlob(url: string, params: Record<string, string>) {
  const token = useAuthStore.getState().accessToken;
  const query = new URLSearchParams(params).toString();
  const response = await fetch(`${BASE_URL}${url}?${query}`, {
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });
  if (!response.ok) {
    throw new Error('エクスポートに失敗しました');
  }
  return response.blob();
}

export const api = {
  // 認証
  auth: {
    login: (data: { email: string; password: string }) =>
      fetchWithAuth('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
    register: (data: { email: string; password: string; first_name: string; last_name: string }) =>
      fetchWithAuth('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
    refresh: (data: { refresh_token: string }) =>
      fetchWithAuth('/auth/refresh', { method: 'POST', body: JSON.stringify(data) }),
    logout: () => fetchWithAuth('/auth/logout', { method: 'POST' }),
  },

  // 勤怠
  attendance: {
    clockIn: (data?: { note?: string }) =>
      fetchWithAuth('/attendance/clock-in', { method: 'POST', body: JSON.stringify(data || {}) }),
    clockOut: (data?: { note?: string }) =>
      fetchWithAuth('/attendance/clock-out', { method: 'POST', body: JSON.stringify(data || {}) }),
    getList: (params?: { start_date?: string; end_date?: string; page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/attendance?${query}`);
    },
    getToday: () => fetchWithAuth('/attendance/today'),
    getSummary: (params?: { start_date?: string; end_date?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/attendance/summary?${query}`);
    },
  },

  // 休暇
  leaves: {
    create: (data: { leave_type: string; start_date: string; end_date: string; reason?: string }) =>
      fetchWithAuth('/leaves', { method: 'POST', body: JSON.stringify(data) }),
    getList: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/leaves?${query}`);
    },
    getPending: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/leaves/pending?${query}`);
    },
    approve: (id: string, data: { status: string; rejected_reason?: string }) =>
      fetchWithAuth(`/leaves/${id}/approve`, { method: 'PUT', body: JSON.stringify(data) }),
  },

  // シフト
  shifts: {
    getList: (params: { start_date: string; end_date: string }) => {
      const query = new URLSearchParams(params).toString();
      return fetchWithAuth(`/shifts?${query}`);
    },
    create: (data: { user_id: string; date: string; shift_type: string }) =>
      fetchWithAuth('/shifts', { method: 'POST', body: JSON.stringify(data) }),
    bulkCreate: (data: { shifts: Array<{ user_id: string; date: string; shift_type: string }> }) =>
      fetchWithAuth('/shifts/bulk', { method: 'POST', body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth(`/shifts/${id}`, { method: 'DELETE' }),
  },

  // ユーザー
  users: {
    getMe: () => fetchWithAuth('/users/me'),
    getAll: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/users?${query}`);
    },
    create: (data: { email: string; password: string; first_name: string; last_name: string; role: string; department_id?: string }) =>
      fetchWithAuth('/users', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) =>
      fetchWithAuth(`/users/${id}`, { method: 'DELETE' }),
  },

  // 部署
  departments: {
    getAll: () => fetchWithAuth('/departments'),
  },

  // ダッシュボード
  dashboard: {
    getStats: () => fetchWithAuth('/dashboard/stats'),
  },

  // 残業申請
  overtime: {
    create: (data: { date: string; planned_minutes: number; reason: string }) =>
      fetchWithAuth('/overtime', { method: 'POST', body: JSON.stringify(data) }),
    getList: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/overtime?${query}`);
    },
    getPending: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/overtime/pending?${query}`);
    },
    approve: (id: string, data: { status: string; rejected_reason?: string }) =>
      fetchWithAuth(`/overtime/${id}/approve`, { method: 'PUT', body: JSON.stringify(data) }),
    getAlerts: () => fetchWithAuth('/overtime/alerts'),
  },

  // 勤怠修正申請
  corrections: {
    create: (data: { date: string; corrected_clock_in?: string; corrected_clock_out?: string; reason: string }) =>
      fetchWithAuth('/corrections', { method: 'POST', body: JSON.stringify(data) }),
    getList: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/corrections?${query}`);
    },
    getPending: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/corrections/pending?${query}`);
    },
    approve: (id: string, data: { status: string; rejected_reason?: string }) =>
      fetchWithAuth(`/corrections/${id}/approve`, { method: 'PUT', body: JSON.stringify(data) }),
  },

  // 通知
  notifications: {
    getList: (params?: { page?: number; page_size?: number; is_read?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/notifications?${query}`);
    },
    getUnreadCount: () => fetchWithAuth('/notifications/unread-count'),
    markAsRead: (id: string) => fetchWithAuth(`/notifications/${id}/read`, { method: 'PUT' }),
    markAllAsRead: () => fetchWithAuth('/notifications/read-all', { method: 'PUT' }),
    delete: (id: string) => fetchWithAuth(`/notifications/${id}`, { method: 'DELETE' }),
  },

  // 有給残日数
  leaveBalances: {
    getMy: (params?: { fiscal_year?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/leave-balances?${query}`);
    },
    getByUser: (userId: string, params?: { fiscal_year?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/leave-balances/${userId}?${query}`);
    },
    setBalance: (userId: string, leaveType: string, data: { total_days?: number; carried_over?: number }) =>
      fetchWithAuth(`/leave-balances/${userId}/${leaveType}`, { method: 'PUT', body: JSON.stringify(data) }),
    initialize: (userId: string, params?: { fiscal_year?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/leave-balances/${userId}/initialize?${query}`, { method: 'POST' });
    },
  },

  // プロジェクト
  projects: {
    getAll: (params?: { page?: number; page_size?: number; status?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/projects?${query}`);
    },
    getByID: (id: string) => fetchWithAuth(`/projects/${id}`),
    create: (data: { name: string; code: string; description?: string; manager_id?: string; budget_hours?: number }) =>
      fetchWithAuth('/projects', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/projects/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth(`/projects/${id}`, { method: 'DELETE' }),
  },

  // 工数管理
  timeEntries: {
    create: (data: { project_id: string; date: string; minutes: number; description?: string }) =>
      fetchWithAuth('/time-entries', { method: 'POST', body: JSON.stringify(data) }),
    getList: (params?: { start_date?: string; end_date?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/time-entries?${query}`);
    },
    getByProject: (projectId: string, params?: { start_date?: string; end_date?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/projects/${projectId}/time-entries?${query}`);
    },
    update: (id: string, data: { minutes?: number; description?: string }) =>
      fetchWithAuth(`/time-entries/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth(`/time-entries/${id}`, { method: 'DELETE' }),
    getSummary: (params?: { start_date?: string; end_date?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/time-entries/summary?${query}`);
    },
  },

  // 祝日・カレンダー
  holidays: {
    getByYear: (params?: { year?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/holidays?${query}`);
    },
    getCalendar: (params?: { year?: string; month?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/holidays/calendar?${query}`);
    },
    getWorkingDays: (params?: { start_date?: string; end_date?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/holidays/working-days?${query}`);
    },
    create: (data: { date: string; name: string; holiday_type: string; is_recurring?: boolean }) =>
      fetchWithAuth('/holidays', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/holidays/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth(`/holidays/${id}`, { method: 'DELETE' }),
  },

  // 承認フロー
  approvalFlows: {
    getAll: () => fetchWithAuth('/approval-flows'),
    getByID: (id: string) => fetchWithAuth(`/approval-flows/${id}`),
    create: (data: { name: string; flow_type: string; steps: Array<{ step_order: number; step_type: string; approver_role?: string; approver_id?: string }> }) =>
      fetchWithAuth('/approval-flows', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/approval-flows/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth(`/approval-flows/${id}`, { method: 'DELETE' }),
  },

  // エクスポート
  export: {
    attendance: (params: { start_date: string; end_date: string; user_id?: string }) =>
      fetchWithAuthBlob('/export/attendance', params),
    leaves: (params: { start_date: string; end_date: string; user_id?: string }) =>
      fetchWithAuthBlob('/export/leaves', params),
    overtime: (params: { start_date: string; end_date: string }) =>
      fetchWithAuthBlob('/export/overtime', params),
    projects: (params: { start_date: string; end_date: string }) =>
      fetchWithAuthBlob('/export/projects', params),
  },

  // 経費精算
  expenses: {
    getList: (params?: { page?: number; page_size?: number; status?: string; category?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/expenses?${query}`);
    },
    getByID: (id: string) => fetchWithAuth(`/expenses/${id}`),
    create: (data: { title: string; status: string; notes?: string; items: Array<{ expense_date: string; category: string; description: string; amount: number; receipt_url?: string }> }) =>
      fetchWithAuth('/expenses', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/expenses/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) => fetchWithAuth(`/expenses/${id}`, { method: 'DELETE' }),
    getPending: (params?: { page?: number; page_size?: number }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/expenses/pending?${query}`);
    },
    approve: (id: string, data: { status: string; rejected_reason?: string }) =>
      fetchWithAuth(`/expenses/${id}/approve`, { method: 'PUT', body: JSON.stringify(data) }),
    getStats: () => fetchWithAuth('/expenses/stats'),

    // コメント
    getComments: (id: string) => fetchWithAuth(`/expenses/${id}/comments`),
    addComment: (id: string, data: { content: string }) =>
      fetchWithAuth(`/expenses/${id}/comments`, { method: 'POST', body: JSON.stringify(data) }),

    // 変更履歴
    getHistory: (id: string) => fetchWithAuth(`/expenses/${id}/history`),

    // レポート
    getReport: (params: { start_date: string; end_date: string }) => {
      const query = new URLSearchParams(params).toString();
      return fetchWithAuth(`/expenses/report?${query}`);
    },
    getMonthlyTrend: (params: { year: string }) => {
      const query = new URLSearchParams(params).toString();
      return fetchWithAuth(`/expenses/report/monthly?${query}`);
    },
    exportCSV: (params: { start_date: string; end_date: string }) =>
      fetchWithAuthBlob('/expenses/export/csv', params),
    exportPDF: (params: { start_date: string; end_date: string }) =>
      fetchWithAuthBlob('/expenses/export/pdf', params),

    // レシートアップロード
    uploadReceipt: async (file: File) => {
      const token = useAuthStore.getState().accessToken;
      const formData = new FormData();
      formData.append('file', file);
      const response = await fetch(`${BASE_URL}/expenses/receipts/upload`, {
        method: 'POST',
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: formData,
      });
      if (!response.ok) throw new Error('Upload failed');
      return response.json();
    },

    // テンプレート
    getTemplates: () => fetchWithAuth('/expenses/templates'),
    createTemplate: (data: Record<string, unknown>) =>
      fetchWithAuth('/expenses/templates', { method: 'POST', body: JSON.stringify(data) }),
    updateTemplate: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/expenses/templates/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteTemplate: (id: string) =>
      fetchWithAuth(`/expenses/templates/${id}`, { method: 'DELETE' }),
    useTemplate: (id: string) =>
      fetchWithAuth(`/expenses/templates/${id}/use`, { method: 'POST' }),

    // ポリシー
    getPolicies: () => fetchWithAuth('/expenses/policies'),
    createPolicy: (data: Record<string, unknown>) =>
      fetchWithAuth('/expenses/policies', { method: 'POST', body: JSON.stringify(data) }),
    updatePolicy: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/expenses/policies/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deletePolicy: (id: string) =>
      fetchWithAuth(`/expenses/policies/${id}`, { method: 'DELETE' }),
    getBudgets: () => fetchWithAuth('/expenses/budgets'),
    getPolicyViolations: () => fetchWithAuth('/expenses/policy-violations'),

    // 通知
    getNotifications: (params?: { filter?: string }) => {
      const query = new URLSearchParams(params as Record<string, string>).toString();
      return fetchWithAuth(`/expenses/notifications?${query}`);
    },
    markNotificationRead: (id: string) =>
      fetchWithAuth(`/expenses/notifications/${id}/read`, { method: 'PUT' }),
    markAllNotificationsRead: () =>
      fetchWithAuth('/expenses/notifications/read-all', { method: 'PUT' }),
    getReminders: () => fetchWithAuth('/expenses/reminders'),
    dismissReminder: (id: string) =>
      fetchWithAuth(`/expenses/reminders/${id}/dismiss`, { method: 'PUT' }),
    getNotificationSettings: () => fetchWithAuth('/expenses/notification-settings'),
    updateNotificationSettings: (data: Record<string, unknown>) =>
      fetchWithAuth('/expenses/notification-settings', { method: 'PUT', body: JSON.stringify(data) }),

    // 高度な承認ワークフロー
    advancedApprove: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/expenses/${id}/advanced-approve`, { method: 'PUT', body: JSON.stringify(data) }),
    getApprovalFlowConfig: () => fetchWithAuth('/expenses/approval-flow'),
    getDelegates: () => fetchWithAuth('/expenses/delegates'),
    setDelegate: (data: Record<string, unknown>) =>
      fetchWithAuth('/expenses/delegates', { method: 'POST', body: JSON.stringify(data) }),
    removeDelegate: (id: string) =>
      fetchWithAuth(`/expenses/delegates/${id}`, { method: 'DELETE' }),
  },

  // 人事管理
  hr: {
    // ダッシュボード
    getStats: () => fetchWithAuth('/hr/stats'),
    getRecentActivities: () => fetchWithAuth('/hr/activities'),

    // 社員
    getEmployees: (params?: { page?: number; page_size?: number; department?: string; status?: string; employment_type?: string; search?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/employees?${query.toString()}`);
    },
    getEmployee: (id: string) => fetchWithAuth(`/hr/employees/${id}`),
    createEmployee: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/employees', { method: 'POST', body: JSON.stringify(data) }),
    updateEmployee: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteEmployee: (id: string) =>
      fetchWithAuth(`/hr/employees/${id}`, { method: 'DELETE' }),

    // 部門
    getDepartments: () => fetchWithAuth('/hr/departments'),
    getDepartment: (id: string) => fetchWithAuth(`/hr/departments/${id}`),
    createDepartment: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/departments', { method: 'POST', body: JSON.stringify(data) }),
    updateDepartment: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/departments/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteDepartment: (id: string) =>
      fetchWithAuth(`/hr/departments/${id}`, { method: 'DELETE' }),

    // 評価
    getEvaluations: (params?: { page?: number; page_size?: number; cycle_id?: string; status?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/evaluations?${query.toString()}`);
    },
    getEvaluation: (id: string) => fetchWithAuth(`/hr/evaluations/${id}`),
    createEvaluation: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/evaluations', { method: 'POST', body: JSON.stringify(data) }),
    updateEvaluation: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/evaluations/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    submitEvaluation: (id: string) =>
      fetchWithAuth(`/hr/evaluations/${id}/submit`, { method: 'PUT' }),
    getEvaluationCycles: () => fetchWithAuth('/hr/evaluation-cycles'),
    createEvaluationCycle: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/evaluation-cycles', { method: 'POST', body: JSON.stringify(data) }),

    // 目標
    getGoals: (params?: { page?: number; page_size?: number; status?: string; category?: string; employee_id?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/goals?${query.toString()}`);
    },
    getGoal: (id: string) => fetchWithAuth(`/hr/goals/${id}`),
    createGoal: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/goals', { method: 'POST', body: JSON.stringify(data) }),
    updateGoal: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/goals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteGoal: (id: string) =>
      fetchWithAuth(`/hr/goals/${id}`, { method: 'DELETE' }),
    updateGoalProgress: (id: string, progress: number) =>
      fetchWithAuth(`/hr/goals/${id}/progress`, { method: 'PUT', body: JSON.stringify({ progress }) }),

    // 研修
    getTrainingPrograms: (params?: { page?: number; page_size?: number; category?: string; status?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/training?${query.toString()}`);
    },
    getTrainingProgram: (id: string) => fetchWithAuth(`/hr/training/${id}`),
    createTrainingProgram: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/training', { method: 'POST', body: JSON.stringify(data) }),
    updateTrainingProgram: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/training/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteTrainingProgram: (id: string) =>
      fetchWithAuth(`/hr/training/${id}`, { method: 'DELETE' }),
    enrollTraining: (id: string) =>
      fetchWithAuth(`/hr/training/${id}/enroll`, { method: 'POST' }),
    completeTraining: (id: string, data?: Record<string, unknown>) =>
      fetchWithAuth(`/hr/training/${id}/complete`, { method: 'PUT', body: JSON.stringify(data || {}) }),

    // 採用
    getPositions: (params?: { page?: number; page_size?: number; status?: string; department?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/positions?${query.toString()}`);
    },
    getPosition: (id: string) => fetchWithAuth(`/hr/positions/${id}`),
    createPosition: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/positions', { method: 'POST', body: JSON.stringify(data) }),
    updatePosition: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/positions/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    getApplicants: (params?: { position_id?: string; stage?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/applicants?${query.toString()}`);
    },
    createApplicant: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/applicants', { method: 'POST', body: JSON.stringify(data) }),
    updateApplicantStage: (id: string, stage: string) =>
      fetchWithAuth(`/hr/applicants/${id}/stage`, { method: 'PUT', body: JSON.stringify({ stage }) }),

    // 書類
    getDocuments: (params?: { page?: number; page_size?: number; type?: string; employee_id?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && v !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/documents?${query.toString()}`);
    },
    uploadDocument: async (formData: FormData) => {
      const token = useAuthStore.getState().accessToken;
      const res = await fetch(`${BASE_URL}/hr/documents`, {
        method: 'POST',
        headers: { ...(token ? { Authorization: `Bearer ${token}` } : {}) },
        body: formData,
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    },
    deleteDocument: (id: string) =>
      fetchWithAuth(`/hr/documents/${id}`, { method: 'DELETE' }),
    downloadDocument: (id: string) =>
      fetchWithAuthBlob(`/hr/documents/${id}/download`, {}),

    // お知らせ
    getAnnouncements: (params?: { page?: number; page_size?: number; priority?: string }) => {
      const query = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([k, v]) => { if (v !== undefined && String(v) !== '') query.set(k, String(v)); });
      }
      return fetchWithAuth(`/hr/announcements?${query.toString()}`);
    },
    getAnnouncement: (id: string) => fetchWithAuth(`/hr/announcements/${id}`),
    createAnnouncement: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/announcements', { method: 'POST', body: JSON.stringify(data) }),
    updateAnnouncement: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/announcements/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteAnnouncement: (id: string) =>
      fetchWithAuth(`/hr/announcements/${id}`, { method: 'DELETE' }),

    // 勤怠連携
    getAttendanceIntegration: (params?: { period?: string; department?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/attendance-integration?${query.toString()}`);
    },
    getAttendanceAlerts: () => fetchWithAuth('/hr/attendance-integration/alerts'),
    getAttendanceTrend: (params?: { period?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/attendance-integration/trend?${query.toString()}`);
    },

    // 組織図
    getOrgChart: () => fetchWithAuth('/hr/org-chart'),
    simulateOrgChange: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/org-chart/simulate', { method: 'POST', body: JSON.stringify(data) }),

    // 1on1
    getOneOnOnes: (params?: { page?: number; status?: string; employee_id?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v !== undefined && String(v) !== '') query.set(k, String(v)); });
      return fetchWithAuth(`/hr/one-on-ones?${query.toString()}`);
    },
    getOneOnOne: (id: string) => fetchWithAuth(`/hr/one-on-ones/${id}`),
    createOneOnOne: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/one-on-ones', { method: 'POST', body: JSON.stringify(data) }),
    updateOneOnOne: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/one-on-ones/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteOneOnOne: (id: string) =>
      fetchWithAuth(`/hr/one-on-ones/${id}`, { method: 'DELETE' }),
    addActionItem: (meetingId: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/one-on-ones/${meetingId}/actions`, { method: 'POST', body: JSON.stringify(data) }),
    toggleActionItem: (meetingId: string, actionId: string) =>
      fetchWithAuth(`/hr/one-on-ones/${meetingId}/actions/${actionId}/toggle`, { method: 'PUT' }),

    // スキルマップ
    getSkillMap: (params?: { department?: string; employee_id?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/skill-map?${query.toString()}`);
    },
    getSkillGapAnalysis: (params?: { department?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/skill-map/gap-analysis?${query.toString()}`);
    },
    addEmployeeSkill: (employeeId: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/skill-map/${employeeId}`, { method: 'POST', body: JSON.stringify(data) }),
    updateEmployeeSkill: (employeeId: string, skillId: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/skill-map/${employeeId}/${skillId}`, { method: 'PUT', body: JSON.stringify(data) }),

    // 給与シミュレーター
    getSalaryOverview: (params?: { department?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/salary?${query.toString()}`);
    },
    simulateSalary: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/salary/simulate', { method: 'POST', body: JSON.stringify(data) }),
    getSalaryHistory: (employeeId: string) =>
      fetchWithAuth(`/hr/salary/${employeeId}/history`),
    getBudgetOverview: (params?: { department?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/salary/budget?${query.toString()}`);
    },

    // オンボーディング
    getOnboardings: (params?: { status?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/onboarding?${query.toString()}`);
    },
    getOnboarding: (id: string) => fetchWithAuth(`/hr/onboarding/${id}`),
    createOnboarding: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/onboarding', { method: 'POST', body: JSON.stringify(data) }),
    updateOnboarding: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/onboarding/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    toggleOnboardingTask: (onboardingId: string, taskId: string) =>
      fetchWithAuth(`/hr/onboarding/${onboardingId}/tasks/${taskId}/toggle`, { method: 'PUT' }),
    getOnboardingTemplates: () => fetchWithAuth('/hr/onboarding/templates'),
    createOnboardingTemplate: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/onboarding/templates', { method: 'POST', body: JSON.stringify(data) }),

    // 退職管理
    getOffboardings: (params?: { status?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/offboarding?${query.toString()}`);
    },
    getOffboarding: (id: string) => fetchWithAuth(`/hr/offboarding/${id}`),
    createOffboarding: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/offboarding', { method: 'POST', body: JSON.stringify(data) }),
    updateOffboarding: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/offboarding/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    toggleOffboardingChecklist: (offboardingId: string, itemKey: string) =>
      fetchWithAuth(`/hr/offboarding/${offboardingId}/checklist/${itemKey}/toggle`, { method: 'PUT' }),
    getTurnoverAnalytics: (params?: { period?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/offboarding/analytics?${query.toString()}`);
    },

    // サーベイ
    getSurveys: (params?: { status?: string; type?: string }) => {
      const query = new URLSearchParams();
      if (params) Object.entries(params).forEach(([k, v]) => { if (v) query.set(k, String(v)); });
      return fetchWithAuth(`/hr/surveys?${query.toString()}`);
    },
    getSurvey: (id: string) => fetchWithAuth(`/hr/surveys/${id}`),
    createSurvey: (data: Record<string, unknown>) =>
      fetchWithAuth('/hr/surveys', { method: 'POST', body: JSON.stringify(data) }),
    updateSurvey: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/surveys/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    deleteSurvey: (id: string) =>
      fetchWithAuth(`/hr/surveys/${id}`, { method: 'DELETE' }),
    publishSurvey: (id: string) =>
      fetchWithAuth(`/hr/surveys/${id}/publish`, { method: 'PUT' }),
    closeSurvey: (id: string) =>
      fetchWithAuth(`/hr/surveys/${id}/close`, { method: 'PUT' }),
    getSurveyResults: (id: string) => fetchWithAuth(`/hr/surveys/${id}/results`),
    submitSurveyResponse: (id: string, data: Record<string, unknown>) =>
      fetchWithAuth(`/hr/surveys/${id}/respond`, { method: 'POST', body: JSON.stringify(data) }),
  },
};
