package domain

import (
	"context"

	"github.com/radius/radius-backend/internal/projects/domain"
)

// TasksRepositories groups task-context persistence ports for one transaction.
type TasksRepositories struct {
	Tasks          TaskRepository
	Subtasks       TaskSubtaskRepository
	ChecklistItems TaskChecklistItemRepository
	Attachments    TaskAttachmentRepository
	ActivityLogs   TaskActivityLogRepository
	Projects       domain.ProjectRepository
}

// RunTasksInTransactionFunc runs fn with repos bound to a single DB transaction.
type RunTasksInTransactionFunc func(
	ctx context.Context,
	fn func(ctx context.Context, repos TasksRepositories) error,
) error
