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
	{Err: domain.ErrEmailAlreadyExists, Status: http.StatusConflict, Code: "email_already_exists", Message: "An account with this email already exists."},
	{Err: domain.ErrInvalidCredentials, Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "Invalid email or password."},
	{Err: domain.ErrSSOProviderDisabled, Status: http.StatusServiceUnavailable, Code: "sso_provider_disabled", Message: "This SSO provider is not configured."},
	{Err: domain.ErrSSOInvalidState, Status: http.StatusBadRequest, Code: "sso_invalid_state", Message: "The SSO state parameter is invalid or expired."},
	{Err: domain.ErrSSOInvalidRedirectURI, Status: http.StatusBadRequest, Code: "sso_invalid_redirect_uri", Message: "The redirect URI is not allowed."},
	{Err: domain.ErrSSOAuthenticationFailed, Status: http.StatusUnauthorized, Code: "sso_authentication_failed", Message: "SSO authentication failed."},
	{Err: domain.ErrSSOGitHubEmailPermission, Status: http.StatusForbidden, Code: "sso_github_email_permission", Message: "GitHub did not grant access to your email address."},
}

func RegisterAuth(api huma.API, svc *services.AuthService, logger *zap.Logger) {
	huma.Register(api, huma.Operation{
		OperationID:   "auth-register",
		Method:        http.MethodPost,
		Path:          "/auth/register",
		Summary:       "Register user",
		Tags:          []string{"auth"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.RegisterInput) (*humaapi.CreatedOutput, error) {
		result, err := svc.HandleRegister(ctx, *in)
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
		result, err := svc.HandleLogin(ctx, *in)
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
		result, err := svc.HandleGoogleSSOAuthURL(ctx, *in)
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
		result, err := svc.HandleGoogleSSOCallback(ctx, *in)
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
		result, err := svc.HandleGitHubSSOAuthURL(ctx, *in)
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
		result, err := svc.HandleGitHubSSOCallback(ctx, *in)
		if err != nil {
			return nil, humaapi.MapError(err, authErrors, logger)
		}
		return humaapi.OK(result), nil
	})
}
