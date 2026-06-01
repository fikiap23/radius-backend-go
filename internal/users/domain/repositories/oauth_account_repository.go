package repositories

import (
	"context"

	"github.com/radius/radius-backend/internal/users/domain/entities"
)

type OAuthAccountRepository interface {
	Create(ctx context.Context, account *entities.OAuthAccount) error
	FindByProviderAccount(ctx context.Context, provider entities.OAuthProvider, providerUserID string) (*entities.OAuthAccount, error)
}
