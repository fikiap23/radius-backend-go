package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthController struct{}

func NewHealthController(e *echo.Echo) *HealthController {
	e.GET("/health", health)
	return &HealthController{}
}

// health godoc
// @Summary      Health check
// @Tags         system
// @Produce      json
// @Success      200  {object}  SwaggerHealthOK
// @Router       /health [get]
func health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
