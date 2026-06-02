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
	entwm "github.com/radius/radius-backend/ent/workspacemember"
	"github.com/radius/radius-backend/internal/shared/pagination"
	"github.com/radius/radius-backend/internal/workspaces/domain"
)

type WorkspaceMemberRepository struct {
	client *ent.Client
}

func NewWorkspaceMemberRepository(client *ent.Client) *WorkspaceMemberRepository {
	return &WorkspaceMemberRepository{client: client}
}

var _ domain.WorkspaceMemberRepository = (*WorkspaceMemberRepository)(nil)

func (r *WorkspaceMemberRepository) Create(ctx context.Context, m *domain.WorkspaceMember) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	created, err := r.client.WorkspaceMember.Create().
		SetID(m.ID).
		SetWorkspaceID(m.WorkspaceID).
		SetNillableUserID(m.UserID).
		SetName(m.Name).
		SetEmail(m.Email).
		SetRole(toEntMemberRole(m.Role)).
		SetStatus(toEntMemberStatus(m.Status)).
		Save(ctx)
	if err != nil {
		if mapped := mapWorkspaceMemberSaveError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("create workspace member: %w", err)
	}
	m.CreatedAt = created.CreatedAt
	m.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *WorkspaceMemberRepository) FindByID(
	ctx context.Context,
	id string,
	fields ...domain.WorkspaceMemberFields,
) (*domain.WorkspaceMember, error) {
	f := domain.WorkspaceMemberFieldsAll
	if len(fields) > 0 {
		f = fields[0]
	}
	return r.FindOne(ctx, domain.WorkspaceMemberQuery{
		Select: f,
		Filter: domain.WorkspaceMemberFilter{ID: &id},
	})
}

func (r *WorkspaceMemberRepository) FindOne(ctx context.Context, q domain.WorkspaceMemberQuery) (*domain.WorkspaceMember, error) {
	row, err := r.client.WorkspaceMember.Query().
		Where(buildWorkspaceMemberFilter(q.Filter)...).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrWorkspaceMemberNotFound
		}
		return nil, fmt.Errorf("find workspace member: %w", err)
	}
	return toDomainWorkspaceMember(row), nil
}

func (r *WorkspaceMemberRepository) FindMany(ctx context.Context, q domain.WorkspaceMemberQuery) ([]*domain.WorkspaceMember, error) {
	rows, err := r.client.WorkspaceMember.Query().
		Where(buildWorkspaceMemberFilter(q.Filter)...).
		Order(entwm.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find workspace members: %w", err)
	}
	return toDomainWorkspaceMembers(rows), nil
}

func (r *WorkspaceMemberRepository) FindManyPaginate(
	ctx context.Context,
	q domain.WorkspaceMemberQuery,
	params pagination.Params,
) (*pagination.Result[*domain.WorkspaceMember], error) {
	base := r.client.WorkspaceMember.Query().Where(buildWorkspaceMemberFilter(q.Filter)...)

	total, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count workspace members: %w", err)
	}

	ol := params.OffsetLimit()
	rows, err := base.
		Order(workspaceMemberOrder(params.Sort)...).
		Limit(ol.Limit).
		Offset(ol.Offset).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find workspace members: %w", err)
	}

	result := pagination.NewResult(toDomainWorkspaceMembers(rows), int64(total), params)
	return &result, nil
}

func (r *WorkspaceMemberRepository) UpdateByID(ctx context.Context, id string, data domain.WorkspaceMemberUpdateData) error {
	update := r.client.WorkspaceMember.Update().
		Where(entwm.IDEQ(id)).
		SetUpdatedAt(time.Now().UTC()).
		SetNillableUserID(data.UserID).
		SetNillableName(data.Name).
		SetNillableEmail(data.Email)

	if data.Role != nil {
		update = update.SetRole(toEntMemberRole(*data.Role))
	}
	if data.Status != nil {
		update = update.SetStatus(toEntMemberStatus(*data.Status))
	}

	n, err := update.Save(ctx)
	if err != nil {
		if mapped := mapWorkspaceMemberSaveError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("update workspace member: %w", err)
	}
	if n == 0 {
		return domain.ErrWorkspaceMemberNotFound
	}
	return nil
}

func (r *WorkspaceMemberRepository) DeleteByID(ctx context.Context, id string) error {
	n, err := r.client.WorkspaceMember.Delete().
		Where(entwm.IDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete workspace member: %w", err)
	}
	if n == 0 {
		return domain.ErrWorkspaceMemberNotFound
	}
	return nil
}

func buildWorkspaceMemberFilter(f domain.WorkspaceMemberFilter) []predicate.WorkspaceMember {
	var preds []predicate.WorkspaceMember
	if f.ID != nil {
		preds = append(preds, entwm.IDEQ(*f.ID))
	}
	if f.WorkspaceID != nil {
		preds = append(preds, entwm.WorkspaceIDEQ(*f.WorkspaceID))
	}
	if f.UserID != nil {
		preds = append(preds, entwm.UserIDEQ(*f.UserID))
	}
	if f.Email != nil {
		preds = append(preds, entwm.EmailEQ(*f.Email))
	}
	if f.Role != nil {
		preds = append(preds, entwm.RoleEQ(toEntMemberRole(*f.Role)))
	}
	if f.Status != nil {
		preds = append(preds, entwm.StatusEQ(toEntMemberStatus(*f.Status)))
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		preds = append(preds, entwm.Or(
			entwm.NameContainsFold(search),
			entwm.EmailContainsFold(search),
		))
	}
	return preds
}

func workspaceMemberOrder(sort pagination.Sort) []entwm.OrderOption {
	dir := sql.OrderDesc()
	if sort.IsAsc() {
		dir = sql.OrderAsc()
	}
	switch sort.By {
	case "name":
		return []entwm.OrderOption{entwm.ByName(dir)}
	case "email":
		return []entwm.OrderOption{entwm.ByEmail(dir)}
	case "role":
		return []entwm.OrderOption{entwm.ByRole(dir)}
	case "status":
		return []entwm.OrderOption{entwm.ByStatus(dir)}
	case "updatedAt":
		return []entwm.OrderOption{entwm.ByUpdatedAt(dir)}
	default:
		return []entwm.OrderOption{entwm.ByCreatedAt(dir)}
	}
}
