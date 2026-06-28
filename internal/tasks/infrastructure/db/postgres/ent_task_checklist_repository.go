package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	entchecklist "github.com/radius/radius-backend/ent/taskchecklistitem"
	"github.com/radius/radius-backend/internal/tasks/domain"
)

type TaskChecklistItemRepository struct {
	client *ent.Client
}

func NewTaskChecklistItemRepository(client *ent.Client) *TaskChecklistItemRepository {
	return &TaskChecklistItemRepository{client: client}
}

var _ domain.TaskChecklistItemRepository = (*TaskChecklistItemRepository)(nil)

func (r *TaskChecklistItemRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskChecklistItem, error) {
	rows, err := r.client.TaskChecklistItem.Query().
		Where(entchecklist.TaskIDEQ(taskID)).
		Order(entchecklist.ByID()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find checklist items: %w", err)
	}
	return toDomainChecklistItems(rows), nil
}

func (r *TaskChecklistItemRepository) ReplaceAll(ctx context.Context, taskID string, items []*domain.TaskChecklistItem) error {
	existing, err := r.FindByTaskID(ctx, taskID)
	if err != nil {
		return err
	}

	existingByID := make(map[string]*domain.TaskChecklistItem, len(existing))
	for _, item := range existing {
		existingByID[item.ID] = item
	}

	keep := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.ID != "" {
			if _, ok := existingByID[item.ID]; !ok {
				return fmt.Errorf("checklist item %s not found for task %s", item.ID, taskID)
			}
			keep[item.ID] = struct{}{}
			_, err := r.client.TaskChecklistItem.Update().
				Where(
					entchecklist.IDEQ(item.ID),
					entchecklist.TaskIDEQ(taskID),
				).
				SetText(item.Text).
				SetChecked(item.Checked).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("update checklist item %s: %w", item.ID, err)
			}
			continue
		}

		id := uuid.NewString()
		item.ID = id
		item.TaskID = taskID
		keep[id] = struct{}{}
		if _, err := r.client.TaskChecklistItem.Create().
			SetID(id).
			SetTaskID(taskID).
			SetText(item.Text).
			SetChecked(item.Checked).
			Save(ctx); err != nil {
			return fmt.Errorf("create checklist item: %w", err)
		}
	}

	for _, item := range existing {
		if _, ok := keep[item.ID]; ok {
			continue
		}
		if _, err := r.client.TaskChecklistItem.Delete().
			Where(
				entchecklist.IDEQ(item.ID),
				entchecklist.TaskIDEQ(taskID),
			).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete checklist item %s: %w", item.ID, err)
		}
	}
	return nil
}
