import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/api/client';
import { useAuthStore } from '@/stores/authStore';
import { useState } from 'react';
import { FolderKanban, Plus, Clock } from 'lucide-react';
import { Pagination } from '@/components/ui/Pagination';

const createProjectSchema = (t: (key: string) => string) => z.object({
  name: z.string().min(1, t('projects.validation.nameRequired')),
  code: z.string().min(1, t('projects.validation.codeRequired')),
  description: z.string().optional(),
  budget_hours: z.coerce.number().optional(),
});

const createTimeEntrySchema = (t: (key: string) => string) => z.object({
  project_id: z.string().min(1, t('projects.validation.projectRequired')),
  date: z.string().min(1, t('projects.validation.dateRequired')),
  minutes: z.coerce.number().min(1, t('projects.validation.minutesRequired')),
  description: z.string().optional(),
});

type ProjectForm = z.infer<ReturnType<typeof createProjectSchema>>;
type TimeEntryForm = z.infer<ReturnType<typeof createTimeEntrySchema>>;

export function ProjectsPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const [showProjectForm, setShowProjectForm] = useState(false);
  const [showTimeEntryForm, setShowTimeEntryForm] = useState(false);
  const [activeTab, setActiveTab] = useState<'projects' | 'timeEntries' | 'summary'>('projects');
  const [projectPage, setProjectPage] = useState(1);
  const [projectPageSize, setProjectPageSize] = useState(12);
  const isAdmin = user?.role === 'admin' || user?.role === 'manager';

  const { data: projects } = useQuery({
    queryKey: ['projects', projectPage, projectPageSize],
    queryFn: () => api.projects.getAll({ page: projectPage, page_size: projectPageSize }),
  });

  const { data: timeEntries } = useQuery({
    queryKey: ['timeEntries', 'my'],
    queryFn: () => api.timeEntries.getList(),
  });

  const { data: summary } = useQuery({
    queryKey: ['timeEntries', 'summary'],
    queryFn: () => api.timeEntries.getSummary(),
    enabled: isAdmin,
  });

  const projectSchema = createProjectSchema(t);
  const timeEntrySchema = createTimeEntrySchema(t);
  const projectForm = useForm<ProjectForm>({ resolver: zodResolver(projectSchema) });
  const timeEntryForm = useForm<TimeEntryForm>({ resolver: zodResolver(timeEntrySchema) });

  const createProjectMutation = useMutation({
    mutationFn: (data: ProjectForm) => api.projects.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      setShowProjectForm(false);
      projectForm.reset();
    },
  });

  const createTimeEntryMutation = useMutation({
    mutationFn: (data: TimeEntryForm) => api.timeEntries.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['timeEntries'] });
      setShowTimeEntryForm(false);
      timeEntryForm.reset();
    },
  });

  const deleteTimeEntryMutation = useMutation({
    mutationFn: (id: string) => api.timeEntries.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['timeEntries'] });
    },
  });

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      active: 'bg-green-500/20 text-green-400 border border-green-500/30',
      completed: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
      archived: 'bg-muted text-muted-foreground border border-border',
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs ${styles[status] || ''}`}>
        {t(`common.${status}`) || status}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <FolderKanban className="h-6 w-6" />
          {t('projects.title')}
        </h1>
        <div className="flex gap-2">
          <button
            onClick={() => setShowTimeEntryForm(!showTimeEntryForm)}
            className="flex items-center gap-2 px-4 py-2 border border-input rounded-md hover:bg-accent"
          >
            <Clock className="h-4 w-4" />
            {t('projects.logTime')}
          </button>
          {isAdmin && (
            <button
              onClick={() => setShowProjectForm(!showProjectForm)}
              className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" />
              {t('projects.newProject')}
            </button>
          )}
        </div>
      </div>

      {/* タブ */}
      <div className="flex gap-1 border-b border-border">
        {(['projects', 'timeEntries', ...(isAdmin ? ['summary'] : [])] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab as typeof activeTab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            {tab === 'projects' ? t('projects.list') : tab === 'timeEntries' ? t('projects.myTimeEntries') : t('projects.summary')}
          </button>
        ))}
      </div>

      {/* プロジェクト作成フォーム */}
      {showProjectForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('projects.newProject')}</h2>
          <form onSubmit={projectForm.handleSubmit((data) => createProjectMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.name')}</label>
                <input type="text" {...projectForm.register('name')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.code')}</label>
                <input type="text" {...projectForm.register('code')} placeholder="PRJ-001" className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.description')}</label>
                <input type="text" {...projectForm.register('description')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.budgetHours')}</label>
                <input type="number" {...projectForm.register('budget_hours')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
            </div>
            <div className="flex gap-2">
              <button type="submit" className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
                {t('common.create')}
              </button>
              <button type="button" onClick={() => setShowProjectForm(false)} className="px-4 py-2 border border-input rounded-md hover:bg-accent">
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* 工数記録フォーム */}
      {showTimeEntryForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('projects.logTime')}</h2>
          <form onSubmit={timeEntryForm.handleSubmit((data) => createTimeEntryMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.project')}</label>
                <select {...timeEntryForm.register('project_id')} className="w-full px-3 py-2 border border-input rounded-md bg-background">
                  <option value="">{t('common.selectPlaceholder')}</option>
                  {projects?.data?.map((p: Record<string, unknown>) => (
                    <option key={p.id as string} value={p.id as string}>{p.name as string}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.date')}</label>
                <input type="date" {...timeEntryForm.register('date')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.minutes')}</label>
                <input type="number" {...timeEntryForm.register('minutes')} placeholder="60" className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('projects.entryDescription')}</label>
                <input type="text" {...timeEntryForm.register('description')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
              </div>
            </div>
            <div className="flex gap-2">
              <button type="submit" className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
                {t('common.submit')}
              </button>
              <button type="button" onClick={() => setShowTimeEntryForm(false)} className="px-4 py-2 border border-input rounded-md hover:bg-accent">
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* プロジェクト一覧 */}
      {activeTab === 'projects' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects?.data?.map((p: Record<string, unknown>) => (
            <div key={p.id as string} className="bg-card border border-border rounded-lg p-5">
              <div className="flex items-center justify-between mb-3">
                <span className="text-xs font-mono text-muted-foreground">{p.code as string}</span>
                {statusBadge(p.status as string)}
              </div>
              <h3 className="font-semibold text-lg mb-2">{p.name as string}</h3>
              {Boolean(p.description) && <p className="text-sm text-muted-foreground mb-3">{p.description as string}</p>}
              {Boolean(p.budget_hours) && (
                <div className="text-sm">
                  <span className="text-muted-foreground">{t('common.budget')}: </span>
                  <span className="font-medium">{p.budget_hours as number}{t('common.hours')}</span>
                </div>
              )}
            </div>
          ))}
          {(!projects?.data || projects.data.length === 0) && (
            <div className="col-span-full py-8 text-center text-muted-foreground">{t('common.noData')}</div>
          )}
          {projects?.total_pages > 0 && (
            <div className="col-span-full">
              <Pagination
                currentPage={projectPage}
                totalPages={projects.total_pages}
                totalItems={projects.total}
                pageSize={projectPageSize}
                onPageChange={(p) => setProjectPage(p)}
                onPageSizeChange={(s) => { setProjectPageSize(s); setProjectPage(1); }}
                pageSizeOptions={[6, 12, 24, 48]}
              />
            </div>
          )}
        </div>
      )}

      {/* 工数記録一覧 */}
      {activeTab === 'timeEntries' && (
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-2 px-4">{t('projects.date')}</th>
                  <th className="text-left py-2 px-4">{t('projects.project')}</th>
                  <th className="text-left py-2 px-4">{t('projects.minutes')}</th>
                  <th className="text-left py-2 px-4">{t('projects.entryDescription')}</th>
                  <th className="text-left py-2 px-4"></th>
                </tr>
              </thead>
              <tbody>
                {timeEntries?.map?.((te: Record<string, unknown>) => (
                  <tr key={te.id as string} className="border-b border-border/50">
                    <td className="py-2 px-4">{(te.date as string)?.slice(0, 10)}</td>
                    <td className="py-2 px-4">{(te.project as Record<string, string>)?.name || '-'}</td>
                    <td className="py-2 px-4">{te.minutes as number}{t('common.minutes')}</td>
                    <td className="py-2 px-4">{(te.description as string) || '-'}</td>
                    <td className="py-2 px-4">
                      <button
                        onClick={() => deleteTimeEntryMutation.mutate(te.id as string)}
                        className="text-destructive hover:underline text-xs"
                      >
                        {t('common.delete')}
                      </button>
                    </td>
                  </tr>
                ))}
                {(!timeEntries || timeEntries.length === 0) && (
                  <tr><td colSpan={5} className="py-4 text-center text-muted-foreground">{t('common.noData')}</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* サマリー */}
      {activeTab === 'summary' && isAdmin && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('projects.summary')}</h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-2 px-4">{t('projects.code')}</th>
                  <th className="text-left py-2 px-4">{t('projects.name')}</th>
                  <th className="text-left py-2 px-4">{t('common.totalHours')}</th>
                  <th className="text-left py-2 px-4">{t('common.budgetHours')}</th>
                  <th className="text-left py-2 px-4">{t('common.memberCount')}</th>
                  <th className="text-left py-2 px-4">{t('common.progress')}</th>
                </tr>
              </thead>
              <tbody>
                {summary?.map?.((s: Record<string, unknown>) => {
                  const totalHours = s.total_hours as number;
                  const budgetHours = s.budget_hours as number | null;
                  const progress = budgetHours ? Math.min(100, (totalHours / budgetHours) * 100) : 0;
                  return (
                    <tr key={s.project_code as string} className="border-b border-border/50">
                      <td className="py-2 px-4 font-mono text-xs">{s.project_code as string}</td>
                      <td className="py-2 px-4">{s.project_name as string}</td>
                      <td className="py-2 px-4">{(s.total_hours as number).toFixed(1)}h</td>
                      <td className="py-2 px-4">{budgetHours ? `${budgetHours}h` : '-'}</td>
                      <td className="py-2 px-4">{s.member_count as number}</td>
                      <td className="py-2 px-4">
                        {budgetHours ? (
                          <div className="flex items-center gap-2">
                            <div className="w-24 bg-muted rounded-full h-2">
                              <div
                                className={`h-2 rounded-full ${progress > 90 ? 'bg-red-500' : progress > 70 ? 'bg-yellow-500' : 'bg-green-500'}`}
                                style={{ width: `${progress}%` }}
                              />
                            </div>
                            <span className="text-xs">{progress.toFixed(0)}%</span>
                          </div>
                        ) : '-'}
                      </td>
                    </tr>
                  );
                })}
                {(!summary || summary.length === 0) && (
                  <tr><td colSpan={6} className="py-4 text-center text-muted-foreground">{t('common.noData')}</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
