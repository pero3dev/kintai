import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function HREmployeeDetailPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { employeeId } = useParams({ strict: false }) as { employeeId: string };
  const [activeTab, setActiveTab] = useState('profile');
  const [editing, setEditing] = useState(false);

  const { data: employee, isLoading } = useQuery({
    queryKey: ['hr-employee', employeeId],
    queryFn: () => api.hr.getEmployee(employeeId),
    enabled: !!employeeId,
  });

  const { data: goalsData } = useQuery({
    queryKey: ['hr-goals', employeeId],
    queryFn: () => api.hr.getGoals({ employee_id: employeeId }),
    enabled: activeTab === 'goals',
  });

  const { data: docsData } = useQuery({
    queryKey: ['hr-documents', employeeId],
    queryFn: () => api.hr.getDocuments({ employee_id: employeeId }),
    enabled: activeTab === 'documents',
  });

  const updateMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.updateEmployee(employeeId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-employee', employeeId] });
      setEditing(false);
    },
  });

  const emp = employee?.data || employee || {};
  const goals: Record<string, unknown>[] = goalsData?.data || goalsData || [];
  const docs: Record<string, unknown>[] = docsData?.data || docsData || [];

  const [form, setForm] = useState<Record<string, string>>({});

  const startEdit = () => {
    setForm({
      first_name: emp.first_name || '',
      last_name: emp.last_name || '',
      email: emp.email || '',
      phone: emp.phone || '',
      position: emp.position || '',
      address: emp.address || '',
    });
    setEditing(true);
  };

  const tabs = [
    { id: 'profile', icon: 'person', label: t('hr.detail.tabs.profile') },
    { id: 'goals', icon: 'flag', label: t('hr.detail.tabs.goals') },
    { id: 'documents', icon: 'description', label: t('hr.detail.tabs.documents') },
    { id: 'history', icon: 'history', label: t('hr.detail.tabs.history') },
  ];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 animate-fade-in">
        <div className="text-center text-muted-foreground">
          <MaterialIcon name="hourglass_empty" className="text-4xl mb-2 block animate-spin" />
          <p>{t('common.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/hr/employees" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div className="size-12 sm:size-14 rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-bold text-lg sm:text-xl">
            {(emp.last_name as string || '?')[0]}
          </div>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">
              {emp.last_name as string} {emp.first_name as string}
            </h1>
            <p className="text-xs sm:text-sm text-muted-foreground">
              {emp.position as string || '-'} · {emp.department_name as string || '-'}
            </p>
          </div>
        </div>
        <button
          onClick={startEdit}
          className="flex items-center gap-2 px-4 py-2 glass-subtle hover:bg-white/10 rounded-xl text-sm transition-all border border-white/5"
        >
          <MaterialIcon name="edit" className="text-base" />
          {t('hr.detail.editProfile')}
        </button>
      </div>

      {/* タブ */}
      <div className="flex gap-1 glass-card rounded-2xl p-1.5 overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium whitespace-nowrap transition-all ${
              activeTab === tab.id
                ? 'bg-indigo-500/20 text-indigo-400'
                : 'text-muted-foreground hover:bg-white/5'
            }`}
          >
            <MaterialIcon name={tab.icon} className="text-base" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* プロフィールタブ */}
      {activeTab === 'profile' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 基本情報 */}
          <div className="glass-card rounded-2xl p-4 sm:p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <MaterialIcon name="badge" className="text-indigo-400" />
              {t('hr.detail.basicInfo')}
            </h2>
            <div className="space-y-3">
              {[
                { label: t('hr.employees.employeeId'), value: emp.employee_id },
                { label: t('users.name'), value: `${emp.last_name || ''} ${emp.first_name || ''}` },
                { label: t('common.email'), value: emp.email },
                { label: t('hr.employees.birthDate'), value: emp.birth_date },
                { label: t('hr.employees.gender'), value: emp.gender ? t(`hr.employees.genders.${emp.gender}`) : '-' },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between py-2 border-b border-white/5 last:border-0">
                  <span className="text-sm text-muted-foreground">{item.label}</span>
                  <span className="text-sm font-medium">{(item.value as string) || '-'}</span>
                </div>
              ))}
            </div>
          </div>

          {/* 雇用情報 */}
          <div className="glass-card rounded-2xl p-4 sm:p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <MaterialIcon name="work" className="text-emerald-400" />
              {t('hr.detail.employment')}
            </h2>
            <div className="space-y-3">
              {[
                { label: t('hr.employees.department'), value: emp.department_name },
                { label: t('hr.employees.position'), value: emp.position },
                { label: t('hr.employees.employmentType'), value: emp.employment_type ? t(`hr.employees.types.${emp.employment_type}`) : '-' },
                { label: t('hr.employees.hireDate'), value: emp.hire_date },
                { label: t('common.status'), value: emp.status ? t(`hr.employees.statuses.${emp.status}`) : '-' },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between py-2 border-b border-white/5 last:border-0">
                  <span className="text-sm text-muted-foreground">{item.label}</span>
                  <span className="text-sm font-medium">{(item.value as string) || '-'}</span>
                </div>
              ))}
            </div>
          </div>

          {/* 連絡先 */}
          <div className="glass-card rounded-2xl p-4 sm:p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <MaterialIcon name="contact_phone" className="text-blue-400" />
              {t('hr.detail.contactInfo')}
            </h2>
            <div className="space-y-3">
              {[
                { label: t('hr.employees.phoneNumber'), value: emp.phone },
                { label: t('hr.employees.address'), value: emp.address },
                { label: t('hr.employees.emergencyContact'), value: emp.emergency_contact },
              ].map((item) => (
                <div key={item.label} className="flex items-center justify-between py-2 border-b border-white/5 last:border-0">
                  <span className="text-sm text-muted-foreground">{item.label}</span>
                  <span className="text-sm font-medium text-right max-w-[60%]">{(item.value as string) || '-'}</span>
                </div>
              ))}
            </div>
          </div>

          {/* スキル */}
          <div className="glass-card rounded-2xl p-4 sm:p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <MaterialIcon name="psychology" className="text-amber-400" />
              {t('hr.detail.skills')}
            </h2>
            {emp.skills && (emp.skills as string[]).length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {(emp.skills as string[]).map((skill: string, i: number) => (
                  <span key={i} className="px-3 py-1.5 bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 rounded-full text-xs font-medium">
                    {skill}
                  </span>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">{t('common.noData')}</p>
            )}
          </div>
        </div>
      )}

      {/* 目標タブ */}
      {activeTab === 'goals' && (
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="flag" className="text-indigo-400" />
            {t('hr.goals.title')}
          </h2>
          {goals.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground text-sm">{t('common.noData')}</p>
          ) : (
            <div className="space-y-3">
              {goals.map((goal) => (
                <div key={goal.id as string} className="glass-subtle rounded-xl p-4">
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <p className="font-semibold text-sm">{goal.title as string}</p>
                    <span className="text-xs text-muted-foreground whitespace-nowrap">{goal.due_date as string}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="flex-1 h-2 rounded-full bg-white/5 overflow-hidden">
                      <div className="h-full rounded-full gradient-primary" style={{ width: `${Number(goal.progress) || 0}%` }} />
                    </div>
                    <span className="text-sm font-bold text-indigo-400">{Number(goal.progress) || 0}%</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 書類タブ */}
      {activeTab === 'documents' && (
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="description" className="text-indigo-400" />
            {t('hr.documents.title')}
          </h2>
          {docs.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground text-sm">{t('common.noData')}</p>
          ) : (
            <div className="space-y-3">
              {docs.map((doc) => (
                <div key={doc.id as string} className="flex items-center justify-between glass-subtle rounded-xl p-3">
                  <div className="flex items-center gap-3">
                    <MaterialIcon name="description" className="text-indigo-400" />
                    <div>
                      <p className="text-sm font-medium">{doc.name as string}</p>
                      <p className="text-xs text-muted-foreground">{doc.type as string} · {doc.upload_date as string}</p>
                    </div>
                  </div>
                  <button className="p-2 hover:bg-white/10 rounded-lg transition-colors">
                    <MaterialIcon name="download" className="text-sm" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 履歴タブ */}
      {activeTab === 'history' && (
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="history" className="text-indigo-400" />
            {t('hr.detail.history')}
          </h2>
          <p className="text-center py-8 text-muted-foreground text-sm">{t('common.noData')}</p>
        </div>
      )}

      {/* 編集モーダル */}
      {editing && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.detail.editProfile')}</h2>
              <button onClick={() => setEditing(false)} className="p-1 hover:bg-white/10 rounded-lg transition-colors">
                <MaterialIcon name="close" className="text-xl" />
              </button>
            </div>
            <div className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.lastName')}</label>
                  <input value={form.last_name || ''} onChange={(e) => setForm(f => ({...f, last_name: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.firstName')}</label>
                  <input value={form.first_name || ''} onChange={(e) => setForm(f => ({...f, first_name: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('common.email')}</label>
                <input type="email" value={form.email || ''} onChange={(e) => setForm(f => ({...f, email: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.employees.phoneNumber')}</label>
                <input value={form.phone || ''} onChange={(e) => setForm(f => ({...f, phone: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.employees.position')}</label>
                <input value={form.position || ''} onChange={(e) => setForm(f => ({...f, position: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.employees.address')}</label>
                <input value={form.address || ''} onChange={(e) => setForm(f => ({...f, address: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => updateMutation.mutate(form)}
                disabled={updateMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50">
                {t('common.save')}
              </button>
              <button onClick={() => setEditing(false)}
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
