package services

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/internal/tasks/application/dto"
	taskdomain "github.com/radius/radius-backend/internal/tasks/domain"
)

type ActivityService struct {
	taskRepo     taskdomain.TaskRepository
	activityRepo taskdomain.TaskActivityLogRepository
	access       *TaskService
}

func NewActivityService(
	taskRepo taskdomain.TaskRepository,
	activityRepo taskdomain.TaskActivityLogRepository,
	access *TaskService,
) *ActivityService {
	return &ActivityService{
		taskRepo:     taskRepo,
		activityRepo: activityRepo,
		access:       access,
	}
}

func (s *ActivityService) HandleListActivities(ctx context.Context, userID, taskID string) ([]dto.ActivityResponse, error) {
	if _, err := s.access.requireTaskAccess(ctx, userID, taskID); err != nil {
		return nil, err
	}

	logs, err := s.activityRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}
	return dto.MapActivities(logs), nil
}
