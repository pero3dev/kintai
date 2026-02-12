package integrationtest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestDatabaseRecoveryAfterConnectionTermination(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	email := fmt.Sprintf("it-db-recovery-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, model.RoleEmployee, email, "password123")
	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, model.RoleEmployee),
	}

	// Baseline: protected endpoint should work before fault injection.
	baseline := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, headers)
	require.Equal(t, http.StatusOK, baseline.Code)

	controlDB, err := gorm.Open(postgres.Open(env.Config.DatabaseURL), &gorm.Config{})
	require.NoError(t, err)

	controlSQLDB, err := controlDB.DB()
	require.NoError(t, err)
	controlSQLDB.SetMaxOpenConns(1)
	controlSQLDB.SetMaxIdleConns(1)
	t.Cleanup(func() {
		_ = controlSQLDB.Close()
	})

	terminated, err := terminateOtherDBConnections(controlDB)
	require.NoError(t, err)
	require.GreaterOrEqual(t, terminated, int64(1), "expected at least one DB session to be terminated")

	recovered := false
	lastStatus := 0
	for attempt := 0; attempt < 12; attempt++ {
		resp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, headers)
		lastStatus = resp.Code
		if resp.Code == http.StatusOK {
			recovered = true
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	require.Truef(
		t,
		recovered,
		"API did not recover after DB connection termination: terminated=%d last_status=%d",
		terminated,
		lastStatus,
	)

	sqlDB, err := env.DB.DB()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, sqlDB.PingContext(ctx))
}

func terminateOtherDBConnections(controlDB *gorm.DB) (int64, error) {
	type terminateRow struct {
		Terminated bool `gorm:"column:terminated"`
	}

	var rows []terminateRow
	err := controlDB.Raw(`
		SELECT pg_terminate_backend(pid) AS terminated
		FROM pg_stat_activity
		WHERE datname = current_database()
		  AND backend_type = 'client backend'
		  AND pid <> pg_backend_pid()
	`).Scan(&rows).Error
	if err != nil {
		return 0, err
	}

	var terminated int64
	for _, row := range rows {
		if row.Terminated {
			terminated++
		}
	}
	return terminated, nil
}
