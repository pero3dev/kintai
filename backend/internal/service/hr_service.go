package service

import apphr "github.com/your-org/kintai/backend/internal/apps/hr"

type HREmployeeService = apphr.HREmployeeService
type HRDepartmentService = apphr.HRDepartmentService
type EvaluationService = apphr.EvaluationService
type GoalService = apphr.GoalService
type TrainingService = apphr.TrainingService
type RecruitmentService = apphr.RecruitmentService
type DocumentService = apphr.DocumentService
type AnnouncementService = apphr.AnnouncementService
type HRDashboardService = apphr.HRDashboardService
type AttendanceIntegrationService = apphr.AttendanceIntegrationService
type OrgChartService = apphr.OrgChartService
type OneOnOneService = apphr.OneOnOneService
type SkillService = apphr.SkillService
type SalaryService = apphr.SalaryService
type OnboardingService = apphr.OnboardingService
type OffboardingService = apphr.OffboardingService
type SurveyService = apphr.SurveyService

func toHRDeps(deps Deps) apphr.Deps {
	return apphr.Deps{
		Repos: &apphr.Repositories{
			User:            deps.Repos.User,
			Attendance:      deps.Repos.Attendance,
			LeaveRequest:    deps.Repos.LeaveRequest,
			OvertimeRequest: deps.Repos.OvertimeRequest,
			HREmployee:      deps.Repos.HREmployee,
			HRDepartment:    deps.Repos.HRDepartment,
			Evaluation:      deps.Repos.Evaluation,
			Goal:            deps.Repos.Goal,
			Training:        deps.Repos.Training,
			Recruitment:     deps.Repos.Recruitment,
			Document:        deps.Repos.Document,
			Announcement:    deps.Repos.Announcement,
			OneOnOne:        deps.Repos.OneOnOne,
			Skill:           deps.Repos.Skill,
			Salary:          deps.Repos.Salary,
			Onboarding:      deps.Repos.Onboarding,
			Offboarding:     deps.Repos.Offboarding,
			Survey:          deps.Repos.Survey,
		},
		Config: deps.Config,
		Logger: deps.Logger,
	}
}

func NewHREmployeeService(deps Deps) HREmployeeService {
	return apphr.NewHREmployeeService(toHRDeps(deps))
}
func NewHRDepartmentService(deps Deps) HRDepartmentService {
	return apphr.NewHRDepartmentService(toHRDeps(deps))
}
func NewEvaluationService(deps Deps) EvaluationService {
	return apphr.NewEvaluationService(toHRDeps(deps))
}
func NewGoalService(deps Deps) GoalService         { return apphr.NewGoalService(toHRDeps(deps)) }
func NewTrainingService(deps Deps) TrainingService { return apphr.NewTrainingService(toHRDeps(deps)) }
func NewRecruitmentService(deps Deps) RecruitmentService {
	return apphr.NewRecruitmentService(toHRDeps(deps))
}
func NewDocumentService(deps Deps) DocumentService { return apphr.NewDocumentService(toHRDeps(deps)) }
func NewAnnouncementService(deps Deps) AnnouncementService {
	return apphr.NewAnnouncementService(toHRDeps(deps))
}
func NewHRDashboardService(deps Deps) HRDashboardService {
	return apphr.NewHRDashboardService(toHRDeps(deps))
}
func NewAttendanceIntegrationService(deps Deps) AttendanceIntegrationService {
	return apphr.NewAttendanceIntegrationService(toHRDeps(deps))
}
func NewOrgChartService(deps Deps) OrgChartService { return apphr.NewOrgChartService(toHRDeps(deps)) }
func NewOneOnOneService(deps Deps) OneOnOneService { return apphr.NewOneOnOneService(toHRDeps(deps)) }
func NewSkillService(deps Deps) SkillService       { return apphr.NewSkillService(toHRDeps(deps)) }
func NewSalaryService(deps Deps) SalaryService     { return apphr.NewSalaryService(toHRDeps(deps)) }
func NewOnboardingService(deps Deps) OnboardingService {
	return apphr.NewOnboardingService(toHRDeps(deps))
}
func NewOffboardingService(deps Deps) OffboardingService {
	return apphr.NewOffboardingService(toHRDeps(deps))
}
func NewSurveyService(deps Deps) SurveyService { return apphr.NewSurveyService(toHRDeps(deps)) }
