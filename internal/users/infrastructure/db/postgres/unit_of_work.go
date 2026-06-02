package postgres

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/users/domain"
)

type UnitOfWork struct {
	client *ent.Client
}

func NewUnitOfWork(client *ent.Client) *UnitOfWork {
	return &UnitOfWork{client: client}
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repos domain.UsersRepositories) error) error {
	tx, err := u.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	repos := domain.UsersRepositories{
		Users:         NewUserRepository(tx.Client()),
		OAuthAccounts: NewOAuthAccountRepository(tx.Client()),
	}

	if err := fn(ctx, repos); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rollback after error: %v (rollback: %w)", err, rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
