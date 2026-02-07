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
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleEmployee Role = "employee"
)

// User はユーザーモデル
type User struct {
	BaseModel
	Email        string `gorm:"uniqueIndex;size:255;not null" json:"email" validate:"required,email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	FirstName    string `gorm:"size:100;not null" json:"first_name" validate:"required"`
	LastName     string `gorm:"size:100;not null" json:"last_name" validate:"required"`
	Role         Role   `gorm:"size:20;not null;default:'employee'" json:"role"`
	DepartmentID *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`

	// リレーション
	Department     *Department     `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Attendances    []Attendance    `gorm:"foreignKey:UserID" json:"attendances,omitempty"`
	LeaveRequests  []LeaveRequest  `gorm:"foreignKey:UserID" json:"leave_requests,omitempty"`
}

// ===== 部署 =====

// Department は部署モデル
type Department struct {
	BaseModel
	Name      string `gorm:"size:100;not null;uniqueIndex" json:"name" validate:"required"`
	ManagerID *uuid.UUID `gorm:"type:uuid" json:"manager_id"`

	Manager *User  `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Users   []User `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
}

// ===== 勤怠 =====

// AttendanceStatus は勤怠ステータス
type AttendanceStatus string

const (
	AttendanceStatusPresent  AttendanceStatus = "present"
	AttendanceStatusAbsent   AttendanceStatus = "absent"
	AttendanceStatusLeave    AttendanceStatus = "leave"
	AttendanceStatusHoliday  AttendanceStatus = "holiday"
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

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ===== 休暇申請 =====

// LeaveType は休暇種別
type LeaveType string

const (
	LeaveTypePaid     LeaveType = "paid"       // 有給休暇
	LeaveTypeSick     LeaveType = "sick"       // 病気休暇
	LeaveTypeSpecial  LeaveType = "special"    // 特別休暇
	LeaveTypeHalf     LeaveType = "half"       // 半休
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
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	LeaveType   LeaveType      `gorm:"size:20;not null" json:"leave_type" validate:"required"`
	StartDate   time.Time      `gorm:"type:date;not null" json:"start_date" validate:"required"`
	EndDate     time.Time      `gorm:"type:date;not null" json:"end_date" validate:"required"`
	Reason      string         `gorm:"size:500" json:"reason"`
	Status      ApprovalStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ApprovedBy  *uuid.UUID     `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt  *time.Time     `json:"approved_at"`
	RejectedReason string      `gorm:"size:500" json:"rejected_reason"`

	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Approver *User `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// ===== シフト =====

// ShiftType はシフト種別
type ShiftType string

const (
	ShiftTypeMorning   ShiftType = "morning"   // 早番
	ShiftTypeDay       ShiftType = "day"       // 日勤
	ShiftTypeEvening   ShiftType = "evening"   // 遅番
	ShiftTypeNight     ShiftType = "night"     // 夜勤
	ShiftTypeOff       ShiftType = "off"       // 休み
)

// Shift はシフトモデル
type Shift struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Date      time.Time `gorm:"type:date;not null;index" json:"date" validate:"required"`
	ShiftType ShiftType `gorm:"size:20;not null" json:"shift_type" validate:"required"`
	StartTime *time.Time `gorm:"type:time" json:"start_time"`
	EndTime   *time.Time `gorm:"type:time" json:"end_time"`
	Note      string    `gorm:"size:500" json:"note"`

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

// AutoMigrate はデータベースのマイグレーションを実行する
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Department{},
		&Attendance{},
		&LeaveRequest{},
		&Shift{},
		&RefreshToken{},
	)
}
