package domain

import (
	"context"
	"time"
)

type TaskUpdateData struct {
	Title       *string
	Description *string
	Status      *TaskStatus
	ColumnID    **string
	Priority    *TaskPriority
	DueAt       **time.Time
	LabelIDs    *[]string
	AssigneeID  **string
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByIDWithChildren(ctx context.Context, id string) (*Task, error)
	FindManyByProjectID(ctx context.Context, projectID string) ([]*Task, error)
	UpdateByID(ctx context.Context, id string, data TaskUpdateData) error
	DeleteByID(ctx context.Context, id string) error
}

type TaskSubtaskRepository interface {
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskSubtask, error)
	ReplaceAll(ctx context.Context, taskID string, items []*TaskSubtask) error
}

type TaskChecklistItemRepository interface {
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskChecklistItem, error)
	ReplaceAll(ctx context.Context, taskID string, items []*TaskChecklistItem) error
}

type TaskAttachmentRepository interface {
	Create(ctx context.Context, attachment *TaskAttachment) error
	FindByID(ctx context.Context, taskID, attachmentID string) (*TaskAttachment, error)
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskAttachment, error)
	DeleteByID(ctx context.Context, taskID, attachmentID string) error
}

type TaskActivityLogRepository interface {
	Create(ctx context.Context, log *TaskActivityLog) error
	FindByTaskID(ctx context.Context, taskID string) ([]*TaskActivityLog, error)
}
