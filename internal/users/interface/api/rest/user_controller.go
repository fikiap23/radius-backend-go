package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	appmiddleware "github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/users/application/dto"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

var userErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrUserNotFound, Status: http.StatusNotFound, Code: "user_not_found", Message: "User not found.", Param: "id"},
}

func RegisterUsers(api huma.API, service *services.UserService, auth *appmiddleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

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

		profile, err := service.HandleGetMe(ctx, userID)
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

		profile, err := service.HandleUpdateMe(ctx, userID, in.ToDomain())
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(profile), nil
	})
}
