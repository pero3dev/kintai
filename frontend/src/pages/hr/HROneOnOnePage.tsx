import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const STATUS_STYLES: Record<string, string> = {
  scheduled: 'bg-blue-500/20 text-blue-400',
  completed: 'bg-green-500/20 text-green-400',
  cancelled: 'bg-red-500/20 text-red-400',
};

const MOOD_ICONS: Record<string, string> = {
  great: 'sentiment_very_satisfied',
  good: 'sentiment_satisfied',
  neutral: 'sentiment_neutral',
  bad: 'sentiment_dissatisfied',
  terrible: 'sentiment_very_dissatisfied',
};

export function HROneOnOnePage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [selectedMeeting, setSelectedMeeting] = useState<Record<string, unknown> | null>(null);
  const [formData, setFormData] = useState({ employee_id: '', scheduled_date: '', agenda: '', notes: '', frequency: 'biweekly', mood: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-one-on-ones', statusFilter],
    queryFn: () => api.hr.getOneOnOnes({ status: statusFilter }),
  });

  const createMutation = useMutation({
    mutationFn: (d: typeof formData) => api.hr.createOneOnOne(d),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-one-on-ones'] }); setShowForm(false); resetForm(); },
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.hr.deleteOneOnOne(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-one-on-ones'] }); setSelectedMeeting(null); },
  });
  const toggleActionMutation = useMutation({
    mutationFn: ({ meetingId, actionId }: { meetingId: string; actionId: string }) =>
      api.hr.toggleActionItem(meetingId, actionId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-one-on-ones'] }),
  });

  const meetings: Record<string, unknown>[] = data?.data || data || [];

  function resetForm() {
    setFormData({ employee_id: '', scheduled_date: '', agenda: '', notes: '', frequency: 'biweekly', mood: '' });
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.oneOnOne.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.oneOnOne.subtitle')}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="px-4 py-2.5 glass-input rounded-xl text-sm">
            <option value="">{t('hr.oneOnOne.allStatuses')}</option>
            {['scheduled', 'completed', 'cancelled'].map(s => (
              <option key={s} value={s}>{t(`hr.oneOnOne.statuses.${s}`)}</option>
            ))}
          </select>
          <button onClick={() => setShowForm(true)} className="gradient-primary px-4 py-2.5 rounded-xl text-sm font-medium flex items-center gap-1 whitespace-nowrap">
            <MaterialIcon name="add" className="text-sm" />
            {t('hr.oneOnOne.schedule')}
          </button>
        </div>
      </div>

      {/* 新規作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setShowForm(false)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-md relative z-10 animate-scale-in" onClick={e => e.stopPropagation()}>
            <h3 className="text-lg font-bold mb-4">{t('hr.oneOnOne.schedule')}</h3>
            <div className="space-y-3">
              <input value={formData.employee_id} onChange={e => setFormData(p => ({ ...p, employee_id: e.target.value }))}
                placeholder={t('hr.oneOnOne.participant')} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <input type="datetime-local" value={formData.scheduled_date} onChange={e => setFormData(p => ({ ...p, scheduled_date: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              <select value={formData.frequency} onChange={e => setFormData(p => ({ ...p, frequency: e.target.value }))}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                {['weekly', 'biweekly', 'monthly'].map(f => (
                  <option key={f} value={f}>{t(`hr.oneOnOne.frequencies.${f}`)}</option>
                ))}
              </select>
              <textarea value={formData.agenda} onChange={e => setFormData(p => ({ ...p, agenda: e.target.value }))}
                placeholder={t('hr.oneOnOne.agenda')} rows={3} className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
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
      {selectedMeeting && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setSelectedMeeting(null)}>
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg relative z-10 animate-scale-in max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-start mb-4">
              <div>
                <h3 className="text-lg font-bold">{String(selectedMeeting.employee_name || '')}</h3>
                <p className="text-xs text-muted-foreground">{String(selectedMeeting.scheduled_date || '').substring(0, 16)}</p>
              </div>
              <button onClick={() => setSelectedMeeting(null)} className="p-1 glass-subtle rounded-lg">
                <MaterialIcon name="close" className="text-sm" />
              </button>
            </div>
            {!!selectedMeeting.mood && (
              <div className="flex items-center gap-2 mb-3">
                <MaterialIcon name={MOOD_ICONS[String(selectedMeeting.mood)] || 'sentiment_neutral'} className="text-2xl text-amber-400" />
                <span className="text-sm">{t(`hr.oneOnOne.moods.${String(selectedMeeting.mood)}`)}</span>
              </div>
            )}
            {!!selectedMeeting.agenda && (
              <div className="mb-3">
                <p className="text-xs text-muted-foreground mb-1">{t('hr.oneOnOne.agenda')}</p>
                <p className="text-sm glass-subtle rounded-xl p-3">{String(selectedMeeting.agenda)}</p>
              </div>
            )}
            {!!selectedMeeting.notes && (
              <div className="mb-3">
                <p className="text-xs text-muted-foreground mb-1">{t('hr.oneOnOne.notes')}</p>
                <p className="text-sm glass-subtle rounded-xl p-3">{String(selectedMeeting.notes)}</p>
              </div>
            )}
            {/* アクションアイテム */}
            {Array.isArray(selectedMeeting.action_items) && selectedMeeting.action_items.length > 0 && (
              <div className="mb-3">
                <p className="text-xs text-muted-foreground mb-2">{t('hr.oneOnOne.actionItems')}</p>
                <div className="space-y-1">
                  {(selectedMeeting.action_items as Record<string, unknown>[]).map((item) => (
                    <div key={String(item.id)} className="flex items-center gap-2 p-2 glass-subtle rounded-lg">
                      <button onClick={() => toggleActionMutation.mutate({ meetingId: String(selectedMeeting.id), actionId: String(item.id) })}
                        className={`size-5 rounded flex items-center justify-center border ${!!item.completed ? 'bg-green-500/30 border-green-500' : 'border-white/20'}`}>
                        {!!item.completed && <MaterialIcon name="check" className="text-xs text-green-400" />}
                      </button>
                      <span className={`text-sm flex-1 ${!!item.completed ? 'line-through text-muted-foreground' : ''}`}>{String(item.title || '')}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => deleteMutation.mutate(String(selectedMeeting.id))} className="px-4 py-2 text-sm text-red-400 glass-subtle rounded-xl hover:bg-red-500/10">
                {t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ミーティングリスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : meetings.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="groups" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {meetings.map((m) => {
            const status = String(m.status || 'scheduled');
            return (
              <div key={String(m.id)} onClick={() => setSelectedMeeting(m)}
                className="glass-card rounded-2xl p-4 cursor-pointer hover:bg-white/10 transition-all">
                <div className="flex items-center gap-3">
                  <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400 font-bold shrink-0">
                    {String(m.employee_name || '?').substring(0, 1)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="font-semibold text-sm truncate">{String(m.employee_name || '')}</p>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${STATUS_STYLES[status] || STATUS_STYLES.scheduled}`}>
                        {t(`hr.oneOnOne.statuses.${status}`)}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground truncate">
                      {String(m.scheduled_date || '').substring(0, 16)} · {t(`hr.oneOnOne.frequencies.${String(m.frequency || 'biweekly')}`)}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    {!!m.mood && <MaterialIcon name={MOOD_ICONS[String(m.mood)] || 'sentiment_neutral'} className="text-xl text-amber-400" />}
                    {Array.isArray(m.action_items) && (
                      <span className="text-[10px] text-muted-foreground">
                        {(m.action_items as Record<string, unknown>[]).filter(a => !!a.completed).length}/{(m.action_items as Record<string, unknown>[]).length}
                      </span>
                    )}
                    <MaterialIcon name="chevron_right" className="text-sm text-muted-foreground" />
                  </div>
                </div>
                {!!m.agenda && (
                  <p className="text-xs text-muted-foreground mt-2 ml-[52px] line-clamp-1">{String(m.agenda)}</p>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
