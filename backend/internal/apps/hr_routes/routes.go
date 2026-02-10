package hr_routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/handler"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, h *handler.Handlers) {
	hr := protected.Group("/hr")
	{
		hr.GET("/stats", h.HRDashboard.GetStats)
		hr.GET("/activities", h.HRDashboard.GetActivities)

		hr.GET("/employees", h.HREmployee.GetAll)
		hr.GET("/employees/:id", h.HREmployee.GetByID)
		hr.POST("/employees", h.HREmployee.Create)
		hr.PUT("/employees/:id", h.HREmployee.Update)
		hr.DELETE("/employees/:id", h.HREmployee.Delete)

		hr.GET("/departments", h.HRDepartment.GetAll)
		hr.GET("/departments/:id", h.HRDepartment.GetByID)
		hr.POST("/departments", h.HRDepartment.Create)
		hr.PUT("/departments/:id", h.HRDepartment.Update)
		hr.DELETE("/departments/:id", h.HRDepartment.Delete)

		hr.GET("/evaluations", h.Evaluation.GetAll)
		hr.GET("/evaluations/:id", h.Evaluation.GetByID)
		hr.POST("/evaluations", h.Evaluation.Create)
		hr.PUT("/evaluations/:id", h.Evaluation.Update)
		hr.PUT("/evaluations/:id/submit", h.Evaluation.Submit)
		hr.GET("/evaluation-cycles", h.Evaluation.GetCycles)
		hr.POST("/evaluation-cycles", h.Evaluation.CreateCycle)

		hr.GET("/goals", h.Goal.GetAll)
		hr.GET("/goals/:id", h.Goal.GetByID)
		hr.POST("/goals", h.Goal.Create)
		hr.PUT("/goals/:id", h.Goal.Update)
		hr.DELETE("/goals/:id", h.Goal.Delete)
		hr.PUT("/goals/:id/progress", h.Goal.UpdateProgress)

		hr.GET("/training", h.Training.GetAll)
		hr.GET("/training/:id", h.Training.GetByID)
		hr.POST("/training", h.Training.Create)
		hr.PUT("/training/:id", h.Training.Update)
		hr.DELETE("/training/:id", h.Training.Delete)
		hr.POST("/training/:id/enroll", h.Training.Enroll)
		hr.PUT("/training/:id/complete", h.Training.Complete)

		hr.GET("/positions", h.Recruitment.GetAllPositions)
		hr.GET("/positions/:id", h.Recruitment.GetPosition)
		hr.POST("/positions", h.Recruitment.CreatePosition)
		hr.PUT("/positions/:id", h.Recruitment.UpdatePosition)
		hr.GET("/applicants", h.Recruitment.GetAllApplicants)
		hr.POST("/applicants", h.Recruitment.CreateApplicant)
		hr.PUT("/applicants/:id/stage", h.Recruitment.UpdateApplicantStage)

		hr.GET("/documents", h.Document.GetAll)
		hr.POST("/documents", h.Document.Upload)
		hr.DELETE("/documents/:id", h.Document.Delete)
		hr.GET("/documents/:id/download", h.Document.Download)

		hr.GET("/announcements", h.Announcement.GetAll)
		hr.GET("/announcements/:id", h.Announcement.GetByID)
		hr.POST("/announcements", h.Announcement.Create)
		hr.PUT("/announcements/:id", h.Announcement.Update)
		hr.DELETE("/announcements/:id", h.Announcement.Delete)

		hr.GET("/attendance-integration", h.AttendanceIntegration.GetIntegration)
		hr.GET("/attendance-integration/alerts", h.AttendanceIntegration.GetAlerts)
		hr.GET("/attendance-integration/trend", h.AttendanceIntegration.GetTrend)

		hr.GET("/org-chart", h.OrgChart.GetOrgChart)
		hr.POST("/org-chart/simulate", h.OrgChart.Simulate)

		hr.GET("/one-on-ones", h.OneOnOne.GetAll)
		hr.GET("/one-on-ones/:id", h.OneOnOne.GetByID)
		hr.POST("/one-on-ones", h.OneOnOne.Create)
		hr.PUT("/one-on-ones/:id", h.OneOnOne.Update)
		hr.DELETE("/one-on-ones/:id", h.OneOnOne.Delete)
		hr.POST("/one-on-ones/:id/actions", h.OneOnOne.AddActionItem)
		hr.PUT("/one-on-ones/:id/actions/:actionId/toggle", h.OneOnOne.ToggleActionItem)

		hr.GET("/skill-map", h.Skill.GetSkillMap)
		hr.GET("/skill-map/gap-analysis", h.Skill.GetGapAnalysis)
		hr.POST("/skill-map/:employeeId", h.Skill.AddSkill)
		hr.PUT("/skill-map/:employeeId/:skillId", h.Skill.UpdateSkill)

		hr.GET("/salary", h.Salary.GetOverview)
		hr.POST("/salary/simulate", h.Salary.Simulate)
		hr.GET("/salary/:employeeId/history", h.Salary.GetHistory)
		hr.GET("/salary/budget", h.Salary.GetBudget)

		hr.GET("/onboarding", h.Onboarding.GetAll)
		hr.GET("/onboarding/templates", h.Onboarding.GetTemplates)
		hr.POST("/onboarding/templates", h.Onboarding.CreateTemplate)
		hr.GET("/onboarding/:id", h.Onboarding.GetByID)
		hr.POST("/onboarding", h.Onboarding.Create)
		hr.PUT("/onboarding/:id", h.Onboarding.Update)
		hr.PUT("/onboarding/:id/tasks/:taskId/toggle", h.Onboarding.ToggleTask)

		hr.GET("/offboarding", h.Offboarding.GetAll)
		hr.GET("/offboarding/analytics", h.Offboarding.GetAnalytics)
		hr.GET("/offboarding/:id", h.Offboarding.GetByID)
		hr.POST("/offboarding", h.Offboarding.Create)
		hr.PUT("/offboarding/:id", h.Offboarding.Update)
		hr.PUT("/offboarding/:id/checklist/:itemKey/toggle", h.Offboarding.ToggleChecklist)

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
