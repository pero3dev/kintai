import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';
import { Clock, Plus, AlertTriangle } from 'lucide-react';
import { Pagination } from '@/components/ui/Pagination';

const createOvertimeSchema = (t: (key: string) => string) => z.object({
  date: z.string().min(1, t('overtime.validation.dateRequired')),
  planned_minutes: z.coerce.number().min(1, t('overtime.validation.minutesRequired')).max(480, t('overtime.validation.minutesMax')),
  reason: z.string().min(1, t('overtime.validation.reasonRequired')),
});

type OvertimeForm = z.infer<ReturnType<typeof createOvertimeSchema>>;

export function OvertimePage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [pendingPage, setPendingPage] = useState(1);
  const isAdmin = user?.role === 'admin' || user?.role === 'manager';

  const { data: myOvertimes } = useQuery({
    queryKey: ['overtime', 'my', page, pageSize],
    queryFn: () => api.overtime.getList({ page, page_size: pageSize }),
  });

  const { data: pendingOvertimes } = useQuery({
    queryKey: ['overtime', 'pending', pendingPage],
    queryFn: () => api.overtime.getPending({ page: pendingPage, page_size: 10 }),
    enabled: isAdmin,
  });

  const { data: alerts } = useQuery({
    queryKey: ['overtime', 'alerts'],
    queryFn: () => api.overtime.getAlerts(),
    enabled: isAdmin,
  });

  const overtimeSchema = createOvertimeSchema(t);
  const { register, handleSubmit, reset, formState: { errors } } = useForm<OvertimeForm>({
    resolver: zodResolver(overtimeSchema),
  });

  const createMutation = useMutation({
    mutationFn: (data: OvertimeForm) => api.overtime.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime'] });
      setShowForm(false);
      reset();
    },
  });

  const approveMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { status: string; rejected_reason?: string } }) =>
      api.overtime.approve(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime'] });
    },
  });

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      pending: 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30',
      approved: 'bg-green-500/20 text-green-400 border border-green-500/30',
      rejected: 'bg-red-500/20 text-red-400 border border-red-500/30',
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs ${styles[status] || ''}`}>
        {t(`common.${status}`) || status}
      </span>
    );
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Clock className="h-6 w-6" />
          {t('overtime.title')}
        </h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 gradient-primary text-white rounded-xl hover:shadow-glow-md transition-all"
        >
          <Plus className="h-4 w-4" />
          {t('overtime.newRequest')}
        </button>
      </div>

      {/* 36 Agreement alerts */}
      {isAdmin && alerts && alerts.length > 0 && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2 text-red-400">
            <AlertTriangle className="h-5 w-5" />
            {t('overtime.alerts')}
          </h2>
          <div className="space-y-3">
            {alerts.map((alert: Record<string, unknown>) => (
              <div key={alert.user_id as string} className="flex items-center justify-between p-3 bg-card rounded border border-border">
                <div>
                  <p className="font-medium">{alert.user_name as string}</p>
                  <p className="text-sm text-muted-foreground">
                    {t('overtime.monthly')}: {(alert.monthly_overtime_hours as number)?.toFixed(1)}h / {alert.monthly_limit_hours as number}h |
                    {t('overtime.yearly')}: {(alert.yearly_overtime_hours as number)?.toFixed(1)}h / {alert.yearly_limit_hours as number}h
                  </p>
                </div>
                <div className="flex gap-2">
                  {Boolean(alert.is_monthly_exceeded) && (
                    <span className="px-2 py-1 bg-red-500/20 text-red-400 rounded text-xs font-semibold">{t('common.monthlyExceeded')}</span>
                  )}
                  {Boolean(alert.is_yearly_exceeded) && (
                    <span className="px-2 py-1 bg-red-500/20 text-red-400 rounded text-xs font-semibold">{t('common.yearlyExceeded')}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 申請フォーム */}
      {showForm && (
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4">{t('overtime.newRequest')}</h2>
          <form onSubmit={handleSubmit((data) => createMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('overtime.date')}</label>
                <input type="date" {...register('date')} className="w-full px-3 py-2 glass-input rounded-xl" />
                {errors.date && <p className="text-sm text-red-400 mt-1">{errors.date.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('overtime.plannedMinutes')}</label>
                <input type="number" {...register('planned_minutes')} placeholder="60" className="w-full px-3 py-2 glass-input rounded-xl" />
                {errors.planned_minutes && <p className="text-sm text-red-400 mt-1">{errors.planned_minutes.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('overtime.reason')}</label>
                <input type="text" {...register('reason')} className="w-full px-3 py-2 glass-input rounded-xl" />
                {errors.reason && <p className="text-sm text-red-400 mt-1">{errors.reason.message}</p>}
              </div>
            </div>
            <div className="flex gap-2">
              <button type="submit" disabled={createMutation.isPending} className="px-4 py-2 gradient-primary text-white rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.submit')}
              </button>
              <button type="button" onClick={() => setShowForm(false)} className="px-4 py-2 glass-input rounded-xl hover:bg-white/10 transition-all">
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* 承認待ち一覧（管理者用） */}
      {isAdmin && pendingOvertimes?.data?.length > 0 && (
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4">{t('overtime.pending')}</h2>
          <div className="space-y-3">
            {pendingOvertimes.data.map((ot: Record<string, unknown>) => (
              <div key={ot.id as string} className="flex items-center justify-between p-4 glass-subtle rounded-xl">
                <div>
                  <p className="font-medium">{(ot.user as Record<string, string>)?.last_name} {(ot.user as Record<string, string>)?.first_name}</p>
                  <p className="text-sm text-muted-foreground">
                    {ot.date as string} | {ot.planned_minutes as number}{t('common.minutes')} | {ot.reason as string}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => approveMutation.mutate({ id: ot.id as string, data: { status: 'approved' } })}
                    className="px-3 py-1 bg-green-600/80 text-white rounded text-sm hover:bg-green-600"
                  >
                    {t('common.approve')}
                  </button>
                  <button
                    onClick={() => approveMutation.mutate({ id: ot.id as string, data: { status: 'rejected' } })}
                    className="px-3 py-1 bg-red-600/80 text-white rounded text-sm hover:bg-red-600"
                  >
                    {t('common.reject')}
                  </button>
                </div>
              </div>
            ))}
          </div>
          {pendingOvertimes?.total_pages > 1 && (
            <Pagination
              currentPage={pendingPage}
              totalPages={pendingOvertimes.total_pages}
              totalItems={pendingOvertimes.total}
              pageSize={10}
              onPageChange={setPendingPage}
            />
          )}
        </div>
      )}

      {/* 自分の申請一覧 */}
      <div className="glass-card rounded-2xl p-6">
        <h2 className="text-lg font-semibold mb-4">{t('overtime.history')}</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5">
                <th className="text-left py-2 px-4">{t('overtime.date')}</th>
                <th className="text-left py-2 px-4">{t('overtime.plannedMinutes')}</th>
                <th className="text-left py-2 px-4">{t('overtime.reason')}</th>
                <th className="text-left py-2 px-4">{t('common.status')}</th>
              </tr>
            </thead>
            <tbody>
              {myOvertimes?.data?.map((ot: Record<string, unknown>) => (
                <tr key={ot.id as string} className="border-b border-white/5">
                  <td className="py-2 px-4">{(ot.date as string)?.slice(0, 10)}</td>
                  <td className="py-2 px-4">{ot.planned_minutes as number}{t('common.minutes')}</td>
                  <td className="py-2 px-4">{ot.reason as string}</td>
                  <td className="py-2 px-4">{statusBadge(ot.status as string)}</td>
                </tr>
              ))}
              {(!myOvertimes?.data || myOvertimes.data.length === 0) && (
                <tr><td colSpan={4} className="py-4 text-center text-muted-foreground">{t('common.noData')}</td></tr>
              )}
            </tbody>
          </table>
        </div>
        {myOvertimes?.total_pages > 0 && (
          <Pagination
            currentPage={page}
            totalPages={myOvertimes.total_pages}
            totalItems={myOvertimes.total}
            pageSize={pageSize}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
          />
        )}
      </div>
    </div>
  );
}
