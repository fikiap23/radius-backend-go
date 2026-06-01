package domain

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserInactive            = errors.New("user inactive")
	ErrPasswordRequired        = errors.New("password required")
	ErrSSOProviderDisabled     = errors.New("sso provider disabled")
	ErrSSOInvalidState         = errors.New("sso invalid state")
	ErrSSOInvalidRedirectURI   = errors.New("sso invalid redirect uri")
	ErrSSOAuthenticationFailed   = errors.New("sso authentication failed")
)
