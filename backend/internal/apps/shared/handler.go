package shared

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// Handlers は共通ハンドラーを束ねる構造体
type Handlers struct {
	Auth         *AuthHandler
	User         *UserHandler
	Department   *DepartmentHandler
	Dashboard    *DashboardHandler
	Health       *HealthHandler
	Shift        *ShiftHandler
	Notification *NotificationHandler
	Project      *ProjectHandler
	TimeEntry    *TimeEntryHandler
	Holiday      *HolidayHandler
	ApprovalFlow *ApprovalFlowHandler
	Export       *ExportHandler
}

// NewHandlers は共通ハンドラーを初期化する
func NewHandlers(services *Services, logger *logger.Logger) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(services.Auth, logger),
		User:         NewUserHandler(services.User, logger),
		Department:   NewDepartmentHandler(services.Department, logger),
		Dashboard:    NewDashboardHandler(services.Dashboard, logger),
		Health:       NewHealthHandler(),
		Shift:        NewShiftHandler(services.Shift, logger),
		Notification: NewNotificationHandler(services.Notification, logger),
		Project:      NewProjectHandler(services.Project, logger),
		TimeEntry:    NewTimeEntryHandler(services.TimeEntry, logger),
		Holiday:      NewHolidayHandler(services.Holiday, logger),
		ApprovalFlow: NewApprovalFlowHandler(services.ApprovalFlow, logger),
		Export:       NewExportHandler(services.Export, logger),
	}
}

// ===== エクスポートされたヘルパー関数 =====

func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.Parse(userIDStr.(string))
}

func ParseUUID(c *gin.Context, param string) (uuid.UUID, error) {
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

func ParsePagination(c *gin.Context) (int, int) {
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

func ParseDateRange(c *gin.Context) (time.Time, time.Time, error) {
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

func PaginatedResponse(c *gin.Context, data interface{}, total int64, page, pageSize int) {
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

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "kintai-api",
		"version": "1.0.0",
	})
}

// ===== AuthHandler =====

type AuthHandler struct {
	service AuthService
	logger  *logger.Logger
}

func NewAuthHandler(service AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

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

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
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

// ===== ShiftHandler =====

type ShiftHandler struct {
	service ShiftService
	logger  *logger.Logger
}

func NewShiftHandler(service ShiftService, logger *logger.Logger) *ShiftHandler {
	return &ShiftHandler{service: service, logger: logger}
}

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

func (h *ShiftHandler) GetByDateRange(c *gin.Context) {
	start, end, err := ParseDateRange(c)
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

func (h *ShiftHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
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
	service UserService
	logger  *logger.Logger
}

func NewUserHandler(service UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
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

func (h *UserHandler) GetAll(c *gin.Context) {
	page, pageSize := ParsePagination(c)
	users, total, err := h.service.GetAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	PaginatedResponse(c, users, total, page, pageSize)
}

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

func (h *UserHandler) Update(c *gin.Context) {
	id, err := ParseUUID(c, "id")
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

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
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
	service DepartmentService
	logger  *logger.Logger
}

func NewDepartmentHandler(service DepartmentService, logger *logger.Logger) *DepartmentHandler {
	return &DepartmentHandler{service: service, logger: logger}
}

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
	service DashboardService
	logger  *logger.Logger
}

func NewDashboardHandler(service DashboardService, logger *logger.Logger) *DashboardHandler {
	return &DashboardHandler{service: service, logger: logger}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "統計の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ===== NotificationHandler =====

type NotificationHandler struct {
	service NotificationService
	logger  *logger.Logger
}

func NewNotificationHandler(service NotificationService, logger *logger.Logger) *NotificationHandler {
	return &NotificationHandler{service: service, logger: logger}
}

func (h *NotificationHandler) GetMy(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}
	page, pageSize := ParsePagination(c)
	var isRead *bool
	if r := c.Query("is_read"); r != "" {
		v := r == "true"
		isRead = &v
	}
	notifications, total, err := h.service.GetByUser(c.Request.Context(), userID, isRead, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	PaginatedResponse(c, notifications, total, page, pageSize)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}
	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, model.NotificationCount{Unread: int(count)})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.service.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}
	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "更新に失敗しました"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== ProjectHandler =====

type ProjectHandler struct {
	service ProjectService
	logger  *logger.Logger
}

func NewProjectHandler(service ProjectService, logger *logger.Logger) *ProjectHandler {
	return &ProjectHandler{service: service, logger: logger}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req model.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}
	project, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) GetAll(c *gin.Context) {
	page, pageSize := ParsePagination(c)
	var status *model.ProjectStatus
	if s := c.Query("status"); s != "" {
		ps := model.ProjectStatus(s)
		status = &ps
	}
	projects, total, err := h.service.GetAll(c.Request.Context(), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	PaginatedResponse(c, projects, total, page, pageSize)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	project, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "プロジェクトが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	project, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== TimeEntryHandler =====

type TimeEntryHandler struct {
	service TimeEntryService
	logger  *logger.Logger
}

func NewTimeEntryHandler(service TimeEntryService, logger *logger.Logger) *TimeEntryHandler {
	return &TimeEntryHandler{service: service, logger: logger}
}

func (h *TimeEntryHandler) Create(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}
	var req model.TimeEntryCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}
	entry, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *TimeEntryHandler) GetMy(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	entries, err := h.service.GetByUserAndDateRange(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *TimeEntryHandler) GetByProject(c *gin.Context) {
	projectID, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	entries, err := h.service.GetByProjectAndDateRange(c.Request.Context(), projectID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *TimeEntryHandler) Update(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.TimeEntryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	entry, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *TimeEntryHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TimeEntryHandler) GetSummary(c *gin.Context) {
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	summaries, err := h.service.GetProjectSummary(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, summaries)
}

// ===== HolidayHandler =====

type HolidayHandler struct {
	service HolidayService
	logger  *logger.Logger
}

func NewHolidayHandler(service HolidayService, logger *logger.Logger) *HolidayHandler {
	return &HolidayHandler{service: service, logger: logger}
}

func (h *HolidayHandler) Create(c *gin.Context) {
	var req model.HolidayCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}
	holiday, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, holiday)
}

func (h *HolidayHandler) GetByYear(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	holidays, err := h.service.GetByYear(c.Request.Context(), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, holidays)
}

func (h *HolidayHandler) Update(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.HolidayUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	holiday, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, holiday)
}

func (h *HolidayHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HolidayHandler) GetCalendar(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	calendar, err := h.service.GetCalendar(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, calendar)
}

func (h *HolidayHandler) GetWorkingDays(c *gin.Context) {
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	summary, err := h.service.GetWorkingDays(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ===== ApprovalFlowHandler =====

type ApprovalFlowHandler struct {
	service ApprovalFlowService
	logger  *logger.Logger
}

func NewApprovalFlowHandler(service ApprovalFlowService, logger *logger.Logger) *ApprovalFlowHandler {
	return &ApprovalFlowHandler{service: service, logger: logger}
}

func (h *ApprovalFlowHandler) Create(c *gin.Context) {
	var req model.ApprovalFlowCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}
	flow, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, flow)
}

func (h *ApprovalFlowHandler) GetAll(c *gin.Context) {
	flows, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, flows)
}

func (h *ApprovalFlowHandler) GetByID(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	flow, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "承認フローが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, flow)
}

func (h *ApprovalFlowHandler) Update(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.ApprovalFlowUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	flow, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, flow)
}

func (h *ApprovalFlowHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "削除に失敗しました"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== ExportHandler =====

type ExportHandler struct {
	service ExportService
	logger  *logger.Logger
}

func NewExportHandler(service ExportService, logger *logger.Logger) *ExportHandler {
	return &ExportHandler{service: service, logger: logger}
}

func (h *ExportHandler) ExportAttendance(c *gin.Context) {
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	var userID *uuid.UUID
	if uid := c.Query("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err == nil {
			userID = &id
		}
	}
	data, err := h.service.ExportAttendanceCSV(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "エクスポートに失敗しました"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=attendance.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *ExportHandler) ExportLeaves(c *gin.Context) {
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	var userID *uuid.UUID
	if uid := c.Query("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err == nil {
			userID = &id
		}
	}
	data, err := h.service.ExportLeavesCSV(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "エクスポートに失敗しました"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=leaves.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *ExportHandler) ExportOvertime(c *gin.Context) {
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	data, err := h.service.ExportOvertimeCSV(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "エクスポートに失敗しました"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=overtime.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *ExportHandler) ExportProjects(c *gin.Context) {
	start, end, err := ParseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "日付フォーマットが不正です"})
		return
	}
	data, err := h.service.ExportProjectsCSV(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "エクスポートに失敗しました"})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=projects.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}
