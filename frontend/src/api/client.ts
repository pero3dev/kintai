import createClient from 'openapi-fetch';
import { useAuthStore } from '@/stores/authStore';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

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
  async onResponse({ response }) {
    if (response.status === 401) {
      // TODO: リフレッシュトークンでリトライ
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
};
