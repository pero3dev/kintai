package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// ===== Mock Repositories for Extended Services =====

// --- OvertimeRequestRepository mock ---
type mockOvertimeRequestRepo struct {
	requests         map[uuid.UUID]*model.OvertimeRequest
	createErr        error
	findByIDErr      error
	updateErr        error
	monthlyOvertime  map[uuid.UUID]int64
	yearlyOvertime   map[uuid.UUID]int64
}

func newMockOvertimeRequestRepo() *mockOvertimeRequestRepo {
	return &mockOvertimeRequestRepo{
		requests:        make(map[uuid.UUID]*model.OvertimeRequest),
		monthlyOvertime: make(map[uuid.UUID]int64),
		yearlyOvertime:  make(map[uuid.UUID]int64),
	}
}

func (m *mockOvertimeRequestRepo) Create(ctx context.Context, req *model.OvertimeRequest) error {
	if m.createErr != nil { return m.createErr }
	if req.ID == uuid.Nil { req.ID = uuid.New() }
	m.requests[req.ID] = req
	return nil
}

func (m *mockOvertimeRequestRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.OvertimeRequest, error) {
	if m.findByIDErr != nil { return nil, m.findByIDErr }
	r, ok := m.requests[id]
	if !ok { return nil, errors.New("not found") }
	return r, nil
}

func (m *mockOvertimeRequestRepo) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	var result []model.OvertimeRequest
	for _, r := range m.requests {
		if r.UserID == userID { result = append(result, *r) }
	}
	return result, int64(len(result)), nil
}

func (m *mockOvertimeRequestRepo) FindPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	var result []model.OvertimeRequest
	for _, r := range m.requests {
		if r.Status == model.OvertimeStatusPending { result = append(result, *r) }
	}
	return result, int64(len(result)), nil
}

func (m *mockOvertimeRequestRepo) Update(ctx context.Context, req *model.OvertimeRequest) error {
	if m.updateErr != nil { return m.updateErr }
	m.requests[req.ID] = req
	return nil
}

func (m *mockOvertimeRequestRepo) CountPending(ctx context.Context) (int64, error) {
	var count int64
	for _, r := range m.requests {
		if r.Status == model.OvertimeStatusPending { count++ }
	}
	return count, nil
}

func (m *mockOvertimeRequestRepo) GetUserMonthlyOvertime(ctx context.Context, userID uuid.UUID, year, month int) (int64, error) {
	return m.monthlyOvertime[userID], nil
}

func (m *mockOvertimeRequestRepo) GetUserYearlyOvertime(ctx context.Context, userID uuid.UUID, year int) (int64, error) {
	return m.yearlyOvertime[userID], nil
}

// --- LeaveBalanceRepository mock ---
type mockLeaveBalanceRepo struct {
	balances    map[string]*model.LeaveBalance // key: userID-year-type
	createErr   error
	updateErr   error
	upsertErr   error
	findErr     error
}

func newMockLeaveBalanceRepo() *mockLeaveBalanceRepo {
	return &mockLeaveBalanceRepo{balances: make(map[string]*model.LeaveBalance)}
}

func (m *mockLeaveBalanceRepo) key(userID uuid.UUID, year int, lt model.LeaveType) string {
	return userID.String() + "-" + string(rune(year)) + "-" + string(lt)
}

func (m *mockLeaveBalanceRepo) Create(ctx context.Context, balance *model.LeaveBalance) error {
	if m.createErr != nil { return m.createErr }
	k := m.key(balance.UserID, balance.FiscalYear, balance.LeaveType)
	m.balances[k] = balance
	return nil
}

func (m *mockLeaveBalanceRepo) FindByUserAndYear(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalance, error) {
	if m.findErr != nil { return nil, m.findErr }
	var result []model.LeaveBalance
	for _, b := range m.balances {
		if b.UserID == userID && b.FiscalYear == fiscalYear {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (m *mockLeaveBalanceRepo) FindByUserYearAndType(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType) (*model.LeaveBalance, error) {
	if m.findErr != nil { return nil, m.findErr }
	for _, b := range m.balances {
		if b.UserID == userID && b.FiscalYear == fiscalYear && b.LeaveType == leaveType {
			return b, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockLeaveBalanceRepo) Update(ctx context.Context, balance *model.LeaveBalance) error {
	if m.updateErr != nil { return m.updateErr }
	k := m.key(balance.UserID, balance.FiscalYear, balance.LeaveType)
	m.balances[k] = balance
	return nil
}

func (m *mockLeaveBalanceRepo) Upsert(ctx context.Context, balance *model.LeaveBalance) error {
	if m.upsertErr != nil { return m.upsertErr }
	k := m.key(balance.UserID, balance.FiscalYear, balance.LeaveType)
	m.balances[k] = balance
	return nil
}

// --- AttendanceCorrectionRepository mock ---
type mockAttendanceCorrectionRepo struct {
	corrections map[uuid.UUID]*model.AttendanceCorrection
	createErr   error
	findByIDErr error
	updateErr   error
}

func newMockAttendanceCorrectionRepo() *mockAttendanceCorrectionRepo {
	return &mockAttendanceCorrectionRepo{corrections: make(map[uuid.UUID]*model.AttendanceCorrection)}
}

func (m *mockAttendanceCorrectionRepo) Create(ctx context.Context, c *model.AttendanceCorrection) error {
	if m.createErr != nil { return m.createErr }
	if c.ID == uuid.Nil { c.ID = uuid.New() }
	m.corrections[c.ID] = c
	return nil
}

func (m *mockAttendanceCorrectionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceCorrection, error) {
	if m.findByIDErr != nil { return nil, m.findByIDErr }
	c, ok := m.corrections[id]
	if !ok { return nil, errors.New("not found") }
	return c, nil
}

func (m *mockAttendanceCorrectionRepo) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	var result []model.AttendanceCorrection
	for _, c := range m.corrections {
		if c.UserID == userID { result = append(result, *c) }
	}
	return result, int64(len(result)), nil
}

func (m *mockAttendanceCorrectionRepo) FindPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	var result []model.AttendanceCorrection
	for _, c := range m.corrections {
		if c.Status == model.CorrectionStatusPending { result = append(result, *c) }
	}
	return result, int64(len(result)), nil
}

func (m *mockAttendanceCorrectionRepo) Update(ctx context.Context, c *model.AttendanceCorrection) error {
	if m.updateErr != nil { return m.updateErr }
	m.corrections[c.ID] = c
	return nil
}

func (m *mockAttendanceCorrectionRepo) CountPending(ctx context.Context) (int64, error) {
	var count int64
	for _, c := range m.corrections {
		if c.Status == model.CorrectionStatusPending { count++ }
	}
	return count, nil
}

// --- NotificationRepository mock ---
type mockNotificationRepo struct {
	notifications map[uuid.UUID]*model.Notification
	createErr     error
	deleteErr     error
	markReadErr   error
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{notifications: make(map[uuid.UUID]*model.Notification)}
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *model.Notification) error {
	if m.createErr != nil { return m.createErr }
	if n.ID == uuid.Nil { n.ID = uuid.New() }
	m.notifications[n.ID] = n
	return nil
}

func (m *mockNotificationRepo) FindByUserID(ctx context.Context, userID uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
	var result []model.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			if isRead != nil && n.IsRead != *isRead { continue }
			result = append(result, *n)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.markReadErr != nil { return m.markReadErr }
	if n, ok := m.notifications[id]; ok { n.IsRead = true }
	return nil
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.markReadErr != nil { return m.markReadErr }
	for _, n := range m.notifications {
		if n.UserID == userID { n.IsRead = true }
	}
	return nil
}

func (m *mockNotificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead { count++ }
	}
	return count, nil
}

func (m *mockNotificationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil { return m.deleteErr }
	delete(m.notifications, id)
	return nil
}

// --- ProjectRepository mock ---
type mockProjectRepo struct {
	projects    map[uuid.UUID]*model.Project
	createErr   error
	findByIDErr error
	updateErr   error
	deleteErr   error
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{projects: make(map[uuid.UUID]*model.Project)}
}

func (m *mockProjectRepo) Create(ctx context.Context, p *model.Project) error {
	if m.createErr != nil { return m.createErr }
	if p.ID == uuid.Nil { p.ID = uuid.New() }
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	if m.findByIDErr != nil { return nil, m.findByIDErr }
	p, ok := m.projects[id]
	if !ok { return nil, errors.New("not found") }
	return p, nil
}

func (m *mockProjectRepo) FindAll(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error) {
	var result []model.Project
	for _, p := range m.projects {
		if status != nil && p.Status != *status { continue }
		result = append(result, *p)
	}
	return result, int64(len(result)), nil
}

func (m *mockProjectRepo) Update(ctx context.Context, p *model.Project) error {
	if m.updateErr != nil { return m.updateErr }
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil { return m.deleteErr }
	delete(m.projects, id)
	return nil
}

// --- TimeEntryRepository mock ---
type mockTimeEntryRepo struct {
	entries     map[uuid.UUID]*model.TimeEntry
	createErr   error
	findByIDErr error
	updateErr   error
	deleteErr   error
}

func newMockTimeEntryRepo() *mockTimeEntryRepo {
	return &mockTimeEntryRepo{entries: make(map[uuid.UUID]*model.TimeEntry)}
}

func (m *mockTimeEntryRepo) Create(ctx context.Context, e *model.TimeEntry) error {
	if m.createErr != nil { return m.createErr }
	if e.ID == uuid.Nil { e.ID = uuid.New() }
	m.entries[e.ID] = e
	return nil
}

func (m *mockTimeEntryRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.TimeEntry, error) {
	if m.findByIDErr != nil { return nil, m.findByIDErr }
	e, ok := m.entries[id]
	if !ok { return nil, errors.New("not found") }
	return e, nil
}

func (m *mockTimeEntryRepo) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	var result []model.TimeEntry
	for _, e := range m.entries {
		if e.UserID == userID && !e.Date.Before(start) && !e.Date.After(end) {
			result = append(result, *e)
		}
	}
	return result, nil
}

func (m *mockTimeEntryRepo) FindByProjectAndDateRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
	var result []model.TimeEntry
	for _, e := range m.entries {
		if e.ProjectID == projectID && !e.Date.Before(start) && !e.Date.After(end) {
			result = append(result, *e)
		}
	}
	return result, nil
}

func (m *mockTimeEntryRepo) Update(ctx context.Context, e *model.TimeEntry) error {
	if m.updateErr != nil { return m.updateErr }
	m.entries[e.ID] = e
	return nil
}

func (m *mockTimeEntryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil { return m.deleteErr }
	delete(m.entries, id)
	return nil
}

func (m *mockTimeEntryRepo) GetProjectSummary(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
	return []model.ProjectSummary{}, nil
}

// --- HolidayRepository mock ---
type mockHolidayRepo struct {
	holidays    map[uuid.UUID]*model.Holiday
	createErr   error
	findByIDErr error
	updateErr   error
	deleteErr   error
}

func newMockHolidayRepo() *mockHolidayRepo {
	return &mockHolidayRepo{holidays: make(map[uuid.UUID]*model.Holiday)}
}

func (m *mockHolidayRepo) Create(ctx context.Context, h *model.Holiday) error {
	if m.createErr != nil { return m.createErr }
	if h.ID == uuid.Nil { h.ID = uuid.New() }
	m.holidays[h.ID] = h
	return nil
}

func (m *mockHolidayRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Holiday, error) {
	if m.findByIDErr != nil { return nil, m.findByIDErr }
	h, ok := m.holidays[id]
	if !ok { return nil, errors.New("not found") }
	return h, nil
}

func (m *mockHolidayRepo) FindByDateRange(ctx context.Context, start, end time.Time) ([]model.Holiday, error) {
	var result []model.Holiday
	for _, h := range m.holidays {
		if !h.Date.Before(start) && !h.Date.After(end) {
			result = append(result, *h)
		}
	}
	return result, nil
}

func (m *mockHolidayRepo) FindByYear(ctx context.Context, year int) ([]model.Holiday, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year, 12, 31, 23, 59, 59, 0, time.Local)
	return m.FindByDateRange(ctx, start, end)
}

func (m *mockHolidayRepo) Update(ctx context.Context, h *model.Holiday) error {
	if m.updateErr != nil { return m.updateErr }
	m.holidays[h.ID] = h
	return nil
}

func (m *mockHolidayRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil { return m.deleteErr }
	delete(m.holidays, id)
	return nil
}

func (m *mockHolidayRepo) IsHoliday(ctx context.Context, date time.Time) (bool, *model.Holiday, error) {
	for _, h := range m.holidays {
		if h.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			return true, h, nil
		}
	}
	return false, nil, nil
}

// --- ApprovalFlowRepository mock ---
type mockApprovalFlowRepo struct {
	flows       map[uuid.UUID]*model.ApprovalFlow
	steps       map[uuid.UUID][]model.ApprovalStep
	createErr   error
	findByIDErr error
	updateErr   error
	deleteErr   error
}

func newMockApprovalFlowRepo() *mockApprovalFlowRepo {
	return &mockApprovalFlowRepo{
		flows: make(map[uuid.UUID]*model.ApprovalFlow),
		steps: make(map[uuid.UUID][]model.ApprovalStep),
	}
}

func (m *mockApprovalFlowRepo) Create(ctx context.Context, f *model.ApprovalFlow) error {
	if m.createErr != nil { return m.createErr }
	if f.ID == uuid.Nil { f.ID = uuid.New() }
	m.flows[f.ID] = f
	return nil
}

func (m *mockApprovalFlowRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error) {
	if m.findByIDErr != nil { return nil, m.findByIDErr }
	f, ok := m.flows[id]
	if !ok { return nil, errors.New("not found") }
	f.Steps = m.steps[id]
	return f, nil
}

func (m *mockApprovalFlowRepo) FindByType(ctx context.Context, flowType model.ApprovalFlowType) ([]model.ApprovalFlow, error) {
	var result []model.ApprovalFlow
	for _, f := range m.flows {
		if f.FlowType == flowType && f.IsActive {
			f.Steps = m.steps[f.ID]
			result = append(result, *f)
		}
	}
	return result, nil
}

func (m *mockApprovalFlowRepo) FindAll(ctx context.Context) ([]model.ApprovalFlow, error) {
	var result []model.ApprovalFlow
	for _, f := range m.flows {
		f.Steps = m.steps[f.ID]
		result = append(result, *f)
	}
	return result, nil
}

func (m *mockApprovalFlowRepo) Update(ctx context.Context, f *model.ApprovalFlow) error {
	if m.updateErr != nil { return m.updateErr }
	m.flows[f.ID] = f
	return nil
}

func (m *mockApprovalFlowRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil { return m.deleteErr }
	delete(m.flows, id)
	return nil
}

func (m *mockApprovalFlowRepo) DeleteStepsByFlowID(ctx context.Context, flowID uuid.UUID) error {
	delete(m.steps, flowID)
	return nil
}

func (m *mockApprovalFlowRepo) CreateSteps(ctx context.Context, steps []model.ApprovalStep) error {
	if len(steps) > 0 {
		m.steps[steps[0].FlowID] = steps
	}
	return nil
}

// ===== Extended Setup =====

func setupExtendedTestDeps(t *testing.T) (Deps, *mockOvertimeRequestRepo, *mockLeaveBalanceRepo, *mockAttendanceCorrectionRepo, *mockNotificationRepo, *mockProjectRepo, *mockTimeEntryRepo, *mockHolidayRepo, *mockApprovalFlowRepo) {
	cfg := &config.Config{
		JWTSecretKey:          "test-secret-key",
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
	}
	log, _ := logger.NewLogger("debug", "test")
	otRepo := newMockOvertimeRequestRepo()
	lbRepo := newMockLeaveBalanceRepo()
	acRepo := newMockAttendanceCorrectionRepo()
	nRepo := newMockNotificationRepo()
	pRepo := newMockProjectRepo()
	teRepo := newMockTimeEntryRepo()
	hRepo := newMockHolidayRepo()
	afRepo := newMockApprovalFlowRepo()

	deps := Deps{
		Config: cfg,
		Logger: log,
		Repos: &repository.Repositories{
			OvertimeRequest:      otRepo,
			LeaveBalance:         lbRepo,
			AttendanceCorrection: acRepo,
			Notification:         nRepo,
			Project:              pRepo,
			TimeEntry:            teRepo,
			Holiday:              hRepo,
			ApprovalFlow:         afRepo,
		},
	}
	return deps, otRepo, lbRepo, acRepo, nRepo, pRepo, teRepo, hRepo, afRepo
}

// ===== NotificationService Tests =====

func TestNotificationService_Send_Success(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewNotificationService(deps)
	ctx := context.Background()
	userID := uuid.New()

	err := svc.Send(ctx, userID, model.NotificationTypeLeaveApproved, "Test", "Message")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if len(nRepo.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(nRepo.notifications))
	}
}

func TestNotificationService_Send_Error(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	nRepo.createErr = errors.New("db error")
	svc := NewNotificationService(deps)

	err := svc.Send(context.Background(), uuid.New(), model.NotificationTypeLeaveApproved, "Test", "Msg")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestNotificationService_GetByUser(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewNotificationService(deps)
	ctx := context.Background()
	userID := uuid.New()

	nRepo.notifications[uuid.New()] = &model.Notification{
		BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID, Title: "Test", IsRead: false,
	}

	notifs, total, err := svc.GetByUser(ctx, userID, nil, 1, 20)
	if err != nil {
		t.Fatalf("GetByUser failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1, got %d", total)
	}
	if len(notifs) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(notifs))
	}
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewNotificationService(deps)
	ctx := context.Background()
	nID := uuid.New()
	nRepo.notifications[nID] = &model.Notification{BaseModel: model.BaseModel{ID: nID}, IsRead: false}

	err := svc.MarkAsRead(ctx, nID)
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}
	if !nRepo.notifications[nID].IsRead {
		t.Error("Expected notification to be read")
	}
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewNotificationService(deps)
	ctx := context.Background()
	userID := uuid.New()
	n1 := uuid.New()
	n2 := uuid.New()
	nRepo.notifications[n1] = &model.Notification{BaseModel: model.BaseModel{ID: n1}, UserID: userID, IsRead: false}
	nRepo.notifications[n2] = &model.Notification{BaseModel: model.BaseModel{ID: n2}, UserID: userID, IsRead: false}

	err := svc.MarkAllAsRead(ctx, userID)
	if err != nil {
		t.Fatalf("MarkAllAsRead failed: %v", err)
	}
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewNotificationService(deps)
	ctx := context.Background()
	userID := uuid.New()
	nRepo.notifications[uuid.New()] = &model.Notification{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID, IsRead: false}
	nRepo.notifications[uuid.New()] = &model.Notification{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID, IsRead: true}

	count, err := svc.GetUnreadCount(ctx, userID)
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1, got %d", count)
	}
}

func TestNotificationService_Delete(t *testing.T) {
	deps, _, _, _, nRepo, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewNotificationService(deps)
	nID := uuid.New()
	nRepo.notifications[nID] = &model.Notification{BaseModel: model.BaseModel{ID: nID}}

	err := svc.Delete(context.Background(), nID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(nRepo.notifications) != 0 {
		t.Error("Expected notification to be deleted")
	}
}

// ===== ProjectService Tests =====

func TestProjectService_Create_Success(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	ctx := context.Background()

	p, err := svc.Create(ctx, &model.ProjectCreateRequest{
		Name: "Test Project", Code: "TP001", Description: "A project",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if p.Name != "Test Project" {
		t.Errorf("Expected 'Test Project', got '%s'", p.Name)
	}
	if len(pRepo.projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(pRepo.projects))
	}
}

func TestProjectService_Create_Error(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	pRepo.createErr = errors.New("db error")
	svc := NewProjectService(deps)

	_, err := svc.Create(context.Background(), &model.ProjectCreateRequest{Name: "Test"})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestProjectService_GetByID_Success(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	pID := uuid.New()
	pRepo.projects[pID] = &model.Project{BaseModel: model.BaseModel{ID: pID}, Name: "Test"}

	p, err := svc.GetByID(context.Background(), pID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if p.Name != "Test" {
		t.Errorf("Expected 'Test', got '%s'", p.Name)
	}
}

func TestProjectService_GetByID_NotFound(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Error("Expected error")
	}
}

func TestProjectService_GetAll(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	pRepo.projects[uuid.New()] = &model.Project{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "P1"}
	pRepo.projects[uuid.New()] = &model.Project{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "P2"}

	projects, total, err := svc.GetAll(context.Background(), nil, 1, 20)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if total != 2 {
		t.Errorf("Expected 2, got %d", total)
	}
	if len(projects) != 2 {
		t.Errorf("Expected 2, got %d", len(projects))
	}
}

func TestProjectService_GetAll_WithStatusFilter(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	pRepo.projects[uuid.New()] = &model.Project{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "P1", Status: model.ProjectStatusActive}
	pRepo.projects[uuid.New()] = &model.Project{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "P2", Status: model.ProjectStatusArchived}

	status := model.ProjectStatusActive
	projects, total, err := svc.GetAll(context.Background(), &status, 1, 20)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected 1, got %d", total)
	}
	if len(projects) != 1 {
		t.Errorf("Expected 1, got %d", len(projects))
	}
}

func TestProjectService_Update_Success(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	pID := uuid.New()
	pRepo.projects[pID] = &model.Project{BaseModel: model.BaseModel{ID: pID}, Name: "Old"}

	newName := "Updated"
	p, err := svc.Update(context.Background(), pID, &model.ProjectUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if p.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", p.Name)
	}
}

func TestProjectService_Update_NotFound(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)

	newName := "Updated"
	_, err := svc.Update(context.Background(), uuid.New(), &model.ProjectUpdateRequest{Name: &newName})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestProjectService_Delete(t *testing.T) {
	deps, _, _, _, _, pRepo, _, _, _ := setupExtendedTestDeps(t)
	svc := NewProjectService(deps)
	pID := uuid.New()
	pRepo.projects[pID] = &model.Project{BaseModel: model.BaseModel{ID: pID}}

	err := svc.Delete(context.Background(), pID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// ===== TimeEntryService Tests =====

func TestTimeEntryService_Create_Success(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	userID := uuid.New()
	projectID := uuid.New()

	e, err := svc.Create(context.Background(), userID, &model.TimeEntryCreate{
		ProjectID: projectID, Date: "2024-01-15", Minutes: 120, Description: "Coding",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if e.Minutes != 120 {
		t.Errorf("Expected 120, got %d", e.Minutes)
	}
	if len(teRepo.entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(teRepo.entries))
	}
}

func TestTimeEntryService_Create_InvalidDate(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)

	_, err := svc.Create(context.Background(), uuid.New(), &model.TimeEntryCreate{
		Date: "invalid-date", Minutes: 120,
	})
	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

func TestTimeEntryService_GetByUserAndDateRange(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	userID := uuid.New()
	teRepo.entries[uuid.New()] = &model.TimeEntry{
		BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID,
		Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	entries, err := svc.GetByUserAndDateRange(context.Background(), userID, start, end)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1, got %d", len(entries))
	}
}

func TestTimeEntryService_Update_Success(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	eID := uuid.New()
	teRepo.entries[eID] = &model.TimeEntry{BaseModel: model.BaseModel{ID: eID}, Minutes: 60}

	newMin := 120
	e, err := svc.Update(context.Background(), eID, &model.TimeEntryUpdate{Minutes: &newMin})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if e.Minutes != 120 {
		t.Errorf("Expected 120, got %d", e.Minutes)
	}
}

func TestTimeEntryService_Update_NotFound(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)

	newMin := 120
	_, err := svc.Update(context.Background(), uuid.New(), &model.TimeEntryUpdate{Minutes: &newMin})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestTimeEntryService_Delete(t *testing.T) {
	deps, _, _, _, _, _, teRepo, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)
	eID := uuid.New()
	teRepo.entries[eID] = &model.TimeEntry{BaseModel: model.BaseModel{ID: eID}}

	err := svc.Delete(context.Background(), eID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTimeEntryService_GetProjectSummary(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewTimeEntryService(deps)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	summaries, err := svc.GetProjectSummary(context.Background(), start, end)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if summaries == nil {
		t.Error("Expected non-nil summaries")
	}
}

// ===== HolidayService Tests =====

func TestHolidayService_Create_Success(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)

	h, err := svc.Create(context.Background(), &model.HolidayCreateRequest{
		Date: "2024-01-01", Name: "元旦", HolidayType: model.HolidayTypeNational,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if h.Name != "元旦" {
		t.Errorf("Expected '元旦', got '%s'", h.Name)
	}
	if len(hRepo.holidays) != 1 {
		t.Errorf("Expected 1 holiday, got %d", len(hRepo.holidays))
	}
}

func TestHolidayService_Create_InvalidDate(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)

	_, err := svc.Create(context.Background(), &model.HolidayCreateRequest{Date: "invalid"})
	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

func TestHolidayService_GetByYear(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hRepo.holidays[uuid.New()] = &model.Holiday{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		Name: "元旦",
	}

	holidays, err := svc.GetByYear(context.Background(), 2024)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(holidays) != 1 {
		t.Errorf("Expected 1, got %d", len(holidays))
	}
}

func TestHolidayService_Update_Success(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hID := uuid.New()
	hRepo.holidays[hID] = &model.Holiday{BaseModel: model.BaseModel{ID: hID}, Name: "Old"}

	newName := "Updated"
	h, err := svc.Update(context.Background(), hID, &model.HolidayUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if h.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", h.Name)
	}
}

func TestHolidayService_Update_NotFound(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)

	newName := "Updated"
	_, err := svc.Update(context.Background(), uuid.New(), &model.HolidayUpdateRequest{Name: &newName})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestHolidayService_Delete(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hID := uuid.New()
	hRepo.holidays[hID] = &model.Holiday{BaseModel: model.BaseModel{ID: hID}}

	err := svc.Delete(context.Background(), hID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestHolidayService_GetCalendar(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hRepo.holidays[uuid.New()] = &model.Holiday{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		Name: "元旦", HolidayType: model.HolidayTypeNational,
	}

	days, err := svc.GetCalendar(context.Background(), 2024, 1)
	if err != nil {
		t.Fatalf("GetCalendar failed: %v", err)
	}
	if len(days) != 31 {
		t.Errorf("Expected 31 days for January, got %d", len(days))
	}
	// Check that Jan 1 is marked as holiday
	found := false
	for _, d := range days {
		if d.Date == "2024-01-01" && d.IsHoliday {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Jan 1 to be marked as holiday")
	}
}

func TestHolidayService_GetWorkingDays(t *testing.T) {
	deps, _, _, _, _, _, _, hRepo, _ := setupExtendedTestDeps(t)
	svc := NewHolidayService(deps)
	hRepo.holidays[uuid.New()] = &model.Holiday{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		Name: "元旦",
	}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.Local)
	summary, err := svc.GetWorkingDays(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetWorkingDays failed: %v", err)
	}
	if summary.TotalDays != 7 {
		t.Errorf("Expected 7 total days, got %d", summary.TotalDays)
	}
	if summary.Holidays < 1 {
		t.Errorf("Expected at least 1 holiday, got %d", summary.Holidays)
	}
}

// ===== ApprovalFlowService Tests =====

func TestApprovalFlowService_Create_Success(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)

	f, err := svc.Create(context.Background(), &model.ApprovalFlowCreateRequest{
		Name: "Test Flow", FlowType: model.ApprovalFlowLeave,
		Steps: []model.ApprovalStepRequest{
			{StepOrder: 1, StepType: model.ApprovalStepRole, ApproverRole: rolePtr(model.RoleManager)},
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if f.Name != "Test Flow" {
		t.Errorf("Expected 'Test Flow', got '%s'", f.Name)
	}
	if len(afRepo.flows) != 1 {
		t.Errorf("Expected 1, got %d", len(afRepo.flows))
	}
}

func TestApprovalFlowService_GetAll(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	afRepo.flows[uuid.New()] = &model.ApprovalFlow{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "F1"}

	flows, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(flows) != 1 {
		t.Errorf("Expected 1, got %d", len(flows))
	}
}

func TestApprovalFlowService_GetByID(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{BaseModel: model.BaseModel{ID: fID}, Name: "Test"}

	f, err := svc.GetByID(context.Background(), fID)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if f.Name != "Test" {
		t.Errorf("Expected 'Test', got '%s'", f.Name)
	}
}

func TestApprovalFlowService_GetByType(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	afRepo.flows[uuid.New()] = &model.ApprovalFlow{
		BaseModel: model.BaseModel{ID: uuid.New()},
		FlowType: model.ApprovalFlowLeave, IsActive: true,
	}

	flows, err := svc.GetByType(context.Background(), model.ApprovalFlowLeave)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(flows) != 1 {
		t.Errorf("Expected 1, got %d", len(flows))
	}
}

func TestApprovalFlowService_Update_Success(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{BaseModel: model.BaseModel{ID: fID}, Name: "Old"}

	newName := "Updated"
	f, err := svc.Update(context.Background(), fID, &model.ApprovalFlowUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if f.Name != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", f.Name)
	}
}

func TestApprovalFlowService_Update_NotFound(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)

	newName := "Updated"
	_, err := svc.Update(context.Background(), uuid.New(), &model.ApprovalFlowUpdateRequest{Name: &newName})
	if err == nil {
		t.Error("Expected error")
	}
}

func TestApprovalFlowService_Delete(t *testing.T) {
	deps, _, _, _, _, _, _, _, afRepo := setupExtendedTestDeps(t)
	svc := NewApprovalFlowService(deps)
	fID := uuid.New()
	afRepo.flows[fID] = &model.ApprovalFlow{BaseModel: model.BaseModel{ID: fID}}

	err := svc.Delete(context.Background(), fID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

// ===== LeaveBalanceService Tests =====

func TestLeaveBalanceService_GetByUser_Success(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()
	lbRepo.balances["test-key"] = &model.LeaveBalance{
		UserID: userID, FiscalYear: 2024, LeaveType: model.LeaveTypePaid,
		TotalDays: 10, UsedDays: 3, CarriedOver: 2,
	}

	balances, err := svc.GetByUser(context.Background(), userID, 2024)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if len(balances) != 1 {
		t.Errorf("Expected 1, got %d", len(balances))
	}
	if balances[0].RemainingDays != 9 { // 10 + 2 - 3
		t.Errorf("Expected 9 remaining, got %.1f", balances[0].RemainingDays)
	}
}

func TestLeaveBalanceService_GetByUser_Error(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	lbRepo.findErr = errors.New("db error")
	svc := NewLeaveBalanceService(deps)

	_, err := svc.GetByUser(context.Background(), uuid.New(), 2024)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestLeaveBalanceService_SetBalance(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()
	totalDays := 15.0

	err := svc.SetBalance(context.Background(), userID, 2024, model.LeaveTypePaid, &model.LeaveBalanceUpdate{
		TotalDays: &totalDays,
	})
	if err != nil {
		t.Fatalf("SetBalance failed: %v", err)
	}
}

func TestLeaveBalanceService_DeductBalance_Success(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()
	k := lbRepo.key(userID, time.Now().Year(), model.LeaveTypePaid)
	lbRepo.balances[k] = &model.LeaveBalance{
		UserID: userID, FiscalYear: time.Now().Year(), LeaveType: model.LeaveTypePaid,
		TotalDays: 10, UsedDays: 0, CarriedOver: 0,
	}

	err := svc.DeductBalance(context.Background(), userID, model.LeaveTypePaid, 3)
	if err != nil {
		t.Fatalf("DeductBalance failed: %v", err)
	}
	if lbRepo.balances[k].UsedDays != 3 {
		t.Errorf("Expected 3 used days, got %.1f", lbRepo.balances[k].UsedDays)
	}
}

func TestLeaveBalanceService_DeductBalance_Insufficient(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()
	k := lbRepo.key(userID, time.Now().Year(), model.LeaveTypePaid)
	lbRepo.balances[k] = &model.LeaveBalance{
		UserID: userID, FiscalYear: time.Now().Year(), LeaveType: model.LeaveTypePaid,
		TotalDays: 5, UsedDays: 4, CarriedOver: 0,
	}

	err := svc.DeductBalance(context.Background(), userID, model.LeaveTypePaid, 3)
	if err == nil {
		t.Error("Expected error for insufficient balance")
	}
}

func TestLeaveBalanceService_DeductBalance_NotFound(t *testing.T) {
	deps, _, _, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)

	err := svc.DeductBalance(context.Background(), uuid.New(), model.LeaveTypePaid, 1)
	if err == nil {
		t.Error("Expected error for not found balance")
	}
}

func TestLeaveBalanceService_InitializeForUser(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	svc := NewLeaveBalanceService(deps)
	userID := uuid.New()

	err := svc.InitializeForUser(context.Background(), userID, 2024)
	if err != nil {
		t.Fatalf("InitializeForUser failed: %v", err)
	}
	if len(lbRepo.balances) != 3 { // paid, sick, special
		t.Errorf("Expected 3 balance entries, got %d", len(lbRepo.balances))
	}
}

func TestLeaveBalanceService_InitializeForUser_UpsertError(t *testing.T) {
	deps, _, lbRepo, _, _, _, _, _, _ := setupExtendedTestDeps(t)
	lbRepo.upsertErr = errors.New("db error")
	svc := NewLeaveBalanceService(deps)

	err := svc.InitializeForUser(context.Background(), uuid.New(), 2024)
	if err == nil {
		t.Error("Expected error")
	}
}

// Helpers
func stringPtr(s string) *string { return &s }
func rolePtr(r model.Role) *model.Role { return &r }
