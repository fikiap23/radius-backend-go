package domain

import (
	"context"
	"fmt"
)

// OAuthUserInfo holds identity data returned by an external SSO provider.
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      *string
	EmailVerified  bool
}

// SSOProvider exchanges an authorization code for user identity (port).
type SSOProvider interface {
	Name() OAuthProvider
	Enabled() bool
	AuthURL(state, redirectURI string) string
	Exchange(ctx context.Context, code, redirectURI string) (*OAuthUserInfo, error)
}

// SSOProviderFrom returns a registered SSO provider or an error.
func SSOProviderFrom(registry map[OAuthProvider]SSOProvider, provider OAuthProvider) (SSOProvider, error) {
	p, ok := registry[provider]
	if !ok || p == nil {
		return nil, fmt.Errorf("unknown oauth provider: %s", provider)
	}
	return p, nil
}
