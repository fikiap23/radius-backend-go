package postgres

import (
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/users/domain"
)

func toDomainUser(row *ent.User) *domain.User {
	return &domain.User{
		ID:              row.ID,
		Name:            row.Name,
		Email:           row.Email,
		PasswordHash:    row.PasswordHash,
		EmailVerifiedAt: row.EmailVerifiedAt,
		AvatarURL:       row.AvatarURL,
		LastLoginAt:     row.LastLoginAt,
		Timezone:        row.Timezone,
		Locale:          row.Locale,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func toDomainUsers(rows []*ent.User) []*domain.User {
	out := make([]*domain.User, len(rows))
	for i, row := range rows {
		out[i] = toDomainUser(row)
	}
	return out
}
