package service

import (
	"context"
	"errors"
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
	ErrInvalidCredentials = errors.New("メールアドレスまたはパスワードが正しくありません")
	ErrUserNotFound       = errors.New("ユーザーが見つかりません")
	ErrEmailAlreadyExists = errors.New("このメールアドレスは既に登録されています")
	ErrAlreadyClockedIn   = errors.New("既に出勤打刻済みです")
	ErrNotClockedIn       = errors.New("出勤打刻がありません")
	ErrAlreadyClockedOut  = errors.New("既に退勤打刻済みです")
	ErrLeaveNotFound      = errors.New("休暇申請が見つかりません")
	ErrLeaveAlreadyProcessed = errors.New("この休暇申請は既に処理済みです")
	ErrUnauthorized       = errors.New("権限がありません")
)

// Deps はサービスの依存関係
type Deps struct {
	Repos  *repository.Repositories
	Config *config.Config
	Logger *logger.Logger
}

// Services は全サービスを束ねる構造体
type Services struct {
	Auth       AuthService
	Attendance AttendanceService
	Leave      LeaveService
	Shift      ShiftService
	User       UserService
	Department DepartmentService
	Dashboard  DashboardService
}

// NewServices は全サービスを初期化する
func NewServices(deps Deps) *Services {
	return &Services{
		Auth:       NewAuthService(deps),
		Attendance: NewAttendanceService(deps),
		Leave:      NewLeaveService(deps),
		Shift:      NewShiftService(deps),
		User:       NewUserService(deps),
		Department: NewDepartmentService(deps),
		Dashboard:  NewDashboardService(deps),
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
		"sub":  user.ID.String(),
		"email": user.Email,
		"role": string(user.Role),
		"exp":  time.Now().Add(expiry).Unix(),
		"iat":  time.Now().Unix(),
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
