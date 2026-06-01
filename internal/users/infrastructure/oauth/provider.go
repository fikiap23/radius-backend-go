package oauth

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/users/domain"
)

type UserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      *string
	EmailVerified  bool
}

type Provider interface {
	Name() domain.OAuthProvider
	Enabled() bool
	AuthURL(state, redirectURI string) string
	Exchange(ctx context.Context, code, redirectURI string) (*UserInfo, error)
}

func NewRegistry(cfg config.OAuthConfig) map[domain.OAuthProvider]Provider {
	return map[domain.OAuthProvider]Provider{
		domain.OAuthProviderGoogle: NewGoogleProvider(cfg.Google),
		domain.OAuthProviderGitHub: NewGitHubProvider(cfg.GitHub),
	}
}

func GetProvider(registry map[domain.OAuthProvider]Provider, provider domain.OAuthProvider) (Provider, error) {
	p, ok := registry[provider]
	if !ok || p == nil {
		return nil, fmt.Errorf("unknown oauth provider: %s", provider)
	}
	return p, nil
}
