package dto

import (
	"time"

	"github.com/radius/radius-backend/internal/projects/domain"
)

type ProjectResponse struct {
	ID            string     `json:"id"`
	WorkspaceID   string     `json:"workspaceId"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	Icon          string     `json:"icon"`
	Cover         string     `json:"cover"`
	CoverImageURL *string    `json:"coverImageUrl"`
	Status        string     `json:"status"`
	IsFavorite    bool       `json:"isFavorite"`
	ArchivedAt    *time.Time `json:"archivedAt"`
	OpenTasks     int        `json:"openTasks"`
	Progress      int        `json:"progress"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func MapProject(p *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:            p.ID,
		WorkspaceID:   p.WorkspaceID,
		Name:          p.Name,
		Description:   p.Description,
		Icon:          p.Icon,
		Cover:         string(p.Cover),
		CoverImageURL: p.CoverImageURL,
		Status:        string(p.Status),
		IsFavorite:    p.IsFavorite,
		ArchivedAt:    p.ArchivedAt,
		OpenTasks:     p.OpenTasks,
		Progress:      p.Progress,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func MapProjects(rows []*domain.Project) []ProjectResponse {
	out := make([]ProjectResponse, len(rows))
	for i, row := range rows {
		out[i] = MapProject(row)
	}
	return out
}

type FavoriteResponse struct {
	ID         string `json:"id"`
	IsFavorite bool   `json:"isFavorite"`
}

type ArchiveResponse struct {
	ID         string     `json:"id"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

type OkResponse struct {
	OK bool `json:"ok"`
}

// --- Huma input types ---

type WorkspaceProjectsPathInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
}

type CreateProjectInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
	Body        struct {
		Name          string  `json:"name" doc:"Project name" minLength:"1" maxLength:"255"`
		Description   *string `json:"description,omitempty" doc:"Project description (HTML)"`
		Icon          *string `json:"icon,omitempty" doc:"Emoji or icon identifier" minLength:"1"`
		Cover         *string `json:"cover,omitempty" doc:"Cover theme" enum:"emerald,ocean,sunset,violet,rose,slate"`
		CoverImageURL *string `json:"coverImageUrl,omitempty" doc:"Custom cover image URL"`
		Status        *string `json:"status,omitempty" doc:"Project status" enum:"active,on_hold,completed"`
	}
}

type ProjectPathInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
}

type UpdateProjectInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
	Body      struct {
		Name          *string `json:"name,omitempty" doc:"Project name" minLength:"1" maxLength:"255"`
		Description   *string `json:"description,omitempty" doc:"Project description (HTML)"`
		Icon          *string `json:"icon,omitempty" doc:"Emoji or icon identifier" minLength:"1"`
		Cover         *string `json:"cover,omitempty" doc:"Cover theme" enum:"emerald,ocean,sunset,violet,rose,slate"`
		CoverImageURL *string `json:"coverImageUrl,omitempty" doc:"Custom cover image URL"`
		Status        *string `json:"status,omitempty" doc:"Project status" enum:"active,on_hold,completed"`
		IsFavorite    *bool   `json:"isFavorite,omitempty" doc:"Whether project is favorited"`
	}
}

func (in *UpdateProjectInput) ToDomain() domain.ProjectUpdateData {
	data := domain.ProjectUpdateData{
		Name:          in.Body.Name,
		Description:   in.Body.Description,
		Icon:          in.Body.Icon,
		CoverImageURL: in.Body.CoverImageURL,
		IsFavorite:    in.Body.IsFavorite,
	}
	if in.Body.Cover != nil {
		c := domain.ProjectCover(*in.Body.Cover)
		data.Cover = &c
	}
	if in.Body.Status != nil {
		s := domain.ProjectStatus(*in.Body.Status)
		data.Status = &s
	}
	return data
}
