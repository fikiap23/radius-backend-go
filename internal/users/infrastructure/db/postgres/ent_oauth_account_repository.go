package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	entoauth "github.com/radius/radius-backend/ent/useroauthaccount"
	"github.com/radius/radius-backend/internal/users/domain"
)

type OAuthAccountRepository struct {
	client *ent.Client
}

func NewOAuthAccountRepository(client *ent.Client) *OAuthAccountRepository {
	return &OAuthAccountRepository{client: client}
}

var _ domain.OAuthAccountRepository = (*OAuthAccountRepository)(nil)

func (r *OAuthAccountRepository) Create(ctx context.Context, account *domain.OAuthAccount) error {
	if account.ID == "" {
		account.ID = uuid.NewString()
	}

	created, err := r.client.UserOAuthAccount.Create().
		SetID(account.ID).
		SetUserID(account.UserID).
		SetProvider(string(account.Provider)).
		SetProviderUserID(account.ProviderUserID).
		Save(ctx)
	if err != nil {
		if mapped := mapOAuthAccountCreateError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("create oauth account: %w", err)
	}

	account.CreatedAt = created.CreatedAt
	account.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *OAuthAccountRepository) FindByProviderAccount(
	ctx context.Context,
	provider domain.OAuthProvider,
	providerUserID string,
) (*domain.OAuthAccount, error) {
	row, err := r.client.UserOAuthAccount.Query().
		Where(
			entoauth.ProviderEQ(string(provider)),
			entoauth.ProviderUserIDEQ(providerUserID),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrOAuthAccountNotFound
		}
		return nil, fmt.Errorf("find oauth account: %w", err)
	}

	return &domain.OAuthAccount{
		ID:             row.ID,
		UserID:         row.UserID,
		Provider:       domain.OAuthProvider(row.Provider),
		ProviderUserID: row.ProviderUserID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}
