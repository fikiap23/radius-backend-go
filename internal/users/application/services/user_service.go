package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/radius/radius-backend/internal/shared/pagination"
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

func (s *UserService) HandleListUsers(ctx context.Context, params pagination.Params) (pagination.Result[dto.UserProfile], error) {
	page, err := s.userRepo.FindManyPaginate(ctx, domain.Query{
		Select: domain.FieldsProfile,
		Filter: domain.Filter{Search: params.Search},
	}, params)
	if err != nil {
		return pagination.Result[dto.UserProfile]{}, fmt.Errorf("list users: %w", err)
	}
	return pagination.Map(*page, dto.MapUserProfile), nil
}

func (s *UserService) HandleGetByID(ctx context.Context, id string) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, id, domain.FieldsProfile)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	profile := dto.MapUserProfile(user)
	return &profile, nil
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
