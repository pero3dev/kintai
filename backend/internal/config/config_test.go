package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// 環境変数をクリア
	envVars := []string{
		"APP_ENV", "APP_PORT", "DATABASE_URL", "REDIS_URL",
		"JWT_SECRET_KEY", "JWT_ACCESS_TOKEN_EXPIRY", "JWT_REFRESH_TOKEN_EXPIRY",
		"ALLOWED_ORIGINS", "RATE_LIMIT_RPS", "RATE_LIMIT_BURST",
		"AWS_REGION", "SES_FROM_EMAIL", "SENTRY_DSN", "OTLP_ENDPOINT", "LOG_LEVEL",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// デフォルト値を確認
	if cfg.Env != "development" {
		t.Errorf("Expected Env 'development', got '%s'", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Errorf("Expected Port '8080', got '%s'", cfg.Port)
	}
	if cfg.JWTAccessTokenExpiry != 15 {
		t.Errorf("Expected JWTAccessTokenExpiry 15, got %d", cfg.JWTAccessTokenExpiry)
	}
	if cfg.JWTRefreshTokenExpiry != 168 {
		t.Errorf("Expected JWTRefreshTokenExpiry 168, got %d", cfg.JWTRefreshTokenExpiry)
	}
	if cfg.RateLimitRPS != 100 {
		t.Errorf("Expected RateLimitRPS 100, got %d", cfg.RateLimitRPS)
	}
	if cfg.RateLimitBurst != 200 {
		t.Errorf("Expected RateLimitBurst 200, got %d", cfg.RateLimitBurst)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	// カスタム環境変数を設定
	os.Setenv("APP_ENV", "staging")
	os.Setenv("APP_PORT", "3000")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("JWT_SECRET_KEY", "custom-secret")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "30")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY", "720")
	os.Setenv("RATE_LIMIT_RPS", "50")
	os.Setenv("RATE_LIMIT_BURST", "100")
	os.Setenv("LOG_LEVEL", "info")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET_KEY")
		os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRY")
		os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRY")
		os.Unsetenv("RATE_LIMIT_RPS")
		os.Unsetenv("RATE_LIMIT_BURST")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Env != "staging" {
		t.Errorf("Expected Env 'staging', got '%s'", cfg.Env)
	}
	if cfg.Port != "3000" {
		t.Errorf("Expected Port '3000', got '%s'", cfg.Port)
	}
	if cfg.JWTSecretKey != "custom-secret" {
		t.Errorf("Expected JWTSecretKey 'custom-secret', got '%s'", cfg.JWTSecretKey)
	}
	if cfg.JWTAccessTokenExpiry != 30 {
		t.Errorf("Expected JWTAccessTokenExpiry 30, got %d", cfg.JWTAccessTokenExpiry)
	}
	if cfg.JWTRefreshTokenExpiry != 720 {
		t.Errorf("Expected JWTRefreshTokenExpiry 720, got %d", cfg.JWTRefreshTokenExpiry)
	}
	if cfg.RateLimitRPS != 50 {
		t.Errorf("Expected RateLimitRPS 50, got %d", cfg.RateLimitRPS)
	}
	if cfg.RateLimitBurst != 100 {
		t.Errorf("Expected RateLimitBurst 100, got %d", cfg.RateLimitBurst)
	}
}

func TestLoad_ProductionWithDefaultSecret(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("JWT_SECRET_KEY")

	defer func() {
		os.Unsetenv("APP_ENV")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for production with default secret, got nil")
	}
}

func TestLoad_ProductionWithCustomSecret(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET_KEY", "production-secret-key")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("JWT_SECRET_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Env != "production" {
		t.Errorf("Expected Env 'production', got '%s'", cfg.Env)
	}
	if cfg.JWTSecretKey != "production-secret-key" {
		t.Errorf("Expected JWTSecretKey 'production-secret-key', got '%s'", cfg.JWTSecretKey)
	}
}

func TestGetEnvAsInt_InvalidValue(t *testing.T) {
	os.Setenv("TEST_INVALID_INT", "not-a-number")
	defer os.Unsetenv("TEST_INVALID_INT")

	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "not-a-number")
	defer os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 無効な値の場合はデフォルト値が使われる
	if cfg.JWTAccessTokenExpiry != 15 {
		t.Errorf("Expected JWTAccessTokenExpiry 15 (default), got %d", cfg.JWTAccessTokenExpiry)
	}
}
