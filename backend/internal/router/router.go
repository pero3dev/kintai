package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	attendanceroutes "github.com/your-org/kintai/backend/internal/apps/attendance_routes"
	expenseroutes "github.com/your-org/kintai/backend/internal/apps/expense_routes"
	hrroutes "github.com/your-org/kintai/backend/internal/apps/hr_routes"
	sharedroutes "github.com/your-org/kintai/backend/internal/apps/shared"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
)

// Setup configures API routes.
func Setup(r *gin.Engine, h *handler.Handlers, mw *middleware.Middleware) {
	r.Use(mw.Recovery())
	r.Use(mw.RequestLogger())
	r.Use(mw.CORS())
	r.Use(mw.SecurityHeaders())
	r.Use(mw.RateLimit())

	r.GET("/health", h.Health.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1")
	sharedroutes.RegisterPublicRoutes(v1, h)

	protected := v1.Group("")
	protected.Use(mw.Auth())

	sharedroutes.RegisterProtectedRoutes(protected, h, mw)
	attendanceroutes.RegisterProtectedRoutes(protected, h, mw)
	expenseroutes.RegisterProtectedRoutes(protected, h)
	hrroutes.RegisterProtectedRoutes(protected, h)
}
