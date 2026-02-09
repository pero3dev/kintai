import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';
import { Pagination } from '@/components/ui/Pagination';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const statusBadge = (status: string) => {
  const styles: Record<string, string> = {
    draft: 'bg-gray-500/20 text-gray-400 border border-gray-500/30',
    pending: 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30',
    step1_approved: 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30',
    step2_approved: 'bg-teal-500/20 text-teal-400 border border-teal-500/30',
    approved: 'bg-green-500/20 text-green-400 border border-green-500/30',
    rejected: 'bg-red-500/20 text-red-400 border border-red-500/30',
    returned: 'bg-orange-500/20 text-orange-400 border border-orange-500/30',
    reimbursed: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
  };
  return styles[status] || styles.draft;
};

export function ExpenseAdvancedApprovePage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const queryClient = useQueryClient();
  const locale = i18n.language === 'ja' ? ja : enUS;
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [action, setAction] = useState<'approve' | 'reject' | 'return' | null>(null);
  const [reason, setReason] = useState('');
  const [delegateMode, setDelegateMode] = useState(false);
  const [delegateTo, setDelegateTo] = useState('');
  const [delegateStart, setDelegateStart] = useState('');
  const [delegateEnd, setDelegateEnd] = useState('');

  const isManager = user?.role === 'admin' || user?.role === 'manager';

  const { data, isLoading } = useQuery({
    queryKey: ['expenses', 'pending-advanced', page, pageSize],
    queryFn: () => api.expenses.getPending({ page, page_size: pageSize }),
    enabled: isManager,
  });

  const { data: flowData } = useQuery({
    queryKey: ['expense-approval-flow-config'],
    queryFn: () => api.expenses.getApprovalFlowConfig(),
    enabled: isManager,
  });

  const { data: delegateData } = useQuery({
    queryKey: ['expense-delegates'],
    queryFn: () => api.expenses.getDelegates(),
    enabled: isManager,
  });

  const { data: usersData } = useQuery({
    queryKey: ['users-for-delegate'],
    queryFn: () => api.users.getAll({ page: 1, page_size: 100 }),
    enabled: delegateMode,
  });

  const expenses = data?.data || data?.expenses || [];
  const totalItems = data?.total || data?.pagination?.total || 0;
  const totalPages = Math.ceil(totalItems / pageSize) || 1;
  const flowConfig = flowData?.data || flowData || {};
  const delegates = delegateData?.data || delegateData?.delegates || [];
  const usersList = usersData?.data || usersData?.users || [];

  const approveMutation = useMutation({
    mutationFn: ({ id, data: actionData }: { id: string; data: Record<string, unknown> }) =>
      api.expenses.advancedApprove(id, actionData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense-stats'] });
      resetAction();
    },
  });

  const delegateMutation = useMutation({
    mutationFn: (data: Record<string, unknown>) => api.expenses.setDelegate(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-delegates'] });
      setDelegateMode(false);
      setDelegateTo('');
      setDelegateStart('');
      setDelegateEnd('');
    },
  });

  const removeDelegateMutation = useMutation({
    mutationFn: (id: string) => api.expenses.removeDelegate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-delegates'] });
    },
  });

  const resetAction = () => {
    setSelectedId(null);
    setAction(null);
    setReason('');
  };

  const handleAction = (id: string, actionType: 'approve' | 'reject' | 'return') => {
    if (actionType === 'approve') {
      approveMutation.mutate({ id, data: { action: 'approve' } });
    } else {
      setSelectedId(id);
      setAction(actionType);
    }
  };

  const submitAction = () => {
    if (!selectedId || !action) return;
    approveMutation.mutate({
      id: selectedId,
      data: {
        action,
        reason: reason.trim(),
      },
    });
  };

  if (!isManager) {
    return (
      <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
        <MaterialIcon name="lock" className="text-6xl text-muted-foreground mb-4" />
        <p className="text-lg text-muted-foreground">{t('common.noPermission')}</p>
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
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{t('expenses.advancedApprove.title')}</h1>
            <p className="text-muted-foreground text-xs sm:text-sm mt-1">{t('expenses.advancedApprove.subtitle')}</p>
          </div>
        </div>
        <button
          onClick={() => setDelegateMode(!delegateMode)}
          className="flex items-center gap-2 px-4 py-2 glass-subtle hover:bg-white/10 rounded-xl text-sm transition-all border border-white/5"
        >
          <MaterialIcon name="swap_horiz" className="text-base" />
          {t('expenses.advancedApprove.delegate')}
        </button>
      </div>

      {/* フロー情報 */}
      {flowConfig.steps && (
        <div className="glass-card rounded-2xl p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <MaterialIcon name="account_tree" className="text-indigo-400" />
            {t('expenses.advancedApprove.flowSteps')}
          </h2>
          <div className="flex items-center gap-2 flex-wrap">
            {(flowConfig.steps as Array<Record<string, unknown>>).map((step, i) => (
              <div key={i} className="flex items-center gap-2">
                <div className="glass-subtle rounded-xl px-4 py-2 text-center">
                  <p className="text-xs text-muted-foreground">{t('expenses.advancedApprove.step', { num: i + 1 })}</p>
                  <p className="text-sm font-semibold">{step.name as string}</p>
                  <p className="text-[10px] text-muted-foreground">{step.approver_role as string}</p>
                  {Number(step.auto_approve_below) > 0 && (
                    <p className="text-[10px] text-emerald-400">
                      ≤¥{Number(step.auto_approve_below).toLocaleString()} {t('expenses.advancedApprove.autoApprove')}
                    </p>
                  )}
                </div>
                {i < (flowConfig.steps as Array<unknown>).length - 1 && (
                  <MaterialIcon name="arrow_forward" className="text-muted-foreground" />
                )}
              </div>
            ))}
            <div className="flex items-center gap-2">
              <MaterialIcon name="arrow_forward" className="text-muted-foreground" />
              <div className="glass-subtle rounded-xl px-4 py-2 text-center border border-emerald-500/20">
                <MaterialIcon name="check_circle" className="text-emerald-400" />
                <p className="text-xs text-emerald-400 font-semibold">{t('expenses.advancedApprove.finalApproval')}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 代理承認設定 */}
      {delegateMode && (
        <div className="glass-card rounded-2xl p-6 space-y-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <MaterialIcon name="person_add" className="text-indigo-400" />
            {t('expenses.advancedApprove.setDelegate')}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">{t('expenses.advancedApprove.delegateTo')}</label>
              <select
                value={delegateTo}
                onChange={(e) => setDelegateTo(e.target.value)}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm"
              >
                <option value="">{t('common.selectPlaceholder')}</option>
                {usersList
                  .filter((u: Record<string, unknown>) => u.id !== user?.id && (u.role === 'admin' || u.role === 'manager'))
                  .map((u: Record<string, unknown>) => (
                    <option key={u.id as string} value={u.id as string}>
                      {u.last_name as string} {u.first_name as string}
                    </option>
                  ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">{t('expenses.advancedApprove.delegateStart')}</label>
              <input type="date" value={delegateStart}
                onChange={(e) => setDelegateStart(e.target.value)}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">{t('expenses.advancedApprove.delegateEnd')}</label>
              <input type="date" value={delegateEnd}
                onChange={(e) => setDelegateEnd(e.target.value)}
                className="w-full px-4 py-2.5 glass-input rounded-xl text-sm" />
            </div>
          </div>
          <div className="flex gap-3">
            <button
              onClick={() => delegateTo && delegateStart && delegateEnd &&
                delegateMutation.mutate({ delegate_to: delegateTo, start_date: delegateStart, end_date: delegateEnd })}
              disabled={!delegateTo || !delegateStart || !delegateEnd || delegateMutation.isPending}
              className="px-5 py-2 gradient-primary text-white text-sm font-semibold rounded-xl hover:shadow-glow-md transition-all disabled:opacity-50"
            >
              {t('common.save')}
            </button>
            <button onClick={() => setDelegateMode(false)}
              className="px-4 py-2 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-all">
              {t('common.cancel')}
            </button>
          </div>
          {/* 既存の代理設定 */}
          {delegates.length > 0 && (
            <div className="mt-4 space-y-2">
              <h3 className="text-sm font-semibold text-muted-foreground">{t('expenses.advancedApprove.currentDelegates')}</h3>
              {delegates.map((d: Record<string, unknown>) => (
                <div key={d.id as string} className="flex items-center justify-between glass-subtle rounded-xl p-3">
                  <div>
                    <p className="text-sm font-medium">{d.delegate_name as string}</p>
                    <p className="text-xs text-muted-foreground">
                      {d.start_date as string} ~ {d.end_date as string}
                    </p>
                  </div>
                  <button
                    onClick={() => removeDelegateMutation.mutate(d.id as string)}
                    className="p-1.5 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors"
                  >
                    <MaterialIcon name="delete" className="text-base" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 承認待ちリスト */}
      <div className="glass-card rounded-2xl p-4 sm:p-6">
        {/* モバイル: カードビュー */}
        <div className="space-y-3 md:hidden">
          {isLoading ? (
            <p className="text-center py-12 text-muted-foreground">{t('common.loading')}</p>
          ) : expenses.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <MaterialIcon name="check_circle" className="text-4xl mb-2 block text-emerald-400 opacity-50" />
              <p>{t('expenses.approve.noPending')}</p>
            </div>
          ) : (
            expenses.map((expense: Record<string, unknown>) => (
              <div key={expense.id as string} className="glass-subtle rounded-xl p-4 space-y-3">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <Link
                      to="/expenses/$expenseId"
                      params={{ expenseId: expense.id as string }}
                      className="text-sm font-semibold text-indigo-400 hover:underline truncate block"
                    >
                      {expense.title as string}
                    </Link>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {expense.user_name as string || '-'} · {expense.created_at ? format(new Date(expense.created_at as string), 'PP', { locale }) : '-'}
                    </p>
                  </div>
                  <span className="text-base font-bold gradient-text whitespace-nowrap">
                    ¥{Number(expense.amount).toLocaleString()}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`inline-block px-2.5 py-1 rounded-full text-xs font-bold ${statusBadge(expense.status as string)}`}>
                    {expense.current_step ? t('expenses.advancedApprove.step', { num: expense.current_step }) : t(`expenses.status.${expense.status}`)}
                  </span>
                </div>
                {selectedId === expense.id as string && action ? (
                  <div className="space-y-2">
                    <textarea
                      value={reason}
                      onChange={(e) => setReason(e.target.value)}
                      className="px-3 py-2 glass-input rounded-lg text-sm w-full"
                      rows={2}
                      placeholder={action === 'reject' ? t('expenses.approve.rejectReason') : t('expenses.advancedApprove.returnReason')}
                    />
                    <div className="flex gap-2">
                      <button
                        onClick={submitAction}
                        disabled={approveMutation.isPending}
                        className={`flex-1 py-2.5 rounded-xl text-sm font-semibold transition-colors ${
                          action === 'reject'
                            ? 'bg-red-500/20 text-red-400 border border-red-500/30 hover:bg-red-500/30'
                            : 'bg-orange-500/20 text-orange-400 border border-orange-500/30 hover:bg-orange-500/30'
                        }`}
                      >
                        {t('common.confirm')}
                      </button>
                      <button onClick={resetAction}
                        className="flex-1 py-2.5 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-colors">
                        {t('common.cancel')}
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleAction(expense.id as string, 'approve')}
                      className="flex-1 py-2.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-xl text-sm font-semibold hover:bg-emerald-500/30 transition-colors"
                    >
                      {t('common.approve')}
                    </button>
                    <button
                      onClick={() => handleAction(expense.id as string, 'return')}
                      className="flex-1 py-2.5 bg-orange-500/20 text-orange-400 border border-orange-500/30 rounded-xl text-sm font-semibold hover:bg-orange-500/30 transition-colors"
                    >
                      {t('expenses.advancedApprove.return')}
                    </button>
                    <button
                      onClick={() => handleAction(expense.id as string, 'reject')}
                      className="flex-1 py-2.5 bg-red-500/20 text-red-400 border border-red-500/30 rounded-xl text-sm font-semibold hover:bg-red-500/30 transition-colors"
                    >
                      {t('common.reject')}
                    </button>
                  </div>
                )}
              </div>
            ))
          )}
        </div>

        {/* デスクトップ: テーブルビュー */}
        <div className="hidden md:block overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5">
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.applicant')}</th>
                <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.title')}</th>
                <th className="text-right py-3 px-4 font-semibold">{t('expenses.fields.amount')}</th>
                <th className="text-center py-3 px-4 font-semibold">{t('expenses.advancedApprove.currentStep')}</th>
                <th className="text-center py-3 px-4 font-semibold">{t('common.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={5} className="text-center py-12 text-muted-foreground">{t('common.loading')}</td>
                </tr>
              ) : expenses.length === 0 ? (
                <tr>
                  <td colSpan={5} className="text-center py-12 text-muted-foreground">
                    <MaterialIcon name="check_circle" className="text-4xl mb-2 block text-emerald-400 opacity-50" />
                    <p>{t('expenses.approve.noPending')}</p>
                  </td>
                </tr>
              ) : (
                expenses.map((expense: Record<string, unknown>) => (
                  <tr key={expense.id as string} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                    <td className="py-3 px-4">
                      <p className="font-medium">{expense.user_name as string || '-'}</p>
                      <p className="text-xs text-muted-foreground">
                        {expense.created_at ? format(new Date(expense.created_at as string), 'PP', { locale }) : '-'}
                      </p>
                    </td>
                    <td className="py-3 px-4">
                      <Link
                        to="/expenses/$expenseId"
                        params={{ expenseId: expense.id as string }}
                        className="text-indigo-400 hover:underline"
                      >
                        {expense.title as string}
                      </Link>
                    </td>
                    <td className="py-3 px-4 text-right font-semibold">
                      ¥{Number(expense.amount).toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-center">
                      <span className={`inline-block px-2.5 py-1 rounded-full text-xs font-bold ${statusBadge(expense.status as string)}`}>
                        {expense.current_step ? t('expenses.advancedApprove.step', { num: expense.current_step }) : t(`expenses.status.${expense.status}`)}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      {selectedId === expense.id as string && action ? (
                        <div className="flex flex-col gap-2">
                          <textarea
                            value={reason}
                            onChange={(e) => setReason(e.target.value)}
                            className="px-2 py-1.5 glass-input rounded-lg text-xs w-full"
                            rows={2}
                            placeholder={action === 'reject' ? t('expenses.approve.rejectReason') : t('expenses.advancedApprove.returnReason')}
                          />
                          <div className="flex gap-1.5">
                            <button
                              onClick={submitAction}
                              disabled={approveMutation.isPending}
                              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors ${
                                action === 'reject'
                                  ? 'bg-red-500/20 text-red-400 border border-red-500/30 hover:bg-red-500/30'
                                  : 'bg-orange-500/20 text-orange-400 border border-orange-500/30 hover:bg-orange-500/30'
                              }`}
                            >
                              {t('common.confirm')}
                            </button>
                            <button onClick={resetAction}
                              className="px-3 py-1.5 glass-subtle rounded-lg text-xs hover:bg-white/10 transition-colors">
                              {t('common.cancel')}
                            </button>
                          </div>
                        </div>
                      ) : (
                        <div className="flex items-center justify-center gap-1.5">
                          <button
                            onClick={() => handleAction(expense.id as string, 'approve')}
                            className="px-3 py-1.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-lg text-xs font-semibold hover:bg-emerald-500/30 transition-colors"
                          >
                            {t('common.approve')}
                          </button>
                          <button
                            onClick={() => handleAction(expense.id as string, 'return')}
                            className="px-3 py-1.5 bg-orange-500/20 text-orange-400 border border-orange-500/30 rounded-lg text-xs font-semibold hover:bg-orange-500/30 transition-colors"
                          >
                            {t('expenses.advancedApprove.return')}
                          </button>
                          <button
                            onClick={() => handleAction(expense.id as string, 'reject')}
                            className="px-3 py-1.5 bg-red-500/20 text-red-400 border border-red-500/30 rounded-lg text-xs font-semibold hover:bg-red-500/30 transition-colors"
                          >
                            {t('common.reject')}
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="mt-4">
            <Pagination
              currentPage={page}
              totalPages={totalPages}
              totalItems={totalItems}
              pageSize={pageSize}
              onPageChange={setPage}
              onPageSizeChange={(size) => { setPageSize(size); setPage(1); }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
