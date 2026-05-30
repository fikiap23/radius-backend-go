package humaapi

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/config"
)

func NewConfig(cfg *config.Config) huma.Config {
	hc := huma.DefaultConfig("Radius Backend API", "1.0.0")
	hc.Info.Description = "Radius monolith API — auth and users."

	if cfg.App.Env == "production" {
		hc.DocsPath = ""
		hc.OpenAPIPath = ""
		hc.SchemasPath = ""
	}

	hc.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT token. Format: Bearer {token}",
		},
	}

	installEnvelopeErrors()

	return hc
}
