package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// createMockServices creates mock services for testing
func createMockServices() *service.Services {
	return &service.Services{
		Auth:       &mocks.MockAuthService{},
		Attendance: &mocks.MockAttendanceService{},
		Leave:      &mocks.MockLeaveService{},
		Shift:      &mocks.MockShiftService{},
		User:       &mocks.MockUserService{},
		Department: &mocks.MockDepartmentService{},
		Dashboard:  &mocks.MockDashboardService{},
	}
}

func TestSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		JWTSecretKey:   "test-secret",
		RateLimitRPS:   100,
		RateLimitBurst: 10,
		AllowedOrigins: []string{"*"},
	}
	log, _ := logger.NewLogger("debug", "test")
	mw := middleware.NewMiddleware(cfg, log)
	services := createMockServices()
	handlers := handler.NewHandlers(services, log)

	// Should not panic
	Setup(r, handlers, mw)

	// Verify some routes are registered
	routes := r.Routes()
	if len(routes) == 0 {
		t.Error("Expected routes to be registered")
	}

	// Check health endpoint exists
	hasHealthRoute := false
	for _, route := range routes {
		if route.Path == "/health" && route.Method == "GET" {
			hasHealthRoute = true
			break
		}
	}
	if !hasHealthRoute {
		t.Error("Expected /health route to be registered")
	}
}

func TestSetup_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		JWTSecretKey:   "test-secret",
		RateLimitRPS:   100,
		RateLimitBurst: 10,
		AllowedOrigins: []string{"*"},
	}
	log, _ := logger.NewLogger("debug", "test")
	mw := middleware.NewMiddleware(cfg, log)
	services := createMockServices()
	handlers := handler.NewHandlers(services, log)

	Setup(r, handlers, mw)

	// Test health endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSetup_RegisteredRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		JWTSecretKey:   "test-secret",
		RateLimitRPS:   100,
		RateLimitBurst: 10,
		AllowedOrigins: []string{"*"},
	}
	log, _ := logger.NewLogger("debug", "test")
	mw := middleware.NewMiddleware(cfg, log)
	services := createMockServices()
	handlers := handler.NewHandlers(services, log)

	Setup(r, handlers, mw)

	// Expected routes
	expectedRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/metrics"},
		{"POST", "/api/v1/auth/login"},
		{"POST", "/api/v1/auth/register"},
		{"POST", "/api/v1/auth/refresh"},
		{"POST", "/api/v1/auth/logout"},
		{"GET", "/api/v1/attendance"},
		{"POST", "/api/v1/attendance/clock-in"},
		{"POST", "/api/v1/attendance/clock-out"},
		{"GET", "/api/v1/leaves"},
		{"POST", "/api/v1/leaves"},
		{"GET", "/api/v1/users/me"},
		{"GET", "/api/v1/departments"},
		{"GET", "/api/v1/shifts"},
		{"GET", "/api/v1/dashboard/stats"},
	}

	routes := r.Routes()
	routeMap := make(map[string]bool)
	for _, route := range routes {
		key := route.Method + ":" + route.Path
		routeMap[key] = true
	}

	for _, expected := range expectedRoutes {
		key := expected.method + ":" + expected.path
		if !routeMap[key] {
			t.Errorf("Expected route %s %s to be registered", expected.method, expected.path)
		}
	}
}
