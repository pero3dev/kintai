package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/gorm"
)

func TestExpenseRepositoryErrBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expenses"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "expenses"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("Taxi"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("template find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseTemplateRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_templates"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_templates"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Template"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("policy find by id and category", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpensePolicyRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_policies"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_policies"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"category"}).AddRow("meals"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_policies"`, 2).WillReturnError(errors.New("boom"))
		_, err = repo.FindByCategory(ctx, "meals")
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_policies"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"category"}).AddRow("meals"))
		_, err = repo.FindByCategory(ctx, "meals")
		require.NoError(t, err)
	})
}

func TestExpenseRepositoryFilterBranchesWithDryRun(t *testing.T) {
	ctx := context.Background()
	db := dryRunDB(t)

	repo := NewExpenseRepository(db)
	_, _, err := repo.FindByUserID(ctx, uuid.New(), 1, 10, "", "")
	require.NoError(t, err)
	_, _, err = repo.FindByUserID(ctx, uuid.New(), 1, 10, "pending", "meals")
	require.NoError(t, err)

	_, _, err = repo.FindAll(ctx, 1, 10, "", "")
	require.NoError(t, err)
	_, _, err = repo.FindAll(ctx, 1, 10, "approved", "transportation")
	require.NoError(t, err)

	itemRepo := NewExpenseItemRepository(db)
	err = itemRepo.CreateBatch(ctx, nil)
	require.NoError(t, err)
	err = itemRepo.CreateBatch(ctx, []model.ExpenseItem{{Description: "x"}})
	require.NoError(t, err)

	notifRepo := NewExpenseNotificationRepository(db)
	_, err = notifRepo.FindByUserID(ctx, uuid.New(), "")
	require.NoError(t, err)
	_, err = notifRepo.FindByUserID(ctx, uuid.New(), "unread")
	require.NoError(t, err)
}

func TestExpenseRepositoryGetReportSwitch(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewExpenseRepository(db)

	expectQuery(mock, `(?i)SELECT .*SUM\(total_amount\).*FROM "expenses"`, 2).
		WillReturnRows(sqlmock.NewRows([]string{"total_amount"}).AddRow(1000.0))

	expectQuery(mock, `(?i)SELECT .*expense_items\.category.*FROM "expense_items"`, 2).
		WillReturnRows(sqlmock.NewRows([]string{"category", "amount"}).AddRow("meals", 200.0))

	expectQuery(mock, `(?i)SELECT .*status, COUNT\(\*\) as count.*FROM "expenses"`, 2).
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("draft", 1).
			AddRow("pending", 2).
			AddRow("approved", 3).
			AddRow("rejected", 4).
			AddRow("reimbursed", 5))

	expectQuery(mock, `(?i)SELECT .*department.*FROM "expenses"`, 2).
		WillReturnRows(sqlmock.NewRows([]string{"department", "amount", "count", "avg"}).
			AddRow("Engineering", 800.0, 4, 200.0))

	report, err := repo.GetReport(context.Background(), time.Now().AddDate(0, -1, 0), time.Now())
	require.NoError(t, err)
	require.Equal(t, 1, report.StatusSummary.Draft)
	require.Equal(t, 2, report.StatusSummary.Pending)
	require.Equal(t, 3, report.StatusSummary.Approved)
	require.Equal(t, 4, report.StatusSummary.Rejected)
	require.Equal(t, 5, report.StatusSummary.Reimbursed)
}

func TestExpenseNotificationSettingRepositoryBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("record not found", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseNotificationSettingRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_notification_settings"`, 2).WillReturnError(gorm.ErrRecordNotFound)
		got, err := repo.FindByUserID(ctx, uuid.New())
		require.NoError(t, err)
		require.NotNil(t, got)
		require.True(t, got.EmailEnabled)
	})

	t.Run("other error", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseNotificationSettingRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_notification_settings"`, 2).WillReturnError(errors.New("boom"))
		got, err := repo.FindByUserID(ctx, uuid.New())
		require.Error(t, err)
		require.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseNotificationSettingRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_notification_settings"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email_enabled"}).AddRow(uuid.NewString(), true))
		got, err := repo.FindByUserID(ctx, uuid.New())
		require.NoError(t, err)
		require.NotNil(t, got)
	})
}

func TestExpenseApprovalFlowRepositoryBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("record not found", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseApprovalFlowRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_approval_flows"`, 1).WillReturnError(gorm.ErrRecordNotFound)
		got, err := repo.FindActive(ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.True(t, got.IsActive)
	})

	t.Run("other error", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseApprovalFlowRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_approval_flows"`, 1).WillReturnError(errors.New("boom"))
		got, err := repo.FindActive(ctx)
		require.Error(t, err)
		require.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewExpenseApprovalFlowRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "expense_approval_flows"`, 1).
			WillReturnRows(sqlmock.NewRows([]string{"name", "is_active"}).AddRow("custom", true))
		got, err := repo.FindActive(ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "custom", got.Name)
	})
}

