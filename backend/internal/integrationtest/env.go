package integrationtest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"github.com/your-org/kintai/backend/internal/router"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultDatabaseURL = "postgres://kintai:kintai@localhost:5432/kintai_test?sslmode=disable"
	defaultJWTSecret   = "integration-test-secret"
)

type Options struct {
	DatabaseURL                  string
	JWTSecretKey                 string
	AllowedOrigins               []string
	RateLimitRPS                 int
	RateLimitBurst               int
	SeedFiles                    []string
	DisableAdditionalAutoMigrate bool
	FailIfDatabaseUnavailable    bool
}

type options struct {
	databaseURL                string
	jwtSecretKey               string
	allowedOrigins             []string
	rateLimitRPS               int
	rateLimitBurst             int
	seedFiles                  []string
	applyAdditionalAutoMigrate bool
	skipIfDatabaseUnavailable  bool
}

type TestEnv struct {
	DB          *gorm.DB
	Router      *gin.Engine
	Config      *config.Config
	Logger      *logger.Logger
	backendRoot string
}

var (
	backendRootOnce sync.Once
	backendRoot     string
	backendRootErr  error
)

func NewTestEnv(t testing.TB, input *Options) *TestEnv {
	t.Helper()

	opts := normalizeOptions(input)
	root, err := detectBackendRoot()
	if err != nil {
		t.Fatalf("failed to detect backend root: %v", err)
	}

	db, err := openDB(opts.databaseURL)
	if err != nil {
		if opts.skipIfDatabaseUnavailable {
			t.Skipf("skipping integration test because database is unavailable: %v", err)
		}
		t.Fatalf("failed to open database: %v", err)
	}

	cfg := &config.Config{
		Env:                   "test",
		Port:                  "0",
		DatabaseURL:           opts.databaseURL,
		RedisURL:              "redis://localhost:6379/0",
		JWTSecretKey:          opts.jwtSecretKey,
		JWTAccessTokenExpiry:  15,
		JWTRefreshTokenExpiry: 168,
		AllowedOrigins:        append([]string(nil), opts.allowedOrigins...),
		RateLimitRPS:          opts.rateLimitRPS,
		RateLimitBurst:        opts.rateLimitBurst,
		LogLevel:              "error",
	}

	log, err := logger.NewLogger(cfg.LogLevel, "development")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	env := &TestEnv{
		DB:          db,
		Config:      cfg,
		Logger:      log,
		backendRoot: root,
	}

	if err := env.ResetSchema(); err != nil {
		t.Fatalf("failed to reset schema: %v", err)
	}
	if err := env.ApplyMigrations(); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	if opts.applyAdditionalAutoMigrate {
		if err := applyAdditionalAutoMigrations(env.DB); err != nil {
			t.Fatalf("failed to apply additional automigrate: %v", err)
		}
	}
	if len(opts.seedFiles) > 0 {
		if err := env.LoadSeeds(opts.seedFiles...); err != nil {
			t.Fatalf("failed to load seed files: %v", err)
		}
	}

	gin.SetMode(gin.TestMode)
	env.Router = buildRouter(env.DB, env.Config, env.Logger)

	t.Cleanup(func() {
		if sqlDB, err := env.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return env
}

func (e *TestEnv) ResetSchema() error {
	if err := e.DB.Exec(`DROP SCHEMA IF EXISTS public CASCADE`).Error; err != nil {
		return fmt.Errorf("drop public schema: %w", err)
	}
	if err := e.DB.Exec(`CREATE SCHEMA public`).Error; err != nil {
		return fmt.Errorf("create public schema: %w", err)
	}
	return nil
}

func (e *TestEnv) ApplyMigrations() error {
	files, err := migrationFiles(e.backendRoot)
	if err != nil {
		return err
	}

	for _, file := range files {
		sqlContent, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", file, err)
		}

		query := strings.TrimSpace(string(sqlContent))
		if query == "" {
			continue
		}

		if err := e.DB.Exec(query).Error; err != nil {
			return fmt.Errorf("apply migration %s: %w", filepath.Base(file), err)
		}
	}

	return nil
}

func (e *TestEnv) ResetDB() error {
	type tableRow struct {
		Name string `gorm:"column:tablename"`
	}

	var rows []tableRow
	if err := e.DB.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename
	`).Scan(&rows).Error; err != nil {
		return fmt.Errorf("list tables: %w", err)
	}

	tables := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Name == "" || row.Name == "schema_migrations" {
			continue
		}
		tables = append(tables, quoteIdentifier(row.Name))
	}

	if len(tables) == 0 {
		return nil
	}

	stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))
	if err := e.DB.Exec(stmt).Error; err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}

	return nil
}

func (e *TestEnv) LoadSeeds(seedPaths ...string) error {
	for _, seedPath := range seedPaths {
		fullPath := seedPath
		if !filepath.IsAbs(seedPath) {
			fullPath = filepath.Join(e.backendRoot, seedPath)
		}

		sqlContent, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read seed file %s: %w", fullPath, err)
		}

		query := strings.TrimSpace(string(sqlContent))
		if query == "" {
			continue
		}

		if err := e.DB.Exec(query).Error; err != nil {
			return fmt.Errorf("apply seed file %s: %w", fullPath, err)
		}
	}
	return nil
}

func buildRouter(db *gorm.DB, cfg *config.Config, log *logger.Logger) *gin.Engine {
	repos := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{
		Repos:  repos,
		Config: cfg,
		Logger: log,
	})
	handlers := handler.NewHandlers(services, log)
	mw := middleware.NewMiddleware(cfg, log)

	engine := gin.New()
	router.Setup(engine, handlers, mw)
	return engine
}

func openDB(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func applyAdditionalAutoMigrations(db *gorm.DB) error {
	if err := model.AutoMigrate(db); err != nil {
		return err
	}
	if err := model.HRAutoMigrate(db); err != nil {
		return err
	}
	if err := model.ExpenseAutoMigrate(db); err != nil {
		return err
	}
	return nil
}

func migrationFiles(root string) ([]string, error) {
	dir := filepath.Join(root, "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}

	sort.Strings(files)
	return files, nil
}

func normalizeOptions(input *Options) options {
	o := options{
		databaseURL:                defaultDatabaseURL,
		jwtSecretKey:               defaultJWTSecret,
		allowedOrigins:             []string{"http://localhost:5173"},
		rateLimitRPS:               100000,
		rateLimitBurst:             100000,
		applyAdditionalAutoMigrate: false,
		skipIfDatabaseUnavailable:  true,
	}

	if input == nil {
		return o
	}

	if input.DatabaseURL != "" {
		o.databaseURL = input.DatabaseURL
	}
	if input.JWTSecretKey != "" {
		o.jwtSecretKey = input.JWTSecretKey
	}
	if len(input.AllowedOrigins) > 0 {
		o.allowedOrigins = append([]string(nil), input.AllowedOrigins...)
	}
	if input.RateLimitRPS > 0 {
		o.rateLimitRPS = input.RateLimitRPS
	}
	if input.RateLimitBurst > 0 {
		o.rateLimitBurst = input.RateLimitBurst
	}
	if input.SeedFiles != nil {
		o.seedFiles = append([]string(nil), input.SeedFiles...)
	}
	o.applyAdditionalAutoMigrate = !input.DisableAdditionalAutoMigrate
	o.skipIfDatabaseUnavailable = !input.FailIfDatabaseUnavailable

	return o
}

func detectBackendRoot() (string, error) {
	backendRootOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			backendRootErr = fmt.Errorf("get working directory: %w", err)
			return
		}
		root, err := walkUpToGoMod(wd)
		if err != nil {
			backendRootErr = err
			return
		}
		backendRoot = root
	})

	return backendRoot, backendRootErr
}

func walkUpToGoMod(start string) (string, error) {
	dir := filepath.Clean(start)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found while walking parent directories")
		}
		dir = parent
	}
}

func quoteIdentifier(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return `"` + escaped + `"`
}
