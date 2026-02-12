package integrationtest

import (
	"encoding/json"
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
	throughputWarmupRequest = 10
	throughputRequestCount  = 200
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

type throughputSLO struct {
	minRPS float64
}

type throughputTarget struct {
	name           string
	method         string
	path           string
	body           any
	headers        map[string]string
	expectedStatus int
	slo            throughputSLO
}

type endpointPerformanceBaseline struct {
	maxP95 time.Duration
	maxP99 time.Duration
	minRPS float64
}

type endpointRegressionTarget struct {
	name           string
	method         string
	path           string
	body           any
	headers        map[string]string
	expectedStatus int
	baseline       endpointPerformanceBaseline
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

func TestPerformanceThroughputSLO(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-throughput-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	targets := []throughputTarget{
		{
			name:           "health",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			slo: throughputSLO{
				minRPS: 300,
			},
		},
		{
			name:           "users_me",
			method:         http.MethodGet,
			path:           "/api/v1/users/me",
			headers:        employeeHeaders,
			expectedStatus: http.StatusOK,
			slo: throughputSLO{
				minRPS: 150,
			},
		},
	}

	for _, target := range targets {
		target := target
		t.Run(target.name, func(t *testing.T) {
			rps := measureEndpointRPS(t, env, target)
			t.Logf("throughput_slo target=%s rps=%.2f", target.name, rps)
			require.GreaterOrEqualf(t, rps, target.slo.minRPS, "%s throughput dropped below SLO", target.name)
		})
	}
}

func TestPerformanceEndpointRegressionBaseline(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-regression-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	// Baseline thresholds are conservative values for httptest-based API execution.
	// Update them only with measured evidence when handlers/repositories are intentionally changed.
	targets := []endpointRegressionTarget{
		{
			name:           "health",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			baseline: endpointPerformanceBaseline{
				maxP95: 10 * time.Millisecond,
				maxP99: 25 * time.Millisecond,
				minRPS: 20000,
			},
		},
		{
			name:           "users_me",
			method:         http.MethodGet,
			path:           "/api/v1/users/me",
			headers:        employeeHeaders,
			expectedStatus: http.StatusOK,
			baseline: endpointPerformanceBaseline{
				maxP95: 15 * time.Millisecond,
				maxP99: 30 * time.Millisecond,
				minRPS: 3000,
			},
		},
		{
			name:           "attendance_today",
			method:         http.MethodGet,
			path:           "/api/v1/attendance/today",
			headers:        employeeHeaders,
			expectedStatus: http.StatusOK,
			baseline: endpointPerformanceBaseline{
				maxP95: 20 * time.Millisecond,
				maxP99: 40 * time.Millisecond,
				minRPS: 1500,
			},
		},
	}

	for _, target := range targets {
		target := target
		t.Run(target.name, func(t *testing.T) {
			latencies := measureEndpointLatencies(t, env, responseTimeTarget{
				name:           target.name,
				method:         target.method,
				path:           target.path,
				body:           target.body,
				headers:        target.headers,
				expectedStatus: target.expectedStatus,
			})
			p95 := percentileDuration(latencies, 95)
			p99 := percentileDuration(latencies, 99)

			rps := measureEndpointRPS(t, env, throughputTarget{
				name:           target.name,
				method:         target.method,
				path:           target.path,
				body:           target.body,
				headers:        target.headers,
				expectedStatus: target.expectedStatus,
			})

			t.Logf(
				"endpoint_regression target=%s p95=%s baseline_p95=%s p99=%s baseline_p99=%s rps=%.2f baseline_rps=%.2f",
				target.name,
				p95,
				target.baseline.maxP95,
				p99,
				target.baseline.maxP99,
				rps,
				target.baseline.minRPS,
			)

			require.LessOrEqualf(t, p95, target.baseline.maxP95, "%s p95 exceeded regression baseline", target.name)
			require.LessOrEqualf(t, p99, target.baseline.maxP99, "%s p99 exceeded regression baseline", target.name)
			require.GreaterOrEqualf(t, rps, target.baseline.minRPS, "%s RPS dropped below regression baseline", target.name)
		})
	}
}

func TestPerformanceLargeDatasetQuerySLO(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-large-dataset-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	noiseEmail := fmt.Sprintf("it-large-dataset-noise-%s@example.com", uuid.NewString())
	noiseUser := createTestUser(t, env, model.RoleEmployee, noiseEmail, "password123")

	startDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	recordCountPerUser := 1500
	seedAttendanceRowsInBatches(t, env, user.ID, startDate, recordCountPerUser)
	seedAttendanceRowsInBatches(t, env, noiseUser.ID, startDate, recordCountPerUser)

	path := "/api/v1/attendance?start_date=2020-01-01&end_date=2030-12-31&page=1&page_size=100"

	verificationResp := env.DoJSON(t, http.MethodGet, path, nil, employeeHeaders)
	require.Equal(t, http.StatusOK, verificationResp.Code)

	var verification attendanceListResponse
	require.NoError(t, json.Unmarshal(verificationResp.Body.Bytes(), &verification))
	require.Equal(t, int64(recordCountPerUser), verification.Total)
	require.Len(t, verification.Data, 100)

	latencies := measureEndpointLatencies(t, env, responseTimeTarget{
		name:           "attendance_list_large_dataset",
		method:         http.MethodGet,
		path:           path,
		headers:        employeeHeaders,
		expectedStatus: http.StatusOK,
	})
	p95 := percentileDuration(latencies, 95)
	p99 := percentileDuration(latencies, 99)

	rps := measureEndpointRPS(t, env, throughputTarget{
		name:           "attendance_list_large_dataset",
		method:         http.MethodGet,
		path:           path,
		headers:        employeeHeaders,
		expectedStatus: http.StatusOK,
	})

	// Thresholds are intentionally conservative for stable CI execution.
	maxP95 := 250 * time.Millisecond
	maxP99 := 500 * time.Millisecond
	minRPS := 80.0

	t.Logf(
		"large_dataset_query_slo target=attendance_list p95=%s p99=%s rps=%.2f",
		p95,
		p99,
		rps,
	)

	require.LessOrEqual(t, p95, maxP95, "attendance list p95 exceeded large-dataset threshold")
	require.LessOrEqual(t, p99, maxP99, "attendance list p99 exceeded large-dataset threshold")
	require.GreaterOrEqual(t, rps, minRPS, "attendance list RPS dropped below large-dataset threshold")
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

func measureEndpointRPS(t testing.TB, env *TestEnv, target throughputTarget) float64 {
	t.Helper()

	for i := 0; i < throughputWarmupRequest; i++ {
		resp := env.DoJSON(t, target.method, target.path, target.body, target.headers)
		require.Equalf(t, target.expectedStatus, resp.Code, "%s warmup status", target.name)
	}

	start := time.Now()
	for i := 0; i < throughputRequestCount; i++ {
		resp := env.DoJSON(t, target.method, target.path, target.body, target.headers)
		require.Equalf(t, target.expectedStatus, resp.Code, "%s measured status", target.name)
	}
	elapsed := time.Since(start)
	if elapsed <= 0 {
		elapsed = time.Nanosecond
	}

	return float64(throughputRequestCount) / elapsed.Seconds()
}

func seedAttendanceRowsInBatches(t testing.TB, env *TestEnv, userID uuid.UUID, startDate time.Time, days int) {
	t.Helper()

	rows := make([]model.Attendance, 0, days)
	for i := 0; i < days; i++ {
		day := startDate.AddDate(0, 0, i)
		rows = append(rows, model.Attendance{
			UserID:          userID,
			Date:            day,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     480,
			OvertimeMinutes: 30,
		})
	}

	require.NoError(t, env.DB.CreateInBatches(rows, 500).Error)
}
