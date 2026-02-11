package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// Ensure imports are used
var (
	_ = bytes.NewBufferString
	_ = context.Background
	_ = json.Unmarshal
	_ = errors.New
	_ = http.StatusOK
	_ = httptest.NewRecorder
	_ = (*testing.T)(nil)
	_ = time.Now
	_ = (*gin.Context)(nil)
	_ = uuid.New
	_ = (*mocks.MockOvertimeRequestService)(nil)
	_ = model.ErrorResponse{}
	_ = (*logger.Logger)(nil)
)

// ===================================================================
// OvertimeRequestHandler Tests
// ===================================================================

func TestOvertimeRequestHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error) {
			return &model.OvertimeRequest{
				BaseModel:      model.BaseModel{ID: uuid.New()},
				UserID:         uid,
				PlannedMinutes: req.PlannedMinutes,
				Reason:         req.Reason,
				Status:         model.OvertimeStatusPending,
			}, nil
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/overtime", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"date":"2026-02-10","planned_minutes":60,"reason":"締め切り対応"}`
	req, _ := http.NewRequest(http.MethodPost, "/overtime", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestOvertimeRequestHandler_Create_Unauthorized(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/overtime", handler.Create)

	body := `{"date":"2026-02-10","planned_minutes":60,"reason":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/overtime", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestOvertimeRequestHandler_Create_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/overtime", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/overtime", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestOvertimeRequestHandler_Create_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/overtime", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"date":"2026-02-10","planned_minutes":60,"reason":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/overtime", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestOvertimeRequestHandler_Approve_Success(t *testing.T) {
	approverID := uuid.New()
	overtimeID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{
		ApproveFunc: func(ctx context.Context, id uuid.UUID, aid uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error) {
			return &model.OvertimeRequest{
				BaseModel: model.BaseModel{ID: id},
				Status:    model.OvertimeStatusApproved,
			}, nil
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/overtime/:id/approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/overtime/"+overtimeID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOvertimeRequestHandler_Approve_InvalidID(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/overtime/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/overtime/invalid-uuid/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestOvertimeRequestHandler_Approve_Unauthorized(t *testing.T) {
	overtimeID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/overtime/:id/approve", handler.Approve)

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/overtime/"+overtimeID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestOvertimeRequestHandler_Approve_BadRequest(t *testing.T) {
	overtimeID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/overtime/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/overtime/"+overtimeID.String()+"/approve", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestOvertimeRequestHandler_Approve_ServiceError(t *testing.T) {
	overtimeID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{
		ApproveFunc: func(ctx context.Context, id uuid.UUID, aid uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/overtime/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/overtime/"+overtimeID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestOvertimeRequestHandler_GetMy_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
			return []model.OvertimeRequest{{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: uid}}, 1, nil
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/my", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/overtime/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOvertimeRequestHandler_GetMy_Unauthorized(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/my", handler.GetMy)

	req, _ := http.NewRequest(http.MethodGet, "/overtime/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestOvertimeRequestHandler_GetMy_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockOvertimeRequestService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/my", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/overtime/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestOvertimeRequestHandler_GetPending_Success(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
			return []model.OvertimeRequest{{BaseModel: model.BaseModel{ID: uuid.New()}}}, 1, nil
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/overtime/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOvertimeRequestHandler_GetPending_ServiceError(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/overtime/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestOvertimeRequestHandler_GetAlerts_Success(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{
		GetOvertimeAlertsFunc: func(ctx context.Context) ([]model.OvertimeAlert, error) {
			return []model.OvertimeAlert{{UserID: uuid.New(), MonthlyOvertimeHours: 50}}, nil
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/alerts", handler.GetAlerts)

	req, _ := http.NewRequest(http.MethodGet, "/overtime/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOvertimeRequestHandler_GetAlerts_ServiceError(t *testing.T) {
	mockService := &mocks.MockOvertimeRequestService{
		GetOvertimeAlertsFunc: func(ctx context.Context) ([]model.OvertimeAlert, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewOvertimeRequestHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/overtime/alerts", handler.GetAlerts)

	req, _ := http.NewRequest(http.MethodGet, "/overtime/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// LeaveBalanceHandler Tests
// ===================================================================

func TestLeaveBalanceHandler_GetByUser_Success(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error) {
			return []model.LeaveBalanceResponse{
				{LeaveType: model.LeaveTypePaid, TotalDays: 20, UsedDays: 5, RemainingDays: 15, FiscalYear: fiscalYear},
			}, nil
		},
	}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leave-balances/:user_id", handler.GetByUser)

	req, _ := http.NewRequest(http.MethodGet, "/leave-balances/"+targetUserID.String()+"?fiscal_year=2026", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLeaveBalanceHandler_GetByUser_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockLeaveBalanceService{}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leave-balances/:user_id", handler.GetByUser)

	req, _ := http.NewRequest(http.MethodGet, "/leave-balances/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveBalanceHandler_GetByUser_ServiceError(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leave-balances/:user_id", handler.GetByUser)

	req, _ := http.NewRequest(http.MethodGet, "/leave-balances/"+targetUserID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLeaveBalanceHandler_SetBalance_Success(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{
		SetBalanceFunc: func(ctx context.Context, uid uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error {
			return nil
		},
	}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leave-balances/:user_id/:leave_type", handler.SetBalance)

	totalDays := 20.0
	body, _ := json.Marshal(model.LeaveBalanceUpdate{TotalDays: &totalDays})
	req, _ := http.NewRequest(http.MethodPut, "/leave-balances/"+targetUserID.String()+"/paid?fiscal_year=2026", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLeaveBalanceHandler_SetBalance_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockLeaveBalanceService{}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leave-balances/:user_id/:leave_type", handler.SetBalance)

	body := `{"total_days":20}`
	req, _ := http.NewRequest(http.MethodPut, "/leave-balances/invalid-uuid/paid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveBalanceHandler_SetBalance_BadRequest(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leave-balances/:user_id/:leave_type", handler.SetBalance)

	req, _ := http.NewRequest(http.MethodPut, "/leave-balances/"+targetUserID.String()+"/paid", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveBalanceHandler_SetBalance_ServiceError(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{
		SetBalanceFunc: func(ctx context.Context, uid uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error {
			return errors.New("service error")
		},
	}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leave-balances/:user_id/:leave_type", handler.SetBalance)

	totalDays := 20.0
	body, _ := json.Marshal(model.LeaveBalanceUpdate{TotalDays: &totalDays})
	req, _ := http.NewRequest(http.MethodPut, "/leave-balances/"+targetUserID.String()+"/paid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveBalanceHandler_Initialize_Success(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{
		InitializeForUserFunc: func(ctx context.Context, uid uuid.UUID, fiscalYear int) error {
			return nil
		},
	}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leave-balances/:user_id/initialize", handler.Initialize)

	req, _ := http.NewRequest(http.MethodPost, "/leave-balances/"+targetUserID.String()+"/initialize?fiscal_year=2026", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestLeaveBalanceHandler_Initialize_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockLeaveBalanceService{}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leave-balances/:user_id/initialize", handler.Initialize)

	req, _ := http.NewRequest(http.MethodPost, "/leave-balances/invalid-uuid/initialize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveBalanceHandler_Initialize_ServiceError(t *testing.T) {
	targetUserID := uuid.New()
	mockService := &mocks.MockLeaveBalanceService{
		InitializeForUserFunc: func(ctx context.Context, uid uuid.UUID, fiscalYear int) error {
			return errors.New("service error")
		},
	}
	handler := NewLeaveBalanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leave-balances/:user_id/initialize", handler.Initialize)

	req, _ := http.NewRequest(http.MethodPost, "/leave-balances/"+targetUserID.String()+"/initialize", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ===================================================================
// AttendanceCorrectionHandler Tests
// ===================================================================

func TestAttendanceCorrectionHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error) {
			return &model.AttendanceCorrection{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    uid,
				Reason:    req.Reason,
				Status:    model.CorrectionStatusPending,
			}, nil
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/corrections", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"date":"2026-02-10","corrected_clock_in":"09:00","reason":"打刻忘れ"}`
	req, _ := http.NewRequest(http.MethodPost, "/corrections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Create_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceCorrectionService{}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/corrections", handler.Create)

	body := `{"date":"2026-02-10","reason":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/corrections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Create_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/corrections", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/corrections", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Create_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/corrections", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"date":"2026-02-10","reason":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/corrections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Approve_Success(t *testing.T) {
	approverID := uuid.New()
	correctionID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{
		ApproveFunc: func(ctx context.Context, id uuid.UUID, aid uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error) {
			return &model.AttendanceCorrection{
				BaseModel: model.BaseModel{ID: id},
				Status:    model.CorrectionStatusApproved,
			}, nil
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/corrections/:id/approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/corrections/"+correctionID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Approve_Unauthorized(t *testing.T) {
	correctionID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/corrections/:id/approve", handler.Approve)

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/corrections/"+correctionID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Approve_BadRequest(t *testing.T) {
	correctionID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/corrections/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/corrections/"+correctionID.String()+"/approve", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceCorrectionHandler_Approve_ServiceError(t *testing.T) {
	correctionID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{
		ApproveFunc: func(ctx context.Context, id uuid.UUID, aid uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/corrections/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/corrections/"+correctionID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceCorrectionHandler_GetMy_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
			return []model.AttendanceCorrection{{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: uid}}, 1, nil
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/corrections/my", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/corrections/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceCorrectionHandler_GetMy_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceCorrectionService{}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/corrections/my", handler.GetMy)

	req, _ := http.NewRequest(http.MethodGet, "/corrections/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceCorrectionHandler_GetMy_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockAttendanceCorrectionService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/corrections/my", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/corrections/my", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAttendanceCorrectionHandler_GetPending_Success(t *testing.T) {
	mockService := &mocks.MockAttendanceCorrectionService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
			return []model.AttendanceCorrection{{BaseModel: model.BaseModel{ID: uuid.New()}}}, 1, nil
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/corrections/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/corrections/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceCorrectionHandler_GetPending_ServiceError(t *testing.T) {
	mockService := &mocks.MockAttendanceCorrectionService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}
	handler := NewAttendanceCorrectionHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/corrections/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/corrections/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// NotificationHandler Tests
// ===================================================================

func TestNotificationHandler_GetMy_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockNotificationService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
			return []model.Notification{{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: uid, Title: "Test"}}, 1, nil
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/notifications", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNotificationHandler_GetMy_Unauthorized(t *testing.T) {
	mockService := &mocks.MockNotificationService{}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/notifications", handler.GetMy)

	req, _ := http.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestNotificationHandler_GetMy_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockNotificationService{
		GetByUserFunc: func(ctx context.Context, uid uuid.UUID, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/notifications", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNotificationHandler_MarkAsRead_Success(t *testing.T) {
	notifID := uuid.New()
	mockService := &mocks.MockNotificationService{
		MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/notifications/"+notifID.String()+"/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestNotificationHandler_MarkAsRead_InvalidID(t *testing.T) {
	mockService := &mocks.MockNotificationService{}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/notifications/invalid-uuid/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNotificationHandler_MarkAsRead_ServiceError(t *testing.T) {
	notifID := uuid.New()
	mockService := &mocks.MockNotificationService{
		MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/notifications/"+notifID.String()+"/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNotificationHandler_MarkAllAsRead_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockNotificationService{
		MarkAllAsReadFunc: func(ctx context.Context, uid uuid.UUID) error {
			return nil
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/notifications/read-all", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.MarkAllAsRead(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestNotificationHandler_MarkAllAsRead_Unauthorized(t *testing.T) {
	mockService := &mocks.MockNotificationService{}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/notifications/read-all", handler.MarkAllAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestNotificationHandler_MarkAllAsRead_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockNotificationService{
		MarkAllAsReadFunc: func(ctx context.Context, uid uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/notifications/read-all", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.MarkAllAsRead(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNotificationHandler_GetUnreadCount_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockNotificationService{
		GetUnreadCountFunc: func(ctx context.Context, uid uuid.UUID) (int64, error) {
			return 5, nil
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/notifications/unread-count", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetUnreadCount(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	var resp model.NotificationCount
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if resp.Unread != 5 {
		t.Errorf("Expected unread count 5, got %d", resp.Unread)
	}
}

func TestNotificationHandler_GetUnreadCount_Unauthorized(t *testing.T) {
	mockService := &mocks.MockNotificationService{}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/notifications/unread-count", handler.GetUnreadCount)

	req, _ := http.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestNotificationHandler_GetUnreadCount_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockNotificationService{
		GetUnreadCountFunc: func(ctx context.Context, uid uuid.UUID) (int64, error) {
			return 0, errors.New("service error")
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/notifications/unread-count", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetUnreadCount(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNotificationHandler_Delete_Success(t *testing.T) {
	notifID := uuid.New()
	mockService := &mocks.MockNotificationService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/notifications/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/notifications/"+notifID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestNotificationHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockNotificationService{}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/notifications/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/notifications/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNotificationHandler_Delete_ServiceError(t *testing.T) {
	notifID := uuid.New()
	mockService := &mocks.MockNotificationService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/notifications/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/notifications/"+notifID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// ProjectHandler Tests
// ===================================================================

func TestProjectHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockProjectService{
		CreateFunc: func(ctx context.Context, req *model.ProjectCreateRequest) (*model.Project, error) {
			return &model.Project{
				BaseModel: model.BaseModel{ID: uuid.New()},
				Name:      req.Name,
				Code:      req.Code,
				Status:    model.ProjectStatusActive,
			}, nil
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/projects", handler.Create)

	body := `{"name":"テストプロジェクト","code":"PROJ-001","description":"テスト用"}`
	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestProjectHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockProjectService{}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/projects", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockProjectService{
		CreateFunc: func(ctx context.Context, req *model.ProjectCreateRequest) (*model.Project, error) {
			return nil, errors.New("duplicate code")
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/projects", handler.Create)

	body := `{"name":"テスト","code":"PROJ-001"}`
	req, _ := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_GetByID_Success(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Project, error) {
			return &model.Project{
				BaseModel: model.BaseModel{ID: id},
				Name:      "テストプロジェクト",
				Code:      "PROJ-001",
			}, nil
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProjectHandler_GetByID_InvalidID(t *testing.T) {
	mockService := &mocks.MockProjectService{}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/projects/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_GetByID_ServiceError(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Project, error) {
			return nil, errors.New("not found")
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestProjectHandler_GetAll_Success(t *testing.T) {
	mockService := &mocks.MockProjectService{
		GetAllFunc: func(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error) {
			return []model.Project{{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "プロジェクト1"}}, 1, nil
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProjectHandler_GetAll_ServiceError(t *testing.T) {
	mockService := &mocks.MockProjectService{
		GetAllFunc: func(ctx context.Context, status *model.ProjectStatus, page, pageSize int) ([]model.Project, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestProjectHandler_Update_Success(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ProjectUpdateRequest) (*model.Project, error) {
			name := "更新プロジェクト"
			if req.Name != nil {
				name = *req.Name
			}
			return &model.Project{BaseModel: model.BaseModel{ID: id}, Name: name}, nil
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/projects/:id", handler.Update)

	body := `{"name":"更新プロジェクト"}`
	req, _ := http.NewRequest(http.MethodPut, "/projects/"+projectID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProjectHandler_Update_InvalidID(t *testing.T) {
	mockService := &mocks.MockProjectService{}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/projects/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/projects/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_Update_BadRequest(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/projects/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/projects/"+projectID.String(), bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_Update_ServiceError(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ProjectUpdateRequest) (*model.Project, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/projects/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/projects/"+projectID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_Delete_Success(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/projects/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestProjectHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockProjectService{}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/projects/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/projects/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProjectHandler_Delete_ServiceError(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockProjectService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewProjectHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/projects/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/projects/"+projectID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// TimeEntryHandler Tests
// ===================================================================

func TestTimeEntryHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.TimeEntryCreate) (*model.TimeEntry, error) {
			return &model.TimeEntry{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    uid,
				ProjectID: req.ProjectID,
				Minutes:   req.Minutes,
			}, nil
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/time-entries", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"project_id":"` + projectID.String() + `","date":"2026-02-10","minutes":120,"description":"開発作業"}`
	req, _ := http.NewRequest(http.MethodPost, "/time-entries", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestTimeEntryHandler_Create_Unauthorized(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/time-entries", handler.Create)

	body := `{"project_id":"` + uuid.New().String() + `","date":"2026-02-10","minutes":60}`
	req, _ := http.NewRequest(http.MethodPost, "/time-entries", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTimeEntryHandler_Create_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/time-entries", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/time-entries", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_Create_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.TimeEntryCreate) (*model.TimeEntry, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/time-entries", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"project_id":"` + uuid.New().String() + `","date":"2026-02-10","minutes":60}`
	req, _ := http.NewRequest(http.MethodPost, "/time-entries", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_GetMy_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		GetByUserAndDateRangeFunc: func(ctx context.Context, uid uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
			return []model.TimeEntry{{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: uid}}, nil
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/time-entries", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/time-entries?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTimeEntryHandler_GetMy_Unauthorized(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/time-entries", handler.GetMy)

	req, _ := http.NewRequest(http.MethodGet, "/time-entries?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTimeEntryHandler_GetMy_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		GetByUserAndDateRangeFunc: func(ctx context.Context, uid uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/time-entries", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/time-entries?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestTimeEntryHandler_GetByProject_Success(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		GetByProjectAndDateRangeFunc: func(ctx context.Context, pid uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
			return []model.TimeEntry{{BaseModel: model.BaseModel{ID: uuid.New()}, ProjectID: pid}}, nil
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects/:id/time-entries", handler.GetByProject)

	req, _ := http.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/time-entries?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTimeEntryHandler_GetByProject_InvalidID(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects/:id/time-entries", handler.GetByProject)

	req, _ := http.NewRequest(http.MethodGet, "/projects/invalid-uuid/time-entries?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_GetByProject_ServiceError(t *testing.T) {
	projectID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		GetByProjectAndDateRangeFunc: func(ctx context.Context, pid uuid.UUID, start, end time.Time) ([]model.TimeEntry, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/projects/:id/time-entries", handler.GetByProject)

	req, _ := http.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/time-entries?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestTimeEntryHandler_Update_Success(t *testing.T) {
	entryID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.TimeEntryUpdate) (*model.TimeEntry, error) {
			return &model.TimeEntry{BaseModel: model.BaseModel{ID: id}}, nil
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/time-entries/:id", handler.Update)

	body := `{"minutes":180}`
	req, _ := http.NewRequest(http.MethodPut, "/time-entries/"+entryID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTimeEntryHandler_Update_InvalidID(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/time-entries/:id", handler.Update)

	body := `{"minutes":180}`
	req, _ := http.NewRequest(http.MethodPut, "/time-entries/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_Update_BadRequest(t *testing.T) {
	entryID := uuid.New()
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/time-entries/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/time-entries/"+entryID.String(), bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_Update_ServiceError(t *testing.T) {
	entryID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.TimeEntryUpdate) (*model.TimeEntry, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/time-entries/:id", handler.Update)

	body := `{"minutes":180}`
	req, _ := http.NewRequest(http.MethodPut, "/time-entries/"+entryID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_Delete_Success(t *testing.T) {
	entryID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/time-entries/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/time-entries/"+entryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestTimeEntryHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/time-entries/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/time-entries/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_Delete_ServiceError(t *testing.T) {
	entryID := uuid.New()
	mockService := &mocks.MockTimeEntryService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/time-entries/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/time-entries/"+entryID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestTimeEntryHandler_GetSummary_Success(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{
		GetProjectSummaryFunc: func(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
			return []model.ProjectSummary{
				{ProjectID: uuid.New(), ProjectName: "プロジェクト1", TotalMinutes: 480, TotalHours: 8.0},
			}, nil
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/time-entries/summary", handler.GetSummary)

	req, _ := http.NewRequest(http.MethodGet, "/time-entries/summary?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTimeEntryHandler_GetSummary_BadDateRange(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/time-entries/summary", handler.GetSummary)

	req, _ := http.NewRequest(http.MethodGet, "/time-entries/summary?start_date=invalid&end_date=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTimeEntryHandler_GetSummary_ServiceError(t *testing.T) {
	mockService := &mocks.MockTimeEntryService{
		GetProjectSummaryFunc: func(ctx context.Context, start, end time.Time) ([]model.ProjectSummary, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewTimeEntryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/time-entries/summary", handler.GetSummary)

	req, _ := http.NewRequest(http.MethodGet, "/time-entries/summary?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// HolidayHandler Tests
// ===================================================================

func TestHolidayHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		CreateFunc: func(ctx context.Context, req *model.HolidayCreateRequest) (*model.Holiday, error) {
			return &model.Holiday{
				BaseModel:   model.BaseModel{ID: uuid.New()},
				Name:        req.Name,
				HolidayType: model.HolidayType(req.HolidayType),
			}, nil
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/holidays", handler.Create)

	body := `{"date":"2026-01-01","name":"元日","holiday_type":"national","is_recurring":true}`
	req, _ := http.NewRequest(http.MethodPost, "/holidays", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHolidayHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockHolidayService{}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/holidays", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/holidays", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		CreateFunc: func(ctx context.Context, req *model.HolidayCreateRequest) (*model.Holiday, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/holidays", handler.Create)

	body := `{"date":"2026-01-01","name":"元日","holiday_type":"national"}`
	req, _ := http.NewRequest(http.MethodPost, "/holidays", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_GetByYear_Success(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		GetByYearFunc: func(ctx context.Context, year int) ([]model.Holiday, error) {
			return []model.Holiday{{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "元日"}}, nil
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays", handler.GetByYear)

	req, _ := http.NewRequest(http.MethodGet, "/holidays?year=2026", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHolidayHandler_GetByYear_ServiceError(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		GetByYearFunc: func(ctx context.Context, year int) ([]model.Holiday, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays", handler.GetByYear)

	req, _ := http.NewRequest(http.MethodGet, "/holidays?year=2026", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHolidayHandler_Update_Success(t *testing.T) {
	holidayID := uuid.New()
	mockService := &mocks.MockHolidayService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.HolidayUpdateRequest) (*model.Holiday, error) {
			return &model.Holiday{BaseModel: model.BaseModel{ID: id}}, nil
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/holidays/:id", handler.Update)

	body := `{"name":"成人の日"}`
	req, _ := http.NewRequest(http.MethodPut, "/holidays/"+holidayID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHolidayHandler_Update_InvalidID(t *testing.T) {
	mockService := &mocks.MockHolidayService{}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/holidays/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/holidays/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_Update_BadRequest(t *testing.T) {
	holidayID := uuid.New()
	mockService := &mocks.MockHolidayService{}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/holidays/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/holidays/"+holidayID.String(), bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_Update_ServiceError(t *testing.T) {
	holidayID := uuid.New()
	mockService := &mocks.MockHolidayService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.HolidayUpdateRequest) (*model.Holiday, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/holidays/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/holidays/"+holidayID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_Delete_Success(t *testing.T) {
	holidayID := uuid.New()
	mockService := &mocks.MockHolidayService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/holidays/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/holidays/"+holidayID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestHolidayHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockHolidayService{}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/holidays/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/holidays/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_Delete_ServiceError(t *testing.T) {
	holidayID := uuid.New()
	mockService := &mocks.MockHolidayService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/holidays/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/holidays/"+holidayID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHolidayHandler_GetCalendar_Success(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		GetCalendarFunc: func(ctx context.Context, year, month int) ([]model.CalendarDay, error) {
			return []model.CalendarDay{
				{Date: "2026-02-10", IsHoliday: false, IsWeekend: false},
				{Date: "2026-02-11", IsHoliday: true, HolidayName: "建国記念の日"},
			}, nil
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays/calendar", handler.GetCalendar)

	req, _ := http.NewRequest(http.MethodGet, "/holidays/calendar?year=2026&month=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHolidayHandler_GetCalendar_ServiceError(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		GetCalendarFunc: func(ctx context.Context, year, month int) ([]model.CalendarDay, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays/calendar", handler.GetCalendar)

	req, _ := http.NewRequest(http.MethodGet, "/holidays/calendar?year=2026&month=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHolidayHandler_GetWorkingDays_Success(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		GetWorkingDaysFunc: func(ctx context.Context, start, end time.Time) (*model.WorkingDaysSummary, error) {
			return &model.WorkingDaysSummary{TotalDays: 28, WorkingDays: 20, Holidays: 1, Weekends: 7}, nil
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays/working-days", handler.GetWorkingDays)

	req, _ := http.NewRequest(http.MethodGet, "/holidays/working-days?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHolidayHandler_GetWorkingDays_BadDateRange(t *testing.T) {
	mockService := &mocks.MockHolidayService{}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays/working-days", handler.GetWorkingDays)

	req, _ := http.NewRequest(http.MethodGet, "/holidays/working-days?start_date=invalid&end_date=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHolidayHandler_GetWorkingDays_ServiceError(t *testing.T) {
	mockService := &mocks.MockHolidayService{
		GetWorkingDaysFunc: func(ctx context.Context, start, end time.Time) (*model.WorkingDaysSummary, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewHolidayHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/holidays/working-days", handler.GetWorkingDays)

	req, _ := http.NewRequest(http.MethodGet, "/holidays/working-days?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// ApprovalFlowHandler Tests
// ===================================================================

func TestApprovalFlowHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{
		CreateFunc: func(ctx context.Context, req *model.ApprovalFlowCreateRequest) (*model.ApprovalFlow, error) {
			return &model.ApprovalFlow{
				BaseModel: model.BaseModel{ID: uuid.New()},
				Name:      req.Name,
				FlowType:  model.ApprovalFlowType(req.FlowType),
				IsActive:  true,
			}, nil
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/approval-flows", handler.Create)

	body := `{"name":"休暇承認フロー","flow_type":"leave","steps":[{"step_order":1,"step_type":"manager"}]}`
	req, _ := http.NewRequest(http.MethodPost, "/approval-flows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestApprovalFlowHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/approval-flows", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/approval-flows", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{
		CreateFunc: func(ctx context.Context, req *model.ApprovalFlowCreateRequest) (*model.ApprovalFlow, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/approval-flows", handler.Create)

	body := `{"name":"テスト","flow_type":"leave","steps":[{"step_order":1,"step_type":"manager"}]}`
	req, _ := http.NewRequest(http.MethodPost, "/approval-flows", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_GetAll_Success(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{
		GetAllFunc: func(ctx context.Context) ([]model.ApprovalFlow, error) {
			return []model.ApprovalFlow{
				{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "休暇承認フロー", FlowType: model.ApprovalFlowLeave},
			}, nil
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/approval-flows", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/approval-flows", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestApprovalFlowHandler_GetAll_ServiceError(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{
		GetAllFunc: func(ctx context.Context) ([]model.ApprovalFlow, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/approval-flows", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/approval-flows", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestApprovalFlowHandler_GetByID_Success(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error) {
			return &model.ApprovalFlow{
				BaseModel: model.BaseModel{ID: id},
				Name:      "休暇承認フロー",
			}, nil
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/approval-flows/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/approval-flows/"+flowID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestApprovalFlowHandler_GetByID_InvalidID(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/approval-flows/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/approval-flows/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_GetByID_ServiceError(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.ApprovalFlow, error) {
			return nil, errors.New("not found")
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/approval-flows/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/approval-flows/"+flowID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestApprovalFlowHandler_Update_Success(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ApprovalFlowUpdateRequest) (*model.ApprovalFlow, error) {
			return &model.ApprovalFlow{BaseModel: model.BaseModel{ID: id}}, nil
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/approval-flows/:id", handler.Update)

	body := `{"name":"更新フロー"}`
	req, _ := http.NewRequest(http.MethodPut, "/approval-flows/"+flowID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestApprovalFlowHandler_Update_InvalidID(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/approval-flows/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/approval-flows/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_Update_BadRequest(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/approval-flows/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/approval-flows/"+flowID.String(), bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_Update_ServiceError(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ApprovalFlowUpdateRequest) (*model.ApprovalFlow, error) {
			return nil, errors.New("service error")
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/approval-flows/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/approval-flows/"+flowID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_Delete_Success(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/approval-flows/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/approval-flows/"+flowID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestApprovalFlowHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockApprovalFlowService{}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/approval-flows/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/approval-flows/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestApprovalFlowHandler_Delete_ServiceError(t *testing.T) {
	flowID := uuid.New()
	mockService := &mocks.MockApprovalFlowService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("service error")
		},
	}
	handler := NewApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/approval-flows/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/approval-flows/"+flowID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===================================================================
// ExportHandler Tests
// ===================================================================

func TestExportHandler_ExportAttendance_Success(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportAttendanceCSVFunc: func(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
			return []byte("user_id,date,clock_in,clock_out\n"), nil
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/attendance", handler.ExportAttendance)

	req, _ := http.NewRequest(http.MethodGet, "/export/attendance?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/csv; charset=utf-8', got '%s'", contentType)
	}
	disposition := w.Header().Get("Content-Disposition")
	if disposition != "attachment; filename=attendance.csv" {
		t.Errorf("Expected Content-Disposition 'attachment; filename=attendance.csv', got '%s'", disposition)
	}
}

func TestExportHandler_ExportAttendance_BadDateRange(t *testing.T) {
	mockService := &mocks.MockExportService{}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/attendance", handler.ExportAttendance)

	req, _ := http.NewRequest(http.MethodGet, "/export/attendance?start_date=invalid&end_date=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExportHandler_ExportAttendance_ServiceError(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportAttendanceCSVFunc: func(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
			return nil, errors.New("export error")
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/attendance", handler.ExportAttendance)

	req, _ := http.NewRequest(http.MethodGet, "/export/attendance?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExportHandler_ExportLeaves_Success(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportLeavesCSVFunc: func(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
			return []byte("user_id,leave_type,start,end\n"), nil
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/leaves", handler.ExportLeaves)

	req, _ := http.NewRequest(http.MethodGet, "/export/leaves?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExportHandler_ExportLeaves_ServiceError(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportLeavesCSVFunc: func(ctx context.Context, userID *uuid.UUID, start, end time.Time) ([]byte, error) {
			return nil, errors.New("export error")
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/leaves", handler.ExportLeaves)

	req, _ := http.NewRequest(http.MethodGet, "/export/leaves?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExportHandler_ExportOvertime_Success(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportOvertimeCSVFunc: func(ctx context.Context, start, end time.Time) ([]byte, error) {
			return []byte("user_id,date,minutes\n"), nil
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/overtime", handler.ExportOvertime)

	req, _ := http.NewRequest(http.MethodGet, "/export/overtime?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExportHandler_ExportOvertime_ServiceError(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportOvertimeCSVFunc: func(ctx context.Context, start, end time.Time) ([]byte, error) {
			return nil, errors.New("export error")
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/overtime", handler.ExportOvertime)

	req, _ := http.NewRequest(http.MethodGet, "/export/overtime?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExportHandler_ExportProjects_Success(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportProjectsCSVFunc: func(ctx context.Context, start, end time.Time) ([]byte, error) {
			return []byte("project_id,name,total_hours\n"), nil
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/projects", handler.ExportProjects)

	req, _ := http.NewRequest(http.MethodGet, "/export/projects?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExportHandler_ExportProjects_ServiceError(t *testing.T) {
	mockService := &mocks.MockExportService{
		ExportProjectsCSVFunc: func(ctx context.Context, start, end time.Time) ([]byte, error) {
			return nil, errors.New("export error")
		},
	}
	handler := NewExportHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/export/projects", handler.ExportProjects)

	req, _ := http.NewRequest(http.MethodGet, "/export/projects?start_date=2026-02-01&end_date=2026-02-28", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
