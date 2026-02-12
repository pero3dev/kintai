package integrationtest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestAppRestartContinuityWithInflightRequest(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-restart-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	// Old instance route that intentionally keeps one request in-flight.
	oldInstanceDB := env.DB
	inflightStarted := make(chan struct{})
	releaseInflight := make(chan struct{})

	env.Router.GET("/_it/restart/inflight", func(c *gin.Context) {
		close(inflightStarted)
		<-releaseInflight

		var userCount int64
		if err := oldInstanceDB.Model(&model.User{}).Count(&userCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_count": userCount})
	})

	inflightReq, err := JSONRequest(http.MethodGet, "/_it/restart/inflight", nil, nil)
	require.NoError(t, err)

	inflightDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		inflightDone <- env.DoRequest(inflightReq)
	}()

	select {
	case <-inflightStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for in-flight request to start")
	}

	oldSQLDB, err := oldInstanceDB.DB()
	require.NoError(t, err)

	// Simulate restart: create new DB session + new router and switch active instance.
	newDB, err := openDB(env.Config.DatabaseURL)
	require.NoError(t, err)
	newRouter := buildRouter(newDB, env.Config, env.Logger)

	env.DB = newDB
	env.Router = newRouter

	healthResp := env.DoJSON(t, http.MethodGet, "/health", nil, nil)
	require.Equal(t, http.StatusOK, healthResp.Code)

	meResp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, headers)
	require.Equal(t, http.StatusOK, meResp.Code)

	close(releaseInflight)

	var inflightResp *httptest.ResponseRecorder
	select {
	case inflightResp = <-inflightDone:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for in-flight request completion")
	}

	require.Equal(t, http.StatusOK, inflightResp.Code)
	require.Contains(t, inflightResp.Body.String(), "user_count")

	// Old instance shutdown should not break the restarted instance.
	require.NoError(t, oldSQLDB.Close())

	postShutdownResp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, headers)
	require.Equal(t, http.StatusOK, postShutdownResp.Code)
}
