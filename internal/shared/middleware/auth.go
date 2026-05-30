package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	bearerPrefix = "Bearer "
	claimsKey    = "claims"
	userIDKey    = "userID"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	secretKey []byte
	logger    *zap.Logger
}

func NewAuthMiddleware(secretKey string, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: []byte(secretKey),
		logger:    logger,
	}
}

func (m *AuthMiddleware) Authenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := m.ExtractBearerToken(c.Request().Header.Get("Authorization"))
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or malformed token")
			}

			if err := m.SetUserOnContext(c, token); err != nil {
				m.logger.Warn("invalid token", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) ExtractBearerToken(authHeader string) (string, error) {
	return m.extractToken(authHeader)
}

func (m *AuthMiddleware) SetUserOnContext(c echo.Context, token string) error {
	claims, err := m.validateToken(token)
	if err != nil {
		return err
	}
	c.Set(claimsKey, claims)
	c.Set(userIDKey, claims.Subject)
	return nil
}

func (m *AuthMiddleware) extractToken(authHeader string) (string, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("missing bearer token")
	}
	return strings.TrimPrefix(authHeader, bearerPrefix), nil
}

func (m *AuthMiddleware) validateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GetClaims(c echo.Context) (*Claims, bool) {
	val := c.Get(claimsKey)
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*Claims)
	return claims, ok
}

func GetUserID(c echo.Context) (string, bool) {
	val := c.Get(userIDKey)
	if val == nil {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok
}
