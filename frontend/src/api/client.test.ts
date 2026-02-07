import { describe, it, expect, beforeEach, vi, Mock } from 'vitest';
import { api } from './client';
import { useAuthStore } from '@/stores/authStore';

// Mock fetch globally
global.fetch = vi.fn();

// Helper to setup fetch mock
function setupFetchMock(response: unknown, options: { ok?: boolean; status?: number } = {}) {
  const { ok = true, status = 200 } = options;
  (global.fetch as Mock).mockResolvedValueOnce({
    ok,
    status,
    json: () => Promise.resolve(response),
  });
}

describe('api client', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset auth store
    useAuthStore.setState({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
    });
  });

  describe('auth', () => {
    describe('login', () => {
      it('should call login endpoint with credentials', async () => {
        const mockResponse = { access_token: 'token', user: { id: '1' } };
        setupFetchMock(mockResponse);

        const result = await api.auth.login({ email: 'test@example.com', password: 'password' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/login'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ email: 'test@example.com', password: 'password' }),
          })
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('register', () => {
      it('should call register endpoint with user data', async () => {
        const mockResponse = { id: '1', email: 'test@example.com' };
        setupFetchMock(mockResponse);

        const userData = {
          email: 'test@example.com',
          password: 'password',
          first_name: 'Test',
          last_name: 'User',
        };
        const result = await api.auth.register(userData);

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/register'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(userData),
          })
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('refresh', () => {
      it('should call refresh endpoint with refresh token', async () => {
        const mockResponse = { access_token: 'new-token' };
        setupFetchMock(mockResponse);

        const result = await api.auth.refresh({ refresh_token: 'old-refresh' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/refresh'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ refresh_token: 'old-refresh' }),
          })
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('logout', () => {
      it('should call logout endpoint', async () => {
        setupFetchMock(null, { status: 204 });

        await api.auth.logout();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/auth/logout'),
          expect.objectContaining({ method: 'POST' })
        );
      });
    });
  });

  describe('attendance', () => {
    beforeEach(() => {
      useAuthStore.getState().setAuth(
        { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User', role: 'employee', is_active: true },
        'test-token',
        'refresh-token'
      );
    });

    describe('clockIn', () => {
      it('should call clock-in endpoint', async () => {
        const mockResponse = { id: '1', clock_in: '2024-01-15T09:00:00Z' };
        setupFetchMock(mockResponse);

        const result = await api.attendance.clockIn({ note: 'Morning shift' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/attendance/clock-in'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ note: 'Morning shift' }),
          })
        );
        expect(result).toEqual(mockResponse);
      });

      it('should handle clock-in without note', async () => {
        const mockResponse = { id: '1' };
        setupFetchMock(mockResponse);

        await api.attendance.clockIn();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/attendance/clock-in'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({}),
          })
        );
      });
    });

    describe('clockOut', () => {
      it('should call clock-out endpoint', async () => {
        const mockResponse = { id: '1', clock_out: '2024-01-15T18:00:00Z' };
        setupFetchMock(mockResponse);

        const result = await api.attendance.clockOut({ note: 'End of shift' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/attendance/clock-out'),
          expect.objectContaining({
            method: 'POST',
          })
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('getList', () => {
      it('should call attendance list endpoint with params', async () => {
        const mockResponse = { data: [], total: 0 };
        setupFetchMock(mockResponse);

        const result = await api.attendance.getList({ start_date: '2024-01-01', end_date: '2024-01-31' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/attendance?'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('getToday', () => {
      it('should call today endpoint', async () => {
        const mockResponse = { id: '1', status: 'present' };
        setupFetchMock(mockResponse);

        const result = await api.attendance.getToday();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/attendance/today'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('getSummary', () => {
      it('should call summary endpoint', async () => {
        const mockResponse = { total_days: 20 };
        setupFetchMock(mockResponse);

        const result = await api.attendance.getSummary({ start_date: '2024-01-01', end_date: '2024-01-31' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/attendance/summary?'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });
  });

  describe('leaves', () => {
    beforeEach(() => {
      useAuthStore.getState().setAuth(
        { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User', role: 'employee', is_active: true },
        'test-token',
        'refresh-token'
      );
    });

    describe('create', () => {
      it('should create a leave request', async () => {
        const mockResponse = { id: '1', status: 'pending' };
        setupFetchMock(mockResponse);

        const result = await api.leaves.create({
          leave_type: 'paid',
          start_date: '2024-02-01',
          end_date: '2024-02-05',
          reason: 'Vacation',
        });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/leaves'),
          expect.objectContaining({ method: 'POST' })
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('getList', () => {
      it('should get leave list', async () => {
        const mockResponse = { data: [], total: 0 };
        setupFetchMock(mockResponse);

        const result = await api.leaves.getList({ page: 1, page_size: 10 });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/leaves?'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('getPending', () => {
      it('should get pending leaves', async () => {
        const mockResponse = { data: [], total: 0 };
        setupFetchMock(mockResponse);

        await api.leaves.getPending();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/leaves/pending'),
          expect.anything()
        );
      });
    });

    describe('approve', () => {
      it('should approve a leave request', async () => {
        const mockResponse = { id: '1', status: 'approved' };
        setupFetchMock(mockResponse);

        const result = await api.leaves.approve('leave-id', { status: 'approved' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/leaves/leave-id/approve'),
          expect.objectContaining({ method: 'PUT' })
        );
        expect(result).toEqual(mockResponse);
      });
    });
  });

  describe('shifts', () => {
    beforeEach(() => {
      useAuthStore.getState().setAuth(
        { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User', role: 'manager', is_active: true },
        'test-token',
        'refresh-token'
      );
    });

    describe('getList', () => {
      it('should get shifts list', async () => {
        const mockResponse = { data: [] };
        setupFetchMock(mockResponse);

        await api.shifts.getList({ start_date: '2024-01-01', end_date: '2024-01-31' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/shifts?'),
          expect.anything()
        );
      });
    });

    describe('create', () => {
      it('should create a shift', async () => {
        const mockResponse = { id: '1' };
        setupFetchMock(mockResponse);

        await api.shifts.create({
          user_id: 'user-1',
          date: '2024-01-15',
          shift_type: 'morning',
        });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/shifts'),
          expect.objectContaining({ method: 'POST' })
        );
      });
    });

    describe('bulkCreate', () => {
      it('should create multiple shifts', async () => {
        const mockResponse = { created: 2 };
        setupFetchMock(mockResponse);

        await api.shifts.bulkCreate({
          shifts: [
            { user_id: 'user-1', date: '2024-01-15', shift_type: 'morning' },
            { user_id: 'user-1', date: '2024-01-16', shift_type: 'day' },
          ],
        });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/shifts/bulk'),
          expect.objectContaining({ method: 'POST' })
        );
      });
    });

    describe('delete', () => {
      it('should delete a shift', async () => {
        setupFetchMock(null, { status: 204 });

        await api.shifts.delete('shift-id');

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/shifts/shift-id'),
          expect.objectContaining({ method: 'DELETE' })
        );
      });
    });
  });

  describe('users', () => {
    beforeEach(() => {
      useAuthStore.getState().setAuth(
        { id: '1', email: 'admin@example.com', first_name: 'Admin', last_name: 'User', role: 'admin', is_active: true },
        'admin-token',
        'refresh-token'
      );
    });

    describe('getMe', () => {
      it('should get current user', async () => {
        const mockResponse = { id: '1', email: 'admin@example.com' };
        setupFetchMock(mockResponse);

        const result = await api.users.getMe();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/users/me'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });

    describe('getAll', () => {
      it('should get all users with pagination', async () => {
        const mockResponse = { data: [], total: 0 };
        setupFetchMock(mockResponse);

        await api.users.getAll({ page: 1, page_size: 10 });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/users?'),
          expect.anything()
        );
      });
    });

    describe('create', () => {
      it('should create a user', async () => {
        const mockResponse = { id: 'new-user' };
        setupFetchMock(mockResponse);

        await api.users.create({
          email: 'new@example.com',
          password: 'password',
          first_name: 'New',
          last_name: 'User',
          role: 'employee',
        });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/users'),
          expect.objectContaining({ method: 'POST' })
        );
      });
    });

    describe('update', () => {
      it('should update a user', async () => {
        const mockResponse = { id: 'user-1' };
        setupFetchMock(mockResponse);

        await api.users.update('user-1', { first_name: 'Updated' });

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/users/user-1'),
          expect.objectContaining({ method: 'PUT' })
        );
      });
    });

    describe('delete', () => {
      it('should delete a user', async () => {
        setupFetchMock(null, { status: 204 });

        await api.users.delete('user-1');

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/users/user-1'),
          expect.objectContaining({ method: 'DELETE' })
        );
      });
    });
  });

  describe('departments', () => {
    describe('getAll', () => {
      it('should get all departments', async () => {
        const mockResponse = [{ id: '1', name: 'Engineering' }];
        setupFetchMock(mockResponse);

        useAuthStore.getState().setAuth(
          { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User', role: 'employee', is_active: true },
          'test-token',
          'refresh-token'
        );

        const result = await api.departments.getAll();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/departments'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });
  });

  describe('dashboard', () => {
    describe('getStats', () => {
      it('should get dashboard stats', async () => {
        const mockResponse = { today_present: 10, today_absent: 2 };
        setupFetchMock(mockResponse);

        useAuthStore.getState().setAuth(
          { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User', role: 'manager', is_active: true },
          'test-token',
          'refresh-token'
        );

        const result = await api.dashboard.getStats();

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/dashboard/stats'),
          expect.anything()
        );
        expect(result).toEqual(mockResponse);
      });
    });
  });

  describe('error handling', () => {
    it('should handle API errors', async () => {
      (global.fetch as Mock).mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ message: 'Bad Request' }),
      });

      await expect(api.auth.login({ email: 'test@example.com', password: 'wrong' }))
        .rejects.toThrow('Bad Request');
    });

    it('should handle generic API error without message', async () => {
      (global.fetch as Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () => Promise.resolve({}),
      });

      await expect(api.auth.login({ email: 'test@example.com', password: 'test' }))
        .rejects.toThrow('APIエラーが発生しました');
    });

    it('should handle 204 No Content response', async () => {
      (global.fetch as Mock).mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: () => Promise.resolve(null),
      });

      const result = await api.auth.logout();
      expect(result).toBeNull();
    });
  });

  describe('authentication header', () => {
    it('should include auth token when authenticated', async () => {
      useAuthStore.getState().setAuth(
        { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User', role: 'employee', is_active: true },
        'my-access-token',
        'refresh-token'
      );

      setupFetchMock({ data: [] });

      await api.attendance.getList();

      expect(fetch).toHaveBeenCalledWith(
        expect.anything(),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer my-access-token',
          }),
        })
      );
    });

    it('should not include auth token when not authenticated', async () => {
      setupFetchMock({ access_token: 'token' });

      await api.auth.login({ email: 'test@example.com', password: 'password' });

      expect(fetch).toHaveBeenCalledWith(
        expect.anything(),
        expect.not.objectContaining({
          headers: expect.objectContaining({
            Authorization: expect.anything(),
          }),
        })
      );
    });
  });
});
