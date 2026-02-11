package expense

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// UserRepository defines the user lookups required by expense services.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// Repositories groups repository dependencies required by the expense app.
type Repositories struct {
	User                       UserRepository
	Expense                    ExpenseRepository
	ExpenseItem                ExpenseItemRepository
	ExpenseComment             ExpenseCommentRepository
	ExpenseHistory             ExpenseHistoryRepository
	ExpenseTemplate            ExpenseTemplateRepository
	ExpensePolicy              ExpensePolicyRepository
	ExpenseBudget              ExpenseBudgetRepository
	ExpenseNotification        ExpenseNotificationRepository
	ExpenseReminder            ExpenseReminderRepository
	ExpenseNotificationSetting ExpenseNotificationSettingRepository
	ExpenseApprovalFlow        ExpenseApprovalFlowRepository
	ExpenseDelegate            ExpenseDelegateRepository
	ExpensePolicyViolation     ExpensePolicyViolationRepository
}

// Deps defines dependencies for expense app services.
type Deps struct {
	Repos  *Repositories
	Config *config.Config
	Logger *logger.Logger
}
