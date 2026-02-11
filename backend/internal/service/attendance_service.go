package service

import appattendance "github.com/your-org/kintai/backend/internal/apps/attendance"

type AttendanceService = appattendance.AttendanceService
type LeaveService = appattendance.LeaveService
type OvertimeRequestService = appattendance.OvertimeRequestService
type LeaveBalanceService = appattendance.LeaveBalanceService
type AttendanceCorrectionService = appattendance.AttendanceCorrectionService

func toAttendanceDeps(deps Deps) appattendance.Deps {
	return appattendance.Deps{
		Repos: &appattendance.Repositories{
			User:                 deps.Repos.User,
			Attendance:           deps.Repos.Attendance,
			LeaveRequest:         deps.Repos.LeaveRequest,
			OvertimeRequest:      deps.Repos.OvertimeRequest,
			AttendanceCorrection: deps.Repos.AttendanceCorrection,
			LeaveBalance:         deps.Repos.LeaveBalance,
		},
		Config: deps.Config,
		Logger: deps.Logger,
	}
}

func NewAttendanceService(deps Deps) AttendanceService {
	return appattendance.NewAttendanceService(toAttendanceDeps(deps))
}

func NewLeaveService(deps Deps, notificationSvc NotificationService) LeaveService {
	return appattendance.NewLeaveService(toAttendanceDeps(deps), notificationSvc)
}

func NewOvertimeRequestService(deps Deps, notificationSvc NotificationService) OvertimeRequestService {
	return appattendance.NewOvertimeRequestService(toAttendanceDeps(deps), notificationSvc)
}

func NewLeaveBalanceService(deps Deps) LeaveBalanceService {
	return appattendance.NewLeaveBalanceService(toAttendanceDeps(deps))
}

func NewAttendanceCorrectionService(deps Deps, notificationSvc NotificationService) AttendanceCorrectionService {
	return appattendance.NewAttendanceCorrectionService(toAttendanceDeps(deps), notificationSvc)
}
