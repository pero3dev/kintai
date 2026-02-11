package integrationtest

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestRegressionFocusIntegration(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, model.ExpenseAutoMigrate(env.DB))
	require.NoError(t, model.HRAutoMigrate(env.DB))

	t.Run("date boundary regression for UTC/JST month-crossing and closing-date with 4xx and 2xx", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)

		closingDay := mustDate(t, "2026-01-31")
		jst := time.FixedZone("JST", 9*60*60)
		utc := time.UTC

		clockInJST := time.Date(2026, 1, 31, 23, 30, 0, 0, jst)
		clockOutJST := time.Date(2026, 2, 1, 1, 0, 0, 0, jst)
		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:          employee.ID,
			Date:            closingDay,
			ClockIn:         &clockInJST,
			ClockOut:        &clockOutJST,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     90,
			OvertimeMinutes: 0,
		}).Error)

		nextDay := mustDate(t, "2026-02-01")
		clockInUTC := time.Date(2026, 2, 1, 0, 30, 0, 0, utc)
		clockOutUTC := time.Date(2026, 2, 1, 8, 0, 0, 0, utc)
		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:          employee.ID,
			Date:            nextDay,
			ClockIn:         &clockInUTC,
			ClockOut:        &clockOutUTC,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     450,
			OvertimeMinutes: 0,
		}).Error)

		invalidResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance/summary?start_date=2026-02-30&end_date=2026-02-31",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusBadRequest, invalidResp.Code)
		assertErrorResponse(t, invalidResp.Body.Bytes(), http.StatusBadRequest)

		validResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance/summary?start_date=2026-01-31&end_date=2026-02-01",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, validResp.Code)

		var summary model.AttendanceSummary
		require.NoError(t, json.Unmarshal(validResp.Body.Bytes(), &summary))
		require.Equal(t, 2, summary.TotalWorkDays)
		require.Equal(t, 540, summary.TotalWorkMinutes)

		listResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance?start_date=2026-01-31&end_date=2026-02-01&page=1&page_size=20",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, listResp.Code)
		var list attendanceListResponse
		require.NoError(t, json.Unmarshal(listResp.Body.Bytes(), &list))
		require.Equal(t, int64(2), list.Total)
		require.Len(t, list.Data, 2)
		require.Equal(t, "2026-02-01", list.Data[0].Date.Format("2006-01-02"))
		require.Equal(t, "2026-01-31", list.Data[1].Date.Format("2006-01-02"))
	})

	t.Run("mixed role data leakage prevention on user-scoped endpoints", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employeeA, headersA := createActorWithHeaders(t, env, model.RoleEmployee)
		employeeB, headersB := createActorWithHeaders(t, env, model.RoleEmployee)
		manager, managerHeaders := createActorWithHeaders(t, env, model.RoleManager)

		project := model.Project{
			Name:   "Regression Leak Project",
			Code:   "REG-LEAK-001",
			Status: model.ProjectStatusActive,
		}
		require.NoError(t, env.DB.Create(&project).Error)

		workDate := mustDate(t, "2026-03-05")
		seedAttendance(t, env, employeeA.ID, workDate, model.AttendanceStatusPresent, 480, 0)
		seedAttendance(t, env, employeeB.ID, workDate, model.AttendanceStatusPresent, 420, 0)
		seedAttendance(t, env, manager.ID, workDate, model.AttendanceStatusPresent, 450, 0)

		require.Equal(t, http.StatusCreated, env.DoJSON(t, http.MethodPost, "/api/v1/time-entries", map[string]any{
			"project_id":  project.ID,
			"date":        "2026-03-05",
			"minutes":     120,
			"description": "employee-a",
		}, headersA).Code)
		require.Equal(t, http.StatusCreated, env.DoJSON(t, http.MethodPost, "/api/v1/time-entries", map[string]any{
			"project_id":  project.ID,
			"date":        "2026-03-05",
			"minutes":     90,
			"description": "employee-b",
		}, headersB).Code)
		require.Equal(t, http.StatusCreated, env.DoJSON(t, http.MethodPost, "/api/v1/time-entries", map[string]any{
			"project_id":  project.ID,
			"date":        "2026-03-05",
			"minutes":     60,
			"description": "manager",
		}, managerHeaders).Code)

		_ = mustCreateExpense(t, env, headersA, model.ExpenseStatusPending)
		_ = mustCreateExpense(t, env, headersB, model.ExpenseStatusPending)
		_ = mustCreateExpense(t, env, managerHeaders, model.ExpenseStatusPending)

		attendanceA := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance?start_date=2026-03-01&end_date=2026-03-31&page=1&page_size=20",
			nil,
			headersA,
		)
		require.Equal(t, http.StatusOK, attendanceA.Code)
		var attendanceListA attendanceListResponse
		require.NoError(t, json.Unmarshal(attendanceA.Body.Bytes(), &attendanceListA))
		require.Equal(t, int64(1), attendanceListA.Total)
		require.Equal(t, employeeA.ID, attendanceListA.Data[0].UserID)

		timeEntriesA := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/time-entries?start_date=2026-03-01&end_date=2026-03-31",
			nil,
			headersA,
		)
		require.Equal(t, http.StatusOK, timeEntriesA.Code)
		var entriesA []model.TimeEntry
		require.NoError(t, json.Unmarshal(timeEntriesA.Body.Bytes(), &entriesA))
		require.Len(t, entriesA, 1)
		require.Equal(t, employeeA.ID, entriesA[0].UserID)

		expensesA := env.DoJSON(t, http.MethodGet, "/api/v1/expenses?page=1&page_size=20", nil, headersA)
		require.Equal(t, http.StatusOK, expensesA.Code)
		var listA expenseListResponse
		require.NoError(t, json.Unmarshal(expensesA.Body.Bytes(), &listA))
		require.Equal(t, int64(1), listA.Total)
		require.Equal(t, employeeA.ID, listA.Data[0].UserID)

		expensesManager := env.DoJSON(t, http.MethodGet, "/api/v1/expenses?page=1&page_size=20", nil, managerHeaders)
		require.Equal(t, http.StatusOK, expensesManager.Code)
		var listManager expenseListResponse
		require.NoError(t, json.Unmarshal(expensesManager.Body.Bytes(), &listManager))
		require.Equal(t, int64(1), listManager.Total)
		require.Equal(t, manager.ID, listManager.Data[0].UserID)
	})

	t.Run("concurrent correction and expense approval keep consistent state", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		_, managerHeaders := createActorWithHeaders(t, env, model.RoleManager)

		baseDate := mustDate(t, "2026-04-10")
		baseIn := time.Date(2026, 4, 10, 9, 0, 0, 0, time.Local)
		baseOut := time.Date(2026, 4, 10, 18, 0, 0, 0, time.Local)
		attendance := model.Attendance{
			UserID:          employee.ID,
			Date:            baseDate,
			ClockIn:         &baseIn,
			ClockOut:        &baseOut,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     540,
			OvertimeMinutes: 60,
		}
		require.NoError(t, env.DB.Create(&attendance).Error)

		correctionCreate := env.DoJSON(t, http.MethodPost, "/api/v1/corrections", map[string]any{
			"date":                "2026-04-10",
			"corrected_clock_in":  "09:30",
			"corrected_clock_out": "18:30",
			"reason":              "concurrency-check",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, correctionCreate.Code)

		var correction model.AttendanceCorrection
		require.NoError(t, json.Unmarshal(correctionCreate.Body.Bytes(), &correction))

		const correctionWorkers = 6
		correctionCodes := runConcurrentRequests(correctionWorkers, func() int {
			resp := env.DoJSON(
				t,
				http.MethodPut,
				"/api/v1/corrections/"+correction.ID.String()+"/approve",
				map[string]any{"status": model.CorrectionStatusApproved},
				managerHeaders,
			)
			return resp.Code
		})

		var correctionOK int
		for _, code := range correctionCodes {
			require.True(t, code == http.StatusOK || code == http.StatusBadRequest, "unexpected status code: %d", code)
			if code == http.StatusOK {
				correctionOK++
			}
		}
		require.GreaterOrEqual(t, correctionOK, 1)

		var corrected model.AttendanceCorrection
		require.NoError(t, env.DB.First(&corrected, "id = ?", correction.ID).Error)
		require.Equal(t, model.CorrectionStatusApproved, corrected.Status)

		var updatedAttendance model.Attendance
		require.NoError(t, env.DB.First(&updatedAttendance, "id = ?", attendance.ID).Error)
		require.NotNil(t, updatedAttendance.ClockIn)
		require.Equal(t, 9, updatedAttendance.ClockIn.Hour())
		require.Equal(t, 30, updatedAttendance.ClockIn.Minute())

		expense := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusPending)

		const expenseWorkers = 6
		expenseCodes := runConcurrentRequests(expenseWorkers, func() int {
			resp := env.DoJSON(
				t,
				http.MethodPut,
				"/api/v1/expenses/"+expense.ID.String()+"/approve",
				map[string]any{"status": model.ExpenseStatusApproved},
				managerHeaders,
			)
			return resp.Code
		})

		for _, code := range expenseCodes {
			require.True(t, code == http.StatusOK || code == http.StatusBadRequest, "unexpected status code: %d", code)
		}

		var approved model.Expense
		require.NoError(t, env.DB.First(&approved, "id = ?", expense.ID).Error)
		require.Equal(t, model.ExpenseStatusApproved, approved.Status)
		require.NotNil(t, approved.ApprovedBy)

		var historyCount int64
		require.NoError(t, env.DB.Model(&model.ExpenseHistory{}).Where("expense_id = ?", expense.ID).Count(&historyCount).Error)
		require.GreaterOrEqual(t, historyCount, int64(2))

		var expenseCount int64
		require.NoError(t, env.DB.Model(&model.Expense{}).Where("id = ?", expense.ID).Count(&expenseCount).Error)
		require.Equal(t, int64(1), expenseCount)
	})
}

func runConcurrentRequests(workers int, fn func() int) []int {
	results := make([]int, workers)
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = fn()
		}(i)
	}

	wg.Wait()
	return results
}
