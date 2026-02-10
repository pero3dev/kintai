package expense

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/gorm"
)

// ===== ExpenseRepository =====

type ExpenseRepository interface {
	Create(ctx context.Context, expense *model.Expense) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Expense, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error)
	FindPending(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error)
	FindAll(ctx context.Context, page, pageSize int, status, category string) ([]model.Expense, int64, error)
	Update(ctx context.Context, expense *model.Expense) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error)
	GetReport(ctx context.Context, startDate, endDate time.Time) (*model.ExpenseReportResponse, error)
	GetMonthlyTrend(ctx context.Context, year int) ([]model.MonthlyTrendItem, error)
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(ctx context.Context, expense *model.Expense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *expenseRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Expense, error) {
	var expense model.Expense
	err := r.db.WithContext(ctx).Preload("Items").Preload("User").Where("id = ?", id).First(&expense).Error
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
	var expenses []model.Expense
	var total int64
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if category != "" {
		q = q.Joins("JOIN expense_items ON expense_items.expense_id = expenses.id AND expense_items.category = ?", category)
	}
	q.Model(&model.Expense{}).Count(&total)
	err := q.Preload("Items").Preload("User").Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&expenses).Error
	return expenses, total, err
}

func (r *expenseRepository) FindPending(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error) {
	var expenses []model.Expense
	var total int64
	q := r.db.WithContext(ctx).Where("status = ?", model.ExpenseStatusPending)
	q.Model(&model.Expense{}).Count(&total)
	err := q.Preload("Items").Preload("User").Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&expenses).Error
	return expenses, total, err
}

func (r *expenseRepository) FindAll(ctx context.Context, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
	var expenses []model.Expense
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Expense{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if category != "" {
		q = q.Joins("JOIN expense_items ON expense_items.expense_id = expenses.id AND expense_items.category = ?", category)
	}
	q.Count(&total)
	err := r.db.WithContext(ctx).Preload("Items").Preload("User").
		Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&expenses).Error
	if status != "" {
		err = r.db.WithContext(ctx).Preload("Items").Preload("User").Where("status = ?", status).
			Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&expenses).Error
	}
	return expenses, total, err
}

func (r *expenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	return r.db.WithContext(ctx).Save(expense).Error
}

func (r *expenseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// 明細も削除
	r.db.WithContext(ctx).Where("expense_id = ?", id).Delete(&model.ExpenseItem{})
	return r.db.WithContext(ctx).Delete(&model.Expense{}, "id = ?", id).Error
}

func (r *expenseRepository) GetStats(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var stats model.ExpenseStatsResponse

	// 今月の申請合計
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("user_id = ? AND created_at >= ?", userID, startOfMonth).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TotalThisMonth)

	// 承認待ち件数
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("user_id = ? AND status = ?", userID, model.ExpenseStatusPending).
		Count(&stats.PendingCount)

	// 今月の承認済み合計
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("user_id = ? AND status = ? AND updated_at >= ?", userID, model.ExpenseStatusApproved, startOfMonth).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.ApprovedThisMonth)

	// 精算済み合計
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("user_id = ? AND status = ?", userID, model.ExpenseStatusReimbursed).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.ReimbursedTotal)

	return &stats, nil
}

func (r *expenseRepository) GetReport(ctx context.Context, startDate, endDate time.Time) (*model.ExpenseReportResponse, error) {
	report := &model.ExpenseReportResponse{}

	// 合計金額
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&report.TotalAmount)

	// カテゴリ別内訳
	r.db.WithContext(ctx).Model(&model.ExpenseItem{}).
		Joins("JOIN expenses ON expenses.id = expense_items.expense_id").
		Where("expenses.created_at BETWEEN ? AND ?", startDate, endDate).
		Select("expense_items.category, COALESCE(SUM(expense_items.amount), 0) as amount").
		Group("expense_items.category").
		Scan(&report.CategoryBreakdown)

	// ステータス別集計
	type statusCount struct {
		Status string
		Count  int
	}
	var counts []statusCount
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("status, COUNT(*) as count").Group("status").Scan(&counts)
	for _, c := range counts {
		switch model.ExpenseStatus(c.Status) {
		case model.ExpenseStatusDraft:
			report.StatusSummary.Draft = c.Count
		case model.ExpenseStatusPending:
			report.StatusSummary.Pending = c.Count
		case model.ExpenseStatusApproved:
			report.StatusSummary.Approved = c.Count
		case model.ExpenseStatusRejected:
			report.StatusSummary.Rejected = c.Count
		case model.ExpenseStatusReimbursed:
			report.StatusSummary.Reimbursed = c.Count
		}
	}

	// 部署別内訳
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Joins("JOIN users ON users.id = expenses.user_id").
		Joins("LEFT JOIN departments ON departments.id = users.department_id").
		Where("expenses.created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(departments.name, 'その他') as department, COALESCE(SUM(expenses.total_amount), 0) as amount, COUNT(*) as count, COALESCE(AVG(expenses.total_amount), 0) as avg").
		Group("departments.name").
		Scan(&report.DepartmentBreakdown)

	return report, nil
}

func (r *expenseRepository) GetMonthlyTrend(ctx context.Context, year int) ([]model.MonthlyTrendItem, error) {
	var items []model.MonthlyTrendItem
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, COALESCE(SUM(total_amount), 0) as amount").
		Group("TO_CHAR(created_at, 'YYYY-MM')").
		Order("month").
		Scan(&items)
	return items, nil
}

// ===== ExpenseItemRepository =====

type ExpenseItemRepository interface {
	DeleteByExpenseID(ctx context.Context, expenseID uuid.UUID) error
	CreateBatch(ctx context.Context, items []model.ExpenseItem) error
}

type expenseItemRepository struct {
	db *gorm.DB
}

func NewExpenseItemRepository(db *gorm.DB) ExpenseItemRepository {
	return &expenseItemRepository{db: db}
}

func (r *expenseItemRepository) DeleteByExpenseID(ctx context.Context, expenseID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("expense_id = ?", expenseID).Delete(&model.ExpenseItem{}).Error
}

func (r *expenseItemRepository) CreateBatch(ctx context.Context, items []model.ExpenseItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&items).Error
}

// ===== ExpenseCommentRepository =====

type ExpenseCommentRepository interface {
	Create(ctx context.Context, comment *model.ExpenseComment) error
	FindByExpenseID(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseComment, error)
}

type expenseCommentRepository struct {
	db *gorm.DB
}

func NewExpenseCommentRepository(db *gorm.DB) ExpenseCommentRepository {
	return &expenseCommentRepository{db: db}
}

func (r *expenseCommentRepository) Create(ctx context.Context, comment *model.ExpenseComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *expenseCommentRepository) FindByExpenseID(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseComment, error) {
	var comments []model.ExpenseComment
	err := r.db.WithContext(ctx).Preload("User").Where("expense_id = ?", expenseID).Order("created_at ASC").Find(&comments).Error
	return comments, err
}

// ===== ExpenseHistoryRepository =====

type ExpenseHistoryRepository interface {
	Create(ctx context.Context, history *model.ExpenseHistory) error
	FindByExpenseID(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistory, error)
}

type expenseHistoryRepository struct {
	db *gorm.DB
}

func NewExpenseHistoryRepository(db *gorm.DB) ExpenseHistoryRepository {
	return &expenseHistoryRepository{db: db}
}

func (r *expenseHistoryRepository) Create(ctx context.Context, history *model.ExpenseHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *expenseHistoryRepository) FindByExpenseID(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistory, error) {
	var histories []model.ExpenseHistory
	err := r.db.WithContext(ctx).Where("expense_id = ?", expenseID).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// ===== ExpenseTemplateRepository =====

type ExpenseTemplateRepository interface {
	Create(ctx context.Context, template *model.ExpenseTemplate) error
	FindAll(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExpenseTemplate, error)
	Update(ctx context.Context, template *model.ExpenseTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type expenseTemplateRepository struct {
	db *gorm.DB
}

func NewExpenseTemplateRepository(db *gorm.DB) ExpenseTemplateRepository {
	return &expenseTemplateRepository{db: db}
}

func (r *expenseTemplateRepository) Create(ctx context.Context, template *model.ExpenseTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *expenseTemplateRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error) {
	var templates []model.ExpenseTemplate
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("name").Find(&templates).Error
	return templates, err
}

func (r *expenseTemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExpenseTemplate, error) {
	var tmpl model.ExpenseTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tmpl).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (r *expenseTemplateRepository) Update(ctx context.Context, template *model.ExpenseTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *expenseTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ExpenseTemplate{}, "id = ?", id).Error
}

// ===== ExpensePolicyRepository =====

type ExpensePolicyRepository interface {
	Create(ctx context.Context, policy *model.ExpensePolicy) error
	FindAll(ctx context.Context) ([]model.ExpensePolicy, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExpensePolicy, error)
	FindByCategory(ctx context.Context, category string) (*model.ExpensePolicy, error)
	Update(ctx context.Context, policy *model.ExpensePolicy) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type expensePolicyRepository struct {
	db *gorm.DB
}

func NewExpensePolicyRepository(db *gorm.DB) ExpensePolicyRepository {
	return &expensePolicyRepository{db: db}
}

func (r *expensePolicyRepository) Create(ctx context.Context, policy *model.ExpensePolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *expensePolicyRepository) FindAll(ctx context.Context) ([]model.ExpensePolicy, error) {
	var policies []model.ExpensePolicy
	err := r.db.WithContext(ctx).Order("category").Find(&policies).Error
	return policies, err
}

func (r *expensePolicyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExpensePolicy, error) {
	var policy model.ExpensePolicy
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *expensePolicyRepository) FindByCategory(ctx context.Context, category string) (*model.ExpensePolicy, error) {
	var policy model.ExpensePolicy
	err := r.db.WithContext(ctx).Where("category = ? AND is_active = true", category).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *expensePolicyRepository) Update(ctx context.Context, policy *model.ExpensePolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *expensePolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ExpensePolicy{}, "id = ?", id).Error
}

// ===== ExpenseBudgetRepository =====

type ExpenseBudgetRepository interface {
	FindAll(ctx context.Context) ([]model.ExpenseBudget, error)
}

type expenseBudgetRepository struct {
	db *gorm.DB
}

func NewExpenseBudgetRepository(db *gorm.DB) ExpenseBudgetRepository {
	return &expenseBudgetRepository{db: db}
}

func (r *expenseBudgetRepository) FindAll(ctx context.Context) ([]model.ExpenseBudget, error) {
	var budgets []model.ExpenseBudget
	err := r.db.WithContext(ctx).Order("fiscal_year DESC").Find(&budgets).Error
	return budgets, err
}

// ===== ExpenseNotificationRepository =====

type ExpenseNotificationRepository interface {
	Create(ctx context.Context, notification *model.ExpenseNotification) error
	FindByUserID(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

type expenseNotificationRepository struct {
	db *gorm.DB
}

func NewExpenseNotificationRepository(db *gorm.DB) ExpenseNotificationRepository {
	return &expenseNotificationRepository{db: db}
}

func (r *expenseNotificationRepository) Create(ctx context.Context, notification *model.ExpenseNotification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *expenseNotificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error) {
	var notifications []model.ExpenseNotification
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if filter == "unread" {
		q = q.Where("is_read = false")
	}
	err := q.Order("created_at DESC").Limit(100).Find(&notifications).Error
	return notifications, err
}

func (r *expenseNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.ExpenseNotification{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *expenseNotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.ExpenseNotification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

// ===== ExpenseReminderRepository =====

type ExpenseReminderRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error)
	Dismiss(ctx context.Context, id uuid.UUID) error
}

type expenseReminderRepository struct {
	db *gorm.DB
}

func NewExpenseReminderRepository(db *gorm.DB) ExpenseReminderRepository {
	return &expenseReminderRepository{db: db}
}

func (r *expenseReminderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error) {
	var reminders []model.ExpenseReminder
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_dismissed = false", userID).Order("due_date ASC").Find(&reminders).Error
	return reminders, err
}

func (r *expenseReminderRepository) Dismiss(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.ExpenseReminder{}).Where("id = ?", id).Update("is_dismissed", true).Error
}

// ===== ExpenseNotificationSettingRepository =====

type ExpenseNotificationSettingRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error)
	Upsert(ctx context.Context, setting *model.ExpenseNotificationSetting) error
}

type expenseNotificationSettingRepository struct {
	db *gorm.DB
}

func NewExpenseNotificationSettingRepository(db *gorm.DB) ExpenseNotificationSettingRepository {
	return &expenseNotificationSettingRepository{db: db}
}

func (r *expenseNotificationSettingRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error) {
	var setting model.ExpenseNotificationSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		// デフォルト設定を返す
		return &model.ExpenseNotificationSetting{
			UserID:              userID,
			EmailEnabled:        true,
			PushEnabled:         true,
			ApprovalAlerts:      true,
			ReimbursementAlerts: true,
			PolicyAlerts:        true,
			WeeklyDigest:        false,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *expenseNotificationSettingRepository) Upsert(ctx context.Context, setting *model.ExpenseNotificationSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

// ===== ExpenseApprovalFlowRepository =====

type ExpenseApprovalFlowRepository interface {
	FindActive(ctx context.Context) (*model.ExpenseApprovalFlow, error)
}

type expenseApprovalFlowRepository struct {
	db *gorm.DB
}

func NewExpenseApprovalFlowRepository(db *gorm.DB) ExpenseApprovalFlowRepository {
	return &expenseApprovalFlowRepository{db: db}
}

func (r *expenseApprovalFlowRepository) FindActive(ctx context.Context) (*model.ExpenseApprovalFlow, error) {
	var flow model.ExpenseApprovalFlow
	err := r.db.WithContext(ctx).Where("is_active = true").First(&flow).Error
	if err == gorm.ErrRecordNotFound {
		return &model.ExpenseApprovalFlow{
			Name:             "デフォルト承認フロー",
			MinAmount:        0,
			MaxAmount:        0,
			RequiredSteps:    1,
			IsActive:         true,
			AutoApproveBelow: 1000,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// ===== ExpenseDelegateRepository =====

type ExpenseDelegateRepository interface {
	Create(ctx context.Context, delegate *model.ExpenseDelegate) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type expenseDelegateRepository struct {
	db *gorm.DB
}

func NewExpenseDelegateRepository(db *gorm.DB) ExpenseDelegateRepository {
	return &expenseDelegateRepository{db: db}
}

func (r *expenseDelegateRepository) Create(ctx context.Context, delegate *model.ExpenseDelegate) error {
	return r.db.WithContext(ctx).Create(delegate).Error
}

func (r *expenseDelegateRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error) {
	var delegates []model.ExpenseDelegate
	err := r.db.WithContext(ctx).Preload("Delegate").Where("user_id = ? AND is_active = true", userID).Find(&delegates).Error
	return delegates, err
}

func (r *expenseDelegateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ExpenseDelegate{}, "id = ?", id).Error
}

// ===== ExpensePolicyViolationRepository =====

type ExpensePolicyViolationRepository interface {
	Create(ctx context.Context, violation *model.ExpensePolicyViolation) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error)
}

type expensePolicyViolationRepository struct {
	db *gorm.DB
}

func NewExpensePolicyViolationRepository(db *gorm.DB) ExpensePolicyViolationRepository {
	return &expensePolicyViolationRepository{db: db}
}

func (r *expensePolicyViolationRepository) Create(ctx context.Context, violation *model.ExpensePolicyViolation) error {
	return r.db.WithContext(ctx).Create(violation).Error
}

func (r *expensePolicyViolationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error) {
	var violations []model.ExpensePolicyViolation
	err := r.db.WithContext(ctx).Preload("Expense").Preload("Policy").
		Where("user_id = ?", userID).Order("created_at DESC").Find(&violations).Error
	return violations, err
}
