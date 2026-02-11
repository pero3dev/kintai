package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
)

// ===================================================================
// ExpenseHandler Tests
// ===================================================================

// ----- GetList -----

func TestExpenseHandler_GetList_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{
		GetListFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
			return []model.Expense{{BaseModel: model.BaseModel{ID: uuid.New()}, Title: "交通費"}}, 1, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetList(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses?page=1&page_size=10&status=pending&category=transportation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	var resp model.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Total != 1 {
		t.Errorf("Expected total 1, got %d", resp.Total)
	}
}

func TestExpenseHandler_GetList_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses", handler.GetList)

	req, _ := http.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseHandler_GetList_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{
		GetListFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetList(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpenseHandler_GetList_DefaultPagination(t *testing.T) {
	userID := uuid.New()
	var capturedPage, capturedPageSize int
	mockService := &mocks.MockExpenseService{
		GetListFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
			capturedPage = page
			capturedPageSize = pageSize
			return []model.Expense{}, 0, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetList(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if capturedPage != 1 {
		t.Errorf("Expected default page 1, got %d", capturedPage)
	}
	if capturedPageSize != 20 {
		t.Errorf("Expected default page_size 20, got %d", capturedPageSize)
	}
}

// ----- GetByID -----

func TestExpenseHandler_GetByID_Success(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Expense, error) {
			return &model.Expense{BaseModel: model.BaseModel{ID: id}, Title: "テスト経費"}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/"+expenseID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_GetByID_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_GetByID_NotFound(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.Expense, error) {
			return nil, errors.New("not found")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ----- Create -----

func TestExpenseHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseCreateRequest) (*model.Expense, error) {
			return &model.Expense{BaseModel: model.BaseModel{ID: uuid.New()}, Title: req.Title, UserID: uid}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"title":"交通費申請","items":[{"expense_date":"2026-02-10","category":"transportation","description":"電車代","amount":500}]}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExpenseHandler_Create_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses", handler.Create)

	body := `{"title":"交通費申請","items":[{"expense_date":"2026-02-10","category":"transportation","description":"電車代","amount":500}]}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseHandler_Create_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_Create_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseCreateRequest) (*model.Expense, error) {
			return nil, errors.New("create failed")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"title":"交通費申請","items":[{"expense_date":"2026-02-10","category":"transportation","description":"電車代","amount":500}]}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Update -----

func TestExpenseHandler_Update_Success(t *testing.T) {
	userID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		UpdateFunc: func(ctx context.Context, id, uid uuid.UUID, req *model.ExpenseUpdateRequest) (*model.Expense, error) {
			return &model.Expense{BaseModel: model.BaseModel{ID: id}, UserID: uid}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Update(c)
	})

	body := `{"title":"更新後タイトル"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_Update_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Update(c)
	})

	body := `{"title":"更新"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_Update_Unauthorized(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id", handler.Update)

	body := `{"title":"更新"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseHandler_Update_BadRequest(t *testing.T) {
	userID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Update(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String(), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_Update_ServiceError(t *testing.T) {
	userID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		UpdateFunc: func(ctx context.Context, id, uid uuid.UUID, req *model.ExpenseUpdateRequest) (*model.Expense, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Update(c)
	})

	body := `{"title":"更新"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Delete -----

func TestExpenseHandler_Delete_Success(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/"+expenseID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_Delete_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_Delete_ServiceError(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("delete failed")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/"+expenseID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetPending -----

func TestExpenseHandler_GetPending_Success(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error) {
			return []model.Expense{{BaseModel: model.BaseModel{ID: uuid.New()}}}, 1, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/pending?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_GetPending_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Approve -----

func TestExpenseHandler_Approve_Success(t *testing.T) {
	approverID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		ApproveFunc: func(ctx context.Context, id, aID uuid.UUID, req *model.ExpenseApproveRequest) error {
			return nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_Approve_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/invalid-uuid/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_Approve_Unauthorized(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/approve", handler.Approve)

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseHandler_Approve_BadRequest(t *testing.T) {
	approverID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.Approve(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/approve", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_Approve_ServiceError(t *testing.T) {
	approverID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		ApproveFunc: func(ctx context.Context, id, aID uuid.UUID, req *model.ExpenseApproveRequest) error {
			return errors.New("approve failed")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- AdvancedApprove -----

func TestExpenseHandler_AdvancedApprove_Success(t *testing.T) {
	approverID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		AdvancedApproveFunc: func(ctx context.Context, id, aID uuid.UUID, req *model.ExpenseAdvancedApproveRequest) error {
			return nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/advanced-approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.AdvancedApprove(c)
	})

	body := `{"action":"approve","step":1}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/advanced-approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_AdvancedApprove_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/advanced-approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.AdvancedApprove(c)
	})

	body := `{"action":"approve"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/invalid-uuid/advanced-approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_AdvancedApprove_Unauthorized(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/advanced-approve", handler.AdvancedApprove)

	body := `{"action":"approve"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/advanced-approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseHandler_AdvancedApprove_BadRequest(t *testing.T) {
	approverID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/advanced-approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.AdvancedApprove(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/advanced-approve", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHandler_AdvancedApprove_ServiceError(t *testing.T) {
	approverID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseService{
		AdvancedApproveFunc: func(ctx context.Context, id, aID uuid.UUID, req *model.ExpenseAdvancedApproveRequest) error {
			return errors.New("advanced approve failed")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/:id/advanced-approve", func(c *gin.Context) {
		c.Set("userID", approverID.String())
		handler.AdvancedApprove(c)
	})

	body := `{"action":"approve","step":1}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/"+expenseID.String()+"/advanced-approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetStats -----

func TestExpenseHandler_GetStats_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{
		GetStatsFunc: func(ctx context.Context, uid uuid.UUID) (*model.ExpenseStatsResponse, error) {
			return &model.ExpenseStatsResponse{TotalThisMonth: 10000, PendingCount: 2}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/stats", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetStats(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_GetStats_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/stats", handler.GetStats)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseHandler_GetStats_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseService{
		GetStatsFunc: func(ctx context.Context, uid uuid.UUID) (*model.ExpenseStatsResponse, error) {
			return nil, errors.New("stats error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/stats", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetStats(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetReport -----

func TestExpenseHandler_GetReport_Success(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetReportFunc: func(ctx context.Context, startDate, endDate string) (*model.ExpenseReportResponse, error) {
			return &model.ExpenseReportResponse{TotalAmount: 50000}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/report", handler.GetReport)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/report?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_GetReport_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetReportFunc: func(ctx context.Context, startDate, endDate string) (*model.ExpenseReportResponse, error) {
			return nil, errors.New("report error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/report", handler.GetReport)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/report?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpenseHandler_GetReport_NoQueryParams(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetReportFunc: func(ctx context.Context, startDate, endDate string) (*model.ExpenseReportResponse, error) {
			return &model.ExpenseReportResponse{}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/report", handler.GetReport)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/report", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ----- GetMonthlyTrend -----

func TestExpenseHandler_GetMonthlyTrend_Success(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetMonthlyTrendFunc: func(ctx context.Context, year string) ([]model.MonthlyTrendItem, error) {
			return []model.MonthlyTrendItem{{Month: "2026-01", Amount: 10000}}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/trend", handler.GetMonthlyTrend)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/trend?year=2026", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHandler_GetMonthlyTrend_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetMonthlyTrendFunc: func(ctx context.Context, year string) ([]model.MonthlyTrendItem, error) {
			return nil, errors.New("trend error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/trend", handler.GetMonthlyTrend)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/trend?year=2026", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpenseHandler_GetMonthlyTrend_NoYear(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		GetMonthlyTrendFunc: func(ctx context.Context, year string) ([]model.MonthlyTrendItem, error) {
			return []model.MonthlyTrendItem{}, nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/trend", handler.GetMonthlyTrend)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/trend", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ----- ExportCSV -----

func TestExpenseHandler_ExportCSV_Success(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		ExportCSVFunc: func(ctx context.Context, startDate, endDate string) ([]byte, error) {
			return []byte("ID,Title\n1,交通費"), nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/export/csv", handler.ExportCSV)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/export/csv?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
		t.Errorf("Expected Content-Type text/csv; charset=utf-8, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd != "attachment; filename=expenses.csv" {
		t.Errorf("Expected Content-Disposition attachment, got %s", cd)
	}
}

func TestExpenseHandler_ExportCSV_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		ExportCSVFunc: func(ctx context.Context, startDate, endDate string) ([]byte, error) {
			return nil, errors.New("csv export error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/export/csv", handler.ExportCSV)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/export/csv?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestExpenseHandler_ExportCSV_NoQueryParams(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		ExportCSVFunc: func(ctx context.Context, startDate, endDate string) ([]byte, error) {
			return []byte("ID,Title\n"), nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/export/csv", handler.ExportCSV)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/export/csv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ----- ExportPDF -----

func TestExpenseHandler_ExportPDF_Success(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		ExportPDFFunc: func(ctx context.Context, startDate, endDate string) ([]byte, error) {
			return []byte("%PDF-1.4 dummy"), nil
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/export/pdf", handler.ExportPDF)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/export/pdf?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Expected Content-Type application/pdf, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd != "attachment; filename=expenses.pdf" {
		t.Errorf("Expected Content-Disposition attachment, got %s", cd)
	}
}

func TestExpenseHandler_ExportPDF_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseService{
		ExportPDFFunc: func(ctx context.Context, startDate, endDate string) ([]byte, error) {
			return nil, errors.New("pdf export error")
		},
	}
	handler := NewExpenseHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/export/pdf", handler.ExportPDF)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/export/pdf?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- NewExpenseHandler -----

func TestNewExpenseHandler(t *testing.T) {
	mockService := &mocks.MockExpenseService{}
	log := getTestLogger()
	handler := NewExpenseHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
	if handler.svc != mockService {
		t.Error("Expected svc to be set")
	}
	if handler.logger != log {
		t.Error("Expected logger to be set")
	}
}

// ===================================================================
// ExpenseCommentHandler Tests
// ===================================================================

// ----- GetComments -----

func TestExpenseCommentHandler_GetComments_Success(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseCommentService{
		GetCommentsFunc: func(ctx context.Context, id uuid.UUID) ([]model.ExpenseCommentResponse, error) {
			return []model.ExpenseCommentResponse{{ID: uuid.New(), Content: "コメント"}}, nil
		},
	}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id/comments", handler.GetComments)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/"+expenseID.String()+"/comments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseCommentHandler_GetComments_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseCommentService{}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id/comments", handler.GetComments)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/invalid-uuid/comments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseCommentHandler_GetComments_ServiceError(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseCommentService{
		GetCommentsFunc: func(ctx context.Context, id uuid.UUID) ([]model.ExpenseCommentResponse, error) {
			return nil, errors.New("comments error")
		},
	}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id/comments", handler.GetComments)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/"+expenseID.String()+"/comments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- AddComment -----

func TestExpenseCommentHandler_AddComment_Success(t *testing.T) {
	userID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseCommentService{
		AddCommentFunc: func(ctx context.Context, eID, uID uuid.UUID, req *model.ExpenseCommentRequest) (*model.ExpenseCommentResponse, error) {
			return &model.ExpenseCommentResponse{ID: uuid.New(), Content: req.Content}, nil
		},
	}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/:id/comments", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.AddComment(c)
	})

	body := `{"content":"テストコメント"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/"+expenseID.String()+"/comments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExpenseCommentHandler_AddComment_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseCommentService{}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/:id/comments", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.AddComment(c)
	})

	body := `{"content":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/invalid-uuid/comments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseCommentHandler_AddComment_Unauthorized(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseCommentService{}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/:id/comments", handler.AddComment)

	body := `{"content":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/"+expenseID.String()+"/comments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseCommentHandler_AddComment_BadRequest(t *testing.T) {
	userID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseCommentService{}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/:id/comments", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.AddComment(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses/"+expenseID.String()+"/comments", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseCommentHandler_AddComment_ServiceError(t *testing.T) {
	userID := uuid.New()
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseCommentService{
		AddCommentFunc: func(ctx context.Context, eID, uID uuid.UUID, req *model.ExpenseCommentRequest) (*model.ExpenseCommentResponse, error) {
			return nil, errors.New("add comment failed")
		},
	}
	handler := NewExpenseCommentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/:id/comments", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.AddComment(c)
	})

	body := `{"content":"テスト"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/"+expenseID.String()+"/comments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- NewExpenseCommentHandler -----

func TestNewExpenseCommentHandler(t *testing.T) {
	mockService := &mocks.MockExpenseCommentService{}
	log := getTestLogger()
	handler := NewExpenseCommentHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

// ===================================================================
// ExpenseHistoryHandler Tests
// ===================================================================

func TestExpenseHistoryHandler_GetHistory_Success(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseHistoryService{
		GetHistoryFunc: func(ctx context.Context, id uuid.UUID) ([]model.ExpenseHistoryResponse, error) {
			return []model.ExpenseHistoryResponse{{ID: uuid.New(), Action: "作成"}}, nil
		},
	}
	handler := NewExpenseHistoryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id/history", handler.GetHistory)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/"+expenseID.String()+"/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseHistoryHandler_GetHistory_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseHistoryService{}
	handler := NewExpenseHistoryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id/history", handler.GetHistory)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/invalid-uuid/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseHistoryHandler_GetHistory_ServiceError(t *testing.T) {
	expenseID := uuid.New()
	mockService := &mocks.MockExpenseHistoryService{
		GetHistoryFunc: func(ctx context.Context, id uuid.UUID) ([]model.ExpenseHistoryResponse, error) {
			return nil, errors.New("history error")
		},
	}
	handler := NewExpenseHistoryHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/:id/history", handler.GetHistory)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/"+expenseID.String()+"/history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewExpenseHistoryHandler(t *testing.T) {
	mockService := &mocks.MockExpenseHistoryService{}
	log := getTestLogger()
	handler := NewExpenseHistoryHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

// ===================================================================
// ExpenseReceiptHandler Tests
// ===================================================================

func TestExpenseReceiptHandler_Upload_Success(t *testing.T) {
	mockService := &mocks.MockExpenseReceiptService{
		UploadFunc: func(ctx context.Context, filename string, data []byte) (string, error) {
			return "https://storage.example.com/receipts/" + filename, nil
		},
	}
	handler := NewExpenseReceiptHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/receipts", handler.Upload)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "receipt.png")
	part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/expenses/receipts", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseReceiptHandler_Upload_NoFile(t *testing.T) {
	mockService := &mocks.MockExpenseReceiptService{}
	handler := NewExpenseReceiptHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/receipts", handler.Upload)

	req, _ := http.NewRequest(http.MethodPost, "/expenses/receipts", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseReceiptHandler_Upload_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseReceiptService{
		UploadFunc: func(ctx context.Context, filename string, data []byte) (string, error) {
			return "", errors.New("upload failed")
		},
	}
	handler := NewExpenseReceiptHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/receipts", handler.Upload)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "receipt.png")
	part.Write([]byte("fake image data"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/expenses/receipts", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewExpenseReceiptHandler(t *testing.T) {
	mockService := &mocks.MockExpenseReceiptService{}
	log := getTestLogger()
	handler := NewExpenseReceiptHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

// ===================================================================
// ExpenseTemplateHandler Tests
// ===================================================================

// ----- GetTemplates -----

func TestExpenseTemplateHandler_GetTemplates_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		GetTemplatesFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpenseTemplate, error) {
			return []model.ExpenseTemplate{{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "通勤費"}}, nil
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/templates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetTemplates(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseTemplateHandler_GetTemplates_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/templates", handler.GetTemplates)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseTemplateHandler_GetTemplates_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		GetTemplatesFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpenseTemplate, error) {
			return nil, errors.New("templates error")
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/templates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetTemplates(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Create Template -----

func TestExpenseTemplateHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
			return &model.ExpenseTemplate{BaseModel: model.BaseModel{ID: uuid.New()}, Name: req.Name}, nil
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"name":"通勤費テンプレート","category":"transportation","amount":500}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExpenseTemplateHandler_Create_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates", handler.Create)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseTemplateHandler_Create_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseTemplateHandler_Create_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
			return nil, errors.New("create failed")
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.Create(c)
	})

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Update Template -----

func TestExpenseTemplateHandler_Update_Success(t *testing.T) {
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
			return &model.ExpenseTemplate{BaseModel: model.BaseModel{ID: id}, Name: req.Name}, nil
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/templates/:id", handler.Update)

	body := `{"name":"更新テンプレート"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/templates/"+templateID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseTemplateHandler_Update_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/templates/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/templates/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseTemplateHandler_Update_BadRequest(t *testing.T) {
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/templates/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/templates/"+templateID.String(), bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseTemplateHandler_Update_ServiceError(t *testing.T) {
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/templates/:id", handler.Update)

	body := `{"name":"test"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/templates/"+templateID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Delete Template -----

func TestExpenseTemplateHandler_Delete_Success(t *testing.T) {
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/templates/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/templates/"+templateID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseTemplateHandler_Delete_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/templates/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/templates/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseTemplateHandler_Delete_ServiceError(t *testing.T) {
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("delete failed")
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/templates/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/templates/"+templateID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- UseTemplate -----

func TestExpenseTemplateHandler_UseTemplate_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		UseTemplateFunc: func(ctx context.Context, id, uid uuid.UUID) (*model.Expense, error) {
			return &model.Expense{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates/:id/use", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.UseTemplate(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates/"+templateID.String()+"/use", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExpenseTemplateHandler_UseTemplate_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates/:id/use", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.UseTemplate(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates/invalid-uuid/use", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseTemplateHandler_UseTemplate_Unauthorized(t *testing.T) {
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates/:id/use", handler.UseTemplate)

	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates/"+templateID.String()+"/use", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseTemplateHandler_UseTemplate_ServiceError(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()
	mockService := &mocks.MockExpenseTemplateService{
		UseTemplateFunc: func(ctx context.Context, id, uid uuid.UUID) (*model.Expense, error) {
			return nil, errors.New("use template failed")
		},
	}
	handler := NewExpenseTemplateHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/templates/:id/use", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.UseTemplate(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses/templates/"+templateID.String()+"/use", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewExpenseTemplateHandler(t *testing.T) {
	mockService := &mocks.MockExpenseTemplateService{}
	log := getTestLogger()
	handler := NewExpenseTemplateHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

// ===================================================================
// ExpensePolicyHandler Tests
// ===================================================================

// ----- GetPolicies -----

func TestExpensePolicyHandler_GetPolicies_Success(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{
		GetPoliciesFunc: func(ctx context.Context) ([]model.ExpensePolicy, error) {
			return []model.ExpensePolicy{{BaseModel: model.BaseModel{ID: uuid.New()}}}, nil
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/policies", handler.GetPolicies)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/policies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpensePolicyHandler_GetPolicies_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{
		GetPoliciesFunc: func(ctx context.Context) ([]model.ExpensePolicy, error) {
			return nil, errors.New("policies error")
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/policies", handler.GetPolicies)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/policies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Create Policy -----

func TestExpensePolicyHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{
		CreateFunc: func(ctx context.Context, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
			return &model.ExpensePolicy{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/policies", handler.Create)

	body := `{"category":"transportation","monthly_limit":50000}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExpensePolicyHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/policies", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/expenses/policies", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpensePolicyHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{
		CreateFunc: func(ctx context.Context, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
			return nil, errors.New("create failed")
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/policies", handler.Create)

	body := `{"category":"transportation"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/policies", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Update Policy -----

func TestExpensePolicyHandler_Update_Success(t *testing.T) {
	policyID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
			return &model.ExpensePolicy{BaseModel: model.BaseModel{ID: id}}, nil
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/policies/:id", handler.Update)

	body := `{"category":"meals","monthly_limit":30000}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/policies/"+policyID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpensePolicyHandler_Update_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/policies/:id", handler.Update)

	body := `{"category":"meals"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/policies/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpensePolicyHandler_Update_BadRequest(t *testing.T) {
	policyID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/policies/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/policies/"+policyID.String(), bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpensePolicyHandler_Update_ServiceError(t *testing.T) {
	policyID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
			return nil, errors.New("update failed")
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/policies/:id", handler.Update)

	body := `{"category":"meals"}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/policies/"+policyID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- Delete Policy -----

func TestExpensePolicyHandler_Delete_Success(t *testing.T) {
	policyID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/policies/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/policies/"+policyID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpensePolicyHandler_Delete_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/policies/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/policies/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpensePolicyHandler_Delete_ServiceError(t *testing.T) {
	policyID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("delete failed")
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/policies/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/policies/"+policyID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetBudgets -----

func TestExpensePolicyHandler_GetBudgets_Success(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{
		GetBudgetsFunc: func(ctx context.Context) ([]model.ExpenseBudget, error) {
			return []model.ExpenseBudget{{BaseModel: model.BaseModel{ID: uuid.New()}}}, nil
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/budgets", handler.GetBudgets)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/budgets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpensePolicyHandler_GetBudgets_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{
		GetBudgetsFunc: func(ctx context.Context) ([]model.ExpenseBudget, error) {
			return nil, errors.New("budgets error")
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/budgets", handler.GetBudgets)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/budgets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetPolicyViolations -----

func TestExpensePolicyHandler_GetPolicyViolations_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{
		GetPolicyViolationsFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpensePolicyViolation, error) {
			return []model.ExpensePolicyViolation{}, nil
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/violations", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetPolicyViolations(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/violations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpensePolicyHandler_GetPolicyViolations_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/violations", handler.GetPolicyViolations)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/violations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpensePolicyHandler_GetPolicyViolations_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpensePolicyService{
		GetPolicyViolationsFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpensePolicyViolation, error) {
			return nil, errors.New("violations error")
		},
	}
	handler := NewExpensePolicyHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/violations", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetPolicyViolations(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/violations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewExpensePolicyHandler(t *testing.T) {
	mockService := &mocks.MockExpensePolicyService{}
	log := getTestLogger()
	handler := NewExpensePolicyHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

// ===================================================================
// ExpenseNotificationHandler Tests
// ===================================================================

// ----- GetNotifications -----

func TestExpenseNotificationHandler_GetNotifications_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		GetNotificationsFunc: func(ctx context.Context, uid uuid.UUID, filter string) ([]model.ExpenseNotification, error) {
			return []model.ExpenseNotification{{BaseModel: model.BaseModel{ID: uuid.New()}, Message: "承認されました"}}, nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/notifications", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetNotifications(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/notifications?filter=unread", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_GetNotifications_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/notifications", handler.GetNotifications)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseNotificationHandler_GetNotifications_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		GetNotificationsFunc: func(ctx context.Context, uid uuid.UUID, filter string) ([]model.ExpenseNotification, error) {
			return nil, errors.New("notifications error")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/notifications", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetNotifications(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- MarkAsRead -----

func TestExpenseNotificationHandler_MarkAsRead_Success(t *testing.T) {
	notifID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notifications/"+notifID.String()+"/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_MarkAsRead_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notifications/invalid-uuid/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseNotificationHandler_MarkAsRead_ServiceError(t *testing.T) {
	notifID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("mark read failed")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notifications/"+notifID.String()+"/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- MarkAllAsRead -----

func TestExpenseNotificationHandler_MarkAllAsRead_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		MarkAllAsReadFunc: func(ctx context.Context, uid uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notifications/read-all", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.MarkAllAsRead(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_MarkAllAsRead_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notifications/read-all", handler.MarkAllAsRead)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseNotificationHandler_MarkAllAsRead_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		MarkAllAsReadFunc: func(ctx context.Context, uid uuid.UUID) error {
			return errors.New("mark all read failed")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notifications/read-all", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.MarkAllAsRead(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetReminders -----

func TestExpenseNotificationHandler_GetReminders_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		GetRemindersFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpenseReminder, error) {
			return []model.ExpenseReminder{{BaseModel: model.BaseModel{ID: uuid.New()}}}, nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/reminders", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetReminders(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/reminders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_GetReminders_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/reminders", handler.GetReminders)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/reminders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseNotificationHandler_GetReminders_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		GetRemindersFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpenseReminder, error) {
			return nil, errors.New("reminders error")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/reminders", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetReminders(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/reminders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- DismissReminder -----

func TestExpenseNotificationHandler_DismissReminder_Success(t *testing.T) {
	reminderID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		DismissReminderFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/reminders/:id/dismiss", handler.DismissReminder)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/reminders/"+reminderID.String()+"/dismiss", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_DismissReminder_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/reminders/:id/dismiss", handler.DismissReminder)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/reminders/invalid-uuid/dismiss", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseNotificationHandler_DismissReminder_ServiceError(t *testing.T) {
	reminderID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		DismissReminderFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("dismiss failed")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/reminders/:id/dismiss", handler.DismissReminder)

	req, _ := http.NewRequest(http.MethodPut, "/expenses/reminders/"+reminderID.String()+"/dismiss", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetSettings -----

func TestExpenseNotificationHandler_GetSettings_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		GetSettingsFunc: func(ctx context.Context, uid uuid.UUID) (*model.ExpenseNotificationSetting, error) {
			return &model.ExpenseNotificationSetting{EmailEnabled: true}, nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/notification-settings", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetSettings(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/notification-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_GetSettings_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/notification-settings", handler.GetSettings)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/notification-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseNotificationHandler_GetSettings_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		GetSettingsFunc: func(ctx context.Context, uid uuid.UUID) (*model.ExpenseNotificationSetting, error) {
			return nil, errors.New("settings error")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/notification-settings", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetSettings(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/notification-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- UpdateSettings -----

func TestExpenseNotificationHandler_UpdateSettings_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		UpdateSettingsFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseNotificationSettingRequest) (*model.ExpenseNotificationSetting, error) {
			return &model.ExpenseNotificationSetting{EmailEnabled: true}, nil
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notification-settings", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.UpdateSettings(c)
	})

	body := `{"email_enabled":true,"push_enabled":false}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/notification-settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseNotificationHandler_UpdateSettings_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notification-settings", handler.UpdateSettings)

	body := `{"email_enabled":true}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/notification-settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseNotificationHandler_UpdateSettings_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notification-settings", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.UpdateSettings(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/expenses/notification-settings", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseNotificationHandler_UpdateSettings_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseNotificationService{
		UpdateSettingsFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseNotificationSettingRequest) (*model.ExpenseNotificationSetting, error) {
			return nil, errors.New("update settings failed")
		},
	}
	handler := NewExpenseNotificationHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/expenses/notification-settings", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.UpdateSettings(c)
	})

	body := `{"email_enabled":true}`
	req, _ := http.NewRequest(http.MethodPut, "/expenses/notification-settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewExpenseNotificationHandler(t *testing.T) {
	mockService := &mocks.MockExpenseNotificationService{}
	log := getTestLogger()
	handler := NewExpenseNotificationHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}

// ===================================================================
// ExpenseApprovalFlowHandler Tests
// ===================================================================

// ----- GetConfig -----

func TestExpenseApprovalFlowHandler_GetConfig_Success(t *testing.T) {
	mockService := &mocks.MockExpenseApprovalFlowService{
		GetConfigFunc: func(ctx context.Context) (*model.ExpenseApprovalFlow, error) {
			return &model.ExpenseApprovalFlow{Name: "default"}, nil
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/approval-flow", handler.GetConfig)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/approval-flow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_GetConfig_ServiceError(t *testing.T) {
	mockService := &mocks.MockExpenseApprovalFlowService{
		GetConfigFunc: func(ctx context.Context) (*model.ExpenseApprovalFlow, error) {
			return nil, errors.New("config error")
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/approval-flow", handler.GetConfig)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/approval-flow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- GetDelegates -----

func TestExpenseApprovalFlowHandler_GetDelegates_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{
		GetDelegatesFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpenseDelegate, error) {
			return []model.ExpenseDelegate{}, nil
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/delegates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetDelegates(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/delegates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_GetDelegates_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseApprovalFlowService{}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/delegates", handler.GetDelegates)

	req, _ := http.NewRequest(http.MethodGet, "/expenses/delegates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_GetDelegates_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{
		GetDelegatesFunc: func(ctx context.Context, uid uuid.UUID) ([]model.ExpenseDelegate, error) {
			return nil, errors.New("delegates error")
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/expenses/delegates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetDelegates(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/expenses/delegates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- SetDelegate -----

func TestExpenseApprovalFlowHandler_SetDelegate_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{
		SetDelegateFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseDelegateRequest) (*model.ExpenseDelegate, error) {
			return &model.ExpenseDelegate{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/delegates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.SetDelegate(c)
	})

	body := `{"delegate_to":"` + uuid.New().String() + `","start_date":"2026-02-10","end_date":"2026-02-20"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/delegates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_SetDelegate_Unauthorized(t *testing.T) {
	mockService := &mocks.MockExpenseApprovalFlowService{}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/delegates", handler.SetDelegate)

	body := `{"delegate_to":"test","start_date":"2026-02-10","end_date":"2026-02-20"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/delegates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_SetDelegate_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/delegates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.SetDelegate(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/expenses/delegates", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_SetDelegate_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{
		SetDelegateFunc: func(ctx context.Context, uid uuid.UUID, req *model.ExpenseDelegateRequest) (*model.ExpenseDelegate, error) {
			return nil, errors.New("set delegate failed")
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/expenses/delegates", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.SetDelegate(c)
	})

	body := `{"delegate_to":"test","start_date":"2026-02-10","end_date":"2026-02-20"}`
	req, _ := http.NewRequest(http.MethodPost, "/expenses/delegates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ----- RemoveDelegate -----

func TestExpenseApprovalFlowHandler_RemoveDelegate_Success(t *testing.T) {
	delegateID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{
		RemoveDelegateFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/delegates/:id", handler.RemoveDelegate)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/delegates/"+delegateID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_RemoveDelegate_InvalidUUID(t *testing.T) {
	mockService := &mocks.MockExpenseApprovalFlowService{}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/delegates/:id", handler.RemoveDelegate)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/delegates/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestExpenseApprovalFlowHandler_RemoveDelegate_ServiceError(t *testing.T) {
	delegateID := uuid.New()
	mockService := &mocks.MockExpenseApprovalFlowService{
		RemoveDelegateFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("remove delegate failed")
		},
	}
	handler := NewExpenseApprovalFlowHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/expenses/delegates/:id", handler.RemoveDelegate)

	req, _ := http.NewRequest(http.MethodDelete, "/expenses/delegates/"+delegateID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestNewExpenseApprovalFlowHandler(t *testing.T) {
	mockService := &mocks.MockExpenseApprovalFlowService{}
	log := getTestLogger()
	handler := NewExpenseApprovalFlowHandler(mockService, log)
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}
