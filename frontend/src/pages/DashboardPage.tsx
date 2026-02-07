import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { Users, UserCheck, UserX, Clock, CalendarDays } from 'lucide-react';

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
      icon: UserCheck,
      color: 'text-green-600',
      bgColor: 'bg-green-50',
    },
    {
      title: t('dashboard.todayAbsent'),
      value: stats?.today_absent_count ?? '-',
      icon: UserX,
      color: 'text-red-600',
      bgColor: 'bg-red-50',
    },
    {
      title: t('dashboard.pendingLeaves'),
      value: stats?.pending_leaves ?? '-',
      icon: CalendarDays,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-50',
    },
    {
      title: t('dashboard.monthlyOvertime'),
      value: stats?.monthly_overtime ? `${Math.round(stats.monthly_overtime / 60)}h` : '-',
      icon: Clock,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
    },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">{t('dashboard.title')}</h1>

      {/* 今日の勤怠状況 */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-4">{t('attendance.todayStatus')}</h2>
        <div className="flex items-center gap-4">
          <div
            className={`px-4 py-2 rounded-full text-sm font-medium ${todayStatus?.clock_in
                ? todayStatus?.clock_out
                  ? 'bg-muted text-muted-foreground'
                  : 'bg-green-100 text-green-800'
                : 'bg-yellow-100 text-yellow-800'
              }`}
          >
            {todayStatus?.clock_in
              ? todayStatus?.clock_out
                ? t('attendance.clockedOut')
                : t('attendance.clockedIn')
              : t('attendance.notClockedIn')}
          </div>
          {todayStatus?.clock_in && (
            <span className="text-sm text-muted-foreground">
              {t('attendance.clockIn')}: {new Date(todayStatus.clock_in).toLocaleTimeString('ja-JP')}
            </span>
          )}
          {todayStatus?.clock_out && (
            <span className="text-sm text-muted-foreground">
              {t('attendance.clockOut')}: {new Date(todayStatus.clock_out).toLocaleTimeString('ja-JP')}
            </span>
          )}
        </div>
      </div>

      {/* 管理者用統計 */}
      {(user?.role === 'admin' || user?.role === 'manager') && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {statCards.map((card) => (
            <div key={card.title} className="bg-card border border-border rounded-lg p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">{card.title}</p>
                  <p className="text-2xl font-bold mt-1">{card.value}</p>
                </div>
                <div className={`p-3 rounded-full ${card.bgColor}`}>
                  <card.icon className={`h-6 w-6 ${card.color}`} />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 部署別統計 */}
      {stats?.department_stats && stats.department_stats.length > 0 && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Users className="h-5 w-5" />
            部署別出勤状況
          </h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-2 px-4">部署名</th>
                  <th className="text-right py-2 px-4">社員数</th>
                  <th className="text-right py-2 px-4">本日出勤</th>
                  <th className="text-right py-2 px-4">出勤率</th>
                </tr>
              </thead>
              <tbody>
                {stats.department_stats.map((dept: { department_name: string; total_employees: number; present_today: number; attendance_rate: number }) => (
                  <tr key={dept.department_name} className="border-b border-border/50">
                    <td className="py-2 px-4">{dept.department_name}</td>
                    <td className="text-right py-2 px-4">{dept.total_employees}</td>
                    <td className="text-right py-2 px-4">{dept.present_today}</td>
                    <td className="text-right py-2 px-4">
                      {(dept.attendance_rate * 100).toFixed(1)}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
