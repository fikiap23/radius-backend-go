package projects

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/projects/application/services"
	"github.com/radius/radius-backend/internal/projects/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/projects/interface/api/rest"
	"github.com/radius/radius-backend/internal/shared/middleware"
	wspostgres "github.com/radius/radius-backend/internal/workspaces/infrastructure/db/postgres"
)

type Module struct {
	projectSvc *services.ProjectService
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
	memberRepo := wspostgres.NewWorkspaceMemberRepository(deps.Ent)

	m.projectSvc = services.NewProjectService(
		projectRepo,
		memberRepo,
		deps.Logger,
	)
}

func (m *Module) RegisterHTTP(e *echo.Echo, api huma.API, deps module.Dependencies, auth *middleware.AuthMiddleware) {
	m.wire(deps)
	rest.RegisterProjects(api, m.projectSvc, auth, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
