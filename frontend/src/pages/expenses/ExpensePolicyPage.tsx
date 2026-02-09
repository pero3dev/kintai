import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const CATEGORIES = [
  'transportation', 'meals', 'accommodation', 'supplies',
  'communication', 'entertainment', 'other',
] as const;

export function ExpensePolicyPage() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const isAdmin = user?.role === 'admin';
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    category: '',
    monthly_limit: 0,
    per_claim_limit: 0,
    auto_approve_limit: 0,
    requires_receipt_above: 0,
    description: '',
    is_active: true,
  });

  const { data: policiesData, isLoading } = useQuery({
    queryKey: ['expense-policies'],
    queryFn: () => api.expenses.getPolicies(),
  });

  const { data: budgetData } = useQuery({
    queryKey: ['expense-budgets'],
    queryFn: () => api.expenses.getBudgets(),
    enabled: isAdmin,
  });

  const policies = policiesData?.data || policiesData?.policies || [];
  const budgets = budgetData?.data || budgetData?.budgets || [];

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) => api.expenses.createPolicy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-policies'] });
      resetForm();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: typeof formData }) => api.expenses.updatePolicy(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-policies'] });
      resetForm();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.expenses.deletePolicy(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-policies'] });
    },
  });

  const resetForm = () => {
    setFormData({ category: '', monthly_limit: 0, per_claim_limit: 0, auto_approve_limit: 0, requires_receipt_above: 0, description: '', is_active: true });
    setShowForm(false);
    setEditingId(null);
  };

  const startEdit = (policy: Record<string, unknown>) => {
    setFormData({
      category: (policy.category as string) || '',
      monthly_limit: Number(policy.monthly_limit) || 0,
      per_claim_limit: Number(policy.per_claim_limit) || 0,
      auto_approve_limit: Number(policy.auto_approve_limit) || 0,
      requires_receipt_above: Number(policy.requires_receipt_above) || 0,
      description: (policy.description as string) || '',
      is_active: Boolean(policy.is_active),
    });
    setEditingId(policy.id as string);
    setShowForm(true);
  };

  const handleSubmit = () => {
    if (!formData.category) return;
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: formData });
    } else {
      createMutation.mutate(formData);
    }
  };

  if (!isAdmin) {
    return (
      <div className="space-y-6 animate-fade-in">
        <div className="flex items-center gap-4">
          <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold gradient-text">{t('expenses.policy.title')}</h1>
            <p className="text-muted-foreground text-sm mt-1">{t('expenses.policy.subtitle')}</p>
          </div>
        </div>

        {/* 従業員ビュー: ポリシー一覧（閲覧のみ） */}
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="policy" className="text-indigo-400" />
            {t('expenses.policy.currentPolicies')}
          </h2>
          {policies.length === 0 ? (
            <p className="text-center py-8 text-muted-foreground text-sm">{t('expenses.policy.noPolicies')}</p>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {policies.filter((p: Record<string, unknown>) => p.is_active).map((policy: Record<string, unknown>) => (
                <PolicyCard key={policy.id as string} policy={policy} t={t} readOnly />
              ))}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Link to="/expenses" className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all">
            <MaterialIcon name="arrow_back" />
          </Link>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('expenses.policy.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('expenses.policy.subtitle')}</p>
          </div>
        </div>
        <button
          onClick={() => { resetForm(); setShowForm(true); }}
          className="flex items-center gap-2 px-5 py-2.5 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all"
        >
          <MaterialIcon name="add" className="text-lg" />
          {t('expenses.policy.addPolicy')}
        </button>
      </div>

      {/* 予算概要ダッシュボード */}
      {budgets.length > 0 && (
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="account_balance" className="text-indigo-400" />
            {t('expenses.policy.budgetOverview')}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {budgets.map((budget: Record<string, unknown>) => {
              const used = Number(budget.used_amount) || 0;
              const total = Number(budget.budget_amount) || 1;
              const pct = Math.min((used / total) * 100, 100);
              const isOver = pct >= 90;
              return (
                <div key={budget.department as string} className="glass-subtle rounded-xl p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-semibold">{budget.department as string}</span>
                    <span className={`text-xs font-bold ${isOver ? 'text-red-400' : 'text-emerald-400'}`}>
                      {pct.toFixed(1)}%
                    </span>
                  </div>
                  <div className="h-2 rounded-full bg-white/5 overflow-hidden mb-2">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${isOver ? 'bg-red-400' : 'bg-emerald-400'}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>¥{used.toLocaleString()}</span>
                    <span>¥{total.toLocaleString()}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* ポリシー作成/編集フォーム */}
      {showForm && (
        <div className="glass-card rounded-2xl p-6 space-y-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <MaterialIcon name={editingId ? 'edit' : 'add_circle'} className="text-indigo-400" />
            {editingId ? t('expenses.policy.editPolicy') : t('expenses.policy.addPolicy')}
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.policy.monthlyLimit')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">¥</span>
                <input
                  type="number"
                  value={formData.monthly_limit}
                  onChange={(e) => setFormData({ ...formData, monthly_limit: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 glass-input rounded-xl text-sm"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.policy.perClaimLimit')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">¥</span>
                <input
                  type="number"
                  value={formData.per_claim_limit}
                  onChange={(e) => setFormData({ ...formData, per_claim_limit: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 glass-input rounded-xl text-sm"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.policy.autoApproveLimit')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">¥</span>
                <input
                  type="number"
                  value={formData.auto_approve_limit}
                  onChange={(e) => setFormData({ ...formData, auto_approve_limit: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 glass-input rounded-xl text-sm"
                  placeholder="0"
                />
              </div>
              <p className="text-[10px] text-muted-foreground mt-1">{t('expenses.policy.autoApproveDesc')}</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.policy.receiptRequired')}</label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">¥</span>
                <input
                  type="number"
                  value={formData.requires_receipt_above}
                  onChange={(e) => setFormData({ ...formData, requires_receipt_above: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 glass-input rounded-xl text-sm"
                />
              </div>
              <p className="text-[10px] text-muted-foreground mt-1">{t('expenses.policy.receiptRequiredDesc')}</p>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1 text-foreground/80">{t('expenses.policy.description')}</label>
              <input
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm"
                placeholder={t('expenses.policy.descriptionPlaceholder')}
              />
            </div>
          </div>

          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={formData.is_active}
              onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
              className="size-4 rounded accent-indigo-500"
            />
            <span className="text-sm">{t('expenses.policy.active')}</span>
          </label>

          <div className="flex gap-3 justify-end">
            <button onClick={resetForm} className="px-4 py-2 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">
              {t('common.cancel')}
            </button>
            <button
              onClick={handleSubmit}
              disabled={!formData.category || createMutation.isPending || updateMutation.isPending}
              className="px-5 py-2 gradient-primary text-white text-sm font-semibold rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50"
            >
              {t('common.save')}
            </button>
          </div>
        </div>
      )}

      {/* ポリシー一覧 */}
      <div className="glass-card rounded-2xl p-6">
        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <MaterialIcon name="policy" className="text-indigo-400" />
          {t('expenses.policy.currentPolicies')}
        </h2>
        {isLoading ? (
          <p className="text-center py-8 text-muted-foreground">{t('common.loading')}</p>
        ) : policies.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            <MaterialIcon name="shield" className="text-4xl mb-2 block opacity-50" />
            <p className="text-sm">{t('expenses.policy.noPolicies')}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {policies.map((policy: Record<string, unknown>) => (
              <PolicyCard
                key={policy.id as string}
                policy={policy}
                t={t}
                onEdit={() => startEdit(policy)}
                onDelete={() => { if (confirm(t('expenses.policy.deleteConfirm'))) deleteMutation.mutate(policy.id as string); }}
              />
            ))}
          </div>
        )}
      </div>

      {/* ポリシー違反アラート */}
      <PolicyViolations t={t} />
    </div>
  );
}

function PolicyCard({
  policy, t, readOnly, onEdit, onDelete,
}: {
  policy: Record<string, unknown>;
  t: (key: string) => string;
  readOnly?: boolean;
  onEdit?: () => void;
  onDelete?: () => void;
}) {
  const getCategoryIcon = (cat: string) => {
    const icons: Record<string, string> = {
      transportation: 'directions_car', meals: 'restaurant', accommodation: 'hotel',
      supplies: 'inventory_2', communication: 'phone', entertainment: 'celebration', other: 'more_horiz',
    };
    return icons[cat] || 'receipt_long';
  };

  return (
    <div className={`glass-subtle rounded-xl p-4 ${!Boolean(policy.is_active) && !readOnly ? 'opacity-50' : ''}`}>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="size-8 rounded-lg bg-indigo-500/20 flex items-center justify-center">
            <MaterialIcon name={getCategoryIcon(policy.category as string)} className="text-indigo-400 text-base" />
          </div>
          <div>
            <span className="text-sm font-semibold">{t(`expenses.categories.${policy.category}`)}</span>
            {!Boolean(policy.is_active) && (
              <span className="ml-2 text-[10px] bg-gray-500/20 text-gray-400 px-1.5 py-0.5 rounded-full">{t('expenses.policy.inactive')}</span>
            )}
          </div>
        </div>
        {!readOnly && (
          <div className="flex gap-1">
            <button onClick={onEdit} className="p-1 rounded-lg hover:bg-white/10 transition-colors">
              <MaterialIcon name="edit" className="text-base text-muted-foreground" />
            </button>
            <button onClick={onDelete} className="p-1 rounded-lg hover:bg-red-500/10 transition-colors">
              <MaterialIcon name="delete" className="text-base text-red-400" />
            </button>
          </div>
        )}
      </div>
      <div className="grid grid-cols-2 gap-2 text-xs">
        {Number(policy.monthly_limit) > 0 && (
          <div className="glass-subtle rounded-lg p-2">
            <p className="text-muted-foreground">{t('expenses.policy.monthlyLimit')}</p>
            <p className="font-bold">¥{Number(policy.monthly_limit).toLocaleString()}</p>
          </div>
        )}
        {Number(policy.per_claim_limit) > 0 && (
          <div className="glass-subtle rounded-lg p-2">
            <p className="text-muted-foreground">{t('expenses.policy.perClaimLimit')}</p>
            <p className="font-bold">¥{Number(policy.per_claim_limit).toLocaleString()}</p>
          </div>
        )}
        {Number(policy.auto_approve_limit) > 0 && (
          <div className="glass-subtle rounded-lg p-2">
            <p className="text-muted-foreground">{t('expenses.policy.autoApproveLimit')}</p>
            <p className="font-bold">¥{Number(policy.auto_approve_limit).toLocaleString()}</p>
          </div>
        )}
        {Number(policy.requires_receipt_above) > 0 && (
          <div className="glass-subtle rounded-lg p-2">
            <p className="text-muted-foreground">{t('expenses.policy.receiptRequired')}</p>
            <p className="font-bold">¥{Number(policy.requires_receipt_above).toLocaleString()}</p>
          </div>
        )}
      </div>
      {!!policy.description && (
        <p className="text-xs text-muted-foreground mt-2">{policy.description as string}</p>
      )}
    </div>
  );
}

function PolicyViolations({ t }: { t: (key: string) => string }) {
  const { data } = useQuery({
    queryKey: ['expense-policy-violations'],
    queryFn: () => api.expenses.getPolicyViolations(),
  });

  const violations = data?.data || data?.violations || [];

  if (violations.length === 0) return null;

  return (
    <div className="glass-card rounded-2xl p-6 border border-red-500/20">
      <h2 className="text-lg font-semibold mb-4 flex items-center gap-2 text-red-400">
        <MaterialIcon name="warning" />
        {t('expenses.policy.violations')}
      </h2>
      <div className="space-y-2">
        {violations.map((v: Record<string, unknown>, i: number) => (
          <div key={i} className="flex items-center gap-3 glass-subtle rounded-xl p-3">
            <MaterialIcon name="error" className="text-red-400" />
            <div className="flex-1">
              <p className="text-sm font-medium">{v.user_name as string} - {v.expense_title as string}</p>
              <p className="text-xs text-red-400/80">{v.violation_message as string}</p>
            </div>
            <span className="text-sm font-bold text-red-400">¥{Number(v.amount).toLocaleString()}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
