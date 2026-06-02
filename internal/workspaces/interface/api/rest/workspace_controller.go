package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/shared/middleware"
	userdomain "github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/workspaces/application/dto"
	"github.com/radius/radius-backend/internal/workspaces/application/services"
	"github.com/radius/radius-backend/internal/workspaces/domain"
	"go.uber.org/zap"
)

var workspaceErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrWorkspaceNotFound, Status: http.StatusNotFound, Code: "workspace_not_found", Message: "Workspace not found.", Param: "workspaceId"},
	{Err: domain.ErrWorkspaceSlugAlreadyExists, Status: http.StatusConflict, Code: "workspace_slug_already_exists", Message: "A workspace with this slug already exists.", Param: "slug"},
	{Err: domain.ErrWorkspaceForbidden, Status: http.StatusForbidden, Code: "workspace_forbidden", Message: "You do not have permission to perform this action on the workspace."},
	{Err: domain.ErrWorkspaceMemberNotFound, Status: http.StatusNotFound, Code: "workspace_member_not_found", Message: "Workspace member not found.", Param: "memberId"},
	{Err: domain.ErrWorkspaceMemberAlreadyExists, Status: http.StatusConflict, Code: "workspace_member_already_exists", Message: "This email is already a member of the workspace.", Param: "email"},
	{Err: domain.ErrInvalidMemberRole, Status: http.StatusBadRequest, Code: "invalid_member_role", Message: "The member role is invalid.", Param: "role"},
	{Err: userdomain.ErrUserNotFound, Status: http.StatusNotFound, Code: "user_not_found", Message: "User not found."},
}

func RegisterWorkspaces(api huma.API, svc *services.WorkspaceService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID: "workspaces-list",
		Method:      http.MethodGet,
		Path:        "/workspaces",
		Summary:     "List workspaces for the current user",
		Tags:        []string{"workspaces"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, _ *struct{}) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := svc.HandleListWorkspaces(ctx, userID)
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.OK(items), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "workspaces-create",
		Method:        http.MethodPost,
		Path:          "/workspaces",
		Summary:       "Create a workspace",
		Tags:          []string{"workspaces"},
		Security:      humaapi.BearerSecurity(),
		Middlewares:   huma.Middlewares{authMW},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.CreateWorkspaceInput) (*humaapi.CreatedOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleCreateWorkspace(ctx, userID, in.Body.Name, in.Body.Slug)
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.Created(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "workspaces-update",
		Method:      http.MethodPatch,
		Path:        "/workspaces/{workspaceId}",
		Summary:     "Update a workspace",
		Tags:        []string{"workspaces"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.UpdateWorkspaceInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleUpdateWorkspace(ctx, userID, in.WorkspaceID, in.ToDomain())
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "workspaces-members-list",
		Method:      http.MethodGet,
		Path:        "/workspaces/{workspaceId}/members",
		Summary:     "List workspace members",
		Tags:        []string{"workspaces"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.WorkspacePathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := svc.HandleListMembers(ctx, userID, in.WorkspaceID)
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.OK(items), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "workspaces-members-create",
		Method:        http.MethodPost,
		Path:          "/workspaces/{workspaceId}/members",
		Summary:       "Invite a workspace member",
		Tags:          []string{"workspaces"},
		Security:      humaapi.BearerSecurity(),
		Middlewares:   huma.Middlewares{authMW},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *dto.InviteMemberInput) (*humaapi.CreatedOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleInviteMember(ctx, userID, in.WorkspaceID, in.Body.Email, in.Body.Role)
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.Created(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "workspaces-members-update",
		Method:      http.MethodPatch,
		Path:        "/workspaces/{workspaceId}/members/{memberId}",
		Summary:     "Update a workspace member",
		Tags:        []string{"workspaces"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.UpdateMemberInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleUpdateMember(ctx, userID, in.WorkspaceID, in.MemberID, in.Body.Role)
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.OK(out), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "workspaces-members-delete",
		Method:      http.MethodDelete,
		Path:        "/workspaces/{workspaceId}/members/{memberId}",
		Summary:     "Remove a workspace member",
		Tags:        []string{"workspaces"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.MemberPathInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := svc.HandleDeleteMember(ctx, userID, in.WorkspaceID, in.MemberID)
		if err != nil {
			return nil, humaapi.MapError(err, workspaceErrors, logger)
		}
		return humaapi.OK(out), nil
	})
}
