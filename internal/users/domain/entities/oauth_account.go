package entities

import "time"

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGitHub OAuthProvider = "github"
)

type OAuthAccount struct {
	ID             string
	UserID         string
	Provider       OAuthProvider
	ProviderUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
