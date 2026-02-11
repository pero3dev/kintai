package integrationtest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"gorm.io/gorm"
)

func TestDBConstraintsAndSQLBehavior(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, model.ExpenseAutoMigrate(env.DB))
	require.NoError(t, model.HRAutoMigrate(env.DB))

	t.Run("UNIQUE constraints on users departments and projects", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_ = createTestUser(t, env, model.RoleEmployee, "it-db-unique-user@example.com", "password123")
		err := env.DB.Create(&model.User{
			Email:        "it-db-unique-user@example.com",
			PasswordHash: "dummy-hash",
			FirstName:    "Dup",
			LastName:     "User",
			Role:         model.RoleEmployee,
			IsActive:     true,
		}).Error
		assertUniqueViolation(t, err)

		dept := model.Department{Name: "IT DB UNIQUE"}
		require.NoError(t, env.DB.Create(&dept).Error)
		err = env.DB.Create(&model.Department{Name: "IT DB UNIQUE"}).Error
		assertUniqueViolation(t, err)

		project := model.Project{Name: "DB Unique Project A", Code: "DB-UNIQUE-001", Status: model.ProjectStatusActive}
		require.NoError(t, env.DB.Create(&project).Error)
		err = env.DB.Create(&model.Project{Name: "DB Unique Project B", Code: "DB-UNIQUE-001", Status: model.ProjectStatusActive}).Error
		assertUniqueViolation(t, err)
	})

	t.Run("composite UNIQUE constraints on attendances and shifts", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		user := createTestUser(t, env, model.RoleEmployee, "it-db-composite@example.com", "password123")
		targetDate := time.Date(2032, 4, 1, 0, 0, 0, 0, time.Local)

		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:  user.ID,
			Date:    targetDate,
			Status:  model.AttendanceStatusPresent,
			Note:    "first",
		}).Error)
		err := env.DB.Create(&model.Attendance{
			UserID: user.ID,
			Date:   targetDate,
			Status: model.AttendanceStatusPresent,
			Note:   "duplicate",
		}).Error
		assertUniqueViolation(t, err)

		require.NoError(t, env.DB.Create(&model.Shift{
			UserID:    user.ID,
			Date:      targetDate,
			ShiftType: model.ShiftTypeDay,
			Note:      "first",
		}).Error)
		err = env.DB.Create(&model.Shift{
			UserID:    user.ID,
			Date:      targetDate,
			ShiftType: model.ShiftTypeDay,
			Note:      "duplicate",
		}).Error
		assertUniqueViolation(t, err)
	})

	t.Run("foreign key behavior CASCADE and SET NULL", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		deptCascade := model.Department{Name: "IT DB CASCADE"}
		require.NoError(t, env.DB.Create(&deptCascade).Error)
		userCascade := model.User{
			Email:        "it-db-cascade@example.com",
			PasswordHash: "dummy-hash",
			FirstName:    "Cascade",
			LastName:     "User",
			Role:         model.RoleEmployee,
			DepartmentID: &deptCascade.ID,
			IsActive:     true,
		}
		require.NoError(t, env.DB.Create(&userCascade).Error)

		attendance := model.Attendance{
			UserID: userCascade.ID,
			Date:   time.Date(2032, 4, 2, 0, 0, 0, 0, time.Local),
			Status: model.AttendanceStatusPresent,
		}
		require.NoError(t, env.DB.Create(&attendance).Error)

		require.NoError(t, env.DB.Unscoped().Delete(&model.User{}, "id = ?", userCascade.ID).Error)

		var attendanceCount int64
		require.NoError(t, env.DB.Unscoped().Model(&model.Attendance{}).Where("id = ?", attendance.ID).Count(&attendanceCount).Error)
		require.Equal(t, int64(0), attendanceCount)

		deptSetNull := model.Department{Name: "IT DB SETNULL"}
		require.NoError(t, env.DB.Create(&deptSetNull).Error)
		userSetNull := model.User{
			Email:        "it-db-setnull@example.com",
			PasswordHash: "dummy-hash",
			FirstName:    "Set",
			LastName:     "Null",
			Role:         model.RoleEmployee,
			DepartmentID: &deptSetNull.ID,
			IsActive:     true,
		}
		require.NoError(t, env.DB.Create(&userSetNull).Error)

		require.NoError(t, env.DB.Unscoped().Delete(&model.Department{}, "id = ?", deptSetNull.ID).Error)

		var reloaded model.User
		require.NoError(t, env.DB.Unscoped().First(&reloaded, "id = ?", userSetNull.ID).Error)
		require.Nil(t, reloaded.DepartmentID)
	})

	t.Run("preload relation retrieval loads related entities", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		dept := model.Department{Name: "IT DB PRELOAD"}
		require.NoError(t, env.DB.Create(&dept).Error)

		user := model.User{
			Email:        "it-db-preload@example.com",
			PasswordHash: "dummy-hash",
			FirstName:    "Pre",
			LastName:     "Load",
			Role:         model.RoleEmployee,
			DepartmentID: &dept.ID,
			IsActive:     true,
		}
		require.NoError(t, env.DB.Create(&user).Error)

		project := model.Project{
			Name:      "IT DB PRELOAD PROJECT",
			Code:      "DB-PRELOAD-001",
			Status:    model.ProjectStatusActive,
			ManagerID: &user.ID,
		}
		require.NoError(t, env.DB.Create(&project).Error)

		entry := model.TimeEntry{
			UserID:      user.ID,
			ProjectID:   project.ID,
			Date:        time.Date(2032, 4, 3, 0, 0, 0, 0, time.Local),
			Minutes:     90,
			Description: "preload check",
		}
		require.NoError(t, env.DB.Create(&entry).Error)

		repos := repository.NewRepositories(env.DB)

		fetchedUser, err := repos.User.FindByID(context.Background(), user.ID)
		require.NoError(t, err)
		require.NotNil(t, fetchedUser.Department)
		require.Equal(t, dept.ID, fetchedUser.Department.ID)

		fetchedEntry, err := repos.TimeEntry.FindByID(context.Background(), entry.ID)
		require.NoError(t, err)
		require.NotNil(t, fetchedEntry.User)
		require.NotNil(t, fetchedEntry.Project)
		require.Equal(t, user.ID, fetchedEntry.User.ID)
		require.Equal(t, project.ID, fetchedEntry.Project.ID)
	})

	t.Run("aggregate SQL endpoints return expected non-zero metrics", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		employee, _ := createActorWithHeaders(t, env, model.RoleEmployee)

		now := time.Now()
		clockIn := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.Local)
		clockOut := clockIn.Add(10 * time.Hour)

		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:          employee.ID,
			Date:            now,
			ClockIn:         &clockIn,
			ClockOut:        &clockOut,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     600,
			OvertimeMinutes: 120,
		}).Error)

		heavyOvertimeUser := createTestUser(t, env, model.RoleEmployee, "it-db-overtime-alert@example.com", "password123")
		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:          heavyOvertimeUser.ID,
			Date:            now,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     600,
			OvertimeMinutes: 2200,
		}).Error)

		require.NoError(t, env.DB.Create(&model.LeaveRequest{
			UserID:    employee.ID,
			LeaveType: model.LeaveTypePaid,
			StartDate: now,
			EndDate:   now,
			Reason:    "pending leave",
			Status:    model.ApprovalStatusPending,
		}).Error)

		expense := model.Expense{
			UserID:      employee.ID,
			Title:       "DB SQL Expense",
			Status:      model.ExpenseStatusPending,
			TotalAmount: 3400,
			Notes:       "aggregate",
		}
		require.NoError(t, env.DB.Create(&expense).Error)
		require.NoError(t, env.DB.Create(&model.ExpenseItem{
			ExpenseID:   expense.ID,
			ExpenseDate: now,
			Category:    model.ExpenseCategoryMeals,
			Description: "lunch",
			Amount:      3400,
		}).Error)

		hrDept := model.HRDepartment{
			Name:   "IT DB HR DEPT",
			Code:   "DBHR",
			Budget: 12000000,
		}
		require.NoError(t, env.DB.Create(&hrDept).Error)

		hrEmployee := model.HREmployee{
			EmployeeCode:   "DBSQL-" + uuid.NewString()[:8],
			FirstName:      "HR",
			LastName:       "Agg",
			Email:          fmt.Sprintf("it-db-hr-%s@example.com", uuid.NewString()[:8]),
			DepartmentID:   &hrDept.ID,
			EmploymentType: model.EmploymentTypeFullTime,
			Status:         model.EmployeeStatusActive,
			BaseSalary:     400000,
		}
		require.NoError(t, env.DB.Create(&hrEmployee).Error)

		require.NoError(t, env.DB.Create(&model.Offboarding{
			EmployeeID:      hrEmployee.ID,
			Reason:          "resignation",
			Status:          model.OffboardingStatusCompleted,
			LastWorkingDate: now,
			Notes:           "aggregate",
		}).Error)

		dashboardResp := env.DoJSON(t, http.MethodGet, "/api/v1/dashboard/stats", nil, adminHeaders)
		require.Equal(t, http.StatusOK, dashboardResp.Code)
		var dashboard model.DashboardStatsExtended
		require.NoError(t, json.Unmarshal(dashboardResp.Body.Bytes(), &dashboard))
		require.GreaterOrEqual(t, dashboard.TodayPresentCount, 1)
		require.GreaterOrEqual(t, dashboard.PendingLeaves, 1)
		require.GreaterOrEqual(t, dashboard.MonthlyOvertime, 120)

		alertResp := env.DoJSON(t, http.MethodGet, "/api/v1/overtime/alerts", nil, adminHeaders)
		require.Equal(t, http.StatusOK, alertResp.Code)
		var alerts []model.OvertimeAlert
		require.NoError(t, json.Unmarshal(alertResp.Body.Bytes(), &alerts))
		require.NotEmpty(t, alerts)
		require.True(t, containsOvertimeAlertForUser(alerts, heavyOvertimeUser.ID))

		startDate := now.AddDate(0, -1, 0).Format("2006-01-02")
		endDate := now.AddDate(0, 1, 0).Format("2006-01-02")
		reportResp := env.DoJSON(
			t,
			http.MethodGet,
			fmt.Sprintf("/api/v1/expenses/report?start_date=%s&end_date=%s", startDate, endDate),
			nil,
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, reportResp.Code)
		var report model.ExpenseReportResponse
		require.NoError(t, json.Unmarshal(reportResp.Body.Bytes(), &report))
		require.Greater(t, report.TotalAmount, 0.0)
		require.NotEmpty(t, report.CategoryBreakdown)
		require.GreaterOrEqual(t, report.StatusSummary.Pending, 1)

		budgetResp := env.DoJSON(t, http.MethodGet, "/api/v1/hr/salary/budget", nil, adminHeaders)
		require.Equal(t, http.StatusOK, budgetResp.Code)
		var budget map[string]any
		require.NoError(t, json.Unmarshal(budgetResp.Body.Bytes(), &budget))
		require.Greater(t, mapNumber(t, budget, "total_budget"), 0.0)
		require.Greater(t, mapNumber(t, budget, "used_budget"), 0.0)
		require.GreaterOrEqual(t, mapNumber(t, budget, "headcount"), 1.0)

		analyticsResp := env.DoJSON(t, http.MethodGet, "/api/v1/hr/offboarding/analytics", nil, adminHeaders)
		require.Equal(t, http.StatusOK, analyticsResp.Code)
		var analytics map[string]any
		require.NoError(t, json.Unmarshal(analyticsResp.Body.Bytes(), &analytics))
		require.GreaterOrEqual(t, mapNumber(t, analytics, "total_departures"), 1.0)
		require.Greater(t, mapNumber(t, analytics, "turnover_rate"), 0.0)
		reasons, ok := analytics["reason_breakdown"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, reasons)
	})

	t.Run("soft delete hides records from search and restore re-exposes them", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		project := model.Project{
			Name:   "IT DB SOFT DELETE",
			Code:   "DB-SOFT-001",
			Status: model.ProjectStatusActive,
		}
		require.NoError(t, env.DB.Create(&project).Error)

		beforeResp := env.DoJSON(t, http.MethodGet, "/api/v1/projects?status=active&page=1&page_size=20", nil, adminHeaders)
		require.Equal(t, http.StatusOK, beforeResp.Code)
		var before paginatedProjectsResponse
		require.NoError(t, json.Unmarshal(beforeResp.Body.Bytes(), &before))
		require.Equal(t, int64(1), before.Total)

		require.NoError(t, env.DB.Delete(&project).Error)

		afterDeleteResp := env.DoJSON(t, http.MethodGet, "/api/v1/projects?status=active&page=1&page_size=20", nil, adminHeaders)
		require.Equal(t, http.StatusOK, afterDeleteResp.Code)
		var afterDelete paginatedProjectsResponse
		require.NoError(t, json.Unmarshal(afterDeleteResp.Body.Bytes(), &afterDelete))
		require.Equal(t, int64(0), afterDelete.Total)

		var softDeleted model.Project
		require.NoError(t, env.DB.Unscoped().First(&softDeleted, "id = ?", project.ID).Error)
		require.True(t, softDeleted.DeletedAt.Valid)

		require.NoError(
			t,
			env.DB.Unscoped().
				Model(&model.Project{}).
				Where("id = ?", project.ID).
				Update("deleted_at", gorm.Expr("NULL")).Error,
		)

		afterRestoreResp := env.DoJSON(t, http.MethodGet, "/api/v1/projects?status=active&page=1&page_size=20", nil, adminHeaders)
		require.Equal(t, http.StatusOK, afterRestoreResp.Code)
		var afterRestore paginatedProjectsResponse
		require.NoError(t, json.Unmarshal(afterRestoreResp.Body.Bytes(), &afterRestore))
		require.Equal(t, int64(1), afterRestore.Total)
	})
}

func assertUniqueViolation(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	lower := strings.ToLower(err.Error())
	require.True(t, strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique"), err.Error())
}

func mapNumber(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()

	raw, exists := m[key]
	require.True(t, exists, "missing key: %s", key)

	switch v := raw.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	default:
		t.Fatalf("key %s is not numeric: %T", key, raw)
		return 0
	}
}
