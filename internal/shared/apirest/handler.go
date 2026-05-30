package apirest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/shared/response"
	"go.uber.org/zap"
)

var (
	ErrInvalidRequest = errors.New("INVALID_REQUEST")
	ErrValidation     = errors.New("VALIDATION_ERROR")
	ErrUnauthorized   = errors.New("UNAUTHORIZED")
)

type ErrorMapping struct {
	Err     error
	Status  int
	Message string
}

func Bind[T any](c echo.Context) (T, error) {
	var req T
	if err := c.Bind(&req); err != nil {
		return req, ErrInvalidRequest
	}
	if err := c.Validate(&req); err != nil {
		return req, ErrValidation
	}
	return req, nil
}

func BindErr(c echo.Context, err error) error {
	switch err {
	case ErrInvalidRequest:
		return Fail(c, http.StatusBadRequest, "INVALID_REQUEST")
	case ErrValidation:
		return Fail(c, http.StatusBadRequest, "VALIDATION_ERROR")
	default:
		return Fail(c, http.StatusBadRequest, "INVALID_REQUEST")
	}
}

func UserID(c echo.Context) (string, error) {
	id, ok := middleware.GetUserID(c)
	if !ok {
		return "", ErrUnauthorized
	}
	return id, nil
}

func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, response.APIResponse{IsSuccess: true, Message: "OK", Data: data})
}

func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, response.APIResponse{IsSuccess: true, Message: "OK", Data: data})
}

func Fail(c echo.Context, status int, message string) error {
	return c.JSON(status, response.APIResponse{IsSuccess: false, Message: message})
}

func Unauthorized(c echo.Context) error {
	return Fail(c, http.StatusUnauthorized, "UNAUTHORIZED")
}

func Internal(c echo.Context) error {
	return Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func Handle(c echo.Context, err error, mappings []ErrorMapping, logger *zap.Logger) error {
	if err == nil {
		return nil
	}
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			return Fail(c, m.Status, m.Message)
		}
	}
	if errors.Is(err, ErrUnauthorized) {
		return Unauthorized(c)
	}
	if logger != nil {
		logger.Error("handler error", zap.Error(err))
	}
	return Internal(c)
}
