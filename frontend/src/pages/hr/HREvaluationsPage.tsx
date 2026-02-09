import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const SCORE_COLORS: Record<string, string> = {
  S: 'bg-amber-500/20 text-amber-400',
  A: 'bg-green-500/20 text-green-400',
  B: 'bg-blue-500/20 text-blue-400',
  C: 'bg-orange-500/20 text-orange-400',
  D: 'bg-red-500/20 text-red-400',
};

const STATUS_MAP: Record<string, { icon: string; color: string }> = {
  draft: { icon: 'edit_note', color: 'text-gray-400' },
  self_review: { icon: 'person', color: 'text-blue-400' },
  manager_review: { icon: 'supervisor_account', color: 'text-indigo-400' },
  completed: { icon: 'check_circle', color: 'text-green-400' },
};

export function HREvaluationsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'evaluations' | 'cycles'>('evaluations');
  const [showForm, setShowForm] = useState(false);
  const [showCycleForm, setShowCycleForm] = useState(false);
  const [filters, setFilters] = useState({ cycle_id: '', status: '' });

  const [form, setForm] = useState({
    employee_id: '', cycle_id: '', self_score: '', manager_score: '', self_comment: '', manager_comment: '',
  });
  const [cycleForm, setCycleForm] = useState({ name: '', start_date: '', end_date: '', description: '' });

  const { data: evaluationsData, isLoading } = useQuery({
    queryKey: ['hr-evaluations', filters],
    queryFn: () => api.hr.getEvaluations(filters),
  });
  const { data: cyclesData } = useQuery({
    queryKey: ['hr-evaluation-cycles'],
    queryFn: () => api.hr.getEvaluationCycles(),
  });

  const evaluations: Record<string, unknown>[] = evaluationsData?.data || evaluationsData || [];
  const cycles: Record<string, unknown>[] = cyclesData?.data || cyclesData || [];

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createEvaluation(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-evaluations'] }); setShowForm(false); },
  });
  const submitMutation = useMutation({
    mutationFn: (id: string) => api.hr.submitEvaluation(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-evaluations'] }),
  });
  const createCycleMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createEvaluationCycle(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-evaluation-cycles'] }); setShowCycleForm(false); },
  });

  const tabs = [
    { id: 'evaluations' as const, label: t('hr.evaluations.title'), icon: 'rate_review' },
    { id: 'cycles' as const, label: t('hr.evaluations.evaluationCycle'), icon: 'event_repeat' },
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.evaluations.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.evaluations.subtitle')}</p>
          </div>
        </div>
        <button onClick={() => { tab === 'evaluations' ? setShowForm(true) : setShowCycleForm(true); }}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
          <MaterialIcon name="add" className="text-lg" />
          {tab === 'evaluations' ? t('hr.evaluations.startEvaluation') : t('hr.evaluations.addCycle')}
        </button>
      </div>

      {/* タブ */}
      <div className="flex gap-1 glass-subtle rounded-xl p-1 w-fit">
        {tabs.map(tb => (
          <button key={tb.id} onClick={() => setTab(tb.id)}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              tab === tb.id ? 'bg-indigo-500/20 text-indigo-400' : 'text-muted-foreground hover:bg-white/5'}`}>
            <MaterialIcon name={tb.icon} className="text-base" />
            {tb.label}
          </button>
        ))}
      </div>

      {tab === 'evaluations' ? (
        <>
          {/* フィルター */}
          <div className="flex flex-wrap gap-3">
            <select value={filters.cycle_id} onChange={(e) => setFilters(f => ({...f, cycle_id: e.target.value}))}
              className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[160px]">
              <option value="">{t('hr.evaluations.evaluationCycle')}</option>
              {cycles.map(c => <option key={c.id as string} value={c.id as string}>{String(c.name)}</option>)}
            </select>
            <select value={filters.status} onChange={(e) => setFilters(f => ({...f, status: e.target.value}))}
              className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[140px]">
              <option value="">{t('hr.evaluations.status')}</option>
              {['draft', 'self_review', 'manager_review', 'completed'].map(s => (
                <option key={s} value={s}>{t(`hr.evaluations.statuses.${s}`)}</option>
              ))}
            </select>
          </div>

          {isLoading ? (
            <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
          ) : evaluations.length === 0 ? (
            <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
              <MaterialIcon name="rate_review" className="text-4xl mb-2 block opacity-50" />
              <p className="text-sm">{t('common.noData')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {evaluations.map((ev) => {
                const status = ev.status as string || 'draft';
                const sm = STATUS_MAP[status] || STATUS_MAP.draft;
                return (
                  <div key={ev.id as string} className="glass-card rounded-2xl p-4 sm:p-5 hover:bg-white/5 transition-all">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center shrink-0">
                          <span className="text-indigo-400 font-bold text-sm">
                            {(ev.employee_name as string || '?').substring(0, 1)}
                          </span>
                        </div>
                        <div>
                          <p className="font-semibold text-sm">{String(ev.employee_name || '-')}</p>
                          <p className="text-xs text-muted-foreground">{String(ev.cycle_name || '-')} · {String(ev.department || '')}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-3 flex-wrap">
                        {/* スコア */}
                        {!!ev.final_score && (
                          <span className={`px-3 py-1 rounded-lg text-xs font-bold ${SCORE_COLORS[ev.final_score as string] || 'bg-gray-500/20 text-gray-400'}`}>
                            {String(ev.final_score)}
                          </span>
                        )}
                        {/* ステータス */}
                        <span className={`flex items-center gap-1 text-xs ${sm.color}`}>
                          <MaterialIcon name={sm.icon} className="text-base" />
                          {t(`hr.evaluations.statuses.${status}`)}
                        </span>
                        {status !== 'completed' && (
                          <button onClick={() => submitMutation.mutate(ev.id as string)}
                            className="px-3 py-1.5 text-xs font-medium glass-subtle rounded-lg hover:bg-white/10 transition-all">
                            {t('hr.evaluations.submit')}
                          </button>
                        )}
                      </div>
                    </div>
                    {/* 評価基準スコア表示 */}
                    {Array.isArray(ev.criteria) && (
                      <div className="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-2">
                        {(ev.criteria as Record<string, unknown>[]).map((c, i) => (
                          <div key={i} className="glass-subtle rounded-lg p-2 text-center">
                            <p className="text-[10px] text-muted-foreground truncate">{String(c.name)}</p>
                            <p className="font-bold text-sm">{c.score as number || '-'}<span className="text-[10px] text-muted-foreground">/5</span></p>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </>
      ) : (
        /* 評価サイクル一覧 */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {cycles.length === 0 ? (
            <div className="col-span-full glass-card rounded-2xl p-12 text-center text-muted-foreground">
              <MaterialIcon name="event_repeat" className="text-4xl mb-2 block opacity-50" />
              <p className="text-sm">{t('common.noData')}</p>
            </div>
          ) : cycles.map((cycle) => (
            <div key={cycle.id as string} className="glass-card rounded-2xl p-5">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div className="size-10 rounded-xl bg-green-500/20 flex items-center justify-center">
                    <MaterialIcon name="event_repeat" className="text-green-400" />
                  </div>
                  <div>
                    <p className="font-semibold text-sm">{String(cycle.name)}</p>
                    <p className="text-xs text-muted-foreground">
                      {String(cycle.start_date)} ～ {String(cycle.end_date)}
                    </p>
                  </div>
                </div>
                <span className={`px-2 py-1 rounded-lg text-xs ${
                  cycle.status === 'active' ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'
                }`}>{t(`hr.evaluations.statuses.${String(cycle.status || 'draft')}`)}</span>
              </div>
              {!!cycle.description && <p className="text-xs text-muted-foreground">{String(cycle.description)}</p>}
              <div className="flex gap-2 mt-3 text-xs text-muted-foreground">
                <span>{t('hr.evaluations.totalEvaluations')}: {String(cycle.evaluation_count || 0)}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 評価作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.evaluations.startEvaluation')}</h2>
              <button onClick={() => setShowForm(false)} className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.evaluations.evaluationCycle')}</label>
                <select value={form.cycle_id} onChange={(e) => setForm(f => ({...f, cycle_id: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                  <option value="">{t('common.selectPlaceholder')}</option>
                  {cycles.map(c => <option key={c.id as string} value={c.id as string}>{String(c.name)}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.evaluations.selfComment')}</label>
                <textarea value={form.self_comment} onChange={(e) => setForm(f => ({...f, self_comment: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={4} />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.evaluations.selfScore')}</label>
                <select value={form.self_score} onChange={(e) => setForm(f => ({...f, self_score: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                  <option value="">−</option>
                  {['S', 'A', 'B', 'C', 'D'].map(s => (
                    <option key={s} value={s}>{s} - {t(`hr.evaluations.scores.${s}`)}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => createMutation.mutate({ ...form })}
                disabled={!form.cycle_id || createMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.create')}
              </button>
              <button onClick={() => setShowForm(false)}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}

      {/* サイクル作成モーダル */}
      {showCycleForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.evaluations.addCycle')}</h2>
              <button onClick={() => setShowCycleForm(false)} className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.evaluations.cycleName')}</label>
                <input value={cycleForm.name} onChange={(e) => setCycleForm(f => ({...f, name: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.evaluations.startDate')}</label>
                  <input type="date" value={cycleForm.start_date} onChange={(e) => setCycleForm(f => ({...f, start_date: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.evaluations.endDate')}</label>
                  <input type="date" value={cycleForm.end_date} onChange={(e) => setCycleForm(f => ({...f, end_date: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.evaluations.description')}</label>
                <textarea value={cycleForm.description} onChange={(e) => setCycleForm(f => ({...f, description: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => createCycleMutation.mutate({ ...cycleForm })}
                disabled={!cycleForm.name || createCycleMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.create')}
              </button>
              <button onClick={() => setShowCycleForm(false)}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
