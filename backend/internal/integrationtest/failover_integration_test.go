package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestFailoverAcrossMultipleInstances(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-failover-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	primaryDB := env.DB
	primaryRouter := env.Router

	secondaryDB, err := openDB(env.Config.DatabaseURL)
	require.NoError(t, err)
	secondaryRouter := buildRouter(secondaryDB, env.Config, env.Logger)
	t.Cleanup(func() {
		if sqlDB, err := secondaryDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	primaryCode, primaryUserID := callUsersMeFromRouter(t, primaryRouter, headers)
	require.Equal(t, http.StatusOK, primaryCode)
	require.Equal(t, user.ID, primaryUserID)

	secondaryCode, secondaryUserID := callUsersMeFromRouter(t, secondaryRouter, headers)
	require.Equal(t, http.StatusOK, secondaryCode)
	require.Equal(t, user.ID, secondaryUserID)

	firstCode, firstUsed, firstPrimaryStatus := callUsersMeWithFailover(t, headers, primaryRouter, secondaryRouter)
	require.Equal(t, http.StatusOK, firstCode)
	require.Equal(t, "primary", firstUsed)
	require.Equal(t, http.StatusOK, firstPrimaryStatus)

	primarySQLDB, err := primaryDB.DB()
	require.NoError(t, err)
	require.NoError(t, primarySQLDB.Close())

	failoverCode, failoverUsed, primaryAfterFailure := callUsersMeWithFailover(t, headers, primaryRouter, secondaryRouter)
	require.Equal(t, http.StatusOK, failoverCode)
	require.Equal(t, "secondary", failoverUsed)
	require.NotEqual(t, http.StatusOK, primaryAfterFailure)

	for i := 0; i < 5; i++ {
		code, used, failedPrimaryStatus := callUsersMeWithFailover(t, headers, primaryRouter, secondaryRouter)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "secondary", used)
		require.NotEqual(t, http.StatusOK, failedPrimaryStatus)
	}
}

func callUsersMeWithFailover(
	t testing.TB,
	headers map[string]string,
	primary *gin.Engine,
	secondary *gin.Engine,
) (status int, usedInstance string, primaryStatus int) {
	t.Helper()

	primaryResp := doRouterJSON(t, primary, http.MethodGet, "/api/v1/users/me", nil, headers)
	if primaryResp.Code == http.StatusOK {
		return primaryResp.Code, "primary", primaryResp.Code
	}

	secondaryResp := doRouterJSON(t, secondary, http.MethodGet, "/api/v1/users/me", nil, headers)
	return secondaryResp.Code, "secondary", primaryResp.Code
}

func callUsersMeFromRouter(t testing.TB, router *gin.Engine, headers map[string]string) (int, uuid.UUID) {
	t.Helper()

	resp := doRouterJSON(t, router, http.MethodGet, "/api/v1/users/me", nil, headers)
	if resp.Code != http.StatusOK {
		return resp.Code, uuid.Nil
	}

	var me model.User
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &me))
	return resp.Code, me.ID
}

func doRouterJSON(
	t testing.TB,
	router *gin.Engine,
	method string,
	path string,
	body any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	req, err := JSONRequest(method, path, body, headers)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
