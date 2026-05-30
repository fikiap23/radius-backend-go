package repositories

import (
	"context"
	"time"

	"github.com/radius/radius-backend/internal/users/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id string) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	UpdateLastLogin(ctx context.Context, id string, at time.Time) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
