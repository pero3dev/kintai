import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const STAGE_PIPELINE = ['applied', 'screening', 'interview', 'technical', 'offer', 'hired', 'rejected'] as const;
const STAGE_COLORS: Record<string, string> = {
  applied: 'bg-gray-500/20 text-gray-400',
  screening: 'bg-blue-500/20 text-blue-400',
  interview: 'bg-indigo-500/20 text-indigo-400',
  technical: 'bg-purple-500/20 text-purple-400',
  offer: 'bg-amber-500/20 text-amber-400',
  hired: 'bg-green-500/20 text-green-400',
  rejected: 'bg-red-500/20 text-red-400',
};

export function HRRecruitmentPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'positions' | 'applicants'>('positions');
  const [showPositionForm, setShowPositionForm] = useState(false);
  const [showApplicantForm, setShowApplicantForm] = useState(false);
  const [selectedPosition, setSelectedPosition] = useState<string>('');

  const [posForm, setPosForm] = useState({
    title: '', department: '', description: '', requirements: '',
    employment_type: 'full_time', salary_range: '', location: '',
  });
  const [appForm, setAppForm] = useState({
    position_id: '', name: '', email: '', phone: '', resume_url: '', cover_letter: '',
  });

  const { data: positionsData, isLoading: posLoading } = useQuery({
    queryKey: ['hr-positions'],
    queryFn: () => api.hr.getPositions({}),
  });
  const { data: applicantsData, isLoading: appLoading } = useQuery({
    queryKey: ['hr-applicants'],
    queryFn: () => api.hr.getApplicants({}),
  });

  const positions: Record<string, unknown>[] = positionsData?.data || positionsData || [];
  const applicants: Record<string, unknown>[] = applicantsData?.data || applicantsData || [];

  const createPosMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createPosition(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-positions'] }); setShowPositionForm(false); },
  });
  const createAppMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createApplicant(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-applicants'] }); setShowApplicantForm(false); },
  });
  const updateStageMutation = useMutation({
    mutationFn: ({ id, stage }: { id: string; stage: string }) => api.hr.updateApplicantStage(id, stage),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-applicants'] }),
  });

  const filteredApplicants = selectedPosition
    ? applicants.filter(a => a.position_id === selectedPosition)
    : applicants;

  // パイプライン集計
  const pipeline = STAGE_PIPELINE.map(stage => ({
    stage,
    count: filteredApplicants.filter(a => a.stage === stage).length,
  }));

  const tabItems = [
    { id: 'positions' as const, label: t('hr.recruitment.openPositions'), icon: 'work' },
    { id: 'applicants' as const, label: t('hr.recruitment.applicants'), icon: 'people' },
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.recruitment.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.recruitment.subtitle')}</p>
          </div>
        </div>
        <button onClick={() => tab === 'positions' ? setShowPositionForm(true) : setShowApplicantForm(true)}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
          <MaterialIcon name="add" className="text-lg" />
          {tab === 'positions' ? t('hr.recruitment.addPosition') : t('hr.recruitment.addApplicant')}
        </button>
      </div>

      {/* タブ */}
      <div className="flex gap-1 glass-subtle rounded-xl p-1 w-fit">
        {tabItems.map(tb => (
          <button key={tb.id} onClick={() => setTab(tb.id)}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              tab === tb.id ? 'bg-indigo-500/20 text-indigo-400' : 'text-muted-foreground hover:bg-white/5'}`}>
            <MaterialIcon name={tb.icon} className="text-base" />
            {tb.label}
          </button>
        ))}
      </div>

      {tab === 'positions' ? (
        /* ポジション一覧 */
        posLoading ? (
          <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
        ) : positions.length === 0 ? (
          <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
            <MaterialIcon name="work" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('common.noData')}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {positions.map((pos) => (
              <div key={pos.id as string} className="glass-card rounded-2xl p-5 hover:bg-white/5 transition-all">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h3 className="font-semibold text-sm">{String(pos.title)}</h3>
                    <p className="text-xs text-muted-foreground">{String(pos.department || '')} · {String(pos.location || '')}</p>
                  </div>
                  <span className={`px-2 py-1 rounded-lg text-[10px] font-medium ${
                    pos.status === 'open' ? 'bg-green-500/20 text-green-400' :
                    pos.status === 'closed' ? 'bg-red-500/20 text-red-400' : 'bg-gray-500/20 text-gray-400'
                  }`}>{t(`hr.recruitment.statuses.${String(pos.status || 'open')}`)}</span>
                </div>
                {!!pos.description && <p className="text-xs text-muted-foreground mb-3 line-clamp-2">{String(pos.description)}</p>}
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-3 text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <MaterialIcon name="badge" className="text-sm" />
                      {pos.employment_type ? t(`hr.employees.types.${pos.employment_type}`) : ''}
                    </span>
                    {!!pos.salary_range && (
                      <span className="flex items-center gap-1">
                        <MaterialIcon name="payments" className="text-sm" />
                        {String(pos.salary_range)}
                      </span>
                    )}
                  </div>
                  <span className="flex items-center gap-1 text-indigo-400 font-medium">
                    <MaterialIcon name="people" className="text-sm" />
                    {String(pos.applicant_count || 0)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )
      ) : (
        /* 応募者パイプライン */
        <>
          {/* ポジションフィルター */}
          <div className="flex flex-wrap gap-3">
            <select value={selectedPosition} onChange={(e) => setSelectedPosition(e.target.value)}
              className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[200px]">
              <option value="">{t('hr.recruitment.allPositions')}</option>
              {positions.map(p => <option key={p.id as string} value={p.id as string}>{String(p.title)}</option>)}
            </select>
          </div>

          {/* パイプラインサマリー */}
          <div className="flex gap-2 overflow-x-auto pb-2">
            {pipeline.filter(p => p.stage !== 'rejected').map(p => (
              <div key={p.stage} className={`flex-shrink-0 px-3 py-2 rounded-xl text-center min-w-[80px] ${STAGE_COLORS[p.stage]}`}>
                <p className="text-lg font-bold">{p.count}</p>
                <p className="text-[10px]">{t(`hr.recruitment.stages.${p.stage}`)}</p>
              </div>
            ))}
          </div>

          {/* 応募者リスト */}
          {appLoading ? (
            <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
          ) : filteredApplicants.length === 0 ? (
            <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
              <MaterialIcon name="people" className="text-4xl mb-2 block opacity-50" />
              <p className="text-sm">{t('common.noData')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredApplicants.map((app) => {
                const stage = (app.stage as string) || 'applied';
                const stageIdx = STAGE_PIPELINE.indexOf(stage as typeof STAGE_PIPELINE[number]);
                return (
                  <div key={app.id as string} className="glass-card rounded-2xl p-4 sm:p-5 hover:bg-white/5 transition-all">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-3">
                      <div className="flex items-center gap-3">
                        <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center shrink-0">
                          <span className="text-indigo-400 font-bold text-sm">
                            {(app.name as string || '?').substring(0, 1)}
                          </span>
                        </div>
                        <div>
                          <p className="font-semibold text-sm">{String(app.name)}</p>
                          <p className="text-xs text-muted-foreground">{String(app.position_title || '')} · {String(app.email || '')}</p>
                        </div>
                      </div>
                      <span className={`px-3 py-1 rounded-lg text-xs font-medium ${STAGE_COLORS[stage]}`}>
                        {t(`hr.recruitment.stages.${stage}`)}
                      </span>
                    </div>

                    {/* ステージ進行バー */}
                    <div className="flex items-center gap-1 mb-3">
                      {STAGE_PIPELINE.filter(s => s !== 'rejected').map((s, i) => (
                        <div key={s} className={`h-1.5 flex-1 rounded-full transition-all ${
                          i <= stageIdx ? 'bg-indigo-400' : 'bg-white/10'
                        }`} title={t(`hr.recruitment.stages.${s}`)} />
                      ))}
                    </div>

                    {/* ステージ変更ボタン */}
                    {stage !== 'hired' && stage !== 'rejected' && (
                      <div className="flex flex-wrap gap-2">
                        {stageIdx < STAGE_PIPELINE.length - 2 && (
                          <button onClick={() => updateStageMutation.mutate({ id: app.id as string, stage: STAGE_PIPELINE[stageIdx + 1] })}
                            className="px-3 py-1.5 text-xs font-medium glass-subtle rounded-lg hover:bg-white/10 transition-all flex items-center gap-1">
                            <MaterialIcon name="arrow_forward" className="text-sm" />
                            {t(`hr.recruitment.stages.${STAGE_PIPELINE[stageIdx + 1]}`)}
                          </button>
                        )}
                        <button onClick={() => updateStageMutation.mutate({ id: app.id as string, stage: 'rejected' })}
                          className="px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-500/10 rounded-lg transition-all flex items-center gap-1">
                          <MaterialIcon name="close" className="text-sm" />
                          {t('hr.recruitment.stages.rejected')}
                        </button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </>
      )}

      {/* ポジション作成モーダル */}
      {showPositionForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.recruitment.addPosition')}</h2>
              <button onClick={() => setShowPositionForm(false)} className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.positionTitle')}</label>
                <input value={posForm.title} onChange={(e) => setPosForm(f => ({...f, title: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.recruitment.department')}</label>
                  <input value={posForm.department} onChange={(e) => setPosForm(f => ({...f, department: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.recruitment.location')}</label>
                  <input value={posForm.location} onChange={(e) => setPosForm(f => ({...f, location: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.description')}</label>
                <textarea value={posForm.description} onChange={(e) => setPosForm(f => ({...f, description: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.requirements')}</label>
                <textarea value={posForm.requirements} onChange={(e) => setPosForm(f => ({...f, requirements: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.salaryRange')}</label>
                <input value={posForm.salary_range} onChange={(e) => setPosForm(f => ({...f, salary_range: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" placeholder="¥3,000,000 〜 ¥5,000,000" />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => createPosMutation.mutate({ ...posForm })}
                disabled={!posForm.title || createPosMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.create')}
              </button>
              <button onClick={() => setShowPositionForm(false)}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}

      {/* 応募者追加モーダル */}
      {showApplicantForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.recruitment.addApplicant')}</h2>
              <button onClick={() => setShowApplicantForm(false)} className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.position')}</label>
                <select value={appForm.position_id} onChange={(e) => setAppForm(f => ({...f, position_id: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                  <option value="">{t('common.selectPlaceholder')}</option>
                  {positions.map(p => <option key={p.id as string} value={p.id as string}>{String(p.title)}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.applicantName')}</label>
                <input value={appForm.name} onChange={(e) => setAppForm(f => ({...f, name: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.recruitment.email')}</label>
                  <input type="email" value={appForm.email} onChange={(e) => setAppForm(f => ({...f, email: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.recruitment.phone')}</label>
                  <input value={appForm.phone} onChange={(e) => setAppForm(f => ({...f, phone: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.recruitment.coverLetter')}</label>
                <textarea value={appForm.cover_letter} onChange={(e) => setAppForm(f => ({...f, cover_letter: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={4} />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => createAppMutation.mutate({ ...appForm })}
                disabled={!appForm.name || !appForm.position_id || createAppMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.create')}
              </button>
              <button onClick={() => setShowApplicantForm(false)}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
