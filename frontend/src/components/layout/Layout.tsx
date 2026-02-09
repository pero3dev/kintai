import { Outlet, Link, useNavigate, useLocation } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useThemeStore, applyTheme } from '@/stores/themeStore';
import { useState, useEffect } from 'react';
import { AppSwitcher } from './AppSwitcher';

// Material Symbols icon component
function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function Layout() {
  const { t, i18n } = useTranslation();
  const { user, logout } = useAuthStore();
  const { theme, setTheme } = useThemeStore();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchQuery, setSearchQuery] = useState('');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    const saved = localStorage.getItem('sidebarCollapsed');
    return saved === 'true';
  });

  useEffect(() => {
    localStorage.setItem('sidebarCollapsed', String(sidebarCollapsed));
  }, [sidebarCollapsed]);

  const toggleSidebar = () => {
    setSidebarCollapsed(!sidebarCollapsed);
  };

  const handleLogout = () => {
    logout();
    navigate({ to: '/login' });
  };

  const toggleLanguage = () => {
    i18n.changeLanguage(i18n.language === 'ja' ? 'en' : 'ja');
  };

  const cycleTheme = () => {
    const nextTheme = theme === 'system' ? 'light' : theme === 'light' ? 'dark' : 'system';
    setTheme(nextTheme);
    applyTheme(nextTheme);
  };

  const getThemeIcon = () => {
    if (theme === 'light') return 'light_mode';
    if (theme === 'dark') return 'dark_mode';
    return 'brightness_auto';
  };

  const getThemeLabel = () => {
    if (theme === 'light') return t('theme.light');
    if (theme === 'dark') return t('theme.dark');
    return t('theme.system');
  };

  const navItems = [
    { to: '/' as const, icon: 'home', label: t('nav.home') },
    { to: '/attendance' as const, icon: 'schedule', label: t('nav.attendance') },
    { to: '/leaves' as const, icon: 'event_available', label: t('nav.leaves') },
    { to: '/overtime' as const, icon: 'more_time', label: t('nav.overtime') },
    { to: '/corrections' as const, icon: 'edit_note', label: t('nav.corrections') },
    { to: '/projects' as const, icon: 'folder_open', label: t('nav.projects') },
    { to: '/shifts' as const, icon: 'calendar_month', label: t('nav.shifts') },
    { to: '/holidays' as const, icon: 'celebration', label: t('nav.holidays') },
    { to: '/notifications' as const, icon: 'notifications', label: t('nav.notifications') },
    ...(user?.role === 'admin' || user?.role === 'manager'
      ? [
          { to: '/dashboard' as const, icon: 'dashboard', label: t('nav.dashboard') },
          { to: '/users' as const, icon: 'group', label: t('nav.users') },
          { to: '/export' as const, icon: 'download', label: t('nav.export') },
          { to: '/approval-flows' as const, icon: 'account_tree', label: t('nav.approvalFlows') },
        ]
      : []),
  ];

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname.startsWith(path);
  };

  return (
    <div className="flex h-screen overflow-hidden">
      {/* サイドバー */}
      <aside className={`${sidebarCollapsed ? 'w-16' : 'w-64'} flex-shrink-0 bg-card border-r border-border flex flex-col transition-all duration-300`}>
        {/* ロゴ */}
        <div className={`p-4 flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3 px-6'}`}>
          <div className="size-10 bg-primary rounded-lg flex items-center justify-center text-primary-foreground flex-shrink-0">
            <MaterialIcon name="schedule" className="text-2xl" />
          </div>
          {!sidebarCollapsed && (
            <div>
              <h1 className="font-bold text-lg leading-tight">{t('common.appName')}</h1>
              <p className="text-xs text-primary/70">{t('common.subtitle')}</p>
            </div>
          )}
        </div>

        {/* アプリスイッチャー */}
        <div className={`${sidebarCollapsed ? 'px-2' : 'px-4'} mb-2`}>
          <AppSwitcher collapsed={sidebarCollapsed} />
        </div>

        {/* 折り畳みトグルボタン */}
        <div className={`px-2 ${sidebarCollapsed ? 'flex justify-center' : 'px-4'}`}>
          <button
            onClick={toggleSidebar}
            className="flex items-center justify-center w-full py-2 rounded-lg text-muted-foreground hover:bg-primary/10 transition-colors"
            title={sidebarCollapsed ? t('sidebar.expand') : t('sidebar.collapse')}
          >
            <MaterialIcon name={sidebarCollapsed ? 'chevron_right' : 'chevron_left'} />
          </button>
        </div>

        {/* ナビゲーション */}
        <nav className={`flex-1 ${sidebarCollapsed ? 'px-2' : 'px-4'} space-y-1 mt-4 overflow-y-auto scrollbar-thin`}>
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              title={sidebarCollapsed ? item.label : undefined}
              className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2 rounded-lg transition-colors text-sm ${isActive(item.to)
                  ? 'bg-primary text-primary-foreground font-semibold'
                  : 'text-muted-foreground hover:bg-primary/10'
                }`}
            >
              <MaterialIcon name={item.icon} />
              {!sidebarCollapsed && <span>{item.label}</span>}
            </Link>
          ))}
        </nav>

        {/* 設定とユーザー情報 */}
        <div className={`p-4 border-t border-border space-y-2 ${sidebarCollapsed ? 'px-2' : ''}`}>
          <button
            onClick={toggleLanguage}
            title={sidebarCollapsed ? (i18n.language === 'ja' ? 'English' : '日本語') : undefined}
            className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-lg text-muted-foreground hover:bg-primary/10 transition-colors w-full`}
          >
            <MaterialIcon name="language" />
            {!sidebarCollapsed && <span>{i18n.language === 'ja' ? 'English' : '日本語'}</span>}
          </button>
          <button
            onClick={cycleTheme}
            title={sidebarCollapsed ? getThemeLabel() : undefined}
            className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-lg text-muted-foreground hover:bg-primary/10 transition-colors w-full`}
          >
            <MaterialIcon name={getThemeIcon()} />
            {!sidebarCollapsed && <span>{getThemeLabel()}</span>}
          </button>
          <button
            onClick={handleLogout}
            title={sidebarCollapsed ? t('common.logout') : undefined}
            className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-lg text-muted-foreground hover:bg-primary/10 transition-colors w-full`}
          >
            <MaterialIcon name="logout" />
            {!sidebarCollapsed && <span>{t('common.logout')}</span>}
          </button>

          {/* ユーザープロフィール */}
          {user && (
            <div className={`flex items-center ${sidebarCollapsed ? 'justify-center py-3' : 'gap-3 px-3 py-4'} mt-2 bg-black/20 rounded-xl`}>
              <div className="size-10 rounded-full bg-primary/20 flex items-center justify-center border-2 border-primary/30 flex-shrink-0">
                <MaterialIcon name="person" className="text-primary" />
              </div>
              {!sidebarCollapsed && (
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold truncate uppercase">
                    {user.last_name} {user.first_name}
                  </p>
                  <p className="text-xs text-muted-foreground truncate">
                    {t(`users.roles.${user.role}`)}
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      </aside>

      {/* メインコンテンツ */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* ヘッダー */}
        <header className="h-16 flex items-center justify-between px-8 bg-card border-b border-border">
          <div className="flex items-center gap-4 flex-1 max-w-xl">
            <div className="relative w-full">
              <MaterialIcon name="search" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('common.search')}
                className="w-full bg-black/20 border-none rounded-lg pl-10 pr-4 py-2 focus:ring-2 focus:ring-primary text-sm placeholder:text-muted-foreground"
              />
            </div>
          </div>
          <div className="flex items-center gap-4">
            <button className="relative p-2 text-muted-foreground hover:bg-primary/10 rounded-full transition-colors">
              <MaterialIcon name="notifications" />
              <span className="absolute top-1.5 right-1.5 size-2.5 bg-destructive border-2 border-card rounded-full"></span>
            </button>
            <Link
              to="/attendance"
              className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground font-bold text-sm rounded-lg hover:brightness-110 transition-all"
            >
              <MaterialIcon name="add" className="text-lg" />
              {t('attendance.clockIn')}
            </Link>
          </div>
        </header>

        {/* スクロール可能なコンテンツエリア */}
        <div className="flex-1 overflow-y-auto p-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
