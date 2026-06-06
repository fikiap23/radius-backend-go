package dto

import (
	"time"

	"github.com/radius/radius-backend/internal/shared/pagination"
	"github.com/radius/radius-backend/internal/users/domain"
)

type ListUsersInput struct {
	pagination.HTTPQuery
}

func (in ListUsersInput) Params() pagination.Params {
	return in.HTTPQuery.ParamsWithSort("createdAt", "createdAt", "updatedAt", "name", "email")
}

type GetUserByIDInput struct {
	ID string `path:"id" doc:"User ID" format:"uuid"`
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

func MapUserProfile(u *domain.User) UserProfile {
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

type UpdateMeInput struct {
	Body struct {
		Name          *string `json:"name,omitempty" doc:"Display name" minLength:"2" maxLength:"255"`
		AvatarURL     *string `json:"avatarUrl,omitempty" doc:"Avatar URL (direct URL; prefer avatarTempKey from presign flow)" format:"uri"`
		AvatarTempKey *string `json:"avatarTempKey,omitempty" doc:"Temp object key from POST /storage/presign-upload (purpose=avatar)" maxLength:"512"`
		Timezone      *string `json:"timezone,omitempty" doc:"IANA timezone" maxLength:"64"`
		Locale        *string `json:"locale,omitempty" doc:"Locale code" minLength:"2" maxLength:"10"`
	}
}

func (in *UpdateMeInput) ToDomain() domain.ProfileUpdate {
	return domain.ProfileUpdate{
		Name:          in.Body.Name,
		AvatarURL:     in.Body.AvatarURL,
		AvatarTempKey: in.Body.AvatarTempKey,
		Timezone:      in.Body.Timezone,
		Locale:        in.Body.Locale,
	}
}
