package repository

import (
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableAutomaticPing:  true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return db, mock, func() {
		t.Helper()
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, _, cleanup := newMockDB(t)
	t.Cleanup(cleanup)
	return db.Session(&gorm.Session{DryRun: true})
}

func expectQuery(mock sqlmock.Sqlmock, pattern string, argCount int) *sqlmock.ExpectedQuery {
	q := mock.ExpectQuery(pattern)
	if argCount > 0 {
		q.WithArgs(anyArgs(argCount)...)
	}
	return q
}

func anyArgs(n int) []driver.Value {
	args := make([]driver.Value, n)
	for i := 0; i < n; i++ {
		args[i] = sqlmock.AnyArg()
	}
	return args
}

func countRows(n int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(n)
}

func emptyRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id"})
}

