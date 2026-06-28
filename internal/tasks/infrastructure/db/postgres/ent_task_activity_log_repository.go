package postgres

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	entactivity "github.com/radius/radius-backend/ent/taskactivitylog"
	"github.com/radius/radius-backend/internal/tasks/domain"
)

type TaskActivityLogRepository struct {
	client *ent.Client
}

func NewTaskActivityLogRepository(client *ent.Client) *TaskActivityLogRepository {
	return &TaskActivityLogRepository{client: client}
}

var _ domain.TaskActivityLogRepository = (*TaskActivityLogRepository)(nil)

func (r *TaskActivityLogRepository) Create(ctx context.Context, log *domain.TaskActivityLog) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}

	created, err := r.client.TaskActivityLog.Create().
		SetID(log.ID).
		SetTaskID(log.TaskID).
		SetTitle(log.Title).
		SetNillableDescription(log.Description).
		SetIcon(log.Icon).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create task activity log: %w", err)
	}
	log.OccurredAt = created.OccurredAt
	return nil
}

func (r *TaskActivityLogRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskActivityLog, error) {
	rows, err := r.client.TaskActivityLog.Query().
		Where(entactivity.TaskIDEQ(taskID)).
		Order(entactivity.ByOccurredAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find task activity logs: %w", err)
	}
	return toDomainActivityLogs(rows), nil
}
