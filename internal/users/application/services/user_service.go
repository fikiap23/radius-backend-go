package services

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/users/domain/entities"
	"github.com/radius/radius-backend/internal/users/domain/repositories"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo repositories.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repositories.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, logger: logger}
}

func (s *UserService) HandleGetMe(ctx context.Context, userID string) (*entities.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	profile := user.ToProfile()
	return &profile, nil
}

func (s *UserService) HandleUpdateMe(ctx context.Context, userID string, input entities.UpdateProfileInput) (*entities.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
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

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	updated, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	profile := updated.ToProfile()
	return &profile, nil
}
