package model

import (
	"time"

	"github.com/google/uuid"
)

// ===== 認証 =====

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ===== 打刻 =====

type ClockInRequest struct {
	Note string `json:"note"`
}

type ClockOutRequest struct {
	Note string `json:"note"`
}

// ===== 勤怠一覧 =====

type AttendanceListQuery struct {
	UserID    *uuid.UUID `form:"user_id"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02"`
	Page      int        `form:"page,default=1"`
	PageSize  int        `form:"page_size,default=20"`
}

type AttendanceSummary struct {
	TotalWorkDays        int     `json:"total_work_days"`
	TotalWorkMinutes     int     `json:"total_work_minutes"`
	TotalOvertimeMinutes int     `json:"total_overtime_minutes"`
	AverageWorkMinutes   float64 `json:"average_work_minutes"`
	AbsentDays           int     `json:"absent_days"`
	LeaveDays            int     `json:"leave_days"`
}

// ===== 休暇申請 =====

type LeaveRequestCreate struct {
	LeaveType LeaveType `json:"leave_type" validate:"required,oneof=paid sick special half"`
	StartDate string    `json:"start_date" validate:"required"`
	EndDate   string    `json:"end_date" validate:"required"`
	Reason    string    `json:"reason"`
}

type LeaveRequestApproval struct {
	Status         ApprovalStatus `json:"status" validate:"required,oneof=approved rejected"`
	RejectedReason string         `json:"rejected_reason"`
}

// ===== シフト =====

type ShiftCreateRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Date      string    `json:"date" validate:"required"`
	ShiftType ShiftType `json:"shift_type" validate:"required,oneof=morning day evening night off"`
	StartTime *string   `json:"start_time"`
	EndTime   *string   `json:"end_time"`
	Note      string    `json:"note"`
}

type ShiftBulkCreateRequest struct {
	Shifts []ShiftCreateRequest `json:"shifts" validate:"required,dive"`
}

// ===== ユーザー管理 =====

type UserCreateRequest struct {
	Email        string     `json:"email" validate:"required,email"`
	Password     string     `json:"password" validate:"required,min=8"`
	FirstName    string     `json:"first_name" validate:"required"`
	LastName     string     `json:"last_name" validate:"required"`
	Role         Role       `json:"role" validate:"required,oneof=admin manager employee"`
	DepartmentID *uuid.UUID `json:"department_id"`
}

type UserUpdateRequest struct {
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	Role         *Role      `json:"role"`
	DepartmentID *uuid.UUID `json:"department_id"`
	IsActive     *bool      `json:"is_active"`
	Password     *string    `json:"password"`
}

// ===== ページネーション =====

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ===== エラーレスポンス =====

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ===== ダッシュボード =====

type DashboardStats struct {
	TodayPresentCount int              `json:"today_present_count"`
	TodayAbsentCount  int              `json:"today_absent_count"`
	PendingLeaves     int              `json:"pending_leaves"`
	MonthlyOvertime   int              `json:"monthly_overtime"`
	DepartmentStats   []DepartmentStat `json:"department_stats"`
}

type DepartmentStat struct {
	DepartmentName string  `json:"department_name"`
	TotalEmployees int     `json:"total_employees"`
	PresentToday   int     `json:"present_today"`
	AttendanceRate float64 `json:"attendance_rate"`
}
