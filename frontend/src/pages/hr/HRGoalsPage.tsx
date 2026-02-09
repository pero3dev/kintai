import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const PRIORITY_COLORS: Record<string, string> = {
  high: 'bg-red-500/20 text-red-400',
  medium: 'bg-amber-500/20 text-amber-400',
  low: 'bg-blue-500/20 text-blue-400',
};

const STATUS_COLORS: Record<string, string> = {
  not_started: 'bg-gray-500/20 text-gray-400',
  in_progress: 'bg-blue-500/20 text-blue-400',
  completed: 'bg-green-500/20 text-green-400',
  cancelled: 'bg-red-500/20 text-red-400',
};

export function HRGoalsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [filters, setFilters] = useState({ category: '', priority: '', status: '' });

  const [form, setForm] = useState({
    title: '', description: '', category: 'performance', priority: 'medium',
    target_date: '', key_results: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-goals', filters],
    queryFn: () => api.hr.getGoals(filters),
  });

  const goals: Record<string, unknown>[] = data?.data || data || [];

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createGoal(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-goals'] }); setShowForm(false); },
  });
  const updateProgressMutation = useMutation({
    mutationFn: ({ id, progress }: { id: string; progress: number }) => api.hr.updateGoalProgress(id, progress),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-goals'] }),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.hr.deleteGoal(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-goals'] }),
  });

  const stats = {
    total: goals.length,
    completed: goals.filter(g => g.status === 'completed').length,
    inProgress: goals.filter(g => g.status === 'in_progress').length,
    avgProgress: goals.length ? Math.round(goals.reduce((s, g) => s + ((g.progress as number) || 0), 0) / goals.length) : 0,
  };

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.goals.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.goals.subtitle')}</p>
          </div>
        </div>
        <button onClick={() => setShowForm(true)}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
          <MaterialIcon name="add" className="text-lg" />
          {t('hr.goals.addGoal')}
        </button>
      </div>

      {/* 統計 */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {[
          { label: t('hr.goals.totalGoals'), value: stats.total, icon: 'flag', color: 'text-indigo-400' },
          { label: t('hr.goals.statuses.in_progress'), value: stats.inProgress, icon: 'pending', color: 'text-blue-400' },
          { label: t('hr.goals.statuses.completed'), value: stats.completed, icon: 'check_circle', color: 'text-green-400' },
          { label: t('hr.goals.progress'), value: `${stats.avgProgress}%`, icon: 'trending_up', color: 'text-amber-400' },
        ].map((s, i) => (
          <div key={i} className="glass-card rounded-2xl p-4 text-center">
            <MaterialIcon name={s.icon} className={`text-2xl ${s.color} mb-1`} />
            <p className="text-2xl font-bold">{s.value}</p>
            <p className="text-xs text-muted-foreground">{s.label}</p>
          </div>
        ))}
      </div>

      {/* フィルター */}
      <div className="flex flex-wrap gap-3">
        <select value={filters.category} onChange={(e) => setFilters(f => ({...f, category: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[140px]">
          <option value="">{t('hr.goals.category')}</option>
          {['performance', 'skill', 'career', 'team'].map(c => (
            <option key={c} value={c}>{t(`hr.goals.categories.${c}`)}</option>
          ))}
        </select>
        <select value={filters.priority} onChange={(e) => setFilters(f => ({...f, priority: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[120px]">
          <option value="">{t('hr.goals.priority')}</option>
          {['high', 'medium', 'low'].map(p => (
            <option key={p} value={p}>{t(`hr.goals.priorities.${p}`)}</option>
          ))}
        </select>
        <select value={filters.status} onChange={(e) => setFilters(f => ({...f, status: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[120px]">
          <option value="">{t('hr.goals.goalStatus')}</option>
          {['not_started', 'in_progress', 'completed', 'cancelled'].map(s => (
            <option key={s} value={s}>{t(`hr.goals.statuses.${s}`)}</option>
          ))}
        </select>
      </div>

      {/* 目標リスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : goals.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="flag" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {goals.map((goal) => {
            const progress = (goal.progress as number) || 0;
            const status = (goal.status as string) || 'not_started';
            const priority = (goal.priority as string) || 'medium';
            return (
              <div key={goal.id as string} className="glass-card rounded-2xl p-4 sm:p-5 hover:bg-white/5 transition-all group">
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 mb-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1 flex-wrap">
                      <h3 className="font-semibold text-sm">{String(goal.title)}</h3>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${PRIORITY_COLORS[priority]}`}>
                        {t(`hr.goals.priorities.${priority}`)}
                      </span>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${STATUS_COLORS[status]}`}>
                        {t(`hr.goals.statuses.${status}`)}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">{String(goal.employee_name || '')} · {goal.category ? t(`hr.goals.categories.${goal.category}`) : ''}</p>
                    {!!goal.description && (
                      <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{String(goal.description)}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-2 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
                    <button onClick={() => { if (confirm(t('common.confirm'))) deleteMutation.mutate(goal.id as string); }}
                      className="p-1.5 hover:bg-red-500/10 text-red-400 rounded-lg text-xs">
                      <MaterialIcon name="delete" className="text-sm" />
                    </button>
                  </div>
                </div>

                {/* プログレスバー */}
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-muted-foreground">{t('hr.goals.progress')}</span>
                    <span className="font-bold">{progress}%</span>
                  </div>
                  <div className="h-2 bg-white/5 rounded-full overflow-hidden">
                    <div className={`h-full rounded-full transition-all duration-500 ${
                      progress >= 100 ? 'bg-green-400' : progress >= 50 ? 'bg-blue-400' : 'bg-amber-400'
                    }`} style={{ width: `${Math.min(progress, 100)}%` }} />
                  </div>
                  {status !== 'completed' && status !== 'cancelled' && (
                    <div className="flex items-center gap-2 mt-2">
                      <input type="range" min={0} max={100} value={progress}
                        onChange={(e) => updateProgressMutation.mutate({ id: goal.id as string, progress: Number(e.target.value) })}
                        className="flex-1 h-1 accent-indigo-400" />
                    </div>
                  )}
                </div>

                {/* KR表示 */}
                {Array.isArray(goal.key_results) ? (
                  <div className="mt-3 space-y-1.5">
                    <p className="text-[10px] text-muted-foreground uppercase tracking-wider">{t('hr.goals.keyResults')}</p>
                    {(goal.key_results as Record<string, unknown>[]).map((kr, i) => (
                      <div key={i} className="flex items-center gap-2 text-xs">
                        <MaterialIcon name={kr.completed ? 'check_circle' : 'radio_button_unchecked'}
                          className={`text-sm ${kr.completed ? 'text-green-400' : 'text-muted-foreground'}`} />
                        <span className={kr.completed ? 'line-through text-muted-foreground' : ''}>{String(kr.title)}</span>
                      </div>
                    ))}
                  </div>
                ) : null}

                {!!goal.target_date && (
                  <p className="text-[10px] text-muted-foreground mt-2 flex items-center gap-1">
                    <MaterialIcon name="event" className="text-xs" />
                    {t('hr.goals.targetDate')}: {String(goal.target_date)}
                  </p>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* 目標作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.goals.addGoal')}</h2>
              <button onClick={() => setShowForm(false)} className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.goals.goalTitle')}</label>
                <input value={form.title} onChange={(e) => setForm(f => ({...f, title: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.goals.description')}</label>
                <textarea value={form.description} onChange={(e) => setForm(f => ({...f, description: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.goals.category')}</label>
                  <select value={form.category} onChange={(e) => setForm(f => ({...f, category: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    {['performance', 'skill', 'career', 'team'].map(c => (
                      <option key={c} value={c}>{t(`hr.goals.categories.${c}`)}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.goals.priority')}</label>
                  <select value={form.priority} onChange={(e) => setForm(f => ({...f, priority: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    {['high', 'medium', 'low'].map(p => (
                      <option key={p} value={p}>{t(`hr.goals.priorities.${p}`)}</option>
                    ))}
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.goals.targetDate')}</label>
                <input type="date" value={form.target_date} onChange={(e) => setForm(f => ({...f, target_date: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.goals.keyResults')}</label>
                <textarea value={form.key_results} onChange={(e) => setForm(f => ({...f, key_results: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3}
                  placeholder={t('hr.goals.keyResults') + ' (1行1項目)'} />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => {
                const keyResults = form.key_results.split('\n').filter(Boolean).map(t => ({ title: t, completed: false }));
                createMutation.mutate({ ...form, key_results: keyResults });
              }}
                disabled={!form.title || createMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.create')}
              </button>
              <button onClick={() => setShowForm(false)}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
