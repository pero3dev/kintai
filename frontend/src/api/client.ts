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
};
