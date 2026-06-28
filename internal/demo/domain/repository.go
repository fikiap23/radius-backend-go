package domain

import "context"

// DemoRepository persists Demo entities.
// TODO: implement in infrastructure/db/postgres.
type DemoRepository interface {
	// TODO: add repository methods.
	FindByID(ctx context.Context, id string) (*Demo, error)
}
