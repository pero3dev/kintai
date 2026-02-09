import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';

// Material Symbols icon component
function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function DashboardPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();

  const { data: todayStatus } = useQuery({
    queryKey: ['attendance', 'today'],
    queryFn: () => api.attendance.getToday(),
  });

  const { data: stats } = useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: () => api.dashboard.getStats(),
    enabled: user?.role === 'admin' || user?.role === 'manager',
  });

  const statCards = [
    {
      title: t('dashboard.todayPresent'),
      value: stats?.today_present_count ?? '-',
      subtitle: `+3 ${t('dashboard.thisMonth')}`,
      icon: 'group',
      color: 'text-primary',
      bgColor: 'bg-primary/10',
      trend: 'up',
    },
    {
      title: t('dashboard.todayAbsent'),
      value: stats?.today_absent_count ?? '-',
      subtitle: stats?.today_absent_count && stats.today_absent_count > 0 ? `${stats.today_absent_count} ${t('dashboard.absent')}` : t('dashboard.allPresent'),
      icon: 'error',
      color: 'text-destructive',
      bgColor: 'bg-destructive/10',
      trend: stats?.today_absent_count && stats.today_absent_count > 0 ? 'warning' : 'ok',
    },
    {
      title: t('dashboard.pendingLeaves'),
      value: stats?.pending_leaves ?? '-',
      subtitle: t('dashboard.awaitingApproval'),
      icon: 'event_available',
      color: 'text-primary',
      bgColor: 'bg-primary/10',
      trend: 'neutral',
    },
    {
      title: t('dashboard.monthlyOvertime'),
      value: stats?.monthly_overtime ? `${Math.round(stats.monthly_overtime / 60)}h` : '-',
      subtitle: t('dashboard.monthlyTotal'),
      icon: 'schedule',
      color: 'text-emerald-500',
      bgColor: 'bg-emerald-500/10',
      trend: 'neutral',
    },
  ];

  return (
    <div className="space-y-8 animate-fade-in">
      {/* サマリー統計カード */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((card) => (
          <div
            key={card.title}
            className="glass-card p-6 rounded-2xl flex items-start justify-between stat-shimmer"
          >
            <div>
              <p className="text-muted-foreground text-sm font-medium">{card.title}</p>
              <h3 className="text-3xl font-bold mt-1 gradient-text">{card.value}</h3>
              <p className={`text-xs font-semibold flex items-center mt-2 ${card.trend === 'up' ? 'text-emerald-500' :
                  card.trend === 'warning' ? 'text-destructive' :
                    'text-muted-foreground'
                }`}>
                {card.trend === 'up' && <MaterialIcon name="trending_up" className="text-sm mr-1" />}
                {card.trend === 'warning' && <MaterialIcon name="warning" className="text-sm mr-1" />}
                {card.subtitle}
              </p>
            </div>
            <div className={`p-3 rounded-lg ${card.bgColor}`}>
              <MaterialIcon name={card.icon} className={card.color} />
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* 左カラム: 勤怠状況と部署統計 */}
        <div className="lg:col-span-2 space-y-8">
          {/* 今日の勤怠状況 */}
          <section className="glass-card rounded-2xl">
            <div className="p-6 border-b border-white/5 flex items-center justify-between">
              <h2 className="font-bold text-lg flex items-center gap-2">
                <MaterialIcon name="analytics" className="text-indigo-400" />
                {t('attendance.todayStatus')}
              </h2>
              <div className="flex gap-2">
                <span className="px-2 py-1 bg-emerald-500/10 text-emerald-400 text-[10px] font-bold rounded-full uppercase">
                  {t('dashboard.present')}: {stats?.today_present_count ?? 0}
                </span>
                <span className="px-2 py-1 bg-amber-500/10 text-amber-400 text-[10px] font-bold rounded-full uppercase">
                  {t('dashboard.leave')}: {stats?.pending_leaves ?? 0}
                </span>
                <span className="px-2 py-1 bg-red-500/10 text-red-400 text-[10px] font-bold rounded-full uppercase">
                  {t('dashboard.absent')}: {stats?.today_absent_count ?? 0}
                </span>
              </div>
            </div>
            <div className="p-6">
              <div className="relative h-48 w-full glass-subtle rounded-xl overflow-hidden flex items-center justify-center">
                <div className="absolute inset-0 opacity-20 bg-[radial-gradient(circle_at_50%_50%,#39E079_0%,transparent_70%)]"></div>
                <div className="relative z-10 text-center">
                  <div className="flex items-end justify-center gap-1 mb-4 h-24">
                    <div className="w-8 bg-emerald-500/40 h-16 rounded-t"></div>
                    <div className="w-8 bg-emerald-500/60 h-20 rounded-t"></div>
                    <div className="w-8 bg-primary h-24 rounded-t"></div>
                    <div className="w-8 bg-primary/80 h-20 rounded-t"></div>
                    <div className="w-8 bg-amber-500 h-12 rounded-t"></div>
                    <div className="w-8 bg-primary/60 h-18 rounded-t"></div>
                    <div className="w-8 bg-destructive h-8 rounded-t"></div>
                  </div>
                  <p className="text-xs text-muted-foreground">{t('dashboard.weeklyTrend')}</p>
                </div>
              </div>
            </div>
          </section>

          {/* 部署別統計 */}
          {stats?.department_stats && stats.department_stats.length > 0 && (
            <section className="glass-card rounded-2xl overflow-hidden">
              <div className="p-6 border-b border-white/5 flex items-center justify-between">
                <h2 className="font-bold text-lg flex items-center gap-2">
                  <MaterialIcon name="corporate_fare" className="text-indigo-400" />
                  {t('dashboard.departmentStatus')}
                </h2>
                <button className="text-indigo-400 text-sm font-semibold hover:text-indigo-300 transition-colors">
                  {t('dashboard.showAll')}
                </button>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead>
                    <tr className="glass-subtle text-muted-foreground font-medium">
                      <th className="px-6 py-3">{t('dashboard.departmentName')}</th>
                      <th className="px-6 py-3">{t('dashboard.employeeCount')}</th>
                      <th className="px-6 py-3">{t('dashboard.presentToday')}</th>
                      <th className="px-6 py-3 text-right">{t('dashboard.attendanceRate')}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {stats.department_stats.map((dept: { department_name: string; total_employees: number; present_today: number; attendance_rate: number }) => (
                      <tr key={dept.department_name} className="hover:bg-white/5 transition-colors">
                        <td className="px-6 py-4 font-semibold">{dept.department_name}</td>
                        <td className="px-6 py-4">{dept.total_employees}</td>
                        <td className="px-6 py-4">{dept.present_today}</td>
                        <td className="px-6 py-4 text-right">
                          <span className={`px-2 py-0.5 rounded-full text-xs font-bold ${dept.attendance_rate >= 0.9
                              ? 'bg-emerald-500/20 text-emerald-400'
                              : dept.attendance_rate >= 0.7
                                ? 'bg-amber-500/20 text-amber-400'
                                : 'bg-destructive/20 text-destructive'
                            }`}>
                            {(dept.attendance_rate * 100).toFixed(1)}%
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          )}
        </div>

        {/* 右カラム: アラートとステータス */}
        <div className="space-y-8">
          {/* 個人の勤怠ステータス */}
          <section className="glass-card rounded-2xl">
            <div className="p-6 border-b border-white/5">
              <h2 className="font-bold text-lg flex items-center gap-2">
                <MaterialIcon name="person" className="text-indigo-400" />
                {t('attendance.todayStatus')}
              </h2>
            </div>
            <div className="p-6 space-y-4">
              <div className={`flex gap-4 p-4 rounded-xl ${todayStatus?.clock_in
                  ? todayStatus?.clock_out
                    ? 'glass-subtle'
                    : 'bg-emerald-500/5 border border-emerald-500/20'
                  : 'bg-amber-500/5 border border-amber-500/20'
                }`}>
                <div className={`size-10 rounded-full flex items-center justify-center shrink-0 ${todayStatus?.clock_in
                    ? todayStatus?.clock_out
                      ? 'bg-muted text-muted-foreground'
                      : 'bg-emerald-500 text-white'
                    : 'bg-amber-500 text-white'
                  }`}>
                  <MaterialIcon name={todayStatus?.clock_in ? 'check_circle' : 'schedule'} className="text-xl" />
                </div>
                <div>
                  <h4 className={`font-bold text-sm ${todayStatus?.clock_in
                      ? todayStatus?.clock_out
                        ? 'text-muted-foreground'
                        : 'text-emerald-400'
                      : 'text-amber-400'
                    }`}>
                    {todayStatus?.clock_in
                      ? todayStatus?.clock_out
                        ? t('attendance.clockedOut')
                        : t('attendance.clockedIn')
                      : t('attendance.notClockedIn')}
                  </h4>
                  <div className="text-xs text-muted-foreground mt-1 space-y-1">
                    {todayStatus?.clock_in && (
                      <p>{t('attendance.clockIn')}: {new Date(todayStatus.clock_in).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })}</p>
                    )}
                    {todayStatus?.clock_out && (
                      <p>{t('attendance.clockOut')}: {new Date(todayStatus.clock_out).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })}</p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* 月間目標 */}
          <section className="glass-card rounded-2xl p-6 relative overflow-hidden">
            <div className="absolute inset-0 gradient-primary opacity-20 pointer-events-none" />
            <div className="relative z-10">
            <h2 className="font-bold text-lg flex items-center gap-2 mb-4">
              <MaterialIcon name="trending_up" />
              {t('dashboard.monthlyGoal')}
            </h2>
            <div className="space-y-4">
              <div className="flex justify-between text-sm font-bold">
                <span>{t('dashboard.monthlyGoal')}</span>
                <span>{stats ? `${Math.round((stats.today_present_count / (stats.today_present_count + stats.today_absent_count)) * 100) || 0}%` : '-'} {t('dashboard.achieved')}</span>
              </div>
              <div className="w-full bg-white/10 h-3 rounded-full overflow-hidden">
                <div
                  className="gradient-primary h-full rounded-full transition-all duration-500"
                  style={{ width: stats ? `${Math.round((stats.today_present_count / (stats.today_present_count + stats.today_absent_count)) * 100) || 0}%` : '0%' }}
                ></div>
              </div>
              <div className="pt-4 border-t border-white/10">
                <div className="flex justify-between items-end">
                  <div>
                    <p className="text-[10px] uppercase font-black text-muted-foreground">{t('dashboard.goalDays')}</p>
                    <p className="text-2xl font-black gradient-text">22 {t('dashboard.days')}</p>
                  </div>
                  <MaterialIcon name="calendar_month" className="text-4xl opacity-20" />
                </div>
              </div>
            </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
