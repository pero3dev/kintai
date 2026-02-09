import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';
import { Pagination } from '@/components/ui/Pagination';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function ExpenseApprovePage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const locale = i18n.language === 'ja' ? ja : enUS;
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [rejectingId, setRejectingId] = useState<string | null>(null);
  const [rejectReason, setRejectReason] = useState('');

  const isManager = user?.role === 'admin' || user?.role === 'manager';

  const { data, isLoading } = useQuery({
    queryKey: ['expenses', 'pending', page, pageSize],
    queryFn: () => api.expenses.getPending({ page, page_size: pageSize }),
    enabled: isManager,
  });

  const expenses = data?.data || data?.expenses || [];
  const totalItems = data?.total || data?.pagination?.total || 0;
  const totalPages = Math.ceil(totalItems / pageSize) || 1;

  const approveMutation = useMutation({
    mutationFn: (id: string) => api.expenses.approve(id, { status: 'approved' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense-stats'] });
    },
  });

  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      api.expenses.approve(id, { status: 'rejected', rejected_reason: reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense-stats'] });
      setRejectingId(null);
      setRejectReason('');
    },
  });

  const getCategoryIcon = (category: string): string => {
    const icons: Record<string, string> = {
      transportation: 'directions_car',
      meals: 'restaurant',
      accommodation: 'hotel',
      supplies: 'inventory_2',
      communication: 'phone',
      entertainment: 'celebration',
      other: 'more_horiz',
    };
    return icons[category] || 'receipt_long';
  };

  if (!isManager) {
    return (
      <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
        <MaterialIcon name="lock" className="text-6xl text-muted-foreground mb-4" />
        <p className="text-lg text-muted-foreground">{t('common.noPermission')}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex items-center gap-4">
        <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
          <MaterialIcon name="arrow_back" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold gradient-text">{t('expenses.approve.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">{t('expenses.approve.subtitle')}</p>
        </div>
      </div>

      {/* 承認待ちリスト */}
      <div className="glass-card rounded-2xl p-6">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5">
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.applicant')}</th>
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.title')}</th>
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.date')}</th>
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.category')}</th>
                <th className="text-right py-3 px-4 font-semibold">{t('expenses.fields.amount')}</th>
                <th className="text-center py-3 px-4 font-semibold">{t('common.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="text-center py-12 text-muted-foreground">
                    {t('common.loading')}
                  </td>
                </tr>
              ) : expenses.length === 0 ? (
                <tr>
                  <td colSpan={6} className="text-center py-12 text-muted-foreground">
                    <MaterialIcon name="check_circle" className="text-4xl mb-2 block text-emerald-400 opacity-50" />
                    <p>{t('expenses.approve.noPending')}</p>
                  </td>
                </tr>
              ) : (
                expenses.map((expense: Record<string, unknown>) => (
                  <tr key={expense.id as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                    <td className="py-3 px-4">
                      <p className="font-medium">{expense.user_name as string || '-'}</p>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <MaterialIcon name={getCategoryIcon(expense.category as string)} className="text-indigo-400 text-base" />
                        <span>{expense.title as string}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-muted-foreground">
                      {expense.expense_date ? format(new Date(expense.expense_date as string), 'PP', { locale }) : '-'}
                    </td>
                    <td className="py-3 px-4">
                      {expense.category ? t(`expenses.categories.${expense.category}`) : '-'}
                    </td>
                    <td className="py-3 px-4 text-right font-semibold">
                      ¥{Number(expense.amount).toLocaleString()}
                    </td>
                    <td className="py-3 px-4">
                      {rejectingId === expense.id ? (
                        <div className="flex items-center gap-2">
                          <input
                            value={rejectReason}
                            onChange={(e) => setRejectReason(e.target.value)}
                            placeholder={t('expenses.approve.rejectReason')}
                            className="px-2 py-1 glass-input rounded-lg text-xs w-32"
                          />
                          <button
                            onClick={() => rejectMutation.mutate({ id: expense.id as string, reason: rejectReason })}
                            disabled={rejectMutation.isPending}
                            className="px-2 py-1 bg-red-500/20 text-red-400 border border-red-500/30 rounded-lg text-xs hover:bg-red-500/30 transition-colors"
                          >
                            {t('common.confirm')}
                          </button>
                          <button
                            onClick={() => { setRejectingId(null); setRejectReason(''); }}
                            className="px-2 py-1 glass-subtle rounded-lg text-xs hover:bg-white/10 transition-colors"
                          >
                            {t('common.cancel')}
                          </button>
                        </div>
                      ) : (
                        <div className="flex items-center justify-center gap-2">
                          <button
                            onClick={() => approveMutation.mutate(expense.id as string)}
                            disabled={approveMutation.isPending}
                            className="px-3 py-1.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-lg text-xs font-semibold hover:bg-emerald-500/30 transition-colors"
                          >
                            {t('common.approve')}
                          </button>
                          <button
                            onClick={() => setRejectingId(expense.id as string)}
                            className="px-3 py-1.5 bg-red-500/20 text-red-400 border border-red-500/30 rounded-lg text-xs font-semibold hover:bg-red-500/30 transition-colors"
                          >
                            {t('common.reject')}
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="mt-4">
            <Pagination
              currentPage={page}
              totalPages={totalPages}
              totalItems={totalItems}
              pageSize={pageSize}
              onPageChange={setPage}
              onPageSizeChange={(size) => { setPageSize(size); setPage(1); }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
