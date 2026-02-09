package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"github.com/your-org/kintai/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// エラー定義
var (
	ErrInvalidCredentials    = errors.New("メールアドレスまたはパスワードが正しくありません")
	ErrUserNotFound          = errors.New("ユーザーが見つかりません")
	ErrEmailAlreadyExists    = errors.New("このメールアドレスは既に登録されています")
	ErrAlreadyClockedIn      = errors.New("既に出勤打刻済みです")
	ErrNotClockedIn          = errors.New("出勤打刻がありません")
	ErrAlreadyClockedOut     = errors.New("既に退勤打刻済みです")
	ErrLeaveNotFound         = errors.New("休暇申請が見つかりません")
	ErrLeaveAlreadyProcessed = errors.New("この休暇申請は既に処理済みです")
	ErrUnauthorized          = errors.New("権限がありません")
)

// Deps はサービスの依存関係
type Deps struct {
	Repos  *repository.Repositories
	Config *config.Config
	Logger *logger.Logger
}

// Services は全サービスを束ねる構造体
type Services struct {
	Auth                 AuthService
	Attendance           AttendanceService
	Leave                LeaveService
	Shift                ShiftService
	User                 UserService
	Department           DepartmentService
	Dashboard            DashboardService
	OvertimeRequest      OvertimeRequestService
	LeaveBalance         LeaveBalanceService
	AttendanceCorrection AttendanceCorrectionService
	Notification         NotificationService
	Project              ProjectService
	TimeEntry            TimeEntryService
	Holiday              HolidayService
	ApprovalFlow         ApprovalFlowService
	Export               ExportService
	// HR
	HREmployee               HREmployeeService
	HRDepartment             HRDepartmentService
	Evaluation               EvaluationService
	Goal                     GoalService
	Training                 TrainingService
	Recruitment              RecruitmentService
	Document                 DocumentService
	Announcement             AnnouncementService
	HRDashboard              HRDashboardService
	AttendanceIntegration    AttendanceIntegrationService
	OrgChart                 OrgChartService
	OneOnOne                 OneOnOneService
	Skill                    SkillService
	Salary                   SalaryService
	Onboarding               OnboardingService
	Offboarding              OffboardingService
	Survey                   SurveyService
	// Expense
	Expense              ExpenseService
	ExpenseComment       ExpenseCommentService
	ExpenseHistory       ExpenseHistoryService
	ExpenseReceipt       ExpenseReceiptService
	ExpenseTemplate      ExpenseTemplateService
	ExpensePolicy        ExpensePolicyService
	ExpenseNotification  ExpenseNotificationService
	ExpenseApprovalFlow  ExpenseApprovalFlowService
}

// NewServices は全サービスを初期化する
func NewServices(deps Deps) *Services {
	notificationSvc := NewNotificationService(deps)
	return &Services{
		Auth:                 NewAuthService(deps),
		Attendance:           NewAttendanceService(deps),
		Leave:                NewLeaveService(deps),
		Shift:                NewShiftService(deps),
		User:                 NewUserService(deps),
		Department:           NewDepartmentService(deps),
		Dashboard:            NewDashboardService(deps),
		OvertimeRequest:      NewOvertimeRequestService(deps, notificationSvc),
		LeaveBalance:         NewLeaveBalanceService(deps),
		AttendanceCorrection: NewAttendanceCorrectionService(deps, notificationSvc),
		Notification:         notificationSvc,
		Project:              NewProjectService(deps),
		TimeEntry:            NewTimeEntryService(deps),
		Holiday:              NewHolidayService(deps),
		ApprovalFlow:         NewApprovalFlowService(deps),
		Export:               NewExportService(deps),
		// HR
		HREmployee:            NewHREmployeeService(deps),
		HRDepartment:          NewHRDepartmentService(deps),
		Evaluation:            NewEvaluationService(deps),
		Goal:                  NewGoalService(deps),
		Training:              NewTrainingService(deps),
		Recruitment:           NewRecruitmentService(deps),
		Document:              NewDocumentService(deps),
		Announcement:          NewAnnouncementService(deps),
		HRDashboard:           NewHRDashboardService(deps),
		AttendanceIntegration: NewAttendanceIntegrationService(deps),
		OrgChart:              NewOrgChartService(deps),
		OneOnOne:              NewOneOnOneService(deps),
		Skill:                 NewSkillService(deps),
		Salary:                NewSalaryService(deps),
		Onboarding:            NewOnboardingService(deps),
		Offboarding:           NewOffboardingService(deps),
		Survey:                NewSurveyService(deps),
		// Expense
		Expense:             NewExpenseService(deps),
		ExpenseComment:      NewExpenseCommentService(deps),
		ExpenseHistory:      NewExpenseHistoryService(deps),
		ExpenseReceipt:      NewExpenseReceiptService(deps),
		ExpenseTemplate:     NewExpenseTemplateService(deps),
		ExpensePolicy:       NewExpensePolicyService(deps),
		ExpenseNotification: NewExpenseNotificationService(deps),
		ExpenseApprovalFlow: NewExpenseApprovalFlowService(deps),
	}
}

// ===== AuthService =====

type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error)
	Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type authService struct {
	deps Deps
}

func NewAuthService(deps Deps) AuthService {
	return &authService{deps: deps}
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	user, err := s.deps.Repos.User.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// JWTアクセストークン生成
	accessToken, err := s.generateToken(user, time.Duration(s.deps.Config.JWTAccessTokenExpiry)*time.Minute)
	if err != nil {
		return nil, err
	}

	// リフレッシュトークン生成
	refreshToken, err := s.generateToken(user, time.Duration(s.deps.Config.JWTRefreshTokenExpiry)*time.Hour)
	if err != nil {
		return nil, err
	}

	// パスワードハッシュをクリアしてから返す
	user.PasswordHash = ""

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.deps.Config.JWTAccessTokenExpiry * 60,
		User:         user,
	}, nil
}

func (s *authService) generateToken(user *model.User, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  string(user.Role),
		"exp":   time.Now().Add(expiry).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.deps.Config.JWTSecretKey))
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	existing, _ := s.deps.Repos.User.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         model.RoleEmployee,
		IsActive:     true,
	}

	if err := s.deps.Repos.User.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
	// リフレッシュトークンの検証
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.deps.Config.JWTSecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidCredentials
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidCredentials
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.deps.Repos.User.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 新しいトークンを生成
	accessToken, err := s.generateToken(user, time.Duration(s.deps.Config.JWTAccessTokenExpiry)*time.Minute)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateToken(user, time.Duration(s.deps.Config.JWTRefreshTokenExpiry)*time.Hour)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.deps.Config.JWTAccessTokenExpiry * 60,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.deps.Repos.RefreshToken.RevokeByUserID(ctx, userID)
}

// ===== AttendanceService =====

type AttendanceService interface {
	ClockIn(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error)
	ClockOut(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error)
	GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error)
	GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error)
	GetTodayStatus(ctx context.Context, userID uuid.UUID) (*model.Attendance, error)
}

type attendanceService struct {
	deps Deps
}

func NewAttendanceService(deps Deps) AttendanceService {
	return &attendanceService{deps: deps}
}

func (s *attendanceService) ClockIn(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)

	existing, _ := s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, today)
	if existing != nil && existing.ClockIn != nil {
		return nil, ErrAlreadyClockedIn
	}

	now := time.Now()
	attendance := &model.Attendance{
		UserID:  userID,
		Date:    today,
		ClockIn: &now,
		Status:  model.AttendanceStatusPresent,
		Note:    req.Note,
	}

	if err := s.deps.Repos.Attendance.Create(ctx, attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func (s *attendanceService) ClockOut(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)

	attendance, err := s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, ErrNotClockedIn
	}

	if attendance.ClockOut != nil {
		return nil, ErrAlreadyClockedOut
	}

	now := time.Now()
	attendance.ClockOut = &now
	if req.Note != "" {
		attendance.Note = req.Note
	}

	// 勤務時間を計算
	if attendance.ClockIn != nil {
		workDuration := now.Sub(*attendance.ClockIn)
		attendance.WorkMinutes = int(workDuration.Minutes()) - attendance.BreakMinutes

		// 8時間超過を残業として計算
		standardMinutes := 8 * 60
		if attendance.WorkMinutes > standardMinutes {
			attendance.OvertimeMinutes = attendance.WorkMinutes - standardMinutes
		}
	}

	if err := s.deps.Repos.Attendance.Update(ctx, attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func (s *attendanceService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
	return s.deps.Repos.Attendance.FindByUserAndDateRange(ctx, userID, start, end, page, pageSize)
}

func (s *attendanceService) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
	return s.deps.Repos.Attendance.GetSummary(ctx, userID, start, end)
}

func (s *attendanceService) GetTodayStatus(ctx context.Context, userID uuid.UUID) (*model.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, today)
}

// ===== LeaveService =====

type LeaveService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error)
	Approve(ctx context.Context, leaveID uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error)
	GetPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error)
}

type leaveService struct {
	deps Deps
}

func NewLeaveService(deps Deps) LeaveService {
	return &leaveService{deps: deps}
}

func (s *leaveService) Create(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("開始日の形式が不正です")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("終了日の形式が不正です")
	}

	leave := &model.LeaveRequest{
		UserID:    userID,
		LeaveType: req.LeaveType,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    req.Reason,
		Status:    model.ApprovalStatusPending,
	}

	if err := s.deps.Repos.LeaveRequest.Create(ctx, leave); err != nil {
		return nil, err
	}

	return leave, nil
}

func (s *leaveService) Approve(ctx context.Context, leaveID uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error) {
	leave, err := s.deps.Repos.LeaveRequest.FindByID(ctx, leaveID)
	if err != nil {
		return nil, ErrLeaveNotFound
	}

	if leave.Status != model.ApprovalStatusPending {
		return nil, ErrLeaveAlreadyProcessed
	}

	now := time.Now()
	leave.Status = req.Status
	leave.ApprovedBy = &approverID
	leave.ApprovedAt = &now
	if req.Status == model.ApprovalStatusRejected {
		leave.RejectedReason = req.RejectedReason
	}

	if err := s.deps.Repos.LeaveRequest.Update(ctx, leave); err != nil {
		return nil, err
	}

	// TODO: メール通知（AWS SES）
	return leave, nil
}

func (s *leaveService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	return s.deps.Repos.LeaveRequest.FindByUserID(ctx, userID, page, pageSize)
}

func (s *leaveService) GetPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	return s.deps.Repos.LeaveRequest.FindPending(ctx, page, pageSize)
}

// ===== ShiftService =====

type ShiftService interface {
	Create(ctx context.Context, req *model.ShiftCreateRequest) (*model.Shift, error)
	BulkCreate(ctx context.Context, req *model.ShiftBulkCreateRequest) error
	GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Shift, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type shiftService struct {
	deps Deps
}

func NewShiftService(deps Deps) ShiftService {
	return &shiftService{deps: deps}
}

func (s *shiftService) Create(ctx context.Context, req *model.ShiftCreateRequest) (*model.Shift, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}

	shift := &model.Shift{
		UserID:    req.UserID,
		Date:      date,
		ShiftType: req.ShiftType,
		Note:      req.Note,
	}

	if err := s.deps.Repos.Shift.Create(ctx, shift); err != nil {
		return nil, err
	}

	return shift, nil
}

func (s *shiftService) BulkCreate(ctx context.Context, req *model.ShiftBulkCreateRequest) error {
	var shifts []model.Shift
	for _, r := range req.Shifts {
		date, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return errors.New("日付の形式が不正です: " + r.Date)
		}
		shifts = append(shifts, model.Shift{
			UserID:    r.UserID,
			Date:      date,
			ShiftType: r.ShiftType,
			Note:      r.Note,
		})
	}
	return s.deps.Repos.Shift.BulkCreate(ctx, shifts)
}

func (s *shiftService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error) {
	return s.deps.Repos.Shift.FindByUserAndDateRange(ctx, userID, start, end)
}

func (s *shiftService) GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Shift, error) {
	return s.deps.Repos.Shift.FindByDateRange(ctx, start, end)
}

func (s *shiftService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Shift.Delete(ctx, id)
}

// ===== UserService =====

type UserService interface {
	Create(ctx context.Context, req *model.UserCreateRequest) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UserUpdateRequest) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	deps Deps
}

func NewUserService(deps Deps) UserService {
	return &userService{deps: deps}
}

func (s *userService) Create(ctx context.Context, req *model.UserCreateRequest) (*model.User, error) {
	// 既存メールチェック
	existing, _ := s.deps.Repos.User.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	if err := s.deps.Repos.User.Create(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.deps.Repos.User.FindByID(ctx, id)
}

func (s *userService) GetAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	return s.deps.Repos.User.FindAll(ctx, page, pageSize)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req *model.UserUpdateRequest) (*model.User, error) {
	user, err := s.deps.Repos.User.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.DepartmentID != nil {
		user.DepartmentID = req.DepartmentID
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := s.deps.Repos.User.Update(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.User.Delete(ctx, id)
}

// ===== DepartmentService =====

type DepartmentService interface {
	Create(ctx context.Context, dept *model.Department) (*model.Department, error)
	GetAll(ctx context.Context) ([]model.Department, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Department, error)
	Update(ctx context.Context, dept *model.Department) (*model.Department, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type departmentService struct {
	deps Deps
}

func NewDepartmentService(deps Deps) DepartmentService {
	return &departmentService{deps: deps}
}

func (s *departmentService) Create(ctx context.Context, dept *model.Department) (*model.Department, error) {
	if err := s.deps.Repos.Department.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) GetAll(ctx context.Context) ([]model.Department, error) {
	return s.deps.Repos.Department.FindAll(ctx)
}

func (s *departmentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Department, error) {
	return s.deps.Repos.Department.FindByID(ctx, id)
}

func (s *departmentService) Update(ctx context.Context, dept *model.Department) (*model.Department, error) {
	if err := s.deps.Repos.Department.Update(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Department.Delete(ctx, id)
}

// ===== DashboardService =====

type DashboardService interface {
	GetStats(ctx context.Context) (*model.DashboardStats, error)
}

type dashboardService struct {
	deps Deps
}

func NewDashboardService(deps Deps) DashboardService {
	return &dashboardService{deps: deps}
}

func (s *dashboardService) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	// 今日の出勤者数
	todayPresent, _ := s.deps.Repos.Attendance.CountTodayPresent(ctx)

	// 総ユーザー数を取得
	_, totalUsers, _ := s.deps.Repos.User.FindAll(ctx, 1, 1)

	// 今日の欠勤者数（総ユーザー数 - 出勤者数）
	todayAbsent := totalUsers - todayPresent
	if todayAbsent < 0 {
		todayAbsent = 0
	}

	// 承認待ちの休暇申請数
	pendingLeaves, _ := s.deps.Repos.LeaveRequest.CountPending(ctx)

	// 今月の残業時間
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)
	monthlyOvertime, _ := s.deps.Repos.Attendance.GetMonthlyOvertime(ctx, monthStart, monthEnd)

	return &model.DashboardStats{
		TodayPresentCount: int(todayPresent),
		TodayAbsentCount:  int(todayAbsent),
		PendingLeaves:     int(pendingLeaves),
		MonthlyOvertime:   int(monthlyOvertime),
	}, nil
}

// ===== OvertimeRequestService =====

type OvertimeRequestService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error)
	Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	GetPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	GetOvertimeAlerts(ctx context.Context) ([]model.OvertimeAlert, error)
}

type overtimeRequestService struct {
	deps            Deps
	notificationSvc NotificationService
}

func NewOvertimeRequestService(deps Deps, notificationSvc NotificationService) OvertimeRequestService {
	return &overtimeRequestService{deps: deps, notificationSvc: notificationSvc}
}

func (s *overtimeRequestService) Create(ctx context.Context, userID uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}
	overtime := &model.OvertimeRequest{
		UserID: userID, Date: date, PlannedMinutes: req.PlannedMinutes,
		Reason: req.Reason, Status: model.OvertimeStatusPending,
	}
	if err := s.deps.Repos.OvertimeRequest.Create(ctx, overtime); err != nil {
		return nil, err
	}
	return overtime, nil
}

func (s *overtimeRequestService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error) {
	overtime, err := s.deps.Repos.OvertimeRequest.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("残業申請が見つかりません")
	}
	if overtime.Status != model.OvertimeStatusPending {
		return nil, errors.New("この残業申請は既に処理済みです")
	}
	now := time.Now()
	overtime.Status = req.Status
	overtime.ApprovedBy = &approverID
	overtime.ApprovedAt = &now
	if req.Status == model.OvertimeStatusRejected {
		overtime.RejectedReason = req.RejectedReason
	}
	if err := s.deps.Repos.OvertimeRequest.Update(ctx, overtime); err != nil {
		return nil, err
	}
	// 通知送信
	notifType := model.NotificationTypeLeaveApproved
	title := "残業申請が承認されました"
	if req.Status == model.OvertimeStatusRejected {
		notifType = model.NotificationTypeLeaveRejected
		title = "残業申請が却下されました"
	}
	_ = s.notificationSvc.Send(ctx, overtime.UserID, notifType, title,
		fmt.Sprintf("%s の残業申請が%sされました", overtime.Date.Format("2006-01-02"),
			map[model.OvertimeRequestStatus]string{model.OvertimeStatusApproved: "承認", model.OvertimeStatusRejected: "却下"}[req.Status]))
	return overtime, nil
}

func (s *overtimeRequestService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	return s.deps.Repos.OvertimeRequest.FindByUserID(ctx, userID, page, pageSize)
}

func (s *overtimeRequestService) GetPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	return s.deps.Repos.OvertimeRequest.FindPending(ctx, page, pageSize)
}

func (s *overtimeRequestService) GetOvertimeAlerts(ctx context.Context) ([]model.OvertimeAlert, error) {
	users, _, _ := s.deps.Repos.User.FindAll(ctx, 1, 1000)
	now := time.Now()
	var alerts []model.OvertimeAlert
	for _, u := range users {
		monthly, _ := s.deps.Repos.OvertimeRequest.GetUserMonthlyOvertime(ctx, u.ID, now.Year(), int(now.Month()))
		yearly, _ := s.deps.Repos.OvertimeRequest.GetUserYearlyOvertime(ctx, u.ID, now.Year())
		monthlyHours := float64(monthly) / 60.0
		yearlyHours := float64(yearly) / 60.0
		// 36協定: 月45時間、年360時間
		if monthlyHours > 35 || yearlyHours > 300 {
			alerts = append(alerts, model.OvertimeAlert{
				UserID: u.ID, UserName: u.LastName + " " + u.FirstName,
				MonthlyOvertimeHours: monthlyHours, YearlyOvertimeHours: yearlyHours,
				MonthlyLimitHours: 45, YearlyLimitHours: 360,
				IsMonthlyExceeded: monthlyHours > 45, IsYearlyExceeded: yearlyHours > 360,
			})
		}
	}
	return alerts, nil
}

// ===== LeaveBalanceService =====

type LeaveBalanceService interface {
	GetByUser(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error)
	SetBalance(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error
	DeductBalance(ctx context.Context, userID uuid.UUID, leaveType model.LeaveType, days float64) error
	InitializeForUser(ctx context.Context, userID uuid.UUID, fiscalYear int) error
}

type leaveBalanceService struct{ deps Deps }

func NewLeaveBalanceService(deps Deps) LeaveBalanceService {
	return &leaveBalanceService{deps: deps}
}

func (s *leaveBalanceService) GetByUser(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error) {
	balances, err := s.deps.Repos.LeaveBalance.FindByUserAndYear(ctx, userID, fiscalYear)
	if err != nil {
		return nil, err
	}
	var responses []model.LeaveBalanceResponse
	for _, b := range balances {
		responses = append(responses, model.LeaveBalanceResponse{
			LeaveType: b.LeaveType, TotalDays: b.TotalDays, UsedDays: b.UsedDays,
			RemainingDays: b.TotalDays + b.CarriedOver - b.UsedDays,
			CarriedOver:   b.CarriedOver, FiscalYear: b.FiscalYear,
		})
	}
	return responses, nil
}

func (s *leaveBalanceService) SetBalance(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error {
	balance, err := s.deps.Repos.LeaveBalance.FindByUserYearAndType(ctx, userID, fiscalYear, leaveType)
	if err != nil {
		balance = &model.LeaveBalance{UserID: userID, FiscalYear: fiscalYear, LeaveType: leaveType}
	}
	if req.TotalDays != nil {
		balance.TotalDays = *req.TotalDays
	}
	if req.CarriedOver != nil {
		balance.CarriedOver = *req.CarriedOver
	}
	return s.deps.Repos.LeaveBalance.Upsert(ctx, balance)
}

func (s *leaveBalanceService) DeductBalance(ctx context.Context, userID uuid.UUID, leaveType model.LeaveType, days float64) error {
	fiscalYear := time.Now().Year()
	balance, err := s.deps.Repos.LeaveBalance.FindByUserYearAndType(ctx, userID, fiscalYear, leaveType)
	if err != nil {
		return errors.New("有給残日数が設定されていません")
	}
	remaining := balance.TotalDays + balance.CarriedOver - balance.UsedDays
	if remaining < days {
		return fmt.Errorf("有給残日数が不足しています（残り: %.1f日）", remaining)
	}
	balance.UsedDays += days
	return s.deps.Repos.LeaveBalance.Update(ctx, balance)
}

func (s *leaveBalanceService) InitializeForUser(ctx context.Context, userID uuid.UUID, fiscalYear int) error {
	defaultDays := map[model.LeaveType]float64{
		model.LeaveTypePaid: 10, model.LeaveTypeSick: 5, model.LeaveTypeSpecial: 3,
	}
	for lt, days := range defaultDays {
		balance := &model.LeaveBalance{
			UserID: userID, FiscalYear: fiscalYear, LeaveType: lt,
			TotalDays: days, UsedDays: 0, CarriedOver: 0,
		}
		if err := s.deps.Repos.LeaveBalance.Upsert(ctx, balance); err != nil {
			return err
		}
	}
	return nil
}

// ===== AttendanceCorrectionService =====

type AttendanceCorrectionService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error)
	Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
	GetPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
}

type attendanceCorrectionService struct {
	deps            Deps
	notificationSvc NotificationService
}

func NewAttendanceCorrectionService(deps Deps, notificationSvc NotificationService) AttendanceCorrectionService {
	return &attendanceCorrectionService{deps: deps, notificationSvc: notificationSvc}
}

func (s *attendanceCorrectionService) Create(ctx context.Context, userID uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}
	correction := &model.AttendanceCorrection{
		UserID: userID, Date: date, Reason: req.Reason,
		Status: model.CorrectionStatusPending,
	}
	// 既存の出勤データを取得
	existing, _ := s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, date)
	if existing != nil {
		correction.AttendanceID = &existing.ID
		correction.OriginalClockIn = existing.ClockIn
		correction.OriginalClockOut = existing.ClockOut
	}
	if req.CorrectedClockIn != nil {
		t, err := time.Parse("2006-01-02T15:04:05", *req.CorrectedClockIn)
		if err != nil {
			t, err = time.Parse("15:04", *req.CorrectedClockIn)
			if err == nil {
				t = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
			}
		}
		if err == nil {
			correction.CorrectedClockIn = &t
		}
	}
	if req.CorrectedClockOut != nil {
		t, err := time.Parse("2006-01-02T15:04:05", *req.CorrectedClockOut)
		if err != nil {
			t, err = time.Parse("15:04", *req.CorrectedClockOut)
			if err == nil {
				t = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
			}
		}
		if err == nil {
			correction.CorrectedClockOut = &t
		}
	}
	if err := s.deps.Repos.AttendanceCorrection.Create(ctx, correction); err != nil {
		return nil, err
	}
	return correction, nil
}

func (s *attendanceCorrectionService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error) {
	correction, err := s.deps.Repos.AttendanceCorrection.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("修正申請が見つかりません")
	}
	if correction.Status != model.CorrectionStatusPending {
		return nil, errors.New("この修正申請は既に処理済みです")
	}
	now := time.Now()
	correction.Status = req.Status
	correction.ApprovedBy = &approverID
	correction.ApprovedAt = &now
	if req.Status == model.CorrectionStatusRejected {
		correction.RejectedReason = req.RejectedReason
	}
	// 承認時：勤怠データを修正
	if req.Status == model.CorrectionStatusApproved {
		if correction.AttendanceID != nil {
			att, _ := s.deps.Repos.Attendance.FindByID(ctx, *correction.AttendanceID)
			if att != nil {
				if correction.CorrectedClockIn != nil {
					att.ClockIn = correction.CorrectedClockIn
				}
				if correction.CorrectedClockOut != nil {
					att.ClockOut = correction.CorrectedClockOut
				}
				if att.ClockIn != nil && att.ClockOut != nil {
					workDuration := att.ClockOut.Sub(*att.ClockIn)
					att.WorkMinutes = int(workDuration.Minutes()) - att.BreakMinutes
					if att.WorkMinutes > 480 {
						att.OvertimeMinutes = att.WorkMinutes - 480
					} else {
						att.OvertimeMinutes = 0
					}
				}
				_ = s.deps.Repos.Attendance.Update(ctx, att)
			}
		} else {
			// 新規作成
			att := &model.Attendance{
				UserID: correction.UserID, Date: correction.Date,
				ClockIn: correction.CorrectedClockIn, ClockOut: correction.CorrectedClockOut,
				Status: model.AttendanceStatusPresent,
			}
			if att.ClockIn != nil && att.ClockOut != nil {
				workDuration := att.ClockOut.Sub(*att.ClockIn)
				att.WorkMinutes = int(workDuration.Minutes())
				if att.WorkMinutes > 480 {
					att.OvertimeMinutes = att.WorkMinutes - 480
				}
			}
			_ = s.deps.Repos.Attendance.Create(ctx, att)
		}
	}
	if err := s.deps.Repos.AttendanceCorrection.Update(ctx, correction); err != nil {
		return nil, err
	}
	// 通知送信
	title := "勤怠修正申請が承認されました"
	if req.Status == model.CorrectionStatusRejected {
		title = "勤怠修正申請が却下されました"
	}
	_ = s.notificationSvc.Send(ctx, correction.UserID, model.NotificationTypeCorrectionResult, title,
		fmt.Sprintf("%s の勤怠修正申請が処理されました", correction.Date.Format("2006-01-02")))
	return correction, nil
}

func (s *attendanceCorrectionService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	return s.deps.Repos.AttendanceCorrection.FindByUserID(ctx, userID, page, pageSize)
}

func (s *attendanceCorrectionService) GetPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	return s.deps.Repos.AttendanceCorrection.FindPending(ctx, page, pageSize)
}

// ===== NotificationService =====

type NotificationService interface {
	Send(ctx context.Context, userID uuid.UUID, notifType model.NotificationType, title, message string) error
	GetByUser(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationService struct{ deps Deps }

func NewNotificationService(deps Deps) NotificationService {
	return &notificationService{deps: deps}
}

func (s *notificationService) Send(ctx context.Context, userID uuid.UUID, notifType model.NotificationType, title, message string) error {
	notification := &model.Notification{
		UserID: userID, Type: notifType, Title: title, Message: message,
	}
	return s.deps.Repos.Notification.Create(ctx, notification)
}

func (s *notificationService) GetByUser(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
	return s.deps.Repos.Notification.FindByUserID(ctx, userID, isRead, page, pageSize)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Notification.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.deps.Repos.Notification.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.deps.Repos.Notification.CountUnread(ctx, userID)
}

func (s *notificationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Notification.Delete(ctx, id)
}

// ===== ProjectService =====

type ProjectService interface {
	Create(ctx context.Context, req *model.ProjectCreateRequest) (*model.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	GetAll(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *model.ProjectUpdateRequest) (*model.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type projectService struct{ deps Deps }

func NewProjectService(deps Deps) ProjectService { return &projectService{deps: deps} }

func (s *projectService) Create(ctx context.Context, req *model.ProjectCreateRequest) (*model.Project, error) {
	project := &model.Project{
		Name: req.Name, Code: req.Code, Description: req.Description,
		Status: model.ProjectStatusActive, ManagerID: req.ManagerID, BudgetHours: req.BudgetHours,
	}
	if err := s.deps.Repos.Project.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	return s.deps.Repos.Project.FindByID(ctx, id)
}

func (s *projectService) GetAll(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error) {
	return s.deps.Repos.Project.FindAll(ctx, status, page, pageSize)
}

func (s *projectService) Update(ctx context.Context, id uuid.UUID, req *model.ProjectUpdateRequest) (*model.Project, error) {
	project, err := s.deps.Repos.Project.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("プロジェクトが見つかりません")
	}
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}
	if req.ManagerID != nil {
		project.ManagerID = req.ManagerID
	}
	if req.BudgetHours != nil {
		project.BudgetHours = req.BudgetHours
	}
	if err := s.deps.Repos.Project.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Project.Delete(ctx, id)
}

// ===== TimeEntryService =====

type TimeEntryService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.TimeEntryCreate) (*model.TimeEntry, error)
	GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error)
	GetByProjectAndDateRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error)
	Update(ctx context.Context, id uuid.UUID, req *model.TimeEntryUpdate) (*model.TimeEntry, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error)
}

type timeEntryService struct{ deps Deps }

func NewTimeEntryService(deps Deps) TimeEntryService { return &timeEntryService{deps: deps} }

func (s *timeEntryService) Create(ctx context.Context, userID uuid.UUID, req *model.TimeEntryCreate) (*model.TimeEntry, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}
	entry := &model.TimeEntry{
		UserID: userID, ProjectID: req.ProjectID, Date: date,
		Minutes: req.Minutes, Description: req.Description,
	}
	if err := s.deps.Repos.TimeEntry.Create(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *timeEntryService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	return s.deps.Repos.TimeEntry.FindByUserAndDateRange(ctx, userID, start, end)
}

func (s *timeEntryService) GetByProjectAndDateRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	return s.deps.Repos.TimeEntry.FindByProjectAndDateRange(ctx, projectID, start, end)
}

func (s *timeEntryService) Update(ctx context.Context, id uuid.UUID, req *model.TimeEntryUpdate) (*model.TimeEntry, error) {
	entry, err := s.deps.Repos.TimeEntry.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("工数記録が見つかりません")
	}
	if req.Minutes != nil {
		entry.Minutes = *req.Minutes
	}
	if req.Description != nil {
		entry.Description = *req.Description
	}
	if err := s.deps.Repos.TimeEntry.Update(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *timeEntryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.TimeEntry.Delete(ctx, id)
}

func (s *timeEntryService) GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
	return s.deps.Repos.TimeEntry.GetProjectSummary(ctx, start, end)
}

// ===== HolidayService =====

type HolidayService interface {
	Create(ctx context.Context, req *model.HolidayCreateRequest) (*model.Holiday, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error)
	GetByYear(ctx context.Context, year int) ([]model.Holiday, error)
	Update(ctx context.Context, id uuid.UUID, req *model.HolidayUpdateRequest) (*model.Holiday, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetCalendar(ctx context.Context, year, month int) ([]model.CalendarDay, error)
	GetWorkingDays(ctx context.Context, start, end time.Time) (*model.WorkingDaysSummary, error)
}

type holidayService struct{ deps Deps }

func NewHolidayService(deps Deps) HolidayService { return &holidayService{deps: deps} }

func (s *holidayService) Create(ctx context.Context, req *model.HolidayCreateRequest) (*model.Holiday, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}
	holiday := &model.Holiday{
		Date: date, Name: req.Name, HolidayType: req.HolidayType, IsRecurring: req.IsRecurring,
	}
	if err := s.deps.Repos.Holiday.Create(ctx, holiday); err != nil {
		return nil, err
	}
	return holiday, nil
}

func (s *holidayService) GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error) {
	return s.deps.Repos.Holiday.FindByDateRange(ctx, start, end)
}

func (s *holidayService) GetByYear(ctx context.Context, year int) ([]model.Holiday, error) {
	return s.deps.Repos.Holiday.FindByYear(ctx, year)
}

func (s *holidayService) Update(ctx context.Context, id uuid.UUID, req *model.HolidayUpdateRequest) (*model.Holiday, error) {
	holiday, err := s.deps.Repos.Holiday.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("祝日が見つかりません")
	}
	if req.Name != nil {
		holiday.Name = *req.Name
	}
	if req.HolidayType != nil {
		holiday.HolidayType = *req.HolidayType
	}
	if req.IsRecurring != nil {
		holiday.IsRecurring = *req.IsRecurring
	}
	if err := s.deps.Repos.Holiday.Update(ctx, holiday); err != nil {
		return nil, err
	}
	return holiday, nil
}

func (s *holidayService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Holiday.Delete(ctx, id)
}

func (s *holidayService) GetCalendar(ctx context.Context, year, month int) ([]model.CalendarDay, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, -1)
	holidays, err := s.deps.Repos.Holiday.FindByDateRange(ctx, start, end)
	if err != nil {
		return nil, err
	}
	holidayMap := make(map[string]model.Holiday)
	for _, h := range holidays {
		holidayMap[h.Date.Format("2006-01-02")] = h
	}
	var days []model.CalendarDay
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		isWeekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
		day := model.CalendarDay{Date: dateStr, IsWeekend: isWeekend}
		if h, ok := holidayMap[dateStr]; ok {
			day.IsHoliday = true
			day.HolidayName = h.Name
			ht := h.HolidayType
			day.HolidayType = &ht
		}
		days = append(days, day)
	}
	return days, nil
}

func (s *holidayService) GetWorkingDays(ctx context.Context, start, end time.Time) (*model.WorkingDaysSummary, error) {
	holidays, err := s.deps.Repos.Holiday.FindByDateRange(ctx, start, end)
	if err != nil {
		return nil, err
	}
	holidaySet := make(map[string]bool)
	for _, h := range holidays {
		holidaySet[h.Date.Format("2006-01-02")] = true
	}
	summary := &model.WorkingDaysSummary{}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		summary.TotalDays++
		isWeekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
		isHoliday := holidaySet[d.Format("2006-01-02")]
		if isWeekend {
			summary.Weekends++
		} else if isHoliday {
			summary.Holidays++
		} else {
			summary.WorkingDays++
		}
	}
	return summary, nil
}

// ===== ApprovalFlowService =====

type ApprovalFlowService interface {
	Create(ctx context.Context, req *model.ApprovalFlowCreateRequest) (*model.ApprovalFlow, error)
	GetAll(ctx context.Context) ([]model.ApprovalFlow, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error)
	GetByType(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error)
	Update(ctx context.Context, id uuid.UUID, req *model.ApprovalFlowUpdateRequest) (*model.ApprovalFlow, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type approvalFlowService struct{ deps Deps }

func NewApprovalFlowService(deps Deps) ApprovalFlowService {
	return &approvalFlowService{deps: deps}
}

func (s *approvalFlowService) Create(ctx context.Context, req *model.ApprovalFlowCreateRequest) (*model.ApprovalFlow, error) {
	flow := &model.ApprovalFlow{Name: req.Name, FlowType: req.FlowType, IsActive: true}
	if err := s.deps.Repos.ApprovalFlow.Create(ctx, flow); err != nil {
		return nil, err
	}
	var steps []model.ApprovalStep
	for _, sr := range req.Steps {
		steps = append(steps, model.ApprovalStep{
			FlowID: flow.ID, StepOrder: sr.StepOrder, StepType: sr.StepType,
			ApproverRole: sr.ApproverRole, ApproverID: sr.ApproverID,
		})
	}
	if err := s.deps.Repos.ApprovalFlow.CreateSteps(ctx, steps); err != nil {
		return nil, err
	}
	return s.deps.Repos.ApprovalFlow.FindByID(ctx, flow.ID)
}

func (s *approvalFlowService) GetAll(ctx context.Context) ([]model.ApprovalFlow, error) {
	return s.deps.Repos.ApprovalFlow.FindAll(ctx)
}

func (s *approvalFlowService) GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error) {
	return s.deps.Repos.ApprovalFlow.FindByID(ctx, id)
}

func (s *approvalFlowService) GetByType(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error) {
	return s.deps.Repos.ApprovalFlow.FindByType(ctx, flowType)
}

func (s *approvalFlowService) Update(ctx context.Context, id uuid.UUID, req *model.ApprovalFlowUpdateRequest) (*model.ApprovalFlow, error) {
	flow, err := s.deps.Repos.ApprovalFlow.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("承認フローが見つかりません")
	}
	if req.Name != nil {
		flow.Name = *req.Name
	}
	if req.IsActive != nil {
		flow.IsActive = *req.IsActive
	}
	if err := s.deps.Repos.ApprovalFlow.Update(ctx, flow); err != nil {
		return nil, err
	}
	if req.Steps != nil {
		_ = s.deps.Repos.ApprovalFlow.DeleteStepsByFlowID(ctx, id)
		var steps []model.ApprovalStep
		for _, sr := range req.Steps {
			steps = append(steps, model.ApprovalStep{
				FlowID: id, StepOrder: sr.StepOrder, StepType: sr.StepType,
				ApproverRole: sr.ApproverRole, ApproverID: sr.ApproverID,
			})
		}
		if err := s.deps.Repos.ApprovalFlow.CreateSteps(ctx, steps); err != nil {
			return nil, err
		}
	}
	return s.deps.Repos.ApprovalFlow.FindByID(ctx, id)
}

func (s *approvalFlowService) Delete(ctx context.Context, id uuid.UUID) error {
	_ = s.deps.Repos.ApprovalFlow.DeleteStepsByFlowID(ctx, id)
	return s.deps.Repos.ApprovalFlow.Delete(ctx, id)
}

// ===== ExportService =====

type ExportService interface {
	ExportAttendanceCSV(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error)
	ExportLeavesCSV(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error)
	ExportOvertimeCSV(ctx context.Context, start, end time.Time) ([]byte, error)
	ExportProjectsCSV(ctx context.Context, start, end time.Time) ([]byte, error)
}

type exportService struct{ deps Deps }

func NewExportService(deps Deps) ExportService { return &exportService{deps: deps} }

func (s *exportService) ExportAttendanceCSV(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
	var buf bytes.Buffer
	// BOM for Excel
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"日付", "ユーザー", "出勤時刻", "退勤時刻", "勤務時間(分)", "残業時間(分)", "ステータス", "メモ"})

	if userID != nil {
		attendances, _, _ := s.deps.Repos.Attendance.FindByUserAndDateRange(ctx, *userID, start, end, 1, 10000)
		user, _ := s.deps.Repos.User.FindByID(ctx, *userID)
		userName := ""
		if user != nil {
			userName = user.LastName + " " + user.FirstName
		}
		for _, a := range attendances {
			clockIn, clockOut := "", ""
			if a.ClockIn != nil {
				clockIn = a.ClockIn.Format("15:04:05")
			}
			if a.ClockOut != nil {
				clockOut = a.ClockOut.Format("15:04:05")
			}
			_ = writer.Write([]string{
				a.Date.Format("2006-01-02"), userName, clockIn, clockOut,
				fmt.Sprintf("%d", a.WorkMinutes), fmt.Sprintf("%d", a.OvertimeMinutes),
				string(a.Status), a.Note,
			})
		}
	} else {
		attendances, _ := s.deps.Repos.Attendance.FindByDateRange(ctx, start, end)
		for _, a := range attendances {
			userName := ""
			if a.User != nil {
				userName = a.User.LastName + " " + a.User.FirstName
			}
			clockIn, clockOut := "", ""
			if a.ClockIn != nil {
				clockIn = a.ClockIn.Format("15:04:05")
			}
			if a.ClockOut != nil {
				clockOut = a.ClockOut.Format("15:04:05")
			}
			_ = writer.Write([]string{
				a.Date.Format("2006-01-02"), userName, clockIn, clockOut,
				fmt.Sprintf("%d", a.WorkMinutes), fmt.Sprintf("%d", a.OvertimeMinutes),
				string(a.Status), a.Note,
			})
		}
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *exportService) ExportLeavesCSV(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"申請者", "休暇種別", "開始日", "終了日", "理由", "ステータス", "承認者"})

	// Get all users' leaves
	users, _, _ := s.deps.Repos.User.FindAll(ctx, 1, 10000)
	for _, u := range users {
		if userID != nil && u.ID != *userID {
			continue
		}
		leaves, _, _ := s.deps.Repos.LeaveRequest.FindByUserID(ctx, u.ID, 1, 10000)
		for _, l := range leaves {
			if l.StartDate.Before(start) || l.EndDate.After(end) {
				continue
			}
			approver := ""
			if l.Approver != nil {
				approver = l.Approver.LastName + " " + l.Approver.FirstName
			}
			_ = writer.Write([]string{
				u.LastName + " " + u.FirstName, string(l.LeaveType),
				l.StartDate.Format("2006-01-02"), l.EndDate.Format("2006-01-02"),
				l.Reason, string(l.Status), approver,
			})
		}
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *exportService) ExportOvertimeCSV(ctx context.Context, start, end time.Time) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"ユーザー", "月間残業時間(時間)", "年間残業時間(時間)", "月間上限(時間)", "年間上限(時間)", "超過警告"})

	users, _, _ := s.deps.Repos.User.FindAll(ctx, 1, 10000)
	now := time.Now()
	for _, u := range users {
		monthly, _ := s.deps.Repos.OvertimeRequest.GetUserMonthlyOvertime(ctx, u.ID, now.Year(), int(now.Month()))
		yearly, _ := s.deps.Repos.OvertimeRequest.GetUserYearlyOvertime(ctx, u.ID, now.Year())
		monthlyH := float64(monthly) / 60.0
		yearlyH := float64(yearly) / 60.0
		warning := ""
		if monthlyH > 45 {
			warning = "月間上限超過"
		}
		if yearlyH > 360 {
			if warning != "" {
				warning += ", "
			}
			warning += "年間上限超過"
		}
		_ = writer.Write([]string{
			u.LastName + " " + u.FirstName,
			fmt.Sprintf("%.1f", monthlyH), fmt.Sprintf("%.1f", yearlyH),
			"45", "360", warning,
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *exportService) ExportProjectsCSV(ctx context.Context, start, end time.Time) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"プロジェクトコード", "プロジェクト名", "合計工数(時間)", "予算(時間)", "メンバー数"})

	summaries, _ := s.deps.Repos.TimeEntry.GetProjectSummary(ctx, start, end)
	for _, ps := range summaries {
		budget := "-"
		if ps.BudgetHours != nil {
			budget = fmt.Sprintf("%.1f", *ps.BudgetHours)
		}
		_ = writer.Write([]string{
			ps.ProjectCode, ps.ProjectName,
			fmt.Sprintf("%.1f", ps.TotalHours), budget,
			fmt.Sprintf("%d", ps.MemberCount),
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}
