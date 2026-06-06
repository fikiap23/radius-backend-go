package users

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/users/infrastructure/oauth"
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

	userRepo := postgres.NewUserRepository(deps.Ent)
	oauthProviders := oauth.NewRegistry(deps.Config.OAuth)

	runUsersTx := func(ctx context.Context, fn func(context.Context, domain.UsersRepositories) error) error {
		return deps.RunInTransaction(ctx, func(ctx context.Context, tx *ent.Client) error {
			return fn(ctx, domain.UsersRepositories{
				Users:         postgres.NewUserRepository(tx),
				OAuthAccounts: postgres.NewOAuthAccountRepository(tx),
			})
		})
	}

	m.authSvc = services.NewAuthService(
		userRepo,
		runUsersTx,
		oauthProviders,
		deps.Config.OAuth,
		deps.Config.JWT,
		deps.Logger,
	)
	m.userSvc = services.NewUserService(userRepo, deps.ObjectStorage, deps.Logger)
}

func (m *Module) RegisterHTTP(e *echo.Echo, api huma.API, deps module.Dependencies, auth *middleware.AuthMiddleware) {
	m.wire(deps)

	rest.RegisterHealth(api)
	rest.RegisterAuth(api, m.authSvc, deps.Logger)
	rest.RegisterUsers(api, m.userSvc, auth, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
