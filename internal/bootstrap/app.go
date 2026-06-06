package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/shared/database"
	"github.com/radius/radius-backend/internal/shared/httplog"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/shared/middleware"
	sharedstorage "github.com/radius/radius-backend/internal/shared/storage"
	"github.com/radius/radius-backend/internal/projects"
	"github.com/radius/radius-backend/internal/storage"
	storageminio "github.com/radius/radius-backend/internal/storage/infrastructure/minio"
	"github.com/radius/radius-backend/internal/users"
	"github.com/radius/radius-backend/internal/workspaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := buildLogger(cfg.App.Env, cfg.App.LogLevel)
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pg, err := database.NewPostgres(cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pg.Close()

	minioClient, err := sharedstorage.NewClient(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("connect minio: %w", err)
	}
	objectStorage := storageminio.NewObjectStorage(minioClient)

	deps := module.Dependencies{
		Config:        cfg,
		Logger:        logger,
		Ent:           pg.Client,
		ObjectStorage: objectStorage,
		RunInTransaction: func(ctx context.Context, fn func(context.Context, *ent.Client) error) error {
			return database.RunInTx(ctx, pg.Client, fn)
		},
	}

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.SecretKey, logger)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(
		httplog.RequestLogger(logger),
		echomw.Recover(),
		echomw.RequestID(),
		middleware.CORS(cfg.HTTP.CORS),
	)

	api := humaecho.New(e, humaapi.NewConfig(cfg))

	contexts := []module.BoundedContext{
		users.NewModule(),
		workspaces.NewModule(),
		projects.NewModule(),
		storage.NewModule(),
	}

	var messagingStops []func()
	for _, bc := range contexts {
		bc.RegisterHTTP(e, api, deps, authMiddleware)
		stopFn, err := bc.StartMessaging(ctx, deps)
		if err != nil {
			return fmt.Errorf("start messaging %s: %w", bc.Name(), err)
		}
		messagingStops = append(messagingStops, stopFn)
	}
	defer func() {
		for _, stopFn := range messagingStops {
			if stopFn != nil {
				stopFn()
			}
		}
	}()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      e,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening",
			zap.Int("port", cfg.HTTP.Port),
			zap.String("env", cfg.App.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

func buildLogger(env, level string) (*zap.Logger, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.Level = zap.NewAtomicLevelAt(lvl)
	return cfg.Build()
}
