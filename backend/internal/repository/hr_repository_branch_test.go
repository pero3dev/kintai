package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestHRRepositoryFilterBranchesWithDryRun(t *testing.T) {
	ctx := context.Background()
	db := dryRunDB(t)

	empRepo := NewHREmployeeRepository(db)
	_, _, err := empRepo.FindAll(ctx, 1, 10, "", "", "", "")
	require.NoError(t, err)
	_, _, err = empRepo.FindAll(ctx, 1, 10, "eng", "active", "full_time", "alice")
	require.NoError(t, err)

	evalRepo := NewEvaluationRepository(db)
	_, _, err = evalRepo.FindAll(ctx, 1, 10, "", "")
	require.NoError(t, err)
	_, _, err = evalRepo.FindAll(ctx, 1, 10, "cycle-1", "submitted")
	require.NoError(t, err)

	goalRepo := NewGoalRepository(db)
	_, _, err = goalRepo.FindAll(ctx, 1, 10, "", "", "")
	require.NoError(t, err)
	_, _, err = goalRepo.FindAll(ctx, 1, 10, "in_progress", "performance", "emp-1")
	require.NoError(t, err)

	trainingRepo := NewTrainingRepository(db)
	_, _, err = trainingRepo.FindAll(ctx, 1, 10, "", "")
	require.NoError(t, err)
	_, _, err = trainingRepo.FindAll(ctx, 1, 10, "security", "scheduled")
	require.NoError(t, err)

	recruitmentRepo := NewRecruitmentRepository(db)
	_, _, err = recruitmentRepo.FindAllPositions(ctx, 1, 10, "", "")
	require.NoError(t, err)
	_, _, err = recruitmentRepo.FindAllPositions(ctx, 1, 10, "open", "eng")
	require.NoError(t, err)

	_, err = recruitmentRepo.FindAllApplicants(ctx, "", "")
	require.NoError(t, err)
	_, err = recruitmentRepo.FindAllApplicants(ctx, "pos-1", "screening")
	require.NoError(t, err)

	docRepo := NewDocumentRepository(db)
	_, _, err = docRepo.FindAll(ctx, 1, 10, "", "")
	require.NoError(t, err)
	_, _, err = docRepo.FindAll(ctx, 1, 10, "contract", "emp-1")
	require.NoError(t, err)

	announceRepo := NewAnnouncementRepository(db)
	_, _, err = announceRepo.FindAll(ctx, 1, 10, "")
	require.NoError(t, err)
	_, _, err = announceRepo.FindAll(ctx, 1, 10, "high")
	require.NoError(t, err)

	oneOnOneRepo := NewOneOnOneRepository(db)
	_, err = oneOnOneRepo.FindAll(ctx, "", "")
	require.NoError(t, err)
	_, err = oneOnOneRepo.FindAll(ctx, "scheduled", "emp-1")
	require.NoError(t, err)

	skillRepo := NewSkillRepository(db)
	_, err = skillRepo.FindAll(ctx, "")
	require.NoError(t, err)
	_, err = skillRepo.FindAll(ctx, "eng")
	require.NoError(t, err)

	salaryRepo := NewSalaryRepository(db)
	_, err = salaryRepo.GetOverview(ctx, "")
	require.NoError(t, err)
	_, err = salaryRepo.GetOverview(ctx, "eng")
	require.NoError(t, err)

	onboardingRepo := NewOnboardingRepository(db)
	_, err = onboardingRepo.FindAll(ctx, "")
	require.NoError(t, err)
	_, err = onboardingRepo.FindAll(ctx, "in_progress")
	require.NoError(t, err)

	offboardingRepo := NewOffboardingRepository(db)
	_, err = offboardingRepo.FindAll(ctx, "")
	require.NoError(t, err)
	_, err = offboardingRepo.FindAll(ctx, "completed")
	require.NoError(t, err)

	surveyRepo := NewSurveyRepository(db)
	_, err = surveyRepo.FindAll(ctx, "", "")
	require.NoError(t, err)
	_, err = surveyRepo.FindAll(ctx, "active", "engagement")
	require.NoError(t, err)
}

func TestSkillRepositoryGetGapAnalysisBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("rows error", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewSkillRepository(db)

		expectQuery(mock, `(?i)SELECT .*FROM "employee_skills"`, 1).WillReturnError(errors.New("boom"))
		_, err := repo.GetGapAnalysis(ctx, "eng")
		require.Error(t, err)
	})

	t.Run("scan continue and append", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewSkillRepository(db)

		rows := sqlmock.NewRows([]string{"skill_name", "category", "current_avg", "employee_count"}).
			AddRow("Go", "tech", nil, 1).
			AddRow("Go", "tech", 3.5, 2)
		expectQuery(mock, `(?i)SELECT .*FROM "employee_skills"`, 0).WillReturnRows(rows)

		got, err := repo.GetGapAnalysis(ctx, "")
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestOffboardingRepositoryGetTurnoverAnalyticsBranch(t *testing.T) {
	ctx := context.Background()

	t.Run("total employees zero", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewOffboardingRepository(db)

		expectQuery(mock, `(?i)SELECT count\(\*\).*FROM "offboardings"`, 0).WillReturnRows(countRows(2))
		expectQuery(mock, `(?i)SELECT count\(\*\).*FROM "hr_employees"`, 0).WillReturnRows(countRows(0))
		expectQuery(mock, `(?i)SELECT .*reason, COUNT\(\*\) as count.*FROM "offboardings"`, 0).
			WillReturnRows(sqlmock.NewRows([]string{"reason", "count"}))

		got, err := repo.GetTurnoverAnalytics(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 0, got["turnover_rate"])
	})

	t.Run("total employees positive", func(t *testing.T) {
		db, mock, cleanup := newMockDB(t)
		defer cleanup()
		repo := NewOffboardingRepository(db)

		expectQuery(mock, `(?i)SELECT count\(\*\).*FROM "offboardings"`, 0).WillReturnRows(countRows(2))
		expectQuery(mock, `(?i)SELECT count\(\*\).*FROM "hr_employees"`, 0).WillReturnRows(countRows(10))
		expectQuery(mock, `(?i)SELECT .*reason, COUNT\(\*\) as count.*FROM "offboardings"`, 0).
			WillReturnRows(sqlmock.NewRows([]string{"reason", "count"}).AddRow("resignation", 2))

		got, err := repo.GetTurnoverAnalytics(ctx)
		require.NoError(t, err)
		require.Equal(t, 20.0, got["turnover_rate"])
	})
}

