package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	entattachment "github.com/radius/radius-backend/ent/taskattachment"
	"github.com/radius/radius-backend/internal/tasks/domain"
)

type TaskAttachmentRepository struct {
	client *ent.Client
}

func NewTaskAttachmentRepository(client *ent.Client) *TaskAttachmentRepository {
	return &TaskAttachmentRepository{client: client}
}

var _ domain.TaskAttachmentRepository = (*TaskAttachmentRepository)(nil)

func (r *TaskAttachmentRepository) Create(ctx context.Context, attachment *domain.TaskAttachment) error {
	if attachment.ID == "" {
		attachment.ID = uuid.NewString()
	}

	created, err := r.client.TaskAttachment.Create().
		SetID(attachment.ID).
		SetTaskID(attachment.TaskID).
		SetName(attachment.Name).
		SetSize(attachment.Size).
		SetMimeType(attachment.MimeType).
		SetStorageKey(attachment.StorageKey).
		SetURL(attachment.URL).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create task attachment: %w", err)
	}
	attachment.UploadedAt = created.UploadedAt
	return nil
}

func (r *TaskAttachmentRepository) FindByID(ctx context.Context, taskID, attachmentID string) (*domain.TaskAttachment, error) {
	row, err := r.client.TaskAttachment.Query().
		Where(
			entattachment.IDEQ(attachmentID),
			entattachment.TaskIDEQ(taskID),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrTaskAttachmentNotFound
		}
		return nil, fmt.Errorf("find task attachment: %w", err)
	}
	return toDomainAttachment(row), nil
}

func (r *TaskAttachmentRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskAttachment, error) {
	rows, err := r.client.TaskAttachment.Query().
		Where(entattachment.TaskIDEQ(taskID)).
		Order(entattachment.ByUploadedAt()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find task attachments: %w", err)
	}
	return toDomainAttachments(rows), nil
}

func (r *TaskAttachmentRepository) DeleteByID(ctx context.Context, taskID, attachmentID string) error {
	n, err := r.client.TaskAttachment.Delete().
		Where(
			entattachment.IDEQ(attachmentID),
			entattachment.TaskIDEQ(taskID),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete task attachment: %w", err)
	}
	if n == 0 {
		return domain.ErrTaskAttachmentNotFound
	}
	return nil
}
