import { Outlet, Link, useNavigate, useLocation } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';

// Material Symbols icon component
function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function Layout() {
  const { t, i18n } = useTranslation();
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchQuery, setSearchQuery] = useState('');

  const handleLogout = () => {
    logout();
    navigate({ to: '/login' });
  };

  const toggleLanguage = () => {
    i18n.changeLanguage(i18n.language === 'ja' ? 'en' : 'ja');
  };

  const navItems = [
    { to: '/' as const, icon: 'dashboard', label: t('nav.dashboard') },
    { to: '/attendance' as const, icon: 'schedule', label: t('nav.attendance') },
    { to: '/leaves' as const, icon: 'event_available', label: t('nav.leaves') },
    { to: '/shifts' as const, icon: 'calendar_month', label: t('nav.shifts') },
    ...(user?.role === 'admin' || user?.role === 'manager'
      ? [{ to: '/users' as const, icon: 'group', label: t('nav.users') }]
      : []),
  ];

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname.startsWith(path);
  };

  return (
    <div className="flex h-screen overflow-hidden">
      {/* サイドバー */}
      <aside className="w-64 flex-shrink-0 bg-card border-r border-border flex flex-col">
        {/* ロゴ */}
        <div className="p-6 flex items-center gap-3">
          <div className="size-10 bg-primary rounded-lg flex items-center justify-center text-primary-foreground">
            <MaterialIcon name="schedule" className="text-2xl" />
          </div>
          <div>
            <h1 className="font-bold text-lg leading-tight">{t('common.appName')}</h1>
            <p className="text-xs text-primary/70">{t('common.subtitle') || '勤怠管理'}</p>
          </div>
        </div>

        {/* ナビゲーション */}
        <nav className="flex-1 px-4 space-y-2 mt-4">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors ${isActive(item.to)
                  ? 'bg-primary text-primary-foreground font-semibold'
                  : 'text-muted-foreground hover:bg-primary/10'
                }`}
            >
              <MaterialIcon name={item.icon} />
              <span>{item.label}</span>
            </Link>
          ))}
        </nav>

        {/* 設定とユーザー情報 */}
        <div className="p-4 border-t border-border space-y-2">
          <button
            onClick={toggleLanguage}
            className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-muted-foreground hover:bg-primary/10 transition-colors w-full"
          >
            <MaterialIcon name="language" />
            <span>{i18n.language === 'ja' ? 'English' : '日本語'}</span>
          </button>
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-muted-foreground hover:bg-primary/10 transition-colors w-full"
          >
            <MaterialIcon name="logout" />
            <span>{t('common.logout')}</span>
          </button>

          {/* ユーザープロフィール */}
          {user && (
            <div className="flex items-center gap-3 px-3 py-4 mt-2 bg-black/20 rounded-xl">
              <div className="size-10 rounded-full bg-primary/20 flex items-center justify-center border-2 border-primary/30">
                <MaterialIcon name="person" className="text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold truncate uppercase">
                  {user.last_name} {user.first_name}
                </p>
                <p className="text-xs text-muted-foreground truncate">
                  {user.role === 'admin' ? '管理者' : user.role === 'manager' ? 'マネージャー' : '従業員'}
                </p>
              </div>
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
                placeholder={t('common.search') || '検索...'}
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
