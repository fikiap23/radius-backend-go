package domain

import "context"

// DemoRepositories groups demo-context persistence ports for one transaction.
type DemoRepositories struct {
	// TODO: add repository fields.
	Demo DemoRepository
}

// RunDemoInTransactionFunc runs fn with repos bound to a single DB transaction.
type RunDemoInTransactionFunc func(
	ctx context.Context,
	fn func(ctx context.Context, repos DemoRepositories) error,
) error
