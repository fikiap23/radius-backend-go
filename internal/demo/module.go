package demo

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/module"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/demo/application/services"
	"github.com/radius/radius-backend/internal/demo/interface/api/rest"
)

type Module struct {
	demoSvc *services.DemoService
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "demo"
}

func (m *Module) wire(deps module.Dependencies) {
	if m.demoSvc != nil {
		return
	}

	// TODO: wire repositories from postgres.New*Repository(deps.Ent)
	m.demoSvc = services.NewDemoService(deps.Logger)
}

func (m *Module) RegisterHTTP(_ *echo.Echo, api huma.API, deps module.Dependencies, _ *middleware.AuthMiddleware) {
	m.wire(deps)
	rest.RegisterDemo(api, m.demoSvc, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
