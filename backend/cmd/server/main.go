package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/handler"
	"github.com/your-org/kintai/backend/internal/middleware"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
	"github.com/your-org/kintai/backend/internal/router"
	"github.com/your-org/kintai/backend/internal/service"
	"github.com/your-org/kintai/backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           勤怠管理システム API
// @version         1.0
// @description     勤怠管理システムのREST API
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// ロガーの初期化
	zapLogger, err := logger.NewLogger(cfg.LogLevel, cfg.Env)
	if err != nil {
		log.Fatalf("ロガーの初期化に失敗: %v", err)
	}
	defer zapLogger.Sync()

	// データベース接続
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		zapLogger.Fatal("データベース接続に失敗", err)
	}

	// 開発環境のみAutoMigrate（本番はgolang-migrate使用）
	// マイグレーションは手動で実行済みの場合はスキップ
	if cfg.Env == "development" {
		// テーブルが存在するかチェック
		var count int64
		db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users'").Scan(&count)
		if count == 0 {
			if err := model.AutoMigrate(db); err != nil {
				zapLogger.Fatal("AutoMigrateに失敗", err)
			}
		} else {
			zapLogger.Info("テーブルは既に存在するため、AutoMigrateをスキップします")
		}
	}

	// リポジトリ層の初期化
	repos := repository.NewRepositories(db)

	// サービス層の初期化
	services := service.NewServices(service.Deps{
		Repos:  repos,
		Config: cfg,
		Logger: zapLogger,
	})

	// ハンドラー層の初期化
	handlers := handler.NewHandlers(services, zapLogger)

	// ミドルウェアの初期化
	mw := middleware.NewMiddleware(cfg, zapLogger)

	// Ginエンジンの設定
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// ルーターの設定
	router.Setup(engine, handlers, mw)

	// HTTPサーバーの起動
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zapLogger.Info("サーバーを起動します", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("サーバーの起動に失敗", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("サーバーをシャットダウンしています...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("サーバーの強制シャットダウン", err)
	}
	zapLogger.Info("サーバーが正常にシャットダウンしました")
}
