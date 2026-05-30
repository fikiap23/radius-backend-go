package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/radius/radius-backend/internal/shared/humaapi"
	appmiddleware "github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"go.uber.org/zap"
)

type updateMeInput struct {
	Body struct {
		Name      *string `json:"name,omitempty" doc:"Display name" minLength:"2" maxLength:"255"`
		AvatarURL *string `json:"avatarUrl,omitempty" doc:"Avatar URL" format:"uri"`
		Timezone  *string `json:"timezone,omitempty" doc:"IANA timezone" maxLength:"64"`
		Locale    *string `json:"locale,omitempty" doc:"Locale code" minLength:"2" maxLength:"10"`
	}
}

var userErrors = []humaapi.ErrorMapping{
	{Err: domain.ErrUserNotFound, Status: http.StatusNotFound, Message: "USER_NOT_FOUND"},
}

func RegisterUsers(api huma.API, service *services.UserService, auth *appmiddleware.AuthMiddleware, logger *zap.Logger) {
	authMW := humaapi.RequireAuth(auth, api)

	huma.Register(api, huma.Operation{
		OperationID: "users-get-me",
		Method:      http.MethodGet,
		Path:        "/v1/users/me",
		Summary:     "Get current user profile",
		Tags:        []string{"users"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, _ *struct{}) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		profile, err := service.GetMe(ctx, userID)
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(profile), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "users-update-me",
		Method:      http.MethodPatch,
		Path:        "/v1/users/me",
		Summary:     "Update current user profile",
		Tags:        []string{"users"},
		Security:    humaapi.BearerSecurity(),
		Middlewares: huma.Middlewares{authMW},
	}, func(ctx context.Context, in *updateMeInput) (*humaapi.OKOutput, error) {
		userID, err := humaapi.UserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		input := entities.UpdateProfileInput{
			Name:      in.Body.Name,
			AvatarURL: in.Body.AvatarURL,
			Timezone:  in.Body.Timezone,
			Locale:    in.Body.Locale,
		}

		profile, err := service.UpdateMe(ctx, userID, input)
		if err != nil {
			return nil, humaapi.MapError(err, userErrors, logger)
		}
		return humaapi.OK(profile), nil
	})
}

