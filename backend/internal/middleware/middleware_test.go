package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

func setupTestMiddleware(t *testing.T) *Middleware {
	cfg := &config.Config{
		AllowedOrigins: []string{"http://localhost:5173", "http://example.com"},
		JWTSecretKey:   "test-secret-key",
		RateLimitRPS:   100,
		RateLimitBurst: 200,
	}
	log, err := logger.NewLogger("debug", "test")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return NewMiddleware(cfg, log)
}

func TestNewMiddleware(t *testing.T) {
	m := setupTestMiddleware(t)
	if m == nil {
		t.Error("Expected middleware to be non-nil")
	}
}

// ===== CORS Tests =====

func TestCORS_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:5173")

	handler := m.CORS()
	handler(c)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("Expected CORS origin to be 'http://localhost:5173', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Origin", "http://malicious.com")

	handler := m.CORS()
	handler(c)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no CORS origin for disallowed domain, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("OPTIONS", "/test", nil)
	c.Request.Header.Set("Origin", "http://localhost:5173")

	handler := m.CORS()
	handler(c)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusNoContent, w.Code)
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		AllowedOrigins: []string{"*"},
		JWTSecretKey:   "test-secret-key",
		RateLimitRPS:   100,
		RateLimitBurst: 200,
	}
	log, _ := logger.NewLogger("debug", "test")
	m := NewMiddleware(cfg, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Origin", "http://anysite.com")

	handler := m.CORS()
	handler(c)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://anysite.com" {
		t.Errorf("Expected wildcard CORS to work, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// ===== Auth Tests =====

func TestAuth_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	handler := m.Auth()
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuth_InvalidBearerFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "InvalidToken")

	handler := m.Auth()
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	handler := m.Auth()
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	// 有効なトークンを生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "test-user-id",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	handler := m.Auth()
	handler(c)

	// トークンが正常に処理されたか確認
	userID, exists := c.Get("userID")
	if !exists {
		t.Error("Expected userID to be set in context")
	}
	if userID != "test-user-id" {
		t.Errorf("Expected userID 'test-user-id', got '%s'", userID)
	}

	role, _ := c.Get("role")
	if role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", role)
	}
}

func TestAuth_WrongSigningMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	// 異なる署名方式のトークンを作成 (none method)
	token := jwt.New(jwt.SigningMethodNone)
	token.Claims = jwt.MapClaims{
		"sub":  "test-user-id",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	handler := m.Auth()
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for wrong signing method, got %d", http.StatusUnauthorized, w.Code)
	}
}

// ===== RequireRole Tests =====

func TestRequireRole_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Set("role", "admin")

	handler := m.RequireRole(model.RoleAdmin, model.RoleManager)
	handler(c)

	if c.IsAborted() {
		t.Error("Request should not be aborted for valid role")
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Set("role", "employee")

	handler := m.RequireRole(model.RoleAdmin)
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRequireRole_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	// roleを設定しない

	handler := m.RequireRole(model.RoleAdmin)
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

// ===== RateLimit Tests =====

func TestRateLimit_AllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "127.0.0.1:1234"

	handler := m.RateLimit()
	handler(c)

	if c.IsAborted() {
		t.Error("First request should not be rate limited")
	}
}

// ===== SecurityHeaders Tests =====

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	handler := m.SecurityHeaders()
	handler(c)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for header, expected := range expectedHeaders {
		if w.Header().Get(header) != expected {
			t.Errorf("Expected %s to be '%s', got '%s'", header, expected, w.Header().Get(header))
		}
	}
}

// ===== RequestLogger Tests =====

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("User-Agent", "test-agent")

	handler := m.RequestLogger()
	// Should not panic
	handler(c)
}

// ===== CSRF Tests =====

func TestCSRF_GetRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	handler := m.CSRF()
	handler(c)

	if c.IsAborted() {
		t.Error("GET request should not be CSRF checked")
	}
}

func TestCSRF_HeadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("HEAD", "/test", nil)

	handler := m.CSRF()
	handler(c)

	if c.IsAborted() {
		t.Error("HEAD request should not be CSRF checked")
	}
}

func TestCSRF_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("OPTIONS", "/test", nil)

	handler := m.CSRF()
	handler(c)

	if c.IsAborted() {
		t.Error("OPTIONS request should not be CSRF checked")
	}
}

func TestCSRF_PostWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/test", nil)

	handler := m.CSRF()
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestCSRF_PostWithMismatchedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/test", nil)
	c.Request.Header.Set("X-CSRF-Token", "token1")
	c.Request.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token2"})

	handler := m.CSRF()
	handler(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for mismatched tokens, got %d", http.StatusForbidden, w.Code)
	}
}

func TestCSRF_PostWithMatchingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/test", nil)
	c.Request.Header.Set("X-CSRF-Token", "valid-token")
	c.Request.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})

	handler := m.CSRF()
	handler(c)

	if c.IsAborted() {
		t.Error("Request with matching CSRF token should not be aborted")
	}
}

// ===== Recovery Tests =====

func TestRecovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	handler := m.Recovery()
	handler(c)

	if c.IsAborted() {
		t.Error("Normal request should not be aborted")
	}
}

func TestRecovery_WithPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	router := gin.New()
	router.Use(m.Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d after panic, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Use a very low rate limit for testing
	cfg := &config.Config{
		JWTSecretKey:   "test-secret",
		RateLimitRPS:   1,     // 1 request per second
		RateLimitBurst: 1,     // burst of 1
		AllowedOrigins: []string{"*"},
	}
	log, _ := logger.NewLogger("debug", "test")
	m := NewMiddleware(cfg, log)

	router := gin.New()
	router.Use(m.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// First request from unique IP should pass
	unique := "192.168.99.99:9999"
	
	// Make multiple quick requests to exceed burst
	var lastCode int
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = unique
		router.ServeHTTP(w, req)
		lastCode = w.Code
		if w.Code == http.StatusTooManyRequests {
			break // Expected
		}
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("Expected status %d after exceeding rate limit, got %d", http.StatusTooManyRequests, lastCode)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	// Create an expired token
	claims := jwt.MapClaims{
		"sub":  "user-123",
		"role": "employee",
		"exp":  time.Now().Add(-1 * time.Hour).Unix(), // expired
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	handler := m.Auth()
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
