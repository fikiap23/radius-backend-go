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
	entworkspace "github.com/radius/radius-backend/ent/workspace"
	entwm "github.com/radius/radius-backend/ent/workspacemember"
	"github.com/radius/radius-backend/internal/shared/pagination"
	"github.com/radius/radius-backend/internal/workspaces/domain"
)

type WorkspaceRepository struct {
	client *ent.Client
}

func NewWorkspaceRepository(client *ent.Client) *WorkspaceRepository {
	return &WorkspaceRepository{client: client}
}

var _ domain.WorkspaceRepository = (*WorkspaceRepository)(nil)

func (r *WorkspaceRepository) Create(ctx context.Context, w *domain.Workspace) error {
	if w.ID == "" {
		w.ID = uuid.NewString()
	}

	created, err := r.client.Workspace.Create().
		SetID(w.ID).
		SetName(w.Name).
		SetSlug(w.Slug).
		Save(ctx)
	if err != nil {
		if mapped := mapWorkspaceSaveError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("create workspace: %w", err)
	}
	w.CreatedAt = created.CreatedAt
	w.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *WorkspaceRepository) FindByID(ctx context.Context, id string, fields ...domain.WorkspaceFields) (*domain.Workspace, error) {
	f := domain.WorkspaceFieldsAll
	if len(fields) > 0 {
		f = fields[0]
	}
	return r.FindOne(ctx, domain.WorkspaceQuery{
		Select: f,
		Filter: domain.WorkspaceFilter{ID: &id},
	})
}

func (r *WorkspaceRepository) FindBySlug(ctx context.Context, slug string, fields ...domain.WorkspaceFields) (*domain.Workspace, error) {
	f := domain.WorkspaceFieldsAll
	if len(fields) > 0 {
		f = fields[0]
	}
	return r.FindOne(ctx, domain.WorkspaceQuery{
		Select: f,
		Filter: domain.WorkspaceFilter{Slug: &slug},
	})
}

func (r *WorkspaceRepository) FindOne(ctx context.Context, q domain.WorkspaceQuery) (*domain.Workspace, error) {
	row, err := r.client.Workspace.Query().
		Where(buildWorkspaceFilter(q.Filter)...).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("find workspace: %w", err)
	}
	return toDomainWorkspace(row), nil
}

func (r *WorkspaceRepository) FindMany(ctx context.Context, q domain.WorkspaceQuery) ([]*domain.Workspace, error) {
	rows, err := r.client.Workspace.Query().
		Where(buildWorkspaceFilter(q.Filter)...).
		Order(entworkspace.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find workspaces: %w", err)
	}
	return toDomainWorkspaces(rows), nil
}

func (r *WorkspaceRepository) FindManyPaginate(
	ctx context.Context,
	q domain.WorkspaceQuery,
	params pagination.Params,
) (*pagination.Result[*domain.Workspace], error) {
	base := r.client.Workspace.Query().Where(buildWorkspaceFilter(q.Filter)...)

	total, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count workspaces: %w", err)
	}

	ol := params.OffsetLimit()
	rows, err := base.
		Order(workspaceOrder(params.Sort)...).
		Limit(ol.Limit).
		Offset(ol.Offset).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find workspaces: %w", err)
	}

	result := pagination.NewResult(toDomainWorkspaces(rows), int64(total), params)
	return &result, nil
}

func (r *WorkspaceRepository) UpdateByID(ctx context.Context, id string, data domain.WorkspaceUpdateData) error {
	update := r.client.Workspace.Update().
		Where(entworkspace.IDEQ(id)).
		SetUpdatedAt(time.Now().UTC()).
		SetNillableName(data.Name).
		SetNillableSlug(data.Slug)

	n, err := update.Save(ctx)
	if err != nil {
		if mapped := mapWorkspaceSaveError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("update workspace: %w", err)
	}
	if n == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

func (r *WorkspaceRepository) DeleteByID(ctx context.Context, id string) error {
	n, err := r.client.Workspace.Delete().
		Where(entworkspace.IDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	if n == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

func buildWorkspaceFilter(f domain.WorkspaceFilter) []predicate.Workspace {
	var preds []predicate.Workspace
	if f.ID != nil {
		preds = append(preds, entworkspace.IDEQ(*f.ID))
	}
	if f.Slug != nil {
		preds = append(preds, entworkspace.SlugEQ(*f.Slug))
	}
	if f.UserID != nil {
		preds = append(preds, entworkspace.HasMembersWith(
			entwm.UserIDEQ(*f.UserID),
			entwm.StatusEQ(entwm.StatusActive),
		))
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		preds = append(preds, entworkspace.Or(
			entworkspace.NameContainsFold(search),
			entworkspace.SlugContainsFold(search),
		))
	}
	return preds
}

func workspaceOrder(sort pagination.Sort) []entworkspace.OrderOption {
	dir := sql.OrderDesc()
	if sort.IsAsc() {
		dir = sql.OrderAsc()
	}
	switch sort.By {
	case "name":
		return []entworkspace.OrderOption{entworkspace.ByName(dir)}
	case "slug":
		return []entworkspace.OrderOption{entworkspace.BySlug(dir)}
	case "updatedAt":
		return []entworkspace.OrderOption{entworkspace.ByUpdatedAt(dir)}
	default:
		return []entworkspace.OrderOption{entworkspace.ByCreatedAt(dir)}
	}
}
