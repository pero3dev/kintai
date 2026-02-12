package integrationtest

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

type soakRequestResult struct {
	elapsed time.Duration
	latency time.Duration
	status  int
	err     error
}

type soakWindowMetric struct {
	index     int
	total     int
	errorRate float64
	p95       time.Duration
	rps       float64
}

func TestLoadSoakEndurance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping soak endurance test in short mode")
	}

	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-soak-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	const (
		soakDuration         = 12 * time.Second
		reportWindow         = 3 * time.Second
		concurrency          = 12
		requestInterval      = 20 * time.Millisecond
		warmupCount          = 20
		maxErrorRate         = 0.01
		maxP95               = 350 * time.Millisecond
		minRPS               = 350.0
		maxP95WindowIncrease = 200 * time.Millisecond
	)

	for i := 0; i < warmupCount; i++ {
		resp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, headers)
		require.Equal(t, http.StatusOK, resp.Code)
	}

	results := make(chan soakRequestResult, concurrency*8)
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), soakDuration)
	defer cancel()

	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ticker := time.NewTicker(requestInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					req, err := JSONRequest(http.MethodGet, "/api/v1/users/me", nil, headers)
					if err != nil {
						results <- soakRequestResult{
							elapsed: time.Since(start),
							err:     err,
						}
						continue
					}

					requestStart := time.Now()
					rec := env.DoRequest(req)

					results <- soakRequestResult{
						elapsed: requestStart.Sub(start),
						latency: time.Since(requestStart),
						status:  rec.Code,
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	windowCount := int(math.Ceil(float64(soakDuration) / float64(reportWindow)))
	windowTotals := make([]int, windowCount)
	windowErrors := make([]int, windowCount)
	windowLatencies := make([][]time.Duration, windowCount)

	total := 0
	success := 0
	errors := 0
	latencies := make([]time.Duration, 0, int(float64(concurrency)*soakDuration.Seconds()/requestInterval.Seconds()))

	for result := range results {
		total++

		windowIndex := int(result.elapsed / reportWindow)
		if windowIndex < 0 {
			windowIndex = 0
		}
		if windowIndex >= windowCount {
			windowIndex = windowCount - 1
		}
		windowTotals[windowIndex]++

		if result.err != nil || result.status != http.StatusOK {
			errors++
			windowErrors[windowIndex]++
			continue
		}

		success++
		latencies = append(latencies, result.latency)
		windowLatencies[windowIndex] = append(windowLatencies[windowIndex], result.latency)
	}

	elapsed := time.Since(start)
	if elapsed <= 0 {
		elapsed = time.Nanosecond
	}

	errorRate := float64(errors) / float64(total)
	p95 := percentileDuration(latencies, 95)
	p99 := percentileDuration(latencies, 99)
	rps := float64(total) / elapsed.Seconds()

	t.Logf(
		"soak_endurance duration=%s concurrency=%d interval=%s total=%d success=%d errors=%d error_rate=%.4f p95=%s p99=%s rps=%.2f",
		elapsed,
		concurrency,
		requestInterval,
		total,
		success,
		errors,
		errorRate,
		p95,
		p99,
		rps,
	)

	windowMetrics := make([]soakWindowMetric, 0, windowCount)
	for i := 0; i < windowCount; i++ {
		if windowTotals[i] == 0 {
			continue
		}

		windowErrorRate := float64(windowErrors[i]) / float64(windowTotals[i])
		windowP95 := percentileDuration(windowLatencies[i], 95)
		windowRPS := float64(windowTotals[i]) / reportWindow.Seconds()

		metric := soakWindowMetric{
			index:     i,
			total:     windowTotals[i],
			errorRate: windowErrorRate,
			p95:       windowP95,
			rps:       windowRPS,
		}
		windowMetrics = append(windowMetrics, metric)

		t.Logf(
			"soak_window index=%d total=%d error_rate=%.4f p95=%s rps=%.2f",
			metric.index,
			metric.total,
			metric.errorRate,
			metric.p95,
			metric.rps,
		)
	}

	require.Greater(t, total, 0, "soak test produced no requests")
	require.Greater(t, success, 0, "soak test produced no successful requests")
	require.GreaterOrEqual(t, len(windowMetrics), 2, "soak test should span multiple windows")
	require.LessOrEqual(t, errorRate, maxErrorRate, "soak error rate exceeded threshold")
	require.LessOrEqual(t, p95, maxP95, "soak p95 exceeded threshold")
	require.GreaterOrEqual(t, rps, minRPS, "soak throughput dropped below threshold")

	var firstWindowP95 time.Duration
	for _, metric := range windowMetrics {
		if metric.p95 > 0 {
			firstWindowP95 = metric.p95
			break
		}
	}

	var lastWindowP95 time.Duration
	for i := len(windowMetrics) - 1; i >= 0; i-- {
		if windowMetrics[i].p95 > 0 {
			lastWindowP95 = windowMetrics[i].p95
			break
		}
	}

	if firstWindowP95 > 0 && lastWindowP95 > 0 {
		require.LessOrEqual(
			t,
			lastWindowP95,
			firstWindowP95+maxP95WindowIncrease,
			"soak p95 degraded too much over time",
		)
	}
}
