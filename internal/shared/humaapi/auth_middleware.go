package humaapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/radius/radius-backend/internal/shared/middleware"
)

type contextKey string

const userIDContextKey contextKey = "radiusUserID"

func UserIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(userIDContextKey).(string)
	if !ok || id == "" {
		return "", huma.Error401Unauthorized("UNAUTHORIZED")
	}
	return id, nil
}

func RequireAuth(auth *middleware.AuthMiddleware, api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		ec := humaecho.Unwrap(ctx)

		token, err := auth.ExtractBearerToken(ec.Request().Header.Get("Authorization"))
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "UNAUTHORIZED", err)
			return
		}

		if err := auth.SetUserOnContext(ec, token); err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "UNAUTHORIZED", err)
			return
		}

		userID, _ := middleware.GetUserID(ec)
		req := ec.Request().WithContext(context.WithValue(ec.Request().Context(), userIDContextKey, userID))
		ec.SetRequest(req)

		next(ctx)
	}
}

func BearerSecurity() []map[string][]string {
	return []map[string][]string{{"bearer": {}}}
}
