import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { AppSwitcher } from './AppSwitcher';

type MockApp = {
  id: string;
  nameKey: string;
  descriptionKey: string;
  icon: string;
  color: string;
  basePath: string;
  enabled: boolean;
  comingSoon?: boolean;
};

const mockNavigate = vi.fn();
const mockGetActiveApp = vi.fn();
const mockGetAvailableApps = vi.fn();

let mockPathname = '/';
let mockUserRole: 'admin' | 'manager' | 'employee' = 'employee';
let mockActiveApp: MockApp | undefined;
let mockAvailableApps: MockApp[] = [];

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
  useLocation: () => ({ pathname: mockPathname }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    user: {
      id: 'u1',
      email: 'test@example.com',
      first_name: 'Taro',
      last_name: 'Yamada',
      role: mockUserRole,
      is_active: true,
    },
  }),
}));

vi.mock('@/config/apps', () => ({
  getActiveApp: (pathname: string) => mockGetActiveApp(pathname),
  getAvailableApps: (role?: string) => mockGetAvailableApps(role),
}));

function renderAppSwitcher(collapsed = false) {
  return render(<AppSwitcher collapsed={collapsed} />);
}

function openDropdown() {
  const trigger = screen.getByText('apps').closest('button') as HTMLButtonElement;
  fireEvent.click(trigger);
}

function setWindowSize(width: number, height: number) {
  Object.defineProperty(window, 'innerWidth', {
    value: width,
    writable: true,
    configurable: true,
  });
  Object.defineProperty(window, 'innerHeight', {
    value: height,
    writable: true,
    configurable: true,
  });
}

describe('AppSwitcher', () => {
  beforeEach(() => {
    mockNavigate.mockReset();
    mockGetActiveApp.mockReset();
    mockGetAvailableApps.mockReset();
    mockPathname = '/';
    mockUserRole = 'employee';
    mockActiveApp = {
      id: 'attendance',
      nameKey: 'apps.attendance.name',
      descriptionKey: 'apps.attendance.description',
      icon: 'schedule',
      color: 'bg-blue-500',
      basePath: '/',
      enabled: true,
    };
    mockAvailableApps = [
      mockActiveApp,
      {
        id: 'expenses',
        nameKey: 'apps.expenses.name',
        descriptionKey: 'apps.expenses.description',
        icon: 'receipt_long',
        color: 'bg-green-500',
        basePath: '/expenses',
        enabled: true,
      },
      {
        id: 'wiki',
        nameKey: 'apps.wiki.name',
        descriptionKey: 'apps.wiki.description',
        icon: 'menu_book',
        color: 'bg-amber-500',
        basePath: '/wiki',
        enabled: false,
        comingSoon: true,
      },
    ];
    mockGetActiveApp.mockImplementation(() => mockActiveApp);
    mockGetAvailableApps.mockImplementation(() => mockAvailableApps);
    setWindowSize(1280, 720);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders expanded trigger with active app label and default title undefined', () => {
    renderAppSwitcher(false);

    expect(screen.getByText('appSwitcher.currentApp')).toBeInTheDocument();
    expect(screen.getByText('apps.attendance.name')).toBeInTheDocument();
    expect((screen.getByText('apps').closest('button') as HTMLButtonElement).title).toBe('');
  });

  it('uses fallback app name when active app is undefined', () => {
    mockActiveApp = undefined;
    mockGetActiveApp.mockImplementation(() => mockActiveApp);

    renderAppSwitcher(false);

    expect(screen.getByText('apps.attendance.name')).toBeInTheDocument();
  });

  it('renders collapsed trigger with title and without expanded labels', () => {
    renderAppSwitcher(true);

    const trigger = screen.getByText('apps').closest('button') as HTMLButtonElement;
    expect(trigger.title).toBe('appSwitcher.title');
    expect(screen.queryByText('appSwitcher.currentApp')).not.toBeInTheDocument();
  });

  it('opens dropdown, renders app cards, and shows coming soon badge', () => {
    renderAppSwitcher();
    openDropdown();

    expect(screen.getByText('appSwitcher.subtitle')).toBeInTheDocument();
    expect(screen.getByText('apps.expenses.name')).toBeInTheDocument();
    expect(screen.getByText('appSwitcher.comingSoon')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /apps\.wiki\.name/i })).toBeDisabled();
  });

  it('navigates and closes dropdown when selecting enabled app', () => {
    renderAppSwitcher();
    openDropdown();

    fireEvent.click(screen.getByRole('button', { name: /apps\.expenses\.name/i }));

    expect(mockNavigate).toHaveBeenCalledWith({ to: '/expenses' });
    expect(screen.queryByText('appSwitcher.subtitle')).not.toBeInTheDocument();
  });

  it('executes non-navigating branch for disabled or coming-soon app', () => {
    renderAppSwitcher();
    openDropdown();

    const wikiButton = screen.getByRole('button', { name: /apps\.wiki\.name/i }) as HTMLButtonElement;
    wikiButton.disabled = false;
    fireEvent.click(wikiButton);

    expect(mockNavigate).not.toHaveBeenCalled();
    expect(screen.getByText('appSwitcher.subtitle')).toBeInTheDocument();
  });

  it('keeps dropdown open for inside click and closes on outside click', () => {
    renderAppSwitcher();
    openDropdown();

    fireEvent.mouseDown(screen.getByText('appSwitcher.title'));
    expect(screen.getByText('appSwitcher.subtitle')).toBeInTheDocument();

    fireEvent.mouseDown(document.body);
    expect(screen.queryByText('appSwitcher.subtitle')).not.toBeInTheDocument();
  });

  it('registers and removes resize/scroll listeners only while open', () => {
    const addSpy = vi.spyOn(window, 'addEventListener');
    const removeSpy = vi.spyOn(window, 'removeEventListener');
    const docAddSpy = vi.spyOn(document, 'addEventListener');
    const docRemoveSpy = vi.spyOn(document, 'removeEventListener');

    const { unmount } = renderAppSwitcher();
    expect(docAddSpy).toHaveBeenCalledWith('mousedown', expect.any(Function));

    openDropdown();
    expect(addSpy).toHaveBeenCalledWith('resize', expect.any(Function));
    expect(addSpy).toHaveBeenCalledWith('scroll', expect.any(Function), true);

    fireEvent.click(screen.getByText('apps').closest('button') as HTMLButtonElement);
    expect(removeSpy).toHaveBeenCalledWith('resize', expect.any(Function));
    expect(removeSpy).toHaveBeenCalledWith('scroll', expect.any(Function), true);

    unmount();
    expect(docRemoveSpy).toHaveBeenCalledWith('mousedown', expect.any(Function));
  });

  it('returns early in position update when trigger ref becomes null', () => {
    const addSpy = vi.spyOn(window, 'addEventListener');
    const rectSpy = vi.spyOn(HTMLButtonElement.prototype, 'getBoundingClientRect');

    const { unmount } = renderAppSwitcher();
    openDropdown();

    const resizeCall = addSpy.mock.calls.find(
      (call) => call[0] === 'resize' && typeof call[1] === 'function',
    );
    expect(resizeCall).toBeTruthy();
    const resizeHandler = resizeCall?.[1] as () => void;

    unmount();
    resizeHandler();

    expect(rectSpy).toHaveBeenCalledTimes(1);
  });

  it('positions dropdown with left and top clamping in expanded mode', () => {
    setWindowSize(400, 200);
    vi.spyOn(HTMLButtonElement.prototype, 'getBoundingClientRect').mockReturnValue({
      top: 20,
      left: 0,
      right: 40,
      bottom: 40,
      width: 40,
      height: 20,
      x: 0,
      y: 20,
      toJSON: () => ({}),
    } as DOMRect);

    renderAppSwitcher(false);
    openDropdown();

    const dropdown = screen.getByText('appSwitcher.title').closest('div.fixed') as HTMLDivElement;
    expect(dropdown.style.left).toBe('8px');
    expect(dropdown.style.top).toBe('8px');
  });

  it('positions dropdown with right-edge clamping in collapsed mode', () => {
    setWindowSize(400, 700);
    vi.spyOn(HTMLButtonElement.prototype, 'getBoundingClientRect').mockReturnValue({
      top: 50,
      left: 320,
      right: 390,
      bottom: 80,
      width: 70,
      height: 30,
      x: 320,
      y: 50,
      toJSON: () => ({}),
    } as DOMRect);

    renderAppSwitcher(true);
    openDropdown();

    const dropdown = screen.getByText('appSwitcher.title').closest('div.fixed') as HTMLDivElement;
    expect(dropdown.style.left).toBe('112px');
    expect(dropdown.style.top).toBe('50px');
  });
});
