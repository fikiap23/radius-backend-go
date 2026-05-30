package swagger

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const routePath = "/swagger/*"

// Register mounts Swagger UI. Import docs package in bootstrap before calling this.
func Register(e *echo.Echo) {
	e.GET(routePath, echoSwagger.WrapHandler)
}
