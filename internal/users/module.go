package users

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/shared/httplog"
	appmiddleware "github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/users/interface/api/rest"
)

type Module struct {
	authSvc *services.AuthService
	userSvc *services.UserService
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "users"
}

func (m *Module) wire(deps module.Dependencies) {
	if m.authSvc != nil {
		return
	}

	userRepo := postgres.NewGormUserRepository(deps.DB, deps.Logger)
	m.authSvc = services.NewAuthService(userRepo, deps.Config.JWT, deps.Logger)
	m.userSvc = services.NewUserService(userRepo, deps.Logger)
}

func (m *Module) RegisterHTTP(e *echo.Echo, deps module.Dependencies, auth *appmiddleware.AuthMiddleware) {
	m.wire(deps)

	e.Use(
		httplog.RequestLogger(deps.Logger),
		middleware.Recover(),
		middleware.RequestID(),
	)

	rest.NewHealthController(e)
	rest.NewAuthController(e, m.authSvc, deps.Logger)
	rest.NewUserController(e, m.userSvc, auth, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
