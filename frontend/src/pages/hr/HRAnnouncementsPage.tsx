import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const PRIORITY_STYLES: Record<string, { icon: string; color: string }> = {
  urgent: { icon: 'error', color: 'text-red-400 bg-red-500/20' },
  high: { icon: 'priority_high', color: 'text-amber-400 bg-amber-500/20' },
  normal: { icon: 'info', color: 'text-blue-400 bg-blue-500/20' },
  low: { icon: 'arrow_downward', color: 'text-gray-400 bg-gray-500/20' },
};

const TARGET_ICONS: Record<string, string> = {
  all: 'groups',
  department: 'corporate_fare',
  role: 'admin_panel_settings',
  individual: 'person',
};

export function HRAnnouncementsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [filters, setFilters] = useState({ priority: '' });

  const [form, setForm] = useState({
    title: '', content: '', priority: 'normal', target: 'all',
    target_value: '', is_pinned: false,
  });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-announcements', filters],
    queryFn: () => api.hr.getAnnouncements(filters),
  });

  const announcements: Record<string, unknown>[] = data?.data || data || [];

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createAnnouncement(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-announcements'] }); setShowForm(false); resetForm(); },
  });
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) => api.hr.updateAnnouncement(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-announcements'] }); setShowForm(false); setEditingId(null); resetForm(); },
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.hr.deleteAnnouncement(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hr-announcements'] }),
  });

  const resetForm = () => setForm({ title: '', content: '', priority: 'normal', target: 'all', target_value: '', is_pinned: false });

  const startEdit = (ann: Record<string, unknown>) => {
    setForm({
      title: ann.title as string || '',
      content: ann.content as string || '',
      priority: ann.priority as string || 'normal',
      target: ann.target as string || 'all',
      target_value: ann.target_value as string || '',
      is_pinned: ann.is_pinned as boolean || false,
    });
    setEditingId(ann.id as string);
    setShowForm(true);
  };

  const handleSubmit = () => {
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: { ...form } });
    } else {
      createMutation.mutate({ ...form });
    }
  };

  // ピン留めを先頭に
  const sorted = [...announcements].sort((a, b) => {
    if (a.is_pinned && !b.is_pinned) return -1;
    if (!a.is_pinned && b.is_pinned) return 1;
    return 0;
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.announcements.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.announcements.subtitle')}</p>
          </div>
        </div>
        <button onClick={() => { resetForm(); setEditingId(null); setShowForm(true); }}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
          <MaterialIcon name="add" className="text-lg" />
          {t('hr.announcements.addAnnouncement')}
        </button>
      </div>

      {/* フィルター */}
      <div className="flex flex-wrap gap-3">
        <select value={filters.priority} onChange={(e) => setFilters(f => ({...f, priority: e.target.value}))}
          className="px-4 py-2.5 glass-input rounded-xl text-sm min-w-[140px]">
          <option value="">{t('hr.announcements.priority')}</option>
          {['urgent', 'high', 'normal', 'low'].map(p => (
            <option key={p} value={p}>{t(`hr.announcements.priorities.${p}`)}</option>
          ))}
        </select>
      </div>

      {/* お知らせリスト */}
      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : sorted.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="campaign" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {sorted.map((ann) => {
            const priority = (ann.priority as string) || 'normal';
            const prioStyle = PRIORITY_STYLES[priority] || PRIORITY_STYLES.normal;
            const target = (ann.target as string) || 'all';
            const isPinned = ann.is_pinned as boolean;

            return (
              <div key={ann.id as string}
                className={`glass-card rounded-2xl p-4 sm:p-5 hover:bg-white/5 transition-all ${isPinned ? 'ring-1 ring-amber-400/30' : ''}`}>
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                  <div className="flex items-start gap-3 flex-1">
                    <div className={`size-10 rounded-xl flex items-center justify-center shrink-0 ${prioStyle.color.split(' ')[1]}`}>
                      <MaterialIcon name={prioStyle.icon} className={prioStyle.color.split(' ')[0]} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1 flex-wrap">
                        {isPinned && (
                          <MaterialIcon name="push_pin" className="text-amber-400 text-sm" />
                        )}
                        <h3 className="font-semibold text-sm">{ann.title as string}</h3>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${prioStyle.color}`}>
                          {t(`hr.announcements.priorities.${priority}`)}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground line-clamp-3 mb-2">{ann.content as string}</p>
                      <div className="flex items-center gap-3 text-[10px] text-muted-foreground flex-wrap">
                        <span className="flex items-center gap-1">
                          <MaterialIcon name={TARGET_ICONS[target] || 'groups'} className="text-sm" />
                          {t(`hr.announcements.targets.${target}`)}
                          {!!ann.target_value && `: ${String(ann.target_value)}`}
                        </span>
                        <span className="flex items-center gap-1">
                          <MaterialIcon name="person" className="text-sm" />
                          {ann.author_name as string || '-'}
                        </span>
                        <span className="flex items-center gap-1">
                          <MaterialIcon name="schedule" className="text-sm" />
                          {ann.created_at as string || ''}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <button onClick={() => startEdit(ann)}
                      className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
                      <MaterialIcon name="edit" className="text-sm" />
                    </button>
                    <button onClick={() => { if (confirm(t('common.confirm'))) deleteMutation.mutate(ann.id as string); }}
                      className="p-1.5 hover:bg-red-500/10 text-red-400 rounded-lg transition-colors">
                      <MaterialIcon name="delete" className="text-sm" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* 作成/編集モーダル */}
      {showForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">
                {editingId ? t('hr.announcements.editAnnouncement') : t('hr.announcements.addAnnouncement')}
              </h2>
              <button onClick={() => { setShowForm(false); setEditingId(null); resetForm(); }}
                className="p-1 hover:bg-white/10 rounded-lg"><MaterialIcon name="close" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.announcements.announcementTitle')}</label>
                <input value={form.title} onChange={(e) => setForm(f => ({...f, title: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.announcements.content')}</label>
                <textarea value={form.content} onChange={(e) => setForm(f => ({...f, content: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={6} />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.announcements.priority')}</label>
                  <select value={form.priority} onChange={(e) => setForm(f => ({...f, priority: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    {['urgent', 'high', 'normal', 'low'].map(p => (
                      <option key={p} value={p}>{t(`hr.announcements.priorities.${p}`)}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.announcements.target')}</label>
                  <select value={form.target} onChange={(e) => setForm(f => ({...f, target: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    {['all', 'department', 'role', 'individual'].map(tg => (
                      <option key={tg} value={tg}>{t(`hr.announcements.targets.${tg}`)}</option>
                    ))}
                  </select>
                </div>
              </div>
              {form.target !== 'all' && (
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.announcements.targetValue')}</label>
                  <input value={form.target_value} onChange={(e) => setForm(f => ({...f, target_value: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              )}
              <label className="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" checked={form.is_pinned}
                  onChange={(e) => setForm(f => ({...f, is_pinned: e.target.checked}))}
                  className="rounded border-white/20" />
                <span className="text-sm">{t('hr.announcements.pinned')}</span>
              </label>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={handleSubmit}
                disabled={!form.title || !form.content || createMutation.isPending || updateMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {editingId ? t('common.save') : t('common.create')}
              </button>
              <button onClick={() => { setShowForm(false); setEditingId(null); resetForm(); }}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">{t('common.cancel')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
