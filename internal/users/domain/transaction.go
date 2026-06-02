package domain

import "context"

// UsersRepositories groups user-context persistence ports for one transaction.
type UsersRepositories struct {
	Users         UserRepository
	OAuthAccounts OAuthAccountRepository
}

// RunUsersInTransactionFunc runs fn with repos bound to a single DB transaction.
type RunUsersInTransactionFunc func(
	ctx context.Context,
	fn func(ctx context.Context, repos UsersRepositories) error,
) error
