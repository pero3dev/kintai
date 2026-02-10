package hr

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/gorm"
)

// ===== HREmployeeRepository =====

type HREmployeeRepository interface {
	Create(ctx context.Context, e *model.HREmployee) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.HREmployee, error)
	FindAll(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error)
	Update(ctx context.Context, e *model.HREmployee) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByDepartmentID(ctx context.Context, deptID uuid.UUID) ([]model.HREmployee, error)
	CountByStatus(ctx context.Context) (active int64, total int64, err error)
}

type hrEmployeeRepository struct{ db *gorm.DB }

func NewHREmployeeRepository(db *gorm.DB) HREmployeeRepository {
	return &hrEmployeeRepository{db: db}
}

func (r *hrEmployeeRepository) Create(ctx context.Context, e *model.HREmployee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *hrEmployeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HREmployee, error) {
	var e model.HREmployee
	err := r.db.WithContext(ctx).Preload("Department").Preload("Manager").First(&e, "id = ?", id).Error
	return &e, err
}

func (r *hrEmployeeRepository) FindAll(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error) {
	var list []model.HREmployee
	var total int64
	q := r.db.WithContext(ctx).Model(&model.HREmployee{})
	if department != "" {
		q = q.Joins("JOIN hr_departments ON hr_departments.id = hr_employees.department_id").Where("hr_departments.name ILIKE ?", "%"+department+"%")
	}
	if status != "" {
		q = q.Where("hr_employees.status = ?", status)
	}
	if employmentType != "" {
		q = q.Where("hr_employees.employment_type = ?", employmentType)
	}
	if search != "" {
		q = q.Where("hr_employees.first_name ILIKE ? OR hr_employees.last_name ILIKE ? OR hr_employees.email ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Department").Preload("Manager").Offset(offset).Limit(pageSize).Order("hr_employees.created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *hrEmployeeRepository) Update(ctx context.Context, e *model.HREmployee) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *hrEmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.HREmployee{}, "id = ?", id).Error
}

func (r *hrEmployeeRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID) ([]model.HREmployee, error) {
	var list []model.HREmployee
	err := r.db.WithContext(ctx).Where("department_id = ?", deptID).Find(&list).Error
	return list, err
}

func (r *hrEmployeeRepository) CountByStatus(ctx context.Context) (int64, int64, error) {
	var active, total int64
	r.db.WithContext(ctx).Model(&model.HREmployee{}).Count(&total)
	r.db.WithContext(ctx).Model(&model.HREmployee{}).Where("status = 'active'").Count(&active)
	return active, total, nil
}

// ===== HRDepartmentRepository =====

type HRDepartmentRepository interface {
	Create(ctx context.Context, d *model.HRDepartment) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error)
	FindAll(ctx context.Context) ([]model.HRDepartment, error)
	Update(ctx context.Context, d *model.HRDepartment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type hrDepartmentRepository struct{ db *gorm.DB }

func NewHRDepartmentRepository(db *gorm.DB) HRDepartmentRepository {
	return &hrDepartmentRepository{db: db}
}

func (r *hrDepartmentRepository) Create(ctx context.Context, d *model.HRDepartment) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *hrDepartmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error) {
	var d model.HRDepartment
	err := r.db.WithContext(ctx).Preload("Manager").Preload("Children").First(&d, "id = ?", id).Error
	return &d, err
}

func (r *hrDepartmentRepository) FindAll(ctx context.Context) ([]model.HRDepartment, error) {
	var list []model.HRDepartment
	err := r.db.WithContext(ctx).Preload("Manager").Preload("Children").Order("name ASC").Find(&list).Error
	return list, err
}

func (r *hrDepartmentRepository) Update(ctx context.Context, d *model.HRDepartment) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *hrDepartmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.HRDepartment{}, "id = ?", id).Error
}

// ===== EvaluationRepository =====

type EvaluationRepository interface {
	Create(ctx context.Context, e *model.Evaluation) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error)
	FindAll(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error)
	Update(ctx context.Context, e *model.Evaluation) error
	CreateCycle(ctx context.Context, c *model.EvaluationCycle) error
	FindAllCycles(ctx context.Context) ([]model.EvaluationCycle, error)
}

type evaluationRepository struct{ db *gorm.DB }

func NewEvaluationRepository(db *gorm.DB) EvaluationRepository {
	return &evaluationRepository{db: db}
}

func (r *evaluationRepository) Create(ctx context.Context, e *model.Evaluation) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *evaluationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	var e model.Evaluation
	err := r.db.WithContext(ctx).Preload("Employee").Preload("Cycle").Preload("Reviewer").First(&e, "id = ?", id).Error
	return &e, err
}

func (r *evaluationRepository) FindAll(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error) {
	var list []model.Evaluation
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Evaluation{})
	if cycleID != "" {
		q = q.Where("cycle_id = ?", cycleID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Employee").Preload("Cycle").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *evaluationRepository) Update(ctx context.Context, e *model.Evaluation) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *evaluationRepository) CreateCycle(ctx context.Context, c *model.EvaluationCycle) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *evaluationRepository) FindAllCycles(ctx context.Context) ([]model.EvaluationCycle, error) {
	var list []model.EvaluationCycle
	err := r.db.WithContext(ctx).Order("start_date DESC").Find(&list).Error
	return list, err
}

// ===== GoalRepository =====

type GoalRepository interface {
	Create(ctx context.Context, g *model.HRGoal) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRGoal, error)
	FindAll(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error)
	Update(ctx context.Context, g *model.HRGoal) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type goalRepository struct{ db *gorm.DB }

func NewGoalRepository(db *gorm.DB) GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) Create(ctx context.Context, g *model.HRGoal) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *goalRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HRGoal, error) {
	var g model.HRGoal
	err := r.db.WithContext(ctx).Preload("Employee").First(&g, "id = ?", id).Error
	return &g, err
}

func (r *goalRepository) FindAll(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error) {
	var list []model.HRGoal
	var total int64
	q := r.db.WithContext(ctx).Model(&model.HRGoal{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Employee").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *goalRepository) Update(ctx context.Context, g *model.HRGoal) error {
	return r.db.WithContext(ctx).Save(g).Error
}

func (r *goalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.HRGoal{}, "id = ?", id).Error
}

// ===== TrainingRepository =====

type TrainingRepository interface {
	Create(ctx context.Context, t *model.TrainingProgram) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error)
	FindAll(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error)
	Update(ctx context.Context, t *model.TrainingProgram) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateEnrollment(ctx context.Context, e *model.TrainingEnrollment) error
	UpdateEnrollment(ctx context.Context, e *model.TrainingEnrollment) error
	FindEnrollment(ctx context.Context, programID, employeeID uuid.UUID) (*model.TrainingEnrollment, error)
}

type trainingRepository struct{ db *gorm.DB }

func NewTrainingRepository(db *gorm.DB) TrainingRepository {
	return &trainingRepository{db: db}
}

func (r *trainingRepository) Create(ctx context.Context, t *model.TrainingProgram) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *trainingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error) {
	var t model.TrainingProgram
	err := r.db.WithContext(ctx).Preload("Enrollments").Preload("Enrollments.Employee").First(&t, "id = ?", id).Error
	return &t, err
}

func (r *trainingRepository) FindAll(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error) {
	var list []model.TrainingProgram
	var total int64
	q := r.db.WithContext(ctx).Model(&model.TrainingProgram{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Enrollments").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *trainingRepository) Update(ctx context.Context, t *model.TrainingProgram) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *trainingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.TrainingProgram{}, "id = ?", id).Error
}

func (r *trainingRepository) CreateEnrollment(ctx context.Context, e *model.TrainingEnrollment) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *trainingRepository) UpdateEnrollment(ctx context.Context, e *model.TrainingEnrollment) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *trainingRepository) FindEnrollment(ctx context.Context, programID, employeeID uuid.UUID) (*model.TrainingEnrollment, error) {
	var e model.TrainingEnrollment
	err := r.db.WithContext(ctx).Where("program_id = ? AND employee_id = ?", programID, employeeID).First(&e).Error
	return &e, err
}

// ===== RecruitmentRepository =====

type RecruitmentRepository interface {
	CreatePosition(ctx context.Context, p *model.RecruitmentPosition) error
	FindPositionByID(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error)
	FindAllPositions(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error)
	UpdatePosition(ctx context.Context, p *model.RecruitmentPosition) error
	CreateApplicant(ctx context.Context, a *model.Applicant) error
	FindAllApplicants(ctx context.Context, positionID, stage string) ([]model.Applicant, error)
	FindApplicantByID(ctx context.Context, id uuid.UUID) (*model.Applicant, error)
	UpdateApplicant(ctx context.Context, a *model.Applicant) error
}

type recruitmentRepository struct{ db *gorm.DB }

func NewRecruitmentRepository(db *gorm.DB) RecruitmentRepository {
	return &recruitmentRepository{db: db}
}

func (r *recruitmentRepository) CreatePosition(ctx context.Context, p *model.RecruitmentPosition) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *recruitmentRepository) FindPositionByID(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error) {
	var p model.RecruitmentPosition
	err := r.db.WithContext(ctx).Preload("Department").Preload("Applicants").First(&p, "id = ?", id).Error
	return &p, err
}

func (r *recruitmentRepository) FindAllPositions(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error) {
	var list []model.RecruitmentPosition
	var total int64
	q := r.db.WithContext(ctx).Model(&model.RecruitmentPosition{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if department != "" {
		q = q.Joins("JOIN hr_departments ON hr_departments.id = recruitment_positions.department_id").Where("hr_departments.name ILIKE ?", "%"+department+"%")
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Department").Preload("Applicants").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *recruitmentRepository) UpdatePosition(ctx context.Context, p *model.RecruitmentPosition) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *recruitmentRepository) CreateApplicant(ctx context.Context, a *model.Applicant) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *recruitmentRepository) FindAllApplicants(ctx context.Context, positionID, stage string) ([]model.Applicant, error) {
	var list []model.Applicant
	q := r.db.WithContext(ctx)
	if positionID != "" {
		q = q.Where("position_id = ?", positionID)
	}
	if stage != "" {
		q = q.Where("stage = ?", stage)
	}
	err := q.Order("applied_at DESC").Find(&list).Error
	return list, err
}

func (r *recruitmentRepository) FindApplicantByID(ctx context.Context, id uuid.UUID) (*model.Applicant, error) {
	var a model.Applicant
	err := r.db.WithContext(ctx).Preload("Position").First(&a, "id = ?", id).Error
	return &a, err
}

func (r *recruitmentRepository) UpdateApplicant(ctx context.Context, a *model.Applicant) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// ===== DocumentRepository =====

type DocumentRepository interface {
	Create(ctx context.Context, d *model.HRDocument) error
	FindAll(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRDocument, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type documentRepository struct{ db *gorm.DB }

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, d *model.HRDocument) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *documentRepository) FindAll(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error) {
	var list []model.HRDocument
	var total int64
	q := r.db.WithContext(ctx).Model(&model.HRDocument{})
	if docType != "" {
		q = q.Where("type = ?", docType)
	}
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Employee").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *documentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDocument, error) {
	var d model.HRDocument
	err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error
	return &d, err
}

func (r *documentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.HRDocument{}, "id = ?", id).Error
}

// ===== AnnouncementRepository =====

type AnnouncementRepository interface {
	Create(ctx context.Context, a *model.HRAnnouncement) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error)
	FindAll(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error)
	Update(ctx context.Context, a *model.HRAnnouncement) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type announcementRepository struct{ db *gorm.DB }

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return &announcementRepository{db: db}
}

func (r *announcementRepository) Create(ctx context.Context, a *model.HRAnnouncement) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *announcementRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error) {
	var a model.HRAnnouncement
	err := r.db.WithContext(ctx).Preload("Author").First(&a, "id = ?", id).Error
	return &a, err
}

func (r *announcementRepository) FindAll(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error) {
	var list []model.HRAnnouncement
	var total int64
	q := r.db.WithContext(ctx).Model(&model.HRAnnouncement{})
	if priority != "" {
		q = q.Where("priority = ?", priority)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Preload("Author").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *announcementRepository) Update(ctx context.Context, a *model.HRAnnouncement) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *announcementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.HRAnnouncement{}, "id = ?", id).Error
}

// ===== OneOnOneRepository =====

type OneOnOneRepository interface {
	Create(ctx context.Context, m *model.OneOnOneMeeting) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error)
	FindAll(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error)
	Update(ctx context.Context, m *model.OneOnOneMeeting) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type oneOnOneRepository struct{ db *gorm.DB }

func NewOneOnOneRepository(db *gorm.DB) OneOnOneRepository {
	return &oneOnOneRepository{db: db}
}

func (r *oneOnOneRepository) Create(ctx context.Context, m *model.OneOnOneMeeting) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *oneOnOneRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error) {
	var m model.OneOnOneMeeting
	err := r.db.WithContext(ctx).Preload("Manager").Preload("Employee").First(&m, "id = ?", id).Error
	return &m, err
}

func (r *oneOnOneRepository) FindAll(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error) {
	var list []model.OneOnOneMeeting
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if employeeID != "" {
		q = q.Where("employee_id = ? OR manager_id = ?", employeeID, employeeID)
	}
	err := q.Preload("Manager").Preload("Employee").Order("scheduled_date DESC").Find(&list).Error
	return list, err
}

func (r *oneOnOneRepository) Update(ctx context.Context, m *model.OneOnOneMeeting) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *oneOnOneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.OneOnOneMeeting{}, "id = ?", id).Error
}

// ===== SkillRepository =====

type SkillRepository interface {
	Create(ctx context.Context, s *model.EmployeeSkill) error
	FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]model.EmployeeSkill, error)
	FindAll(ctx context.Context, department string) ([]model.EmployeeSkill, error)
	Update(ctx context.Context, s *model.EmployeeSkill) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeSkill, error)
	GetGapAnalysis(ctx context.Context, department string) ([]map[string]interface{}, error)
}

type skillRepository struct{ db *gorm.DB }

func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepository{db: db}
}

func (r *skillRepository) Create(ctx context.Context, s *model.EmployeeSkill) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *skillRepository) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]model.EmployeeSkill, error) {
	var list []model.EmployeeSkill
	err := r.db.WithContext(ctx).Where("employee_id = ?", employeeID).Order("category, skill_name").Find(&list).Error
	return list, err
}

func (r *skillRepository) FindAll(ctx context.Context, department string) ([]model.EmployeeSkill, error) {
	var list []model.EmployeeSkill
	q := r.db.WithContext(ctx)
	if department != "" {
		q = q.Joins("JOIN hr_employees ON hr_employees.id = employee_skills.employee_id").
			Joins("JOIN hr_departments ON hr_departments.id = hr_employees.department_id").
			Where("hr_departments.name ILIKE ?", "%"+department+"%")
	}
	err := q.Preload("Employee").Order("employee_id, category, skill_name").Find(&list).Error
	return list, err
}

func (r *skillRepository) Update(ctx context.Context, s *model.EmployeeSkill) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *skillRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeSkill, error) {
	var s model.EmployeeSkill
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *skillRepository) GetGapAnalysis(ctx context.Context, department string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	q := r.db.WithContext(ctx).Model(&model.EmployeeSkill{}).
		Select("skill_name, category, AVG(level) as current_avg, COUNT(*) as employee_count").
		Group("skill_name, category").
		Order("skill_name")
	if department != "" {
		q = q.Joins("JOIN hr_employees ON hr_employees.id = employee_skills.employee_id").
			Joins("JOIN hr_departments ON hr_departments.id = hr_employees.department_id").
			Where("hr_departments.name ILIKE ?", "%"+department+"%")
	}
	rows, err := q.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, cat string
		var avg float64
		var count int
		if err := rows.Scan(&name, &cat, &avg, &count); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"skill_name":     name,
			"category":       cat,
			"current_avg":    avg,
			"required_level": 4,
			"gap":            avg - 4,
			"employee_count": count,
		})
	}
	return results, nil
}

// ===== SalaryRepository =====

type SalaryRepository interface {
	FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error)
	Create(ctx context.Context, s *model.SalaryRecord) error
	GetOverview(ctx context.Context, department string) (map[string]interface{}, error)
}

type salaryRepository struct{ db *gorm.DB }

func NewSalaryRepository(db *gorm.DB) SalaryRepository {
	return &salaryRepository{db: db}
}

func (r *salaryRepository) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error) {
	var list []model.SalaryRecord
	err := r.db.WithContext(ctx).Where("employee_id = ?", employeeID).Order("effective_date DESC").Find(&list).Error
	return list, err
}

func (r *salaryRepository) Create(ctx context.Context, s *model.SalaryRecord) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *salaryRepository) GetOverview(ctx context.Context, department string) (map[string]interface{}, error) {
	q := r.db.WithContext(ctx).Model(&model.HREmployee{}).Where("status = 'active'")
	if department != "" {
		q = q.Joins("JOIN hr_departments ON hr_departments.id = hr_employees.department_id").
			Where("hr_departments.name ILIKE ?", "%"+department+"%")
	}
	var result struct {
		AvgSalary    float64 `gorm:"column:avg_salary"`
		MedianSalary float64 `gorm:"column:median_salary"`
		TotalPayroll float64 `gorm:"column:total_payroll"`
		Headcount    int64   `gorm:"column:headcount"`
	}
	q.Select("AVG(base_salary) as avg_salary, PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY base_salary) as median_salary, SUM(base_salary) as total_payroll, COUNT(*) as headcount").Scan(&result)
	return map[string]interface{}{
		"avg_salary":    result.AvgSalary,
		"median_salary": result.MedianSalary,
		"total_payroll": result.TotalPayroll,
		"headcount":     result.Headcount,
	}, nil
}

// ===== OnboardingRepository =====

type OnboardingRepository interface {
	Create(ctx context.Context, o *model.Onboarding) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Onboarding, error)
	FindAll(ctx context.Context, status string) ([]model.Onboarding, error)
	Update(ctx context.Context, o *model.Onboarding) error
	CreateTemplate(ctx context.Context, t *model.OnboardingTemplate) error
	FindAllTemplates(ctx context.Context) ([]model.OnboardingTemplate, error)
}

type onboardingRepository struct{ db *gorm.DB }

func NewOnboardingRepository(db *gorm.DB) OnboardingRepository {
	return &onboardingRepository{db: db}
}

func (r *onboardingRepository) Create(ctx context.Context, o *model.Onboarding) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *onboardingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Onboarding, error) {
	var o model.Onboarding
	err := r.db.WithContext(ctx).Preload("Employee").Preload("Mentor").First(&o, "id = ?", id).Error
	return &o, err
}

func (r *onboardingRepository) FindAll(ctx context.Context, status string) ([]model.Onboarding, error) {
	var list []model.Onboarding
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Preload("Employee").Preload("Mentor").Order("start_date DESC").Find(&list).Error
	return list, err
}

func (r *onboardingRepository) Update(ctx context.Context, o *model.Onboarding) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *onboardingRepository) CreateTemplate(ctx context.Context, t *model.OnboardingTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *onboardingRepository) FindAllTemplates(ctx context.Context) ([]model.OnboardingTemplate, error) {
	var list []model.OnboardingTemplate
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

// ===== OffboardingRepository =====

type OffboardingRepository interface {
	Create(ctx context.Context, o *model.Offboarding) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Offboarding, error)
	FindAll(ctx context.Context, status string) ([]model.Offboarding, error)
	Update(ctx context.Context, o *model.Offboarding) error
	GetTurnoverAnalytics(ctx context.Context) (map[string]interface{}, error)
}

type offboardingRepository struct{ db *gorm.DB }

func NewOffboardingRepository(db *gorm.DB) OffboardingRepository {
	return &offboardingRepository{db: db}
}

func (r *offboardingRepository) Create(ctx context.Context, o *model.Offboarding) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *offboardingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Offboarding, error) {
	var o model.Offboarding
	err := r.db.WithContext(ctx).Preload("Employee").First(&o, "id = ?", id).Error
	return &o, err
}

func (r *offboardingRepository) FindAll(ctx context.Context, status string) ([]model.Offboarding, error) {
	var list []model.Offboarding
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Preload("Employee").Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *offboardingRepository) Update(ctx context.Context, o *model.Offboarding) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *offboardingRepository) GetTurnoverAnalytics(ctx context.Context) (map[string]interface{}, error) {
	var totalDepartures int64
	r.db.WithContext(ctx).Model(&model.Offboarding{}).Count(&totalDepartures)

	var totalEmployees int64
	r.db.WithContext(ctx).Model(&model.HREmployee{}).Count(&totalEmployees)

	var turnoverRate float64
	if totalEmployees > 0 {
		turnoverRate = float64(totalDepartures) / float64(totalEmployees) * 100
	}

	// 理由別内訳
	type reasonCount struct {
		Reason string
		Count  int64
	}
	var reasons []reasonCount
	r.db.WithContext(ctx).Model(&model.Offboarding{}).Select("reason, COUNT(*) as count").Group("reason").Order("count DESC").Scan(&reasons)

	reasonBreakdown := make([]map[string]interface{}, len(reasons))
	for i, rc := range reasons {
		reasonBreakdown[i] = map[string]interface{}{"reason": rc.Reason, "count": rc.Count}
	}

	return map[string]interface{}{
		"turnover_rate":    turnoverRate,
		"total_departures": totalDepartures,
		"reason_breakdown": reasonBreakdown,
	}, nil
}

// ===== SurveyRepository =====

type SurveyRepository interface {
	Create(ctx context.Context, s *model.Survey) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Survey, error)
	FindAll(ctx context.Context, status, surveyType string) ([]model.Survey, error)
	Update(ctx context.Context, s *model.Survey) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateResponse(ctx context.Context, r *model.SurveyResponse) error
	FindResponsesBySurveyID(ctx context.Context, surveyID uuid.UUID) ([]model.SurveyResponse, error)
	CountResponsesBySurveyID(ctx context.Context, surveyID uuid.UUID) (int64, error)
}

type surveyRepository struct{ db *gorm.DB }

func NewSurveyRepository(db *gorm.DB) SurveyRepository {
	return &surveyRepository{db: db}
}

func (r *surveyRepository) Create(ctx context.Context, s *model.Survey) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *surveyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	var s model.Survey
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *surveyRepository) FindAll(ctx context.Context, status, surveyType string) ([]model.Survey, error) {
	var list []model.Survey
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if surveyType != "" {
		q = q.Where("type = ?", surveyType)
	}
	err := q.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *surveyRepository) Update(ctx context.Context, s *model.Survey) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *surveyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Survey{}, "id = ?", id).Error
}

func (r *surveyRepository) CreateResponse(ctx context.Context, resp *model.SurveyResponse) error {
	return r.db.WithContext(ctx).Create(resp).Error
}

func (r *surveyRepository) FindResponsesBySurveyID(ctx context.Context, surveyID uuid.UUID) ([]model.SurveyResponse, error) {
	var list []model.SurveyResponse
	err := r.db.WithContext(ctx).Where("survey_id = ?", surveyID).Find(&list).Error
	return list, err
}

func (r *surveyRepository) CountResponsesBySurveyID(ctx context.Context, surveyID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SurveyResponse{}).Where("survey_id = ?", surveyID).Count(&count).Error
	return count, err
}
