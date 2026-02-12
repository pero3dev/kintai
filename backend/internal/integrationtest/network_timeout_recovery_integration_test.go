package integrationtest

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestNetworkPartitionAndTimeoutRecovery(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-network-recovery-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	authHeader := env.MustBearerToken(t, user.ID, model.RoleEmployee)

	t.Run("request timeout recovers and subsequent API calls succeed", func(t *testing.T) {
		releaseSlowHandler := make(chan struct{})
		env.Router.GET("/_it/network/slow", func(c *gin.Context) {
			<-releaseSlowHandler
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		server := httptest.NewServer(env.Router)
		defer server.Close()

		timeoutClient := &http.Client{Timeout: 150 * time.Millisecond}
		_, err := requestStatus(timeoutClient, http.MethodGet, server.URL+"/_it/network/slow", nil)
		require.Error(t, err, "expected timeout during simulated network delay")

		var netErr net.Error
		require.True(t, errors.As(err, &netErr), "expected net.Error, got %T", err)
		require.True(t, netErr.Timeout(), "expected timeout error, got %v", err)

		close(releaseSlowHandler)

		recovered := false
		for attempt := 0; attempt < 10; attempt++ {
			status, reqErr := requestStatus(http.DefaultClient, http.MethodGet, server.URL+"/api/v1/users/me", map[string]string{
				"Authorization": authHeader,
			})
			if reqErr == nil && status == http.StatusOK {
				recovered = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		require.True(t, recovered, "API did not recover after timeout event")
	})

	t.Run("connection outage recovers after server restart", func(t *testing.T) {
		server := httptest.NewServer(env.Router)

		baseline, err := requestStatus(http.DefaultClient, http.MethodGet, server.URL+"/health", nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, baseline)

		server.Close()

		_, err = requestStatus(http.DefaultClient, http.MethodGet, server.URL+"/health", nil)
		require.Error(t, err, "expected connection error after simulated network partition")

		recoveredServer := httptest.NewServer(env.Router)
		defer recoveredServer.Close()

		recoveredStatus, recoveredErr := requestStatus(http.DefaultClient, http.MethodGet, recoveredServer.URL+"/health", nil)
		require.NoError(t, recoveredErr)
		require.Equal(t, http.StatusOK, recoveredStatus)

		meStatus, meErr := requestStatus(http.DefaultClient, http.MethodGet, recoveredServer.URL+"/api/v1/users/me", map[string]string{
			"Authorization": authHeader,
		})
		require.NoError(t, meErr)
		require.Equal(t, http.StatusOK, meStatus)
	})
}

func requestStatus(client *http.Client, method, rawURL string, headers map[string]string) (int, error) {
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		return 0, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}
