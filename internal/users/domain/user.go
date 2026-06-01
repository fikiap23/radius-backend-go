package domain

import "time"

type User struct {
	ID              string
	Name            string
	Email           string
	PasswordHash    *string
	EmailVerifiedAt *time.Time
	AvatarURL       *string
	LastLoginAt     *time.Time
	Timezone        *string
	Locale          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProfileUpdate struct {
	Name      *string
	AvatarURL *string
	Timezone  *string
	Locale    *string
}

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
