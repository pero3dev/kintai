package integrationtest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

type loadProfileScenario struct {
	name          string
	concurrency   int
	totalRequests int
	maxP95        time.Duration
	minRPS        float64
	maxErrorRate  float64
}

type loadProfileResult struct {
	totalCount    int
	successCount  int
	statusErrors  int
	requestErrors int
	errorRate     float64
	rps           float64
	p95           time.Duration
	p99           time.Duration
}

type loadRequestResult struct {
	latency time.Duration
	status  int
	err     error
}

func TestLoadProfileScenarios(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-load-profile-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	target := responseTimeTarget{
		name:           "users_me",
		method:         "GET",
		path:           "/api/v1/users/me",
		headers:        employeeHeaders,
		expectedStatus: 200,
	}

	scenarios := []loadProfileScenario{
		{
			name:          "normal_load",
			concurrency:   5,
			totalRequests: 200,
			maxP95:        80 * time.Millisecond,
			minRPS:        500,
			maxErrorRate:  0.01,
		},
		{
			name:          "peak_load",
			concurrency:   20,
			totalRequests: 1000,
			maxP95:        150 * time.Millisecond,
			minRPS:        700,
			maxErrorRate:  0.01,
		},
		{
			name:          "spike_load",
			concurrency:   50,
			totalRequests: 1500,
			maxP95:        350 * time.Millisecond,
			minRPS:        800,
			maxErrorRate:  0.02,
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			result := runLoadProfileScenario(t, env, target, scenario.concurrency, scenario.totalRequests)

			t.Logf(
				"load_profile scenario=%s concurrency=%d requests=%d success=%d errors=%d status_errors=%d request_errors=%d error_rate=%.4f p95=%s p99=%s rps=%.2f",
				scenario.name,
				scenario.concurrency,
				result.totalCount,
				result.successCount,
				result.statusErrors+result.requestErrors,
				result.statusErrors,
				result.requestErrors,
				result.errorRate,
				result.p95,
				result.p99,
				result.rps,
			)

			require.Greater(t, result.successCount, 0, "%s should have successful requests", scenario.name)
			require.LessOrEqualf(
				t,
				result.errorRate,
				scenario.maxErrorRate,
				"%s error rate exceeded",
				scenario.name,
			)
			require.LessOrEqualf(t, result.p95, scenario.maxP95, "%s p95 exceeded", scenario.name)
			require.GreaterOrEqualf(t, result.rps, scenario.minRPS, "%s throughput dropped", scenario.name)
		})
	}
}

func runLoadProfileScenario(
	t testing.TB,
	env *TestEnv,
	target responseTimeTarget,
	concurrency int,
	totalRequests int,
) loadProfileResult {
	t.Helper()

	const warmupCount = 10
	for i := 0; i < warmupCount; i++ {
		resp := env.DoJSON(t, target.method, target.path, target.body, target.headers)
		require.Equal(t, target.expectedStatus, resp.Code)
	}

	jobs := make(chan struct{}, totalRequests)
	results := make(chan loadRequestResult, totalRequests)

	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				req, err := JSONRequest(target.method, target.path, target.body, target.headers)
				if err != nil {
					results <- loadRequestResult{err: err}
					continue
				}

				start := time.Now()
				rec := env.DoRequest(req)
				results <- loadRequestResult{
					latency: time.Since(start),
					status:  rec.Code,
				}
			}
		}()
	}

	start := time.Now()
	for i := 0; i < totalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	wg.Wait()
	close(results)
	elapsed := time.Since(start)
	if elapsed <= 0 {
		elapsed = time.Nanosecond
	}

	latencies := make([]time.Duration, 0, totalRequests)
	statusErrors := 0
	requestErrors := 0
	successCount := 0

	for result := range results {
		if result.err != nil {
			requestErrors++
			continue
		}
		if result.status != target.expectedStatus {
			statusErrors++
			continue
		}

		successCount++
		latencies = append(latencies, result.latency)
	}

	totalErrors := statusErrors + requestErrors
	errorRate := float64(totalErrors) / float64(totalRequests)

	return loadProfileResult{
		totalCount:    totalRequests,
		successCount:  successCount,
		statusErrors:  statusErrors,
		requestErrors: requestErrors,
		errorRate:     errorRate,
		rps:           float64(totalRequests) / elapsed.Seconds(),
		p95:           percentileDuration(latencies, 95),
		p99:           percentileDuration(latencies, 99),
	}
}
