package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/projects/application/dto"
	"github.com/radius/radius-backend/internal/projects/domain"
	wsdomain "github.com/radius/radius-backend/internal/workspaces/domain"
)

type BoardColumnService struct {
	columnRepo       domain.BoardColumnRepository
	projectRepo      domain.ProjectRepository
	memberRepo       wsdomain.WorkspaceMemberRepository
	runInTransaction domain.RunProjectsInTransactionFunc
}

func NewBoardColumnService(
	columnRepo domain.BoardColumnRepository,
	projectRepo domain.ProjectRepository,
	memberRepo wsdomain.WorkspaceMemberRepository,
	runInTransaction domain.RunProjectsInTransactionFunc,
) *BoardColumnService {
	return &BoardColumnService{
		columnRepo:       columnRepo,
		projectRepo:      projectRepo,
		memberRepo:       memberRepo,
		runInTransaction: runInTransaction,
	}
}

func (s *BoardColumnService) HandleListColumns(
	ctx context.Context,
	userID, projectID string,
) ([]dto.BoardColumnResponse, error) {
	if _, err := s.requireProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	rows, err := s.columnRepo.FindMany(ctx, domain.BoardColumnQuery{
		Filter: domain.BoardColumnFilter{ProjectID: &projectID},
	})
	if err != nil {
		return nil, fmt.Errorf("list board columns: %w", err)
	}
	return dto.MapBoardColumns(rows), nil
}

func (s *BoardColumnService) HandleCreateColumn(
	ctx context.Context,
	userID, projectID string,
	in *dto.CreateBoardColumnInput,
) (*dto.BoardColumnResponse, error) {
	if _, err := s.requireProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	existing, err := s.columnRepo.FindMany(ctx, domain.BoardColumnQuery{
		Filter: domain.BoardColumnFilter{ProjectID: &projectID},
	})
	if err != nil {
		return nil, fmt.Errorf("load board columns: %w", err)
	}

	col := in.ToDomain(projectID)
	col.ID = uuid.NewString()
	col.Order = nextColumnOrder(existing)

	if err := s.columnRepo.Create(ctx, &col); err != nil {
		if errors.Is(err, domain.ErrBoardColumnStatusExists) {
			return nil, domain.ErrBoardColumnStatusExists
		}
		return nil, fmt.Errorf("create board column: %w", err)
	}

	out := dto.MapBoardColumn(&col)
	return &out, nil
}

func (s *BoardColumnService) HandleUpdateColumn(
	ctx context.Context,
	userID, projectID, columnID string,
	data domain.BoardColumnUpdateData,
) (*dto.BoardColumnResponse, error) {
	if _, err := s.requireProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	if err := s.columnRepo.UpdateByID(ctx, projectID, columnID, data); err != nil {
		if errors.Is(err, domain.ErrBoardColumnNotFound) {
			return nil, domain.ErrBoardColumnNotFound
		}
		if errors.Is(err, domain.ErrBoardColumnStatusExists) {
			return nil, domain.ErrBoardColumnStatusExists
		}
		return nil, fmt.Errorf("update board column: %w", err)
	}

	updated, err := s.columnRepo.FindByID(ctx, projectID, columnID)
	if err != nil {
		return nil, fmt.Errorf("reload board column: %w", err)
	}
	out := dto.MapBoardColumn(updated)
	return &out, nil
}

func (s *BoardColumnService) HandleReorderColumns(
	ctx context.Context,
	userID, projectID string,
	columnIDs []string,
) (*dto.OkResponse, error) {
	if _, err := s.requireProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	existing, err := s.columnRepo.FindMany(ctx, domain.BoardColumnQuery{
		Filter: domain.BoardColumnFilter{ProjectID: &projectID},
	})
	if err != nil {
		return nil, fmt.Errorf("load board columns: %w", err)
	}

	if err := validateReorder(existing, columnIDs); err != nil {
		return nil, err
	}

	if err := s.runInTransaction(ctx, func(ctx context.Context, repos domain.ProjectsRepositories) error {
		return repos.BoardColumns.Reorder(ctx, projectID, columnIDs)
	}); err != nil {
		if errors.Is(err, domain.ErrBoardColumnNotFound) {
			return nil, domain.ErrBoardColumnInvalidReorder
		}
		return nil, fmt.Errorf("reorder board columns: %w", err)
	}

	return &dto.OkResponse{OK: true}, nil
}

func (s *BoardColumnService) HandleDeleteColumn(
	ctx context.Context,
	userID, projectID, columnID string,
) (*dto.DeleteBoardColumnResponse, error) {
	if _, err := s.requireProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	count, err := s.columnRepo.CountByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("count board columns: %w", err)
	}
	if count <= 1 {
		return nil, domain.ErrBoardColumnLastColumn
	}

	existing, err := s.columnRepo.FindMany(ctx, domain.BoardColumnQuery{
		Filter: domain.BoardColumnFilter{ProjectID: &projectID},
	})
	if err != nil {
		return nil, fmt.Errorf("load board columns: %w", err)
	}

	fallbackID, err := resolveFallbackColumnID(existing, columnID)
	if err != nil {
		return nil, err
	}

	if err := s.columnRepo.DeleteByID(ctx, projectID, columnID); err != nil {
		if errors.Is(err, domain.ErrBoardColumnNotFound) {
			return nil, domain.ErrBoardColumnNotFound
		}
		return nil, fmt.Errorf("delete board column: %w", err)
	}

	return &dto.DeleteBoardColumnResponse{
		OK:               true,
		FallbackColumnID: fallbackID,
	}, nil
}

func (s *BoardColumnService) requireProjectAccess(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	active := wsdomain.MemberStatusActive
	_, err = s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &project.WorkspaceID,
			UserID:      &userID,
			Status:      &active,
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return nil, domain.ErrProjectForbidden
		}
		return nil, fmt.Errorf("check workspace membership: %w", err)
	}
	return project, nil
}

func nextColumnOrder(columns []*domain.BoardColumn) int {
	maxOrder := -1
	for _, col := range columns {
		if col.Order > maxOrder {
			maxOrder = col.Order
		}
	}
	return maxOrder + 1
}

func validateReorder(existing []*domain.BoardColumn, columnIDs []string) error {
	if len(columnIDs) != len(existing) {
		return domain.ErrBoardColumnInvalidReorder
	}

	known := make(map[string]struct{}, len(existing))
	for _, col := range existing {
		known[col.ID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(columnIDs))
	for _, id := range columnIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return domain.ErrBoardColumnInvalidReorder
		}
		if _, ok := known[id]; !ok {
			return domain.ErrBoardColumnInvalidReorder
		}
		if _, dup := seen[id]; dup {
			return domain.ErrBoardColumnInvalidReorder
		}
		seen[id] = struct{}{}
	}
	return nil
}

func resolveFallbackColumnID(columns []*domain.BoardColumn, deleteID string) (string, error) {
	var target *domain.BoardColumn
	for i, col := range columns {
		if col.ID != deleteID {
			continue
		}
		if i > 0 {
			target = columns[i-1]
		} else if i+1 < len(columns) {
			target = columns[i+1]
		}
		break
	}
	if target == nil {
		return "", domain.ErrBoardColumnNotFound
	}
	return target.ID, nil
}
