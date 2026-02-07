import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { Users } from 'lucide-react';

export function UsersPage() {
  const { t } = useTranslation();

  const { data: users } = useQuery({
    queryKey: ['users'],
    queryFn: () => api.users.getAll(),
  });

  const roleBadge = (role: string) => {
    const styles: Record<string, string> = {
      admin: 'bg-purple-100 text-purple-800',
      manager: 'bg-blue-100 text-blue-800',
      employee: 'bg-gray-100 text-gray-800',
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs ${styles[role] || ''}`}>
        {t(`users.roles.${role}`)}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <Users className="h-6 w-6" />
        {t('users.title')}
      </h1>

      <div className="bg-card border border-border rounded-lg p-6">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 px-4">名前</th>
                <th className="text-left py-2 px-4">{t('common.email')}</th>
                <th className="text-left py-2 px-4">役割</th>
                <th className="text-left py-2 px-4">ステータス</th>
                <th className="text-right py-2 px-4">操作</th>
              </tr>
            </thead>
            <tbody>
              {users?.data?.map((user: Record<string, unknown>) => (
                <tr key={user.id as string} className="border-b border-border/50 hover:bg-accent/50">
                  <td className="py-2 px-4">{user.last_name as string} {user.first_name as string}</td>
                  <td className="py-2 px-4">{user.email as string}</td>
                  <td className="py-2 px-4">{roleBadge(user.role as string)}</td>
                  <td className="py-2 px-4">
                    <span className={`px-2 py-1 rounded-full text-xs ${user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                      {user.is_active ? '有効' : '無効'}
                    </span>
                  </td>
                  <td className="text-right py-2 px-4">
                    <button className="text-sm text-primary hover:text-primary/80">
                      {t('common.edit')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
