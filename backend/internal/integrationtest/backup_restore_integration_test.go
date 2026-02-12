package integrationtest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

const (
	maxAllowedRTO = 5 * time.Second
	maxAllowedRPO = 1 * time.Minute
)

type userBackupSnapshot struct {
	CapturedAt time.Time
	Users      []model.User
}

func TestBackupRestoreRecoveryMeetsRPOAndRTO(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, env.ResetDB())

	baselineEmailA := fmt.Sprintf("it-backup-a-%s@example.com", uuid.NewString())
	baselineEmailB := fmt.Sprintf("it-backup-b-%s@example.com", uuid.NewString())
	postBackupEmail := fmt.Sprintf("it-backup-lost-%s@example.com", uuid.NewString())

	baselineUser := createTestUser(t, env, model.RoleEmployee, baselineEmailA, "password123")
	_ = createTestUser(t, env, model.RoleEmployee, baselineEmailB, "password123")

	snapshot := captureUserBackupSnapshot(t, env)
	require.NotEmpty(t, snapshot.Users)

	_ = createTestUser(t, env, model.RoleEmployee, postBackupEmail, "password123")

	disasterAt := time.Now()
	require.NoError(t, env.ResetSchema())
	require.NoError(t, env.ApplyMigrations())

	restoreStartedAt := time.Now()
	restoreUserBackupSnapshot(t, env, snapshot)
	restoreDuration := time.Since(restoreStartedAt)

	rpo := disasterAt.Sub(snapshot.CapturedAt)
	require.Greater(t, rpo, time.Duration(0))
	require.LessOrEqualf(t, rpo, maxAllowedRPO, "RPO exceeded threshold: %s > %s", rpo, maxAllowedRPO)
	require.LessOrEqualf(
		t,
		restoreDuration,
		maxAllowedRTO,
		"RTO exceeded threshold: %s > %s",
		restoreDuration,
		maxAllowedRTO,
	)

	require.True(t, userExistsByEmail(t, env, baselineEmailA), "baseline user A should be restored")
	require.True(t, userExistsByEmail(t, env, baselineEmailB), "baseline user B should be restored")
	require.False(t, userExistsByEmail(t, env, postBackupEmail), "post-backup user should be lost as expected")

	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, baselineUser.ID, model.RoleEmployee),
	}
	meResp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, headers)
	require.Equal(t, http.StatusOK, meResp.Code)
}

func captureUserBackupSnapshot(t testing.TB, env *TestEnv) userBackupSnapshot {
	t.Helper()

	var users []model.User
	require.NoError(t, env.DB.Order("created_at ASC").Find(&users).Error)

	return userBackupSnapshot{
		CapturedAt: time.Now(),
		Users:      users,
	}
}

func restoreUserBackupSnapshot(t testing.TB, env *TestEnv, snapshot userBackupSnapshot) {
	t.Helper()

	if len(snapshot.Users) == 0 {
		return
	}

	require.NoError(
		t,
		env.DB.Omit("Department", "Attendances", "LeaveRequests").Create(&snapshot.Users).Error,
	)
}

func userExistsByEmail(t testing.TB, env *TestEnv, email string) bool {
	t.Helper()

	var count int64
	require.NoError(t, env.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error)
	return count > 0
}
