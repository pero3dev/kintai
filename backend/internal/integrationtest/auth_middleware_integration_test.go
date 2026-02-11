package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestAuthAndMiddlewareIntegration(t *testing.T) {
	t.Run("register_login_refresh_logout", func(t *testing.T) {
		env := NewTestEnv(t, nil)
		require.NoError(t, env.ResetDB())

		email := fmt.Sprintf("it-auth-%s@example.com", uuid.NewString())
		password := "password123"

		registerResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]any{
			"email":      email,
			"password":   password,
			"first_name": "IT",
			"last_name":  "Auth",
		}, nil)
		require.Equal(t, http.StatusCreated, registerResp.Code)

		loginResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"email":    email,
			"password": password,
		}, nil)
		require.Equal(t, http.StatusOK, loginResp.Code)

		var loginToken model.TokenResponse
		require.NoError(t, json.Unmarshal(loginResp.Body.Bytes(), &loginToken))
		require.NotEmpty(t, loginToken.AccessToken)
		require.NotEmpty(t, loginToken.RefreshToken)

		refreshResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": loginToken.RefreshToken,
		}, nil)
		require.Equal(t, http.StatusOK, refreshResp.Code)

		var refreshed model.TokenResponse
		require.NoError(t, json.Unmarshal(refreshResp.Body.Bytes(), &refreshed))
		require.NotEmpty(t, refreshed.AccessToken)
		require.NotEmpty(t, refreshed.RefreshToken)

		logoutResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": "Bearer " + loginToken.AccessToken,
		})
		require.Equal(t, http.StatusNoContent, logoutResp.Code)
	})

	t.Run("jwt_error_cases", func(t *testing.T) {
		env := NewTestEnv(t, nil)
		require.NoError(t, env.ResetDB())

		user := createTestUser(t, env, model.RoleEmployee, "it-jwt@example.com", "password123")

		expiredResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": env.MustExpiredBearerToken(t, user.ID, model.RoleEmployee),
		})
		require.Equal(t, http.StatusUnauthorized, expiredResp.Code)

		invalidSigResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": env.MustInvalidSignatureBearerToken(t, user.ID, model.RoleEmployee),
		})
		require.Equal(t, http.StatusUnauthorized, invalidSigResp.Code)

		malformedBearerResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": "Token malformed",
		})
		require.Equal(t, http.StatusUnauthorized, malformedBearerResp.Code)
	})

	t.Run("cors_allowed_disallowed_and_options", func(t *testing.T) {
		env := NewTestEnv(t, &Options{
			AllowedOrigins:               []string{"http://allowed.example.com"},
			DisableAdditionalAutoMigrate: true,
		})

		allowedResp := env.DoJSON(t, http.MethodGet, "/health", nil, map[string]string{
			"Origin": "http://allowed.example.com",
		})
		require.Equal(t, http.StatusOK, allowedResp.Code)
		require.Equal(t, "http://allowed.example.com", allowedResp.Header().Get("Access-Control-Allow-Origin"))

		disallowedResp := env.DoJSON(t, http.MethodGet, "/health", nil, map[string]string{
			"Origin": "http://blocked.example.com",
		})
		require.Equal(t, http.StatusOK, disallowedResp.Code)
		require.Empty(t, disallowedResp.Header().Get("Access-Control-Allow-Origin"))

		optionsResp := env.DoJSON(t, http.MethodOptions, "/api/v1/auth/logout", nil, map[string]string{
			"Origin":                        "http://allowed.example.com",
			"Access-Control-Request-Method": http.MethodPost,
		})
		require.Equal(t, http.StatusNoContent, optionsResp.Code)
		require.Contains(t, optionsResp.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
	})

	t.Run("security_headers", func(t *testing.T) {
		env := NewTestEnv(t, nil)

		resp := env.DoJSON(t, http.MethodGet, "/health", nil, nil)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, "nosniff", resp.Header().Get("X-Content-Type-Options"))
		require.Equal(t, "DENY", resp.Header().Get("X-Frame-Options"))
		require.Equal(t, "1; mode=block", resp.Header().Get("X-XSS-Protection"))
		require.Equal(t, "strict-origin-when-cross-origin", resp.Header().Get("Referrer-Policy"))
		require.Equal(t, "default-src 'self'", resp.Header().Get("Content-Security-Policy"))
		require.Contains(t, resp.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	})

	t.Run("rate_limit_returns_429", func(t *testing.T) {
		env := NewTestEnv(t, &Options{
			RateLimitRPS:                 1,
			RateLimitBurst:               1,
			DisableAdditionalAutoMigrate: true,
		})

		remoteAddr := uniqueRemoteAddr()
		got429 := false
		for i := 0; i < 20; i++ {
			req, err := JSONRequest(http.MethodGet, "/health", nil, nil)
			require.NoError(t, err)
			req.RemoteAddr = remoteAddr

			resp := env.DoRequest(req)
			if resp.Code == http.StatusTooManyRequests {
				got429 = true
				break
			}
		}
		require.True(t, got429, "expected at least one 429 from rate limiter")
	})

	t.Run("recovery_returns_500_on_panic", func(t *testing.T) {
		env := NewTestEnv(t, nil)
		env.Router.GET("/_it/panic", func(c *gin.Context) {
			panic("it panic")
		})

		resp := env.DoJSON(t, http.MethodGet, "/_it/panic", nil, nil)
		require.Equal(t, http.StatusInternalServerError, resp.Code)

		var errResp model.ErrorResponse
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &errResp))
		require.Equal(t, 500, errResp.Code)
		require.NotEmpty(t, errResp.Message)
	})
}

func uniqueRemoteAddr() string {
	u := uuid.New()
	return fmt.Sprintf("%d.%d.%d.%d:12345", u[0], u[1], u[2], u[3])
}
