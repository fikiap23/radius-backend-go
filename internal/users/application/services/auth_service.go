package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/shared/config"
	appjwt "github.com/radius/radius-backend/internal/shared/jwt"
	appoauth "github.com/radius/radius-backend/internal/shared/oauth"
	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       domain.UserRepository
	uow            domain.UnitOfWork
	oauthProviders map[domain.OAuthProvider]domain.SSOProvider
	oauthCfg       config.OAuthConfig
	jwtCfg         config.JWTConfig
	logger         *zap.Logger
}

func NewAuthService(
	userRepo domain.UserRepository,
	uow domain.UnitOfWork,
	oauthProviders map[domain.OAuthProvider]domain.SSOProvider,
	oauthCfg config.OAuthConfig,
	jwtCfg config.JWTConfig,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		uow:            uow,
		oauthProviders: oauthProviders,
		oauthCfg:       oauthCfg,
		jwtCfg:         jwtCfg,
		logger:         logger,
	}
}

func (s *AuthService) HandleRegister(ctx context.Context, in dto.RegisterInput) (*dto.AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	_, err := s.userRepo.FindOne(ctx, domain.Query{
		Select: domain.FieldsExists,
		Filter: domain.Filter{Email: &email},
	})
	if err == nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("check email exists: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)

	user := &domain.User{
		ID:           uuid.NewString(),
		Name:         in.Body.Name,
		Email:        email,
		PasswordHash: &hashStr,
		Locale:       "en",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.PasswordHash = nil

	return s.issueAuthResult(user)
}

func (s *AuthService) HandleLogin(ctx context.Context, in dto.LoginInput) (*dto.AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	user, err := s.userRepo.FindOne(ctx, domain.Query{
		Select: domain.FieldsLogin,
		Filter: domain.Filter{Email: &email},
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(in.Body.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err := s.userRepo.UpdateByID(ctx, user.ID, domain.Update{LastLoginAt: &now}); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

	return s.issueAuthResult(user)
}

func (s *AuthService) HandleGoogleSSOAuthURL(_ context.Context, in dto.GoogleSSOAuthURLInput) (*dto.SSOAuthURLResult, error) {
	return s.buildSSOAuthURL(domain.OAuthProviderGoogle, in.RedirectURI)
}

func (s *AuthService) HandleGitHubSSOAuthURL(_ context.Context, in dto.GitHubSSOAuthURLInput) (*dto.SSOAuthURLResult, error) {
	return s.buildSSOAuthURL(domain.OAuthProviderGitHub, in.RedirectURI)
}

func (s *AuthService) HandleGoogleSSOCallback(ctx context.Context, in dto.GoogleSSOCallbackInput) (*dto.AuthResult, error) {
	return s.handleSSOCallback(ctx, domain.OAuthProviderGoogle, in.Body.Code, in.Body.State)
}

func (s *AuthService) HandleGitHubSSOCallback(ctx context.Context, in dto.GitHubSSOCallbackInput) (*dto.AuthResult, error) {
	return s.handleSSOCallback(ctx, domain.OAuthProviderGitHub, in.Body.Code, in.Body.State)
}

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

func (s *AuthService) handleSSOCallback(ctx context.Context, provider domain.OAuthProvider, code, state string) (*dto.AuthResult, error) {
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

	return s.issueAuthResult(user)
}

func (s *AuthService) resolveSSOUser(ctx context.Context, provider domain.OAuthProvider, info *domain.OAuthUserInfo) (*domain.User, error) {
	var resolved *domain.User

	err := s.uow.Do(ctx, func(ctx context.Context, repos domain.UsersRepositories) error {
		existing, err := repos.OAuthAccounts.FindByProviderAccount(ctx, provider, info.ProviderUserID)
		if err != nil && !errors.Is(err, domain.ErrOAuthAccountNotFound) {
			return fmt.Errorf("find oauth account: %w", err)
		}
		if err == nil {
			user, err := repos.Users.FindByID(ctx, existing.UserID, domain.FieldsProfile)
			if err != nil {
				return fmt.Errorf("find linked user: %w", err)
			}
			resolved = user
			return nil
		}

		user, err := repos.Users.FindOne(ctx, domain.Query{
			Select: domain.FieldsProfile,
			Filter: domain.Filter{Email: &info.Email},
		})
		if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
			return fmt.Errorf("find user by email: %w", err)
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			user = nil
		}

		if user == nil {
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
				return fmt.Errorf("create sso user: %w", err)
			}
		} else if info.AvatarURL != nil && user.AvatarURL == nil {
			user.AvatarURL = info.AvatarURL
			if err := repos.Users.UpdateByID(ctx, user.ID, domain.Update{AvatarURL: user.AvatarURL}); err != nil {
				return fmt.Errorf("update user avatar: %w", err)
			}
		}

		if err := repos.OAuthAccounts.Create(ctx, &domain.OAuthAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: info.ProviderUserID,
		}); err != nil {
			return fmt.Errorf("link oauth account: %w", err)
		}

		resolved = user
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resolved, nil
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

func (s *AuthService) issueAuthResult(user *domain.User) (*dto.AuthResult, error) {
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
