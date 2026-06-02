package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	appjwt "github.com/radius/radius-backend/internal/shared/jwt"
	appoauth "github.com/radius/radius-backend/internal/shared/oauth"
	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

func (s *AuthService) buildSSOAuthURL(provider domain.OAuthProvider, redirectURI string) (*dto.SSOAuthURLResult, error) {
	p, err := s.requireEnabledProvider(provider)
	if err != nil {
		return nil, err
	}

	redirectURI = strings.TrimSpace(redirectURI)
	if err := s.validateRedirectURI(redirectURI); err != nil {
		return nil, err
	}

	state, err := appoauth.SignState(s.jwtCfg, string(provider), redirectURI, s.oauthCfg.StateExpiry)
	if err != nil {
		return nil, fmt.Errorf("sign oauth state: %w", err)
	}

	return &dto.SSOAuthURLResult{
		AuthURL: p.AuthURL(state, redirectURI),
		State:   state,
	}, nil
}

func (s *AuthService) processSSOCallback(
	ctx context.Context,
	provider domain.OAuthProvider,
	code, state string,
) (*dto.AuthResult, error) {
	p, err := s.requireEnabledProvider(provider)
	if err != nil {
		return nil, err
	}

	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if code == "" || state == "" {
		return nil, domain.ErrSSOAuthenticationFailed
	}

	redirectURI, err := appoauth.VerifyState(s.jwtCfg, state, string(provider))
	if err != nil {
		return nil, domain.ErrSSOInvalidState
	}
	if err := s.validateRedirectURI(redirectURI); err != nil {
		return nil, err
	}

	userInfo, err := p.Exchange(ctx, code, redirectURI)
	if err != nil {
		s.logger.Warn("oauth exchange failed", zap.String("provider", string(provider)), zap.Error(err))
		if provider == domain.OAuthProviderGitHub && strings.Contains(err.Error(), "github emails forbidden") {
			return nil, domain.ErrSSOGitHubEmailPermission
		}
		return nil, domain.ErrSSOAuthenticationFailed
	}

	user, err := s.resolveSSOUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.userRepo.UpdateByID(ctx, user.ID, domain.Update{LastLoginAt: &now}); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

	return s.authResultFromUser(user)
}

func (s *AuthService) resolveSSOUser(
	ctx context.Context,
	provider domain.OAuthProvider,
	info *domain.OAuthUserInfo,
) (*domain.User, error) {
	var resolved *domain.User

	err := s.runUsersTx(ctx, func(ctx context.Context, repos domain.UsersRepositories) error {
		user, err := s.linkSSOUserInTx(ctx, repos, provider, info)
		if err != nil {
			return err
		}
		resolved = user
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resolved, nil
}

func (s *AuthService) linkSSOUserInTx(
	ctx context.Context,
	repos domain.UsersRepositories,
	provider domain.OAuthProvider,
	info *domain.OAuthUserInfo,
) (*domain.User, error) {
	if user, err := s.findSSOUserByOAuthAccount(ctx, repos, provider, info.ProviderUserID); err != nil {
		return nil, err
	} else if user != nil {
		return user, nil
	}

	user, err := s.findOrCreateSSOUserByEmail(ctx, repos, info)
	if err != nil {
		return nil, err
	}

	if err := repos.OAuthAccounts.Create(ctx, &domain.OAuthAccount{
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
	}); err != nil {
		if errors.Is(err, domain.ErrOAuthAccountAlreadyExists) {
			return s.findSSOUserByOAuthAccount(ctx, repos, provider, info.ProviderUserID)
		}
		return nil, fmt.Errorf("link oauth account: %w", err)
	}

	return user, nil
}

func (s *AuthService) findSSOUserByOAuthAccount(
	ctx context.Context,
	repos domain.UsersRepositories,
	provider domain.OAuthProvider,
	providerUserID string,
) (*domain.User, error) {
	existing, err := repos.OAuthAccounts.FindByProviderAccount(ctx, provider, providerUserID)
	if err != nil {
		if errors.Is(err, domain.ErrOAuthAccountNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find oauth account: %w", err)
	}

	user, err := repos.Users.FindByID(ctx, existing.UserID, domain.FieldsProfile)
	if err != nil {
		return nil, fmt.Errorf("find linked user: %w", err)
	}
	return user, nil
}

func (s *AuthService) findOrCreateSSOUserByEmail(
	ctx context.Context,
	repos domain.UsersRepositories,
	info *domain.OAuthUserInfo,
) (*domain.User, error) {
	user, err := repos.Users.FindOne(ctx, domain.Query{
		Select: domain.FieldsProfile,
		Filter: domain.Filter{Email: &info.Email},
	})
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	if err == nil {
		if info.AvatarURL != nil && user.AvatarURL == nil {
			user.AvatarURL = info.AvatarURL
			if err := repos.Users.UpdateByID(ctx, user.ID, domain.Update{AvatarURL: user.AvatarURL}); err != nil {
				return nil, fmt.Errorf("update user avatar: %w", err)
			}
		}
		return user, nil
	}

	user = &domain.User{
		ID:        uuid.NewString(),
		Name:      info.Name,
		Email:     info.Email,
		AvatarURL: info.AvatarURL,
		Locale:    "en",
	}
	if info.EmailVerified {
		now := time.Now().UTC()
		user.EmailVerifiedAt = &now
	}
	if err := repos.Users.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return repos.Users.FindOne(ctx, domain.Query{
				Select: domain.FieldsProfile,
				Filter: domain.Filter{Email: &info.Email},
			})
		}
		return nil, fmt.Errorf("create sso user: %w", err)
	}
	return user, nil
}

func (s *AuthService) requireEnabledProvider(provider domain.OAuthProvider) (domain.SSOProvider, error) {
	p, err := domain.SSOProviderFrom(s.oauthProviders, provider)
	if err != nil {
		return nil, err
	}
	if !p.Enabled() {
		return nil, domain.ErrSSOProviderDisabled
	}
	return p, nil
}

func (s *AuthService) validateRedirectURI(redirectURI string) error {
	if redirectURI == "" {
		return domain.ErrSSOInvalidRedirectURI
	}

	parsed, err := url.Parse(redirectURI)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return domain.ErrSSOInvalidRedirectURI
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return domain.ErrSSOInvalidRedirectURI
	}

	if len(s.oauthCfg.AllowedRedirectURIs) == 0 {
		return nil
	}

	for _, allowed := range s.oauthCfg.AllowedRedirectURIs {
		if strings.TrimSpace(allowed) == redirectURI {
			return nil
		}
	}

	return domain.ErrSSOInvalidRedirectURI
}

func (s *AuthService) authResultFromUser(user *domain.User) (*dto.AuthResult, error) {
	token, expiresIn, err := appjwt.SignAccessToken(s.jwtCfg, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        dto.MapUserProfile(user),
	}, nil
}
