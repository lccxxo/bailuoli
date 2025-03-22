package start

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lccxxo/bailuoli/internal/controller"
	"github.com/lccxxo/bailuoli/internal/model"

	"github.com/lccxxo/bailuoli/internal/config"
	"github.com/lccxxo/bailuoli/internal/logger"
	"go.uber.org/zap"
)

func Run() {
	// 加载配置
	cfg, err := config.Load("configs/gateway.yaml")
	if err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	// 初始化日志
	logger.InitLogger(cfg.Log.Level, cfg.Log.Outputs, logger.RotationConfig{
		MaxSize:    cfg.Log.Rotation.MaxSize,
		MaxAge:     cfg.Log.Rotation.MaxAge,
		MaxBackups: cfg.Log.Rotation.MaxBackups,
		Compress:   cfg.Log.Rotation.Compress,
	})
	defer logger.Sync()

	// 初始化路由
	router := controller.NewRouter(cfg.Routes)

	// 启动配置热更新监听
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.Watch(ctx, "configs/gateway.yaml", func(newCfg *model.Config) {
		logger.Logger.Info("检测到配置变更，开始热更新")

		// 更新路由
		if err := router.UpdateRoutes(newCfg.Routes); err != nil {
			logger.Logger.Error("路由更新失败", zap.Error(err))
			return
		}

		// 更新日志配置
		logger.UpdateLogLevel(newCfg.Log.Level)

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

	// 启动服务
	startServer(server, cfg)

	// 优雅关闭
	waitForShutdown(server, cfg.Server.ShutdownTimeout)
}

func startServer(server *http.Server, cfg *model.Config) {
	go func() {
		logger.Logger.Info("Starting API Gateway",
			zap.String("address", server.Addr))

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("Server crashed",
				zap.String("error", err.Error()))
		}
	}()
}

func waitForShutdown(server *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Logger.Info("Shutting down server...",
		zap.Time("timestamp", time.Now()))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Error("Shutdown error",
			zap.String("error", err.Error()))
	}
}
