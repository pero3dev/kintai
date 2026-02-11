import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { Layout } from './Layout';

type MockUser = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: 'admin' | 'manager' | 'employee';
  is_active: boolean;
};

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  changeLanguage: vi.fn(),
  logout: vi.fn(),
  setTheme: vi.fn(),
  applyTheme: vi.fn(),
  getActiveApp: vi.fn(),
  pathname: '/',
  language: 'ja' as 'ja' | 'en',
  theme: 'system' as 'system' | 'light' | 'dark',
  user: null as MockUser | null,
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
  Outlet: () => <div data-testid="outlet">outlet</div>,
  useNavigate: () => mocks.navigate,
  useLocation: () => ({ pathname: mocks.pathname }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: {
      language: mocks.language,
      changeLanguage: mocks.changeLanguage,
    },
  }),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    user: mocks.user,
    logout: mocks.logout,
  }),
}));

vi.mock('@/stores/themeStore', () => ({
  useThemeStore: () => ({
    theme: mocks.theme,
    setTheme: mocks.setTheme,
  }),
  applyTheme: mocks.applyTheme,
}));

vi.mock('@/config/apps', () => ({
  getActiveApp: (pathname: string) => mocks.getActiveApp(pathname),
}));

vi.mock('./AppSwitcher', () => ({
  AppSwitcher: ({ collapsed = false }: { collapsed?: boolean }) => (
    <div data-testid={collapsed ? 'app-switcher-collapsed' : 'app-switcher-expanded'} />
  ),
}));

function setActiveApp(id?: 'attendance' | 'expenses' | 'hr' | 'wiki') {
  if (!id) {
    mocks.getActiveApp.mockReturnValue(undefined);
    return;
  }
  mocks.getActiveApp.mockReturnValue({ id });
}

function renderLayout() {
  return render(<Layout />);
}

function clickButtonByIcon(iconName: string) {
  const icon = screen.getAllByText(iconName).find((n) => n.closest('button'));
  if (!icon) throw new Error(`button icon not found: ${iconName}`);
  fireEvent.click(icon.closest('button') as HTMLButtonElement);
}

describe('Layout', () => {
  beforeEach(() => {
    localStorage.clear();
    document.body.style.overflow = '';
    mocks.pathname = '/';
    mocks.language = 'ja';
    mocks.theme = 'system';
    mocks.user = {
      id: 'u1',
      email: 'user@example.com',
      first_name: 'Taro',
      last_name: 'Yamada',
      role: 'employee',
      is_active: true,
    };
    setActiveApp(undefined);
    mocks.navigate.mockReset();
    mocks.changeLanguage.mockReset();
    mocks.logout.mockReset();
    mocks.setTheme.mockReset();
    mocks.applyTheme.mockReset();
    mocks.getActiveApp.mockClear();
  });

  it('renders default navigation for employee and marks root as active at "/"', () => {
    mocks.pathname = '/';
    setActiveApp(undefined);

    renderLayout();

    expect(screen.getAllByText('nav.home').length).toBeGreaterThan(0);
    expect(screen.queryByText('nav.dashboard')).not.toBeInTheDocument();
    expect(
      screen
        .getAllByRole('link')
        .some((l) => l.getAttribute('data-to') === '/' && l.className.includes('nav-item-active')),
    ).toBe(true);
    expect(screen.getAllByText('Yamada Taro').length).toBeGreaterThan(0);
  });

  it('includes admin-only items in default app for admin role and active prefix match', () => {
    mocks.pathname = '/attendance/history';
    mocks.user = {
      ...mocks.user!,
      role: 'admin',
    };
    setActiveApp(undefined);

    renderLayout();

    expect(screen.getAllByText('nav.dashboard').length).toBeGreaterThan(0);
    expect(
      screen
        .getAllByRole('link')
        .some(
          (l) =>
            l.getAttribute('data-to') === '/attendance' && l.className.includes('nav-item-active'),
        ),
    ).toBe(true);
  });

  it('renders expenses navigation and omits approval menus for employee', () => {
    mocks.pathname = '/expenses';
    setActiveApp('expenses');

    renderLayout();

    expect(screen.getAllByText('expenses.nav.dashboard').length).toBeGreaterThan(0);
    expect(screen.queryByText('expenses.nav.approve')).not.toBeInTheDocument();
    expect(screen.queryByText('expenses.nav.advancedApprove')).not.toBeInTheDocument();
  });

  it('renders expenses admin approval menus for admin role', () => {
    mocks.pathname = '/expenses';
    mocks.user = {
      ...mocks.user!,
      role: 'admin',
    };
    setActiveApp('expenses');

    renderLayout();

    expect(screen.getAllByText('expenses.nav.approve').length).toBeGreaterThan(0);
    expect(screen.getAllByText('expenses.nav.policy').length).toBeGreaterThan(0);
  });

  it('renders hr navigation items when HR app is active', () => {
    mocks.pathname = '/hr/evaluations';
    setActiveApp('hr');

    renderLayout();

    expect(screen.getAllByText('hr.nav.dashboard').length).toBeGreaterThan(0);
    expect(screen.getAllByText('hr.nav.employees').length).toBeGreaterThan(0);
    expect(screen.getAllByText('hr.nav.survey').length).toBeGreaterThan(0);
  });

  it('renders wiki navigation items in Japanese when Wiki app is active', () => {
    mocks.pathname = '/wiki/backend';
    mocks.language = 'ja';
    setActiveApp('wiki');

    renderLayout();

    expect(screen.getAllByText('概要').length).toBeGreaterThan(0);
    expect(screen.getAllByText('アーキテクチャ').length).toBeGreaterThan(0);
    expect(screen.getAllByText('バックエンド').length).toBeGreaterThan(0);
    expect(
      screen
        .getAllByRole('link')
        .some((l) => l.getAttribute('data-to') === '/wiki/backend' && l.className.includes('nav-item-active')),
    ).toBe(true);
  });

  it('renders wiki navigation items in English when locale is en', () => {
    mocks.pathname = '/wiki/backend';
    mocks.language = 'en';
    setActiveApp('wiki');

    renderLayout();

    expect(screen.getAllByText('Overview').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Architecture').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Backend').length).toBeGreaterThan(0);
  });

  it('uses sidebarCollapsed value from localStorage and toggles it', () => {
    localStorage.setItem('sidebarCollapsed', 'true');

    const { container } = renderLayout();
    const desktopAside = container.querySelectorAll('aside')[1];
    expect(desktopAside?.className.includes('w-16')).toBe(true);

    clickButtonByIcon('chevron_right');
    expect(localStorage.getItem('sidebarCollapsed')).toBe('false');
  });

  it('sets collapsed language button title for english locale', () => {
    localStorage.setItem('sidebarCollapsed', 'true');
    mocks.language = 'en';

    renderLayout();

    const languageButton = screen
      .getAllByText('language')
      .map((node) => node.closest('button'))
      .find((button) => button?.getAttribute('title') !== null) as HTMLButtonElement;

    expect(languageButton).toBeInTheDocument();
    expect(languageButton.getAttribute('title')).not.toBe('English');
  });

  it('toggles language ja -> en and en -> ja', () => {
    mocks.language = 'ja';
    const { unmount } = renderLayout();
    clickButtonByIcon('language');
    expect(mocks.changeLanguage).toHaveBeenCalledWith('en');

    unmount();
    mocks.changeLanguage.mockReset();
    mocks.language = 'en';
    renderLayout();
    clickButtonByIcon('language');
    expect(mocks.changeLanguage).toHaveBeenCalledWith('ja');
  });

  it('cycles theme from system to light and applies it', () => {
    mocks.theme = 'system';

    renderLayout();
    clickButtonByIcon('brightness_auto');

    expect(mocks.setTheme).toHaveBeenCalledWith('light');
    expect(mocks.applyTheme).toHaveBeenCalledWith('light');
  });

  it('cycles theme from light to dark and from dark to system', () => {
    mocks.theme = 'light';
    const { unmount } = renderLayout();
    clickButtonByIcon('light_mode');
    expect(mocks.setTheme).toHaveBeenCalledWith('dark');
    expect(mocks.applyTheme).toHaveBeenCalledWith('dark');

    unmount();
    mocks.setTheme.mockReset();
    mocks.applyTheme.mockReset();
    mocks.theme = 'dark';
    renderLayout();
    clickButtonByIcon('dark_mode');
    expect(mocks.setTheme).toHaveBeenCalledWith('system');
    expect(mocks.applyTheme).toHaveBeenCalledWith('system');
  });

  it('logs out and navigates to login', () => {
    renderLayout();
    clickButtonByIcon('logout');

    expect(mocks.logout).toHaveBeenCalledTimes(1);
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/login' });
  });

  it('opens mobile drawer, closes by overlay click, and resets body overflow on unmount', () => {
    const { container, unmount } = renderLayout();

    clickButtonByIcon('menu');
    expect(document.body.style.overflow).toBe('hidden');

    const overlay = container.querySelector('div.bg-black\\/50');
    expect(overlay).toBeInTheDocument();
    fireEvent.click(overlay as HTMLDivElement);
    expect(document.body.style.overflow).toBe('');

    clickButtonByIcon('menu');
    expect(document.body.style.overflow).toBe('hidden');
    unmount();
    expect(document.body.style.overflow).toBe('');
  });

  it('closes mobile drawer when location path changes', () => {
    const { container, rerender } = renderLayout();

    clickButtonByIcon('menu');
    expect(container.querySelector('div.bg-black\\/50')).toBeInTheDocument();

    mocks.pathname = '/attendance';
    rerender(<Layout />);

    expect(container.querySelector('div.bg-black\\/50')).not.toBeInTheDocument();
  });

  it('updates desktop search input value', () => {
    renderLayout();
    const input = screen.getByPlaceholderText('common.search') as HTMLInputElement;

    fireEvent.change(input, { target: { value: 'keyword' } });
    expect(input.value).toBe('keyword');
  });

  it('renders without user profile blocks when user is null', () => {
    mocks.user = null;

    renderLayout();

    expect(screen.queryByText('Yamada Taro')).not.toBeInTheDocument();
  });
});
