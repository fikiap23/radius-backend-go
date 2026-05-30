package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/shared/config"
	appjwt "github.com/radius/radius-backend/internal/shared/jwt"
	appdto "github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"github.com/radius/radius-backend/internal/users/domain/repositories"
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
	userRepo repositories.UserRepository
	cfg      config.JWTConfig
	logger   *zap.Logger
}

func NewAuthService(userRepo repositories.UserRepository, cfg config.JWTConfig, logger *zap.Logger) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg, logger: logger}
}

func (s *AuthService) HandleRegister(ctx context.Context, in appdto.RegisterInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
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

func (s *AuthService) HandleLogin(ctx context.Context, in appdto.LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	user, err := s.userRepo.FindByEmail(ctx, email)
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
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

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
