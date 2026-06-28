package postgres

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	enttask "github.com/radius/radius-backend/ent/task"
	entattachment "github.com/radius/radius-backend/ent/taskattachment"
	entchecklist "github.com/radius/radius-backend/ent/taskchecklistitem"
	entsubtask "github.com/radius/radius-backend/ent/tasksubtask"
	"github.com/radius/radius-backend/internal/tasks/domain"
)

type TaskRepository struct {
	client *ent.Client
}

func NewTaskRepository(client *ent.Client) *TaskRepository {
	return &TaskRepository{client: client}
}

var _ domain.TaskRepository = (*TaskRepository)(nil)

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if task.ID == "" {
		task.ID = uuid.NewString()
	}

	labelIDs := task.LabelIDs
	if labelIDs == nil {
		labelIDs = []string{}
	}

	builder := r.client.Task.Create().
		SetID(task.ID).
		SetProjectID(task.ProjectID).
		SetWorkspaceID(task.WorkspaceID).
		SetTitle(task.Title).
		SetNillableDescription(task.Description).
		SetStatus(toEntStatus(task.Status)).
		SetPriority(toEntPriority(task.Priority)).
		SetLabelIds(labelIDs).
		SetNillableDueAt(task.DueAt).
		SetNillableAssigneeID(task.AssigneeID)

	if task.ColumnID != nil {
		builder = builder.SetColumnID(*task.ColumnID)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	task.CreatedAt = created.CreatedAt
	task.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	row, err := r.client.Task.Query().
		Where(enttask.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}
	return toDomainTask(row), nil
}

func (r *TaskRepository) FindByIDWithChildren(ctx context.Context, id string) (*domain.Task, error) {
	row, err := r.client.Task.Query().
		Where(enttask.IDEQ(id)).
		WithSubtasks(func(q *ent.TaskSubtaskQuery) {
			q.Order(entsubtask.ByID(sql.OrderAsc()))
		}).
		WithChecklistItems(func(q *ent.TaskChecklistItemQuery) {
			q.Order(entchecklist.ByID(sql.OrderAsc()))
		}).
		WithAttachments(func(q *ent.TaskAttachmentQuery) {
			q.Order(entattachment.ByUploadedAt(sql.OrderAsc()))
		}).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task with children: %w", err)
	}
	return toDomainTaskWithChildren(row), nil
}

func (r *TaskRepository) FindManyByProjectID(ctx context.Context, projectID string) ([]*domain.Task, error) {
	rows, err := r.client.Task.Query().
		Where(enttask.ProjectIDEQ(projectID)).
		WithSubtasks(func(q *ent.TaskSubtaskQuery) {
			q.Order(entsubtask.ByID(sql.OrderAsc()))
		}).
		WithChecklistItems(func(q *ent.TaskChecklistItemQuery) {
			q.Order(entchecklist.ByID(sql.OrderAsc()))
		}).
		WithAttachments(func(q *ent.TaskAttachmentQuery) {
			q.Order(entattachment.ByUploadedAt(sql.OrderAsc()))
		}).
		Order(enttask.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find tasks by project: %w", err)
	}
	return toDomainTasks(rows), nil
}

func (r *TaskRepository) UpdateByID(ctx context.Context, id string, data domain.TaskUpdateData) error {
	update := r.client.Task.Update().
		Where(enttask.IDEQ(id)).
		SetUpdatedAt(time.Now().UTC())

	if data.Title != nil {
		update = update.SetTitle(*data.Title)
	}
	if data.Description != nil {
		update = update.SetNillableDescription(data.Description)
	}
	if data.Status != nil {
		update = update.SetStatus(toEntStatus(*data.Status))
	}
	if data.ColumnID != nil {
		if *data.ColumnID == nil {
			update = update.ClearColumnID()
		} else {
			update = update.SetColumnID(**data.ColumnID)
		}
	}
	if data.Priority != nil {
		update = update.SetPriority(toEntPriority(*data.Priority))
	}
	if data.DueAt != nil {
		if *data.DueAt == nil {
			update = update.ClearDueAt()
		} else {
			update = update.SetDueAt(**data.DueAt)
		}
	}
	if data.LabelIDs != nil {
		labelIDs := *data.LabelIDs
		if labelIDs == nil {
			labelIDs = []string{}
		}
		update = update.SetLabelIds(labelIDs)
	}
	if data.AssigneeID != nil {
		if *data.AssigneeID == nil {
			update = update.ClearAssigneeID()
		} else {
			update = update.SetAssigneeID(**data.AssigneeID)
		}
	}

	n, err := update.Save(ctx)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if n == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepository) DeleteByID(ctx context.Context, id string) error {
	n, err := r.client.Task.Delete().
		Where(enttask.IDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if n == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}
