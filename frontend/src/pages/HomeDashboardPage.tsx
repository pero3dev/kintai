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
    <div className="space-y-8">
      {/* ウェルカムヘッダー */}
      <div className="bg-gradient-to-r from-primary to-primary/70 rounded-2xl p-8 text-primary-foreground">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-primary-foreground/80 text-sm">
              {format(today, i18n.language === 'ja' ? 'yyyy年M月d日 (EEEE)' : 'EEEE, MMMM d, yyyy', { locale: dateLocale })}
            </p>
            <h1 className="text-3xl font-bold mt-1">
              {greeting()}, {user?.first_name || 'User'}
            </h1>
            <p className="text-primary-foreground/80 mt-2">{t('home.welcomeMessage')}</p>
          </div>
          <div className="hidden md:block">
            <MaterialIcon name="waving_hand" className="text-6xl opacity-80" />
          </div>
        </div>
      </div>

      {/* クイックアクション */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* 出退勤状態 */}
        <div className="bg-card border border-border rounded-xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="size-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
              <MaterialIcon name="schedule" className="text-blue-500" />
            </div>
            <h2 className="font-semibold">{t('home.todayStatus')}</h2>
          </div>
          {attendanceStatus?.clock_in_time ? (
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground text-sm">{t('attendance.clockIn')}</span>
                <span className="font-mono font-semibold">
                  {new Date(attendanceStatus.clock_in_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
              {attendanceStatus.clock_out_time ? (
                <div className="flex justify-between items-center">
                  <span className="text-muted-foreground text-sm">{t('attendance.clockOut')}</span>
                  <span className="font-mono font-semibold">
                    {new Date(attendanceStatus.clock_out_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
              ) : (
                <Link
                  to="/attendance"
                  className="block w-full mt-3 px-4 py-2 bg-destructive text-destructive-foreground text-center rounded-lg hover:brightness-110 transition-all font-medium"
                >
                  {t('attendance.clockOut')}
                </Link>
              )}
            </div>
          ) : (
            <div>
              <p className="text-muted-foreground text-sm mb-3">{t('home.notClockedIn')}</p>
              <Link
                to="/attendance"
                className="block w-full px-4 py-2 bg-primary text-primary-foreground text-center rounded-lg hover:brightness-110 transition-all font-medium"
              >
                {t('attendance.clockIn')}
              </Link>
            </div>
          )}
        </div>

        {/* 未読通知 */}
        <div className="bg-card border border-border rounded-xl p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="size-10 rounded-lg bg-amber-500/20 flex items-center justify-center">
              <MaterialIcon name="notifications" className="text-amber-500" />
            </div>
            <h2 className="font-semibold">{t('home.notifications')}</h2>
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-3xl font-bold">{unreadCount}</p>
              <p className="text-muted-foreground text-sm">{t('home.unreadNotifications')}</p>
            </div>
            <Link
              to="/notifications"
              className="px-4 py-2 bg-primary/10 text-primary rounded-lg hover:bg-primary/20 transition-colors text-sm font-medium"
            >
              {t('home.viewAll')}
            </Link>
          </div>
        </div>

        {/* 承認待ち（管理者のみ） */}
        {(user?.role === 'admin' || user?.role === 'manager') && (
          <div className="bg-card border border-border rounded-xl p-6">
            <div className="flex items-center gap-3 mb-4">
              <div className="size-10 rounded-lg bg-purple-500/20 flex items-center justify-center">
                <MaterialIcon name="pending_actions" className="text-purple-500" />
              </div>
              <h2 className="font-semibold">{t('home.pendingApprovals')}</h2>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-3xl font-bold">{pendingCount}</p>
                <p className="text-muted-foreground text-sm">{t('home.awaitingReview')}</p>
              </div>
              <Link
                to="/leaves"
                className="px-4 py-2 bg-primary/10 text-primary rounded-lg hover:bg-primary/20 transition-colors text-sm font-medium"
              >
                {t('home.review')}
              </Link>
            </div>
          </div>
        )}
      </div>

      {/* アプリ一覧 */}
      <div>
        <h2 className="text-xl font-bold mb-4">{t('home.apps')}</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {availableApps.map((app) => {
          const isClickable = app.enabled && !app.comingSoon;
          if (isClickable) {
            return (
              <Link
                key={app.id}
                to={app.basePath as '/'}
                className="flex flex-col items-center gap-3 p-6 bg-card border border-border rounded-xl transition-all hover:border-primary hover:shadow-lg hover:shadow-primary/10 cursor-pointer"
              >
                <div className={`size-14 rounded-xl flex items-center justify-center text-white ${app.color}`}>
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
              className="flex flex-col items-center gap-3 p-6 bg-card border border-border rounded-xl opacity-50 cursor-not-allowed"
            >
              <div className={`size-14 rounded-xl flex items-center justify-center text-white ${app.color}`}>
                <MaterialIcon name={app.icon} className="text-2xl" />
              </div>
              <div className="text-center">
                <p className="font-medium text-sm">{t(app.nameKey)}</p>
                <span className="inline-block mt-1 px-2 py-0.5 bg-amber-500/20 text-amber-500 text-[10px] font-bold rounded-full">
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
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="font-bold mb-4 flex items-center gap-2">
            <MaterialIcon name="star" className="text-amber-500" />
            {t('home.quickAccess')}
          </h2>
          <div className="grid grid-cols-2 gap-3">
            <Link
              to="/leaves"
              className="flex items-center gap-3 p-3 rounded-lg hover:bg-primary/10 transition-colors"
            >
              <MaterialIcon name="event_available" className="text-green-500" />
              <span className="text-sm font-medium">{t('nav.leaves')}</span>
            </Link>
            <Link
              to="/overtime"
              className="flex items-center gap-3 p-3 rounded-lg hover:bg-primary/10 transition-colors"
            >
              <MaterialIcon name="more_time" className="text-orange-500" />
              <span className="text-sm font-medium">{t('nav.overtime')}</span>
            </Link>
            <Link
              to="/projects"
              className="flex items-center gap-3 p-3 rounded-lg hover:bg-primary/10 transition-colors"
            >
              <MaterialIcon name="folder_open" className="text-blue-500" />
              <span className="text-sm font-medium">{t('nav.projects')}</span>
            </Link>
            <Link
              to="/holidays"
              className="flex items-center gap-3 p-3 rounded-lg hover:bg-primary/10 transition-colors"
            >
              <MaterialIcon name="celebration" className="text-pink-500" />
              <span className="text-sm font-medium">{t('nav.holidays')}</span>
            </Link>
          </div>
        </div>

        {/* 最近の通知 */}
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="font-bold mb-4 flex items-center gap-2">
            <MaterialIcon name="notifications_active" className="text-blue-500" />
            {t('home.recentNotifications')}
          </h2>
          <div className="space-y-3">
            {(notifications as { data?: { id: string; title: string; created_at: string; is_read: boolean }[] } | undefined)?.data?.slice(0, 3).map((n) => (
              <div
                key={n.id}
                className={`p-3 rounded-lg border ${n.is_read ? 'border-border bg-muted/30' : 'border-primary/30 bg-primary/5'}`}
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
