package handler

import (
	appexpense "github.com/your-org/kintai/backend/internal/apps/expense"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

type ExpenseHandler struct {
	svc    service.ExpenseService
	logger *logger.Logger
	*appexpense.ExpenseHandler
}

type ExpenseCommentHandler struct {
	svc    service.ExpenseCommentService
	logger *logger.Logger
	*appexpense.ExpenseCommentHandler
}

type ExpenseHistoryHandler struct {
	svc    service.ExpenseHistoryService
	logger *logger.Logger
	*appexpense.ExpenseHistoryHandler
}

type ExpenseReceiptHandler struct {
	svc    service.ExpenseReceiptService
	logger *logger.Logger
	*appexpense.ExpenseReceiptHandler
}

type ExpenseTemplateHandler struct {
	svc    service.ExpenseTemplateService
	logger *logger.Logger
	*appexpense.ExpenseTemplateHandler
}

type ExpensePolicyHandler struct {
	svc    service.ExpensePolicyService
	logger *logger.Logger
	*appexpense.ExpensePolicyHandler
}

type ExpenseNotificationHandler struct {
	svc    service.ExpenseNotificationService
	logger *logger.Logger
	*appexpense.ExpenseNotificationHandler
}

type ExpenseApprovalFlowHandler struct {
	svc    service.ExpenseApprovalFlowService
	logger *logger.Logger
	*appexpense.ExpenseApprovalFlowHandler
}

func NewExpenseHandler(svc service.ExpenseService, logger *logger.Logger) *ExpenseHandler {
	return &ExpenseHandler{svc: svc, logger: logger, ExpenseHandler: appexpense.NewExpenseHandler(svc, logger)}
}

func NewExpenseCommentHandler(svc service.ExpenseCommentService, logger *logger.Logger) *ExpenseCommentHandler {
	return &ExpenseCommentHandler{svc: svc, logger: logger, ExpenseCommentHandler: appexpense.NewExpenseCommentHandler(svc, logger)}
}

func NewExpenseHistoryHandler(svc service.ExpenseHistoryService, logger *logger.Logger) *ExpenseHistoryHandler {
	return &ExpenseHistoryHandler{svc: svc, logger: logger, ExpenseHistoryHandler: appexpense.NewExpenseHistoryHandler(svc, logger)}
}

func NewExpenseReceiptHandler(svc service.ExpenseReceiptService, logger *logger.Logger) *ExpenseReceiptHandler {
	return &ExpenseReceiptHandler{svc: svc, logger: logger, ExpenseReceiptHandler: appexpense.NewExpenseReceiptHandler(svc, logger)}
}

func NewExpenseTemplateHandler(svc service.ExpenseTemplateService, logger *logger.Logger) *ExpenseTemplateHandler {
	return &ExpenseTemplateHandler{svc: svc, logger: logger, ExpenseTemplateHandler: appexpense.NewExpenseTemplateHandler(svc, logger)}
}

func NewExpensePolicyHandler(svc service.ExpensePolicyService, logger *logger.Logger) *ExpensePolicyHandler {
	return &ExpensePolicyHandler{svc: svc, logger: logger, ExpensePolicyHandler: appexpense.NewExpensePolicyHandler(svc, logger)}
}

func NewExpenseNotificationHandler(svc service.ExpenseNotificationService, logger *logger.Logger) *ExpenseNotificationHandler {
	return &ExpenseNotificationHandler{svc: svc, logger: logger, ExpenseNotificationHandler: appexpense.NewExpenseNotificationHandler(svc, logger)}
}

func NewExpenseApprovalFlowHandler(svc service.ExpenseApprovalFlowService, logger *logger.Logger) *ExpenseApprovalFlowHandler {
	return &ExpenseApprovalFlowHandler{svc: svc, logger: logger, ExpenseApprovalFlowHandler: appexpense.NewExpenseApprovalFlowHandler(svc, logger)}
}
