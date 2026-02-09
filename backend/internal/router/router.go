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

			// 経費精算 (認証済みユーザー全員)
			expenses := protected.Group("/expenses")
			{
				// CRUD
				expenses.GET("", h.Expense.GetList)
				expenses.POST("", h.Expense.Create)
				expenses.GET("/stats", h.Expense.GetStats)
				expenses.GET("/pending", h.Expense.GetPending)
				expenses.GET("/report", h.Expense.GetReport)
				expenses.GET("/report/monthly", h.Expense.GetMonthlyTrend)
				expenses.GET("/export/csv", h.Expense.ExportCSV)
				expenses.GET("/export/pdf", h.Expense.ExportPDF)

				// レシート
				expenses.POST("/receipts/upload", h.ExpenseReceipt.Upload)

				// テンプレート
				expenses.GET("/templates", h.ExpenseTemplate.GetTemplates)
				expenses.POST("/templates", h.ExpenseTemplate.Create)
				expenses.PUT("/templates/:id", h.ExpenseTemplate.Update)
				expenses.DELETE("/templates/:id", h.ExpenseTemplate.Delete)
				expenses.POST("/templates/:id/use", h.ExpenseTemplate.UseTemplate)

				// ポリシー
				expenses.GET("/policies", h.ExpensePolicy.GetPolicies)
				expenses.POST("/policies", h.ExpensePolicy.Create)
				expenses.PUT("/policies/:id", h.ExpensePolicy.Update)
				expenses.DELETE("/policies/:id", h.ExpensePolicy.Delete)
				expenses.GET("/budgets", h.ExpensePolicy.GetBudgets)
				expenses.GET("/policy-violations", h.ExpensePolicy.GetPolicyViolations)

				// 通知
				expenses.GET("/notifications", h.ExpenseNotification.GetNotifications)
				expenses.PUT("/notifications/read-all", h.ExpenseNotification.MarkAllAsRead)
				expenses.PUT("/notifications/:id/read", h.ExpenseNotification.MarkAsRead)
				expenses.GET("/reminders", h.ExpenseNotification.GetReminders)
				expenses.PUT("/reminders/:id/dismiss", h.ExpenseNotification.DismissReminder)
				expenses.GET("/notification-settings", h.ExpenseNotification.GetSettings)
				expenses.PUT("/notification-settings", h.ExpenseNotification.UpdateSettings)

				// 承認ワークフロー・代理
				expenses.GET("/approval-flow", h.ExpenseApprovalFlow.GetConfig)
				expenses.GET("/delegates", h.ExpenseApprovalFlow.GetDelegates)
				expenses.POST("/delegates", h.ExpenseApprovalFlow.SetDelegate)
				expenses.DELETE("/delegates/:id", h.ExpenseApprovalFlow.RemoveDelegate)

				// 個別経費操作（:idルート）
				expenses.GET("/:id", h.Expense.GetByID)
				expenses.PUT("/:id", h.Expense.Update)
				expenses.DELETE("/:id", h.Expense.Delete)
				expenses.PUT("/:id/approve", h.Expense.Approve)
				expenses.PUT("/:id/advanced-approve", h.Expense.AdvancedApprove)
				expenses.GET("/:id/comments", h.ExpenseComment.GetComments)
				expenses.POST("/:id/comments", h.ExpenseComment.AddComment)
				expenses.GET("/:id/history", h.ExpenseHistory.GetHistory)
			}

			// HR (認証済みユーザー全員)
			hr := protected.Group("/hr")
			{
				// ダッシュボード
				hr.GET("/stats", h.HRDashboard.GetStats)
				hr.GET("/activities", h.HRDashboard.GetActivities)

				// 社員
				hr.GET("/employees", h.HREmployee.GetAll)
				hr.GET("/employees/:id", h.HREmployee.GetByID)
				hr.POST("/employees", h.HREmployee.Create)
				hr.PUT("/employees/:id", h.HREmployee.Update)
				hr.DELETE("/employees/:id", h.HREmployee.Delete)

				// 部門
				hr.GET("/departments", h.HRDepartment.GetAll)
				hr.GET("/departments/:id", h.HRDepartment.GetByID)
				hr.POST("/departments", h.HRDepartment.Create)
				hr.PUT("/departments/:id", h.HRDepartment.Update)
				hr.DELETE("/departments/:id", h.HRDepartment.Delete)

				// 評価
				hr.GET("/evaluations", h.Evaluation.GetAll)
				hr.GET("/evaluations/:id", h.Evaluation.GetByID)
				hr.POST("/evaluations", h.Evaluation.Create)
				hr.PUT("/evaluations/:id", h.Evaluation.Update)
				hr.PUT("/evaluations/:id/submit", h.Evaluation.Submit)
				hr.GET("/evaluation-cycles", h.Evaluation.GetCycles)
				hr.POST("/evaluation-cycles", h.Evaluation.CreateCycle)

				// 目標
				hr.GET("/goals", h.Goal.GetAll)
				hr.GET("/goals/:id", h.Goal.GetByID)
				hr.POST("/goals", h.Goal.Create)
				hr.PUT("/goals/:id", h.Goal.Update)
				hr.DELETE("/goals/:id", h.Goal.Delete)
				hr.PUT("/goals/:id/progress", h.Goal.UpdateProgress)

				// 研修
				hr.GET("/training", h.Training.GetAll)
				hr.GET("/training/:id", h.Training.GetByID)
				hr.POST("/training", h.Training.Create)
				hr.PUT("/training/:id", h.Training.Update)
				hr.DELETE("/training/:id", h.Training.Delete)
				hr.POST("/training/:id/enroll", h.Training.Enroll)
				hr.PUT("/training/:id/complete", h.Training.Complete)

				// 採用
				hr.GET("/positions", h.Recruitment.GetAllPositions)
				hr.GET("/positions/:id", h.Recruitment.GetPosition)
				hr.POST("/positions", h.Recruitment.CreatePosition)
				hr.PUT("/positions/:id", h.Recruitment.UpdatePosition)
				hr.GET("/applicants", h.Recruitment.GetAllApplicants)
				hr.POST("/applicants", h.Recruitment.CreateApplicant)
				hr.PUT("/applicants/:id/stage", h.Recruitment.UpdateApplicantStage)

				// 書類
				hr.GET("/documents", h.Document.GetAll)
				hr.POST("/documents", h.Document.Upload)
				hr.DELETE("/documents/:id", h.Document.Delete)
				hr.GET("/documents/:id/download", h.Document.Download)

				// お知らせ
				hr.GET("/announcements", h.Announcement.GetAll)
				hr.GET("/announcements/:id", h.Announcement.GetByID)
				hr.POST("/announcements", h.Announcement.Create)
				hr.PUT("/announcements/:id", h.Announcement.Update)
				hr.DELETE("/announcements/:id", h.Announcement.Delete)

				// 勤怠連携
				hr.GET("/attendance-integration", h.AttendanceIntegration.GetIntegration)
				hr.GET("/attendance-integration/alerts", h.AttendanceIntegration.GetAlerts)
				hr.GET("/attendance-integration/trend", h.AttendanceIntegration.GetTrend)

				// 組織図
				hr.GET("/org-chart", h.OrgChart.GetOrgChart)
				hr.POST("/org-chart/simulate", h.OrgChart.Simulate)

				// 1on1
				hr.GET("/one-on-ones", h.OneOnOne.GetAll)
				hr.GET("/one-on-ones/:id", h.OneOnOne.GetByID)
				hr.POST("/one-on-ones", h.OneOnOne.Create)
				hr.PUT("/one-on-ones/:id", h.OneOnOne.Update)
				hr.DELETE("/one-on-ones/:id", h.OneOnOne.Delete)
				hr.POST("/one-on-ones/:id/actions", h.OneOnOne.AddActionItem)
				hr.PUT("/one-on-ones/:id/actions/:actionId/toggle", h.OneOnOne.ToggleActionItem)

				// スキルマップ
				hr.GET("/skill-map", h.Skill.GetSkillMap)
				hr.GET("/skill-map/gap-analysis", h.Skill.GetGapAnalysis)
				hr.POST("/skill-map/:employeeId", h.Skill.AddSkill)
				hr.PUT("/skill-map/:employeeId/:skillId", h.Skill.UpdateSkill)

				// 給与
				hr.GET("/salary", h.Salary.GetOverview)
				hr.POST("/salary/simulate", h.Salary.Simulate)
				hr.GET("/salary/:employeeId/history", h.Salary.GetHistory)
				hr.GET("/salary/budget", h.Salary.GetBudget)

				// オンボーディング
				hr.GET("/onboarding", h.Onboarding.GetAll)
				hr.GET("/onboarding/templates", h.Onboarding.GetTemplates)
				hr.POST("/onboarding/templates", h.Onboarding.CreateTemplate)
				hr.GET("/onboarding/:id", h.Onboarding.GetByID)
				hr.POST("/onboarding", h.Onboarding.Create)
				hr.PUT("/onboarding/:id", h.Onboarding.Update)
				hr.PUT("/onboarding/:id/tasks/:taskId/toggle", h.Onboarding.ToggleTask)

				// オフボーディング
				hr.GET("/offboarding", h.Offboarding.GetAll)
				hr.GET("/offboarding/analytics", h.Offboarding.GetAnalytics)
				hr.GET("/offboarding/:id", h.Offboarding.GetByID)
				hr.POST("/offboarding", h.Offboarding.Create)
				hr.PUT("/offboarding/:id", h.Offboarding.Update)
				hr.PUT("/offboarding/:id/checklist/:itemKey/toggle", h.Offboarding.ToggleChecklist)

				// サーベイ
				hr.GET("/surveys", h.Survey.GetAll)
				hr.GET("/surveys/:id", h.Survey.GetByID)
				hr.POST("/surveys", h.Survey.Create)
				hr.PUT("/surveys/:id", h.Survey.Update)
				hr.DELETE("/surveys/:id", h.Survey.Delete)
				hr.PUT("/surveys/:id/publish", h.Survey.Publish)
				hr.PUT("/surveys/:id/close", h.Survey.Close)
				hr.GET("/surveys/:id/results", h.Survey.GetResults)
				hr.POST("/surveys/:id/respond", h.Survey.SubmitResponse)
			}
		}
	}
}
