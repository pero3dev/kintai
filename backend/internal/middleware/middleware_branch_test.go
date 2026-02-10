package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

func TestAuth_ClaimsTypeMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := setupTestMiddleware(t)

	originalJWTParse := jwtParse
	t.Cleanup(func() {
		jwtParse = originalJWTParse
	})

	jwtParse = func(_ string, _ jwt.Keyfunc, _ ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{
			Valid:  true,
			Claims: jwt.RegisteredClaims{},
		}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer any-token")

	m.Auth()(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCleanupStaleClients(t *testing.T) {
	now := time.Now()

	mu.Lock()
	clients = map[string]*clientLimiter{
		"stale": {
			limiter:  rate.NewLimiter(rate.Limit(1), 1),
			lastSeen: now.Add(-4 * time.Minute),
		},
		"active": {
			limiter:  rate.NewLimiter(rate.Limit(1), 1),
			lastSeen: now.Add(-2 * time.Minute),
		},
	}
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		clients = make(map[string]*clientLimiter)
		mu.Unlock()
	})

	cleanupStaleClients(now)

	mu.Lock()
	_, staleExists := clients["stale"]
	_, activeExists := clients["active"]
	mu.Unlock()

	if staleExists {
		t.Fatal("stale client should be removed")
	}
	if !activeExists {
		t.Fatal("active client should remain")
	}
}
