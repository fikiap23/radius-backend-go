package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/users/domain"
	"gorm.io/gorm"
)

type GormOAuthAccountRepository struct {
	db *gorm.DB
}

func NewGormOAuthAccountRepository(db *gorm.DB) *GormOAuthAccountRepository {
	return &GormOAuthAccountRepository{db: db}
}

var _ domain.OAuthAccountRepository = (*GormOAuthAccountRepository)(nil)

func (r *GormOAuthAccountRepository) Create(ctx context.Context, account *domain.OAuthAccount) error {
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
	provider domain.OAuthProvider,
	providerUserID string,
) (*domain.OAuthAccount, error) {
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

	return &domain.OAuthAccount{
		ID:             model.ID,
		UserID:         model.UserID,
		Provider:       domain.OAuthProvider(model.Provider),
		ProviderUserID: model.ProviderUserID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}
