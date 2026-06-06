package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/storage/application/services"
	"github.com/radius/radius-backend/internal/storage/interface/api/rest"
	"go.uber.org/zap"
)

type Module struct {
	storageSvc      *services.StorageService
	tempCleanupSvc  *services.TempCleanupService
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "storage"
}

func (m *Module) wire(deps module.Dependencies) error {
	if m.storageSvc != nil {
		return nil
	}
	if deps.ObjectStorage == nil {
		return fmt.Errorf("object storage is not configured")
	}

	m.storageSvc = services.NewStorageService(deps.ObjectStorage, deps.Logger)
	m.tempCleanupSvc = services.NewTempCleanupService(
		deps.ObjectStorage,
		deps.Config.MinIO.TempMaxAge,
		deps.Logger,
	)
	return nil
}

func (m *Module) RegisterHTTP(_ *echo.Echo, api huma.API, deps module.Dependencies, _ *middleware.AuthMiddleware) {
	if err := m.wire(deps); err != nil {
		deps.Logger.Fatal("storage module wire failed", zap.Error(err))
	}
	rest.RegisterStorage(api, m.storageSvc, deps.Logger)
}

func (m *Module) StartMessaging(ctx context.Context, deps module.Dependencies) (func(), error) {
	if err := m.wire(deps); err != nil {
		return nil, err
	}

	interval := deps.Config.MinIO.TempCleanupInterval
	if interval <= 0 {
		interval = time.Hour
	}

	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		if err := m.tempCleanupSvc.RunOnce(ctx); err != nil {
			deps.Logger.Warn("initial temp upload cleanup failed", zap.Error(err))
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				if err := m.tempCleanupSvc.RunOnce(ctx); err != nil {
					deps.Logger.Warn("temp upload cleanup failed", zap.Error(err))
				}
			}
		}
	}()

	return func() {
		close(done)
		ticker.Stop()
	}, nil
}

var _ module.BoundedContext = (*Module)(nil)
