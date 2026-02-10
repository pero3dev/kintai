package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"gorm.io/datatypes"
)

// =============================================================================
// HREmployeeService branch coverage
// =============================================================================

func TestBranch_HREmployee_Create(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHREmployeeService(deps)

	t.Run("with_employment_type_set", func(t *testing.T) {
		got, err := svc.Create(ctx, model.HREmployeeCreateRequest{
			EmployeeCode:   "BC-001",
			EmploymentType: "part_time",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.EmploymentType != model.EmploymentType("part_time") {
			t.Fatalf("expected part_time, got %s", got.EmploymentType)
		}
	})

	t.Run("without_dates", func(t *testing.T) {
		got, err := svc.Create(ctx, model.HREmployeeCreateRequest{EmployeeCode: "BC-002"})
		if err != nil {
			t.Fatal(err)
		}
		if got.HireDate != nil || got.BirthDate != nil {
			t.Fatal("expected nil dates")
		}
	})
}

func TestBranch_HREmployee_Update_NoFields(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHREmployeeService(deps)

	id := uuid.New()
	r.hrEmployee.items[id] = &model.HREmployee{
		BaseModel: model.BaseModel{ID: id},
		FirstName: "Keep", LastName: "Same",
	}

	got, err := svc.Update(ctx, id, model.HREmployeeUpdateRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got.FirstName != "Keep" {
		t.Fatal("fields should remain unchanged")
	}
}

// =============================================================================
// HRDepartmentService branch coverage
// =============================================================================

func TestBranch_HRDepartment_CreateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHRDepartmentService(deps)

	r.hrDepartment.createErr = errors.New("create error")
	if _, err := svc.Create(ctx, model.HRDepartmentCreateRequest{Name: "X"}); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_HRDepartment_Update_ParentAndManager(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHRDepartmentService(deps)

	id := uuid.New()
	r.hrDepartment.items[id] = &model.HRDepartment{BaseModel: model.BaseModel{ID: id}, Name: "Dept"}

	parentID := uuid.New()
	managerID := uuid.New()
	got, err := svc.Update(ctx, id, model.HRDepartmentUpdateRequest{
		ParentID:  &parentID,
		ManagerID: &managerID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ParentID == nil || *got.ParentID != parentID {
		t.Fatal("ParentID not updated")
	}
	if got.ManagerID == nil || *got.ManagerID != managerID {
		t.Fatal("ManagerID not updated")
	}
}

// =============================================================================
// EvaluationService branch coverage
// =============================================================================

func TestBranch_Evaluation_CreateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	r.evaluation.createErr = errors.New("create error")
	if _, err := svc.Create(ctx, model.EvaluationCreateRequest{}, uuid.New()); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_Evaluation_Update_AllFields(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	id := uuid.New()
	r.evaluation.evals[id] = &model.Evaluation{
		BaseModel: model.BaseModel{ID: id},
		Status:    model.EvaluationStatusDraft,
	}

	score := 4.5
	comment := "manager comment"
	goals := "new goals"
	got, err := svc.Update(ctx, id, model.EvaluationUpdateRequest{
		ManagerComment: &comment,
		Goals:          &goals,
		FinalScore:     &score,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ManagerComment != comment || got.Goals != goals {
		t.Fatal("ManagerComment/Goals not updated")
	}
}

func TestBranch_Evaluation_Update_FindError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	r.evaluation.findByIDErr = errors.New("find error")
	if _, err := svc.Update(ctx, uuid.New(), model.EvaluationUpdateRequest{}); err == nil {
		t.Fatal("expected find error")
	}
}

func TestBranch_Evaluation_Update_UpdateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	id := uuid.New()
	r.evaluation.evals[id] = &model.Evaluation{BaseModel: model.BaseModel{ID: id}}
	r.evaluation.updateErr = errors.New("update error")
	if _, err := svc.Update(ctx, id, model.EvaluationUpdateRequest{}); err == nil {
		t.Fatal("expected update error")
	}
}

func TestBranch_Evaluation_Submit_FindError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	r.evaluation.findByIDErr = errors.New("find error")
	if _, err := svc.Submit(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
}

func TestBranch_Evaluation_CreateCycle_Success(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	c, err := svc.CreateCycle(ctx, model.EvaluationCycleCreateRequest{
		Name:      "2025H1",
		StartDate: "2025-01-01",
		EndDate:   "2025-06-30",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "2025H1" {
		t.Fatal("cycle name mismatch")
	}
}

// =============================================================================
// GoalService branch coverage
// =============================================================================

func TestBranch_Goal_Create_AllBranches(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewGoalService(deps)

	t.Run("with_employee_id_and_all_fields", func(t *testing.T) {
		empID := uuid.New()
		got, err := svc.Create(ctx, model.HRGoalCreateRequest{
			Title:      "Goal X",
			EmployeeID: &empID,
			Category:   "development",
			Weight:     5,
			StartDate:  "2025-01-01",
			DueDate:    "2025-06-30",
		}, uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		if got.EmployeeID != empID {
			t.Fatal("EmployeeID not set from request")
		}
		if got.Category != model.GoalCategory("development") {
			t.Fatal("category should be development")
		}
		if got.Weight != 5 {
			t.Fatal("weight should be 5")
		}
		if got.StartDate == nil || got.DueDate == nil {
			t.Fatal("dates should be parsed")
		}
	})

	t.Run("create_error", func(t *testing.T) {
		deps2, r2 := setupHRServiceDeps(t)
		svc2 := NewGoalService(deps2)
		r2.goal.createErr = errors.New("create error")
		if _, err := svc2.Create(ctx, model.HRGoalCreateRequest{Title: "X"}, uuid.New()); err == nil {
			t.Fatal("expected create error")
		}
	})
}

func TestBranch_Goal_Update_AllFields(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewGoalService(deps)

	id := uuid.New()
	r.goal.items[id] = &model.HRGoal{BaseModel: model.BaseModel{ID: id}, Title: "Old"}

	title := "New"
	desc := "desc"
	cat := "behavior"
	status := "in_progress"
	dueDate := "2025-12-31"
	weight := 3
	got, err := svc.Update(ctx, id, model.HRGoalUpdateRequest{
		Title:       &title,
		Description: &desc,
		Category:    &cat,
		Status:      &status,
		DueDate:     &dueDate,
		Weight:      &weight,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "New" || got.Description != "desc" || got.Weight != 3 {
		t.Fatal("fields not updated")
	}
	if got.DueDate == nil {
		t.Fatal("DueDate should be parsed")
	}
}

func TestBranch_Goal_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewGoalService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.goal.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), model.HRGoalUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.goal.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.goal.items[id] = &model.HRGoal{BaseModel: model.BaseModel{ID: id}}
		r.goal.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, model.HRGoalUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.goal.updateErr = nil
	})
}

func TestBranch_Goal_UpdateProgress_ZeroAndFindError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewGoalService(deps)

	t.Run("progress_zero_no_status_change", func(t *testing.T) {
		id := uuid.New()
		r.goal.items[id] = &model.HRGoal{
			BaseModel: model.BaseModel{ID: id},
			Status:    model.GoalStatusNotStarted,
		}
		got, err := svc.UpdateProgress(ctx, id, 0)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != model.GoalStatusNotStarted {
			t.Fatalf("status should remain not_started, got %s", got.Status)
		}
	})

	t.Run("find_error", func(t *testing.T) {
		r.goal.findByIDErr = errors.New("find error")
		if _, err := svc.UpdateProgress(ctx, uuid.New(), 50); err == nil {
			t.Fatal("expected find error")
		}
		r.goal.findByIDErr = nil
	})
}

// =============================================================================
// TrainingService branch coverage
// =============================================================================

func TestBranch_Training_Create_Branches(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewTrainingService(deps)

	t.Run("without_dates", func(t *testing.T) {
		got, err := svc.Create(ctx, model.TrainingProgramCreateRequest{Title: "Training"})
		if err != nil {
			t.Fatal(err)
		}
		if got.StartDate != nil || got.EndDate != nil {
			t.Fatal("expected nil dates when not provided")
		}
	})

	t.Run("create_error", func(t *testing.T) {
		r.training.createErr = errors.New("create error")
		if _, err := svc.Create(ctx, model.TrainingProgramCreateRequest{Title: "X"}); err == nil {
			t.Fatal("expected create error")
		}
		r.training.createErr = nil
	})
}

func TestBranch_Training_Update_AllFields(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewTrainingService(deps)

	id := uuid.New()
	r.training.programs[id] = &model.TrainingProgram{BaseModel: model.BaseModel{ID: id}, Title: "Old"}

	title := "New"
	desc := "desc"
	cat := "tech"
	instructor := "Alice"
	status := "completed"
	maxP := 30
	loc := "Tokyo"
	isOnline := true
	got, err := svc.Update(ctx, id, model.TrainingProgramUpdateRequest{
		Title:           &title,
		Description:     &desc,
		Category:        &cat,
		InstructorName:  &instructor,
		Status:          &status,
		MaxParticipants: &maxP,
		Location:        &loc,
		IsOnline:        &isOnline,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "New" || got.Category != "tech" || got.MaxParticipants != 30 || !got.IsOnline {
		t.Fatal("fields not fully updated")
	}
}

func TestBranch_Training_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewTrainingService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.training.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), model.TrainingProgramUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.training.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.training.programs[id] = &model.TrainingProgram{BaseModel: model.BaseModel{ID: id}}
		r.training.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, model.TrainingProgramUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.training.updateErr = nil
	})
}

func TestBranch_Training_Complete_UpdateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewTrainingService(deps)

	pid := uuid.New()
	empID := uuid.New()
	r.training.enrollments[enrollmentKey(pid, empID)] = &model.TrainingEnrollment{
		BaseModel: model.BaseModel{ID: uuid.New()}, ProgramID: pid, EmployeeID: empID, Status: "enrolled",
	}
	r.training.updateEnrollErr = errors.New("update error")
	if err := svc.Complete(ctx, pid, empID); err == nil {
		t.Fatal("expected update enrollment error")
	}
}

// =============================================================================
// RecruitmentService branch coverage
// =============================================================================

func TestBranch_Recruitment_CreatePosition_Branches(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewRecruitmentService(deps)

	t.Run("with_openings", func(t *testing.T) {
		got, err := svc.CreatePosition(ctx, model.PositionCreateRequest{Title: "Dev", Openings: 3})
		if err != nil {
			t.Fatal(err)
		}
		if got.Openings != 3 {
			t.Fatalf("expected openings=3, got %d", got.Openings)
		}
	})

	t.Run("create_error", func(t *testing.T) {
		r.recruitment.createPositionErr = errors.New("create error")
		if _, err := svc.CreatePosition(ctx, model.PositionCreateRequest{Title: "X"}); err == nil {
			t.Fatal("expected create error")
		}
		r.recruitment.createPositionErr = nil
	})
}

func TestBranch_Recruitment_UpdatePosition_AllFields(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewRecruitmentService(deps)

	id := uuid.New()
	r.recruitment.positions[id] = &model.RecruitmentPosition{BaseModel: model.BaseModel{ID: id}, Title: "Old"}

	title := "New"
	deptID := uuid.New()
	desc := "desc"
	reqs := "reqs"
	status := "closed"
	openings := 5
	loc := "Osaka"
	salMin := 300000.0
	salMax := 500000.0
	got, err := svc.UpdatePosition(ctx, id, model.PositionUpdateRequest{
		Title:        &title,
		DepartmentID: &deptID,
		Description:  &desc,
		Requirements: &reqs,
		Status:       &status,
		Openings:     &openings,
		Location:     &loc,
		SalaryMin:    &salMin,
		SalaryMax:    &salMax,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "New" || got.Openings != 5 || *got.SalaryMin != 300000 {
		t.Fatal("fields not fully updated")
	}
}

func TestBranch_Recruitment_UpdatePosition_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewRecruitmentService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.recruitment.findPositionErr = errors.New("find error")
		if _, err := svc.UpdatePosition(ctx, uuid.New(), model.PositionUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.recruitment.findPositionErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.recruitment.positions[id] = &model.RecruitmentPosition{BaseModel: model.BaseModel{ID: id}}
		r.recruitment.updatePositionErr = errors.New("update error")
		if _, err := svc.UpdatePosition(ctx, id, model.PositionUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.recruitment.updatePositionErr = nil
	})
}

func TestBranch_Recruitment_Applicant_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewRecruitmentService(deps)

	t.Run("create_applicant_error", func(t *testing.T) {
		r.recruitment.createApplicantErr = errors.New("create error")
		if _, err := svc.CreateApplicant(ctx, model.ApplicantCreateRequest{Name: "X"}); err == nil {
			t.Fatal("expected create error")
		}
		r.recruitment.createApplicantErr = nil
	})

	t.Run("update_applicant_error", func(t *testing.T) {
		appID := uuid.New()
		r.recruitment.applicants[appID] = &model.Applicant{BaseModel: model.BaseModel{ID: appID}, Stage: "new"}
		r.recruitment.updateApplicantErr = errors.New("update error")
		if _, err := svc.UpdateApplicantStage(ctx, appID, "interview"); err == nil {
			t.Fatal("expected update error")
		}
		r.recruitment.updateApplicantErr = nil
	})
}

// =============================================================================
// DocumentService branch coverage
// =============================================================================

func TestBranch_Document_UploadError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewDocumentService(deps)

	r.document.createErr = errors.New("upload error")
	if err := svc.Upload(ctx, &model.HRDocument{Title: "X"}); err == nil {
		t.Fatal("expected upload error")
	}
}

// =============================================================================
// AnnouncementService branch coverage
// =============================================================================

func TestBranch_Announcement_Create_WithPriority(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAnnouncementService(deps)

	got, err := svc.Create(ctx, model.AnnouncementCreateRequest{
		Title: "A", Content: "B", Priority: "high",
	}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if got.Priority != model.AnnouncementPriority("high") {
		t.Fatal("priority should be high, not default")
	}
}

func TestBranch_Announcement_CreateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAnnouncementService(deps)

	r.announcement.createErr = errors.New("create error")
	if _, err := svc.Create(ctx, model.AnnouncementCreateRequest{Title: "X"}, uuid.New()); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_Announcement_Update_ContentPriority(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAnnouncementService(deps)

	id := uuid.New()
	r.announcement.items[id] = &model.HRAnnouncement{BaseModel: model.BaseModel{ID: id}, Title: "Old"}

	content := "new content"
	priority := "urgent"
	got, err := svc.Update(ctx, id, model.AnnouncementUpdateRequest{
		Content:  &content,
		Priority: &priority,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "new content" || got.Priority != model.AnnouncementPriority("urgent") {
		t.Fatal("content/priority not updated")
	}
}

func TestBranch_Announcement_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAnnouncementService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.announcement.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), model.AnnouncementUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.announcement.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.announcement.items[id] = &model.HRAnnouncement{BaseModel: model.BaseModel{ID: id}}
		r.announcement.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, model.AnnouncementUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.announcement.updateErr = nil
	})
}

// =============================================================================
// HRDashboardService branch coverage
// =============================================================================

func TestBranch_Dashboard_GetRecentActivities_NilHireDate(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHRDashboardService(deps)

	id := uuid.New()
	r.hrEmployee.items[id] = &model.HREmployee{
		BaseModel: model.BaseModel{ID: id},
		FirstName: "A", LastName: "B",
		HireDate: nil, // nil HireDate → skip new_hire activity
	}

	activities, err := svc.GetRecentActivities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Should have fallback info message since no valid activities
	if len(activities) != 1 || activities[0]["type"] != "info" {
		t.Fatal("expected fallback info activity for nil hire date")
	}
}

// =============================================================================
// AttendanceIntegrationService branch coverage
// =============================================================================

func TestBranch_AttendanceIntegration_GetIntegration_NoPresent(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAttendanceIntegrationService(deps)

	r.hrEmployee.useCount = true
	r.hrEmployee.total = 5
	r.hrEmployee.active = 0 // active=0 → skip avg calculation branch

	v, err := svc.GetIntegration(ctx, "day", "")
	if err != nil {
		t.Fatal(err)
	}
	if v["avg_working_hours"].(float64) != 8.0 {
		t.Fatal("expected default avg_working_hours=8.0 when no present/active")
	}
}

func TestBranch_AttendanceIntegration_GetAlerts_Warning(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAttendanceIntegrationService(deps)

	userID := uuid.New()
	r.user.Users[userID] = &model.User{BaseModel: model.BaseModel{ID: userID}, FirstName: "T", LastName: "Y"}
	// 42 hours = warning (40 < 42 <= 45)
	r.overtime.monthlyOvertime[userID] = 42 * 60
	// Add attendance so no absence alert
	today := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 3; i++ {
		d := today.AddDate(0, 0, -i)
		aid := uuid.New()
		r.attendance.Attendances[aid] = &model.Attendance{
			BaseModel: model.BaseModel{ID: aid}, UserID: userID, Date: d,
		}
	}

	alerts, err := svc.GetAlerts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range alerts {
		if a["type"] == "overtime" && a["severity"] == "warning" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected overtime warning alert for 42h")
	}
}

func TestBranch_AttendanceIntegration_GetAlerts_HasApprovedLeave(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAttendanceIntegrationService(deps)

	userID := uuid.New()
	r.user.Users[userID] = &model.User{BaseModel: model.BaseModel{ID: userID}, FirstName: "L", LastName: "V"}
	// No attendance records → would normally trigger absence alert
	// But has approved leave → should NOT trigger absence alert
	now := time.Now()
	leaveID := uuid.New()
	r.leave.LeaveRequests[leaveID] = &model.LeaveRequest{
		BaseModel: model.BaseModel{ID: leaveID},
		UserID:    userID,
		Status:    model.ApprovalStatusApproved,
		StartDate: now.AddDate(0, 0, -5),
		EndDate:   now.AddDate(0, 0, 1),
	}

	alerts, err := svc.GetAlerts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range alerts {
		if a["type"] == "absence" && a["employee_id"] == userID {
			t.Fatal("should NOT have absence alert when approved leave exists")
		}
	}
}

func TestBranch_AttendanceIntegration_GetTrend_Week(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAttendanceIntegrationService(deps)

	r.hrEmployee.useCount = true
	r.hrEmployee.total = 1
	r.hrEmployee.active = 1

	trend, err := svc.GetTrend(ctx, "week") // not "month" → 7 days
	if err != nil {
		t.Fatal(err)
	}
	if len(trend) != 7 {
		t.Fatalf("expected 7 trend points, got %d", len(trend))
	}
}

func TestBranch_AttendanceIntegration_GetTrend_TotalZero(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAttendanceIntegrationService(deps)

	r.hrEmployee.useCount = true
	r.hrEmployee.total = 0 // should be set to 1 to prevent division by zero

	trend, err := svc.GetTrend(ctx, "week")
	if err != nil {
		t.Fatal(err)
	}
	if len(trend) != 7 {
		t.Fatalf("expected 7 trend points, got %d", len(trend))
	}
}

func TestBranch_AttendanceIntegration_GetTrend_RateCap(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAttendanceIntegrationService(deps)

	r.hrEmployee.useCount = true
	r.hrEmployee.total = 1
	r.hrEmployee.active = 1

	// Add many attendances for today to make rate > 100
	today := time.Now().Truncate(24 * time.Hour)
	for i := 0; i < 5; i++ {
		aid := uuid.New()
		r.attendance.Attendances[aid] = &model.Attendance{
			BaseModel: model.BaseModel{ID: aid}, Date: today,
		}
	}

	trend, err := svc.GetTrend(ctx, "week")
	if err != nil {
		t.Fatal(err)
	}
	lastRate := trend[len(trend)-1]["attendance_rate"].(float64)
	if lastRate > 100 {
		t.Fatal("rate should be capped at 100")
	}
}

// =============================================================================
// OrgChartService branch coverage
// =============================================================================

func TestBranch_OrgChart_Simulate_EmptyIDs(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOrgChartService(deps)

	deptID := uuid.New()
	r.hrDepartment.items[deptID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: deptID}, Name: "D1"}

	// empID="" → continue
	sim, err := svc.Simulate(ctx, map[string]interface{}{
		"moves": []interface{}{
			map[string]interface{}{
				"employee_id":        "",
				"from_department_id": deptID.String(),
				"to_department_id":   deptID.String(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sim) == 0 {
		t.Fatal("expected non-empty chart")
	}
}

func TestBranch_OrgChart_Simulate_EmployeeNotFound(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOrgChartService(deps)

	fromID := uuid.New()
	toID := uuid.New()
	r.hrDepartment.items[fromID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: fromID}, Name: "From"}
	r.hrDepartment.items[toID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: toID}, Name: "To"}

	// Employee doesn't exist in from dept → movedEmployee stays nil
	sim, err := svc.Simulate(ctx, map[string]interface{}{
		"moves": []interface{}{
			map[string]interface{}{
				"employee_id":        uuid.New().String(),
				"from_department_id": fromID.String(),
				"to_department_id":   toID.String(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sim) < 2 {
		t.Fatal("expected at least 2 departments")
	}
}

func TestBranch_OrgChart_Simulate_InvalidRename(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOrgChartService(deps)

	deptID := uuid.New()
	r.hrDepartment.items[deptID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: deptID}, Name: "D1"}

	sim, err := svc.Simulate(ctx, map[string]interface{}{
		"renames": []interface{}{
			"invalid_rename", // not a map → continue
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sim) == 0 {
		t.Fatal("expected non-empty chart")
	}
}

func TestBranch_OrgChart_Simulate_GetOrgChartError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOrgChartService(deps)

	r.hrDepartment.findAllErr = errors.New("find all error")
	if _, err := svc.Simulate(ctx, map[string]interface{}{}); err == nil {
		t.Fatal("expected GetOrgChart error propagation")
	}
}

func TestBranch_OrgChart_Simulate_NoMovesOrRenames(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOrgChartService(deps)

	deptID := uuid.New()
	r.hrDepartment.items[deptID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: deptID}, Name: "D1"}

	// data without "moves" or "renames" keys → skip both blocks
	sim, err := svc.Simulate(ctx, map[string]interface{}{"other": "data"})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range sim {
		if d["simulated"] != true {
			t.Fatal("simulated flag should be set")
		}
	}
}

// =============================================================================
// OneOnOneService branch coverage
// =============================================================================

func TestBranch_OneOnOne_Create_RFC3339Date(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	empID := uuid.New()
	got, err := svc.Create(ctx, model.OneOnOneCreateRequest{
		EmployeeID:    empID.String(),
		ScheduledDate: "2025-03-15T10:00:00+09:00",
		Frequency:     "weekly",
	}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if got.Frequency != "weekly" {
		t.Fatal("frequency should be weekly")
	}
	if got.ScheduledDate.IsZero() {
		t.Fatal("scheduledDate should be parsed")
	}
}

func TestBranch_OneOnOne_Create_Error(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	r.oneOnOne.createErr = errors.New("create error")
	empID := uuid.New()
	if _, err := svc.Create(ctx, model.OneOnOneCreateRequest{
		EmployeeID: empID.String(), ScheduledDate: "2025-01-01",
	}, uuid.New()); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_OneOnOne_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.oneOnOne.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), model.OneOnOneUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.oneOnOne.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.oneOnOne.items[id] = &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: id}}
		r.oneOnOne.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, model.OneOnOneUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.oneOnOne.updateErr = nil
	})

	t.Run("empty_status_noop", func(t *testing.T) {
		id := uuid.New()
		r.oneOnOne.items[id] = &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: id}, Status: "scheduled"}
		got, err := svc.Update(ctx, id, model.OneOnOneUpdateRequest{Status: ""})
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != "scheduled" {
			t.Fatal("status should remain unchanged for empty string")
		}
	})
}

func TestBranch_OneOnOne_AddActionItem_NilActions(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	id := uuid.New()
	r.oneOnOne.items[id] = &model.OneOnOneMeeting{
		BaseModel:   model.BaseModel{ID: id},
		ActionItems: nil, // nil ActionItems
	}
	got, err := svc.AddActionItem(ctx, id, model.ActionItemRequest{Title: "Task"})
	if err != nil {
		t.Fatal(err)
	}
	var items []map[string]interface{}
	json.Unmarshal(got.ActionItems, &items)
	if len(items) != 1 {
		t.Fatal("expected 1 action item")
	}
}

func TestBranch_OneOnOne_AddActionItem_UpdateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	id := uuid.New()
	r.oneOnOne.items[id] = &model.OneOnOneMeeting{
		BaseModel:   model.BaseModel{ID: id},
		ActionItems: datatypes.JSON([]byte("[]")),
	}
	r.oneOnOne.updateErr = errors.New("update error")
	if _, err := svc.AddActionItem(ctx, id, model.ActionItemRequest{Title: "X"}); err == nil {
		t.Fatal("expected update error")
	}
}

func TestBranch_OneOnOne_ToggleActionItem_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.oneOnOne.findByIDErr = errors.New("find error")
		if _, err := svc.ToggleActionItem(ctx, uuid.New(), "x"); err == nil {
			t.Fatal("expected find error")
		}
		r.oneOnOne.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.oneOnOne.items[id] = &model.OneOnOneMeeting{
			BaseModel:   model.BaseModel{ID: id},
			ActionItems: datatypes.JSON([]byte(`[{"id":"a1","completed":false}]`)),
		}
		r.oneOnOne.updateErr = errors.New("update error")
		if _, err := svc.ToggleActionItem(ctx, id, "a1"); err == nil {
			t.Fatal("expected update error")
		}
		r.oneOnOne.updateErr = nil
	})

	t.Run("nil_action_items", func(t *testing.T) {
		id := uuid.New()
		r.oneOnOne.items[id] = &model.OneOnOneMeeting{
			BaseModel:   model.BaseModel{ID: id},
			ActionItems: nil,
		}
		got, err := svc.ToggleActionItem(ctx, id, "nonexistent")
		if err != nil {
			t.Fatal(err)
		}
		_ = got
	})

	t.Run("non_matching_id", func(t *testing.T) {
		id := uuid.New()
		r.oneOnOne.items[id] = &model.OneOnOneMeeting{
			BaseModel:   model.BaseModel{ID: id},
			ActionItems: datatypes.JSON([]byte(`[{"id":"a1","completed":false}]`)),
		}
		got, err := svc.ToggleActionItem(ctx, id, "no-match")
		if err != nil {
			t.Fatal(err)
		}
		var items []map[string]interface{}
		json.Unmarshal(got.ActionItems, &items)
		if items[0]["completed"] != false {
			t.Fatal("should not toggle non-matching item")
		}
	})
}

// =============================================================================
// SkillService branch coverage
// =============================================================================

func TestBranch_Skill_AddSkill_NonDefaults(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSkillService(deps)

	got, err := svc.AddSkill(ctx, uuid.New(), model.SkillAddRequest{
		SkillName: "Python", Category: "business", Level: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Category != "business" || got.Level != 3 {
		t.Fatal("non-default values should be used")
	}
}

func TestBranch_Skill_AddSkill_Error(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSkillService(deps)

	r.skill.createErr = errors.New("create error")
	if _, err := svc.AddSkill(ctx, uuid.New(), model.SkillAddRequest{SkillName: "X"}); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_Skill_UpdateSkill_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSkillService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.skill.findByIDErr = errors.New("find error")
		if _, err := svc.UpdateSkill(ctx, uuid.New(), model.SkillUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.skill.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.skill.items[id] = &model.EmployeeSkill{BaseModel: model.BaseModel{ID: id}}
		r.skill.updateErr = errors.New("update error")
		if _, err := svc.UpdateSkill(ctx, id, model.SkillUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.skill.updateErr = nil
	})

	t.Run("partial_fields_only_level", func(t *testing.T) {
		id := uuid.New()
		r.skill.items[id] = &model.EmployeeSkill{
			BaseModel: model.BaseModel{ID: id}, Category: "tech", Level: 1,
		}
		level := 5
		got, err := svc.UpdateSkill(ctx, id, model.SkillUpdateRequest{Level: &level})
		if err != nil {
			t.Fatal(err)
		}
		if got.Level != 5 || got.Category != "tech" {
			t.Fatal("only level should change")
		}
	})

	t.Run("no_fields", func(t *testing.T) {
		id := uuid.New()
		r.skill.items[id] = &model.EmployeeSkill{
			BaseModel: model.BaseModel{ID: id}, Category: "tech", Level: 2,
		}
		got, err := svc.UpdateSkill(ctx, id, model.SkillUpdateRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Level != 2 || got.Category != "tech" {
			t.Fatal("no fields should change")
		}
	})
}

// =============================================================================
// SalaryService branch coverage
// =============================================================================

func TestBranch_Salary_Simulate_NoGrade(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	got, err := svc.Simulate(ctx, model.SalarySimulateRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got["base_salary"].(float64) != 300000 {
		t.Fatal("expected default base 300000 when no grade")
	}
}

func TestBranch_Salary_Simulate_NoEvalScore(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	got, err := svc.Simulate(ctx, model.SalarySimulateRequest{
		Grade:          "S1",
		YearsOfService: "3",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["evaluation_bonus"].(float64) != 0 {
		t.Fatal("expected evaluation_bonus=0 when no score")
	}
}

func TestBranch_Salary_Simulate_PositionNoData(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	got, err := svc.Simulate(ctx, model.SalarySimulateRequest{
		Grade:    "M1",
		Position: "NonExistentPosition",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["position_adjustment"].(float64) != 0 {
		t.Fatal("expected position_adjustment=0 when no data")
	}
}

func TestBranch_Salary_Simulate_S2Grade(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	got, err := svc.Simulate(ctx, model.SalarySimulateRequest{Grade: "S2"})
	if err != nil {
		t.Fatal(err)
	}
	if got["base_salary"].(float64) != 250000 {
		t.Fatal("S2 should have base 250000")
	}
}

// TestBranch_Salary_Simulate_WithActiveEmployees covers the for-loop body
// and gradeCount > 0 / posCount > 0 branches inside Simulate.
func TestBranch_Salary_Simulate_WithActiveEmployees(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()
	r.hrEmployee.items = map[uuid.UUID]*model.HREmployee{
		id1: {BaseModel: model.BaseModel{ID: id1}, Grade: "M1", Position: "Engineer", BaseSalary: 400000, Status: model.EmployeeStatusActive},
		id2: {BaseModel: model.BaseModel{ID: id2}, Grade: "M1", Position: "Engineer", BaseSalary: 500000, Status: model.EmployeeStatusActive},
		id3: {BaseModel: model.BaseModel{ID: id3}, Grade: "S1", Position: "Designer", BaseSalary: 250000, Status: model.EmployeeStatusActive},
	}

	t.Run("grade_match_with_data", func(t *testing.T) {
		got, err := svc.Simulate(ctx, model.SalarySimulateRequest{
			Grade:           "M1",
			YearsOfService:  "5",
			EvaluationScore: "80",
		})
		if err != nil {
			t.Fatal(err)
		}
		// gradeCount=2, average=(400000+500000)/2=450000
		if got["base_salary"].(float64) != 450000 {
			t.Fatalf("expected base_salary=450000, got %v", got["base_salary"])
		}
	})

	t.Run("position_match_with_data", func(t *testing.T) {
		got, err := svc.Simulate(ctx, model.SalarySimulateRequest{
			Grade:    "M1",
			Position: "Engineer",
		})
		if err != nil {
			t.Fatal(err)
		}
		// posCount=2, posAvg=(400000+500000)/2=450000, baseSalary=450000
		// positionAdjustment = (450000 - 450000) * 0.3 = 0
		if got["position_adjustment"].(float64) != 0 {
			t.Fatalf("expected position_adjustment=0, got %v", got["position_adjustment"])
		}
	})

	t.Run("position_nonzero_adjustment", func(t *testing.T) {
		// Use grade S1 (base from data: 250000) with position Engineer (posAvg: 450000)
		got, err := svc.Simulate(ctx, model.SalarySimulateRequest{
			Grade:    "S1",
			Position: "Engineer",
		})
		if err != nil {
			t.Fatal(err)
		}
		// baseSalary = 250000 (from S1 employee average)
		// posAvg = (400000+500000)/2 = 450000
		// positionAdjustment = (450000 - 250000) * 0.3 = 60000
		if got["position_adjustment"].(float64) != 60000 {
			t.Fatalf("expected position_adjustment=60000, got %v", got["position_adjustment"])
		}
	})
}

func TestBranch_Salary_GetBudget_AllBranches(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	t.Run("budget_positive_department_by_code", func(t *testing.T) {
		r.salary.overview = map[string]interface{}{"total_payroll": 100000.0, "headcount": int64(2)}
		deptID := uuid.New()
		r.hrDepartment.items[deptID] = &model.HRDepartment{
			BaseModel: model.BaseModel{ID: deptID}, Name: "Eng", Code: "ENG", Budget: 5000000,
		}
		got, err := svc.GetBudget(ctx, "ENG") // match by code
		if err != nil {
			t.Fatal(err)
		}
		if got["total_budget"].(float64) != 5000000 {
			t.Fatalf("expected budget=5000000, got %v", got["total_budget"])
		}
	})

	t.Run("department_empty_sums_all", func(t *testing.T) {
		r.salary.overview = map[string]interface{}{"total_payroll": 100000.0, "headcount": int64(1)}
		r.hrDepartment.items = map[uuid.UUID]*model.HRDepartment{}
		d1 := uuid.New()
		d2 := uuid.New()
		r.hrDepartment.items[d1] = &model.HRDepartment{BaseModel: model.BaseModel{ID: d1}, Budget: 1000000}
		r.hrDepartment.items[d2] = &model.HRDepartment{BaseModel: model.BaseModel{ID: d2}, Budget: 2000000}

		got, err := svc.GetBudget(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		if got["total_budget"].(float64) != 3000000 {
			t.Fatalf("expected budget=3000000, got %v", got["total_budget"])
		}
	})

	t.Run("remaining_negative_clamped", func(t *testing.T) {
		// very high payroll, low budget → remaining < 0 → clamped to 0
		r.salary.overview = map[string]interface{}{"total_payroll": 1000000.0, "headcount": int64(1)}
		r.hrDepartment.items = map[uuid.UUID]*model.HRDepartment{}
		deptID := uuid.New()
		r.hrDepartment.items[deptID] = &model.HRDepartment{
			BaseModel: model.BaseModel{ID: deptID}, Name: "X", Budget: 100000, // much less than annual payroll
		}

		got, err := svc.GetBudget(ctx, "X")
		if err != nil {
			t.Fatal(err)
		}
		if got["remaining"].(float64) != 0 {
			t.Fatalf("expected remaining=0 (clamped), got %v", got["remaining"])
		}
	})

	t.Run("total_budget_zero_utilization", func(t *testing.T) {
		r.salary.overview = map[string]interface{}{"total_payroll": 0.0, "headcount": int64(0)}
		r.hrDepartment.items = map[uuid.UUID]*model.HRDepartment{}

		got, err := svc.GetBudget(ctx, "NoMatch")
		if err != nil {
			t.Fatal(err)
		}
		// totalBudget = 0*12*1.3 = 0, so utilization should be 0
		if got["utilization"].(float64) != 0 {
			t.Fatalf("expected utilization=0, got %v", got["utilization"])
		}
	})
}

// =============================================================================
// OnboardingService branch coverage
// =============================================================================

func TestBranch_Onboarding_CreateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOnboardingService(deps)

	r.onboarding.createErr = errors.New("create error")
	if _, err := svc.Create(ctx, model.OnboardingCreateRequest{
		EmployeeID: uuid.New().String(), StartDate: "2025-01-01",
	}); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_Onboarding_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOnboardingService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.onboarding.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), map[string]interface{}{}); err == nil {
			t.Fatal("expected find error")
		}
		r.onboarding.findByIDErr = nil
	})

	t.Run("no_status_key", func(t *testing.T) {
		id := uuid.New()
		r.onboarding.items[id] = &model.Onboarding{
			BaseModel: model.BaseModel{ID: id}, Status: model.OnboardingStatusPending,
		}
		got, err := svc.Update(ctx, id, map[string]interface{}{"other": "data"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != model.OnboardingStatusPending {
			t.Fatal("status should not change when no status key")
		}
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.onboarding.items[id] = &model.Onboarding{BaseModel: model.BaseModel{ID: id}}
		r.onboarding.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, map[string]interface{}{}); err == nil {
			t.Fatal("expected update error")
		}
		r.onboarding.updateErr = nil
	})
}

func TestBranch_Onboarding_ToggleTask_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOnboardingService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.onboarding.findByIDErr = errors.New("find error")
		if _, err := svc.ToggleTask(ctx, uuid.New(), "x"); err == nil {
			t.Fatal("expected find error")
		}
		r.onboarding.findByIDErr = nil
	})

	t.Run("nil_tasks", func(t *testing.T) {
		id := uuid.New()
		r.onboarding.items[id] = &model.Onboarding{
			BaseModel: model.BaseModel{ID: id}, Tasks: nil,
		}
		got, err := svc.ToggleTask(ctx, id, "x")
		if err != nil {
			t.Fatal(err)
		}
		_ = got
	})

	t.Run("non_matching_task", func(t *testing.T) {
		id := uuid.New()
		r.onboarding.items[id] = &model.Onboarding{
			BaseModel: model.BaseModel{ID: id},
			Tasks:     datatypes.JSON([]byte(`[{"id":"a1","completed":false}]`)),
		}
		got, err := svc.ToggleTask(ctx, id, "no-match")
		if err != nil {
			t.Fatal(err)
		}
		var tasks []map[string]interface{}
		json.Unmarshal(got.Tasks, &tasks)
		if tasks[0]["completed"] != false {
			t.Fatal("non-matching task should not toggle")
		}
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.onboarding.items[id] = &model.Onboarding{
			BaseModel: model.BaseModel{ID: id},
			Tasks:     datatypes.JSON([]byte("[]")),
		}
		r.onboarding.updateErr = errors.New("update error")
		if _, err := svc.ToggleTask(ctx, id, "x"); err == nil {
			t.Fatal("expected update error")
		}
		r.onboarding.updateErr = nil
	})
}

func TestBranch_Onboarding_CreateTemplate_Error(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOnboardingService(deps)

	r.onboarding.createTemplateErr = errors.New("template error")
	if _, err := svc.CreateTemplate(ctx, model.OnboardingTemplateCreateRequest{Name: "X"}); err == nil {
		t.Fatal("expected template create error")
	}
}

// =============================================================================
// OffboardingService branch coverage
// =============================================================================

func TestBranch_Offboarding_CreateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOffboardingService(deps)

	r.offboarding.createErr = errors.New("create error")
	if _, err := svc.Create(ctx, model.OffboardingCreateRequest{
		EmployeeID: uuid.New().String(), LastWorkingDate: "2025-03-31",
	}); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_Offboarding_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOffboardingService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.offboarding.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), model.OffboardingUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.offboarding.findByIDErr = nil
	})

	t.Run("partial_fields", func(t *testing.T) {
		id := uuid.New()
		r.offboarding.items[id] = &model.Offboarding{
			BaseModel: model.BaseModel{ID: id}, Status: model.OffboardingStatusPending, Notes: "old",
		}
		got, err := svc.Update(ctx, id, model.OffboardingUpdateRequest{}) // no fields set
		if err != nil {
			t.Fatal(err)
		}
		if got.Notes != "old" {
			t.Fatal("notes should remain unchanged")
		}
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.offboarding.items[id] = &model.Offboarding{BaseModel: model.BaseModel{ID: id}}
		r.offboarding.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, model.OffboardingUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.offboarding.updateErr = nil
	})
}

func TestBranch_Offboarding_ToggleChecklist_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOffboardingService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.offboarding.findByIDErr = errors.New("find error")
		if _, err := svc.ToggleChecklist(ctx, uuid.New(), "x"); err == nil {
			t.Fatal("expected find error")
		}
		r.offboarding.findByIDErr = nil
	})

	t.Run("nil_checklist", func(t *testing.T) {
		id := uuid.New()
		r.offboarding.items[id] = &model.Offboarding{
			BaseModel: model.BaseModel{ID: id}, Checklist: nil,
		}
		got, err := svc.ToggleChecklist(ctx, id, "x")
		if err != nil {
			t.Fatal(err)
		}
		_ = got
	})

	t.Run("non_matching_key", func(t *testing.T) {
		id := uuid.New()
		r.offboarding.items[id] = &model.Offboarding{
			BaseModel: model.BaseModel{ID: id},
			Checklist: datatypes.JSON([]byte(`[{"key":"pc","completed":false}]`)),
		}
		got, err := svc.ToggleChecklist(ctx, id, "no-match")
		if err != nil {
			t.Fatal(err)
		}
		var list []map[string]interface{}
		json.Unmarshal(got.Checklist, &list)
		if list[0]["completed"] != false {
			t.Fatal("non-matching item should not toggle")
		}
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.offboarding.items[id] = &model.Offboarding{
			BaseModel: model.BaseModel{ID: id},
			Checklist: datatypes.JSON([]byte("[]")),
		}
		r.offboarding.updateErr = errors.New("update error")
		if _, err := svc.ToggleChecklist(ctx, id, "x"); err == nil {
			t.Fatal("expected update error")
		}
		r.offboarding.updateErr = nil
	})
}

// =============================================================================
// SurveyService branch coverage
// =============================================================================

func TestBranch_Survey_Create_WithType(t *testing.T) {
	deps, _ := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	got, err := svc.Create(ctx, model.SurveyCreateRequest{
		Title: "S", Type: "pulse",
	}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "pulse" {
		t.Fatal("type should be pulse, not default")
	}
}

func TestBranch_Survey_CreateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	r.survey.createErr = errors.New("create error")
	if _, err := svc.Create(ctx, model.SurveyCreateRequest{Title: "X"}, uuid.New()); err == nil {
		t.Fatal("expected create error")
	}
}

func TestBranch_Survey_Update_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.survey.findByIDErr = errors.New("find error")
		if _, err := svc.Update(ctx, uuid.New(), model.SurveyUpdateRequest{}); err == nil {
			t.Fatal("expected find error")
		}
		r.survey.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.survey.surveys[id] = &model.Survey{BaseModel: model.BaseModel{ID: id}}
		r.survey.updateErr = errors.New("update error")
		if _, err := svc.Update(ctx, id, model.SurveyUpdateRequest{}); err == nil {
			t.Fatal("expected update error")
		}
		r.survey.updateErr = nil
	})

	t.Run("partial_description", func(t *testing.T) {
		id := uuid.New()
		r.survey.surveys[id] = &model.Survey{BaseModel: model.BaseModel{ID: id}, Title: "Old"}
		desc := "new desc"
		got, err := svc.Update(ctx, id, model.SurveyUpdateRequest{Description: &desc})
		if err != nil {
			t.Fatal(err)
		}
		if got.Description != "new desc" || got.Title != "Old" {
			t.Fatal("only description should change")
		}
	})

	t.Run("no_fields", func(t *testing.T) {
		id := uuid.New()
		r.survey.surveys[id] = &model.Survey{BaseModel: model.BaseModel{ID: id}, Title: "Keep"}
		got, err := svc.Update(ctx, id, model.SurveyUpdateRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Title != "Keep" {
			t.Fatal("title should remain unchanged")
		}
	})
}

func TestBranch_Survey_Publish_UpdateError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	id := uuid.New()
	r.survey.surveys[id] = &model.Survey{BaseModel: model.BaseModel{ID: id}}
	r.survey.updateErr = errors.New("update error")
	if _, err := svc.Publish(ctx, id); err == nil {
		t.Fatal("expected publish update error")
	}
}

func TestBranch_Survey_Close_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.survey.findByIDErr = errors.New("find error")
		if _, err := svc.Close(ctx, uuid.New()); err == nil {
			t.Fatal("expected close find error")
		}
		r.survey.findByIDErr = nil
	})

	t.Run("update_error", func(t *testing.T) {
		id := uuid.New()
		r.survey.surveys[id] = &model.Survey{BaseModel: model.BaseModel{ID: id}}
		r.survey.updateErr = errors.New("update error")
		if _, err := svc.Close(ctx, id); err == nil {
			t.Fatal("expected close update error")
		}
		r.survey.updateErr = nil
	})
}

func TestBranch_Survey_GetResults_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	t.Run("find_error", func(t *testing.T) {
		r.survey.findByIDErr = errors.New("find error")
		if _, err := svc.GetResults(ctx, uuid.New()); err == nil {
			t.Fatal("expected find error")
		}
		r.survey.findByIDErr = nil
	})

	t.Run("responses_error", func(t *testing.T) {
		id := uuid.New()
		r.survey.surveys[id] = &model.Survey{BaseModel: model.BaseModel{ID: id}}
		r.survey.findResponsesErr = errors.New("responses error")
		if _, err := svc.GetResults(ctx, id); err == nil {
			t.Fatal("expected responses error")
		}
		r.survey.findResponsesErr = nil
	})
}

// =============================================================================
// Announcement Delete error
// =============================================================================

func TestBranch_Announcement_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAnnouncementService(deps)

	r.announcement.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// Employee Delete error
// =============================================================================

func TestBranch_Employee_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHREmployeeService(deps)

	r.hrEmployee.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// Department Delete error
// =============================================================================

func TestBranch_Department_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHRDepartmentService(deps)

	r.hrDepartment.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// Goal Delete error
// =============================================================================

func TestBranch_Goal_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewGoalService(deps)

	r.goal.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// Training Delete error
// =============================================================================

func TestBranch_Training_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewTrainingService(deps)

	r.training.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// OneOnOne Delete error
// =============================================================================

func TestBranch_OneOnOne_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	r.oneOnOne.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// Survey Delete error
// =============================================================================

func TestBranch_Survey_DeleteError(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	r.survey.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
}

// =============================================================================
// Document Delete error & FindAll/FindByID errors
// =============================================================================

func TestBranch_Document_Errors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewDocumentService(deps)

	r.document.deleteErr = errors.New("delete error")
	if err := svc.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("expected delete error")
	}
	r.document.deleteErr = nil

	r.document.findAllErr = errors.New("findall error")
	if _, _, err := svc.FindAll(ctx, 1, 10, "", ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.document.findAllErr = nil

	r.document.findByIDErr = errors.New("findbyid error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected findbyid error")
	}
}

// =============================================================================
// FindByID/FindAll error paths for various services
// =============================================================================

func TestBranch_Employee_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewHREmployeeService(deps)

	r.hrEmployee.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.hrEmployee.findByIDErr = nil

	r.hrEmployee.findAllErr = errors.New("findall error")
	if _, _, err := svc.FindAll(ctx, 1, 10, "", "", "", ""); err == nil {
		t.Fatal("expected findall error")
	}
}

func TestBranch_Evaluation_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewEvaluationService(deps)

	r.evaluation.findAllErr = errors.New("findall error")
	if _, _, err := svc.FindAll(ctx, 1, 10, "", ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.evaluation.findAllErr = nil

	r.evaluation.findCyclesErr = errors.New("cycles error")
	if _, err := svc.FindAllCycles(ctx); err == nil {
		t.Fatal("expected cycles error")
	}
}

func TestBranch_Goal_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewGoalService(deps)

	r.goal.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.goal.findByIDErr = nil

	r.goal.findAllErr = errors.New("findall error")
	if _, _, err := svc.FindAll(ctx, 1, 10, "", "", ""); err == nil {
		t.Fatal("expected findall error")
	}
}

func TestBranch_Training_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewTrainingService(deps)

	r.training.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.training.findByIDErr = nil

	r.training.findAllErr = errors.New("findall error")
	if _, _, err := svc.FindAll(ctx, 1, 10, "", ""); err == nil {
		t.Fatal("expected findall error")
	}
}

func TestBranch_Recruitment_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewRecruitmentService(deps)

	r.recruitment.findAllPositionErr = errors.New("findall error")
	if _, _, err := svc.FindAllPositions(ctx, 1, 10, "", ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.recruitment.findAllPositionErr = nil

	r.recruitment.findApplicantsErr = errors.New("find applicants error")
	if _, err := svc.FindAllApplicants(ctx, "", ""); err == nil {
		t.Fatal("expected find applicants error")
	}
}

func TestBranch_Announcement_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewAnnouncementService(deps)

	r.announcement.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.announcement.findByIDErr = nil

	r.announcement.findAllErr = errors.New("findall error")
	if _, _, err := svc.FindAll(ctx, 1, 10, ""); err == nil {
		t.Fatal("expected findall error")
	}
}

func TestBranch_OneOnOne_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOneOnOneService(deps)

	r.oneOnOne.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.oneOnOne.findByIDErr = nil

	r.oneOnOne.findAllErr = errors.New("findall error")
	if _, err := svc.FindAll(ctx, "", ""); err == nil {
		t.Fatal("expected findall error")
	}
}

func TestBranch_Skill_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSkillService(deps)

	r.skill.gapErr = errors.New("gap error")
	if _, err := svc.GetGapAnalysis(ctx, ""); err == nil {
		t.Fatal("expected gap error")
	}
	r.skill.gapErr = nil

	r.skill.findAllErr = errors.New("findall error")
	if _, err := svc.GetSkillMap(ctx, "dept", ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.skill.findAllErr = nil

	empID := uuid.New()
	r.skill.findEmployeeErr = errors.New("employee error")
	if _, err := svc.GetSkillMap(ctx, "", empID.String()); err == nil {
		t.Fatal("expected employee error")
	}
}

func TestBranch_Salary_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSalaryService(deps)

	r.salary.overviewErr = errors.New("overview error")
	if _, err := svc.GetOverview(ctx, ""); err == nil {
		t.Fatal("expected overview error")
	}
	r.salary.overviewErr = nil

	r.salary.findByIDErr = errors.New("findbyid error")
	if _, err := svc.GetHistory(ctx, uuid.New()); err == nil {
		t.Fatal("expected findbyid error")
	}
}

func TestBranch_Onboarding_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOnboardingService(deps)

	r.onboarding.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.onboarding.findByIDErr = nil

	r.onboarding.findAllErr = errors.New("findall error")
	if _, err := svc.FindAll(ctx, ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.onboarding.findAllErr = nil

	r.onboarding.findTemplatesErr = errors.New("templates error")
	if _, err := svc.FindAllTemplates(ctx); err == nil {
		t.Fatal("expected templates error")
	}
}

func TestBranch_Offboarding_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewOffboardingService(deps)

	r.offboarding.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.offboarding.findByIDErr = nil

	r.offboarding.findAllErr = errors.New("findall error")
	if _, err := svc.FindAll(ctx, ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.offboarding.findAllErr = nil

	r.offboarding.analyticsErr = errors.New("analytics error")
	if _, err := svc.GetAnalytics(ctx); err == nil {
		t.Fatal("expected analytics error")
	}
}

func TestBranch_Survey_FindErrors(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()
	svc := NewSurveyService(deps)

	r.survey.findByIDErr = errors.New("find error")
	if _, err := svc.FindByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected find error")
	}
	r.survey.findByIDErr = nil

	r.survey.findAllErr = errors.New("findall error")
	if _, err := svc.FindAll(ctx, "", ""); err == nil {
		t.Fatal("expected findall error")
	}
	r.survey.findAllErr = nil

	r.survey.createResponseErr = errors.New("response error")
	if err := svc.SubmitResponse(ctx, uuid.New(), nil, model.SurveyResponseRequest{}); err == nil {
		t.Fatal("expected submit response error")
	}
}
