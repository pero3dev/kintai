package service

import appexpense "github.com/your-org/kintai/backend/internal/apps/expense"

type ExpenseService = appexpense.ExpenseService
type ExpenseCommentService = appexpense.ExpenseCommentService
type ExpenseHistoryService = appexpense.ExpenseHistoryService
type ExpenseReceiptService = appexpense.ExpenseReceiptService
type ExpenseTemplateService = appexpense.ExpenseTemplateService
type ExpensePolicyService = appexpense.ExpensePolicyService
type ExpenseNotificationService = appexpense.ExpenseNotificationService
type ExpenseApprovalFlowService = appexpense.ExpenseApprovalFlowService

func toExpenseDeps(deps Deps) appexpense.Deps {
	return appexpense.Deps{
		Repos: &appexpense.Repositories{
			User:                       deps.Repos.User,
			Expense:                    deps.Repos.Expense,
			ExpenseItem:                deps.Repos.ExpenseItem,
			ExpenseComment:             deps.Repos.ExpenseComment,
			ExpenseHistory:             deps.Repos.ExpenseHistory,
			ExpenseTemplate:            deps.Repos.ExpenseTemplate,
			ExpensePolicy:              deps.Repos.ExpensePolicy,
			ExpenseBudget:              deps.Repos.ExpenseBudget,
			ExpenseNotification:        deps.Repos.ExpenseNotification,
			ExpenseReminder:            deps.Repos.ExpenseReminder,
			ExpenseNotificationSetting: deps.Repos.ExpenseNotificationSetting,
			ExpenseApprovalFlow:        deps.Repos.ExpenseApprovalFlow,
			ExpenseDelegate:            deps.Repos.ExpenseDelegate,
			ExpensePolicyViolation:     deps.Repos.ExpensePolicyViolation,
		},
		Config: deps.Config,
		Logger: deps.Logger,
	}
}

func NewExpenseService(deps Deps) ExpenseService {
	return appexpense.NewExpenseService(toExpenseDeps(deps))
}

func NewExpenseCommentService(deps Deps) ExpenseCommentService {
	return appexpense.NewExpenseCommentService(toExpenseDeps(deps))
}

func NewExpenseHistoryService(deps Deps) ExpenseHistoryService {
	return appexpense.NewExpenseHistoryService(toExpenseDeps(deps))
}

func NewExpenseReceiptService(deps Deps) ExpenseReceiptService {
	return appexpense.NewExpenseReceiptService(toExpenseDeps(deps))
}

func NewExpenseTemplateService(deps Deps) ExpenseTemplateService {
	return appexpense.NewExpenseTemplateService(toExpenseDeps(deps))
}

func NewExpensePolicyService(deps Deps) ExpensePolicyService {
	return appexpense.NewExpensePolicyService(toExpenseDeps(deps))
}

func NewExpenseNotificationService(deps Deps) ExpenseNotificationService {
	return appexpense.NewExpenseNotificationService(toExpenseDeps(deps))
}

func NewExpenseApprovalFlowService(deps Deps) ExpenseApprovalFlowService {
	return appexpense.NewExpenseApprovalFlowService(toExpenseDeps(deps))
}
