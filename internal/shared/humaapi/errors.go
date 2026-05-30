package humaapi

import (
	"errors"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

type ErrorMapping struct {
	Err     error
	Status  int
	Message string
}

type envelopeError struct {
	status int
	Body   struct {
		IsSuccess bool   `json:"isSuccess"`
		Message   string `json:"message"`
	}
}

func (e *envelopeError) Error() string {
	return e.Body.Message
}

func (e *envelopeError) GetStatus() int {
	return e.status
}

var installOnce sync.Once

func installEnvelopeErrors() {
	installOnce.Do(func() {
		huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
			message := msg
			if status == 422 {
				message = "VALIDATION_ERROR"
			}
			_ = errs
			return &envelopeError{
				status: status,
				Body: struct {
					IsSuccess bool   `json:"isSuccess"`
					Message   string `json:"message"`
				}{
					IsSuccess: false,
					Message:   message,
				},
			}
		}
	})
}

func MapError(err error, mappings []ErrorMapping, logger *zap.Logger) error {
	if err == nil {
		return nil
	}
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			return huma.NewError(m.Status, m.Message)
		}
	}
	if logger != nil {
		logger.Error("handler error", zap.Error(err))
	}
	return huma.NewError(500, "INTERNAL_ERROR")
}
