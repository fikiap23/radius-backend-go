package domain

import (
	"context"
	"time"
)

type BoardColumn struct {
	ID        string
	ProjectID string
	Title     string
	Status    string
	WipLimit  *int
	Order     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BoardColumnUpdateData struct {
	Title    *string
	Status   *string
	WipLimit *int
}

type BoardColumnFilter struct {
	ProjectID *string
	ID        *string
}

type BoardColumnQuery struct {
	Filter BoardColumnFilter
}

type BoardColumnRepository interface {
	Create(ctx context.Context, column *BoardColumn) error
	FindByID(ctx context.Context, projectID, columnID string) (*BoardColumn, error)
	FindMany(ctx context.Context, q BoardColumnQuery) ([]*BoardColumn, error)
	UpdateByID(ctx context.Context, projectID, columnID string, data BoardColumnUpdateData) error
	Reorder(ctx context.Context, projectID string, columnIDs []string) error
	DeleteByID(ctx context.Context, projectID, columnID string) error
	CountByProjectID(ctx context.Context, projectID string) (int, error)
}
