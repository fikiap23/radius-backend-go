package postgres

import (
	"context"
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

func (r *GormUserRepository) FindByID(ctx context.Context, id string, fields ...repositories.UserFields) (*entities.User, error) {
	f := repositories.SelectAll
	if len(fields) > 0 {
		f = fields[0]
	}
	return r.FindOne(ctx, repositories.Query{
		Select: f,
		Where:  repositories.Where{ID: &id},
	})
}

func (r *GormUserRepository) FindOne(ctx context.Context, q repositories.Query) (*entities.User, error) {
	var model userModel
	db := applyWhere(r.db.WithContext(ctx), q.Where)
	db = applySelect(db, q.Fields())

	if err := db.First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return toEntity(&model), nil
}

func (r *GormUserRepository) FindMany(ctx context.Context, q repositories.Query) ([]*entities.User, error) {
	db := applyWhere(r.db.WithContext(ctx), q.Where)
	db = applySelect(db, q.Fields())

	var models []userModel
	if err := db.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}
	return toEntities(models), nil
}

func (r *GormUserRepository) FindManyPaginate(ctx context.Context, q repositories.Query, page repositories.Page) (*repositories.PageResult, error) {
	base := applyWhere(r.db.WithContext(ctx).Model(&userModel{}), q.Where)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	limit, offset := normalizePage(page)
	db := applySelect(base, q.Fields())

	var models []userModel
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}

	return &repositories.PageResult{
		Items: toEntities(models),
		Total: total,
	}, nil
}

func (r *GormUserRepository) UpdateByID(ctx context.Context, id string, data repositories.Update) error {
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

// --- helpers ---

func normalizePage(page repositories.Page) (int, int) {
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

func toEntities(models []userModel) []*entities.User {
	out := make([]*entities.User, len(models))
	for i := range models {
		out[i] = toEntity(&models[i])
	}
	return out
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
