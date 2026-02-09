import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';
import { format, startOfMonth, endOfMonth, subMonths } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const CATEGORY_COLORS: Record<string, string> = {
  transportation: '#818cf8',
  meals: '#f472b6',
  accommodation: '#fb923c',
  supplies: '#34d399',
  communication: '#60a5fa',
  entertainment: '#a78bfa',
  other: '#94a3b8',
};

export function ExpenseReportPage() {
  const { t, i18n } = useTranslation();
  const locale = i18n.language === 'ja' ? ja : enUS;
  const [period, setPeriod] = useState<'month' | 'quarter' | 'year'>('month');
  const [selectedMonth, setSelectedMonth] = useState(() => format(new Date(), 'yyyy-MM'));

  const dateRange = useMemo(() => {
    const base = new Date(selectedMonth + '-01');
    if (period === 'month') {
      return { start: startOfMonth(base), end: endOfMonth(base) };
    } else if (period === 'quarter') {
      return { start: subMonths(startOfMonth(base), 2), end: endOfMonth(base) };
    } else {
      const yearStart = new Date(base.getFullYear(), 0, 1);
      const yearEnd = new Date(base.getFullYear(), 11, 31);
      return { start: yearStart, end: yearEnd };
    }
  }, [selectedMonth, period]);

  const { data: reportData } = useQuery({
    queryKey: ['expense-report', dateRange.start, dateRange.end],
    queryFn: () => api.expenses.getReport({
      start_date: format(dateRange.start, 'yyyy-MM-dd'),
      end_date: format(dateRange.end, 'yyyy-MM-dd'),
    }),
  });

  const { data: monthlyTrend } = useQuery({
    queryKey: ['expense-monthly-trend', selectedMonth],
    queryFn: () => api.expenses.getMonthlyTrend({
      year: new Date(selectedMonth + '-01').getFullYear().toString(),
    }),
  });

  const report = reportData || {
    total_amount: 0,
    category_breakdown: [],
    department_breakdown: [],
    status_summary: { draft: 0, pending: 0, approved: 0, rejected: 0, reimbursed: 0 },
  };

  const trend = monthlyTrend?.data || monthlyTrend?.months || [];
  const categoryData = report.category_breakdown || [];
  const deptData = report.department_breakdown || [];
  const maxCategoryAmount = Math.max(...categoryData.map((c: Record<string, unknown>) => Number(c.amount) || 0), 1);
  const maxTrend = Math.max(...trend.map((m: Record<string, unknown>) => Number(m.amount) || 0), 1);

  const handleExportCSV = async () => {
    try {
      const blob = await api.expenses.exportCSV({
        start_date: format(dateRange.start, 'yyyy-MM-dd'),
        end_date: format(dateRange.end, 'yyyy-MM-dd'),
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `expense-report-${format(dateRange.start, 'yyyyMMdd')}-${format(dateRange.end, 'yyyyMMdd')}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      // Export failed silently
    }
  };

  const handleExportPDF = async () => {
    try {
      const blob = await api.expenses.exportPDF({
        start_date: format(dateRange.start, 'yyyy-MM-dd'),
        end_date: format(dateRange.end, 'yyyy-MM-dd'),
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `expense-report-${format(dateRange.start, 'yyyyMMdd')}-${format(dateRange.end, 'yyyyMMdd')}.pdf`;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      // Export failed silently
    }
  };

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('expenses.report.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('expenses.report.subtitle')}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={handleExportCSV}
            className="flex items-center gap-2 px-4 py-2 glass-subtle hover:bg-white/10 rounded-xl text-sm transition-all border border-white/5">
            <MaterialIcon name="download" className="text-base" />
            CSV
          </button>
          <button onClick={handleExportPDF}
            className="flex items-center gap-2 px-4 py-2 glass-subtle hover:bg-white/10 rounded-xl text-sm transition-all border border-white/5">
            <MaterialIcon name="picture_as_pdf" className="text-base text-red-400" />
            PDF
          </button>
        </div>
      </div>

      {/* フィルター */}
      <div className="glass-card rounded-2xl p-4 flex flex-wrap items-center gap-3">
        <input
          type="month"
          value={selectedMonth}
          onChange={(e) => setSelectedMonth(e.target.value)}
          className="px-3 py-2 glass-input rounded-xl text-sm"
        />
        <div className="flex rounded-xl overflow-hidden border border-white/10">
          {(['month', 'quarter', 'year'] as const).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-4 py-2 text-xs font-medium transition-all ${
                period === p
                  ? 'bg-indigo-500/20 text-indigo-400'
                  : 'text-muted-foreground hover:bg-white/5'
              }`}
            >
              {t(`expenses.report.period.${p}`)}
            </button>
          ))}
        </div>
        <span className="text-xs text-muted-foreground">
          {format(dateRange.start, 'PP', { locale })} ~ {format(dateRange.end, 'PP', { locale })}
        </span>
      </div>

      {/* サマリーカード */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center">
              <MaterialIcon name="savings" className="text-indigo-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.report.totalExpenses')}</p>
          </div>
          <p className="text-2xl font-bold gradient-text">¥{Number(report.total_amount).toLocaleString()}</p>
        </div>
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <MaterialIcon name="check_circle" className="text-emerald-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.report.approvedAmount')}</p>
          </div>
          <p className="text-2xl font-bold gradient-text">¥{Number(report.status_summary?.approved_amount || 0).toLocaleString()}</p>
        </div>
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-amber-500/20 flex items-center justify-center">
              <MaterialIcon name="pending_actions" className="text-amber-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.report.pendingAmount')}</p>
          </div>
          <p className="text-2xl font-bold gradient-text">¥{Number(report.status_summary?.pending_amount || 0).toLocaleString()}</p>
        </div>
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-blue-500/20 flex items-center justify-center">
              <MaterialIcon name="receipt" className="text-blue-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.report.totalCount')}</p>
          </div>
          <p className="text-2xl font-bold gradient-text">{report.total_count || 0}{t('expenses.report.claims')}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* カテゴリ別内訳 (横棒グラフ) */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="donut_large" className="text-indigo-400" />
            {t('expenses.report.categoryBreakdown')}
          </h2>
          {categoryData.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground text-sm">{t('common.noData')}</p>
          ) : (
            <div className="space-y-3">
              {categoryData.map((cat: Record<string, unknown>) => {
                const amount = Number(cat.amount) || 0;
                const pct = maxCategoryAmount > 0 ? (amount / maxCategoryAmount) * 100 : 0;
                const color = CATEGORY_COLORS[(cat.category as string) || 'other'] || CATEGORY_COLORS.other;
                return (
                  <div key={cat.category as string}>
                    <div className="flex items-center justify-between text-sm mb-1">
                      <span className="flex items-center gap-2">
                        <span className="size-3 rounded-full" style={{ backgroundColor: color }} />
                        {t(`expenses.categories.${cat.category}`)}
                      </span>
                      <span className="font-semibold">¥{amount.toLocaleString()}</span>
                    </div>
                    <div className="h-2 rounded-full bg-white/5 overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all duration-700"
                        style={{ width: `${pct}%`, backgroundColor: color }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {/* ドーナツ風パイチャート */}
          {categoryData.length > 0 && (
            <div className="mt-6 flex justify-center">
              <PieChart data={categoryData} t={t} />
            </div>
          )}
        </div>

        {/* 月次トレンド (棒グラフ) */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="trending_up" className="text-indigo-400" />
            {t('expenses.report.monthlyTrend')}
          </h2>
          {trend.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground text-sm">{t('common.noData')}</p>
          ) : (
            <div className="flex items-end gap-2 h-48">
              {trend.map((month: Record<string, unknown>, i: number) => {
                const amount = Number(month.amount) || 0;
                const pct = maxTrend > 0 ? (amount / maxTrend) * 100 : 0;
                return (
                  <div key={i} className="flex-1 flex flex-col items-center gap-1 h-full justify-end group">
                    <div className="opacity-0 group-hover:opacity-100 transition-opacity text-xs font-semibold text-center whitespace-nowrap">
                      ¥{amount.toLocaleString()}
                    </div>
                    <div
                      className="w-full rounded-t-lg transition-all duration-500 gradient-primary hover:shadow-glow-sm cursor-pointer"
                      style={{ height: `${Math.max(pct, 2)}%` }}
                      title={`¥${amount.toLocaleString()}`}
                    />
                    <span className="text-[10px] text-muted-foreground mt-1">
                      {month.month as string}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* 部門別内訳テーブル */}
      {deptData.length > 0 && (
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="corporate_fare" className="text-indigo-400" />
            {t('expenses.report.departmentBreakdown')}
          </h2>

          {/* モバイル: カードビュー */}
          <div className="space-y-3 md:hidden">
            {deptData.map((dept: Record<string, unknown>) => {
              const amount = Number(dept.amount) || 0;
              const count = Number(dept.count) || 1;
              const totalAll = Number(report.total_amount) || 1;
              const ratio = (amount / totalAll) * 100;
              return (
                <div key={dept.department as string} className="glass-subtle rounded-xl p-4 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{dept.department as string}</span>
                    <span className="font-bold gradient-text">¥{amount.toLocaleString()}</span>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div><span className="text-muted-foreground">{t('expenses.report.claimCount')}: </span>{count}</div>
                    <div><span className="text-muted-foreground">{t('expenses.report.avgPerClaim')}: </span>¥{Math.round(amount / count).toLocaleString()}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 h-2 rounded-full bg-white/5 overflow-hidden">
                      <div className="h-full rounded-full gradient-primary" style={{ width: `${ratio}%` }} />
                    </div>
                    <span className="text-xs text-muted-foreground w-10 text-right">{ratio.toFixed(1)}%</span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* デスクトップ: テーブルビュー */}
          <div className="hidden md:block overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/5">
                  <th className="text-left py-3 px-4 font-semibold">{t('expenses.report.department')}</th>
                  <th className="text-right py-3 px-4 font-semibold">{t('expenses.report.totalExpenses')}</th>
                  <th className="text-right py-3 px-4 font-semibold">{t('expenses.report.claimCount')}</th>
                  <th className="text-right py-3 px-4 font-semibold">{t('expenses.report.avgPerClaim')}</th>
                  <th className="text-left py-3 px-4 font-semibold">{t('expenses.report.ratio')}</th>
                </tr>
              </thead>
              <tbody>
                {deptData.map((dept: Record<string, unknown>) => {
                  const amount = Number(dept.amount) || 0;
                  const count = Number(dept.count) || 1;
                  const totalAll = Number(report.total_amount) || 1;
                  const ratio = (amount / totalAll) * 100;
                  return (
                    <tr key={dept.department as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="py-3 px-4 font-medium">{dept.department as string}</td>
                      <td className="py-3 px-4 text-right font-semibold">¥{amount.toLocaleString()}</td>
                      <td className="py-3 px-4 text-right">{count}</td>
                      <td className="py-3 px-4 text-right">¥{Math.round(amount / count).toLocaleString()}</td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2">
                          <div className="flex-1 h-2 rounded-full bg-white/5 overflow-hidden">
                            <div className="h-full rounded-full gradient-primary" style={{ width: `${ratio}%` }} />
                          </div>
                          <span className="text-xs text-muted-foreground w-10 text-right">{ratio.toFixed(1)}%</span>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ステータス別件数 */}
      <div className="glass-card rounded-2xl p-6">
        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <MaterialIcon name="pie_chart" className="text-indigo-400" />
          {t('expenses.report.statusSummary')}
        </h2>
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          {(['draft', 'pending', 'approved', 'rejected', 'reimbursed'] as const).map((status) => {
            const statusColors: Record<string, string> = {
              draft: 'bg-gray-500/20 text-gray-400',
              pending: 'bg-amber-500/20 text-amber-400',
              approved: 'bg-emerald-500/20 text-emerald-400',
              rejected: 'bg-red-500/20 text-red-400',
              reimbursed: 'bg-blue-500/20 text-blue-400',
            };
            return (
              <div key={status} className="glass-subtle rounded-xl p-4 text-center">
                <span className={`inline-flex items-center justify-center size-10 rounded-xl ${statusColors[status]} mb-2`}>
                  <MaterialIcon name={status === 'draft' ? 'edit_note' : status === 'pending' ? 'pending' : status === 'approved' ? 'check' : status === 'rejected' ? 'close' : 'payments'} />
                </span>
                <p className="text-xl font-bold">{report.status_summary?.[status] || 0}</p>
                <p className="text-xs text-muted-foreground">{t(`expenses.status.${status}`)}</p>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

/* CSS-only ドーナツチャート */
function PieChart({ data, t }: { data: Record<string, unknown>[]; t: (key: string) => string }) {
  const total = data.reduce((s, c) => s + (Number(c.amount) || 0), 0);
  if (total === 0) return null;

  let accumulated = 0;
  const segments = data.map((cat) => {
    const amount = Number(cat.amount) || 0;
    const pct = (amount / total) * 100;
    const start = accumulated;
    accumulated += pct;
    return {
      category: cat.category as string,
      pct,
      start,
      color: CATEGORY_COLORS[(cat.category as string)] || CATEGORY_COLORS.other,
    };
  });

  const gradientParts = segments.map(
    (s) => `${s.color} ${s.start}% ${s.start + s.pct}%`
  );

  return (
    <div className="relative">
      <div
        className="size-40 rounded-full"
        style={{
          background: `conic-gradient(${gradientParts.join(', ')})`,
        }}
      >
        <div className="absolute inset-4 rounded-full bg-card flex items-center justify-center glass-subtle">
          <div className="text-center">
            <p className="text-lg font-bold gradient-text">¥{total.toLocaleString()}</p>
            <p className="text-[10px] text-muted-foreground">{t('expenses.fields.totalAmount')}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
