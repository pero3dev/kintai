package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
)

func jsonReq(method, url, body string) *http.Request {
	req, _ := http.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withUser(userID uuid.UUID, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID.String())
		h(c)
	}
}

func serve(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var errTest = errors.New("test error")

// ===================================================================
// HREmployeeHandler Tests
// ===================================================================

func TestHREmployeeHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			CreateFunc: func(ctx context.Context, req model.HREmployeeCreateRequest) (*model.HREmployee, error) {
				return &model.HREmployee{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/e", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/e", `{"employee_code":"E001","first_name":"太郎","last_name":"山田","email":"t@example.com"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewHREmployeeHandler(&mocks.MockHREmployeeService{}, getTestLogger())
		r := setupRouter()
		r.POST("/e", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/e", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			CreateFunc: func(ctx context.Context, req model.HREmployeeCreateRequest) (*model.HREmployee, error) {
				return nil, errTest
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/e", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/e", `{"employee_code":"E001","first_name":"太郎","last_name":"山田","email":"t@example.com"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHREmployeeHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HREmployee, error) {
				return &model.HREmployee{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/e/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/e/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewHREmployeeHandler(&mocks.MockHREmployeeService{}, getTestLogger())
		r := setupRouter()
		r.GET("/e/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/e/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HREmployee, error) {
				return nil, errTest
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/e/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/e/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestHREmployeeHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, dept, status, empType, search string) ([]model.HREmployee, int64, error) {
				return []model.HREmployee{}, 0, nil
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/e", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/e?page=1&page_size=10&department=dev&status=active&employment_type=full_time&search=test", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, dept, status, empType, search string) ([]model.HREmployee, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/e", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/e", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHREmployeeHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.HREmployeeUpdateRequest) (*model.HREmployee, error) {
				return &model.HREmployee{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/e/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/e/"+id.String(), `{"first_name":"次郎"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewHREmployeeHandler(&mocks.MockHREmployeeService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/e/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/e/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewHREmployeeHandler(&mocks.MockHREmployeeService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/e/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/e/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.HREmployeeUpdateRequest) (*model.HREmployee, error) {
				return nil, errTest
			},
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/e/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/e/"+id.String(), `{"first_name":"次郎"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHREmployeeHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil },
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/e/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/e/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewHREmployeeHandler(&mocks.MockHREmployeeService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/e/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/e/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHREmployeeService{
			DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest },
		}
		h := NewHREmployeeHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/e/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/e/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// HRDepartmentHandler Tests
// ===================================================================

func TestHRDepartmentHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			CreateFunc: func(ctx context.Context, req model.HRDepartmentCreateRequest) (*model.HRDepartment, error) {
				return &model.HRDepartment{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/d", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/d", `{"name":"開発部"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewHRDepartmentHandler(&mocks.MockHRDepartmentService{}, getTestLogger())
		r := setupRouter()
		r.POST("/d", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/d", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			CreateFunc: func(ctx context.Context, req model.HRDepartmentCreateRequest) (*model.HRDepartment, error) {
				return nil, errTest
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/d", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/d", `{"name":"開発部"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHRDepartmentHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRDepartment, error) {
				return &model.HRDepartment{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/d/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/d/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewHRDepartmentHandler(&mocks.MockHRDepartmentService{}, getTestLogger())
		r := setupRouter()
		r.GET("/d/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/d/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRDepartment, error) {
				return nil, errTest
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/d/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/d/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestHRDepartmentHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			FindAllFunc: func(ctx context.Context) ([]model.HRDepartment, error) {
				return []model.HRDepartment{}, nil
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/d", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/d", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			FindAllFunc: func(ctx context.Context) ([]model.HRDepartment, error) {
				return nil, errTest
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/d", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/d", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHRDepartmentHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.HRDepartmentUpdateRequest) (*model.HRDepartment, error) {
				return &model.HRDepartment{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/d/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/d/"+id.String(), `{"name":"営業部"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewHRDepartmentHandler(&mocks.MockHRDepartmentService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/d/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/d/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewHRDepartmentHandler(&mocks.MockHRDepartmentService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/d/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/d/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.HRDepartmentUpdateRequest) (*model.HRDepartment, error) {
				return nil, errTest
			},
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/d/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/d/"+id.String(), `{"name":"営業部"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHRDepartmentHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil },
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/d/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/d/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewHRDepartmentHandler(&mocks.MockHRDepartmentService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/d/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/d/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHRDepartmentService{
			DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest },
		}
		h := NewHRDepartmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/d/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/d/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// EvaluationHandler Tests
// ===================================================================

func TestEvaluationHandler_Create(t *testing.T) {
	userID := uuid.New()
	body := `{"employee_id":"` + uuid.New().String() + `","cycle_id":"` + uuid.New().String() + `"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			CreateFunc: func(ctx context.Context, req model.EvaluationCreateRequest, rid uuid.UUID) (*model.Evaluation, error) {
				return &model.Evaluation{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ev", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/ev", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ev", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/ev", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ev", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/ev", body))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			CreateFunc: func(ctx context.Context, req model.EvaluationCreateRequest, rid uuid.UUID) (*model.Evaluation, error) {
				return nil, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ev", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/ev", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestEvaluationHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Evaluation, error) {
				return &model.Evaluation{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ev/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ev/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.GET("/ev/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ev/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Evaluation, error) {
				return nil, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ev/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ev/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestEvaluationHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error) {
				return []model.Evaluation{}, 0, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ev", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/ev?cycle_id=abc&status=draft", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ev", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/ev", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestEvaluationHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.EvaluationUpdateRequest) (*model.Evaluation, error) {
				return &model.Evaluation{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ev/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ev/"+id.String(), `{"self_comment":"good"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ev/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ev/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ev/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ev/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.EvaluationUpdateRequest) (*model.Evaluation, error) {
				return nil, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ev/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ev/"+id.String(), `{"self_comment":"good"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestEvaluationHandler_Submit(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			SubmitFunc: func(ctx context.Context, i uuid.UUID) (*model.Evaluation, error) {
				return &model.Evaluation{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ev/:id/submit", h.Submit)
		w := serve(r, jsonReq(http.MethodPost, "/ev/"+id.String()+"/submit", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ev/:id/submit", h.Submit)
		w := serve(r, jsonReq(http.MethodPost, "/ev/invalid/submit", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			SubmitFunc: func(ctx context.Context, i uuid.UUID) (*model.Evaluation, error) {
				return nil, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ev/:id/submit", h.Submit)
		w := serve(r, jsonReq(http.MethodPost, "/ev/"+id.String()+"/submit", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestEvaluationHandler_GetCycles(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			FindAllCyclesFunc: func(ctx context.Context) ([]model.EvaluationCycle, error) {
				return []model.EvaluationCycle{}, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/cycles", h.GetCycles)
		w := serve(r, jsonReq(http.MethodGet, "/cycles", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			FindAllCyclesFunc: func(ctx context.Context) ([]model.EvaluationCycle, error) {
				return nil, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/cycles", h.GetCycles)
		w := serve(r, jsonReq(http.MethodGet, "/cycles", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestEvaluationHandler_CreateCycle(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			CreateCycleFunc: func(ctx context.Context, req model.EvaluationCycleCreateRequest) (*model.EvaluationCycle, error) {
				return &model.EvaluationCycle{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/cycles", h.CreateCycle)
		w := serve(r, jsonReq(http.MethodPost, "/cycles", `{"name":"2025H1","start_date":"2025-01-01","end_date":"2025-06-30"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewEvaluationHandler(&mocks.MockEvaluationService{}, getTestLogger())
		r := setupRouter()
		r.POST("/cycles", h.CreateCycle)
		w := serve(r, jsonReq(http.MethodPost, "/cycles", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockEvaluationService{
			CreateCycleFunc: func(ctx context.Context, req model.EvaluationCycleCreateRequest) (*model.EvaluationCycle, error) {
				return nil, errTest
			},
		}
		h := NewEvaluationHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/cycles", h.CreateCycle)
		w := serve(r, jsonReq(http.MethodPost, "/cycles", `{"name":"2025H1","start_date":"2025-01-01","end_date":"2025-06-30"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// GoalHandler Tests
// ===================================================================

func TestGoalHandler_Create(t *testing.T) {
	userID := uuid.New()
	body := `{"title":"目標1"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			CreateFunc: func(ctx context.Context, req model.HRGoalCreateRequest, uid uuid.UUID) (*model.HRGoal, error) {
				return &model.HRGoal{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/g", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/g", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.POST("/g", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/g", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.POST("/g", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/g", body))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			CreateFunc: func(ctx context.Context, req model.HRGoalCreateRequest, uid uuid.UUID) (*model.HRGoal, error) {
				return nil, errTest
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/g", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/g", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestGoalHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRGoal, error) {
				return &model.HRGoal{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/g/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/g/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.GET("/g/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/g/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRGoal, error) { return nil, errTest },
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/g/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/g/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestGoalHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error) {
				return []model.HRGoal{}, 0, nil
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/g", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/g?status=in_progress&category=performance&employee_id=abc", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/g", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/g", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestGoalHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.HRGoalUpdateRequest) (*model.HRGoal, error) {
				return &model.HRGoal{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/g/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/g/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/g/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.HRGoalUpdateRequest) (*model.HRGoal, error) {
				return nil, errTest
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/g/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestGoalHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil },
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/g/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/g/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/g/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/g/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest },
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/g/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/g/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestGoalHandler_UpdateProgress(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			UpdateProgressFunc: func(ctx context.Context, i uuid.UUID, progress int) (*model.HRGoal, error) {
				return &model.HRGoal{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id/progress", h.UpdateProgress)
		w := serve(r, jsonReq(http.MethodPut, "/g/"+id.String()+"/progress", `{"progress":50}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id/progress", h.UpdateProgress)
		w := serve(r, jsonReq(http.MethodPut, "/g/invalid/progress", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewGoalHandler(&mocks.MockGoalService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id/progress", h.UpdateProgress)
		w := serve(r, jsonReq(http.MethodPut, "/g/"+id.String()+"/progress", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockGoalService{
			UpdateProgressFunc: func(ctx context.Context, i uuid.UUID, progress int) (*model.HRGoal, error) {
				return nil, errTest
			},
		}
		h := NewGoalHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/g/:id/progress", h.UpdateProgress)
		w := serve(r, jsonReq(http.MethodPut, "/g/"+id.String()+"/progress", `{"progress":50}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// TrainingHandler Tests
// ===================================================================

func TestTrainingHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			CreateFunc: func(ctx context.Context, req model.TrainingProgramCreateRequest) (*model.TrainingProgram, error) {
				return &model.TrainingProgram{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/tr", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/tr", `{"title":"研修1"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/tr", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/tr", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			CreateFunc: func(ctx context.Context, req model.TrainingProgramCreateRequest) (*model.TrainingProgram, error) {
				return nil, errTest
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/tr", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/tr", `{"title":"研修1"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestTrainingHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.TrainingProgram, error) {
				return &model.TrainingProgram{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/tr/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/tr/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.GET("/tr/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/tr/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.TrainingProgram, error) { return nil, errTest },
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/tr/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/tr/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestTrainingHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error) {
				return []model.TrainingProgram{}, 0, nil
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/tr", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/tr?category=tech&status=scheduled", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/tr", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/tr", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestTrainingHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.TrainingProgramUpdateRequest) (*model.TrainingProgram, error) {
				return &model.TrainingProgram{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/tr/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/tr/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/tr/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/tr/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/tr/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/tr/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockTrainingService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.TrainingProgramUpdateRequest) (*model.TrainingProgram, error) {
				return nil, errTest
			},
		}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/tr/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/tr/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestTrainingHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil }}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/tr/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/tr/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/tr/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/tr/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockTrainingService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest }}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/tr/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/tr/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestTrainingHandler_Enroll(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{EnrollFunc: func(ctx context.Context, pid, eid uuid.UUID) error { return nil }}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/enroll", withUser(userID, h.Enroll))
		w := serve(r, jsonReq(http.MethodPost, "/tr/"+id.String()+"/enroll", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/enroll", withUser(userID, h.Enroll))
		w := serve(r, jsonReq(http.MethodPost, "/tr/invalid/enroll", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/enroll", h.Enroll)
		w := serve(r, jsonReq(http.MethodPost, "/tr/"+id.String()+"/enroll", ""))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockTrainingService{EnrollFunc: func(ctx context.Context, pid, eid uuid.UUID) error { return errTest }}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/enroll", withUser(userID, h.Enroll))
		w := serve(r, jsonReq(http.MethodPost, "/tr/"+id.String()+"/enroll", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestTrainingHandler_Complete(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockTrainingService{CompleteFunc: func(ctx context.Context, pid, eid uuid.UUID) error { return nil }}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/complete", withUser(userID, h.Complete))
		w := serve(r, jsonReq(http.MethodPost, "/tr/"+id.String()+"/complete", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/complete", withUser(userID, h.Complete))
		w := serve(r, jsonReq(http.MethodPost, "/tr/invalid/complete", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewTrainingHandler(&mocks.MockTrainingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/complete", h.Complete)
		w := serve(r, jsonReq(http.MethodPost, "/tr/"+id.String()+"/complete", ""))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockTrainingService{CompleteFunc: func(ctx context.Context, pid, eid uuid.UUID) error { return errTest }}
		h := NewTrainingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/tr/:id/complete", withUser(userID, h.Complete))
		w := serve(r, jsonReq(http.MethodPost, "/tr/"+id.String()+"/complete", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// RecruitmentHandler Tests
// ===================================================================

func TestRecruitmentHandler_CreatePosition(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			CreatePositionFunc: func(ctx context.Context, req model.PositionCreateRequest) (*model.RecruitmentPosition, error) {
				return &model.RecruitmentPosition{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/pos", h.CreatePosition)
		w := serve(r, jsonReq(http.MethodPost, "/pos", `{"title":"エンジニア"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.POST("/pos", h.CreatePosition)
		w := serve(r, jsonReq(http.MethodPost, "/pos", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			CreatePositionFunc: func(ctx context.Context, req model.PositionCreateRequest) (*model.RecruitmentPosition, error) {
				return nil, errTest
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/pos", h.CreatePosition)
		w := serve(r, jsonReq(http.MethodPost, "/pos", `{"title":"エンジニア"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestRecruitmentHandler_GetPosition(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			FindPositionByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.RecruitmentPosition, error) {
				return &model.RecruitmentPosition{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/pos/:id", h.GetPosition)
		w := serve(r, jsonReq(http.MethodGet, "/pos/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.GET("/pos/:id", h.GetPosition)
		w := serve(r, jsonReq(http.MethodGet, "/pos/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			FindPositionByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.RecruitmentPosition, error) { return nil, errTest },
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/pos/:id", h.GetPosition)
		w := serve(r, jsonReq(http.MethodGet, "/pos/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestRecruitmentHandler_GetAllPositions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			FindAllPositionsFunc: func(ctx context.Context, page, pageSize int, status, dept string) ([]model.RecruitmentPosition, int64, error) {
				return []model.RecruitmentPosition{}, 0, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/pos", h.GetAllPositions)
		w := serve(r, jsonReq(http.MethodGet, "/pos?status=open&department=dev", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			FindAllPositionsFunc: func(ctx context.Context, page, pageSize int, status, dept string) ([]model.RecruitmentPosition, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/pos", h.GetAllPositions)
		w := serve(r, jsonReq(http.MethodGet, "/pos", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestRecruitmentHandler_UpdatePosition(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			UpdatePositionFunc: func(ctx context.Context, i uuid.UUID, req model.PositionUpdateRequest) (*model.RecruitmentPosition, error) {
				return &model.RecruitmentPosition{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/pos/:id", h.UpdatePosition)
		w := serve(r, jsonReq(http.MethodPut, "/pos/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/pos/:id", h.UpdatePosition)
		w := serve(r, jsonReq(http.MethodPut, "/pos/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/pos/:id", h.UpdatePosition)
		w := serve(r, jsonReq(http.MethodPut, "/pos/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			UpdatePositionFunc: func(ctx context.Context, i uuid.UUID, req model.PositionUpdateRequest) (*model.RecruitmentPosition, error) {
				return nil, errTest
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/pos/:id", h.UpdatePosition)
		w := serve(r, jsonReq(http.MethodPut, "/pos/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestRecruitmentHandler_CreateApplicant(t *testing.T) {
	body := `{"position_id":"` + uuid.New().String() + `","name":"田中","email":"t@example.com"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			CreateApplicantFunc: func(ctx context.Context, req model.ApplicantCreateRequest) (*model.Applicant, error) {
				return &model.Applicant{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/app", h.CreateApplicant)
		w := serve(r, jsonReq(http.MethodPost, "/app", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.POST("/app", h.CreateApplicant)
		w := serve(r, jsonReq(http.MethodPost, "/app", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			CreateApplicantFunc: func(ctx context.Context, req model.ApplicantCreateRequest) (*model.Applicant, error) {
				return nil, errTest
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/app", h.CreateApplicant)
		w := serve(r, jsonReq(http.MethodPost, "/app", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestRecruitmentHandler_GetAllApplicants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			FindAllApplicantsFunc: func(ctx context.Context, posID, stage string) ([]model.Applicant, error) {
				return []model.Applicant{}, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/app", h.GetAllApplicants)
		w := serve(r, jsonReq(http.MethodGet, "/app?position_id=abc&stage=new", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			FindAllApplicantsFunc: func(ctx context.Context, posID, stage string) ([]model.Applicant, error) {
				return nil, errTest
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/app", h.GetAllApplicants)
		w := serve(r, jsonReq(http.MethodGet, "/app", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestRecruitmentHandler_UpdateApplicantStage(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			UpdateApplicantStageFunc: func(ctx context.Context, i uuid.UUID, stage string) (*model.Applicant, error) {
				return &model.Applicant{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/app/:id/stage", h.UpdateApplicantStage)
		w := serve(r, jsonReq(http.MethodPut, "/app/"+id.String()+"/stage", `{"stage":"interview"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/app/:id/stage", h.UpdateApplicantStage)
		w := serve(r, jsonReq(http.MethodPut, "/app/invalid/stage", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewRecruitmentHandler(&mocks.MockRecruitmentService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/app/:id/stage", h.UpdateApplicantStage)
		w := serve(r, jsonReq(http.MethodPut, "/app/"+id.String()+"/stage", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockRecruitmentService{
			UpdateApplicantStageFunc: func(ctx context.Context, i uuid.UUID, stage string) (*model.Applicant, error) {
				return nil, errTest
			},
		}
		h := NewRecruitmentHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/app/:id/stage", h.UpdateApplicantStage)
		w := serve(r, jsonReq(http.MethodPut, "/app/"+id.String()+"/stage", `{"stage":"interview"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// DocumentHandler Tests
// ===================================================================

func createMultipartRequest(url string, fields map[string]string, fileName string, fileContent []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if fileName != "" {
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			return nil, err
		}
		part.Write(fileContent)
	}
	for k, v := range fields {
		writer.WriteField(k, v)
	}
	writer.Close()
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestDocumentHandler_Upload(t *testing.T) {
	userID := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockDocumentService{
			UploadFunc: func(ctx context.Context, doc *model.HRDocument) error { return nil },
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/doc", withUser(userID, h.Upload))
		req, _ := createMultipartRequest("/doc", map[string]string{"title": "test", "type": "contract"}, "test.txt", []byte("hello"))
		w := serve(r, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
		os.RemoveAll("uploads")
	})
	t.Run("WithEmployeeID", func(t *testing.T) {
		mock := &mocks.MockDocumentService{
			UploadFunc: func(ctx context.Context, doc *model.HRDocument) error { return nil },
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/doc", withUser(userID, h.Upload))
		eid := uuid.New()
		req, _ := createMultipartRequest("/doc", map[string]string{"employee_id": eid.String()}, "test.txt", []byte("hello"))
		w := serve(r, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
		os.RemoveAll("uploads")
	})
	t.Run("NoFile", func(t *testing.T) {
		h := NewDocumentHandler(&mocks.MockDocumentService{}, getTestLogger())
		r := setupRouter()
		r.POST("/doc", withUser(userID, h.Upload))
		req, _ := http.NewRequest(http.MethodPost, "/doc", nil)
		req.Header.Set("Content-Type", "multipart/form-data")
		w := serve(r, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockDocumentService{
			UploadFunc: func(ctx context.Context, doc *model.HRDocument) error { return errTest },
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/doc", withUser(userID, h.Upload))
		req, _ := createMultipartRequest("/doc", nil, "test.txt", []byte("hello"))
		w := serve(r, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
		os.RemoveAll("uploads")
	})
	t.Run("FileCreateError", func(t *testing.T) {
		// uploadsパスにファイルを作成してディレクトリ作成を妨害
		os.RemoveAll("uploads")
		os.MkdirAll("uploads", 0755)
		os.WriteFile("uploads/documents", []byte("block"), 0644)
		defer os.RemoveAll("uploads")

		h := NewDocumentHandler(&mocks.MockDocumentService{}, getTestLogger())
		r := setupRouter()
		r.POST("/doc", withUser(userID, h.Upload))
		req, _ := createMultipartRequest("/doc", nil, "test.txt", []byte("hello"))
		w := serve(r, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestDocumentHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockDocumentService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, docType, empID string) ([]model.HRDocument, int64, error) {
				return []model.HRDocument{}, 0, nil
			},
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/doc", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/doc?type=contract&employee_id=abc", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockDocumentService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, docType, empID string) ([]model.HRDocument, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/doc", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/doc", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestDocumentHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockDocumentService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil }}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/doc/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/doc/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewDocumentHandler(&mocks.MockDocumentService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/doc/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/doc/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockDocumentService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest }}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/doc/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/doc/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestDocumentHandler_Download(t *testing.T) {
	id := uuid.New()
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewDocumentHandler(&mocks.MockDocumentService{}, getTestLogger())
		r := setupRouter()
		r.GET("/doc/:id/download", h.Download)
		w := serve(r, jsonReq(http.MethodGet, "/doc/invalid/download", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockDocumentService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRDocument, error) { return nil, errTest },
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/doc/:id/download", h.Download)
		w := serve(r, jsonReq(http.MethodGet, "/doc/"+id.String()+"/download", ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
	t.Run("Success", func(t *testing.T) {
		tmpFile := "test_download_file.txt"
		os.WriteFile(tmpFile, []byte("test"), 0644)
		defer os.Remove(tmpFile)
		mock := &mocks.MockDocumentService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRDocument, error) {
				return &model.HRDocument{FilePath: tmpFile, FileName: "test.txt"}, nil
			},
		}
		h := NewDocumentHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/doc/:id/download", h.Download)
		w := serve(r, jsonReq(http.MethodGet, "/doc/"+id.String()+"/download", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

// ===================================================================
// AnnouncementHandler Tests
// ===================================================================

func TestAnnouncementHandler_Create(t *testing.T) {
	userID := uuid.New()
	body := `{"title":"お知らせ","content":"内容"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			CreateFunc: func(ctx context.Context, req model.AnnouncementCreateRequest, aid uuid.UUID) (*model.HRAnnouncement, error) {
				return &model.HRAnnouncement{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ann", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/ann", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewAnnouncementHandler(&mocks.MockAnnouncementService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ann", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/ann", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewAnnouncementHandler(&mocks.MockAnnouncementService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ann", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/ann", body))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			CreateFunc: func(ctx context.Context, req model.AnnouncementCreateRequest, aid uuid.UUID) (*model.HRAnnouncement, error) {
				return nil, errTest
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ann", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/ann", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestAnnouncementHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRAnnouncement, error) {
				return &model.HRAnnouncement{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ann/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ann/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewAnnouncementHandler(&mocks.MockAnnouncementService{}, getTestLogger())
		r := setupRouter()
		r.GET("/ann/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ann/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.HRAnnouncement, error) { return nil, errTest },
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ann/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ann/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestAnnouncementHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error) {
				return []model.HRAnnouncement{}, 0, nil
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ann", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/ann?priority=high", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			FindAllFunc: func(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error) {
				return nil, 0, errTest
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ann", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/ann", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestAnnouncementHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.AnnouncementUpdateRequest) (*model.HRAnnouncement, error) {
				return &model.HRAnnouncement{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ann/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ann/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewAnnouncementHandler(&mocks.MockAnnouncementService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ann/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ann/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewAnnouncementHandler(&mocks.MockAnnouncementService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ann/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ann/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.AnnouncementUpdateRequest) (*model.HRAnnouncement, error) {
				return nil, errTest
			},
		}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ann/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ann/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestAnnouncementHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil }}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/ann/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/ann/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewAnnouncementHandler(&mocks.MockAnnouncementService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/ann/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/ann/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAnnouncementService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest }}
		h := NewAnnouncementHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/ann/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/ann/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// HRDashboardHandler Tests
// ===================================================================

func TestHRDashboardHandler_GetStats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDashboardService{
			GetStatsFunc: func(ctx context.Context) (map[string]interface{}, error) { return map[string]interface{}{"total": 10}, nil },
		}
		h := NewHRDashboardHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/stats", h.GetStats)
		w := serve(r, jsonReq(http.MethodGet, "/stats", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHRDashboardService{
			GetStatsFunc: func(ctx context.Context) (map[string]interface{}, error) { return nil, errTest },
		}
		h := NewHRDashboardHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/stats", h.GetStats)
		w := serve(r, jsonReq(http.MethodGet, "/stats", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestHRDashboardHandler_GetActivities(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockHRDashboardService{
			GetRecentActivitiesFunc: func(ctx context.Context) ([]map[string]interface{}, error) { return []map[string]interface{}{}, nil },
		}
		h := NewHRDashboardHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/act", h.GetActivities)
		w := serve(r, jsonReq(http.MethodGet, "/act", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockHRDashboardService{
			GetRecentActivitiesFunc: func(ctx context.Context) ([]map[string]interface{}, error) { return nil, errTest },
		}
		h := NewHRDashboardHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/act", h.GetActivities)
		w := serve(r, jsonReq(http.MethodGet, "/act", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// AttendanceIntegrationHandler Tests
// ===================================================================

func TestAttendanceIntegrationHandler_GetIntegration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAttendanceIntegrationService{
			GetIntegrationFunc: func(ctx context.Context, period, dept string) (map[string]interface{}, error) {
				return map[string]interface{}{}, nil
			},
		}
		h := NewAttendanceIntegrationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/int", h.GetIntegration)
		w := serve(r, jsonReq(http.MethodGet, "/int?period=2025-01&department=dev", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAttendanceIntegrationService{
			GetIntegrationFunc: func(ctx context.Context, period, dept string) (map[string]interface{}, error) { return nil, errTest },
		}
		h := NewAttendanceIntegrationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/int", h.GetIntegration)
		w := serve(r, jsonReq(http.MethodGet, "/int", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestAttendanceIntegrationHandler_GetAlerts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAttendanceIntegrationService{
			GetAlertsFunc: func(ctx context.Context) ([]map[string]interface{}, error) { return []map[string]interface{}{}, nil },
		}
		h := NewAttendanceIntegrationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/alerts", h.GetAlerts)
		w := serve(r, jsonReq(http.MethodGet, "/alerts", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAttendanceIntegrationService{
			GetAlertsFunc: func(ctx context.Context) ([]map[string]interface{}, error) { return nil, errTest },
		}
		h := NewAttendanceIntegrationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/alerts", h.GetAlerts)
		w := serve(r, jsonReq(http.MethodGet, "/alerts", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestAttendanceIntegrationHandler_GetTrend(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockAttendanceIntegrationService{
			GetTrendFunc: func(ctx context.Context, period string) ([]map[string]interface{}, error) { return []map[string]interface{}{}, nil },
		}
		h := NewAttendanceIntegrationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/trend", h.GetTrend)
		w := serve(r, jsonReq(http.MethodGet, "/trend?period=monthly", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockAttendanceIntegrationService{
			GetTrendFunc: func(ctx context.Context, period string) ([]map[string]interface{}, error) { return nil, errTest },
		}
		h := NewAttendanceIntegrationHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/trend", h.GetTrend)
		w := serve(r, jsonReq(http.MethodGet, "/trend", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// OrgChartHandler Tests
// ===================================================================

func TestOrgChartHandler_GetOrgChart(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOrgChartService{
			GetOrgChartFunc: func(ctx context.Context) ([]map[string]interface{}, error) { return []map[string]interface{}{}, nil },
		}
		h := NewOrgChartHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/org", h.GetOrgChart)
		w := serve(r, jsonReq(http.MethodGet, "/org", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOrgChartService{
			GetOrgChartFunc: func(ctx context.Context) ([]map[string]interface{}, error) { return nil, errTest },
		}
		h := NewOrgChartHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/org", h.GetOrgChart)
		w := serve(r, jsonReq(http.MethodGet, "/org", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOrgChartHandler_Simulate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOrgChartService{
			SimulateFunc: func(ctx context.Context, data map[string]interface{}) ([]map[string]interface{}, error) {
				return []map[string]interface{}{}, nil
			},
		}
		h := NewOrgChartHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/org/sim", h.Simulate)
		w := serve(r, jsonReq(http.MethodPost, "/org/sim", `{"move":"dept1"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOrgChartService{
			SimulateFunc: func(ctx context.Context, data map[string]interface{}) ([]map[string]interface{}, error) {
				return nil, errTest
			},
		}
		h := NewOrgChartHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/org/sim", h.Simulate)
		w := serve(r, jsonReq(http.MethodPost, "/org/sim", `{}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// suppress unused imports
var (
	_ = (*multipart.Writer)(nil)
	_ = os.Remove
)
