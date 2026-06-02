package workspaces

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/shared/middleware"
	userpostgres "github.com/radius/radius-backend/internal/users/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/workspaces/application/services"
	"github.com/radius/radius-backend/internal/workspaces/domain"
	"github.com/radius/radius-backend/internal/workspaces/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/workspaces/interface/api/rest"
)

type Module struct {
	workspaceSvc *services.WorkspaceService
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "workspaces"
}

func (m *Module) wire(deps module.Dependencies) {
	if m.workspaceSvc != nil {
		return
	}

	workspaceRepo := postgres.NewWorkspaceRepository(deps.Ent)
	memberRepo := postgres.NewWorkspaceMemberRepository(deps.Ent)
	userRepo := userpostgres.NewUserRepository(deps.Ent)

	runWorkspacesTx := func(ctx context.Context, fn func(context.Context, domain.WorkspacesRepositories) error) error {
		return deps.RunInTransaction(ctx, func(ctx context.Context, tx *ent.Client) error {
			return fn(ctx, domain.WorkspacesRepositories{
				Workspaces:       postgres.NewWorkspaceRepository(tx),
				WorkspaceMembers: postgres.NewWorkspaceMemberRepository(tx),
			})
		})
	}

	m.workspaceSvc = services.NewWorkspaceService(
		workspaceRepo,
		memberRepo,
		userRepo,
		runWorkspacesTx,
		deps.Logger,
	)
}

func (m *Module) RegisterHTTP(e *echo.Echo, api huma.API, deps module.Dependencies, auth *middleware.AuthMiddleware) {
	m.wire(deps)
	rest.RegisterWorkspaces(api, m.workspaceSvc, auth, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
