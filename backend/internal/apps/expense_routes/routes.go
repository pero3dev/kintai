package expense_routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/handler"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, h *handler.Handlers) {
	expenses := protected.Group("/expenses")
	{
		expenses.GET("", h.Expense.GetList)
		expenses.POST("", h.Expense.Create)
		expenses.GET("/stats", h.Expense.GetStats)
		expenses.GET("/pending", h.Expense.GetPending)
		expenses.GET("/report", h.Expense.GetReport)
		expenses.GET("/report/monthly", h.Expense.GetMonthlyTrend)
		expenses.GET("/export/csv", h.Expense.ExportCSV)
		expenses.GET("/export/pdf", h.Expense.ExportPDF)

		expenses.POST("/receipts/upload", h.ExpenseReceipt.Upload)

		expenses.GET("/templates", h.ExpenseTemplate.GetTemplates)
		expenses.POST("/templates", h.ExpenseTemplate.Create)
		expenses.PUT("/templates/:id", h.ExpenseTemplate.Update)
		expenses.DELETE("/templates/:id", h.ExpenseTemplate.Delete)
		expenses.POST("/templates/:id/use", h.ExpenseTemplate.UseTemplate)

		expenses.GET("/policies", h.ExpensePolicy.GetPolicies)
		expenses.POST("/policies", h.ExpensePolicy.Create)
		expenses.PUT("/policies/:id", h.ExpensePolicy.Update)
		expenses.DELETE("/policies/:id", h.ExpensePolicy.Delete)
		expenses.GET("/budgets", h.ExpensePolicy.GetBudgets)
		expenses.GET("/policy-violations", h.ExpensePolicy.GetPolicyViolations)

		expenses.GET("/notifications", h.ExpenseNotification.GetNotifications)
		expenses.PUT("/notifications/read-all", h.ExpenseNotification.MarkAllAsRead)
		expenses.PUT("/notifications/:id/read", h.ExpenseNotification.MarkAsRead)
		expenses.GET("/reminders", h.ExpenseNotification.GetReminders)
		expenses.PUT("/reminders/:id/dismiss", h.ExpenseNotification.DismissReminder)
		expenses.GET("/notification-settings", h.ExpenseNotification.GetSettings)
		expenses.PUT("/notification-settings", h.ExpenseNotification.UpdateSettings)

		expenses.GET("/approval-flow", h.ExpenseApprovalFlow.GetConfig)
		expenses.GET("/delegates", h.ExpenseApprovalFlow.GetDelegates)
		expenses.POST("/delegates", h.ExpenseApprovalFlow.SetDelegate)
		expenses.DELETE("/delegates/:id", h.ExpenseApprovalFlow.RemoveDelegate)

		expenses.GET("/:id", h.Expense.GetByID)
		expenses.PUT("/:id", h.Expense.Update)
		expenses.DELETE("/:id", h.Expense.Delete)
		expenses.PUT("/:id/approve", h.Expense.Approve)
		expenses.PUT("/:id/advanced-approve", h.Expense.AdvancedApprove)
		expenses.GET("/:id/comments", h.ExpenseComment.GetComments)
		expenses.POST("/:id/comments", h.ExpenseComment.AddComment)
		expenses.GET("/:id/history", h.ExpenseHistory.GetHistory)
	}
}
