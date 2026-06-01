package postgres

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/ent/predicate"
	entuser "github.com/radius/radius-backend/ent/user"
	"github.com/radius/radius-backend/internal/users/domain"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	created, err := r.client.User.Create().
		SetID(u.ID).
		SetName(u.Name).
		SetEmail(u.Email).
		SetNillablePasswordHash(u.PasswordHash).
		SetNillableEmailVerifiedAt(u.EmailVerifiedAt).
		SetNillableAvatarURL(u.AvatarURL).
		SetNillableLastLoginAt(u.LastLoginAt).
		SetNillableTimezone(u.Timezone).
		SetLocale(u.Locale).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	u.CreatedAt = created.CreatedAt
	u.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string, fields ...domain.Fields) (*domain.User, error) {
	f := domain.FieldsAll
	if len(fields) > 0 {
		f = fields[0]
	}
	return r.FindOne(ctx, domain.Query{
		Select: f,
		Filter: domain.Filter{ID: &id},
	})
}

func (r *UserRepository) FindOne(ctx context.Context, q domain.Query) (*domain.User, error) {
	row, err := r.activeUsers().
		Where(buildFilter(q.Filter)...).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) FindMany(ctx context.Context, q domain.Query) ([]*domain.User, error) {
	rows, err := r.activeUsers().
		Where(buildFilter(q.Filter)...).
		Order(entuser.ByCreatedAt(sql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}
	return toDomainUsers(rows), nil
}

func (r *UserRepository) FindManyPaginate(ctx context.Context, q domain.Query, page domain.Page) (*domain.PageResult, error) {
	base := r.activeUsers().Where(buildFilter(q.Filter)...)

	total, err := base.Clone().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	limit, offset := normalizePage(page)
	rows, err := base.
		Order(entuser.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}

	return &domain.PageResult{
		Items: toDomainUsers(rows),
		Total: int64(total),
	}, nil
}

func (r *UserRepository) UpdateByID(ctx context.Context, id string, data domain.Update) error {
	update := r.client.User.Update().
		Where(entuser.IDEQ(id), entuser.DeletedAtIsNil()).
		SetUpdatedAt(time.Now().UTC())

	update = applyUpdate(update, data)

	n, err := update.Save(ctx)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id string) error {
	now := time.Now().UTC()
	n, err := r.client.User.Update().
		Where(entuser.IDEQ(id), entuser.DeletedAtIsNil()).
		SetDeletedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) activeUsers() *ent.UserQuery {
	return r.client.User.Query().Where(entuser.DeletedAtIsNil())
}

func buildFilter(f domain.Filter) []predicate.User {
	var preds []predicate.User
	if f.ID != nil {
		preds = append(preds, entuser.IDEQ(*f.ID))
	}
	if f.Email != nil {
		preds = append(preds, entuser.EmailEQ(*f.Email))
	}
	return preds
}

func applyUpdate(u *ent.UserUpdate, data domain.Update) *ent.UserUpdate {
	return u.
		SetNillableName(data.Name).
		SetNillableEmail(data.Email).
		SetNillablePasswordHash(data.PasswordHash).
		SetNillableEmailVerifiedAt(data.EmailVerifiedAt).
		SetNillableAvatarURL(data.AvatarURL).
		SetNillableLastLoginAt(data.LastLoginAt).
		SetNillableTimezone(data.Timezone).
		SetNillableLocale(data.Locale)
}

func normalizePage(page domain.Page) (int, int) {
	limit := page.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := page.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
