package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/internal/shared/apirest"
	"github.com/radius/radius-backend/internal/users/application/services"
	"github.com/radius/radius-backend/internal/users/domain"
	"go.uber.org/zap"
)

type AuthController struct {
	service *services.AuthService
	logger  *zap.Logger
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

var authErrors = []apirest.ErrorMapping{
	{Err: domain.ErrEmailAlreadyExists, Status: http.StatusConflict, Message: "EMAIL_ALREADY_EXISTS"},
	{Err: domain.ErrInvalidCredentials, Status: http.StatusUnauthorized, Message: "INVALID_CREDENTIALS"},
}

func NewAuthController(e *echo.Echo, service *services.AuthService, logger *zap.Logger) *AuthController {
	c := &AuthController{service: service, logger: logger}

	v1 := e.Group("/v1")
	auth := v1.Group("/auth")
	auth.POST("/register", c.Register)
	auth.POST("/login", c.Login)

	return c
}

// Register godoc
// @Summary      Register user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Register payload"
// @Success      201   {object}  SwaggerAuthOK
// @Failure      400   {object}  SwaggerErr
// @Failure      409   {object}  SwaggerErr
// @Failure      500   {object}  SwaggerErr
// @Router       /v1/auth/register [post]
func (c *AuthController) Register(ctx echo.Context) error {
	req, err := apirest.Bind[RegisterRequest](ctx)
	if err != nil {
		return apirest.BindErr(ctx, err)
	}

	result, err := c.service.Register(ctx.Request().Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return apirest.Handle(ctx, err, authErrors, c.logger)
	}

	return apirest.Created(ctx, result)
}

// Login godoc
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login payload"
// @Success      200   {object}  SwaggerAuthOK
// @Failure      400   {object}  SwaggerErr
// @Failure      401   {object}  SwaggerErr
// @Failure      500   {object}  SwaggerErr
// @Router       /v1/auth/login [post]
func (c *AuthController) Login(ctx echo.Context) error {
	req, err := apirest.Bind[LoginRequest](ctx)
	if err != nil {
		return apirest.BindErr(ctx, err)
	}

	result, err := c.service.Login(ctx.Request().Context(), req.Email, req.Password)
	if err != nil {
		return apirest.Handle(ctx, err, authErrors, c.logger)
	}

	return apirest.OK(ctx, result)
}
