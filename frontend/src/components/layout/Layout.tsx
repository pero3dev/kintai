import { Outlet, Link, useNavigate, useLocation } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useThemeStore, applyTheme } from '@/stores/themeStore';
import { useState, useEffect, useMemo, useCallback } from 'react';
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
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false);

  // モバイルドロワーが開いている時はスクロールを無効化
  useEffect(() => {
    if (mobileDrawerOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => { document.body.style.overflow = ''; };
  }, [mobileDrawerOpen]);

  // ルート変更時にドロワーを閉じる
  useEffect(() => {
    setMobileDrawerOpen(false);
  }, [location.pathname]);

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

  const cycleTheme = useCallback(() => {
    const nextTheme = theme === 'system' ? 'light' : theme === 'light' ? 'dark' : 'system';
    setTheme(nextTheme);
    applyTheme(nextTheme);
  }, [theme, setTheme]);

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
    const wikiLabels = i18n.language.startsWith('ja')
      ? {
          overview: '概要',
          architecture: 'アーキテクチャ',
          backend: 'バックエンド',
          frontend: 'フロントエンド',
          infrastructure: 'インフラ',
          testing: 'テスト',
        }
      : {
          overview: 'Overview',
          architecture: 'Architecture',
          backend: 'Backend',
          frontend: 'Frontend',
          infrastructure: 'Infrastructure',
          testing: 'Testing',
        };

    // 経費精算アプリのナビ
    if (activeApp?.id === 'expenses') {
      return [
        { to: '/expenses' as const, icon: 'dashboard', label: t('expenses.nav.dashboard'), mobile: true },
        { to: '/expenses/new' as const, icon: 'add_card', label: t('expenses.nav.newExpense'), mobile: true },
        { to: '/expenses/history' as const, icon: 'history', label: t('expenses.nav.history'), mobile: true },
        { to: '/expenses/report' as const, icon: 'bar_chart', label: t('expenses.nav.report'), mobile: false },
        { to: '/expenses/templates' as const, icon: 'content_copy', label: t('expenses.nav.templates'), mobile: false },
        { to: '/expenses/notifications' as const, icon: 'notifications', label: t('expenses.nav.notifications'), mobile: true },
        ...(user?.role === 'admin' || user?.role === 'manager'
          ? [
              { to: '/expenses/approve' as const, icon: 'fact_check', label: t('expenses.nav.approve'), mobile: false },
              { to: '/expenses/advanced-approve' as const, icon: 'account_tree', label: t('expenses.nav.advancedApprove'), mobile: false },
              { to: '/expenses/policy' as const, icon: 'policy', label: t('expenses.nav.policy'), mobile: false },
            ]
          : []),
      ];
    }

    // 人事管理アプリのナビ
    if (activeApp?.id === 'hr') {
      return [
        { to: '/hr' as const, icon: 'dashboard', label: t('hr.nav.dashboard'), mobile: true },
        { to: '/hr/employees' as const, icon: 'people', label: t('hr.nav.employees'), mobile: true },
        { to: '/hr/departments' as const, icon: 'account_tree', label: t('hr.nav.departments'), mobile: false },
        { to: '/hr/evaluations' as const, icon: 'rate_review', label: t('hr.nav.evaluations'), mobile: true },
        { to: '/hr/goals' as const, icon: 'flag', label: t('hr.nav.goals'), mobile: true },
        { to: '/hr/training' as const, icon: 'school', label: t('hr.nav.training'), mobile: false },
        { to: '/hr/recruitment' as const, icon: 'work', label: t('hr.nav.recruitment'), mobile: false },
        { to: '/hr/documents' as const, icon: 'description', label: t('hr.nav.documents'), mobile: false },
        { to: '/hr/announcements' as const, icon: 'campaign', label: t('hr.nav.announcements'), mobile: true },
        { to: '/hr/attendance-integration' as const, icon: 'schedule', label: t('hr.nav.attendanceIntegration'), mobile: false },
        { to: '/hr/org-chart' as const, icon: 'lan', label: t('hr.nav.orgChart'), mobile: false },
        { to: '/hr/one-on-one' as const, icon: 'groups', label: t('hr.nav.oneOnOne'), mobile: false },
        { to: '/hr/skill-map' as const, icon: 'psychology', label: t('hr.nav.skillMap'), mobile: false },
        { to: '/hr/salary' as const, icon: 'payments', label: t('hr.nav.salary'), mobile: false },
        { to: '/hr/onboarding' as const, icon: 'waving_hand', label: t('hr.nav.onboarding'), mobile: false },
        { to: '/hr/offboarding' as const, icon: 'logout', label: t('hr.nav.offboarding'), mobile: false },
        { to: '/hr/survey' as const, icon: 'poll', label: t('hr.nav.survey'), mobile: false },
      ];
    }

    // 社内Wikiアプリのナビ
    if (activeApp?.id === 'wiki') {
      return [
        { to: '/wiki' as const, icon: 'home', label: wikiLabels.overview, mobile: true },
        { to: '/wiki/architecture' as const, icon: 'lan', label: wikiLabels.architecture, mobile: true },
        { to: '/wiki/backend' as const, icon: 'dns', label: wikiLabels.backend, mobile: true },
        { to: '/wiki/frontend' as const, icon: 'web', label: wikiLabels.frontend, mobile: true },
        { to: '/wiki/infrastructure' as const, icon: 'cloud', label: wikiLabels.infrastructure, mobile: false },
        { to: '/wiki/testing' as const, icon: 'fact_check', label: wikiLabels.testing, mobile: false },
      ];
    }

    // デフォルト: 勤怠管理アプリのナビ
    return [
      { to: '/' as const, icon: 'home', label: t('nav.home'), mobile: true },
      { to: '/attendance' as const, icon: 'schedule', label: t('nav.attendance'), mobile: true },
      { to: '/leaves' as const, icon: 'event_available', label: t('nav.leaves'), mobile: true },
      { to: '/overtime' as const, icon: 'more_time', label: t('nav.overtime'), mobile: false },
      { to: '/corrections' as const, icon: 'edit_note', label: t('nav.corrections'), mobile: false },
      { to: '/projects' as const, icon: 'folder_open', label: t('nav.projects'), mobile: false },
      { to: '/shifts' as const, icon: 'calendar_month', label: t('nav.shifts'), mobile: false },
      { to: '/holidays' as const, icon: 'celebration', label: t('nav.holidays'), mobile: false },
      { to: '/notifications' as const, icon: 'notifications', label: t('nav.notifications'), mobile: true },
      ...(user?.role === 'admin' || user?.role === 'manager'
        ? [
            { to: '/dashboard' as const, icon: 'dashboard', label: t('nav.dashboard'), mobile: false },
            { to: '/users' as const, icon: 'group', label: t('nav.users'), mobile: false },
            { to: '/export' as const, icon: 'download', label: t('nav.export'), mobile: false },
            { to: '/approval-flows' as const, icon: 'account_tree', label: t('nav.approvalFlows'), mobile: false },
          ]
        : []),
    ];
  }, [location.pathname, user?.role, t, i18n.language]);

  // ボトムタブ用（最大4つ + Moreボタン）
  const mobileTabItems = useMemo(() => navItems.filter(i => i.mobile).slice(0, 4), [navItems]);

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

      {/* ====== モバイルドロワーオーバーレイ ====== */}
      {mobileDrawerOpen && (
        <div
          className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 md:hidden"
          onClick={() => setMobileDrawerOpen(false)}
        />
      )}

      {/* ====== モバイルドロワー ====== */}
      <aside className={`
        fixed inset-y-0 left-0 w-72 sidebar-glass flex flex-col z-50 md:hidden
        transform transition-transform duration-300 ease-out
        ${mobileDrawerOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        {/* ドロワーヘッダー */}
        <div className="p-4 flex items-center gap-3 px-5">
          <div className="size-10 rounded-xl flex items-center justify-center flex-shrink-0 gradient-primary shadow-glow-sm">
            <MaterialIcon name="schedule" className="text-2xl text-white" />
          </div>
          <div className="flex-1">
            <h1 className="font-bold text-lg leading-tight gradient-text">{t('common.appName')}</h1>
            <p className="text-xs text-muted-foreground">{t('common.subtitle')}</p>
          </div>
          <button
            onClick={() => setMobileDrawerOpen(false)}
            className="p-2 rounded-xl text-muted-foreground hover:text-foreground nav-item-hover"
          >
            <MaterialIcon name="close" />
          </button>
        </div>

        {/* アプリスイッチャー */}
        <div className="px-4 mb-2">
          <AppSwitcher collapsed={false} />
        </div>

        {/* ナビゲーション */}
        <nav className="flex-1 px-3 space-y-1 mt-2 overflow-y-auto scrollbar-thin">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={`flex items-center gap-3 px-3 py-3 rounded-xl transition-all duration-200 text-sm ${isActive(item.to)
                  ? 'nav-item-active font-semibold'
                  : 'text-muted-foreground hover:text-foreground nav-item-hover'
                }`}
            >
              <MaterialIcon name={item.icon} className={isActive(item.to) ? 'text-indigo-400' : ''} />
              <span>{item.label}</span>
            </Link>
          ))}
        </nav>

        {/* ドロワーフッター */}
        <div className="p-4 border-t border-white/5 space-y-1">
          <button
            onClick={toggleLanguage}
            className="flex items-center gap-3 px-3 py-3 rounded-xl text-muted-foreground hover:text-foreground nav-item-hover transition-all w-full text-sm"
          >
            <MaterialIcon name="language" />
            <span>{i18n.language === 'ja' ? 'English' : '日本語'}</span>
          </button>
          <button
            onClick={cycleTheme}
            className="flex items-center gap-3 px-3 py-3 rounded-xl text-muted-foreground hover:text-foreground nav-item-hover transition-all w-full text-sm"
          >
            <MaterialIcon name={getThemeIcon()} />
            <span>{getThemeLabel()}</span>
          </button>
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-3 py-3 rounded-xl text-muted-foreground hover:text-red-400 nav-item-hover transition-all w-full text-sm"
          >
            <MaterialIcon name="logout" />
            <span>{t('common.logout')}</span>
          </button>

          {user && (
            <div className="flex items-center gap-3 px-3 py-4 mt-2 glass-subtle rounded-2xl">
              <div className="size-10 rounded-full flex items-center justify-center flex-shrink-0 gradient-primary-subtle border border-indigo-400/20">
                <MaterialIcon name="person" className="text-indigo-400" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold truncate">{user.last_name} {user.first_name}</p>
                <p className="text-xs text-muted-foreground truncate">{t(`users.roles.${user.role}`)}</p>
              </div>
            </div>
          )}
        </div>
      </aside>

      {/* ====== デスクトップサイドバー ====== */}
      <aside className={`${sidebarCollapsed ? 'w-16' : 'w-64'} flex-shrink-0 sidebar-glass hidden md:flex flex-col transition-all duration-300 relative z-10`}>
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
        {/* ====== モバイルヘッダー ====== */}
        <header className="h-14 flex items-center justify-between px-4 glass border-b border-white/5 md:hidden">
          <button
            onClick={() => setMobileDrawerOpen(true)}
            className="p-2 -ml-1 rounded-xl text-muted-foreground hover:text-foreground transition-colors"
          >
            <MaterialIcon name="menu" className="text-2xl" />
          </button>
          <h1 className="font-bold text-base gradient-text">{t('common.appName')}</h1>
          <div className="flex items-center gap-1">
            <button
              onClick={cycleTheme}
              className="p-2 rounded-xl text-muted-foreground hover:text-foreground transition-colors"
            >
              <MaterialIcon name={getThemeIcon()} className="text-xl" />
            </button>
            {user && (
              <div className="size-8 rounded-full flex items-center justify-center gradient-primary-subtle border border-indigo-400/20">
                <MaterialIcon name="person" className="text-indigo-400 text-sm" />
              </div>
            )}
          </div>
        </header>

        {/* ====== デスクトップヘッダー ====== */}
        <header className="h-16 hidden md:flex items-center justify-between px-8 glass border-b border-white/5">
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
        <div className="flex-1 overflow-y-auto p-4 md:p-8 pb-20 md:pb-8 scrollbar-thin">
          <Outlet />
        </div>
      </main>

      {/* ====== モバイルボトムタブバー ====== */}
      <nav className="fixed bottom-0 left-0 right-0 z-30 md:hidden mobile-bottom-nav safe-area-bottom">
        <div className="flex items-stretch">
          {mobileTabItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={`flex-1 flex flex-col items-center justify-center py-2 gap-0.5 transition-colors min-h-[56px] ${
                isActive(item.to)
                  ? 'text-indigo-500 dark:text-indigo-400'
                  : 'text-muted-foreground'
              }`}
            >
              <MaterialIcon name={item.icon} className={`text-xl ${isActive(item.to) ? 'text-indigo-500 dark:text-indigo-400' : ''}`} />
              <span className="text-[10px] font-medium leading-tight">{item.label}</span>
              {isActive(item.to) && (
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full gradient-primary" />
              )}
            </Link>
          ))}
          {/* Moreボタン */}
          <button
            onClick={() => setMobileDrawerOpen(true)}
            className="flex-1 flex flex-col items-center justify-center py-2 gap-0.5 text-muted-foreground min-h-[56px]"
          >
            <MaterialIcon name="more_horiz" className="text-xl" />
            <span className="text-[10px] font-medium leading-tight">{t('common.more')}</span>
          </button>
        </div>
      </nav>
    </div>
  );
}
