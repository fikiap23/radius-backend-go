package services

import (
	"context"
	"errors"
	"fmt"

	projectdomain "github.com/radius/radius-backend/internal/projects/domain"
	taskdomain "github.com/radius/radius-backend/internal/tasks/domain"
	wsdomain "github.com/radius/radius-backend/internal/workspaces/domain"
)

type TaskAccessValidator struct {
	taskRepo    taskdomain.TaskRepository
	projectRepo projectdomain.ProjectRepository
	memberRepo  wsdomain.WorkspaceMemberRepository
}

func NewTaskAccessValidator(
	taskRepo taskdomain.TaskRepository,
	projectRepo projectdomain.ProjectRepository,
	memberRepo wsdomain.WorkspaceMemberRepository,
) *TaskAccessValidator {
	return &TaskAccessValidator{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
	}
}

var _ taskdomain.TaskAccessChecker = (*TaskAccessValidator)(nil)

func (v *TaskAccessValidator) CanAccessTask(ctx context.Context, userID, taskID string) error {
	task, err := v.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, taskdomain.ErrTaskNotFound) {
			return taskdomain.ErrTaskNotFound
		}
		return fmt.Errorf("load task: %w", err)
	}

	project, err := v.projectRepo.FindByID(ctx, task.ProjectID)
	if err != nil {
		if errors.Is(err, projectdomain.ErrProjectNotFound) {
			return projectdomain.ErrProjectNotFound
		}
		return fmt.Errorf("load project: %w", err)
	}

	active := wsdomain.MemberStatusActive
	_, err = v.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &project.WorkspaceID,
			UserID:      &userID,
			Status:      &active,
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return projectdomain.ErrProjectForbidden
		}
		return fmt.Errorf("check workspace membership: %w", err)
	}
	return nil
}
