package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
	"golang.org/x/time/rate"
)

// Middleware は全ミドルウェアを束ねる構造体
type Middleware struct {
	config *config.Config
	logger *logger.Logger
}

// NewMiddleware はミドルウェアを初期化する
func NewMiddleware(cfg *config.Config, logger *logger.Logger) *Middleware {
	return &Middleware{config: cfg, logger: logger}
}

// ===== CORS =====

func (m *Middleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range m.config.AllowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ===== JWT認証 =====

func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Code:    401,
				Message: "認証トークンが必要です",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Code:    401,
				Message: "Bearerトークンの形式が不正です",
			})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.config.JWTSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Code:    401,
				Message: "無効なトークンです",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Code:    401,
				Message: "トークンのクレームが不正です",
			})
			return
		}

		c.Set("userID", claims["sub"].(string))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

// ===== ロールベースアクセス制御 =====

func (m *Middleware) RequireRole(roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse{
				Code:    403,
				Message: "権限がありません",
			})
			return
		}

		userRole := model.Role(roleStr.(string))
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse{
			Code:    403,
			Message: "この操作を行う権限がありません",
		})
	}
}

// ===== レート制限 =====

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*clientLimiter)
	mu      sync.Mutex
)

func (m *Middleware) RateLimit() gin.HandlerFunc {
	// クリーンアップゴルーチン
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		if _, exists := clients[ip]; !exists {
			clients[ip] = &clientLimiter{
				limiter: rate.NewLimiter(rate.Limit(m.config.RateLimitRPS), m.config.RateLimitBurst),
			}
		}
		clients[ip].lastSeen = time.Now()
		limiter := clients[ip].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, model.ErrorResponse{
				Code:    429,
				Message: "リクエスト回数が制限を超えました。しばらくしてからお試しください",
			})
			return
		}

		c.Next()
	}
}

// ===== セキュリティヘッダー =====

func (m *Middleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// ===== リクエストロギング =====

func (m *Middleware) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		m.logger.Info("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// ===== CSRF =====

func (m *Middleware) CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		cookie, err := c.Cookie("csrf_token")

		if err != nil || token == "" || token != cookie {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse{
				Code:    403,
				Message: "CSRFトークンが不正です",
			})
			return
		}

		c.Next()
	}
}

// ===== リカバリー =====

func (m *Middleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error("パニックが発生しました", "error", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{
					Code:    500,
					Message: "内部サーバーエラーが発生しました",
				})
			}
		}()
		c.Next()
	}
}
