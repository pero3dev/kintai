package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Liveness(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/live", nil)

	h.Liveness(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "alive", body["status"])
}

func TestHealthHandler_Readiness_Ready(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandlerWithChecks(func(ctx context.Context) error { return nil })
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/ready", nil)

	h.Readiness(c)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "ready", body["status"])
}

func TestHealthHandler_Readiness_NotReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHealthHandlerWithChecks(func(ctx context.Context) error {
		return errors.New("db unavailable")
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/ready", nil)

	h.Readiness(c)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, "not_ready", body["status"])
}
