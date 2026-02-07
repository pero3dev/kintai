package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config はアプリケーション設定を保持する構造体
type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	RedisURL    string

	// JWT
	JWTSecretKey          string
	JWTAccessTokenExpiry  int // 分
	JWTRefreshTokenExpiry int // 時間

	// CORS
	AllowedOrigins []string

	// Rate Limiting
	RateLimitRPS   int
	RateLimitBurst int

	// AWS SES
	AWSRegion       string
	SESFromEmail    string

	// Sentry
	SentryDSN string

	// OpenTelemetry
	OTLPEndpoint string

	// ログレベル
	LogLevel string
}

// Load は環境変数から設定を読み込む
func Load() (*Config, error) {
	cfg := &Config{
		Env:                   getEnv("APP_ENV", "development"),
		Port:                  getEnv("APP_PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://kintai:kintai@localhost:5432/kintai?sslmode=disable"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecretKey:          getEnv("JWT_SECRET_KEY", "dev-secret-key-change-in-production"),
		JWTAccessTokenExpiry:  getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRY", 15),
		JWTRefreshTokenExpiry: getEnvAsInt("JWT_REFRESH_TOKEN_EXPIRY", 168),
		AllowedOrigins:        []string{getEnv("ALLOWED_ORIGINS", "http://localhost:5173")},
		RateLimitRPS:          getEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst:        getEnvAsInt("RATE_LIMIT_BURST", 200),
		AWSRegion:             getEnv("AWS_REGION", "ap-northeast-1"),
		SESFromEmail:          getEnv("SES_FROM_EMAIL", "noreply@example.com"),
		SentryDSN:             getEnv("SENTRY_DSN", ""),
		OTLPEndpoint:          getEnv("OTLP_ENDPOINT", "localhost:4317"),
		LogLevel:              getEnv("LOG_LEVEL", "debug"),
	}

	if cfg.Env == "production" && cfg.JWTSecretKey == "dev-secret-key-change-in-production" {
		return nil, fmt.Errorf("本番環境ではJWT_SECRET_KEYを設定してください")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
