package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/radius/radius-backend/internal/shared/config"
)

func CORS(cfg config.CORSConfig) echo.MiddlewareFunc {
	if len(cfg.AllowedOrigins) == 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	methods := cfg.AllowMethods
	if len(methods) == 0 {
		methods = []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.PATCH,
			echo.DELETE,
			echo.OPTIONS,
		}
	}

	headers := cfg.AllowHeaders
	if len(headers) == 0 {
		headers = []string{
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderContentType,
		}
	}

	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: methods,
		AllowHeaders: headers,
	})
}
