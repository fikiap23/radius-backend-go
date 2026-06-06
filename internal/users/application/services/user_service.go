package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/radius/radius-backend/internal/shared/pagination"
	"github.com/radius/radius-backend/internal/shared/storage"
	"github.com/radius/radius-backend/internal/storage/domain"
	"github.com/radius/radius-backend/internal/users/application/dto"
	userdomain "github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo      userdomain.UserRepository
	objectStorage domain.ObjectStorage
	logger        *zap.Logger
}

func NewUserService(
	userRepo userdomain.UserRepository,
	objectStorage domain.ObjectStorage,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		objectStorage: objectStorage,
		logger:        logger,
	}
}

func (s *UserService) HandleListUsers(ctx context.Context, params pagination.Params) (pagination.Result[dto.UserProfile], error) {
	page, err := s.userRepo.FindManyPaginate(ctx, userdomain.Query{
		Select: userdomain.FieldsProfile,
		Filter: userdomain.Filter{Search: params.Search},
	}, params)
	if err != nil {
		return pagination.Result[dto.UserProfile]{}, fmt.Errorf("list users: %w", err)
	}
	return pagination.Map(*page, dto.MapUserProfile), nil
}

func (s *UserService) HandleGetByID(ctx context.Context, id string) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, id, userdomain.FieldsProfile)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return nil, userdomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	profile := dto.MapUserProfile(user)
	return &profile, nil
}

func (s *UserService) HandleGetMe(ctx context.Context, userID string) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID, userdomain.FieldsProfile)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	profile := dto.MapUserProfile(user)
	return &profile, nil
}

func (s *UserService) HandleUpdateMe(ctx context.Context, userID string, input userdomain.ProfileUpdate) (*dto.UserProfile, error) {
	update := userdomain.Update{
		Name:     input.Name,
		Timezone: input.Timezone,
		Locale:   input.Locale,
	}

	if tempKey := storage.TrimTempKey(input.AvatarTempKey); tempKey != "" {
		publicURL, err := s.objectStorage.PromoteUserAvatar(ctx, tempKey)
		if err != nil {
			return nil, err
		}
		update.AvatarURL = &publicURL
	} else if input.AvatarURL != nil {
		update.AvatarURL = input.AvatarURL
	}

	if err := s.userRepo.UpdateByID(ctx, userID, update); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, userID, userdomain.FieldsProfile)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	profile := dto.MapUserProfile(user)
	return &profile, nil
}
