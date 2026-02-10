package handler

import (
	apphr "github.com/your-org/kintai/backend/internal/apps/hr"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

type HREmployeeHandler = apphr.HREmployeeHandler
type HRDepartmentHandler = apphr.HRDepartmentHandler
type EvaluationHandler = apphr.EvaluationHandler
type GoalHandler = apphr.GoalHandler
type TrainingHandler = apphr.TrainingHandler
type RecruitmentHandler = apphr.RecruitmentHandler
type DocumentHandler = apphr.DocumentHandler
type AnnouncementHandler = apphr.AnnouncementHandler
type HRDashboardHandler = apphr.HRDashboardHandler
type AttendanceIntegrationHandler = apphr.AttendanceIntegrationHandler
type OrgChartHandler = apphr.OrgChartHandler
type OneOnOneHandler = apphr.OneOnOneHandler
type SkillHandler = apphr.SkillHandler
type SalaryHandler = apphr.SalaryHandler
type OnboardingHandler = apphr.OnboardingHandler
type OffboardingHandler = apphr.OffboardingHandler
type SurveyHandler = apphr.SurveyHandler

func NewHREmployeeHandler(svc service.HREmployeeService, logger *logger.Logger) *HREmployeeHandler {
	return apphr.NewHREmployeeHandler(svc, logger)
}
func NewHRDepartmentHandler(svc service.HRDepartmentService, logger *logger.Logger) *HRDepartmentHandler {
	return apphr.NewHRDepartmentHandler(svc, logger)
}
func NewEvaluationHandler(svc service.EvaluationService, logger *logger.Logger) *EvaluationHandler {
	return apphr.NewEvaluationHandler(svc, logger)
}
func NewGoalHandler(svc service.GoalService, logger *logger.Logger) *GoalHandler {
	return apphr.NewGoalHandler(svc, logger)
}
func NewTrainingHandler(svc service.TrainingService, logger *logger.Logger) *TrainingHandler {
	return apphr.NewTrainingHandler(svc, logger)
}
func NewRecruitmentHandler(svc service.RecruitmentService, logger *logger.Logger) *RecruitmentHandler {
	return apphr.NewRecruitmentHandler(svc, logger)
}
func NewDocumentHandler(svc service.DocumentService, logger *logger.Logger) *DocumentHandler {
	return apphr.NewDocumentHandler(svc, logger)
}
func NewAnnouncementHandler(svc service.AnnouncementService, logger *logger.Logger) *AnnouncementHandler {
	return apphr.NewAnnouncementHandler(svc, logger)
}
func NewHRDashboardHandler(svc service.HRDashboardService, logger *logger.Logger) *HRDashboardHandler {
	return apphr.NewHRDashboardHandler(svc, logger)
}
func NewAttendanceIntegrationHandler(svc service.AttendanceIntegrationService, logger *logger.Logger) *AttendanceIntegrationHandler {
	return apphr.NewAttendanceIntegrationHandler(svc, logger)
}
func NewOrgChartHandler(svc service.OrgChartService, logger *logger.Logger) *OrgChartHandler {
	return apphr.NewOrgChartHandler(svc, logger)
}
func NewOneOnOneHandler(svc service.OneOnOneService, logger *logger.Logger) *OneOnOneHandler {
	return apphr.NewOneOnOneHandler(svc, logger)
}
func NewSkillHandler(svc service.SkillService, logger *logger.Logger) *SkillHandler {
	return apphr.NewSkillHandler(svc, logger)
}
func NewSalaryHandler(svc service.SalaryService, logger *logger.Logger) *SalaryHandler {
	return apphr.NewSalaryHandler(svc, logger)
}
func NewOnboardingHandler(svc service.OnboardingService, logger *logger.Logger) *OnboardingHandler {
	return apphr.NewOnboardingHandler(svc, logger)
}
func NewOffboardingHandler(svc service.OffboardingService, logger *logger.Logger) *OffboardingHandler {
	return apphr.NewOffboardingHandler(svc, logger)
}
func NewSurveyHandler(svc service.SurveyService, logger *logger.Logger) *SurveyHandler {
	return apphr.NewSurveyHandler(svc, logger)
}
