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
	User         UserRepository
	Attendance   AttendanceRepository
	LeaveRequest LeaveRequestRepository
	Shift        ShiftRepository
	Department   DepartmentRepository
	RefreshToken RefreshTokenRepository
}

// NewRepositories は全リポジトリを初期化する
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db),
		Attendance:   NewAttendanceRepository(db),
		LeaveRequest: NewLeaveRequestRepository(db),
		Shift:        NewShiftRepository(db),
		Department:   NewDepartmentRepository(db),
		RefreshToken: NewRefreshTokenRepository(db),
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
