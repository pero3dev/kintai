package model

import (
	"time"

	"github.com/google/uuid"
)

// ===== 経費申請 DTO =====

// ExpenseCreateRequest は経費申請作成リクエスト
type ExpenseCreateRequest struct {
	Title  string               `json:"title" validate:"required"`
	Status string               `json:"status"`
	Notes  string               `json:"notes"`
	Items  []ExpenseItemRequest `json:"items" validate:"required,min=1"`
}

// ExpenseItemRequest は経費明細リクエスト
type ExpenseItemRequest struct {
	ExpenseDate string  `json:"expense_date" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	ReceiptURL  string  `json:"receipt_url"`
}

// ExpenseUpdateRequest は経費申請更新リクエスト
type ExpenseUpdateRequest struct {
	Title  *string              `json:"title"`
	Status *string              `json:"status"`
	Notes  *string              `json:"notes"`
	Items  []ExpenseItemRequest `json:"items"`
}

// ExpenseApproveRequest は経費承認リクエスト
type ExpenseApproveRequest struct {
	Status         string `json:"status" validate:"required"`
	RejectedReason string `json:"rejected_reason"`
}

// ExpenseAdvancedApproveRequest は高度な経費承認リクエスト
type ExpenseAdvancedApproveRequest struct {
	Action string `json:"action" validate:"required"`
	Reason string `json:"reason"`
	Step   int    `json:"step"`
}

// ===== 経費統計 DTO =====

// ExpenseStatsResponse は経費統計レスポンス
type ExpenseStatsResponse struct {
	TotalThisMonth    float64 `json:"total_this_month"`
	PendingCount      int64   `json:"pending_count"`
	ApprovedThisMonth float64 `json:"approved_this_month"`
	ReimbursedTotal   float64 `json:"reimbursed_total"`
}

// ===== 経費レポート DTO =====

// ExpenseReportResponse は経費レポートレスポンス
type ExpenseReportResponse struct {
	TotalAmount         float64                    `json:"total_amount"`
	CategoryBreakdown   []CategoryBreakdownItem    `json:"category_breakdown"`
	DepartmentBreakdown []DepartmentBreakdownItem  `json:"department_breakdown"`
	StatusSummary       ExpenseStatusSummary       `json:"status_summary"`
}

// CategoryBreakdownItem はカテゴリ内訳
type CategoryBreakdownItem struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

// DepartmentBreakdownItem は部署内訳
type DepartmentBreakdownItem struct {
	Department string  `json:"department"`
	Amount     float64 `json:"amount"`
	Count      int     `json:"count"`
	Avg        float64 `json:"avg"`
}

// ExpenseStatusSummary はステータス集計
type ExpenseStatusSummary struct {
	Draft      int `json:"draft"`
	Pending    int `json:"pending"`
	Approved   int `json:"approved"`
	Rejected   int `json:"rejected"`
	Reimbursed int `json:"reimbursed"`
}

// MonthlyTrendItem は月別トレンド項目
type MonthlyTrendItem struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

// ===== 経費コメント DTO =====

// ExpenseCommentRequest はコメント作成リクエスト
type ExpenseCommentRequest struct {
	Content string `json:"content" validate:"required"`
}

// ExpenseCommentResponse はコメントレスポンス
type ExpenseCommentResponse struct {
	ID        uuid.UUID `json:"id"`
	ExpenseID uuid.UUID `json:"expense_id"`
	UserID    uuid.UUID `json:"user_id"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ===== 経費履歴 DTO =====

// ExpenseHistoryResponse は変更履歴レスポンス
type ExpenseHistoryResponse struct {
	ID        uuid.UUID `json:"id"`
	ExpenseID uuid.UUID `json:"expense_id"`
	Action    string    `json:"action"`
	ChangedBy string    `json:"changed_by"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

// ===== レシートアップロード DTO =====

// ReceiptUploadResponse はレシートアップロードレスポンス
type ReceiptUploadResponse struct {
	URL string `json:"url"`
}

// ===== テンプレート DTO =====

// ExpenseTemplateRequest はテンプレート作成/更新リクエスト
type ExpenseTemplateRequest struct {
	Name         string  `json:"name" validate:"required"`
	Title        string  `json:"title"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	IsRecurring  bool    `json:"is_recurring"`
	RecurringDay int     `json:"recurring_day"`
}

// ===== ポリシー DTO =====

// ExpensePolicyRequest はポリシー作成/更新リクエスト
type ExpensePolicyRequest struct {
	Category             string  `json:"category" validate:"required"`
	MonthlyLimit         float64 `json:"monthly_limit"`
	PerClaimLimit        float64 `json:"per_claim_limit"`
	AutoApproveLimit     float64 `json:"auto_approve_limit"`
	RequiresReceiptAbove float64 `json:"requires_receipt_above"`
	Description          string  `json:"description"`
	IsActive             *bool   `json:"is_active"`
}

// ===== 通知 DTO =====

// ExpenseNotificationResponse は経費通知レスポンス
type ExpenseNotificationResponse struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	Message   string     `json:"message"`
	IsRead    bool       `json:"is_read"`
	ExpenseID *uuid.UUID `json:"expense_id"`
	CreatedAt time.Time  `json:"created_at"`
}

// ===== 代理承認 DTO =====

// ExpenseDelegateRequest は代理承認リクエスト
type ExpenseDelegateRequest struct {
	DelegateTo string `json:"delegate_to" validate:"required"`
	StartDate  string `json:"start_date" validate:"required"`
	EndDate    string `json:"end_date" validate:"required"`
}

// ===== 通知設定 DTO =====

// ExpenseNotificationSettingRequest は通知設定更新リクエスト
type ExpenseNotificationSettingRequest struct {
	EmailEnabled        *bool `json:"email_enabled"`
	PushEnabled         *bool `json:"push_enabled"`
	ApprovalAlerts      *bool `json:"approval_alerts"`
	ReimbursementAlerts *bool `json:"reimbursement_alerts"`
	PolicyAlerts        *bool `json:"policy_alerts"`
	WeeklyDigest        *bool `json:"weekly_digest"`
}

// ===== 経費一覧用 レスポンス =====

// ExpenseListItem は一覧表示用の経費レスポンス
type ExpenseListItem struct {
	ID          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	Status      ExpenseStatus `json:"status"`
	TotalAmount float64       `json:"total_amount"`
	UserID      uuid.UUID     `json:"user_id"`
	UserName    string        `json:"user_name"`
	Category    string        `json:"category"`
	ExpenseDate string        `json:"expense_date"`
	Notes       string        `json:"notes"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
