import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { getAvailableApps } from '@/config/apps';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

// Material Symbols icon component
function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function HomeDashboardPage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const dateLocale = i18n.language === 'ja' ? ja : enUS;
  const availableApps = getAvailableApps(user?.role);

  // 今日の勤怠状態を取得
  const { data: todayAttendance } = useQuery({
    queryKey: ['attendance', 'today'],
    queryFn: () => api.attendance.getToday(),
  });

  // 通知を取得
  const { data: notifications } = useQuery({
    queryKey: ['notifications', 'home'],
    queryFn: () => api.notifications.getList({ page: 1, page_size: 5 }),
  });

  // 未読カウント
  const { data: unreadCountData } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () => api.notifications.getUnreadCount(),
  });

  // 承認待ち件数（管理者のみ）
  const { data: pendingLeaves } = useQuery({
    queryKey: ['leaves', 'pending'],
    queryFn: () => api.leaves.getPending(),
    enabled: user?.role === 'admin' || user?.role === 'manager',
  });

  const unreadCount = (unreadCountData as { count: number } | undefined)?.count || 0;
  const pendingCount = (pendingLeaves as { data?: unknown[] } | undefined)?.data?.length || 0;

  const today = new Date();
  const greeting = () => {
    const hour = today.getHours();
    if (hour < 12) return t('home.greetingMorning');
    if (hour < 18) return t('home.greetingAfternoon');
    return t('home.greetingEvening');
  };

  const attendanceStatus = todayAttendance as { clock_in_time?: string; clock_out_time?: string } | undefined;

  return (
    <div className="space-y-8 animate-fade-in">
      {/* ウェルカムヘッダー */}
      <div className="glass-card rounded-2xl p-8 relative overflow-hidden">
        {/* 内部のAuroraエフェクト */}
        <div className="absolute inset-0 pointer-events-none">
          <div className="absolute -top-1/2 -right-1/4 w-1/2 h-full bg-indigo-400/10 rounded-full blur-[80px]" />
          <div className="absolute -bottom-1/2 -left-1/4 w-1/2 h-full bg-emerald-400/8 rounded-full blur-[100px]" />
        </div>
        <div className="flex items-center justify-between relative z-10">
          <div>
            <p className="text-muted-foreground text-sm">
              {format(today, i18n.language === 'ja' ? 'yyyy年M月d日 (EEEE)' : 'EEEE, MMMM d, yyyy', { locale: dateLocale })}
            </p>
            <h1 className="text-3xl font-bold mt-2 gradient-text">
              {greeting()}, {user?.first_name || 'User'}
            </h1>
            <p className="text-muted-foreground mt-2">{t('home.welcomeMessage')}</p>
          </div>
          <div className="hidden md:block">
            <div className="size-20 rounded-2xl gradient-primary-subtle flex items-center justify-center">
              <MaterialIcon name="waving_hand" className="text-5xl text-indigo-400" />
            </div>
          </div>
        </div>
      </div>

      {/* クイックアクション */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {/* 出退勤状態 */}
        <div className="glass-card rounded-2xl p-6 stat-shimmer">
          <div className="flex items-center gap-3 mb-5">
            <div className="size-10 rounded-xl bg-indigo-400/10 flex items-center justify-center border border-indigo-400/20">
              <MaterialIcon name="schedule" className="text-indigo-400" />
            </div>
            <h2 className="font-semibold">{t('home.todayStatus')}</h2>
          </div>
          {attendanceStatus?.clock_in_time ? (
            <div className="space-y-3">
              <div className="flex justify-between items-center py-2 border-b border-white/5">
                <span className="text-muted-foreground text-sm">{t('attendance.clockIn')}</span>
                <span className="font-mono font-semibold text-emerald-400">
                  {new Date(attendanceStatus.clock_in_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
              {attendanceStatus.clock_out_time ? (
                <div className="flex justify-between items-center py-2">
                  <span className="text-muted-foreground text-sm">{t('attendance.clockOut')}</span>
                  <span className="font-mono font-semibold text-indigo-400">
                    {new Date(attendanceStatus.clock_out_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
              ) : (
                <Link
                  to="/attendance"
                  className="block w-full mt-3 px-4 py-2.5 bg-red-500/10 text-red-400 border border-red-500/20 text-center rounded-xl hover:bg-red-500/20 transition-all font-medium text-sm"
                >
                  {t('attendance.clockOut')}
                </Link>
              )}
            </div>
          ) : (
            <div>
              <p className="text-muted-foreground text-sm mb-4">{t('home.notClockedIn')}</p>
              <Link
                to="/attendance"
                className="block w-full px-4 py-2.5 gradient-primary text-white text-center rounded-xl hover:shadow-glow-md transition-all font-medium text-sm"
              >
                {t('attendance.clockIn')}
              </Link>
            </div>
          )}
        </div>

        {/* 未読通知 */}
        <div className="glass-card rounded-2xl p-6 stat-shimmer">
          <div className="flex items-center gap-3 mb-5">
            <div className="size-10 rounded-xl bg-amber-400/10 flex items-center justify-center border border-amber-400/20">
              <MaterialIcon name="notifications" className="text-amber-400" />
            </div>
            <h2 className="font-semibold">{t('home.notifications')}</h2>
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-4xl font-bold gradient-text">{unreadCount}</p>
              <p className="text-muted-foreground text-sm mt-1">{t('home.unreadNotifications')}</p>
            </div>
            <Link
              to="/notifications"
              className="px-4 py-2 glass-input rounded-xl text-indigo-400 text-sm font-medium hover:border-indigo-400/30 transition-all"
            >
              {t('home.viewAll')}
            </Link>
          </div>
        </div>

        {/* 承認待ち（管理者のみ） */}
        {(user?.role === 'admin' || user?.role === 'manager') && (
          <div className="glass-card rounded-2xl p-6 stat-shimmer">
            <div className="flex items-center gap-3 mb-5">
              <div className="size-10 rounded-xl bg-purple-400/10 flex items-center justify-center border border-purple-400/20">
                <MaterialIcon name="pending_actions" className="text-purple-400" />
              </div>
              <h2 className="font-semibold">{t('home.pendingApprovals')}</h2>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-4xl font-bold gradient-text">{pendingCount}</p>
                <p className="text-muted-foreground text-sm mt-1">{t('home.awaitingReview')}</p>
              </div>
              <Link
                to="/leaves"
                className="px-4 py-2 glass-input rounded-xl text-indigo-400 text-sm font-medium hover:border-indigo-400/30 transition-all"
              >
                {t('home.review')}
              </Link>
            </div>
          </div>
        )}
      </div>

      {/* アプリ一覧 */}
      <div>
        <h2 className="text-xl font-bold mb-5">{t('home.apps')}</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {availableApps.map((app) => {
          const isClickable = app.enabled && !app.comingSoon;
          if (isClickable) {
            return (
              <Link
                key={app.id}
                to={app.basePath as '/'}
                className="group flex flex-col items-center gap-3 p-6 glass-card rounded-2xl transition-all hover:shadow-glow-sm cursor-pointer"
              >
                <div className={`size-14 rounded-xl flex items-center justify-center text-white ${app.color} group-hover:scale-110 transition-transform duration-300`}>
                  <MaterialIcon name={app.icon} className="text-2xl" />
                </div>
                <div className="text-center">
                  <p className="font-medium text-sm">{t(app.nameKey)}</p>
                </div>
              </Link>
            );
          }
          return (
            <div
              key={app.id}
              className="flex flex-col items-center gap-3 p-6 glass-subtle rounded-2xl opacity-40 cursor-not-allowed"
            >
              <div className={`size-14 rounded-xl flex items-center justify-center text-white ${app.color}`}>
                <MaterialIcon name={app.icon} className="text-2xl" />
              </div>
              <div className="text-center">
                <p className="font-medium text-sm">{t(app.nameKey)}</p>
                <span className="inline-block mt-1 px-2 py-0.5 bg-amber-400/10 text-amber-400 text-[10px] font-bold rounded-full border border-amber-400/20">
                  {t('appSwitcher.comingSoon')}
                </span>
              </div>
            </div>
          );
        })}
        </div>
      </div>

      {/* クイックリンク */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* よく使う機能 */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="font-bold mb-5 flex items-center gap-2">
            <MaterialIcon name="star" className="text-amber-400" />
            {t('home.quickAccess')}
          </h2>
          <div className="grid grid-cols-2 gap-2">
            <Link
              to="/leaves"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <MaterialIcon name="event_available" className="text-emerald-400" />
              <span className="text-sm font-medium">{t('nav.leaves')}</span>
            </Link>
            <Link
              to="/overtime"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <MaterialIcon name="more_time" className="text-orange-400" />
              <span className="text-sm font-medium">{t('nav.overtime')}</span>
            </Link>
            <Link
              to="/projects"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <MaterialIcon name="folder_open" className="text-indigo-400" />
              <span className="text-sm font-medium">{t('nav.projects')}</span>
            </Link>
            <Link
              to="/holidays"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <MaterialIcon name="celebration" className="text-pink-400" />
              <span className="text-sm font-medium">{t('nav.holidays')}</span>
            </Link>
          </div>
        </div>

        {/* 最近の通知 */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="font-bold mb-5 flex items-center gap-2">
            <MaterialIcon name="notifications_active" className="text-indigo-400" />
            {t('home.recentNotifications')}
          </h2>
          <div className="space-y-3">
            {(notifications as { data?: { id: string; title: string; created_at: string; is_read: boolean }[] } | undefined)?.data?.slice(0, 3).map((n) => (
              <div
                key={n.id}
                className={`p-3 rounded-xl transition-all ${n.is_read ? 'glass-subtle' : 'glass-card border-indigo-400/20'}`}
              >
                <p className="text-sm font-medium truncate">{n.title}</p>
                <p className="text-xs text-muted-foreground mt-1">
                  {format(new Date(n.created_at), i18n.language === 'ja' ? 'M月d日 HH:mm' : 'MMM d, HH:mm', { locale: dateLocale })}
                </p>
              </div>
            ))}
            {(!notifications || !(notifications as { data?: unknown[] }).data?.length) && (
              <p className="text-center text-muted-foreground text-sm py-4">{t('notifications.empty')}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
