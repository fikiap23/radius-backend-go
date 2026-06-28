package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	projectdomain "github.com/radius/radius-backend/internal/projects/domain"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/shared/middleware"
	storagedomain "github.com/radius/radius-backend/internal/storage/domain"
	"github.com/radius/radius-backend/internal/tasks/application/dto"
	"github.com/radius/radius-backend/internal/tasks/application/services"
	"github.com/radius/radius-backend/internal/tasks/domain"
	"go.uber.org/zap"
)

var taskErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrTaskNotFound, Status: http.StatusNotFound, Code: "task_not_found", Message: "Task not found.", Param: "taskId"},
	{Err: domain.ErrTaskForbidden, Status: http.StatusForbidden, Code: "task_forbidden", Message: "You do not have access to this task."},
	{Err: domain.ErrTaskInvalidColumn, Status: http.StatusBadRequest, Code: "task_invalid_column", Message: "Column does not belong to this project.", Param: "columnId"},
	{Err: domain.ErrTaskInvalidAssignee, Status: http.StatusBadRequest, Code: "task_invalid_assignee", Message: "Assignee is not a workspace member.", Param: "assigneeId"},
	{Err: projectdomain.ErrProjectNotFound, Status: http.StatusNotFound, Code: "project_not_found", Message: "Project not found.", Param: "projectId"},
	{Err: projectdomain.ErrProjectForbidden, Status: http.StatusForbidden, Code: "project_forbidden", Message: "You do not have access to this project."},
}

func RegisterTasks(api huma.API, svc *services.TaskService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID: "projects-tasks-list",
		Method:      http.MethodGet,
		Path:        "/projects/{projectId}/tasks",
		Summary:     "List project tasks",
		Tags:        []string{"tasks"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ProjectTasksPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := svc.HandleListTasks(ctx, userID, in.ProjectID)
		if err != nil {
			return nil, humaapi.MapError(err, taskErrors, logger)
		}
		return humaapi.OK(items), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "projects-tasks-create",
		Method:        http.MethodPost,
		Path:          "/projects/{projectId}/tasks",
		Summary:       "Create a task",
		Tags:          []string{"tasks"},
		Security:      humaapi.BearerSecurity(),
		Middlewares:   huma.Middlewares{authMW},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.CreateTaskInput) (*humaapi.CreatedOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := svc.HandleCreateTask(ctx, userID, in.ProjectID, in)
		if err != nil {
			return nil, humaapi.MapError(err, taskErrors, logger)
		}
		return humaapi.Created(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "tasks-update",
		Method:      http.MethodPatch,
		Path:        "/tasks/{taskId}",
		Summary:     "Update a task",
		Tags:        []string{"tasks"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.UpdateTaskInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := svc.HandleUpdateTask(ctx, userID, in.TaskID, in)
		if err != nil {
			return nil, humaapi.MapError(err, taskErrors, logger)
		}
		return humaapi.OK(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "tasks-delete",
		Method:      http.MethodDelete,
		Path:        "/tasks/{taskId}",
		Summary:     "Delete a task",
		Tags:        []string{"tasks"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.TaskPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := svc.HandleDeleteTask(ctx, userID, in.TaskID)
		if err != nil {
			return nil, humaapi.MapError(err, taskErrors, logger)
		}
		return humaapi.OK(result), nil
	})
}

func RegisterActivities(api huma.API, svc *services.ActivityService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID: "tasks-activities-list",
		Method:      http.MethodGet,
		Path:        "/tasks/{taskId}/activities",
		Summary:     "List task activity history",
		Tags:        []string{"tasks"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.TaskPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := svc.HandleListActivities(ctx, userID, in.TaskID)
		if err != nil {
			return nil, humaapi.MapError(err, taskErrors, logger)
		}
		return humaapi.OK(items), nil
	})
}

func RegisterAttachments(api huma.API, svc *services.AttachmentService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)
	errors := append(taskErrors, []humaapi.ErrorMapping{
		{Err: domain.ErrTaskAttachmentNotFound, Status: http.StatusNotFound, Code: "task_attachment_not_found", Message: "Attachment not found.", Param: "attachmentId"},
		{Err: storagedomain.ErrTempFileNotFound, Status: http.StatusBadRequest, Code: "temp_file_not_found", Message: "Uploaded temp file not found.", Param: "tempKey"},
		{Err: storagedomain.ErrInvalidTempKey, Status: http.StatusBadRequest, Code: "invalid_temp_key", Message: "Invalid temp key.", Param: "tempKey"},
	}...)

	huma.Register(api, huma.Operation{
		OperationID:   "tasks-attachments-create",
		Method:        http.MethodPost,
		Path:          "/tasks/{taskId}/attachments",
		Summary:       "Confirm task attachment after presigned upload",
		Tags:          []string{"tasks"},
		Security:      humaapi.BearerSecurity(),
		Middlewares:   huma.Middlewares{authMW},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.ConfirmAttachmentInput) (*humaapi.CreatedOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := svc.HandleConfirmAttachment(ctx, userID, in.TaskID, in)
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.Created(result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "tasks-attachments-delete",
		Method:      http.MethodDelete,
		Path:        "/tasks/{taskId}/attachments/{attachmentId}",
		Summary:     "Delete a task attachment",
		Tags:        []string{"tasks"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.DeleteAttachmentPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := svc.HandleDeleteAttachment(ctx, userID, in.TaskID, in.AttachmentID)
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.OK(result), nil
	})
}
