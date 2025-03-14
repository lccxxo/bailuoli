package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lccxxo/bailuoli/internal/config"
	"github.com/lccxxo/bailuoli/internal/controller"
	"github.com/lccxxo/bailuoli/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 入口函数

func main() {
	cfg, err := config.Load("configs/gateway.yaml")
	if err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	logger.InitLogger(cfg.Log.Level, cfg.Log.Outputs, logger.RotationConfig{
		MaxSize:    cfg.Log.Rotation.MaxSize,
		MaxAge:     cfg.Log.Rotation.MaxAge,
		MaxBackups: cfg.Log.Rotation.MaxBackups,
		Compress:   cfg.Log.Rotation.Compress,
	})
	defer logger.Sync()

	router := controller.NewRouter(cfg.Routes)

	// 启动配置热更新监听
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Watch(ctx, "configs/gateway.yaml", func(newCfg *config.Config) {
		logger.Logger.Info("检测到配置变更，开始热更新")

		// 更新路由
		if err := router.UpdateRoutes(newCfg.Routes); err != nil {
			logger.Logger.Error("路由更新失败", zap.Error(err))
			return
		}

		// 更新日志配置（示例）
		logger.UpdateLogLevel(newCfg.Log.Level)

		logger.Logger.Info("当前path", zap.String("path", cfg.Routes[0].Path))
		logger.Logger.Info("配置热更新完成")
	}, 5*time.Second)

	server := &http.Server{
		Addr:         cfg.Server.Addr,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		Handler: logger.LoggingMiddleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				route, handler := router.MatchRoute(r)
				if route == nil {
					http.NotFound(w, r)
					return
				}

				// 传递路由信息到上下文
				ctx := context.WithValue(r.Context(), "route", route)
				handler.ServeHTTP(w, r.WithContext(ctx))
			}),
		),
	}

	go func() {
		logger.Logger.Info("Starting API Gateway",
			zap.String("address", server.Addr))

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("Server crashed",
				zap.String("error", err.Error()))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Logger.Info("Shutting down server...",
		zap.Time("timestamp", time.Now()))

	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Error("Shutdown error",
			zap.String("error", err.Error()))
	}
}
