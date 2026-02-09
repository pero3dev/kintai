import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useParams, useNavigate } from '@tanstack/react-router';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';
import { format } from 'date-fns';
import { ja, enUS } from 'date-fns/locale';

function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

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

const statusBadge = (status: string) => {
  const styles: Record<string, string> = {
    draft: 'bg-gray-500/20 text-gray-400 border border-gray-500/30',
    pending: 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30',
    approved: 'bg-green-500/20 text-green-400 border border-green-500/30',
    rejected: 'bg-red-500/20 text-red-400 border border-red-500/30',
    reimbursed: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
  };
  return styles[status] || styles.draft;
};

const statusIcon = (status: string) => {
  const icons: Record<string, string> = {
    draft: 'edit_note',
    pending: 'pending_actions',
    approved: 'check_circle',
    rejected: 'cancel',
    reimbursed: 'account_balance_wallet',
  };
  return icons[status] || 'help';
};

export function ExpenseDetailPage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const locale = i18n.language === 'ja' ? ja : enUS;
  const { expenseId } = useParams({ strict: false }) as { expenseId: string };

  const [newComment, setNewComment] = useState('');
  const [activeTab, setActiveTab] = useState<'details' | 'comments' | 'history'>('details');

  const { data: expense, isLoading } = useQuery({
    queryKey: ['expense', expenseId],
    queryFn: () => api.expenses.getByID(expenseId),
    enabled: !!expenseId,
  });

  const { data: commentsData } = useQuery({
    queryKey: ['expense-comments', expenseId],
    queryFn: () => api.expenses.getComments(expenseId),
    enabled: !!expenseId,
  });

  const { data: historyData } = useQuery({
    queryKey: ['expense-history', expenseId],
    queryFn: () => api.expenses.getHistory(expenseId),
    enabled: !!expenseId,
  });

  const comments = commentsData?.data || commentsData?.comments || [];
  const history = historyData?.data || historyData?.history || [];

  const addCommentMutation = useMutation({
    mutationFn: (content: string) => api.expenses.addComment(expenseId, { content }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-comments', expenseId] });
      setNewComment('');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => api.expenses.delete(expenseId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      navigate({ to: '/expenses/history' });
    },
  });

  const submitMutation = useMutation({
    mutationFn: () => api.expenses.update(expenseId, { status: 'pending' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense', expenseId] });
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
    },
  });

  const isManager = user?.role === 'admin' || user?.role === 'manager';
  const items = expense?.items || [];
  const totalAmount = items.reduce((sum: number, item: Record<string, unknown>) => sum + (Number(item.amount) || 0), 0);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 animate-fade-in">
        <MaterialIcon name="hourglass_empty" className="text-4xl text-muted-foreground animate-spin" />
      </div>
    );
  }

  if (!expense) {
    return (
      <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
        <MaterialIcon name="search_off" className="text-6xl text-muted-foreground mb-4" />
        <p className="text-lg text-muted-foreground">{t('common.noData')}</p>
        <Link to="/expenses/history" className="mt-4 text-indigo-400 hover:underline text-sm">
          {t('expenses.history.title')}
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto space-y-6 animate-fade-in">
      {/* ヘッダー */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate({ to: '/expenses/history' })}
            className="p-2 rounded-xl glass-subtle hover:bg-white/10 transition-all"
          >
            <MaterialIcon name="arrow_back" />
          </button>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold gradient-text">{expense.title}</h1>
            <div className="flex items-center gap-2 sm:gap-3 mt-1 flex-wrap">
              <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold ${statusBadge(expense.status)}`}>
                <MaterialIcon name={statusIcon(expense.status)} className="text-sm" />
                {t(`expenses.status.${expense.status}`)}
              </span>
              {expense.created_at && (
                <span className="text-xs text-muted-foreground">
                  {format(new Date(expense.created_at), 'PPp', { locale })}
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          {expense.status === 'draft' && (
            <>
              <button
                onClick={() => submitMutation.mutate()}
                disabled={submitMutation.isPending}
                className="flex items-center gap-2 px-4 py-2 gradient-primary text-white font-semibold text-sm rounded-xl hover:shadow-glow-md transition-all"
              >
                <MaterialIcon name="send" className="text-base" />
                {t('expenses.actions.submit')}
              </button>
              <Link
                to={"/expenses/new" as string}
                search={{ edit: expenseId }}
                className="flex items-center gap-2 px-4 py-2 glass-subtle hover:bg-white/10 rounded-xl text-sm transition-all border border-white/5"
              >
                <MaterialIcon name="edit" className="text-base" />
                {t('common.edit')}
              </Link>
              <button
                onClick={() => {
                  if (confirm(t('expenses.detail.deleteConfirm'))) {
                    deleteMutation.mutate();
                  }
                }}
                className="flex items-center gap-2 px-4 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-xl text-sm transition-all border border-red-500/20"
              >
                <MaterialIcon name="delete" className="text-base" />
                {t('common.delete')}
              </button>
            </>
          )}
        </div>
      </div>

      {/* サマリーカード */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-indigo-500/20 flex items-center justify-center">
              <MaterialIcon name="receipt_long" className="text-indigo-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.fields.totalAmount')}</p>
          </div>
          <p className="text-2xl font-bold gradient-text">¥{totalAmount.toLocaleString()}</p>
        </div>
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <MaterialIcon name="format_list_numbered" className="text-emerald-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.detail.itemCount')}</p>
          </div>
          <p className="text-2xl font-bold gradient-text">{items.length}</p>
        </div>
        <div className="glass-card stat-shimmer rounded-2xl p-5">
          <div className="flex items-center gap-3 mb-2">
            <div className="size-10 rounded-xl bg-amber-500/20 flex items-center justify-center">
              <MaterialIcon name="person" className="text-amber-400" />
            </div>
            <p className="text-sm text-muted-foreground">{t('expenses.fields.applicant')}</p>
          </div>
          <p className="text-lg font-bold">{expense.user_name || '-'}</p>
        </div>
      </div>

      {/* タブ */}
      <div className="glass-card rounded-2xl overflow-hidden">
        <div className="flex border-b border-white/5">
          {(['details', 'comments', 'history'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-3.5 text-sm font-medium transition-all ${
                activeTab === tab
                  ? 'border-b-2 border-indigo-400 text-indigo-400 bg-indigo-500/5'
                  : 'text-muted-foreground hover:text-foreground hover:bg-white/5'
              }`}
            >
              <MaterialIcon
                name={tab === 'details' ? 'list_alt' : tab === 'comments' ? 'chat' : 'history'}
                className="text-base"
              />
              {t(`expenses.detail.tabs.${tab}`)}
              {tab === 'comments' && comments.length > 0 && (
                <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-indigo-500/20 text-indigo-400 font-bold">
                  {comments.length}
                </span>
              )}
            </button>
          ))}
        </div>

        <div className="p-6">
          {/* 明細タブ */}
          {activeTab === 'details' && (
            <div className="space-y-4">
              {items.length === 0 ? (
                <p className="text-center py-8 text-muted-foreground">{t('common.noData')}</p>
              ) : (
                <>
                  {/* デスクトップテーブル */}
                  <div className="hidden md:block overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-white/5">
                          <th className="text-left py-3 px-4 font-semibold">#</th>
                          <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.date')}</th>
                          <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.category')}</th>
                          <th className="text-left py-3 px-4 font-semibold">{t('expenses.fields.description')}</th>
                          <th className="text-right py-3 px-4 font-semibold">{t('expenses.fields.amount')}</th>
                          <th className="text-center py-3 px-4 font-semibold">{t('expenses.detail.receipt')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {items.map((item: Record<string, unknown>, index: number) => (
                          <tr key={index} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                            <td className="py-3 px-4 text-muted-foreground">{index + 1}</td>
                            <td className="py-3 px-4">
                              {item.expense_date ? format(new Date(item.expense_date as string), 'PP', { locale }) : '-'}
                            </td>
                            <td className="py-3 px-4">
                              <div className="flex items-center gap-2">
                                <MaterialIcon name={getCategoryIcon(item.category as string)} className="text-indigo-400 text-base" />
                                {item.category ? t(`expenses.categories.${item.category}`) : '-'}
                              </div>
                            </td>
                            <td className="py-3 px-4">{(item.description as string) || '-'}</td>
                            <td className="py-3 px-4 text-right font-semibold">¥{Number(item.amount).toLocaleString()}</td>
                            <td className="py-3 px-4 text-center">
                              {item.receipt_url ? (
                                <a href={item.receipt_url as string} target="_blank" rel="noopener noreferrer"
                                  className="text-indigo-400 hover:underline inline-flex items-center gap-1">
                                  <MaterialIcon name="image" className="text-base" />
                                </a>
                              ) : (
                                <span className="text-muted-foreground text-xs">-</span>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                      <tfoot>
                        <tr className="border-t border-white/10">
                          <td colSpan={4} className="py-3 px-4 text-right font-semibold">{t('expenses.fields.totalAmount')}</td>
                          <td className="py-3 px-4 text-right font-bold text-lg gradient-text">¥{totalAmount.toLocaleString()}</td>
                          <td></td>
                        </tr>
                      </tfoot>
                    </table>
                  </div>

                  {/* モバイルカード */}
                  <div className="md:hidden space-y-3">
                    {items.map((item: Record<string, unknown>, index: number) => (
                      <div key={index} className="glass-subtle rounded-xl p-4">
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center gap-2">
                            <MaterialIcon name={getCategoryIcon(item.category as string)} className="text-indigo-400" />
                            <span className="text-sm font-medium">
                              {item.category ? t(`expenses.categories.${item.category}`) : '-'}
                            </span>
                          </div>
                          <span className="font-bold">¥{Number(item.amount).toLocaleString()}</span>
                        </div>
                        <p className="text-sm text-muted-foreground">{(item.description as string) || '-'}</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          {item.expense_date ? format(new Date(item.expense_date as string), 'PP', { locale }) : ''}
                        </p>
                        {!!item.receipt_url && (
                          <a href={item.receipt_url as string} target="_blank" rel="noopener noreferrer"
                            className="mt-2 inline-flex items-center gap-1 text-xs text-indigo-400 hover:underline">
                            <MaterialIcon name="image" className="text-sm" />
                            {t('expenses.detail.viewReceipt')}
                          </a>
                        )}
                      </div>
                    ))}
                    <div className="flex items-center justify-between pt-3 border-t border-white/5">
                      <span className="font-semibold">{t('expenses.fields.totalAmount')}</span>
                      <span className="text-xl font-bold gradient-text">¥{totalAmount.toLocaleString()}</span>
                    </div>
                  </div>
                </>
              )}

              {/* 備考 */}
              {expense.notes && (
                <div className="mt-4 glass-subtle rounded-xl p-4">
                  <h3 className="text-sm font-semibold mb-2 flex items-center gap-2">
                    <MaterialIcon name="notes" className="text-indigo-400 text-base" />
                    {t('expenses.fields.notes')}
                  </h3>
                  <p className="text-sm text-muted-foreground whitespace-pre-wrap">{expense.notes}</p>
                </div>
              )}

              {/* 却下理由 */}
              {expense.status === 'rejected' && expense.rejected_reason && (
                <div className="mt-4 glass-subtle rounded-xl p-4 border border-red-500/20">
                  <h3 className="text-sm font-semibold mb-2 flex items-center gap-2 text-red-400">
                    <MaterialIcon name="error" className="text-base" />
                    {t('expenses.detail.rejectionReason')}
                  </h3>
                  <p className="text-sm text-muted-foreground">{expense.rejected_reason}</p>
                </div>
              )}
            </div>
          )}

          {/* コメントタブ */}
          {activeTab === 'comments' && (
            <div className="space-y-4">
              {/* コメント入力 */}
              <div className="glass-subtle rounded-xl p-4">
                <div className="flex gap-3">
                  <div className="size-9 rounded-full flex items-center justify-center flex-shrink-0 gradient-primary-subtle border border-indigo-400/20">
                    <MaterialIcon name="person" className="text-indigo-400 text-sm" />
                  </div>
                  <div className="flex-1">
                    <textarea
                      value={newComment}
                      onChange={(e) => setNewComment(e.target.value)}
                      rows={2}
                      className="w-full px-3 py-2 glass-input rounded-xl text-sm resize-none"
                      placeholder={t('expenses.detail.commentPlaceholder')}
                    />
                    <div className="flex justify-end mt-2">
                      <button
                        onClick={() => newComment.trim() && addCommentMutation.mutate(newComment.trim())}
                        disabled={!newComment.trim() || addCommentMutation.isPending}
                        className="flex items-center gap-1.5 px-4 py-1.5 gradient-primary text-white text-xs font-semibold rounded-lg hover:shadow-glow-sm transition-all disabled:opacity-50"
                      >
                        <MaterialIcon name="send" className="text-sm" />
                        {t('expenses.detail.addComment')}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              {/* コメント一覧 */}
              {comments.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <MaterialIcon name="chat_bubble_outline" className="text-4xl mb-2 block opacity-50" />
                  <p className="text-sm">{t('expenses.detail.noComments')}</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {comments.map((comment: Record<string, unknown>) => (
                    <div key={comment.id as string} className="flex gap-3">
                      <div className="size-9 rounded-full flex items-center justify-center flex-shrink-0 bg-indigo-500/10 border border-indigo-400/20">
                        <MaterialIcon name="person" className="text-indigo-400 text-sm" />
                      </div>
                      <div className="flex-1 glass-subtle rounded-xl p-3">
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-sm font-semibold">{comment.user_name as string}</span>
                          <span className="text-xs text-muted-foreground">
                            {comment.created_at ? format(new Date(comment.created_at as string), 'PPp', { locale }) : ''}
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground whitespace-pre-wrap">{comment.content as string}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 履歴タブ */}
          {activeTab === 'history' && (
            <div className="space-y-1">
              {history.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                  <MaterialIcon name="history" className="text-4xl mb-2 block opacity-50" />
                  <p className="text-sm">{t('expenses.detail.noHistory')}</p>
                </div>
              ) : (
                <div className="relative pl-6">
                  <div className="absolute left-[11px] top-2 bottom-2 w-0.5 bg-white/10" />
                  {history.map((entry: Record<string, unknown>, index: number) => (
                    <div key={index} className="relative flex gap-4 py-3">
                      <div className="absolute left-[-13px] top-4 size-3 rounded-full bg-indigo-400 border-2 border-background z-10" />
                      <div className="flex-1">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="text-sm font-medium">{entry.action as string}</span>
                          {!!entry.user_name && (
                            <span className="text-xs text-muted-foreground">by {entry.user_name as string}</span>
                          )}
                        </div>
                        {!!entry.details && (
                          <p className="text-xs text-muted-foreground mt-0.5">{entry.details as string}</p>
                        )}
                        <p className="text-xs text-muted-foreground mt-1">
                          {entry.created_at ? format(new Date(entry.created_at as string), 'PPp', { locale }) : ''}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 承認アクション（管理者用） */}
      {isManager && expense.status === 'pending' && (
        <ApprovalActions expenseId={expenseId} />
      )}
    </div>
  );
}

function ApprovalActions({ expenseId }: { expenseId: string }) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [showReject, setShowReject] = useState(false);
  const [rejectReason, setRejectReason] = useState('');

  const approveMutation = useMutation({
    mutationFn: () => api.expenses.approve(expenseId, { status: 'approved' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense', expenseId] });
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
    },
  });

  const rejectMutation = useMutation({
    mutationFn: (reason: string) => api.expenses.approve(expenseId, { status: 'rejected', rejected_reason: reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense', expenseId] });
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      setShowReject(false);
    },
  });

  return (
    <div className="glass-card rounded-2xl p-6">
      <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
        <MaterialIcon name="gavel" className="text-indigo-400" />
        {t('expenses.detail.approvalAction')}
      </h3>
      {showReject ? (
        <div className="space-y-3">
          <textarea
            value={rejectReason}
            onChange={(e) => setRejectReason(e.target.value)}
            rows={2}
            className="w-full px-4 py-2.5 glass-input rounded-xl resize-none text-sm"
            placeholder={t('expenses.approve.rejectReason')}
          />
          <div className="flex gap-3">
            <button
              onClick={() => rejectReason.trim() && rejectMutation.mutate(rejectReason.trim())}
              disabled={!rejectReason.trim() || rejectMutation.isPending}
              className="px-4 py-2 bg-red-500/20 text-red-400 border border-red-500/30 rounded-xl text-sm font-semibold hover:bg-red-500/30 transition-colors"
            >
              {t('expenses.detail.confirmReject')}
            </button>
            <button
              onClick={() => { setShowReject(false); setRejectReason(''); }}
              className="px-4 py-2 glass-subtle rounded-xl text-sm hover:bg-white/10 transition-colors"
            >
              {t('common.cancel')}
            </button>
          </div>
        </div>
      ) : (
        <div className="flex gap-3">
          <button
            onClick={() => approveMutation.mutate()}
            disabled={approveMutation.isPending}
            className="flex items-center gap-2 px-5 py-2.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-xl text-sm font-semibold hover:bg-emerald-500/30 transition-colors"
          >
            <MaterialIcon name="check_circle" className="text-base" />
            {t('common.approve')}
          </button>
          <button
            onClick={() => setShowReject(true)}
            className="flex items-center gap-2 px-5 py-2.5 bg-red-500/20 text-red-400 border border-red-500/30 rounded-xl text-sm font-semibold hover:bg-red-500/30 transition-colors"
          >
            <MaterialIcon name="cancel" className="text-base" />
            {t('common.reject')}
          </button>
        </div>
      )}
    </div>
  );
}
