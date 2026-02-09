package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ===== 経費申請ステータス =====

type ExpenseStatus string

const (
	ExpenseStatusDraft         ExpenseStatus = "draft"
	ExpenseStatusPending       ExpenseStatus = "pending"
	ExpenseStatusApproved      ExpenseStatus = "approved"
	ExpenseStatusRejected      ExpenseStatus = "rejected"
	ExpenseStatusReimbursed    ExpenseStatus = "reimbursed"
	ExpenseStatusStep1Approved ExpenseStatus = "step1_approved"
	ExpenseStatusStep2Approved ExpenseStatus = "step2_approved"
	ExpenseStatusReturned      ExpenseStatus = "returned"
)

// ===== 経費カテゴリ =====

type ExpenseCategory string

const (
	ExpenseCategoryTransportation ExpenseCategory = "transportation"
	ExpenseCategoryMeals          ExpenseCategory = "meals"
	ExpenseCategoryAccommodation  ExpenseCategory = "accommodation"
	ExpenseCategorySupplies       ExpenseCategory = "supplies"
	ExpenseCategoryCommunication  ExpenseCategory = "communication"
	ExpenseCategoryEntertainment  ExpenseCategory = "entertainment"
	ExpenseCategoryOther          ExpenseCategory = "other"
)

// ===== 経費申請 =====

// Expense は経費申請モデル
type Expense struct {
	BaseModel
	UserID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"user_id"`
	Title          string        `gorm:"size:200;not null" json:"title"`
	Status         ExpenseStatus `gorm:"size:30;not null;default:'draft'" json:"status"`
	Notes          string        `gorm:"size:2000" json:"notes"`
	TotalAmount    float64       `gorm:"not null;default:0" json:"total_amount"`
	ApprovedBy     *uuid.UUID    `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt     *time.Time    `json:"approved_at"`
	RejectedReason string        `gorm:"size:500" json:"rejected_reason"`
	ReimbursedAt   *time.Time    `json:"reimbursed_at"`

	User     *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User         `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	Items    []ExpenseItem `gorm:"foreignKey:ExpenseID" json:"items,omitempty"`
}

// ===== 経費明細 =====

// ExpenseItem は経費明細モデル
type ExpenseItem struct {
	BaseModel
	ExpenseID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"expense_id"`
	ExpenseDate time.Time       `gorm:"type:date;not null" json:"expense_date"`
	Category    ExpenseCategory `gorm:"size:30;not null" json:"category"`
	Description string          `gorm:"size:500;not null" json:"description"`
	Amount      float64         `gorm:"not null" json:"amount"`
	ReceiptURL  string          `gorm:"size:500" json:"receipt_url"`

	Expense *Expense `gorm:"foreignKey:ExpenseID" json:"-"`
}

// ===== 経費コメント =====

// ExpenseComment は経費申請コメントモデル
type ExpenseComment struct {
	BaseModel
	ExpenseID uuid.UUID `gorm:"type:uuid;not null;index" json:"expense_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Content   string    `gorm:"size:2000;not null" json:"content"`

	Expense *Expense `gorm:"foreignKey:ExpenseID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 経費変更履歴 =====

// ExpenseHistory は経費変更履歴モデル
type ExpenseHistory struct {
	BaseModel
	ExpenseID  uuid.UUID `gorm:"type:uuid;not null;index" json:"expense_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Action     string    `gorm:"size:100;not null" json:"action"`
	OldValue   string    `gorm:"size:2000" json:"old_value"`
	NewValue   string    `gorm:"size:2000" json:"new_value"`
	ChangedBy  string    `gorm:"size:200" json:"changed_by"`

	Expense *Expense `gorm:"foreignKey:ExpenseID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 経費テンプレート =====

// ExpenseTemplate は経費テンプレートモデル
type ExpenseTemplate struct {
	BaseModel
	UserID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	Name         string          `gorm:"size:200;not null" json:"name"`
	Title        string          `gorm:"size:200" json:"title"`
	Category     ExpenseCategory `gorm:"size:30" json:"category"`
	Description  string          `gorm:"size:500" json:"description"`
	Amount       float64         `gorm:"default:0" json:"amount"`
	IsRecurring  bool            `gorm:"default:false" json:"is_recurring"`
	RecurringDay int             `gorm:"default:0" json:"recurring_day"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 経費ポリシー =====

// ExpensePolicy は経費ポリシーモデル
type ExpensePolicy struct {
	BaseModel
	Category            ExpenseCategory `gorm:"size:30;not null" json:"category"`
	MonthlyLimit        float64         `gorm:"default:0" json:"monthly_limit"`
	PerClaimLimit       float64         `gorm:"default:0" json:"per_claim_limit"`
	AutoApproveLimit    float64         `gorm:"default:0" json:"auto_approve_limit"`
	RequiresReceiptAbove float64        `gorm:"default:0" json:"requires_receipt_above"`
	Description         string          `gorm:"size:500" json:"description"`
	IsActive            bool            `gorm:"default:true" json:"is_active"`
}

// ===== 経費予算 =====

// ExpenseBudget は経費予算モデル
type ExpenseBudget struct {
	BaseModel
	DepartmentID *uuid.UUID      `gorm:"type:uuid" json:"department_id"`
	Category     ExpenseCategory `gorm:"size:30" json:"category"`
	FiscalYear   int             `gorm:"not null" json:"fiscal_year"`
	BudgetAmount float64         `gorm:"not null;default:0" json:"budget_amount"`
	SpentAmount  float64         `gorm:"not null;default:0" json:"spent_amount"`
}

// ===== 経費通知 =====

// ExpenseNotification は経費通知モデル
type ExpenseNotification struct {
	BaseModel
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ExpenseID *uuid.UUID `gorm:"type:uuid" json:"expense_id"`
	Type      string     `gorm:"size:50;not null" json:"type"`
	Message   string     `gorm:"size:1000;not null" json:"message"`
	IsRead    bool       `gorm:"default:false" json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Expense *Expense `gorm:"foreignKey:ExpenseID" json:"expense,omitempty"`
}

// ===== 経費リマインダー =====

// ExpenseReminder は経費リマインダーモデル
type ExpenseReminder struct {
	BaseModel
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ExpenseID   *uuid.UUID `gorm:"type:uuid" json:"expense_id"`
	Message     string     `gorm:"size:500;not null" json:"message"`
	DueDate     *time.Time `gorm:"type:date" json:"due_date"`
	IsDismissed bool       `gorm:"default:false" json:"is_dismissed"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 経費通知設定 =====

// ExpenseNotificationSetting は経費通知設定モデル
type ExpenseNotificationSetting struct {
	BaseModel
	UserID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	EmailEnabled      bool      `gorm:"default:true" json:"email_enabled"`
	PushEnabled       bool      `gorm:"default:true" json:"push_enabled"`
	ApprovalAlerts    bool      `gorm:"default:true" json:"approval_alerts"`
	ReimbursementAlerts bool    `gorm:"default:true" json:"reimbursement_alerts"`
	PolicyAlerts      bool      `gorm:"default:true" json:"policy_alerts"`
	WeeklyDigest      bool      `gorm:"default:false" json:"weekly_digest"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 経費承認フロー設定 =====

// ExpenseApprovalFlow は経費承認フロー設定モデル
type ExpenseApprovalFlow struct {
	BaseModel
	Name            string `gorm:"size:200;not null" json:"name"`
	MinAmount       float64 `gorm:"default:0" json:"min_amount"`
	MaxAmount       float64 `gorm:"default:0" json:"max_amount"`
	RequiredSteps   int     `gorm:"default:1" json:"required_steps"`
	IsActive        bool    `gorm:"default:true" json:"is_active"`
	AutoApproveBelow float64 `gorm:"default:0" json:"auto_approve_below"`
}

// ===== 経費代理承認 =====

// ExpenseDelegate は経費代理承認モデル
type ExpenseDelegate struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	DelegateID uuid.UUID `gorm:"type:uuid;not null" json:"delegate_to"`
	StartDate  time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:date;not null" json:"end_date"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`

	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Delegate *User `gorm:"foreignKey:DelegateID" json:"delegate,omitempty"`
}

// ===== 経費ポリシー違反 =====

// ExpensePolicyViolation は経費ポリシー違反モデル
type ExpensePolicyViolation struct {
	BaseModel
	ExpenseID uuid.UUID `gorm:"type:uuid;not null;index" json:"expense_id"`
	PolicyID  uuid.UUID `gorm:"type:uuid;not null" json:"policy_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Reason    string    `gorm:"size:500;not null" json:"reason"`
	Severity  string    `gorm:"size:20;not null;default:'warning'" json:"severity"`

	Expense *Expense        `gorm:"foreignKey:ExpenseID" json:"expense,omitempty"`
	Policy  *ExpensePolicy  `gorm:"foreignKey:PolicyID" json:"policy,omitempty"`
	User    *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ExpenseAutoMigrate は経費関連テーブルのマイグレーションを実行する
func ExpenseAutoMigrate(db *gorm.DB) error {
	// FK制約の自動生成を無効化してマイグレーション
	prev := db.Config.DisableForeignKeyConstraintWhenMigrating
	db.Config.DisableForeignKeyConstraintWhenMigrating = true
	defer func() { db.Config.DisableForeignKeyConstraintWhenMigrating = prev }()

	models := []interface{}{
		&Expense{},
		&ExpenseItem{},
		&ExpenseComment{},
		&ExpenseHistory{},
		&ExpenseTemplate{},
		&ExpensePolicy{},
		&ExpenseBudget{},
		&ExpenseNotification{},
		&ExpenseReminder{},
		&ExpenseNotificationSetting{},
		&ExpenseApprovalFlow{},
		&ExpenseDelegate{},
		&ExpensePolicyViolation{},
	}

	m := db.Migrator()
	for _, model := range models {
		if !m.HasTable(model) {
			if err := m.CreateTable(model); err != nil {
				return err
			}
		}
	}
	return nil
}
