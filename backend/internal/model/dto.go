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

// ===== 打刻（位置情報付き） =====

type ClockInWithLocationRequest struct {
	Note      string   `json:"note"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type ClockOutWithLocationRequest struct {
	Note      string   `json:"note"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// ===== 残業申請 =====

type OvertimeRequestCreate struct {
	Date           string `json:"date" validate:"required"`
	PlannedMinutes int    `json:"planned_minutes" validate:"required,min=1"`
	Reason         string `json:"reason" validate:"required"`
}

type OvertimeRequestApproval struct {
	Status         OvertimeRequestStatus `json:"status" validate:"required,oneof=approved rejected"`
	RejectedReason string                `json:"rejected_reason"`
}

type OvertimeAlert struct {
	UserID               uuid.UUID `json:"user_id"`
	UserName             string    `json:"user_name"`
	MonthlyOvertimeHours float64   `json:"monthly_overtime_hours"`
	YearlyOvertimeHours  float64   `json:"yearly_overtime_hours"`
	MonthlyLimitHours    float64   `json:"monthly_limit_hours"`
	YearlyLimitHours     float64   `json:"yearly_limit_hours"`
	IsMonthlyExceeded    bool      `json:"is_monthly_exceeded"`
	IsYearlyExceeded     bool      `json:"is_yearly_exceeded"`
}

// ===== 有給休暇残日数 =====

type LeaveBalanceResponse struct {
	LeaveType     LeaveType `json:"leave_type"`
	TotalDays     float64   `json:"total_days"`
	UsedDays      float64   `json:"used_days"`
	RemainingDays float64   `json:"remaining_days"`
	CarriedOver   float64   `json:"carried_over"`
	FiscalYear    int       `json:"fiscal_year"`
}

type LeaveBalanceUpdate struct {
	TotalDays   *float64 `json:"total_days"`
	CarriedOver *float64 `json:"carried_over"`
}

// ===== 勤怠修正申請 =====

type AttendanceCorrectionCreate struct {
	Date              string  `json:"date" validate:"required"`
	CorrectedClockIn  *string `json:"corrected_clock_in"`
	CorrectedClockOut *string `json:"corrected_clock_out"`
	Reason            string  `json:"reason" validate:"required"`
}

type AttendanceCorrectionApproval struct {
	Status         CorrectionStatus `json:"status" validate:"required,oneof=approved rejected"`
	RejectedReason string           `json:"rejected_reason"`
}

// ===== 通知 =====

type NotificationListQuery struct {
	IsRead   *bool `form:"is_read"`
	Page     int   `form:"page,default=1"`
	PageSize int   `form:"page_size,default=20"`
}

type NotificationCount struct {
	Total  int `json:"total"`
	Unread int `json:"unread"`
}

// ===== プロジェクト・工数管理 =====

type ProjectCreateRequest struct {
	Name        string     `json:"name" validate:"required"`
	Code        string     `json:"code" validate:"required"`
	Description string     `json:"description"`
	ManagerID   *uuid.UUID `json:"manager_id"`
	BudgetHours *float64   `json:"budget_hours"`
}

type ProjectUpdateRequest struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	Status      *ProjectStatus `json:"status"`
	ManagerID   *uuid.UUID     `json:"manager_id"`
	BudgetHours *float64       `json:"budget_hours"`
}

type TimeEntryCreate struct {
	ProjectID   uuid.UUID `json:"project_id" validate:"required"`
	Date        string    `json:"date" validate:"required"`
	Minutes     int       `json:"minutes" validate:"required,min=1"`
	Description string    `json:"description"`
}

type TimeEntryUpdate struct {
	Minutes     *int    `json:"minutes"`
	Description *string `json:"description"`
}

type ProjectSummary struct {
	ProjectID    uuid.UUID `json:"project_id"`
	ProjectName  string    `json:"project_name"`
	ProjectCode  string    `json:"project_code"`
	TotalMinutes int       `json:"total_minutes"`
	TotalHours   float64   `json:"total_hours"`
	BudgetHours  *float64  `json:"budget_hours"`
	MemberCount  int       `json:"member_count"`
}

// ===== 祝日・会社カレンダー =====

type HolidayCreateRequest struct {
	Date        string      `json:"date" validate:"required"`
	Name        string      `json:"name" validate:"required"`
	HolidayType HolidayType `json:"holiday_type" validate:"required,oneof=national company special"`
	IsRecurring bool        `json:"is_recurring"`
}

type HolidayUpdateRequest struct {
	Name        *string      `json:"name"`
	HolidayType *HolidayType `json:"holiday_type"`
	IsRecurring *bool        `json:"is_recurring"`
}

type CalendarDay struct {
	Date        string       `json:"date"`
	IsHoliday   bool         `json:"is_holiday"`
	IsWeekend   bool         `json:"is_weekend"`
	HolidayName string       `json:"holiday_name,omitempty"`
	HolidayType *HolidayType `json:"holiday_type,omitempty"`
}

type WorkingDaysSummary struct {
	TotalDays   int `json:"total_days"`
	WorkingDays int `json:"working_days"`
	Holidays    int `json:"holidays"`
	Weekends    int `json:"weekends"`
}

// ===== 承認フロー =====

type ApprovalFlowCreateRequest struct {
	Name     string                `json:"name" validate:"required"`
	FlowType ApprovalFlowType      `json:"flow_type" validate:"required,oneof=leave overtime correction"`
	Steps    []ApprovalStepRequest `json:"steps" validate:"required,min=1"`
}

type ApprovalStepRequest struct {
	StepOrder    int              `json:"step_order" validate:"required"`
	StepType     ApprovalStepType `json:"step_type" validate:"required,oneof=manager role user"`
	ApproverRole *Role            `json:"approver_role"`
	ApproverID   *uuid.UUID       `json:"approver_id"`
}

type ApprovalFlowUpdateRequest struct {
	Name     *string               `json:"name"`
	IsActive *bool                 `json:"is_active"`
	Steps    []ApprovalStepRequest `json:"steps"`
}

// ===== CSVエクスポート =====

type ExportRequest struct {
	Type      string     `form:"type" validate:"required,oneof=attendance leaves overtime projects"`
	StartDate string     `form:"start_date" validate:"required"`
	EndDate   string     `form:"end_date" validate:"required"`
	UserID    *uuid.UUID `form:"user_id"`
}

// ===== ダッシュボード拡張 =====

type DashboardTrend struct {
	Date            string  `json:"date"`
	PresentCount    int     `json:"present_count"`
	AbsentCount     int     `json:"absent_count"`
	LeaveCount      int     `json:"leave_count"`
	OvertimeMinutes int     `json:"overtime_minutes"`
	AttendanceRate  float64 `json:"attendance_rate"`
}

type DashboardStatsExtended struct {
	DashboardStats
	WeeklyTrend        []DashboardTrend `json:"weekly_trend"`
	MonthlyTrend       []DashboardTrend `json:"monthly_trend"`
	OvertimeAlerts     []OvertimeAlert  `json:"overtime_alerts"`
	UpcomingHolidays   []Holiday        `json:"upcoming_holidays"`
	PendingCorrections int              `json:"pending_corrections"`
	PendingOvertimes   int              `json:"pending_overtimes"`
}
