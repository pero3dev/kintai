package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ===== HR社員 =====

type EmploymentType string

const (
	EmploymentTypeFullTime EmploymentType = "full_time"
	EmploymentTypePartTime EmploymentType = "part_time"
	EmploymentTypeContract EmploymentType = "contract"
	EmploymentTypeIntern   EmploymentType = "intern"
)

type EmployeeStatus string

const (
	EmployeeStatusActive   EmployeeStatus = "active"
	EmployeeStatusInactive EmployeeStatus = "inactive"
	EmployeeStatusOnLeave  EmployeeStatus = "on_leave"
)

type HREmployee struct {
	BaseModel
	UserID         *uuid.UUID     `gorm:"type:uuid;uniqueIndex" json:"user_id"`
	EmployeeCode   string         `gorm:"size:50;uniqueIndex;not null" json:"employee_code"`
	FirstName      string         `gorm:"size:100;not null" json:"first_name"`
	LastName       string         `gorm:"size:100;not null" json:"last_name"`
	Email          string         `gorm:"size:255;not null" json:"email"`
	Phone          string         `gorm:"size:50" json:"phone"`
	Position       string         `gorm:"size:100" json:"position"`
	Grade          string         `gorm:"size:50" json:"grade"`
	DepartmentID   *uuid.UUID     `gorm:"type:uuid" json:"department_id"`
	ManagerID      *uuid.UUID     `gorm:"type:uuid" json:"manager_id"`
	EmploymentType EmploymentType `gorm:"size:20;not null;default:'full_time'" json:"employment_type"`
	Status         EmployeeStatus `gorm:"size:20;not null;default:'active'" json:"status"`
	HireDate       *time.Time     `gorm:"type:date" json:"hire_date"`
	BirthDate      *time.Time     `gorm:"type:date" json:"birth_date"`
	Address        string         `gorm:"size:500" json:"address"`
	BaseSalary     float64        `gorm:"default:0" json:"base_salary"`

	Department *HRDepartment `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Manager    *HREmployee   `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

// ===== HR部門 =====

type HRDepartment struct {
	BaseModel
	Name        string     `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Code        string     `gorm:"size:20;uniqueIndex" json:"code"`
	Description string     `gorm:"size:500" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	ManagerID   *uuid.UUID `gorm:"type:uuid" json:"manager_id"`
	Budget      float64    `gorm:"default:0" json:"budget"`

	Parent   *HRDepartment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Manager  *HREmployee    `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Children []HRDepartment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// ===== 評価 =====

type EvaluationStatus string

const (
	EvaluationStatusDraft     EvaluationStatus = "draft"
	EvaluationStatusSubmitted EvaluationStatus = "submitted"
	EvaluationStatusReviewed  EvaluationStatus = "reviewed"
	EvaluationStatusFinalized EvaluationStatus = "finalized"
)

type EvaluationCycle struct {
	BaseModel
	Name      string    `gorm:"size:200;not null" json:"name"`
	StartDate time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate   time.Time `gorm:"type:date;not null" json:"end_date"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

type Evaluation struct {
	BaseModel
	EmployeeID  uuid.UUID        `gorm:"type:uuid;not null;index" json:"employee_id"`
	CycleID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"cycle_id"`
	ReviewerID  *uuid.UUID       `gorm:"type:uuid" json:"reviewer_id"`
	Status      EvaluationStatus `gorm:"size:20;not null;default:'draft'" json:"status"`
	SelfScore   *float64         `json:"self_score"`
	ManagerScore *float64        `json:"manager_score"`
	FinalScore  *float64         `json:"final_score"`
	SelfComment string           `gorm:"type:text" json:"self_comment"`
	ManagerComment string        `gorm:"type:text" json:"manager_comment"`
	Goals       string           `gorm:"type:text" json:"goals"`

	Employee *HREmployee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Cycle    *EvaluationCycle  `gorm:"foreignKey:CycleID" json:"cycle,omitempty"`
	Reviewer *HREmployee      `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
}

// ===== 目標 =====

type GoalStatus string

const (
	GoalStatusNotStarted GoalStatus = "not_started"
	GoalStatusInProgress GoalStatus = "in_progress"
	GoalStatusCompleted  GoalStatus = "completed"
	GoalStatusCancelled  GoalStatus = "cancelled"
)

type GoalCategory string

const (
	GoalCategoryPerformance GoalCategory = "performance"
	GoalCategoryDevelopment GoalCategory = "development"
	GoalCategoryBehavior    GoalCategory = "behavior"
)

type HRGoal struct {
	BaseModel
	EmployeeID  uuid.UUID    `gorm:"type:uuid;not null;index" json:"employee_id"`
	Title       string       `gorm:"size:200;not null" json:"title"`
	Description string       `gorm:"type:text" json:"description"`
	Category    GoalCategory `gorm:"size:30;not null;default:'performance'" json:"category"`
	Status      GoalStatus   `gorm:"size:20;not null;default:'not_started'" json:"status"`
	Progress    int          `gorm:"default:0" json:"progress"`
	StartDate   *time.Time   `gorm:"type:date" json:"start_date"`
	DueDate     *time.Time   `gorm:"type:date" json:"due_date"`
	Weight      int          `gorm:"default:1" json:"weight"`

	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== 研修 =====

type TrainingStatus string

const (
	TrainingStatusScheduled  TrainingStatus = "scheduled"
	TrainingStatusInProgress TrainingStatus = "in_progress"
	TrainingStatusCompleted  TrainingStatus = "completed"
	TrainingStatusCancelled  TrainingStatus = "cancelled"
)

type TrainingProgram struct {
	BaseModel
	Title          string         `gorm:"size:200;not null" json:"title"`
	Description    string         `gorm:"type:text" json:"description"`
	Category       string         `gorm:"size:50" json:"category"`
	InstructorName string         `gorm:"size:100" json:"instructor_name"`
	Status         TrainingStatus `gorm:"size:20;not null;default:'scheduled'" json:"status"`
	StartDate      *time.Time     `gorm:"type:date" json:"start_date"`
	EndDate        *time.Time     `gorm:"type:date" json:"end_date"`
	MaxParticipants int           `gorm:"default:0" json:"max_participants"`
	Location       string         `gorm:"size:200" json:"location"`
	IsOnline       bool           `gorm:"default:false" json:"is_online"`

	Enrollments []TrainingEnrollment `gorm:"foreignKey:ProgramID" json:"enrollments,omitempty"`
}

type TrainingEnrollment struct {
	BaseModel
	ProgramID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"program_id"`
	EmployeeID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	Status      string     `gorm:"size:20;not null;default:'enrolled'" json:"status"`
	CompletedAt *time.Time `json:"completed_at"`
	Score       *float64   `json:"score"`
	Feedback    string     `gorm:"type:text" json:"feedback"`

	Program  *TrainingProgram `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	Employee *HREmployee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== 採用 =====

type PositionStatus string

const (
	PositionStatusOpen   PositionStatus = "open"
	PositionStatusClosed PositionStatus = "closed"
	PositionStatusOnHold PositionStatus = "on_hold"
)

type RecruitmentPosition struct {
	BaseModel
	Title        string         `gorm:"size:200;not null" json:"title"`
	DepartmentID *uuid.UUID     `gorm:"type:uuid" json:"department_id"`
	Description  string         `gorm:"type:text" json:"description"`
	Requirements string         `gorm:"type:text" json:"requirements"`
	Status       PositionStatus `gorm:"size:20;not null;default:'open'" json:"status"`
	Openings     int            `gorm:"default:1" json:"openings"`
	Location     string         `gorm:"size:200" json:"location"`
	SalaryMin    *float64       `json:"salary_min"`
	SalaryMax    *float64       `json:"salary_max"`

	Department *HRDepartment `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Applicants []Applicant   `gorm:"foreignKey:PositionID" json:"applicants,omitempty"`
}

type ApplicantStage string

const (
	ApplicantStageNew       ApplicantStage = "new"
	ApplicantStageScreening ApplicantStage = "screening"
	ApplicantStageInterview ApplicantStage = "interview"
	ApplicantStageOffer     ApplicantStage = "offer"
	ApplicantStageHired     ApplicantStage = "hired"
	ApplicantStageRejected  ApplicantStage = "rejected"
)

type Applicant struct {
	BaseModel
	PositionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"position_id"`
	Name       string         `gorm:"size:200;not null" json:"name"`
	Email      string         `gorm:"size:255;not null" json:"email"`
	Phone      string         `gorm:"size:50" json:"phone"`
	ResumeURL  string         `gorm:"size:500" json:"resume_url"`
	Stage      ApplicantStage `gorm:"size:20;not null;default:'new'" json:"stage"`
	Notes      string         `gorm:"type:text" json:"notes"`
	Rating     *int           `json:"rating"`
	AppliedAt  time.Time      `gorm:"not null;default:now()" json:"applied_at"`

	Position *RecruitmentPosition `gorm:"foreignKey:PositionID" json:"position,omitempty"`
}

// ===== 書類管理 =====

type HRDocument struct {
	BaseModel
	EmployeeID *uuid.UUID `gorm:"type:uuid;index" json:"employee_id"`
	Title      string     `gorm:"size:200;not null" json:"title"`
	Type       string     `gorm:"size:50;not null" json:"type"`
	FileName   string     `gorm:"size:255;not null" json:"file_name"`
	FilePath   string     `gorm:"size:500;not null" json:"file_path"`
	FileSize   int64      `json:"file_size"`
	MimeType   string     `gorm:"size:100" json:"mime_type"`
	UploadedBy uuid.UUID  `gorm:"type:uuid;not null" json:"uploaded_by"`

	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== お知らせ =====

type AnnouncementPriority string

const (
	AnnouncementPriorityLow    AnnouncementPriority = "low"
	AnnouncementPriorityNormal AnnouncementPriority = "normal"
	AnnouncementPriorityHigh   AnnouncementPriority = "high"
	AnnouncementPriorityUrgent AnnouncementPriority = "urgent"
)

type HRAnnouncement struct {
	BaseModel
	Title     string               `gorm:"size:200;not null" json:"title"`
	Content   string               `gorm:"type:text;not null" json:"content"`
	Priority  AnnouncementPriority `gorm:"size:20;not null;default:'normal'" json:"priority"`
	AuthorID  uuid.UUID            `gorm:"type:uuid;not null" json:"author_id"`
	IsPublished bool               `gorm:"default:false" json:"is_published"`
	PublishedAt *time.Time         `json:"published_at"`
	ExpiresAt  *time.Time          `json:"expires_at"`

	Author *HREmployee `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// ===== 1on1ミーティング =====

type OneOnOneMeeting struct {
	BaseModel
	ManagerID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"manager_id"`
	EmployeeID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	ScheduledDate time.Time  `gorm:"not null" json:"scheduled_date"`
	Status        string     `gorm:"size:20;not null;default:'scheduled'" json:"status"`
	Frequency     string     `gorm:"size:20;not null;default:'biweekly'" json:"frequency"`
	Agenda        string     `gorm:"type:text" json:"agenda"`
	Notes         string     `gorm:"type:text" json:"notes"`
	Mood          string     `gorm:"size:20" json:"mood"`
	ActionItems   datatypes.JSON `gorm:"type:jsonb" json:"action_items"`

	Manager  *HREmployee `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== スキルマップ =====

type EmployeeSkill struct {
	BaseModel
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	SkillName  string    `gorm:"size:100;not null" json:"skill_name"`
	Category   string    `gorm:"size:50;not null;default:'technical'" json:"category"`
	Level      int       `gorm:"not null;default:1" json:"level"`

	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== 給与 =====

type SalaryRecord struct {
	BaseModel
	EmployeeID  uuid.UUID `gorm:"type:uuid;not null;index" json:"employee_id"`
	BaseSalary  float64   `gorm:"not null" json:"base_salary"`
	Allowances  float64   `gorm:"default:0" json:"allowances"`
	Deductions  float64   `gorm:"default:0" json:"deductions"`
	NetSalary   float64   `gorm:"not null" json:"net_salary"`
	EffectiveDate time.Time `gorm:"type:date;not null" json:"effective_date"`
	Reason      string    `gorm:"size:200" json:"reason"`

	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== オンボーディング =====

type OnboardingStatus string

const (
	OnboardingStatusPending    OnboardingStatus = "pending"
	OnboardingStatusInProgress OnboardingStatus = "in_progress"
	OnboardingStatusCompleted  OnboardingStatus = "completed"
	OnboardingStatusOverdue    OnboardingStatus = "overdue"
)

type Onboarding struct {
	BaseModel
	EmployeeID uuid.UUID        `gorm:"type:uuid;not null;index" json:"employee_id"`
	TemplateID *uuid.UUID       `gorm:"type:uuid" json:"template_id"`
	MentorID   *uuid.UUID       `gorm:"type:uuid" json:"mentor_id"`
	Status     OnboardingStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	StartDate  time.Time        `gorm:"type:date;not null" json:"start_date"`
	Tasks      datatypes.JSON   `gorm:"type:jsonb" json:"tasks"`

	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Mentor   *HREmployee `gorm:"foreignKey:MentorID" json:"mentor,omitempty"`
}

type OnboardingTemplate struct {
	BaseModel
	Name        string         `gorm:"size:200;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Tasks       datatypes.JSON `gorm:"type:jsonb" json:"tasks"`
}

// ===== オフボーディング =====

type OffboardingStatus string

const (
	OffboardingStatusPending    OffboardingStatus = "pending"
	OffboardingStatusInProgress OffboardingStatus = "in_progress"
	OffboardingStatusCompleted  OffboardingStatus = "completed"
)

type Offboarding struct {
	BaseModel
	EmployeeID      uuid.UUID         `gorm:"type:uuid;not null;index" json:"employee_id"`
	Reason          string            `gorm:"size:50;not null;default:'resignation'" json:"reason"`
	Status          OffboardingStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	LastWorkingDate time.Time         `gorm:"type:date;not null" json:"last_working_date"`
	Notes           string            `gorm:"type:text" json:"notes"`
	ExitInterview   string            `gorm:"type:text" json:"exit_interview"`
	Checklist       datatypes.JSON    `gorm:"type:jsonb" json:"checklist"`

	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// ===== サーベイ =====

type SurveyStatus string

const (
	SurveyStatusDraft  SurveyStatus = "draft"
	SurveyStatusActive SurveyStatus = "active"
	SurveyStatusClosed SurveyStatus = "closed"
)

type Survey struct {
	BaseModel
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Type        string         `gorm:"size:30;not null;default:'engagement'" json:"type"`
	Status      SurveyStatus   `gorm:"size:20;not null;default:'draft'" json:"status"`
	IsAnonymous bool           `gorm:"default:true" json:"is_anonymous"`
	Questions   datatypes.JSON `gorm:"type:jsonb" json:"questions"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	PublishedAt *time.Time     `json:"published_at"`
	ClosedAt    *time.Time     `json:"closed_at"`

	Responses []SurveyResponse `gorm:"foreignKey:SurveyID" json:"responses,omitempty"`
}

type SurveyResponse struct {
	BaseModel
	SurveyID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"survey_id"`
	EmployeeID *uuid.UUID     `gorm:"type:uuid" json:"employee_id"`
	Answers    datatypes.JSON `gorm:"type:jsonb" json:"answers"`

	Survey   *Survey     `gorm:"foreignKey:SurveyID" json:"survey,omitempty"`
	Employee *HREmployee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

// HRAutoMigrate はHR関連テーブルのマイグレーションを実行する
func HRAutoMigrate(db *gorm.DB) error {
	// FK制約の自動生成を無効化してマイグレーション（循環参照を避ける）
	prev := db.Config.DisableForeignKeyConstraintWhenMigrating
	db.Config.DisableForeignKeyConstraintWhenMigrating = true
	defer func() { db.Config.DisableForeignKeyConstraintWhenMigrating = prev }()

	return db.AutoMigrate(
		&HRDepartment{},
		&HREmployee{},
		&EvaluationCycle{},
		&Evaluation{},
		&HRGoal{},
		&TrainingProgram{},
		&TrainingEnrollment{},
		&RecruitmentPosition{},
		&Applicant{},
		&HRDocument{},
		&HRAnnouncement{},
		&OneOnOneMeeting{},
		&EmployeeSkill{},
		&SalaryRecord{},
		&Onboarding{},
		&OnboardingTemplate{},
		&Offboarding{},
		&Survey{},
		&SurveyResponse{},
	)
}
