package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/datatypes"
)

func TestHRDomainIntegration(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, model.HRAutoMigrate(env.DB))
	t.Cleanup(func() { _ = os.RemoveAll("uploads") })

	t.Run("hr dashboard stats and activities", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleAdmin)
		seedHREmployee(t, env, uuid.New(), "HR-ACTIVE", nil, model.EmployeeStatusActive)
		seedHREmployee(t, env, uuid.New(), "HR-LEAVE", nil, model.EmployeeStatusOnLeave)

		statsResp := env.DoJSON(t, http.MethodGet, "/api/v1/hr/stats", nil, headers)
		require.Equal(t, http.StatusOK, statsResp.Code)

		activitiesResp := env.DoJSON(t, http.MethodGet, "/api/v1/hr/activities", nil, headers)
		require.Equal(t, http.StatusOK, activitiesResp.Code)
	})

	t.Run("employee CRUD and filter search", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleAdmin)
		dept := seedHRDepartment(t, env, "Engineering", "ENG", nil, 5000000)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/employees", map[string]any{
			"employee_code": uniqueCode("EMP"),
			"first_name":    "Hanako",
			"last_name":     "Yamada",
			"email":         "hr-employee@example.com",
			"department_id": dept.ID,
			"position":      "Developer",
			"base_salary":   420000,
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var employee model.HREmployee
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &employee))

		listResp := env.DoJSON(t, http.MethodGet, "/api/v1/hr/employees?search=Hanako&page=1&page_size=20", nil, headers)
		require.Equal(t, http.StatusOK, listResp.Code)

		getResp := env.DoJSON(t, http.MethodGet, "/api/v1/hr/employees/"+employee.ID.String(), nil, headers)
		require.Equal(t, http.StatusOK, getResp.Code)

		updateResp := env.DoJSON(t, http.MethodPut, "/api/v1/hr/employees/"+employee.ID.String(), map[string]any{
			"status":   model.EmployeeStatusInactive,
			"position": "Senior Developer",
		}, headers)
		require.Equal(t, http.StatusOK, updateResp.Code)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/hr/employees/"+employee.ID.String(), nil, headers)
		require.Equal(t, http.StatusOK, deleteResp.Code)
	})

	t.Run("department CRUD with hierarchy", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleAdmin)

		parentResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/departments", map[string]any{
			"name": "Corporate", "code": "CORP",
		}, headers)
		require.Equal(t, http.StatusCreated, parentResp.Code)
		var parent model.HRDepartment
		require.NoError(t, json.Unmarshal(parentResp.Body.Bytes(), &parent))

		childResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/departments", map[string]any{
			"name": "Corporate IT", "code": "CORP-IT", "parent_id": parent.ID,
		}, headers)
		require.Equal(t, http.StatusCreated, childResp.Code)
		var child model.HRDepartment
		require.NoError(t, json.Unmarshal(childResp.Body.Bytes(), &child))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/departments", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/departments/"+child.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/departments/"+child.ID.String(), map[string]any{"name": "Corporate Platform"}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/departments/"+child.ID.String(), nil, headers).Code)
	})

	t.Run("evaluation CRUD submit and cycle CRUD", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		employee := seedHREmployee(t, env, uuid.New(), "HR-EVAL-EMP", nil, model.EmployeeStatusActive)

		cycleResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/evaluation-cycles", map[string]any{
			"name": "2026 Annual", "start_date": "2026-01-01", "end_date": "2026-12-31",
		}, headers)
		require.Equal(t, http.StatusCreated, cycleResp.Code)
		var cycle model.EvaluationCycle
		require.NoError(t, json.Unmarshal(cycleResp.Body.Bytes(), &cycle))

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/evaluations", map[string]any{
			"employee_id": employee.ID, "cycle_id": cycle.ID, "self_score": 3.5,
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var evaluation model.Evaluation
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &evaluation))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/evaluations?cycle_id="+cycle.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/evaluations/"+evaluation.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/evaluations/"+evaluation.ID.String(), map[string]any{"manager_score": 4.2}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/evaluations/"+evaluation.ID.String()+"/submit", nil, headers).Code)
	})

	t.Run("goal CRUD and progress update", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		employee := seedHREmployee(t, env, uuid.New(), "HR-GOAL-EMP", nil, model.EmployeeStatusActive)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/goals", map[string]any{
			"employee_id": employee.ID, "title": "Integration Goal", "due_date": "2026-09-30",
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var goal model.HRGoal
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &goal))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/goals?employee_id="+employee.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/goals/"+goal.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/goals/"+goal.ID.String(), map[string]any{"title": "Integration Goal Updated"}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/goals/"+goal.ID.String()+"/progress", map[string]any{"progress": 60}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/goals/"+goal.ID.String(), nil, headers).Code)
	})

	t.Run("training CRUD with enroll and complete", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, adminHeaders := createActorWithHeaders(t, env, model.RoleAdmin)
		trainee, traineeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		seedHREmployee(t, env, trainee.ID, "HR-TRAINEE", nil, model.EmployeeStatusActive)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/training", map[string]any{
			"title": "Security Basics", "start_date": "2026-05-01", "end_date": "2026-05-05",
		}, adminHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var program model.TrainingProgram
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &program))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/training?page=1&page_size=20", nil, adminHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/training/"+program.ID.String(), nil, adminHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/training/"+program.ID.String(), map[string]any{"status": model.TrainingStatusInProgress}, adminHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPost, "/api/v1/hr/training/"+program.ID.String()+"/enroll", nil, traineeHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/training/"+program.ID.String()+"/complete", nil, traineeHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/training/"+program.ID.String(), nil, adminHeaders).Code)
	})

	t.Run("recruitment positions applicants and stage update", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		dept := seedHRDepartment(t, env, "Talent", "TAL", nil, 4000000)

		positionResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/positions", map[string]any{
			"title": "Backend Engineer", "department_id": dept.ID, "openings": 2,
		}, headers)
		require.Equal(t, http.StatusCreated, positionResp.Code)
		var position model.RecruitmentPosition
		require.NoError(t, json.Unmarshal(positionResp.Body.Bytes(), &position))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/positions?department=Talent&page=1&page_size=20", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/positions/"+position.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/positions/"+position.ID.String(), map[string]any{"status": model.PositionStatusOnHold}, headers).Code)

		applicantResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/applicants", map[string]any{
			"position_id": position.ID, "name": "Applicant A", "email": "applicant@example.com",
		}, headers)
		require.Equal(t, http.StatusCreated, applicantResp.Code)
		var applicant model.Applicant
		require.NoError(t, json.Unmarshal(applicantResp.Body.Bytes(), &applicant))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/applicants?position_id="+position.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/applicants/"+applicant.ID.String()+"/stage", map[string]any{"stage": model.ApplicantStageInterview}, headers).Code)
	})

	t.Run("document upload list delete and download", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleEmployee)

		uploadResp := env.DoMultipart(
			t,
			http.MethodPost,
			"/api/v1/hr/documents",
			map[string]string{"title": "Offer Letter", "type": "offer"},
			map[string]MultipartFile{"file": {FileName: "offer.txt", Content: []byte("offer-content")}},
			headers,
		)
		require.Equal(t, http.StatusCreated, uploadResp.Code)
		var document model.HRDocument
		require.NoError(t, json.Unmarshal(uploadResp.Body.Bytes(), &document))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/documents?type=offer&page=1&page_size=20", nil, headers).Code)
		_, _, code := env.DoDownload(t, http.MethodGet, "/api/v1/hr/documents/"+document.ID.String()+"/download", headers)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/documents/"+document.ID.String(), nil, headers).Code)
	})

	t.Run("announcement CRUD", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		author, headers := createActorWithHeaders(t, env, model.RoleManager)
		seedHREmployee(t, env, author.ID, "HR-AUTHOR", nil, model.EmployeeStatusActive)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/announcements", map[string]any{
			"title": "Maintenance", "content": "maintenance window",
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var announcement model.HRAnnouncement
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &announcement))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/announcements?page=1&page_size=20", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/announcements/"+announcement.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/announcements/"+announcement.ID.String(), map[string]any{"is_published": true}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/announcements/"+announcement.ID.String(), nil, headers).Code)
	})

	t.Run("attendance integration endpoints", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		userPresent, headers := createActorWithHeaders(t, env, model.RoleAdmin)
		userLeave, _ := createActorWithHeaders(t, env, model.RoleEmployee)
		seedHREmployee(t, env, userPresent.ID, "HR-PRESENT", nil, model.EmployeeStatusActive)
		seedHREmployee(t, env, userLeave.ID, "HR-LEAVE-2", nil, model.EmployeeStatusActive)

		now := time.Now()
		day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		in := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, time.Local)
		out := in.Add(10 * time.Hour)
		require.NoError(t, env.DB.Create(&model.Attendance{
			UserID:          userPresent.ID,
			Date:            day,
			ClockIn:         &in,
			ClockOut:        &out,
			Status:          model.AttendanceStatusPresent,
			WorkMinutes:     600,
			OvertimeMinutes: 2500,
		}).Error)
		require.NoError(t, env.DB.Create(&model.LeaveRequest{
			UserID: userLeave.ID, LeaveType: model.LeaveTypePaid, StartDate: day, EndDate: day, Status: model.ApprovalStatusApproved,
		}).Error)

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/attendance-integration", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/attendance-integration/alerts", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/attendance-integration/trend?period=week", nil, headers).Code)
	})

	t.Run("org chart and simulation", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		parent := seedHRDepartment(t, env, "Delivery", "DLY", nil, 2000000)
		target := seedHRDepartment(t, env, "Product", "PRD", nil, 3000000)
		employee := seedHREmployee(t, env, uuid.New(), "HR-ORG-EMP", &parent.ID, model.EmployeeStatusActive)

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/org-chart", nil, headers).Code)
		simResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/org-chart/simulate", map[string]any{
			"moves": []map[string]any{
				{"employee_id": employee.ID.String(), "from_department_id": parent.ID.String(), "to_department_id": target.ID.String()},
			},
		}, headers)
		require.Equal(t, http.StatusOK, simResp.Code)
	})

	t.Run("one on one CRUD action add and toggle", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		manager, managerHeaders := createActorWithHeaders(t, env, model.RoleManager)
		seedHREmployee(t, env, manager.ID, "HR-MANAGER", nil, model.EmployeeStatusActive)
		employee := seedHREmployee(t, env, uuid.New(), "HR-1ON1", nil, model.EmployeeStatusActive)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/one-on-ones", map[string]any{
			"employee_id": employee.ID.String(), "scheduled_date": "2026-06-10", "agenda": "weekly sync",
		}, managerHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var meeting model.OneOnOneMeeting
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &meeting))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/one-on-ones", nil, managerHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/one-on-ones/"+meeting.ID.String(), nil, managerHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/one-on-ones/"+meeting.ID.String(), map[string]any{"status": "done"}, managerHeaders).Code)

		addActionResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/one-on-ones/"+meeting.ID.String()+"/actions", map[string]any{"title": "Follow up"}, managerHeaders)
		require.Equal(t, http.StatusOK, addActionResp.Code)
		var meetingWithAction model.OneOnOneMeeting
		require.NoError(t, json.Unmarshal(addActionResp.Body.Bytes(), &meetingWithAction))
		actionID := firstActionID(t, meetingWithAction.ActionItems)
		require.NotEmpty(t, actionID)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/one-on-ones/"+meeting.ID.String()+"/actions/"+actionID+"/toggle", nil, managerHeaders).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/one-on-ones/"+meeting.ID.String(), nil, managerHeaders).Code)
	})

	t.Run("skill map gap analysis add and update", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		employee := seedHREmployee(t, env, uuid.New(), "HR-SKILL", nil, model.EmployeeStatusActive)

		addResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/skill-map/"+employee.ID.String(), map[string]any{
			"skill_name": "Go", "category": "technical", "level": 2,
		}, headers)
		require.Equal(t, http.StatusCreated, addResp.Code)
		var skill model.EmployeeSkill
		require.NoError(t, json.Unmarshal(addResp.Body.Bytes(), &skill))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/skill-map?employee_id="+employee.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/skill-map/"+employee.ID.String()+"/"+skill.ID.String(), map[string]any{"level": 4}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/skill-map/gap-analysis", nil, headers).Code)
	})

	t.Run("salary overview simulate history and budget", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleAdmin)
		dept := seedHRDepartment(t, env, "Finance", "FIN", nil, 9000000)
		employee := seedHREmployee(t, env, uuid.New(), "HR-SALARY", &dept.ID, model.EmployeeStatusActive)
		require.NoError(t, env.DB.Create(&model.SalaryRecord{
			EmployeeID: employee.ID, BaseSalary: 500000, NetSalary: 500000, EffectiveDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local),
		}).Error)

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/salary", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPost, "/api/v1/hr/salary/simulate", map[string]any{"grade": "M1", "years_of_service": "3"}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/salary/"+employee.ID.String()+"/history", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/salary/budget", nil, headers).Code)
	})

	t.Run("onboarding CRUD template CRUD and task toggle", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		employee := seedHREmployee(t, env, uuid.New(), "HR-ONB", nil, model.EmployeeStatusActive)

		templateResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/onboarding/templates", map[string]any{
			"name": "Engineer Onboarding", "tasks": []map[string]any{{"id": "t-1", "completed": false}},
		}, headers)
		require.Equal(t, http.StatusCreated, templateResp.Code)
		var template model.OnboardingTemplate
		require.NoError(t, json.Unmarshal(templateResp.Body.Bytes(), &template))

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/onboarding", map[string]any{
			"employee_id": employee.ID.String(), "template_id": template.ID.String(), "start_date": "2026-07-01",
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var onboarding model.Onboarding
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &onboarding))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/onboarding", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/onboarding/templates", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/onboarding/"+onboarding.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/onboarding/"+onboarding.ID.String(), map[string]any{"status": model.OnboardingStatusInProgress}, headers).Code)
		require.NoError(t, env.DB.Model(&model.Onboarding{}).Where("id = ?", onboarding.ID).Update("tasks", datatypes.JSON([]byte(`[{"id":"task-1","completed":false}]`))).Error)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/onboarding/"+onboarding.ID.String()+"/tasks/task-1/toggle", nil, headers).Code)
	})

	t.Run("offboarding CRUD completion and checklist toggle", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		_, headers := createActorWithHeaders(t, env, model.RoleManager)
		employee := seedHREmployee(t, env, uuid.New(), "HR-OFF", nil, model.EmployeeStatusActive)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/offboarding", map[string]any{
			"employee_id": employee.ID.String(), "last_working_date": "2026-08-31", "reason": "resignation",
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var offboarding model.Offboarding
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &offboarding))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/offboarding", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/offboarding/"+offboarding.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/offboarding/"+offboarding.ID.String(), map[string]any{"status": model.OffboardingStatusCompleted}, headers).Code)
		require.NoError(t, env.DB.Model(&model.Offboarding{}).Where("id = ?", offboarding.ID).Update("checklist", datatypes.JSON([]byte(`[{"key":"asset_return","completed":false}]`))).Error)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/offboarding/"+offboarding.ID.String()+"/checklist/asset_return/toggle", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/offboarding/analytics", nil, headers).Code)
	})

	t.Run("survey CRUD publish close results and respond", func(t *testing.T) {
		require.NoError(t, env.ResetDB())
		user, headers := createActorWithHeaders(t, env, model.RoleAdmin)
		seedHREmployee(t, env, user.ID, "HR-SURVEY", nil, model.EmployeeStatusActive)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/hr/surveys", map[string]any{
			"title": "Engagement Survey", "questions": []map[string]any{{"id": "q1", "type": "rating"}},
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var survey model.Survey
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &survey))

		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/surveys", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/surveys/"+survey.ID.String(), nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/surveys/"+survey.ID.String(), map[string]any{"description": "updated"}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/surveys/"+survey.ID.String()+"/publish", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPost, "/api/v1/hr/surveys/"+survey.ID.String()+"/respond", map[string]any{"answers": []map[string]any{{"question_id": "q1", "value": 5}}}, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodGet, "/api/v1/hr/surveys/"+survey.ID.String()+"/results", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodPut, "/api/v1/hr/surveys/"+survey.ID.String()+"/close", nil, headers).Code)
		require.Equal(t, http.StatusOK, env.DoJSON(t, http.MethodDelete, "/api/v1/hr/surveys/"+survey.ID.String(), nil, headers).Code)
	})
}

func uniqueCode(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString()[:8])
}

func seedHRDepartment(t *testing.T, env *TestEnv, name, code string, parentID *uuid.UUID, budget float64) *model.HRDepartment {
	t.Helper()
	dept := &model.HRDepartment{Name: name, Code: code, ParentID: parentID, Budget: budget}
	require.NoError(t, env.DB.Create(dept).Error)
	return dept
}

func seedHREmployee(t *testing.T, env *TestEnv, id uuid.UUID, codePrefix string, departmentID *uuid.UUID, status model.EmployeeStatus) *model.HREmployee {
	t.Helper()
	hireDate := time.Now().AddDate(0, 0, -10)
	employee := &model.HREmployee{
		BaseModel:      model.BaseModel{ID: id},
		EmployeeCode:   uniqueCode(codePrefix),
		FirstName:      "HR",
		LastName:       "User",
		Email:          fmt.Sprintf("%s-%s@example.com", codePrefix, uuid.NewString()[:8]),
		DepartmentID:   departmentID,
		EmploymentType: model.EmploymentTypeFullTime,
		Status:         status,
		HireDate:       &hireDate,
		BaseSalary:     350000,
	}
	require.NoError(t, env.DB.Create(employee).Error)
	return employee
}

func firstActionID(t *testing.T, actions datatypes.JSON) string {
	t.Helper()
	var items []map[string]any
	require.NoError(t, json.Unmarshal(actions, &items))
	if len(items) == 0 {
		return ""
	}
	id, _ := items[0]["id"].(string)
	return id
}
