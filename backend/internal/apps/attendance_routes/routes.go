package attendance_routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/model"
)

func RegisterProtectedRoutes(protected *gin.RouterGroup, h *handler.Handlers, mw *middleware.Middleware) {
	attendance := protected.Group("/attendance")
	{
		attendance.POST("/clock-in", h.Attendance.ClockIn)
		attendance.POST("/clock-out", h.Attendance.ClockOut)
		attendance.GET("", h.Attendance.GetMyAttendances)
		attendance.GET("/today", h.Attendance.GetTodayStatus)
		attendance.GET("/summary", h.Attendance.GetSummary)
	}

	leaves := protected.Group("/leaves")
	{
		leaves.POST("", h.Leave.Create)
		leaves.GET("", h.Leave.GetMy)
	}

	overtime := protected.Group("/overtime")
	{
		overtime.POST("", h.OvertimeRequest.Create)
		overtime.GET("", h.OvertimeRequest.GetMy)
	}

	corrections := protected.Group("/corrections")
	{
		corrections.POST("", h.AttendanceCorrection.Create)
		corrections.GET("", h.AttendanceCorrection.GetMy)
	}

	leaveBalance := protected.Group("/leave-balances")
	{
		leaveBalance.GET("", h.LeaveBalance.GetMy)
	}

	admin := protected.Group("")
	admin.Use(mw.RequireRole(model.RoleAdmin, model.RoleManager))
	{
		admin.GET("/leaves/pending", h.Leave.GetPending)
		admin.PUT("/leaves/:id/approve", h.Leave.Approve)

		admin.GET("/overtime/pending", h.OvertimeRequest.GetPending)
		admin.PUT("/overtime/:id/approve", h.OvertimeRequest.Approve)
		admin.GET("/overtime/alerts", h.OvertimeRequest.GetAlerts)

		admin.GET("/corrections/pending", h.AttendanceCorrection.GetPending)
		admin.PUT("/corrections/:id/approve", h.AttendanceCorrection.Approve)

		admin.GET("/leave-balances/:user_id", h.LeaveBalance.GetByUser)
		admin.PUT("/leave-balances/:user_id/:leave_type", h.LeaveBalance.SetBalance)
		admin.POST("/leave-balances/:user_id/initialize", h.LeaveBalance.Initialize)
	}
}
