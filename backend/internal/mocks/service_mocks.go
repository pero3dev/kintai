package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

// ===== MockAuthService =====

type MockAuthService struct {
	LoginFunc        func(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error)
	RegisterFunc     func(ctx context.Context, req *model.RegisterRequest) (*model.User, error)
	RefreshTokenFunc func(ctx context.Context, refreshToken string) (*model.TokenResponse, error)
	LogoutFunc       func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockAuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(ctx, userID)
	}
	return nil
}

// ===== MockAttendanceService =====

type MockAttendanceService struct {
	ClockInFunc               func(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error)
	ClockOutFunc              func(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error)
	GetByUserAndDateRangeFunc func(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error)
	GetSummaryFunc            func(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error)
	GetTodayStatusFunc        func(ctx context.Context, userID uuid.UUID) (*model.Attendance, error)
}

func (m *MockAttendanceService) ClockIn(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error) {
	if m.ClockInFunc != nil {
		return m.ClockInFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockAttendanceService) ClockOut(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error) {
	if m.ClockOutFunc != nil {
		return m.ClockOutFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockAttendanceService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
	if m.GetByUserAndDateRangeFunc != nil {
		return m.GetByUserAndDateRangeFunc(ctx, userID, start, end, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockAttendanceService) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
	if m.GetSummaryFunc != nil {
		return m.GetSummaryFunc(ctx, userID, start, end)
	}
	return nil, nil
}

func (m *MockAttendanceService) GetTodayStatus(ctx context.Context, userID uuid.UUID) (*model.Attendance, error) {
	if m.GetTodayStatusFunc != nil {
		return m.GetTodayStatusFunc(ctx, userID)
	}
	return nil, nil
}

// ===== MockLeaveService =====

type MockLeaveService struct {
	CreateFunc     func(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error)
	ApproveFunc    func(ctx context.Context, leaveID uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error)
	GetByUserFunc  func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error)
	GetPendingFunc func(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error)
}

func (m *MockLeaveService) Create(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockLeaveService) Approve(ctx context.Context, leaveID uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error) {
	if m.ApproveFunc != nil {
		return m.ApproveFunc(ctx, leaveID, approverID, req)
	}
	return nil, nil
}

func (m *MockLeaveService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	if m.GetByUserFunc != nil {
		return m.GetByUserFunc(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockLeaveService) GetPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	if m.GetPendingFunc != nil {
		return m.GetPendingFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

// ===== MockShiftService =====

type MockShiftService struct {
	CreateFunc                func(ctx context.Context, req *model.ShiftCreateRequest) (*model.Shift, error)
	BulkCreateFunc            func(ctx context.Context, req *model.ShiftBulkCreateRequest) error
	GetByUserAndDateRangeFunc func(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error)
	GetByDateRangeFunc        func(ctx context.Context, start, end time.Time) ([]model.Shift, error)
	DeleteFunc                func(ctx context.Context, id uuid.UUID) error
}

func (m *MockShiftService) Create(ctx context.Context, req *model.ShiftCreateRequest) (*model.Shift, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockShiftService) BulkCreate(ctx context.Context, req *model.ShiftBulkCreateRequest) error {
	if m.BulkCreateFunc != nil {
		return m.BulkCreateFunc(ctx, req)
	}
	return nil
}

func (m *MockShiftService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error) {
	if m.GetByUserAndDateRangeFunc != nil {
		return m.GetByUserAndDateRangeFunc(ctx, userID, start, end)
	}
	return nil, nil
}

func (m *MockShiftService) GetByDateRange(ctx context.Context, start, end time.Time) ([]model.Shift, error) {
	if m.GetByDateRangeFunc != nil {
		return m.GetByDateRangeFunc(ctx, start, end)
	}
	return nil, nil
}

func (m *MockShiftService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockUserService =====

type MockUserService struct {
	CreateFunc  func(ctx context.Context, req *model.UserCreateRequest) (*model.User, error)
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetAllFunc  func(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
	UpdateFunc  func(ctx context.Context, id uuid.UUID, req *model.UserUpdateRequest) (*model.User, error)
	DeleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserService) Create(ctx context.Context, req *model.UserCreateRequest) (*model.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserService) GetAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockUserService) Update(ctx context.Context, id uuid.UUID, req *model.UserUpdateRequest) (*model.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil, nil
}

func (m *MockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockDepartmentService =====

type MockDepartmentService struct {
	CreateFunc  func(ctx context.Context, dept *model.Department) (*model.Department, error)
	GetAllFunc  func(ctx context.Context) ([]model.Department, error)
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*model.Department, error)
	UpdateFunc  func(ctx context.Context, dept *model.Department) (*model.Department, error)
	DeleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockDepartmentService) Create(ctx context.Context, dept *model.Department) (*model.Department, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, dept)
	}
	return nil, nil
}

func (m *MockDepartmentService) GetAll(ctx context.Context) ([]model.Department, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockDepartmentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Department, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDepartmentService) Update(ctx context.Context, dept *model.Department) (*model.Department, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, dept)
	}
	return nil, nil
}

func (m *MockDepartmentService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// ===== MockDashboardService =====

type MockDashboardService struct {
	GetStatsFunc func(ctx context.Context) (*model.DashboardStatsExtended, error)
}

func (m *MockDashboardService) GetStats(ctx context.Context) (*model.DashboardStatsExtended, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx)
	}
	return nil, nil
}

// ===== MockNotificationService =====

type MockNotificationService struct {
	SendFunc          func(ctx context.Context, userID uuid.UUID, notifType model.NotificationType, title, message string) error
	GetByUserFunc     func(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error)
	MarkAsReadFunc    func(ctx context.Context, id uuid.UUID) error
	MarkAllAsReadFunc func(ctx context.Context, userID uuid.UUID) error
	GetUnreadCountFunc func(ctx context.Context, userID uuid.UUID) (int64, error)
	DeleteFunc        func(ctx context.Context, id uuid.UUID) error
}

func (m *MockNotificationService) Send(ctx context.Context, userID uuid.UUID, notifType model.NotificationType, title, message string) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, userID, notifType, title, message)
	}
	return nil
}

func (m *MockNotificationService) GetByUser(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
	if m.GetByUserFunc != nil {
		return m.GetByUserFunc(ctx, userID, isRead, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(ctx, id)
	}
	return nil
}

func (m *MockNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.MarkAllAsReadFunc != nil {
		return m.MarkAllAsReadFunc(ctx, userID)
	}
	return nil
}

func (m *MockNotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.GetUnreadCountFunc != nil {
		return m.GetUnreadCountFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockNotificationService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

