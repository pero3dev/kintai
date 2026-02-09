import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const CATEGORY_ICONS: Record<string, { icon: string; color: string }> = {
  compliance: { icon: 'gavel', color: 'text-red-400 bg-red-500/20' },
  technical: { icon: 'code', color: 'text-blue-400 bg-blue-500/20' },
  leadership: { icon: 'psychology', color: 'text-purple-400 bg-purple-500/20' },
  communication: { icon: 'forum', color: 'text-green-400 bg-green-500/20' },
  safety: { icon: 'health_and_safety', color: 'text-amber-400 bg-amber-500/20' },
};

const STATUS_COLORS: Record<string, string> = {
  upcoming: 'bg-blue-500/20 text-blue-400',
  in_progress: 'bg-indigo-500/20 text-indigo-400',
  completed: 'bg-green-500/20 text-green-400',
  cancelled: 'bg-red-500/20 text-red-400',
};

export function HRTrainingPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [filters, setFilters] = useState({ category: '', status: '' });

  const [form, setForm] = useState({
    title: '', description: '', category: 'technical', instructor: '',
    start_date: '', end_date: '', max_participants: '', location: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-training', filters],
    queryFn: () => api.hr.getTrainingPrograms(filters),
  });

  const programs: Record<string, unknown>[] = data?.data || data || [];

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createTrainingProgram(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-training'] }); setShowForm(false); },
  });
  const enrollMutation = useMutation({
    mutationFn: (id: string) => api.hr.enrollTraining(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-training'] }),
  });
  const completeMutation = useMutation({
    mutationFn: (id: string) => api.hr.completeTraining(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-training'] }),
  });

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.training.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.training.subtitle')}</p>
          </div>
        </div>
        <button onClick={() => setShowForm(true)}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
          <MaterialIcon name="add" className="text-lg" />
          {t('hr.training.addProgram')}
        </button>
      </div>

      {/* フィルター */}
      <div className="flex flex-wrap gap-3">
        <select value={filters.category} onChange={(e) => setFilters(f => ({...f, category: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[140px]">
          <option value="">{t('hr.training.category')}</option>
          {['compliance', 'technical', 'leadership', 'communication', 'safety'].map(c => (
            <option key={c} value={c}>{t(`hr.training.categories.${c}`)}</option>
          ))}
        </select>
        <select value={filters.status} onChange={(e) => setFilters(f => ({...f, status: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[120px]">
          <option value="">{t('hr.training.trainingStatus')}</option>
          {['upcoming', 'in_progress', 'completed', 'cancelled'].map(s => (
            <option key={s} value={s}>{t(`hr.training.statuses.${s}`)}</option>
          ))}
        </select>
      </div>

      {/* プログラム一覧 */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : programs.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="school" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {programs.map((prog) => {
            const category = (prog.category as string) || 'technical';
            const catInfo = CATEGORY_ICONS[category] || CATEGORY_ICONS.technical;
            const status = (prog.status as string) || 'upcoming';
            const enrolled = (prog.enrolled_count as number) || 0;
            const max = (prog.max_participants as number) || 0;
            const isEnrolled = prog.is_enrolled as boolean;

            return (
              <div key={prog.id as string} className="glass-card rounded-2xl overflow-hidden hover:bg-white/5 transition-all">
                {/* カテゴリヘッダー */}
                <div className={`p-3 flex items-center gap-2 ${catInfo.color.split(' ')[1]}`}>
                  <MaterialIcon name={catInfo.icon} className={`text-xl ${catInfo.color.split(' ')[0]}`} />
                  <span className="text-xs font-medium">{t(`hr.training.categories.${category}`)}</span>
                  <span className={`ml-auto px-2 py-0.5 rounded text-[10px] font-medium ${STATUS_COLORS[status]}`}>
                    {t(`hr.training.statuses.${status}`)}
                  </span>
                </div>

                <div className="p-4 sm:p-5 space-y-3">
                  <div>
                    <h3 className="font-semibold text-sm mb-1">{String(prog.title)}</h3>
                    {!!prog.description && (
                      <p className="text-xs text-muted-foreground line-clamp-2">{String(prog.description)}</p>
                    )}
                  </div>

                  <div className="space-y-1.5 text-xs text-muted-foreground">
                    {!!prog.instructor && (
                      <div className="flex items-center gap-2">
                        <MaterialIcon name="person" className="text-sm" />
                        <span>{t('hr.training.instructor')}: {String(prog.instructor)}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2">
                      <MaterialIcon name="calendar_month" className="text-sm" />
                      <span>{String(prog.start_date)} ~ {String(prog.end_date)}</span>
                    </div>
                    {!!prog.location && (
                      <div className="flex items-center gap-2">
                        <MaterialIcon name="location_on" className="text-sm" />
                        <span>{String(prog.location)}</span>
                      </div>
                    )}
                  </div>

                  {/* 参加者プログレス */}
                  {max > 0 && (
                    <div>
                      <div className="flex items-center justify-between text-xs mb-1">
                        <span className="text-muted-foreground">{t('hr.training.participants')}</span>
                        <span className="font-medium">{enrolled}/{max}</span>
                      </div>
                      <div className="h-1.5 bg-white/5 rounded-full overflow-hidden">
                        <div className="h-full bg-indigo-400 rounded-full transition-all duration-300"
                          style={{ width: `${Math.min((enrolled / max) * 100, 100)}%` }} />
                      </div>
                    </div>
                  )}

                  {/* アクションボタン */}
                  <div className="flex gap-2 pt-1">
                    {status === 'upcoming' || status === 'in_progress' ? (
                      isEnrolled ? (
                        <button onClick={() => completeMutation.mutate(prog.id as string)}
                          className="flex-1 py-2 text-xs font-medium glass-subtle rounded-xl hover:bg-white/10 transition-all text-center flex items-center justify-center gap-1">
                          <MaterialIcon name="check" className="text-sm" />
                          {t('hr.training.complete')}
                        </button>
                      ) : (
                        <button onClick={() => enrollMutation.mutate(prog.id as string)}
                          disabled={max > 0 && enrolled >= max}
                          className="flex-1 py-2 text-xs font-medium gradient-primary text-white rounded-xl hover:shadow-glow-md transition-all text-center disabled:opacity-50">
                          {t('hr.training.enroll')}
                        </button>
                      )
                    ) : (
                      <span className="text-xs text-muted-foreground">{t(`hr.training.statuses.${status}`)}</span>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* プログラム作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.training.addProgram')}</h2>
              <button onClick={() => setShowForm(false)} className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.training.programName')}</label>
                <input value={form.title} onChange={(e) => setForm(f => ({...f, title: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.training.description')}</label>
                <textarea value={form.description} onChange={(e) => setForm(f => ({...f, description: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.training.category')}</label>
                  <select value={form.category} onChange={(e) => setForm(f => ({...f, category: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    {['compliance', 'technical', 'leadership', 'communication', 'safety'].map(c => (
                      <option key={c} value={c}>{t(`hr.training.categories.${c}`)}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.training.instructor')}</label>
                  <input value={form.instructor} onChange={(e) => setForm(f => ({...f, instructor: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.training.startDate')}</label>
                  <input type="date" value={form.start_date} onChange={(e) => setForm(f => ({...f, start_date: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.training.endDate')}</label>
                  <input type="date" value={form.end_date} onChange={(e) => setForm(f => ({...f, end_date: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.training.maxParticipants')}</label>
                  <input type="number" value={form.max_participants} onChange={(e) => setForm(f => ({...f, max_participants: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.training.location')}</label>
                  <input value={form.location} onChange={(e) => setForm(f => ({...f, location: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => createMutation.mutate({ ...form, max_participants: form.max_participants ? Number(form.max_participants) : undefined })}
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
