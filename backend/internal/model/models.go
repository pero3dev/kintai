package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ===== 基底モデル =====

// BaseModel は全モデル共通のフィールド
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ===== ユーザー =====

// Role はユーザーの権限を表す
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleManager  Role = "manager"
	RoleEmployee Role = "employee"
)

// User はユーザーモデル
type User struct {
	BaseModel
	Email        string     `gorm:"uniqueIndex;size:255;not null" json:"email" validate:"required,email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	FirstName    string     `gorm:"size:100;not null" json:"first_name" validate:"required"`
	LastName     string     `gorm:"size:100;not null" json:"last_name" validate:"required"`
	Role         Role       `gorm:"size:20;not null;default:'employee'" json:"role"`
	DepartmentID *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`

	// リレーション
	Department    *Department    `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Attendances   []Attendance   `gorm:"foreignKey:UserID" json:"attendances,omitempty"`
	LeaveRequests []LeaveRequest `gorm:"foreignKey:UserID" json:"leave_requests,omitempty"`
}

// ===== 部署 =====

// Department は部署モデル
type Department struct {
	BaseModel
	Name      string     `gorm:"size:100;not null;uniqueIndex" json:"name" validate:"required"`
	ManagerID *uuid.UUID `gorm:"type:uuid" json:"manager_id"`

	Manager *User  `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Users   []User `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
}

// ===== 勤怠 =====

// AttendanceStatus は勤怠ステータス
type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusLeave   AttendanceStatus = "leave"
	AttendanceStatusHoliday AttendanceStatus = "holiday"
)

// Attendance は勤怠レコード
type Attendance struct {
	BaseModel
	UserID       uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Date         time.Time        `gorm:"type:date;not null;index" json:"date"`
	ClockIn      *time.Time       `json:"clock_in"`
	ClockOut     *time.Time       `json:"clock_out"`
	BreakMinutes int              `gorm:"default:0" json:"break_minutes"`
	Status       AttendanceStatus `gorm:"size:20;not null;default:'present'" json:"status"`
	Note         string           `gorm:"size:500" json:"note"`

	// 計算フィールド
	WorkMinutes     int `gorm:"default:0" json:"work_minutes"`
	OvertimeMinutes int `gorm:"default:0" json:"overtime_minutes"`

	// GPS位置情報
	ClockInLatitude   *float64 `gorm:"type:decimal(10,8)" json:"clock_in_latitude"`
	ClockInLongitude  *float64 `gorm:"type:decimal(11,8)" json:"clock_in_longitude"`
	ClockOutLatitude  *float64 `gorm:"type:decimal(10,8)" json:"clock_out_latitude"`
	ClockOutLongitude *float64 `gorm:"type:decimal(11,8)" json:"clock_out_longitude"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 休暇申請 =====

// LeaveType は休暇種別
type LeaveType string

const (
	LeaveTypePaid    LeaveType = "paid"    // 有給休暇
	LeaveTypeSick    LeaveType = "sick"    // 病気休暇
	LeaveTypeSpecial LeaveType = "special" // 特別休暇
	LeaveTypeHalf    LeaveType = "half"    // 半休
)

// ApprovalStatus は承認ステータス
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// LeaveRequest は休暇申請モデル
type LeaveRequest struct {
	BaseModel
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	LeaveType      LeaveType      `gorm:"size:20;not null" json:"leave_type" validate:"required"`
	StartDate      time.Time      `gorm:"type:date;not null" json:"start_date" validate:"required"`
	EndDate        time.Time      `gorm:"type:date;not null" json:"end_date" validate:"required"`
	Reason         string         `gorm:"size:500" json:"reason"`
	Status         ApprovalStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ApprovedBy     *uuid.UUID     `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt     *time.Time     `json:"approved_at"`
	RejectedReason string         `gorm:"size:500" json:"rejected_reason"`

	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// ===== シフト =====

// ShiftType はシフト種別
type ShiftType string

const (
	ShiftTypeMorning ShiftType = "morning" // 早番
	ShiftTypeDay     ShiftType = "day"     // 日勤
	ShiftTypeEvening ShiftType = "evening" // 遅番
	ShiftTypeNight   ShiftType = "night"   // 夜勤
	ShiftTypeOff     ShiftType = "off"     // 休み
)

// Shift はシフトモデル
type Shift struct {
	BaseModel
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Date      time.Time  `gorm:"type:date;not null;index" json:"date" validate:"required"`
	ShiftType ShiftType  `gorm:"size:20;not null" json:"shift_type" validate:"required"`
	StartTime *time.Time `gorm:"type:time" json:"start_time"`
	EndTime   *time.Time `gorm:"type:time" json:"end_time"`
	Note      string     `gorm:"size:500" json:"note"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== リフレッシュトークン =====

// RefreshToken はリフレッシュトークン管理用モデル
type RefreshToken struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"size:500;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 残業申請 =====

// OvertimeRequestStatus は残業申請ステータス
type OvertimeRequestStatus string

const (
	OvertimeStatusPending  OvertimeRequestStatus = "pending"
	OvertimeStatusApproved OvertimeRequestStatus = "approved"
	OvertimeStatusRejected OvertimeRequestStatus = "rejected"
)

// OvertimeRequest は残業申請モデル
type OvertimeRequest struct {
	BaseModel
	UserID          uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	Date            time.Time             `gorm:"type:date;not null" json:"date"`
	PlannedMinutes  int                   `gorm:"not null" json:"planned_minutes"`
	ActualMinutes   *int                  `json:"actual_minutes"`
	Reason          string                `gorm:"size:500;not null" json:"reason"`
	Status          OvertimeRequestStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ApprovedBy      *uuid.UUID            `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt      *time.Time            `json:"approved_at"`
	RejectedReason  string                `gorm:"size:500" json:"rejected_reason"`

	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// ===== 有給休暇残日数 =====

// LeaveBalance は有給休暇残日数モデル
type LeaveBalance struct {
	BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	FiscalYear   int       `gorm:"not null" json:"fiscal_year"`
	LeaveType    LeaveType `gorm:"size:20;not null" json:"leave_type"`
	TotalDays    float64   `gorm:"not null;default:0" json:"total_days"`
	UsedDays     float64   `gorm:"not null;default:0" json:"used_days"`
	CarriedOver  float64   `gorm:"not null;default:0" json:"carried_over"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 勤怠修正申請 =====

// CorrectionStatus は修正申請ステータス
type CorrectionStatus string

const (
	CorrectionStatusPending  CorrectionStatus = "pending"
	CorrectionStatusApproved CorrectionStatus = "approved"
	CorrectionStatusRejected CorrectionStatus = "rejected"
)

// AttendanceCorrection は勤怠修正申請モデル
type AttendanceCorrection struct {
	BaseModel
	UserID         uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	AttendanceID   *uuid.UUID       `gorm:"type:uuid" json:"attendance_id"`
	Date           time.Time        `gorm:"type:date;not null" json:"date"`
	OriginalClockIn  *time.Time     `json:"original_clock_in"`
	OriginalClockOut *time.Time     `json:"original_clock_out"`
	CorrectedClockIn  *time.Time    `json:"corrected_clock_in"`
	CorrectedClockOut *time.Time    `json:"corrected_clock_out"`
	Reason          string          `gorm:"size:500;not null" json:"reason"`
	Status          CorrectionStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ApprovedBy      *uuid.UUID      `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt      *time.Time      `json:"approved_at"`
	RejectedReason  string          `gorm:"size:500" json:"rejected_reason"`

	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Attendance *Attendance `gorm:"foreignKey:AttendanceID" json:"attendance,omitempty"`
	Approver   *User       `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// ===== 通知 =====

// NotificationType は通知種別
type NotificationType string

const (
	NotificationTypeLeaveApproved    NotificationType = "leave_approved"
	NotificationTypeLeaveRejected    NotificationType = "leave_rejected"
	NotificationTypeLeaveRequested   NotificationType = "leave_requested"
	NotificationTypeOvertimeAlert    NotificationType = "overtime_alert"
	NotificationTypeCorrectionResult NotificationType = "correction_result"
	NotificationTypeCorrectionReq    NotificationType = "correction_requested"
	NotificationTypeShiftChanged     NotificationType = "shift_changed"
	NotificationTypeClockReminder    NotificationType = "clock_reminder"
	NotificationTypeGeneral          NotificationType = "general"
)

// Notification は通知モデル
type Notification struct {
	BaseModel
	UserID   uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Type     NotificationType `gorm:"size:30;not null" json:"type"`
	Title    string           `gorm:"size:200;not null" json:"title"`
	Message  string           `gorm:"size:1000;not null" json:"message"`
	IsRead   bool             `gorm:"default:false" json:"is_read"`
	ReadAt   *time.Time       `json:"read_at"`
	LinkURL  string           `gorm:"size:500" json:"link_url"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== プロジェクト・工数管理 =====

// ProjectStatus はプロジェクトステータス
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusInactive ProjectStatus = "inactive"
	ProjectStatusArchived ProjectStatus = "archived"
)

// Project はプロジェクトモデル
type Project struct {
	BaseModel
	Name        string        `gorm:"size:200;not null" json:"name" validate:"required"`
	Code        string        `gorm:"size:50;uniqueIndex;not null" json:"code" validate:"required"`
	Description string        `gorm:"size:1000" json:"description"`
	Status      ProjectStatus `gorm:"size:20;not null;default:'active'" json:"status"`
	ManagerID   *uuid.UUID    `gorm:"type:uuid" json:"manager_id"`
	BudgetHours *float64      `json:"budget_hours"`

	Manager     *User       `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	TimeEntries []TimeEntry `gorm:"foreignKey:ProjectID" json:"time_entries,omitempty"`
}

// TimeEntry は工数記録モデル
type TimeEntry struct {
	BaseModel
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	Date        time.Time  `gorm:"type:date;not null;index" json:"date"`
	Minutes     int        `gorm:"not null" json:"minutes"`
	Description string     `gorm:"size:500" json:"description"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// ===== 祝日・会社カレンダー =====

// HolidayType は祝日種別
type HolidayType string

const (
	HolidayTypeNational HolidayType = "national" // 国民の祝日
	HolidayTypeCompany  HolidayType = "company"  // 会社休業日
	HolidayTypeSpecial  HolidayType = "special"  // 特別休日
)

// Holiday は祝日・会社カレンダーモデル
type Holiday struct {
	BaseModel
	Date        time.Time   `gorm:"type:date;not null;index" json:"date"`
	Name        string      `gorm:"size:200;not null" json:"name"`
	HolidayType HolidayType `gorm:"size:20;not null" json:"holiday_type"`
	IsRecurring bool        `gorm:"default:false" json:"is_recurring"`
}

// ===== 承認フロー =====

// ApprovalFlowType は承認フロー対象種別
type ApprovalFlowType string

const (
	ApprovalFlowLeave      ApprovalFlowType = "leave"
	ApprovalFlowOvertime   ApprovalFlowType = "overtime"
	ApprovalFlowCorrection ApprovalFlowType = "correction"
)

// ApprovalFlow は承認フロー定義モデル
type ApprovalFlow struct {
	BaseModel
	Name     string           `gorm:"size:200;not null" json:"name"`
	FlowType ApprovalFlowType `gorm:"size:30;not null" json:"flow_type"`
	IsActive bool             `gorm:"default:true" json:"is_active"`

	Steps []ApprovalStep `gorm:"foreignKey:FlowID" json:"steps,omitempty"`
}

// ApprovalStepType は承認ステップの承認者種別
type ApprovalStepType string

const (
	ApprovalStepManager  ApprovalStepType = "manager"   // 直属マネージャー
	ApprovalStepRole     ApprovalStepType = "role"      // 特定ロール
	ApprovalStepUser     ApprovalStepType = "user"      // 特定ユーザー
)

// ApprovalStep は承認ステップモデル
type ApprovalStep struct {
	BaseModel
	FlowID       uuid.UUID        `gorm:"type:uuid;not null;index" json:"flow_id"`
	StepOrder    int              `gorm:"not null" json:"step_order"`
	StepType     ApprovalStepType `gorm:"size:20;not null" json:"step_type"`
	ApproverRole *Role            `gorm:"size:20" json:"approver_role"`
	ApproverID   *uuid.UUID       `gorm:"type:uuid" json:"approver_id"`

	Flow     *ApprovalFlow `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
	Approver *User         `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}

// AutoMigrate はデータベースのマイグレーションを実行する
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Department{},
		&Attendance{},
		&LeaveRequest{},
		&Shift{},
		&RefreshToken{},
		&OvertimeRequest{},
		&LeaveBalance{},
		&AttendanceCorrection{},
		&Notification{},
		&Project{},
		&TimeEntry{},
		&Holiday{},
		&ApprovalFlow{},
		&ApprovalStep{},
	)
}
