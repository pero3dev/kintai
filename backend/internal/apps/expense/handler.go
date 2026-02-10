package expense

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// ===== ExpenseHandler =====

type ExpenseHandler struct {
	svc    ExpenseService
	logger *logger.Logger
}

func NewExpenseHandler(svc ExpenseService, logger *logger.Logger) *ExpenseHandler {
	return &ExpenseHandler{svc: svc, logger: logger}
}

func (h *ExpenseHandler) GetList(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	category := c.Query("category")

	expenses, total, err := h.svc.GetList(c.Request.Context(), userID, page, pageSize, status, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     expenses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ExpenseHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	expense, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "経費申請が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です", Details: err.Error()})
		return
	}
	expense, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, expense)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	expense, err := h.svc.Update(c.Request.Context(), id, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "削除しました"})
}

func (h *ExpenseHandler) GetPending(c *gin.Context) {
	page, pageSize := parsePagination(c)
	expenses, total, err := h.svc.GetPending(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data:     expenses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ExpenseHandler) Approve(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	approverID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	if err := h.svc.Approve(c.Request.Context(), id, approverID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "処理しました"})
}

func (h *ExpenseHandler) AdvancedApprove(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	approverID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseAdvancedApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	if err := h.svc.AdvancedApprove(c.Request.Context(), id, approverID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "処理しました"})
}

func (h *ExpenseHandler) GetStats(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	stats, err := h.svc.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *ExpenseHandler) GetReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	report, err := h.svc.GetReport(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ExpenseHandler) GetMonthlyTrend(c *gin.Context) {
	year := c.Query("year")
	items, err := h.svc.GetMonthlyTrend(c.Request.Context(), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *ExpenseHandler) ExportCSV(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	data, err := h.svc.ExportCSV(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=expenses.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func (h *ExpenseHandler) ExportPDF(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	data, err := h.svc.ExportPDF(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=expenses.pdf")
	c.Data(http.StatusOK, "application/pdf", data)
}

// ===== ExpenseCommentHandler =====

type ExpenseCommentHandler struct {
	svc    ExpenseCommentService
	logger *logger.Logger
}

func NewExpenseCommentHandler(svc ExpenseCommentService, logger *logger.Logger) *ExpenseCommentHandler {
	return &ExpenseCommentHandler{svc: svc, logger: logger}
}

func (h *ExpenseCommentHandler) GetComments(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	comments, err := h.svc.GetComments(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": comments})
}

func (h *ExpenseCommentHandler) AddComment(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	comment, err := h.svc.AddComment(c.Request.Context(), id, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// ===== ExpenseHistoryHandler =====

type ExpenseHistoryHandler struct {
	svc    ExpenseHistoryService
	logger *logger.Logger
}

func NewExpenseHistoryHandler(svc ExpenseHistoryService, logger *logger.Logger) *ExpenseHistoryHandler {
	return &ExpenseHistoryHandler{svc: svc, logger: logger}
}

func (h *ExpenseHistoryHandler) GetHistory(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	history, err := h.svc.GetHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history})
}

// ===== ExpenseReceiptHandler =====

type ExpenseReceiptHandler struct {
	svc    ExpenseReceiptService
	logger *logger.Logger
}

func NewExpenseReceiptHandler(svc ExpenseReceiptService, logger *logger.Logger) *ExpenseReceiptHandler {
	return &ExpenseReceiptHandler{svc: svc, logger: logger}
}

func (h *ExpenseReceiptHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "ファイルが見つかりません"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "ファイル読み込みエラー"})
		return
	}

	url, err := h.svc.Upload(c.Request.Context(), header.Filename, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.ReceiptUploadResponse{URL: url})
}

// ===== ExpenseTemplateHandler =====

type ExpenseTemplateHandler struct {
	svc    ExpenseTemplateService
	logger *logger.Logger
}

func NewExpenseTemplateHandler(svc ExpenseTemplateService, logger *logger.Logger) *ExpenseTemplateHandler {
	return &ExpenseTemplateHandler{svc: svc, logger: logger}
}

func (h *ExpenseTemplateHandler) GetTemplates(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	templates, err := h.svc.GetTemplates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": templates})
}

func (h *ExpenseTemplateHandler) Create(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	tmpl, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tmpl)
}

func (h *ExpenseTemplateHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.ExpenseTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	tmpl, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

func (h *ExpenseTemplateHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "削除しました"})
}

func (h *ExpenseTemplateHandler) UseTemplate(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	expense, err := h.svc.UseTemplate(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": expense.ID})
}

// ===== ExpensePolicyHandler =====

type ExpensePolicyHandler struct {
	svc    ExpensePolicyService
	logger *logger.Logger
}

func NewExpensePolicyHandler(svc ExpensePolicyService, logger *logger.Logger) *ExpensePolicyHandler {
	return &ExpensePolicyHandler{svc: svc, logger: logger}
}

func (h *ExpensePolicyHandler) GetPolicies(c *gin.Context) {
	policies, err := h.svc.GetPolicies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policies})
}

func (h *ExpensePolicyHandler) Create(c *gin.Context) {
	var req model.ExpensePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	policy, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, policy)
}

func (h *ExpensePolicyHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.ExpensePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	policy, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, policy)
}

func (h *ExpensePolicyHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "削除しました"})
}

func (h *ExpensePolicyHandler) GetBudgets(c *gin.Context) {
	budgets, err := h.svc.GetBudgets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": budgets})
}

func (h *ExpensePolicyHandler) GetPolicyViolations(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	violations, err := h.svc.GetPolicyViolations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": violations})
}

// ===== ExpenseNotificationHandler =====

type ExpenseNotificationHandler struct {
	svc    ExpenseNotificationService
	logger *logger.Logger
}

func NewExpenseNotificationHandler(svc ExpenseNotificationService, logger *logger.Logger) *ExpenseNotificationHandler {
	return &ExpenseNotificationHandler{svc: svc, logger: logger}
}

func (h *ExpenseNotificationHandler) GetNotifications(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	filter := c.Query("filter")
	notifications, err := h.svc.GetNotifications(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": notifications})
}

func (h *ExpenseNotificationHandler) MarkAsRead(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "既読にしました"})
}

func (h *ExpenseNotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	if err := h.svc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "全て既読にしました"})
}

func (h *ExpenseNotificationHandler) GetReminders(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	reminders, err := h.svc.GetReminders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": reminders})
}

func (h *ExpenseNotificationHandler) DismissReminder(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.DismissReminder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "非表示にしました"})
}

func (h *ExpenseNotificationHandler) GetSettings(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	settings, err := h.svc.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *ExpenseNotificationHandler) UpdateSettings(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseNotificationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	settings, err := h.svc.UpdateSettings(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// ===== ExpenseApprovalFlowHandler =====

type ExpenseApprovalFlowHandler struct {
	svc    ExpenseApprovalFlowService
	logger *logger.Logger
}

func NewExpenseApprovalFlowHandler(svc ExpenseApprovalFlowService, logger *logger.Logger) *ExpenseApprovalFlowHandler {
	return &ExpenseApprovalFlowHandler{svc: svc, logger: logger}
}

func (h *ExpenseApprovalFlowHandler) GetConfig(c *gin.Context) {
	config, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *ExpenseApprovalFlowHandler) GetDelegates(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	delegates, err := h.svc.GetDelegates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": delegates})
}

func (h *ExpenseApprovalFlowHandler) SetDelegate(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	var req model.ExpenseDelegateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "リクエストが不正です"})
		return
	}
	delegate, err := h.svc.SetDelegate(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, delegate)
}

func (h *ExpenseApprovalFlowHandler) RemoveDelegate(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.RemoveDelegate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "削除しました"})
}
