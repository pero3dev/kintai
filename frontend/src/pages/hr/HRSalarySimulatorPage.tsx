import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function HRSalarySimulatorPage() {
  const { t } = useTranslation();
  const [department, setDepartment] = useState('');
  const [simForm, setSimForm] = useState({ grade: '', position: '', evaluation_score: '', years_of_service: '' });
  const [simResult, setSimResult] = useState<Record<string, unknown> | null>(null);

  const { data } = useQuery({
    queryKey: ['hr-salary-overview', department],
    queryFn: () => api.hr.getSalaryOverview({ department }),
  });
  const { data: budgetData } = useQuery({
    queryKey: ['hr-budget-overview', department],
    queryFn: () => api.hr.getBudgetOverview({ department }),
  });

  const simulateMutation = useMutation({
    mutationFn: (d: typeof simForm) => api.hr.simulateSalary(d),
    onSuccess: (res) => setSimResult(res?.data || res || {}),
  });

  const overview = data?.data || data || {};
  const budget = budgetData?.data || budgetData || {};
  const distributions: Record<string, unknown>[] = overview.department_breakdown || [];
  const budgetDepts: Record<string, unknown>[] = budget.departments || [];

  const stats = [
    { label: t('hr.salary.avgSalary'), value: overview.avg_salary ? `¥${Number(overview.avg_salary).toLocaleString()}` : '-', icon: 'payments', color: 'text-green-400' },
    { label: t('hr.salary.medianSalary'), value: overview.median_salary ? `¥${Number(overview.median_salary).toLocaleString()}` : '-', icon: 'trending_flat', color: 'text-blue-400' },
    { label: t('hr.salary.totalPayroll'), value: overview.total_payroll ? `¥${(Number(overview.total_payroll) / 1000000).toFixed(1)}M` : '-', icon: 'account_balance', color: 'text-amber-400' },
    { label: t('hr.salary.headcount'), value: overview.headcount ? String(overview.headcount) : '-', icon: 'group', color: 'text-indigo-400' },
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.salary.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.salary.subtitle')}</p>
          </div>
        </div>
        <input value={department} onChange={e => setDepartment(e.target.value)}
          placeholder={t('hr.salary.department')} className="px-4 py-2.5 glass-input rounded-xl text-sm w-32" />
      </div>

      {/* 概要統計 */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {stats.map((s, i) => (
          <div key={i} className="glass-card rounded-2xl p-4 text-center">
            <MaterialIcon name={s.icon} className={`text-2xl ${s.color} mb-1`} />
            <p className="text-xl font-bold">{s.value}</p>
            <p className="text-xs text-muted-foreground">{s.label}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* シミュレーター */}
        <div className="glass-card rounded-2xl p-4 sm:p-5">
          <h3 className="font-semibold text-sm mb-4 flex items-center gap-2">
            <MaterialIcon name="calculate" className="text-indigo-400" />
            {t('hr.salary.simulate')}
          </h3>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <input value={simForm.grade} onChange={e => setSimForm(p => ({ ...p, grade: e.target.value }))}
                placeholder={t('hr.salary.grade')} className="px-4 py-2.5 glass-input rounded-xl text-sm" />
              <input value={simForm.position} onChange={e => setSimForm(p => ({ ...p, position: e.target.value }))}
                placeholder={t('hr.salary.position')} className="px-4 py-2.5 glass-input rounded-xl text-sm" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <input type="number" value={simForm.evaluation_score}
                onChange={e => setSimForm(p => ({ ...p, evaluation_score: e.target.value }))}
                placeholder={t('hr.salary.evaluationScore')} className="px-4 py-2.5 glass-input rounded-xl text-sm" />
              <input type="number" value={simForm.years_of_service}
                onChange={e => setSimForm(p => ({ ...p, years_of_service: e.target.value }))}
                placeholder={t('hr.salary.yearsOfService')} className="px-4 py-2.5 glass-input rounded-xl text-sm" />
            </div>
            <button onClick={() => simulateMutation.mutate(simForm)}
              className="w-full gradient-primary px-4 py-2.5 rounded-xl text-sm font-medium flex items-center justify-center gap-1">
              <MaterialIcon name="play_arrow" className="text-sm" />
              {t('hr.salary.runSimulation')}
            </button>
          </div>
          {simResult && (
            <div className="mt-4 p-4 glass-subtle rounded-xl animate-scale-in">
              <p className="text-xs text-muted-foreground mb-2">{t('hr.salary.simulationResult')}</p>
              <div className="grid grid-cols-2 gap-3 text-center">
                <div>
                  <p className="text-xl font-bold text-green-400">¥{Number(simResult.proposed_salary || 0).toLocaleString()}</p>
                  <p className="text-[10px] text-muted-foreground">{t('hr.salary.proposedSalary')}</p>
                </div>
                <div>
                  <p className="text-xl font-bold text-amber-400">{simResult.percentile ? `${simResult.percentile}%` : '-'}</p>
                  <p className="text-[10px] text-muted-foreground">{t('hr.salary.marketPercentile')}</p>
                </div>
              </div>
              {!!simResult.explanation && (
                <p className="text-xs text-muted-foreground mt-3 glass-subtle rounded-lg p-2">{String(simResult.explanation)}</p>
              )}
            </div>
          )}
        </div>

        {/* 予算概要 */}
        <div className="glass-card rounded-2xl p-4 sm:p-5">
          <h3 className="font-semibold text-sm mb-4 flex items-center gap-2">
            <MaterialIcon name="account_balance_wallet" className="text-amber-400" />
            {t('hr.salary.budgetOverview')}
          </h3>
          {budget.total_budget && (
            <div className="mb-4 text-center">
              <p className="text-xs text-muted-foreground">{t('hr.salary.budgetUsed')}</p>
              <p className="text-2xl font-bold">
                ¥{(Number(budget.used_budget || 0) / 1000000).toFixed(1)}M
                <span className="text-sm text-muted-foreground"> / ¥{(Number(budget.total_budget) / 1000000).toFixed(1)}M</span>
              </p>
              <div className="h-3 rounded-full bg-white/10 mt-2 relative overflow-hidden">
                <div className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all"
                  style={{ width: `${Math.min((Number(budget.used_budget || 0) / Number(budget.total_budget)) * 100, 100)}%` }} />
              </div>
            </div>
          )}
          {budgetDepts.length > 0 && (
            <div className="space-y-2">
              {budgetDepts.map((dep, i) => (
                <div key={i} className="flex items-center gap-3">
                  <span className="text-xs w-24 truncate">{String(dep.name || '')}</span>
                  <div className="flex-1 h-2 rounded-full bg-white/10 relative">
                    <div className="h-full rounded-full bg-indigo-400/50 transition-all"
                      style={{ width: `${Math.min(Number(dep.usage_rate || 0), 100)}%` }} />
                  </div>
                  <span className="text-xs text-muted-foreground w-12 text-right">{Number(dep.usage_rate || 0).toFixed(0)}%</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* 部門別内訳 */}
      {distributions.length > 0 && (
        <div className="glass-card rounded-2xl p-4 sm:p-5">
          <h3 className="font-semibold text-sm mb-4">{t('hr.salary.departmentBreakdown')}</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {distributions.map((dep, i) => (
              <div key={i} className="glass-subtle rounded-xl p-3">
                <p className="font-semibold text-sm mb-1">{String(dep.name || '')}</p>
                <div className="grid grid-cols-2 gap-2 text-xs">
                  <div>
                    <p className="text-muted-foreground">{t('hr.salary.avgSalary')}</p>
                    <p className="font-bold">¥{Number(dep.avg_salary || 0).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">{t('hr.salary.headcount')}</p>
                    <p className="font-bold">{String(dep.headcount || 0)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
