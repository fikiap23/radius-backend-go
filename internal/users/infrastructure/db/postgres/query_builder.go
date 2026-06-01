package postgres

import (
	"github.com/radius/radius-backend/internal/users/domain/repositories"
	"gorm.io/gorm"
)

// --- Select ---

var fieldColumnMap = []struct {
	field  repositories.UserFields
	column string
}{
	{repositories.FieldID, "id"},
	{repositories.FieldName, "name"},
	{repositories.FieldEmail, "email"},
	{repositories.FieldPasswordHash, "password_hash"},
	{repositories.FieldEmailVerifiedAt, "email_verified_at"},
	{repositories.FieldAvatarURL, "avatar_url"},
	{repositories.FieldLastLoginAt, "last_login_at"},
	{repositories.FieldTimezone, "timezone"},
	{repositories.FieldLocale, "locale"},
	{repositories.FieldCreatedAt, "created_at"},
	{repositories.FieldUpdatedAt, "updated_at"},
}

func applySelect(db *gorm.DB, fields repositories.UserFields) *gorm.DB {
	if fields == repositories.SelectAll {
		return db
	}
	cols := make([]string, 0, 11)
	for _, m := range fieldColumnMap {
		if fields&m.field != 0 {
			cols = append(cols, m.column)
		}
	}
	if len(cols) == 0 {
		return db
	}
	return db.Select(cols)
}

// --- Where ---

func applyWhere(db *gorm.DB, where repositories.Where) *gorm.DB {
	db = db.Where("deleted_at IS NULL")
	if where.ID != nil {
		db = db.Where("id = ?", *where.ID)
	}
	if where.Email != nil {
		db = db.Where("email = ?", *where.Email)
	}
	return db
}

// --- Update ---

func updateToMap(data repositories.Update) map[string]interface{} {
	m := make(map[string]interface{})
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
