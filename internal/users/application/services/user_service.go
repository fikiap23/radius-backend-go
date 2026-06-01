package services

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo domain.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo domain.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, logger: logger}
}

func (s *UserService) HandleGetMe(ctx context.Context, userID string) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID, domain.FieldsProfile)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	profile := dto.MapUserProfile(user)
	return &profile, nil
}

func (s *UserService) HandleUpdateMe(ctx context.Context, userID string, input domain.ProfileUpdate) (*dto.UserProfile, error) {
	if err := s.userRepo.UpdateByID(ctx, userID, domain.Update{
		Name:      input.Name,
		AvatarURL: input.AvatarURL,
		Timezone:  input.Timezone,
		Locale:    input.Locale,
	}); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, userID, domain.FieldsProfile)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	profile := dto.MapUserProfile(user)
	return &profile, nil
}
