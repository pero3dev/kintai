package integrationtest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadinessLivenessBehaviorWhenDBDependencyIsDown(t *testing.T) {
	env := NewTestEnv(t, nil)

	readyBefore := env.DoJSON(t, http.MethodGet, "/ready", nil, nil)
	require.Equal(t, http.StatusOK, readyBefore.Code)

	liveBefore := env.DoJSON(t, http.MethodGet, "/live", nil, nil)
	require.Equal(t, http.StatusOK, liveBefore.Code)

	sqlDB, err := env.DB.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	readyAfter := env.DoJSON(t, http.MethodGet, "/ready", nil, nil)
	require.Equal(t, http.StatusServiceUnavailable, readyAfter.Code)
	require.Contains(t, readyAfter.Body.String(), "not_ready")

	liveAfter := env.DoJSON(t, http.MethodGet, "/live", nil, nil)
	require.Equal(t, http.StatusOK, liveAfter.Code)
	require.Contains(t, liveAfter.Body.String(), "alive")
}
