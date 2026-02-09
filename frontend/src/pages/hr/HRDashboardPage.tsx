import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

export function HRDashboardPage() {
  const { t } = useTranslation();

  const { data: stats } = useQuery({
    queryKey: ['hr-stats'],
    queryFn: () => api.hr.getStats(),
  });

  const { data: activitiesData } = useQuery({
    queryKey: ['hr-activities'],
    queryFn: () => api.hr.getRecentActivities(),
  });

  const s = stats || {
    total_employees: 0,
    active_employees: 0,
    new_hires_this_month: 0,
    turnover_rate: 0,
    open_positions: 0,
    upcoming_reviews: 0,
    training_completion: 0,
    pending_documents: 0,
  };

  const activities: Record<string, unknown>[] = activitiesData?.data || activitiesData || [];

  const statCards = [
    { label: t('hr.dashboard.totalEmployees'), value: s.total_employees, icon: 'groups', color: 'bg-indigo-500/20', iconColor: 'text-indigo-400' },
    { label: t('hr.dashboard.activeEmployees'), value: s.active_employees, icon: 'person', color: 'bg-emerald-500/20', iconColor: 'text-emerald-400' },
    { label: t('hr.dashboard.newHires'), value: s.new_hires_this_month, icon: 'person_add', color: 'bg-blue-500/20', iconColor: 'text-blue-400' },
    { label: t('hr.dashboard.turnoverRate'), value: `${(Number(s.turnover_rate) || 0).toFixed(1)}%`, icon: 'trending_down', color: 'bg-red-500/20', iconColor: 'text-red-400' },
    { label: t('hr.dashboard.openPositions'), value: s.open_positions, icon: 'work', color: 'bg-amber-500/20', iconColor: 'text-amber-400' },
    { label: t('hr.dashboard.upcomingReviews'), value: s.upcoming_reviews, icon: 'rate_review', color: 'bg-purple-500/20', iconColor: 'text-purple-400' },
    { label: t('hr.dashboard.trainingCompletion'), value: `${(Number(s.training_completion) || 0).toFixed(0)}%`, icon: 'school', color: 'bg-cyan-500/20', iconColor: 'text-cyan-400' },
    { label: t('hr.dashboard.pendingDocuments'), value: s.pending_documents, icon: 'description', color: 'bg-orange-500/20', iconColor: 'text-orange-400' },
  ];

  const quickActions = [
    { icon: 'person_add', label: t('hr.employees.addEmployee'), desc: t('hr.dashboard.addEmployeeDesc'), to: '/hr/employees', color: 'text-indigo-400' },
    { icon: 'rate_review', label: t('hr.evaluations.newCycle'), desc: t('hr.dashboard.startEvalDesc'), to: '/hr/evaluations', color: 'text-purple-400' },
    { icon: 'campaign', label: t('hr.announcements.addAnnouncement'), desc: t('hr.dashboard.postAnnouncementDesc'), to: '/hr/announcements', color: 'text-amber-400' },
    { icon: 'bar_chart', label: t('hr.dashboard.viewReportsDesc'), desc: t('hr.dashboard.viewReportsDesc'), to: '/hr/departments', color: 'text-cyan-400' },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('hr.dashboard.title')}</h1>
          <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('hr.dashboard.subtitle')}</p>
        </div>
      </div>

      {/* 統計カード */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
        {statCards.map((card) => (
          <div key={card.label} className="glass-card stat-shimmer rounded-2xl p-4 sm:p-5">
            <div className="flex items-center gap-2 sm:gap-3 mb-2">
              <div className={`size-9 sm:size-10 rounded-xl ${card.color} flex items-center justify-center`}>
                <MaterialIcon name={card.icon} className={`${card.iconColor} text-lg sm:text-xl`} />
              </div>
              <p className="text-xs text-muted-foreground leading-tight">{card.label}</p>
            </div>
            <p className="text-xl sm:text-2xl font-bold gradient-text">{card.value}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* クイックアクション */}
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="bolt" className="text-indigo-400" />
            {t('hr.dashboard.quickActions')}
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {quickActions.map((action) => (
              <Link
                key={action.label}
                to={action.to as '/'}
                className="glass-subtle rounded-xl p-4 hover:bg-white/5 transition-all group"
              >
                <div className="flex items-center gap-3">
                  <div className="size-10 rounded-xl bg-white/5 flex items-center justify-center group-hover:scale-110 transition-transform">
                    <MaterialIcon name={action.icon} className={action.color} />
                  </div>
                  <div>
                    <p className="text-sm font-semibold">{action.label}</p>
                    <p className="text-xs text-muted-foreground">{action.desc}</p>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </div>

        {/* 最近の活動 */}
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="history" className="text-indigo-400" />
            {t('hr.dashboard.recentActivities')}
          </h2>
          {activities.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground text-sm">{t('common.noData')}</p>
          ) : (
            <div className="space-y-3">
              {activities.slice(0, 8).map((activity, i) => (
                <div key={i} className="flex items-start gap-3 glass-subtle rounded-xl p-3">
                  <div className="size-8 rounded-lg bg-indigo-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <MaterialIcon name={activity.icon as string || 'info'} className="text-indigo-400 text-sm" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm">{activity.message as string}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">{activity.timestamp as string}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
