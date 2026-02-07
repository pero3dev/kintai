import { describe, it, expect, beforeEach } from 'vitest';
import { useAuthStore, User } from './authStore';

describe('authStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
    });
  });

  const mockUser: User = {
    id: 'test-user-id',
    email: 'test@example.com',
    first_name: 'Test',
    last_name: 'User',
    role: 'employee',
    is_active: true,
  };

  describe('initial state', () => {
    it('should have null user initially', () => {
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
    });

    it('should have null tokens initially', () => {
      const state = useAuthStore.getState();
      expect(state.accessToken).toBeNull();
      expect(state.refreshToken).toBeNull();
    });

    it('should not be authenticated initially', () => {
      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe('setAuth', () => {
    it('should set user and tokens', () => {
      useAuthStore.getState().setAuth(mockUser, 'access-token', 'refresh-token');
      
      const state = useAuthStore.getState();
      expect(state.user).toEqual(mockUser);
      expect(state.accessToken).toBe('access-token');
      expect(state.refreshToken).toBe('refresh-token');
      expect(state.isAuthenticated).toBe(true);
    });

    it('should handle null user', () => {
      useAuthStore.getState().setAuth(null, 'access-token', 'refresh-token');
      
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.isAuthenticated).toBe(true);
    });

    it('should handle user with admin role', () => {
      const adminUser: User = { ...mockUser, role: 'admin' };
      useAuthStore.getState().setAuth(adminUser, 'token', 'refresh');
      
      expect(useAuthStore.getState().user?.role).toBe('admin');
    });

    it('should handle user with manager role', () => {
      const managerUser: User = { ...mockUser, role: 'manager' };
      useAuthStore.getState().setAuth(managerUser, 'token', 'refresh');
      
      expect(useAuthStore.getState().user?.role).toBe('manager');
    });

    it('should handle user with department_id', () => {
      const userWithDept: User = { ...mockUser, department_id: 'dept-123' };
      useAuthStore.getState().setAuth(userWithDept, 'token', 'refresh');
      
      expect(useAuthStore.getState().user?.department_id).toBe('dept-123');
    });
  });

  describe('logout', () => {
    it('should clear all auth state', () => {
      // First set auth
      useAuthStore.getState().setAuth(mockUser, 'access-token', 'refresh-token');
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      
      // Then logout
      useAuthStore.getState().logout();
      
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.refreshToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });

    it('should work even if not authenticated', () => {
      useAuthStore.getState().logout();
      
      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe('updateUser', () => {
    it('should update user first_name', () => {
      useAuthStore.getState().setAuth(mockUser, 'token', 'refresh');
      useAuthStore.getState().updateUser({ first_name: 'Updated' });
      
      expect(useAuthStore.getState().user?.first_name).toBe('Updated');
    });

    it('should update user last_name', () => {
      useAuthStore.getState().setAuth(mockUser, 'token', 'refresh');
      useAuthStore.getState().updateUser({ last_name: 'NewLast' });
      
      expect(useAuthStore.getState().user?.last_name).toBe('NewLast');
    });

    it('should update multiple fields at once', () => {
      useAuthStore.getState().setAuth(mockUser, 'token', 'refresh');
      useAuthStore.getState().updateUser({ 
        first_name: 'New', 
        last_name: 'Name',
        email: 'new@example.com'
      });
      
      const user = useAuthStore.getState().user;
      expect(user?.first_name).toBe('New');
      expect(user?.last_name).toBe('Name');
      expect(user?.email).toBe('new@example.com');
    });

    it('should not affect other user fields', () => {
      useAuthStore.getState().setAuth(mockUser, 'token', 'refresh');
      useAuthStore.getState().updateUser({ first_name: 'Updated' });
      
      const user = useAuthStore.getState().user;
      expect(user?.email).toBe(mockUser.email);
      expect(user?.role).toBe(mockUser.role);
    });

    it('should return null when no user is set', () => {
      useAuthStore.getState().updateUser({ first_name: 'New' });
      
      expect(useAuthStore.getState().user).toBeNull();
    });

    it('should update is_active status', () => {
      useAuthStore.getState().setAuth(mockUser, 'token', 'refresh');
      useAuthStore.getState().updateUser({ is_active: false });
      
      expect(useAuthStore.getState().user?.is_active).toBe(false);
    });

    it('should update role', () => {
      useAuthStore.getState().setAuth(mockUser, 'token', 'refresh');
      useAuthStore.getState().updateUser({ role: 'manager' });
      
      expect(useAuthStore.getState().user?.role).toBe('manager');
    });
  });
});
