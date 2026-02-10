package model

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newModelTestDB(t *testing.T, dryRun bool) (*gorm.DB, *sql.DB) {
	t.Helper()

	sqlDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	if dryRun {
		db = db.Session(&gorm.Session{DryRun: true})
	}

	return db, sqlDB
}

func TestAutoMigrate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, _ := newModelTestDB(t, true)

		if err := AutoMigrate(db); err != nil {
			t.Fatalf("AutoMigrate() returned error: %v", err)
		}
	})

	t.Run("error on closed db", func(t *testing.T) {
		db, sqlDB := newModelTestDB(t, false)
		_ = sqlDB.Close()

		if err := AutoMigrate(db); err == nil {
			t.Fatal("AutoMigrate() expected error on closed db, got nil")
		}
	})
}

func TestHRAutoMigrate(t *testing.T) {
	t.Run("success restores fk flag", func(t *testing.T) {
		db, _ := newModelTestDB(t, true)
		db.Config.DisableForeignKeyConstraintWhenMigrating = false

		if err := HRAutoMigrate(db); err != nil {
			t.Fatalf("HRAutoMigrate() returned error: %v", err)
		}
		if db.Config.DisableForeignKeyConstraintWhenMigrating != false {
			t.Fatal("fk flag was not restored after success")
		}
	})

	t.Run("error restores fk flag", func(t *testing.T) {
		db, sqlDB := newModelTestDB(t, false)
		db.Config.DisableForeignKeyConstraintWhenMigrating = true
		_ = sqlDB.Close()

		if err := HRAutoMigrate(db); err == nil {
			t.Fatal("HRAutoMigrate() expected error on closed db, got nil")
		}
		if db.Config.DisableForeignKeyConstraintWhenMigrating != true {
			t.Fatal("fk flag was not restored after error")
		}
	})
}

func TestExpenseAutoMigrate(t *testing.T) {
	t.Run("success restores fk flag", func(t *testing.T) {
		db, _ := newModelTestDB(t, true)
		db.Config.DisableForeignKeyConstraintWhenMigrating = true

		if err := ExpenseAutoMigrate(db); err != nil {
			t.Fatalf("ExpenseAutoMigrate() returned error: %v", err)
		}
		if db.Config.DisableForeignKeyConstraintWhenMigrating != true {
			t.Fatal("fk flag was not restored after success")
		}
	})

	t.Run("error restores fk flag", func(t *testing.T) {
		db, sqlDB := newModelTestDB(t, false)
		db.Config.DisableForeignKeyConstraintWhenMigrating = false
		_ = sqlDB.Close()

		if err := ExpenseAutoMigrate(db); err == nil {
			t.Fatal("ExpenseAutoMigrate() expected error on closed db, got nil")
		}
		if db.Config.DisableForeignKeyConstraintWhenMigrating != false {
			t.Fatal("fk flag was not restored after error")
		}
	})
}
