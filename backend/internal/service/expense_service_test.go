package service

import (
	"context"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
)

type expTestUserRepo struct {
	users        map[uuid.UUID]*model.User
	usersByEmail map[string]*model.User

	createErr      error
	findByIDErr    error
	findByEmailErr error
	findAllErr     error
	updateErr      error
	deleteErr      error
}

func newExpTestUserRepo() *expTestUserRepo {
	return &expTestUserRepo{
		users:        map[uuid.UUID]*model.User{},
		usersByEmail: map[string]*model.User{},
	}
}

func (m *expTestUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *expTestUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *expTestUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *expTestUserRepo) FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	out := make([]model.User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, int64(len(out)), nil
}

func (m *expTestUserRepo) Update(ctx context.Context, user *model.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *expTestUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.users, id)
	return nil
}

func (m *expTestUserRepo) FindByDepartmentID(ctx context.Context, departmentID uuid.UUID) ([]model.User, error) {
	var out []model.User
	for _, u := range m.users {
		if u.DepartmentID != nil && *u.DepartmentID == departmentID {
			out = append(out, *u)
		}
	}
	return out, nil
}

type expTestExpenseRepo struct {
	expenses map[uuid.UUID]*model.Expense

	findAllList []model.Expense
	statsResp   *model.ExpenseStatsResponse
	reportResp  *model.ExpenseReportResponse
	trendResp   []model.MonthlyTrendItem

	lastReportStart time.Time
	lastReportEnd   time.Time
	lastTrendYear   int

	createErr      error
	findByIDErr    error
	findByUserErr  error
	findPendingErr error
	findAllErr     error
	updateErr      error
	deleteErr      error
	statsErr       error
	reportErr      error
	trendErr       error
}

func newExpTestExpenseRepo() *expTestExpenseRepo {
	return &expTestExpenseRepo{
		expenses: map[uuid.UUID]*model.Expense{},
	}
}

func (m *expTestExpenseRepo) Create(ctx context.Context, expense *model.Expense) error {
	if m.createErr != nil {
		return m.createErr
	}
	if expense.ID == uuid.Nil {
		expense.ID = uuid.New()
	}
	if expense.CreatedAt.IsZero() {
		expense.CreatedAt = time.Now()
	}
	m.expenses[expense.ID] = expense
	return nil
}

func (m *expTestExpenseRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Expense, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	e, ok := m.expenses[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return e, nil
}

func (m *expTestExpenseRepo) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
	if m.findByUserErr != nil {
		return nil, 0, m.findByUserErr
	}
	out := make([]model.Expense, 0)
	for _, e := range m.expenses {
		if e.UserID != userID {
			continue
		}
		if status != "" && string(e.Status) != status {
			continue
		}
		if category != "" {
			matched := false
			for _, it := range e.Items {
				if string(it.Category) == category {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		out = append(out, *e)
	}
	return out, int64(len(out)), nil
}

func (m *expTestExpenseRepo) FindPending(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error) {
	if m.findPendingErr != nil {
		return nil, 0, m.findPendingErr
	}
	out := make([]model.Expense, 0)
	for _, e := range m.expenses {
		if e.Status == model.ExpenseStatusPending {
			out = append(out, *e)
		}
	}
	return out, int64(len(out)), nil
}

func (m *expTestExpenseRepo) FindAll(ctx context.Context, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	if m.findAllList != nil {
		return m.findAllList, int64(len(m.findAllList)), nil
	}
	out := make([]model.Expense, 0, len(m.expenses))
	for _, e := range m.expenses {
		out = append(out, *e)
	}
	return out, int64(len(out)), nil
}

func (m *expTestExpenseRepo) Update(ctx context.Context, expense *model.Expense) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.expenses[expense.ID] = expense
	return nil
}

func (m *expTestExpenseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.expenses, id)
	return nil
}

func (m *expTestExpenseRepo) GetStats(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error) {
	if m.statsErr != nil {
		return nil, m.statsErr
	}
	if m.statsResp != nil {
		return m.statsResp, nil
	}
	return &model.ExpenseStatsResponse{}, nil
}

func (m *expTestExpenseRepo) GetReport(ctx context.Context, startDate, endDate time.Time) (*model.ExpenseReportResponse, error) {
	m.lastReportStart = startDate
	m.lastReportEnd = endDate
	if m.reportErr != nil {
		return nil, m.reportErr
	}
	if m.reportResp != nil {
		return m.reportResp, nil
	}
	return &model.ExpenseReportResponse{}, nil
}

func (m *expTestExpenseRepo) GetMonthlyTrend(ctx context.Context, year int) ([]model.MonthlyTrendItem, error) {
	m.lastTrendYear = year
	if m.trendErr != nil {
		return nil, m.trendErr
	}
	if m.trendResp != nil {
		return m.trendResp, nil
	}
	return []model.MonthlyTrendItem{}, nil
}

type expTestExpenseItemRepo struct {
	deletedExpenseIDs []uuid.UUID
	createdBatches    [][]model.ExpenseItem
	deleteErr         error
	createErr         error
}

func (m *expTestExpenseItemRepo) DeleteByExpenseID(ctx context.Context, expenseID uuid.UUID) error {
	m.deletedExpenseIDs = append(m.deletedExpenseIDs, expenseID)
	return m.deleteErr
}

func (m *expTestExpenseItemRepo) CreateBatch(ctx context.Context, items []model.ExpenseItem) error {
	m.createdBatches = append(m.createdBatches, items)
	return m.createErr
}

type expTestExpenseCommentRepo struct {
	commentsByExpense map[uuid.UUID][]model.ExpenseComment
	createErr         error
	findErr           error
}

func newExpTestExpenseCommentRepo() *expTestExpenseCommentRepo {
	return &expTestExpenseCommentRepo{
		commentsByExpense: map[uuid.UUID][]model.ExpenseComment{},
	}
}

func (m *expTestExpenseCommentRepo) Create(ctx context.Context, comment *model.ExpenseComment) error {
	if m.createErr != nil {
		return m.createErr
	}
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now()
	}
	m.commentsByExpense[comment.ExpenseID] = append(m.commentsByExpense[comment.ExpenseID], *comment)
	return nil
}

func (m *expTestExpenseCommentRepo) FindByExpenseID(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseComment, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.commentsByExpense[expenseID], nil
}

type expTestExpenseHistoryRepo struct {
	historiesByExpense map[uuid.UUID][]model.ExpenseHistory
	createErr          error
	findErr            error
}

func newExpTestExpenseHistoryRepo() *expTestExpenseHistoryRepo {
	return &expTestExpenseHistoryRepo{
		historiesByExpense: map[uuid.UUID][]model.ExpenseHistory{},
	}
}

func (m *expTestExpenseHistoryRepo) Create(ctx context.Context, history *model.ExpenseHistory) error {
	if m.createErr != nil {
		return m.createErr
	}
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}
	m.historiesByExpense[history.ExpenseID] = append(m.historiesByExpense[history.ExpenseID], *history)
	return nil
}

func (m *expTestExpenseHistoryRepo) FindByExpenseID(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistory, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.historiesByExpense[expenseID], nil
}

type expTestExpenseTemplateRepo struct {
	templates   map[uuid.UUID]*model.ExpenseTemplate
	createErr   error
	findAllErr  error
	findByIDErr error
	updateErr   error
	deleteErr   error
}

func newExpTestExpenseTemplateRepo() *expTestExpenseTemplateRepo {
	return &expTestExpenseTemplateRepo{templates: map[uuid.UUID]*model.ExpenseTemplate{}}
}

func (m *expTestExpenseTemplateRepo) Create(ctx context.Context, template *model.ExpenseTemplate) error {
	if m.createErr != nil {
		return m.createErr
	}
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	m.templates[template.ID] = template
	return nil
}

func (m *expTestExpenseTemplateRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	out := make([]model.ExpenseTemplate, 0)
	for _, t := range m.templates {
		if t.UserID == userID {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (m *expTestExpenseTemplateRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ExpenseTemplate, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	t, ok := m.templates[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (m *expTestExpenseTemplateRepo) Update(ctx context.Context, template *model.ExpenseTemplate) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.templates[template.ID] = template
	return nil
}

func (m *expTestExpenseTemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.templates, id)
	return nil
}

type expTestExpensePolicyRepo struct {
	policies          map[uuid.UUID]*model.ExpensePolicy
	createErr         error
	findAllErr        error
	findByIDErr       error
	findByCategoryErr error
	updateErr         error
	deleteErr         error
}

func newExpTestExpensePolicyRepo() *expTestExpensePolicyRepo {
	return &expTestExpensePolicyRepo{policies: map[uuid.UUID]*model.ExpensePolicy{}}
}

func (m *expTestExpensePolicyRepo) Create(ctx context.Context, policy *model.ExpensePolicy) error {
	if m.createErr != nil {
		return m.createErr
	}
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	m.policies[policy.ID] = policy
	return nil
}

func (m *expTestExpensePolicyRepo) FindAll(ctx context.Context) ([]model.ExpensePolicy, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	out := make([]model.ExpensePolicy, 0, len(m.policies))
	for _, p := range m.policies {
		out = append(out, *p)
	}
	return out, nil
}

func (m *expTestExpensePolicyRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ExpensePolicy, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	p, ok := m.policies[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *expTestExpensePolicyRepo) FindByCategory(ctx context.Context, category string) (*model.ExpensePolicy, error) {
	if m.findByCategoryErr != nil {
		return nil, m.findByCategoryErr
	}
	for _, p := range m.policies {
		if string(p.Category) == category {
			return p, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *expTestExpensePolicyRepo) Update(ctx context.Context, policy *model.ExpensePolicy) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.policies[policy.ID] = policy
	return nil
}

func (m *expTestExpensePolicyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.policies, id)
	return nil
}

type expTestExpenseBudgetRepo struct {
	budgets []model.ExpenseBudget
	findErr error
}

func (m *expTestExpenseBudgetRepo) FindAll(ctx context.Context) ([]model.ExpenseBudget, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.budgets, nil
}

type expTestExpenseNotificationRepo struct {
	notifications map[uuid.UUID]*model.ExpenseNotification
	createErr     error
	findErr       error
	markErr       error
	markAllErr    error
}

func newExpTestExpenseNotificationRepo() *expTestExpenseNotificationRepo {
	return &expTestExpenseNotificationRepo{notifications: map[uuid.UUID]*model.ExpenseNotification{}}
}

func (m *expTestExpenseNotificationRepo) Create(ctx context.Context, notification *model.ExpenseNotification) error {
	if m.createErr != nil {
		return m.createErr
	}
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	m.notifications[notification.ID] = notification
	return nil
}

func (m *expTestExpenseNotificationRepo) FindByUserID(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var out []model.ExpenseNotification
	for _, n := range m.notifications {
		if n.UserID != userID {
			continue
		}
		if filter == "unread" && n.IsRead {
			continue
		}
		out = append(out, *n)
	}
	return out, nil
}

func (m *expTestExpenseNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.markErr != nil {
		return m.markErr
	}
	if n, ok := m.notifications[id]; ok {
		n.IsRead = true
		now := time.Now()
		n.ReadAt = &now
	}
	return nil
}

func (m *expTestExpenseNotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.markAllErr != nil {
		return m.markAllErr
	}
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			n.IsRead = true
			now := time.Now()
			n.ReadAt = &now
		}
	}
	return nil
}

type expTestExpenseReminderRepo struct {
	remindersByUser map[uuid.UUID][]model.ExpenseReminder
	findErr         error
	dismissErr      error
	dismissedIDs    []uuid.UUID
}

func newExpTestExpenseReminderRepo() *expTestExpenseReminderRepo {
	return &expTestExpenseReminderRepo{remindersByUser: map[uuid.UUID][]model.ExpenseReminder{}}
}

func (m *expTestExpenseReminderRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.remindersByUser[userID], nil
}

func (m *expTestExpenseReminderRepo) Dismiss(ctx context.Context, id uuid.UUID) error {
	if m.dismissErr != nil {
		return m.dismissErr
	}
	m.dismissedIDs = append(m.dismissedIDs, id)
	return nil
}

type expTestExpenseNotificationSettingRepo struct {
	settingsByUser map[uuid.UUID]*model.ExpenseNotificationSetting
	findErr        error
	upsertErr      error
}

func newExpTestExpenseNotificationSettingRepo() *expTestExpenseNotificationSettingRepo {
	return &expTestExpenseNotificationSettingRepo{settingsByUser: map[uuid.UUID]*model.ExpenseNotificationSetting{}}
}

func (m *expTestExpenseNotificationSettingRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if s, ok := m.settingsByUser[userID]; ok {
		return s, nil
	}
	s := &model.ExpenseNotificationSetting{UserID: userID}
	m.settingsByUser[userID] = s
	return s, nil
}

func (m *expTestExpenseNotificationSettingRepo) Upsert(ctx context.Context, setting *model.ExpenseNotificationSetting) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.settingsByUser[setting.UserID] = setting
	return nil
}

type expTestExpenseApprovalFlowRepo struct {
	active  *model.ExpenseApprovalFlow
	findErr error
}

func (m *expTestExpenseApprovalFlowRepo) FindActive(ctx context.Context) (*model.ExpenseApprovalFlow, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.active != nil {
		return m.active, nil
	}
	return &model.ExpenseApprovalFlow{}, nil
}

type expTestExpenseDelegateRepo struct {
	delegates map[uuid.UUID]*model.ExpenseDelegate
	createErr error
	findErr   error
	deleteErr error
}

func newExpTestExpenseDelegateRepo() *expTestExpenseDelegateRepo {
	return &expTestExpenseDelegateRepo{delegates: map[uuid.UUID]*model.ExpenseDelegate{}}
}

func (m *expTestExpenseDelegateRepo) Create(ctx context.Context, delegate *model.ExpenseDelegate) error {
	if m.createErr != nil {
		return m.createErr
	}
	if delegate.ID == uuid.Nil {
		delegate.ID = uuid.New()
	}
	m.delegates[delegate.ID] = delegate
	return nil
}

func (m *expTestExpenseDelegateRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var out []model.ExpenseDelegate
	for _, d := range m.delegates {
		if d.UserID == userID {
			out = append(out, *d)
		}
	}
	return out, nil
}

func (m *expTestExpenseDelegateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.delegates, id)
	return nil
}

type expTestExpensePolicyViolationRepo struct {
	violationsByUser map[uuid.UUID][]model.ExpensePolicyViolation
	createErr        error
	findErr          error
}

func newExpTestExpensePolicyViolationRepo() *expTestExpensePolicyViolationRepo {
	return &expTestExpensePolicyViolationRepo{
		violationsByUser: map[uuid.UUID][]model.ExpensePolicyViolation{},
	}
}

func (m *expTestExpensePolicyViolationRepo) Create(ctx context.Context, violation *model.ExpensePolicyViolation) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.violationsByUser[violation.UserID] = append(m.violationsByUser[violation.UserID], *violation)
	return nil
}

func (m *expTestExpensePolicyViolationRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.violationsByUser[userID], nil
}

type expTestEnv struct {
	deps         Deps
	userRepo     *expTestUserRepo
	expenseRepo  *expTestExpenseRepo
	itemRepo     *expTestExpenseItemRepo
	commentRepo  *expTestExpenseCommentRepo
	historyRepo  *expTestExpenseHistoryRepo
	templateRepo *expTestExpenseTemplateRepo
	policyRepo   *expTestExpensePolicyRepo
	budgetRepo   *expTestExpenseBudgetRepo
	notifRepo    *expTestExpenseNotificationRepo
	reminderRepo *expTestExpenseReminderRepo
	settingRepo  *expTestExpenseNotificationSettingRepo
	flowRepo     *expTestExpenseApprovalFlowRepo
	delegateRepo *expTestExpenseDelegateRepo
	violateRepo  *expTestExpensePolicyViolationRepo
}

func newExpTestEnv() *expTestEnv {
	userRepo := newExpTestUserRepo()
	expenseRepo := newExpTestExpenseRepo()
	itemRepo := &expTestExpenseItemRepo{}
	commentRepo := newExpTestExpenseCommentRepo()
	historyRepo := newExpTestExpenseHistoryRepo()
	templateRepo := newExpTestExpenseTemplateRepo()
	policyRepo := newExpTestExpensePolicyRepo()
	budgetRepo := &expTestExpenseBudgetRepo{}
	notifRepo := newExpTestExpenseNotificationRepo()
	reminderRepo := newExpTestExpenseReminderRepo()
	settingRepo := newExpTestExpenseNotificationSettingRepo()
	flowRepo := &expTestExpenseApprovalFlowRepo{}
	delegateRepo := newExpTestExpenseDelegateRepo()
	violateRepo := newExpTestExpensePolicyViolationRepo()

	deps := Deps{
		Repos: &repository.Repositories{
			User:                       userRepo,
			Expense:                    expenseRepo,
			ExpenseItem:                itemRepo,
			ExpenseComment:             commentRepo,
			ExpenseHistory:             historyRepo,
			ExpenseTemplate:            templateRepo,
			ExpensePolicy:              policyRepo,
			ExpenseBudget:              budgetRepo,
			ExpenseNotification:        notifRepo,
			ExpenseReminder:            reminderRepo,
			ExpenseNotificationSetting: settingRepo,
			ExpenseApprovalFlow:        flowRepo,
			ExpenseDelegate:            delegateRepo,
			ExpensePolicyViolation:     violateRepo,
		},
	}

	return &expTestEnv{
		deps:         deps,
		userRepo:     userRepo,
		expenseRepo:  expenseRepo,
		itemRepo:     itemRepo,
		commentRepo:  commentRepo,
		historyRepo:  historyRepo,
		templateRepo: templateRepo,
		policyRepo:   policyRepo,
		budgetRepo:   budgetRepo,
		notifRepo:    notifRepo,
		reminderRepo: reminderRepo,
		settingRepo:  settingRepo,
		flowRepo:     flowRepo,
		delegateRepo: delegateRepo,
		violateRepo:  violateRepo,
	}
}

func expParseCSVRows(t *testing.T, b []byte) [][]string {
	t.Helper()
	s := strings.TrimPrefix(string(b), "\uFEFF")
	rows, err := csv.NewReader(strings.NewReader(s)).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv: %v", err)
	}
	return rows
}

func expPtrString(v string) *string { return &v }
func expPtrBool(v bool) *bool       { return &v }

func TestExpenseService_Create(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("default status and history user", func(t *testing.T) {
		env := newExpTestEnv()
		env.userRepo.users[userID] = &model.User{
			BaseModel: model.BaseModel{ID: userID},
			FirstName: "Taro",
			LastName:  "Test",
		}
		svc := NewExpenseService(env.deps)
		req := &model.ExpenseCreateRequest{
			Title: "Taxi",
			Items: []model.ExpenseItemRequest{
				{ExpenseDate: "2024-01-10", Category: "transportation", Description: "one", Amount: 1200},
				{ExpenseDate: "2024-01-11", Category: "meals", Description: "two", Amount: 800},
			},
		}

		got, err := svc.Create(ctx, userID, req)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if got.Status != model.ExpenseStatusDraft {
			t.Fatalf("expected draft, got %s", got.Status)
		}
		if got.TotalAmount != 2000 {
			t.Fatalf("expected total 2000, got %.0f", got.TotalAmount)
		}
		if len(got.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(got.Items))
		}
		h := env.historyRepo.historiesByExpense[got.ID]
		if len(h) != 1 {
			t.Fatalf("expected 1 history, got %d", len(h))
		}
		if h[0].ChangedBy != "Test Taro" {
			t.Fatalf("unexpected changer: %q", h[0].ChangedBy)
		}
	})

	t.Run("provided status and missing user", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseService(env.deps)
		req := &model.ExpenseCreateRequest{
			Title:  "Hotel",
			Status: "pending",
			Items: []model.ExpenseItemRequest{
				{ExpenseDate: "2024-02-01", Category: "accommodation", Description: "stay", Amount: 5000},
			},
		}

		got, err := svc.Create(ctx, userID, req)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if got.Status != model.ExpenseStatusPending {
			t.Fatalf("expected pending, got %s", got.Status)
		}
		h := env.historyRepo.historiesByExpense[got.ID]
		if len(h) != 1 || h[0].ChangedBy != "" {
			t.Fatalf("expected empty changed by when user missing")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		env := newExpTestEnv()
		env.expenseRepo.createErr = errors.New("create error")
		svc := NewExpenseService(env.deps)
		_, err := svc.Create(ctx, userID, &model.ExpenseCreateRequest{
			Title: "Err",
			Items: []model.ExpenseItemRequest{{ExpenseDate: "2024-01-01", Category: "other", Description: "x", Amount: 1}},
		})
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestExpenseService_BasicWrappers(t *testing.T) {
	ctx := context.Background()
	env := newExpTestEnv()
	svc := NewExpenseService(env.deps)

	userID := uuid.New()
	expenseID := uuid.New()
	env.expenseRepo.expenses[expenseID] = &model.Expense{
		BaseModel: model.BaseModel{ID: expenseID},
		UserID:    userID,
		Status:    model.ExpenseStatusPending,
	}

	if _, err := svc.GetByID(ctx, expenseID); err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	list, _, err := svc.GetList(ctx, userID, 1, 10, "pending", "")
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 list item, got %d", len(list))
	}
	if err := svc.Delete(ctx, expenseID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(env.expenseRepo.expenses) != 0 {
		t.Fatalf("expected empty repo after delete")
	}

	pendingID := uuid.New()
	env.expenseRepo.expenses[pendingID] = &model.Expense{
		BaseModel: model.BaseModel{ID: pendingID},
		UserID:    userID,
		Status:    model.ExpenseStatusPending,
	}
	pending, _, err := svc.GetPending(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}

	env.expenseRepo.statsResp = &model.ExpenseStatsResponse{PendingCount: 3}
	stats, err := svc.GetStats(ctx, userID)
	if err != nil || stats.PendingCount != 3 {
		t.Fatalf("GetStats failed: %v / %v", err, stats)
	}
}

func TestExpenseService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("find by id error", func(t *testing.T) {
		env := newExpTestEnv()
		env.expenseRepo.findByIDErr = errors.New("find error")
		svc := NewExpenseService(env.deps)
		_, err := svc.Update(ctx, uuid.New(), uuid.New(), &model.ExpenseUpdateRequest{})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("update error", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		userID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    userID,
			Title:     "Old",
			Status:    model.ExpenseStatusDraft,
		}
		env.expenseRepo.updateErr = errors.New("update error")
		svc := NewExpenseService(env.deps)
		_, err := svc.Update(ctx, expenseID, userID, &model.ExpenseUpdateRequest{
			Title: expPtrString("New"),
		})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("full update with items and user history name", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		userID := uuid.New()
		env.userRepo.users[userID] = &model.User{
			BaseModel: model.BaseModel{ID: userID},
			FirstName: "Hanako",
			LastName:  "Yamada",
		}
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    userID,
			Title:     "Old",
			Status:    model.ExpenseStatusDraft,
			Notes:     "old notes",
		}
		svc := NewExpenseService(env.deps)

		got, err := svc.Update(ctx, expenseID, userID, &model.ExpenseUpdateRequest{
			Title:  expPtrString("New Title"),
			Status: expPtrString("pending"),
			Notes:  expPtrString("new notes"),
			Items: []model.ExpenseItemRequest{
				{ExpenseDate: "2024-02-01", Category: "meals", Description: "lunch", Amount: 1000},
				{ExpenseDate: "2024-02-02", Category: "transportation", Description: "train", Amount: 700},
			},
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if got.Title != "New Title" || got.Notes != "new notes" || got.Status != model.ExpenseStatusPending {
			t.Fatalf("fields not updated: %+v", got)
		}
		if got.TotalAmount != 1700 {
			t.Fatalf("expected total 1700, got %.0f", got.TotalAmount)
		}
		if len(env.itemRepo.deletedExpenseIDs) != 1 {
			t.Fatalf("expected DeleteByExpenseID call")
		}
		if len(env.itemRepo.createdBatches) != 1 || len(env.itemRepo.createdBatches[0]) != 2 {
			t.Fatalf("expected item batch call")
		}
		h := env.historyRepo.historiesByExpense[expenseID]
		if len(h) != 1 || h[0].ChangedBy != "Yamada Hanako" || h[0].OldValue != "draft" || h[0].NewValue != "pending" {
			t.Fatalf("unexpected history: %+v", h)
		}
	})

	t.Run("no optional fields and no user", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		userID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    userID,
			Title:     "Keep",
			Status:    model.ExpenseStatusDraft,
		}
		svc := NewExpenseService(env.deps)

		got, err := svc.Update(ctx, expenseID, userID, &model.ExpenseUpdateRequest{})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if got.Title != "Keep" || got.Status != model.ExpenseStatusDraft {
			t.Fatalf("should keep values")
		}
		if len(env.itemRepo.deletedExpenseIDs) != 0 || len(env.itemRepo.createdBatches) != 0 {
			t.Fatalf("items should not be touched")
		}
		h := env.historyRepo.historiesByExpense[expenseID]
		if len(h) != 1 || h[0].ChangedBy != "" {
			t.Fatalf("expected empty changed by")
		}
	})
}

func TestExpenseService_Approve(t *testing.T) {
	ctx := context.Background()

	t.Run("find error", func(t *testing.T) {
		env := newExpTestEnv()
		env.expenseRepo.findByIDErr = errors.New("find err")
		svc := NewExpenseService(env.deps)
		err := svc.Approve(ctx, uuid.New(), uuid.New(), &model.ExpenseApproveRequest{Status: "approved"})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("update error", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    uuid.New(),
			Title:     "A",
		}
		env.expenseRepo.updateErr = errors.New("update err")
		svc := NewExpenseService(env.deps)
		err := svc.Approve(ctx, expenseID, uuid.New(), &model.ExpenseApproveRequest{Status: "approved"})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("approved branch with missing approver user", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		ownerID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    ownerID,
			Title:     "Expense A",
		}
		approverID := uuid.New()
		svc := NewExpenseService(env.deps)

		if err := svc.Approve(ctx, expenseID, approverID, &model.ExpenseApproveRequest{Status: "approved"}); err != nil {
			t.Fatalf("Approve failed: %v", err)
		}
		e := env.expenseRepo.expenses[expenseID]
		if e.Status != model.ExpenseStatusApproved || e.ApprovedBy == nil || *e.ApprovedBy != approverID || e.ApprovedAt == nil {
			t.Fatalf("expense not approved: %+v", e)
		}
		if len(env.notifRepo.notifications) != 1 {
			t.Fatalf("expected 1 notification")
		}
		for _, n := range env.notifRepo.notifications {
			if n.Type != "approved" {
				t.Fatalf("unexpected notif type: %s", n.Type)
			}
		}
		h := env.historyRepo.historiesByExpense[expenseID]
		if len(h) != 1 || h[0].ChangedBy != "" {
			t.Fatalf("expected empty approver name in history")
		}
	})

	t.Run("rejected branch with approver name and reason", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		ownerID := uuid.New()
		approverID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    ownerID,
			Title:     "Expense B",
		}
		env.userRepo.users[approverID] = &model.User{
			BaseModel: model.BaseModel{ID: approverID},
			FirstName: "Ken",
			LastName:  "Sato",
		}
		svc := NewExpenseService(env.deps)

		if err := svc.Approve(ctx, expenseID, approverID, &model.ExpenseApproveRequest{
			Status: "rejected", RejectedReason: "policy",
		}); err != nil {
			t.Fatalf("Approve failed: %v", err)
		}

		e := env.expenseRepo.expenses[expenseID]
		if e.Status != model.ExpenseStatusRejected || e.RejectedReason != "policy" {
			t.Fatalf("unexpected reject state: %+v", e)
		}
		for _, n := range env.notifRepo.notifications {
			if n.Type != "rejected" {
				t.Fatalf("expected rejected notification")
			}
		}
		h := env.historyRepo.historiesByExpense[expenseID]
		if len(h) != 1 || h[0].ChangedBy != "Sato Ken" {
			t.Fatalf("unexpected history approver: %+v", h)
		}
	})
}

func TestExpenseService_AdvancedApprove(t *testing.T) {
	ctx := context.Background()

	t.Run("find error", func(t *testing.T) {
		env := newExpTestEnv()
		env.expenseRepo.findByIDErr = errors.New("find err")
		svc := NewExpenseService(env.deps)
		err := svc.AdvancedApprove(ctx, uuid.New(), uuid.New(), &model.ExpenseAdvancedApproveRequest{Action: "approve", Step: 1})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("update error", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{BaseModel: model.BaseModel{ID: expenseID}}
		env.expenseRepo.updateErr = errors.New("update err")
		svc := NewExpenseService(env.deps)
		err := svc.AdvancedApprove(ctx, expenseID, uuid.New(), &model.ExpenseAdvancedApproveRequest{Action: "approve", Step: 1})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("approve step1", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{BaseModel: model.BaseModel{ID: expenseID}}
		svc := NewExpenseService(env.deps)
		if err := svc.AdvancedApprove(ctx, expenseID, uuid.New(), &model.ExpenseAdvancedApproveRequest{Action: "approve", Step: 1}); err != nil {
			t.Fatalf("failed: %v", err)
		}
		if env.expenseRepo.expenses[expenseID].Status != model.ExpenseStatusStep1Approved {
			t.Fatalf("expected step1 status")
		}
	})

	t.Run("approve final step with approver", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		approverID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{BaseModel: model.BaseModel{ID: expenseID}}
		env.userRepo.users[approverID] = &model.User{
			BaseModel: model.BaseModel{ID: approverID},
			FirstName: "Jiro",
			LastName:  "Abe",
		}
		svc := NewExpenseService(env.deps)
		if err := svc.AdvancedApprove(ctx, expenseID, approverID, &model.ExpenseAdvancedApproveRequest{Action: "approve", Step: 2}); err != nil {
			t.Fatalf("failed: %v", err)
		}
		e := env.expenseRepo.expenses[expenseID]
		if e.Status != model.ExpenseStatusApproved || e.ApprovedBy == nil || e.ApprovedAt == nil {
			t.Fatalf("expected final approve state")
		}
		h := env.historyRepo.historiesByExpense[expenseID]
		if len(h) != 1 || h[0].ChangedBy != "Abe Jiro" {
			t.Fatalf("unexpected history user: %+v", h)
		}
	})

	t.Run("reject and return", func(t *testing.T) {
		env := newExpTestEnv()
		expenseRejectID := uuid.New()
		expenseReturnID := uuid.New()
		env.expenseRepo.expenses[expenseRejectID] = &model.Expense{BaseModel: model.BaseModel{ID: expenseRejectID}}
		env.expenseRepo.expenses[expenseReturnID] = &model.Expense{BaseModel: model.BaseModel{ID: expenseReturnID}}
		svc := NewExpenseService(env.deps)

		if err := svc.AdvancedApprove(ctx, expenseRejectID, uuid.New(), &model.ExpenseAdvancedApproveRequest{
			Action: "reject", Reason: "invalid",
		}); err != nil {
			t.Fatalf("failed reject: %v", err)
		}
		if env.expenseRepo.expenses[expenseRejectID].Status != model.ExpenseStatusRejected || env.expenseRepo.expenses[expenseRejectID].RejectedReason != "invalid" {
			t.Fatalf("expected reject state")
		}
		if err := svc.AdvancedApprove(ctx, expenseReturnID, uuid.New(), &model.ExpenseAdvancedApproveRequest{Action: "return"}); err != nil {
			t.Fatalf("failed return: %v", err)
		}
		if env.expenseRepo.expenses[expenseReturnID].Status != model.ExpenseStatusReturned {
			t.Fatalf("expected returned state")
		}
	})
}

func TestExpenseService_GetReportAndTrend(t *testing.T) {
	ctx := context.Background()
	env := newExpTestEnv()
	svc := NewExpenseService(env.deps)

	env.expenseRepo.reportResp = &model.ExpenseReportResponse{TotalAmount: 999}
	report, err := svc.GetReport(ctx, "2024-01-01", "2024-01-31")
	if err != nil || report.TotalAmount != 999 {
		t.Fatalf("GetReport failed: %v", err)
	}
	if env.expenseRepo.lastReportStart.Format("2006-01-02") != "2024-01-01" {
		t.Fatalf("unexpected start: %s", env.expenseRepo.lastReportStart)
	}
	if env.expenseRepo.lastReportEnd.Format("2006-01-02 15:04:05") != "2024-01-31 23:59:59" {
		t.Fatalf("unexpected end: %s", env.expenseRepo.lastReportEnd)
	}

	env.expenseRepo.trendResp = []model.MonthlyTrendItem{{Month: "2024-01", Amount: 100}}
	trend, err := svc.GetMonthlyTrend(ctx, "2024")
	if err != nil || len(trend) != 1 || env.expenseRepo.lastTrendYear != 2024 {
		t.Fatalf("GetMonthlyTrend failed")
	}

	_, _ = svc.GetMonthlyTrend(ctx, "bad-year")
	if env.expenseRepo.lastTrendYear != time.Now().Year() {
		t.Fatalf("expected fallback current year")
	}
}

func TestExpenseService_ExportCSV(t *testing.T) {
	ctx := context.Background()

	t.Run("repo error", func(t *testing.T) {
		env := newExpTestEnv()
		env.expenseRepo.findAllErr = errors.New("find all error")
		svc := NewExpenseService(env.deps)
		_, err := svc.ExportCSV(ctx, "2024-01-01", "2024-01-31")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("success with filter and nil user", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseService(env.deps)
		inRange1 := model.Expense{
			BaseModel:   model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)},
			Title:       "in-1",
			Status:      model.ExpenseStatusApproved,
			TotalAmount: 1000,
			User:        &model.User{FirstName: "A", LastName: "B"},
		}
		inRange2 := model.Expense{
			BaseModel:   model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)},
			Title:       "in-2",
			Status:      model.ExpenseStatusPending,
			TotalAmount: 500,
			User:        nil,
		}
		outRange := model.Expense{
			BaseModel:   model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 2, 20, 12, 0, 0, 0, time.UTC)},
			Title:       "out",
			Status:      model.ExpenseStatusPending,
			TotalAmount: 700,
		}
		env.expenseRepo.findAllList = []model.Expense{inRange1, inRange2, outRange}

		b, err := svc.ExportCSV(ctx, "2024-01-01", "2024-01-31")
		if err != nil {
			t.Fatalf("ExportCSV failed: %v", err)
		}
		rows := expParseCSVRows(t, b)
		if len(rows) != 3 {
			t.Fatalf("expected 3 rows(header+2), got %d", len(rows))
		}
		if rows[1][1] != "in-1" || rows[2][1] != "in-2" {
			t.Fatalf("unexpected rows: %+v", rows)
		}
		if rows[2][2] != "" {
			t.Fatalf("expected empty user name for nil user")
		}
	})
}

func TestExpenseService_ExportPDF(t *testing.T) {
	ctx := context.Background()

	t.Run("repo error", func(t *testing.T) {
		env := newExpTestEnv()
		env.expenseRepo.findAllErr = errors.New("find all error")
		svc := NewExpenseService(env.deps)
		_, err := svc.ExportPDF(ctx, "2024-01-01", "2024-01-31")
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("success with break and skip", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseService(env.deps)
		var list []model.Expense
		for i := 0; i < 50; i++ {
			list = append(list, model.Expense{
				BaseModel:   model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
				Title:       "title-" + strings.Repeat("x", i%3) + string(rune('A'+(i%26))),
				TotalAmount: float64(i + 1),
				User:        &model.User{FirstName: "F", LastName: "L"},
			})
		}
		list = append(list, model.Expense{
			BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
			Title:     "out-of-range",
		})
		list = append(list, model.Expense{
			BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
			Title:     "no-user",
			User:      nil,
		})
		env.expenseRepo.findAllList = list

		b, err := svc.ExportPDF(ctx, "2024-01-01", "2024-01-31")
		if err != nil {
			t.Fatalf("ExportPDF failed: %v", err)
		}
		s := string(b)
		if !strings.HasPrefix(s, "%PDF-1.4") {
			t.Fatalf("invalid pdf header")
		}
		if !strings.Contains(s, "Total:") {
			t.Fatalf("missing total line")
		}
		if strings.Contains(s, "out-of-range") {
			t.Fatalf("out-of-range entry should be skipped")
		}
	})

	t.Run("success with nil user branch", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseService(env.deps)
		env.expenseRepo.findAllList = []model.Expense{
			{
				BaseModel:   model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)},
				Title:       "skip-me",
				TotalAmount: 999,
				User:        &model.User{FirstName: "X", LastName: "Y"},
			},
			{
				BaseModel:   model.BaseModel{ID: uuid.New(), CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
				Title:       "nil-user",
				TotalAmount: 123,
				User:        nil,
			},
		}

		b, err := svc.ExportPDF(ctx, "2024-01-01", "2024-01-31")
		if err != nil {
			t.Fatalf("ExportPDF failed: %v", err)
		}
		s := string(b)
		if !strings.Contains(s, "nil-user") {
			t.Fatalf("expected pdf content to include title")
		}
		if strings.Contains(s, "skip-me") {
			t.Fatalf("expected out-of-range record to be skipped")
		}
	})
}

func TestExpenseCommentService_GetComments(t *testing.T) {
	ctx := context.Background()

	t.Run("repo error", func(t *testing.T) {
		env := newExpTestEnv()
		env.commentRepo.findErr = errors.New("find error")
		svc := NewExpenseCommentService(env.deps)
		_, err := svc.GetComments(ctx, uuid.New())
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("empty list should return empty slice", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseCommentService(env.deps)
		got, err := svc.GetComments(ctx, uuid.New())
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Fatalf("expected empty non-nil slice")
		}
	})

	t.Run("with user and nil user", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		env.commentRepo.commentsByExpense[expenseID] = []model.ExpenseComment{
			{
				BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Now()},
				ExpenseID: expenseID,
				UserID:    uuid.New(),
				Content:   "a",
				User:      &model.User{FirstName: "A", LastName: "B"},
			},
			{
				BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Now()},
				ExpenseID: expenseID,
				UserID:    uuid.New(),
				Content:   "b",
				User:      nil,
			},
		}
		svc := NewExpenseCommentService(env.deps)
		got, err := svc.GetComments(ctx, expenseID)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if len(got) != 2 || got[0].UserName != "B A" || got[1].UserName != "" {
			t.Fatalf("unexpected response: %+v", got)
		}
	})
}

func TestExpenseCommentService_AddComment(t *testing.T) {
	ctx := context.Background()

	t.Run("create error", func(t *testing.T) {
		env := newExpTestEnv()
		env.commentRepo.createErr = errors.New("create err")
		svc := NewExpenseCommentService(env.deps)
		_, err := svc.AddComment(ctx, uuid.New(), uuid.New(), &model.ExpenseCommentRequest{Content: "x"})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("notification when commenter is different user", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		commentUserID := uuid.New()
		ownerID := uuid.New()
		env.userRepo.users[commentUserID] = &model.User{
			BaseModel: model.BaseModel{ID: commentUserID},
			FirstName: "Hanako",
			LastName:  "Ito",
		}
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    ownerID,
		}
		svc := NewExpenseCommentService(env.deps)

		got, err := svc.AddComment(ctx, expenseID, commentUserID, &model.ExpenseCommentRequest{Content: "hello"})
		if err != nil {
			t.Fatalf("AddComment failed: %v", err)
		}
		if got.UserName != "Ito Hanako" {
			t.Fatalf("unexpected user name: %s", got.UserName)
		}
		if len(env.notifRepo.notifications) != 1 {
			t.Fatalf("expected notification")
		}
	})

	t.Run("no notification when self comment and no user profile", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		userID := uuid.New()
		env.expenseRepo.expenses[expenseID] = &model.Expense{
			BaseModel: model.BaseModel{ID: expenseID},
			UserID:    userID,
		}
		svc := NewExpenseCommentService(env.deps)
		got, err := svc.AddComment(ctx, expenseID, userID, &model.ExpenseCommentRequest{Content: "self"})
		if err != nil {
			t.Fatalf("AddComment failed: %v", err)
		}
		if got.UserName != "" {
			t.Fatalf("expected empty user name")
		}
		if len(env.notifRepo.notifications) != 0 {
			t.Fatalf("should not create notification")
		}
	})
}

func TestExpenseHistoryService_GetHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("repo error", func(t *testing.T) {
		env := newExpTestEnv()
		env.historyRepo.findErr = errors.New("find err")
		svc := NewExpenseHistoryService(env.deps)
		_, err := svc.GetHistory(ctx, uuid.New())
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseHistoryService(env.deps)
		got, err := svc.GetHistory(ctx, uuid.New())
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Fatalf("expected empty non-nil slice")
		}
	})

	t.Run("with history", func(t *testing.T) {
		env := newExpTestEnv()
		expenseID := uuid.New()
		env.historyRepo.historiesByExpense[expenseID] = []model.ExpenseHistory{
			{
				BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Now()},
				ExpenseID: expenseID,
				Action:    "update",
				ChangedBy: "A",
				OldValue:  "draft",
				NewValue:  "pending",
			},
		}
		svc := NewExpenseHistoryService(env.deps)
		got, err := svc.GetHistory(ctx, expenseID)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		if len(got) != 1 || got[0].Action != "update" {
			t.Fatalf("unexpected result: %+v", got)
		}
	})
}

func expChdirTemp(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestExpenseReceiptService_Upload(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expChdirTemp(t, t.TempDir())
		svc := NewExpenseReceiptService(newExpTestEnv().deps)
		url, err := svc.Upload(ctx, "receipt.png", []byte("abc"))
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
		if !strings.HasPrefix(url, "/uploads/receipts/") && !strings.HasPrefix(url, "/uploads\\receipts\\") {
			t.Fatalf("unexpected url: %s", url)
		}
		if _, err := os.Stat(filepath.Join(strings.TrimPrefix(url, "/"))); err != nil {
			t.Fatalf("saved file not found: %v", err)
		}
	})

	t.Run("mkdir error", func(t *testing.T) {
		expChdirTemp(t, t.TempDir())
		if err := os.WriteFile("uploads", []byte("x"), 0644); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		svc := NewExpenseReceiptService(newExpTestEnv().deps)
		_, err := svc.Upload(ctx, "receipt.png", []byte("abc"))
		if err == nil {
			t.Fatalf("expected mkdir error")
		}
	})

	t.Run("write file error", func(t *testing.T) {
		expChdirTemp(t, t.TempDir())
		svc := NewExpenseReceiptService(newExpTestEnv().deps)
		_, err := svc.Upload(ctx, "bad.\x00txt", []byte("abc"))
		if err == nil {
			t.Fatalf("expected write error")
		}
	})
}

func TestExpenseTemplateService(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tmplID := uuid.New()

	t.Run("get templates", func(t *testing.T) {
		env := newExpTestEnv()
		env.templateRepo.templates[tmplID] = &model.ExpenseTemplate{
			BaseModel: model.BaseModel{ID: tmplID},
			UserID:    userID,
			Name:      "t1",
		}
		svc := NewExpenseTemplateService(env.deps)
		out, err := svc.GetTemplates(ctx, userID)
		if err != nil || len(out) != 1 {
			t.Fatalf("GetTemplates failed: %v", err)
		}
	})

	t.Run("create success and error", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseTemplateService(env.deps)
		got, err := svc.Create(ctx, userID, &model.ExpenseTemplateRequest{
			Name: "name", Title: "title", Category: "meals", Description: "desc", Amount: 1200,
		})
		if err != nil || got.Name != "name" {
			t.Fatalf("Create failed: %v", err)
		}
		env.templateRepo.createErr = errors.New("create err")
		if _, err := svc.Create(ctx, userID, &model.ExpenseTemplateRequest{Name: "x"}); err == nil {
			t.Fatalf("expected create error")
		}
	})

	t.Run("update find and update errors", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseTemplateService(env.deps)
		env.templateRepo.findByIDErr = errors.New("find err")
		if _, err := svc.Update(ctx, tmplID, &model.ExpenseTemplateRequest{Name: "x"}); err == nil {
			t.Fatalf("expected find error")
		}
		env.templateRepo.findByIDErr = nil
		env.templateRepo.templates[tmplID] = &model.ExpenseTemplate{
			BaseModel: model.BaseModel{ID: tmplID},
			UserID:    userID,
			Name:      "old",
		}
		env.templateRepo.updateErr = errors.New("update err")
		if _, err := svc.Update(ctx, tmplID, &model.ExpenseTemplateRequest{Name: "new"}); err == nil {
			t.Fatalf("expected update error")
		}
	})

	t.Run("update success, delete and use template", func(t *testing.T) {
		env := newExpTestEnv()
		env.templateRepo.templates[tmplID] = &model.ExpenseTemplate{
			BaseModel: model.BaseModel{ID: tmplID},
			UserID:    userID,
			Name:      "old",
			Title:     "t",
			Category:  model.ExpenseCategoryMeals,
			Amount:    500,
		}
		svc := NewExpenseTemplateService(env.deps)

		updated, err := svc.Update(ctx, tmplID, &model.ExpenseTemplateRequest{
			Name: "new", Title: "title2", Category: "transportation", Description: "d2", Amount: 999, IsRecurring: true, RecurringDay: 5,
		})
		if err != nil || updated.Name != "new" || updated.Amount != 999 {
			t.Fatalf("Update failed: %v", err)
		}
		if err := svc.Delete(ctx, tmplID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		env.templateRepo.findByIDErr = errors.New("find err")
		if _, err := svc.UseTemplate(ctx, tmplID, userID); err == nil {
			t.Fatalf("expected find error")
		}
		env.templateRepo.findByIDErr = nil
		env.templateRepo.templates[tmplID] = &model.ExpenseTemplate{
			BaseModel:   model.BaseModel{ID: tmplID},
			UserID:      userID,
			Title:       "ByTmpl",
			Category:    model.ExpenseCategorySupplies,
			Description: "desk",
			Amount:      1500,
		}
		env.expenseRepo.createErr = errors.New("create err")
		if _, err := svc.UseTemplate(ctx, tmplID, userID); err == nil {
			t.Fatalf("expected create error")
		}
		env.expenseRepo.createErr = nil
		created, err := svc.UseTemplate(ctx, tmplID, userID)
		if err != nil || created.Status != model.ExpenseStatusDraft || len(created.Items) != 1 {
			t.Fatalf("UseTemplate failed: %v", err)
		}
	})
}

func TestExpensePolicyService(t *testing.T) {
	ctx := context.Background()
	policyID := uuid.New()

	t.Run("wrappers", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpensePolicyService(env.deps)
		env.policyRepo.policies[policyID] = &model.ExpensePolicy{
			BaseModel: model.BaseModel{ID: policyID},
			Category:  model.ExpenseCategoryMeals,
		}
		env.budgetRepo.budgets = []model.ExpenseBudget{{FiscalYear: 2024}}
		userID := uuid.New()
		env.violateRepo.violationsByUser[userID] = []model.ExpensePolicyViolation{{UserID: userID}}

		policies, err := svc.GetPolicies(ctx)
		if err != nil || len(policies) != 1 {
			t.Fatalf("GetPolicies failed")
		}
		budgets, err := svc.GetBudgets(ctx)
		if err != nil || len(budgets) != 1 {
			t.Fatalf("GetBudgets failed")
		}
		violations, err := svc.GetPolicyViolations(ctx, userID)
		if err != nil || len(violations) != 1 {
			t.Fatalf("GetPolicyViolations failed")
		}
		if err := svc.Delete(ctx, policyID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("create branches", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpensePolicyService(env.deps)

		p1, err := svc.Create(ctx, &model.ExpensePolicyRequest{
			Category: "meals", MonthlyLimit: 10000, Description: "d",
		})
		if err != nil || !p1.IsActive {
			t.Fatalf("Create default active failed")
		}
		p2, err := svc.Create(ctx, &model.ExpensePolicyRequest{
			Category: "other", IsActive: expPtrBool(false),
		})
		if err != nil || p2.IsActive {
			t.Fatalf("Create with explicit IsActive failed")
		}
		env.policyRepo.createErr = errors.New("create err")
		if _, err := svc.Create(ctx, &model.ExpensePolicyRequest{Category: "x"}); err == nil {
			t.Fatalf("expected create error")
		}
	})

	t.Run("update branches", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpensePolicyService(env.deps)
		env.policyRepo.findByIDErr = errors.New("find err")
		if _, err := svc.Update(ctx, policyID, &model.ExpensePolicyRequest{Category: "x"}); err == nil {
			t.Fatalf("expected find error")
		}

		env.policyRepo.findByIDErr = nil
		env.policyRepo.policies[policyID] = &model.ExpensePolicy{
			BaseModel: model.BaseModel{ID: policyID},
			Category:  model.ExpenseCategoryMeals,
			IsActive:  true,
		}
		env.policyRepo.updateErr = errors.New("update err")
		if _, err := svc.Update(ctx, policyID, &model.ExpensePolicyRequest{Category: "supplies"}); err == nil {
			t.Fatalf("expected update error")
		}

		env.policyRepo.updateErr = nil
		out, err := svc.Update(ctx, policyID, &model.ExpensePolicyRequest{
			Category: "supplies",
			IsActive: expPtrBool(false),
		})
		if err != nil || out.Category != model.ExpenseCategorySupplies || out.IsActive {
			t.Fatalf("Update with isActive failed")
		}

		out2, err := svc.Update(ctx, policyID, &model.ExpensePolicyRequest{Category: "meals"})
		if err != nil || out2.Category != model.ExpenseCategoryMeals {
			t.Fatalf("Update nil isActive failed")
		}
	})
}

func TestExpenseNotificationService(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	notifID := uuid.New()
	reminderID := uuid.New()

	t.Run("wrappers", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseNotificationService(env.deps)
		env.notifRepo.notifications[notifID] = &model.ExpenseNotification{
			BaseModel: model.BaseModel{ID: notifID},
			UserID:    userID,
		}
		env.reminderRepo.remindersByUser[userID] = []model.ExpenseReminder{
			{BaseModel: model.BaseModel{ID: reminderID}, UserID: userID},
		}
		env.settingRepo.settingsByUser[userID] = &model.ExpenseNotificationSetting{UserID: userID}

		list, err := svc.GetNotifications(ctx, userID, "")
		if err != nil || len(list) != 1 {
			t.Fatalf("GetNotifications failed")
		}
		if err := svc.MarkAsRead(ctx, notifID); err != nil {
			t.Fatalf("MarkAsRead failed: %v", err)
		}
		if err := svc.MarkAllAsRead(ctx, userID); err != nil {
			t.Fatalf("MarkAllAsRead failed: %v", err)
		}
		reminders, err := svc.GetReminders(ctx, userID)
		if err != nil || len(reminders) != 1 {
			t.Fatalf("GetReminders failed")
		}
		if err := svc.DismissReminder(ctx, reminderID); err != nil {
			t.Fatalf("DismissReminder failed: %v", err)
		}
		setting, err := svc.GetSettings(ctx, userID)
		if err != nil || setting.UserID != userID {
			t.Fatalf("GetSettings failed")
		}
	})

	t.Run("update settings find and upsert errors", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseNotificationService(env.deps)

		env.settingRepo.findErr = errors.New("find err")
		if _, err := svc.UpdateSettings(ctx, userID, &model.ExpenseNotificationSettingRequest{}); err == nil {
			t.Fatalf("expected find error")
		}
		env.settingRepo.findErr = nil

		env.settingRepo.settingsByUser[userID] = &model.ExpenseNotificationSetting{
			UserID:              userID,
			EmailEnabled:        true,
			PushEnabled:         true,
			ApprovalAlerts:      true,
			ReimbursementAlerts: true,
			PolicyAlerts:        true,
			WeeklyDigest:        false,
		}
		env.settingRepo.upsertErr = errors.New("upsert err")
		if _, err := svc.UpdateSettings(ctx, userID, &model.ExpenseNotificationSettingRequest{}); err == nil {
			t.Fatalf("expected upsert error")
		}
	})

	t.Run("update settings all nil and all set", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseNotificationService(env.deps)
		env.settingRepo.settingsByUser[userID] = &model.ExpenseNotificationSetting{
			UserID:              userID,
			EmailEnabled:        true,
			PushEnabled:         true,
			ApprovalAlerts:      true,
			ReimbursementAlerts: true,
			PolicyAlerts:        true,
			WeeklyDigest:        false,
		}

		s1, err := svc.UpdateSettings(ctx, userID, &model.ExpenseNotificationSettingRequest{})
		if err != nil || !s1.EmailEnabled || !s1.PushEnabled {
			t.Fatalf("UpdateSettings nil pointers failed: %v", err)
		}

		s2, err := svc.UpdateSettings(ctx, userID, &model.ExpenseNotificationSettingRequest{
			EmailEnabled:        expPtrBool(false),
			PushEnabled:         expPtrBool(false),
			ApprovalAlerts:      expPtrBool(false),
			ReimbursementAlerts: expPtrBool(false),
			PolicyAlerts:        expPtrBool(false),
			WeeklyDigest:        expPtrBool(true),
		})
		if err != nil {
			t.Fatalf("UpdateSettings failed: %v", err)
		}
		if s2.EmailEnabled || s2.PushEnabled || s2.ApprovalAlerts || s2.ReimbursementAlerts || s2.PolicyAlerts || !s2.WeeklyDigest {
			t.Fatalf("unexpected setting values: %+v", s2)
		}
	})
}

func TestExpenseApprovalFlowService(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	delegateID := uuid.New()

	t.Run("wrappers", func(t *testing.T) {
		env := newExpTestEnv()
		env.flowRepo.active = &model.ExpenseApprovalFlow{Name: "default"}
		env.delegateRepo.delegates[uuid.New()] = &model.ExpenseDelegate{
			BaseModel: model.BaseModel{ID: uuid.New()},
			UserID:    userID,
		}
		svc := NewExpenseApprovalFlowService(env.deps)
		cfg, err := svc.GetConfig(ctx)
		if err != nil || cfg.Name != "default" {
			t.Fatalf("GetConfig failed")
		}
		list, err := svc.GetDelegates(ctx, userID)
		if err != nil || len(list) != 1 {
			t.Fatalf("GetDelegates failed")
		}
		for id := range env.delegateRepo.delegates {
			if err := svc.RemoveDelegate(ctx, id); err != nil {
				t.Fatalf("RemoveDelegate failed: %v", err)
			}
			break
		}
	})

	t.Run("set delegate invalid id", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseApprovalFlowService(env.deps)
		_, err := svc.SetDelegate(ctx, userID, &model.ExpenseDelegateRequest{
			DelegateTo: "invalid",
			StartDate:  "2024-01-01",
			EndDate:    "2024-01-31",
		})
		if err == nil {
			t.Fatalf("expected parse error")
		}
	})

	t.Run("set delegate create error", func(t *testing.T) {
		env := newExpTestEnv()
		env.delegateRepo.createErr = errors.New("create err")
		svc := NewExpenseApprovalFlowService(env.deps)
		_, err := svc.SetDelegate(ctx, userID, &model.ExpenseDelegateRequest{
			DelegateTo: delegateID.String(),
			StartDate:  "2024-01-01",
			EndDate:    "2024-01-31",
		})
		if err == nil {
			t.Fatalf("expected create error")
		}
	})

	t.Run("set delegate success", func(t *testing.T) {
		env := newExpTestEnv()
		svc := NewExpenseApprovalFlowService(env.deps)
		d, err := svc.SetDelegate(ctx, userID, &model.ExpenseDelegateRequest{
			DelegateTo: delegateID.String(),
			StartDate:  "2024-01-01",
			EndDate:    "2024-01-31",
		})
		if err != nil {
			t.Fatalf("SetDelegate failed: %v", err)
		}
		if d.UserID != userID || d.DelegateID != delegateID || !d.IsActive {
			t.Fatalf("unexpected delegate: %+v", d)
		}
	})
}
