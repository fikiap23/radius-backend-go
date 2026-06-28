package dto

import (
	"strings"

	"github.com/radius/radius-backend/internal/projects/domain"
)

type BoardColumnResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	WipLimit *int   `json:"wipLimit,omitempty"`
	Order    int    `json:"order"`
}

func MapBoardColumn(col *domain.BoardColumn) BoardColumnResponse {
	return BoardColumnResponse{
		ID:       col.ID,
		Title:    col.Title,
		Status:   col.Status,
		WipLimit: col.WipLimit,
		Order:    col.Order,
	}
}

func MapBoardColumns(rows []*domain.BoardColumn) []BoardColumnResponse {
	out := make([]BoardColumnResponse, len(rows))
	for i, row := range rows {
		out[i] = MapBoardColumn(row)
	}
	return out
}

type DeleteBoardColumnResponse struct {
	OK               bool   `json:"ok"`
	FallbackColumnID string `json:"fallbackColumnId"`
}

// --- Huma input types ---

type ProjectBoardColumnsPathInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
}

type CreateBoardColumnInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
	Body      struct {
		Title    string `json:"title" doc:"Column title" minLength:"1" maxLength:"255"`
		Status   string `json:"status" doc:"Column status slug" minLength:"1" maxLength:"64" pattern:"^[a-z][a-z0-9_]*$"`
		WipLimit *int   `json:"wipLimit" required:"false" nullable:"true" doc:"Work-in-progress limit" minimum:"0"`
	}
}

func (in *CreateBoardColumnInput) ToDomain(projectID string) domain.BoardColumn {
	return domain.BoardColumn{
		ProjectID: projectID,
		Title:     strings.TrimSpace(in.Body.Title),
		Status:    strings.TrimSpace(in.Body.Status),
		WipLimit:  in.Body.WipLimit,
	}
}

type BoardColumnPathInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
	ColumnID  string `path:"columnId" doc:"Board column ID" format:"uuid"`
}

type UpdateBoardColumnInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
	ColumnID  string `path:"columnId" doc:"Board column ID" format:"uuid"`
	Body      struct {
		Title    *string `json:"title,omitempty" required:"false" doc:"Column title" minLength:"1" maxLength:"255"`
		Status   *string `json:"status,omitempty" required:"false" doc:"Column status slug" minLength:"1" maxLength:"64" pattern:"^[a-z][a-z0-9_]*$"`
		WipLimit *int    `json:"wipLimit" required:"false" nullable:"true" doc:"Work-in-progress limit" minimum:"0"`
	}
}

func (in *UpdateBoardColumnInput) ToDomain() domain.BoardColumnUpdateData {
	data := domain.BoardColumnUpdateData{
		WipLimit: in.Body.WipLimit,
	}
	if in.Body.Title != nil {
		title := strings.TrimSpace(*in.Body.Title)
		data.Title = &title
	}
	if in.Body.Status != nil {
		status := strings.TrimSpace(*in.Body.Status)
		data.Status = &status
	}
	return data
}

type ReorderBoardColumnsInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
	Body      struct {
		ColumnIDs []string `json:"columnIds" doc:"Ordered column IDs" minItems:"1"`
	}
}
