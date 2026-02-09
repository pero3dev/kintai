import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function ExpenseDashboardPage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const locale = i18n.language === 'ja' ? ja : enUS;

  const { data: stats } = useQuery({
    queryKey: ['expense-stats'],
    queryFn: () => api.expenses.getStats(),
  });

  const { data: recentData } = useQuery({
    queryKey: ['expenses', 'recent'],
    queryFn: () => api.expenses.getList({ page: 1, page_size: 5 }),
  });

  const recent = recentData?.data || recentData?.expenses || [];

  const statCards = [
    {
      icon: 'receipt_long',
      label: t('expenses.stats.totalThisMonth'),
      value: stats?.total_this_month != null ? `¥${Number(stats.total_this_month).toLocaleString()}` : '¥0',
      color: 'text-indigo-400',
      bg: 'bg-indigo-500/20',
    },
    {
      icon: 'pending_actions',
      label: t('expenses.stats.pendingCount'),
      value: stats?.pending_count ?? 0,
      color: 'text-amber-400',
      bg: 'bg-amber-500/20',
    },
    {
      icon: 'check_circle',
      label: t('expenses.stats.approvedThisMonth'),
      value: stats?.approved_this_month != null ? `¥${Number(stats.approved_this_month).toLocaleString()}` : '¥0',
      color: 'text-emerald-400',
      bg: 'bg-emerald-500/20',
    },
    {
      icon: 'account_balance_wallet',
      label: t('expenses.stats.reimbursedTotal'),
      value: stats?.reimbursed_total != null ? `¥${Number(stats.reimbursed_total).toLocaleString()}` : '¥0',
      color: 'text-blue-400',
      bg: 'bg-blue-500/20',
    },
  ];

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

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('expenses.dashboard.title')}</h1>
          <p className="text-muted-foreground text-xs sm:text-sm mt-0.5">{t('expenses.dashboard.subtitle')}</p>
        </div>
        <Link
          to="/expenses/new"
          className="flex items-center justify-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all duration-300"
        >
          <MaterialIcon name="add" className="text-lg" />
          {t('expenses.newExpense')}
        </Link>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((card, i) => (
          <div key={i} className="glass-card stat-shimmer rounded-2xl p-5">
            <div className="flex items-center gap-3 mb-3">
              <div className={`size-10 rounded-xl flex items-center justify-center ${card.bg}`}>
                <MaterialIcon name={card.icon} className={card.color} />
              </div>
              <p className="text-sm text-muted-foreground">{card.label}</p>
            </div>
            <p className={`text-2xl font-bold gradient-text`}>{card.value}</p>
          </div>
        ))}
      </div>

      {/* クイックアクション + 最近の経費 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* クイックアクション */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="bolt" className="text-indigo-400" />
            {t('expenses.dashboard.quickActions')}
          </h2>
          <div className="space-y-3">
            <Link
              to="/expenses/new"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <div className="size-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
                <MaterialIcon name="add_card" className="text-emerald-400" />
              </div>
              <div>
                <p className="text-sm font-medium">{t('expenses.newExpense')}</p>
                <p className="text-xs text-muted-foreground">{t('expenses.dashboard.newExpenseDesc')}</p>
              </div>
            </Link>
            <Link
                to="/expenses/history"
                className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
              >
              <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center">
                <MaterialIcon name="history" className="text-indigo-400" />
              </div>
              <div>
                <p className="text-sm font-medium">{t('expenses.history.title')}</p>
                <p className="text-xs text-muted-foreground">{t('expenses.dashboard.historyDesc')}</p>
              </div>
            </Link>
            <Link
              to="/expenses/report"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <div className="size-10 rounded-xl bg-blue-500/20 flex items-center justify-center">
                <MaterialIcon name="bar_chart" className="text-blue-400" />
              </div>
              <div>
                <p className="text-sm font-medium">{t('expenses.nav.report')}</p>
                <p className="text-xs text-muted-foreground">{t('expenses.dashboard.reportDesc')}</p>
              </div>
            </Link>
            <Link
              to="/expenses/templates"
              className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
            >
              <div className="size-10 rounded-xl bg-purple-500/20 flex items-center justify-center">
                <MaterialIcon name="content_copy" className="text-purple-400" />
              </div>
              <div>
                <p className="text-sm font-medium">{t('expenses.nav.templates')}</p>
                <p className="text-xs text-muted-foreground">{t('expenses.dashboard.templatesDesc')}</p>
              </div>
            </Link>
            {(user?.role === 'admin' || user?.role === 'manager') && (
              <Link
                to="/expenses/approve"
                className="flex items-center gap-3 p-3 rounded-xl nav-item-hover transition-all"
              >
                <div className="size-10 rounded-xl bg-amber-500/20 flex items-center justify-center">
                  <MaterialIcon name="fact_check" className="text-amber-400" />
                </div>
                <div>
                  <p className="text-sm font-medium">{t('expenses.approve.title')}</p>
                  <p className="text-xs text-muted-foreground">{t('expenses.dashboard.approveDesc')}</p>
                </div>
              </Link>
            )}
          </div>
        </div>

        {/* 最近の経費精算 */}
        <div className="lg:col-span-2 glass-card rounded-2xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <MaterialIcon name="receipt" className="text-indigo-400" />
              {t('expenses.dashboard.recentExpenses')}
            </h2>
            <Link
              to="/expenses/history"
              className="text-sm text-indigo-400 hover:underline flex items-center gap-1"
            >
              {t('common.viewAll')}
              <MaterialIcon name="arrow_forward" className="text-base" />
            </Link>
          </div>

          {recent.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <MaterialIcon name="receipt_long" className="text-4xl mb-2 block opacity-50" />
              <p>{t('common.noData')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {recent.map((expense: Record<string, unknown>) => (
                <div
                  key={expense.id as string}
                  className="flex items-center justify-between p-3 rounded-xl glass-subtle hover:bg-white/5 transition-all"
                >
                  <div className="flex items-center gap-3">
                    <div className="size-10 rounded-xl bg-indigo-500/10 flex items-center justify-center">
                      <MaterialIcon name={getCategoryIcon(expense.category as string)} className="text-indigo-400" />
                    </div>
                    <div>
                      <p className="text-sm font-medium">{expense.title as string}</p>
                      <p className="text-xs text-muted-foreground">
                        {expense.expense_date ? format(new Date(expense.expense_date as string), 'PP', { locale }) : ''}
                        {expense.category ? ` · ${t(`expenses.categories.${expense.category}`)}` : ''}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-semibold">¥{Number(expense.amount).toLocaleString()}</p>
                    <span className={`inline-block px-2 py-0.5 rounded-full text-[10px] font-bold ${statusBadge(expense.status as string)}`}>
                      {t(`expenses.status.${expense.status}`)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function getCategoryIcon(category: string): string {
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
}
