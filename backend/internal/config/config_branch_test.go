package config

import (
	"os"
	"strings"
	"testing"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	orig, ok := os.LookupEnv(key)
	_ = os.Unsetenv(key)
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, orig)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func TestGetEnv_Branches(t *testing.T) {
	t.Run("returns env value when key exists", func(t *testing.T) {
		t.Setenv("GET_ENV_KEY", "set-value")

		got := getEnv("GET_ENV_KEY", "default-value")
		if got != "set-value" {
			t.Fatalf("expected set-value, got %s", got)
		}
	})

	t.Run("returns default when key does not exist", func(t *testing.T) {
		unsetEnv(t, "GET_ENV_KEY_MISSING")

		got := getEnv("GET_ENV_KEY_MISSING", "default-value")
		if got != "default-value" {
			t.Fatalf("expected default-value, got %s", got)
		}
	})
}

func TestGetEnvAsInt_Branches(t *testing.T) {
	t.Run("returns parsed int when key exists and valid", func(t *testing.T) {
		t.Setenv("GET_ENV_INT_KEY", "42")

		got := getEnvAsInt("GET_ENV_INT_KEY", 7)
		if got != 42 {
			t.Fatalf("expected 42, got %d", got)
		}
	})

	t.Run("returns default when key exists and invalid", func(t *testing.T) {
		t.Setenv("GET_ENV_INT_KEY_INVALID", "abc")

		got := getEnvAsInt("GET_ENV_INT_KEY_INVALID", 7)
		if got != 7 {
			t.Fatalf("expected default 7, got %d", got)
		}
	})

	t.Run("returns default when key does not exist", func(t *testing.T) {
		unsetEnv(t, "GET_ENV_INT_KEY_MISSING")

		got := getEnvAsInt("GET_ENV_INT_KEY_MISSING", 9)
		if got != 9 {
			t.Fatalf("expected default 9, got %d", got)
		}
	})
}

func TestLoad_Branches(t *testing.T) {
	t.Run("production with default jwt secret returns error", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		unsetEnv(t, "JWT_SECRET_KEY")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !strings.Contains(err.Error(), "JWT_SECRET_KEY") {
			t.Fatalf("expected JWT_SECRET_KEY error, got %v", err)
		}
	})

	t.Run("allowed origins uses env value", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("ALLOWED_ORIGINS", "https://example.com")
		t.Setenv("JWT_SECRET_KEY", "custom-secret")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}
		if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "https://example.com" {
			t.Fatalf("unexpected AllowedOrigins: %#v", cfg.AllowedOrigins)
		}
	})
}
