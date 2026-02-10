package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
)

// ===================================================================
// OneOnOneHandler Tests
// ===================================================================

func TestOneOnOneHandler_Create(t *testing.T) {
	userID := uuid.New()
	body := `{"employee_id":"` + uuid.New().String() + `","scheduled_date":"2026-03-01"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			CreateFunc: func(ctx context.Context, req model.OneOnOneCreateRequest, mid uuid.UUID) (*model.OneOnOneMeeting, error) {
				return &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/oo", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/oo", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.POST("/oo", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/oo", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.POST("/oo", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/oo", body))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			CreateFunc: func(ctx context.Context, req model.OneOnOneCreateRequest, mid uuid.UUID) (*model.OneOnOneMeeting, error) {
				return nil, errTest
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/oo", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/oo", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOneOnOneHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.OneOnOneMeeting, error) {
				return &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/oo/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/oo/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.GET("/oo/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/oo/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.OneOnOneMeeting, error) { return nil, errTest },
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/oo/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/oo/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestOneOnOneHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			FindAllFunc: func(ctx context.Context, status, empID string) ([]model.OneOnOneMeeting, error) {
				return []model.OneOnOneMeeting{}, nil
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/oo", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/oo?status=scheduled&employee_id=abc", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			FindAllFunc: func(ctx context.Context, status, empID string) ([]model.OneOnOneMeeting, error) { return nil, errTest },
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/oo", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/oo", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOneOnOneHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.OneOnOneUpdateRequest) (*model.OneOnOneMeeting, error) {
				return &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/oo/"+id.String(), `{"status":"completed"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/oo/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/oo/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.OneOnOneUpdateRequest) (*model.OneOnOneMeeting, error) {
				return nil, errTest
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/oo/"+id.String(), `{"status":"completed"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOneOnOneHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil }}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/oo/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/oo/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/oo/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/oo/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest }}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/oo/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/oo/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOneOnOneHandler_AddActionItem(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			AddActionItemFunc: func(ctx context.Context, mid uuid.UUID, req model.ActionItemRequest) (*model.OneOnOneMeeting, error) {
				return &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: mid}}, nil
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/oo/:id/action", h.AddActionItem)
		w := serve(r, jsonReq(http.MethodPost, "/oo/"+id.String()+"/action", `{"title":"タスク1"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.POST("/oo/:id/action", h.AddActionItem)
		w := serve(r, jsonReq(http.MethodPost, "/oo/invalid/action", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.POST("/oo/:id/action", h.AddActionItem)
		w := serve(r, jsonReq(http.MethodPost, "/oo/"+id.String()+"/action", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			AddActionItemFunc: func(ctx context.Context, mid uuid.UUID, req model.ActionItemRequest) (*model.OneOnOneMeeting, error) {
				return nil, errTest
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/oo/:id/action", h.AddActionItem)
		w := serve(r, jsonReq(http.MethodPost, "/oo/"+id.String()+"/action", `{"title":"タスク1"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOneOnOneHandler_ToggleActionItem(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			ToggleActionItemFunc: func(ctx context.Context, mid uuid.UUID, actionID string) (*model.OneOnOneMeeting, error) {
				return &model.OneOnOneMeeting{BaseModel: model.BaseModel{ID: mid}}, nil
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id/action/:actionId", h.ToggleActionItem)
		w := serve(r, jsonReq(http.MethodPut, "/oo/"+id.String()+"/action/act1", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOneOnOneHandler(&mocks.MockOneOnOneService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id/action/:actionId", h.ToggleActionItem)
		w := serve(r, jsonReq(http.MethodPut, "/oo/invalid/action/act1", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOneOnOneService{
			ToggleActionItemFunc: func(ctx context.Context, mid uuid.UUID, actionID string) (*model.OneOnOneMeeting, error) {
				return nil, errTest
			},
		}
		h := NewOneOnOneHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/oo/:id/action/:actionId", h.ToggleActionItem)
		w := serve(r, jsonReq(http.MethodPut, "/oo/"+id.String()+"/action/act1", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// SkillHandler Tests
// ===================================================================

func TestSkillHandler_GetSkillMap(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			GetSkillMapFunc: func(ctx context.Context, dept, empID string) ([]model.EmployeeSkill, error) {
				return []model.EmployeeSkill{}, nil
			},
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sk", h.GetSkillMap)
		w := serve(r, jsonReq(http.MethodGet, "/sk?department=dev&employee_id=abc", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			GetSkillMapFunc: func(ctx context.Context, dept, empID string) ([]model.EmployeeSkill, error) { return nil, errTest },
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sk", h.GetSkillMap)
		w := serve(r, jsonReq(http.MethodGet, "/sk", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSkillHandler_GetGapAnalysis(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			GetGapAnalysisFunc: func(ctx context.Context, dept string) ([]map[string]interface{}, error) {
				return []map[string]interface{}{}, nil
			},
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sk/gap", h.GetGapAnalysis)
		w := serve(r, jsonReq(http.MethodGet, "/sk/gap?department=dev", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			GetGapAnalysisFunc: func(ctx context.Context, dept string) ([]map[string]interface{}, error) { return nil, errTest },
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sk/gap", h.GetGapAnalysis)
		w := serve(r, jsonReq(http.MethodGet, "/sk/gap", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSkillHandler_AddSkill(t *testing.T) {
	empID := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			AddSkillFunc: func(ctx context.Context, eid uuid.UUID, req model.SkillAddRequest) (*model.EmployeeSkill, error) {
				return &model.EmployeeSkill{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sk/:employeeId", h.AddSkill)
		w := serve(r, jsonReq(http.MethodPost, "/sk/"+empID.String(), `{"skill_name":"Go","level":3}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSkillHandler(&mocks.MockSkillService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sk/:employeeId", h.AddSkill)
		w := serve(r, jsonReq(http.MethodPost, "/sk/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewSkillHandler(&mocks.MockSkillService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sk/:employeeId", h.AddSkill)
		w := serve(r, jsonReq(http.MethodPost, "/sk/"+empID.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			AddSkillFunc: func(ctx context.Context, eid uuid.UUID, req model.SkillAddRequest) (*model.EmployeeSkill, error) {
				return nil, errTest
			},
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sk/:employeeId", h.AddSkill)
		w := serve(r, jsonReq(http.MethodPost, "/sk/"+empID.String(), `{"skill_name":"Go","level":3}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSkillHandler_UpdateSkill(t *testing.T) {
	skillID := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			UpdateSkillFunc: func(ctx context.Context, sid uuid.UUID, req model.SkillUpdateRequest) (*model.EmployeeSkill, error) {
				return &model.EmployeeSkill{BaseModel: model.BaseModel{ID: sid}}, nil
			},
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/sk/:skillId", h.UpdateSkill)
		w := serve(r, jsonReq(http.MethodPut, "/sk/"+skillID.String(), `{"level":5}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSkillHandler(&mocks.MockSkillService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/sk/:skillId", h.UpdateSkill)
		w := serve(r, jsonReq(http.MethodPut, "/sk/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewSkillHandler(&mocks.MockSkillService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/sk/:skillId", h.UpdateSkill)
		w := serve(r, jsonReq(http.MethodPut, "/sk/"+skillID.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSkillService{
			UpdateSkillFunc: func(ctx context.Context, sid uuid.UUID, req model.SkillUpdateRequest) (*model.EmployeeSkill, error) {
				return nil, errTest
			},
		}
		h := NewSkillHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/sk/:skillId", h.UpdateSkill)
		w := serve(r, jsonReq(http.MethodPut, "/sk/"+skillID.String(), `{"level":5}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// SalaryHandler Tests
// ===================================================================

func TestSalaryHandler_GetOverview(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			GetOverviewFunc: func(ctx context.Context, dept string) (map[string]interface{}, error) { return map[string]interface{}{}, nil },
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sal", h.GetOverview)
		w := serve(r, jsonReq(http.MethodGet, "/sal?department=dev", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			GetOverviewFunc: func(ctx context.Context, dept string) (map[string]interface{}, error) { return nil, errTest },
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sal", h.GetOverview)
		w := serve(r, jsonReq(http.MethodGet, "/sal", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSalaryHandler_Simulate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			SimulateFunc: func(ctx context.Context, req model.SalarySimulateRequest) (map[string]interface{}, error) {
				return map[string]interface{}{}, nil
			},
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sal/sim", h.Simulate)
		w := serve(r, jsonReq(http.MethodPost, "/sal/sim", `{"grade":"M1"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewSalaryHandler(&mocks.MockSalaryService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sal/sim", h.Simulate)
		w := serve(r, jsonReq(http.MethodPost, "/sal/sim", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			SimulateFunc: func(ctx context.Context, req model.SalarySimulateRequest) (map[string]interface{}, error) {
				return nil, errTest
			},
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sal/sim", h.Simulate)
		w := serve(r, jsonReq(http.MethodPost, "/sal/sim", `{"grade":"M1"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSalaryHandler_GetHistory(t *testing.T) {
	empID := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			GetHistoryFunc: func(ctx context.Context, eid uuid.UUID) ([]model.SalaryRecord, error) {
				return []model.SalaryRecord{}, nil
			},
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sal/:employeeId/history", h.GetHistory)
		w := serve(r, jsonReq(http.MethodGet, "/sal/"+empID.String()+"/history", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSalaryHandler(&mocks.MockSalaryService{}, getTestLogger())
		r := setupRouter()
		r.GET("/sal/:employeeId/history", h.GetHistory)
		w := serve(r, jsonReq(http.MethodGet, "/sal/invalid/history", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			GetHistoryFunc: func(ctx context.Context, eid uuid.UUID) ([]model.SalaryRecord, error) { return nil, errTest },
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sal/:employeeId/history", h.GetHistory)
		w := serve(r, jsonReq(http.MethodGet, "/sal/"+empID.String()+"/history", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSalaryHandler_GetBudget(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			GetBudgetFunc: func(ctx context.Context, dept string) (map[string]interface{}, error) { return map[string]interface{}{}, nil },
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sal/budget", h.GetBudget)
		w := serve(r, jsonReq(http.MethodGet, "/sal/budget?department=dev", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSalaryService{
			GetBudgetFunc: func(ctx context.Context, dept string) (map[string]interface{}, error) { return nil, errTest },
		}
		h := NewSalaryHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sal/budget", h.GetBudget)
		w := serve(r, jsonReq(http.MethodGet, "/sal/budget", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// OnboardingHandler Tests
// ===================================================================

func TestOnboardingHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			CreateFunc: func(ctx context.Context, req model.OnboardingCreateRequest) (*model.Onboarding, error) {
				return &model.Onboarding{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ob", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/ob", `{"employee_id":"`+uuid.New().String()+`","start_date":"2026-03-01"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOnboardingHandler(&mocks.MockOnboardingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ob", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/ob", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			CreateFunc: func(ctx context.Context, req model.OnboardingCreateRequest) (*model.Onboarding, error) { return nil, errTest },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ob", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/ob", `{"employee_id":"`+uuid.New().String()+`","start_date":"2026-03-01"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOnboardingHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Onboarding, error) {
				return &model.Onboarding{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ob/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ob/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOnboardingHandler(&mocks.MockOnboardingService{}, getTestLogger())
		r := setupRouter()
		r.GET("/ob/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ob/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Onboarding, error) { return nil, errTest },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ob/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/ob/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestOnboardingHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			FindAllFunc: func(ctx context.Context, status string) ([]model.Onboarding, error) { return []model.Onboarding{}, nil },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ob", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/ob?status=pending", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			FindAllFunc: func(ctx context.Context, status string) ([]model.Onboarding, error) { return nil, errTest },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ob", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/ob", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOnboardingHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, data map[string]interface{}) (*model.Onboarding, error) {
				return &model.Onboarding{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ob/"+id.String(), `{"status":"in_progress"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOnboardingHandler(&mocks.MockOnboardingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ob/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOnboardingHandler(&mocks.MockOnboardingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ob/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, data map[string]interface{}) (*model.Onboarding, error) {
				return nil, errTest
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/ob/"+id.String(), `{"status":"in_progress"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOnboardingHandler_ToggleTask(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			ToggleTaskFunc: func(ctx context.Context, i uuid.UUID, taskID string) (*model.Onboarding, error) {
				return &model.Onboarding{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id/task/:taskId", h.ToggleTask)
		w := serve(r, jsonReq(http.MethodPut, "/ob/"+id.String()+"/task/t1", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOnboardingHandler(&mocks.MockOnboardingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id/task/:taskId", h.ToggleTask)
		w := serve(r, jsonReq(http.MethodPut, "/ob/invalid/task/t1", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			ToggleTaskFunc: func(ctx context.Context, i uuid.UUID, taskID string) (*model.Onboarding, error) { return nil, errTest },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/ob/:id/task/:taskId", h.ToggleTask)
		w := serve(r, jsonReq(http.MethodPut, "/ob/"+id.String()+"/task/t1", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOnboardingHandler_GetTemplates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			FindAllTemplatesFunc: func(ctx context.Context) ([]model.OnboardingTemplate, error) { return []model.OnboardingTemplate{}, nil },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ob/tpl", h.GetTemplates)
		w := serve(r, jsonReq(http.MethodGet, "/ob/tpl", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			FindAllTemplatesFunc: func(ctx context.Context) ([]model.OnboardingTemplate, error) { return nil, errTest },
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/ob/tpl", h.GetTemplates)
		w := serve(r, jsonReq(http.MethodGet, "/ob/tpl", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOnboardingHandler_CreateTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			CreateTemplateFunc: func(ctx context.Context, req model.OnboardingTemplateCreateRequest) (*model.OnboardingTemplate, error) {
				return &model.OnboardingTemplate{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ob/tpl", h.CreateTemplate)
		w := serve(r, jsonReq(http.MethodPost, "/ob/tpl", `{"name":"テンプレート1"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOnboardingHandler(&mocks.MockOnboardingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/ob/tpl", h.CreateTemplate)
		w := serve(r, jsonReq(http.MethodPost, "/ob/tpl", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOnboardingService{
			CreateTemplateFunc: func(ctx context.Context, req model.OnboardingTemplateCreateRequest) (*model.OnboardingTemplate, error) {
				return nil, errTest
			},
		}
		h := NewOnboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/ob/tpl", h.CreateTemplate)
		w := serve(r, jsonReq(http.MethodPost, "/ob/tpl", `{"name":"テンプレート1"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// OffboardingHandler Tests
// ===================================================================

func TestOffboardingHandler_Create(t *testing.T) {
	body := `{"employee_id":"` + uuid.New().String() + `","last_working_date":"2026-03-31"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			CreateFunc: func(ctx context.Context, req model.OffboardingCreateRequest) (*model.Offboarding, error) {
				return &model.Offboarding{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/off", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/off", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOffboardingHandler(&mocks.MockOffboardingService{}, getTestLogger())
		r := setupRouter()
		r.POST("/off", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/off", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			CreateFunc: func(ctx context.Context, req model.OffboardingCreateRequest) (*model.Offboarding, error) { return nil, errTest },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/off", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/off", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOffboardingHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Offboarding, error) {
				return &model.Offboarding{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/off/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/off/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOffboardingHandler(&mocks.MockOffboardingService{}, getTestLogger())
		r := setupRouter()
		r.GET("/off/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/off/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Offboarding, error) { return nil, errTest },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/off/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/off/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestOffboardingHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			FindAllFunc: func(ctx context.Context, status string) ([]model.Offboarding, error) { return []model.Offboarding{}, nil },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/off", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/off?status=pending", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			FindAllFunc: func(ctx context.Context, status string) ([]model.Offboarding, error) { return nil, errTest },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/off", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/off", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOffboardingHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.OffboardingUpdateRequest) (*model.Offboarding, error) {
				return &model.Offboarding{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/off/"+id.String(), `{"status":"completed"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOffboardingHandler(&mocks.MockOffboardingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/off/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewOffboardingHandler(&mocks.MockOffboardingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/off/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.OffboardingUpdateRequest) (*model.Offboarding, error) {
				return nil, errTest
			},
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/off/"+id.String(), `{"status":"completed"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOffboardingHandler_ToggleChecklist(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			ToggleChecklistFunc: func(ctx context.Context, i uuid.UUID, key string) (*model.Offboarding, error) {
				return &model.Offboarding{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id/check/:itemKey", h.ToggleChecklist)
		w := serve(r, jsonReq(http.MethodPut, "/off/"+id.String()+"/check/key1", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewOffboardingHandler(&mocks.MockOffboardingService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id/check/:itemKey", h.ToggleChecklist)
		w := serve(r, jsonReq(http.MethodPut, "/off/invalid/check/key1", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			ToggleChecklistFunc: func(ctx context.Context, i uuid.UUID, key string) (*model.Offboarding, error) { return nil, errTest },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/off/:id/check/:itemKey", h.ToggleChecklist)
		w := serve(r, jsonReq(http.MethodPut, "/off/"+id.String()+"/check/key1", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestOffboardingHandler_GetAnalytics(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			GetAnalyticsFunc: func(ctx context.Context) (map[string]interface{}, error) { return map[string]interface{}{}, nil },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/off/analytics", h.GetAnalytics)
		w := serve(r, jsonReq(http.MethodGet, "/off/analytics", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockOffboardingService{
			GetAnalyticsFunc: func(ctx context.Context) (map[string]interface{}, error) { return nil, errTest },
		}
		h := NewOffboardingHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/off/analytics", h.GetAnalytics)
		w := serve(r, jsonReq(http.MethodGet, "/off/analytics", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

// ===================================================================
// SurveyHandler Tests
// ===================================================================

func TestSurveyHandler_Create(t *testing.T) {
	userID := uuid.New()
	body := `{"title":"サーベイ1"}`
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			CreateFunc: func(ctx context.Context, req model.SurveyCreateRequest, uid uuid.UUID) (*model.Survey, error) {
				return &model.Survey{BaseModel: model.BaseModel{ID: uuid.New()}}, nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/sv", body))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sv", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/sv", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("Unauthorized", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sv", h.Create)
		w := serve(r, jsonReq(http.MethodPost, "/sv", body))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			CreateFunc: func(ctx context.Context, req model.SurveyCreateRequest, uid uuid.UUID) (*model.Survey, error) {
				return nil, errTest
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv", withUser(userID, h.Create))
		w := serve(r, jsonReq(http.MethodPost, "/sv", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_GetByID(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Survey, error) {
				return &model.Survey{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sv/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/sv/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.GET("/sv/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/sv/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			FindByIDFunc: func(ctx context.Context, i uuid.UUID) (*model.Survey, error) { return nil, errTest },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sv/:id", h.GetByID)
		w := serve(r, jsonReq(http.MethodGet, "/sv/"+id.String(), ""))
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			FindAllFunc: func(ctx context.Context, status, sType string) ([]model.Survey, error) { return []model.Survey{}, nil },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sv", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/sv?status=active&type=engagement", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			FindAllFunc: func(ctx context.Context, status, sType string) ([]model.Survey, error) { return nil, errTest },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sv", h.GetAll)
		w := serve(r, jsonReq(http.MethodGet, "/sv", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_Update(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.SurveyUpdateRequest) (*model.Survey, error) {
				return &model.Survey{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/sv/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/sv/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/sv/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/sv/invalid", `{}`))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.PUT("/sv/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/sv/"+id.String(), "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			UpdateFunc: func(ctx context.Context, i uuid.UUID, req model.SurveyUpdateRequest) (*model.Survey, error) { return nil, errTest },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.PUT("/sv/:id", h.Update)
		w := serve(r, jsonReq(http.MethodPut, "/sv/"+id.String(), `{"title":"updated"}`))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_Delete(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return nil }}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/sv/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/sv/"+id.String(), ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.DELETE("/sv/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/sv/invalid", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{DeleteFunc: func(ctx context.Context, i uuid.UUID) error { return errTest }}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.DELETE("/sv/:id", h.Delete)
		w := serve(r, jsonReq(http.MethodDelete, "/sv/"+id.String(), ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_Publish(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			PublishFunc: func(ctx context.Context, i uuid.UUID) (*model.Survey, error) {
				return &model.Survey{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/publish", h.Publish)
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/publish", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/publish", h.Publish)
		w := serve(r, jsonReq(http.MethodPost, "/sv/invalid/publish", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			PublishFunc: func(ctx context.Context, i uuid.UUID) (*model.Survey, error) { return nil, errTest },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/publish", h.Publish)
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/publish", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_Close(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			CloseFunc: func(ctx context.Context, i uuid.UUID) (*model.Survey, error) {
				return &model.Survey{BaseModel: model.BaseModel{ID: i}}, nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/close", h.Close)
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/close", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/close", h.Close)
		w := serve(r, jsonReq(http.MethodPost, "/sv/invalid/close", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			CloseFunc: func(ctx context.Context, i uuid.UUID) (*model.Survey, error) { return nil, errTest },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/close", h.Close)
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/close", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_GetResults(t *testing.T) {
	id := uuid.New()
	t.Run("Success", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			GetResultsFunc: func(ctx context.Context, i uuid.UUID) (map[string]interface{}, error) { return map[string]interface{}{}, nil },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sv/:id/results", h.GetResults)
		w := serve(r, jsonReq(http.MethodGet, "/sv/"+id.String()+"/results", ""))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.GET("/sv/:id/results", h.GetResults)
		w := serve(r, jsonReq(http.MethodGet, "/sv/invalid/results", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			GetResultsFunc: func(ctx context.Context, i uuid.UUID) (map[string]interface{}, error) { return nil, errTest },
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.GET("/sv/:id/results", h.GetResults)
		w := serve(r, jsonReq(http.MethodGet, "/sv/"+id.String()+"/results", ""))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestSurveyHandler_SubmitResponse(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	body := `{"answers":{"q1":"yes"}}`
	t.Run("Success_WithUser", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			SubmitResponseFunc: func(ctx context.Context, sid uuid.UUID, empID *uuid.UUID, req model.SurveyResponseRequest) error {
				if empID == nil {
					t.Error("expected empID to be non-nil")
				}
				return nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/respond", withUser(userID, h.SubmitResponse))
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/respond", body))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("Success_WithoutUser", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			SubmitResponseFunc: func(ctx context.Context, sid uuid.UUID, empID *uuid.UUID, req model.SurveyResponseRequest) error {
				if empID != nil {
					t.Error("expected empID to be nil")
				}
				return nil
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/respond", h.SubmitResponse)
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/respond", body))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/respond", h.SubmitResponse)
		w := serve(r, jsonReq(http.MethodPost, "/sv/invalid/respond", body))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("BindError", func(t *testing.T) {
		h := NewSurveyHandler(&mocks.MockSurveyService{}, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/respond", h.SubmitResponse)
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/respond", "invalid"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
	t.Run("ServiceError", func(t *testing.T) {
		mock := &mocks.MockSurveyService{
			SubmitResponseFunc: func(ctx context.Context, sid uuid.UUID, empID *uuid.UUID, req model.SurveyResponseRequest) error {
				return errTest
			},
		}
		h := NewSurveyHandler(mock, getTestLogger())
		r := setupRouter()
		r.POST("/sv/:id/respond", withUser(userID, h.SubmitResponse))
		w := serve(r, jsonReq(http.MethodPost, "/sv/"+id.String()+"/respond", body))
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}
