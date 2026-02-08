import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/api/client';
import { useState } from 'react';
import { GitBranch, Plus, Trash2 } from 'lucide-react';

const stepSchema = z.object({
  step_order: z.coerce.number().min(1),
  step_type: z.enum(['role', 'specific_user']),
  approver_role: z.string().optional(),
  approver_id: z.string().optional(),
});

const createFlowSchema = (t: (key: string) => string) => z.object({
  name: z.string().min(1, t('approvalFlows.validation.nameRequired')),
  flow_type: z.enum(['leave', 'overtime', 'correction']),
  steps: z.array(stepSchema).min(1, t('approvalFlows.validation.stepsRequired')),
});

type FlowForm = z.infer<ReturnType<typeof createFlowSchema>>;

export function ApprovalFlowsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [, setEditingId] = useState<string | null>(null);
  const flowSchema = createFlowSchema(t);

  const { data: flows } = useQuery({
    queryKey: ['approval-flows'],
    queryFn: () => api.approvalFlows.getAll(),
  });

  const { register, handleSubmit, control, reset, formState: { errors } } = useForm<FlowForm>({
    resolver: zodResolver(flowSchema),
    defaultValues: {
      flow_type: 'leave',
      steps: [{ step_order: 1, step_type: 'role', approver_role: 'manager' }],
    },
  });

  const { fields, append, remove } = useFieldArray({ control, name: 'steps' });

  const createMutation = useMutation({
    mutationFn: (data: FlowForm) => api.approvalFlows.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['approval-flows'] });
      setShowForm(false);
      reset();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) =>
      api.approvalFlows.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['approval-flows'] });
      setEditingId(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.approvalFlows.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['approval-flows'] });
    },
  });

  const flowTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      leave: t('approvalFlows.flowTypes.leave'),
      overtime: t('approvalFlows.flowTypes.overtime'),
      correction: t('approvalFlows.flowTypes.correction'),
    };
    return labels[type] || type;
  };

  const stepTypeLabel = (type: string) => {
    return type === 'role' ? t('approvalFlows.stepTypes.role') : t('approvalFlows.stepTypes.specificUser');
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <GitBranch className="h-6 w-6" />
          {t('approvalFlows.title')}
        </h1>
        <button
          onClick={() => { setShowForm(!showForm); setEditingId(null); }}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          {t('approvalFlows.create')}
        </button>
      </div>

      {/* 作成フォーム */}
      {showForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4">{t('approvalFlows.create')}</h2>
          <form onSubmit={handleSubmit((data) => createMutation.mutate(data))} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('approvalFlows.name')}</label>
                <input type="text" {...register('name')} placeholder={t('approvalFlows.placeholder')} className="w-full px-3 py-2 border border-input rounded-md bg-background" />
                {errors.name && <p className="text-sm text-destructive mt-1">{errors.name.message}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">{t('approvalFlows.flowType')}</label>
                <select {...register('flow_type')} className="w-full px-3 py-2 border border-input rounded-md bg-background">
                  <option value="leave">{t('approvalFlows.flowTypes.leave')}</option>
                  <option value="overtime">{t('approvalFlows.flowTypes.overtime')}</option>
                  <option value="correction">{t('approvalFlows.flowTypes.correction')}</option>
                </select>
              </div>
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="text-sm font-medium">{t('approvalFlows.steps')}</label>
                <button
                  type="button"
                  onClick={() => append({ step_order: fields.length + 1, step_type: 'role', approver_role: 'manager' })}
                  className="text-sm text-primary hover:underline flex items-center gap-1"
                >
                  <Plus className="h-3 w-3" /> {t('approvalFlows.addStep')}
                </button>
              </div>
              <div className="space-y-3">
                {fields.map((field, index) => (
                  <div key={field.id} className="flex items-center gap-3 p-3 border border-border rounded-md">
                    <span className="text-sm font-bold text-muted-foreground w-8">#{index + 1}</span>
                    <input type="hidden" {...register(`steps.${index}.step_order`)} value={index + 1} />
                    <select
                      {...register(`steps.${index}.step_type`)}
                      className="px-3 py-2 border border-input rounded-md bg-background text-sm"
                    >
                      <option value="role">{t('approvalFlows.stepTypes.role')}</option>
                      <option value="specific_user">{t('approvalFlows.stepTypes.specificUser')}</option>
                    </select>
                    <input
                      type="text"
                      {...register(`steps.${index}.approver_role`)}
                      placeholder="manager / admin"
                      className="flex-1 px-3 py-2 border border-input rounded-md bg-background text-sm"
                    />
                    {fields.length > 1 && (
                      <button type="button" onClick={() => remove(index)} className="p-1 text-destructive hover:bg-destructive/10 rounded">
                        <Trash2 className="h-4 w-4" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="flex gap-2">
              <button type="submit" className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
                {t('common.create')}
              </button>
              <button type="button" onClick={() => setShowForm(false)} className="px-4 py-2 border border-input rounded-md hover:bg-accent">
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* フロー一覧 */}
      <div className="space-y-4">
        {(flows as Record<string, unknown>[] | undefined)?.map((flow: Record<string, unknown>) => (
          <div key={flow.id as string} className="bg-card border border-border rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-semibold text-lg">{flow.name as string}</h3>
                <div className="flex items-center gap-2 mt-1">
                  <span className="px-2 py-1 bg-primary/10 text-primary rounded text-xs">
                    {flowTypeLabel(flow.flow_type as string)}
                  </span>
                  <span className={`px-2 py-1 rounded text-xs ${flow.is_active ? 'bg-green-500/20 text-green-400' : 'bg-muted text-muted-foreground'}`}>
                    {flow.is_active ? t('approvalFlows.active') : t('approvalFlows.inactive')}
                  </span>
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => updateMutation.mutate({
                    id: flow.id as string,
                    data: { is_active: !flow.is_active },
                  })}
                  className="px-3 py-1 border border-input rounded text-sm hover:bg-accent"
                >
                  {flow.is_active ? t('approvalFlows.disable') : t('approvalFlows.enable')}
                </button>
                <button
                  onClick={() => deleteMutation.mutate(flow.id as string)}
                  className="px-3 py-1 text-destructive border border-destructive/30 rounded text-sm hover:bg-destructive/10"
                >
                  {t('common.delete')}
                </button>
              </div>
            </div>

            {/* Approval Steps Visualization */}
            <div className="flex items-center gap-2 overflow-x-auto">
              {(flow.steps as Record<string, unknown>[] || [])
                .sort((a, b) => (a.step_order as number) - (b.step_order as number))
                .map((step, idx, arr) => (
                  <div key={step.id as string || idx} className="flex items-center gap-2">
                    <div className="px-4 py-2 bg-primary/10 border border-primary/20 rounded-lg text-center min-w-[120px]">
                      <p className="text-xs text-muted-foreground">{t('approvalFlows.step', { num: step.step_order as number })}</p>
                      <p className="font-medium text-sm">{stepTypeLabel(step.step_type as string)}</p>
                      <p className="text-xs text-primary">{(step.approver_role as string) || t('approvalFlows.specifiedUser')}</p>
                    </div>
                    {idx < arr.length - 1 && (
                      <span className="text-muted-foreground">→</span>
                    )}
                  </div>
                ))}
            </div>
          </div>
        ))}
        {(!flows || (flows as unknown[]).length === 0) && (
          <div className="text-center py-8 text-muted-foreground">{t('common.noData')}</div>
        )}
      </div>
    </div>
  );
}
