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
	"go.uber.org/zap"
)

var boardColumnErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrBoardColumnNotFound, Status: http.StatusNotFound, Code: "board_column_not_found", Message: "Board column not found.", Param: "columnId"},
	{Err: domain.ErrBoardColumnStatusExists, Status: http.StatusConflict, Code: "board_column_status_exists", Message: "A column with this status already exists.", Param: "status"},
	{Err: domain.ErrBoardColumnLastColumn, Status: http.StatusConflict, Code: "board_column_last_column", Message: "Cannot delete the last board column."},
	{Err: domain.ErrBoardColumnInvalidReorder, Status: http.StatusBadRequest, Code: "board_column_invalid_reorder", Message: "Invalid column reorder request.", Param: "columnIds"},
}

func RegisterBoardColumns(api huma.API, svc *services.BoardColumnService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)
	errors := append(projectErrors, boardColumnErrors...)

	huma.Register(api, huma.Operation{
		OperationID: "projects-board-columns-list",
		Method:      http.MethodGet,
		Path:        "/projects/{projectId}/board/columns",
		Summary:     "List kanban board columns",
		Tags:        []string{"projects", "board"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ProjectBoardColumnsPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := svc.HandleListColumns(ctx, userID, in.ProjectID)
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.OK(items), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "projects-board-columns-create",
		Method:        http.MethodPost,
		Path:          "/projects/{projectId}/board/columns",
		Summary:       "Create a kanban board column",
		Tags:          []string{"projects", "board"},
		Security:      humaapi.BearerSecurity(),
		Middlewares:   huma.Middlewares{authMW},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.CreateBoardColumnInput) (*humaapi.CreatedOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleCreateColumn(ctx, userID, in.ProjectID, in)
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.Created(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-board-columns-reorder",
		Method:      http.MethodPut,
		Path:        "/projects/{projectId}/board/columns/reorder",
		Summary:     "Reorder kanban board columns",
		Tags:        []string{"projects", "board"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ReorderBoardColumnsInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleReorderColumns(ctx, userID, in.ProjectID, in.Body.ColumnIDs)
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-board-columns-update",
		Method:      http.MethodPatch,
		Path:        "/projects/{projectId}/board/columns/{columnId}",
		Summary:     "Update a kanban board column",
		Tags:        []string{"projects", "board"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.UpdateBoardColumnInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleUpdateColumn(ctx, userID, in.ProjectID, in.ColumnID, in.ToDomain())
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "projects-board-columns-delete",
		Method:      http.MethodDelete,
		Path:        "/projects/{projectId}/board/columns/{columnId}",
		Summary:     "Delete a kanban board column",
		Tags:        []string{"projects", "board"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.BoardColumnPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleDeleteColumn(ctx, userID, in.ProjectID, in.ColumnID)
		if err != nil {
			return nil, humaapi.MapError(err, errors, logger)
		}
		return humaapi.OK(out), nil
	})
}
