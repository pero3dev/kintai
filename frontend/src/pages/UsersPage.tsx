import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { Users, X, Save, Plus, Trash2, UserPlus } from 'lucide-react';
import { Pagination } from '@/components/ui/Pagination';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: 'admin' | 'manager' | 'employee';
  is_active: boolean;
  department_id?: string;
}

interface Department {
  id: string;
  name: string;
}

interface CreateUserForm {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  role: 'admin' | 'manager' | 'employee';
  department_id?: string;
}

export function UsersPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [formData, setFormData] = useState<Partial<User & { password?: string }>>({});
  const [createData, setCreateData] = useState<CreateUserForm>({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    role: 'employee',
  });
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const { data: usersResponse } = useQuery({
    queryKey: ['users', page, pageSize],
    queryFn: () => api.users.getAll({ page, page_size: pageSize }),
  });

  const { data: departments = [] } = useQuery<Department[]>({
    queryKey: ['departments'],
    queryFn: () => api.departments.getAll(),
  });

  const users: User[] = usersResponse?.data || [];

  const createMutation = useMutation({
    mutationFn: (data: CreateUserForm) => api.users.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setShowCreateModal(false);
      setCreateData({ email: '', password: '', first_name: '', last_name: '', role: 'employee' });
      setError(null);
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) =>
      api.users.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditingUser(null);
      setError(null);
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.users.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const roleBadge = (role: string) => {
    const styles: Record<string, string> = {
      admin: 'bg-purple-500/20 text-purple-400 border border-purple-500/30',
      manager: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
      employee: 'bg-muted text-muted-foreground border border-border',
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs ${styles[role] || ''}`}>
        {t(`users.roles.${role}`)}
      </span>
    );
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setFormData({
      first_name: user.first_name,
      last_name: user.last_name,
      role: user.role,
      is_active: user.is_active,
      department_id: user.department_id,
    });
    setError(null);
  };

  const handleSave = () => {
    if (!editingUser) return;
    updateMutation.mutate({
      id: editingUser.id,
      data: formData,
    });
  };

  const handleCreate = () => {
    if (!createData.email || !createData.password || !createData.first_name || !createData.last_name) {
      setError(t('users.requiredFieldsError'));
      return;
    }
    createMutation.mutate(createData);
  };

  const handleDelete = (user: User) => {
    if (confirm(t('users.deleteConfirm', { name: `${user.last_name} ${user.first_name}` }))) {
      deleteMutation.mutate(user.id);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Users className="h-6 w-6" />
          {t('users.title')}
        </h1>
        <button
          onClick={() => { setShowCreateModal(true); setError(null); }}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90"
        >
          <UserPlus className="h-4 w-4" />
          {t('users.addNew')}
        </button>
      </div>

      <div className="bg-card border border-border rounded-lg p-6">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 px-4">{t('users.name')}</th>
                <th className="text-left py-2 px-4">{t('common.email')}</th>
                <th className="text-left py-2 px-4">{t('users.role')}</th>
                <th className="text-left py-2 px-4">{t('users.status')}</th>
                <th className="text-right py-2 px-4">{t('users.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-b border-border/50 hover:bg-accent/50">
                  <td className="py-2 px-4">{user.last_name} {user.first_name}</td>
                  <td className="py-2 px-4">{user.email}</td>
                  <td className="py-2 px-4">{roleBadge(user.role)}</td>
                  <td className="py-2 px-4">
                    <span className={`px-2 py-1 rounded-full text-xs ${user.is_active ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                      {user.is_active ? t('users.active') : t('users.inactive')}
                    </span>
                  </td>
                  <td className="text-right py-2 px-4">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handleEdit(user)}
                        className="text-sm text-primary hover:text-primary/80"
                      >
                        {t('common.edit')}
                      </button>
                      <button
                        onClick={() => handleDelete(user)}
                        className="text-sm text-red-600 hover:text-red-800"
                        title={t('common.delete')}
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {usersResponse?.total_pages > 0 && (
          <Pagination
            currentPage={page}
            totalPages={usersResponse.total_pages}
            totalItems={usersResponse.total}
            pageSize={pageSize}
            onPageChange={(p) => setPage(p)}
            onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
          />
        )}
      </div>

      {/* 新規作成モーダル */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold flex items-center gap-2">
                <UserPlus className="h-5 w-5" />
                {t('users.createNew')}
              </h2>
              <button onClick={() => setShowCreateModal(false)} className="p-1 hover:bg-accent rounded">
                <X className="h-5 w-5" />
              </button>
            </div>

            {error && (
              <div className="mb-4 p-3 bg-red-100 text-red-800 rounded-lg text-sm">
                {error}
              </div>
            )}

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('common.email')} *</label>
                <input
                  type="email"
                  value={createData.email}
                  onChange={(e) => setCreateData({ ...createData, email: e.target.value })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="user@example.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('common.password')} *</label>
                <input
                  type="password"
                  value={createData.password}
                  onChange={(e) => setCreateData({ ...createData, password: e.target.value })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder={t('users.passwordRequirement')}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.lastName')} *</label>
                  <input
                    type="text"
                    value={createData.last_name}
                    onChange={(e) => setCreateData({ ...createData, last_name: e.target.value })}
                    className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.firstName')} *</label>
                  <input
                    type="text"
                    value={createData.first_name}
                    onChange={(e) => setCreateData({ ...createData, first_name: e.target.value })}
                    className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('users.role')}</label>
                <select
                  value={createData.role}
                  onChange={(e) => setCreateData({ ...createData, role: e.target.value as User['role'] })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="employee">{t('users.roles.employee')}</option>
                  <option value="manager">{t('users.roles.manager')}</option>
                  <option value="admin">{t('users.roles.admin')}</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('users.department')}</label>
                <select
                  value={createData.department_id || ''}
                  onChange={(e) => setCreateData({ ...createData, department_id: e.target.value || undefined })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="">{t('users.unassigned')}</option>
                  {departments.map((dept) => (
                    <option key={dept.id} value={dept.id}>{dept.name}</option>
                  ))}
                </select>
              </div>

              <div className="flex gap-2 pt-4">
                <button
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-2 border border-border rounded-lg hover:bg-accent"
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={handleCreate}
                  disabled={createMutation.isPending}
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  <Plus className="h-4 w-4" />
                  {t('common.create')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 編集モーダル */}
      {editingUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">{t('users.editUser')}</h2>
              <button onClick={() => setEditingUser(null)} className="p-1 hover:bg-accent rounded">
                <X className="h-5 w-5" />
              </button>
            </div>

            {error && (
              <div className="mb-4 p-3 bg-red-100 text-red-800 rounded-lg text-sm">
                {error}
              </div>
            )}

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('common.email')}</label>
                <div className="text-muted-foreground">{editingUser.email}</div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.lastName')}</label>
                  <input
                    type="text"
                    value={formData.last_name || ''}
                    onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
                    className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">{t('users.firstName')}</label>
                  <input
                    type="text"
                    value={formData.first_name || ''}
                    onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
                    className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('users.newPassword')}</label>
                <input
                  type="password"
                  value={formData.password || ''}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder={t('users.passwordPlaceholder')}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('users.role')}</label>
                <select
                  value={formData.role || 'employee'}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value as User['role'] })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="employee">{t('users.roles.employee')}</option>
                  <option value="manager">{t('users.roles.manager')}</option>
                  <option value="admin">{t('users.roles.admin')}</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('users.department')}</label>
                <select
                  value={formData.department_id || ''}
                  onChange={(e) => setFormData({ ...formData, department_id: e.target.value || undefined })}
                  className="w-full px-3 py-2 border border-border rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="">{t('users.unassigned')}</option>
                  {departments.map((dept) => (
                    <option key={dept.id} value={dept.id}>{dept.name}</option>
                  ))}
                </select>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.is_active ?? true}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  className="rounded border-border"
                />
                <label htmlFor="is_active" className="text-sm font-medium">{t('users.active')}</label>
              </div>

              <div className="flex gap-2 pt-4">
                <button
                  onClick={() => setEditingUser(null)}
                  className="flex-1 px-4 py-2 border border-border rounded-lg hover:bg-accent"
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={handleSave}
                  disabled={updateMutation.isPending}
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  <Save className="h-4 w-4" />
                  {t('common.save')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
