package oauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/radius/radius-backend/internal/shared/config"
)

const stateIssuer = "radius-oauth-state"

type StateClaims struct {
	Provider    string `json:"provider"`
	RedirectURI string `json:"redirectUri"`
	jwt.RegisteredClaims
}

func SignState(jwtCfg config.JWTConfig, provider, redirectURI string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := StateClaims{
		Provider:    provider,
		RedirectURI: redirectURI,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    stateIssuer,
			Subject:   provider,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(jwtCfg.SecretKey))
	if err != nil {
		return "", fmt.Errorf("sign oauth state: %w", err)
	}
	return signed, nil
}

func VerifyState(jwtCfg config.JWTConfig, state, expectedProvider string) (redirectURI string, err error) {
	token, err := jwt.ParseWithClaims(state, &StateClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtCfg.SecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("parse oauth state: %w", err)
	}

	claims, ok := token.Claims.(*StateClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid oauth state claims")
	}
	if claims.Issuer != stateIssuer {
		return "", errors.New("invalid oauth state issuer")
	}
	if claims.Provider != expectedProvider {
		return "", errors.New("oauth state provider mismatch")
	}
	if claims.RedirectURI == "" {
		return "", errors.New("oauth state missing redirect uri")
	}

	return claims.RedirectURI, nil
}
