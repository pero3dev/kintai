package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

// MockUserRepository はUserRepositoryのモック
type MockUserRepository struct {
	Users         map[uuid.UUID]*model.User
	UsersByEmail  map[string]*model.User
	CreateErr     error
	FindByIDErr   error
	FindByEmailErr error
	FindAllErr    error
	UpdateErr     error
	DeleteErr     error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:        make(map[uuid.UUID]*model.User),
		UsersByEmail: make(map[string]*model.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.Users[user.ID] = user
	m.UsersByEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	user, ok := m.Users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.FindByEmailErr != nil {
		return nil, m.FindByEmailErr
	}
	user, ok := m.UsersByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	if m.FindAllErr != nil {
		return nil, 0, m.FindAllErr
	}
	users := make([]model.User, 0, len(m.Users))
	for _, u := range m.Users {
		users = append(users, *u)
	}
	return users, int64(len(users)), nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.Users[user.ID] = user
	m.UsersByEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	user, ok := m.Users[id]
	if ok {
		delete(m.UsersByEmail, user.Email)
	}
	delete(m.Users, id)
	return nil
}

func (m *MockUserRepository) FindByDepartmentID(ctx context.Context, departmentID uuid.UUID) ([]model.User, error) {
	users := make([]model.User, 0)
	for _, u := range m.Users {
		if u.DepartmentID != nil && *u.DepartmentID == departmentID {
			users = append(users, *u)
		}
	}
	return users, nil
}

// MockAttendanceRepository はAttendanceRepositoryのモック
type MockAttendanceRepository struct {
	Attendances       map[uuid.UUID]*model.Attendance
	UserDateIndex     map[string]*model.Attendance
	CreateErr         error
	FindByIDErr       error
	FindByUserDateErr error
	UpdateErr         error
}

func NewMockAttendanceRepository() *MockAttendanceRepository {
	return &MockAttendanceRepository{
		Attendances:   make(map[uuid.UUID]*model.Attendance),
		UserDateIndex: make(map[string]*model.Attendance),
	}
}

func (m *MockAttendanceRepository) Create(ctx context.Context, attendance *model.Attendance) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	if attendance.ID == uuid.Nil {
		attendance.ID = uuid.New()
	}
	m.Attendances[attendance.ID] = attendance
	key := attendance.UserID.String() + attendance.Date.Format("2006-01-02")
	m.UserDateIndex[key] = attendance
	return nil
}

func (m *MockAttendanceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Attendance, error) {
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	att, ok := m.Attendances[id]
	if !ok {
		return nil, ErrNotFound
	}
	return att, nil
}

func (m *MockAttendanceRepository) FindByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*model.Attendance, error) {
	if m.FindByUserDateErr != nil {
		return nil, m.FindByUserDateErr
	}
	key := userID.String() + date.Format("2006-01-02")
	att, ok := m.UserDateIndex[key]
	if !ok {
		return nil, ErrNotFound
	}
	return att, nil
}

func (m *MockAttendanceRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
	attendances := make([]model.Attendance, 0)
	for _, att := range m.Attendances {
		if att.UserID == userID && !att.Date.Before(start) && !att.Date.After(end) {
			attendances = append(attendances, *att)
		}
	}
	return attendances, int64(len(attendances)), nil
}

func (m *MockAttendanceRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Attendance, error) {
	attendances := make([]model.Attendance, 0)
	for _, att := range m.Attendances {
		if !att.Date.Before(start) && !att.Date.After(end) {
			attendances = append(attendances, *att)
		}
	}
	return attendances, nil
}

func (m *MockAttendanceRepository) Update(ctx context.Context, attendance *model.Attendance) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.Attendances[attendance.ID] = attendance
	key := attendance.UserID.String() + attendance.Date.Format("2006-01-02")
	m.UserDateIndex[key] = attendance
	return nil
}

func (m *MockAttendanceRepository) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
	summary := &model.AttendanceSummary{}
	for _, att := range m.Attendances {
		if att.UserID == userID && !att.Date.Before(start) && !att.Date.After(end) {
			if att.Status == model.AttendanceStatusPresent {
				summary.TotalWorkDays++
				summary.TotalWorkMinutes += att.WorkMinutes
				summary.TotalOvertimeMinutes += att.OvertimeMinutes
			} else if att.Status == model.AttendanceStatusAbsent {
				summary.AbsentDays++
			} else if att.Status == model.AttendanceStatusLeave {
				summary.LeaveDays++
			}
		}
	}
	if summary.TotalWorkDays > 0 {
		summary.AverageWorkMinutes = float64(summary.TotalWorkMinutes) / float64(summary.TotalWorkDays)
	}
	return summary, nil
}

func (m *MockAttendanceRepository) CountTodayPresent(ctx context.Context) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	for _, att := range m.Attendances {
		if att.Date.Equal(today) && att.Status == model.AttendanceStatusPresent {
			count++
		}
	}
	return count, nil
}

func (m *MockAttendanceRepository) CountTodayAbsent(ctx context.Context, totalUsers int64) (int64, error) {
	present, _ := m.CountTodayPresent(ctx)
	return totalUsers - present, nil
}

func (m *MockAttendanceRepository) GetMonthlyOvertime(ctx context.Context, start, end time.Time) (int64, error) {
	var total int64
	for _, att := range m.Attendances {
		if !att.Date.Before(start) && !att.Date.After(end) {
			total += int64(att.OvertimeMinutes)
		}
	}
	return total, nil
}

// MockLeaveRequestRepository はLeaveRequestRepositoryのモック
type MockLeaveRequestRepository struct {
	LeaveRequests map[uuid.UUID]*model.LeaveRequest
	CreateErr     error
	FindByIDErr   error
	UpdateErr     error
}

func NewMockLeaveRequestRepository() *MockLeaveRequestRepository {
	return &MockLeaveRequestRepository{
		LeaveRequests: make(map[uuid.UUID]*model.LeaveRequest),
	}
}

func (m *MockLeaveRequestRepository) Create(ctx context.Context, leave *model.LeaveRequest) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	if leave.ID == uuid.Nil {
		leave.ID = uuid.New()
	}
	m.LeaveRequests[leave.ID] = leave
	return nil
}

func (m *MockLeaveRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveRequest, error) {
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	leave, ok := m.LeaveRequests[id]
	if !ok {
		return nil, ErrNotFound
	}
	return leave, nil
}

func (m *MockLeaveRequestRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	leaves := make([]model.LeaveRequest, 0)
	for _, l := range m.LeaveRequests {
		if l.UserID == userID {
			leaves = append(leaves, *l)
		}
	}
	return leaves, int64(len(leaves)), nil
}

func (m *MockLeaveRequestRepository) FindPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	leaves := make([]model.LeaveRequest, 0)
	for _, l := range m.LeaveRequests {
		if l.Status == model.ApprovalStatusPending {
			leaves = append(leaves, *l)
		}
	}
	return leaves, int64(len(leaves)), nil
}

func (m *MockLeaveRequestRepository) Update(ctx context.Context, leave *model.LeaveRequest) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.LeaveRequests[leave.ID] = leave
	return nil
}

func (m *MockLeaveRequestRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	for _, l := range m.LeaveRequests {
		if l.Status == model.ApprovalStatusPending {
			count++
		}
	}
	return count, nil
}

// MockShiftRepository はShiftRepositoryのモック
type MockShiftRepository struct {
	Shifts        map[uuid.UUID]*model.Shift
	CreateErr     error
	BulkCreateErr error
	DeleteErr     error
}

func NewMockShiftRepository() *MockShiftRepository {
	return &MockShiftRepository{
		Shifts: make(map[uuid.UUID]*model.Shift),
	}
}

func (m *MockShiftRepository) Create(ctx context.Context, shift *model.Shift) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	if shift.ID == uuid.Nil {
		shift.ID = uuid.New()
	}
	m.Shifts[shift.ID] = shift
	return nil
}

func (m *MockShiftRepository) BulkCreate(ctx context.Context, shifts []model.Shift) error {
	if m.BulkCreateErr != nil {
		return m.BulkCreateErr
	}
	for i := range shifts {
		shifts[i].ID = uuid.New()
		m.Shifts[shifts[i].ID] = &shifts[i]
	}
	return nil
}

func (m *MockShiftRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Shift, error) {
	shift, ok := m.Shifts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return shift, nil
}

func (m *MockShiftRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.Shift, error) {
	shifts := make([]model.Shift, 0)
	for _, s := range m.Shifts {
		if s.UserID == userID && !s.Date.Before(start) && !s.Date.After(end) {
			shifts = append(shifts, *s)
		}
	}
	return shifts, nil
}

func (m *MockShiftRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Shift, error) {
	shifts := make([]model.Shift, 0)
	for _, s := range m.Shifts {
		if !s.Date.Before(start) && !s.Date.After(end) {
			shifts = append(shifts, *s)
		}
	}
	return shifts, nil
}

func (m *MockShiftRepository) Update(ctx context.Context, shift *model.Shift) error {
	m.Shifts[shift.ID] = shift
	return nil
}

func (m *MockShiftRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Shifts, id)
	return nil
}

func (m *MockShiftRepository) DeleteByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) error {
	for id, s := range m.Shifts {
		if s.UserID == userID && !s.Date.Before(start) && !s.Date.After(end) {
			delete(m.Shifts, id)
		}
	}
	return nil
}

// MockDepartmentRepository はDepartmentRepositoryのモック
type MockDepartmentRepository struct {
	Departments map[uuid.UUID]*model.Department
	CreateErr   error
	UpdateErr   error
	DeleteErr   error
}

func NewMockDepartmentRepository() *MockDepartmentRepository {
	return &MockDepartmentRepository{
		Departments: make(map[uuid.UUID]*model.Department),
	}
}

func (m *MockDepartmentRepository) Create(ctx context.Context, dept *model.Department) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	if dept.ID == uuid.Nil {
		dept.ID = uuid.New()
	}
	m.Departments[dept.ID] = dept
	return nil
}

func (m *MockDepartmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Department, error) {
	dept, ok := m.Departments[id]
	if !ok {
		return nil, ErrNotFound
	}
	return dept, nil
}

func (m *MockDepartmentRepository) FindAll(ctx context.Context) ([]model.Department, error) {
	depts := make([]model.Department, 0, len(m.Departments))
	for _, d := range m.Departments {
		depts = append(depts, *d)
	}
	return depts, nil
}

func (m *MockDepartmentRepository) Update(ctx context.Context, dept *model.Department) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	m.Departments[dept.ID] = dept
	return nil
}

func (m *MockDepartmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Departments, id)
	return nil
}

// MockRefreshTokenRepository はRefreshTokenRepositoryのモック
type MockRefreshTokenRepository struct {
	Tokens    map[string]*model.RefreshToken
	UserTokens map[uuid.UUID][]string
	RevokeErr error
	CreateErr error
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		Tokens:     make(map[string]*model.RefreshToken),
		UserTokens: make(map[uuid.UUID][]string),
	}
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	if m.CreateErr != nil {
		return m.CreateErr
	}
	m.Tokens[token.Token] = token
	m.UserTokens[token.UserID] = append(m.UserTokens[token.UserID], token.Token)
	return nil
}

func (m *MockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	rt, ok := m.Tokens[token]
	if !ok || rt.IsRevoked {
		return nil, ErrNotFound
	}
	return rt, nil
}

func (m *MockRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.RevokeErr != nil {
		return m.RevokeErr
	}
	for _, tokenStr := range m.UserTokens[userID] {
		if rt, ok := m.Tokens[tokenStr]; ok {
			rt.IsRevoked = true
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	if m.RevokeErr != nil {
		return m.RevokeErr
	}
	if rt, ok := m.Tokens[token]; ok {
		rt.IsRevoked = true
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for key, rt := range m.Tokens {
		if rt.ExpiresAt.Before(now) || rt.IsRevoked {
			delete(m.Tokens, key)
		}
	}
	return nil
}

// エラー定義
var ErrNotFound = errors.New("not found")
