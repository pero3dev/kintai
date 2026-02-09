import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function HRDepartmentsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [viewMode, setViewMode] = useState<'list' | 'chart'>('list');
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  const [form, setForm] = useState({
    name: '', code: '', description: '', manager_id: '', parent_id: '', budget: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-departments'],
    queryFn: () => api.hr.getDepartments(),
  });

  const departments: Record<string, unknown>[] = data?.data || data || [];

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createDepartment(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-departments'] }); setShowForm(false); resetForm(); },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) => api.hr.updateDepartment(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-departments'] }); setShowForm(false); setEditingId(null); resetForm(); },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.hr.deleteDepartment(id),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['hr-departments'] }); },
  });

  const resetForm = () => setForm({ name: '', code: '', description: '', manager_id: '', parent_id: '', budget: '' });

  const startEdit = (dept: Record<string, unknown>) => {
    setForm({
      name: dept.name as string || '',
      code: dept.code as string || '',
      description: dept.description as string || '',
      manager_id: dept.manager_id as string || '',
      parent_id: dept.parent_id as string || '',
      budget: dept.budget ? String(dept.budget) : '',
    });
    setEditingId(dept.id as string);
    setShowForm(true);
  };

  const handleSubmit = () => {
    const data = { ...form, budget: form.budget ? Number(form.budget) : undefined };
    if (editingId) {
      updateMutation.mutate({ id: editingId, data });
    } else {
      createMutation.mutate(data);
    }
  };

  // 組織図用のツリー構造を構築
  const rootDepts = departments.filter(d => !d.parent_id);
  const getChildren = (parentId: string) => departments.filter(d => d.parent_id === parentId);

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.departments.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.departments.subtitle')}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex rounded-xl overflow-hidden border border-white/10">
            <button onClick={() => setViewMode('list')}
              className={`px-3 py-2 text-xs font-medium transition-all ${viewMode === 'list' ? 'bg-indigo-500/20 text-indigo-400' : 'text-muted-foreground hover:bg-white/5'}`}>
              <MaterialIcon name="list" className="text-base" />
            </button>
            <button onClick={() => setViewMode('chart')}
              className={`px-3 py-2 text-xs font-medium transition-all ${viewMode === 'chart' ? 'bg-indigo-500/20 text-indigo-400' : 'text-muted-foreground hover:bg-white/5'}`}>
              <MaterialIcon name="account_tree" className="text-base" />
            </button>
          </div>
          <button onClick={() => { resetForm(); setEditingId(null); setShowForm(true); }}
            className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all">
            <MaterialIcon name="add" className="text-lg" />
            {t('hr.departments.addDepartment')}
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
      ) : departments.length === 0 ? (
        <div className="glass-card rounded-2xl p-12 text-center text-muted-foreground">
          <MaterialIcon name="corporate_fare" className="text-4xl mb-2 block opacity-50" />
          <p className="text-sm">{t('common.noData')}</p>
        </div>
      ) : viewMode === 'list' ? (
        /* リストビュー */
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {departments.map((dept) => (
            <div key={dept.id as string} className="glass-card rounded-2xl p-5 hover:bg-white/5 transition-all group">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div className="size-10 rounded-xl bg-purple-500/20 flex items-center justify-center">
                    <MaterialIcon name="corporate_fare" className="text-purple-400" />
                  </div>
                  <div>
                    <p className="font-semibold text-sm">{String(dept.name)}</p>
                    <p className="text-xs text-muted-foreground">{String(dept.code || '')}</p>
                  </div>
                </div>
                <div className="flex items-center gap-1 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
                  <button onClick={() => startEdit(dept)} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
                    <MaterialIcon name="edit" className="text-sm" />
                  </button>
                  <button onClick={() => { if (confirm(t('common.confirm'))) deleteMutation.mutate(dept.id as string); }}
                    className="p-1.5 hover:bg-red-500/10 text-red-400 rounded-lg transition-colors">
                    <MaterialIcon name="delete" className="text-sm" />
                  </button>
                </div>
              </div>
              {!!dept.description && (
                <p className="text-xs text-muted-foreground mb-3 line-clamp-2">{String(dept.description)}</p>
              )}
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div className="glass-subtle rounded-lg p-2 text-center">
                  <p className="text-muted-foreground">{t('hr.departments.memberCount')}</p>
                  <p className="font-bold text-base">{String(dept.member_count || 0)}</p>
                </div>
                <div className="glass-subtle rounded-lg p-2 text-center">
                  <p className="text-muted-foreground">{t('hr.departments.manager')}</p>
                  <p className="font-medium text-xs truncate">{String(dept.manager_name || '-')}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* 組織図ビュー */
        <div className="glass-card rounded-2xl p-4 sm:p-6 overflow-x-auto">
          <div className="min-w-[300px]">
            {rootDepts.map((dept) => (
              <OrgNode key={dept.id as string} dept={dept} getChildren={getChildren} level={0} t={t} />
            ))}
          </div>
        </div>
      )}

      {/* 作成/編集モーダル */}
      {showForm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-md max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">
                {editingId ? t('hr.departments.editDepartment') : t('hr.departments.addDepartment')}
              </h2>
              <button onClick={() => { setShowForm(false); setEditingId(null); resetForm(); }}
                className="p-1 hover:bg-white/10 rounded-lg transition-colors">
                <MaterialIcon name="close" className="text-xl" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.departments.departmentName')}</label>
                <input value={form.name} onChange={(e) => setForm(f => ({...f, name: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.departments.code')}</label>
                <input value={form.code} onChange={(e) => setForm(f => ({...f, code: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" placeholder="DEPT-001" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.departments.description')}</label>
                <textarea value={form.description} onChange={(e) => setForm(f => ({...f, description: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" rows={3} />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.departments.parentDepartment')}</label>
                <select value={form.parent_id} onChange={(e) => setForm(f => ({...f, parent_id: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                  <option value="">({t('common.selectPlaceholder')})</option>
                  {departments.filter(d => d.id !== editingId).map((d) => (
                    <option key={d.id as string} value={d.id as string}>{String(d.name)}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.departments.budget')}</label>
                <input type="number" value={form.budget} onChange={(e) => setForm(f => ({...f, budget: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" placeholder="¥0" />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={handleSubmit}
                disabled={!form.name || createMutation.isPending || updateMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {editingId ? t('common.save') : t('common.create')}
              </button>
              <button onClick={() => { setShowForm(false); setEditingId(null); resetForm(); }}
                className="px-6 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">
                {t('common.cancel')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function OrgNode({ dept, getChildren, level, t }: {
  dept: Record<string, unknown>;
  getChildren: (id: string) => Record<string, unknown>[];
  level: number;
  t: (key: string) => string;
}) {
  const children = getChildren(dept.id as string);
  return (
    <div className={`${level > 0 ? 'ml-6 sm:ml-10 border-l-2 border-white/10 pl-4 sm:pl-6' : ''} mb-3`}>
      <div className="glass-subtle rounded-xl p-3 sm:p-4 inline-block">
        <div className="flex items-center gap-3">
          <div className="size-8 rounded-lg bg-purple-500/20 flex items-center justify-center">
            <MaterialIcon name="corporate_fare" className="text-purple-400 text-sm" />
          </div>
          <div>
            <p className="font-semibold text-sm">{String(dept.name)}</p>
            <p className="text-xs text-muted-foreground">
              {String(dept.manager_name || t('hr.departments.manager') + ': -')} · {String(dept.member_count || 0)}{t('hr.departments.memberCount').replace('所属人数', '名')}
            </p>
          </div>
        </div>
      </div>
      {children.map((child) => (
        <OrgNode key={child.id as string} dept={child} getChildren={getChildren} level={level + 1} t={t} />
      ))}
    </div>
  );
}
