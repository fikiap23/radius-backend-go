package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/radius/radius-backend/internal/projects/application/dto"
	"github.com/radius/radius-backend/internal/projects/domain"
	"github.com/radius/radius-backend/internal/shared/storage"
	storagedomain "github.com/radius/radius-backend/internal/storage/domain"
	wsdomain "github.com/radius/radius-backend/internal/workspaces/domain"
	"go.uber.org/zap"
)

type ProjectService struct {
	projectRepo   domain.ProjectRepository
	memberRepo    wsdomain.WorkspaceMemberRepository
	objectStorage storagedomain.ObjectStorage
	logger        *zap.Logger
}

func NewProjectService(
	projectRepo domain.ProjectRepository,
	memberRepo wsdomain.WorkspaceMemberRepository,
	objectStorage storagedomain.ObjectStorage,
	logger *zap.Logger,
) *ProjectService {
	return &ProjectService{
		projectRepo:   projectRepo,
		memberRepo:    memberRepo,
		objectStorage: objectStorage,
		logger:        logger,
	}
}

func (s *ProjectService) HandleListProjects(
	ctx context.Context,
	userID, workspaceID string,
) ([]dto.ProjectResponse, error) {
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	rows, err := s.projectRepo.FindMany(ctx, domain.ProjectQuery{
		Filter: domain.ProjectFilter{WorkspaceID: &workspaceID},
	})
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return dto.MapProjects(rows), nil
}

func (s *ProjectService) HandleCreateProject(
	ctx context.Context,
	userID, workspaceID string,
	in *dto.CreateProjectInput,
) (*dto.ProjectResponse, error) {
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	coverImageURL, err := s.resolveCoverImageURL(ctx, in.Body.CoverImageTempKey, in.Body.CoverImageURL)
	if err != nil {
		return nil, err
	}

	project := &domain.Project{
		ID:            uuid.NewString(),
		WorkspaceID:   workspaceID,
		Name:          strings.TrimSpace(in.Body.Name),
		Description:   in.Body.Description,
		Icon:          "🚀",
		Cover:         domain.ProjectCoverEmerald,
		CoverImageURL: coverImageURL,
		Status:        domain.ProjectStatusActive,
	}

	if in.Body.Icon != nil {
		project.Icon = *in.Body.Icon
	}
	if in.Body.Cover != nil {
		project.Cover = domain.ProjectCover(*in.Body.Cover)
	}
	if in.Body.Status != nil {
		project.Status = domain.ProjectStatus(*in.Body.Status)
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	out := dto.MapProject(project)
	return &out, nil
}

func (s *ProjectService) HandleUpdateProject(
	ctx context.Context,
	userID, projectID string,
	data domain.ProjectUpdateData,
) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	if err := s.requireWorkspaceMember(ctx, project.WorkspaceID, userID); err != nil {
		return nil, err
	}

	coverImageURL, err := s.resolveCoverImageURL(ctx, data.CoverImageTempKey, data.CoverImageURL)
	if err != nil {
		return nil, err
	}
	if coverImageURL != nil {
		data.CoverImageURL = coverImageURL
	}
	data.CoverImageTempKey = nil

	if err := s.projectRepo.UpdateByID(ctx, projectID, data); err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("update project: %w", err)
	}

	updated, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("reload project: %w", err)
	}
	out := dto.MapProject(updated)
	return &out, nil
}

func (s *ProjectService) resolveCoverImageURL(ctx context.Context, tempKey, directURL *string) (*string, error) {
	if key := storage.TrimTempKey(tempKey); key != "" {
		publicURL, err := s.objectStorage.PromoteProjectCover(ctx, key)
		if err != nil {
			return nil, err
		}
		return &publicURL, nil
	}
	return directURL, nil
}

func (s *ProjectService) HandleToggleFavorite(
	ctx context.Context,
	userID, projectID string,
) (*dto.FavoriteResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	if err := s.requireWorkspaceMember(ctx, project.WorkspaceID, userID); err != nil {
		return nil, err
	}

	newFav := !project.IsFavorite
	if err := s.projectRepo.SetFavorite(ctx, projectID, newFav); err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("toggle favorite: %w", err)
	}

	return &dto.FavoriteResponse{ID: projectID, IsFavorite: newFav}, nil
}

func (s *ProjectService) HandleArchive(
	ctx context.Context,
	userID, projectID string,
) (*dto.ArchiveResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	if err := s.requireWorkspaceMember(ctx, project.WorkspaceID, userID); err != nil {
		return nil, err
	}

	if project.ArchivedAt != nil {
		return nil, domain.ErrProjectArchived
	}

	if err := s.projectRepo.Archive(ctx, projectID); err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("archive project: %w", err)
	}

	updated, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("reload project: %w", err)
	}

	return &dto.ArchiveResponse{ID: projectID, ArchivedAt: updated.ArchivedAt}, nil
}

func (s *ProjectService) HandleUnarchive(
	ctx context.Context,
	userID, projectID string,
) (*dto.ArchiveResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	if err := s.requireWorkspaceMember(ctx, project.WorkspaceID, userID); err != nil {
		return nil, err
	}

	if project.ArchivedAt == nil {
		return nil, domain.ErrProjectNotArchived
	}

	if err := s.projectRepo.Unarchive(ctx, projectID); err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("unarchive project: %w", err)
	}

	return &dto.ArchiveResponse{ID: projectID, ArchivedAt: nil}, nil
}

func (s *ProjectService) HandleDeleteProject(
	ctx context.Context,
	userID, projectID string,
) (*dto.OkResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("load project: %w", err)
	}

	if err := s.requireWorkspaceMember(ctx, project.WorkspaceID, userID); err != nil {
		return nil, err
	}

	if err := s.projectRepo.DeleteByID(ctx, projectID); err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("delete project: %w", err)
	}

	return &dto.OkResponse{OK: true}, nil
}

func (s *ProjectService) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	active := wsdomain.MemberStatusActive
	_, err := s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &workspaceID,
			UserID:      &userID,
			Status:      &active,
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return domain.ErrProjectForbidden
		}
		return fmt.Errorf("check workspace membership: %w", err)
	}
	return nil
}
