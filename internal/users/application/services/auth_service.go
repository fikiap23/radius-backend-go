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
	appdto "github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	repo "github.com/radius/radius-backend/internal/users/domain/repositories"
	infraoauth "github.com/radius/radius-backend/internal/users/infrastructure/oauth"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthResult struct {
	AccessToken string               `json:"accessToken"`
	TokenType   string               `json:"tokenType"`
	ExpiresIn   int64                `json:"expiresIn"`
	User        entities.UserProfile `json:"user"`
}

type AuthService struct {
	userRepo         repo.UserRepository
	oauthAccountRepo repo.OAuthAccountRepository
	oauthProviders   map[entities.OAuthProvider]infraoauth.Provider
	oauthCfg         config.OAuthConfig
	cfg              config.JWTConfig
	logger           *zap.Logger
}

func NewAuthService(
	userRepo repo.UserRepository,
	oauthAccountRepo repo.OAuthAccountRepository,
	oauthProviders map[entities.OAuthProvider]infraoauth.Provider,
	oauthCfg config.OAuthConfig,
	cfg config.JWTConfig,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		oauthAccountRepo: oauthAccountRepo,
		oauthProviders:   oauthProviders,
		oauthCfg:         oauthCfg,
		cfg:              cfg,
		logger:           logger,
	}
}

func (s *AuthService) HandleRegister(ctx context.Context, in appdto.RegisterInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	_, err := s.userRepo.FindOne(ctx, repo.Query{
		Select: repo.SelectExists,
		Where:  repo.Where{Email: &email},
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

	user := &entities.User{
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

func (s *AuthService) HandleLogin(ctx context.Context, in appdto.LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	user, err := s.userRepo.FindOne(ctx, repo.Query{
		Select: repo.SelectLogin,
		Where:  repo.Where{Email: &email},
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
	if err := s.userRepo.UpdateByID(ctx, user.ID, repo.Update{LastLoginAt: &now}); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

	return s.issueAuthResult(user)
}

func (s *AuthService) HandleGoogleSSOAuthURL(ctx context.Context, in appdto.GoogleSSOAuthURLInput) (*appdto.SSOAuthURLResult, error) {
	return s.handleSSOAuthURL(ctx, entities.OAuthProviderGoogle, in.RedirectURI)
}

func (s *AuthService) HandleGitHubSSOAuthURL(ctx context.Context, in appdto.GitHubSSOAuthURLInput) (*appdto.SSOAuthURLResult, error) {
	return s.handleSSOAuthURL(ctx, entities.OAuthProviderGitHub, in.RedirectURI)
}

func (s *AuthService) HandleGoogleSSOCallback(ctx context.Context, in appdto.GoogleSSOCallbackInput) (*AuthResult, error) {
	return s.handleSSOCallback(ctx, entities.OAuthProviderGoogle, in.Body.Code, in.Body.State)
}

func (s *AuthService) HandleGitHubSSOCallback(ctx context.Context, in appdto.GitHubSSOCallbackInput) (*AuthResult, error) {
	return s.handleSSOCallback(ctx, entities.OAuthProviderGitHub, in.Body.Code, in.Body.State)
}

func (s *AuthService) handleSSOAuthURL(_ context.Context, provider entities.OAuthProvider, redirectURI string) (*appdto.SSOAuthURLResult, error) {
	p, err := s.requireEnabledProvider(provider)
	if err != nil {
		return nil, err
	}

	redirectURI = strings.TrimSpace(redirectURI)
	if err := s.validateRedirectURI(redirectURI); err != nil {
		return nil, err
	}

	state, err := appoauth.SignState(s.cfg, string(provider), redirectURI, s.oauthCfg.StateExpiry)
	if err != nil {
		return nil, fmt.Errorf("sign oauth state: %w", err)
	}

	return &appdto.SSOAuthURLResult{
		AuthURL: p.AuthURL(state, redirectURI),
		State:   state,
	}, nil
}

func (s *AuthService) handleSSOCallback(ctx context.Context, provider entities.OAuthProvider, code, state string) (*AuthResult, error) {
	p, err := s.requireEnabledProvider(provider)
	if err != nil {
		return nil, err
	}

	code = strings.TrimSpace(code)
	state = strings.TrimSpace(state)
	if code == "" || state == "" {
		return nil, domain.ErrSSOAuthenticationFailed
	}

	redirectURI, err := appoauth.VerifyState(s.cfg, state, string(provider))
	if err != nil {
		return nil, domain.ErrSSOInvalidState
	}
	if err := s.validateRedirectURI(redirectURI); err != nil {
		return nil, err
	}

	userInfo, err := p.Exchange(ctx, code, redirectURI)
	if err != nil {
		s.logger.Warn("oauth exchange failed", zap.String("provider", string(provider)), zap.Error(err))
		if provider == entities.OAuthProviderGitHub && strings.Contains(err.Error(), "github emails forbidden") {
			return nil, domain.ErrSSOGitHubEmailPermission
		}
		return nil, domain.ErrSSOAuthenticationFailed
	}

	user, err := s.resolveSSOUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.userRepo.UpdateByID(ctx, user.ID, repo.Update{LastLoginAt: &now}); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

	return s.issueAuthResult(user)
}

func (s *AuthService) resolveSSOUser(ctx context.Context, provider entities.OAuthProvider, info *infraoauth.UserInfo) (*entities.User, error) {
	existingOAuth, err := s.oauthAccountRepo.FindByProviderAccount(ctx, provider, info.ProviderUserID)
	if err != nil {
		return nil, fmt.Errorf("find oauth account: %w", err)
	}
	if existingOAuth != nil {
		user, err := s.userRepo.FindByID(ctx, existingOAuth.UserID, repo.SelectProfile)
		if err != nil {
			return nil, fmt.Errorf("find linked user: %w", err)
		}
		return user, nil
	}

	user, err := s.userRepo.FindOne(ctx, repo.Query{
		Select: repo.SelectProfile,
		Where:  repo.Where{Email: &info.Email},
	})
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if errors.Is(err, domain.ErrUserNotFound) {
		user = nil
	}

	if user == nil {
		user = &entities.User{
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

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create sso user: %w", err)
		}
	} else if info.AvatarURL != nil && user.AvatarURL == nil {
		user.AvatarURL = info.AvatarURL
		if err := s.userRepo.UpdateByID(ctx, user.ID, repo.Update{AvatarURL: user.AvatarURL}); err != nil {
			return nil, fmt.Errorf("update user avatar: %w", err)
		}
	}

	if err := s.oauthAccountRepo.Create(ctx, &entities.OAuthAccount{
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
	}); err != nil {
		return nil, fmt.Errorf("link oauth account: %w", err)
	}

	return user, nil
}

func (s *AuthService) requireEnabledProvider(provider entities.OAuthProvider) (infraoauth.Provider, error) {
	p, err := infraoauth.GetProvider(s.oauthProviders, provider)
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

func (s *AuthService) issueAuthResult(user *entities.User) (*AuthResult, error) {
	token, expiresIn, err := appjwt.SignAccessToken(s.cfg, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        user.ToProfile(),
	}, nil
}
