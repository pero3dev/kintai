package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestCommonEndpointQualityAxes(t *testing.T) {
	t.Run("CA-05 validation error returns 4xx with ErrorResponse", func(t *testing.T) {
		env := NewTestEnv(t, nil)
		require.NoError(t, env.ResetDB())

		adminHeaders := map[string]string{
			"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleAdmin),
		}

		type tc struct {
			name    string
			method  string
			path    string
			body    any
			headers map[string]string
		}

		cases := []tc{
			{
				name:   "login invalid body type",
				method: http.MethodPost,
				path:   "/api/v1/auth/login",
				body:   "invalid",
			},
			{
				name:   "refresh invalid body type",
				method: http.MethodPost,
				path:   "/api/v1/auth/refresh",
				body:   "invalid",
			},
			{
				name:    "users update invalid path param",
				method:  http.MethodPut,
				path:    "/api/v1/users/not-a-uuid",
				body:    map[string]any{"first_name": "X"},
				headers: adminHeaders,
			},
		}

		for _, c := range cases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				resp := env.DoJSON(t, c.method, c.path, c.body, c.headers)
				require.GreaterOrEqual(t, resp.Code, 400)
				require.Less(t, resp.Code, 500)

				var errResp model.ErrorResponse
				require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &errResp))
				require.NotZero(t, errResp.Code)
				require.NotEmpty(t, errResp.Message)
			})
		}
	})

	t.Run("CA-06 happy path returns expected status and response shape", func(t *testing.T) {
		env := NewTestEnv(t, nil)
		require.NoError(t, env.ResetDB())

		healthResp := env.DoJSON(t, http.MethodGet, "/health", nil, nil)
		require.Equal(t, http.StatusOK, healthResp.Code)
		var health map[string]any
		require.NoError(t, json.Unmarshal(healthResp.Body.Bytes(), &health))
		require.Equal(t, "ok", health["status"])
		require.NotEmpty(t, health["service"])

		metricsResp := env.DoJSON(t, http.MethodGet, "/metrics", nil, nil)
		require.Equal(t, http.StatusOK, metricsResp.Code)
		require.Contains(t, metricsResp.Header().Get("Content-Type"), "text/plain")

		email := fmt.Sprintf("it-ca06-%s@example.com", uuid.NewString())
		password := "password123"

		registerResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]any{
			"email":      email,
			"password":   password,
			"first_name": "Happy",
			"last_name":  "Path",
		}, nil)
		require.Equal(t, http.StatusCreated, registerResp.Code)

		var registered model.User
		require.NoError(t, json.Unmarshal(registerResp.Body.Bytes(), &registered))
		require.NotEqual(t, uuid.Nil, registered.ID)
		require.Equal(t, email, registered.Email)

		loginResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"email":    email,
			"password": password,
		}, nil)
		require.Equal(t, http.StatusOK, loginResp.Code)

		var token model.TokenResponse
		require.NoError(t, json.Unmarshal(loginResp.Body.Bytes(), &token))
		require.NotEmpty(t, token.AccessToken)
		require.NotEmpty(t, token.RefreshToken)
	})

	t.Run("CA-07 write endpoint reflects db mutation", func(t *testing.T) {
		env := NewTestEnv(t, nil)
		require.NoError(t, env.ResetDB())

		adminHeaders := map[string]string{
			"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleAdmin),
		}

		var beforeCount int64
		require.NoError(t, env.DB.Model(&model.User{}).Count(&beforeCount).Error)

		createEmail := fmt.Sprintf("it-ca07-%s@example.com", uuid.NewString())
		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/users", map[string]any{
			"email":      createEmail,
			"password":   "password123",
			"first_name": "Before",
			"last_name":  "Mutation",
			"role":       "employee",
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var created model.User
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
		require.NotEqual(t, uuid.Nil, created.ID)

		var afterCreateCount int64
		require.NoError(t, env.DB.Model(&model.User{}).Count(&afterCreateCount).Error)
		require.Equal(t, beforeCount+1, afterCreateCount)

		updateResp := env.DoJSON(t, http.MethodPut, "/api/v1/users/"+created.ID.String(), map[string]any{
			"first_name": "After",
		}, adminHeaders)
		require.Equal(t, http.StatusOK, updateResp.Code)

		var updated model.User
		require.NoError(t, env.DB.Where("id = ?", created.ID).First(&updated).Error)
		require.Equal(t, "After", updated.FirstName)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/users/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)

		var deleted model.User
		err := env.DB.Unscoped().Where("id = ?", created.ID).First(&deleted).Error
		require.NoError(t, err)
		require.NotNil(t, deleted.DeletedAt)
		require.NotEqual(t, time.Time{}, deleted.DeletedAt.Time)
	})
}
