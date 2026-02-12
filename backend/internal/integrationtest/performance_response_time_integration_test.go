package integrationtest

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

const (
	latencyWarmupRequests   = 5
	latencyMeasuredRequests = 80
)

type responseTimeSLO struct {
	p50Max time.Duration
	p95Max time.Duration
	p99Max time.Duration
}

type responseTimeTarget struct {
	name           string
	method         string
	path           string
	body           any
	headers        map[string]string
	expectedStatus int
	slo            responseTimeSLO
}

func TestPerformanceResponseTimeSLO(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-perf-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	targets := []responseTimeTarget{
		{
			name:           "health",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			slo: responseTimeSLO{
				p50Max: 30 * time.Millisecond,
				p95Max: 80 * time.Millisecond,
				p99Max: 150 * time.Millisecond,
			},
		},
		{
			name:           "users_me",
			method:         http.MethodGet,
			path:           "/api/v1/users/me",
			headers:        employeeHeaders,
			expectedStatus: http.StatusOK,
			slo: responseTimeSLO{
				p50Max: 80 * time.Millisecond,
				p95Max: 160 * time.Millisecond,
				p99Max: 250 * time.Millisecond,
			},
		},
	}

	for _, target := range targets {
		target := target
		t.Run(target.name, func(t *testing.T) {
			latencies := measureEndpointLatencies(t, env, target)

			p50 := percentileDuration(latencies, 50)
			p95 := percentileDuration(latencies, 95)
			p99 := percentileDuration(latencies, 99)

			t.Logf(
				"response_time_slo target=%s p50=%s p95=%s p99=%s",
				target.name,
				p50,
				p95,
				p99,
			)

			require.LessOrEqualf(t, p50, target.slo.p50Max, "%s p50 exceeded", target.name)
			require.LessOrEqualf(t, p95, target.slo.p95Max, "%s p95 exceeded", target.name)
			require.LessOrEqualf(t, p99, target.slo.p99Max, "%s p99 exceeded", target.name)
		})
	}
}

func measureEndpointLatencies(t testing.TB, env *TestEnv, target responseTimeTarget) []time.Duration {
	t.Helper()

	for i := 0; i < latencyWarmupRequests; i++ {
		resp := env.DoJSON(t, target.method, target.path, target.body, target.headers)
		require.Equalf(t, target.expectedStatus, resp.Code, "%s warmup status", target.name)
	}

	latencies := make([]time.Duration, 0, latencyMeasuredRequests)
	for i := 0; i < latencyMeasuredRequests; i++ {
		start := time.Now()
		resp := env.DoJSON(t, target.method, target.path, target.body, target.headers)
		latencies = append(latencies, time.Since(start))
		require.Equalf(t, target.expectedStatus, resp.Code, "%s measured status", target.name)
	}
	return latencies
}

func percentileDuration(samples []time.Duration, percentile float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}

	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	rank := int(math.Ceil((percentile/100)*float64(len(sorted)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}

