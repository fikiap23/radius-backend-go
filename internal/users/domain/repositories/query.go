package repositories

import (
	"time"

	"github.com/radius/radius-backend/internal/users/domain/entities"
)

// --- Select (column mask) ---

type UserFields uint32

const (
	FieldID UserFields = 1 << iota
	FieldName
	FieldEmail
	FieldPasswordHash
	FieldEmailVerifiedAt
	FieldAvatarURL
	FieldLastLoginAt
	FieldTimezone
	FieldLocale
	FieldCreatedAt
	FieldUpdatedAt

	SelectAll UserFields = 0

	SelectProfile = FieldID | FieldName | FieldEmail |
		FieldEmailVerifiedAt | FieldAvatarURL | FieldLastLoginAt |
		FieldTimezone | FieldLocale | FieldCreatedAt | FieldUpdatedAt

	SelectLogin = SelectProfile | FieldPasswordHash

	SelectExists = FieldID
)

// --- Where (filter) ---

type Where struct {
	ID    *string
	Email *string
}

// --- Query (select + where) ---

type Query struct {
	Select UserFields
	Where  Where
}

func (q Query) Fields() UserFields {
	return q.Select
}

// --- Update (partial write) ---

type Update struct {
	Name            *string
	Email           *string
	PasswordHash    *string
	EmailVerifiedAt *time.Time
	AvatarURL       *string
	LastLoginAt     *time.Time
	Timezone        *string
	Locale          *string
}

// --- Pagination ---

type Page struct {
	Limit  int
	Offset int
}

type PageResult struct {
	Items []*entities.User
	Total int64
}
