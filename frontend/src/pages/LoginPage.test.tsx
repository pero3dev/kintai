import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { LoginPage } from './LoginPage';

const state = vi.hoisted(() => ({
  navigate: vi.fn(),
  setAuth: vi.fn(),
}));

const apiMocks = vi.hoisted(() => ({
  auth: {
    login: vi.fn(),
  },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => state.navigate,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: {
      language: 'ja',
      changeLanguage: vi.fn(),
    },
  }),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    setAuth: state.setAuth,
  }),
}));

vi.mock('@/api/client', () => ({
  api: apiMocks,
}));

describe('LoginPage', () => {
  beforeEach(() => {
    state.navigate.mockReset();
    state.setAuth.mockReset();
    apiMocks.auth.login.mockReset();
  });

  it('submits login form and navigates to home on success', async () => {
    apiMocks.auth.login.mockResolvedValue({
      user: {
        id: 'u1',
        email: 'admin@example.com',
        first_name: 'Admin',
        last_name: 'User',
        role: 'admin',
        is_active: true,
      },
      access_token: 'access-token',
      refresh_token: 'refresh-token',
    });

    render(<LoginPage />);

    fireEvent.change(screen.getByPlaceholderText('user@example.com'), {
      target: { value: 'admin@example.com' },
    });
    fireEvent.change(
      document.querySelector('input[name="password"]') as HTMLInputElement,
      {
        target: { value: 'password123' },
      },
    );
    fireEvent.click(screen.getByRole('button'));

    await waitFor(() => {
      expect(apiMocks.auth.login).toHaveBeenCalledWith({
        email: 'admin@example.com',
        password: 'password123',
      });
      expect(state.setAuth).toHaveBeenCalledWith(
        expect.objectContaining({ email: 'admin@example.com' }),
        'access-token',
        'refresh-token',
      );
      expect(state.navigate).toHaveBeenCalledWith({ to: '/' });
    });
  });
});
