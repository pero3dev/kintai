package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

type concurrentSession struct {
	userID  uuid.UUID
	headers map[string]string
}

type concurrentSessionResult struct {
	latency      time.Duration
	status       int
	expectedUser uuid.UUID
	actualUser   uuid.UUID
	err          error
}

func TestHighConcurrentMultiUserSessions(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	const (
		userCount       = 80
		requestsPerUser = 20
		totalRequests   = userCount * requestsPerUser
		concurrency     = userCount
	)

	sessions := make([]concurrentSession, 0, userCount)
	for i := 0; i < userCount; i++ {
		email := fmt.Sprintf("it-concurrent-%d-%s@example.com", i, uuid.NewString())
		user := createTestUser(t, env, model.RoleEmployee, email, "password123")
		sessions = append(sessions, concurrentSession{
			userID: user.ID,
			headers: map[string]string{
				"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
			},
		})
	}

	for i := 0; i < userCount; i++ {
		resp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, sessions[i].headers)
		require.Equal(t, http.StatusOK, resp.Code)
	}

	jobs := make(chan concurrentSession, totalRequests)
	results := make(chan concurrentSessionResult, totalRequests)

	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for session := range jobs {
				req, err := JSONRequest(http.MethodGet, "/api/v1/users/me", nil, session.headers)
				if err != nil {
					results <- concurrentSessionResult{
						expectedUser: session.userID,
						err:          err,
					}
					continue
				}

				start := time.Now()
				rec := env.DoRequest(req)
				latency := time.Since(start)

				if rec.Code != http.StatusOK {
					results <- concurrentSessionResult{
						latency:      latency,
						status:       rec.Code,
						expectedUser: session.userID,
					}
					continue
				}

				var me model.User
				if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
					results <- concurrentSessionResult{
						latency:      latency,
						status:       rec.Code,
						expectedUser: session.userID,
						err:          err,
					}
					continue
				}

				results <- concurrentSessionResult{
					latency:      latency,
					status:       rec.Code,
					expectedUser: session.userID,
					actualUser:   me.ID,
				}
			}
		}()
	}

	start := time.Now()
	for _, session := range sessions {
		for i := 0; i < requestsPerUser; i++ {
			jobs <- session
		}
	}
	close(jobs)

	wg.Wait()
	close(results)
	elapsed := time.Since(start)
	if elapsed <= 0 {
		elapsed = time.Nanosecond
	}

	statusErrors := 0
	requestErrors := 0
	sessionMismatches := 0
	successCount := 0
	latencies := make([]time.Duration, 0, totalRequests)

	for result := range results {
		if result.err != nil {
			requestErrors++
			continue
		}
		if result.status != http.StatusOK {
			statusErrors++
			continue
		}
		if result.expectedUser != result.actualUser {
			sessionMismatches++
			continue
		}

		successCount++
		latencies = append(latencies, result.latency)
	}

	totalErrors := statusErrors + requestErrors + sessionMismatches
	errorRate := float64(totalErrors) / float64(totalRequests)
	rps := float64(totalRequests) / elapsed.Seconds()
	p95 := percentileDuration(latencies, 95)
	p99 := percentileDuration(latencies, 99)

	t.Logf(
		"high_concurrency_sessions users=%d requests=%d success=%d status_errors=%d request_errors=%d session_mismatches=%d error_rate=%.4f p95=%s p99=%s rps=%.2f",
		userCount,
		totalRequests,
		successCount,
		statusErrors,
		requestErrors,
		sessionMismatches,
		errorRate,
		p95,
		p99,
		rps,
	)

	require.Equal(t, 0, sessionMismatches, "detected session/user mismatch across concurrent requests")
	require.Greater(t, successCount, 0, "no successful request under high concurrency")
	require.LessOrEqual(t, errorRate, 0.01, "error rate exceeded high-concurrency threshold")
	require.LessOrEqual(t, p95, 500*time.Millisecond, "p95 exceeded high-concurrency threshold")
	require.GreaterOrEqual(t, rps, 700.0, "throughput dropped under high concurrency")
}
