import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { api } from '@/api/client';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';
import { Clock, LogIn, LogOut } from 'lucide-react';
import { Pagination } from '@/components/ui/Pagination';

export function AttendancePage() {
  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const [note, setNote] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const dateLocale = i18n.language === 'ja' ? ja : enUS;
  const [currentTime, setCurrentTime] = useState(new Date());

  // 毎秒時刻を更新
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  const { data: todayStatus } = useQuery({
    queryKey: ['attendance', 'today'],
    queryFn: () => api.attendance.getToday(),
    refetchInterval: 30000,
  });

  const { data: attendanceList } = useQuery({
    queryKey: ['attendance', 'list', page, pageSize],
    queryFn: () => api.attendance.getList({ page, page_size: pageSize }),
  });

  const { data: summary } = useQuery({
    queryKey: ['attendance', 'summary'],
    queryFn: () => api.attendance.getSummary(),
  });

  const clockInMutation = useMutation({
    mutationFn: () => api.attendance.clockIn({ note }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      setNote('');
    },
  });

  const clockOutMutation = useMutation({
    mutationFn: () => api.attendance.clockOut({ note }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      setNote('');
    },
  });

  const isClockedIn = todayStatus?.clock_in && !todayStatus?.clock_out;
  const isClockedOut = todayStatus?.clock_in && todayStatus?.clock_out;

  return (
    <div className="space-y-6 animate-fade-in">
      <h1 className="text-2xl font-bold gradient-text">{t('nav.attendance')}</h1>

      {/* 打刻セクション */}
      <div className="glass-card rounded-2xl p-6">
        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <Clock className="h-5 w-5" />
          {t('attendance.todayStatus')}
        </h2>

        <div className="text-center py-6">
          <p className="text-4xl font-mono font-bold gradient-text mb-2">
            {format(currentTime, 'HH:mm:ss')}
          </p>
          <p className="text-muted-foreground">
            {format(currentTime, i18n.language === 'ja' ? 'yyyy年MM月dd日 (EEEE)' : 'MMMM d, yyyy (EEEE)', { locale: dateLocale })}
          </p>
        </div>

        {/* 打刻ボタン */}
        <div className="flex flex-col items-center gap-4">
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder={t('attendance.note')}
            className="w-full max-w-md px-3 py-2.5 glass-input rounded-xl"
          />
          <div className="flex flex-col sm:flex-row gap-3 w-full sm:w-auto sm:gap-4">
            <button
              onClick={() => clockInMutation.mutate()}
              disabled={!!isClockedIn || !!isClockedOut || clockInMutation.isPending}
              className="flex items-center justify-center gap-2 px-6 sm:px-8 py-3 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-xl hover:bg-emerald-500/30 hover:shadow-glow-sm disabled:opacity-50 disabled:cursor-not-allowed transition-all text-base sm:text-lg"
            >
              <LogIn className="h-5 w-5" />
              {t('attendance.clockIn')}
            </button>
            <button
              onClick={() => clockOutMutation.mutate()}
              disabled={!isClockedIn || clockOutMutation.isPending}
              className="flex items-center justify-center gap-2 px-6 sm:px-8 py-3 bg-red-500/20 text-red-400 border border-red-500/30 rounded-xl hover:bg-red-500/30 hover:shadow-glow-sm disabled:opacity-50 disabled:cursor-not-allowed transition-all text-base sm:text-lg"
            >
              <LogOut className="h-5 w-5" />
              {t('attendance.clockOut')}
            </button>
          </div>

          {todayStatus?.clock_in && (
            <div className="text-sm text-muted-foreground">
              {t('attendance.clockIn')}: {new Date(todayStatus.clock_in).toLocaleTimeString('ja-JP')}
              {todayStatus.clock_out && (
                <> | {t('attendance.clockOut')}: {new Date(todayStatus.clock_out).toLocaleTimeString('ja-JP')}</>
              )}
            </div>
          )}
        </div>
      </div>

      {/* サマリー */}
      {summary && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="glass-card rounded-2xl p-4 stat-shimmer">
            <p className="text-sm text-muted-foreground">{t('attendance.totalWorkDays')}</p>
            <p className="text-2xl font-bold gradient-text">{summary.total_work_days}{t('dashboard.days')}</p>
          </div>
          <div className="glass-card rounded-2xl p-4 stat-shimmer">
            <p className="text-sm text-muted-foreground">{t('attendance.totalWorkHours')}</p>
            <p className="text-2xl font-bold gradient-text">{Math.round(summary.total_work_minutes / 60)}h</p>
          </div>
          <div className="glass-card rounded-2xl p-4 stat-shimmer">
            <p className="text-sm text-muted-foreground">{t('attendance.totalOvertime')}</p>
            <p className="text-2xl font-bold gradient-text">{Math.round(summary.total_overtime_minutes / 60)}h</p>
          </div>
          <div className="glass-card rounded-2xl p-4 stat-shimmer">
            <p className="text-sm text-muted-foreground">{t('attendance.averageWorkHours')}</p>
            <p className="text-2xl font-bold gradient-text">{(summary.average_work_minutes / 60).toFixed(1)}h</p>
          </div>
        </div>
      )}

      {/* 勤怠履歴 */}
      <div className="glass-card rounded-2xl p-4 sm:p-6">
        <h2 className="text-lg font-semibold mb-4">{t('attendance.history')}</h2>

        {/* モバイルカードビュー */}
        <div className="space-y-3 md:hidden">
          {attendanceList?.data?.map((record: Record<string, unknown>) => (
            <div key={record.id as string} className="glass-subtle rounded-xl p-4 space-y-2">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-sm">
                  {format(new Date(record.date as string), 'MM/dd (EEE)', { locale: dateLocale })}
                </span>
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${record.status === 'present' ? 'bg-emerald-500/20 text-emerald-400' :
                    record.status === 'absent' ? 'bg-red-500/20 text-red-400' :
                      record.status === 'leave' ? 'bg-indigo-500/20 text-indigo-400' :
                        'bg-white/10 text-muted-foreground'
                  }`}>
                  {record.status as string}
                </span>
              </div>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <p className="text-[10px] text-muted-foreground uppercase">{t('attendance.clockIn')}</p>
                  <p className="font-medium">{record.clock_in ? new Date(record.clock_in as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}</p>
                </div>
                <div>
                  <p className="text-[10px] text-muted-foreground uppercase">{t('attendance.clockOut')}</p>
                  <p className="font-medium">{record.clock_out ? new Date(record.clock_out as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}</p>
                </div>
                <div>
                  <p className="text-[10px] text-muted-foreground uppercase">{t('attendance.workTime')}</p>
                  <p className="font-medium">{record.work_minutes ? `${Math.floor(record.work_minutes as number / 60)}h ${(record.work_minutes as number) % 60}m` : '-'}</p>
                </div>
                <div>
                  <p className="text-[10px] text-muted-foreground uppercase">{t('attendance.overtime')}</p>
                  <p className="font-medium">{(record.overtime_minutes as number) > 0 ? `${Math.floor(record.overtime_minutes as number / 60)}h ${(record.overtime_minutes as number) % 60}m` : '-'}</p>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* デスクトップテーブルビュー */}
        <div className="hidden md:block overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5">
                <th className="text-left py-2 px-4">{t('common.date')}</th>
                <th className="text-left py-2 px-4">{t('attendance.clockIn')}</th>
                <th className="text-left py-2 px-4">{t('attendance.clockOut')}</th>
                <th className="text-right py-2 px-4">{t('attendance.workTime')}</th>
                <th className="text-right py-2 px-4">{t('attendance.overtime')}</th>
                <th className="text-left py-2 px-4">{t('common.status')}</th>
              </tr>
            </thead>
            <tbody>
              {attendanceList?.data?.map((record: Record<string, unknown>) => (
                <tr key={record.id as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                  <td className="py-2 px-4">
                    {format(new Date(record.date as string), 'MM/dd (EEE)', { locale: dateLocale })}
                  </td>
                  <td className="py-2 px-4">
                    {record.clock_in ? new Date(record.clock_in as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td className="py-2 px-4">
                    {record.clock_out ? new Date(record.clock_out as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td className="text-right py-2 px-4">
                    {record.work_minutes ? `${Math.floor(record.work_minutes as number / 60)}h ${(record.work_minutes as number) % 60}m` : '-'}
                  </td>
                  <td className="text-right py-2 px-4">
                    {(record.overtime_minutes as number) > 0 ? `${Math.floor(record.overtime_minutes as number / 60)}h ${(record.overtime_minutes as number) % 60}m` : '-'}
                  </td>
                  <td className="py-2 px-4">
                    <span className={`px-2 py-1 rounded-full text-xs ${record.status === 'present' ? 'bg-emerald-500/20 text-emerald-400' :
                        record.status === 'absent' ? 'bg-red-500/20 text-red-400' :
                          record.status === 'leave' ? 'bg-indigo-500/20 text-indigo-400' :
                            'bg-white/10 text-muted-foreground'
                      }`}>
                      {record.status as string}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {attendanceList?.total_pages > 0 && (
          <Pagination
            currentPage={page}
            totalPages={attendanceList.total_pages}
            totalItems={attendanceList.total}
            pageSize={pageSize}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
          />
        )}
      </div>
    </div>
  );
}
