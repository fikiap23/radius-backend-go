package module

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"go.uber.org/zap"
)

// RunInTransactionFunc runs fn inside a single DB transaction (see database.RunInTx).
type RunInTransactionFunc func(
	ctx context.Context,
	fn func(ctx context.Context, tx *ent.Client) error,
) error

type Dependencies struct {
	Config           *config.Config
	Logger           *zap.Logger
	Ent              *ent.Client
	RunInTransaction RunInTransactionFunc
}

type BoundedContext interface {
	Name() string
	RegisterHTTP(e *echo.Echo, api huma.API, deps Dependencies, auth *middleware.AuthMiddleware)
	StartMessaging(ctx context.Context, deps Dependencies) (stop func(), err error)
}
