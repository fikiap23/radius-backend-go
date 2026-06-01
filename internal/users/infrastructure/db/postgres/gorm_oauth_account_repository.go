package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"github.com/radius/radius-backend/internal/users/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GormOAuthAccountRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewGormOAuthAccountRepository(db *gorm.DB, logger *zap.Logger) *GormOAuthAccountRepository {
	return &GormOAuthAccountRepository{db: db, logger: logger}
}

var _ repositories.OAuthAccountRepository = (*GormOAuthAccountRepository)(nil)

func (r *GormOAuthAccountRepository) Create(ctx context.Context, account *entities.OAuthAccount) error {
	if account.ID == "" {
		account.ID = uuid.NewString()
	}

	model := oauthAccountModel{
		ID:             account.ID,
		UserID:         account.UserID,
		Provider:       string(account.Provider),
		ProviderUserID: account.ProviderUserID,
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("create oauth account: %w", err)
	}

	account.CreatedAt = model.CreatedAt
	account.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *GormOAuthAccountRepository) FindByProviderAccount(
	ctx context.Context,
	provider entities.OAuthProvider,
	providerUserID string,
) (*entities.OAuthAccount, error) {
	var model oauthAccountModel
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", string(provider), providerUserID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find oauth account: %w", err)
	}

	return &entities.OAuthAccount{
		ID:             model.ID,
		UserID:         model.UserID,
		Provider:       entities.OAuthProvider(model.Provider),
		ProviderUserID: model.ProviderUserID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}
