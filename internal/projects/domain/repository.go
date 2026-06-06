package domain

import "context"

type ProjectFilter struct {
	ID          *string
	WorkspaceID *string
	Search      string
}

type ProjectQuery struct {
	Filter ProjectFilter
}

type ProjectUpdateData struct {
	Name               *string
	Description        *string
	Icon               *string
	Cover              *ProjectCover
	CoverImageURL      *string
	CoverImageTempKey  *string
	Status             *ProjectStatus
	IsFavorite         *bool
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id string) (*Project, error)
	FindMany(ctx context.Context, q ProjectQuery) ([]*Project, error)
	UpdateByID(ctx context.Context, id string, data ProjectUpdateData) error
	SetFavorite(ctx context.Context, id string, isFavorite bool) error
	Archive(ctx context.Context, id string) error
	Unarchive(ctx context.Context, id string) error
	DeleteByID(ctx context.Context, id string) error
}
