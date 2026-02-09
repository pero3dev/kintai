import { useState, useRef, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { useForm, useFieldArray } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const CATEGORIES = [
  'transportation',
  'meals',
  'accommodation',
  'supplies',
  'communication',
  'entertainment',
  'other',
] as const;

function createExpenseSchema(t: (key: string) => string) {
  return z.object({
    title: z.string().min(1, t('expenses.validation.titleRequired')),
    items: z.array(
      z.object({
        expense_date: z.string().min(1, t('expenses.validation.dateRequired')),
        category: z.string().min(1, t('expenses.validation.categoryRequired')),
        description: z.string().min(1, t('expenses.validation.descriptionRequired')),
        amount: z.coerce.number().min(1, t('expenses.validation.amountRequired')),
        receipt_url: z.string().optional(),
      })
    ).min(1, t('expenses.validation.itemsRequired')),
    notes: z.string().optional(),
  });
}

type ExpenseFormData = z.infer<ReturnType<typeof createExpenseSchema>>;

function ReceiptUpload({ index, onUpload }: { index: number; onUpload: (index: number, url: string) => void }) {
  const { t } = useTranslation();
  const [isDragging, setIsDragging] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(async (file: File) => {
    if (!file.type.startsWith('image/') && file.type !== 'application/pdf') return;
    setIsUploading(true);
    if (file.type.startsWith('image/')) {
      const reader = new FileReader();
      reader.onload = (e) => setPreview(e.target?.result as string);
      reader.readAsDataURL(file);
    }
    try {
      const result = await api.expenses.uploadReceipt(file);
      onUpload(index, result?.url || '');
    } catch {
      // upload failed
    } finally {
      setIsUploading(false);
    }
  }, [index, onUpload]);

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files[0];
    if (file) handleFile(file);
  };

  return (
    <div
      onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
      onDragLeave={() => setIsDragging(false)}
      onDrop={handleDrop}
      onClick={() => fileInputRef.current?.click()}
      className={`relative flex flex-col items-center justify-center p-3 rounded-xl border-2 border-dashed cursor-pointer transition-all ${
        isDragging ? 'border-indigo-400 bg-indigo-500/10' : 'border-white/10 hover:border-white/20 hover:bg-white/5'
      }`}
    >
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*,application/pdf"
        capture="environment"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleFile(file);
        }}
      />
      {isUploading ? (
        <MaterialIcon name="hourglass_empty" className="text-2xl text-indigo-400 animate-spin" />
      ) : preview ? (
        <img src={preview} alt="Receipt" className="max-h-20 rounded-lg object-contain" />
      ) : (
        <>
          <MaterialIcon name="cloud_upload" className="text-2xl text-muted-foreground" />
          <span className="text-[10px] text-muted-foreground mt-1 text-center">{t('expenses.receipt.dragOrTap')}</span>
        </>
      )}
    </div>
  );
}

export function ExpenseNewPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [submitType, setSubmitType] = useState<'draft' | 'submit'>('submit');
  const [fromTemplate, setFromTemplate] = useState(false);

  // テンプレート読み込み
  const { data: templatesData } = useQuery({
    queryKey: ['expense-templates'],
    queryFn: () => api.expenses.getTemplates(),
  });
  const templates = templatesData?.data || templatesData?.templates || [];

  const schema = createExpenseSchema(t);

  const {
    register,
    handleSubmit,
    control,
    watch,
    formState: { errors },
  } = useForm<ExpenseFormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: '',
      items: [{ expense_date: '', category: '', description: '', amount: 0, receipt_url: '' }],
      notes: '',
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'items',
  });

  const items = watch('items');
  const totalAmount = items.reduce((sum, item) => sum + (Number(item.amount) || 0), 0);

  const createMutation = useMutation({
    mutationFn: (data: ExpenseFormData & { status: string }) => api.expenses.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense-stats'] });
      navigate({ to: '/expenses' });
    },
  });

  const onSubmit = (data: ExpenseFormData) => {
    createMutation.mutate({ ...data, status: submitType === 'draft' ? 'draft' : 'pending' });
  };

  const handleReceiptUpload = (index: number, url: string) => {
    if (url) {
      // Update form value for receipt_url at the given index
      const currentItems = watch('items');
      if (currentItems[index]) {
        currentItems[index].receipt_url = url;
      }
    }
  };

  const applyTemplate = (template: Record<string, unknown>) => {
    if (template.title) {
      // Fill form with template data
      const currentItems = watch('items');
      if (currentItems.length === 1 && !currentItems[0].description) {
        remove(0);
      }
      append({
        expense_date: '',
        category: (template.category as string) || '',
        description: (template.description as string) || '',
        amount: Number(template.amount) || 0,
        receipt_url: '',
      });
      setFromTemplate(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate({ to: '/expenses' })}
          className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all"
        >
          <MaterialIcon name="arrow_back" />
        </button>
        <div>
          <h1 className="text-2xl font-bold gradient-text">{t('expenses.new.title')}</h1>
          <p className="text-muted-foreground text-sm mt-1">{t('expenses.new.subtitle')}</p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* テンプレートから作成 */}
        {templates.length > 0 && (
          <div className="glass-card rounded-2xl p-4">
            <button
              type="button"
              onClick={() => setFromTemplate(!fromTemplate)}
              className="flex items-center gap-2 text-sm text-indigo-400 hover:underline w-full"
            >
              <MaterialIcon name="content_copy" className="text-base" />
              {t('expenses.receipt.fromTemplate')}
              <MaterialIcon name={fromTemplate ? 'expand_less' : 'expand_more'} className="text-base ml-auto" />
            </button>
            {fromTemplate && (
              <div className="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-2">
                {templates.map((tpl: Record<string, unknown>) => (
                  <button
                    key={tpl.id as string}
                    type="button"
                    onClick={() => applyTemplate(tpl)}
                    className="flex items-center gap-3 p-3 glass-subtle rounded-xl text-left hover:bg-white/5 transition-all"
                  >
                    <MaterialIcon name="description" className="text-indigo-400" />
                    <div>
                      <p className="text-sm font-medium">{tpl.name as string}</p>
                      <p className="text-xs text-muted-foreground">¥{Number(tpl.amount).toLocaleString()}</p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}

        {/* 基本情報 */}
        <div className="glass-card rounded-2xl p-4 sm:p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="info" className="text-indigo-400" />
            {t('expenses.new.basicInfo')}
          </h2>
          <div>
            <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.fields.title')}</label>
            <input
              {...register('title')}
              className="w-full px-4 py-2.5 glass-input rounded-xl"
              placeholder={t('expenses.placeholders.title')}
            />
            {errors.title && <p className="text-sm text-red-400 mt-1">{errors.title.message}</p>}
          </div>
        </div>

        {/* 明細 */}
        <div className="glass-card rounded-2xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <MaterialIcon name="list_alt" className="text-indigo-400" />
              {t('expenses.new.items')}
            </h2>
            <button
              type="button"
              onClick={() => append({ expense_date: '', category: '', description: '', amount: 0, receipt_url: '' })}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-xl glass-subtle hover:bg-white/10 transition-all text-indigo-400"
            >
              <MaterialIcon name="add" className="text-base" />
              {t('expenses.new.addItem')}
            </button>
          </div>
          {errors.items?.root && <p className="text-sm text-red-400 mb-3">{errors.items.root.message}</p>}

          <div className="space-y-4">
            {fields.map((field, index) => (
              <div key={field.id} className="glass-subtle rounded-xl p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-muted-foreground">
                    {t('expenses.new.itemNumber', { number: index + 1 })}
                  </span>
                  {fields.length > 1 && (
                    <button
                      type="button"
                      onClick={() => remove(index)}
                      className="p-1 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors"
                    >
                      <MaterialIcon name="close" className="text-base" />
                    </button>
                  )}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium mb-1 text-foreground/80">{t('expenses.fields.date')}</label>
                    <input
                      type="date"
                      {...register(`items.${index}.expense_date`)}
                      className="w-full px-3 py-2 glass-input rounded-xl text-sm"
                    />
                    {errors.items?.[index]?.expense_date && (
                      <p className="text-xs text-red-400 mt-1">{errors.items[index].expense_date?.message}</p>
                    )}
                  </div>
                  <div>
                    <label className="block text-xs font-medium mb-1 text-foreground/80">{t('expenses.fields.category')}</label>
                    <select
                      {...register(`items.${index}.category`)}
                      className="w-full px-3 py-2 glass-input rounded-xl text-sm"
                    >
                      <option value="">{t('expenses.placeholders.category')}</option>
                      {CATEGORIES.map((cat) => (
                        <option key={cat} value={cat}>{t(`expenses.categories.${cat}`)}</option>
                      ))}
                    </select>
                    {errors.items?.[index]?.category && (
                      <p className="text-xs text-red-400 mt-1">{errors.items[index].category?.message}</p>
                    )}
                  </div>
                  <div>
                    <label className="block text-xs font-medium mb-1 text-foreground/80">{t('expenses.fields.description')}</label>
                    <input
                      {...register(`items.${index}.description`)}
                      className="w-full px-3 py-2 glass-input rounded-xl text-sm"
                      placeholder={t('expenses.placeholders.description')}
                    />
                    {errors.items?.[index]?.description && (
                      <p className="text-xs text-red-400 mt-1">{errors.items[index].description?.message}</p>
                    )}
                  </div>
                  <div>
                    <label className="block text-xs font-medium mb-1 text-foreground/80">{t('expenses.fields.amount')}</label>
                    <div className="relative">
                      <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">¥</span>
                      <input
                        type="number"
                        {...register(`items.${index}.amount`)}
                        className="w-full pl-8 pr-3 py-2 glass-input rounded-xl text-sm"
                        placeholder="0"
                      />
                    </div>
                    {errors.items?.[index]?.amount && (
                      <p className="text-xs text-red-400 mt-1">{errors.items[index].amount?.message}</p>
                    )}
                  </div>
                  {/* レシートアップロード */}
                  <div className="md:col-span-2">
                    <label className="block text-xs font-medium mb-1 text-foreground/80">{t('expenses.receipt.label')}</label>
                    <ReceiptUpload index={index} onUpload={handleReceiptUpload} />
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* 合計 */}
          <div className="mt-4 pt-4 border-t border-white/5 flex items-center justify-between">
            <span className="text-sm font-medium text-muted-foreground">{t('expenses.fields.totalAmount')}</span>
            <span className="text-xl font-bold gradient-text">¥{totalAmount.toLocaleString()}</span>
          </div>
        </div>

        {/* 備考 */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="notes" className="text-indigo-400" />
            {t('expenses.fields.notes')}
          </h2>
          <textarea
            {...register('notes')}
            rows={3}
            className="w-full px-4 py-2.5 glass-input rounded-xl resize-none"
            placeholder={t('expenses.placeholders.notes')}
          />
        </div>

        {/* エラー */}
        {createMutation.isError && (
          <div className="glass-card rounded-2xl p-4 border-red-500/30">
            <p className="text-sm text-red-400">{(createMutation.error as Error).message}</p>
          </div>
        )}

        {/* アクション - モバイル対応 */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 sm:gap-4">
          <button
            type="button"
            onClick={() => navigate({ to: '/expenses' })}
            className="px-5 py-2.5 glass-input hover:bg-white/10 rounded-xl text-sm transition-all order-3 sm:order-1"
          >
            {t('common.cancel')}
          </button>
          <div className="flex gap-3 order-1 sm:order-2">
            <button
              type="submit"
              onClick={() => setSubmitType('draft')}
              disabled={createMutation.isPending}
              className="flex-1 sm:flex-initial flex items-center justify-center gap-2 px-5 py-2.5 glass-subtle hover:bg-white/10 rounded-xl text-sm font-medium transition-all border border-white/5"
            >
              <MaterialIcon name="save" className="text-base" />
              {t('expenses.actions.saveDraft')}
            </button>
            <button
              type="submit"
              onClick={() => setSubmitType('submit')}
              disabled={createMutation.isPending}
              className="flex-1 sm:flex-initial flex items-center justify-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all duration-300"
            >
              <MaterialIcon name="send" className="text-base" />
              {createMutation.isPending ? t('common.submitting') : t('expenses.actions.submit')}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
