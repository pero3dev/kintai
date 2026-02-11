package integrationtest

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func TestFoundation(t *testing.T) {
	env := NewTestEnv(t, nil)

	t.Run("ResetDB truncates all tables", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		user := createTestUser(t, env, model.RoleEmployee, "reset-user@example.com", "password123")
		require.NotEqual(t, uuid.Nil, user.ID)

		var count int64
		require.NoError(t, env.DB.Model(&model.User{}).Count(&count).Error)
		require.Equal(t, int64(1), count)

		require.NoError(t, env.ResetDB())
		require.NoError(t, env.DB.Model(&model.User{}).Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("JWT helper and JSON helper work with protected route", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		user := createTestUser(t, env, model.RoleEmployee, "login-user@example.com", "password123")

		loginResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
			"email":    "login-user@example.com",
			"password": "password123",
		}, nil)
		require.Equal(t, http.StatusOK, loginResp.Code)

		var tokenResp model.TokenResponse
		require.NoError(t, json.Unmarshal(loginResp.Body.Bytes(), &tokenResp))
		require.NotEmpty(t, tokenResp.AccessToken)

		logoutResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
		})
		require.Equal(t, http.StatusNoContent, logoutResp.Code)

		expiredResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
			"Authorization": env.MustExpiredBearerToken(t, user.ID, model.RoleEmployee),
		})
		require.Equal(t, http.StatusUnauthorized, expiredResp.Code)
	})

	t.Run("Multipart and download helpers", func(t *testing.T) {
		env.Router.POST("/_it/multipart", func(c *gin.Context) {
			file, err := c.FormFile("file")
			if err != nil {
				c.String(http.StatusBadRequest, "file missing")
				return
			}
			title := c.PostForm("title")
			c.JSON(http.StatusOK, gin.H{
				"title":    title,
				"filename": file.Filename,
			})
		})
		env.Router.GET("/_it/download", func(c *gin.Context) {
			c.Header("Content-Type", "text/plain")
			c.String(http.StatusOK, "download-ok")
		})

		multipartResp := env.DoMultipart(
			t,
			http.MethodPost,
			"/_it/multipart",
			map[string]string{"title": "receipt"},
			map[string]MultipartFile{
				"file": {
					FileName: "receipt.txt",
					Content:  []byte("receipt content"),
				},
			},
			nil,
		)
		require.Equal(t, http.StatusOK, multipartResp.Code)
		require.Contains(t, multipartResp.Body.String(), "receipt")
		require.Contains(t, multipartResp.Body.String(), "receipt.txt")

		body, headers, code := env.DoDownload(t, http.MethodGet, "/_it/download", nil)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "download-ok", string(body))
		require.Contains(t, headers.Get("Content-Type"), "text/plain")
	})
}

func createTestUser(t *testing.T, env *TestEnv, role model.Role, email, password string) *model.User {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &model.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    "Integration",
		LastName:     "Tester",
		Role:         role,
		IsActive:     true,
	}
	require.NoError(t, env.DB.Create(user).Error)
	return user
}
