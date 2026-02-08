import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';
import { CalendarDays, Plus } from 'lucide-react';
import { Pagination } from '@/components/ui/Pagination';

const createLeaveSchema = (t: (key: string) => string) => z.object({
  leave_type: z.enum(['paid', 'sick', 'special', 'half']),
  start_date: z.string().min(1, t('leaves.validation.startDateRequired')),
  end_date: z.string().min(1, t('leaves.validation.endDateRequired')),
  reason: z.string().optional(),
});

type LeaveForm = z.infer<ReturnType<typeof createLeaveSchema>>;

export function LeavesPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [pendingPage, setPendingPage] = useState(1);

  const { data: myLeaves } = useQuery({
    queryKey: ['leaves', 'my', page, pageSize],
    queryFn: () => api.leaves.getList({ page, page_size: pageSize }),
  });

  const { data: pendingLeaves } = useQuery({
    queryKey: ['leaves', 'pending', pendingPage],
    queryFn: () => api.leaves.getPending({ page: pendingPage, page_size: 10 }),
    enabled: user?.role === 'admin' || user?.role === 'manager',
  });

  const leaveSchema = createLeaveSchema(t);
  const { register, handleSubmit, reset, formState: { errors } } = useForm<LeaveForm>({
    resolver: zodResolver(leaveSchema),
  });

  const createMutation = useMutation({
    mutationFn: (data: LeaveForm) => api.leaves.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leaves'] });
      setShowForm(false);
      reset();
    },
  });

  const approveMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { status: string; rejected_reason?: string } }) =>
      api.leaves.approve(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leaves'] });
    },
  });

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      pending: 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30',
      approved: 'bg-green-500/20 text-green-400 border border-green-500/30',
      rejected: 'bg-red-500/20 text-red-400 border border-red-500/30',
    };
    const labels: Record<string, string> = {
      pending: t('leaves.pending'),
      approved: t('leaves.approved'),
      rejected: t('leaves.rejected'),
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs ${styles[status] || ''}`}>
        {labels[status] || status}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <CalendarDays className="h-6 w-6" />
          {t('leaves.title')}
        </h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          {t('leaves.newRequest')}
        </button>
      </div>

      {/* 申請フォーム */}
      {showForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('leaves.newRequest')}</h2>
          <form onSubmit={handleSubmit((data) => createMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('leaves.type')}</label>
                <select {...register('leave_type')} className="w-full px-3 py-2 border border-input rounded-md bg-background">
                  <option value="paid">{t('leaves.types.paid')}</option>
                  <option value="sick">{t('leaves.types.sick')}</option>
                  <option value="special">{t('leaves.types.special')}</option>
                  <option value="half">{t('leaves.types.half')}</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('leaves.reason')}</label>
                <input type="text" {...register('reason')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('leaves.startDate')}</label>
                <input type="date" {...register('start_date')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
                {errors.start_date && <p className="text-sm text-destructive mt-1">{errors.start_date.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('leaves.endDate')}</label>
                <input type="date" {...register('end_date')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
                {errors.end_date && <p className="text-sm text-destructive mt-1">{errors.end_date.message}</p>}
              </div>
            </div>
            <div className="flex gap-2">
              <button type="submit" disabled={createMutation.isPending} className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50">
                {t('common.submit')}
              </button>
              <button type="button" onClick={() => setShowForm(false)} className="px-4 py-2 border border-input rounded-md hover:bg-accent">
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* 承認待ち一覧（管理者用） */}
      {(user?.role === 'admin' || user?.role === 'manager') && pendingLeaves?.data?.length > 0 && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('leaves.pending')}</h2>
          <div className="space-y-3">
            {pendingLeaves.data.map((leave: Record<string, unknown>) => (
              <div key={leave.id as string} className="flex items-center justify-between p-4 border border-border rounded-md">
                <div>
                  <p className="font-medium">{(leave.user as Record<string, string>)?.last_name} {(leave.user as Record<string, string>)?.first_name}</p>
                  <p className="text-sm text-muted-foreground">
                    {leave.start_date as string} ~ {leave.end_date as string} | {t(`leaves.types.${leave.leave_type as string}`)}
                  </p>
                  {leave.reason ? <p className="text-sm mt-1">{String(leave.reason)}</p> : null}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => approveMutation.mutate({ id: leave.id as string, data: { status: 'approved' } })}
                    className="px-3 py-1 bg-green-600/80 text-white rounded text-sm hover:bg-green-600"
                  >
                    {t('leaves.approve')}
                  </button>
                  <button
                    onClick={() => approveMutation.mutate({ id: leave.id as string, data: { status: 'rejected' } })}
                    className="px-3 py-1 bg-red-600/80 text-white rounded text-sm hover:bg-red-600"
                  >
                    {t('leaves.reject')}
                  </button>
                </div>
              </div>
            ))}
          </div>
          {pendingLeaves?.total_pages > 1 && (
            <Pagination
              currentPage={pendingPage}
              totalPages={pendingLeaves.total_pages}
              totalItems={pendingLeaves.total}
              pageSize={10}
              onPageChange={setPendingPage}
            />
          )}
        </div>
      )}

      {/* 自分の申請一覧 */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-4">{t('leaves.history')}</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 px-4">{t('leaves.type')}</th>
                <th className="text-left py-2 px-4">{t('leaves.startDate')}</th>
                <th className="text-left py-2 px-4">{t('leaves.endDate')}</th>
                <th className="text-left py-2 px-4">{t('leaves.reason')}</th>
                <th className="text-left py-2 px-4">{t('leaves.status')}</th>
              </tr>
            </thead>
            <tbody>
              {myLeaves?.data?.map((leave: Record<string, unknown>) => (
                <tr key={leave.id as string} className="border-b border-border/50">
                  <td className="py-2 px-4">{t(`leaves.types.${leave.leave_type as string}`)}</td>
                  <td className="py-2 px-4">{leave.start_date as string}</td>
                  <td className="py-2 px-4">{leave.end_date as string}</td>
                  <td className="py-2 px-4">{(leave.reason as string) || '-'}</td>
                  <td className="py-2 px-4">{statusBadge(leave.status as string)}</td>
                </tr>
              ))}
              {(!myLeaves?.data || myLeaves.data.length === 0) && (
                <tr><td colSpan={5} className="py-4 text-center text-muted-foreground">{t('common.noData')}</td></tr>
              )}
            </tbody>
          </table>
        </div>
        {myLeaves?.total_pages > 0 && (
          <Pagination
            currentPage={page}
            totalPages={myLeaves.total_pages}
            totalItems={myLeaves.total}
            pageSize={pageSize}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
          />
        )}
      </div>
    </div>
  );
}
