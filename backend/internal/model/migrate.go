package model

import "gorm.io/gorm"

// AutoMigrate はデータベースのマイグレーションを実行する
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Department{},
		&Attendance{},
		&LeaveRequest{},
		&Shift{},
		&RefreshToken{},
		&OvertimeRequest{},
		&LeaveBalance{},
		&AttendanceCorrection{},
		&Notification{},
		&Project{},
		&TimeEntry{},
		&Holiday{},
		&ApprovalFlow{},
		&ApprovalStep{},
	)
}

// HRAutoMigrate はHR関連テーブルのマイグレーションを実行する
func HRAutoMigrate(db *gorm.DB) error {
	// FK制約の自動生成を無効化してマイグレーション（循環参照を避ける）
	prev := db.Config.DisableForeignKeyConstraintWhenMigrating
	db.Config.DisableForeignKeyConstraintWhenMigrating = true
	defer func() { db.Config.DisableForeignKeyConstraintWhenMigrating = prev }()

	return db.AutoMigrate(
		&HRDepartment{},
		&HREmployee{},
		&EvaluationCycle{},
		&Evaluation{},
		&HRGoal{},
		&TrainingProgram{},
		&TrainingEnrollment{},
		&RecruitmentPosition{},
		&Applicant{},
		&HRDocument{},
		&HRAnnouncement{},
		&OneOnOneMeeting{},
		&EmployeeSkill{},
		&SalaryRecord{},
		&Onboarding{},
		&OnboardingTemplate{},
		&Offboarding{},
		&Survey{},
		&SurveyResponse{},
	)
}

// ExpenseAutoMigrate は経費関連テーブルのマイグレーションを実行する
func ExpenseAutoMigrate(db *gorm.DB) error {
	// FK制約の自動生成を無効化してマイグレーション
	prev := db.Config.DisableForeignKeyConstraintWhenMigrating
	db.Config.DisableForeignKeyConstraintWhenMigrating = true
	defer func() { db.Config.DisableForeignKeyConstraintWhenMigrating = prev }()

	models := []interface{}{
		&Expense{},
		&ExpenseItem{},
		&ExpenseComment{},
		&ExpenseHistory{},
		&ExpenseTemplate{},
		&ExpensePolicy{},
		&ExpenseBudget{},
		&ExpenseNotification{},
		&ExpenseReminder{},
		&ExpenseNotificationSetting{},
		&ExpenseApprovalFlow{},
		&ExpenseDelegate{},
		&ExpensePolicyViolation{},
	}

	m := db.Migrator()
	for _, model := range models {
		if !m.HasTable(model) {
			if err := m.CreateTable(model); err != nil {
				return err
			}
		}
	}
	return nil
}
