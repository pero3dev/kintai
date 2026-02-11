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

type paginatedNotificationsResponse struct {
	Data       []model.Notification `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

type paginatedUsersResponse struct {
	Data       []model.User `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

type paginatedProjectsResponse struct {
	Data       []model.Project `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func TestSharedDomainIntegration(t *testing.T) {
	env := NewTestEnv(t, nil)

	t.Run("notifications list unread mark read mark all delete and create flow", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		readAt := time.Now()
		unread := model.Notification{
			UserID:  employee.ID,
			Type:    model.NotificationTypeGeneral,
			Title:   "unread",
			Message: "unread",
			IsRead:  false,
		}
		read := model.Notification{
			UserID:  employee.ID,
			Type:    model.NotificationTypeGeneral,
			Title:   "read",
			Message: "read",
			IsRead:  true,
			ReadAt:  &readAt,
		}
		require.NoError(t, env.DB.Create(&unread).Error)
		require.NoError(t, env.DB.Create(&read).Error)

		leaveResp := env.DoJSON(t, http.MethodPost, "/api/v1/leaves", map[string]any{
			"leave_type": model.LeaveTypePaid,
			"start_date": "2030-01-10",
			"end_date":   "2030-01-10",
			"reason":     "integration",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, leaveResp.Code)

		var leave model.LeaveRequest
		require.NoError(t, json.Unmarshal(leaveResp.Body.Bytes(), &leave))

		approveResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/leaves/"+leave.ID.String()+"/approve",
			map[string]any{"status": model.ApprovalStatusApproved},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, approveResp.Code)

		var generated model.Notification
		require.NoError(
			t,
			env.DB.Where("user_id = ? AND type = ?", employee.ID, model.NotificationTypeLeaveApproved).
				Order("created_at DESC").
				First(&generated).Error,
		)

		unreadCountResp := env.DoJSON(t, http.MethodGet, "/api/v1/notifications/unread-count", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, unreadCountResp.Code)
		var unreadCount model.NotificationCount
		require.NoError(t, json.Unmarshal(unreadCountResp.Body.Bytes(), &unreadCount))
		require.Equal(t, 2, unreadCount.Unread)

		listResp := env.DoJSON(t, http.MethodGet, "/api/v1/notifications?is_read=false&page=1&page_size=20", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, listResp.Code)
		var list paginatedNotificationsResponse
		require.NoError(t, json.Unmarshal(listResp.Body.Bytes(), &list))
		require.Equal(t, int64(2), list.Total)
		require.Len(t, list.Data, 2)

		markReadResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/notifications/"+unread.ID.String()+"/read",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusNoContent, markReadResp.Code)

		unreadCountAfterMarkResp := env.DoJSON(t, http.MethodGet, "/api/v1/notifications/unread-count", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, unreadCountAfterMarkResp.Code)
		var unreadCountAfterMark model.NotificationCount
		require.NoError(t, json.Unmarshal(unreadCountAfterMarkResp.Body.Bytes(), &unreadCountAfterMark))
		require.Equal(t, 1, unreadCountAfterMark.Unread)

		markAllResp := env.DoJSON(t, http.MethodPut, "/api/v1/notifications/read-all", nil, employeeHeaders)
		require.Equal(t, http.StatusNoContent, markAllResp.Code)

		unreadCountAfterAllResp := env.DoJSON(t, http.MethodGet, "/api/v1/notifications/unread-count", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, unreadCountAfterAllResp.Code)
		var unreadCountAfterAll model.NotificationCount
		require.NoError(t, json.Unmarshal(unreadCountAfterAllResp.Body.Bytes(), &unreadCountAfterAll))
		require.Equal(t, 0, unreadCountAfterAll.Unread)

		deleteResp := env.DoJSON(
			t,
			http.MethodDelete,
			"/api/v1/notifications/"+generated.ID.String(),
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)

		var deleted model.Notification
		require.NoError(t, env.DB.Unscoped().First(&deleted, "id = ?", generated.ID).Error)
		require.True(t, deleted.DeletedAt.Valid)
	})

	t.Run("user CRUD role update and users me", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		department := model.Department{Name: "Engineering"}
		require.NoError(t, env.DB.Create(&department).Error)

		email := fmt.Sprintf("it-shared-user-%s@example.com", uuid.NewString())
		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/users", map[string]any{
			"email":         email,
			"password":      "password123",
			"first_name":    "Shared",
			"last_name":     "Domain",
			"role":          model.RoleEmployee,
			"department_id": department.ID,
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var created model.User
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
		require.Equal(t, email, created.Email)
		require.Equal(t, model.RoleEmployee, created.Role)

		allResp := env.DoJSON(t, http.MethodGet, "/api/v1/users?page=1&page_size=20", nil, adminHeaders)
		require.Equal(t, http.StatusOK, allResp.Code)
		var users paginatedUsersResponse
		require.NoError(t, json.Unmarshal(allResp.Body.Bytes(), &users))
		require.True(t, containsUser(users.Data, created.ID))

		employeeHeaders := map[string]string{
			"Authorization": env.MustBearerToken(t, created.ID, model.RoleEmployee),
		}
		meResp := env.DoJSON(t, http.MethodGet, "/api/v1/users/me", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, meResp.Code)
		var me model.User
		require.NoError(t, json.Unmarshal(meResp.Body.Bytes(), &me))
		require.Equal(t, created.ID, me.ID)

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/users/"+created.ID.String(),
			map[string]any{
				"first_name": "Updated",
				"role":       model.RoleManager,
			},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)
		var updated model.User
		require.NoError(t, json.Unmarshal(updateResp.Body.Bytes(), &updated))
		require.Equal(t, "Updated", updated.FirstName)
		require.Equal(t, model.RoleManager, updated.Role)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/users/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)

		var deleted model.User
		require.NoError(t, env.DB.Unscoped().First(&deleted, "id = ?", created.ID).Error)
		require.True(t, deleted.DeletedAt.Valid)
	})

	t.Run("department management state reflects create update delete", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, headers := createActorWithHeaders(t, env, model.RoleEmployee)

		createDepartment := model.Department{Name: "Temp Department"}
		updateDepartment := model.Department{Name: "Before Update"}
		require.NoError(t, env.DB.Create(&createDepartment).Error)
		require.NoError(t, env.DB.Create(&updateDepartment).Error)

		require.NoError(t, env.DB.Model(&updateDepartment).Update("name", "After Update").Error)
		require.NoError(t, env.DB.Delete(&createDepartment).Error)

		resp := env.DoJSON(t, http.MethodGet, "/api/v1/departments", nil, headers)
		require.Equal(t, http.StatusOK, resp.Code)

		var departments []model.Department
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &departments))
		require.Len(t, departments, 1)
		require.Equal(t, "After Update", departments[0].Name)
	})

	t.Run("shift weekly create bulk create individual update and delete", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/shifts", map[string]any{
			"user_id":    employee.ID,
			"date":       "2030-02-03",
			"shift_type": model.ShiftTypeDay,
			"note":       "single",
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var created model.Shift
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))

		bulkResp := env.DoJSON(t, http.MethodPost, "/api/v1/shifts/bulk", map[string]any{
			"shifts": []map[string]any{
				{"user_id": employee.ID, "date": "2030-02-04", "shift_type": model.ShiftTypeDay, "note": "d1"},
				{"user_id": employee.ID, "date": "2030-02-05", "shift_type": model.ShiftTypeDay, "note": "d2"},
				{"user_id": employee.ID, "date": "2030-02-06", "shift_type": model.ShiftTypeEvening, "note": "d3"},
				{"user_id": employee.ID, "date": "2030-02-07", "shift_type": model.ShiftTypeOff, "note": "d4"},
			},
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, bulkResp.Code)

		require.NoError(t, env.DB.Model(&model.Shift{}).Where("id = ?", created.ID).Update("note", "updated-single").Error)

		listResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/shifts?start_date=2030-02-03&end_date=2030-02-09",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, listResp.Code)
		var shifts []model.Shift
		require.NoError(t, json.Unmarshal(listResp.Body.Bytes(), &shifts))
		require.Len(t, shifts, 5)
		require.Equal(t, "updated-single", shiftByID(t, shifts, created.ID).Note)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/shifts/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)

		listAfterDeleteResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/shifts?start_date=2030-02-03&end_date=2030-02-09",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, listAfterDeleteResp.Code)
		var afterDelete []model.Shift
		require.NoError(t, json.Unmarshal(listAfterDeleteResp.Body.Bytes(), &afterDelete))
		require.Len(t, afterDelete, 4)
	})

	t.Run("project CRUD filter and cross project search", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)

		createAResp := env.DoJSON(t, http.MethodPost, "/api/v1/projects", map[string]any{
			"name":        "Alpha Project",
			"code":        "ALPHA-001",
			"description": "alpha",
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createAResp.Code)
		var projectA model.Project
		require.NoError(t, json.Unmarshal(createAResp.Body.Bytes(), &projectA))

		createBResp := env.DoJSON(t, http.MethodPost, "/api/v1/projects", map[string]any{
			"name":        "Beta Project",
			"code":        "BETA-001",
			"description": "beta",
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createBResp.Code)
		var projectB model.Project
		require.NoError(t, json.Unmarshal(createBResp.Body.Bytes(), &projectB))

		entryAResp := env.DoJSON(t, http.MethodPost, "/api/v1/time-entries", map[string]any{
			"project_id":  projectA.ID,
			"date":        "2030-03-01",
			"minutes":     120,
			"description": "feature-a",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, entryAResp.Code)

		entryBResp := env.DoJSON(t, http.MethodPost, "/api/v1/time-entries", map[string]any{
			"project_id":  projectB.ID,
			"date":        "2030-03-02",
			"minutes":     60,
			"description": "feature-b",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, entryBResp.Code)

		activeResp := env.DoJSON(t, http.MethodGet, "/api/v1/projects?status=active&page=1&page_size=20", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, activeResp.Code)
		var activeProjects paginatedProjectsResponse
		require.NoError(t, json.Unmarshal(activeResp.Body.Bytes(), &activeProjects))
		require.Equal(t, int64(2), activeProjects.Total)

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/projects/"+projectB.ID.String(),
			map[string]any{"status": model.ProjectStatusInactive},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)
		var updatedProjectB model.Project
		require.NoError(t, json.Unmarshal(updateResp.Body.Bytes(), &updatedProjectB))
		require.Equal(t, model.ProjectStatusInactive, updatedProjectB.Status)

		inactiveResp := env.DoJSON(t, http.MethodGet, "/api/v1/projects?status=inactive&page=1&page_size=20", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, inactiveResp.Code)
		var inactiveProjects paginatedProjectsResponse
		require.NoError(t, json.Unmarshal(inactiveResp.Body.Bytes(), &inactiveProjects))
		require.Equal(t, int64(1), inactiveProjects.Total)
		require.Equal(t, projectB.ID, inactiveProjects.Data[0].ID)

		getByIDResp := env.DoJSON(t, http.MethodGet, "/api/v1/projects/"+projectA.ID.String(), nil, employeeHeaders)
		require.Equal(t, http.StatusOK, getByIDResp.Code)
		var gotProjectA model.Project
		require.NoError(t, json.Unmarshal(getByIDResp.Body.Bytes(), &gotProjectA))
		require.Equal(t, projectA.ID, gotProjectA.ID)

		byProjectResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/projects/"+projectA.ID.String()+"/time-entries?start_date=2030-03-01&end_date=2030-03-31",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, byProjectResp.Code)
		var byProjectEntries []model.TimeEntry
		require.NoError(t, json.Unmarshal(byProjectResp.Body.Bytes(), &byProjectEntries))
		require.Len(t, byProjectEntries, 1)
		require.Equal(t, projectA.ID, byProjectEntries[0].ProjectID)
		require.Equal(t, employee.ID, byProjectEntries[0].UserID)

		summaryResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/time-entries/summary?start_date=2030-03-01&end_date=2030-03-31",
			nil,
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, summaryResp.Code)
		var summaries []model.ProjectSummary
		require.NoError(t, json.Unmarshal(summaryResp.Body.Bytes(), &summaries))
		require.GreaterOrEqual(t, len(summaries), 2)
		require.True(t, containsProjectSummary(summaries, projectA.ID))
		require.True(t, containsProjectSummary(summaries, projectB.ID))

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/projects/"+projectA.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)
		var deletedProject model.Project
		require.NoError(t, env.DB.Unscoped().First(&deletedProject, "id = ?", projectA.ID).Error)
		require.True(t, deletedProject.DeletedAt.Valid)
	})

	t.Run("holiday CRUD duplicate detection year calendar and working days", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/holidays", map[string]any{
			"date":         "2030-05-01",
			"name":         "Integration Holiday",
			"holiday_type": model.HolidayTypeCompany,
			"is_recurring": false,
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var created model.Holiday
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))

		duplicateResp := env.DoJSON(t, http.MethodPost, "/api/v1/holidays", map[string]any{
			"date":         "2030-05-01",
			"name":         "Duplicate",
			"holiday_type": model.HolidayTypeCompany,
			"is_recurring": false,
		}, adminHeaders)
		require.Equal(t, http.StatusBadRequest, duplicateResp.Code)
		assertErrorResponse(t, duplicateResp.Body.Bytes(), http.StatusBadRequest)

		yearResp := env.DoJSON(t, http.MethodGet, "/api/v1/holidays?year=2030", nil, adminHeaders)
		require.Equal(t, http.StatusOK, yearResp.Code)
		var byYear []model.Holiday
		require.NoError(t, json.Unmarshal(yearResp.Body.Bytes(), &byYear))
		require.True(t, containsHoliday(byYear, created.ID))

		calendarResp := env.DoJSON(t, http.MethodGet, "/api/v1/holidays/calendar?year=2030&month=5", nil, adminHeaders)
		require.Equal(t, http.StatusOK, calendarResp.Code)
		var calendar []model.CalendarDay
		require.NoError(t, json.Unmarshal(calendarResp.Body.Bytes(), &calendar))
		require.NotEmpty(t, calendar)

		holidayDay := findCalendarDay(calendar, "2030-05-01")
		require.NotNil(t, holidayDay)
		require.True(t, holidayDay.IsHoliday)

		workingDaysResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/holidays/working-days?start_date=2030-05-01&end_date=2030-05-07",
			nil,
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, workingDaysResp.Code)
		var working model.WorkingDaysSummary
		require.NoError(t, json.Unmarshal(workingDaysResp.Body.Bytes(), &working))
		require.Equal(t, 7, working.TotalDays)
		require.GreaterOrEqual(t, working.Holidays, 1)

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/holidays/"+created.ID.String(),
			map[string]any{"name": "Integration Holiday Updated"},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)
		var updated model.Holiday
		require.NoError(t, json.Unmarshal(updateResp.Body.Bytes(), &updated))
		require.Equal(t, "Integration Holiday Updated", updated.Name)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/holidays/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)
		var deleted model.Holiday
		require.NoError(t, env.DB.Unscoped().First(&deleted, "id = ?", created.ID).Error)
		require.True(t, deleted.DeletedAt.Valid)
	})

	t.Run("approval flow CRUD and status transition", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/approval-flows", map[string]any{
			"name":      "Leave Flow",
			"flow_type": model.ApprovalFlowLeave,
			"steps": []map[string]any{
				{"step_order": 1, "step_type": model.ApprovalStepRole, "approver_role": model.RoleManager},
			},
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var created model.ApprovalFlow
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
		require.True(t, created.IsActive)
		require.Len(t, created.Steps, 1)

		allResp := env.DoJSON(t, http.MethodGet, "/api/v1/approval-flows", nil, adminHeaders)
		require.Equal(t, http.StatusOK, allResp.Code)
		var all []model.ApprovalFlow
		require.NoError(t, json.Unmarshal(allResp.Body.Bytes(), &all))
		require.True(t, containsApprovalFlow(all, created.ID))

		byIDResp := env.DoJSON(t, http.MethodGet, "/api/v1/approval-flows/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusOK, byIDResp.Code)
		var byID model.ApprovalFlow
		require.NoError(t, json.Unmarshal(byIDResp.Body.Bytes(), &byID))
		require.Equal(t, created.ID, byID.ID)

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/approval-flows/"+created.ID.String(),
			map[string]any{
				"name":      "Leave Flow Updated",
				"is_active": false,
				"steps": []map[string]any{
					{"step_order": 1, "step_type": model.ApprovalStepManager},
					{"step_order": 2, "step_type": model.ApprovalStepRole, "approver_role": model.RoleAdmin},
				},
			},
			adminHeaders,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)
		var updated model.ApprovalFlow
		require.NoError(t, json.Unmarshal(updateResp.Body.Bytes(), &updated))
		require.Equal(t, "Leave Flow Updated", updated.Name)
		require.False(t, updated.IsActive)
		require.Len(t, updated.Steps, 2)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/approval-flows/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNoContent, deleteResp.Code)

		getAfterDeleteResp := env.DoJSON(t, http.MethodGet, "/api/v1/approval-flows/"+created.ID.String(), nil, adminHeaders)
		require.Equal(t, http.StatusNotFound, getAfterDeleteResp.Code)
	})

	t.Run("export attendance leaves overtime projects as csv", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		employee, _ := createActorWithHeaders(t, env, model.RoleEmployee)

		attendanceDate := mustDate(t, "2030-06-10")
		clockIn := time.Date(2030, 6, 10, 9, 0, 0, 0, time.Local)
		clockOut := time.Date(2030, 6, 10, 19, 0, 0, 0, time.Local)
		attendance := model.Attendance{
			UserID:          employee.ID,
			Date:            attendanceDate,
			ClockIn:         &clockIn,
			ClockOut:        &clockOut,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     600,
			OvertimeMinutes: 120,
		}
		require.NoError(t, env.DB.Create(&attendance).Error)

		leave := model.LeaveRequest{
			UserID:    employee.ID,
			LeaveType: model.LeaveTypePaid,
			StartDate: mustDate(t, "2030-06-11"),
			EndDate:   mustDate(t, "2030-06-11"),
			Reason:    "vacation",
			Status:    model.ApprovalStatusPending,
		}
		require.NoError(t, env.DB.Create(&leave).Error)

		project := model.Project{
			Name:   "Export Project",
			Code:   "EXPORT-001",
			Status: model.ProjectStatusActive,
		}
		require.NoError(t, env.DB.Create(&project).Error)

		timeEntry := model.TimeEntry{
			UserID:    employee.ID,
			ProjectID: project.ID,
			Date:      mustDate(t, "2030-06-12"),
			Minutes:   180,
		}
		require.NoError(t, env.DB.Create(&timeEntry).Error)

		assertCSVDownload(
			t,
			env,
			"/api/v1/export/attendance?start_date=2030-06-01&end_date=2030-06-30",
			adminHeaders,
			"attendance.csv",
		)
		assertCSVDownload(
			t,
			env,
			"/api/v1/export/leaves?start_date=2030-06-01&end_date=2030-06-30",
			adminHeaders,
			"leaves.csv",
		)
		assertCSVDownload(
			t,
			env,
			"/api/v1/export/overtime?start_date=2030-06-01&end_date=2030-06-30",
			adminHeaders,
			"overtime.csv",
		)
		assertCSVDownload(
			t,
			env,
			"/api/v1/export/projects?start_date=2030-06-01&end_date=2030-06-30",
			adminHeaders,
			"projects.csv",
		)
	})

	t.Run("dashboard stats aggregation", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		employee, _ := createActorWithHeaders(t, env, model.RoleEmployee)

		attendance := model.Attendance{
			UserID:          employee.ID,
			Date:            time.Now(),
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     600,
			OvertimeMinutes: 120,
		}
		nowIn := time.Now()
		nowOut := nowIn.Add(10 * time.Hour)
		attendance.ClockIn = &nowIn
		attendance.ClockOut = &nowOut
		require.NoError(t, env.DB.Create(&attendance).Error)

		pendingLeave := model.LeaveRequest{
			UserID:    employee.ID,
			LeaveType: model.LeaveTypePaid,
			StartDate: time.Now(),
			EndDate:   time.Now(),
			Reason:    "pending",
			Status:    model.ApprovalStatusPending,
		}
		require.NoError(t, env.DB.Create(&pendingLeave).Error)

		resp := env.DoJSON(t, http.MethodGet, "/api/v1/dashboard/stats", nil, adminHeaders)
		require.Equal(t, http.StatusOK, resp.Code)

		var stats model.DashboardStatsExtended
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &stats))
		require.Equal(t, 1, stats.TodayPresentCount)
		require.Equal(t, 1, stats.PendingLeaves)
		require.GreaterOrEqual(t, stats.MonthlyOvertime, 120)
		require.Len(t, stats.WeeklyTrend, 7)
	})
}

func containsUser(users []model.User, id uuid.UUID) bool {
	for _, user := range users {
		if user.ID == id {
			return true
		}
	}
	return false
}

func shiftByID(t *testing.T, shifts []model.Shift, id uuid.UUID) model.Shift {
	t.Helper()
	for _, shift := range shifts {
		if shift.ID == id {
			return shift
		}
	}
	t.Fatalf("shift not found: %s", id)
	return model.Shift{}
}

func containsProjectSummary(summaries []model.ProjectSummary, projectID uuid.UUID) bool {
	for _, summary := range summaries {
		if summary.ProjectID == projectID {
			return true
		}
	}
	return false
}

func containsHoliday(holidays []model.Holiday, id uuid.UUID) bool {
	for _, holiday := range holidays {
		if holiday.ID == id {
			return true
		}
	}
	return false
}

func findCalendarDay(days []model.CalendarDay, date string) *model.CalendarDay {
	for idx := range days {
		if days[idx].Date == date {
			return &days[idx]
		}
	}
	return nil
}

func containsApprovalFlow(flows []model.ApprovalFlow, id uuid.UUID) bool {
	for _, flow := range flows {
		if flow.ID == id {
			return true
		}
	}
	return false
}

func assertCSVDownload(t *testing.T, env *TestEnv, path string, headers map[string]string, fileName string) {
	t.Helper()

	body, respHeaders, code := env.DoDownload(t, http.MethodGet, path, headers)
	require.Equal(t, http.StatusOK, code)
	require.Contains(t, respHeaders.Get("Content-Type"), "text/csv")
	require.Contains(t, respHeaders.Get("Content-Disposition"), fileName)
	require.GreaterOrEqual(t, len(body), 3)
	require.Equal(t, byte(0xEF), body[0])
	require.Equal(t, byte(0xBB), body[1])
	require.Equal(t, byte(0xBF), body[2])
}
