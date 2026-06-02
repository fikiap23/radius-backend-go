package domain

import "context"

// UsersRepositories groups persistence ports used inside a transaction.
type UsersRepositories struct {
	Users         UserRepository
	OAuthAccounts OAuthAccountRepository
}

// UnitOfWork runs a function with repositories bound to a single transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repos UsersRepositories) error) error
}
