package oauth

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/users/domain/entities"
)

type UserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      *string
	EmailVerified  bool
}

type Provider interface {
	Name() entities.OAuthProvider
	Enabled() bool
	AuthURL(state, redirectURI string) string
	Exchange(ctx context.Context, code, redirectURI string) (*UserInfo, error)
}

func NewRegistry(cfg config.OAuthConfig) map[entities.OAuthProvider]Provider {
	return map[entities.OAuthProvider]Provider{
		entities.OAuthProviderGoogle: NewGoogleProvider(cfg.Google),
		entities.OAuthProviderGitHub: NewGitHubProvider(cfg.GitHub),
	}
}

func GetProvider(registry map[entities.OAuthProvider]Provider, provider entities.OAuthProvider) (Provider, error) {
	p, ok := registry[provider]
	if !ok || p == nil {
		return nil, fmt.Errorf("unknown oauth provider: %s", provider)
	}
	return p, nil
}
