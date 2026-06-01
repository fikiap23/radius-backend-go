package humaapi

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

type ErrorMapping struct {
	Err     error
	Status  int
	Code    string
	Message string
	Type    string
	Param   string
}

// APIErrorDetail is the nested error object (Stripe-style).
type APIErrorDetail struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// apiErrorResponse is serialized as the HTTP body as-is (Huma does not unwrap a Body field on errors).
type apiErrorResponse struct {
	status  int
	ErrBody APIErrorDetail `json:"error"`
}

func (e *apiErrorResponse) Error() string {
	return e.ErrBody.Message
}

func (e *apiErrorResponse) GetStatus() int {
	return e.status
}

var installOnce sync.Once

func installEnvelopeErrors() {
	installOnce.Do(func() {
		huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
			_ = errs
			if status == http.StatusUnprocessableEntity {
				return newAPIError(status, "validation_error", "Request validation failed.", "", "invalid_request_error")
			}
			if status == http.StatusInternalServerError {
				return newAPIError(status, "internal_error", "An unexpected error occurred.", "", "api_error")
			}
			code := toSnakeCode(msg)
			return newAPIError(status, code, defaultErrorMessage(code), "", errorTypeForStatus(status))
		}
	})
}

func newAPIError(status int, code, message, param, errType string) huma.StatusError {
	if errType == "" {
		errType = errorTypeForStatus(status)
	}
	if message == "" {
		message = defaultErrorMessage(code)
	}
	return &apiErrorResponse{
		status: status,
		ErrBody: APIErrorDetail{
			Type:    errType,
			Code:    code,
			Message: message,
			Param:   param,
		},
	}
}

func errorTypeForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "authentication_error"
	case http.StatusForbidden:
		return "permission_error"
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		return "api_error"
	default:
		return "invalid_request_error"
	}
}

func toSnakeCode(msg string) string {
	if msg == "" {
		return "unknown_error"
	}
	return strings.ToLower(msg)
}

func defaultErrorMessage(code string) string {
	words := strings.Split(code, "_")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ") + "."
}

func MapError(err error, mappings []ErrorMapping, logger *zap.Logger) error {
	if err == nil {
		return nil
	}
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			return newAPIError(m.Status, m.Code, m.Message, m.Param, m.Type)
		}
	}
	if logger != nil {
		logger.Error("handler error", zap.Error(err))
	}
	return newAPIError(http.StatusInternalServerError, "internal_error", "An unexpected error occurred.", "", "api_error")
}
