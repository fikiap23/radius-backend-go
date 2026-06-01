package module

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"go.uber.org/zap"
)

type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger
	Ent    *ent.Client
}

type BoundedContext interface {
	Name() string
	RegisterHTTP(e *echo.Echo, deps Dependencies, auth *middleware.AuthMiddleware)
	StartMessaging(ctx context.Context, deps Dependencies) (stop func(), err error)
}
