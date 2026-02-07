package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// Handlers は全ハンドラーを束ねる構造体
type Handlers struct {
	Auth       *AuthHandler
	Attendance *AttendanceHandler
	Leave      *LeaveHandler
	Shift      *ShiftHandler
	User       *UserHandler
	Department *DepartmentHandler
	Dashboard  *DashboardHandler
	Health     *HealthHandler
}

// NewHandlers は全ハンドラーを初期化する
func NewHandlers(services *service.Services, logger *logger.Logger) *Handlers {
	return &Handlers{
		Auth:       NewAuthHandler(services.Auth, logger),
		Attendance: NewAttendanceHandler(services.Attendance, logger),
		Leave:      NewLeaveHandler(services.Leave, logger),
		Shift:      NewShiftHandler(services.Shift, logger),
		User:       NewUserHandler(services.User, logger),
		Department: NewDepartmentHandler(services.Department, logger),
		Dashboard:  NewDashboardHandler(services.Dashboard, logger),
		Health:     NewHealthHandler(),
	}
}

// ===== ヘルパー関数 =====

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, service.ErrUnauthorized
	}
	return uuid.Parse(userIDStr.(string))
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "無効なIDフォーマットです",
		})
		return uuid.Nil, err
	}
	return id, nil
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func paginatedResponse(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// ===== HealthHandler =====

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// @Summary ヘルスチェック
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "kintai-api",
		"version": "1.0.0",
	})
}

// ===== AuthHandler =====

type AuthHandler struct {
	service service.AuthService
	logger  *logger.Logger
}

func NewAuthHandler(service service.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

// Login godoc
// @Summary ログイン
// @Tags auth
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "ログイン情報"
// @Success 200 {object} model.TokenResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}

	token, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

// Register godoc
// @Summary ユーザー登録
// @Tags auth
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "登録情報"
// @Success 201 {object} model.User
// @Failure 400 {object} model.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}

	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// RefreshToken godoc
// @Summary トークンリフレッシュ
// @Tags auth
// @Accept json
// @Produce json
// @Param body body model.RefreshTokenRequest true "リフレッシュトークン"
// @Success 200 {object} model.TokenResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}

	token, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

// Logout godoc
// @Summary ログアウト
// @Tags auth
// @Security BearerAuth
// @Success 204
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	if err := h.service.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "ログアウトに失敗しました"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ===== AttendanceHandler =====

type AttendanceHandler struct {
	service service.AttendanceService
	logger  *logger.Logger
}

func NewAttendanceHandler(service service.AttendanceService, logger *logger.Logger) *AttendanceHandler {
	return &AttendanceHandler{service: service, logger: logger}
}

// ClockIn godoc
// @Summary 出勤打刻
// @Tags attendance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.ClockInRequest false "出勤メモ"
// @Success 201 {object} model.Attendance
// @Router /attendance/clock-in [post]
func (h *AttendanceHandler) ClockIn(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	var req model.ClockInRequest
	_ = c.ShouldBindJSON(&req)

	attendance, err := h.service.ClockIn(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

// ClockOut godoc
// @Summary 退勤打刻
// @Tags attendance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.ClockOutRequest false "退勤メモ"
// @Success 200 {object} model.Attendance
// @Router /attendance/clock-out [post]
func (h *AttendanceHandler) ClockOut(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	var req model.ClockOutRequest
	_ = c.ShouldBindJSON(&req)

	attendance, err := h.service.ClockOut(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// GetMyAttendances godoc
// @Summary 自分の勤怠一覧取得
// @Tags attendance
// @Security BearerAuth
// @Produce json
// @Param start_date query string false "開始日 (YYYY-MM-DD)"
// @Param end_date query string false "終了日 (YYYY-MM-DD)"
// @Param page query int false "ページ番号"
// @Param page_size query int false "ページサイズ"
// @Success 200 {object} model.PaginatedResponse
// @Router /attendance [get]
func (h *AttendanceHandler) GetMyAttendances(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	start, end, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}

	page, pageSize := parsePagination(c)

	attendances, total, err := h.service.GetByUserAndDateRange(c.Request.Context(), userID, start, end, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "勤怠データの取得に失敗しました"})
		return
	}

	paginatedResponse(c, attendances, total, page, pageSize)
}

// GetTodayStatus godoc
// @Summary 今日の勤怠状況を取得
// @Tags attendance
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.Attendance
// @Router /attendance/today [get]
func (h *AttendanceHandler) GetTodayStatus(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	attendance, err := h.service.GetTodayStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "not_clocked_in"})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// GetSummary godoc
// @Summary 勤怠サマリーを取得
// @Tags attendance
// @Security BearerAuth
// @Produce json
// @Param start_date query string false "開始日"
// @Param end_date query string false "終了日"
// @Success 200 {object} model.AttendanceSummary
// @Router /attendance/summary [get]
func (h *AttendanceHandler) GetSummary(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	start, end, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}

	summary, err := h.service.GetSummary(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "サマリーの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// ===== LeaveHandler =====

type LeaveHandler struct {
	service service.LeaveService
	logger  *logger.Logger
}

func NewLeaveHandler(service service.LeaveService, logger *logger.Logger) *LeaveHandler {
	return &LeaveHandler{service: service, logger: logger}
}

// CreateLeaveRequest godoc
// @Summary 休暇申請を作成
// @Tags leaves
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.LeaveRequestCreate true "休暇申請情報"
// @Success 201 {object} model.LeaveRequest
// @Router /leaves [post]
func (h *LeaveHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	var req model.LeaveRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}

	leave, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, leave)
}

// ApproveLeaveRequest godoc
// @Summary 休暇申請を承認/却下
// @Tags leaves
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "休暇申請ID"
// @Param body body model.LeaveRequestApproval true "承認情報"
// @Success 200 {object} model.LeaveRequest
// @Router /leaves/{id}/approve [put]
func (h *LeaveHandler) Approve(c *gin.Context) {
	leaveID, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	approverID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	var req model.LeaveRequestApproval
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}

	leave, err := h.service.Approve(c.Request.Context(), leaveID, approverID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, leave)
}

// GetMyLeaves godoc
// @Summary 自分の休暇申請一覧を取得
// @Tags leaves
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.PaginatedResponse
// @Router /leaves [get]
func (h *LeaveHandler) GetMy(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	page, pageSize := parsePagination(c)
	leaves, total, err := h.service.GetByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}

	paginatedResponse(c, leaves, total, page, pageSize)
}

// GetPending godoc
// @Summary 未承認の休暇申請一覧を取得
// @Tags leaves
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.PaginatedResponse
// @Router /leaves/pending [get]
func (h *LeaveHandler) GetPending(c *gin.Context) {
	page, pageSize := parsePagination(c)
	leaves, total, err := h.service.GetPending(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}

	paginatedResponse(c, leaves, total, page, pageSize)
}

// ===== ShiftHandler =====

type ShiftHandler struct {
	service service.ShiftService
	logger  *logger.Logger
}

func NewShiftHandler(service service.ShiftService, logger *logger.Logger) *ShiftHandler {
	return &ShiftHandler{service: service, logger: logger}
}

// CreateShift godoc
// @Summary シフトを作成
// @Tags shifts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.ShiftCreateRequest true "シフト情報"
// @Success 201 {object} model.Shift
// @Router /shifts [post]
func (h *ShiftHandler) Create(c *gin.Context) {
	var req model.ShiftCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}

	shift, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shift)
}

// BulkCreateShifts godoc
// @Summary シフトを一括作成
// @Tags shifts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.ShiftBulkCreateRequest true "シフト一括情報"
// @Success 201 {object} map[string]string
// @Router /shifts/bulk [post]
func (h *ShiftHandler) BulkCreate(c *gin.Context) {
	var req model.ShiftBulkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}

	if err := h.service.BulkCreate(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "シフトを一括作成しました"})
}

// GetShifts godoc
// @Summary シフト一覧を取得
// @Tags shifts
// @Security BearerAuth
// @Produce json
// @Param start_date query string true "開始日"
// @Param end_date query string true "終了日"
// @Success 200 {array} model.Shift
// @Router /shifts [get]
func (h *ShiftHandler) GetByDateRange(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}

	shifts, err := h.service.GetByDateRange(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, shifts)
}

// DeleteShift godoc
// @Summary シフトを削除
// @Tags shifts
// @Security BearerAuth
// @Param id path string true "シフトID"
// @Success 204
// @Router /shifts/{id} [delete]
func (h *ShiftHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ===== UserHandler =====

type UserHandler struct {
	service service.UserService
	logger  *logger.Logger
}

func NewUserHandler(service service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

// GetMe godoc
// @Summary 自分のプロフィールを取得
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.User
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "ユーザーが見つかりません"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsers godoc
// @Summary 全ユーザー一覧を取得 (管理者用)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.PaginatedResponse
// @Router /users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	users, total, err := h.service.GetAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}

	paginatedResponse(c, users, total, page, pageSize)
}

// UpdateUser godoc
// @Summary ユーザー情報を更新 (管理者用)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ユーザーID"
// @Param body body model.UserUpdateRequest true "更新情報"
// @Success 200 {object} model.User
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req model.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}

	user, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser godoc
// @Summary ユーザーを作成 (管理者用)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body model.UserCreateRequest true "ユーザー情報"
// @Success 201 {object} model.User
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req model.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}

	user, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// DeleteUser godoc
// @Summary ユーザーを削除 (管理者用)
// @Tags users
// @Security BearerAuth
// @Param id path string true "ユーザーID"
// @Success 204
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ===== DepartmentHandler =====

type DepartmentHandler struct {
	service service.DepartmentService
	logger  *logger.Logger
}

func NewDepartmentHandler(service service.DepartmentService, logger *logger.Logger) *DepartmentHandler {
	return &DepartmentHandler{service: service, logger: logger}
}

// GetAllDepartments godoc
// @Summary 部署一覧を取得
// @Tags departments
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.Department
// @Router /departments [get]
func (h *DepartmentHandler) GetAll(c *gin.Context) {
	departments, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, departments)
}

// ===== DashboardHandler =====

type DashboardHandler struct {
	service service.DashboardService
	logger  *logger.Logger
}

func NewDashboardHandler(service service.DashboardService, logger *logger.Logger) *DashboardHandler {
	return &DashboardHandler{service: service, logger: logger}
}

// GetDashboardStats godoc
// @Summary ダッシュボード統計を取得
// @Tags dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.DashboardStats
// @Router /dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "統計の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
