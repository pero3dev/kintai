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
