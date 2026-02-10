package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

// ===== MockHREmployeeService =====

type MockHREmployeeService struct {
	CreateFunc   func(ctx context.Context, req model.HREmployeeCreateRequest) (*model.HREmployee, error)
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*model.HREmployee, error)
	FindAllFunc  func(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error)
	UpdateFunc   func(ctx context.Context, id uuid.UUID, req model.HREmployeeUpdateRequest) (*model.HREmployee, error)
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *MockHREmployeeService) Create(ctx context.Context, req model.HREmployeeCreateRequest) (*model.HREmployee, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockHREmployeeService) FindByID(ctx context.Context, id uuid.UUID) (*model.HREmployee, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockHREmployeeService) FindAll(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, page, pageSize, department, status, employmentType, search)
	}
	return nil, 0, nil
}

func (m *MockHREmployeeService) Update(ctx context.Context, id uuid.UUID, req model.HREmployeeUpdateRequest) (*model.HREmployee, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockHREmployeeService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockHRDepartmentService =====

type MockHRDepartmentService struct {
	CreateFunc   func(ctx context.Context, req model.HRDepartmentCreateRequest) (*model.HRDepartment, error)
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error)
	FindAllFunc  func(ctx context.Context) ([]model.HRDepartment, error)
	UpdateFunc   func(ctx context.Context, id uuid.UUID, req model.HRDepartmentUpdateRequest) (*model.HRDepartment, error)
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *MockHRDepartmentService) Create(ctx context.Context, req model.HRDepartmentCreateRequest) (*model.HRDepartment, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockHRDepartmentService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockHRDepartmentService) FindAll(ctx context.Context) ([]model.HRDepartment, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockHRDepartmentService) Update(ctx context.Context, id uuid.UUID, req model.HRDepartmentUpdateRequest) (*model.HRDepartment, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockHRDepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockEvaluationService =====

type MockEvaluationService struct {
	CreateFunc        func(ctx context.Context, req model.EvaluationCreateRequest, reviewerID uuid.UUID) (*model.Evaluation, error)
	FindByIDFunc      func(ctx context.Context, id uuid.UUID) (*model.Evaluation, error)
	FindAllFunc       func(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error)
	UpdateFunc        func(ctx context.Context, id uuid.UUID, req model.EvaluationUpdateRequest) (*model.Evaluation, error)
	SubmitFunc        func(ctx context.Context, id uuid.UUID) (*model.Evaluation, error)
	CreateCycleFunc   func(ctx context.Context, req model.EvaluationCycleCreateRequest) (*model.EvaluationCycle, error)
	FindAllCyclesFunc func(ctx context.Context) ([]model.EvaluationCycle, error)
}

func (m *MockEvaluationService) Create(ctx context.Context, req model.EvaluationCreateRequest, reviewerID uuid.UUID) (*model.Evaluation, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req, reviewerID)
	}
	return nil, nil
}

func (m *MockEvaluationService) FindByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockEvaluationService) FindAll(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, page, pageSize, cycleID, status)
	}
	return nil, 0, nil
}

func (m *MockEvaluationService) Update(ctx context.Context, id uuid.UUID, req model.EvaluationUpdateRequest) (*model.Evaluation, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockEvaluationService) Submit(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	if m.SubmitFunc != nil {
		return m.SubmitFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockEvaluationService) CreateCycle(ctx context.Context, req model.EvaluationCycleCreateRequest) (*model.EvaluationCycle, error) {
	if m.CreateCycleFunc != nil {
		return m.CreateCycleFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockEvaluationService) FindAllCycles(ctx context.Context) ([]model.EvaluationCycle, error) {
	if m.FindAllCyclesFunc != nil {
		return m.FindAllCyclesFunc(ctx)
	}
	return nil, nil
}

// ===== MockGoalService =====

type MockGoalService struct {
	CreateFunc         func(ctx context.Context, req model.HRGoalCreateRequest, userID uuid.UUID) (*model.HRGoal, error)
	FindByIDFunc       func(ctx context.Context, id uuid.UUID) (*model.HRGoal, error)
	FindAllFunc        func(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error)
	UpdateFunc         func(ctx context.Context, id uuid.UUID, req model.HRGoalUpdateRequest) (*model.HRGoal, error)
	UpdateProgressFunc func(ctx context.Context, id uuid.UUID, progress int) (*model.HRGoal, error)
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
}

func (m *MockGoalService) Create(ctx context.Context, req model.HRGoalCreateRequest, userID uuid.UUID) (*model.HRGoal, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req, userID)
	}
	return nil, nil
}

func (m *MockGoalService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRGoal, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockGoalService) FindAll(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, page, pageSize, status, category, employeeID)
	}
	return nil, 0, nil
}

func (m *MockGoalService) Update(ctx context.Context, id uuid.UUID, req model.HRGoalUpdateRequest) (*model.HRGoal, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockGoalService) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) (*model.HRGoal, error) {
	if m.UpdateProgressFunc != nil {
		return m.UpdateProgressFunc(ctx, id, progress)
	}
	return nil, nil
}

func (m *MockGoalService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockTrainingService =====

type MockTrainingService struct {
	CreateFunc   func(ctx context.Context, req model.TrainingProgramCreateRequest) (*model.TrainingProgram, error)
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error)
	FindAllFunc  func(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error)
	UpdateFunc   func(ctx context.Context, id uuid.UUID, req model.TrainingProgramUpdateRequest) (*model.TrainingProgram, error)
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
	EnrollFunc   func(ctx context.Context, programID, employeeID uuid.UUID) error
	CompleteFunc func(ctx context.Context, programID, employeeID uuid.UUID) error
}

func (m *MockTrainingService) Create(ctx context.Context, req model.TrainingProgramCreateRequest) (*model.TrainingProgram, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockTrainingService) FindByID(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockTrainingService) FindAll(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, page, pageSize, category, status)
	}
	return nil, 0, nil
}

func (m *MockTrainingService) Update(ctx context.Context, id uuid.UUID, req model.TrainingProgramUpdateRequest) (*model.TrainingProgram, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockTrainingService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockTrainingService) Enroll(ctx context.Context, programID, employeeID uuid.UUID) error {
	if m.EnrollFunc != nil {
		return m.EnrollFunc(ctx, programID, employeeID)
	}
	return nil
}

func (m *MockTrainingService) Complete(ctx context.Context, programID, employeeID uuid.UUID) error {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, programID, employeeID)
	}
	return nil
}

// ===== MockRecruitmentService =====

type MockRecruitmentService struct {
	CreatePositionFunc       func(ctx context.Context, req model.PositionCreateRequest) (*model.RecruitmentPosition, error)
	FindPositionByIDFunc     func(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error)
	FindAllPositionsFunc     func(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error)
	UpdatePositionFunc       func(ctx context.Context, id uuid.UUID, req model.PositionUpdateRequest) (*model.RecruitmentPosition, error)
	CreateApplicantFunc      func(ctx context.Context, req model.ApplicantCreateRequest) (*model.Applicant, error)
	FindAllApplicantsFunc    func(ctx context.Context, positionID, stage string) ([]model.Applicant, error)
	UpdateApplicantStageFunc func(ctx context.Context, id uuid.UUID, stage string) (*model.Applicant, error)
}

func (m *MockRecruitmentService) CreatePosition(ctx context.Context, req model.PositionCreateRequest) (*model.RecruitmentPosition, error) {
	if m.CreatePositionFunc != nil {
		return m.CreatePositionFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockRecruitmentService) FindPositionByID(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error) {
	if m.FindPositionByIDFunc != nil {
		return m.FindPositionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRecruitmentService) FindAllPositions(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error) {
	if m.FindAllPositionsFunc != nil {
		return m.FindAllPositionsFunc(ctx, page, pageSize, status, department)
	}
	return nil, 0, nil
}

func (m *MockRecruitmentService) UpdatePosition(ctx context.Context, id uuid.UUID, req model.PositionUpdateRequest) (*model.RecruitmentPosition, error) {
	if m.UpdatePositionFunc != nil {
		return m.UpdatePositionFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockRecruitmentService) CreateApplicant(ctx context.Context, req model.ApplicantCreateRequest) (*model.Applicant, error) {
	if m.CreateApplicantFunc != nil {
		return m.CreateApplicantFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockRecruitmentService) FindAllApplicants(ctx context.Context, positionID, stage string) ([]model.Applicant, error) {
	if m.FindAllApplicantsFunc != nil {
		return m.FindAllApplicantsFunc(ctx, positionID, stage)
	}
	return nil, nil
}

func (m *MockRecruitmentService) UpdateApplicantStage(ctx context.Context, id uuid.UUID, stage string) (*model.Applicant, error) {
	if m.UpdateApplicantStageFunc != nil {
		return m.UpdateApplicantStageFunc(ctx, id, stage)
	}
	return nil, nil
}

// ===== MockDocumentService =====

type MockDocumentService struct {
	UploadFunc   func(ctx context.Context, doc *model.HRDocument) error
	FindAllFunc  func(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error)
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*model.HRDocument, error)
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *MockDocumentService) Upload(ctx context.Context, doc *model.HRDocument) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, doc)
	}
	return nil
}

func (m *MockDocumentService) FindAll(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, page, pageSize, docType, employeeID)
	}
	return nil, 0, nil
}

func (m *MockDocumentService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDocument, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDocumentService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockAnnouncementService =====

type MockAnnouncementService struct {
	CreateFunc   func(ctx context.Context, req model.AnnouncementCreateRequest, authorID uuid.UUID) (*model.HRAnnouncement, error)
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error)
	FindAllFunc  func(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error)
	UpdateFunc   func(ctx context.Context, id uuid.UUID, req model.AnnouncementUpdateRequest) (*model.HRAnnouncement, error)
	DeleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *MockAnnouncementService) Create(ctx context.Context, req model.AnnouncementCreateRequest, authorID uuid.UUID) (*model.HRAnnouncement, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req, authorID)
	}
	return nil, nil
}

func (m *MockAnnouncementService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockAnnouncementService) FindAll(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, page, pageSize, priority)
	}
	return nil, 0, nil
}

func (m *MockAnnouncementService) Update(ctx context.Context, id uuid.UUID, req model.AnnouncementUpdateRequest) (*model.HRAnnouncement, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockAnnouncementService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockHRDashboardService =====

type MockHRDashboardService struct {
	GetStatsFunc            func(ctx context.Context) (map[string]interface{}, error)
	GetRecentActivitiesFunc func(ctx context.Context) ([]map[string]interface{}, error)
}

func (m *MockHRDashboardService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx)
	}
	return nil, nil
}

func (m *MockHRDashboardService) GetRecentActivities(ctx context.Context) ([]map[string]interface{}, error) {
	if m.GetRecentActivitiesFunc != nil {
		return m.GetRecentActivitiesFunc(ctx)
	}
	return nil, nil
}

// ===== MockAttendanceIntegrationService =====

type MockAttendanceIntegrationService struct {
	GetIntegrationFunc func(ctx context.Context, period, department string) (map[string]interface{}, error)
	GetAlertsFunc      func(ctx context.Context) ([]map[string]interface{}, error)
	GetTrendFunc       func(ctx context.Context, period string) ([]map[string]interface{}, error)
}

func (m *MockAttendanceIntegrationService) GetIntegration(ctx context.Context, period, department string) (map[string]interface{}, error) {
	if m.GetIntegrationFunc != nil {
		return m.GetIntegrationFunc(ctx, period, department)
	}
	return nil, nil
}

func (m *MockAttendanceIntegrationService) GetAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	if m.GetAlertsFunc != nil {
		return m.GetAlertsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAttendanceIntegrationService) GetTrend(ctx context.Context, period string) ([]map[string]interface{}, error) {
	if m.GetTrendFunc != nil {
		return m.GetTrendFunc(ctx, period)
	}
	return nil, nil
}

// ===== MockOrgChartService =====

type MockOrgChartService struct {
	GetOrgChartFunc func(ctx context.Context) ([]map[string]interface{}, error)
	SimulateFunc    func(ctx context.Context, data map[string]interface{}) ([]map[string]interface{}, error)
}

func (m *MockOrgChartService) GetOrgChart(ctx context.Context) ([]map[string]interface{}, error) {
	if m.GetOrgChartFunc != nil {
		return m.GetOrgChartFunc(ctx)
	}
	return nil, nil
}

func (m *MockOrgChartService) Simulate(ctx context.Context, data map[string]interface{}) ([]map[string]interface{}, error) {
	if m.SimulateFunc != nil {
		return m.SimulateFunc(ctx, data)
	}
	return nil, nil
}

// ===== MockOneOnOneService =====

type MockOneOnOneService struct {
	CreateFunc           func(ctx context.Context, req model.OneOnOneCreateRequest, managerID uuid.UUID) (*model.OneOnOneMeeting, error)
	FindByIDFunc         func(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error)
	FindAllFunc          func(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error)
	UpdateFunc           func(ctx context.Context, id uuid.UUID, req model.OneOnOneUpdateRequest) (*model.OneOnOneMeeting, error)
	DeleteFunc           func(ctx context.Context, id uuid.UUID) error
	AddActionItemFunc    func(ctx context.Context, meetingID uuid.UUID, req model.ActionItemRequest) (*model.OneOnOneMeeting, error)
	ToggleActionItemFunc func(ctx context.Context, meetingID uuid.UUID, actionID string) (*model.OneOnOneMeeting, error)
}

func (m *MockOneOnOneService) Create(ctx context.Context, req model.OneOnOneCreateRequest, managerID uuid.UUID) (*model.OneOnOneMeeting, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req, managerID)
	}
	return nil, nil
}

func (m *MockOneOnOneService) FindByID(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockOneOnOneService) FindAll(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, status, employeeID)
	}
	return nil, nil
}

func (m *MockOneOnOneService) Update(ctx context.Context, id uuid.UUID, req model.OneOnOneUpdateRequest) (*model.OneOnOneMeeting, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockOneOnOneService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockOneOnOneService) AddActionItem(ctx context.Context, meetingID uuid.UUID, req model.ActionItemRequest) (*model.OneOnOneMeeting, error) {
	if m.AddActionItemFunc != nil {
		return m.AddActionItemFunc(ctx, meetingID, req)
	}
	return nil, nil
}

func (m *MockOneOnOneService) ToggleActionItem(ctx context.Context, meetingID uuid.UUID, actionID string) (*model.OneOnOneMeeting, error) {
	if m.ToggleActionItemFunc != nil {
		return m.ToggleActionItemFunc(ctx, meetingID, actionID)
	}
	return nil, nil
}

// ===== MockSkillService =====

type MockSkillService struct {
	GetSkillMapFunc    func(ctx context.Context, department, employeeID string) ([]model.EmployeeSkill, error)
	GetGapAnalysisFunc func(ctx context.Context, department string) ([]map[string]interface{}, error)
	AddSkillFunc       func(ctx context.Context, employeeID uuid.UUID, req model.SkillAddRequest) (*model.EmployeeSkill, error)
	UpdateSkillFunc    func(ctx context.Context, skillID uuid.UUID, req model.SkillUpdateRequest) (*model.EmployeeSkill, error)
}

func (m *MockSkillService) GetSkillMap(ctx context.Context, department, employeeID string) ([]model.EmployeeSkill, error) {
	if m.GetSkillMapFunc != nil {
		return m.GetSkillMapFunc(ctx, department, employeeID)
	}
	return nil, nil
}

func (m *MockSkillService) GetGapAnalysis(ctx context.Context, department string) ([]map[string]interface{}, error) {
	if m.GetGapAnalysisFunc != nil {
		return m.GetGapAnalysisFunc(ctx, department)
	}
	return nil, nil
}

func (m *MockSkillService) AddSkill(ctx context.Context, employeeID uuid.UUID, req model.SkillAddRequest) (*model.EmployeeSkill, error) {
	if m.AddSkillFunc != nil {
		return m.AddSkillFunc(ctx, employeeID, req)
	}
	return nil, nil
}

func (m *MockSkillService) UpdateSkill(ctx context.Context, skillID uuid.UUID, req model.SkillUpdateRequest) (*model.EmployeeSkill, error) {
	if m.UpdateSkillFunc != nil {
		return m.UpdateSkillFunc(ctx, skillID, req)
	}
	return nil, nil
}

// ===== MockSalaryService =====

type MockSalaryService struct {
	GetOverviewFunc func(ctx context.Context, department string) (map[string]interface{}, error)
	SimulateFunc    func(ctx context.Context, req model.SalarySimulateRequest) (map[string]interface{}, error)
	GetHistoryFunc  func(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error)
	GetBudgetFunc   func(ctx context.Context, department string) (map[string]interface{}, error)
}

func (m *MockSalaryService) GetOverview(ctx context.Context, department string) (map[string]interface{}, error) {
	if m.GetOverviewFunc != nil {
		return m.GetOverviewFunc(ctx, department)
	}
	return nil, nil
}

func (m *MockSalaryService) Simulate(ctx context.Context, req model.SalarySimulateRequest) (map[string]interface{}, error) {
	if m.SimulateFunc != nil {
		return m.SimulateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockSalaryService) GetHistory(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(ctx, employeeID)
	}
	return nil, nil
}

func (m *MockSalaryService) GetBudget(ctx context.Context, department string) (map[string]interface{}, error) {
	if m.GetBudgetFunc != nil {
		return m.GetBudgetFunc(ctx, department)
	}
	return nil, nil
}

// ===== MockOnboardingService =====

type MockOnboardingService struct {
	CreateFunc           func(ctx context.Context, req model.OnboardingCreateRequest) (*model.Onboarding, error)
	FindByIDFunc         func(ctx context.Context, id uuid.UUID) (*model.Onboarding, error)
	FindAllFunc          func(ctx context.Context, status string) ([]model.Onboarding, error)
	UpdateFunc           func(ctx context.Context, id uuid.UUID, data map[string]interface{}) (*model.Onboarding, error)
	ToggleTaskFunc       func(ctx context.Context, id uuid.UUID, taskID string) (*model.Onboarding, error)
	CreateTemplateFunc   func(ctx context.Context, req model.OnboardingTemplateCreateRequest) (*model.OnboardingTemplate, error)
	FindAllTemplatesFunc func(ctx context.Context) ([]model.OnboardingTemplate, error)
}

func (m *MockOnboardingService) Create(ctx context.Context, req model.OnboardingCreateRequest) (*model.Onboarding, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockOnboardingService) FindByID(ctx context.Context, id uuid.UUID) (*model.Onboarding, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockOnboardingService) FindAll(ctx context.Context, status string) ([]model.Onboarding, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, status)
	}
	return nil, nil
}

func (m *MockOnboardingService) Update(ctx context.Context, id uuid.UUID, data map[string]interface{}) (*model.Onboarding, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, data)
	}
	return nil, nil
}

func (m *MockOnboardingService) ToggleTask(ctx context.Context, id uuid.UUID, taskID string) (*model.Onboarding, error) {
	if m.ToggleTaskFunc != nil {
		return m.ToggleTaskFunc(ctx, id, taskID)
	}
	return nil, nil
}

func (m *MockOnboardingService) CreateTemplate(ctx context.Context, req model.OnboardingTemplateCreateRequest) (*model.OnboardingTemplate, error) {
	if m.CreateTemplateFunc != nil {
		return m.CreateTemplateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockOnboardingService) FindAllTemplates(ctx context.Context) ([]model.OnboardingTemplate, error) {
	if m.FindAllTemplatesFunc != nil {
		return m.FindAllTemplatesFunc(ctx)
	}
	return nil, nil
}

// ===== MockOffboardingService =====

type MockOffboardingService struct {
	CreateFunc          func(ctx context.Context, req model.OffboardingCreateRequest) (*model.Offboarding, error)
	FindByIDFunc        func(ctx context.Context, id uuid.UUID) (*model.Offboarding, error)
	FindAllFunc         func(ctx context.Context, status string) ([]model.Offboarding, error)
	UpdateFunc          func(ctx context.Context, id uuid.UUID, req model.OffboardingUpdateRequest) (*model.Offboarding, error)
	ToggleChecklistFunc func(ctx context.Context, id uuid.UUID, itemKey string) (*model.Offboarding, error)
	GetAnalyticsFunc    func(ctx context.Context) (map[string]interface{}, error)
}

func (m *MockOffboardingService) Create(ctx context.Context, req model.OffboardingCreateRequest) (*model.Offboarding, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockOffboardingService) FindByID(ctx context.Context, id uuid.UUID) (*model.Offboarding, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockOffboardingService) FindAll(ctx context.Context, status string) ([]model.Offboarding, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, status)
	}
	return nil, nil
}

func (m *MockOffboardingService) Update(ctx context.Context, id uuid.UUID, req model.OffboardingUpdateRequest) (*model.Offboarding, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockOffboardingService) ToggleChecklist(ctx context.Context, id uuid.UUID, itemKey string) (*model.Offboarding, error) {
	if m.ToggleChecklistFunc != nil {
		return m.ToggleChecklistFunc(ctx, id, itemKey)
	}
	return nil, nil
}

func (m *MockOffboardingService) GetAnalytics(ctx context.Context) (map[string]interface{}, error) {
	if m.GetAnalyticsFunc != nil {
		return m.GetAnalyticsFunc(ctx)
	}
	return nil, nil
}

// ===== MockSurveyService =====

type MockSurveyService struct {
	CreateFunc         func(ctx context.Context, req model.SurveyCreateRequest, createdBy uuid.UUID) (*model.Survey, error)
	FindByIDFunc       func(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	FindAllFunc        func(ctx context.Context, status, surveyType string) ([]model.Survey, error)
	UpdateFunc         func(ctx context.Context, id uuid.UUID, req model.SurveyUpdateRequest) (*model.Survey, error)
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
	PublishFunc        func(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	CloseFunc          func(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	GetResultsFunc     func(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	SubmitResponseFunc func(ctx context.Context, surveyID uuid.UUID, employeeID *uuid.UUID, req model.SurveyResponseRequest) error
}

func (m *MockSurveyService) Create(ctx context.Context, req model.SurveyCreateRequest, createdBy uuid.UUID) (*model.Survey, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req, createdBy)
	}
	return nil, nil
}

func (m *MockSurveyService) FindByID(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSurveyService) FindAll(ctx context.Context, status, surveyType string) ([]model.Survey, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, status, surveyType)
	}
	return nil, nil
}

func (m *MockSurveyService) Update(ctx context.Context, id uuid.UUID, req model.SurveyUpdateRequest) (*model.Survey, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockSurveyService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSurveyService) Publish(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSurveyService) Close(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSurveyService) GetResults(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	if m.GetResultsFunc != nil {
		return m.GetResultsFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSurveyService) SubmitResponse(ctx context.Context, surveyID uuid.UUID, employeeID *uuid.UUID, req model.SurveyResponseRequest) error {
	if m.SubmitResponseFunc != nil {
		return m.SubmitResponseFunc(ctx, surveyID, employeeID, req)
	}
	return nil
}

// ===== MockExpenseService =====

type MockExpenseService struct {
	CreateFunc          func(ctx context.Context, userID uuid.UUID, req *model.ExpenseCreateRequest) (*model.Expense, error)
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*model.Expense, error)
	GetListFunc         func(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error)
	UpdateFunc          func(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *model.ExpenseUpdateRequest) (*model.Expense, error)
	DeleteFunc          func(ctx context.Context, id uuid.UUID) error
	GetPendingFunc      func(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error)
	ApproveFunc         func(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseApproveRequest) error
	AdvancedApproveFunc func(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseAdvancedApproveRequest) error
	GetStatsFunc        func(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error)
	GetReportFunc       func(ctx context.Context, startDate, endDate string) (*model.ExpenseReportResponse, error)
	GetMonthlyTrendFunc func(ctx context.Context, year string) ([]model.MonthlyTrendItem, error)
	ExportCSVFunc       func(ctx context.Context, startDate, endDate string) ([]byte, error)
	ExportPDFFunc       func(ctx context.Context, startDate, endDate string) ([]byte, error)
}

func (m *MockExpenseService) Create(ctx context.Context, userID uuid.UUID, req *model.ExpenseCreateRequest) (*model.Expense, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockExpenseService) GetByID(ctx context.Context, id uuid.UUID) (*model.Expense, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockExpenseService) GetList(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
	if m.GetListFunc != nil {
		return m.GetListFunc(ctx, userID, page, pageSize, status, category)
	}
	return nil, 0, nil
}

func (m *MockExpenseService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *model.ExpenseUpdateRequest) (*model.Expense, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, userID, req)
	}
	return nil, nil
}

func (m *MockExpenseService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockExpenseService) GetPending(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error) {
	if m.GetPendingFunc != nil {
		return m.GetPendingFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockExpenseService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseApproveRequest) error {
	if m.ApproveFunc != nil {
		return m.ApproveFunc(ctx, id, approverID, req)
	}
	return nil
}

func (m *MockExpenseService) AdvancedApprove(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseAdvancedApproveRequest) error {
	if m.AdvancedApproveFunc != nil {
		return m.AdvancedApproveFunc(ctx, id, approverID, req)
	}
	return nil
}

func (m *MockExpenseService) GetStats(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockExpenseService) GetReport(ctx context.Context, startDate, endDate string) (*model.ExpenseReportResponse, error) {
	if m.GetReportFunc != nil {
		return m.GetReportFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func (m *MockExpenseService) GetMonthlyTrend(ctx context.Context, year string) ([]model.MonthlyTrendItem, error) {
	if m.GetMonthlyTrendFunc != nil {
		return m.GetMonthlyTrendFunc(ctx, year)
	}
	return nil, nil
}

func (m *MockExpenseService) ExportCSV(ctx context.Context, startDate, endDate string) ([]byte, error) {
	if m.ExportCSVFunc != nil {
		return m.ExportCSVFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func (m *MockExpenseService) ExportPDF(ctx context.Context, startDate, endDate string) ([]byte, error) {
	if m.ExportPDFFunc != nil {
		return m.ExportPDFFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

// ===== MockExpenseCommentService =====

type MockExpenseCommentService struct {
	GetCommentsFunc func(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseCommentResponse, error)
	AddCommentFunc  func(ctx context.Context, expenseID, userID uuid.UUID, req *model.ExpenseCommentRequest) (*model.ExpenseCommentResponse, error)
}

func (m *MockExpenseCommentService) GetComments(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseCommentResponse, error) {
	if m.GetCommentsFunc != nil {
		return m.GetCommentsFunc(ctx, expenseID)
	}
	return nil, nil
}

func (m *MockExpenseCommentService) AddComment(ctx context.Context, expenseID, userID uuid.UUID, req *model.ExpenseCommentRequest) (*model.ExpenseCommentResponse, error) {
	if m.AddCommentFunc != nil {
		return m.AddCommentFunc(ctx, expenseID, userID, req)
	}
	return nil, nil
}

// ===== MockExpenseHistoryService =====

type MockExpenseHistoryService struct {
	GetHistoryFunc func(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistoryResponse, error)
}

func (m *MockExpenseHistoryService) GetHistory(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistoryResponse, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(ctx, expenseID)
	}
	return nil, nil
}

// ===== MockExpenseReceiptService =====

type MockExpenseReceiptService struct {
	UploadFunc func(ctx context.Context, filename string, data []byte) (string, error)
}

func (m *MockExpenseReceiptService) Upload(ctx context.Context, filename string, data []byte) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, filename, data)
	}
	return "", nil
}

// ===== MockExpenseTemplateService =====

type MockExpenseTemplateService struct {
	GetTemplatesFunc func(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error)
	CreateFunc       func(ctx context.Context, userID uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error)
	UpdateFunc       func(ctx context.Context, id uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error)
	DeleteFunc       func(ctx context.Context, id uuid.UUID) error
	UseTemplateFunc  func(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Expense, error)
}

func (m *MockExpenseTemplateService) GetTemplates(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error) {
	if m.GetTemplatesFunc != nil {
		return m.GetTemplatesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockExpenseTemplateService) Create(ctx context.Context, userID uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockExpenseTemplateService) Update(ctx context.Context, id uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockExpenseTemplateService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockExpenseTemplateService) UseTemplate(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Expense, error) {
	if m.UseTemplateFunc != nil {
		return m.UseTemplateFunc(ctx, id, userID)
	}
	return nil, nil
}

// ===== MockExpensePolicyService =====

type MockExpensePolicyService struct {
	GetPoliciesFunc         func(ctx context.Context) ([]model.ExpensePolicy, error)
	CreateFunc              func(ctx context.Context, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error)
	UpdateFunc              func(ctx context.Context, id uuid.UUID, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error)
	DeleteFunc              func(ctx context.Context, id uuid.UUID) error
	GetBudgetsFunc          func(ctx context.Context) ([]model.ExpenseBudget, error)
	GetPolicyViolationsFunc func(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error)
}

func (m *MockExpensePolicyService) GetPolicies(ctx context.Context) ([]model.ExpensePolicy, error) {
	if m.GetPoliciesFunc != nil {
		return m.GetPoliciesFunc(ctx)
	}
	return nil, nil
}

func (m *MockExpensePolicyService) Create(ctx context.Context, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockExpensePolicyService) Update(ctx context.Context, id uuid.UUID, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockExpensePolicyService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockExpensePolicyService) GetBudgets(ctx context.Context) ([]model.ExpenseBudget, error) {
	if m.GetBudgetsFunc != nil {
		return m.GetBudgetsFunc(ctx)
	}
	return nil, nil
}

func (m *MockExpensePolicyService) GetPolicyViolations(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error) {
	if m.GetPolicyViolationsFunc != nil {
		return m.GetPolicyViolationsFunc(ctx, userID)
	}
	return nil, nil
}

// ===== MockExpenseNotificationService =====

type MockExpenseNotificationService struct {
	GetNotificationsFunc func(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error)
	MarkAsReadFunc       func(ctx context.Context, id uuid.UUID) error
	MarkAllAsReadFunc    func(ctx context.Context, userID uuid.UUID) error
	GetRemindersFunc     func(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error)
	DismissReminderFunc  func(ctx context.Context, id uuid.UUID) error
	GetSettingsFunc      func(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error)
	UpdateSettingsFunc   func(ctx context.Context, userID uuid.UUID, req *model.ExpenseNotificationSettingRequest) (*model.ExpenseNotificationSetting, error)
}

func (m *MockExpenseNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error) {
	if m.GetNotificationsFunc != nil {
		return m.GetNotificationsFunc(ctx, userID, filter)
	}
	return nil, nil
}

func (m *MockExpenseNotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(ctx, id)
	}
	return nil
}

func (m *MockExpenseNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.MarkAllAsReadFunc != nil {
		return m.MarkAllAsReadFunc(ctx, userID)
	}
	return nil
}

func (m *MockExpenseNotificationService) GetReminders(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error) {
	if m.GetRemindersFunc != nil {
		return m.GetRemindersFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockExpenseNotificationService) DismissReminder(ctx context.Context, id uuid.UUID) error {
	if m.DismissReminderFunc != nil {
		return m.DismissReminderFunc(ctx, id)
	}
	return nil
}

func (m *MockExpenseNotificationService) GetSettings(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error) {
	if m.GetSettingsFunc != nil {
		return m.GetSettingsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockExpenseNotificationService) UpdateSettings(ctx context.Context, userID uuid.UUID, req *model.ExpenseNotificationSettingRequest) (*model.ExpenseNotificationSetting, error) {
	if m.UpdateSettingsFunc != nil {
		return m.UpdateSettingsFunc(ctx, userID, req)
	}
	return nil, nil
}

// ===== MockExpenseApprovalFlowService =====

type MockExpenseApprovalFlowService struct {
	GetConfigFunc      func(ctx context.Context) (*model.ExpenseApprovalFlow, error)
	GetDelegatesFunc   func(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error)
	SetDelegateFunc    func(ctx context.Context, userID uuid.UUID, req *model.ExpenseDelegateRequest) (*model.ExpenseDelegate, error)
	RemoveDelegateFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *MockExpenseApprovalFlowService) GetConfig(ctx context.Context) (*model.ExpenseApprovalFlow, error) {
	if m.GetConfigFunc != nil {
		return m.GetConfigFunc(ctx)
	}
	return nil, nil
}

func (m *MockExpenseApprovalFlowService) GetDelegates(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error) {
	if m.GetDelegatesFunc != nil {
		return m.GetDelegatesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockExpenseApprovalFlowService) SetDelegate(ctx context.Context, userID uuid.UUID, req *model.ExpenseDelegateRequest) (*model.ExpenseDelegate, error) {
	if m.SetDelegateFunc != nil {
		return m.SetDelegateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockExpenseApprovalFlowService) RemoveDelegate(ctx context.Context, id uuid.UUID) error {
	if m.RemoveDelegateFunc != nil {
		return m.RemoveDelegateFunc(ctx, id)
	}
	return nil
}

// ===== MockOvertimeRequestService =====

type MockOvertimeRequestService struct {
	CreateFunc            func(ctx context.Context, userID uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error)
	ApproveFunc           func(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error)
	GetByUserFunc         func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	GetPendingFunc        func(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	GetOvertimeAlertsFunc func(ctx context.Context) ([]model.OvertimeAlert, error)
}

func (m *MockOvertimeRequestService) Create(ctx context.Context, userID uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockOvertimeRequestService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error) {
	if m.ApproveFunc != nil {
		return m.ApproveFunc(ctx, id, approverID, req)
	}
	return nil, nil
}

func (m *MockOvertimeRequestService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	if m.GetByUserFunc != nil {
		return m.GetByUserFunc(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockOvertimeRequestService) GetPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	if m.GetPendingFunc != nil {
		return m.GetPendingFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockOvertimeRequestService) GetOvertimeAlerts(ctx context.Context) ([]model.OvertimeAlert, error) {
	if m.GetOvertimeAlertsFunc != nil {
		return m.GetOvertimeAlertsFunc(ctx)
	}
	return nil, nil
}

// ===== MockLeaveBalanceService =====

type MockLeaveBalanceService struct {
	GetByUserFunc         func(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error)
	SetBalanceFunc        func(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error
	DeductBalanceFunc     func(ctx context.Context, userID uuid.UUID, leaveType model.LeaveType, days float64) error
	InitializeForUserFunc func(ctx context.Context, userID uuid.UUID, fiscalYear int) error
}

func (m *MockLeaveBalanceService) GetByUser(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error) {
	if m.GetByUserFunc != nil {
		return m.GetByUserFunc(ctx, userID, fiscalYear)
	}
	return nil, nil
}

func (m *MockLeaveBalanceService) SetBalance(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error {
	if m.SetBalanceFunc != nil {
		return m.SetBalanceFunc(ctx, userID, fiscalYear, leaveType, req)
	}
	return nil
}

func (m *MockLeaveBalanceService) DeductBalance(ctx context.Context, userID uuid.UUID, leaveType model.LeaveType, days float64) error {
	if m.DeductBalanceFunc != nil {
		return m.DeductBalanceFunc(ctx, userID, leaveType, days)
	}
	return nil
}

func (m *MockLeaveBalanceService) InitializeForUser(ctx context.Context, userID uuid.UUID, fiscalYear int) error {
	if m.InitializeForUserFunc != nil {
		return m.InitializeForUserFunc(ctx, userID, fiscalYear)
	}
	return nil
}

// ===== MockAttendanceCorrectionService =====

type MockAttendanceCorrectionService struct {
	CreateFunc     func(ctx context.Context, userID uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error)
	ApproveFunc    func(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error)
	GetByUserFunc  func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
	GetPendingFunc func(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
}

func (m *MockAttendanceCorrectionService) Create(ctx context.Context, userID uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockAttendanceCorrectionService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error) {
	if m.ApproveFunc != nil {
		return m.ApproveFunc(ctx, id, approverID, req)
	}
	return nil, nil
}

func (m *MockAttendanceCorrectionService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	if m.GetByUserFunc != nil {
		return m.GetByUserFunc(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockAttendanceCorrectionService) GetPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	if m.GetPendingFunc != nil {
		return m.GetPendingFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

// ===== MockProjectService =====

type MockProjectService struct {
	CreateFunc  func(ctx context.Context, req *model.ProjectCreateRequest) (*model.Project, error)
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*model.Project, error)
	GetAllFunc  func(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error)
	UpdateFunc  func(ctx context.Context, id uuid.UUID, req *model.ProjectUpdateRequest) (*model.Project, error)
	DeleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockProjectService) Create(ctx context.Context, req *model.ProjectCreateRequest) (*model.Project, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockProjectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProjectService) GetAll(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx, status, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockProjectService) Update(ctx context.Context, id uuid.UUID, req *model.ProjectUpdateRequest) (*model.Project, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockTimeEntryService =====

type MockTimeEntryService struct {
	CreateFunc                   func(ctx context.Context, userID uuid.UUID, req *model.TimeEntryCreate) (*model.TimeEntry, error)
	GetByUserAndDateRangeFunc    func(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error)
	GetByProjectAndDateRangeFunc func(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error)
	UpdateFunc                   func(ctx context.Context, id uuid.UUID, req *model.TimeEntryUpdate) (*model.TimeEntry, error)
	DeleteFunc                   func(ctx context.Context, id uuid.UUID) error
	GetProjectSummaryFunc        func(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error)
}

func (m *MockTimeEntryService) Create(ctx context.Context, userID uuid.UUID, req *model.TimeEntryCreate) (*model.TimeEntry, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockTimeEntryService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	if m.GetByUserAndDateRangeFunc != nil {
		return m.GetByUserAndDateRangeFunc(ctx, userID, start, end)
	}
	return nil, nil
}

func (m *MockTimeEntryService) GetByProjectAndDateRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	if m.GetByProjectAndDateRangeFunc != nil {
		return m.GetByProjectAndDateRangeFunc(ctx, projectID, start, end)
	}
	return nil, nil
}

func (m *MockTimeEntryService) Update(ctx context.Context, id uuid.UUID, req *model.TimeEntryUpdate) (*model.TimeEntry, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockTimeEntryService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockTimeEntryService) GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
	if m.GetProjectSummaryFunc != nil {
		return m.GetProjectSummaryFunc(ctx, start, end)
	}
	return nil, nil
}

// ===== MockHolidayService =====

type MockHolidayService struct {
	CreateFunc         func(ctx context.Context, req *model.HolidayCreateRequest) (*model.Holiday, error)
	GetByDateRangeFunc func(ctx context.Context, start, end time.Time) ([]model.Holiday, error)
	GetByYearFunc      func(ctx context.Context, year int) ([]model.Holiday, error)
	UpdateFunc         func(ctx context.Context, id uuid.UUID, req *model.HolidayUpdateRequest) (*model.Holiday, error)
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
	GetCalendarFunc    func(ctx context.Context, year, month int) ([]model.CalendarDay, error)
	GetWorkingDaysFunc func(ctx context.Context, start, end time.Time) (*model.WorkingDaysSummary, error)
}

func (m *MockHolidayService) Create(ctx context.Context, req *model.HolidayCreateRequest) (*model.Holiday, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockHolidayService) GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error) {
	if m.GetByDateRangeFunc != nil {
		return m.GetByDateRangeFunc(ctx, start, end)
	}
	return nil, nil
}

func (m *MockHolidayService) GetByYear(ctx context.Context, year int) ([]model.Holiday, error) {
	if m.GetByYearFunc != nil {
		return m.GetByYearFunc(ctx, year)
	}
	return nil, nil
}

func (m *MockHolidayService) Update(ctx context.Context, id uuid.UUID, req *model.HolidayUpdateRequest) (*model.Holiday, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockHolidayService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockHolidayService) GetCalendar(ctx context.Context, year, month int) ([]model.CalendarDay, error) {
	if m.GetCalendarFunc != nil {
		return m.GetCalendarFunc(ctx, year, month)
	}
	return nil, nil
}

func (m *MockHolidayService) GetWorkingDays(ctx context.Context, start, end time.Time) (*model.WorkingDaysSummary, error) {
	if m.GetWorkingDaysFunc != nil {
		return m.GetWorkingDaysFunc(ctx, start, end)
	}
	return nil, nil
}

// ===== MockApprovalFlowService =====

type MockApprovalFlowService struct {
	CreateFunc    func(ctx context.Context, req *model.ApprovalFlowCreateRequest) (*model.ApprovalFlow, error)
	GetAllFunc    func(ctx context.Context) ([]model.ApprovalFlow, error)
	GetByIDFunc   func(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error)
	GetByTypeFunc func(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error)
	UpdateFunc    func(ctx context.Context, id uuid.UUID, req *model.ApprovalFlowUpdateRequest) (*model.ApprovalFlow, error)
	DeleteFunc    func(ctx context.Context, id uuid.UUID) error
}

func (m *MockApprovalFlowService) Create(ctx context.Context, req *model.ApprovalFlowCreateRequest) (*model.ApprovalFlow, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockApprovalFlowService) GetAll(ctx context.Context) ([]model.ApprovalFlow, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockApprovalFlowService) GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockApprovalFlowService) GetByType(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error) {
	if m.GetByTypeFunc != nil {
		return m.GetByTypeFunc(ctx, flowType)
	}
	return nil, nil
}

func (m *MockApprovalFlowService) Update(ctx context.Context, id uuid.UUID, req *model.ApprovalFlowUpdateRequest) (*model.ApprovalFlow, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockApprovalFlowService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockExportService =====

type MockExportService struct {
	ExportAttendanceCSVFunc func(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error)
	ExportLeavesCSVFunc     func(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error)
	ExportOvertimeCSVFunc   func(ctx context.Context, start, end time.Time) ([]byte, error)
	ExportProjectsCSVFunc   func(ctx context.Context, start, end time.Time) ([]byte, error)
}

func (m *MockExportService) ExportAttendanceCSV(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
	if m.ExportAttendanceCSVFunc != nil {
		return m.ExportAttendanceCSVFunc(ctx, userID, start, end)
	}
	return nil, nil
}

func (m *MockExportService) ExportLeavesCSV(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
	if m.ExportLeavesCSVFunc != nil {
		return m.ExportLeavesCSVFunc(ctx, userID, start, end)
	}
	return nil, nil
}

func (m *MockExportService) ExportOvertimeCSV(ctx context.Context, start, end time.Time) ([]byte, error) {
	if m.ExportOvertimeCSVFunc != nil {
		return m.ExportOvertimeCSVFunc(ctx, start, end)
	}
	return nil, nil
}

func (m *MockExportService) ExportProjectsCSV(ctx context.Context, start, end time.Time) ([]byte, error) {
	if m.ExportProjectsCSVFunc != nil {
		return m.ExportProjectsCSVFunc(ctx, start, end)
	}
	return nil, nil
}
