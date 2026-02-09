import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const RISK_COLORS: Record<string, string> = {
  high: 'bg-red-500/20 text-red-400',
  medium: 'bg-amber-500/20 text-amber-400',
  low: 'bg-green-500/20 text-green-400',
  none: 'bg-gray-500/20 text-gray-400',
};

export function HRAttendanceIntegrationPage() {
  const { t } = useTranslation();
  const [period, setPeriod] = useState('thisMonth');
  const [department, setDepartment] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['hr-attendance-integration', period, department],
    queryFn: () => api.hr.getAttendanceIntegration({ period, department }),
  });
  const { data: alertsData } = useQuery({
    queryKey: ['hr-attendance-alerts'],
    queryFn: () => api.hr.getAttendanceAlerts(),
  });
  const { data: trendData } = useQuery({
    queryKey: ['hr-attendance-trend', period],
    queryFn: () => api.hr.getAttendanceTrend({ period }),
  });

  const stats = data?.data || data || {};
  const alerts: Record<string, unknown>[] = alertsData?.data || alertsData || [];
  const employees: Record<string, unknown>[] = stats.employees || [];
  const trend: Record<string, unknown>[] = trendData?.data || trendData || [];

  const summaryStats = [
    { label: t('hr.attendanceIntegration.avgWorkHours'), value: stats.avg_work_hours ? `${Number(stats.avg_work_hours).toFixed(1)}h` : '-', icon: 'schedule', color: 'text-blue-400' },
    { label: t('hr.attendanceIntegration.totalOvertime'), value: stats.total_overtime ? `${Number(stats.total_overtime).toFixed(0)}h` : '-', icon: 'more_time', color: 'text-amber-400' },
    { label: t('hr.attendanceIntegration.lateRate'), value: stats.late_rate ? `${Number(stats.late_rate).toFixed(1)}%` : '-', icon: 'running_with_errors', color: 'text-red-400' },
    { label: t('hr.attendanceIntegration.paidLeaveUsage'), value: stats.leave_usage ? `${Number(stats.leave_usage).toFixed(1)}%` : '-', icon: 'event_available', color: 'text-green-400' },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.attendanceIntegration.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.attendanceIntegration.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <select value={period} onChange={(e) => setPeriod(e.target.value)} className="px-4 py-2.5 glass-input rounded-xl text-sm">
            {['thisMonth', 'lastMonth', 'last3Months', 'last6Months', 'thisYear'].map(p => (
              <option key={p} value={p}>{t(`hr.attendanceIntegration.periods.${p}`)}</option>
            ))}
          </select>
          <input value={department} onChange={(e) => setDepartment(e.target.value)}
            placeholder={t('hr.attendanceIntegration.department')}
            className="px-4 py-2.5 glass-input rounded-xl text-sm w-32" />
        </div>
      </div>

      {/* サマリー統計 */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {summaryStats.map((s, i) => (
          <div key={i} className="glass-card rounded-2xl p-4 text-center">
            <MaterialIcon name={s.icon} className={`text-2xl ${s.color} mb-1`} />
            <p className="text-2xl font-bold">{s.value}</p>
            <p className="text-xs text-muted-foreground">{s.label}</p>
          </div>
        ))}
      </div>

      {/* アラート */}
      {alerts.length > 0 && (
        <div className="glass-card rounded-2xl p-4 sm:p-5">
          <h3 className="font-semibold text-sm mb-3 flex items-center gap-2">
            <MaterialIcon name="warning" className="text-amber-400" />
            {t('hr.attendanceIntegration.alerts')} ({alerts.length})
          </h3>
          <div className="space-y-2">
            {alerts.slice(0, 5).map((alert, i) => (
              <div key={i} className="flex items-center gap-3 p-2 glass-subtle rounded-xl">
                <MaterialIcon name={
                  alert.type === 'overtime' ? 'more_time' :
                  alert.type === 'late' ? 'running_with_errors' : 'event_busy'
                } className={`text-sm ${
                  alert.severity === 'high' ? 'text-red-400' : 'text-amber-400'
                }`} />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{String(alert.employee_name || '')}</p>
                  <p className="text-xs text-muted-foreground">{String(alert.message || '')}</p>
                </div>
                <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${RISK_COLORS[String(alert.severity || 'medium')]}`}>
                  {t(`hr.attendanceIntegration.riskLevels.${String(alert.severity || 'medium')}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* トレンドグラフ（簡易バーチャート） */}
      {trend.length > 0 && (
        <div className="glass-card rounded-2xl p-4 sm:p-5">
          <h3 className="font-semibold text-sm mb-4">{t('hr.attendanceIntegration.trend')}</h3>
          <div className="flex items-end gap-2 h-32">
            {trend.map((t_item, i) => {
              const val = Number(t_item.overtime_hours || 0);
              const maxVal = Math.max(...trend.map(x => Number(x.overtime_hours || 0)), 1);
              return (
                <div key={i} className="flex-1 flex flex-col items-center gap-1">
                  <span className="text-[10px] text-muted-foreground">{val}h</span>
                  <div className="w-full rounded-t-lg bg-indigo-400/60 transition-all"
                    style={{ height: `${(val / maxVal) * 100}%`, minHeight: '4px' }} />
                  <span className="text-[10px] text-muted-foreground truncate w-full text-center">
                    {String(t_item.label || '')}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* 社員別テーブル */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : employees.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="schedule" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <>
          {/* モバイルカード */}
          <div className="md:hidden space-y-3">
            {employees.map((emp) => {
              const risk = String(emp.risk_level || 'none');
              return (
                <div key={String(emp.id)} className="glass-card rounded-2xl p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <div className="size-8 rounded-lg bg-indigo-500/20 flex items-center justify-center text-indigo-400 font-bold text-xs">
                        {String(emp.name || '?').substring(0, 1)}
                      </div>
                      <div>
                        <p className="font-semibold text-sm">{String(emp.name)}</p>
                        <p className="text-[10px] text-muted-foreground">{String(emp.department || '')}</p>
                      </div>
                    </div>
                    <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${RISK_COLORS[risk]}`}>
                      {t(`hr.attendanceIntegration.riskLevels.${risk}`)}
                    </span>
                  </div>
                  <div className="grid grid-cols-3 gap-2 text-center text-xs">
                    <div className="glass-subtle rounded-lg p-1.5">
                      <p className="text-muted-foreground">{t('hr.attendanceIntegration.overtimeHours')}</p>
                      <p className="font-bold">{String(emp.overtime_hours || 0)}h</p>
                    </div>
                    <div className="glass-subtle rounded-lg p-1.5">
                      <p className="text-muted-foreground">{t('hr.attendanceIntegration.lateRate')}</p>
                      <p className="font-bold">{String(emp.late_count || 0)}</p>
                    </div>
                    <div className="glass-subtle rounded-lg p-1.5">
                      <p className="text-muted-foreground">{t('hr.attendanceIntegration.paidLeaveUsage')}</p>
                      <p className="font-bold">{String(emp.leave_usage || 0)}%</p>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* デスクトップテーブル */}
          <div className="hidden md:block glass-card rounded-2xl overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/10 text-xs text-muted-foreground">
                  <th className="text-left px-5 py-3 font-medium">{t('hr.employees.employeeName')}</th>
                  <th className="text-left px-5 py-3 font-medium">{t('hr.attendanceIntegration.department')}</th>
                  <th className="text-right px-5 py-3 font-medium">{t('hr.attendanceIntegration.overtimeHours')}</th>
                  <th className="text-right px-5 py-3 font-medium">{t('hr.attendanceIntegration.lateRate')}</th>
                  <th className="text-right px-5 py-3 font-medium">{t('hr.attendanceIntegration.paidLeaveUsage')}</th>
                  <th className="text-right px-5 py-3 font-medium">{t('hr.attendanceIntegration.absentDays')}</th>
                  <th className="text-center px-5 py-3 font-medium">{t('hr.attendanceIntegration.riskLevel')}</th>
                </tr>
              </thead>
              <tbody>
                {employees.map((emp) => {
                  const risk = String(emp.risk_level || 'none');
                  return (
                    <tr key={String(emp.id)} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="px-5 py-3 text-sm font-medium">{String(emp.name)}</td>
                      <td className="px-5 py-3 text-sm text-muted-foreground">{String(emp.department || '')}</td>
                      <td className="px-5 py-3 text-sm text-right">{String(emp.overtime_hours || 0)}h</td>
                      <td className="px-5 py-3 text-sm text-right">{String(emp.late_count || 0)}</td>
                      <td className="px-5 py-3 text-sm text-right">{String(emp.leave_usage || 0)}%</td>
                      <td className="px-5 py-3 text-sm text-right">{String(emp.absent_days || 0)}</td>
                      <td className="px-5 py-3 text-center">
                        <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${RISK_COLORS[risk]}`}>
                          {t(`hr.attendanceIntegration.riskLevels.${risk}`)}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}
