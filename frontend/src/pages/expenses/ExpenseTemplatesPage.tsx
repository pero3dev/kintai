import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const CATEGORIES = [
  'transportation', 'meals', 'accommodation', 'supplies',
  'communication', 'entertainment', 'other',
] as const;

function getCategoryIcon(category: string): string {
  const icons: Record<string, string> = {
    transportation: 'directions_car',
    meals: 'restaurant',
    accommodation: 'hotel',
    supplies: 'inventory_2',
    communication: 'phone',
    entertainment: 'celebration',
    other: 'more_horiz',
  };
  return icons[category] || 'receipt_long';
}

export function ExpenseTemplatesPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    title: '',
    category: '',
    description: '',
    amount: 0,
    is_recurring: false,
    recurring_day: 1,
  });

  const { data: templatesData, isLoading } = useQuery({
    queryKey: ['expense-templates'],
    queryFn: () => api.expenses.getTemplates(),
  });

  const templates = templatesData?.data || templatesData?.templates || [];

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) => api.expenses.createTemplate(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-templates'] });
      resetForm();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: typeof formData }) => api.expenses.updateTemplate(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-templates'] });
      resetForm();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.expenses.deleteTemplate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-templates'] });
    },
  });

  const useTemplateMutation = useMutation({
    mutationFn: (templateId: string) => api.expenses.useTemplate(templateId),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      if (data?.id) {
        navigate({ to: '/expenses/$expenseId', params: { expenseId: data.id } });
      } else {
        navigate({ to: '/expenses/new' });
      }
    },
  });

  const resetForm = () => {
    setFormData({ name: '', title: '', category: '', description: '', amount: 0, is_recurring: false, recurring_day: 1 });
    setShowCreateForm(false);
    setEditingId(null);
  };

  const startEdit = (template: Record<string, unknown>) => {
    setFormData({
      name: (template.name as string) || '',
      title: (template.title as string) || '',
      category: (template.category as string) || '',
      description: (template.description as string) || '',
      amount: Number(template.amount) || 0,
      is_recurring: Boolean(template.is_recurring),
      recurring_day: Number(template.recurring_day) || 1,
    });
    setEditingId(template.id as string);
    setShowCreateForm(true);
  };

  const handleSubmit = () => {
    if (!formData.name || !formData.title) return;
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: formData });
    } else {
      createMutation.mutate(formData);
    }
  };

  const recurringTemplates = useMemo(
    () => templates.filter((t: Record<string, unknown>) => t.is_recurring),
    [templates]
  );

  const regularTemplates = useMemo(
    () => templates.filter((t: Record<string, unknown>) => !t.is_recurring),
    [templates]
  );

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div className="flex items-center gap-4">
          <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold gradient-text">{t('expenses.templates.title')}</h1>
            <p className="text-muted-foreground text-sm mt-1">{t('expenses.templates.subtitle')}</p>
          </div>
        </div>
        <button
          onClick={() => { resetForm(); setShowCreateForm(true); }}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all"
        >
          <MaterialIcon name="add" className="text-lg" />
          {t('expenses.templates.create')}
        </button>
      </div>

      {/* 作成/編集フォーム */}
      {showCreateForm && (
        <div className="glass-card rounded-2xl p-6 space-y-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <MaterialIcon name={editingId ? 'edit' : 'add_card'} className="text-indigo-400" />
            {editingId ? t('expenses.templates.edit') : t('expenses.templates.create')}
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.templates.templateName')}</label>
              <input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm"
                placeholder={t('expenses.templates.templateNamePlaceholder')}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.fields.title')}</label>
              <input
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm"
                placeholder={t('expenses.placeholders.title')}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.fields.category')}</label>
              <select
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm"
              >
                <option value="">{t('expenses.placeholders.category')}</option>
                {CATEGORIES.map((cat) => (
                  <option key={cat} value={cat}>{t(`expenses.categories.${cat}`)}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.fields.amount')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">¥</span>
                <input
                  type="number"
                  value={formData.amount}
                  onChange={(e) => setFormData({ ...formData, amount: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 glass-input rounded-xl text-sm"
                />
              </div>
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.fields.description')}</label>
              <input
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm"
                placeholder={t('expenses.placeholders.description')}
              />
            </div>
          </div>

          {/* 定期経費オプション */}
          <div className="glass-subtle rounded-xl p-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.is_recurring}
                onChange={(e) => setFormData({ ...formData, is_recurring: e.target.checked })}
                className="size-4 rounded accent-indigo-500"
              />
              <div>
                <span className="text-sm font-medium">{t('expenses.templates.recurring')}</span>
                <p className="text-xs text-muted-foreground">{t('expenses.templates.recurringDesc')}</p>
              </div>
            </label>
            {formData.is_recurring && (
              <div className="mt-3 ml-7">
                <label className="block text-xs font-medium mb-1 text-foreground/80">{t('expenses.templates.recurringDay')}</label>
                <select
                  value={formData.recurring_day}
                  onChange={(e) => setFormData({ ...formData, recurring_day: Number(e.target.value) })}
                  className="px-3 py-2 glass-input rounded-xl text-sm w-32"
                >
                  {Array.from({ length: 28 }, (_, i) => i + 1).map((day) => (
                    <option key={day} value={day}>{t('expenses.templates.dayOfMonth', { day })}</option>
                  ))}
                </select>
              </div>
            )}
          </div>

          <div className="flex gap-3 justify-end">
            <button onClick={resetForm} className="px-4 py-2 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">
              {t('common.cancel')}
            </button>
            <button
              onClick={handleSubmit}
              disabled={!formData.name || !formData.title || createMutation.isPending || updateMutation.isPending}
              className="px-5 py-2 gradient-primary text-white text-sm font-semibold rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50"
            >
              {t('common.save')}
            </button>
          </div>
        </div>
      )}

      {/* 定期経費テンプレート */}
      {recurringTemplates.length > 0 && (
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="autorenew" className="text-emerald-400" />
            {t('expenses.templates.recurringTemplates')}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {recurringTemplates.map((template: Record<string, unknown>) => (
              <TemplateCard
                key={template.id as string}
                template={template}
                onEdit={() => startEdit(template)}
                onDelete={() => deleteMutation.mutate(template.id as string)}
                onUse={() => useTemplateMutation.mutate(template.id as string)}
                t={t}
                isRecurring
              />
            ))}
          </div>
        </div>
      )}

      {/* 通常テンプレート */}
      <div className="glass-card rounded-2xl p-6">
        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <MaterialIcon name="content_copy" className="text-indigo-400" />
          {t('expenses.templates.savedTemplates')}
        </h2>
        {isLoading ? (
          <p className="text-center py-8 text-muted-foreground">{t('common.loading')}</p>
        ) : regularTemplates.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            <MaterialIcon name="content_paste" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('expenses.templates.noTemplates')}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {regularTemplates.map((template: Record<string, unknown>) => (
              <TemplateCard
                key={template.id as string}
                template={template}
                onEdit={() => startEdit(template)}
                onDelete={() => deleteMutation.mutate(template.id as string)}
                onUse={() => useTemplateMutation.mutate(template.id as string)}
                t={t}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function TemplateCard({
  template, onEdit, onDelete, onUse, t, isRecurring,
}: {
  template: Record<string, unknown>;
  onEdit: () => void;
  onDelete: () => void;
  onUse: () => void;
  t: (key: string, opts?: Record<string, unknown>) => string;
  isRecurring?: boolean;
}) {
  return (
    <div className="glass-subtle rounded-xl p-4 hover:bg-white/5 transition-all group">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className={`size-10 rounded-xl flex items-center justify-center ${isRecurring ? 'bg-emerald-500/20' : 'bg-indigo-500/20'}`}>
            <MaterialIcon
              name={template.category ? getCategoryIcon(template.category as string) : 'description'}
              className={isRecurring ? 'text-emerald-400' : 'text-indigo-400'}
            />
          </div>
          <div>
            <p className="text-sm font-semibold">{template.name as string}</p>
            <p className="text-xs text-muted-foreground">{template.title as string}</p>
            {isRecurring && (
              <p className="text-xs text-emerald-400 mt-0.5">
                <MaterialIcon name="autorenew" className="text-xs align-middle" />
                {' '}{t('expenses.templates.everyMonth', { day: template.recurring_day })}
              </p>
            )}
          </div>
        </div>
        <div className="text-right">
          {!!template.amount && Number(template.amount) > 0 && (
            <p className="text-sm font-bold">¥{Number(template.amount).toLocaleString()}</p>
          )}
          {!!template.category && (
            <span className="text-[10px] text-muted-foreground">{t(`expenses.categories.${template.category}`)}</span>
          )}
        </div>
      </div>
      <div className="flex items-center gap-2 mt-3 pt-3 border-t border-white/5 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={onUse}
          className="flex items-center gap-1.5 px-3 py-1.5 gradient-primary text-white text-xs font-semibold rounded-lg hover:shadow-glow-sm transition-all"
        >
          <MaterialIcon name="add_card" className="text-sm" />
          {t('expenses.templates.useTemplate')}
        </button>
        <button onClick={onEdit}
          className="flex items-center gap-1 px-2 py-1.5 glass-subtle rounded-lg text-xs hover:bg-white/10 transition-all">
          <MaterialIcon name="edit" className="text-sm" />
        </button>
        <button onClick={() => { if (confirm(t('expenses.templates.deleteConfirm'))) onDelete(); }}
          className="flex items-center gap-1 px-2 py-1.5 hover:bg-red-500/10 text-red-400 rounded-lg text-xs transition-all">
          <MaterialIcon name="delete" className="text-sm" />
        </button>
      </div>
    </div>
  );
}
