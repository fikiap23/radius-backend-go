package projects

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/projects/application/services"
	"github.com/radius/radius-backend/internal/projects/domain"
	"github.com/radius/radius-backend/internal/projects/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/projects/interface/api/rest"
	"github.com/radius/radius-backend/internal/shared/middleware"
	wspostgres "github.com/radius/radius-backend/internal/workspaces/infrastructure/db/postgres"
)

type Module struct {
	projectSvc     *services.ProjectService
	boardColumnSvc *services.BoardColumnService
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "projects"
}

func (m *Module) wire(deps module.Dependencies) {
	if m.projectSvc != nil {
		return
	}

	projectRepo := postgres.NewProjectRepository(deps.Ent)
	columnRepo := postgres.NewBoardColumnRepository(deps.Ent)
	memberRepo := wspostgres.NewWorkspaceMemberRepository(deps.Ent)

	runProjectsTx := func(ctx context.Context, fn func(context.Context, domain.ProjectsRepositories) error) error {
		return deps.RunInTransaction(ctx, func(ctx context.Context, tx *ent.Client) error {
			return fn(ctx, domain.ProjectsRepositories{
				BoardColumns: postgres.NewBoardColumnRepository(tx),
				Projects:     postgres.NewProjectRepository(tx),
			})
		})
	}

	m.projectSvc = services.NewProjectService(
		projectRepo,
		memberRepo,
		deps.ObjectStorage,
		deps.Logger,
	)
	m.boardColumnSvc = services.NewBoardColumnService(
		columnRepo,
		projectRepo,
		memberRepo,
		runProjectsTx,
	)
}

func (m *Module) RegisterHTTP(e *echo.Echo, api huma.API, deps module.Dependencies, auth *middleware.AuthMiddleware) {
	m.wire(deps)
	rest.RegisterProjects(api, m.projectSvc, auth, deps.Logger)
	rest.RegisterBoardColumns(api, m.boardColumnSvc, auth, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
