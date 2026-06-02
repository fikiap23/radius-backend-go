package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/shared/config"
	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       domain.UserRepository
	runUsersTx     domain.RunUsersInTransactionFunc
	oauthProviders map[domain.OAuthProvider]domain.SSOProvider
	oauthCfg       config.OAuthConfig
	jwtCfg         config.JWTConfig
	logger         *zap.Logger
}

func NewAuthService(
	userRepo domain.UserRepository,
	runUsersTx domain.RunUsersInTransactionFunc,
	oauthProviders map[domain.OAuthProvider]domain.SSOProvider,
	oauthCfg config.OAuthConfig,
	jwtCfg config.JWTConfig,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		runUsersTx:     runUsersTx,
		oauthProviders: oauthProviders,
		oauthCfg:       oauthCfg,
		jwtCfg:         jwtCfg,
		logger:         logger,
	}
}

func (s *AuthService) HandleRegister(ctx context.Context, in dto.RegisterInput) (*dto.AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

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
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.PasswordHash = nil

	return s.authResultFromUser(user)
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

	return s.authResultFromUser(user)
}

func (s *AuthService) HandleGoogleSSOAuthURL(_ context.Context, in dto.GoogleSSOAuthURLInput) (*dto.SSOAuthURLResult, error) {
	return s.buildSSOAuthURL(domain.OAuthProviderGoogle, in.RedirectURI)
}

func (s *AuthService) HandleGitHubSSOAuthURL(_ context.Context, in dto.GitHubSSOAuthURLInput) (*dto.SSOAuthURLResult, error) {
	return s.buildSSOAuthURL(domain.OAuthProviderGitHub, in.RedirectURI)
}

func (s *AuthService) HandleGoogleSSOCallback(ctx context.Context, in dto.GoogleSSOCallbackInput) (*dto.AuthResult, error) {
	return s.processSSOCallback(ctx, domain.OAuthProviderGoogle, in.Body.Code, in.Body.State)
}

func (s *AuthService) HandleGitHubSSOCallback(ctx context.Context, in dto.GitHubSSOCallbackInput) (*dto.AuthResult, error) {
	return s.processSSOCallback(ctx, domain.OAuthProviderGitHub, in.Body.Code, in.Body.State)
}
