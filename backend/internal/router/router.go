package router

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Setup はルーターを設定する
func Setup(r *gin.Engine, h *handler.Handlers, mw *middleware.Middleware) {
	// グローバルミドルウェア
	r.Use(mw.Recovery())
	r.Use(mw.RequestLogger())
	r.Use(mw.CORS())
	r.Use(mw.SecurityHeaders())
	r.Use(mw.RateLimit())

	// ヘルスチェック・メトリクス（認証不要）
	r.GET("/health", h.Health.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 認証（認証不要）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", h.Auth.Login)
			auth.POST("/register", h.Auth.Register)
			auth.POST("/refresh", h.Auth.RefreshToken)
		}

		// 認証が必要なルート
		protected := v1.Group("")
		protected.Use(mw.Auth())
		{
			// ログアウト
			protected.POST("/auth/logout", h.Auth.Logout)

			// 勤怠
			attendance := protected.Group("/attendance")
			{
				attendance.POST("/clock-in", h.Attendance.ClockIn)
				attendance.POST("/clock-out", h.Attendance.ClockOut)
				attendance.GET("", h.Attendance.GetMyAttendances)
				attendance.GET("/today", h.Attendance.GetTodayStatus)
				attendance.GET("/summary", h.Attendance.GetSummary)
			}

			// 休暇申請
			leaves := protected.Group("/leaves")
			{
				leaves.POST("", h.Leave.Create)
				leaves.GET("", h.Leave.GetMy)
			}

			// ユーザー
			users := protected.Group("/users")
			{
				users.GET("/me", h.User.GetMe)
			}

			// 部署
			departments := protected.Group("/departments")
			{
				departments.GET("", h.Department.GetAll)
			}

			// シフト
			shifts := protected.Group("/shifts")
			{
				shifts.GET("", h.Shift.GetByDateRange)
			}

			// 管理者・マネージャー向け
			admin := protected.Group("")
			admin.Use(mw.RequireRole(model.RoleAdmin, model.RoleManager))
			{
				// 休暇承認
				admin.GET("/leaves/pending", h.Leave.GetPending)
				admin.PUT("/leaves/:id/approve", h.Leave.Approve)

				// ユーザー管理
				admin.GET("/users", h.User.GetAll)
				admin.POST("/users", h.User.Create)
				admin.PUT("/users/:id", h.User.Update)
				admin.DELETE("/users/:id", h.User.Delete)

				// シフト管理
				admin.POST("/shifts", h.Shift.Create)
				admin.POST("/shifts/bulk", h.Shift.BulkCreate)
				admin.DELETE("/shifts/:id", h.Shift.Delete)

				// ダッシュボード
				admin.GET("/dashboard/stats", h.Dashboard.GetStats)
			}
		}
	}
}
