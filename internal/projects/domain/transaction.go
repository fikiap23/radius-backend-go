package domain

import "context"

// ProjectsRepositories groups project-context persistence ports for one transaction.
type ProjectsRepositories struct {
	BoardColumns BoardColumnRepository
	Projects     ProjectRepository
}

// RunProjectsInTransactionFunc runs fn with repos bound to a single DB transaction.
type RunProjectsInTransactionFunc func(
	ctx context.Context,
	fn func(ctx context.Context, repos ProjectsRepositories) error,
) error
