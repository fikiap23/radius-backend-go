package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/ent/predicate"
	entproject "github.com/radius/radius-backend/ent/project"
	"github.com/radius/radius-backend/internal/projects/domain"
)

type ProjectRepository struct {
	client *ent.Client
}

func NewProjectRepository(client *ent.Client) *ProjectRepository {
	return &ProjectRepository{client: client}
}

var _ domain.ProjectRepository = (*ProjectRepository)(nil)

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}

	created, err := r.client.Project.Create().
		SetID(p.ID).
		SetWorkspaceID(p.WorkspaceID).
		SetName(p.Name).
		SetNillableDescription(p.Description).
		SetIcon(p.Icon).
		SetCover(toEntCover(p.Cover)).
		SetNillableCoverImageURL(p.CoverImageURL).
		SetStatus(toEntStatus(p.Status)).
		SetIsFavorite(p.IsFavorite).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	p.CreatedAt = created.CreatedAt
	p.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	row, err := r.client.Project.Query().
		Where(entproject.IDEQ(id)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("find project: %w", err)
	}
	return toDomainProject(row), nil
}

func (r *ProjectRepository) FindMany(ctx context.Context, q domain.ProjectQuery) ([]*domain.Project, error) {
	rows, err := r.client.Project.Query().
		Where(buildProjectFilter(q.Filter)...).
		Order(entproject.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find projects: %w", err)
	}
	return toDomainProjects(rows), nil
}

func (r *ProjectRepository) UpdateByID(ctx context.Context, id string, data domain.ProjectUpdateData) error {
	update := r.client.Project.Update().
		Where(entproject.IDEQ(id)).
		SetUpdatedAt(time.Now().UTC())

	if data.Name != nil {
		update = update.SetName(*data.Name)
	}
	if data.Description != nil {
		update = update.SetDescription(*data.Description)
	}
	if data.Icon != nil {
		update = update.SetIcon(*data.Icon)
	}
	if data.Cover != nil {
		update = update.SetCover(toEntCover(*data.Cover))
	}
	if data.CoverImageURL != nil {
		update = update.SetCoverImageURL(*data.CoverImageURL)
	}
	if data.Status != nil {
		update = update.SetStatus(toEntStatus(*data.Status))
	}
	if data.IsFavorite != nil {
		update = update.SetIsFavorite(*data.IsFavorite)
	}

	n, err := update.Save(ctx)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	if n == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) SetFavorite(ctx context.Context, id string, isFavorite bool) error {
	n, err := r.client.Project.Update().
		Where(entproject.IDEQ(id)).
		SetIsFavorite(isFavorite).
		SetUpdatedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("set project favorite: %w", err)
	}
	if n == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) Archive(ctx context.Context, id string) error {
	now := time.Now().UTC()
	n, err := r.client.Project.Update().
		Where(entproject.IDEQ(id)).
		SetArchivedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("archive project: %w", err)
	}
	if n == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) Unarchive(ctx context.Context, id string) error {
	n, err := r.client.Project.Update().
		Where(entproject.IDEQ(id)).
		ClearArchivedAt().
		SetUpdatedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("unarchive project: %w", err)
	}
	if n == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) DeleteByID(ctx context.Context, id string) error {
	n, err := r.client.Project.Delete().
		Where(entproject.IDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if n == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func buildProjectFilter(f domain.ProjectFilter) []predicate.Project {
	var preds []predicate.Project
	if f.ID != nil {
		preds = append(preds, entproject.IDEQ(*f.ID))
	}
	if f.WorkspaceID != nil {
		preds = append(preds, entproject.WorkspaceIDEQ(*f.WorkspaceID))
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		preds = append(preds, entproject.NameContainsFold(search))
	}
	return preds
}
