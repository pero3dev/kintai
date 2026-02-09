import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const STATUS_STYLES: Record<string, string> = {
  pending: 'bg-amber-500/20 text-amber-400',
  in_progress: 'bg-blue-500/20 text-blue-400',
  completed: 'bg-green-500/20 text-green-400',
};

export function HROffboardingPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [selectedOffboarding, setSelectedOffboarding] = useState<Record<string, unknown> | null>(null);
  const [showAnalytics, setShowAnalytics] = useState(false);
  const [formData, setFormData] = useState({ employee_id: '', reason: 'resignation', last_working_date: '', notes: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-offboardings', statusFilter],
    queryFn: () => api.hr.getOffboardings({ status: statusFilter }),
  });
  const { data: analyticsData } = useQuery({
    queryKey: ['hr-turnover-analytics'],
    queryFn: () => api.hr.getTurnoverAnalytics({}),
    enabled: showAnalytics,
  });

  const createMutation = useMutation({
    mutationFn: (d: typeof formData) => api.hr.createOffboarding(d),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-offboardings'] }); setShowForm(false); },
  });
  const toggleChecklistMutation = useMutation({
    mutationFn: ({ obId, itemKey }: { obId: string; itemKey: string }) => api.hr.toggleOffboardingChecklist(obId, itemKey),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-offboardings'] });
      if (selectedOffboarding) {
        api.hr.getOffboarding(String(selectedOffboarding.id)).then(res => setSelectedOffboarding(res?.data || res));
      }
    },
  });

  const offboardings: Record<string, unknown>[] = data?.data || data || [];
  const analytics = analyticsData?.data || analyticsData || {};
  const reasonBreakdown: Record<string, unknown>[] = analytics.reason_breakdown || [];
  const deptBreakdown: Record<string, unknown>[] = analytics.department_breakdown || [];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.offboarding.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.offboarding.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <button onClick={() => setShowAnalytics(!showAnalytics)}
            className={`px-4 py-2.5 glass-subtle rounded-xl text-sm flex items-center gap-1 ${showAnalytics ? 'bg-white/10' : ''}`}>
            <MaterialIcon name="analytics" className="text-sm" />
            {t('hr.offboarding.analytics')}
          </button>
          <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="px-4 py-2.5 glass-input rounded-xl text-sm">
            <option value="">{t('hr.offboarding.allStatuses')}</option>
            {['pending', 'in_progress', 'completed'].map(s => (
              <option key={s} value={s}>{t(`hr.offboarding.statuses.${s}`)}</option>
            ))}
          </select>
          <button onClick={() => setShowForm(true)} className="gradient-primary px-4 py-2.5 rounded-xl text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            <MaterialIcon name="add" className="text-sm" />
            {t('hr.offboarding.startOffboarding')}
          </button>
        </div>
      </div>

      {/* 離職分析 */}
      {showAnalytics && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 animate-fade-in">
          <div className="glass-card rounded-2xl p-4 sm:p-5">
            <h3 className="font-semibold text-sm mb-4 flex items-center gap-2">
              <MaterialIcon name="trending_up" className="text-amber-400" />
              {t('hr.offboarding.turnoverRate')}
            </h3>
            <div className="text-center mb-4">
              <p className="text-3xl font-bold text-amber-400">{analytics.turnover_rate ? `${Number(analytics.turnover_rate).toFixed(1)}%` : '-'}</p>
              <p className="text-xs text-muted-foreground">{t('hr.offboarding.annualTurnover')}</p>
            </div>
            <div className="grid grid-cols-2 gap-3 text-center">
              <div className="glass-subtle rounded-lg p-2">
                <p className="text-lg font-bold">{String(analytics.total_departures || 0)}</p>
                <p className="text-[10px] text-muted-foreground">{t('hr.offboarding.totalDepartures')}</p>
              </div>
              <div className="glass-subtle rounded-lg p-2">
                <p className="text-lg font-bold">{String(analytics.avg_tenure || '-')}</p>
                <p className="text-[10px] text-muted-foreground">{t('hr.offboarding.avgTenure')}</p>
              </div>
            </div>
          </div>

          <div className="glass-card rounded-2xl p-4 sm:p-5">
            <h3 className="font-semibold text-sm mb-4">{t('hr.offboarding.reasonBreakdown')}</h3>
            {reasonBreakdown.length > 0 ? (
              <div className="space-y-2">
                {reasonBreakdown.map((r, i) => {
                  const maxCount = Math.max(...reasonBreakdown.map(x => Number(x.count || 0)), 1);
                  return (
                    <div key={i} className="flex items-center gap-3">
                      <span className="text-xs w-20 truncate">{t(`hr.offboarding.reasons.${String(r.reason || 'other')}`)}</span>
                      <div className="flex-1 h-2 rounded-full bg-white/10">
                        <div className="h-full rounded-full bg-red-400/50 transition-all"
                          style={{ width: `${(Number(r.count || 0) / maxCount) * 100}%` }} />
                      </div>
                      <span className="text-xs text-muted-foreground">{String(r.count || 0)}</span>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground text-center py-8">{t('common.noData')}</p>
            )}
          </div>

          {deptBreakdown.length > 0 && (
            <div className="glass-card rounded-2xl p-4 sm:p-5 lg:col-span-2">
              <h3 className="font-semibold text-sm mb-4">{t('hr.offboarding.departmentBreakdown')}</h3>
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
                {deptBreakdown.map((dep, i) => (
                  <div key={i} className="glass-subtle rounded-xl p-3 text-center">
                    <p className="text-sm font-medium mb-1">{String(dep.name || '')}</p>
                    <p className="text-lg font-bold text-red-400">{String(dep.count || 0)}</p>
                    <p className="text-[10px] text-muted-foreground">{Number(dep.rate || 0).toFixed(1)}%</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* 新規作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setShowForm(false)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-md relative z-10 animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">{t('hr.offboarding.startOffboarding')}</h3>
            <div className="space-y-3">
              <input value={formData.employee_id} onChange={e => setFormData(p => ({ ...p, employee_id: e.target.value }))}
                placeholder={t('hr.offboarding.employee')} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <select value={formData.reason} onChange={e => setFormData(p => ({ ...p, reason: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                {['resignation', 'retirement', 'termination', 'contract_end', 'transfer', 'other'].map(r => (
                  <option key={r} value={r}>{t(`hr.offboarding.reasons.${r}`)}</option>
                ))}
              </select>
              <input type="date" value={formData.last_working_date} onChange={e => setFormData(p => ({ ...p, last_working_date: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <textarea value={formData.notes} onChange={e => setFormData(p => ({ ...p, notes: e.target.value }))}
                placeholder={t('hr.offboarding.exitInterviewNotes')} rows={3} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
            </div>
            <div className="flex gap-2 mt-4 justify-end">
              <button onClick={() => setShowForm(false)} className="px-4 py-2 glass-subtle rounded-xl text-sm">{t('common.cancel')}</button>
              <button onClick={() => createMutation.mutate(formData)} className="gradient-primary px-4 py-2 rounded-xl text-sm font-medium">
                {t('common.save')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 詳細モーダル */}
      {selectedOffboarding && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setSelectedOffboarding(null)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg relative z-10 animate-scale-in max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="text-lg font-bold">{String(selectedOffboarding.employee_name || '')}</h3>
                <p className="text-xs text-muted-foreground">
                  {t(`hr.offboarding.reasons.${String(selectedOffboarding.reason || 'other')}`)} · {t('hr.offboarding.lastDay')}: {String(selectedOffboarding.last_working_date || '').substring(0, 10)}
                </p>
              </div>
              <button onClick={() => setSelectedOffboarding(null)} className="p-1 glass-subtle rounded-lg">
                <MaterialIcon name="close" className="text-sm" />
              </button>
            </div>

            {/* チェックリスト */}
            {Array.isArray(selectedOffboarding.checklist) && selectedOffboarding.checklist.length > 0 && (
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground mb-2">{t('hr.offboarding.checklist')}</p>
                {(selectedOffboarding.checklist as Record<string, unknown>[]).map((item) => (
                  <div key={String(item.key)} className="flex items-center gap-3 p-2.5 glass-subtle rounded-xl">
                    <button onClick={() => toggleChecklistMutation.mutate({ obId: String(selectedOffboarding.id), itemKey: String(item.key) })}
                      className={`size-5 rounded flex items-center justify-center border shrink-0 ${!!item.completed ? 'bg-green-500/30 border-green-500' : 'border-white/20'}`}>
                      {!!item.completed && <MaterialIcon name="check" className="text-xs text-green-400" />}
                    </button>
                    <span className={`text-sm flex-1 ${!!item.completed ? 'line-through text-muted-foreground' : ''}`}>
                      {t(`hr.offboarding.checklistItems.${String(item.key)}`) !== `hr.offboarding.checklistItems.${String(item.key)}`
                        ? t(`hr.offboarding.checklistItems.${String(item.key)}`)
                        : String(item.label || item.key)}
                    </span>
                  </div>
                ))}
              </div>
            )}

            {!!selectedOffboarding.notes && (
              <div className="mt-4">
                <p className="text-xs text-muted-foreground mb-1">{t('hr.offboarding.exitInterviewNotes')}</p>
                <p className="text-sm glass-subtle rounded-xl p-3">{String(selectedOffboarding.notes)}</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* オフボーディングリスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : offboardings.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="logout" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {offboardings.map((ob) => {
            const status = String(ob.status || 'pending');
            const checklist: Record<string, unknown>[] = Array.isArray(ob.checklist) ? ob.checklist as Record<string, unknown>[] : [];
            const completed = checklist.filter(c => !!c.completed).length;
            const pct = checklist.length ? (completed / checklist.length) * 100 : 0;
            return (
              <div key={String(ob.id)} onClick={() => setSelectedOffboarding(ob)}
                className="glass-card rounded-2xl p-4 cursor-pointer hover:bg-white/10 transition-all">
                <div className="flex items-center gap-3 mb-2">
                  <div className="size-10 rounded-xl bg-red-500/20 flex items-center justify-center text-red-400 font-bold shrink-0">
                    {String(ob.employee_name || '?').substring(0, 1)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-semibold text-sm truncate">{String(ob.employee_name || '')}</p>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${STATUS_STYLES[status] || STATUS_STYLES.pending}`}>
                        {t(`hr.offboarding.statuses.${status}`)}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {t(`hr.offboarding.reasons.${String(ob.reason || 'other')}`)} · {t('hr.offboarding.lastDay')}: {String(ob.last_working_date || '').substring(0, 10)}
                    </p>
                  </div>
                  <MaterialIcon name="chevron_right" className="text-sm text-muted-foreground" />
                </div>
                {checklist.length > 0 && (
                  <div className="ml-[52px]">
                    <div className="flex justify-between text-[10px] text-muted-foreground mb-1">
                      <span>{t('hr.offboarding.checklist')}</span>
                      <span>{completed}/{checklist.length}</span>
                    </div>
                    <div className="h-1.5 rounded-full bg-white/10 overflow-hidden">
                      <div className="h-full rounded-full bg-gradient-to-r from-red-500 to-amber-500 transition-all" style={{ width: `${pct}%` }} />
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
