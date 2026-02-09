import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';
import { Pagination } from '@/components/ui/Pagination';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function ExpenseHistoryPage() {
  const { t, i18n } = useTranslation();
  const locale = i18n.language === 'ja' ? ja : enUS;
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [statusFilter, setStatusFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['expenses', 'history', page, pageSize, statusFilter, categoryFilter],
    queryFn: () => api.expenses.getList({
      page,
      page_size: pageSize,
      ...(statusFilter ? { status: statusFilter } : {}),
      ...(categoryFilter ? { category: categoryFilter } : {}),
    }),
  });

  const expenses = data?.data || data?.expenses || [];
  const totalItems = data?.total || data?.pagination?.total || 0;
  const totalPages = Math.ceil(totalItems / pageSize) || 1;

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      draft: 'bg-gray-500/20 text-gray-400 border border-gray-500/30',
      pending: 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30',
      approved: 'bg-green-500/20 text-green-400 border border-green-500/30',
      rejected: 'bg-red-500/20 text-red-400 border border-red-500/30',
      reimbursed: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
    };
    return styles[status] || styles.draft;
  };

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

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold gradient-text">{t('expenses.history.title')}</h1>
            <p className="text-muted-foreground text-sm mt-1">{t('expenses.history.subtitle')}</p>
          </div>
        </div>
        <Link
          to="/expenses/new"
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all duration-300"
        >
          <MaterialIcon name="add" className="text-lg" />
          {t('expenses.newExpense')}
        </Link>
      </div>

      {/* フィルター */}
      <div className="glass-card rounded-2xl p-4 flex flex-wrap gap-3">
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="px-3 py-2 glass-input rounded-xl text-sm"
        >
          <option value="">{t('expenses.filters.allStatuses')}</option>
          <option value="draft">{t('expenses.status.draft')}</option>
          <option value="pending">{t('expenses.status.pending')}</option>
          <option value="approved">{t('expenses.status.approved')}</option>
          <option value="rejected">{t('expenses.status.rejected')}</option>
          <option value="reimbursed">{t('expenses.status.reimbursed')}</option>
        </select>
        <select
          value={categoryFilter}
          onChange={(e) => { setCategoryFilter(e.target.value); setPage(1); }}
          className="px-3 py-2 glass-input rounded-xl text-sm"
        >
          <option value="">{t('expenses.filters.allCategories')}</option>
          <option value="transportation">{t('expenses.categories.transportation')}</option>
          <option value="meals">{t('expenses.categories.meals')}</option>
          <option value="accommodation">{t('expenses.categories.accommodation')}</option>
          <option value="supplies">{t('expenses.categories.supplies')}</option>
          <option value="communication">{t('expenses.categories.communication')}</option>
          <option value="entertainment">{t('expenses.categories.entertainment')}</option>
          <option value="other">{t('expenses.categories.other')}</option>
        </select>
      </div>

      {/* テーブル */}
      <div className="glass-card rounded-2xl p-6">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5">
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.title')}</th>
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.date')}</th>
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.category')}</th>
                <th className="text-right py-3 px-4 font-semibold">{t('expenses.fields.amount')}</th>
                <th className="text-center py-3 px-4 font-semibold">{t('expenses.fields.status')}</th>
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
                    <MaterialIcon name="receipt_long" className="text-4xl mb-2 block opacity-50" />
                    {t('common.noData')}
                  </td>
                </tr>
              ) : (
                expenses.map((expense: Record<string, unknown>) => (
                  <tr key={expense.id as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <MaterialIcon name={getCategoryIcon(expense.category as string)} className="text-indigo-400 text-base" />
                        <Link
                          to="/expenses/$expenseId"
                          params={{ expenseId: expense.id as string }}
                          className="font-medium text-indigo-400 hover:underline"
                        >
                          {expense.title as string}
                        </Link>
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
                    <td className="py-3 px-4 text-center">
                      <span className={`inline-block px-2.5 py-1 rounded-full text-xs font-bold ${statusBadge(expense.status as string)}`}>
                        {t(`expenses.status.${expense.status}`)}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-center">
                      <Link
                        to="/expenses/$expenseId"
                        params={{ expenseId: expense.id as string }}
                        className="text-indigo-400 hover:underline text-xs mr-2"
                      >
                        {t('expenses.detail.view')}
                      </Link>
                      {expense.status === 'draft' && (
                        <button className="text-indigo-400 hover:underline text-xs">
                          {t('common.edit')}
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 0 && (
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
