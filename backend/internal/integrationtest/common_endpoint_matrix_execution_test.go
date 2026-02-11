package integrationtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestCommonEndpointMatrixExecution(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, model.HRAutoMigrate(env.DB))
	require.NoError(t, model.ExpenseAutoMigrate(env.DB))
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	allEndpoints := collectRegisteredEndpoints(env)
	publicSet := toEndpointSet(publicEndpointDefs)
	roleSet := toEndpointSet(roleEndpointDefs)

	loginEmail := fmt.Sprintf("it-matrix-%s@example.com", uuid.NewString())
	loginPassword := "password123"
	createTestUser(t, env, model.RoleEmployee, loginEmail, loginPassword)

	loginResp := env.DoJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email":    loginEmail,
		"password": loginPassword,
	}, nil)
	require.Equal(t, http.StatusOK, loginResp.Code)
	var loginToken model.TokenResponse
	require.NoError(t, json.Unmarshal(loginResp.Body.Bytes(), &loginToken))

	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleEmployee),
	}
	adminHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleAdmin),
	}

	t.Run("CA-06 all endpoints are reachable with proper auth context", func(t *testing.T) {
		for _, ep := range allEndpoints {
			key := endpointKey(ep.Method, ep.Path)
			headers := chooseHeadersForEndpoint(key, publicSet, roleSet, employeeHeaders, adminHeaders)

			resp := doHappyPathLikeRequest(t, env, ep, headers, happyPathContext{
				loginEmail:    loginEmail,
				loginPassword: loginPassword,
				refreshToken:  loginToken.RefreshToken,
			})
			require.NotEqualf(t, http.StatusUnauthorized, resp.Code, "%s", key)
			require.NotEqualf(t, http.StatusForbidden, resp.Code, "%s", key)
		}
	})

	t.Run("CA-05 validatable endpoints return 4xx on invalid input", func(t *testing.T) {
		for _, ep := range allEndpoints {
			if !isValidationTarget(ep) {
				continue
			}

			key := endpointKey(ep.Method, ep.Path)
			headers := chooseHeadersForEndpoint(key, publicSet, roleSet, employeeHeaders, adminHeaders)
			resp := doInvalidInputRequest(t, env, ep, headers)

			require.GreaterOrEqualf(t, resp.Code, 400, "%s", key)
			require.Lessf(t, resp.Code, 500, "%s", key)

			if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
				var errResp model.ErrorResponse
				require.NoErrorf(t, json.Unmarshal(resp.Body, &errResp), "%s", key)
				require.NotZerof(t, errResp.Code, "%s", key)
				require.NotEmptyf(t, errResp.Message, "%s", key)
			}
		}
	})
}

func chooseHeadersForEndpoint(
	key string,
	publicSet map[string]endpointDef,
	roleSet map[string]endpointDef,
	employeeHeaders map[string]string,
	adminHeaders map[string]string,
) map[string]string {
	if _, ok := publicSet[key]; ok {
		return nil
	}
	if _, ok := roleSet[key]; ok {
		return adminHeaders
	}
	return employeeHeaders
}

type happyPathContext struct {
	loginEmail    string
	loginPassword string
	refreshToken  string
}

func doHappyPathLikeRequest(
	t testing.TB,
	env *TestEnv,
	ep endpointDef,
	headers map[string]string,
	ctx happyPathContext,
) *httpResponse {
	t.Helper()

	path := materializeEndpointPath(ep.Path)

	switch ep.Path {
	case "/api/v1/auth/login":
		rec := env.DoJSON(t, http.MethodPost, path, map[string]any{
			"email":    ctx.loginEmail,
			"password": ctx.loginPassword,
		}, headers)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	case "/api/v1/auth/register":
		rec := env.DoJSON(t, http.MethodPost, path, map[string]any{
			"email":      fmt.Sprintf("it-matrix-register-%s@example.com", uuid.NewString()),
			"password":   "password123",
			"first_name": "Matrix",
			"last_name":  "Register",
		}, headers)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	case "/api/v1/auth/refresh":
		rec := env.DoJSON(t, http.MethodPost, path, map[string]any{
			"refresh_token": ctx.refreshToken,
		}, headers)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	}

	switch {
	case ep.Method == http.MethodPost && ep.Path == "/api/v1/expenses/receipts/upload":
		rec := env.DoMultipart(
			t,
			http.MethodPost,
			path,
			map[string]string{"title": "it-receipt"},
			map[string]MultipartFile{
				"file": {
					FileName: "receipt.txt",
					Content:  []byte("receipt"),
				},
			},
			headers,
		)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	case ep.Method == http.MethodPost || ep.Method == http.MethodPut || ep.Method == http.MethodPatch:
		rec := env.DoJSON(t, ep.Method, path, map[string]any{}, headers)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	default:
		rec := env.DoJSON(t, ep.Method, path, nil, headers)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	}
}

func doInvalidInputRequest(t testing.TB, env *TestEnv, ep endpointDef, headers map[string]string) *httpResponse {
	t.Helper()

	path := materializeEndpointPath(ep.Path)
	if strings.Contains(ep.Path, ":") {
		path = invalidizeEndpointPath(ep.Path)
		rec := env.DoJSON(t, ep.Method, path, nil, headers)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	}

	if ep.Method == http.MethodPost || ep.Method == http.MethodPut || ep.Method == http.MethodPatch {
		req, err := http.NewRequest(ep.Method, path, bytes.NewBufferString("{invalid-json"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rec := env.DoRequest(req)
		return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
	}

	rec := env.DoJSON(t, ep.Method, path+"?start_date=invalid&end_date=invalid", nil, headers)
	return &httpResponse{Code: rec.Code, Body: rec.Body.Bytes(), Header: rec.Header()}
}

func invalidizeEndpointPath(path string) string {
	return pathParamPattern.ReplaceAllStringFunc(path, func(raw string) string {
		name := strings.ToLower(strings.TrimPrefix(raw, ":"))
		switch name {
		case "leave_type":
			return "invalid-leave-type"
		case "taskid", "actionid":
			return "invalid-id"
		case "itemkey":
			return "invalid-item"
		default:
			return "not-a-uuid"
		}
	})
}

func isValidationTarget(ep endpointDef) bool {
	if ep.Path == "/health" || ep.Path == "/metrics" {
		return false
	}

	key := endpointKey(ep.Method, ep.Path)
	if _, ok := noBodyValidationEndpoints[key]; ok {
		return false
	}
	if strings.Contains(ep.Path, ":") {
		return true
	}
	return ep.Method == http.MethodPost || ep.Method == http.MethodPut || ep.Method == http.MethodPatch
}

var noBodyValidationEndpoints = map[string]struct{}{
	"POST /api/v1/auth/logout":                    {},
	"PUT /api/v1/notifications/read-all":          {},
	"PUT /api/v1/expenses/notifications/read-all": {},
	"POST /api/v1/hr/org-chart/simulate":          {},
}

type httpResponse struct {
	Code   int
	Body   []byte
	Header http.Header
}
