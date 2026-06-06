package postgres

import (
	"github.com/radius/radius-backend/ent"
	entproject "github.com/radius/radius-backend/ent/project"
	"github.com/radius/radius-backend/internal/projects/domain"
)

func toDomainProject(row *ent.Project) *domain.Project {
	return &domain.Project{
		ID:            row.ID,
		WorkspaceID:   row.WorkspaceID,
		Name:          row.Name,
		Description:   row.Description,
		Icon:          row.Icon,
		Cover:         domain.ProjectCover(row.Cover),
		CoverImageURL: row.CoverImageURL,
		Status:        domain.ProjectStatus(row.Status),
		IsFavorite:    row.IsFavorite,
		ArchivedAt:    row.ArchivedAt,
		OpenTasks:     row.OpenTasks,
		Progress:      row.Progress,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func toDomainProjects(rows []*ent.Project) []*domain.Project {
	out := make([]*domain.Project, len(rows))
	for i, row := range rows {
		out[i] = toDomainProject(row)
	}
	return out
}

func toEntCover(c domain.ProjectCover) entproject.Cover {
	return entproject.Cover(c)
}

func toEntStatus(s domain.ProjectStatus) entproject.Status {
	return entproject.Status(s)
}
