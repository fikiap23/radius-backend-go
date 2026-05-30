package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
)

func RegisterHealth(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Tags:        []string{"system"},
	}, func(_ context.Context, _ *struct{}) (*humaapi.HealthOutput, error) {
		out := &humaapi.HealthOutput{}
		out.Body.Status = "ok"
		return out, nil
	})
}
