package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// ===== HREmployeeHandler =====

type HREmployeeHandler struct {
	svc    service.HREmployeeService
	logger *logger.Logger
}

func NewHREmployeeHandler(svc service.HREmployeeService, logger *logger.Logger) *HREmployeeHandler {
	return &HREmployeeHandler{svc: svc, logger: logger}
}

func (h *HREmployeeHandler) Create(c *gin.Context) {
	var req model.HREmployeeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	e, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *HREmployeeHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	e, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "社員が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *HREmployeeHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	department := c.Query("department")
	status := c.Query("status")
	employmentType := c.Query("employment_type")
	search := c.Query("search")
	list, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, department, status, employmentType, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *HREmployeeHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.HREmployeeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	e, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *HREmployeeHandler) Delete(c *gin.Context) {
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

// ===== HRDepartmentHandler =====

type HRDepartmentHandler struct {
	svc    service.HRDepartmentService
	logger *logger.Logger
}

func NewHRDepartmentHandler(svc service.HRDepartmentService, logger *logger.Logger) *HRDepartmentHandler {
	return &HRDepartmentHandler{svc: svc, logger: logger}
}

func (h *HRDepartmentHandler) Create(c *gin.Context) {
	var req model.HRDepartmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	d, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *HRDepartmentHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	d, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "部門が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *HRDepartmentHandler) GetAll(c *gin.Context) {
	list, err := h.svc.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *HRDepartmentHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.HRDepartmentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	d, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *HRDepartmentHandler) Delete(c *gin.Context) {
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

// ===== EvaluationHandler =====

type EvaluationHandler struct {
	svc    service.EvaluationService
	logger *logger.Logger
}

func NewEvaluationHandler(svc service.EvaluationService, logger *logger.Logger) *EvaluationHandler {
	return &EvaluationHandler{svc: svc, logger: logger}
}

func (h *EvaluationHandler) Create(c *gin.Context) {
	var req model.EvaluationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	e, err := h.svc.Create(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *EvaluationHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	e, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "評価が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *EvaluationHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	cycleID := c.Query("cycle_id")
	status := c.Query("status")
	list, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, cycleID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *EvaluationHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.EvaluationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	e, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *EvaluationHandler) Submit(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	e, err := h.svc.Submit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *EvaluationHandler) GetCycles(c *gin.Context) {
	list, err := h.svc.FindAllCycles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *EvaluationHandler) CreateCycle(c *gin.Context) {
	var req model.EvaluationCycleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	cycle, err := h.svc.CreateCycle(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cycle)
}

// ===== GoalHandler =====

type GoalHandler struct {
	svc    service.GoalService
	logger *logger.Logger
}

func NewGoalHandler(svc service.GoalService, logger *logger.Logger) *GoalHandler {
	return &GoalHandler{svc: svc, logger: logger}
}

func (h *GoalHandler) Create(c *gin.Context) {
	var req model.HRGoalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	g, err := h.svc.Create(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (h *GoalHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	g, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "目標が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GoalHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	category := c.Query("category")
	employeeID := c.Query("employee_id")
	list, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, status, category, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *GoalHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.HRGoalUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	g, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GoalHandler) Delete(c *gin.Context) {
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

func (h *GoalHandler) UpdateProgress(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var body struct {
		Progress int `json:"progress"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	g, err := h.svc.UpdateProgress(c.Request.Context(), id, body.Progress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

// ===== TrainingHandler =====

type TrainingHandler struct {
	svc    service.TrainingService
	logger *logger.Logger
}

func NewTrainingHandler(svc service.TrainingService, logger *logger.Logger) *TrainingHandler {
	return &TrainingHandler{svc: svc, logger: logger}
}

func (h *TrainingHandler) Create(c *gin.Context) {
	var req model.TrainingProgramCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TrainingHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	t, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "研修が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TrainingHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	category := c.Query("category")
	status := c.Query("status")
	list, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, category, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *TrainingHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.TrainingProgramUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	t, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TrainingHandler) Delete(c *gin.Context) {
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

func (h *TrainingHandler) Enroll(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	if err := h.svc.Enroll(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "登録しました"})
}

func (h *TrainingHandler) Complete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	if err := h.svc.Complete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "完了しました"})
}

// ===== RecruitmentHandler =====

type RecruitmentHandler struct {
	svc    service.RecruitmentService
	logger *logger.Logger
}

func NewRecruitmentHandler(svc service.RecruitmentService, logger *logger.Logger) *RecruitmentHandler {
	return &RecruitmentHandler{svc: svc, logger: logger}
}

func (h *RecruitmentHandler) CreatePosition(c *gin.Context) {
	var req model.PositionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	p, err := h.svc.CreatePosition(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *RecruitmentHandler) GetPosition(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	p, err := h.svc.FindPositionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "ポジションが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *RecruitmentHandler) GetAllPositions(c *gin.Context) {
	page, pageSize := parsePagination(c)
	status := c.Query("status")
	department := c.Query("department")
	list, total, err := h.svc.FindAllPositions(c.Request.Context(), page, pageSize, status, department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *RecruitmentHandler) UpdatePosition(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.PositionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	p, err := h.svc.UpdatePosition(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *RecruitmentHandler) CreateApplicant(c *gin.Context) {
	var req model.ApplicantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	a, err := h.svc.CreateApplicant(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *RecruitmentHandler) GetAllApplicants(c *gin.Context) {
	positionID := c.Query("position_id")
	stage := c.Query("stage")
	list, err := h.svc.FindAllApplicants(c.Request.Context(), positionID, stage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *RecruitmentHandler) UpdateApplicantStage(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.ApplicantStageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	a, err := h.svc.UpdateApplicantStage(c.Request.Context(), id, req.Stage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

// ===== DocumentHandler =====

type DocumentHandler struct {
	svc    service.DocumentService
	logger *logger.Logger
}

func NewDocumentHandler(svc service.DocumentService, logger *logger.Logger) *DocumentHandler {
	return &DocumentHandler{svc: svc, logger: logger}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "ファイルが必要です"})
		return
	}
	defer file.Close()

	userID, _ := getUserIDFromContext(c)

	uploadDir := "uploads/documents"
	os.MkdirAll(uploadDir, 0755)
	fileName := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadDir, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: "ファイル保存に失敗しました"})
		return
	}
	defer out.Close()
	io.Copy(out, file)

	var employeeID *uuid.UUID
	if eid := c.PostForm("employee_id"); eid != "" {
		id, _ := uuid.Parse(eid)
		employeeID = &id
	}

	doc := &model.HRDocument{
		EmployeeID: employeeID,
		Title:      c.DefaultPostForm("title", header.Filename),
		Type:       c.DefaultPostForm("type", "other"),
		FileName:   header.Filename,
		FilePath:   filePath,
		FileSize:   header.Size,
		MimeType:   header.Header.Get("Content-Type"),
		UploadedBy: userID,
	}
	if err := h.svc.Upload(c.Request.Context(), doc); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	docType := c.Query("type")
	employeeID := c.Query("employee_id")
	list, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, docType, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
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

func (h *DocumentHandler) Download(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	doc, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "書類が見つかりません"})
		return
	}
	c.FileAttachment(doc.FilePath, doc.FileName)
}

// ===== AnnouncementHandler =====

type AnnouncementHandler struct {
	svc    service.AnnouncementService
	logger *logger.Logger
}

func NewAnnouncementHandler(svc service.AnnouncementService, logger *logger.Logger) *AnnouncementHandler {
	return &AnnouncementHandler{svc: svc, logger: logger}
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var req model.AnnouncementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	a, err := h.svc.Create(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *AnnouncementHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	a, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "お知らせが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AnnouncementHandler) GetAll(c *gin.Context) {
	page, pageSize := parsePagination(c)
	priority := c.Query("priority")
	list, total, err := h.svc.FindAll(c.Request.Context(), page, pageSize, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.AnnouncementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
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

// ===== HRDashboardHandler =====

type HRDashboardHandler struct {
	svc    service.HRDashboardService
	logger *logger.Logger
}

func NewHRDashboardHandler(svc service.HRDashboardService, logger *logger.Logger) *HRDashboardHandler {
	return &HRDashboardHandler{svc: svc, logger: logger}
}

func (h *HRDashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *HRDashboardHandler) GetActivities(c *gin.Context) {
	activities, err := h.svc.GetRecentActivities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, activities)
}

// ===== AttendanceIntegrationHandler =====

type AttendanceIntegrationHandler struct {
	svc    service.AttendanceIntegrationService
	logger *logger.Logger
}

func NewAttendanceIntegrationHandler(svc service.AttendanceIntegrationService, logger *logger.Logger) *AttendanceIntegrationHandler {
	return &AttendanceIntegrationHandler{svc: svc, logger: logger}
}

func (h *AttendanceIntegrationHandler) GetIntegration(c *gin.Context) {
	period := c.Query("period")
	department := c.Query("department")
	data, err := h.svc.GetIntegration(c.Request.Context(), period, department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AttendanceIntegrationHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.svc.GetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *AttendanceIntegrationHandler) GetTrend(c *gin.Context) {
	period := c.Query("period")
	trend, err := h.svc.GetTrend(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

// ===== OrgChartHandler =====

type OrgChartHandler struct {
	svc    service.OrgChartService
	logger *logger.Logger
}

func NewOrgChartHandler(svc service.OrgChartService, logger *logger.Logger) *OrgChartHandler {
	return &OrgChartHandler{svc: svc, logger: logger}
}

func (h *OrgChartHandler) GetOrgChart(c *gin.Context) {
	chart, err := h.svc.GetOrgChart(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, chart)
}

func (h *OrgChartHandler) Simulate(c *gin.Context) {
	var data map[string]interface{}
	c.ShouldBindJSON(&data)
	result, err := h.svc.Simulate(c.Request.Context(), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ===== OneOnOneHandler =====

type OneOnOneHandler struct {
	svc    service.OneOnOneService
	logger *logger.Logger
}

func NewOneOnOneHandler(svc service.OneOnOneService, logger *logger.Logger) *OneOnOneHandler {
	return &OneOnOneHandler{svc: svc, logger: logger}
}

func (h *OneOnOneHandler) Create(c *gin.Context) {
	var req model.OneOnOneCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	managerID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	m, err := h.svc.Create(c.Request.Context(), req, managerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *OneOnOneHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	m, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "1on1が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *OneOnOneHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	employeeID := c.Query("employee_id")
	list, err := h.svc.FindAll(c.Request.Context(), status, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *OneOnOneHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.OneOnOneUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	m, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *OneOnOneHandler) Delete(c *gin.Context) {
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

func (h *OneOnOneHandler) AddActionItem(c *gin.Context) {
	meetingID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.ActionItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	m, err := h.svc.AddActionItem(c.Request.Context(), meetingID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *OneOnOneHandler) ToggleActionItem(c *gin.Context) {
	meetingID, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	actionID := c.Param("actionId")
	m, err := h.svc.ToggleActionItem(c.Request.Context(), meetingID, actionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

// ===== SkillHandler =====

type SkillHandler struct {
	svc    service.SkillService
	logger *logger.Logger
}

func NewSkillHandler(svc service.SkillService, logger *logger.Logger) *SkillHandler {
	return &SkillHandler{svc: svc, logger: logger}
}

func (h *SkillHandler) GetSkillMap(c *gin.Context) {
	department := c.Query("department")
	employeeID := c.Query("employee_id")
	skills, err := h.svc.GetSkillMap(c.Request.Context(), department, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, skills)
}

func (h *SkillHandler) GetGapAnalysis(c *gin.Context) {
	department := c.Query("department")
	analysis, err := h.svc.GetGapAnalysis(c.Request.Context(), department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, analysis)
}

func (h *SkillHandler) AddSkill(c *gin.Context) {
	employeeID, err := parseUUID(c, "employeeId")
	if err != nil {
		return
	}
	var req model.SkillAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	skill, err := h.svc.AddSkill(c.Request.Context(), employeeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, skill)
}

func (h *SkillHandler) UpdateSkill(c *gin.Context) {
	skillID, err := parseUUID(c, "skillId")
	if err != nil {
		return
	}
	var req model.SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	skill, err := h.svc.UpdateSkill(c.Request.Context(), skillID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, skill)
}

// ===== SalaryHandler =====

type SalaryHandler struct {
	svc    service.SalaryService
	logger *logger.Logger
}

func NewSalaryHandler(svc service.SalaryService, logger *logger.Logger) *SalaryHandler {
	return &SalaryHandler{svc: svc, logger: logger}
}

func (h *SalaryHandler) GetOverview(c *gin.Context) {
	department := c.Query("department")
	overview, err := h.svc.GetOverview(c.Request.Context(), department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *SalaryHandler) Simulate(c *gin.Context) {
	var req model.SalarySimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	result, err := h.svc.Simulate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SalaryHandler) GetHistory(c *gin.Context) {
	empIDStr := c.Param("employeeId")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: "無効なIDフォーマットです"})
		return
	}
	records, err := h.svc.GetHistory(c.Request.Context(), empID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func (h *SalaryHandler) GetBudget(c *gin.Context) {
	department := c.Query("department")
	budget, err := h.svc.GetBudget(c.Request.Context(), department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, budget)
}

// ===== OnboardingHandler =====

type OnboardingHandler struct {
	svc    service.OnboardingService
	logger *logger.Logger
}

func NewOnboardingHandler(svc service.OnboardingService, logger *logger.Logger) *OnboardingHandler {
	return &OnboardingHandler{svc: svc, logger: logger}
}

func (h *OnboardingHandler) Create(c *gin.Context) {
	var req model.OnboardingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	o, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *OnboardingHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	o, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "オンボーディングが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OnboardingHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	list, err := h.svc.FindAll(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *OnboardingHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	o, err := h.svc.Update(c.Request.Context(), id, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OnboardingHandler) ToggleTask(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	taskID := c.Param("taskId")
	o, err := h.svc.ToggleTask(c.Request.Context(), id, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OnboardingHandler) GetTemplates(c *gin.Context) {
	list, err := h.svc.FindAllTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *OnboardingHandler) CreateTemplate(c *gin.Context) {
	var req model.OnboardingTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	t, err := h.svc.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// ===== OffboardingHandler =====

type OffboardingHandler struct {
	svc    service.OffboardingService
	logger *logger.Logger
}

func NewOffboardingHandler(svc service.OffboardingService, logger *logger.Logger) *OffboardingHandler {
	return &OffboardingHandler{svc: svc, logger: logger}
}

func (h *OffboardingHandler) Create(c *gin.Context) {
	var req model.OffboardingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	o, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *OffboardingHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	o, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "退職管理データが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OffboardingHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	list, err := h.svc.FindAll(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *OffboardingHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.OffboardingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	o, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OffboardingHandler) ToggleChecklist(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	itemKey := c.Param("itemKey")
	o, err := h.svc.ToggleChecklist(c.Request.Context(), id, itemKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OffboardingHandler) GetAnalytics(c *gin.Context) {
	analytics, err := h.svc.GetAnalytics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, analytics)
}

// ===== SurveyHandler =====

type SurveyHandler struct {
	svc    service.SurveyService
	logger *logger.Logger
}

func NewSurveyHandler(svc service.SurveyService, logger *logger.Logger) *SurveyHandler {
	return &SurveyHandler{svc: svc, logger: logger}
}

func (h *SurveyHandler) Create(c *gin.Context) {
	var req model.SurveyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Code: 401, Message: "認証エラー"})
		return
	}
	s, err := h.svc.Create(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *SurveyHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	s, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Code: 404, Message: "サーベイが見つかりません"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SurveyHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	surveyType := c.Query("type")
	list, err := h.svc.FindAll(c.Request.Context(), status, surveyType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *SurveyHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.SurveyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	s, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SurveyHandler) Delete(c *gin.Context) {
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

func (h *SurveyHandler) Publish(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	s, err := h.svc.Publish(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SurveyHandler) Close(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	s, err := h.svc.Close(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *SurveyHandler) GetResults(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	results, err := h.svc.GetResults(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *SurveyHandler) SubmitResponse(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}
	var req model.SurveyResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	userID, _ := getUserIDFromContext(c)
	var empID *uuid.UUID
	if userID != uuid.Nil {
		empID = &userID
	}
	if err := h.svc.SubmitResponse(c.Request.Context(), id, empID, req); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "回答を送信しました"})
}

// suppress unused import
var _ = strconv.Itoa
