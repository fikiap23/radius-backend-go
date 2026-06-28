package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	storagedomain "github.com/radius/radius-backend/internal/storage/domain"
	"github.com/radius/radius-backend/internal/tasks/application/dto"
	taskdomain "github.com/radius/radius-backend/internal/tasks/domain"
)

type AttachmentService struct {
	taskRepo       taskdomain.TaskRepository
	attachmentRepo taskdomain.TaskAttachmentRepository
	activityRepo   taskdomain.TaskActivityLogRepository
	objectStorage  storagedomain.ObjectStorage
	access         *TaskService
}

func NewAttachmentService(
	taskRepo taskdomain.TaskRepository,
	attachmentRepo taskdomain.TaskAttachmentRepository,
	activityRepo taskdomain.TaskActivityLogRepository,
	objectStorage storagedomain.ObjectStorage,
	access *TaskService,
) *AttachmentService {
	return &AttachmentService{
		taskRepo:       taskRepo,
		attachmentRepo: attachmentRepo,
		activityRepo:   activityRepo,
		objectStorage:  objectStorage,
		access:         access,
	}
}

func (s *AttachmentService) HandleConfirmAttachment(
	ctx context.Context,
	userID, taskID string,
	in *dto.ConfirmAttachmentInput,
) (*dto.AttachmentResponse, error) {
	if _, err := s.access.requireTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}

	attachmentID := uuid.NewString()
	storageKey, publicURL, err := s.objectStorage.PromoteTaskAttachment(
		ctx,
		in.Body.TempKey,
		taskID,
		attachmentID,
		in.Body.FileName,
	)
	if err != nil {
		return nil, err
	}

	attachment := &taskdomain.TaskAttachment{
		ID:         attachmentID,
		TaskID:     taskID,
		Name:       in.Body.FileName,
		Size:       in.Body.Size,
		MimeType:   in.Body.ContentType,
		StorageKey: storageKey,
		URL:        publicURL,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		return nil, fmt.Errorf("save attachment: %w", err)
	}

	desc := in.Body.FileName
	_ = s.activityRepo.Create(ctx, &taskdomain.TaskActivityLog{
		ID:          uuid.NewString(),
		TaskID:      taskID,
		Title:       "Attachment added",
		Description: &desc,
		Icon:        "i-lucide-paperclip",
	})

	out := dto.MapAttachment(attachment)
	return &out, nil
}

func (s *AttachmentService) HandleDeleteAttachment(
	ctx context.Context,
	userID, taskID, attachmentID string,
) (*dto.OkResponse, error) {
	if _, err := s.access.requireTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}

	attachment, err := s.attachmentRepo.FindByID(ctx, taskID, attachmentID)
	if err != nil {
		return nil, err
	}

	if err := s.attachmentRepo.DeleteByID(ctx, taskID, attachmentID); err != nil {
		return nil, fmt.Errorf("delete attachment: %w", err)
	}

	if err := s.objectStorage.DeleteObject(ctx, attachment.StorageKey); err != nil {
		return nil, fmt.Errorf("delete attachment object: %w", err)
	}

	return &dto.OkResponse{OK: true}, nil
}
