package module

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/shared/config"
	appmiddleware "github.com/radius/radius-backend/internal/shared/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
}

type BoundedContext interface {
	Name() string
	RegisterHTTP(e *echo.Echo, deps Dependencies, auth *appmiddleware.AuthMiddleware)
	StartMessaging(ctx context.Context, deps Dependencies) (stop func(), err error)
}
