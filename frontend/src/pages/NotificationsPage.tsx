import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { Bell, Check, CheckCheck, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { Pagination } from '@/components/ui/Pagination';

export function NotificationsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const { data: notifications } = useQuery({
    queryKey: ['notifications', page, pageSize],
    queryFn: () => api.notifications.getList({ page, page_size: pageSize }),
  });

  const { data: unreadCount } = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () => api.notifications.getUnreadCount(),
  });

  const markReadMutation = useMutation({
    mutationFn: (id: string) => api.notifications.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: () => api.notifications.markAllAsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.notifications.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const typeIcon = (type: string) => {
    const icons: Record<string, string> = {
      leave_approved: '✅',
      leave_rejected: '❌',
      overtime_approved: '✅',
      overtime_rejected: '❌',
      correction_result: '📝',
      shift_assigned: '📅',
      general: '🔔',
    };
    return icons[type] || '🔔';
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Bell className="h-6 w-6" />
          {t('notifications.title')}
          {(unreadCount?.unread ?? 0) > 0 && (
            <span className="ml-2 px-2 py-1 bg-destructive text-destructive-foreground rounded-full text-xs">
              {unreadCount.unread}
            </span>
          )}
        </h1>
        {(unreadCount?.unread ?? 0) > 0 && (
          <button
            onClick={() => markAllReadMutation.mutate()}
            className="flex items-center gap-2 px-4 py-2 glass-input rounded-xl hover:bg-white/10 transition-all text-sm"
          >
            <CheckCheck className="h-4 w-4" />
            {t('notifications.markAllRead')}
          </button>
        )}
      </div>

      <div className="bg-card border border-border rounded-lg divide-y divide-border">
        {notifications?.data?.map((n: Record<string, unknown>) => (
          <div
            key={n.id as string}
            className={`p-4 flex items-start gap-4 ${!n.is_read ? 'bg-primary/5' : ''}`}
          >
            <span className="text-2xl">{typeIcon(n.type as string)}</span>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className={`font-medium ${!n.is_read ? 'text-foreground' : 'text-muted-foreground'}`}>
                  {n.title as string}
                </p>
                {!n.is_read && (
                  <span className="size-2 bg-primary rounded-full flex-shrink-0" />
                )}
              </div>
              <p className="text-sm text-muted-foreground mt-1">{n.message as string}</p>
              <p className="text-xs text-muted-foreground mt-2">
                {new Date(n.created_at as string).toLocaleString('ja-JP')}
              </p>
            </div>
            <div className="flex gap-1 flex-shrink-0">
              {!n.is_read && (
                <button
                  onClick={() => markReadMutation.mutate(n.id as string)}
                  className="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded"
                  title={t('common.markAsRead')}
                >
                  <Check className="h-4 w-4" />
                </button>
              )}
              <button
                onClick={() => deleteMutation.mutate(n.id as string)}
                className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded"
                title={t('common.delete')}
              >
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          </div>
        ))}
        {(!notifications?.data || notifications.data.length === 0) && (
          <div className="p-8 text-center text-muted-foreground">
            {t('common.noData')}
          </div>
        )}
      </div>
      {notifications?.total_pages > 0 && (
        <Pagination
          currentPage={page}
          totalPages={notifications.total_pages}
          totalItems={notifications.total}
          pageSize={pageSize}
          onPageChange={(p) => setPage(p)}
          onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
        />
      )}
    </div>
  );
}
