package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"github.com/radius/radius-backend/internal/users/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewGormUserRepository(db *gorm.DB, logger *zap.Logger) *GormUserRepository {
	return &GormUserRepository{db: db, logger: logger}
}

var _ repositories.UserRepository = (*GormUserRepository)(nil)

func (r *GormUserRepository) Create(ctx context.Context, user *entities.User) error {
	model := toModel(user)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return toEntity(&model), nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("email = ?", email).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return toEntity(&model), nil
}

func (r *GormUserRepository) Update(ctx context.Context, user *entities.User) error {
	model := toModel(user)
	result := r.db.WithContext(ctx).
		Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(map[string]interface{}{
			"name":        model.Name,
			"avatar_url":  model.AvatarURL,
			"timezone":    model.Timezone,
			"locale":      model.Locale,
			"updated_at":  time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *GormUserRepository) UpdateLastLogin(ctx context.Context, id string, at time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"last_login_at": at,
			"updated_at":    time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("update last login: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&userModel{}).
		Where("deleted_at IS NULL").
		Where("email = ?", email).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("exists by email: %w", err)
	}
	return count > 0, nil
}

func toModel(user *entities.User) userModel {
	return userModel{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		EmailVerifiedAt: user.EmailVerifiedAt,
		AvatarURL:       user.AvatarURL,
		LastLoginAt:     user.LastLoginAt,
		Timezone:        user.Timezone,
		Locale:          user.Locale,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

func toEntity(model *userModel) *entities.User {
	return &entities.User{
		ID:              model.ID,
		Name:            model.Name,
		Email:           model.Email,
		PasswordHash:    model.PasswordHash,
		EmailVerifiedAt: model.EmailVerifiedAt,
		AvatarURL:       model.AvatarURL,
		LastLoginAt:     model.LastLoginAt,
		Timezone:        model.Timezone,
		Locale:          model.Locale,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
