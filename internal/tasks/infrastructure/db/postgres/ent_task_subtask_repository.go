package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	entsubtask "github.com/radius/radius-backend/ent/tasksubtask"
	"github.com/radius/radius-backend/internal/tasks/domain"
)

type TaskSubtaskRepository struct {
	client *ent.Client
}

func NewTaskSubtaskRepository(client *ent.Client) *TaskSubtaskRepository {
	return &TaskSubtaskRepository{client: client}
}

var _ domain.TaskSubtaskRepository = (*TaskSubtaskRepository)(nil)

func (r *TaskSubtaskRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskSubtask, error) {
	rows, err := r.client.TaskSubtask.Query().
		Where(entsubtask.TaskIDEQ(taskID)).
		Order(entsubtask.ByID()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find subtasks: %w", err)
	}
	return toDomainSubtasks(rows), nil
}

func (r *TaskSubtaskRepository) ReplaceAll(ctx context.Context, taskID string, items []*domain.TaskSubtask) error {
	existing, err := r.FindByTaskID(ctx, taskID)
	if err != nil {
		return err
	}

	existingByID := make(map[string]*domain.TaskSubtask, len(existing))
	for _, item := range existing {
		existingByID[item.ID] = item
	}

	keep := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.ID != "" {
			if _, ok := existingByID[item.ID]; !ok {
				return fmt.Errorf("subtask %s not found for task %s", item.ID, taskID)
			}
			keep[item.ID] = struct{}{}
			_, err := r.client.TaskSubtask.Update().
				Where(
					entsubtask.IDEQ(item.ID),
					entsubtask.TaskIDEQ(taskID),
				).
				SetTitle(item.Title).
				SetDone(item.Done).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("update subtask %s: %w", item.ID, err)
			}
			continue
		}

		id := uuid.NewString()
		item.ID = id
		item.TaskID = taskID
		keep[id] = struct{}{}
		if _, err := r.client.TaskSubtask.Create().
			SetID(id).
			SetTaskID(taskID).
			SetTitle(item.Title).
			SetDone(item.Done).
			Save(ctx); err != nil {
			return fmt.Errorf("create subtask: %w", err)
		}
	}

	for _, item := range existing {
		if _, ok := keep[item.ID]; ok {
			continue
		}
		if _, err := r.client.TaskSubtask.Delete().
			Where(
				entsubtask.IDEQ(item.ID),
				entsubtask.TaskIDEQ(taskID),
			).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete subtask %s: %w", item.ID, err)
		}
	}
	return nil
}
