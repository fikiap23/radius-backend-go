package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	"github.com/radius/radius-backend/internal/shared/middleware"
	storagedomain "github.com/radius/radius-backend/internal/storage/domain"
	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

var userErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrUserNotFound, Status: http.StatusNotFound, Code: "user_not_found", Message: "User not found.", Param: "id"},
	{Err: storagedomain.ErrTempFileNotFound, Status: http.StatusBadRequest, Code: "temp_file_not_found", Message: "Uploaded temp file not found.", Param: "avatarTempKey"},
	{Err: storagedomain.ErrInvalidTempKey, Status: http.StatusBadRequest, Code: "invalid_temp_key", Message: "Invalid temp key.", Param: "avatarTempKey"},
}

func RegisterUsers(api huma.API, svc *services.UserService, auth *middleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID: "users-list",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List users (paginated)",
		Tags:        []string{"users"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.ListUsersInput) (*humaapi.OKOutput, error) {
		page, err := svc.HandleListUsers(ctx, in.Params())
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(page), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "users-get-me",
		Method:      http.MethodGet,
		Path:        "/users/me",
		Summary:     "Get current user profile",
		Tags:        []string{"users"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, _ *struct{}) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		profile, err := svc.HandleGetMe(ctx, userID)
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(profile), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "users-get-by-id",
		Method:      http.MethodGet,
		Path:        "/users/{id}",
		Summary:     "Get user by ID",
		Tags:        []string{"users"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.GetUserByIDInput) (*humaapi.OKOutput, error) {
		profile, err := svc.HandleGetByID(ctx, in.ID)
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(profile), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "users-update-me",
		Method:      http.MethodPatch,
		Path:        "/users/me",
		Summary:     "Update current user profile",
		Tags:        []string{"users"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *dto.UpdateMeInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		profile, err := svc.HandleUpdateMe(ctx, userID, in.ToDomain())
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(profile), nil
	})
}
