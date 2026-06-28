package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/projects/domain"
	storagedomain "github.com/radius/radius-backend/internal/storage/domain"
	"github.com/radius/radius-backend/internal/tasks/application/dto"
	taskdomain "github.com/radius/radius-backend/internal/tasks/domain"
	wsdomain "github.com/radius/radius-backend/internal/workspaces/domain"
)

type TaskService struct {
	taskRepo         taskdomain.TaskRepository
	subtaskRepo      taskdomain.TaskSubtaskRepository
	checklistRepo    taskdomain.TaskChecklistItemRepository
	activityRepo     taskdomain.TaskActivityLogRepository
	attachmentRepo   taskdomain.TaskAttachmentRepository
	projectRepo      domain.ProjectRepository
	columnRepo       domain.BoardColumnRepository
	memberRepo       wsdomain.WorkspaceMemberRepository
	runInTransaction taskdomain.RunTasksInTransactionFunc
	objectStorage    storagedomain.ObjectStorage
}

func NewTaskService(
	taskRepo taskdomain.TaskRepository,
	subtaskRepo taskdomain.TaskSubtaskRepository,
	checklistRepo taskdomain.TaskChecklistItemRepository,
	activityRepo taskdomain.TaskActivityLogRepository,
	attachmentRepo taskdomain.TaskAttachmentRepository,
	projectRepo domain.ProjectRepository,
	columnRepo domain.BoardColumnRepository,
	memberRepo wsdomain.WorkspaceMemberRepository,
	runInTransaction taskdomain.RunTasksInTransactionFunc,
	objectStorage storagedomain.ObjectStorage,
) *TaskService {
	return &TaskService{
		taskRepo:         taskRepo,
		subtaskRepo:      subtaskRepo,
		checklistRepo:    checklistRepo,
		activityRepo:     activityRepo,
		attachmentRepo:   attachmentRepo,
		projectRepo:      projectRepo,
		columnRepo:       columnRepo,
		memberRepo:       memberRepo,
		runInTransaction: runInTransaction,
		objectStorage:    objectStorage,
	}
}

var _ taskdomain.TaskAccessChecker = (*TaskService)(nil)

func (s *TaskService) CanAccessTask(ctx context.Context, userID, taskID string) error {
	_, err := s.requireTaskAccess(ctx, userID, taskID)
	return err
}

func (s *TaskService) HandleListTasks(ctx context.Context, userID, projectID string) ([]dto.TaskResponse, error) {
	if _, err := s.requireProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.FindManyByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return dto.MapTasks(tasks), nil
}

func (s *TaskService) HandleCreateTask(ctx context.Context, userID, projectID string, in *dto.CreateTaskInput) (*dto.TaskResponse, error) {
	project, err := s.requireProjectAccess(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	task := in.ToDomain(projectID, project.WorkspaceID)
	task.ID = uuid.NewString()

	if err := s.validateColumn(ctx, projectID, task.ColumnID); err != nil {
		return nil, err
	}
	if err := s.validateAssignee(ctx, project.WorkspaceID, task.AssigneeID); err != nil {
		return nil, err
	}

	openDelta := openTasksDeltaForCreate(task.Status)

	err = s.runInTransaction(ctx, func(ctx context.Context, repos taskdomain.TasksRepositories) error {
		if err := repos.Tasks.Create(ctx, &task); err != nil {
			return err
		}
		log := &taskdomain.TaskActivityLog{
			ID:          uuid.NewString(),
			TaskID:      task.ID,
			Title:       "Task created",
			Description: strPtr(task.Title),
			Icon:        "i-lucide-plus-circle",
		}
		if err := repos.ActivityLogs.Create(ctx, log); err != nil {
			return err
		}
		if openDelta != 0 {
			return repos.Projects.AdjustOpenTasks(ctx, projectID, openDelta)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	created, err := s.taskRepo.FindByIDWithChildren(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("load created task: %w", err)
	}
	out := dto.MapTask(created)
	return &out, nil
}

func (s *TaskService) HandleUpdateTask(ctx context.Context, userID, taskID string, in *dto.UpdateTaskInput) (*dto.TaskPatchResponse, error) {
	existing, err := s.requireTaskAccess(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	update := taskdomain.TaskUpdateData{}
	var activityLogs []*taskdomain.TaskActivityLog

	if in.Body.Title != nil {
		update.Title = in.Body.Title
	}
	if in.Body.Description != nil {
		update.Description = in.Body.Description
	}
	if in.Body.Status != nil {
		if !in.Body.Status.Valid() {
			return nil, fmt.Errorf("invalid status")
		}
		update.Status = in.Body.Status
		if *in.Body.Status != existing.Status {
			desc := fmt.Sprintf("%s -> %s", existing.Status, *in.Body.Status)
			activityLogs = append(activityLogs, newActivityLog(taskID, "Status changed", &desc, "i-lucide-arrow-right-left"))
		}
	}
	if in.Body.ColumnID != nil {
		if err := s.validateColumn(ctx, existing.ProjectID, in.Body.ColumnID); err != nil {
			return nil, err
		}
		colID := in.Body.ColumnID
		update.ColumnID = &colID
		if !ptrStringEqual(existing.ColumnID, in.Body.ColumnID) {
			activityLogs = append(activityLogs, newActivityLog(taskID, "Column changed", nil, "i-lucide-columns"))
		}
	}
	if in.Body.Priority != nil {
		if !in.Body.Priority.Valid() {
			return nil, fmt.Errorf("invalid priority")
		}
		update.Priority = in.Body.Priority
		if *in.Body.Priority != existing.Priority {
			desc := fmt.Sprintf("%s -> %s", existing.Priority, *in.Body.Priority)
			activityLogs = append(activityLogs, newActivityLog(taskID, "Priority updated", &desc, "i-lucide-flag"))
		}
	}
	if in.Body.DueAt != nil {
		dueAt := in.Body.DueAt
		update.DueAt = &dueAt
	}
	if in.Body.LabelIDs != nil {
		update.LabelIDs = in.Body.LabelIDs
	}
	if in.Body.AssigneeID != nil {
		if err := s.validateAssignee(ctx, existing.WorkspaceID, in.Body.AssigneeID); err != nil {
			return nil, err
		}
		assigneeID := in.Body.AssigneeID
		update.AssigneeID = &assigneeID
		if !ptrStringEqual(existing.AssigneeID, in.Body.AssigneeID) {
			activityLogs = append(activityLogs, newActivityLog(taskID, "Assignee updated", nil, "i-lucide-user"))
		}
	}

	openDelta := openTasksDeltaForStatusChange(existing.Status, in.Body.Status)

	var subtasks []*taskdomain.TaskSubtask
	if in.Body.Subtasks != nil {
		subtasks = make([]*taskdomain.TaskSubtask, len(*in.Body.Subtasks))
		for i, item := range *in.Body.Subtasks {
			subtasks[i] = &taskdomain.TaskSubtask{
				ID:    strings.TrimSpace(item.ID),
				Title: strings.TrimSpace(item.Title),
				Done:  item.Done,
			}
		}
	}

	var checklist []*taskdomain.TaskChecklistItem
	if in.Body.Checklist != nil {
		checklist = make([]*taskdomain.TaskChecklistItem, len(*in.Body.Checklist))
		for i, item := range *in.Body.Checklist {
			checklist[i] = &taskdomain.TaskChecklistItem{
				ID:      strings.TrimSpace(item.ID),
				Text:    strings.TrimSpace(item.Text),
				Checked: item.Checked,
			}
		}
	}

	err = s.runInTransaction(ctx, func(ctx context.Context, repos taskdomain.TasksRepositories) error {
		if hasTaskFieldUpdate(update) {
			if err := repos.Tasks.UpdateByID(ctx, taskID, update); err != nil {
				return err
			}
		}
		if in.Body.Subtasks != nil {
			if err := repos.Subtasks.ReplaceAll(ctx, taskID, subtasks); err != nil {
				return err
			}
		}
		if in.Body.Checklist != nil {
			if err := repos.ChecklistItems.ReplaceAll(ctx, taskID, checklist); err != nil {
				return err
			}
		}
		for _, log := range activityLogs {
			if err := repos.ActivityLogs.Create(ctx, log); err != nil {
				return err
			}
		}
		if openDelta != 0 {
			return repos.Projects.AdjustOpenTasks(ctx, existing.ProjectID, openDelta)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	out := &dto.TaskPatchResponse{ID: taskID}
	if in.Body.Status != nil {
		out.Status = in.Body.Status
	}
	if in.Body.ColumnID != nil {
		out.ColumnID = in.Body.ColumnID
	}
	if in.Body.Subtasks != nil {
		rows, err := s.subtaskRepo.FindByTaskID(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("load subtasks: %w", err)
		}
		out.Subtasks = dto.MapSubtasks(rows)
	}
	if in.Body.Checklist != nil {
		rows, err := s.checklistRepo.FindByTaskID(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("load checklist: %w", err)
		}
		out.Checklist = dto.MapChecklistItems(rows)
	}
	return out, nil
}

func (s *TaskService) HandleDeleteTask(ctx context.Context, userID, taskID string) (*dto.OkResponse, error) {
	existing, err := s.requireTaskAccess(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	attachments, err := s.attachmentRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("load attachments: %w", err)
	}

	openDelta := openTasksDeltaForDelete(existing.Status)

	err = s.runInTransaction(ctx, func(ctx context.Context, repos taskdomain.TasksRepositories) error {
		if err := repos.Tasks.DeleteByID(ctx, taskID); err != nil {
			return err
		}
		if openDelta != 0 {
			return repos.Projects.AdjustOpenTasks(ctx, existing.ProjectID, openDelta)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("delete task: %w", err)
	}

	for _, att := range attachments {
		if s.objectStorage != nil {
			_ = s.objectStorage.DeleteObject(ctx, att.StorageKey)
		}
	}

	return &dto.OkResponse{OK: true}, nil
}

func (s *TaskService) requireProjectAccess(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	active := wsdomain.MemberStatusActive
	_, err = s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &project.WorkspaceID,
			UserID:      &userID,
			Status:      &active,
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return nil, domain.ErrProjectForbidden
		}
		return nil, fmt.Errorf("check workspace membership: %w", err)
	}
	return project, nil
}

func (s *TaskService) requireTaskAccess(ctx context.Context, userID, taskID string) (*taskdomain.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, taskdomain.ErrTaskNotFound) {
			return nil, taskdomain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("load task: %w", err)
	}
	if _, err := s.requireProjectAccess(ctx, userID, task.ProjectID); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) validateColumn(ctx context.Context, projectID string, columnID *string) error {
	if columnID == nil || strings.TrimSpace(*columnID) == "" {
		return nil
	}
	_, err := s.columnRepo.FindByID(ctx, projectID, *columnID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardColumnNotFound) {
			return taskdomain.ErrTaskInvalidColumn
		}
		return fmt.Errorf("validate column: %w", err)
	}
	return nil
}

func (s *TaskService) validateAssignee(ctx context.Context, workspaceID string, assigneeID *string) error {
	if assigneeID == nil || strings.TrimSpace(*assigneeID) == "" {
		return nil
	}
	active := wsdomain.MemberStatusActive
	_, err := s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &workspaceID,
			UserID:      assigneeID,
			Status:      &active,
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return taskdomain.ErrTaskInvalidAssignee
		}
		return fmt.Errorf("validate assignee: %w", err)
	}
	return nil
}

func openTasksDeltaForCreate(status taskdomain.TaskStatus) int {
	if status.IsDone() {
		return 0
	}
	return 1
}

func openTasksDeltaForDelete(status taskdomain.TaskStatus) int {
	if status.IsDone() {
		return 0
	}
	return -1
}

func openTasksDeltaForStatusChange(oldStatus taskdomain.TaskStatus, newStatus *taskdomain.TaskStatus) int {
	if newStatus == nil {
		return 0
	}
	if oldStatus.IsDone() && !newStatus.IsDone() {
		return 1
	}
	if !oldStatus.IsDone() && newStatus.IsDone() {
		return -1
	}
	return 0
}

func hasTaskFieldUpdate(data taskdomain.TaskUpdateData) bool {
	return data.Title != nil || data.Description != nil || data.Status != nil ||
		data.ColumnID != nil || data.Priority != nil || data.DueAt != nil ||
		data.LabelIDs != nil || data.AssigneeID != nil
}

func newActivityLog(taskID, title string, description *string, icon string) *taskdomain.TaskActivityLog {
	return &taskdomain.TaskActivityLog{
		ID:          uuid.NewString(),
		TaskID:      taskID,
		Title:       title,
		Description: description,
		Icon:        icon,
	}
}

func strPtr(s string) *string {
	return &s
}

func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
