package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"github.com/your-org/kintai/backend/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// generateTestToken はテスト用にJWTトークンを生成するヘルパー関数
func generateTestToken(user *model.User, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  string(user.Role),
		"exp":   time.Now().Add(expiry).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func setupTestDeps(t *testing.T) Deps {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")
	
	return Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			User:         mocks.NewMockUserRepository(),
			Attendance:   mocks.NewMockAttendanceRepository(),
			LeaveRequest: mocks.NewMockLeaveRequestRepository(),
			Shift:        mocks.NewMockShiftRepository(),
			Department:   mocks.NewMockDepartmentRepository(),
			RefreshToken: mocks.NewMockRefreshTokenRepository(),
		},
	}
}

// ===== AuthService Tests =====

func TestAuthService_Login_Success(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// テストユーザーを作成
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := uuid.New()
	testUser := &model.User{
		BaseModel:    model.BaseModel{ID: userID},
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FirstName:    "Test",
		LastName:     "User",
		Role:         model.RoleEmployee,
		IsActive:     true,
	}
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.Users[userID] = testUser
	mockUserRepo.UsersByEmail[testUser.Email] = testUser

	// ログイン
	resp, err := authService.Login(ctx, &model.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if resp.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
	if resp.User == nil {
		t.Error("User should not be nil")
	}
	if resp.User.PasswordHash != "" {
		t.Error("Password hash should be cleared")
	}
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	_, err := authService.Login(ctx, &model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// テストユーザーを作成
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := uuid.New()
	testUser := &model.User{
		BaseModel:    model.BaseModel{ID: userID},
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.Users[userID] = testUser
	mockUserRepo.UsersByEmail[testUser.Email] = testUser

	_, err := authService.Login(ctx, &model.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	user, err := authService.Register(ctx, &model.RegisterRequest{
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	})

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user == nil {
		t.Fatal("User should not be nil")
	}
	if user.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", user.Email)
	}
	if user.Role != model.RoleEmployee {
		t.Errorf("Expected role 'employee', got '%s'", user.Role)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// 既存ユーザーを作成
	existingUserID := uuid.New()
	existingUser := &model.User{
		BaseModel: model.BaseModel{ID: existingUserID},
		Email:     "existing@example.com",
	}
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.Users[existingUserID] = existingUser
	mockUserRepo.UsersByEmail[existingUser.Email] = existingUser

	_, err := authService.Register(ctx, &model.RegisterRequest{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	})

	if err != ErrEmailAlreadyExists {
		t.Errorf("Expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// テストユーザーを作成
	userID := uuid.New()
	testUser := &model.User{
		BaseModel: model.BaseModel{ID: userID},
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	}
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.Users[userID] = testUser

	// リフレッシュトークンを生成
	refreshToken, err := generateTestToken(testUser, deps.Config.JWTSecretKey, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// リフレッシュ
	resp, err := authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	_, err := authService.RefreshToken(ctx, "invalid-token")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_WrongSigningMethod(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// Create token with different signing method (RS256 instead of HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": "employee",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := authService.RefreshToken(ctx, tokenString)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// Create token for non-existent user
	nonExistentUserID := uuid.New()
	testUser := &model.User{
		BaseModel: model.BaseModel{ID: nonExistentUserID},
		Email:     "test@example.com",
		Role:      model.RoleEmployee,
	}
	refreshToken, _ := generateTestToken(testUser, deps.Config.JWTSecretKey, time.Hour)

	_, err := authService.RefreshToken(ctx, refreshToken)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestAuthService_RefreshToken_InvalidUserID(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// Create token with invalid user ID (not a UUID)
	claims := jwt.MapClaims{
		"sub":  "not-a-uuid",
		"role": "employee",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(deps.Config.JWTSecretKey))

	_, err := authService.RefreshToken(ctx, tokenString)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_NoSubClaim(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	// Create token without sub claim
	claims := jwt.MapClaims{
		"role": "employee",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(deps.Config.JWTSecretKey))

	_, err := authService.RefreshToken(ctx, tokenString)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	deps := setupTestDeps(t)
	authService := NewAuthService(deps)
	ctx := context.Background()

	userID := uuid.New()
	err := authService.Logout(ctx, userID)
	if err != nil {
		t.Errorf("Logout failed: %v", err)
	}
}

// ===== AttendanceService Tests =====

func TestAttendanceService_ClockIn_Success(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	attendance, err := attService.ClockIn(ctx, userID, &model.ClockInRequest{Note: "Test note"})

	if err != nil {
		t.Fatalf("ClockIn failed: %v", err)
	}
	if attendance == nil {
		t.Fatal("Attendance should not be nil")
	}
	if attendance.ClockIn == nil {
		t.Error("ClockIn time should not be nil")
	}
	if attendance.Note != "Test note" {
		t.Errorf("Expected note 'Test note', got '%s'", attendance.Note)
	}
}

func TestAttendanceService_ClockIn_AlreadyClockedIn(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	
	// 最初の出勤打刻
	_, err := attService.ClockIn(ctx, userID, &model.ClockInRequest{})
	if err != nil {
		t.Fatalf("First ClockIn failed: %v", err)
	}

	// 2回目の出勤打刻
	_, err = attService.ClockIn(ctx, userID, &model.ClockInRequest{})
	if err != ErrAlreadyClockedIn {
		t.Errorf("Expected ErrAlreadyClockedIn, got %v", err)
	}
}

func TestAttendanceService_ClockOut_Success(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	
	// 出勤打刻
	_, err := attService.ClockIn(ctx, userID, &model.ClockInRequest{})
	if err != nil {
		t.Fatalf("ClockIn failed: %v", err)
	}

	// 退勤打刻
	attendance, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{Note: "Leaving"})
	if err != nil {
		t.Fatalf("ClockOut failed: %v", err)
	}
	if attendance.ClockOut == nil {
		t.Error("ClockOut time should not be nil")
	}
}

func TestAttendanceService_ClockOut_NotClockedIn(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{})
	if err != ErrNotClockedIn {
		t.Errorf("Expected ErrNotClockedIn, got %v", err)
	}
}

func TestAttendanceService_ClockOut_AlreadyClockedOut(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	
	// 出勤・退勤打刻
	_, _ = attService.ClockIn(ctx, userID, &model.ClockInRequest{})
	_, _ = attService.ClockOut(ctx, userID, &model.ClockOutRequest{})

	// 2回目の退勤打刻
	_, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{})
	if err != ErrAlreadyClockedOut {
		t.Errorf("Expected ErrAlreadyClockedOut, got %v", err)
	}
}

func TestAttendanceService_GetByUserAndDateRange(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = attService.ClockIn(ctx, userID, &model.ClockInRequest{})

	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)

	attendances, total, err := attService.GetByUserAndDateRange(ctx, userID, start, end, 1, 10)
	if err != nil {
		t.Fatalf("GetByUserAndDateRange failed: %v", err)
	}
	if total == 0 {
		t.Error("Expected at least one attendance record")
	}
	if len(attendances) == 0 {
		t.Error("Expected at least one attendance record in results")
	}
}

func TestAttendanceService_GetSummary(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = attService.ClockIn(ctx, userID, &model.ClockInRequest{})
	_, _ = attService.ClockOut(ctx, userID, &model.ClockOutRequest{})

	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)

	summary, err := attService.GetSummary(ctx, userID, start, end)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary == nil {
		t.Error("Summary should not be nil")
	}
}

func TestAttendanceService_GetTodayStatus(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = attService.ClockIn(ctx, userID, &model.ClockInRequest{})

	attendance, err := attService.GetTodayStatus(ctx, userID)
	if err != nil {
		t.Fatalf("GetTodayStatus failed: %v", err)
	}
	if attendance == nil {
		t.Error("Attendance should not be nil")
	}
}

// ===== LeaveService Tests =====

func TestLeaveService_Create_Success(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	userID := uuid.New()
	leave, err := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Family event",
	})

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if leave == nil {
		t.Fatal("Leave should not be nil")
	}
	if leave.Status != model.ApprovalStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", leave.Status)
	}
}

func TestLeaveService_Create_InvalidStartDate(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, err := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "invalid-date",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	if err == nil {
		t.Error("Expected error for invalid start date")
	}
}

func TestLeaveService_Create_InvalidEndDate(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, err := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "invalid-date",
		Reason:    "Test",
	})

	if err == nil {
		t.Error("Expected error for invalid end date")
	}
}

func TestLeaveService_Approve_Success(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	// 休暇申請を作成
	userID := uuid.New()
	leave, _ := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	// 承認
	approverID := uuid.New()
	approved, err := leaveService.Approve(ctx, leave.ID, approverID, &model.LeaveRequestApproval{
		Status: model.ApprovalStatusApproved,
	})

	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if approved.Status != model.ApprovalStatusApproved {
		t.Errorf("Expected status 'approved', got '%s'", approved.Status)
	}
}

func TestLeaveService_Approve_NotFound(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	approverID := uuid.New()
	_, err := leaveService.Approve(ctx, uuid.New(), approverID, &model.LeaveRequestApproval{
		Status: model.ApprovalStatusApproved,
	})

	if err != ErrLeaveNotFound {
		t.Errorf("Expected ErrLeaveNotFound, got %v", err)
	}
}

func TestLeaveService_Approve_AlreadyProcessed(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	// 休暇申請を作成して承認
	userID := uuid.New()
	leave, _ := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	approverID := uuid.New()
	_, _ = leaveService.Approve(ctx, leave.ID, approverID, &model.LeaveRequestApproval{
		Status: model.ApprovalStatusApproved,
	})

	// 再度承認試行
	_, err := leaveService.Approve(ctx, leave.ID, approverID, &model.LeaveRequestApproval{
		Status: model.ApprovalStatusApproved,
	})

	if err != ErrLeaveAlreadyProcessed {
		t.Errorf("Expected ErrLeaveAlreadyProcessed, got %v", err)
	}
}

func TestLeaveService_Approve_Rejected(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	// 休暇申請を作成
	userID := uuid.New()
	leave, _ := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	// 却下
	approverID := uuid.New()
	rejected, err := leaveService.Approve(ctx, leave.ID, approverID, &model.LeaveRequestApproval{
		Status:         model.ApprovalStatusRejected,
		RejectedReason: "Not enough notice",
	})

	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	if rejected.Status != model.ApprovalStatusRejected {
		t.Errorf("Expected status 'rejected', got '%s'", rejected.Status)
	}
	if rejected.RejectedReason != "Not enough notice" {
		t.Errorf("Expected rejected reason, got '%s'", rejected.RejectedReason)
	}
}

func TestLeaveService_GetByUser(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	leaves, total, err := leaveService.GetByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("GetByUser failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 leave, got %d", total)
	}
	if len(leaves) != 1 {
		t.Errorf("Expected 1 leave in results, got %d", len(leaves))
	}
}

func TestLeaveService_GetPending(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	leaves, total, err := leaveService.GetPending(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1 pending leave, got %d", total)
	}
	if len(leaves) != 1 {
		t.Errorf("Expected 1 pending leave in results, got %d", len(leaves))
	}
}

// ===== ShiftService Tests =====

func TestShiftService_Create_Success(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	userID := uuid.New()
	shift, err := shiftService.Create(ctx, &model.ShiftCreateRequest{
		UserID:    userID,
		Date:      "2026-02-10",
		ShiftType: model.ShiftTypeMorning,
		Note:      "Morning shift",
	})

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if shift == nil {
		t.Fatal("Shift should not be nil")
	}
	if shift.ShiftType != model.ShiftTypeMorning {
		t.Errorf("Expected shift type 'morning', got '%s'", shift.ShiftType)
	}
}

func TestShiftService_Create_InvalidDate(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	_, err := shiftService.Create(ctx, &model.ShiftCreateRequest{
		UserID:    uuid.New(),
		Date:      "invalid-date",
		ShiftType: model.ShiftTypeMorning,
	})

	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

func TestShiftService_BulkCreate_Success(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	userID := uuid.New()
	err := shiftService.BulkCreate(ctx, &model.ShiftBulkCreateRequest{
		Shifts: []model.ShiftCreateRequest{
			{UserID: userID, Date: "2026-02-10", ShiftType: model.ShiftTypeMorning},
			{UserID: userID, Date: "2026-02-11", ShiftType: model.ShiftTypeDay},
		},
	})

	if err != nil {
		t.Fatalf("BulkCreate failed: %v", err)
	}
}

func TestShiftService_BulkCreate_InvalidDate(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	err := shiftService.BulkCreate(ctx, &model.ShiftBulkCreateRequest{
		Shifts: []model.ShiftCreateRequest{
			{UserID: uuid.New(), Date: "invalid-date", ShiftType: model.ShiftTypeMorning},
		},
	})

	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

func TestShiftService_GetByUserAndDateRange(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = shiftService.Create(ctx, &model.ShiftCreateRequest{
		UserID:    userID,
		Date:      "2026-02-10",
		ShiftType: model.ShiftTypeMorning,
	})

	start, _ := time.Parse("2006-01-02", "2026-02-01")
	end, _ := time.Parse("2006-01-02", "2026-02-28")

	shifts, err := shiftService.GetByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		t.Fatalf("GetByUserAndDateRange failed: %v", err)
	}
	if len(shifts) != 1 {
		t.Errorf("Expected 1 shift, got %d", len(shifts))
	}
}

func TestShiftService_GetByDateRange(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	_, _ = shiftService.Create(ctx, &model.ShiftCreateRequest{
		UserID:    uuid.New(),
		Date:      "2026-02-10",
		ShiftType: model.ShiftTypeMorning,
	})

	start, _ := time.Parse("2006-01-02", "2026-02-01")
	end, _ := time.Parse("2006-01-02", "2026-02-28")

	shifts, err := shiftService.GetByDateRange(ctx, start, end)
	if err != nil {
		t.Fatalf("GetByDateRange failed: %v", err)
	}
	if len(shifts) != 1 {
		t.Errorf("Expected 1 shift, got %d", len(shifts))
	}
}

func TestShiftService_Delete(t *testing.T) {
	deps := setupTestDeps(t)
	shiftService := NewShiftService(deps)
	ctx := context.Background()

	shift, _ := shiftService.Create(ctx, &model.ShiftCreateRequest{
		UserID:    uuid.New(),
		Date:      "2026-02-10",
		ShiftType: model.ShiftTypeMorning,
	})

	err := shiftService.Delete(ctx, shift.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// ===== UserService Tests =====

func TestUserService_Create_Success(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	user, err := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user == nil {
		t.Fatal("User should not be nil")
	}
	if user.PasswordHash != "" {
		t.Error("Password hash should be cleared")
	}
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	// 最初のユーザーを作成
	_, _ = userService.Create(ctx, &model.UserCreateRequest{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Existing",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	// 重複メールで作成試行
	_, err := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	if err != ErrEmailAlreadyExists {
		t.Errorf("Expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestUserService_GetByID_Success(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	user, err := userService.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestUserService_GetAll(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	_, _ = userService.Create(ctx, &model.UserCreateRequest{
		Email:     "user1@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "One",
		Role:      model.RoleEmployee,
	})
	_, _ = userService.Create(ctx, &model.UserCreateRequest{
		Email:     "user2@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "Two",
		Role:      model.RoleEmployee,
	})

	users, total, err := userService.GetAll(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if total != 2 {
		t.Errorf("Expected 2 users, got %d", total)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users in results, got %d", len(users))
	}
}

func TestUserService_Update_Success(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	newFirstName := "Updated"
	newRole := model.RoleManager
	updated, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		FirstName: &newFirstName,
		Role:      &newRole,
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.FirstName != "Updated" {
		t.Errorf("Expected first name 'Updated', got '%s'", updated.FirstName)
	}
	if updated.Role != model.RoleManager {
		t.Errorf("Expected role 'manager', got '%s'", updated.Role)
	}
}

func TestUserService_Update_WithPassword(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	newPassword := "newpassword456"
	_, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		Password: &newPassword,
	})

	if err != nil {
		t.Fatalf("Update with password failed: %v", err)
	}
}

func TestUserService_Update_NotFound(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	newFirstName := "Updated"
	_, err := userService.Update(ctx, uuid.New(), &model.UserUpdateRequest{
		FirstName: &newFirstName,
	})

	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_Delete(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	err := userService.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// ===== DepartmentService Tests =====

func TestDepartmentService_Create(t *testing.T) {
	deps := setupTestDeps(t)
	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	dept, err := deptService.Create(ctx, &model.Department{Name: "Engineering"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dept.Name != "Engineering" {
		t.Errorf("Expected name 'Engineering', got '%s'", dept.Name)
	}
}

func TestDepartmentService_GetAll(t *testing.T) {
	deps := setupTestDeps(t)
	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	_, _ = deptService.Create(ctx, &model.Department{Name: "Engineering"})
	_, _ = deptService.Create(ctx, &model.Department{Name: "Sales"})

	depts, err := deptService.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(depts) != 2 {
		t.Errorf("Expected 2 departments, got %d", len(depts))
	}
}

func TestDepartmentService_GetByID(t *testing.T) {
	deps := setupTestDeps(t)
	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	created, _ := deptService.Create(ctx, &model.Department{Name: "Engineering"})

	dept, err := deptService.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if dept.Name != "Engineering" {
		t.Errorf("Expected name 'Engineering', got '%s'", dept.Name)
	}
}

func TestDepartmentService_Update(t *testing.T) {
	deps := setupTestDeps(t)
	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	created, _ := deptService.Create(ctx, &model.Department{Name: "Engineering"})
	created.Name = "Tech"

	updated, err := deptService.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Tech" {
		t.Errorf("Expected name 'Tech', got '%s'", updated.Name)
	}
}

func TestDepartmentService_Delete(t *testing.T) {
	deps := setupTestDeps(t)
	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	created, _ := deptService.Create(ctx, &model.Department{Name: "Engineering"})

	err := deptService.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestDepartmentService_Create_Error(t *testing.T) {
	deps := setupTestDeps(t)
	mockDeptRepo := deps.Repos.Department.(*mocks.MockDepartmentRepository)
	mockDeptRepo.CreateErr = errors.New("database error")

	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	_, err := deptService.Create(ctx, &model.Department{Name: "Engineering"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDepartmentService_Update_Error(t *testing.T) {
	deps := setupTestDeps(t)
	mockDeptRepo := deps.Repos.Department.(*mocks.MockDepartmentRepository)
	mockDeptRepo.UpdateErr = errors.New("database error")

	deptService := NewDepartmentService(deps)
	ctx := context.Background()

	dept := &model.Department{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Name:      "Engineering",
	}

	_, err := deptService.Update(ctx, dept)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// ===== DashboardService Tests =====

func TestDashboardService_GetStats(t *testing.T) {
	deps := setupTestDeps(t)
	dashService := NewDashboardService(deps)
	ctx := context.Background()

	stats, err := dashService.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
}

// ===== NewServices Test =====

func TestNewServices(t *testing.T) {
	deps := setupTestDeps(t)
	services := NewServices(deps)

	if services == nil {
		t.Fatal("Services should not be nil")
	}
	if services.Auth == nil {
		t.Error("Auth service should not be nil")
	}
	if services.Attendance == nil {
		t.Error("Attendance service should not be nil")
	}
	if services.Leave == nil {
		t.Error("Leave service should not be nil")
	}
	if services.Shift == nil {
		t.Error("Shift service should not be nil")
	}
	if services.User == nil {
		t.Error("User service should not be nil")
	}
	if services.Department == nil {
		t.Error("Department service should not be nil")
	}
	if services.Dashboard == nil {
		t.Error("Dashboard service should not be nil")
	}
}

// ===== Additional Edge Case Tests =====

func TestAttendanceService_ClockIn_CreateError(t *testing.T) {
	deps := setupTestDeps(t)
	mockAttRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	mockAttRepo.CreateErr = errors.New("database error")

	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, err := attService.ClockIn(ctx, userID, &model.ClockInRequest{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAttendanceService_ClockOut_UpdateError(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, _ = attService.ClockIn(ctx, userID, &model.ClockInRequest{})

	// Update errorを設定
	mockAttRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	mockAttRepo.UpdateErr = errors.New("database error")

	_, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{Note: "Test note"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAuthService_Register_CreateError(t *testing.T) {
	deps := setupTestDeps(t)
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.CreateErr = errors.New("database error")

	authService := NewAuthService(deps)
	ctx := context.Background()

	_, err := authService.Register(ctx, &model.RegisterRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLeaveService_Create_RepoError(t *testing.T) {
	deps := setupTestDeps(t)
	mockLeaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	mockLeaveRepo.CreateErr = errors.New("database error")

	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	userID := uuid.New()
	_, err := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLeaveService_Approve_UpdateError(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps)
	ctx := context.Background()

	// Create a pending leave
	userID := uuid.New()
	leave, _ := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	// Set update error
	mockLeaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	mockLeaveRepo.UpdateErr = errors.New("database error")

	approverID := uuid.New()
	_, err := leaveService.Approve(ctx, leave.ID, approverID, &model.LeaveRequestApproval{
		Status: model.ApprovalStatusApproved,
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestShiftService_Create_RepoError(t *testing.T) {
	deps := setupTestDeps(t)
	mockShiftRepo := deps.Repos.Shift.(*mocks.MockShiftRepository)
	mockShiftRepo.CreateErr = errors.New("database error")

	shiftService := NewShiftService(deps)
	ctx := context.Background()

	_, err := shiftService.Create(ctx, &model.ShiftCreateRequest{
		UserID:    uuid.New(),
		Date:      "2026-02-10",
		ShiftType: model.ShiftTypeMorning,
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestShiftService_BulkCreate_RepoError(t *testing.T) {
	deps := setupTestDeps(t)
	mockShiftRepo := deps.Repos.Shift.(*mocks.MockShiftRepository)
	mockShiftRepo.BulkCreateErr = errors.New("database error")

	shiftService := NewShiftService(deps)
	ctx := context.Background()

	err := shiftService.BulkCreate(ctx, &model.ShiftBulkCreateRequest{
		Shifts: []model.ShiftCreateRequest{
			{UserID: uuid.New(), Date: "2026-02-10", ShiftType: model.ShiftTypeMorning},
		},
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUserService_Create_RepoError(t *testing.T) {
	deps := setupTestDeps(t)
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.CreateErr = errors.New("database error")

	userService := NewUserService(deps)
	ctx := context.Background()

	_, err := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUserService_Update_AllFields(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	newFirstName := "Updated"
	newLastName := "LastName"
	newRole := model.RoleManager
	isActive := false
	deptID := uuid.New()

	updated, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		FirstName:    &newFirstName,
		LastName:     &newLastName,
		Role:         &newRole,
		DepartmentID: &deptID,
		IsActive:     &isActive,
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.LastName != "LastName" {
		t.Errorf("Expected last name 'LastName', got '%s'", updated.LastName)
	}
	if updated.IsActive != false {
		t.Error("Expected IsActive to be false")
	}
}

func TestUserService_Update_RepoError(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
	})

	// Set update error
	mockUserRepo := deps.Repos.User.(*mocks.MockUserRepository)
	mockUserRepo.UpdateErr = errors.New("database error")

	newFirstName := "Updated"
	_, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		FirstName: &newFirstName,
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestAuthService_Logout_Error(t *testing.T) {
	deps := setupTestDeps(t)
	mockRTRepo := deps.Repos.RefreshToken.(*mocks.MockRefreshTokenRepository)
	mockRTRepo.RevokeErr = errors.New("database error")

	authService := NewAuthService(deps)
	ctx := context.Background()

	err := authService.Logout(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
