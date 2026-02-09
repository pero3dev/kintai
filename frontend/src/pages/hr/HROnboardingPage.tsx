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
  overdue: 'bg-red-500/20 text-red-400',
};

export function HROnboardingPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [selectedOnboarding, setSelectedOnboarding] = useState<Record<string, unknown> | null>(null);
  const [formData, setFormData] = useState({ employee_id: '', template_id: '', start_date: '', mentor_id: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-onboardings', statusFilter],
    queryFn: () => api.hr.getOnboardings({ status: statusFilter }),
  });
  const { data: templatesData } = useQuery({
    queryKey: ['hr-onboarding-templates'],
    queryFn: () => api.hr.getOnboardingTemplates(),
  });

  const createMutation = useMutation({
    mutationFn: (d: typeof formData) => api.hr.createOnboarding(d),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-onboardings'] }); setShowForm(false); },
  });
  const toggleTaskMutation = useMutation({
    mutationFn: ({ obId, taskId }: { obId: string; taskId: string }) => api.hr.toggleOnboardingTask(obId, taskId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-onboardings'] });
      if (selectedOnboarding) {
        api.hr.getOnboarding(String(selectedOnboarding.id)).then(res => setSelectedOnboarding(res?.data || res));
      }
    },
  });

  const onboardings: Record<string, unknown>[] = data?.data || data || [];
  const templates: Record<string, unknown>[] = templatesData?.data || templatesData || [];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.onboarding.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.onboarding.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="px-4 py-2.5 glass-input rounded-xl text-sm">
            <option value="">{t('hr.onboarding.allStatuses')}</option>
            {['pending', 'in_progress', 'completed', 'overdue'].map(s => (
              <option key={s} value={s}>{t(`hr.onboarding.statuses.${s}`)}</option>
            ))}
          </select>
          <button onClick={() => setShowForm(true)} className="gradient-primary px-4 py-2.5 rounded-xl text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            <MaterialIcon name="add" className="text-sm" />
            {t('hr.onboarding.startOnboarding')}
          </button>
        </div>
      </div>

      {/* 新規作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setShowForm(false)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-md relative z-10 animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">{t('hr.onboarding.startOnboarding')}</h3>
            <div className="space-y-3">
              <input value={formData.employee_id} onChange={e => setFormData(p => ({ ...p, employee_id: e.target.value }))}
                placeholder={t('hr.onboarding.employee')} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <select value={formData.template_id} onChange={e => setFormData(p => ({ ...p, template_id: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                <option value="">{t('hr.onboarding.selectTemplate')}</option>
                {templates.map(tmpl => (
                  <option key={String(tmpl.id)} value={String(tmpl.id)}>{String(tmpl.name || '')}</option>
                ))}
              </select>
              <input type="date" value={formData.start_date} onChange={e => setFormData(p => ({ ...p, start_date: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <input value={formData.mentor_id} onChange={e => setFormData(p => ({ ...p, mentor_id: e.target.value }))}
                placeholder={t('hr.onboarding.mentor')} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
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
      {selectedOnboarding && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setSelectedOnboarding(null)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg relative z-10 animate-scale-in max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="text-lg font-bold">{String(selectedOnboarding.employee_name || '')}</h3>
                <p className="text-xs text-muted-foreground">
                  {t('hr.onboarding.startDate')}: {String(selectedOnboarding.start_date || '').substring(0, 10)}
                  {!!selectedOnboarding.mentor_name && ` · ${t('hr.onboarding.mentor')}: ${String(selectedOnboarding.mentor_name)}`}
                </p>
              </div>
              <button onClick={() => setSelectedOnboarding(null)} className="p-1 glass-subtle rounded-lg">
                <MaterialIcon name="close" className="text-sm" />
              </button>
            </div>

            {/* 進捗バー */}
            {(() => {
              const tasks: Record<string, unknown>[] = Array.isArray(selectedOnboarding.tasks) ? selectedOnboarding.tasks as Record<string, unknown>[] : [];
              const completed = tasks.filter(t => !!t.completed).length;
              const pct = tasks.length ? (completed / tasks.length) * 100 : 0;
              return (
                <div className="mb-4">
                  <div className="flex justify-between text-xs text-muted-foreground mb-1">
                    <span>{t('hr.onboarding.progress')}</span>
                    <span>{completed}/{tasks.length} ({pct.toFixed(0)}%)</span>
                  </div>
                  <div className="h-2 rounded-full bg-white/10 overflow-hidden">
                    <div className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all" style={{ width: `${pct}%` }} />
                  </div>
                </div>
              );
            })()}

            {/* タスクリスト */}
            {Array.isArray(selectedOnboarding.tasks) && (
              <div className="space-y-2">
                {(selectedOnboarding.tasks as Record<string, unknown>[]).map((task) => {
                  const cat = String(task.category || 'general');
                  return (
                    <div key={String(task.id)} className="flex items-center gap-3 p-2.5 glass-subtle rounded-xl">
                      <button onClick={() => toggleTaskMutation.mutate({ obId: String(selectedOnboarding.id), taskId: String(task.id) })}
                        className={`size-5 rounded flex items-center justify-center border shrink-0 ${!!task.completed ? 'bg-green-500/30 border-green-500' : 'border-white/20'}`}>
                        {!!task.completed && <MaterialIcon name="check" className="text-xs text-green-400" />}
                      </button>
                      <div className="flex-1 min-w-0">
                        <p className={`text-sm ${!!task.completed ? 'line-through text-muted-foreground' : ''}`}>{String(task.title || '')}</p>
                        {!!task.due_date && <p className="text-[10px] text-muted-foreground">{String(task.due_date).substring(0, 10)}</p>}
                      </div>
                      <span className="text-[10px] glass-subtle px-2 py-0.5 rounded shrink-0">
                        {t(`hr.onboarding.categories.${cat}`)}
                      </span>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}

      {/* オンボーディングリスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : onboardings.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="waving_hand" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {onboardings.map((ob) => {
            const status = String(ob.status || 'pending');
            const tasks: Record<string, unknown>[] = Array.isArray(ob.tasks) ? ob.tasks as Record<string, unknown>[] : [];
            const completed = tasks.filter(t => !!t.completed).length;
            const pct = tasks.length ? (completed / tasks.length) * 100 : 0;
            return (
              <div key={String(ob.id)} onClick={() => setSelectedOnboarding(ob)}
                className="glass-card rounded-2xl p-4 cursor-pointer hover:bg-white/10 transition-all">
                <div className="flex items-center gap-3 mb-2">
                  <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400 font-bold shrink-0">
                    {String(ob.employee_name || '?').substring(0, 1)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-semibold text-sm truncate">{String(ob.employee_name || '')}</p>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${STATUS_STYLES[status] || STATUS_STYLES.pending}`}>
                        {t(`hr.onboarding.statuses.${status}`)}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {String(ob.start_date || '').substring(0, 10)}
                      {!!ob.department && ` · ${String(ob.department)}`}
                    </p>
                  </div>
                  <MaterialIcon name="chevron_right" className="text-sm text-muted-foreground" />
                </div>
                <div className="ml-[52px]">
                  <div className="flex justify-between text-[10px] text-muted-foreground mb-1">
                    <span>{t('hr.onboarding.progress')}</span>
                    <span>{completed}/{tasks.length}</span>
                  </div>
                  <div className="h-1.5 rounded-full bg-white/10 overflow-hidden">
                    <div className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all" style={{ width: `${pct}%` }} />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
