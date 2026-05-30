package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/radius/radius-backend/internal/shared/config"
)

func SignAccessToken(cfg config.JWTConfig, userID, email string) (token string, expiresIn int64, err error) {
	expiresAt := time.Now().UTC().Add(cfg.Expiry)
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iss":   cfg.Issuer,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().UTC().Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	return signed, int64(cfg.Expiry.Seconds()), nil
}
