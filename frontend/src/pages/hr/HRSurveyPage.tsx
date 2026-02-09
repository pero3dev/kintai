import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const STATUS_STYLES: Record<string, string> = {
  draft: 'bg-gray-500/20 text-gray-400',
  active: 'bg-green-500/20 text-green-400',
  closed: 'bg-blue-500/20 text-blue-400',
};

const TYPE_ICONS: Record<string, string> = {
  engagement: 'favorite',
  satisfaction: 'thumb_up',
  pulse: 'speed',
  exit: 'logout',
  onboarding: 'waving_hand',
  custom: 'tune',
};

export function HRSurveyPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [selectedSurvey, setSelectedSurvey] = useState<Record<string, unknown> | null>(null);
  const [showResults, setShowResults] = useState(false);
  const [formData, setFormData] = useState({ title: '', description: '', type: 'engagement', is_anonymous: true, questions: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-surveys', statusFilter],
    queryFn: () => api.hr.getSurveys({ status: statusFilter }),
  });
  const { data: resultsData } = useQuery({
    queryKey: ['hr-survey-results', selectedSurvey?.id],
    queryFn: () => api.hr.getSurveyResults(String(selectedSurvey!.id)),
    enabled: showResults && !!selectedSurvey,
  });

  const createMutation = useMutation({
    mutationFn: (d: Record<string, unknown>) => api.hr.createSurvey(d),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-surveys'] }); setShowForm(false); resetForm(); },
  });
  const publishMutation = useMutation({
    mutationFn: (id: string) => api.hr.publishSurvey(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-surveys'] }),
  });
  const closeMutation = useMutation({
    mutationFn: (id: string) => api.hr.closeSurvey(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-surveys'] }),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.hr.deleteSurvey(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-surveys'] }); setSelectedSurvey(null); },
  });

  const surveys: Record<string, unknown>[] = data?.data || data || [];
  const results = resultsData?.data || resultsData || {};
  const questionResults: Record<string, unknown>[] = results.questions || [];

  function resetForm() {
    setFormData({ title: '', description: '', type: 'engagement', is_anonymous: true, questions: '' });
  }

  function handleCreate() {
    const questions = formData.questions
      .split('\n')
      .filter(q => q.trim())
      .map((q, i) => ({ id: `q${i + 1}`, text: q.trim(), type: 'rating' }));
    createMutation.mutate({ ...formData, questions });
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.survey.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.survey.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="px-4 py-2.5 glass-input rounded-xl text-sm">
            <option value="">{t('hr.survey.allStatuses')}</option>
            {['draft', 'active', 'closed'].map(s => (
              <option key={s} value={s}>{t(`hr.survey.statuses.${s}`)}</option>
            ))}
          </select>
          <button onClick={() => setShowForm(true)} className="gradient-primary px-4 py-2.5 rounded-xl text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            <MaterialIcon name="add" className="text-sm" />
            {t('hr.survey.createSurvey')}
          </button>
        </div>
      </div>

      {/* 新規作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setShowForm(false)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg relative z-10 animate-scale-in max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">{t('hr.survey.createSurvey')}</h3>
            <div className="space-y-3">
              <input value={formData.title} onChange={e => setFormData(p => ({ ...p, title: e.target.value }))}
                placeholder={t('hr.survey.surveyTitle')} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <textarea value={formData.description} onChange={e => setFormData(p => ({ ...p, description: e.target.value }))}
                placeholder={t('hr.survey.description')} rows={2} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <select value={formData.type} onChange={e => setFormData(p => ({ ...p, type: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                {['engagement', 'satisfaction', 'pulse', 'exit', 'onboarding', 'custom'].map(t_type => (
                  <option key={t_type} value={t_type}>{t(`hr.survey.types.${t_type}`)}</option>
                ))}
              </select>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={formData.is_anonymous}
                  onChange={e => setFormData(p => ({ ...p, is_anonymous: e.target.checked }))} className="rounded" />
                {t('hr.survey.anonymous')}
              </label>
              <div>
                <p className="text-xs text-muted-foreground mb-1">{t('hr.survey.questions')} ({t('hr.survey.questionsPlaceholder')})</p>
                <textarea value={formData.questions}
                  onChange={e => setFormData(p => ({ ...p, questions: e.target.value }))}
                  placeholder={`${t('hr.survey.questionsPlaceholder')}`}
                  rows={5} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
            </div>
            <div className="flex gap-2 mt-4 justify-end">
              <button onClick={() => setShowForm(false)} className="px-4 py-2 glass-subtle rounded-xl text-sm">{t('common.cancel')}</button>
              <button onClick={handleCreate} className="gradient-primary px-4 py-2 rounded-xl text-sm font-medium">{t('common.save')}</button>
            </div>
          </div>
        </div>
      )}

      {/* 結果表示モーダル */}
      {showResults && selectedSurvey && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => { setShowResults(false); setSelectedSurvey(null); }}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-2xl relative z-10 animate-scale-in max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="text-lg font-bold">{String(selectedSurvey.title || '')}</h3>
                <p className="text-xs text-muted-foreground">{t('hr.survey.results')}</p>
              </div>
              <button onClick={() => { setShowResults(false); setSelectedSurvey(null); }} className="p-1 glass-subtle rounded-lg">
                <MaterialIcon name="close" className="text-sm" />
              </button>
            </div>

            {/* 概要統計 */}
            <div className="grid grid-cols-3 gap-3 mb-4">
              <div className="glass-subtle rounded-xl p-3 text-center">
                <p className="text-xl font-bold text-green-400">{results.response_rate ? `${Number(results.response_rate).toFixed(0)}%` : '-'}</p>
                <p className="text-[10px] text-muted-foreground">{t('hr.survey.responseRate')}</p>
              </div>
              <div className="glass-subtle rounded-xl p-3 text-center">
                <p className="text-xl font-bold text-blue-400">{results.avg_score ? Number(results.avg_score).toFixed(1) : '-'}</p>
                <p className="text-[10px] text-muted-foreground">{t('hr.survey.avgScore')}</p>
              </div>
              <div className="glass-subtle rounded-xl p-3 text-center">
                <p className="text-xl font-bold text-amber-400">{results.enps != null ? String(results.enps) : '-'}</p>
                <p className="text-[10px] text-muted-foreground">eNPS</p>
              </div>
            </div>

            {/* 質問別結果 */}
            {questionResults.length > 0 && (
              <div className="space-y-3">
                {questionResults.map((q, i) => (
                  <div key={i} className="glass-subtle rounded-xl p-3">
                    <p className="text-sm font-medium mb-2">{String(q.text || `Q${i + 1}`)}</p>
                    <div className="flex items-center gap-3">
                      <div className="flex-1 h-2 rounded-full bg-white/10">
                        <div className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all"
                          style={{ width: `${(Number(q.avg_score || 0) / 5) * 100}%` }} />
                      </div>
                      <span className="text-sm font-bold w-10 text-right">{Number(q.avg_score || 0).toFixed(1)}</span>
                    </div>
                    {q.responses != null && (
                      <p className="text-[10px] text-muted-foreground mt-1">{String(q.responses)} {t('hr.survey.responses')}</p>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* サーベイリスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : surveys.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="poll" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {surveys.map((survey) => {
            const status = String(survey.status || 'draft');
            const type = String(survey.type || 'custom');
            return (
              <div key={String(survey.id)} className="glass-card rounded-2xl p-4 hover:bg-white/10 transition-all">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <div className="size-9 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400">
                      <MaterialIcon name={TYPE_ICONS[type] || 'poll'} className="text-lg" />
                    </div>
                    <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${STATUS_STYLES[status] || STATUS_STYLES.draft}`}>
                      {t(`hr.survey.statuses.${status}`)}
                    </span>
                  </div>
                  <span className="text-[10px] text-muted-foreground">{t(`hr.survey.types.${type}`)}</span>
                </div>
                <h3 className="font-semibold text-sm mb-1 truncate">{String(survey.title || '')}</h3>
                {!!survey.description && (
                  <p className="text-xs text-muted-foreground line-clamp-2 mb-3">{String(survey.description)}</p>
                )}
                <div className="flex items-center justify-between text-[10px] text-muted-foreground mb-3">
                  <span>{Array.isArray(survey.questions) ? `${(survey.questions as unknown[]).length} ${t('hr.survey.questions')}` : ''}</span>
                  {survey.response_count != null && <span>{String(survey.response_count)} {t('hr.survey.responses')}</span>}
                </div>
                <div className="flex gap-2">
                  {status === 'draft' && (
                    <button onClick={() => publishMutation.mutate(String(survey.id))}
                      className="flex-1 px-3 py-1.5 text-xs gradient-primary rounded-lg font-medium flex items-center justify-center gap-1">
                      <MaterialIcon name="send" className="text-xs" /> {t('hr.survey.publish')}
                    </button>
                  )}
                  {status === 'active' && (
                    <button onClick={() => closeMutation.mutate(String(survey.id))}
                      className="flex-1 px-3 py-1.5 text-xs glass-subtle rounded-lg flex items-center justify-center gap-1 hover:bg-white/10">
                      <MaterialIcon name="lock" className="text-xs" /> {t('hr.survey.close')}
                    </button>
                  )}
                  {(status === 'active' || status === 'closed') && (
                    <button onClick={() => { setSelectedSurvey(survey); setShowResults(true); }}
                      className="flex-1 px-3 py-1.5 text-xs glass-subtle rounded-lg flex items-center justify-center gap-1 hover:bg-white/10">
                      <MaterialIcon name="bar_chart" className="text-xs" /> {t('hr.survey.results')}
                    </button>
                  )}
                  {status === 'draft' && (
                    <button onClick={() => deleteMutation.mutate(String(survey.id))}
                      className="px-3 py-1.5 text-xs text-red-400 glass-subtle rounded-lg hover:bg-red-500/10">
                      <MaterialIcon name="delete" className="text-xs" />
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
