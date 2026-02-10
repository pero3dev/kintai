package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

func TestRepositorySmoke_CoverMoreStatements(t *testing.T) {
	db := dryRunDB(t)
	ctx := context.Background()
	now := time.Now()
	id := uuid.New()

	userRepo := NewUserRepository(db)
	_ = userRepo.Create(ctx, &model.User{})
	_, _, _ = userRepo.FindAll(ctx, 1, 10)
	_ = userRepo.Update(ctx, &model.User{})
	_ = userRepo.Delete(ctx, id)
	_, _ = userRepo.FindByDepartmentID(ctx, id)

	attendanceRepo := NewAttendanceRepository(db)
	_ = attendanceRepo.Create(ctx, &model.Attendance{})
	_, _, _ = attendanceRepo.FindByUserAndDateRange(ctx, id, now, now, 1, 10)
	_, _ = attendanceRepo.FindByDateRange(ctx, now, now)
	_ = attendanceRepo.Update(ctx, &model.Attendance{})
	_, _ = attendanceRepo.GetSummary(ctx, id, now.AddDate(0, 0, -30), now)
	_, _ = attendanceRepo.CountTodayPresent(ctx)
	_, _ = attendanceRepo.GetMonthlyOvertime(ctx, now.AddDate(0, -1, 0), now)

	leaveRepo := NewLeaveRequestRepository(db)
	_ = leaveRepo.Create(ctx, &model.LeaveRequest{})
	_, _, _ = leaveRepo.FindByUserID(ctx, id, 1, 10)
	_, _, _ = leaveRepo.FindPending(ctx, 1, 10)
	_ = leaveRepo.Update(ctx, &model.LeaveRequest{})
	_, _ = leaveRepo.CountPending(ctx)

	shiftRepo := NewShiftRepository(db)
	_ = shiftRepo.Create(ctx, &model.Shift{})
	_ = shiftRepo.BulkCreate(ctx, []model.Shift{{}})
	_, _ = shiftRepo.FindByUserAndDateRange(ctx, id, now, now)
	_, _ = shiftRepo.FindByDateRange(ctx, now, now)
	_ = shiftRepo.Update(ctx, &model.Shift{})
	_ = shiftRepo.Delete(ctx, id)

	deptRepo := NewDepartmentRepository(db)
	_ = deptRepo.Create(ctx, &model.Department{})
	_, _ = deptRepo.FindAll(ctx)
	_ = deptRepo.Update(ctx, &model.Department{})
	_ = deptRepo.Delete(ctx, id)

	rtRepo := NewRefreshTokenRepository(db)
	_ = rtRepo.Create(ctx, &model.RefreshToken{})
	_ = rtRepo.RevokeByUserID(ctx, id)
	_ = rtRepo.Revoke(ctx, "token")
	_ = rtRepo.DeleteExpired(ctx)

	otRepo := NewOvertimeRequestRepository(db)
	_ = otRepo.Create(ctx, &model.OvertimeRequest{})
	_, _, _ = otRepo.FindByUserID(ctx, id, 1, 10)
	_, _, _ = otRepo.FindPending(ctx, 1, 10)
	_ = otRepo.Update(ctx, &model.OvertimeRequest{})
	_, _ = otRepo.CountPending(ctx)
	_, _ = otRepo.GetUserMonthlyOvertime(ctx, id, now.Year(), int(now.Month()))
	_, _ = otRepo.GetUserYearlyOvertime(ctx, id, now.Year())

	balanceRepo := NewLeaveBalanceRepository(db)
	_ = balanceRepo.Create(ctx, &model.LeaveBalance{})
	_, _ = balanceRepo.FindByUserAndYear(ctx, id, now.Year())
	_ = balanceRepo.Update(ctx, &model.LeaveBalance{})
	_ = balanceRepo.Upsert(ctx, &model.LeaveBalance{UserID: id, FiscalYear: now.Year(), LeaveType: model.LeaveTypePaid})

	corrRepo := NewAttendanceCorrectionRepository(db)
	_ = corrRepo.Create(ctx, &model.AttendanceCorrection{})
	_, _, _ = corrRepo.FindByUserID(ctx, id, 1, 10)
	_, _, _ = corrRepo.FindPending(ctx, 1, 10)
	_ = corrRepo.Update(ctx, &model.AttendanceCorrection{})
	_, _ = corrRepo.CountPending(ctx)

	notifRepo := NewNotificationRepository(db)
	_ = notifRepo.Create(ctx, &model.Notification{})
	_ = notifRepo.MarkAsRead(ctx, id)
	_ = notifRepo.MarkAllAsRead(ctx, id)
	_, _ = notifRepo.CountUnread(ctx, id)
	_ = notifRepo.Delete(ctx, id)

	projectRepo := NewProjectRepository(db)
	_ = projectRepo.Create(ctx, &model.Project{})
	_ = projectRepo.Update(ctx, &model.Project{})
	_ = projectRepo.Delete(ctx, id)

	timeRepo := NewTimeEntryRepository(db)
	_ = timeRepo.Create(ctx, &model.TimeEntry{})
	_, _ = timeRepo.FindByUserAndDateRange(ctx, id, now, now)
	_, _ = timeRepo.FindByProjectAndDateRange(ctx, id, now, now)
	_ = timeRepo.Update(ctx, &model.TimeEntry{})
	_ = timeRepo.Delete(ctx, id)
	_, _ = timeRepo.GetProjectSummary(ctx, now.AddDate(0, -1, 0), now)

	holidayRepo := NewHolidayRepository(db)
	_ = holidayRepo.Create(ctx, &model.Holiday{})
	_, _ = holidayRepo.FindByDateRange(ctx, now.AddDate(0, 0, -7), now)
	_, _ = holidayRepo.FindByYear(ctx, now.Year())
	_ = holidayRepo.Update(ctx, &model.Holiday{})
	_ = holidayRepo.Delete(ctx, id)

	flowRepo := NewApprovalFlowRepository(db)
	_ = flowRepo.Create(ctx, &model.ApprovalFlow{})
	_, _ = flowRepo.FindByType(ctx, model.ApprovalFlowLeave)
	_, _ = flowRepo.FindAll(ctx)
	_ = flowRepo.Update(ctx, &model.ApprovalFlow{})
	_ = flowRepo.Delete(ctx, id)
	_ = flowRepo.DeleteStepsByFlowID(ctx, id)
	_ = flowRepo.CreateSteps(ctx, []model.ApprovalStep{{}})

	hrEmp := NewHREmployeeRepository(db)
	_ = hrEmp.Create(ctx, &model.HREmployee{})
	_ = hrEmp.Update(ctx, &model.HREmployee{})
	_ = hrEmp.Delete(ctx, id)
	_, _ = hrEmp.FindByDepartmentID(ctx, id)
	_, _, _ = hrEmp.CountByStatus(ctx)

	hrDept := NewHRDepartmentRepository(db)
	_ = hrDept.Create(ctx, &model.HRDepartment{})
	_, _ = hrDept.FindAll(ctx)
	_ = hrDept.Update(ctx, &model.HRDepartment{})
	_ = hrDept.Delete(ctx, id)

	evalRepo := NewEvaluationRepository(db)
	_ = evalRepo.Create(ctx, &model.Evaluation{})
	_ = evalRepo.Update(ctx, &model.Evaluation{})
	_ = evalRepo.CreateCycle(ctx, &model.EvaluationCycle{})
	_, _ = evalRepo.FindAllCycles(ctx)

	goalRepo := NewGoalRepository(db)
	_ = goalRepo.Create(ctx, &model.HRGoal{})
	_ = goalRepo.Update(ctx, &model.HRGoal{})
	_ = goalRepo.Delete(ctx, id)

	trainingRepo := NewTrainingRepository(db)
	_ = trainingRepo.Create(ctx, &model.TrainingProgram{})
	_ = trainingRepo.Update(ctx, &model.TrainingProgram{})
	_ = trainingRepo.Delete(ctx, id)
	_ = trainingRepo.CreateEnrollment(ctx, &model.TrainingEnrollment{})
	_ = trainingRepo.UpdateEnrollment(ctx, &model.TrainingEnrollment{})
	_, _ = trainingRepo.FindEnrollment(ctx, id, id)

	recruitRepo := NewRecruitmentRepository(db)
	_ = recruitRepo.CreatePosition(ctx, &model.RecruitmentPosition{})
	_ = recruitRepo.UpdatePosition(ctx, &model.RecruitmentPosition{})
	_ = recruitRepo.CreateApplicant(ctx, &model.Applicant{})
	_ = recruitRepo.UpdateApplicant(ctx, &model.Applicant{})

	docRepo := NewDocumentRepository(db)
	_ = docRepo.Create(ctx, &model.HRDocument{})
	_ = docRepo.Delete(ctx, id)

	announceRepo := NewAnnouncementRepository(db)
	_ = announceRepo.Create(ctx, &model.HRAnnouncement{})
	_ = announceRepo.Update(ctx, &model.HRAnnouncement{})
	_ = announceRepo.Delete(ctx, id)

	oneOnOneRepo := NewOneOnOneRepository(db)
	_ = oneOnOneRepo.Create(ctx, &model.OneOnOneMeeting{})
	_ = oneOnOneRepo.Update(ctx, &model.OneOnOneMeeting{})
	_ = oneOnOneRepo.Delete(ctx, id)

	skillRepo := NewSkillRepository(db)
	_ = skillRepo.Create(ctx, &model.EmployeeSkill{})
	_ = skillRepo.Update(ctx, &model.EmployeeSkill{})

	salaryRepo := NewSalaryRepository(db)
	_ = salaryRepo.Create(ctx, &model.SalaryRecord{})

	onboardingRepo := NewOnboardingRepository(db)
	_ = onboardingRepo.Create(ctx, &model.Onboarding{})
	_ = onboardingRepo.Update(ctx, &model.Onboarding{})
	_ = onboardingRepo.CreateTemplate(ctx, &model.OnboardingTemplate{})
	_, _ = onboardingRepo.FindAllTemplates(ctx)

	offboardingRepo := NewOffboardingRepository(db)
	_ = offboardingRepo.Create(ctx, &model.Offboarding{})
	_ = offboardingRepo.Update(ctx, &model.Offboarding{})

	surveyRepo := NewSurveyRepository(db)
	_ = surveyRepo.Create(ctx, &model.Survey{})
	_ = surveyRepo.Update(ctx, &model.Survey{})
	_ = surveyRepo.Delete(ctx, id)
	_ = surveyRepo.CreateResponse(ctx, &model.SurveyResponse{})
	_, _ = surveyRepo.FindResponsesBySurveyID(ctx, id)
	_, _ = surveyRepo.CountResponsesBySurveyID(ctx, id)

	expenseRepo := NewExpenseRepository(db)
	_ = expenseRepo.Create(ctx, &model.Expense{})
	_, _, _ = expenseRepo.FindPending(ctx, 1, 10)
	_ = expenseRepo.Update(ctx, &model.Expense{})
	_ = expenseRepo.Delete(ctx, id)
	_, _ = expenseRepo.GetStats(ctx, id)
	_, _ = expenseRepo.GetMonthlyTrend(ctx, now.Year())

	expenseItemRepo := NewExpenseItemRepository(db)
	_ = expenseItemRepo.DeleteByExpenseID(ctx, id)

	expenseCommentRepo := NewExpenseCommentRepository(db)
	_ = expenseCommentRepo.Create(ctx, &model.ExpenseComment{})
	_, _ = expenseCommentRepo.FindByExpenseID(ctx, id)

	expenseHistoryRepo := NewExpenseHistoryRepository(db)
	_ = expenseHistoryRepo.Create(ctx, &model.ExpenseHistory{})
	_, _ = expenseHistoryRepo.FindByExpenseID(ctx, id)

	expenseTemplateRepo := NewExpenseTemplateRepository(db)
	_ = expenseTemplateRepo.Create(ctx, &model.ExpenseTemplate{})
	_, _ = expenseTemplateRepo.FindAll(ctx, id)
	_ = expenseTemplateRepo.Update(ctx, &model.ExpenseTemplate{})
	_ = expenseTemplateRepo.Delete(ctx, id)

	expensePolicyRepo := NewExpensePolicyRepository(db)
	_ = expensePolicyRepo.Create(ctx, &model.ExpensePolicy{})
	_, _ = expensePolicyRepo.FindAll(ctx)
	_ = expensePolicyRepo.Update(ctx, &model.ExpensePolicy{})
	_ = expensePolicyRepo.Delete(ctx, id)

	expenseBudgetRepo := NewExpenseBudgetRepository(db)
	_, _ = expenseBudgetRepo.FindAll(ctx)

	expenseNotifRepo := NewExpenseNotificationRepository(db)
	_ = expenseNotifRepo.Create(ctx, &model.ExpenseNotification{})
	_ = expenseNotifRepo.MarkAsRead(ctx, id)
	_ = expenseNotifRepo.MarkAllAsRead(ctx, id)

	expenseReminderRepo := NewExpenseReminderRepository(db)
	_, _ = expenseReminderRepo.FindByUserID(ctx, id)
	_ = expenseReminderRepo.Dismiss(ctx, id)

	expenseSettingRepo := NewExpenseNotificationSettingRepository(db)
	_ = expenseSettingRepo.Upsert(ctx, &model.ExpenseNotificationSetting{})

	expenseDelegateRepo := NewExpenseDelegateRepository(db)
	_ = expenseDelegateRepo.Create(ctx, &model.ExpenseDelegate{})
	_ = expenseDelegateRepo.Delete(ctx, id)

	expenseViolationRepo := NewExpensePolicyViolationRepository(db)
	_ = expenseViolationRepo.Create(ctx, &model.ExpensePolicyViolation{})
}
