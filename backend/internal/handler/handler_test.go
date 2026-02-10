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
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func getTestLogger() *logger.Logger {
	log, _ := logger.NewLogger("debug", "test")
	return log
}

func TestHealthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodGet, "/health", nil)

	handler.Health(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}

	if response["service"] != "kintai-api" {
		t.Errorf("Expected service 'kintai-api', got '%s'", response["service"])
	}
}

// ===== AuthHandler Tests =====

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := &mocks.MockAuthService{
		LoginFunc: func(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
			return &model.TokenResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    3600,
			}, nil
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/login", handler.Login)

	body := `{"email":"test@example.com","password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := &mocks.MockAuthService{
		LoginFunc: func(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
			return nil, service.ErrInvalidCredentials
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/login", handler.Login)

	body := `{"email":"test@example.com","password":"wrong"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_Login_BadRequest(t *testing.T) {
	mockService := &mocks.MockAuthService{}
	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/login", handler.Login)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := &mocks.MockAuthService{
		RegisterFunc: func(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
			return &model.User{
				BaseModel: model.BaseModel{ID: uuid.New()},
				Email:     req.Email,
				FirstName: req.FirstName,
				LastName:  req.LastName,
			}, nil
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/register", handler.Register)

	body := `{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestAuthHandler_Register_BadRequest(t *testing.T) {
	mockService := &mocks.MockAuthService{}
	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/register", handler.Register)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	mockService := &mocks.MockAuthService{
		RegisterFunc: func(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
			return nil, service.ErrEmailAlreadyExists
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/register", handler.Register)

	body := `{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockService := &mocks.MockAuthService{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
			return &model.TokenResponse{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				ExpiresIn:    3600,
			}, nil
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	body := `{"refresh_token":"old-token"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthHandler_RefreshToken_BadRequest(t *testing.T) {
	mockService := &mocks.MockAuthService{}
	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	mockService := &mocks.MockAuthService{
		RefreshTokenFunc: func(ctx context.Context, refreshToken string) (*model.TokenResponse, error) {
			return nil, service.ErrInvalidCredentials
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	body := `{"refresh_token":"invalid"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockService := &mocks.MockAuthService{
		LogoutFunc: func(ctx context.Context, userID uuid.UUID) error {
			return nil
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/logout", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Logout(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAuthService{}
	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/logout", handler.Logout)

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_Logout_Error(t *testing.T) {
	mockService := &mocks.MockAuthService{
		LogoutFunc: func(ctx context.Context, userID uuid.UUID) error {
			return errors.New("logout error")
		},
	}

	handler := NewAuthHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/auth/logout", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Logout(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== AttendanceHandler Tests =====

func TestAttendanceHandler_ClockIn_Success(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		ClockInFunc: func(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error) {
			now := time.Now()
			return &model.Attendance{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    userID,
				ClockIn:   &now,
				Status:    model.AttendanceStatusPresent,
			}, nil
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/attendance/clock-in", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.ClockIn(c)
	})

	body := `{"note":"出勤"}`
	req, _ := http.NewRequest(http.MethodPost, "/attendance/clock-in", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestAttendanceHandler_ClockIn_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/attendance/clock-in", handler.ClockIn)

	req, _ := http.NewRequest(http.MethodPost, "/attendance/clock-in", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceHandler_ClockIn_AlreadyClockedIn(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		ClockInFunc: func(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error) {
			return nil, service.ErrAlreadyClockedIn
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/attendance/clock-in", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.ClockIn(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/attendance/clock-in", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceHandler_ClockOut_Success(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		ClockOutFunc: func(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error) {
			now := time.Now()
			return &model.Attendance{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    userID,
				ClockOut:  &now,
				Status:    model.AttendanceStatusPresent,
			}, nil
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/attendance/clock-out", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.ClockOut(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/attendance/clock-out", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceHandler_ClockOut_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/attendance/clock-out", handler.ClockOut)

	req, _ := http.NewRequest(http.MethodPost, "/attendance/clock-out", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceHandler_ClockOut_NotClockedIn(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		ClockOutFunc: func(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error) {
			return nil, service.ErrNotClockedIn
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/attendance/clock-out", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.ClockOut(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/attendance/clock-out", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceHandler_GetMyAttendances_Success(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		GetByUserAndDateRangeFunc: func(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
			return []model.Attendance{}, 0, nil
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetMyAttendances(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceHandler_GetMyAttendances_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance", handler.GetMyAttendances)

	req, _ := http.NewRequest(http.MethodGet, "/attendance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceHandler_GetMyAttendances_InvalidDate(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetMyAttendances(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance?start_date=invalid&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceHandler_GetMyAttendances_Error(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		GetByUserAndDateRangeFunc: func(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetMyAttendances(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAttendanceHandler_GetTodayStatus_Success(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		GetTodayStatusFunc: func(ctx context.Context, userID uuid.UUID) (*model.Attendance, error) {
			now := time.Now()
			return &model.Attendance{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    userID,
				ClockIn:   &now,
				Status:    model.AttendanceStatusPresent,
			}, nil
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/today", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetTodayStatus(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance/today", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceHandler_GetTodayStatus_NotClockedIn(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		GetTodayStatusFunc: func(ctx context.Context, userID uuid.UUID) (*model.Attendance, error) {
			return nil, errors.New("not found")
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/today", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetTodayStatus(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance/today", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 with not_clocked_in status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceHandler_GetTodayStatus_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/today", handler.GetTodayStatus)

	req, _ := http.NewRequest(http.MethodGet, "/attendance/today", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceHandler_GetSummary_Success(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		GetSummaryFunc: func(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
			return &model.AttendanceSummary{
				TotalWorkDays:    20,
				TotalWorkMinutes: 9600,
				AbsentDays:       0,
				LeaveDays:        2,
			}, nil
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/summary", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetSummary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance/summary?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAttendanceHandler_GetSummary_Unauthorized(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/summary", handler.GetSummary)

	req, _ := http.NewRequest(http.MethodGet, "/attendance/summary", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAttendanceHandler_GetSummary_InvalidDate(t *testing.T) {
	mockService := &mocks.MockAttendanceService{}
	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/summary", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetSummary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance/summary?start_date=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAttendanceHandler_GetSummary_Error(t *testing.T) {
	mockService := &mocks.MockAttendanceService{
		GetSummaryFunc: func(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewAttendanceHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/attendance/summary", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetSummary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/attendance/summary?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== LeaveHandler Tests =====

func TestLeaveHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockLeaveService{
		CreateFunc: func(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error) {
			return &model.LeaveRequest{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    userID,
				LeaveType: req.LeaveType,
				Status:    model.ApprovalStatusPending,
			}, nil
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leaves", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Create(c)
	})

	body := `{"leave_type":"annual","start_date":"2024-12-25","end_date":"2024-12-26","reason":"休暇"}`
	req, _ := http.NewRequest(http.MethodPost, "/leaves", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestLeaveHandler_Create_Unauthorized(t *testing.T) {
	mockService := &mocks.MockLeaveService{}
	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leaves", handler.Create)

	body := `{"leave_type":"annual","start_date":"2024-12-25","end_date":"2024-12-26","reason":"休暇"}`
	req, _ := http.NewRequest(http.MethodPost, "/leaves", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLeaveHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockLeaveService{}
	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leaves", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Create(c)
	})

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/leaves", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockLeaveService{
		CreateFunc: func(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error) {
			return nil, errors.New("service error")
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/leaves", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Create(c)
	})

	body := `{"leave_type":"annual","start_date":"2024-12-25","end_date":"2024-12-26","reason":"休暇"}`
	req, _ := http.NewRequest(http.MethodPost, "/leaves", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveHandler_Approve_Success(t *testing.T) {
	leaveID := uuid.New()
	mockService := &mocks.MockLeaveService{
		ApproveFunc: func(ctx context.Context, lid uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error) {
			return &model.LeaveRequest{
				BaseModel: model.BaseModel{ID: lid},
				Status:    req.Status,
			}, nil
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leaves/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/leaves/"+leaveID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLeaveHandler_Approve_InvalidID(t *testing.T) {
	mockService := &mocks.MockLeaveService{}
	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leaves/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/leaves/invalid-uuid/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveHandler_Approve_Unauthorized(t *testing.T) {
	leaveID := uuid.New()
	mockService := &mocks.MockLeaveService{}
	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leaves/:id/approve", handler.Approve)

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/leaves/"+leaveID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLeaveHandler_Approve_BadRequest(t *testing.T) {
	leaveID := uuid.New()
	mockService := &mocks.MockLeaveService{}
	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leaves/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPut, "/leaves/"+leaveID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveHandler_Approve_ServiceError(t *testing.T) {
	leaveID := uuid.New()
	mockService := &mocks.MockLeaveService{
		ApproveFunc: func(ctx context.Context, lid uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error) {
			return nil, service.ErrLeaveNotFound
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/leaves/:id/approve", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.Approve(c)
	})

	body := `{"status":"approved"}`
	req, _ := http.NewRequest(http.MethodPut, "/leaves/"+leaveID.String()+"/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLeaveHandler_GetMy_Success(t *testing.T) {
	mockService := &mocks.MockLeaveService{
		GetByUserFunc: func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
			return []model.LeaveRequest{}, 0, nil
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leaves", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/leaves", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLeaveHandler_GetMy_Unauthorized(t *testing.T) {
	mockService := &mocks.MockLeaveService{}
	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leaves", handler.GetMy)

	req, _ := http.NewRequest(http.MethodGet, "/leaves", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLeaveHandler_GetMy_Error(t *testing.T) {
	mockService := &mocks.MockLeaveService{
		GetByUserFunc: func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leaves", func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		handler.GetMy(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/leaves", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestLeaveHandler_GetPending_Success(t *testing.T) {
	mockService := &mocks.MockLeaveService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
			return []model.LeaveRequest{}, 0, nil
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leaves/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/leaves/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestLeaveHandler_GetPending_Error(t *testing.T) {
	mockService := &mocks.MockLeaveService{
		GetPendingFunc: func(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	handler := NewLeaveHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/leaves/pending", handler.GetPending)

	req, _ := http.NewRequest(http.MethodGet, "/leaves/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== ShiftHandler Tests =====

func TestShiftHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockShiftService{
		CreateFunc: func(ctx context.Context, req *model.ShiftCreateRequest) (*model.Shift, error) {
			return &model.Shift{
				BaseModel: model.BaseModel{ID: uuid.New()},
				UserID:    req.UserID,
				ShiftType: req.ShiftType,
			}, nil
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/shifts", handler.Create)

	userID := uuid.New()
	body := `{"user_id":"` + userID.String() + `","date":"2024-12-25","shift_type":"morning"}`
	req, _ := http.NewRequest(http.MethodPost, "/shifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestShiftHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockShiftService{}
	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/shifts", handler.Create)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/shifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShiftHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockShiftService{
		CreateFunc: func(ctx context.Context, req *model.ShiftCreateRequest) (*model.Shift, error) {
			return nil, errors.New("service error")
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/shifts", handler.Create)

	userID := uuid.New()
	body := `{"user_id":"` + userID.String() + `","date":"2024-12-25","shift_type":"morning"}`
	req, _ := http.NewRequest(http.MethodPost, "/shifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShiftHandler_BulkCreate_Success(t *testing.T) {
	mockService := &mocks.MockShiftService{
		BulkCreateFunc: func(ctx context.Context, req *model.ShiftBulkCreateRequest) error {
			return nil
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/shifts/bulk", handler.BulkCreate)

	userID := uuid.New()
	body := `{"shifts":[{"user_id":"` + userID.String() + `","date":"2024-12-25","shift_type":"morning"}]}`
	req, _ := http.NewRequest(http.MethodPost, "/shifts/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestShiftHandler_BulkCreate_BadRequest(t *testing.T) {
	mockService := &mocks.MockShiftService{}
	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/shifts/bulk", handler.BulkCreate)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/shifts/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShiftHandler_BulkCreate_ServiceError(t *testing.T) {
	mockService := &mocks.MockShiftService{
		BulkCreateFunc: func(ctx context.Context, req *model.ShiftBulkCreateRequest) error {
			return errors.New("service error")
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/shifts/bulk", handler.BulkCreate)

	userID := uuid.New()
	body := `{"shifts":[{"user_id":"` + userID.String() + `","date":"2024-12-25","shift_type":"morning"}]}`
	req, _ := http.NewRequest(http.MethodPost, "/shifts/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShiftHandler_GetByDateRange_Success(t *testing.T) {
	mockService := &mocks.MockShiftService{
		GetByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]model.Shift, error) {
			return []model.Shift{}, nil
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/shifts", handler.GetByDateRange)

	req, _ := http.NewRequest(http.MethodGet, "/shifts?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestShiftHandler_GetByDateRange_InvalidDate(t *testing.T) {
	mockService := &mocks.MockShiftService{}
	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/shifts", handler.GetByDateRange)

	req, _ := http.NewRequest(http.MethodGet, "/shifts?start_date=invalid&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShiftHandler_GetByDateRange_Error(t *testing.T) {
	mockService := &mocks.MockShiftService{
		GetByDateRangeFunc: func(ctx context.Context, start, end time.Time) ([]model.Shift, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/shifts", handler.GetByDateRange)

	req, _ := http.NewRequest(http.MethodGet, "/shifts?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestShiftHandler_Delete_Success(t *testing.T) {
	mockService := &mocks.MockShiftService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/shifts/:id", handler.Delete)

	shiftID := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, "/shifts/"+shiftID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestShiftHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockShiftService{}
	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/shifts/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/shifts/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShiftHandler_Delete_Error(t *testing.T) {
	mockService := &mocks.MockShiftService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("delete error")
		},
	}

	handler := NewShiftHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/shifts/:id", handler.Delete)

	shiftID := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, "/shifts/"+shiftID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== UserHandler Tests =====

func TestUserHandler_GetMe_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return &model.User{
				BaseModel: model.BaseModel{ID: id},
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
			}, nil
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/users/me", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMe(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUserHandler_GetMe_Unauthorized(t *testing.T) {
	mockService := &mocks.MockUserService{}
	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/users/me", handler.GetMe)

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUserHandler_GetMe_NotFound(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*model.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/users/me", func(c *gin.Context) {
		c.Set("userID", userID.String())
		handler.GetMe(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUserHandler_GetAll_Success(t *testing.T) {
	mockService := &mocks.MockUserService{
		GetAllFunc: func(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
			return []model.User{}, 0, nil
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/users", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUserHandler_GetAll_Error(t *testing.T) {
	mockService := &mocks.MockUserService{
		GetAllFunc: func(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/users", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUserHandler_Create_Success(t *testing.T) {
	mockService := &mocks.MockUserService{
		CreateFunc: func(ctx context.Context, req *model.UserCreateRequest) (*model.User, error) {
			return &model.User{
				BaseModel: model.BaseModel{ID: uuid.New()},
				Email:     req.Email,
				FirstName: req.FirstName,
				LastName:  req.LastName,
			}, nil
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/users", handler.Create)

	body := `{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User","role":"employee"}`
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestUserHandler_Create_BadRequest(t *testing.T) {
	mockService := &mocks.MockUserService{}
	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/users", handler.Create)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Create_ServiceError(t *testing.T) {
	mockService := &mocks.MockUserService{
		CreateFunc: func(ctx context.Context, req *model.UserCreateRequest) (*model.User, error) {
			return nil, service.ErrEmailAlreadyExists
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.POST("/users", handler.Create)

	body := `{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User","role":"employee"}`
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Update_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.UserUpdateRequest) (*model.User, error) {
			firstName := "Updated"
			return &model.User{
				BaseModel: model.BaseModel{ID: id},
				Email:     "test@example.com",
				FirstName: firstName,
				LastName:  "User",
			}, nil
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/users/:id", handler.Update)

	body := `{"first_name":"Updated"}`
	req, _ := http.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	mockService := &mocks.MockUserService{}
	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/users/:id", handler.Update)

	body := `{"first_name":"Updated"}`
	req, _ := http.NewRequest(http.MethodPut, "/users/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Update_BadRequest(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{}
	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/users/:id", handler.Update)

	body := `invalid json`
	req, _ := http.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Update_ServiceError(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{
		UpdateFunc: func(ctx context.Context, id uuid.UUID, req *model.UserUpdateRequest) (*model.User, error) {
			return nil, service.ErrUserNotFound
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.PUT("/users/:id", handler.Update)

	body := `{"first_name":"Updated"}`
	req, _ := http.NewRequest(http.MethodPut, "/users/"+userID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Delete_Success(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/users/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	mockService := &mocks.MockUserService{}
	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/users/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Delete_Error(t *testing.T) {
	userID := uuid.New()
	mockService := &mocks.MockUserService{
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("delete error")
		},
	}

	handler := NewUserHandler(mockService, getTestLogger())
	router := setupRouter()
	router.DELETE("/users/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== DepartmentHandler Tests =====

func TestDepartmentHandler_GetAll_Success(t *testing.T) {
	mockService := &mocks.MockDepartmentService{
		GetAllFunc: func(ctx context.Context) ([]model.Department, error) {
			return []model.Department{
				{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "Engineering"},
			}, nil
		},
	}

	handler := NewDepartmentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/departments", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/departments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDepartmentHandler_GetAll_Error(t *testing.T) {
	mockService := &mocks.MockDepartmentService{
		GetAllFunc: func(ctx context.Context) ([]model.Department, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewDepartmentHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/departments", handler.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/departments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== DashboardHandler Tests =====

func TestDashboardHandler_GetStats_Success(t *testing.T) {
	mockService := &mocks.MockDashboardService{
		GetStatsFunc: func(ctx context.Context) (*model.DashboardStatsExtended, error) {
			return &model.DashboardStatsExtended{
				DashboardStats: model.DashboardStats{
					TodayPresentCount: 10,
					TodayAbsentCount:  10,
					PendingLeaves:     5,
				},
			}, nil
		},
	}

	handler := NewDashboardHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/dashboard/stats", handler.GetStats)

	req, _ := http.NewRequest(http.MethodGet, "/dashboard/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDashboardHandler_GetStats_Error(t *testing.T) {
	mockService := &mocks.MockDashboardService{
		GetStatsFunc: func(ctx context.Context) (*model.DashboardStatsExtended, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewDashboardHandler(mockService, getTestLogger())
	router := setupRouter()
	router.GET("/dashboard/stats", handler.GetStats)

	req, _ := http.NewRequest(http.MethodGet, "/dashboard/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ===== Helper Function Tests =====

func TestParsePagination(t *testing.T) {
	router := setupRouter()
	router.GET("/test", func(c *gin.Context) {
		page, pageSize := parsePagination(c)
		c.JSON(http.StatusOK, gin.H{"page": page, "page_size": pageSize})
	})

	// Default values
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]int
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["page"] != 1 {
		t.Errorf("Expected page 1, got %d", resp["page"])
	}
	if resp["page_size"] != 20 {
		t.Errorf("Expected page_size 20, got %d", resp["page_size"])
	}

	// Custom values
	req, _ = http.NewRequest(http.MethodGet, "/test?page=2&page_size=50", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["page"] != 2 {
		t.Errorf("Expected page 2, got %d", resp["page"])
	}
	if resp["page_size"] != 50 {
		t.Errorf("Expected page_size 50, got %d", resp["page_size"])
	}

	// Invalid values (too high page size)
	req, _ = http.NewRequest(http.MethodGet, "/test?page=-1&page_size=200", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["page"] != 1 {
		t.Errorf("Expected page 1, got %d", resp["page"])
	}
	if resp["page_size"] != 20 {
		t.Errorf("Expected page_size 20, got %d", resp["page_size"])
	}
}

func TestNewHandlers(t *testing.T) {
	services := &service.Services{
		Auth:       &mocks.MockAuthService{},
		Attendance: &mocks.MockAttendanceService{},
		Leave:      &mocks.MockLeaveService{},
		Shift:      &mocks.MockShiftService{},
		User:       &mocks.MockUserService{},
		Department: &mocks.MockDepartmentService{},
		Dashboard:  &mocks.MockDashboardService{},
	}

	handlers := NewHandlers(services, getTestLogger())

	if handlers.Auth == nil {
		t.Error("AuthHandler should not be nil")
	}
	if handlers.Attendance == nil {
		t.Error("AttendanceHandler should not be nil")
	}
	if handlers.Leave == nil {
		t.Error("LeaveHandler should not be nil")
	}
	if handlers.Shift == nil {
		t.Error("ShiftHandler should not be nil")
	}
	if handlers.User == nil {
		t.Error("UserHandler should not be nil")
	}
	if handlers.Department == nil {
		t.Error("DepartmentHandler should not be nil")
	}
	if handlers.Dashboard == nil {
		t.Error("DashboardHandler should not be nil")
	}
	if handlers.Health == nil {
		t.Error("HealthHandler should not be nil")
	}
}
