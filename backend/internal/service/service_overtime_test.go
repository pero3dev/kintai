package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"github.com/your-org/kintai/backend/pkg/logger"
)

func setupOvertimeDeps(t *testing.T) (Deps, *mockOvertimeRequestRepo, *mocks.MockUserRepository) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")
	otRepo := newMockOvertimeRequestRepo()
	userRepo := mocks.NewMockUserRepository()
	attRepo := mocks.NewMockAttendanceRepository()
	leaveRepo := mocks.NewMockLeaveRequestRepository()
	shiftRepo := mocks.NewMockShiftRepository()
	deptRepo := mocks.NewMockDepartmentRepository()
	rtRepo := mocks.NewMockRefreshTokenRepository()
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
	return deps, otRepo, userRepo
}

// ===== OvertimeRequestService Tests =====

func TestOvertimeRequestService_Create_Success(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	userID := uuid.New()
	ot, err := svc.Create(context.Background(), userID, &model.OvertimeRequestCreate{
		Date: "2024-01-15", PlannedMinutes: 120, Reason: "Deadline",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ot.PlannedMinutes != 120 {
		t.Errorf("Expected 120, got %d", ot.PlannedMinutes)
	}
	if ot.Status != model.OvertimeStatusPending {
		t.Errorf("Expected pending, got %s", ot.Status)
	}
	if len(otRepo.requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(otRepo.requests))
	}
}

func TestOvertimeRequestService_Create_InvalidDate(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	_, err := svc.Create(context.Background(), uuid.New(), &model.OvertimeRequestCreate{
		Date: "invalid", PlannedMinutes: 60, Reason: "Test",
	})
	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

func TestOvertimeRequestService_Create_RepoError(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	otRepo.createErr = errors.New("db error")
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	_, err := svc.Create(context.Background(), uuid.New(), &model.OvertimeRequestCreate{
		Date: "2024-01-15", PlannedMinutes: 60, Reason: "Test",
	})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOvertimeRequestService_Approve_Success(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	reqID := uuid.New()
	userID := uuid.New()
	otRepo.requests[reqID] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: reqID}, UserID: userID,
		Status: model.OvertimeStatusPending,
		Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	approverID := uuid.New()
	result, err := svc.Approve(context.Background(), reqID, approverID, &model.OvertimeRequestApproval{
		Status: model.OvertimeStatusApproved,
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if result.Status != model.OvertimeStatusApproved {
		t.Errorf("Expected approved, got %s", result.Status)
	}
	if result.ApprovedBy == nil || *result.ApprovedBy != approverID {
		t.Error("Expected approver to be set")
	}
}

func TestOvertimeRequestService_Approve_Rejected(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	reqID := uuid.New()
	otRepo.requests[reqID] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: reqID}, UserID: uuid.New(),
		Status: model.OvertimeStatusPending,
		Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	result, err := svc.Approve(context.Background(), reqID, uuid.New(), &model.OvertimeRequestApproval{
		Status: model.OvertimeStatusRejected, RejectedReason: "Not needed",
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if result.Status != model.OvertimeStatusRejected {
		t.Errorf("Expected rejected, got %s", result.Status)
	}
	if result.RejectedReason != "Not needed" {
		t.Errorf("Expected reason 'Not needed', got '%s'", result.RejectedReason)
	}
}

func TestOvertimeRequestService_Approve_NotFound(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	_, err := svc.Approve(context.Background(), uuid.New(), uuid.New(), &model.OvertimeRequestApproval{
		Status: model.OvertimeStatusApproved,
	})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOvertimeRequestService_Approve_AlreadyProcessed(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	reqID := uuid.New()
	otRepo.requests[reqID] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: reqID},
		Status:    model.OvertimeStatusApproved,
	}

	_, err := svc.Approve(context.Background(), reqID, uuid.New(), &model.OvertimeRequestApproval{
		Status: model.OvertimeStatusApproved,
	})
	if err == nil {
		t.Error("Expected error for already processed")
	}
}

func TestOvertimeRequestService_Approve_UpdateError(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	reqID := uuid.New()
	otRepo.requests[reqID] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: reqID}, UserID: uuid.New(),
		Status: model.OvertimeStatusPending,
		Date:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	otRepo.updateErr = errors.New("update failed")

	_, err := svc.Approve(context.Background(), reqID, uuid.New(), &model.OvertimeRequestApproval{
		Status: model.OvertimeStatusApproved,
	})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOvertimeRequestService_GetByUser(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	userID := uuid.New()
	otRepo.requests[uuid.New()] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID,
		Status: model.OvertimeStatusPending,
	}

	results, total, err := svc.GetByUser(context.Background(), userID, 1, 20)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1, got %d", len(results))
	}
}

func TestOvertimeRequestService_GetPending(t *testing.T) {
	deps, otRepo, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	otRepo.requests[uuid.New()] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Status:    model.OvertimeStatusPending,
	}
	otRepo.requests[uuid.New()] = &model.OvertimeRequest{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Status:    model.OvertimeStatusApproved,
	}

	results, total, err := svc.GetPending(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1, got %d", len(results))
	}
}

func TestOvertimeRequestService_GetOvertimeAlerts(t *testing.T) {
	deps, otRepo, userRepo := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	// monthly overtime > 35h (2100 minutes)
	otRepo.monthlyOvertime[userID] = 2400

	alerts, err := svc.GetOvertimeAlerts(context.Background())
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
	if len(alerts) > 0 && alerts[0].MonthlyOvertimeHours != 40.0 {
		t.Errorf("Expected 40h, got %.1f", alerts[0].MonthlyOvertimeHours)
	}
}

func TestOvertimeRequestService_GetOvertimeAlerts_NoAlerts(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	// monthly overtime < 35h

	alerts, err := svc.GetOvertimeAlerts(context.Background())
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("Expected 0 alerts, got %d", len(alerts))
	}
}

func TestOvertimeRequestService_GetOvertimeAlerts_YearlyExceeded(t *testing.T) {
	deps, otRepo, userRepo := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewOvertimeRequestService(deps, notifSvc)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	// yearly overtime > 360h (21600 minutes) but monthly < 35h
	otRepo.yearlyOvertime[userID] = 22000

	alerts, err := svc.GetOvertimeAlerts(context.Background())
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
}

// ===== AttendanceCorrectionService Tests =====

func TestAttendanceCorrectionService_Create_Success(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	userID := uuid.New()
	clockIn := "09:00"
	clockOut := "18:00"
	result, err := svc.Create(context.Background(), userID, &model.AttendanceCorrectionCreate{
		Date: "2024-01-15", CorrectedClockIn: &clockIn, CorrectedClockOut: &clockOut, Reason: "間違い",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.Status != model.CorrectionStatusPending {
		t.Errorf("Expected pending, got %s", result.Status)
	}
	if result.CorrectedClockIn == nil {
		t.Error("Expected corrected clock in to be set")
	}
	if result.CorrectedClockOut == nil {
		t.Error("Expected corrected clock out to be set")
	}
}

func TestAttendanceCorrectionService_Create_WithFullTimestamp(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	clockIn := "2024-01-15T09:00:00"
	clockOut := "2024-01-15T18:00:00"
	result, err := svc.Create(context.Background(), uuid.New(), &model.AttendanceCorrectionCreate{
		Date: "2024-01-15", CorrectedClockIn: &clockIn, CorrectedClockOut: &clockOut, Reason: "修正",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.CorrectedClockIn == nil {
		t.Error("Expected corrected clock in")
	}
}

func TestAttendanceCorrectionService_Create_WithExistingAttendance(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	userID := uuid.New()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	clockInTime := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	attID := uuid.New()
	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	att := &model.Attendance{
		BaseModel: model.BaseModel{ID: attID}, UserID: userID, Date: date,
		ClockIn: &clockInTime,
	}
	attRepo.Attendances[attID] = att
	attRepo.UserDateIndex[userID.String()+date.Format("2006-01-02")] = att

	clockIn := "09:00"
	result, err := svc.Create(context.Background(), userID, &model.AttendanceCorrectionCreate{
		Date: "2024-01-15", CorrectedClockIn: &clockIn, Reason: "修正",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if result.AttendanceID == nil {
		t.Error("Expected attendance ID to be set")
	}
	if result.OriginalClockIn == nil {
		t.Error("Expected original clock in")
	}
}

func TestAttendanceCorrectionService_Create_InvalidDate(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	_, err := svc.Create(context.Background(), uuid.New(), &model.AttendanceCorrectionCreate{
		Date: "invalid", Reason: "Test",
	})
	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

func TestAttendanceCorrectionService_Create_RepoError(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	acRepo.createErr = errors.New("db error")
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	_, err := svc.Create(context.Background(), uuid.New(), &model.AttendanceCorrectionCreate{
		Date: "2024-01-15", Reason: "Test",
	})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestAttendanceCorrectionService_Approve_Success(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	cID := uuid.New()
	attID := uuid.New()
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: cID}, UserID: uuid.New(),
		AttendanceID: &attID, Status: model.CorrectionStatusPending,
		Date:              time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn:  &clockIn,
		CorrectedClockOut: &clockOut,
	}

	// Add attendance record
	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	origIn := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	origOut := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		ClockIn:   &origIn, ClockOut: &origOut,
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
	// Check attendance was updated
	att := attRepo.Attendances[attID]
	if att.ClockIn.Hour() != 9 {
		t.Errorf("Expected clock in hour 9, got %d", att.ClockIn.Hour())
	}
}

func TestAttendanceCorrectionService_Approve_Rejected(t *testing.T) {
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

	result, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusRejected, RejectedReason: "Not valid",
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}
	if result.Status != model.CorrectionStatusRejected {
		t.Errorf("Expected rejected, got %s", result.Status)
	}
}

func TestAttendanceCorrectionService_Approve_NotFound(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	_, err := svc.Approve(context.Background(), uuid.New(), uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestAttendanceCorrectionService_Approve_AlreadyProcessed(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	cID := uuid.New()
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: cID},
		Status:    model.CorrectionStatusApproved,
	}

	_, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err == nil {
		t.Error("Expected error for already processed")
	}
}

func TestAttendanceCorrectionService_Approve_NoAttendanceID(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	cID := uuid.New()
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: cID}, UserID: uuid.New(),
		Status:            model.CorrectionStatusPending,
		Date:              time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn:  &clockIn,
		CorrectedClockOut: &clockOut,
		// No AttendanceID - triggers new attendance creation branch
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

func TestAttendanceCorrectionService_Approve_OvertimeCalculation(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	cID := uuid.New()
	attID := uuid.New()
	// 10 hours of work (600 minutes > 480 = overtime 120)
	clockIn := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	acRepo.corrections[cID] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: cID}, UserID: uuid.New(),
		AttendanceID:      &attID,
		Status:            model.CorrectionStatusPending,
		Date:              time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		CorrectedClockIn:  &clockIn,
		CorrectedClockOut: &clockOut,
	}

	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	origIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	origOut := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)
	attRepo.Attendances[attID] = &model.Attendance{
		BaseModel: model.BaseModel{ID: attID},
		ClockIn:   &origIn, ClockOut: &origOut, BreakMinutes: 0,
	}

	_, err := svc.Approve(context.Background(), cID, uuid.New(), &model.AttendanceCorrectionApproval{
		Status: model.CorrectionStatusApproved,
	})
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	att := attRepo.Attendances[attID]
	if att.OvertimeMinutes <= 0 {
		t.Errorf("Expected overtime > 0, got %d", att.OvertimeMinutes)
	}
}

func TestAttendanceCorrectionService_GetByUser(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	userID := uuid.New()
	acRepo.corrections[uuid.New()] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID,
	}

	results, total, err := svc.GetByUser(context.Background(), userID, 1, 20)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Errorf("Expected 1, got total=%d len=%d", total, len(results))
	}
}

func TestAttendanceCorrectionService_GetPending(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	notifSvc := NewNotificationService(deps)
	svc := NewAttendanceCorrectionService(deps, notifSvc)

	acRepo := deps.Repos.AttendanceCorrection.(*mockAttendanceCorrectionRepo)
	acRepo.corrections[uuid.New()] = &model.AttendanceCorrection{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Status:    model.CorrectionStatusPending,
	}

	results, total, err := svc.GetPending(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Errorf("Expected 1, got total=%d len=%d", total, len(results))
	}
}

// ===== ExportService Tests =====

func TestExportService_ExportAttendanceCSV_WithUserID(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	attRepo := deps.Repos.Attendance.(*mocks.MockAttendanceRepository)
	clockIn := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	clockOut := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	attRepo.Attendances[uuid.New()] = &model.Attendance{
		BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID,
		Date:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		ClockIn: &clockIn, ClockOut: &clockOut,
		WorkMinutes: 540, OvertimeMinutes: 60,
		Status: model.AttendanceStatusPresent,
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
	// Verify BOM
	if data[0] != 0xEF || data[1] != 0xBB || data[2] != 0xBF {
		t.Error("Expected BOM header")
	}
}

func TestExportService_ExportAttendanceCSV_AllUsers(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

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

func TestExportService_ExportLeavesCSV(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
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

func TestExportService_ExportLeavesCSV_WithSpecificUser(t *testing.T) {
	deps, _, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)
	ctx := context.Background()

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportLeavesCSV(ctx, &userID, start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
}

func TestExportService_ExportOvertimeCSV(t *testing.T) {
	deps, otRepo, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	otRepo.monthlyOvertime[userID] = 3000 // 50h > 45h monthly limit

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportOvertimeCSV(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty CSV")
	}
	content := string(data)
	if !contains(content, "月間上限超過") {
		t.Error("Expected monthly limit exceeded warning")
	}
}

func TestExportService_ExportOvertimeCSV_YearlyExceeded(t *testing.T) {
	deps, otRepo, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	otRepo.yearlyOvertime[userID] = 22000 // 366h > 360h limit

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportOvertimeCSV(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	if !contains(content, "年間上限超過") {
		t.Error("Expected yearly limit exceeded warning")
	}
}

func TestExportService_ExportOvertimeCSV_BothExceeded(t *testing.T) {
	deps, otRepo, userRepo := setupOvertimeDeps(t)
	svc := NewExportService(deps)

	userID := uuid.New()
	userRepo.Users[userID] = &model.User{
		BaseModel: model.BaseModel{ID: userID},
		FirstName: "Taro", LastName: "Test",
	}
	otRepo.monthlyOvertime[userID] = 3000
	otRepo.yearlyOvertime[userID] = 22000

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	data, err := svc.ExportOvertimeCSV(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	content := string(data)
	if !contains(content, "月間上限超過") || !contains(content, "年間上限超過") {
		t.Error("Expected both warning messages")
	}
}

func TestExportService_ExportProjectsCSV(t *testing.T) {
	deps, _, _ := setupOvertimeDeps(t)
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

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
