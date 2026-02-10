package shared

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/model"
)

func RegisterPublicRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/register", h.Auth.Register)
		auth.POST("/refresh", h.Auth.RefreshToken)
	}
}

func RegisterProtectedRoutes(protected *gin.RouterGroup, h *handler.Handlers, mw *middleware.Middleware) {
	protected.POST("/auth/logout", h.Auth.Logout)

	notifications := protected.Group("/notifications")
	{
		notifications.GET("", h.Notification.GetMy)
		notifications.GET("/unread-count", h.Notification.GetUnreadCount)
		notifications.PUT("/:id/read", h.Notification.MarkAsRead)
		notifications.PUT("/read-all", h.Notification.MarkAllAsRead)
		notifications.DELETE("/:id", h.Notification.Delete)
	}

	timeEntries := protected.Group("/time-entries")
	{
		timeEntries.POST("", h.TimeEntry.Create)
		timeEntries.GET("", h.TimeEntry.GetMy)
		timeEntries.PUT("/:id", h.TimeEntry.Update)
		timeEntries.DELETE("/:id", h.TimeEntry.Delete)
	}

	users := protected.Group("/users")
	{
		users.GET("/me", h.User.GetMe)
	}

	departments := protected.Group("/departments")
	{
		departments.GET("", h.Department.GetAll)
	}

	shifts := protected.Group("/shifts")
	{
		shifts.GET("", h.Shift.GetByDateRange)
	}

	projects := protected.Group("/projects")
	{
		projects.GET("", h.Project.GetAll)
		projects.GET("/:id", h.Project.GetByID)
		projects.GET("/:id/time-entries", h.TimeEntry.GetByProject)
	}

	holidays := protected.Group("/holidays")
	{
		holidays.GET("", h.Holiday.GetByYear)
		holidays.GET("/calendar", h.Holiday.GetCalendar)
		holidays.GET("/working-days", h.Holiday.GetWorkingDays)
	}

	admin := protected.Group("")
	admin.Use(mw.RequireRole(model.RoleAdmin, model.RoleManager))
	{
		admin.GET("/users", h.User.GetAll)
		admin.POST("/users", h.User.Create)
		admin.PUT("/users/:id", h.User.Update)
		admin.DELETE("/users/:id", h.User.Delete)

		admin.POST("/shifts", h.Shift.Create)
		admin.POST("/shifts/bulk", h.Shift.BulkCreate)
		admin.DELETE("/shifts/:id", h.Shift.Delete)

		admin.POST("/projects", h.Project.Create)
		admin.PUT("/projects/:id", h.Project.Update)
		admin.DELETE("/projects/:id", h.Project.Delete)

		admin.GET("/time-entries/summary", h.TimeEntry.GetSummary)

		admin.POST("/holidays", h.Holiday.Create)
		admin.PUT("/holidays/:id", h.Holiday.Update)
		admin.DELETE("/holidays/:id", h.Holiday.Delete)

		admin.GET("/approval-flows", h.ApprovalFlow.GetAll)
		admin.GET("/approval-flows/:id", h.ApprovalFlow.GetByID)
		admin.POST("/approval-flows", h.ApprovalFlow.Create)
		admin.PUT("/approval-flows/:id", h.ApprovalFlow.Update)
		admin.DELETE("/approval-flows/:id", h.ApprovalFlow.Delete)

		admin.GET("/export/attendance", h.Export.ExportAttendance)
		admin.GET("/export/leaves", h.Export.ExportLeaves)
		admin.GET("/export/overtime", h.Export.ExportOvertime)
		admin.GET("/export/projects", h.Export.ExportProjects)

		admin.GET("/dashboard/stats", h.Dashboard.GetStats)
	}
}
