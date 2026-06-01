package postgres

import (
	"github.com/radius/radius-backend/internal/users/domain"
	"gorm.io/gorm"
)

var fieldColumns = map[domain.Fields][]string{
	domain.FieldsProfile: {
		"id", "name", "email", "email_verified_at", "avatar_url",
		"last_login_at", "timezone", "locale", "created_at", "updated_at",
	},
	domain.FieldsLogin: {
		"id", "name", "email", "password_hash", "email_verified_at", "avatar_url",
		"last_login_at", "timezone", "locale", "created_at", "updated_at",
	},
	domain.FieldsExists: {"id"},
}

func applySelect(db *gorm.DB, fields domain.Fields) *gorm.DB {
	cols, ok := fieldColumns[fields]
	if !ok || fields == domain.FieldsAll {
		return db
	}
	return db.Select(cols)
}

func applyFilter(db *gorm.DB, f domain.Filter) *gorm.DB {
	db = db.Where("deleted_at IS NULL")
	if f.ID != nil {
		db = db.Where("id = ?", *f.ID)
	}
	if f.Email != nil {
		db = db.Where("email = ?", *f.Email)
	}
	return db
}

func updateToMap(data domain.Update) map[string]any {
	m := make(map[string]any)
	if data.Name != nil {
		m["name"] = *data.Name
	}
	if data.Email != nil {
		m["email"] = *data.Email
	}
	if data.PasswordHash != nil {
		m["password_hash"] = *data.PasswordHash
	}
	if data.EmailVerifiedAt != nil {
		m["email_verified_at"] = *data.EmailVerifiedAt
	}
	if data.AvatarURL != nil {
		m["avatar_url"] = *data.AvatarURL
	}
	if data.LastLoginAt != nil {
		m["last_login_at"] = *data.LastLoginAt
	}
	if data.Timezone != nil {
		m["timezone"] = *data.Timezone
	}
	if data.Locale != nil {
		m["locale"] = *data.Locale
	}
	return m
}
