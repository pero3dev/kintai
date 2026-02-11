package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

type attendanceListResponse struct {
	Data       []model.Attendance `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type leaveListResponse struct {
	Data       []model.LeaveRequest `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

type overtimeListResponse struct {
	Data       []model.OvertimeRequest `json:"data"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

func TestAttendanceDomainIntegration(t *testing.T) {
	env := NewTestEnv(t, nil)

	t.Run("clock-in to clock-out flow with attendance list and today status", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		startDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		clockInResp := env.DoJSON(
			t,
			http.MethodPost,
			"/api/v1/attendance/clock-in",
			map[string]any{"note": "start"},
			employeeHeaders,
		)
		require.Equal(t, http.StatusCreated, clockInResp.Code)

		var clockedIn model.Attendance
		require.NoError(t, json.Unmarshal(clockInResp.Body.Bytes(), &clockedIn))
		require.Equal(t, employee.ID, clockedIn.UserID)
		require.NotNil(t, clockedIn.ClockIn)
		require.Nil(t, clockedIn.ClockOut)
		require.Equal(t, model.AttendanceStatusPresent, clockedIn.Status)

		clockOutResp := env.DoJSON(
			t,
			http.MethodPost,
			"/api/v1/attendance/clock-out",
			map[string]any{"note": "end"},
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, clockOutResp.Code)

		var clockedOut model.Attendance
		require.NoError(t, json.Unmarshal(clockOutResp.Body.Bytes(), &clockedOut))
		require.Equal(t, clockedIn.ID, clockedOut.ID)
		require.NotNil(t, clockedOut.ClockOut)
		require.GreaterOrEqual(t, clockedOut.WorkMinutes, 0)

		listResp := env.DoJSON(
			t,
			http.MethodGet,
			fmt.Sprintf("/api/v1/attendance?start_date=%s&end_date=%s&page=1&page_size=20", startDate, endDate),
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, listResp.Code)
		var list attendanceListResponse
		require.NoError(t, json.Unmarshal(listResp.Body.Bytes(), &list))
		require.Equal(t, int64(1), list.Total)
		require.Len(t, list.Data, 1)
		require.Equal(t, clockedIn.ID, list.Data[0].ID)

		todayResp := env.DoJSON(t, http.MethodGet, "/api/v1/attendance/today", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, todayResp.Code)
		var todayStatus model.Attendance
		require.NoError(t, json.Unmarshal(todayResp.Body.Bytes(), &todayStatus))
		require.Equal(t, clockedIn.ID, todayStatus.ID)
	})

	t.Run("clock-in and clock-out error cases", func(t *testing.T) {
		t.Run("duplicate clock-in returns 400", func(t *testing.T) {
			require.NoError(t, env.ResetDB())
			_, headers := createActorWithHeaders(t, env, model.RoleEmployee)

			first := env.DoJSON(t, http.MethodPost, "/api/v1/attendance/clock-in", map[string]any{"note": "first"}, headers)
			require.Equal(t, http.StatusCreated, first.Code)

			second := env.DoJSON(t, http.MethodPost, "/api/v1/attendance/clock-in", map[string]any{"note": "second"}, headers)
			require.Equal(t, http.StatusBadRequest, second.Code)
			assertErrorResponse(t, second.Body.Bytes(), http.StatusBadRequest)
		})

		t.Run("clock-out without clock-in returns 400", func(t *testing.T) {
			require.NoError(t, env.ResetDB())
			_, headers := createActorWithHeaders(t, env, model.RoleEmployee)

			resp := env.DoJSON(t, http.MethodPost, "/api/v1/attendance/clock-out", map[string]any{"note": "end"}, headers)
			require.Equal(t, http.StatusBadRequest, resp.Code)
			assertErrorResponse(t, resp.Body.Bytes(), http.StatusBadRequest)
		})

		t.Run("duplicate clock-out returns 400", func(t *testing.T) {
			require.NoError(t, env.ResetDB())
			_, headers := createActorWithHeaders(t, env, model.RoleEmployee)

			in := env.DoJSON(t, http.MethodPost, "/api/v1/attendance/clock-in", map[string]any{"note": "start"}, headers)
			require.Equal(t, http.StatusCreated, in.Code)

			firstOut := env.DoJSON(t, http.MethodPost, "/api/v1/attendance/clock-out", map[string]any{"note": "first"}, headers)
			require.Equal(t, http.StatusOK, firstOut.Code)

			secondOut := env.DoJSON(t, http.MethodPost, "/api/v1/attendance/clock-out", map[string]any{"note": "second"}, headers)
			require.Equal(t, http.StatusBadRequest, secondOut.Code)
			assertErrorResponse(t, secondOut.Body.Bytes(), http.StatusBadRequest)
		})
	})

	t.Run("attendance list supports date range, paging and boundary dates", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		otherUser, _ := createActorWithHeaders(t, env, model.RoleEmployee)

		seedAttendance(t, env, employee.ID, mustDate(t, "2026-01-31"), model.AttendanceStatusPresent, 480, 0)
		seedAttendance(t, env, employee.ID, mustDate(t, "2026-02-01"), model.AttendanceStatusPresent, 450, 0)
		seedAttendance(t, env, employee.ID, mustDate(t, "2026-02-02"), model.AttendanceStatusPresent, 420, 0)
		seedAttendance(t, env, employee.ID, mustDate(t, "2026-02-03"), model.AttendanceStatusPresent, 480, 0)
		seedAttendance(t, env, otherUser.ID, mustDate(t, "2026-02-01"), model.AttendanceStatusPresent, 480, 0)

		page1Resp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance?start_date=2026-01-31&end_date=2026-02-02&page=1&page_size=2",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, page1Resp.Code)

		var page1 attendanceListResponse
		require.NoError(t, json.Unmarshal(page1Resp.Body.Bytes(), &page1))
		require.Equal(t, int64(3), page1.Total)
		require.Equal(t, 1, page1.Page)
		require.Equal(t, 2, page1.PageSize)
		require.Equal(t, 2, page1.TotalPages)
		require.Len(t, page1.Data, 2)
		require.Equal(t, "2026-02-02", page1.Data[0].Date.Format("2006-01-02"))
		require.Equal(t, "2026-02-01", page1.Data[1].Date.Format("2006-01-02"))

		page2Resp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance?start_date=2026-01-31&end_date=2026-02-02&page=2&page_size=2",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, page2Resp.Code)

		var page2 attendanceListResponse
		require.NoError(t, json.Unmarshal(page2Resp.Body.Bytes(), &page2))
		require.Len(t, page2.Data, 1)
		require.Equal(t, "2026-01-31", page2.Data[0].Date.Format("2006-01-02"))
	})

	t.Run("leave balance initialize and leave request approval/rejection flow", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		admin, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		fiscalYear := time.Now().Year()

		initResp := env.DoJSON(
			t,
			http.MethodPost,
			fmt.Sprintf("/api/v1/leave-balances/%s/initialize?fiscal_year=%d", employee.ID, fiscalYear),
			nil,
			adminHeaders,
		)
		require.Equal(t, http.StatusCreated, initResp.Code)

		balanceResp := env.DoJSON(
			t,
			http.MethodGet,
			fmt.Sprintf("/api/v1/leave-balances?fiscal_year=%d", fiscalYear),
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, balanceResp.Code)

		var balances []model.LeaveBalanceResponse
		require.NoError(t, json.Unmarshal(balanceResp.Body.Bytes(), &balances))
		require.GreaterOrEqual(t, len(balances), 3)
		require.True(t, containsLeaveType(balances, model.LeaveTypePaid))

		createApprovedResp := env.DoJSON(t, http.MethodPost, "/api/v1/leaves", map[string]any{
			"leave_type": model.LeaveTypePaid,
			"start_date": "2026-03-10",
			"end_date":   "2026-03-10",
			"reason":     "doctor",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, createApprovedResp.Code)

		var approvedTarget model.LeaveRequest
		require.NoError(t, json.Unmarshal(createApprovedResp.Body.Bytes(), &approvedTarget))
		require.Equal(t, model.ApprovalStatusPending, approvedTarget.Status)

		approveResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/leaves/"+approvedTarget.ID.String()+"/approve",
			map[string]any{"status": model.ApprovalStatusApproved},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, approveResp.Code)
		var approved model.LeaveRequest
		require.NoError(t, json.Unmarshal(approveResp.Body.Bytes(), &approved))
		require.Equal(t, model.ApprovalStatusApproved, approved.Status)
		require.NotNil(t, approved.ApprovedBy)
		require.Equal(t, admin.ID, *approved.ApprovedBy)

		createRejectedResp := env.DoJSON(t, http.MethodPost, "/api/v1/leaves", map[string]any{
			"leave_type": model.LeaveTypePaid,
			"start_date": "2026-03-11",
			"end_date":   "2026-03-11",
			"reason":     "private",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, createRejectedResp.Code)
		var rejectedTarget model.LeaveRequest
		require.NoError(t, json.Unmarshal(createRejectedResp.Body.Bytes(), &rejectedTarget))

		rejectResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/leaves/"+rejectedTarget.ID.String()+"/approve",
			map[string]any{
				"status":          model.ApprovalStatusRejected,
				"rejected_reason": "conflict",
			},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, rejectResp.Code)
		var rejected model.LeaveRequest
		require.NoError(t, json.Unmarshal(rejectResp.Body.Bytes(), &rejected))
		require.Equal(t, model.ApprovalStatusRejected, rejected.Status)
		require.Equal(t, "conflict", rejected.RejectedReason)

		pendingResp := env.DoJSON(t, http.MethodGet, "/api/v1/leaves/pending?page=1&page_size=20", nil, adminHeaders)
		require.Equal(t, http.StatusOK, pendingResp.Code)
		var pending leaveListResponse
		require.NoError(t, json.Unmarshal(pendingResp.Body.Bytes(), &pending))
		require.Equal(t, int64(0), pending.Total)
	})

	t.Run("overtime request create/approve and overtime alerts threshold", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		admin, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/overtime", map[string]any{
			"date":            "2026-03-12",
			"planned_minutes": 120,
			"reason":          "release",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var request model.OvertimeRequest
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &request))
		require.Equal(t, model.OvertimeStatusPending, request.Status)

		approveResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/overtime/"+request.ID.String()+"/approve",
			map[string]any{"status": model.OvertimeStatusApproved},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, approveResp.Code)

		var approved model.OvertimeRequest
		require.NoError(t, json.Unmarshal(approveResp.Body.Bytes(), &approved))
		require.Equal(t, model.OvertimeStatusApproved, approved.Status)
		require.NotNil(t, approved.ApprovedBy)
		require.Equal(t, admin.ID, *approved.ApprovedBy)

		now := time.Now()
		seedAttendance(t, env, employee.ID, time.Date(now.Year(), now.Month(), 10, 0, 0, 0, 0, time.Local), model.AttendanceStatusPresent, 600, 2200)

		alertResp := env.DoJSON(t, http.MethodGet, "/api/v1/overtime/alerts", nil, adminHeaders)
		require.Equal(t, http.StatusOK, alertResp.Code)

		var alerts []model.OvertimeAlert
		require.NoError(t, json.Unmarshal(alertResp.Body.Bytes(), &alerts))
		require.NotEmpty(t, alerts)
		require.True(t, containsOvertimeAlertForUser(alerts, employee.ID))

		pendingResp := env.DoJSON(t, http.MethodGet, "/api/v1/overtime/pending?page=1&page_size=20", nil, adminHeaders)
		require.Equal(t, http.StatusOK, pendingResp.Code)

		var pending overtimeListResponse
		require.NoError(t, json.Unmarshal(pendingResp.Body.Bytes(), &pending))
		require.Equal(t, int64(0), pending.Total)
	})

	t.Run("attendance correction request approve/reject flow", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		day := mustDate(t, "2026-03-15")
		clockIn := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, time.Local)
		clockOut := time.Date(day.Year(), day.Month(), day.Day(), 18, 0, 0, 0, time.Local)
		baseAttendance := &model.Attendance{
			UserID:          employee.ID,
			Date:            day,
			ClockIn:         &clockIn,
			ClockOut:        &clockOut,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     540,
			OvertimeMinutes: 60,
		}
		require.NoError(t, env.DB.Create(baseAttendance).Error)

		createCorrectionResp := env.DoJSON(t, http.MethodPost, "/api/v1/corrections", map[string]any{
			"date":                "2026-03-15",
			"corrected_clock_in":  "09:30",
			"corrected_clock_out": "18:30",
			"reason":              "device issue",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, createCorrectionResp.Code)

		var correction model.AttendanceCorrection
		require.NoError(t, json.Unmarshal(createCorrectionResp.Body.Bytes(), &correction))
		require.Equal(t, model.CorrectionStatusPending, correction.Status)

		approveResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/corrections/"+correction.ID.String()+"/approve",
			map[string]any{"status": model.CorrectionStatusApproved},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, approveResp.Code)

		var approved model.AttendanceCorrection
		require.NoError(t, json.Unmarshal(approveResp.Body.Bytes(), &approved))
		require.Equal(t, model.CorrectionStatusApproved, approved.Status)

		var updated model.Attendance
		require.NoError(t, env.DB.First(&updated, "id = ?", baseAttendance.ID).Error)
		require.NotNil(t, updated.ClockIn)
		require.NotNil(t, updated.ClockOut)
		require.Equal(t, 9, updated.ClockIn.Hour())
		require.Equal(t, 30, updated.ClockIn.Minute())

		createRejectedResp := env.DoJSON(t, http.MethodPost, "/api/v1/corrections", map[string]any{
			"date":                "2026-03-15",
			"corrected_clock_in":  "10:00",
			"corrected_clock_out": "19:00",
			"reason":              "second try",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, createRejectedResp.Code)
		var rejectedTarget model.AttendanceCorrection
		require.NoError(t, json.Unmarshal(createRejectedResp.Body.Bytes(), &rejectedTarget))

		rejectResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/corrections/"+rejectedTarget.ID.String()+"/approve",
			map[string]any{
				"status":          model.CorrectionStatusRejected,
				"rejected_reason": "invalid proof",
			},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, rejectResp.Code)
		var rejected model.AttendanceCorrection
		require.NoError(t, json.Unmarshal(rejectResp.Body.Bytes(), &rejected))
		require.Equal(t, model.CorrectionStatusRejected, rejected.Status)

		var afterReject model.Attendance
		require.NoError(t, env.DB.First(&afterReject, "id = ?", baseAttendance.ID).Error)
		require.Equal(t, 9, afterReject.ClockIn.Hour())
		require.Equal(t, 30, afterReject.ClockIn.Minute())
	})

	t.Run("monthly summary boundary with late-night, holiday and closing date", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)

		closingDay := mustDate(t, "2026-01-31")
		lateNightIn := time.Date(2026, 1, 31, 23, 30, 0, 0, time.Local)
		lateNightOut := time.Date(2026, 2, 1, 1, 30, 0, 0, time.Local)

		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:          employee.ID,
			Date:            closingDay,
			ClockIn:         &lateNightIn,
			ClockOut:        &lateNightOut,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     120,
			OvertimeMinutes: 0,
		}).Error)

		seedAttendance(t, env, employee.ID, mustDate(t, "2026-02-01"), model.AttendanceStatusPresent, 480, 60)
		seedAttendance(t, env, employee.ID, mustDate(t, "2026-02-02"), model.AttendanceStatusLeave, 0, 0)
		seedAttendance(t, env, employee.ID, mustDate(t, "2026-02-03"), model.AttendanceStatusHoliday, 0, 0)
		seedAttendance(t, env, employee.ID, mustDate(t, "2026-01-30"), model.AttendanceStatusAbsent, 0, 0)

		summaryResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance/summary?start_date=2026-01-31&end_date=2026-02-03",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, summaryResp.Code)

		var summary model.AttendanceSummary
		require.NoError(t, json.Unmarshal(summaryResp.Body.Bytes(), &summary))
		require.Equal(t, 2, summary.TotalWorkDays)
		require.Equal(t, 600, summary.TotalWorkMinutes)
		require.Equal(t, 60, summary.TotalOvertimeMinutes)
		require.Equal(t, 1, summary.LeaveDays)
		require.Equal(t, 0, summary.AbsentDays)

		febSummaryResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/attendance/summary?start_date=2026-02-01&end_date=2026-02-03",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, febSummaryResp.Code)

		var febSummary model.AttendanceSummary
		require.NoError(t, json.Unmarshal(febSummaryResp.Body.Bytes(), &febSummary))
		require.Equal(t, 1, febSummary.TotalWorkDays)
		require.Equal(t, 480, febSummary.TotalWorkMinutes)
		require.Equal(t, 60, febSummary.TotalOvertimeMinutes)
		require.Equal(t, 1, febSummary.LeaveDays)
	})
}

func createActorWithHeaders(t *testing.T, env *TestEnv, role model.Role) (*model.User, map[string]string) {
	t.Helper()

	email := fmt.Sprintf("it-attendance-%s@example.com", uuid.NewString())
	user := createTestUser(t, env, role, email, "password123")
	headers := map[string]string{
		"Authorization": env.MustBearerToken(t, user.ID, role),
	}
	return user, headers
}

func assertErrorResponse(t *testing.T, body []byte, expectedCode int) {
	t.Helper()

	var errResp model.ErrorResponse
	require.NoError(t, json.Unmarshal(body, &errResp))
	require.Equal(t, expectedCode, errResp.Code)
	require.NotEmpty(t, errResp.Message)
}

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()

	parsed, err := time.Parse("2006-01-02", raw)
	require.NoError(t, err)
	return parsed
}

func seedAttendance(
	t *testing.T,
	env *TestEnv,
	userID uuid.UUID,
	date time.Time,
	status model.AttendanceStatus,
	workMinutes int,
	overtimeMinutes int,
) {
	t.Helper()

	in := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.Local)
	out := in.Add(time.Duration(workMinutes) * time.Minute)

	row := &model.Attendance{
		UserID:          userID,
		Date:            date,
		ClockIn:         &in,
		ClockOut:        &out,
		Status:          status,
		WorkMinutes:     workMinutes,
		OvertimeMinutes: overtimeMinutes,
	}
	require.NoError(t, env.DB.Create(row).Error)
}

func containsLeaveType(balances []model.LeaveBalanceResponse, leaveType model.LeaveType) bool {
	for _, balance := range balances {
		if balance.LeaveType == leaveType {
			return true
		}
	}
	return false
}

func containsOvertimeAlertForUser(alerts []model.OvertimeAlert, userID uuid.UUID) bool {
	for _, alert := range alerts {
		if alert.UserID == userID {
			return alert.IsMonthlyExceeded || alert.IsYearlyExceeded
		}
	}
	return false
}
