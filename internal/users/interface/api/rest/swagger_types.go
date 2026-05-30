package rest

import swagschema "github.com/radius/radius-backend/internal/shared/swagger"

// Swagger response aliases — reuse in @Success / @Failure annotations.
type (
	SwaggerErr      = swagschema.Err
	SwaggerAuthOK   = swagschema.AuthOK
	SwaggerUserOK   = swagschema.UserOK
	SwaggerHealthOK = swagschema.HealthOK
)
