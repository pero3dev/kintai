package model

import "github.com/google/uuid"

// ===== HR社員 =====

type HREmployeeCreateRequest struct {
	EmployeeCode   string     `json:"employee_code" validate:"required"`
	FirstName      string     `json:"first_name" validate:"required"`
	LastName       string     `json:"last_name" validate:"required"`
	Email          string     `json:"email" validate:"required,email"`
	Phone          string     `json:"phone"`
	Position       string     `json:"position"`
	Grade          string     `json:"grade"`
	DepartmentID   *uuid.UUID `json:"department_id"`
	ManagerID      *uuid.UUID `json:"manager_id"`
	EmploymentType string     `json:"employment_type"`
	HireDate       string     `json:"hire_date"`
	BirthDate      string     `json:"birth_date"`
	Address        string     `json:"address"`
	BaseSalary     float64    `json:"base_salary"`
}

type HREmployeeUpdateRequest struct {
	FirstName      *string    `json:"first_name"`
	LastName       *string    `json:"last_name"`
	Email          *string    `json:"email"`
	Phone          *string    `json:"phone"`
	Position       *string    `json:"position"`
	Grade          *string    `json:"grade"`
	DepartmentID   *uuid.UUID `json:"department_id"`
	ManagerID      *uuid.UUID `json:"manager_id"`
	EmploymentType *string    `json:"employment_type"`
	Status         *string    `json:"status"`
	Address        *string    `json:"address"`
	BaseSalary     *float64   `json:"base_salary"`
}

// ===== HR部門 =====

type HRDepartmentCreateRequest struct {
	Name        string     `json:"name" validate:"required"`
	Code        string     `json:"code"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	ManagerID   *uuid.UUID `json:"manager_id"`
	Budget      float64    `json:"budget"`
}

type HRDepartmentUpdateRequest struct {
	Name        *string    `json:"name"`
	Code        *string    `json:"code"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	ManagerID   *uuid.UUID `json:"manager_id"`
	Budget      *float64   `json:"budget"`
}

// ===== 評価 =====

type EvaluationCreateRequest struct {
	EmployeeID     uuid.UUID `json:"employee_id" validate:"required"`
	CycleID        uuid.UUID `json:"cycle_id" validate:"required"`
	SelfScore      *float64  `json:"self_score"`
	ManagerScore   *float64  `json:"manager_score"`
	SelfComment    string    `json:"self_comment"`
	ManagerComment string    `json:"manager_comment"`
	Goals          string    `json:"goals"`
}

type EvaluationUpdateRequest struct {
	SelfScore      *float64 `json:"self_score"`
	ManagerScore   *float64 `json:"manager_score"`
	FinalScore     *float64 `json:"final_score"`
	SelfComment    *string  `json:"self_comment"`
	ManagerComment *string  `json:"manager_comment"`
	Goals          *string  `json:"goals"`
}

type EvaluationCycleCreateRequest struct {
	Name      string `json:"name" validate:"required"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}

// ===== 目標 =====

type HRGoalCreateRequest struct {
	EmployeeID  *uuid.UUID `json:"employee_id"`
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	StartDate   string     `json:"start_date"`
	DueDate     string     `json:"due_date"`
	Weight      int        `json:"weight"`
}

type HRGoalUpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Status      *string `json:"status"`
	DueDate     *string `json:"due_date"`
	Weight      *int    `json:"weight"`
}

// ===== 研修 =====

type TrainingProgramCreateRequest struct {
	Title           string `json:"title" validate:"required"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	InstructorName  string `json:"instructor_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	MaxParticipants int    `json:"max_participants"`
	Location        string `json:"location"`
	IsOnline        bool   `json:"is_online"`
}

type TrainingProgramUpdateRequest struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Category        *string `json:"category"`
	InstructorName  *string `json:"instructor_name"`
	Status          *string `json:"status"`
	MaxParticipants *int    `json:"max_participants"`
	Location        *string `json:"location"`
	IsOnline        *bool   `json:"is_online"`
}

// ===== 採用 =====

type PositionCreateRequest struct {
	Title        string     `json:"title" validate:"required"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Description  string     `json:"description"`
	Requirements string     `json:"requirements"`
	Openings     int        `json:"openings"`
	Location     string     `json:"location"`
	SalaryMin    *float64   `json:"salary_min"`
	SalaryMax    *float64   `json:"salary_max"`
}

type PositionUpdateRequest struct {
	Title        *string    `json:"title"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Description  *string    `json:"description"`
	Requirements *string    `json:"requirements"`
	Status       *string    `json:"status"`
	Openings     *int       `json:"openings"`
	Location     *string    `json:"location"`
	SalaryMin    *float64   `json:"salary_min"`
	SalaryMax    *float64   `json:"salary_max"`
}

type ApplicantCreateRequest struct {
	PositionID uuid.UUID `json:"position_id" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Email      string    `json:"email" validate:"required"`
	Phone      string    `json:"phone"`
	ResumeURL  string    `json:"resume_url"`
	Notes      string    `json:"notes"`
}

type ApplicantStageUpdateRequest struct {
	Stage string `json:"stage" validate:"required"`
}

// ===== 書類 =====

// (ファイルアップロードはmultipart/form-dataで処理)

// ===== お知らせ =====

type AnnouncementCreateRequest struct {
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Priority string `json:"priority"`
}

type AnnouncementUpdateRequest struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	Priority    *string `json:"priority"`
	IsPublished *bool   `json:"is_published"`
}

// ===== 1on1ミーティング =====

type OneOnOneCreateRequest struct {
	EmployeeID    string `json:"employee_id" validate:"required"`
	ScheduledDate string `json:"scheduled_date" validate:"required"`
	Frequency     string `json:"frequency"`
	Agenda        string `json:"agenda"`
}

type OneOnOneUpdateRequest struct {
	Status string `json:"status"`
	Agenda *string `json:"agenda"`
	Notes  *string `json:"notes"`
	Mood   *string `json:"mood"`
}

type ActionItemRequest struct {
	Title string `json:"title" validate:"required"`
}

// ===== スキルマップ =====

type SkillAddRequest struct {
	SkillName string `json:"skill_name" validate:"required"`
	Category  string `json:"category"`
	Level     int    `json:"level"`
}

type SkillUpdateRequest struct {
	Level    *int    `json:"level"`
	Category *string `json:"category"`
}

// ===== 給与 =====

type SalarySimulateRequest struct {
	Grade           string  `json:"grade"`
	Position        string  `json:"position"`
	EvaluationScore string  `json:"evaluation_score"`
	YearsOfService  string  `json:"years_of_service"`
}

// ===== オンボーディング =====

type OnboardingCreateRequest struct {
	EmployeeID string `json:"employee_id" validate:"required"`
	TemplateID string `json:"template_id"`
	StartDate  string `json:"start_date" validate:"required"`
	MentorID   string `json:"mentor_id"`
}

type OnboardingTemplateCreateRequest struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Tasks       interface{} `json:"tasks"`
}

// ===== オフボーディング =====

type OffboardingCreateRequest struct {
	EmployeeID      string `json:"employee_id" validate:"required"`
	Reason          string `json:"reason"`
	LastWorkingDate string `json:"last_working_date" validate:"required"`
	Notes           string `json:"notes"`
}

type OffboardingUpdateRequest struct {
	Status        *string `json:"status"`
	Notes         *string `json:"notes"`
	ExitInterview *string `json:"exit_interview"`
}

// ===== サーベイ =====

type SurveyCreateRequest struct {
	Title       string      `json:"title" validate:"required"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	IsAnonymous bool        `json:"is_anonymous"`
	Questions   interface{} `json:"questions"`
}

type SurveyUpdateRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	Type        *string     `json:"type"`
	IsAnonymous *bool       `json:"is_anonymous"`
	Questions   interface{} `json:"questions"`
}

type SurveyResponseRequest struct {
	Answers interface{} `json:"answers"`
}
