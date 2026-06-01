package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/radius/radius-backend/internal/users/domain"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

var _ domain.UserRepository = (*GormUserRepository)(nil)

func (r *GormUserRepository) Create(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id string, fields ...domain.Fields) (*domain.User, error) {
	f := domain.FieldsAll
	if len(fields) > 0 {
		f = fields[0]
	}
	return r.FindOne(ctx, domain.Query{
		Select: f,
		Filter: domain.Filter{ID: &id},
	})
}

func (r *GormUserRepository) FindOne(ctx context.Context, q domain.Query) (*domain.User, error) {
	var model userModel
	db := applyFilter(r.db.WithContext(ctx), q.Filter)
	db = applySelect(db, q.Select)

	if err := db.First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return toEntity(&model), nil
}

func (r *GormUserRepository) FindMany(ctx context.Context, q domain.Query) ([]*domain.User, error) {
	db := applyFilter(r.db.WithContext(ctx), q.Filter)
	db = applySelect(db, q.Select)

	var models []userModel
	if err := db.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}
	return toEntities(models), nil
}

func (r *GormUserRepository) FindManyPaginate(ctx context.Context, q domain.Query, page domain.Page) (*domain.PageResult, error) {
	base := applyFilter(r.db.WithContext(ctx).Model(&userModel{}), q.Filter)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	limit, offset := normalizePage(page)
	db := applySelect(base, q.Select)

	var models []userModel
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}

	return &domain.PageResult{
		Items: toEntities(models),
		Total: total,
	}, nil
}

func (r *GormUserRepository) UpdateByID(ctx context.Context, id string, data domain.Update) error {
	values := updateToMap(data)
	if len(values) == 0 {
		return nil
	}
	values["updated_at"] = time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(values)
	if result.Error != nil {
		return fmt.Errorf("update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *GormUserRepository) DeleteByID(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now().UTC())
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func normalizePage(page domain.Page) (int, int) {
	limit := page.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := page.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func toEntities(models []userModel) []*domain.User {
	out := make([]*domain.User, len(models))
	for i := range models {
		out[i] = toEntity(&models[i])
	}
	return out
}

func toModel(user *domain.User) userModel {
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

func toEntity(model *userModel) *domain.User {
	return &domain.User{
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
