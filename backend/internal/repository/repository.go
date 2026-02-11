package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/gorm"
)

// Repositories は全リポジトリを束ねる構造体
type Repositories struct {
	User                 UserRepository
	Attendance           AttendanceRepository
	LeaveRequest         LeaveRequestRepository
	Shift                ShiftRepository
	Department           DepartmentRepository
	RefreshToken         RefreshTokenRepository
	OvertimeRequest      OvertimeRequestRepository
	LeaveBalance         LeaveBalanceRepository
	AttendanceCorrection AttendanceCorrectionRepository
	Notification         NotificationRepository
	Project              ProjectRepository
	TimeEntry            TimeEntryRepository
	Holiday              HolidayRepository
	ApprovalFlow         ApprovalFlowRepository
	// HR
	HREmployee   HREmployeeRepository
	HRDepartment HRDepartmentRepository
	Evaluation   EvaluationRepository
	Goal         GoalRepository
	Training     TrainingRepository
	Recruitment  RecruitmentRepository
	Document     DocumentRepository
	Announcement AnnouncementRepository
	OneOnOne     OneOnOneRepository
	Skill        SkillRepository
	Salary       SalaryRepository
	Onboarding   OnboardingRepository
	Offboarding  OffboardingRepository
	Survey       SurveyRepository
	// Expense
	Expense                    ExpenseRepository
	ExpenseItem                ExpenseItemRepository
	ExpenseComment             ExpenseCommentRepository
	ExpenseHistory             ExpenseHistoryRepository
	ExpenseTemplate            ExpenseTemplateRepository
	ExpensePolicy              ExpensePolicyRepository
	ExpenseBudget              ExpenseBudgetRepository
	ExpenseNotification        ExpenseNotificationRepository
	ExpenseReminder            ExpenseReminderRepository
	ExpenseNotificationSetting ExpenseNotificationSettingRepository
	ExpenseApprovalFlow        ExpenseApprovalFlowRepository
	ExpenseDelegate            ExpenseDelegateRepository
	ExpensePolicyViolation     ExpensePolicyViolationRepository
}

// NewRepositories は全リポジトリを初期化する
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:                 NewUserRepository(db),
		Attendance:           NewAttendanceRepository(db),
		LeaveRequest:         NewLeaveRequestRepository(db),
		Shift:                NewShiftRepository(db),
		Department:           NewDepartmentRepository(db),
		RefreshToken:         NewRefreshTokenRepository(db),
		OvertimeRequest:      NewOvertimeRequestRepository(db),
		LeaveBalance:         NewLeaveBalanceRepository(db),
		AttendanceCorrection: NewAttendanceCorrectionRepository(db),
		Notification:         NewNotificationRepository(db),
		Project:              NewProjectRepository(db),
		TimeEntry:            NewTimeEntryRepository(db),
		Holiday:              NewHolidayRepository(db),
		ApprovalFlow:         NewApprovalFlowRepository(db),
		// HR
		HREmployee:   NewHREmployeeRepository(db),
		HRDepartment: NewHRDepartmentRepository(db),
		Evaluation:   NewEvaluationRepository(db),
		Goal:         NewGoalRepository(db),
		Training:     NewTrainingRepository(db),
		Recruitment:  NewRecruitmentRepository(db),
		Document:     NewDocumentRepository(db),
		Announcement: NewAnnouncementRepository(db),
		OneOnOne:     NewOneOnOneRepository(db),
		Skill:        NewSkillRepository(db),
		Salary:       NewSalaryRepository(db),
		Onboarding:   NewOnboardingRepository(db),
		Offboarding:  NewOffboardingRepository(db),
		Survey:       NewSurveyRepository(db),
		// Expense
		Expense:                    NewExpenseRepository(db),
		ExpenseItem:                NewExpenseItemRepository(db),
		ExpenseComment:             NewExpenseCommentRepository(db),
		ExpenseHistory:             NewExpenseHistoryRepository(db),
		ExpenseTemplate:            NewExpenseTemplateRepository(db),
		ExpensePolicy:              NewExpensePolicyRepository(db),
		ExpenseBudget:              NewExpenseBudgetRepository(db),
		ExpenseNotification:        NewExpenseNotificationRepository(db),
		ExpenseReminder:            NewExpenseReminderRepository(db),
		ExpenseNotificationSetting: NewExpenseNotificationSettingRepository(db),
		ExpenseApprovalFlow:        NewExpenseApprovalFlowRepository(db),
		ExpenseDelegate:            NewExpenseDelegateRepository(db),
		ExpensePolicyViolation:     NewExpensePolicyViolationRepository(db),
	}
}

// ===== UserRepository =====

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByDepartmentID(ctx context.Context, departmentID uuid.UUID) ([]model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Department").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.WithContext(ctx).Model(&model.User{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Preload("Department").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (r *userRepository) FindByDepartmentID(ctx context.Context, departmentID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Where("department_id = ?", departmentID).Find(&users).Error
	return users, err
}

// ===== ShiftRepository =====

type ShiftRepository interface {
	Create(ctx context.Context, shift *model.Shift) error
	BulkCreate(ctx context.Context, shifts []model.Shift) error
	FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error)
	FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Shift, error)
	Update(ctx context.Context, shift *model.Shift) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) Create(ctx context.Context, shift *model.Shift) error {
	return r.db.WithContext(ctx).Create(shift).Error
}

func (r *shiftRepository) BulkCreate(ctx context.Context, shifts []model.Shift) error {
	return r.db.WithContext(ctx).CreateInBatches(shifts, 100).Error
}

func (r *shiftRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error) {
	var shifts []model.Shift
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date ASC").
		Find(&shifts).Error
	return shifts, err
}

func (r *shiftRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Shift, error) {
	var shifts []model.Shift
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("date BETWEEN ? AND ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date ASC, user_id ASC").
		Find(&shifts).Error
	return shifts, err
}

func (r *shiftRepository) Update(ctx context.Context, shift *model.Shift) error {
	return r.db.WithContext(ctx).Save(shift).Error
}

func (r *shiftRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Shift{}, "id = ?", id).Error
}

// ===== DepartmentRepository =====

type DepartmentRepository interface {
	Create(ctx context.Context, dept *model.Department) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Department, error)
	FindAll(ctx context.Context) ([]model.Department, error)
	Update(ctx context.Context, dept *model.Department) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *departmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Department, error) {
	var dept model.Department
	err := r.db.WithContext(ctx).Preload("Manager").First(&dept, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) FindAll(ctx context.Context) ([]model.Department, error) {
	var departments []model.Department
	err := r.db.WithContext(ctx).Preload("Manager").Order("name ASC").Find(&departments).Error
	return departments, err
}

func (r *departmentRepository) Update(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *departmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Department{}, "id = ?", id).Error
}

// ===== RefreshTokenRepository =====

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	Revoke(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND is_revoked = false AND expires_at > ?", token, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR is_revoked = true", time.Now()).
		Delete(&model.RefreshToken{}).Error
}

// ===== NotificationRepository =====

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	FindByUserID(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}
	query.Model(&model.Notification{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error
	return notifications, total, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": now}).Error
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Notification{}, "id = ?", id).Error
}

// ===== ProjectRepository =====

type ProjectRepository interface {
	Create(ctx context.Context, project *model.Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	FindAll(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error)
	Update(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type projectRepository struct{ db *gorm.DB }

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Preload("Manager").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindAll(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64
	query := r.db.WithContext(ctx)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Model(&model.Project{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("Manager").Offset(offset).Limit(pageSize).Order("name ASC").Find(&projects).Error
	return projects, total, err
}

func (r *projectRepository) Update(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, "id = ?", id).Error
}

// ===== TimeEntryRepository =====

type TimeEntryRepository interface {
	Create(ctx context.Context, entry *model.TimeEntry) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.TimeEntry, error)
	FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error)
	FindByProjectAndDateRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error)
	Update(ctx context.Context, entry *model.TimeEntry) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error)
}

type timeEntryRepository struct{ db *gorm.DB }

func NewTimeEntryRepository(db *gorm.DB) TimeEntryRepository {
	return &timeEntryRepository{db: db}
}

func (r *timeEntryRepository) Create(ctx context.Context, entry *model.TimeEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *timeEntryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TimeEntry, error) {
	var entry model.TimeEntry
	err := r.db.WithContext(ctx).Preload("User").Preload("Project").First(&entry, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *timeEntryRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	var entries []model.TimeEntry
	err := r.db.WithContext(ctx).Preload("Project").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date DESC").Find(&entries).Error
	return entries, err
}

func (r *timeEntryRepository) FindByProjectAndDateRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	var entries []model.TimeEntry
	err := r.db.WithContext(ctx).Preload("User").
		Where("project_id = ? AND date BETWEEN ? AND ?", projectID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date DESC").Find(&entries).Error
	return entries, err
}

func (r *timeEntryRepository) Update(ctx context.Context, entry *model.TimeEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *timeEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.TimeEntry{}, "id = ?", id).Error
}

func (r *timeEntryRepository) GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
	var summaries []model.ProjectSummary
	err := r.db.WithContext(ctx).
		Table("time_entries").
		Select("time_entries.project_id, projects.name as project_name, projects.code as project_code, COALESCE(SUM(time_entries.minutes), 0) as total_minutes, COALESCE(SUM(time_entries.minutes) / 60.0, 0) as total_hours, projects.budget_hours, COUNT(DISTINCT time_entries.user_id) as member_count").
		Joins("JOIN projects ON projects.id = time_entries.project_id").
		Where("time_entries.date BETWEEN ? AND ? AND time_entries.deleted_at IS NULL", start.Format("2006-01-02"), end.Format("2006-01-02")).
		Group("time_entries.project_id, projects.name, projects.code, projects.budget_hours").
		Scan(&summaries).Error
	return summaries, err
}

// ===== HolidayRepository =====

type HolidayRepository interface {
	Create(ctx context.Context, holiday *model.Holiday) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Holiday, error)
	FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error)
	FindByYear(ctx context.Context, year int) ([]model.Holiday, error)
	Update(ctx context.Context, holiday *model.Holiday) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsHoliday(ctx context.Context, date time.Time) (bool, *model.Holiday, error)
}

type holidayRepository struct{ db *gorm.DB }

func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepository{db: db}
}

func (r *holidayRepository) Create(ctx context.Context, holiday *model.Holiday) error {
	return r.db.WithContext(ctx).Create(holiday).Error
}

func (r *holidayRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Holiday, error) {
	var holiday model.Holiday
	err := r.db.WithContext(ctx).First(&holiday, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &holiday, nil
}

func (r *holidayRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error) {
	var holidays []model.Holiday
	err := r.db.WithContext(ctx).Where("date BETWEEN ? AND ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date ASC").Find(&holidays).Error
	return holidays, err
}

func (r *holidayRepository) FindByYear(ctx context.Context, year int) ([]model.Holiday, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.Local)
	return r.FindByDateRange(ctx, start, end)
}

func (r *holidayRepository) Update(ctx context.Context, holiday *model.Holiday) error {
	return r.db.WithContext(ctx).Save(holiday).Error
}

func (r *holidayRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Holiday{}, "id = ?", id).Error
}

func (r *holidayRepository) IsHoliday(ctx context.Context, date time.Time) (bool, *model.Holiday, error) {
	var holiday model.Holiday
	err := r.db.WithContext(ctx).Where("date = ?", date.Format("2006-01-02")).First(&holiday).Error
	if err != nil {
		return false, nil, nil
	}
	return true, &holiday, nil
}

// ===== ApprovalFlowRepository =====

type ApprovalFlowRepository interface {
	Create(ctx context.Context, flow *model.ApprovalFlow) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error)
	FindByType(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error)
	FindAll(ctx context.Context) ([]model.ApprovalFlow, error)
	Update(ctx context.Context, flow *model.ApprovalFlow) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteStepsByFlowID(ctx context.Context, flowID uuid.UUID) error
	CreateSteps(ctx context.Context, steps []model.ApprovalStep) error
}

type approvalFlowRepository struct{ db *gorm.DB }

func NewApprovalFlowRepository(db *gorm.DB) ApprovalFlowRepository {
	return &approvalFlowRepository{db: db}
}

func (r *approvalFlowRepository) Create(ctx context.Context, flow *model.ApprovalFlow) error {
	return r.db.WithContext(ctx).Create(flow).Error
}

func (r *approvalFlowRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error) {
	var flow model.ApprovalFlow
	err := r.db.WithContext(ctx).Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Preload("Steps.Approver").First(&flow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

func (r *approvalFlowRepository) FindByType(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error) {
	var flows []model.ApprovalFlow
	err := r.db.WithContext(ctx).Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Where("flow_type = ? AND is_active = true", flowType).Find(&flows).Error
	return flows, err
}

func (r *approvalFlowRepository) FindAll(ctx context.Context) ([]model.ApprovalFlow, error) {
	var flows []model.ApprovalFlow
	err := r.db.WithContext(ctx).Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Order("name ASC").Find(&flows).Error
	return flows, err
}

func (r *approvalFlowRepository) Update(ctx context.Context, flow *model.ApprovalFlow) error {
	return r.db.WithContext(ctx).Save(flow).Error
}

func (r *approvalFlowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ApprovalFlow{}, "id = ?", id).Error
}

func (r *approvalFlowRepository) DeleteStepsByFlowID(ctx context.Context, flowID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("flow_id = ?", flowID).Delete(&model.ApprovalStep{}).Error
}

func (r *approvalFlowRepository) CreateSteps(ctx context.Context, steps []model.ApprovalStep) error {
	return r.db.WithContext(ctx).CreateInBatches(steps, 100).Error
}
