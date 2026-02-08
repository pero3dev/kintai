package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/model"
)

// Setup はルーターを設定する
func Setup(r *gin.Engine, h *handler.Handlers, mw *middleware.Middleware) {
	// グローバルミドルウェア
	r.Use(mw.Recovery())
	r.Use(mw.RequestLogger())
	r.Use(mw.CORS())
	r.Use(mw.SecurityHeaders())
	r.Use(mw.RateLimit())

	// ヘルスチェック・メトリクス（認証不要）
	r.GET("/health", h.Health.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 認証（認証不要）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", h.Auth.Login)
			auth.POST("/register", h.Auth.Register)
			auth.POST("/refresh", h.Auth.RefreshToken)
		}

		// 認証が必要なルート
		protected := v1.Group("")
		protected.Use(mw.Auth())
		{
			// ログアウト
			protected.POST("/auth/logout", h.Auth.Logout)

			// 勤怠
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/clock-in", h.Attendance.ClockIn)
				attendance.POST("/clock-out", h.Attendance.ClockOut)
				attendance.GET("", h.Attendance.GetMyAttendances)
				attendance.GET("/today", h.Attendance.GetTodayStatus)
				attendance.GET("/summary", h.Attendance.GetSummary)
			}

			// 休暇申請
			leaves := protected.Group("/leaves")
			{
				leaves.POST("", h.Leave.Create)
				leaves.GET("", h.Leave.GetMy)
			}

			// 残業申請
			overtime := protected.Group("/overtime")
			{
				overtime.POST("", h.OvertimeRequest.Create)
				overtime.GET("", h.OvertimeRequest.GetMy)
			}

			// 勤怠修正申請
			corrections := protected.Group("/corrections")
			{
				corrections.POST("", h.AttendanceCorrection.Create)
				corrections.GET("", h.AttendanceCorrection.GetMy)
			}

			// 通知
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", h.Notification.GetMy)
				notifications.GET("/unread-count", h.Notification.GetUnreadCount)
				notifications.PUT("/:id/read", h.Notification.MarkAsRead)
				notifications.PUT("/read-all", h.Notification.MarkAllAsRead)
				notifications.DELETE("/:id", h.Notification.Delete)
			}

			// 有給残日数
			leaveBalance := protected.Group("/leave-balances")
			{
				leaveBalance.GET("", h.LeaveBalance.GetMy)
			}

			// 工数管理
			timeEntries := protected.Group("/time-entries")
			{
				timeEntries.POST("", h.TimeEntry.Create)
				timeEntries.GET("", h.TimeEntry.GetMy)
				timeEntries.PUT("/:id", h.TimeEntry.Update)
				timeEntries.DELETE("/:id", h.TimeEntry.Delete)
			}

			// ユーザー
			users := protected.Group("/users")
			{
				users.GET("/me", h.User.GetMe)
			}

			// 部署
			departments := protected.Group("/departments")
			{
				departments.GET("", h.Department.GetAll)
			}

			// シフト
			shifts := protected.Group("/shifts")
			{
				shifts.GET("", h.Shift.GetByDateRange)
			}

			// プロジェクト（閲覧は全員可）
			projects := protected.Group("/projects")
			{
				projects.GET("", h.Project.GetAll)
				projects.GET("/:id", h.Project.GetByID)
				projects.GET("/:id/time-entries", h.TimeEntry.GetByProject)
			}

			// 祝日・カレンダー（閲覧は全員可）
			holidays := protected.Group("/holidays")
			{
				holidays.GET("", h.Holiday.GetByYear)
				holidays.GET("/calendar", h.Holiday.GetCalendar)
				holidays.GET("/working-days", h.Holiday.GetWorkingDays)
			}

			// 管理者・マネージャー向け
			admin := protected.Group("")
			admin.Use(mw.RequireRole(model.RoleAdmin, model.RoleManager))
			{
				// 休暇承認
				admin.GET("/leaves/pending", h.Leave.GetPending)
				admin.PUT("/leaves/:id/approve", h.Leave.Approve)

				// 残業申請承認
				admin.GET("/overtime/pending", h.OvertimeRequest.GetPending)
				admin.PUT("/overtime/:id/approve", h.OvertimeRequest.Approve)
				admin.GET("/overtime/alerts", h.OvertimeRequest.GetAlerts)

				// 勤怠修正承認
				admin.GET("/corrections/pending", h.AttendanceCorrection.GetPending)
				admin.PUT("/corrections/:id/approve", h.AttendanceCorrection.Approve)

				// 有給残日数管理
				admin.GET("/leave-balances/:user_id", h.LeaveBalance.GetByUser)
				admin.PUT("/leave-balances/:user_id/:leave_type", h.LeaveBalance.SetBalance)
				admin.POST("/leave-balances/:user_id/initialize", h.LeaveBalance.Initialize)

				// ユーザー管理
				admin.GET("/users", h.User.GetAll)
				admin.POST("/users", h.User.Create)
				admin.PUT("/users/:id", h.User.Update)
				admin.DELETE("/users/:id", h.User.Delete)

				// シフト管理
				admin.POST("/shifts", h.Shift.Create)
				admin.POST("/shifts/bulk", h.Shift.BulkCreate)
				admin.DELETE("/shifts/:id", h.Shift.Delete)

				// プロジェクト管理
				admin.POST("/projects", h.Project.Create)
				admin.PUT("/projects/:id", h.Project.Update)
				admin.DELETE("/projects/:id", h.Project.Delete)

				// 工数サマリー
				admin.GET("/time-entries/summary", h.TimeEntry.GetSummary)

				// 祝日管理
				admin.POST("/holidays", h.Holiday.Create)
				admin.PUT("/holidays/:id", h.Holiday.Update)
				admin.DELETE("/holidays/:id", h.Holiday.Delete)

				// 承認フロー管理
				admin.GET("/approval-flows", h.ApprovalFlow.GetAll)
				admin.GET("/approval-flows/:id", h.ApprovalFlow.GetByID)
				admin.POST("/approval-flows", h.ApprovalFlow.Create)
				admin.PUT("/approval-flows/:id", h.ApprovalFlow.Update)
				admin.DELETE("/approval-flows/:id", h.ApprovalFlow.Delete)

				// エクスポート
				admin.GET("/export/attendance", h.Export.ExportAttendance)
				admin.GET("/export/leaves", h.Export.ExportLeaves)
				admin.GET("/export/overtime", h.Export.ExportOvertime)
				admin.GET("/export/projects", h.Export.ExportProjects)

				// ダッシュボード
				admin.GET("/dashboard/stats", h.Dashboard.GetStats)
			}
		}
	}
}
