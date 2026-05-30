package httplog

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func RequestLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()
			fields := []zap.Field{
				zap.Int("status", res.Status),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("query", req.URL.RawQuery),
				zap.String("ip", c.RealIP()),
				zap.Duration("latency", time.Since(start)),
				zap.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
			}

			switch {
			case res.Status >= http.StatusInternalServerError:
				logger.Error("server error", fields...)
			case res.Status >= http.StatusBadRequest:
				logger.Warn("client error", fields...)
			default:
				logger.Info("request handled", fields...)
			}

			return err
		}
	}
}
