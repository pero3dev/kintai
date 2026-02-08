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

// ===== AttendanceRepository =====

type AttendanceRepository interface {
	Create(ctx context.Context, attendance *model.Attendance) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Attendance, error)
	FindByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*model.Attendance, error)
	FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error)
	FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Attendance, error)
	Update(ctx context.Context, attendance *model.Attendance) error
	GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error)
	CountTodayPresent(ctx context.Context) (int64, error)
	CountTodayAbsent(ctx context.Context, totalUsers int64) (int64, error)
	GetMonthlyOvertime(ctx context.Context, start, end time.Time) (int64, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) Create(ctx context.Context, attendance *model.Attendance) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

func (r *attendanceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Attendance, error) {
	var attendance model.Attendance
	err := r.db.WithContext(ctx).Preload("User").First(&attendance, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

func (r *attendanceRepository) FindByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*model.Attendance, error) {
	var attendance model.Attendance
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).
		First(&attendance).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

func (r *attendanceRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
	var attendances []model.Attendance
	var total int64

	query := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02"))

	query.Model(&model.Attendance{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("date DESC").
		Find(&attendances).Error

	return attendances, total, err
}

func (r *attendanceRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Attendance, error) {
	var attendances []model.Attendance
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("date BETWEEN ? AND ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
		Find(&attendances).Error
	return attendances, err
}

func (r *attendanceRepository) Update(ctx context.Context, attendance *model.Attendance) error {
	return r.db.WithContext(ctx).Save(attendance).Error
}

func (r *attendanceRepository) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
	var summary model.AttendanceSummary

	r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start, end).
		Select("COUNT(*) as total_work_days, COALESCE(SUM(work_minutes), 0) as total_work_minutes, COALESCE(SUM(overtime_minutes), 0) as total_overtime_minutes, COALESCE(AVG(work_minutes), 0) as average_work_minutes").
		Where("status = ?", model.AttendanceStatusPresent).
		Scan(&summary)

	var absentCount, leaveCount int64
	r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("user_id = ? AND date BETWEEN ? AND ? AND status = ?", userID, start, end, model.AttendanceStatusAbsent).
		Count(&absentCount)
	summary.AbsentDays = int(absentCount)

	r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("user_id = ? AND date BETWEEN ? AND ? AND status = ?", userID, start, end, model.AttendanceStatusLeave).
		Count(&leaveCount)
	summary.LeaveDays = int(leaveCount)

	return &summary, nil
}

func (r *attendanceRepository) CountTodayPresent(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("date = ? AND clock_in IS NOT NULL", today).
		Count(&count).Error
	return count, err
}

func (r *attendanceRepository) CountTodayAbsent(ctx context.Context, totalUsers int64) (int64, error) {
	presentCount, err := r.CountTodayPresent(ctx)
	if err != nil {
		return 0, err
	}
	return totalUsers - presentCount, nil
}

func (r *attendanceRepository) GetMonthlyOvertime(ctx context.Context, start, end time.Time) (int64, error) {
	var totalOvertime int64
	err := r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("date BETWEEN ? AND ?", start.Format("2006-01-02"), end.Format("2006-01-02")).
		Select("COALESCE(SUM(overtime_minutes), 0)").
		Scan(&totalOvertime).Error
	return totalOvertime, err
}

// ===== LeaveRequestRepository =====

type LeaveRequestRepository interface {
	Create(ctx context.Context, req *model.LeaveRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveRequest, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error)
	FindPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error)
	Update(ctx context.Context, req *model.LeaveRequest) error
	CountPending(ctx context.Context) (int64, error)
}

type leaveRequestRepository struct {
	db *gorm.DB
}

func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return &leaveRequestRepository{db: db}
}

func (r *leaveRequestRepository) Create(ctx context.Context, req *model.LeaveRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *leaveRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveRequest, error) {
	var req model.LeaveRequest
	err := r.db.WithContext(ctx).Preload("User").Preload("Approver").First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *leaveRequestRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	var requests []model.LeaveRequest
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	query.Model(&model.LeaveRequest{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Approver").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&requests).Error

	return requests, total, err
}

func (r *leaveRequestRepository) FindPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	var requests []model.LeaveRequest
	var total int64

	query := r.db.WithContext(ctx).Where("status = ?", model.ApprovalStatusPending)
	query.Model(&model.LeaveRequest{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").
		Offset(offset).Limit(pageSize).
		Order("created_at ASC").
		Find(&requests).Error

	return requests, total, err
}

func (r *leaveRequestRepository) Update(ctx context.Context, req *model.LeaveRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *leaveRequestRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.LeaveRequest{}).Where("status = ?", model.ApprovalStatusPending).Count(&count).Error
	return count, err
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

// ===== OvertimeRequestRepository =====

type OvertimeRequestRepository interface {
	Create(ctx context.Context, req *model.OvertimeRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.OvertimeRequest, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	FindPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	Update(ctx context.Context, req *model.OvertimeRequest) error
	CountPending(ctx context.Context) (int64, error)
	GetUserMonthlyOvertime(ctx context.Context, userID uuid.UUID, year, month int) (int64, error)
	GetUserYearlyOvertime(ctx context.Context, userID uuid.UUID, year int) (int64, error)
}

type overtimeRequestRepository struct{ db *gorm.DB }

func NewOvertimeRequestRepository(db *gorm.DB) OvertimeRequestRepository {
	return &overtimeRequestRepository{db: db}
}

func (r *overtimeRequestRepository) Create(ctx context.Context, req *model.OvertimeRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *overtimeRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.OvertimeRequest, error) {
	var req model.OvertimeRequest
	err := r.db.WithContext(ctx).Preload("User").Preload("Approver").First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *overtimeRequestRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	var requests []model.OvertimeRequest
	var total int64
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	query.Model(&model.OvertimeRequest{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("Approver").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&requests).Error
	return requests, total, err
}

func (r *overtimeRequestRepository) FindPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	var requests []model.OvertimeRequest
	var total int64
	query := r.db.WithContext(ctx).Where("status = ?", model.OvertimeStatusPending)
	query.Model(&model.OvertimeRequest{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&requests).Error
	return requests, total, err
}

func (r *overtimeRequestRepository) Update(ctx context.Context, req *model.OvertimeRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *overtimeRequestRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.OvertimeRequest{}).Where("status = ?", model.OvertimeStatusPending).Count(&count).Error
	return count, err
}

func (r *overtimeRequestRepository) GetUserMonthlyOvertime(ctx context.Context, userID uuid.UUID, year, month int) (int64, error) {
	var total int64
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	err := r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Select("COALESCE(SUM(overtime_minutes), 0)").Scan(&total).Error
	return total, err
}

func (r *overtimeRequestRepository) GetUserYearlyOvertime(ctx context.Context, userID uuid.UUID, year int) (int64, error) {
	var total int64
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.Local)
	err := r.db.WithContext(ctx).Model(&model.Attendance{}).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Select("COALESCE(SUM(overtime_minutes), 0)").Scan(&total).Error
	return total, err
}

// ===== LeaveBalanceRepository =====

type LeaveBalanceRepository interface {
	Create(ctx context.Context, balance *model.LeaveBalance) error
	FindByUserAndYear(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalance, error)
	FindByUserYearAndType(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType) (*model.LeaveBalance, error)
	Update(ctx context.Context, balance *model.LeaveBalance) error
	Upsert(ctx context.Context, balance *model.LeaveBalance) error
}

type leaveBalanceRepository struct{ db *gorm.DB }

func NewLeaveBalanceRepository(db *gorm.DB) LeaveBalanceRepository {
	return &leaveBalanceRepository{db: db}
}

func (r *leaveBalanceRepository) Create(ctx context.Context, balance *model.LeaveBalance) error {
	return r.db.WithContext(ctx).Create(balance).Error
}

func (r *leaveBalanceRepository) FindByUserAndYear(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalance, error) {
	var balances []model.LeaveBalance
	err := r.db.WithContext(ctx).Where("user_id = ? AND fiscal_year = ?", userID, fiscalYear).Find(&balances).Error
	return balances, err
}

func (r *leaveBalanceRepository) FindByUserYearAndType(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType) (*model.LeaveBalance, error) {
	var balance model.LeaveBalance
	err := r.db.WithContext(ctx).Where("user_id = ? AND fiscal_year = ? AND leave_type = ?", userID, fiscalYear, leaveType).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *leaveBalanceRepository) Update(ctx context.Context, balance *model.LeaveBalance) error {
	return r.db.WithContext(ctx).Save(balance).Error
}

func (r *leaveBalanceRepository) Upsert(ctx context.Context, balance *model.LeaveBalance) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND fiscal_year = ? AND leave_type = ?", balance.UserID, balance.FiscalYear, balance.LeaveType).
		Assign(model.LeaveBalance{TotalDays: balance.TotalDays, UsedDays: balance.UsedDays, CarriedOver: balance.CarriedOver}).
		FirstOrCreate(balance).Error
}

// ===== AttendanceCorrectionRepository =====

type AttendanceCorrectionRepository interface {
	Create(ctx context.Context, correction *model.AttendanceCorrection) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceCorrection, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
	FindPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
	Update(ctx context.Context, correction *model.AttendanceCorrection) error
	CountPending(ctx context.Context) (int64, error)
}

type attendanceCorrectionRepository struct{ db *gorm.DB }

func NewAttendanceCorrectionRepository(db *gorm.DB) AttendanceCorrectionRepository {
	return &attendanceCorrectionRepository{db: db}
}

func (r *attendanceCorrectionRepository) Create(ctx context.Context, correction *model.AttendanceCorrection) error {
	return r.db.WithContext(ctx).Create(correction).Error
}

func (r *attendanceCorrectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceCorrection, error) {
	var correction model.AttendanceCorrection
	err := r.db.WithContext(ctx).Preload("User").Preload("Attendance").Preload("Approver").First(&correction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &correction, nil
}

func (r *attendanceCorrectionRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	var corrections []model.AttendanceCorrection
	var total int64
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	query.Model(&model.AttendanceCorrection{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("Approver").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&corrections).Error
	return corrections, total, err
}

func (r *attendanceCorrectionRepository) FindPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	var corrections []model.AttendanceCorrection
	var total int64
	query := r.db.WithContext(ctx).Where("status = ?", model.CorrectionStatusPending)
	query.Model(&model.AttendanceCorrection{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&corrections).Error
	return corrections, total, err
}

func (r *attendanceCorrectionRepository) Update(ctx context.Context, correction *model.AttendanceCorrection) error {
	return r.db.WithContext(ctx).Save(correction).Error
}

func (r *attendanceCorrectionRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AttendanceCorrection{}).Where("status = ?", model.CorrectionStatusPending).Count(&count).Error
	return count, err
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
