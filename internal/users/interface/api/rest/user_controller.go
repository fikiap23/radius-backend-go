package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/shared/apirest"
	appmiddleware "github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/users/domain/entities"
	"go.uber.org/zap"
)

type UserController struct {
	service *services.UserService
	logger  *zap.Logger
}

type UpdateMeRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=2,max=255"`
	AvatarURL *string `json:"avatarUrl" validate:"omitempty,url"`
	Timezone  *string `json:"timezone" validate:"omitempty,max=64"`
	Locale    *string `json:"locale" validate:"omitempty,min=2,max=10"`
}

var userErrors = []apirest.ErrorMapping{
	{Err: domain.ErrUserNotFound, Status: http.StatusNotFound, Message: "USER_NOT_FOUND"},
}

func NewUserController(e *echo.Echo, service *services.UserService, auth *appmiddleware.AuthMiddleware, logger *zap.Logger) *UserController {
	c := &UserController{service: service, logger: logger}

	v1 := e.Group("/v1")
	users := v1.Group("/users", auth.Authenticate())
	users.GET("/me", c.GetMe)
	users.PATCH("/me", c.UpdateMe)

	return c
}

// GetMe godoc
// @Summary      Get current user profile
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SwaggerUserOK
// @Failure      401  {object}  SwaggerErr
// @Failure      404  {object}  SwaggerErr
// @Failure      500  {object}  SwaggerErr
// @Router       /v1/users/me [get]
func (c *UserController) GetMe(ctx echo.Context) error {
	userID, err := apirest.UserID(ctx)
	if err != nil {
		return apirest.Unauthorized(ctx)
	}

	profile, err := c.service.GetMe(ctx.Request().Context(), userID)
	if err != nil {
		return apirest.Handle(ctx, err, userErrors, c.logger)
	}

	return apirest.OK(ctx, profile)
}

// UpdateMe godoc
// @Summary      Update current user profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      UpdateMeRequest  true  "Profile fields to update"
// @Success      200   {object}  SwaggerUserOK
// @Failure      400   {object}  SwaggerErr
// @Failure      401   {object}  SwaggerErr
// @Failure      404   {object}  SwaggerErr
// @Failure      500   {object}  SwaggerErr
// @Router       /v1/users/me [patch]
func (c *UserController) UpdateMe(ctx echo.Context) error {
	userID, err := apirest.UserID(ctx)
	if err != nil {
		return apirest.Unauthorized(ctx)
	}

	req, err := apirest.Bind[UpdateMeRequest](ctx)
	if err != nil {
		return apirest.BindErr(ctx, err)
	}

	input := entities.UpdateProfileInput{
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Timezone:  req.Timezone,
		Locale:    req.Locale,
	}

	profile, err := c.service.UpdateMe(ctx.Request().Context(), userID, input)
	if err != nil {
		return apirest.Handle(ctx, err, userErrors, c.logger)
	}

	return apirest.OK(ctx, profile)
}
