import { Outlet, Link, useNavigate, useLocation } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useThemeStore, applyTheme } from '@/stores/themeStore';
import { useState, useEffect, useMemo } from 'react';
import { AppSwitcher } from './AppSwitcher';
import { getActiveApp } from '@/config/apps';

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

  const navItems = useMemo(() => {
    const activeApp = getActiveApp(location.pathname);

    // 経費精算アプリのナビ
    if (activeApp?.id === 'expenses') {
      return [
        { to: '/expenses' as const, icon: 'dashboard', label: t('expenses.nav.dashboard') },
        { to: '/expenses/new' as const, icon: 'add_card', label: t('expenses.nav.newExpense') },
        { to: '/expenses/history' as const, icon: 'history', label: t('expenses.nav.history') },
        { to: '/expenses/report' as const, icon: 'bar_chart', label: t('expenses.nav.report') },
        { to: '/expenses/templates' as const, icon: 'content_copy', label: t('expenses.nav.templates') },
        { to: '/expenses/notifications' as const, icon: 'notifications', label: t('expenses.nav.notifications') },
        ...(user?.role === 'admin' || user?.role === 'manager'
          ? [
              { to: '/expenses/approve' as const, icon: 'fact_check', label: t('expenses.nav.approve') },
              { to: '/expenses/advanced-approve' as const, icon: 'account_tree', label: t('expenses.nav.advancedApprove') },
              { to: '/expenses/policy' as const, icon: 'policy', label: t('expenses.nav.policy') },
            ]
          : []),
      ];
    }

    // デフォルト: 勤怠管理アプリのナビ
    return [
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
  }, [location.pathname, user?.role, t]);

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname.startsWith(path);
  };

  return (
    <div className="flex h-screen overflow-hidden relative">
      {/* Aurora Background */}
      <div className="aurora-bg">
        <div className="aurora-orb-1" />
        <div className="aurora-orb-2" />
      </div>

      {/* Noise Overlay */}
      <div className="noise-overlay" />

      {/* サイドバー */}
      <aside className={`${sidebarCollapsed ? 'w-16' : 'w-64'} flex-shrink-0 sidebar-glass flex flex-col transition-all duration-300 relative z-10`}>
        {/* ロゴ */}
        <div className={`p-4 flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3 px-6'}`}>
          <div className="size-10 rounded-xl flex items-center justify-center flex-shrink-0 gradient-primary shadow-glow-sm">
            <MaterialIcon name="schedule" className="text-2xl text-white" />
          </div>
          {!sidebarCollapsed && (
            <div>
              <h1 className="font-bold text-lg leading-tight gradient-text">{t('common.appName')}</h1>
              <p className="text-xs text-muted-foreground">{t('common.subtitle')}</p>
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
            className="flex items-center justify-center w-full py-2 rounded-lg text-muted-foreground hover:text-foreground nav-item-hover transition-colors"
            title={sidebarCollapsed ? t('sidebar.expand') : t('sidebar.collapse')}
          >
            <MaterialIcon name={sidebarCollapsed ? 'chevron_right' : 'chevron_left'} />
          </button>
        </div>

        {/* ナビゲーション */}
        <nav className={`flex-1 ${sidebarCollapsed ? 'px-2' : 'px-3'} space-y-1 mt-4 overflow-y-auto scrollbar-thin`}>
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              title={sidebarCollapsed ? item.label : undefined}
              className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-xl transition-all duration-200 text-sm ${isActive(item.to)
                  ? 'nav-item-active font-semibold'
                  : 'text-muted-foreground hover:text-foreground nav-item-hover'
                }`}
            >
              <MaterialIcon name={item.icon} className={isActive(item.to) ? 'text-indigo-400' : ''} />
              {!sidebarCollapsed && <span>{item.label}</span>}
            </Link>
          ))}
        </nav>

        {/* 設定とユーザー情報 */}
        <div className={`p-4 border-t border-white/5 space-y-1 ${sidebarCollapsed ? 'px-2' : ''}`}>
          <button
            onClick={toggleLanguage}
            title={sidebarCollapsed ? (i18n.language === 'ja' ? 'English' : '日本語') : undefined}
            className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-xl text-muted-foreground hover:text-foreground nav-item-hover transition-all w-full text-sm`}
          >
            <MaterialIcon name="language" />
            {!sidebarCollapsed && <span>{i18n.language === 'ja' ? 'English' : '日本語'}</span>}
          </button>
          <button
            onClick={cycleTheme}
            title={sidebarCollapsed ? getThemeLabel() : undefined}
            className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-xl text-muted-foreground hover:text-foreground nav-item-hover transition-all w-full text-sm`}
          >
            <MaterialIcon name={getThemeIcon()} />
            {!sidebarCollapsed && <span>{getThemeLabel()}</span>}
          </button>
          <button
            onClick={handleLogout}
            title={sidebarCollapsed ? t('common.logout') : undefined}
            className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-xl text-muted-foreground hover:text-red-400 nav-item-hover transition-all w-full text-sm`}
          >
            <MaterialIcon name="logout" />
            {!sidebarCollapsed && <span>{t('common.logout')}</span>}
          </button>

          {/* ユーザープロフィール */}
          {user && (
            <div className={`flex items-center ${sidebarCollapsed ? 'justify-center py-3' : 'gap-3 px-3 py-4'} mt-2 glass-subtle rounded-2xl`}>
              <div className="size-10 rounded-full flex items-center justify-center flex-shrink-0 gradient-primary-subtle border border-indigo-400/20">
                <MaterialIcon name="person" className="text-indigo-400" />
              </div>
              {!sidebarCollapsed && (
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold truncate">
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
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative z-10">
        {/* ヘッダー */}
        <header className="h-16 flex items-center justify-between px-8 glass border-b border-white/5">
          <div className="flex items-center gap-4 flex-1 max-w-xl">
            <div className="relative w-full">
              <MaterialIcon name="search" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('common.search')}
                className="w-full glass-input rounded-xl pl-10 pr-4 py-2.5 text-sm text-foreground placeholder:text-muted-foreground"
              />
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button className="relative p-2.5 text-muted-foreground hover:text-foreground glass-subtle rounded-xl transition-all hover:shadow-glow-sm">
              <MaterialIcon name="notifications" />
              <span className="absolute top-1.5 right-1.5 size-2.5 bg-emerald-400 rounded-full pulse-dot"></span>
            </button>
            <Link
              to="/attendance"
              className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all duration-300"
            >
              <MaterialIcon name="add" className="text-lg" />
              {t('attendance.clockIn')}
            </Link>
          </div>
        </header>

        {/* スクロール可能なコンテンツエリア */}
        <div className="flex-1 overflow-y-auto p-8 scrollbar-thin">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
