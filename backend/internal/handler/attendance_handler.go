package handler

import (
	appattendance "github.com/your-org/kintai/backend/internal/apps/attendance"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

type AttendanceHandler = appattendance.AttendanceHandler
type LeaveHandler = appattendance.LeaveHandler
type OvertimeRequestHandler = appattendance.OvertimeRequestHandler
type LeaveBalanceHandler = appattendance.LeaveBalanceHandler
type AttendanceCorrectionHandler = appattendance.AttendanceCorrectionHandler

func NewAttendanceHandler(svc service.AttendanceService, logger *logger.Logger) *AttendanceHandler {
	return appattendance.NewAttendanceHandler(svc, logger)
}

func NewLeaveHandler(svc service.LeaveService, logger *logger.Logger) *LeaveHandler {
	return appattendance.NewLeaveHandler(svc, logger)
}

func NewOvertimeRequestHandler(svc service.OvertimeRequestService, logger *logger.Logger) *OvertimeRequestHandler {
	return appattendance.NewOvertimeRequestHandler(svc, logger)
}

func NewLeaveBalanceHandler(svc service.LeaveBalanceService, logger *logger.Logger) *LeaveBalanceHandler {
	return appattendance.NewLeaveBalanceHandler(svc, logger)
}

func NewAttendanceCorrectionHandler(svc service.AttendanceCorrectionService, logger *logger.Logger) *AttendanceCorrectionHandler {
	return appattendance.NewAttendanceCorrectionHandler(svc, logger)
}
