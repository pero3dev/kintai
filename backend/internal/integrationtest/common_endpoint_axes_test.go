package integrationtest

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

const (
	expectedEndpointCount     = 199
	expectedRoleEndpointCount = 34
)

type endpointDef struct {
	Method string
	Path   string
}

var (
	pathParamPattern = regexp.MustCompile(`:[A-Za-z_][A-Za-z0-9_]*`)

	publicEndpointDefs = []endpointDef{
		{Method: http.MethodGet, Path: "/health"},
		{Method: http.MethodGet, Path: "/metrics"},
		{Method: http.MethodPost, Path: "/api/v1/auth/login"},
		{Method: http.MethodPost, Path: "/api/v1/auth/register"},
		{Method: http.MethodPost, Path: "/api/v1/auth/refresh"},
	}

	roleEndpointDefs = []endpointDef{
		{Method: http.MethodGet, Path: "/api/v1/users"},
		{Method: http.MethodPost, Path: "/api/v1/users"},
		{Method: http.MethodPut, Path: "/api/v1/users/:id"},
		{Method: http.MethodDelete, Path: "/api/v1/users/:id"},
		{Method: http.MethodPost, Path: "/api/v1/shifts"},
		{Method: http.MethodPost, Path: "/api/v1/shifts/bulk"},
		{Method: http.MethodDelete, Path: "/api/v1/shifts/:id"},
		{Method: http.MethodPost, Path: "/api/v1/projects"},
		{Method: http.MethodPut, Path: "/api/v1/projects/:id"},
		{Method: http.MethodDelete, Path: "/api/v1/projects/:id"},
		{Method: http.MethodGet, Path: "/api/v1/time-entries/summary"},
		{Method: http.MethodPost, Path: "/api/v1/holidays"},
		{Method: http.MethodPut, Path: "/api/v1/holidays/:id"},
		{Method: http.MethodDelete, Path: "/api/v1/holidays/:id"},
		{Method: http.MethodGet, Path: "/api/v1/approval-flows"},
		{Method: http.MethodGet, Path: "/api/v1/approval-flows/:id"},
		{Method: http.MethodPost, Path: "/api/v1/approval-flows"},
		{Method: http.MethodPut, Path: "/api/v1/approval-flows/:id"},
		{Method: http.MethodDelete, Path: "/api/v1/approval-flows/:id"},
		{Method: http.MethodGet, Path: "/api/v1/export/attendance"},
		{Method: http.MethodGet, Path: "/api/v1/export/leaves"},
		{Method: http.MethodGet, Path: "/api/v1/export/overtime"},
		{Method: http.MethodGet, Path: "/api/v1/export/projects"},
		{Method: http.MethodGet, Path: "/api/v1/dashboard/stats"},
		{Method: http.MethodGet, Path: "/api/v1/leaves/pending"},
		{Method: http.MethodPut, Path: "/api/v1/leaves/:id/approve"},
		{Method: http.MethodGet, Path: "/api/v1/overtime/pending"},
		{Method: http.MethodPut, Path: "/api/v1/overtime/:id/approve"},
		{Method: http.MethodGet, Path: "/api/v1/overtime/alerts"},
		{Method: http.MethodGet, Path: "/api/v1/corrections/pending"},
		{Method: http.MethodPut, Path: "/api/v1/corrections/:id/approve"},
		{Method: http.MethodGet, Path: "/api/v1/leave-balances/:user_id"},
		{Method: http.MethodPut, Path: "/api/v1/leave-balances/:user_id/:leave_type"},
		{Method: http.MethodPost, Path: "/api/v1/leave-balances/:user_id/initialize"},
	}
)

func TestCommonEndpointAxes(t *testing.T) {
	env := NewTestEnv(t, nil)

	allEndpoints := collectRegisteredEndpoints(env)
	require.Len(t, allEndpoints, expectedEndpointCount, "registered endpoint count changed; update checklist + common axis tests")

	publicSet := toEndpointSet(publicEndpointDefs)
	roleSet := toEndpointSet(roleEndpointDefs)
	require.Len(t, roleSet, expectedRoleEndpointCount, "role endpoint definition count changed")

	assertSubset(t, allEndpoints, publicSet, "public")
	assertSubset(t, allEndpoints, roleSet, "role")

	protectedEndpoints := make([]endpointDef, 0, len(allEndpoints))
	for _, ep := range allEndpoints {
		if _, ok := publicSet[endpointKey(ep.Method, ep.Path)]; ok {
			continue
		}
		protectedEndpoints = append(protectedEndpoints, ep)
	}
	require.Len(t, protectedEndpoints, expectedEndpointCount-len(publicEndpointDefs))

	t.Run("CA-01 public endpoint allows unauthenticated access", func(t *testing.T) {
		for _, ep := range sortedEndpointDefs(publicEndpointDefs) {
			resp := doEndpointRequest(t, env, ep, nil)
			require.NotEqualf(t, http.StatusUnauthorized, resp.Code, "%s", endpointKey(ep.Method, ep.Path))
			require.NotEqualf(t, http.StatusForbidden, resp.Code, "%s", endpointKey(ep.Method, ep.Path))
		}
	})

	t.Run("CA-02 protected endpoint rejects unauthenticated access", func(t *testing.T) {
		for _, ep := range protectedEndpoints {
			resp := doEndpointRequest(t, env, ep, nil)
			require.Equalf(t, http.StatusUnauthorized, resp.Code, "%s", endpointKey(ep.Method, ep.Path))
		}
	})

	t.Run("CA-03 role protected endpoint rejects employee role", func(t *testing.T) {
		employeeHeaders := map[string]string{
			"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleEmployee),
		}
		for _, ep := range sortedEndpointDefs(roleEndpointDefs) {
			resp := doEndpointRequest(t, env, ep, employeeHeaders)
			require.Equalf(t, http.StatusForbidden, resp.Code, "%s", endpointKey(ep.Method, ep.Path))
		}
	})

	t.Run("CA-04 role protected endpoint allows admin and manager", func(t *testing.T) {
		testCases := []struct {
			name    string
			headers map[string]string
		}{
			{
				name: "admin",
				headers: map[string]string{
					"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleAdmin),
				},
			},
			{
				name: "manager",
				headers: map[string]string{
					"Authorization": env.MustBearerToken(t, uuid.New(), model.RoleManager),
				},
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				for _, ep := range sortedEndpointDefs(roleEndpointDefs) {
					resp := doEndpointRequest(t, env, ep, tc.headers)
					require.NotEqualf(t, http.StatusForbidden, resp.Code, "%s", endpointKey(ep.Method, ep.Path))
				}
			})
		}
	})
}

func collectRegisteredEndpoints(env *TestEnv) []endpointDef {
	allowedMethods := map[string]struct{}{
		http.MethodGet:    {},
		http.MethodPost:   {},
		http.MethodPut:    {},
		http.MethodPatch:  {},
		http.MethodDelete: {},
	}

	seen := make(map[string]struct{})
	endpoints := make([]endpointDef, 0, len(env.Router.Routes()))

	for _, route := range env.Router.Routes() {
		if _, ok := allowedMethods[route.Method]; !ok {
			continue
		}
		key := endpointKey(route.Method, route.Path)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		endpoints = append(endpoints, endpointDef{
			Method: route.Method,
			Path:   route.Path,
		})
	}

	return sortedEndpointDefs(endpoints)
}

func doEndpointRequest(t testing.TB, env *TestEnv, ep endpointDef, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := JSONRequest(ep.Method, materializeEndpointPath(ep.Path), nil, headers)
	require.NoError(t, err)
	return env.DoRequest(req)
}

func materializeEndpointPath(path string) string {
	return pathParamPattern.ReplaceAllStringFunc(path, func(raw string) string {
		name := strings.ToLower(strings.TrimPrefix(raw, ":"))
		switch name {
		case "leave_type":
			return "paid"
		case "taskid", "actionid":
			return "1"
		case "itemkey":
			return "equipment"
		default:
			return "00000000-0000-0000-0000-000000000001"
		}
	})
}

func sortedEndpointDefs(endpoints []endpointDef) []endpointDef {
	sorted := append([]endpointDef(nil), endpoints...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Path == sorted[j].Path {
			return sorted[i].Method < sorted[j].Method
		}
		return sorted[i].Path < sorted[j].Path
	})
	return sorted
}

func toEndpointSet(endpoints []endpointDef) map[string]endpointDef {
	set := make(map[string]endpointDef, len(endpoints))
	for _, ep := range endpoints {
		set[endpointKey(ep.Method, ep.Path)] = ep
	}
	return set
}

func assertSubset(t *testing.T, all []endpointDef, subset map[string]endpointDef, kind string) {
	t.Helper()

	allSet := toEndpointSet(all)
	for key := range subset {
		_, ok := allSet[key]
		require.Truef(t, ok, "missing %s endpoint: %s", kind, key)
	}
}

func endpointKey(method string, path string) string {
	return method + " " + path
}
