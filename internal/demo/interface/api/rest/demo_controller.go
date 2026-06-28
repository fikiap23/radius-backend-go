package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/demo/application/services"
	"go.uber.org/zap"
)

func RegisterDemo(api huma.API, svc *services.DemoService, logger *zap.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "demo-hello",
		Method:      http.MethodGet,
		Path:        "/demo/hello",
		Summary:     "Hello from demo",
		Tags:        []string{"demo"},
	}, func(ctx context.Context, _ *struct{}) (*humaapi.OKOutput, error) {
		out, err := svc.HandleHello(ctx)
		if err != nil {
			return nil, humaapi.MapError(err, nil, logger)
		}
		return humaapi.OK(out), nil
	})
}
