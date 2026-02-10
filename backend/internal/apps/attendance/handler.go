package attendance

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

type AttendanceHandler struct {
	svc    AttendanceService
	logger *logger.Logger
}

func NewAttendanceHandler(svc AttendanceService, logger *logger.Logger) *AttendanceHandler {
	return &AttendanceHandler{svc: svc, logger: logger}
}

func (h *AttendanceHandler) ClockIn(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	var req model.ClockInRequest
	_ = c.ShouldBindJSON(&req)

	attendance, err := h.svc.ClockIn(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

func (h *AttendanceHandler) ClockOut(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	var req model.ClockOutRequest
	_ = c.ShouldBindJSON(&req)

	attendance, err := h.svc.ClockOut(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

func (h *AttendanceHandler) GetMyAttendances(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	start, end, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid date range format"})
		return
	}

	page, pageSize := parsePagination(c)
	attendances, total, err := h.svc.GetByUserAndDateRange(c.Request.Context(), userID, start, end, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	paginatedResponse(c, attendances, total, page, pageSize)
}

func (h *AttendanceHandler) GetTodayStatus(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	attendance, err := h.svc.GetTodayStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "not_clocked_in"})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

func (h *AttendanceHandler) GetSummary(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	start, end, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid date range format"})
		return
	}

	summary, err := h.svc.GetSummary(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

type LeaveHandler struct {
	svc    LeaveService
	logger *logger.Logger
}

func NewLeaveHandler(svc LeaveService, logger *logger.Logger) *LeaveHandler {
	return &LeaveHandler{svc: svc, logger: logger}
}

func (h *LeaveHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	var req model.LeaveRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request", Details: err.Error()})
		return
	}

	leave, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, leave)
}

func (h *LeaveHandler) Approve(c *gin.Context) {
	leaveID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	approverID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	var req model.LeaveRequestApproval
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request"})
		return
	}

	leave, err := h.svc.Approve(c.Request.Context(), leaveID, approverID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, leave)
}

func (h *LeaveHandler) GetMy(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}

	page, pageSize := parsePagination(c)
	leaves, total, err := h.svc.GetByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	paginatedResponse(c, leaves, total, page, pageSize)
}

func (h *LeaveHandler) GetPending(c *gin.Context) {
	page, pageSize := parsePagination(c)
	leaves, total, err := h.svc.GetPending(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	paginatedResponse(c, leaves, total, page, pageSize)
}

type OvertimeRequestHandler struct {
	svc    OvertimeRequestService
	logger *logger.Logger
}

func NewOvertimeRequestHandler(svc OvertimeRequestService, logger *logger.Logger) *OvertimeRequestHandler {
	return &OvertimeRequestHandler{svc: svc, logger: logger}
}

func (h *OvertimeRequestHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	var req model.OvertimeRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request", Details: err.Error()})
		return
	}
	overtime, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, overtime)
}

func (h *OvertimeRequestHandler) Approve(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	approverID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	var req model.OvertimeRequestApproval
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request"})
		return
	}
	overtime, err := h.svc.Approve(c.Request.Context(), id, approverID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, overtime)
}

func (h *OvertimeRequestHandler) GetMy(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	page, pageSize := parsePagination(c)
	overtimes, total, err := h.svc.GetByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	paginatedResponse(c, overtimes, total, page, pageSize)
}

func (h *OvertimeRequestHandler) GetPending(c *gin.Context) {
	page, pageSize := parsePagination(c)
	overtimes, total, err := h.svc.GetPending(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	paginatedResponse(c, overtimes, total, page, pageSize)
}

func (h *OvertimeRequestHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.svc.GetOvertimeAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

type LeaveBalanceHandler struct {
	svc    LeaveBalanceService
	logger *logger.Logger
}

func NewLeaveBalanceHandler(svc LeaveBalanceService, logger *logger.Logger) *LeaveBalanceHandler {
	return &LeaveBalanceHandler{svc: svc, logger: logger}
}

func (h *LeaveBalanceHandler) GetMy(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("fiscal_year", strconv.Itoa(time.Now().Year())))
	balances, err := h.svc.GetByUser(c.Request.Context(), userID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, balances)
}

func (h *LeaveBalanceHandler) GetByUser(c *gin.Context) {
	userID, err := parseUUID(c, "user_id")
	if err != nil {
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("fiscal_year", strconv.Itoa(time.Now().Year())))
	balances, err := h.svc.GetByUser(c.Request.Context(), userID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, balances)
}

func (h *LeaveBalanceHandler) SetBalance(c *gin.Context) {
	userID, err := parseUUID(c, "user_id")
	if err != nil {
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("fiscal_year", strconv.Itoa(time.Now().Year())))
	leaveType := model.LeaveType(c.Param("leave_type"))
	var req model.LeaveBalanceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request"})
		return
	}
	if err := h.svc.SetBalance(c.Request.Context(), userID, year, leaveType, &req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *LeaveBalanceHandler) Initialize(c *gin.Context) {
	userID, err := parseUUID(c, "user_id")
	if err != nil {
		return
	}
	year, _ := strconv.Atoi(c.DefaultQuery("fiscal_year", strconv.Itoa(time.Now().Year())))
	if err := h.svc.InitializeForUser(c.Request.Context(), userID, year); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "initialized"})
}

type AttendanceCorrectionHandler struct {
	svc    AttendanceCorrectionService
	logger *logger.Logger
}

func NewAttendanceCorrectionHandler(svc AttendanceCorrectionService, logger *logger.Logger) *AttendanceCorrectionHandler {
	return &AttendanceCorrectionHandler{svc: svc, logger: logger}
}

func (h *AttendanceCorrectionHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	var req model.AttendanceCorrectionCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request", Details: err.Error()})
		return
	}
	correction, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, correction)
}

func (h *AttendanceCorrectionHandler) Approve(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	approverID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	var req model.AttendanceCorrectionApproval
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "invalid request"})
		return
	}
	correction, err := h.svc.Approve(c.Request.Context(), id, approverID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, correction)
}

func (h *AttendanceCorrectionHandler) GetMy(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "unauthorized"})
		return
	}
	page, pageSize := parsePagination(c)
	corrections, total, err := h.svc.GetByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	paginatedResponse(c, corrections, total, page, pageSize)
}

func (h *AttendanceCorrectionHandler) GetPending(c *gin.Context) {
	page, pageSize := parsePagination(c)
	corrections, total, err := h.svc.GetPending(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	paginatedResponse(c, corrections, total, page, pageSize)
}
