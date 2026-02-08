import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';
import { FileEdit, Plus } from 'lucide-react';
import { Pagination } from '@/components/ui/Pagination';

const createCorrectionSchema = (t: (key: string) => string) => z.object({
  date: z.string().min(1, t('corrections.validation.dateRequired')),
  corrected_clock_in: z.string().optional(),
  corrected_clock_out: z.string().optional(),
  reason: z.string().min(1, t('corrections.validation.reasonRequired')),
});

type CorrectionForm = z.infer<ReturnType<typeof createCorrectionSchema>>;

export function CorrectionsPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [pendingPage, setPendingPage] = useState(1);
  const isAdmin = user?.role === 'admin' || user?.role === 'manager';

  const { data: myCorrections } = useQuery({
    queryKey: ['corrections', 'my', page, pageSize],
    queryFn: () => api.corrections.getList({ page, page_size: pageSize }),
  });

  const { data: pendingCorrections } = useQuery({
    queryKey: ['corrections', 'pending', pendingPage],
    queryFn: () => api.corrections.getPending({ page: pendingPage, page_size: 10 }),
    enabled: isAdmin,
  });

  const correctionSchema = createCorrectionSchema(t);
  const { register, handleSubmit, reset, formState: { errors } } = useForm<CorrectionForm>({
    resolver: zodResolver(correctionSchema),
  });

  const createMutation = useMutation({
    mutationFn: (data: CorrectionForm) => api.corrections.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['corrections'] });
      setShowForm(false);
      reset();
    },
  });

  const approveMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { status: string; rejected_reason?: string } }) =>
      api.corrections.approve(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['corrections'] });
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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <FileEdit className="h-6 w-6" />
          {t('corrections.title')}
        </h1>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          {t('corrections.newRequest')}
        </button>
      </div>

      {/* 申請フォーム */}
      {showForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('corrections.newRequest')}</h2>
          <form onSubmit={handleSubmit((data) => createMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('corrections.date')}</label>
                <input type="date" {...register('date')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
                {errors.date && <p className="text-sm text-destructive mt-1">{errors.date.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('corrections.reason')}</label>
                <input type="text" {...register('reason')} placeholder={t('common.forgotClockPlaceholder')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
                {errors.reason && <p className="text-sm text-destructive mt-1">{errors.reason.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('corrections.correctedClockIn')}</label>
                <input type="time" {...register('corrected_clock_in')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('corrections.correctedClockOut')}</label>
                <input type="time" {...register('corrected_clock_out')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
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
      {isAdmin && pendingCorrections?.data?.length > 0 && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('corrections.pending')}</h2>
          <div className="space-y-3">
            {pendingCorrections.data.map((c: Record<string, unknown>) => (
              <div key={c.id as string} className="flex items-center justify-between p-4 border border-border rounded-md">
                <div>
                  <p className="font-medium">{(c.user as Record<string, string>)?.last_name} {(c.user as Record<string, string>)?.first_name}</p>
                  <p className="text-sm text-muted-foreground">
                    {(c.date as string)?.slice(0, 10)} | {c.reason as string}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {t('common.correctedTime')}: {c.corrected_clock_in ? new Date(c.corrected_clock_in as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'} ~{' '}
                    {c.corrected_clock_out ? new Date(c.corrected_clock_out as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => approveMutation.mutate({ id: c.id as string, data: { status: 'approved' } })}
                    className="px-3 py-1 bg-green-600/80 text-white rounded text-sm hover:bg-green-600"
                  >
                    {t('common.approve')}
                  </button>
                  <button
                    onClick={() => approveMutation.mutate({ id: c.id as string, data: { status: 'rejected' } })}
                    className="px-3 py-1 bg-red-600/80 text-white rounded text-sm hover:bg-red-600"
                  >
                    {t('common.reject')}
                  </button>
                </div>
              </div>
            ))}
          </div>
          {pendingCorrections?.total_pages > 1 && (
            <Pagination
              currentPage={pendingPage}
              totalPages={pendingCorrections.total_pages}
              totalItems={pendingCorrections.total}
              pageSize={10}
              onPageChange={setPendingPage}
            />
          )}
        </div>
      )}

      {/* 自分の申請一覧 */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h2 className="text-lg font-semibold mb-4">{t('corrections.history')}</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 px-4">{t('corrections.date')}</th>
                <th className="text-left py-2 px-4">{t('corrections.correctedClockIn')}</th>
                <th className="text-left py-2 px-4">{t('corrections.correctedClockOut')}</th>
                <th className="text-left py-2 px-4">{t('corrections.reason')}</th>
                <th className="text-left py-2 px-4">{t('common.status')}</th>
              </tr>
            </thead>
            <tbody>
              {myCorrections?.data?.map((c: Record<string, unknown>) => (
                <tr key={c.id as string} className="border-b border-border/50">
                  <td className="py-2 px-4">{(c.date as string)?.slice(0, 10)}</td>
                  <td className="py-2 px-4">
                    {c.corrected_clock_in ? new Date(c.corrected_clock_in as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td className="py-2 px-4">
                    {c.corrected_clock_out ? new Date(c.corrected_clock_out as string).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }) : '-'}
                  </td>
                  <td className="py-2 px-4">{c.reason as string}</td>
                  <td className="py-2 px-4">{statusBadge(c.status as string)}</td>
                </tr>
              ))}
              {(!myCorrections?.data || myCorrections.data.length === 0) && (
                <tr><td colSpan={5} className="py-4 text-center text-muted-foreground">{t('common.noData')}</td></tr>
              )}
            </tbody>
          </table>
        </div>
        {myCorrections?.total_pages > 0 && (
          <Pagination
            currentPage={page}
            totalPages={myCorrections.total_pages}
            totalItems={myCorrections.total}
            pageSize={pageSize}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
          />
        )}
      </div>
    </div>
  );
}
