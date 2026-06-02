package oauth

import (
	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/users/domain"
)

func NewRegistry(cfg config.OAuthConfig) map[domain.OAuthProvider]domain.SSOProvider {
	return map[domain.OAuthProvider]domain.SSOProvider{
		domain.OAuthProviderGoogle: NewGoogleProvider(cfg.Google),
		domain.OAuthProviderGitHub: NewGitHubProvider(cfg.GitHub),
	}
}
