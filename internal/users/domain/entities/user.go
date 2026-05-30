package entities

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

type UserProfile struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	AvatarURL       *string    `json:"avatarUrl,omitempty"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	Locale          string     `json:"locale"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (u *User) ToProfile() UserProfile {
	return UserProfile{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		EmailVerifiedAt: u.EmailVerifiedAt,
		AvatarURL:       u.AvatarURL,
		LastLoginAt:     u.LastLoginAt,
		Timezone:        u.Timezone,
		Locale:          u.Locale,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

type UpdateProfileInput struct {
	Name      *string
	AvatarURL *string
	Timezone  *string
	Locale    *string
}
