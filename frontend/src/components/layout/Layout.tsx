import { Outlet, Link, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import {
  LayoutDashboard,
  Clock,
  CalendarDays,
  CalendarRange,
  Users,
  LogOut,
  Menu,
} from 'lucide-react';
import { useState } from 'react';

export function Layout() {
  const { t, i18n } = useTranslation();
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const handleLogout = () => {
    logout();
    navigate({ to: '/login' });
  };

  const toggleLanguage = () => {
    i18n.changeLanguage(i18n.language === 'ja' ? 'en' : 'ja');
  };

  const navItems = [
    { to: '/' as const, icon: LayoutDashboard, label: t('nav.dashboard') },
    { to: '/attendance' as const, icon: Clock, label: t('nav.attendance') },
    { to: '/leaves' as const, icon: CalendarDays, label: t('nav.leaves') },
    { to: '/shifts' as const, icon: CalendarRange, label: t('nav.shifts') },
    ...(user?.role === 'admin' || user?.role === 'manager'
      ? [{ to: '/users' as const, icon: Users, label: t('nav.users') }]
      : []),
  ];

  return (
    <div className="flex h-screen bg-background">
      {/* サイドバー */}
      <aside
        className={`${sidebarOpen ? 'w-64' : 'w-16'
          } bg-card border-r border-border transition-all duration-300 flex flex-col`}
      >
        {/* ロゴ */}
        <div className="h-16 flex items-center justify-between px-4 border-b border-border">
          {sidebarOpen && (
            <h1 className="text-lg font-bold text-primary truncate">
              {t('common.appName')}
            </h1>
          )}
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 rounded-md hover:bg-accent"
          >
            <Menu className="h-5 w-5" />
          </button>
        </div>

        {/* ナビゲーション */}
        <nav className="flex-1 p-2 space-y-1">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className="flex items-center gap-3 px-3 py-2 rounded-md text-sm hover:bg-accent transition-colors [&.active]:bg-primary [&.active]:text-primary-foreground"
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {sidebarOpen && <span>{item.label}</span>}
            </Link>
          ))}
        </nav>

        {/* ユーザー情報 */}
        <div className="border-t border-border p-4">
          {sidebarOpen && user && (
            <div className="mb-2 text-sm">
              <p className="font-medium">
                {user.last_name} {user.first_name}
              </p>
              <p className="text-muted-foreground text-xs">{user.email}</p>
            </div>
          )}
          <div className="flex items-center gap-2">
            <button
              onClick={toggleLanguage}
              className="text-xs px-2 py-1 rounded border border-border hover:bg-accent"
            >
              {i18n.language === 'ja' ? 'EN' : 'JA'}
            </button>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 text-sm text-destructive hover:text-destructive/80"
            >
              <LogOut className="h-4 w-4" />
              {sidebarOpen && t('common.logout')}
            </button>
          </div>
        </div>
      </aside>

      {/* メインコンテンツ */}
      <main className="flex-1 overflow-auto">
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
