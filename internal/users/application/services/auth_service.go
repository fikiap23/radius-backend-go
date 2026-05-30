package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/shared/config"
	appdto "github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"github.com/radius/radius-backend/internal/users/domain/repositories"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthResult struct {
	AccessToken string                `json:"accessToken"`
	TokenType   string                `json:"tokenType"`
	ExpiresIn   int64                 `json:"expiresIn"`
	User        entities.UserProfile  `json:"user"`
}

type AuthService struct {
	users  repositories.UserRepository
	cfg    config.JWTConfig
	logger *zap.Logger
}

func NewAuthService(users repositories.UserRepository, cfg config.JWTConfig, logger *zap.Logger) *AuthService {
	return &AuthService{users: users, cfg: cfg, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, in appdto.RegisterInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	exists, err := s.users.ExistsByEmail(ctx, email)
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

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.buildAuthResult(user)
}

func (s *AuthService) Login(ctx context.Context, in appdto.LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Body.Email))

	user, err := s.users.FindByEmail(ctx, email)
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
	if err := s.users.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

	return s.buildAuthResult(user)
}

func (s *AuthService) GenerateToken(userID, email string) (string, int64, error) {
	expiresAt := time.Now().UTC().Add(s.cfg.Expiry)
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iss":   s.cfg.Issuer,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	return signed, int64(s.cfg.Expiry.Seconds()), nil
}

func (s *AuthService) buildAuthResult(user *entities.User) (*AuthResult, error) {
	token, expiresIn, err := s.GenerateToken(user.ID, user.Email)
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
