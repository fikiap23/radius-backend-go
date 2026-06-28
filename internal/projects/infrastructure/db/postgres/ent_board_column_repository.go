package postgres

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/radius/radius-backend/ent"
	entboardcolumn "github.com/radius/radius-backend/ent/boardcolumn"
	"github.com/radius/radius-backend/ent/predicate"
	"github.com/radius/radius-backend/internal/projects/domain"
)

type BoardColumnRepository struct {
	client *ent.Client
}

func NewBoardColumnRepository(client *ent.Client) *BoardColumnRepository {
	return &BoardColumnRepository{client: client}
}

var _ domain.BoardColumnRepository = (*BoardColumnRepository)(nil)

func (r *BoardColumnRepository) Create(ctx context.Context, col *domain.BoardColumn) error {
	if col.ID == "" {
		col.ID = uuid.NewString()
	}

	created, err := r.client.BoardColumn.Create().
		SetID(col.ID).
		SetProjectID(col.ProjectID).
		SetTitle(col.Title).
		SetStatus(col.Status).
		SetNillableWipLimit(col.WipLimit).
		SetSortOrder(col.Order).
		Save(ctx)
	if err != nil {
		if mapped := mapBoardColumnSaveError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("create board column: %w", err)
	}
	col.CreatedAt = created.CreatedAt
	col.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *BoardColumnRepository) FindByID(ctx context.Context, projectID, columnID string) (*domain.BoardColumn, error) {
	row, err := r.client.BoardColumn.Query().
		Where(
			entboardcolumn.IDEQ(columnID),
			entboardcolumn.ProjectIDEQ(projectID),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrBoardColumnNotFound
		}
		return nil, fmt.Errorf("find board column: %w", err)
	}
	return toDomainBoardColumn(row), nil
}

func (r *BoardColumnRepository) FindMany(ctx context.Context, q domain.BoardColumnQuery) ([]*domain.BoardColumn, error) {
	rows, err := r.client.BoardColumn.Query().
		Where(buildBoardColumnFilter(q.Filter)...).
		Order(entboardcolumn.BySortOrder(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("find board columns: %w", err)
	}
	return toDomainBoardColumns(rows), nil
}

func (r *BoardColumnRepository) UpdateByID(ctx context.Context, projectID, columnID string, data domain.BoardColumnUpdateData) error {
	update := r.client.BoardColumn.Update().
		Where(
			entboardcolumn.IDEQ(columnID),
			entboardcolumn.ProjectIDEQ(projectID),
		).
		SetUpdatedAt(time.Now().UTC())

	if data.Title != nil {
		update = update.SetTitle(*data.Title)
	}
	if data.Status != nil {
		update = update.SetStatus(*data.Status)
	}
	if data.WipLimit != nil {
		update = update.SetWipLimit(*data.WipLimit)
	}

	n, err := update.Save(ctx)
	if err != nil {
		if mapped := mapBoardColumnSaveError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("update board column: %w", err)
	}
	if n == 0 {
		return domain.ErrBoardColumnNotFound
	}
	return nil
}

func (r *BoardColumnRepository) Reorder(ctx context.Context, projectID string, columnIDs []string) error {
	for i, columnID := range columnIDs {
		n, err := r.client.BoardColumn.Update().
			Where(
				entboardcolumn.IDEQ(columnID),
				entboardcolumn.ProjectIDEQ(projectID),
			).
			SetSortOrder(i).
			SetUpdatedAt(time.Now().UTC()).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("reorder board column %s: %w", columnID, err)
		}
		if n == 0 {
			return domain.ErrBoardColumnNotFound
		}
	}
	return nil
}

func (r *BoardColumnRepository) DeleteByID(ctx context.Context, projectID, columnID string) error {
	n, err := r.client.BoardColumn.Delete().
		Where(
			entboardcolumn.IDEQ(columnID),
			entboardcolumn.ProjectIDEQ(projectID),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete board column: %w", err)
	}
	if n == 0 {
		return domain.ErrBoardColumnNotFound
	}
	return nil
}

func (r *BoardColumnRepository) CountByProjectID(ctx context.Context, projectID string) (int, error) {
	count, err := r.client.BoardColumn.Query().
		Where(entboardcolumn.ProjectIDEQ(projectID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count board columns: %w", err)
	}
	return count, nil
}

func buildBoardColumnFilter(f domain.BoardColumnFilter) []predicate.BoardColumn {
	var preds []predicate.BoardColumn
	if f.ProjectID != nil {
		preds = append(preds, entboardcolumn.ProjectIDEQ(*f.ProjectID))
	}
	if f.ID != nil {
		preds = append(preds, entboardcolumn.IDEQ(*f.ID))
	}
	return preds
}

func toDomainBoardColumn(row *ent.BoardColumn) *domain.BoardColumn {
	return &domain.BoardColumn{
		ID:        row.ID,
		ProjectID: row.ProjectID,
		Title:     row.Title,
		Status:    row.Status,
		WipLimit:  row.WipLimit,
		Order:     row.SortOrder,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toDomainBoardColumns(rows []*ent.BoardColumn) []*domain.BoardColumn {
	out := make([]*domain.BoardColumn, len(rows))
	for i, row := range rows {
		out[i] = toDomainBoardColumn(row)
	}
	return out
}
