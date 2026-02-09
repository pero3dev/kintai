package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/datatypes"
)

// ===== HREmployeeService =====

type HREmployeeService interface {
	Create(ctx context.Context, req model.HREmployeeCreateRequest) (*model.HREmployee, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HREmployee, error)
	FindAll(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error)
	Update(ctx context.Context, id uuid.UUID, req model.HREmployeeUpdateRequest) (*model.HREmployee, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type hrEmployeeService struct{ deps Deps }

func NewHREmployeeService(deps Deps) HREmployeeService {
	return &hrEmployeeService{deps: deps}
}

func (s *hrEmployeeService) Create(ctx context.Context, req model.HREmployeeCreateRequest) (*model.HREmployee, error) {
	e := &model.HREmployee{
		EmployeeCode:   req.EmployeeCode,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		Position:       req.Position,
		Grade:          req.Grade,
		DepartmentID:   req.DepartmentID,
		ManagerID:      req.ManagerID,
		EmploymentType: model.EmploymentType(req.EmploymentType),
		Address:        req.Address,
		BaseSalary:     req.BaseSalary,
	}
	if req.EmploymentType == "" {
		e.EmploymentType = model.EmploymentTypeFullTime
	}
	if req.HireDate != "" {
		t, _ := time.Parse("2006-01-02", req.HireDate)
		e.HireDate = &t
	}
	if req.BirthDate != "" {
		t, _ := time.Parse("2006-01-02", req.BirthDate)
		e.BirthDate = &t
	}
	if err := s.deps.Repos.HREmployee.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *hrEmployeeService) FindByID(ctx context.Context, id uuid.UUID) (*model.HREmployee, error) {
	return s.deps.Repos.HREmployee.FindByID(ctx, id)
}

func (s *hrEmployeeService) FindAll(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error) {
	return s.deps.Repos.HREmployee.FindAll(ctx, page, pageSize, department, status, employmentType, search)
}

func (s *hrEmployeeService) Update(ctx context.Context, id uuid.UUID, req model.HREmployeeUpdateRequest) (*model.HREmployee, error) {
	e, err := s.deps.Repos.HREmployee.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.FirstName != nil {
		e.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		e.LastName = *req.LastName
	}
	if req.Email != nil {
		e.Email = *req.Email
	}
	if req.Phone != nil {
		e.Phone = *req.Phone
	}
	if req.Position != nil {
		e.Position = *req.Position
	}
	if req.Grade != nil {
		e.Grade = *req.Grade
	}
	if req.DepartmentID != nil {
		e.DepartmentID = req.DepartmentID
	}
	if req.ManagerID != nil {
		e.ManagerID = req.ManagerID
	}
	if req.EmploymentType != nil {
		e.EmploymentType = model.EmploymentType(*req.EmploymentType)
	}
	if req.Status != nil {
		e.Status = model.EmployeeStatus(*req.Status)
	}
	if req.Address != nil {
		e.Address = *req.Address
	}
	if req.BaseSalary != nil {
		e.BaseSalary = *req.BaseSalary
	}
	if err := s.deps.Repos.HREmployee.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *hrEmployeeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.HREmployee.Delete(ctx, id)
}

// ===== HRDepartmentService =====

type HRDepartmentService interface {
	Create(ctx context.Context, req model.HRDepartmentCreateRequest) (*model.HRDepartment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error)
	FindAll(ctx context.Context) ([]model.HRDepartment, error)
	Update(ctx context.Context, id uuid.UUID, req model.HRDepartmentUpdateRequest) (*model.HRDepartment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type hrDepartmentService struct{ deps Deps }

func NewHRDepartmentService(deps Deps) HRDepartmentService {
	return &hrDepartmentService{deps: deps}
}

func (s *hrDepartmentService) Create(ctx context.Context, req model.HRDepartmentCreateRequest) (*model.HRDepartment, error) {
	d := &model.HRDepartment{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		ParentID:    req.ParentID,
		ManagerID:   req.ManagerID,
		Budget:      req.Budget,
	}
	if err := s.deps.Repos.HRDepartment.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *hrDepartmentService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error) {
	return s.deps.Repos.HRDepartment.FindByID(ctx, id)
}

func (s *hrDepartmentService) FindAll(ctx context.Context) ([]model.HRDepartment, error) {
	return s.deps.Repos.HRDepartment.FindAll(ctx)
}

func (s *hrDepartmentService) Update(ctx context.Context, id uuid.UUID, req model.HRDepartmentUpdateRequest) (*model.HRDepartment, error) {
	d, err := s.deps.Repos.HRDepartment.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Code != nil {
		d.Code = *req.Code
	}
	if req.Description != nil {
		d.Description = *req.Description
	}
	if req.ParentID != nil {
		d.ParentID = req.ParentID
	}
	if req.ManagerID != nil {
		d.ManagerID = req.ManagerID
	}
	if req.Budget != nil {
		d.Budget = *req.Budget
	}
	if err := s.deps.Repos.HRDepartment.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *hrDepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.HRDepartment.Delete(ctx, id)
}

// ===== EvaluationService =====

type EvaluationService interface {
	Create(ctx context.Context, req model.EvaluationCreateRequest, reviewerID uuid.UUID) (*model.Evaluation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error)
	FindAll(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error)
	Update(ctx context.Context, id uuid.UUID, req model.EvaluationUpdateRequest) (*model.Evaluation, error)
	Submit(ctx context.Context, id uuid.UUID) (*model.Evaluation, error)
	CreateCycle(ctx context.Context, req model.EvaluationCycleCreateRequest) (*model.EvaluationCycle, error)
	FindAllCycles(ctx context.Context) ([]model.EvaluationCycle, error)
}

type evaluationService struct{ deps Deps }

func NewEvaluationService(deps Deps) EvaluationService {
	return &evaluationService{deps: deps}
}

func (s *evaluationService) Create(ctx context.Context, req model.EvaluationCreateRequest, reviewerID uuid.UUID) (*model.Evaluation, error) {
	e := &model.Evaluation{
		EmployeeID:     req.EmployeeID,
		CycleID:        req.CycleID,
		ReviewerID:     &reviewerID,
		Status:         model.EvaluationStatusDraft,
		SelfScore:      req.SelfScore,
		ManagerScore:   req.ManagerScore,
		SelfComment:    req.SelfComment,
		ManagerComment: req.ManagerComment,
		Goals:          req.Goals,
	}
	if err := s.deps.Repos.Evaluation.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *evaluationService) FindByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	return s.deps.Repos.Evaluation.FindByID(ctx, id)
}

func (s *evaluationService) FindAll(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error) {
	return s.deps.Repos.Evaluation.FindAll(ctx, page, pageSize, cycleID, status)
}

func (s *evaluationService) Update(ctx context.Context, id uuid.UUID, req model.EvaluationUpdateRequest) (*model.Evaluation, error) {
	e, err := s.deps.Repos.Evaluation.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.SelfScore != nil {
		e.SelfScore = req.SelfScore
	}
	if req.ManagerScore != nil {
		e.ManagerScore = req.ManagerScore
	}
	if req.FinalScore != nil {
		e.FinalScore = req.FinalScore
	}
	if req.SelfComment != nil {
		e.SelfComment = *req.SelfComment
	}
	if req.ManagerComment != nil {
		e.ManagerComment = *req.ManagerComment
	}
	if req.Goals != nil {
		e.Goals = *req.Goals
	}
	if err := s.deps.Repos.Evaluation.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *evaluationService) Submit(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	e, err := s.deps.Repos.Evaluation.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	e.Status = model.EvaluationStatusSubmitted
	if err := s.deps.Repos.Evaluation.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *evaluationService) CreateCycle(ctx context.Context, req model.EvaluationCycleCreateRequest) (*model.EvaluationCycle, error) {
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)
	c := &model.EvaluationCycle{
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		IsActive:  true,
	}
	if err := s.deps.Repos.Evaluation.CreateCycle(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *evaluationService) FindAllCycles(ctx context.Context) ([]model.EvaluationCycle, error) {
	return s.deps.Repos.Evaluation.FindAllCycles(ctx)
}

// ===== GoalService =====

type GoalService interface {
	Create(ctx context.Context, req model.HRGoalCreateRequest, userID uuid.UUID) (*model.HRGoal, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRGoal, error)
	FindAll(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error)
	Update(ctx context.Context, id uuid.UUID, req model.HRGoalUpdateRequest) (*model.HRGoal, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress int) (*model.HRGoal, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type goalService struct{ deps Deps }

func NewGoalService(deps Deps) GoalService {
	return &goalService{deps: deps}
}

func (s *goalService) Create(ctx context.Context, req model.HRGoalCreateRequest, userID uuid.UUID) (*model.HRGoal, error) {
	empID := userID
	if req.EmployeeID != nil {
		empID = *req.EmployeeID
	}
	g := &model.HRGoal{
		EmployeeID:  empID,
		Title:       req.Title,
		Description: req.Description,
		Category:    model.GoalCategory(req.Category),
		Status:      model.GoalStatusNotStarted,
		Weight:      req.Weight,
	}
	if g.Category == "" {
		g.Category = model.GoalCategoryPerformance
	}
	if g.Weight == 0 {
		g.Weight = 1
	}
	if req.StartDate != "" {
		t, _ := time.Parse("2006-01-02", req.StartDate)
		g.StartDate = &t
	}
	if req.DueDate != "" {
		t, _ := time.Parse("2006-01-02", req.DueDate)
		g.DueDate = &t
	}
	if err := s.deps.Repos.Goal.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *goalService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRGoal, error) {
	return s.deps.Repos.Goal.FindByID(ctx, id)
}

func (s *goalService) FindAll(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error) {
	return s.deps.Repos.Goal.FindAll(ctx, page, pageSize, status, category, employeeID)
}

func (s *goalService) Update(ctx context.Context, id uuid.UUID, req model.HRGoalUpdateRequest) (*model.HRGoal, error) {
	g, err := s.deps.Repos.Goal.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		g.Title = *req.Title
	}
	if req.Description != nil {
		g.Description = *req.Description
	}
	if req.Category != nil {
		g.Category = model.GoalCategory(*req.Category)
	}
	if req.Status != nil {
		g.Status = model.GoalStatus(*req.Status)
	}
	if req.DueDate != nil {
		t, _ := time.Parse("2006-01-02", *req.DueDate)
		g.DueDate = &t
	}
	if req.Weight != nil {
		g.Weight = *req.Weight
	}
	if err := s.deps.Repos.Goal.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *goalService) UpdateProgress(ctx context.Context, id uuid.UUID, progress int) (*model.HRGoal, error) {
	g, err := s.deps.Repos.Goal.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	g.Progress = progress
	if progress >= 100 {
		g.Status = model.GoalStatusCompleted
	} else if progress > 0 {
		g.Status = model.GoalStatusInProgress
	}
	if err := s.deps.Repos.Goal.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *goalService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Goal.Delete(ctx, id)
}

// ===== TrainingService =====

type TrainingService interface {
	Create(ctx context.Context, req model.TrainingProgramCreateRequest) (*model.TrainingProgram, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error)
	FindAll(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error)
	Update(ctx context.Context, id uuid.UUID, req model.TrainingProgramUpdateRequest) (*model.TrainingProgram, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Enroll(ctx context.Context, programID, employeeID uuid.UUID) error
	Complete(ctx context.Context, programID, employeeID uuid.UUID) error
}

type trainingService struct{ deps Deps }

func NewTrainingService(deps Deps) TrainingService {
	return &trainingService{deps: deps}
}

func (s *trainingService) Create(ctx context.Context, req model.TrainingProgramCreateRequest) (*model.TrainingProgram, error) {
	t := &model.TrainingProgram{
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		InstructorName:  req.InstructorName,
		Status:          model.TrainingStatusScheduled,
		MaxParticipants: req.MaxParticipants,
		Location:        req.Location,
		IsOnline:        req.IsOnline,
	}
	if req.StartDate != "" {
		sd, _ := time.Parse("2006-01-02", req.StartDate)
		t.StartDate = &sd
	}
	if req.EndDate != "" {
		ed, _ := time.Parse("2006-01-02", req.EndDate)
		t.EndDate = &ed
	}
	if err := s.deps.Repos.Training.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *trainingService) FindByID(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error) {
	return s.deps.Repos.Training.FindByID(ctx, id)
}

func (s *trainingService) FindAll(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error) {
	return s.deps.Repos.Training.FindAll(ctx, page, pageSize, category, status)
}

func (s *trainingService) Update(ctx context.Context, id uuid.UUID, req model.TrainingProgramUpdateRequest) (*model.TrainingProgram, error) {
	t, err := s.deps.Repos.Training.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Category != nil {
		t.Category = *req.Category
	}
	if req.InstructorName != nil {
		t.InstructorName = *req.InstructorName
	}
	if req.Status != nil {
		t.Status = model.TrainingStatus(*req.Status)
	}
	if req.MaxParticipants != nil {
		t.MaxParticipants = *req.MaxParticipants
	}
	if req.Location != nil {
		t.Location = *req.Location
	}
	if req.IsOnline != nil {
		t.IsOnline = *req.IsOnline
	}
	if err := s.deps.Repos.Training.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *trainingService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Training.Delete(ctx, id)
}

func (s *trainingService) Enroll(ctx context.Context, programID, employeeID uuid.UUID) error {
	enrollment := &model.TrainingEnrollment{
		ProgramID:  programID,
		EmployeeID: employeeID,
		Status:     "enrolled",
	}
	return s.deps.Repos.Training.CreateEnrollment(ctx, enrollment)
}

func (s *trainingService) Complete(ctx context.Context, programID, employeeID uuid.UUID) error {
	enrollment, err := s.deps.Repos.Training.FindEnrollment(ctx, programID, employeeID)
	if err != nil {
		return err
	}
	now := time.Now()
	enrollment.Status = "completed"
	enrollment.CompletedAt = &now
	return s.deps.Repos.Training.UpdateEnrollment(ctx, enrollment)
}

// ===== RecruitmentService =====

type RecruitmentService interface {
	CreatePosition(ctx context.Context, req model.PositionCreateRequest) (*model.RecruitmentPosition, error)
	FindPositionByID(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error)
	FindAllPositions(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error)
	UpdatePosition(ctx context.Context, id uuid.UUID, req model.PositionUpdateRequest) (*model.RecruitmentPosition, error)
	CreateApplicant(ctx context.Context, req model.ApplicantCreateRequest) (*model.Applicant, error)
	FindAllApplicants(ctx context.Context, positionID, stage string) ([]model.Applicant, error)
	UpdateApplicantStage(ctx context.Context, id uuid.UUID, stage string) (*model.Applicant, error)
}

type recruitmentService struct{ deps Deps }

func NewRecruitmentService(deps Deps) RecruitmentService {
	return &recruitmentService{deps: deps}
}

func (s *recruitmentService) CreatePosition(ctx context.Context, req model.PositionCreateRequest) (*model.RecruitmentPosition, error) {
	p := &model.RecruitmentPosition{
		Title:        req.Title,
		DepartmentID: req.DepartmentID,
		Description:  req.Description,
		Requirements: req.Requirements,
		Status:       model.PositionStatusOpen,
		Openings:     req.Openings,
		Location:     req.Location,
		SalaryMin:    req.SalaryMin,
		SalaryMax:    req.SalaryMax,
	}
	if p.Openings == 0 {
		p.Openings = 1
	}
	if err := s.deps.Repos.Recruitment.CreatePosition(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *recruitmentService) FindPositionByID(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error) {
	return s.deps.Repos.Recruitment.FindPositionByID(ctx, id)
}

func (s *recruitmentService) FindAllPositions(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error) {
	return s.deps.Repos.Recruitment.FindAllPositions(ctx, page, pageSize, status, department)
}

func (s *recruitmentService) UpdatePosition(ctx context.Context, id uuid.UUID, req model.PositionUpdateRequest) (*model.RecruitmentPosition, error) {
	p, err := s.deps.Repos.Recruitment.FindPositionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		p.Title = *req.Title
	}
	if req.DepartmentID != nil {
		p.DepartmentID = req.DepartmentID
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Requirements != nil {
		p.Requirements = *req.Requirements
	}
	if req.Status != nil {
		p.Status = model.PositionStatus(*req.Status)
	}
	if req.Openings != nil {
		p.Openings = *req.Openings
	}
	if req.Location != nil {
		p.Location = *req.Location
	}
	if req.SalaryMin != nil {
		p.SalaryMin = req.SalaryMin
	}
	if req.SalaryMax != nil {
		p.SalaryMax = req.SalaryMax
	}
	if err := s.deps.Repos.Recruitment.UpdatePosition(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *recruitmentService) CreateApplicant(ctx context.Context, req model.ApplicantCreateRequest) (*model.Applicant, error) {
	a := &model.Applicant{
		PositionID: req.PositionID,
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		ResumeURL:  req.ResumeURL,
		Stage:      model.ApplicantStageNew,
		Notes:      req.Notes,
		AppliedAt:  time.Now(),
	}
	if err := s.deps.Repos.Recruitment.CreateApplicant(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *recruitmentService) FindAllApplicants(ctx context.Context, positionID, stage string) ([]model.Applicant, error) {
	return s.deps.Repos.Recruitment.FindAllApplicants(ctx, positionID, stage)
}

func (s *recruitmentService) UpdateApplicantStage(ctx context.Context, id uuid.UUID, stage string) (*model.Applicant, error) {
	a, err := s.deps.Repos.Recruitment.FindApplicantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a.Stage = model.ApplicantStage(stage)
	if err := s.deps.Repos.Recruitment.UpdateApplicant(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// ===== DocumentService =====

type DocumentService interface {
	Upload(ctx context.Context, doc *model.HRDocument) error
	FindAll(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRDocument, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type documentService struct{ deps Deps }

func NewDocumentService(deps Deps) DocumentService {
	return &documentService{deps: deps}
}

func (s *documentService) Upload(ctx context.Context, doc *model.HRDocument) error {
	return s.deps.Repos.Document.Create(ctx, doc)
}

func (s *documentService) FindAll(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error) {
	return s.deps.Repos.Document.FindAll(ctx, page, pageSize, docType, employeeID)
}

func (s *documentService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDocument, error) {
	return s.deps.Repos.Document.FindByID(ctx, id)
}

func (s *documentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Document.Delete(ctx, id)
}

// ===== AnnouncementService =====

type AnnouncementService interface {
	Create(ctx context.Context, req model.AnnouncementCreateRequest, authorID uuid.UUID) (*model.HRAnnouncement, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error)
	FindAll(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error)
	Update(ctx context.Context, id uuid.UUID, req model.AnnouncementUpdateRequest) (*model.HRAnnouncement, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type announcementService struct{ deps Deps }

func NewAnnouncementService(deps Deps) AnnouncementService {
	return &announcementService{deps: deps}
}

func (s *announcementService) Create(ctx context.Context, req model.AnnouncementCreateRequest, authorID uuid.UUID) (*model.HRAnnouncement, error) {
	a := &model.HRAnnouncement{
		Title:    req.Title,
		Content:  req.Content,
		Priority: model.AnnouncementPriority(req.Priority),
		AuthorID: authorID,
	}
	if a.Priority == "" {
		a.Priority = model.AnnouncementPriorityNormal
	}
	if err := s.deps.Repos.Announcement.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *announcementService) FindByID(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error) {
	return s.deps.Repos.Announcement.FindByID(ctx, id)
}

func (s *announcementService) FindAll(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error) {
	return s.deps.Repos.Announcement.FindAll(ctx, page, pageSize, priority)
}

func (s *announcementService) Update(ctx context.Context, id uuid.UUID, req model.AnnouncementUpdateRequest) (*model.HRAnnouncement, error) {
	a, err := s.deps.Repos.Announcement.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Content != nil {
		a.Content = *req.Content
	}
	if req.Priority != nil {
		a.Priority = model.AnnouncementPriority(*req.Priority)
	}
	if req.IsPublished != nil {
		a.IsPublished = *req.IsPublished
		if *req.IsPublished {
			now := time.Now()
			a.PublishedAt = &now
		}
	}
	if err := s.deps.Repos.Announcement.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *announcementService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Announcement.Delete(ctx, id)
}

// ===== HRDashboardService =====

type HRDashboardService interface {
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetRecentActivities(ctx context.Context) ([]map[string]interface{}, error)
}

type hrDashboardService struct{ deps Deps }

func NewHRDashboardService(deps Deps) HRDashboardService {
	return &hrDashboardService{deps: deps}
}

func (s *hrDashboardService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	active, total, _ := s.deps.Repos.HREmployee.CountByStatus(ctx)
	depts, _ := s.deps.Repos.HRDepartment.FindAll(ctx)
	return map[string]interface{}{
		"total_employees":  total,
		"active_employees": active,
		"departments":      len(depts),
		"new_hires":        0,
		"on_leave":         total - active,
	}, nil
}

func (s *hrDashboardService) GetRecentActivities(ctx context.Context) ([]map[string]interface{}, error) {
	activities := []map[string]interface{}{
		{"type": "info", "message": "HR system initialized", "timestamp": time.Now().Format(time.RFC3339)},
	}
	return activities, nil
}

// ===== AttendanceIntegrationService =====

type AttendanceIntegrationService interface {
	GetIntegration(ctx context.Context, period, department string) (map[string]interface{}, error)
	GetAlerts(ctx context.Context) ([]map[string]interface{}, error)
	GetTrend(ctx context.Context, period string) ([]map[string]interface{}, error)
}

type attendanceIntegrationService struct{ deps Deps }

func NewAttendanceIntegrationService(deps Deps) AttendanceIntegrationService {
	return &attendanceIntegrationService{deps: deps}
}

func (s *attendanceIntegrationService) GetIntegration(ctx context.Context, period, department string) (map[string]interface{}, error) {
	active, total, _ := s.deps.Repos.HREmployee.CountByStatus(ctx)
	return map[string]interface{}{
		"total_employees":   total,
		"present_today":     active,
		"absent_today":      total - active,
		"on_leave_today":    0,
		"attendance_rate":   95.5,
		"avg_working_hours": 8.2,
	}, nil
}

func (s *attendanceIntegrationService) GetAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (s *attendanceIntegrationService) GetTrend(ctx context.Context, period string) ([]map[string]interface{}, error) {
	trend := []map[string]interface{}{}
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i)
		trend = append(trend, map[string]interface{}{
			"date":            d.Format("2006-01-02"),
			"attendance_rate": 90 + float64(i%10),
		})
	}
	return trend, nil
}

// ===== OrgChartService =====

type OrgChartService interface {
	GetOrgChart(ctx context.Context) ([]map[string]interface{}, error)
	Simulate(ctx context.Context, data map[string]interface{}) ([]map[string]interface{}, error)
}

type orgChartService struct{ deps Deps }

func NewOrgChartService(deps Deps) OrgChartService {
	return &orgChartService{deps: deps}
}

func (s *orgChartService) GetOrgChart(ctx context.Context) ([]map[string]interface{}, error) {
	depts, err := s.deps.Repos.HRDepartment.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, d := range depts {
		employees, _ := s.deps.Repos.HREmployee.FindByDepartmentID(ctx, d.ID)
		empList := make([]map[string]interface{}, len(employees))
		for i, e := range employees {
			empList[i] = map[string]interface{}{
				"id":       e.ID,
				"name":     e.FirstName + " " + e.LastName,
				"position": e.Position,
			}
		}
		node := map[string]interface{}{
			"id":         d.ID,
			"name":       d.Name,
			"parent_id":  d.ParentID,
			"manager_id": d.ManagerID,
			"employees":  empList,
		}
		result = append(result, node)
	}
	return result, nil
}

func (s *orgChartService) Simulate(ctx context.Context, data map[string]interface{}) ([]map[string]interface{}, error) {
	return s.GetOrgChart(ctx)
}

// ===== OneOnOneService =====

type OneOnOneService interface {
	Create(ctx context.Context, req model.OneOnOneCreateRequest, managerID uuid.UUID) (*model.OneOnOneMeeting, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error)
	FindAll(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error)
	Update(ctx context.Context, id uuid.UUID, req model.OneOnOneUpdateRequest) (*model.OneOnOneMeeting, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddActionItem(ctx context.Context, meetingID uuid.UUID, req model.ActionItemRequest) (*model.OneOnOneMeeting, error)
	ToggleActionItem(ctx context.Context, meetingID uuid.UUID, actionID string) (*model.OneOnOneMeeting, error)
}

type oneOnOneService struct{ deps Deps }

func NewOneOnOneService(deps Deps) OneOnOneService {
	return &oneOnOneService{deps: deps}
}

func (s *oneOnOneService) Create(ctx context.Context, req model.OneOnOneCreateRequest, managerID uuid.UUID) (*model.OneOnOneMeeting, error) {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}
	scheduledDate, _ := time.Parse("2006-01-02T15:04:05Z07:00", req.ScheduledDate)
	if scheduledDate.IsZero() {
		scheduledDate, _ = time.Parse("2006-01-02", req.ScheduledDate)
	}
	freq := req.Frequency
	if freq == "" {
		freq = "biweekly"
	}
	m := &model.OneOnOneMeeting{
		ManagerID:     managerID,
		EmployeeID:    empID,
		ScheduledDate: scheduledDate,
		Status:        "scheduled",
		Frequency:     freq,
		Agenda:        req.Agenda,
		ActionItems:   datatypes.JSON([]byte("[]")),
	}
	if err := s.deps.Repos.OneOnOne.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *oneOnOneService) FindByID(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error) {
	return s.deps.Repos.OneOnOne.FindByID(ctx, id)
}

func (s *oneOnOneService) FindAll(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error) {
	return s.deps.Repos.OneOnOne.FindAll(ctx, status, employeeID)
}

func (s *oneOnOneService) Update(ctx context.Context, id uuid.UUID, req model.OneOnOneUpdateRequest) (*model.OneOnOneMeeting, error) {
	m, err := s.deps.Repos.OneOnOne.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Status != "" {
		m.Status = req.Status
	}
	if req.Agenda != nil {
		m.Agenda = *req.Agenda
	}
	if req.Notes != nil {
		m.Notes = *req.Notes
	}
	if req.Mood != nil {
		m.Mood = *req.Mood
	}
	if err := s.deps.Repos.OneOnOne.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *oneOnOneService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.OneOnOne.Delete(ctx, id)
}

func (s *oneOnOneService) AddActionItem(ctx context.Context, meetingID uuid.UUID, req model.ActionItemRequest) (*model.OneOnOneMeeting, error) {
	m, err := s.deps.Repos.OneOnOne.FindByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	var items []map[string]interface{}
	if m.ActionItems != nil {
		json.Unmarshal([]byte(m.ActionItems), &items)
	}
	newItem := map[string]interface{}{
		"id":        uuid.New().String(),
		"title":     req.Title,
		"completed": false,
	}
	items = append(items, newItem)
	b, _ := json.Marshal(items)
	m.ActionItems = datatypes.JSON(b)
	if err := s.deps.Repos.OneOnOne.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *oneOnOneService) ToggleActionItem(ctx context.Context, meetingID uuid.UUID, actionID string) (*model.OneOnOneMeeting, error) {
	m, err := s.deps.Repos.OneOnOne.FindByID(ctx, meetingID)
	if err != nil {
		return nil, err
	}
	var items []map[string]interface{}
	if m.ActionItems != nil {
		json.Unmarshal([]byte(m.ActionItems), &items)
	}
	for i, item := range items {
		if id, ok := item["id"].(string); ok && id == actionID {
			if completed, ok := item["completed"].(bool); ok {
				items[i]["completed"] = !completed
			}
		}
	}
	b, _ := json.Marshal(items)
	m.ActionItems = datatypes.JSON(b)
	if err := s.deps.Repos.OneOnOne.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// ===== SkillService =====

type SkillService interface {
	GetSkillMap(ctx context.Context, department, employeeID string) ([]model.EmployeeSkill, error)
	GetGapAnalysis(ctx context.Context, department string) ([]map[string]interface{}, error)
	AddSkill(ctx context.Context, employeeID uuid.UUID, req model.SkillAddRequest) (*model.EmployeeSkill, error)
	UpdateSkill(ctx context.Context, skillID uuid.UUID, req model.SkillUpdateRequest) (*model.EmployeeSkill, error)
}

type skillService struct{ deps Deps }

func NewSkillService(deps Deps) SkillService {
	return &skillService{deps: deps}
}

func (s *skillService) GetSkillMap(ctx context.Context, department, employeeID string) ([]model.EmployeeSkill, error) {
	if employeeID != "" {
		empID, err := uuid.Parse(employeeID)
		if err != nil {
			return nil, err
		}
		return s.deps.Repos.Skill.FindByEmployeeID(ctx, empID)
	}
	return s.deps.Repos.Skill.FindAll(ctx, department)
}

func (s *skillService) GetGapAnalysis(ctx context.Context, department string) ([]map[string]interface{}, error) {
	return s.deps.Repos.Skill.GetGapAnalysis(ctx, department)
}

func (s *skillService) AddSkill(ctx context.Context, employeeID uuid.UUID, req model.SkillAddRequest) (*model.EmployeeSkill, error) {
	skill := &model.EmployeeSkill{
		EmployeeID: employeeID,
		SkillName:  req.SkillName,
		Category:   req.Category,
		Level:      req.Level,
	}
	if skill.Category == "" {
		skill.Category = "technical"
	}
	if skill.Level == 0 {
		skill.Level = 1
	}
	if err := s.deps.Repos.Skill.Create(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func (s *skillService) UpdateSkill(ctx context.Context, skillID uuid.UUID, req model.SkillUpdateRequest) (*model.EmployeeSkill, error) {
	skill, err := s.deps.Repos.Skill.FindByID(ctx, skillID)
	if err != nil {
		return nil, err
	}
	if req.Level != nil {
		skill.Level = *req.Level
	}
	if req.Category != nil {
		skill.Category = *req.Category
	}
	if err := s.deps.Repos.Skill.Update(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

// ===== SalaryService =====

type SalaryService interface {
	GetOverview(ctx context.Context, department string) (map[string]interface{}, error)
	Simulate(ctx context.Context, req model.SalarySimulateRequest) (map[string]interface{}, error)
	GetHistory(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error)
	GetBudget(ctx context.Context, department string) (map[string]interface{}, error)
}

type salaryService struct{ deps Deps }

func NewSalaryService(deps Deps) SalaryService {
	return &salaryService{deps: deps}
}

func (s *salaryService) GetOverview(ctx context.Context, department string) (map[string]interface{}, error) {
	return s.deps.Repos.Salary.GetOverview(ctx, department)
}

func (s *salaryService) Simulate(ctx context.Context, req model.SalarySimulateRequest) (map[string]interface{}, error) {
	baseSalary := 300000.0
	if req.Grade != "" {
		switch req.Grade {
		case "S1", "S2":
			baseSalary = 250000
		case "M1":
			baseSalary = 400000
		case "M2":
			baseSalary = 500000
		case "L1":
			baseSalary = 600000
		case "L2":
			baseSalary = 750000
		default:
			baseSalary = 350000
		}
	}
	years, _ := strconv.ParseFloat(req.YearsOfService, 64)
	seniority := baseSalary * 0.02 * years

	evalBonus := 0.0
	if req.EvaluationScore != "" {
		score, _ := strconv.ParseFloat(req.EvaluationScore, 64)
		evalBonus = baseSalary * (score / 100) * 0.3
	}

	projected := baseSalary + seniority + evalBonus
	return map[string]interface{}{
		"base_salary":      baseSalary,
		"seniority_bonus":  math.Round(seniority*100) / 100,
		"evaluation_bonus": math.Round(evalBonus*100) / 100,
		"projected_salary": math.Round(projected*100) / 100,
		"annual_salary":    math.Round(projected*12*100) / 100,
	}, nil
}

func (s *salaryService) GetHistory(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error) {
	return s.deps.Repos.Salary.FindByEmployeeID(ctx, employeeID)
}

func (s *salaryService) GetBudget(ctx context.Context, department string) (map[string]interface{}, error) {
	overview, _ := s.deps.Repos.Salary.GetOverview(ctx, department)
	totalPayroll, _ := overview["total_payroll"].(float64)
	headcount, _ := overview["headcount"].(int64)
	return map[string]interface{}{
		"total_budget": totalPayroll * 12 * 1.3,
		"used_budget":  totalPayroll * 12,
		"remaining":    totalPayroll * 12 * 0.3,
		"utilization":  76.9,
		"headcount":    headcount,
		"avg_cost":     totalPayroll / float64(max(headcount, 1)),
	}, nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ===== OnboardingService =====

type OnboardingService interface {
	Create(ctx context.Context, req model.OnboardingCreateRequest) (*model.Onboarding, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Onboarding, error)
	FindAll(ctx context.Context, status string) ([]model.Onboarding, error)
	Update(ctx context.Context, id uuid.UUID, data map[string]interface{}) (*model.Onboarding, error)
	ToggleTask(ctx context.Context, id uuid.UUID, taskID string) (*model.Onboarding, error)
	CreateTemplate(ctx context.Context, req model.OnboardingTemplateCreateRequest) (*model.OnboardingTemplate, error)
	FindAllTemplates(ctx context.Context) ([]model.OnboardingTemplate, error)
}

type onboardingService struct{ deps Deps }

func NewOnboardingService(deps Deps) OnboardingService {
	return &onboardingService{deps: deps}
}

func (s *onboardingService) Create(ctx context.Context, req model.OnboardingCreateRequest) (*model.Onboarding, error) {
	empID, _ := uuid.Parse(req.EmployeeID)
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	o := &model.Onboarding{
		EmployeeID: empID,
		Status:     model.OnboardingStatusPending,
		StartDate:  startDate,
		Tasks:      datatypes.JSON([]byte("[]")),
	}
	if req.TemplateID != "" {
		tid, _ := uuid.Parse(req.TemplateID)
		o.TemplateID = &tid
	}
	if req.MentorID != "" {
		mid, _ := uuid.Parse(req.MentorID)
		o.MentorID = &mid
	}
	if err := s.deps.Repos.Onboarding.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *onboardingService) FindByID(ctx context.Context, id uuid.UUID) (*model.Onboarding, error) {
	return s.deps.Repos.Onboarding.FindByID(ctx, id)
}

func (s *onboardingService) FindAll(ctx context.Context, status string) ([]model.Onboarding, error) {
	return s.deps.Repos.Onboarding.FindAll(ctx, status)
}

func (s *onboardingService) Update(ctx context.Context, id uuid.UUID, data map[string]interface{}) (*model.Onboarding, error) {
	o, err := s.deps.Repos.Onboarding.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if status, ok := data["status"].(string); ok {
		o.Status = model.OnboardingStatus(status)
	}
	if err := s.deps.Repos.Onboarding.Update(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *onboardingService) ToggleTask(ctx context.Context, id uuid.UUID, taskID string) (*model.Onboarding, error) {
	o, err := s.deps.Repos.Onboarding.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var tasks []map[string]interface{}
	if o.Tasks != nil {
		json.Unmarshal([]byte(o.Tasks), &tasks)
	}
	for i, task := range tasks {
		if tid, ok := task["id"].(string); ok && tid == taskID {
			if completed, ok := task["completed"].(bool); ok {
				tasks[i]["completed"] = !completed
			}
		}
	}
	b, _ := json.Marshal(tasks)
	o.Tasks = datatypes.JSON(b)
	if err := s.deps.Repos.Onboarding.Update(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *onboardingService) CreateTemplate(ctx context.Context, req model.OnboardingTemplateCreateRequest) (*model.OnboardingTemplate, error) {
	tasksJSON, _ := json.Marshal(req.Tasks)
	t := &model.OnboardingTemplate{
		Name:        req.Name,
		Description: req.Description,
		Tasks:       datatypes.JSON(tasksJSON),
	}
	if err := s.deps.Repos.Onboarding.CreateTemplate(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *onboardingService) FindAllTemplates(ctx context.Context) ([]model.OnboardingTemplate, error) {
	return s.deps.Repos.Onboarding.FindAllTemplates(ctx)
}

// ===== OffboardingService =====

type OffboardingService interface {
	Create(ctx context.Context, req model.OffboardingCreateRequest) (*model.Offboarding, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Offboarding, error)
	FindAll(ctx context.Context, status string) ([]model.Offboarding, error)
	Update(ctx context.Context, id uuid.UUID, req model.OffboardingUpdateRequest) (*model.Offboarding, error)
	ToggleChecklist(ctx context.Context, id uuid.UUID, itemKey string) (*model.Offboarding, error)
	GetAnalytics(ctx context.Context) (map[string]interface{}, error)
}

type offboardingService struct{ deps Deps }

func NewOffboardingService(deps Deps) OffboardingService {
	return &offboardingService{deps: deps}
}

func (s *offboardingService) Create(ctx context.Context, req model.OffboardingCreateRequest) (*model.Offboarding, error) {
	empID, _ := uuid.Parse(req.EmployeeID)
	lastDate, _ := time.Parse("2006-01-02", req.LastWorkingDate)
	o := &model.Offboarding{
		EmployeeID:      empID,
		Reason:          req.Reason,
		Status:          model.OffboardingStatusPending,
		LastWorkingDate: lastDate,
		Notes:           req.Notes,
		Checklist:       datatypes.JSON([]byte("[]")),
	}
	if err := s.deps.Repos.Offboarding.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *offboardingService) FindByID(ctx context.Context, id uuid.UUID) (*model.Offboarding, error) {
	return s.deps.Repos.Offboarding.FindByID(ctx, id)
}

func (s *offboardingService) FindAll(ctx context.Context, status string) ([]model.Offboarding, error) {
	return s.deps.Repos.Offboarding.FindAll(ctx, status)
}

func (s *offboardingService) Update(ctx context.Context, id uuid.UUID, req model.OffboardingUpdateRequest) (*model.Offboarding, error) {
	o, err := s.deps.Repos.Offboarding.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Status != nil {
		o.Status = model.OffboardingStatus(*req.Status)
	}
	if req.Notes != nil {
		o.Notes = *req.Notes
	}
	if req.ExitInterview != nil {
		o.ExitInterview = *req.ExitInterview
	}
	if err := s.deps.Repos.Offboarding.Update(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *offboardingService) ToggleChecklist(ctx context.Context, id uuid.UUID, itemKey string) (*model.Offboarding, error) {
	o, err := s.deps.Repos.Offboarding.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	var checklist []map[string]interface{}
	if o.Checklist != nil {
		json.Unmarshal([]byte(o.Checklist), &checklist)
	}
	for i, item := range checklist {
		if key, ok := item["key"].(string); ok && key == itemKey {
			if completed, ok := item["completed"].(bool); ok {
				checklist[i]["completed"] = !completed
			}
		}
	}
	b, _ := json.Marshal(checklist)
	o.Checklist = datatypes.JSON(b)
	if err := s.deps.Repos.Offboarding.Update(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *offboardingService) GetAnalytics(ctx context.Context) (map[string]interface{}, error) {
	return s.deps.Repos.Offboarding.GetTurnoverAnalytics(ctx)
}

// ===== SurveyService =====

type SurveyService interface {
	Create(ctx context.Context, req model.SurveyCreateRequest, createdBy uuid.UUID) (*model.Survey, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	FindAll(ctx context.Context, status, surveyType string) ([]model.Survey, error)
	Update(ctx context.Context, id uuid.UUID, req model.SurveyUpdateRequest) (*model.Survey, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Publish(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	Close(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	GetResults(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	SubmitResponse(ctx context.Context, surveyID uuid.UUID, employeeID *uuid.UUID, req model.SurveyResponseRequest) error
}

type surveyService struct{ deps Deps }

func NewSurveyService(deps Deps) SurveyService {
	return &surveyService{deps: deps}
}

func (s *surveyService) Create(ctx context.Context, req model.SurveyCreateRequest, createdBy uuid.UUID) (*model.Survey, error) {
	questionsJSON, _ := json.Marshal(req.Questions)
	survey := &model.Survey{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      model.SurveyStatusDraft,
		IsAnonymous: req.IsAnonymous,
		Questions:   datatypes.JSON(questionsJSON),
		CreatedBy:   createdBy,
	}
	if survey.Type == "" {
		survey.Type = "engagement"
	}
	if err := s.deps.Repos.Survey.Create(ctx, survey); err != nil {
		return nil, err
	}
	return survey, nil
}

func (s *surveyService) FindByID(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	return s.deps.Repos.Survey.FindByID(ctx, id)
}

func (s *surveyService) FindAll(ctx context.Context, status, surveyType string) ([]model.Survey, error) {
	return s.deps.Repos.Survey.FindAll(ctx, status, surveyType)
}

func (s *surveyService) Update(ctx context.Context, id uuid.UUID, req model.SurveyUpdateRequest) (*model.Survey, error) {
	survey, err := s.deps.Repos.Survey.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		survey.Title = *req.Title
	}
	if req.Description != nil {
		survey.Description = *req.Description
	}
	if req.Type != nil {
		survey.Type = *req.Type
	}
	if req.IsAnonymous != nil {
		survey.IsAnonymous = *req.IsAnonymous
	}
	if req.Questions != nil {
		questionsJSON, _ := json.Marshal(req.Questions)
		survey.Questions = datatypes.JSON(questionsJSON)
	}
	if err := s.deps.Repos.Survey.Update(ctx, survey); err != nil {
		return nil, err
	}
	return survey, nil
}

func (s *surveyService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Survey.Delete(ctx, id)
}

func (s *surveyService) Publish(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	survey, err := s.deps.Repos.Survey.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	survey.Status = model.SurveyStatusActive
	survey.PublishedAt = &now
	if err := s.deps.Repos.Survey.Update(ctx, survey); err != nil {
		return nil, err
	}
	return survey, nil
}

func (s *surveyService) Close(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	survey, err := s.deps.Repos.Survey.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	survey.Status = model.SurveyStatusClosed
	survey.ClosedAt = &now
	if err := s.deps.Repos.Survey.Update(ctx, survey); err != nil {
		return nil, err
	}
	return survey, nil
}

func (s *surveyService) GetResults(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	survey, err := s.deps.Repos.Survey.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	responses, err := s.deps.Repos.Survey.FindResponsesBySurveyID(ctx, id)
	if err != nil {
		return nil, err
	}
	count, _ := s.deps.Repos.Survey.CountResponsesBySurveyID(ctx, id)
	return map[string]interface{}{
		"survey":          survey,
		"total_responses": count,
		"responses":       responses,
	}, nil
}

func (s *surveyService) SubmitResponse(ctx context.Context, surveyID uuid.UUID, employeeID *uuid.UUID, req model.SurveyResponseRequest) error {
	answersJSON, _ := json.Marshal(req.Answers)
	resp := &model.SurveyResponse{
		SurveyID:   surveyID,
		EmployeeID: employeeID,
		Answers:    datatypes.JSON(answersJSON),
	}
	return s.deps.Repos.Survey.CreateResponse(ctx, resp)
}
