package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

func setupRouterForBranchTest(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		JWTSecretKey:   "test-secret",
		RateLimitRPS:   10000,
		RateLimitBurst: 10000,
		AllowedOrigins: []string{"*"},
	}
	log, err := logger.NewLogger("debug", "test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	services := &service.Services{
		Auth:       nil,
		Attendance: nil,
		Leave:      nil,
		Shift:      nil,
		User:       nil,
		Department: nil,
		Dashboard:  nil,
	}
	handlers := handler.NewHandlers(services, log)
	Setup(r, handlers, middleware.NewMiddleware(cfg, log))

	return r
}

func makeBearerToken(t *testing.T, secret, userID, role string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return "Bearer " + signed
}

func TestSetup_RouteAccessBranches(t *testing.T) {
	r := setupRouterForBranchTest(t)

	t.Run("public endpoint is accessible", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("auth endpoint does not require auth middleware", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		r.ServeHTTP(w, req)

		// Handler bind error is expected; key point is not blocked by Auth middleware.
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("protected endpoint requires auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("admin endpoint rejects non-admin role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/leaves/not-a-uuid/approve", nil)
		req.Header.Set("Authorization", makeBearerToken(t, "test-secret", "user-1", "employee"))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("admin endpoint allows admin role and reaches handler", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/leaves/not-a-uuid/approve", nil)
		req.Header.Set("Authorization", makeBearerToken(t, "test-secret", "admin-1", "admin"))
		r.ServeHTTP(w, req)

		// Invalid UUID response means RequireRole was passed and handler was executed.
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
