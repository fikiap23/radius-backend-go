package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

type registerInput struct {
	Body struct {
		Name     string `json:"name" doc:"Display name" minLength:"2" maxLength:"255"`
		Email    string `json:"email" doc:"Email address" format:"email"`
		Password string `json:"password" doc:"Password" minLength:"8" maxLength:"72"`
	}
}

type loginInput struct {
	Body struct {
		Email    string `json:"email" doc:"Email address" format:"email"`
		Password string `json:"password" doc:"Password" minLength:"8" maxLength:"72"`
	}
}

var authErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrEmailAlreadyExists, Status: http.StatusConflict, Message: "EMAIL_ALREADY_EXISTS"},
	{Err: domain.ErrInvalidCredentials, Status: http.StatusUnauthorized, Message: "INVALID_CREDENTIALS"},
}

func RegisterAuth(api huma.API, service *services.AuthService, logger *zap.Logger) {
	huma.Register(api, huma.Operation{
		OperationID:   "auth-register",
		Method:        http.MethodPost,
		Path:          "/v1/auth/register",
		Summary:       "Register user",
		Tags:          []string{"auth"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *registerInput) (*humaapi.CreatedOutput, error) {
		result, err := service.Register(ctx, in.Body.Name, in.Body.Email, in.Body.Password)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.Created(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/v1/auth/login",
		Summary:     "Login",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *loginInput) (*humaapi.OKOutput, error) {
		result, err := service.Login(ctx, in.Body.Email, in.Body.Password)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})
}
