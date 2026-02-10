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
	Auth                 *AuthHandler
	Attendance           *AttendanceHandler
	Leave                *LeaveHandler
	Shift                *ShiftHandler
	User                 *UserHandler
	Department           *DepartmentHandler
	Dashboard            *DashboardHandler
	Health               *HealthHandler
	OvertimeRequest      *OvertimeRequestHandler
	LeaveBalance         *LeaveBalanceHandler
	AttendanceCorrection *AttendanceCorrectionHandler
	Notification         *NotificationHandler
	Project              *ProjectHandler
	TimeEntry            *TimeEntryHandler
	Holiday              *HolidayHandler
	ApprovalFlow         *ApprovalFlowHandler
	Export               *ExportHandler
	// HR
	HREmployee            *HREmployeeHandler
	HRDepartment          *HRDepartmentHandler
	Evaluation            *EvaluationHandler
	Goal                  *GoalHandler
	Training              *TrainingHandler
	Recruitment           *RecruitmentHandler
	Document              *DocumentHandler
	Announcement          *AnnouncementHandler
	HRDashboard           *HRDashboardHandler
	AttendanceIntegration *AttendanceIntegrationHandler
	OrgChart              *OrgChartHandler
	OneOnOne              *OneOnOneHandler
	Skill                 *SkillHandler
	Salary                *SalaryHandler
	Onboarding            *OnboardingHandler
	Offboarding           *OffboardingHandler
	Survey                *SurveyHandler
	// Expense
	Expense             *ExpenseHandler
	ExpenseComment      *ExpenseCommentHandler
	ExpenseHistory      *ExpenseHistoryHandler
	ExpenseReceipt      *ExpenseReceiptHandler
	ExpenseTemplate     *ExpenseTemplateHandler
	ExpensePolicy       *ExpensePolicyHandler
	ExpenseNotification *ExpenseNotificationHandler
	ExpenseApprovalFlow *ExpenseApprovalFlowHandler
}

// NewHandlers は全ハンドラーを初期化する
func NewHandlers(services *service.Services, logger *logger.Logger) *Handlers {
	return &Handlers{
		Auth:                 NewAuthHandler(services.Auth, logger),
		Attendance:           NewAttendanceHandler(services.Attendance, logger),
		Leave:                NewLeaveHandler(services.Leave, logger),
		Shift:                NewShiftHandler(services.Shift, logger),
		User:                 NewUserHandler(services.User, logger),
		Department:           NewDepartmentHandler(services.Department, logger),
		Dashboard:            NewDashboardHandler(services.Dashboard, logger),
		Health:               NewHealthHandler(),
		OvertimeRequest:      NewOvertimeRequestHandler(services.OvertimeRequest, logger),
		LeaveBalance:         NewLeaveBalanceHandler(services.LeaveBalance, logger),
		AttendanceCorrection: NewAttendanceCorrectionHandler(services.AttendanceCorrection, logger),
		Notification:         NewNotificationHandler(services.Notification, logger),
		Project:              NewProjectHandler(services.Project, logger),
		TimeEntry:            NewTimeEntryHandler(services.TimeEntry, logger),
		Holiday:              NewHolidayHandler(services.Holiday, logger),
		ApprovalFlow:         NewApprovalFlowHandler(services.ApprovalFlow, logger),
		Export:               NewExportHandler(services.Export, logger),
		// HR
		HREmployee:            NewHREmployeeHandler(services.HREmployee, logger),
		HRDepartment:          NewHRDepartmentHandler(services.HRDepartment, logger),
		Evaluation:            NewEvaluationHandler(services.Evaluation, logger),
		Goal:                  NewGoalHandler(services.Goal, logger),
		Training:              NewTrainingHandler(services.Training, logger),
		Recruitment:           NewRecruitmentHandler(services.Recruitment, logger),
		Document:              NewDocumentHandler(services.Document, logger),
		Announcement:          NewAnnouncementHandler(services.Announcement, logger),
		HRDashboard:           NewHRDashboardHandler(services.HRDashboard, logger),
		AttendanceIntegration: NewAttendanceIntegrationHandler(services.AttendanceIntegration, logger),
		OrgChart:              NewOrgChartHandler(services.OrgChart, logger),
		OneOnOne:              NewOneOnOneHandler(services.OneOnOne, logger),
		Skill:                 NewSkillHandler(services.Skill, logger),
		Salary:                NewSalaryHandler(services.Salary, logger),
		Onboarding:            NewOnboardingHandler(services.Onboarding, logger),
		Offboarding:           NewOffboardingHandler(services.Offboarding, logger),
		Survey:                NewSurveyHandler(services.Survey, logger),
		// Expense
		Expense:             NewExpenseHandler(services.Expense, logger),
		ExpenseComment:      NewExpenseCommentHandler(services.ExpenseComment, logger),
		ExpenseHistory:      NewExpenseHistoryHandler(services.ExpenseHistory, logger),
		ExpenseReceipt:      NewExpenseReceiptHandler(services.ExpenseReceipt, logger),
		ExpenseTemplate:     NewExpenseTemplateHandler(services.ExpenseTemplate, logger),
		ExpensePolicy:       NewExpensePolicyHandler(services.ExpensePolicy, logger),
		ExpenseNotification: NewExpenseNotificationHandler(services.ExpenseNotification, logger),
		ExpenseApprovalFlow: NewExpenseApprovalFlowHandler(services.ExpenseApprovalFlow, logger),
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

// ===== NotificationHandler =====

type NotificationHandler struct {
	service service.NotificationService
	logger  *logger.Logger
}

func NewNotificationHandler(service service.NotificationService, logger *logger.Logger) *NotificationHandler {
	return &NotificationHandler{service: service, logger: logger}
}

func (h *NotificationHandler) GetMy(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証が必要です"})
		return
	}
	page, pageSize := parsePagination(c)
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
	paginatedResponse(c, notifications, total, page, pageSize)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
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
	id, err := parseUUID(c, "id")
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
	userID, err := getUserIDFromContext(c)
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

// ===== ProjectHandler =====

type ProjectHandler struct {
	service service.ProjectService
	logger  *logger.Logger
}

func NewProjectHandler(service service.ProjectService, logger *logger.Logger) *ProjectHandler {
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
	page, pageSize := parsePagination(c)
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
	paginatedResponse(c, projects, total, page, pageSize)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
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
	id, err := parseUUID(c, "id")
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

// ===== TimeEntryHandler =====

type TimeEntryHandler struct {
	service service.TimeEntryService
	logger  *logger.Logger
}

func NewTimeEntryHandler(service service.TimeEntryService, logger *logger.Logger) *TimeEntryHandler {
	return &TimeEntryHandler{service: service, logger: logger}
}

func (h *TimeEntryHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
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
	entries, err := h.service.GetByUserAndDateRange(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *TimeEntryHandler) GetByProject(c *gin.Context) {
	projectID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	start, end, err := parseDateRange(c)
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
	id, err := parseUUID(c, "id")
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

func (h *TimeEntryHandler) GetSummary(c *gin.Context) {
	start, end, err := parseDateRange(c)
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
	service service.HolidayService
	logger  *logger.Logger
}

func NewHolidayHandler(service service.HolidayService, logger *logger.Logger) *HolidayHandler {
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
	id, err := parseUUID(c, "id")
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
	start, end, err := parseDateRange(c)
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
	service service.ApprovalFlowService
	logger  *logger.Logger
}

func NewApprovalFlowHandler(service service.ApprovalFlowService, logger *logger.Logger) *ApprovalFlowHandler {
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
	id, err := parseUUID(c, "id")
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
	id, err := parseUUID(c, "id")
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

// ===== ExportHandler =====

type ExportHandler struct {
	service service.ExportService
	logger  *logger.Logger
}

func NewExportHandler(service service.ExportService, logger *logger.Logger) *ExportHandler {
	return &ExportHandler{service: service, logger: logger}
}

func (h *ExportHandler) ExportAttendance(c *gin.Context) {
	start, end, err := parseDateRange(c)
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
	start, end, err := parseDateRange(c)
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
	start, end, err := parseDateRange(c)
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
	start, end, err := parseDateRange(c)
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
