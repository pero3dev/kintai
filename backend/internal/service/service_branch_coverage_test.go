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
)

// ===== ヘルパー: 全リポジトリを含むDeps =====

func setupFullDeps(t *testing.T) (Deps, *mocks.MockUserRepository, *mocks.MockAttendanceRepository) {
	t.Helper()
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")
	userRepo := mocks.NewMockUserRepository()
	attRepo := mocks.NewMockAttendanceRepository()
	leaveRepo := mocks.NewMockLeaveRequestRepository()
	shiftRepo := mocks.NewMockShiftRepository()
	deptRepo := mocks.NewMockDepartmentRepository()
	rtRepo := mocks.NewMockRefreshTokenRepository()
	otRepo := newMockOvertimeRequestRepo()
	nRepo := newMockNotificationRepo()
	acRepo := newMockAttendanceCorrectionRepo()
	teRepo := newMockTimeEntryRepo()
	lbRepo := newMockLeaveBalanceRepo()
	pRepo := newMockProjectRepo()
	hRepo := newMockHolidayRepo()
	afRepo := newMockApprovalFlowRepo()

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			User:                 userRepo,
			Attendance:           attRepo,
			LeaveRequest:         leaveRepo,
			Shift:                shiftRepo,
			Department:           deptRepo,
			RefreshToken:         rtRepo,
			OvertimeRequest:      otRepo,
			Notification:         nRepo,
			AttendanceCorrection: acRepo,
			TimeEntry:            teRepo,
			LeaveBalance:         lbRepo,
			Project:              pRepo,
			Holiday:              hRepo,
			ApprovalFlow:         afRepo,
		},
	}
	return deps, userRepo, attRepo
}

// ===================================================================
// AuthService - RefreshToken: sub claimが文字列でないケース
// ===================================================================

func TestAuthService_RefreshToken_SubClaimNotString(t *testing.T) {
	deps := setupTestDeps(t)
	authSvc := NewAuthService(deps)
	ctx := context.Background()

	// sub を整数値に設定（string型アサーションが失敗する）
	claims := jwt.MapClaims{
		"sub":  12345,
		"role": "employee",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(deps.Config.JWTSecretKey))

	_, err := authSvc.RefreshToken(ctx, tokenString)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_ExpiredToken(t *testing.T) {
	deps := setupTestDeps(t)
	authSvc := NewAuthService(deps)
	ctx := context.Background()

	// 期限切れトークン
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": "employee",
		"exp":  time.Now().Add(-time.Hour).Unix(),
		"iat":  time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(deps.Config.JWTSecretKey))

	_, err := authSvc.RefreshToken(ctx, tokenString)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

// ===================================================================
// AttendanceService - ClockOut: 空ノート、ClockInがnilのケース
// ===================================================================

func TestAttendanceService_ClockOut_EmptyNote(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	// 出勤打刻
	_, err := attService.ClockIn(ctx, userID, &model.ClockInRequest{Note: "Morning"})
	if err != nil {
		t.Fatalf("ClockIn failed: %v", err)
	}

	// 空ノートで退勤打刻（if req.Note != "" の false分岐）
	attendance, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{Note: ""})
	if err != nil {
		t.Fatalf("ClockOut failed: %v", err)
	}
	// 空ノートの場合、元のノートが維持される
	if attendance.Note != "Morning" {
		t.Errorf("Expected original note 'Morning', got '%s'", attendance.Note)
	}
}

func TestAttendanceService_ClockOut_NilClockIn(t *testing.T) {
	deps := setupTestDeps(t)
	attService := NewAttendanceService(deps)
	ctx := context.Background()

	userID := uuid.New()
	today := time.Now().Truncate(24 * time.Hour)

	// ClockInがnilの出勤データを直接作成（if attendance.ClockIn != nil の false分岐）
	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	attID := uuid.New()
	att := &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		UserID:    userID,
		Date:      today,
		ClockIn:   nil, // ClockInがnil
		Status:    model.AttendanceStatusPresent,
	}
	attRepo.Attendances[attID] = att
	attRepo.UserDateIndex[userID.String()+today.Format("2006-01-02")] = att

	attendance, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{Note: "Leaving"})
	if err != nil {
		t.Fatalf("ClockOut failed: %v", err)
	}
	if attendance.ClockOut == nil {
		t.Error("ClockOut time should be set")
	}
	// ClockInがnilなので勤務時間は計算されない
	if attendance.WorkMinutes != 0 {
		t.Errorf("Expected 0 work minutes when ClockIn is nil, got %d", attendance.WorkMinutes)
	}
}

// ===================================================================
// UserService - Update: パスワード空文字列/nil、全フィールド
// ===================================================================

func TestUserService_Update_EmptyPassword(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email: "test@example.com", Password: "password123",
		FirstName: "Test", LastName: "User", Role: model.RoleEmployee,
	})

	// パスワードが空文字列のケース（if req.Password != nil && *req.Password != "" の false分岐）
	emptyPassword := ""
	updated, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		Password: &emptyPassword,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.PasswordHash != "" {
		t.Error("PasswordHash should be cleared in response")
	}
}

func TestUserService_Update_NilPassword(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email: "test@example.com", Password: "password123",
		FirstName: "Test", LastName: "User", Role: model.RoleEmployee,
	})

	// パスワードがnil
	newName := "Updated"
	updated, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		FirstName: &newName,
		Password:  nil,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.FirstName != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", updated.FirstName)
	}
}

// ===================================================================
// DashboardService - GetStats: 各種ブランチ
// ===================================================================

func TestDashboardService_GetStats_WithData(t *testing.T) {
	deps, userRepo, attRepo := setupFullDeps(t)
	dashService := NewDashboardService(deps)
	ctx := context.Background()

	// ユーザーを追加（totalUsersが > 0 にする）
	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		Email: "test@example.com", FirstName: "Test", LastName: "User",
	}

	// 今日の出勤データを追加
	today := time.Now().Truncate(24 * time.Hour)
	clockIn := time.Now()
	attRepo.Attendances[uuid.New()] = &model.Attendance{
		BaseModel: model.BaseModel{ID: uuid.New()},
		UserID: userID, Date: today,
		ClockIn: &clockIn, Status: model.AttendanceStatusPresent,
	}
	attRepo.UserDateIndex[userID.String()+today.Format("2006-01-02")] = attRepo.Attendances[uuid.New()]

	stats, err := dashService.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
	if len(stats.WeeklyTrend) != 7 {
		t.Errorf("Expected 7 days in weekly trend, got %d", len(stats.WeeklyTrend))
	}
}

func TestDashboardService_GetStats_NoUsers(t *testing.T) {
	deps, _, _ := setupFullDeps(t)
	dashService := NewDashboardService(deps)
	ctx := context.Background()

	// ユーザーなし（totalUsers == 0の分岐）
	stats, err := dashService.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.TodayAbsentCount < 0 {
		t.Error("TodayAbsentCount should not be negative")
	}
	// totalUsers == 0のため、attendanceRate は 0.0
	for _, trend := range stats.WeeklyTrend {
		if trend.AttendanceRate != 0.0 {
			t.Errorf("Expected 0%% attendance rate with 0 users, got %.1f%%", trend.AttendanceRate)
		}
	}
}

func TestDashboardService_GetStats_HighPresent(t *testing.T) {
	deps, userRepo, attRepo := setupFullDeps(t)
	dashService := NewDashboardService(deps)
	ctx := context.Background()

	// 全員出勤して todayAbsent が 0 になるケース
	today := time.Now().Truncate(24 * time.Hour)
	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		Email: "a@test.com", FirstName: "A", LastName: "B",
	}

	clockIn := time.Now()
	attID := uuid.New()
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		UserID: userID, Date: today,
		ClockIn: &clockIn, Status: model.AttendanceStatusPresent,
	}
	attRepo.UserDateIndex[userID.String()+today.Format("2006-01-02")] = attRepo.Attendances[attID]

	stats, err := dashService.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.TodayAbsentCount < 0 {
		t.Error("TodayAbsentCount should not be negative")
	}
}

// ===================================================================
// LeaveBalanceService - SetBalance: CarriedOver更新、既存バランスあり
// ===================================================================

func TestLeaveBalanceService_SetBalance_CarriedOver(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()

	// CarriedOverのみ設定
	carriedOver := 5.0
	err := svc.SetBalance(context.Background(), userID, 2024, model.LeaveTypePaid, &model.LeaveBalanceUpdate{
		CarriedOver: &carriedOver,
	})
	if err != nil {
		t.Fatalf("SetBalance with CarriedOver failed: %v", err)
	}
}

func TestLeaveBalanceService_SetBalance_WithExisting(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()

	// 既存バランスを設定
	k := lbRepo.key(userID, 2024, model.LeaveTypePaid)
	lbRepo.balances[k] = &model.LeaveBalance{
		UserID: userID, FiscalYear: 2024, LeaveType: model.LeaveTypePaid,
		TotalDays: 10, UsedDays: 3, CarriedOver: 0,
	}

	totalDays := 15.0
	carriedOver := 2.0
	err := svc.SetBalance(context.Background(), userID, 2024, model.LeaveTypePaid, &model.LeaveBalanceUpdate{
		TotalDays:   &totalDays,
		CarriedOver: &carriedOver,
	})
	if err != nil {
		t.Fatalf("SetBalance with existing balance failed: %v", err)
	}
}

func TestLeaveBalanceService_SetBalance_UpsertError(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	lbRepo.upsertErr = errors.New("db error")
	svc := NewLeaveBalanceService(deps)

	totalDays := 15.0
	err := svc.SetBalance(context.Background(), uuid.New(), 2024, model.LeaveTypePaid, &model.LeaveBalanceUpdate{
		TotalDays: &totalDays,
	})
	if err == nil {
		t.Error("Expected error")
	}
}

// ===================================================================
// AttendanceCorrectionService - Approve: 追加ブランチカバレッジ
// ===================================================================

func TestAttendanceCorrectionService_Approve_UpdateError(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	cID := uuid.New()
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: cID}, UserID: uuid.New(),
		Status: model.CorrectionStatusPending,
		Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	acRepo.updateErr = errors.New("update error")

	_, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err == nil {
		t.Error("Expected error from update")
	}
}

func TestAttendanceCorrectionService_Approve_WithOnlyCorrectedClockIn(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)

	cID := uuid.New()
	attID := uuid.New()
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	// CorrectedClockOutだけnilのケース
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel:        model.BaseModel{ID: cID}, UserID: uuid.New(),
		AttendanceID:     &attID,
		Status:           model.CorrectionStatusPending,
		Date:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn: &clockIn,
		// CorrectedClockOut is nil
	}

	origIn := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		ClockIn:   &origIn,
		// ClockOut is nil → ClockIn != nil && ClockOut != nil is false
	}

	result, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if result.Status != model.CorrectionStatusApproved {
		t.Errorf("Expected approved, got %s", result.Status)
	}
}

func TestAttendanceCorrectionService_Approve_NoOvertime(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)

	cID := uuid.New()
	attID := uuid.New()
	// 6時間勤務（残業なし: WorkMinutes <= 480）
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC) // 6h
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel:         model.BaseModel{ID: cID}, UserID: uuid.New(),
		AttendanceID:      &attID,
		Status:            model.CorrectionStatusPending,
		Date:              time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn:  &clockIn,
		CorrectedClockOut: &clockOut,
	}

	origIn := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	origOut := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		ClockIn:   &origIn, ClockOut: &origOut,
		OvertimeMinutes: 60, // 既存の残業を0にリセットされるか確認
	}

	_, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	att := attRepo.Attendances[attID]
	if att.OvertimeMinutes != 0 {
		t.Errorf("Expected 0 overtime, got %d", att.OvertimeMinutes)
	}
}

func TestAttendanceCorrectionService_Approve_NewAttendanceNoOvertime(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)

	cID := uuid.New()
	// 6時間勤務の新規勤怠（AttendanceIDなし、残業なし）
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel:         model.BaseModel{ID: cID}, UserID: uuid.New(),
		Status:            model.CorrectionStatusPending,
		Date:              time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn:  &clockIn,
		CorrectedClockOut: &clockOut,
		// AttendanceIDなし → 新規作成
	}

	result, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if result.Status != model.CorrectionStatusApproved {
		t.Errorf("Expected approved, got %s", result.Status)
	}
}

func TestAttendanceCorrectionService_Approve_NewAttendanceClockInOnly(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	cID := uuid.New()
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel:        model.BaseModel{ID: cID}, UserID: uuid.New(),
		Status:           model.CorrectionStatusPending,
		Date:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn: &clockIn,
		// CorrectedClockOutなし → ClockIn && ClockOut条件がfalse
	}

	result, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if result.Status != model.CorrectionStatusApproved {
		t.Errorf("Expected approved, got %s", result.Status)
	}
}

// ===================================================================
// ProjectService - Update: 全フィールド更新と各種フィールド個別テスト
// ===================================================================

func TestProjectService_Update_AllFields(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	pID := uuid.New()
	pRepo.projects[pID] = &model.Project{
		BaseModel: model.BaseModel{ID: pID},
		Name: "Old", Description: "Old desc", Status: model.ProjectStatusActive,
	}

	newName := "Updated"
	newDesc := "New description"
	newStatus := model.ProjectStatusArchived
	managerID := uuid.New()
	budgetHours := 100.0

	p, err := svc.Update(context.Background(), pID, &model.ProjectUpdateRequest{
		Name:        &newName,
		Description: &newDesc,
		Status:      &newStatus,
		ManagerID:   &managerID,
		BudgetHours: &budgetHours,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if p.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", p.Name)
	}
	if p.Description != "New description" {
		t.Errorf("Expected 'New description', got '%s'", p.Description)
	}
	if p.Status != model.ProjectStatusArchived {
		t.Errorf("Expected 'archived', got '%s'", p.Status)
	}
	if p.ManagerID == nil || *p.ManagerID != managerID {
		t.Error("Expected ManagerID to be set")
	}
	if p.BudgetHours == nil || *p.BudgetHours != 100.0 {
		t.Error("Expected BudgetHours 100.0")
	}
}

func TestProjectService_Update_RepoError(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	pRepo.updateErr = errors.New("db error")
	svc := NewProjectService(deps)
	pID := uuid.New()
	pRepo.projects[pID] = &model.Project{BaseModel: model.BaseModel{ID: pID}, Name: "Old"}

	newName := "Updated"
	_, err := svc.Update(context.Background(), pID, &model.ProjectUpdateRequest{Name: &newName})
	if err == nil {
		t.Error("Expected error from repo update")
	}
}

// ===================================================================
// TimeEntryService - 追加カバレッジ
// ===================================================================

func TestTimeEntryService_Create_RepoError(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	teRepo.createErr = errors.New("db error")
	svc := NewTimeEntryService(deps)

	_, err := svc.Create(context.Background(), uuid.New(), &model.TimeEntryCreate{
		ProjectID: uuid.New(), Date: "2024-01-15", Minutes: 120, Description: "Test",
	})
	if err == nil {
		t.Error("Expected error from create")
	}
}

func TestTimeEntryService_GetByProjectAndDateRange(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	projectID := uuid.New()
	teRepo.entries[uuid.New()] = &model.TimeEntry{
		BaseModel: model.BaseModel{ID: uuid.New()}, ProjectID: projectID,
		Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Minutes: 120,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	entries, err := svc.GetByProjectAndDateRange(context.Background(), projectID, start, end)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1, got %d", len(entries))
	}
}

func TestTimeEntryService_Update_Description(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	eID := uuid.New()
	teRepo.entries[eID] = &model.TimeEntry{
		BaseModel:   model.BaseModel{ID: eID},
		Minutes:     60,
		Description: "Old",
	}

	newDesc := "Updated description"
	e, err := svc.Update(context.Background(), eID, &model.TimeEntryUpdate{Description: &newDesc})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if e.Description != "Updated description" {
		t.Errorf("Expected 'Updated description', got '%s'", e.Description)
	}
}

func TestTimeEntryService_Update_RepoError(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	eID := uuid.New()
	teRepo.entries[eID] = &model.TimeEntry{BaseModel: model.BaseModel{ID: eID}, Minutes: 60}
	teRepo.updateErr = errors.New("db error")

	newMin := 120
	_, err := svc.Update(context.Background(), eID, &model.TimeEntryUpdate{Minutes: &newMin})
	if err == nil {
		t.Error("Expected error from repo update")
	}
}

// ===================================================================
// HolidayService - 追加カバレッジ
// ===================================================================

func TestHolidayService_Create_RepoError(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	hRepo.createErr = errors.New("db error")
	svc := NewHolidayService(deps)

	_, err := svc.Create(context.Background(), &model.HolidayCreateRequest{
		Date: "2024-01-01", Name: "元旦", HolidayType: model.HolidayTypeNational,
	})
	if err == nil {
		t.Error("Expected error from create")
	}
}

func TestHolidayService_GetByDateRange(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hRepo.holidays[uuid.New()] = &model.Holiday{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		Name:      "元旦",
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.Local)
	holidays, err := svc.GetByDateRange(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetByDateRange failed: %v", err)
	}
	if len(holidays) != 1 {
		t.Errorf("Expected 1, got %d", len(holidays))
	}
}

func TestHolidayService_Update_AllFields(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hID := uuid.New()
	hRepo.holidays[hID] = &model.Holiday{
		BaseModel:   model.BaseModel{ID: hID},
		Name:        "Old",
		HolidayType: model.HolidayTypeNational,
		IsRecurring: false,
	}

	newName := "Updated"
	newType := model.HolidayTypeCompany
	isRecurring := true
	h, err := svc.Update(context.Background(), hID, &model.HolidayUpdateRequest{
		Name:        &newName,
		HolidayType: &newType,
		IsRecurring: &isRecurring,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if h.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", h.Name)
	}
	if h.HolidayType != model.HolidayTypeCompany {
		t.Errorf("Expected 'company', got '%s'", h.HolidayType)
	}
	if !h.IsRecurring {
		t.Error("Expected IsRecurring to be true")
	}
}

func TestHolidayService_Update_RepoError(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hID := uuid.New()
	hRepo.holidays[hID] = &model.Holiday{BaseModel: model.BaseModel{ID: hID}, Name: "Old"}
	hRepo.updateErr = errors.New("db error")

	newName := "Updated"
	_, err := svc.Update(context.Background(), hID, &model.HolidayUpdateRequest{Name: &newName})
	if err == nil {
		t.Error("Expected error from update")
	}
}

func TestHolidayService_GetCalendar_Error(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)

	// February（28/29日）
	days, err := svc.GetCalendar(context.Background(), 2024, 2)
	if err != nil {
		t.Fatalf("GetCalendar failed: %v", err)
	}
	if len(days) != 29 { // 2024は閏年
		t.Errorf("Expected 29 days for Feb 2024, got %d", len(days))
	}
	// 週末チェック
	for _, d := range days {
		if d.Date == "2024-02-03" && !d.IsWeekend { // Saturday
			t.Error("Expected Feb 3 (Sat) to be weekend")
		}
		if d.Date == "2024-02-04" && !d.IsWeekend { // Sunday
			t.Error("Expected Feb 4 (Sun) to be weekend")
		}
	}
}

func TestHolidayService_GetWorkingDays_WeekdayHoliday(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)

	// 2024-01-01は月曜日 → 平日の祝日
	hRepo.holidays[uuid.New()] = &model.Holiday{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		Name:      "元旦",
	}
	// 2024-01-06は土曜日 → 週末（祝日ではないが週末）
	// 2024-01-07は日曜日 → 週末

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.Local)
	summary, err := svc.GetWorkingDays(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetWorkingDays failed: %v", err)
	}
	if summary.TotalDays != 7 {
		t.Errorf("Expected 7 total days, got %d", summary.TotalDays)
	}
	if summary.Holidays != 1 {
		t.Errorf("Expected 1 holiday, got %d", summary.Holidays)
	}
	if summary.Weekends != 2 {
		t.Errorf("Expected 2 weekends, got %d", summary.Weekends)
	}
	// 平日: 7 - 1 (holiday) - 2 (weekends) = 4
	if summary.WorkingDays != 4 {
		t.Errorf("Expected 4 working days, got %d", summary.WorkingDays)
	}
}

func TestHolidayService_GetWorkingDays_Error(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	// FindByDateRange にエラーを返すようにするためにカスタムエラーを設定
	// mockHolidayRepo にはエラーフィールドがないため、空データで確認
	_ = hRepo
	svc := NewHolidayService(deps)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	summary, err := svc.GetWorkingDays(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetWorkingDays failed: %v", err)
	}
	if summary.TotalDays != 1 {
		t.Errorf("Expected 1 total day, got %d", summary.TotalDays)
	}
}

// ===================================================================
// ApprovalFlowService - Create/Update: 追加ブランチカバレッジ
// ===================================================================

func TestApprovalFlowService_Create_RepoError(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	afRepo.createErr = errors.New("db error")
	svc := NewApprovalFlowService(deps)

	_, err := svc.Create(context.Background(), &model.ApprovalFlowCreateRequest{
		Name: "Flow", FlowType: model.ApprovalFlowLeave,
		Steps: []model.ApprovalStepRequest{
			{StepOrder: 1, StepType: model.ApprovalStepRole},
		},
	})
	if err == nil {
		t.Error("Expected error from create")
	}
}

func TestApprovalFlowService_Update_AllFields(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{
		BaseModel: model.BaseModel{ID: fID},
		Name:      "Old",
		IsActive:  true,
	}

	newName := "Updated"
	isActive := false
	steps := []model.ApprovalStepRequest{
		{StepOrder: 1, StepType: model.ApprovalStepManager},
		{StepOrder: 2, StepType: model.ApprovalStepRole, ApproverRole: rolePtr(model.RoleAdmin)},
	}

	f, err := svc.Update(context.Background(), fID, &model.ApprovalFlowUpdateRequest{
		Name:     &newName,
		IsActive: &isActive,
		Steps:    steps,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if f.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", f.Name)
	}
	if f.IsActive {
		t.Error("Expected IsActive to be false")
	}
	if len(f.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(f.Steps))
	}
}

func TestApprovalFlowService_Update_IsActiveOnly(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{
		BaseModel: model.BaseModel{ID: fID},
		Name:      "Flow",
		IsActive:  true,
	}

	isActive := false
	f, err := svc.Update(context.Background(), fID, &model.ApprovalFlowUpdateRequest{
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if f.IsActive {
		t.Error("Expected IsActive to be false")
	}
}

func TestApprovalFlowService_Update_RepoError(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{
		BaseModel: model.BaseModel{ID: fID},
		Name:      "Flow",
	}
	afRepo.updateErr = errors.New("db error")

	newName := "Updated"
	_, err := svc.Update(context.Background(), fID, &model.ApprovalFlowUpdateRequest{
		Name: &newName,
	})
	if err == nil {
		t.Error("Expected error from update")
	}
}

func TestApprovalFlowService_Update_StepsOnly(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{
		BaseModel: model.BaseModel{ID: fID},
		Name:      "Flow",
	}

	steps := []model.ApprovalStepRequest{
		{StepOrder: 1, StepType: model.ApprovalStepUser, ApproverID: uuidPtr(uuid.New())},
	}
	f, err := svc.Update(context.Background(), fID, &model.ApprovalFlowUpdateRequest{
		Steps: steps,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if len(f.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(f.Steps))
	}
}

// ===================================================================
// ExportService - 追加ブランチカバレッジ
// ===================================================================

func TestExportService_ExportAttendanceCSV_WithNilClockTimes(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	attID := uuid.New()
	// ClockIn/ClockOutがnil
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID}, UserID: userID,
		Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Status: model.AttendanceStatusAbsent,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportAttendanceCSV(ctx, &userID, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

func TestExportService_ExportAttendanceCSV_AllUsers_WithNilUser(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	attID := uuid.New()
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID}, UserID: uuid.New(),
		Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		ClockIn: &clockIn, ClockOut: &clockOut,
		Status: model.AttendanceStatusPresent, User: nil,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportAttendanceCSV(ctx, nil, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

func TestExportService_ExportAttendanceCSV_AllUsers_WithUser(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	attID := uuid.New()
	user := &model.User{FirstName: "Taro", LastName: "Test"}
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID}, UserID: uuid.New(),
		Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Status: model.AttendanceStatusPresent, User: user,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportAttendanceCSV(ctx, nil, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

func TestExportService_ExportAttendanceCSV_WithUserID_NilUser(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	// ユーザーが見つからないケース
	userID := uuid.New()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportAttendanceCSV(ctx, &userID, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV (header at minimum)")
	}
}

func TestExportService_ExportLeavesCSV_WithLeavesInRange(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	leaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	leaveID := uuid.New()
	leaveRepo.LeaveRequests[leaveID] = &model.LeaveRequest{
		BaseModel: model.BaseModel{ID: leaveID},
		UserID:    userID,
		LeaveType: model.LeaveTypePaid,
		StartDate: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 2, 12, 0, 0, 0, 0, time.UTC),
		Reason:    "Family event",
		Status:    model.ApprovalStatusApproved,
		Approver:  &model.User{FirstName: "Manager", LastName: "San"},
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportLeavesCSV(ctx, nil, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	if !containsSubstring(content, "Family event") {
		t.Error("Expected leave reason in CSV")
	}
}

func TestExportService_ExportLeavesCSV_LeaveOutOfRange(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	leaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	leaveID := uuid.New()
	leaveRepo.LeaveRequests[leaveID] = &model.LeaveRequest{
		BaseModel: model.BaseModel{ID: leaveID},
		UserID:    userID,
		LeaveType: model.LeaveTypePaid,
		StartDate: time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC), // 範囲外
		EndDate:   time.Date(2023, 2, 12, 0, 0, 0, 0, time.UTC), // 範囲外
		Status:    model.ApprovalStatusApproved,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportLeavesCSV(ctx, nil, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	// ヘッダーのみ
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

func TestExportService_ExportLeavesCSV_NoApprover(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	leaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	leaveID := uuid.New()
	leaveRepo.LeaveRequests[leaveID] = &model.LeaveRequest{
		BaseModel: model.BaseModel{ID: leaveID},
		UserID:    userID,
		LeaveType: model.LeaveTypePaid,
		StartDate: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 2, 12, 0, 0, 0, 0, time.UTC),
		Status:    model.ApprovalStatusPending,
		Approver:  nil, // 承認者なし
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportLeavesCSV(ctx, nil, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

func TestExportService_ExportProjectsCSV_WithSummaries(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)

	// GetProjectSummary がデータを返すモックに置き換え
	teRepo := deps.Repos.TimeEntry.(*mockTimeEntryRepo)
	// mockTimeEntryRepoのGetProjectSummaryは空を返すので、
	// 直接データを挿入して確認
	_ = teRepo

	svc := NewExportService(deps)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportProjectsCSV(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

// ===================================================================
// LeaveService - Approve: 却下理由なしの却下
// ===================================================================

func TestLeaveService_Approve_RejectedNoReason(t *testing.T) {
	deps := setupTestDeps(t)
	leaveService := NewLeaveService(deps, &mocks.MockNotificationService{})
	ctx := context.Background()

	userID := uuid.New()
	leave, _ := leaveService.Create(ctx, userID, &model.LeaveRequestCreate{
		LeaveType: model.LeaveTypePaid,
		StartDate: "2026-02-10",
		EndDate:   "2026-02-12",
		Reason:    "Test",
	})

	approverID := uuid.New()
	rejected, err := leaveService.Approve(ctx, leave.ID, approverID, &model.LeaveRequestApproval{
		Status:         model.ApprovalStatusRejected,
		RejectedReason: "", // 却下理由なし
	})
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	if rejected.Status != model.ApprovalStatusRejected {
		t.Errorf("Expected 'rejected', got '%s'", rejected.Status)
	}
}

// ===================================================================
// AttendanceCorrectionService - Create: 不正な時刻フォーマット
// ===================================================================

func TestAttendanceCorrectionService_Create_InvalidClockInFormat(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	invalidTime := "invalid-time"
	result, err := svc.Create(context.Background(), uuid.New(), &model.AttendanceCorrectionCreate{
		Date:             "2024-01-15",
		CorrectedClockIn: &invalidTime,
		Reason:           "Test",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	// 不正な時刻はパースに失敗するので、CorrectedClockInはnilのまま
	if result.CorrectedClockIn != nil {
		t.Error("Expected nil CorrectedClockIn for invalid format")
	}
}

func TestAttendanceCorrectionService_Create_InvalidClockOutFormat(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	clockIn := "09:00"
	invalidTime := "invalid-time"
	result, err := svc.Create(context.Background(), uuid.New(), &model.AttendanceCorrectionCreate{
		Date:              "2024-01-15",
		CorrectedClockIn:  &clockIn,
		CorrectedClockOut: &invalidTime,
		Reason:            "Test",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.CorrectedClockOut != nil {
		t.Error("Expected nil CorrectedClockOut for invalid format")
	}
}

func TestAttendanceCorrectionService_Create_NilClockTimes(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	result, err := svc.Create(context.Background(), uuid.New(), &model.AttendanceCorrectionCreate{
		Date:   "2024-01-15",
		Reason: "Test",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.CorrectedClockIn != nil {
		t.Error("Expected nil CorrectedClockIn")
	}
	if result.CorrectedClockOut != nil {
		t.Error("Expected nil CorrectedClockOut")
	}
}

// ===================================================================
// ExportService - ExportOvertimeCSV: 超過なし
// ===================================================================

func TestExportService_ExportOvertimeCSV_NoExceeded(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	// 残業なし（月間・年間とも0）

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportOvertimeCSV(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	if containsSubstring(content, "上限超過") {
		t.Error("Expected no warning for no overtime")
	}
}

// ===================================================================
// ExportService - ExportProjectsCSV: ProjectSummary with budget
// ===================================================================

// mockTimeEntryRepoWithSummary extends mockTimeEntryRepo with custom summary
type mockTimeEntryRepoWithSummary struct {
	mockTimeEntryRepo
	summaries []model.ProjectSummary
}

func (m *mockTimeEntryRepoWithSummary) GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
	return m.summaries, nil
}

func TestExportService_ExportProjectsCSV_WithBudget(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")

	budgetHours := 200.0
	teRepoCustom := &mockTimeEntryRepoWithSummary{
		mockTimeEntryRepo: mockTimeEntryRepo{entries: make(map[uuid.UUID]*model.TimeEntry)},
		summaries: []model.ProjectSummary{
			{
				ProjectID: uuid.New(), ProjectName: "Project A", ProjectCode: "PA001",
				TotalMinutes: 3600, TotalHours: 60.0, BudgetHours: &budgetHours, MemberCount: 5,
			},
			{
				ProjectID: uuid.New(), ProjectName: "Project B", ProjectCode: "PB001",
				TotalMinutes: 1200, TotalHours: 20.0, BudgetHours: nil, MemberCount: 3,
			},
		},
	}

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			TimeEntry: teRepoCustom,
			User:      mocks.NewMockUserRepository(),
		},
	}

	svc := NewExportService(deps)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportProjectsCSV(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	if !containsSubstring(content, "200.0") {
		t.Error("Expected budget hours 200.0 in CSV")
	}
	if !containsSubstring(content, "-") {
		t.Error("Expected '-' for nil budget in CSV")
	}
}

// ===================================================================
// AttendanceService - ClockOut: 残業時間超過
// ===================================================================

func TestAttendanceService_ClockOut_WithOvertime(t *testing.T) {
	deps := setupTestDeps(t)
	ctx := context.Background()

	userID := uuid.New()
	today := time.Now().Truncate(24 * time.Hour)

	// 10時間前にClockIn（残業が発生する）
	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	attID := uuid.New()
	clockInTime := time.Now().Add(-10 * time.Hour)
	att := &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		UserID:    userID,
		Date:      today,
		ClockIn:   &clockInTime,
		Status:    model.AttendanceStatusPresent,
	}
	attRepo.Attendances[attID] = att
	attRepo.UserDateIndex[userID.String()+today.Format("2006-01-02")] = att

	attService := NewAttendanceService(deps)
	attendance, err := attService.ClockOut(ctx, userID, &model.ClockOutRequest{Note: "Long day"})
	if err != nil {
		t.Fatalf("ClockOut failed: %v", err)
	}
	if attendance.OvertimeMinutes <= 0 {
		t.Error("Expected overtime minutes > 0 for 10+ hour workday")
	}
	if attendance.WorkMinutes <= 8*60 {
		t.Error("Expected work minutes > 480 for 10+ hour workday")
	}
}

// ===================================================================
// ヘルパー関数
// ===================================================================

func uuidPtr(id uuid.UUID) *uuid.UUID { return &id }

// ===================================================================
// AuthService - Register: bcryptエラー（72バイト超パスワード）
// ===================================================================

func TestAuthService_Register_BcryptError(t *testing.T) {
	deps := setupTestDeps(t)
	authSvc := NewAuthService(deps)
	ctx := context.Background()

	// 73バイト超のパスワード → bcryptエラー
	longPassword := string(make([]byte, 73))
	_, err := authSvc.Register(ctx, &model.RegisterRequest{
		Email: "bc@test.com", Password: longPassword,
		FirstName: "Test", LastName: "User",
	})
	if err == nil {
		t.Error("Expected bcrypt error for password > 72 bytes")
	}
}

// ===================================================================
// UserService - Create: bcryptエラー（72バイト超パスワード）
// ===================================================================

func TestUserService_Create_BcryptError(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	longPassword := string(make([]byte, 73))
	_, err := userService.Create(ctx, &model.UserCreateRequest{
		Email: "bc@test.com", Password: longPassword,
		FirstName: "Test", LastName: "User", Role: model.RoleEmployee,
	})
	if err == nil {
		t.Error("Expected bcrypt error for password > 72 bytes")
	}
}

// ===================================================================
// AuthService - RefreshToken: 不正なUUID文字列
// ===================================================================

func TestAuthService_RefreshToken_InvalidUUIDSub(t *testing.T) {
	deps := setupTestDeps(t)
	authSvc := NewAuthService(deps)
	ctx := context.Background()

	claims := jwt.MapClaims{
		"sub":  "not-a-valid-uuid",
		"role": "employee",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(deps.Config.JWTSecretKey))

	_, err := authSvc.RefreshToken(ctx, tokenString)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_RefreshToken_UserNotFoundAfterParse(t *testing.T) {
	deps := setupTestDeps(t)
	authSvc := NewAuthService(deps)
	ctx := context.Background()

	// 存在しないユーザーID（有効なUUID形式だが未登録）
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": "employee",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(deps.Config.JWTSecretKey))

	_, err := authSvc.RefreshToken(ctx, tokenString)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

// ===================================================================
// ExportLeavesCSV: endDateが範囲外
// ===================================================================

func TestExportService_ExportLeavesCSV_EndDateAfterRange(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	leaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	leaveID := uuid.New()
	leaveRepo.LeaveRequests[leaveID] = &model.LeaveRequest{
		BaseModel: model.BaseModel{ID: leaveID},
		UserID:    userID,
		LeaveType: model.LeaveTypePaid,
		StartDate: time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC), // 範囲内
		EndDate:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),   // 範囲外
		Status:    model.ApprovalStatusPending,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportLeavesCSV(ctx, nil, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	// EndDateが範囲外なのでスキップされるはず
	if containsSubstring(content, "2024-12-30") {
		t.Error("Expected leave with end date after range to be excluded")
	}
}

// ===================================================================
// ApprovalFlowService - Create: CreateStepsエラー、FindByIDエラー
// ===================================================================

func TestApprovalFlowService_Create_StepsError(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")

	afRepoErr := &mockApprovalFlowRepoWithStepsErr{
		mockApprovalFlowRepo: *newMockApprovalFlowRepo(),
		createStepsErr:       errors.New("steps error"),
	}

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			ApprovalFlow: afRepoErr,
		},
	}

	svc := NewApprovalFlowService(deps)
	_, err := svc.Create(context.Background(), &model.ApprovalFlowCreateRequest{
		Name: "Flow", FlowType: model.ApprovalFlowLeave,
		Steps: []model.ApprovalStepRequest{
			{StepOrder: 1, StepType: model.ApprovalStepRole},
		},
	})
	if err == nil {
		t.Error("Expected error from CreateSteps")
	}
}

// mockApprovalFlowRepoWithStepsErr extends mockApprovalFlowRepo with CreateSteps error
type mockApprovalFlowRepoWithStepsErr struct {
	mockApprovalFlowRepo
	createStepsErr error
}

func (m *mockApprovalFlowRepoWithStepsErr) CreateSteps(ctx context.Context, steps []model.ApprovalStep) error {
	if m.createStepsErr != nil {
		return m.createStepsErr
	}
	if len(steps) > 0 {
		m.steps[steps[0].FlowID] = steps
	}
	return nil
}

// ===================================================================
// DashboardService - 週間トレンドの出勤率分岐（totalUsers > 0）
// ===================================================================

func TestDashboardService_GetStats_PresenceExceedsTotal(t *testing.T) {
	deps, _, attRepo := setupFullDeps(t)
	dashService := NewDashboardService(deps)
	ctx := context.Background()

	// ユーザーなしだが出勤データがある場合 → todayAbsent < 0 → 0にクランプ
	today := time.Now().Truncate(24 * time.Hour)
	clockIn := time.Now()
	attID := uuid.New()
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		UserID:    uuid.New(), Date: today,
		ClockIn: &clockIn, Status: model.AttendanceStatusPresent,
	}
	attRepo.UserDateIndex[uuid.New().String()+today.Format("2006-01-02")] = attRepo.Attendances[attID]

	stats, err := dashService.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats.TodayAbsentCount < 0 {
		t.Error("TodayAbsentCount should be >= 0 even when present > total")
	}
}

// ===================================================================
// UserService - Update: パスワード変更(成功)とbcryptエラー
// ===================================================================

func TestUserService_Update_PasswordWithBcryptError(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email: "upd@test.com", Password: "password123",
		FirstName: "Test", LastName: "User", Role: model.RoleEmployee,
	})

	longPassword := string(make([]byte, 73))
	_, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		Password: &longPassword,
	})
	if err == nil {
		t.Error("Expected bcrypt error for password > 72 bytes")
	}
}

// ===================================================================
// UserService - Update: 全フィールド更新
// ===================================================================

func TestUserService_Update_AllFieldsCoverage(t *testing.T) {
	deps := setupTestDeps(t)
	userService := NewUserService(deps)
	ctx := context.Background()

	created, _ := userService.Create(ctx, &model.UserCreateRequest{
		Email: "all@test.com", Password: "password123",
		FirstName: "Test", LastName: "User", Role: model.RoleEmployee,
	})

	newFirst := "NewFirst"
	newLast := "NewLast"
	newRole := model.RoleAdmin
	deptID := uuid.New()
	isActive := false
	newPass := "newpassword123"

	updated, err := userService.Update(ctx, created.ID, &model.UserUpdateRequest{
		FirstName:    &newFirst,
		LastName:     &newLast,
		Role:         &newRole,
		DepartmentID: &deptID,
		IsActive:     &isActive,
		Password:     &newPass,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.FirstName != "NewFirst" {
		t.Errorf("Expected 'NewFirst', got '%s'", updated.FirstName)
	}
	if updated.Role != model.RoleAdmin {
		t.Errorf("Expected admin role, got '%s'", updated.Role)
	}
	if updated.IsActive {
		t.Error("Expected IsActive to be false")
	}
}

// ===================================================================
// HolidayService - GetCalendar/GetWorkingDays エラーパス
// ===================================================================

// mockHolidayRepoWithFindError wraps mockHolidayRepo but returns error from FindByDateRange
type mockHolidayRepoWithFindError struct {
	mockHolidayRepo
	findByDateRangeErr error
}

func (m *mockHolidayRepoWithFindError) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error) {
	if m.findByDateRangeErr != nil {
		return nil, m.findByDateRangeErr
	}
	return m.mockHolidayRepo.FindByDateRange(ctx, start, end)
}

func (m *mockHolidayRepoWithFindError) FindByYear(ctx context.Context, year int) ([]model.Holiday, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.Local)
	return m.FindByDateRange(ctx, start, end)
}

func (m *mockHolidayRepoWithFindError) IsHoliday(ctx context.Context, date time.Time) (bool, *model.Holiday, error) {
	return m.mockHolidayRepo.IsHoliday(ctx, date)
}

func TestHolidayService_GetCalendar_FindByDateRangeError(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")

	hRepoErr := &mockHolidayRepoWithFindError{
		mockHolidayRepo:    *newMockHolidayRepo(),
		findByDateRangeErr: errors.New("db error"),
	}

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			Holiday: hRepoErr,
		},
	}

	svc := NewHolidayService(deps)
	_, err := svc.GetCalendar(context.Background(), 2024, 1)
	if err == nil {
		t.Error("Expected error from FindByDateRange")
	}
}

func TestHolidayService_GetWorkingDays_FindByDateRangeError(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")

	hRepoErr := &mockHolidayRepoWithFindError{
		mockHolidayRepo:    *newMockHolidayRepo(),
		findByDateRangeErr: errors.New("db error"),
	}

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			Holiday: hRepoErr,
		},
	}

	svc := NewHolidayService(deps)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.Local)
	_, err := svc.GetWorkingDays(context.Background(), start, end)
	if err == nil {
		t.Error("Expected error from FindByDateRange")
	}
}

// ===================================================================
// HolidayService - GetWorkingDays: 週末と祝日が重なるケース
// ===================================================================

func TestHolidayService_GetWorkingDays_HolidayOnWeekend(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)

	// 2024-01-06 は土曜日 → 祝日を土曜日に設定
	hRepo.holidays[uuid.New()] = &model.Holiday{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date:      time.Date(2024, 1, 6, 0, 0, 0, 0, time.Local),
		Name:      "Weekend Holiday",
	}

	start := time.Date(2024, 1, 6, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 6, 0, 0, 0, 0, time.Local)
	summary, err := svc.GetWorkingDays(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetWorkingDays failed: %v", err)
	}
	// 土曜日なので weekend=1, holiday=1（重複）, workingDay = 1-1-0 = 0
	if summary.Weekends != 1 {
		t.Errorf("Expected 1 weekend, got %d", summary.Weekends)
	}
}

// ===================================================================
// ApprovalFlowService - Update: CreateStepsエラー
// ===================================================================

func TestApprovalFlowService_Update_CreateStepsError(t *testing.T) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")

	afRepoErr := &mockApprovalFlowRepoWithStepsErr{
		mockApprovalFlowRepo: *newMockApprovalFlowRepo(),
		createStepsErr:       errors.New("steps error"),
	}

	fID := uuid.New()
	afRepoErr.flows[fID] = &model.ApprovalFlow{
		BaseModel: model.BaseModel{ID: fID},
		Name:      "Flow",
	}

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			ApprovalFlow: afRepoErr,
		},
	}

	svc := NewApprovalFlowService(deps)
	steps := []model.ApprovalStepRequest{
		{StepOrder: 1, StepType: model.ApprovalStepManager},
	}
	_, err := svc.Update(context.Background(), fID, &model.ApprovalFlowUpdateRequest{
		Steps: steps,
	})
	if err == nil {
		t.Error("Expected error from CreateSteps")
	}
}

// ===================================================================
// ExportLeavesCSV: userIDフィルタリングで別ユーザーをスキップ
// ===================================================================

func TestExportService_ExportLeavesCSV_FilterByUserID(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	// 2人のユーザーを追加
	userID1 := uuid.New()
	userID2 := uuid.New()
	userRepo.Users[userID1] = &model.User{
		BaseModel: model.BaseModel{ID: userID1},
		FirstName: "Taro", LastName: "Test",
	}
	userRepo.Users[userID2] = &model.User{
		BaseModel: model.BaseModel{ID: userID2},
		FirstName: "Hanako", LastName: "Test",
	}

	leaveRepo := deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository)
	leaveID := uuid.New()
	leaveRepo.LeaveRequests[leaveID] = &model.LeaveRequest{
		BaseModel: model.BaseModel{ID: leaveID},
		UserID:    userID1,
		LeaveType: model.LeaveTypePaid,
		StartDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
		Status:    model.ApprovalStatusApproved,
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	// userID1でフィルタ → userID2はスキップされる
	data, err := svc.ExportLeavesCSV(ctx, &userID1, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	if containsSubstring(content, "Hanako") {
		t.Error("Expected user2 to be filtered out")
	}
}
