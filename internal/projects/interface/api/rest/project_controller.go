package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/projects/application/dto"
	"github.com/radius/radius-backend/internal/projects/application/services"
	"github.com/radius/radius-backend/internal/projects/domain"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/shared/middleware"
	storagedomain "github.com/radius/radius-backend/internal/storage/domain"
	"go.uber.org/zap"
)

var projectErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrProjectNotFound, Status: http.StatusNotFound, Code: "project_not_found", Message: "Project not found.", Param: "projectId"},
	{Err: domain.ErrProjectForbidden, Status: http.StatusForbidden, Code: "project_forbidden", Message: "You do not have permission to perform this action."},
	{Err: domain.ErrProjectArchived, Status: http.StatusConflict, Code: "project_already_archived", Message: "Project is already archived."},
	{Err: domain.ErrProjectNotArchived, Status: http.StatusConflict, Code: "project_not_archived", Message: "Project is not archived."},
	{Err: storagedomain.ErrTempFileNotFound, Status: http.StatusBadRequest, Code: "temp_file_not_found", Message: "Uploaded temp file not found.", Param: "coverImageTempKey"},
	{Err: storagedomain.ErrInvalidTempKey, Status: http.StatusBadRequest, Code: "invalid_temp_key", Message: "Invalid temp key.", Param: "coverImageTempKey"},
}

func RegisterProjects(api huma.API, svc *services.ProjectService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID: "projects-list",
		Method:      http.MethodGet,
		Path:        "/workspaces/{workspaceId}/projects",
		Summary:     "List projects in a workspace",
		Tags:        []string{"projects"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.WorkspaceProjectsPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := svc.HandleListProjects(ctx, userID, in.WorkspaceID)
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.OK(items), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "projects-create",
		Method:        http.MethodPost,
		Path:          "/workspaces/{workspaceId}/projects",
		Summary:       "Create a project",
		Tags:          []string{"projects"},
		Security:      humaapi.BearerSecurity(),
		Middlewares:   huma.Middlewares{authMW},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.CreateProjectInput) (*humaapi.CreatedOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleCreateProject(ctx, userID, in.WorkspaceID, in)
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.Created(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-update",
		Method:      http.MethodPatch,
		Path:        "/projects/{projectId}",
		Summary:     "Update a project",
		Tags:        []string{"projects"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.UpdateProjectInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleUpdateProject(ctx, userID, in.ProjectID, in.ToDomain())
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-toggle-favorite",
		Method:      http.MethodPatch,
		Path:        "/projects/{projectId}/favorite",
		Summary:     "Toggle project favorite status",
		Tags:        []string{"projects"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ProjectPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleToggleFavorite(ctx, userID, in.ProjectID)
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-archive",
		Method:      http.MethodPatch,
		Path:        "/projects/{projectId}/archive",
		Summary:     "Archive a project",
		Tags:        []string{"projects"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ProjectPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleArchive(ctx, userID, in.ProjectID)
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-unarchive",
		Method:      http.MethodPatch,
		Path:        "/projects/{projectId}/unarchive",
		Summary:     "Unarchive a project",
		Tags:        []string{"projects"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ProjectPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleUnarchive(ctx, userID, in.ProjectID)
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-delete",
		Method:      http.MethodDelete,
		Path:        "/projects/{projectId}",
		Summary:     "Delete a project",
		Tags:        []string{"projects"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ProjectPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleDeleteProject(ctx, userID, in.ProjectID)
		if err != nil {
			return nil, humaapi.MapError(err, projectErrors, logger)
		}
		return humaapi.OK(out), nil
	})
}
