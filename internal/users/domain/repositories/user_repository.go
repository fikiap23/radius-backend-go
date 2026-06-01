package repositories

import (
	"context"

	"github.com/radius/radius-backend/internal/users/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id string, fields ...UserFields) (*entities.User, error)
	FindOne(ctx context.Context, q Query) (*entities.User, error)
	FindMany(ctx context.Context, q Query) ([]*entities.User, error)
	FindManyPaginate(ctx context.Context, q Query, page Page) (*PageResult, error)
	UpdateByID(ctx context.Context, id string, data Update) error
	DeleteByID(ctx context.Context, id string) error
}
