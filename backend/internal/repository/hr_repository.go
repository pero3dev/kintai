package repository

import (
	apphr "github.com/your-org/kintai/backend/internal/apps/hr"
	"gorm.io/gorm"
)

type HREmployeeRepository = apphr.HREmployeeRepository
type HRDepartmentRepository = apphr.HRDepartmentRepository
type EvaluationRepository = apphr.EvaluationRepository
type GoalRepository = apphr.GoalRepository
type TrainingRepository = apphr.TrainingRepository
type RecruitmentRepository = apphr.RecruitmentRepository
type DocumentRepository = apphr.DocumentRepository
type AnnouncementRepository = apphr.AnnouncementRepository
type OneOnOneRepository = apphr.OneOnOneRepository
type SkillRepository = apphr.SkillRepository
type SalaryRepository = apphr.SalaryRepository
type OnboardingRepository = apphr.OnboardingRepository
type OffboardingRepository = apphr.OffboardingRepository
type SurveyRepository = apphr.SurveyRepository

func NewHREmployeeRepository(db *gorm.DB) HREmployeeRepository {
	return apphr.NewHREmployeeRepository(db)
}
func NewHRDepartmentRepository(db *gorm.DB) HRDepartmentRepository {
	return apphr.NewHRDepartmentRepository(db)
}
func NewEvaluationRepository(db *gorm.DB) EvaluationRepository {
	return apphr.NewEvaluationRepository(db)
}
func NewGoalRepository(db *gorm.DB) GoalRepository         { return apphr.NewGoalRepository(db) }
func NewTrainingRepository(db *gorm.DB) TrainingRepository { return apphr.NewTrainingRepository(db) }
func NewRecruitmentRepository(db *gorm.DB) RecruitmentRepository {
	return apphr.NewRecruitmentRepository(db)
}
func NewDocumentRepository(db *gorm.DB) DocumentRepository { return apphr.NewDocumentRepository(db) }
func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return apphr.NewAnnouncementRepository(db)
}
func NewOneOnOneRepository(db *gorm.DB) OneOnOneRepository { return apphr.NewOneOnOneRepository(db) }
func NewSkillRepository(db *gorm.DB) SkillRepository       { return apphr.NewSkillRepository(db) }
func NewSalaryRepository(db *gorm.DB) SalaryRepository     { return apphr.NewSalaryRepository(db) }
func NewOnboardingRepository(db *gorm.DB) OnboardingRepository {
	return apphr.NewOnboardingRepository(db)
}
func NewOffboardingRepository(db *gorm.DB) OffboardingRepository {
	return apphr.NewOffboardingRepository(db)
}
func NewSurveyRepository(db *gorm.DB) SurveyRepository { return apphr.NewSurveyRepository(db) }
