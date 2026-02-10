package model

import (
	"time"

	"github.com/google/uuid"
)

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
	UserID         uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	Date           time.Time             `gorm:"type:date;not null" json:"date"`
	PlannedMinutes int                   `gorm:"not null" json:"planned_minutes"`
	ActualMinutes  *int                  `json:"actual_minutes"`
	Reason         string                `gorm:"size:500;not null" json:"reason"`
	Status         OvertimeRequestStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ApprovedBy     *uuid.UUID            `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt     *time.Time            `json:"approved_at"`
	RejectedReason string                `gorm:"size:500" json:"rejected_reason"`

	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// ===== 有給休暇残日数 =====

// LeaveBalance は有給休暇残日数モデル
type LeaveBalance struct {
	BaseModel
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	FiscalYear  int       `gorm:"not null" json:"fiscal_year"`
	LeaveType   LeaveType `gorm:"size:20;not null" json:"leave_type"`
	TotalDays   float64   `gorm:"not null;default:0" json:"total_days"`
	UsedDays    float64   `gorm:"not null;default:0" json:"used_days"`
	CarriedOver float64   `gorm:"not null;default:0" json:"carried_over"`

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
	UserID            uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	AttendanceID      *uuid.UUID       `gorm:"type:uuid" json:"attendance_id"`
	Date              time.Time        `gorm:"type:date;not null" json:"date"`
	OriginalClockIn   *time.Time       `json:"original_clock_in"`
	OriginalClockOut  *time.Time       `json:"original_clock_out"`
	CorrectedClockIn  *time.Time       `json:"corrected_clock_in"`
	CorrectedClockOut *time.Time       `json:"corrected_clock_out"`
	Reason            string           `gorm:"size:500;not null" json:"reason"`
	Status            CorrectionStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ApprovedBy        *uuid.UUID       `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt        *time.Time       `json:"approved_at"`
	RejectedReason    string           `gorm:"size:500" json:"rejected_reason"`

	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Attendance *Attendance `gorm:"foreignKey:AttendanceID" json:"attendance,omitempty"`
	Approver   *User       `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}
