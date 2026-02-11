package model

import (
	"time"

	"github.com/google/uuid"
)

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
	UserID  uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	Type    NotificationType `gorm:"size:30;not null" json:"type"`
	Title   string           `gorm:"size:200;not null" json:"title"`
	Message string           `gorm:"size:1000;not null" json:"message"`
	IsRead  bool             `gorm:"default:false" json:"is_read"`
	ReadAt  *time.Time       `json:"read_at"`
	LinkURL string           `gorm:"size:500" json:"link_url"`

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
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Date        time.Time `gorm:"type:date;not null;index" json:"date"`
	Minutes     int       `gorm:"not null" json:"minutes"`
	Description string    `gorm:"size:500" json:"description"`

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
	ApprovalStepManager ApprovalStepType = "manager" // 直属マネージャー
	ApprovalStepRole    ApprovalStepType = "role"    // 特定ロール
	ApprovalStepUser    ApprovalStepType = "user"    // 特定ユーザー
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
