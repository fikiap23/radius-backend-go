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
	{Err: domain.ErrSSOProviderDisabled, Status: http.StatusServiceUnavailable, Message: "SSO_PROVIDER_DISABLED"},
	{Err: domain.ErrSSOInvalidState, Status: http.StatusBadRequest, Message: "SSO_INVALID_STATE"},
	{Err: domain.ErrSSOInvalidRedirectURI, Status: http.StatusBadRequest, Message: "SSO_INVALID_REDIRECT_URI"},
	{Err: domain.ErrSSOAuthenticationFailed, Status: http.StatusUnauthorized, Message: "SSO_AUTHENTICATION_FAILED"},
	{Err: domain.ErrSSOGitHubEmailPermission, Status: http.StatusForbidden, Message: "SSO_GITHUB_EMAIL_PERMISSION"},
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

	huma.Register(api, huma.Operation{
		OperationID: "auth-sso-google-url",
		Method:      http.MethodGet,
		Path:        "/auth/sso/google/url",
		Summary:     "Get Google SSO authorization URL",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *dto.GoogleSSOAuthURLInput) (*humaapi.OKOutput, error) {
		result, err := service.HandleGoogleSSOAuthURL(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-sso-google-callback",
		Method:      http.MethodPost,
		Path:        "/auth/sso/google/callback",
		Summary:     "Complete Google SSO login",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *dto.GoogleSSOCallbackInput) (*humaapi.OKOutput, error) {
		result, err := service.HandleGoogleSSOCallback(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-sso-github-url",
		Method:      http.MethodGet,
		Path:        "/auth/sso/github/url",
		Summary:     "Get GitHub SSO authorization URL",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *dto.GitHubSSOAuthURLInput) (*humaapi.OKOutput, error) {
		result, err := service.HandleGitHubSSOAuthURL(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-sso-github-callback",
		Method:      http.MethodPost,
		Path:        "/auth/sso/github/callback",
		Summary:     "Complete GitHub SSO login",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, in *dto.GitHubSSOCallbackInput) (*humaapi.OKOutput, error) {
		result, err := service.HandleGitHubSSOCallback(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})
}
