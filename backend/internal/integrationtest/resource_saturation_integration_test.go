package integrationtest

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestLoadResourceSaturation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping resource saturation test in short mode")
	}

	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-resource-saturation-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	target := responseTimeTarget{
		name:           "users_me_under_resource_saturation",
		method:         http.MethodGet,
		path:           "/api/v1/users/me",
		headers:        headers,
		expectedStatus: http.StatusOK,
	}

	sqlDB, err := env.DB.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxIdleTime(0)
	sqlDB.SetConnMaxLifetime(0)

	stressCtx, stopStress := context.WithCancel(context.Background())
	var stressWG sync.WaitGroup

	// CPU pressure
	cpuWorkers := runtime.NumCPU() / 2
	if cpuWorkers < 2 {
		cpuWorkers = 2
	}
	if cpuWorkers > 4 {
		cpuWorkers = 4
	}
	for i := 0; i < cpuWorkers; i++ {
		stressWG.Add(1)
		go func() {
			defer stressWG.Done()
			runCPUBurner(stressCtx)
		}()
	}

	// Memory pressure
	for i := 0; i < 2; i++ {
		stressWG.Add(1)
		go func() {
			defer stressWG.Done()
			runMemoryPressure(stressCtx, 2*1024*1024, 12)
		}()
	}

	// DB connection pressure: continuously occupy the limited DB pool.
	for i := 0; i < 2; i++ {
		stressWG.Add(1)
		go func() {
			defer stressWG.Done()
			runDBConnectionPressure(stressCtx, env)
		}()
	}

	result := runLoadProfileScenario(t, env, target, 50, 1500)

	stopStress()
	stressWG.Wait()

	t.Logf(
		"resource_saturation cpu_workers=%d db_max_open_conns=%d requests=%d success=%d status_errors=%d request_errors=%d error_rate=%.4f p95=%s p99=%s rps=%.2f",
		cpuWorkers,
		1,
		result.totalCount,
		result.successCount,
		result.statusErrors,
		result.requestErrors,
		result.errorRate,
		result.p95,
		result.p99,
		result.rps,
	)

	require.Greater(t, result.successCount, 0, "resource saturation test produced no successful requests")
	require.LessOrEqual(t, result.errorRate, 0.05, "resource saturation error rate exceeded threshold")
	require.LessOrEqual(t, result.p95, 900*time.Millisecond, "resource saturation p95 exceeded threshold")
	require.GreaterOrEqual(t, result.rps, 100.0, "resource saturation throughput dropped below threshold")
}

func runCPUBurner(ctx context.Context) {
	var x uint64
	for {
		select {
		case <-ctx.Done():
			return
		default:
			x = x*1664525 + 1013904223
			if x&0x3fff == 0 {
				runtime.Gosched()
			}
		}
	}
}

func runMemoryPressure(ctx context.Context, chunkSize int, keepChunks int) {
	if keepChunks < 1 {
		keepChunks = 1
	}

	chunks := make([][]byte, keepChunks)
	index := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf := make([]byte, chunkSize)
			for i := 0; i < len(buf); i += 4096 {
				buf[i] = byte(i)
			}
			chunks[index%keepChunks] = buf
			index++
			time.Sleep(2 * time.Millisecond)
		}
	}
}

func runDBConnectionPressure(ctx context.Context, env *TestEnv) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = env.DB.Exec("SELECT pg_sleep(0.02)").Error
		}
	}
}
