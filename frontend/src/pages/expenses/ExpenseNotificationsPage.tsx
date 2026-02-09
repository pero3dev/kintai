import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function ExpenseNotificationsPage() {
  const { t, i18n } = useTranslation();
  useAuthStore();
  const queryClient = useQueryClient();
  const locale = i18n.language === 'ja' ? ja : enUS;
  const [filter, setFilter] = useState<'all' | 'unread' | 'action_required'>('all');

  const { data: notificationsData, isLoading } = useQuery({
    queryKey: ['expense-notifications', filter],
    queryFn: () => api.expenses.getNotifications({ filter }),
  });

  const { data: remindersData } = useQuery({
    queryKey: ['expense-reminders'],
    queryFn: () => api.expenses.getReminders(),
  });

  const notifications = notificationsData?.data || notificationsData?.notifications || [];
  const reminders = remindersData?.data || remindersData?.reminders || [];

  const markReadMutation = useMutation({
    mutationFn: (id: string) => api.expenses.markNotificationRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-notifications'] });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: () => api.expenses.markAllNotificationsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-notifications'] });
    },
  });

  const dismissReminderMutation = useMutation({
    mutationFn: (id: string) => api.expenses.dismissReminder(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-reminders'] });
    },
  });

  const getNotificationIcon = (type: string) => {
    const icons: Record<string, { icon: string; color: string; bg: string }> = {
      approved: { icon: 'check_circle', color: 'text-emerald-400', bg: 'bg-emerald-500/20' },
      rejected: { icon: 'cancel', color: 'text-red-400', bg: 'bg-red-500/20' },
      pending: { icon: 'pending_actions', color: 'text-amber-400', bg: 'bg-amber-500/20' },
      reimbursed: { icon: 'payments', color: 'text-blue-400', bg: 'bg-blue-500/20' },
      comment: { icon: 'chat', color: 'text-indigo-400', bg: 'bg-indigo-500/20' },
      reminder: { icon: 'alarm', color: 'text-amber-400', bg: 'bg-amber-500/20' },
      policy_violation: { icon: 'warning', color: 'text-red-400', bg: 'bg-red-500/20' },
      returned: { icon: 'replay', color: 'text-orange-400', bg: 'bg-orange-500/20' },
    };
    return icons[type] || { icon: 'notifications', color: 'text-indigo-400', bg: 'bg-indigo-500/20' };
  };

  const unreadCount = notifications.filter((n: Record<string, unknown>) => !n.is_read).length;

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('expenses.notifications.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">
              {unreadCount > 0
                ? t('expenses.notifications.unreadCount', { count: unreadCount })
                : t('expenses.notifications.allRead')
              }
            </p>
          </div>
        </div>
        {unreadCount > 0 && (
          <button
            onClick={() => markAllReadMutation.mutate()}
            disabled={markAllReadMutation.isPending}
            className="flex items-center gap-2 px-4 py-2 glass-subtle hover:bg-white/10 rounded-xl text-sm transition-all border border-white/5"
          >
            <MaterialIcon name="done_all" className="text-base" />
            {t('expenses.notifications.markAllRead')}
          </button>
        )}
      </div>

      {/* リマインダー */}
      {reminders.length > 0 && (
        <div className="glass-card rounded-2xl p-6 border border-amber-500/20">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2 text-amber-400">
            <MaterialIcon name="alarm" />
            {t('expenses.notifications.reminders')}
          </h2>
          <div className="space-y-3">
            {reminders.map((reminder: Record<string, unknown>) => (
              <div key={reminder.id as string} className="flex flex-col sm:flex-row sm:items-center gap-3 glass-subtle rounded-xl p-3">
                <div className="size-10 rounded-xl bg-amber-500/20 flex items-center justify-center flex-shrink-0">
                  <MaterialIcon name={
                    reminder.type === 'month_end' ? 'calendar_month' :
                    reminder.type === 'overdue' ? 'schedule' :
                    'notifications_active'
                  } className="text-amber-400" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium">{reminder.title as string}</p>
                  <p className="text-xs text-muted-foreground">{reminder.message as string}</p>
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  {!!reminder.action_url && (
                    <Link
                      to={reminder.action_url as string}
                      className="px-3 py-1.5 gradient-primary text-white text-xs font-semibold rounded-lg hover:shadow-glow-sm transition-all"
                    >
                      {t('expenses.notifications.takeAction')}
                    </Link>
                  )}
                  <button
                    onClick={() => dismissReminderMutation.mutate(reminder.id as string)}
                    className="p-1.5 rounded-lg hover:bg-white/10 text-muted-foreground transition-colors"
                  >
                    <MaterialIcon name="close" className="text-base" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* フィルター */}
      <div className="flex gap-2 flex-wrap">
        {(['all', 'unread', 'action_required'] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-4 py-2 rounded-xl text-sm font-medium transition-all ${
              filter === f
                ? 'bg-indigo-500/20 text-indigo-400 border border-indigo-500/30'
                : 'glass-subtle text-muted-foreground hover:bg-white/5'
            }`}
          >
            {t(`expenses.notifications.filter.${f}`)}
          </button>
        ))}
      </div>

      {/* 通知一覧 */}
      <div className="glass-card rounded-2xl p-6">
        {isLoading ? (
          <p className="text-center py-8 text-muted-foreground">{t('common.loading')}</p>
        ) : notifications.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            <MaterialIcon name="notifications_off" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('expenses.notifications.empty')}</p>
          </div>
        ) : (
          <div className="space-y-2">
            {notifications.map((notification: Record<string, unknown>) => {
              const { icon, color, bg } = getNotificationIcon(notification.type as string);
              const isUnread = !notification.is_read;
              return (
                <div
                  key={notification.id as string}
                  className={`flex items-start gap-3 rounded-xl p-4 transition-all cursor-pointer hover:bg-white/5 ${
                    isUnread ? 'glass-subtle border-l-2 border-l-indigo-400' : ''
                  }`}
                  onClick={() => {
                    if (isUnread) markReadMutation.mutate(notification.id as string);
                  }}
                >
                  <div className={`size-10 rounded-xl ${bg} flex items-center justify-center flex-shrink-0`}>
                    <MaterialIcon name={icon} className={color} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <p className={`text-sm ${isUnread ? 'font-semibold' : 'font-medium'}`}>
                        {notification.title as string}
                      </p>
                      {isUnread && (
                        <span className="size-2 rounded-full bg-indigo-400 flex-shrink-0" />
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground line-clamp-2">{notification.message as string}</p>
                    <p className="text-[10px] text-muted-foreground mt-1">
                      {notification.created_at ? format(new Date(notification.created_at as string), 'PPp', { locale }) : ''}
                    </p>
                  </div>
                  {!!notification.expense_id && (
                    <Link
                      to="/expenses/$expenseId"
                      params={{ expenseId: notification.expense_id as string }}
                      className="flex-shrink-0 p-2 rounded-lg hover:bg-white/10 transition-colors"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <MaterialIcon name="open_in_new" className="text-base text-muted-foreground" />
                    </Link>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* 通知設定 */}
      <NotificationSettings t={t} />
    </div>
  );
}

function NotificationSettings({ t }: { t: (key: string) => string }) {
  const queryClient = useQueryClient();

  const { data: settingsData } = useQuery({
    queryKey: ['expense-notification-settings'],
    queryFn: () => api.expenses.getNotificationSettings(),
  });

  const settings = settingsData || {
    on_approved: true,
    on_rejected: true,
    on_comment: true,
    on_reimbursed: true,
    month_end_reminder: true,
    overdue_reminder: true,
    reminder_days_before: 3,
  };

  const updateMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.expenses.updateNotificationSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-notification-settings'] });
    },
  });

  const toggleSetting = (key: string) => {
    updateMutation.mutate({ ...settings, [key]: !settings[key as keyof typeof settings] });
  };

  const settingsList = [
    { key: 'on_approved', icon: 'check_circle', label: t('expenses.notifications.settings.onApproved') },
    { key: 'on_rejected', icon: 'cancel', label: t('expenses.notifications.settings.onRejected') },
    { key: 'on_comment', icon: 'chat', label: t('expenses.notifications.settings.onComment') },
    { key: 'on_reimbursed', icon: 'payments', label: t('expenses.notifications.settings.onReimbursed') },
    { key: 'month_end_reminder', icon: 'calendar_month', label: t('expenses.notifications.settings.monthEnd') },
    { key: 'overdue_reminder', icon: 'schedule', label: t('expenses.notifications.settings.overdue') },
  ];

  return (
    <div className="glass-card rounded-2xl p-6">
      <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
        <MaterialIcon name="settings" className="text-indigo-400" />
        {t('expenses.notifications.settings.title')}
      </h2>
      <div className="space-y-2">
        {settingsList.map((s) => (
          <label key={s.key} className="flex items-center justify-between p-3 glass-subtle rounded-xl cursor-pointer hover:bg-white/5 transition-all">
            <div className="flex items-center gap-3">
              <MaterialIcon name={s.icon} className="text-muted-foreground" />
              <span className="text-sm">{s.label}</span>
            </div>
            <div
              onClick={() => toggleSetting(s.key)}
              className={`relative w-11 h-6 rounded-full transition-colors cursor-pointer ${
                settings[s.key as keyof typeof settings] ? 'bg-indigo-500' : 'bg-white/10'
              }`}
            >
              <div
                className={`absolute top-1 size-4 rounded-full bg-white transition-transform ${
                  settings[s.key as keyof typeof settings] ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </div>
          </label>
        ))}
      </div>
    </div>
  );
}
