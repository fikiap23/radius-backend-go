package domain

import "context"

// WorkspacesRepositories groups workspace-context persistence ports for one transaction.
type WorkspacesRepositories struct {
	Workspaces       WorkspaceRepository
	WorkspaceMembers WorkspaceMemberRepository
}

// RunWorkspacesInTransactionFunc runs fn with repos bound to a single DB transaction.
type RunWorkspacesInTransactionFunc func(
	ctx context.Context,
	fn func(ctx context.Context, repos WorkspacesRepositories) error,
) error
