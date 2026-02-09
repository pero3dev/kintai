import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const STATUS_BADGE: Record<string, string> = {
  active: 'bg-emerald-500/20 text-emerald-400',
  onLeave: 'bg-amber-500/20 text-amber-400',
  resigned: 'bg-gray-500/20 text-gray-400',
  terminated: 'bg-red-500/20 text-red-400',
};

const TYPE_BADGE: Record<string, string> = {
  fullTime: 'bg-blue-500/20 text-blue-400',
  partTime: 'bg-cyan-500/20 text-cyan-400',
  contract: 'bg-purple-500/20 text-purple-400',
  intern: 'bg-pink-500/20 text-pink-400',
  temporary: 'bg-orange-500/20 text-orange-400',
};

export function HREmployeesPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [departmentFilter, setDepartmentFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [page, setPage] = useState(1);

  // フォーム
  const [form, setForm] = useState({
    employee_id: '', first_name: '', last_name: '', email: '',
    department: '', position: '', employment_type: 'fullTime',
    hire_date: '', phone: '', status: 'active',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['hr-employees', page, search, departmentFilter, statusFilter, typeFilter],
    queryFn: () => api.hr.getEmployees({
      page, page_size: 20, search, department: departmentFilter,
      status: statusFilter, employment_type: typeFilter,
    }),
  });

  const { data: depts } = useQuery({
    queryKey: ['hr-departments'],
    queryFn: () => api.hr.getDepartments(),
  });

  const employees: Record<string, unknown>[] = data?.data || data?.employees || [];
  const departments: Record<string, unknown>[] = depts?.data || depts || [];
  const totalCount = data?.total || employees.length;

  const createMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.hr.createEmployee(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-employees'] });
      setShowAddModal(false);
      resetForm();
    },
  });

  const resetForm = () => {
    setForm({
      employee_id: '', first_name: '', last_name: '', email: '',
      department: '', position: '', employment_type: 'fullTime',
      hire_date: '', phone: '', status: 'active',
    });
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.employees.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">
              {t('hr.employees.subtitle')} · {t('hr.employees.totalCount', { count: totalCount })}
            </p>
          </div>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all"
        >
          <MaterialIcon name="person_add" className="text-lg" />
          {t('hr.employees.addEmployee')}
        </button>
      </div>

      {/* フィルター */}
      <div className="glass-card rounded-2xl p-4 flex flex-wrap gap-3">
        <div className="flex-1 min-w-[200px]">
          <div className="relative">
            <MaterialIcon name="search" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground text-lg" />
            <input
              type="text"
              value={search}
              onChange={(e) => { setSearch(e.target.value); setPage(1); }}
              placeholder={t('common.search')}
              className="w-full pl-10 pr-4 py-2.5 glass-input rounded-xl text-sm"
            />
          </div>
        </div>
        <select value={departmentFilter} onChange={(e) => { setDepartmentFilter(e.target.value); setPage(1); }}
          className="px-3 py-2.5 glass-input rounded-xl text-sm">
          <option value="">{t('hr.employees.department')}</option>
          {departments.map((d) => (
            <option key={d.id as string} value={d.id as string}>{d.name as string}</option>
          ))}
        </select>
        <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="px-3 py-2.5 glass-input rounded-xl text-sm">
          <option value="">{t('common.status')}</option>
          {['active', 'onLeave', 'resigned', 'terminated'].map((s) => (
            <option key={s} value={s}>{t(`hr.employees.statuses.${s}`)}</option>
          ))}
        </select>
        <select value={typeFilter} onChange={(e) => { setTypeFilter(e.target.value); setPage(1); }}
          className="px-3 py-2.5 glass-input rounded-xl text-sm">
          <option value="">{t('hr.employees.employmentType')}</option>
          {['fullTime', 'partTime', 'contract', 'intern', 'temporary'].map((t2) => (
            <option key={t2} value={t2}>{t(`hr.employees.types.${t2}`)}</option>
          ))}
        </select>
      </div>

      {/* 社員リスト */}
      <div className="glass-card rounded-2xl p-4 sm:p-6">
        {isLoading ? (
          <p className="text-center py-12 text-muted-foreground">{t('common.loading')}</p>
        ) : employees.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            <MaterialIcon name="person_search" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('common.noData')}</p>
          </div>
        ) : (
          <>
            {/* モバイル: カードビュー */}
            <div className="space-y-3 md:hidden">
              {employees.map((emp) => (
                <Link
                  key={emp.id as string}
                  to={'/hr/employees/$employeeId' as string}
                  params={{ employeeId: emp.id as string }}
                  className="block glass-subtle rounded-xl p-4 hover:bg-white/5 transition-all active:scale-[0.98]"
                >
                  <div className="flex items-center gap-3">
                    <div className="size-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                      {(emp.last_name as string || '?')[0]}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-semibold text-sm truncate">{emp.last_name as string} {emp.first_name as string}</p>
                      <p className="text-xs text-muted-foreground truncate">{emp.position as string || '-'} · {emp.department_name as string || '-'}</p>
                    </div>
                    <div className="flex flex-col items-end gap-1">
                      <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${STATUS_BADGE[emp.status as string] || STATUS_BADGE.active}`}>
                        {t(`hr.employees.statuses.${emp.status}`)}
                      </span>
                      <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${TYPE_BADGE[emp.employment_type as string] || TYPE_BADGE.fullTime}`}>
                        {t(`hr.employees.types.${emp.employment_type}`)}
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>

            {/* デスクトップ: テーブルビュー */}
            <div className="hidden md:block overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/5">
                    <th className="text-left py-3 px-4 font-semibold">{t('users.name')}</th>
                    <th className="text-left py-3 px-4 font-semibold">{t('hr.employees.employeeId')}</th>
                    <th className="text-left py-3 px-4 font-semibold">{t('hr.employees.department')}</th>
                    <th className="text-left py-3 px-4 font-semibold">{t('hr.employees.position')}</th>
                    <th className="text-center py-3 px-4 font-semibold">{t('hr.employees.employmentType')}</th>
                    <th className="text-center py-3 px-4 font-semibold">{t('common.status')}</th>
                    <th className="text-left py-3 px-4 font-semibold">{t('hr.employees.hireDate')}</th>
                  </tr>
                </thead>
                <tbody>
                  {employees.map((emp) => (
                    <tr key={emp.id as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="py-3 px-4">
                        <Link
                          to={'/hr/employees/$employeeId' as string}
                          params={{ employeeId: emp.id as string }}
                          className="flex items-center gap-3 text-indigo-400 hover:underline"
                        >
                          <div className="size-8 rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-bold text-xs flex-shrink-0">
                            {(emp.last_name as string || '?')[0]}
                          </div>
                          {emp.last_name as string} {emp.first_name as string}
                        </Link>
                      </td>
                      <td className="py-3 px-4 text-muted-foreground">{emp.employee_id as string || '-'}</td>
                      <td className="py-3 px-4">{emp.department_name as string || '-'}</td>
                      <td className="py-3 px-4">{emp.position as string || '-'}</td>
                      <td className="py-3 px-4 text-center">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-bold ${TYPE_BADGE[emp.employment_type as string] || TYPE_BADGE.fullTime}`}>
                          {t(`hr.employees.types.${emp.employment_type}`)}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span className={`px-2.5 py-1 rounded-full text-xs font-bold ${STATUS_BADGE[emp.status as string] || STATUS_BADGE.active}`}>
                          {t(`hr.employees.statuses.${emp.status}`)}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-muted-foreground">{emp.hire_date as string || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* ページネーション */}
            {totalCount > 20 && (
              <div className="flex items-center justify-center gap-2 mt-4 pt-4 border-t border-white/5">
                <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page <= 1}
                  className="px-3 py-1.5 glass-subtle rounded-lg text-sm disabled:opacity-50 hover:bg-white/10 transition-colors">
                  <MaterialIcon name="chevron_left" className="text-sm" />
                </button>
                <span className="text-sm text-muted-foreground">{page}</span>
                <button onClick={() => setPage(p => p + 1)} disabled={employees.length < 20}
                  className="px-3 py-1.5 glass-subtle rounded-lg text-sm disabled:opacity-50 hover:bg-white/10 transition-colors">
                  <MaterialIcon name="chevron_right" className="text-sm" />
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* 追加モーダル */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="glass-card rounded-2xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto animate-scale-in">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">{t('hr.employees.addEmployee')}</h2>
              <button onClick={() => { setShowAddModal(false); resetForm(); }} className="p-1 hover:bg-white/10 rounded-lg transition-colors">
                <MaterialIcon name="close" className="text-xl" />
              </button>
            </div>
            <div className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.employees.employeeId')}</label>
                  <input value={form.employee_id} onChange={(e) => setForm(f => ({...f, employee_id: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" placeholder="EMP-001" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('common.email')}</label>
                  <input type="email" value={form.email} onChange={(e) => setForm(f => ({...f, email: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.lastName')}</label>
                  <input value={form.last_name} onChange={(e) => setForm(f => ({...f, last_name: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.firstName')}</label>
                  <input value={form.first_name} onChange={(e) => setForm(f => ({...f, first_name: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.employees.department')}</label>
                  <select value={form.department} onChange={(e) => setForm(f => ({...f, department: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    <option value="">{t('common.selectPlaceholder')}</option>
                    {departments.map((d) => (
                      <option key={d.id as string} value={d.id as string}>{d.name as string}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.employees.position')}</label>
                  <input value={form.position} onChange={(e) => setForm(f => ({...f, position: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.employees.employmentType')}</label>
                  <select value={form.employment_type} onChange={(e) => setForm(f => ({...f, employment_type: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm">
                    {['fullTime', 'partTime', 'contract', 'intern', 'temporary'].map((t2) => (
                      <option key={t2} value={t2}>{t(`hr.employees.types.${t2}`)}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('hr.employees.hireDate')}</label>
                  <input type="date" value={form.hire_date} onChange={(e) => setForm(f => ({...f, hire_date: e.target.value}))}
                    className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('hr.employees.phoneNumber')}</label>
                <input value={form.phone} onChange={(e) => setForm(f => ({...f, phone: e.target.value}))}
                  className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => createMutation.mutate(form)}
                disabled={!form.last_name || !form.first_name || !form.email || createMutation.isPending}
                className="flex-1 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50"
              >
                {t('common.create')}
              </button>
              <button onClick={() => { setShowAddModal(false); resetForm(); }}
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
