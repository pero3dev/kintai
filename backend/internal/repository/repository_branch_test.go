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

func TestRepositoryCoreErrBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("user find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewUserRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "users"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "users"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("a@example.com"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("user find by email", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewUserRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "users"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByEmail(ctx, "x@example.com")
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "users"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("x@example.com"))
		_, err = repo.FindByEmail(ctx, "x@example.com")
		require.NoError(t, err)
	})

	t.Run("attendance find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewAttendanceRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "attendances"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "attendances"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"date"}).AddRow(time.Now()))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("attendance find by user and date", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewAttendanceRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "attendances"`, 3).WillReturnError(errors.New("boom"))
		_, err := repo.FindByUserAndDate(ctx, uuid.New(), time.Now())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "attendances"`, 3).
			WillReturnRows(sqlmock.NewRows([]string{"date"}).AddRow(time.Now()))
		_, err = repo.FindByUserAndDate(ctx, uuid.New(), time.Now())
		require.NoError(t, err)
	})

	t.Run("count today absent", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewAttendanceRepository(db)

		expectQuery(mock, `(?i)SELECT count\(\*\).*FROM "attendances"`, 1).WillReturnError(errors.New("boom"))
		_, err := repo.CountTodayAbsent(ctx, 10)
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT count\(\*\).*FROM "attendances"`, 1).WillReturnRows(countRows(3))
		got, err := repo.CountTodayAbsent(ctx, 10)
		require.NoError(t, err)
		require.EqualValues(t, 7, got)
	})

	t.Run("leave request find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewLeaveRequestRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "leave_requests"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "leave_requests"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"reason"}).AddRow("private"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("department find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewDepartmentRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "departments"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "departments"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Eng"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("refresh token find by token", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewRefreshTokenRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "refresh_tokens"`, 3).WillReturnError(errors.New("boom"))
		_, err := repo.FindByToken(ctx, "token")
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "refresh_tokens"`, 3).
			WillReturnRows(sqlmock.NewRows([]string{"token"}).AddRow("token"))
		_, err = repo.FindByToken(ctx, "token")
		require.NoError(t, err)
	})

	t.Run("overtime request find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewOvertimeRequestRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "overtime_requests"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "overtime_requests"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"reason"}).AddRow("release"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("leave balance find by user year and type", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewLeaveBalanceRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "leave_balances"`, 4).WillReturnError(errors.New("boom"))
		_, err := repo.FindByUserYearAndType(ctx, uuid.New(), 2026, model.LeaveTypePaid)
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "leave_balances"`, 4).
			WillReturnRows(sqlmock.NewRows([]string{"leave_type"}).AddRow("paid"))
		_, err = repo.FindByUserYearAndType(ctx, uuid.New(), 2026, model.LeaveTypePaid)
		require.NoError(t, err)
	})

	t.Run("attendance correction find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewAttendanceCorrectionRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "attendance_corrections"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "attendance_corrections"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"reason"}).AddRow("fix"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("project find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewProjectRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "projects"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "projects"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("P"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("time entry find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewTimeEntryRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "time_entries"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "time_entries"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"minutes"}).AddRow(60))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("holiday find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewHolidayRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "holidays"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "holidays"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("New Year"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})

	t.Run("is holiday", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewHolidayRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "holidays"`, 2).WillReturnError(gorm.ErrRecordNotFound)
		ok, holiday, err := repo.IsHoliday(ctx, time.Now())
		require.NoError(t, err)
		require.False(t, ok)
		require.Nil(t, holiday)

		expectQuery(mock, `(?i)SELECT .*FROM "holidays"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"name", "holiday_type"}).AddRow("Holiday", "national"))
		ok, holiday, err = repo.IsHoliday(ctx, time.Now())
		require.NoError(t, err)
		require.True(t, ok)
		require.NotNil(t, holiday)
	})

	t.Run("approval flow find by id", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewApprovalFlowRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "approval_flows"`, 2).WillReturnError(errors.New("boom"))
		_, err := repo.FindByID(ctx, uuid.New())
		require.Error(t, err)

		expectQuery(mock, `(?i)SELECT .*FROM "approval_flows"`, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("00000000-0000-0000-0000-000000000000", "flow"))
		_, err = repo.FindByID(ctx, uuid.New())
		require.NoError(t, err)
	})
}

func TestRepositoryFilterBranchesWithDryRun(t *testing.T) {
	ctx := context.Background()
	db := dryRunDB(t)

	notifRepo := NewNotificationRepository(db)
	_, _, err := notifRepo.FindByUserID(ctx, uuid.New(), nil, 1, 10)
	require.NoError(t, err)

	read := true
	_, _, err = notifRepo.FindByUserID(ctx, uuid.New(), &read, 1, 10)
	require.NoError(t, err)

	projectRepo := NewProjectRepository(db)
	_, _, err = projectRepo.FindAll(ctx, nil, 1, 10)
	require.NoError(t, err)

	status := model.ProjectStatusActive
	_, _, err = projectRepo.FindAll(ctx, &status, 1, 10)
	require.NoError(t, err)
}

