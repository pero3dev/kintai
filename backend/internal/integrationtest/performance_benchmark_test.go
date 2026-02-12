package integrationtest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type benchmarkTarget struct {
	name           string
	method         string
	path           string
	headers        map[string]string
	expectedStatus int
}

func BenchmarkAPIEndpoints(b *testing.B) {
	env := NewTestEnv(b, &Options{
		RateLimitRPS:                 10000000,
		RateLimitBurst:               10000000,
		DisableAdditionalAutoMigrate: true,
	})
	if err := env.ResetDB(); err != nil {
		b.Fatalf("reset db: %v", err)
	}

	email := fmt.Sprintf("it-bench-%s@example.com", uuid.NewString())
	user := createBenchmarkUser(b, env, model.RoleEmployee, email, "password123")
	employeeHeaders := map[string]string{
		"Authorization": env.MustBearerToken(b, user.ID, model.RoleEmployee),
	}

	seedAttendanceRowsInBatches(
		b,
		env,
		user.ID,
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
		1500,
	)

	targets := []benchmarkTarget{
		{
			name:           "health",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "users_me",
			method:         http.MethodGet,
			path:           "/api/v1/users/me",
			headers:        employeeHeaders,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "attendance_list_large_dataset",
			method:         http.MethodGet,
			path:           "/api/v1/attendance?start_date=2020-01-01&end_date=2030-12-31&page=1&page_size=100",
			headers:        employeeHeaders,
			expectedStatus: http.StatusOK,
		},
	}

	for _, target := range targets {
		target := target
		b.Run(target.name, func(b *testing.B) {
			for i := 0; i < 10; i++ {
				resp := env.DoJSON(b, target.method, target.path, nil, target.headers)
				if resp.Code != target.expectedStatus {
					b.Fatalf("warmup status mismatch: got=%d want=%d", resp.Code, target.expectedStatus)
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp := env.DoJSON(b, target.method, target.path, nil, target.headers)
				if resp.Code != target.expectedStatus {
					b.Fatalf("status mismatch: got=%d want=%d", resp.Code, target.expectedStatus)
				}
			}
		})
	}
}

func createBenchmarkUser(tb testing.TB, env *TestEnv, role model.Role, email, password string) *model.User {
	tb.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		tb.Fatalf("hash password: %v", err)
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    "Integration",
		LastName:     "Benchmark",
		Role:         role,
		IsActive:     true,
	}
	if err := env.DB.Create(user).Error; err != nil {
		tb.Fatalf("create user: %v", err)
	}

	return user
}
