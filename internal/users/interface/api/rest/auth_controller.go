package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

var authErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrEmailAlreadyExists, Status: http.StatusConflict, Message: "EMAIL_ALREADY_EXISTS"},
	{Err: domain.ErrInvalidCredentials, Status: http.StatusUnauthorized, Message: "INVALID_CREDENTIALS"},
}

func RegisterAuth(api huma.API, service *services.AuthService, logger *zap.Logger) {
	huma.Register(api, huma.Operation{
		OperationID:   "auth-register",
		Method:        http.MethodPost,
		Path:          "/auth/register",
		Summary:       "Register user",
		Tags:          []string{"auth"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.RegisterInput) (*humaapi.CreatedOutput, error) {
		result, err := service.HandleRegister(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.Created(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *dto.LoginInput) (*humaapi.OKOutput, error) {
		result, err := service.HandleLogin(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})
}
