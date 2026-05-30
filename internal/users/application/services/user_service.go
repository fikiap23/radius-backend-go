package services

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"github.com/radius/radius-backend/internal/users/domain/repositories"
	"go.uber.org/zap"
)

type UserService struct {
	users  repositories.UserRepository
	logger *zap.Logger
}

func NewUserService(users repositories.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{users: users, logger: logger}
}

func (s *UserService) GetMe(ctx context.Context, userID string) (*entities.UserProfile, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	profile := user.ToProfile()
	return &profile, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID string, input entities.UpdateProfileInput) (*entities.UserProfile, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	if input.Timezone != nil {
		user.Timezone = input.Timezone
	}
	if input.Locale != nil {
		user.Locale = *input.Locale
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	updated, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	profile := updated.ToProfile()
	return &profile, nil
}

func (s *UserService) MapError(err error) error {
	if err == domain.ErrUserNotFound {
		return domain.ErrUserNotFound
	}
	return err
}
