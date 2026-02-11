package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
}

// Repositories は勤怠関連リポジトリを束ねる構造体
type Repositories struct {
	User                 UserRepository
	Attendance           AttendanceRepository
	LeaveRequest         LeaveRequestRepository
	OvertimeRequest      OvertimeRequestRepository
	AttendanceCorrection AttendanceCorrectionRepository
	LeaveBalance         LeaveBalanceRepository
}

// NewRepositories は勤怠関連リポジトリを初期化する
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Attendance:           NewAttendanceRepository(db),
		LeaveRequest:         NewLeaveRequestRepository(db),
		OvertimeRequest:      NewOvertimeRequestRepository(db),
		AttendanceCorrection: NewAttendanceCorrectionRepository(db),
		LeaveBalance:         NewLeaveBalanceRepository(db),
	}
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
