package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/service"
)

// ===================================================================
// getUserIDFromContext Tests
// ===================================================================

func TestGetUserIDFromContext_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expected := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", expected.String())

	got, err := getUserIDFromContext(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestGetUserIDFromContext_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// userID をセットしない

	_, err := getUserIDFromContext(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != service.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGetUserIDFromContext_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "not-a-valid-uuid")

	_, err := getUserIDFromContext(c)
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
}

// ===================================================================
// parseUUID Tests
// ===================================================================

func TestParseUUID_ValidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expected := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: expected.String()}}

	got, err := parseUUID(c, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestParseUUID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	got, err := parseUUID(c, "id")
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
	if got != uuid.Nil {
		t.Errorf("expected uuid.Nil, got %s", got)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected error code %d, got %d", http.StatusBadRequest, resp.Code)
	}
	if resp.Message != "無効なIDフォーマットです" {
		t.Errorf("unexpected error message: %s", resp.Message)
	}
}

func TestParseUUID_EmptyParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: ""}}

	_, err := parseUUID(c, "id")
	if err == nil {
		t.Fatal("expected error for empty param, got nil")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestParseUUID_MissingParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// パラメータなし

	_, err := parseUUID(c, "id")
	if err == nil {
		t.Fatal("expected error for missing param, got nil")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// ===================================================================
// parseDateRange Tests
// ===================================================================

func TestParseDateRange_BothDatesProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?start_date=2025-01-01&end_date=2025-12-31", nil)

	start, end, err := parseDateRange(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if start.Format("2006-01-02") != "2025-01-01" {
		t.Errorf("expected start 2025-01-01, got %s", start.Format("2006-01-02"))
	}
	if end.Format("2006-01-02") != "2025-12-31" {
		t.Errorf("expected end 2025-12-31, got %s", end.Format("2006-01-02"))
	}
}

func TestParseDateRange_DefaultDates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	start, end, err := parseDateRange(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// デフォルトは1ヶ月前〜今日
	if start.IsZero() {
		t.Error("start should not be zero")
	}
	if end.IsZero() {
		t.Error("end should not be zero")
	}
	if !start.Before(end) {
		t.Errorf("start (%s) should be before end (%s)", start, end)
	}
}

func TestParseDateRange_InvalidStartDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?start_date=invalid&end_date=2025-12-31", nil)

	_, _, err := parseDateRange(c)
	if err == nil {
		t.Fatal("expected error for invalid start_date, got nil")
	}
}

func TestParseDateRange_InvalidEndDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?start_date=2025-01-01&end_date=not-a-date", nil)

	_, _, err := parseDateRange(c)
	if err == nil {
		t.Fatal("expected error for invalid end_date, got nil")
	}
}

func TestParseDateRange_OnlyStartDateProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?start_date=2025-06-15", nil)

	start, end, err := parseDateRange(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if start.Format("2006-01-02") != "2025-06-15" {
		t.Errorf("expected start 2025-06-15, got %s", start.Format("2006-01-02"))
	}
	// end_date はデフォルト（今日）
	if end.IsZero() {
		t.Error("end should not be zero")
	}
}

func TestParseDateRange_OnlyEndDateProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?end_date=2025-12-31", nil)

	start, end, err := parseDateRange(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// start_date はデフォルト（1ヶ月前）
	if start.IsZero() {
		t.Error("start should not be zero")
	}
	if end.Format("2006-01-02") != "2025-12-31" {
		t.Errorf("expected end 2025-12-31, got %s", end.Format("2006-01-02"))
	}
}

func TestParseDateRange_InvalidStartValidEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/?start_date=2025/01/01&end_date=2025-12-31", nil)

	_, _, err := parseDateRange(c)
	if err == nil {
		t.Fatal("expected error for invalid start_date format, got nil")
	}
}

// ===================================================================
// paginatedResponse Tests
// ===================================================================

func TestPaginatedResponse_ExactDivision(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	data := []string{"a", "b", "c"}
	paginatedResponse(c, data, 60, 1, 20)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Total != 60 {
		t.Errorf("expected total 60, got %d", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Page)
	}
	if resp.PageSize != 20 {
		t.Errorf("expected page_size 20, got %d", resp.PageSize)
	}
	if resp.TotalPages != 3 {
		t.Errorf("expected total_pages 3, got %d", resp.TotalPages)
	}
}

func TestPaginatedResponse_WithRemainder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	data := []string{"a"}
	paginatedResponse(c, data, 25, 2, 10)

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// 25 / 10 = 2 余り 5 → totalPages = 3
	if resp.TotalPages != 3 {
		t.Errorf("expected total_pages 3, got %d", resp.TotalPages)
	}
	if resp.Page != 2 {
		t.Errorf("expected page 2, got %d", resp.Page)
	}
}

func TestPaginatedResponse_ZeroTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	data := []string{}
	paginatedResponse(c, data, 0, 1, 20)

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.TotalPages != 0 {
		t.Errorf("expected total_pages 0, got %d", resp.TotalPages)
	}
	if resp.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Total)
	}
}

func TestPaginatedResponse_SingleItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	data := []string{"only"}
	paginatedResponse(c, data, 1, 1, 10)

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// 1 / 10 = 0 余り 1 → totalPages = 1
	if resp.TotalPages != 1 {
		t.Errorf("expected total_pages 1, got %d", resp.TotalPages)
	}
}

func TestPaginatedResponse_PageSizeEqualsTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	data := []int{1, 2, 3, 4, 5}
	paginatedResponse(c, data, 5, 1, 5)

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// 5 / 5 = 1 余り 0 → totalPages = 1
	if resp.TotalPages != 1 {
		t.Errorf("expected total_pages 1, got %d", resp.TotalPages)
	}
}

func TestPaginatedResponse_NilData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	paginatedResponse(c, nil, 0, 1, 20)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Data != nil {
		t.Errorf("expected nil data, got %v", resp.Data)
	}
}

func TestPaginatedResponse_LargeDataset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

	data := []string{"item"}
	paginatedResponse(c, data, 1001, 50, 20)

	var resp model.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	// 1001 / 20 = 50 余り 1 → totalPages = 51
	if resp.TotalPages != 51 {
		t.Errorf("expected total_pages 51, got %d", resp.TotalPages)
	}
	if resp.Page != 50 {
		t.Errorf("expected page 50, got %d", resp.Page)
	}
}
