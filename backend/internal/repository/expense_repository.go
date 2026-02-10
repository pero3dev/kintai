package repository

import (
	appexpense "github.com/your-org/kintai/backend/internal/apps/expense"
	"gorm.io/gorm"
)

type ExpenseRepository = appexpense.ExpenseRepository
type ExpenseItemRepository = appexpense.ExpenseItemRepository
type ExpenseCommentRepository = appexpense.ExpenseCommentRepository
type ExpenseHistoryRepository = appexpense.ExpenseHistoryRepository
type ExpenseTemplateRepository = appexpense.ExpenseTemplateRepository
type ExpensePolicyRepository = appexpense.ExpensePolicyRepository
type ExpenseBudgetRepository = appexpense.ExpenseBudgetRepository
type ExpenseNotificationRepository = appexpense.ExpenseNotificationRepository
type ExpenseReminderRepository = appexpense.ExpenseReminderRepository
type ExpenseNotificationSettingRepository = appexpense.ExpenseNotificationSettingRepository
type ExpenseApprovalFlowRepository = appexpense.ExpenseApprovalFlowRepository
type ExpenseDelegateRepository = appexpense.ExpenseDelegateRepository
type ExpensePolicyViolationRepository = appexpense.ExpensePolicyViolationRepository

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return appexpense.NewExpenseRepository(db)
}

func NewExpenseItemRepository(db *gorm.DB) ExpenseItemRepository {
	return appexpense.NewExpenseItemRepository(db)
}

func NewExpenseCommentRepository(db *gorm.DB) ExpenseCommentRepository {
	return appexpense.NewExpenseCommentRepository(db)
}

func NewExpenseHistoryRepository(db *gorm.DB) ExpenseHistoryRepository {
	return appexpense.NewExpenseHistoryRepository(db)
}

func NewExpenseTemplateRepository(db *gorm.DB) ExpenseTemplateRepository {
	return appexpense.NewExpenseTemplateRepository(db)
}

func NewExpensePolicyRepository(db *gorm.DB) ExpensePolicyRepository {
	return appexpense.NewExpensePolicyRepository(db)
}

func NewExpenseBudgetRepository(db *gorm.DB) ExpenseBudgetRepository {
	return appexpense.NewExpenseBudgetRepository(db)
}

func NewExpenseNotificationRepository(db *gorm.DB) ExpenseNotificationRepository {
	return appexpense.NewExpenseNotificationRepository(db)
}

func NewExpenseReminderRepository(db *gorm.DB) ExpenseReminderRepository {
	return appexpense.NewExpenseReminderRepository(db)
}

func NewExpenseNotificationSettingRepository(db *gorm.DB) ExpenseNotificationSettingRepository {
	return appexpense.NewExpenseNotificationSettingRepository(db)
}

func NewExpenseApprovalFlowRepository(db *gorm.DB) ExpenseApprovalFlowRepository {
	return appexpense.NewExpenseApprovalFlowRepository(db)
}

func NewExpenseDelegateRepository(db *gorm.DB) ExpenseDelegateRepository {
	return appexpense.NewExpenseDelegateRepository(db)
}

func NewExpensePolicyViolationRepository(db *gorm.DB) ExpensePolicyViolationRepository {
	return appexpense.NewExpensePolicyViolationRepository(db)
}
