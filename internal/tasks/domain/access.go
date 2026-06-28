package domain

import "context"

// TaskAccessChecker validates whether a user may access a task.
type TaskAccessChecker interface {
	CanAccessTask(ctx context.Context, userID, taskID string) error
}
