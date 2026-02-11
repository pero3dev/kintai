package repository

import (
	appattendance "github.com/your-org/kintai/backend/internal/apps/attendance"
	"gorm.io/gorm"
)

type AttendanceRepository = appattendance.AttendanceRepository
type LeaveRequestRepository = appattendance.LeaveRequestRepository
type OvertimeRequestRepository = appattendance.OvertimeRequestRepository
type LeaveBalanceRepository = appattendance.LeaveBalanceRepository
type AttendanceCorrectionRepository = appattendance.AttendanceCorrectionRepository

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return appattendance.NewAttendanceRepository(db)
}

func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return appattendance.NewLeaveRequestRepository(db)
}

func NewOvertimeRequestRepository(db *gorm.DB) OvertimeRequestRepository {
	return appattendance.NewOvertimeRequestRepository(db)
}

func NewLeaveBalanceRepository(db *gorm.DB) LeaveBalanceRepository {
	return appattendance.NewLeaveBalanceRepository(db)
}

func NewAttendanceCorrectionRepository(db *gorm.DB) AttendanceCorrectionRepository {
	return appattendance.NewAttendanceCorrectionRepository(db)
}
