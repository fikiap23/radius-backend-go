package services

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/users/domain/entities"
	repo "github.com/radius/radius-backend/internal/users/domain/repositories"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo repo.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repo.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, logger: logger}
}

func (s *UserService) HandleGetMe(ctx context.Context, userID string) (*entities.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID, repo.SelectProfile)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	profile := user.ToProfile()
	return &profile, nil
}

func (s *UserService) HandleUpdateMe(ctx context.Context, userID string, input entities.UpdateProfileInput) (*entities.UserProfile, error) {
	if err := s.userRepo.UpdateByID(ctx, userID, repo.Update{
		Name:      input.Name,
		AvatarURL: input.AvatarURL,
		Timezone:  input.Timezone,
		Locale:    input.Locale,
	}); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, userID, repo.SelectProfile)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	profile := user.ToProfile()
	return &profile, nil
}
