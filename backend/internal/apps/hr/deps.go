package hr

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// UserRepository defines user lookups required by HR services.
type UserRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
}

// AttendanceRepository defines attendance queries required by HR services.
type AttendanceRepository interface {
	CountTodayPresent(ctx context.Context) (int64, error)
	GetMonthlyOvertime(ctx context.Context, start, end time.Time) (int64, error)
	FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error)
	FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Attendance, error)
}

// LeaveRequestRepository defines leave queries required by HR services.
type LeaveRequestRepository interface {
	FindPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error)
}

// OvertimeRequestRepository defines overtime queries required by HR services.
type OvertimeRequestRepository interface {
	GetUserMonthlyOvertime(ctx context.Context, userID uuid.UUID, year, month int) (int64, error)
}

// Repositories groups repository dependencies required by HR app.
type Repositories struct {
	User            UserRepository
	Attendance      AttendanceRepository
	LeaveRequest    LeaveRequestRepository
	OvertimeRequest OvertimeRequestRepository

	HREmployee   HREmployeeRepository
	HRDepartment HRDepartmentRepository
	Evaluation   EvaluationRepository
	Goal         GoalRepository
	Training     TrainingRepository
	Recruitment  RecruitmentRepository
	Document     DocumentRepository
	Announcement AnnouncementRepository
	OneOnOne     OneOnOneRepository
	Skill        SkillRepository
	Salary       SalaryRepository
	Onboarding   OnboardingRepository
	Offboarding  OffboardingRepository
	Survey       SurveyRepository
}

// Deps defines dependencies for HR app services.
type Deps struct {
	Repos  *Repositories
	Config *config.Config
	Logger *logger.Logger
}
