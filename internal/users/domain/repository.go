package domain

import (
	"context"
	"time"
)

// Fields represents a named column selection preset for user queries.
type Fields int

const (
	FieldsAll     Fields = iota // all columns
	FieldsProfile              // public profile fields
	FieldsLogin                // profile + password hash
	FieldsExists               // id only (existence check)
)

type Filter struct {
	ID    *string
	Email *string
}

type Query struct {
	Select Fields
	Filter Filter
}

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

type Page struct {
	Limit  int
	Offset int
}

type PageResult struct {
	Items []*User
	Total int64
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string, fields ...Fields) (*User, error)
	FindOne(ctx context.Context, q Query) (*User, error)
	FindMany(ctx context.Context, q Query) ([]*User, error)
	FindManyPaginate(ctx context.Context, q Query, page Page) (*PageResult, error)
	UpdateByID(ctx context.Context, id string, data Update) error
	DeleteByID(ctx context.Context, id string) error
}

type OAuthAccountRepository interface {
	Create(ctx context.Context, account *OAuthAccount) error
	FindByProviderAccount(ctx context.Context, provider OAuthProvider, providerUserID string) (*OAuthAccount, error)
}
