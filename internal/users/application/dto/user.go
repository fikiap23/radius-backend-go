package dto

import "github.com/radius/radius-backend/internal/users/domain/entities"

type UpdateMeInput struct {
	Body struct {
		Name      *string `json:"name,omitempty" doc:"Display name" minLength:"2" maxLength:"255"`
		AvatarURL *string `json:"avatarUrl,omitempty" doc:"Avatar URL" format:"uri"`
		Timezone  *string `json:"timezone,omitempty" doc:"IANA timezone" maxLength:"64"`
		Locale    *string `json:"locale,omitempty" doc:"Locale code" minLength:"2" maxLength:"10"`
	}
}

func (in *UpdateMeInput) ToDomain() entities.UpdateProfileInput {
	return entities.UpdateProfileInput{
		Name:      in.Body.Name,
		AvatarURL: in.Body.AvatarURL,
		Timezone:  in.Body.Timezone,
		Locale:    in.Body.Locale,
	}
}
