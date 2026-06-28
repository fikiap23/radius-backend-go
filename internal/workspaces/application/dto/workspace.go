package dto

import (
	"strings"
	"time"

	"github.com/radius/radius-backend/internal/workspaces/domain"
)

type WorkspaceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
}

func MapWorkspace(w *domain.Workspace) WorkspaceResponse {
	return WorkspaceResponse{
		ID:        w.ID,
		Name:      w.Name,
		Slug:      w.Slug,
		CreatedAt: w.CreatedAt,
	}
}

func MapWorkspaces(rows []*domain.Workspace) []WorkspaceResponse {
	out := make([]WorkspaceResponse, len(rows))
	for i, row := range rows {
		out[i] = MapWorkspace(row)
	}
	return out
}

type CreateWorkspaceInput struct {
	Body struct {
		Name string `json:"name" doc:"Workspace display name" minLength:"1" maxLength:"255"`
		Slug string `json:"slug" doc:"URL-safe workspace slug" minLength:"1" maxLength:"255" pattern:"^[a-z0-9]+(?:-[a-z0-9]+)*$"`
	}
}

type UpdateWorkspaceInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
	Body        struct {
		Name *string `json:"name,omitempty" doc:"Workspace display name" minLength:"1" maxLength:"255"`
		Slug *string `json:"slug,omitempty" doc:"URL-safe workspace slug" minLength:"1" maxLength:"255" pattern:"^[a-z0-9]+(?:-[a-z0-9]+)*$"`
	}
}

func (in *UpdateWorkspaceInput) ToDomain() domain.WorkspaceUpdateData {
	var slug *string
	if in.Body.Slug != nil {
		s := NormalizeSlug(*in.Body.Slug)
		slug = &s
	}
	return domain.WorkspaceUpdateData{
		Name: in.Body.Name,
		Slug: slug,
	}
}

type WorkspacePathInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
}

type MemberResponse struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspaceId"`
	UserID      *string `json:"userId,omitempty"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
}

func MapMember(m *domain.WorkspaceMember) MemberResponse {
	return MemberResponse{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Name:        m.Name,
		Email:       m.Email,
		Role:        string(m.Role),
		Status:      string(m.Status),
	}
}

func MapMembers(rows []*domain.WorkspaceMember) []MemberResponse {
	out := make([]MemberResponse, len(rows))
	for i, row := range rows {
		out[i] = MapMember(row)
	}
	return out
}

type InviteMemberInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
	Body        struct {
		Email string `json:"email" doc:"Member email" format:"email"`
		Role  string `json:"role" doc:"Member role" enum:"admin,member,viewer"`
	}
}

type UpdateMemberInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
	MemberID    string `path:"memberId" doc:"Member ID" format:"uuid"`
	Body        struct {
		Role string `json:"role" doc:"Member role" enum:"owner,admin,member,viewer"`
	}
}

type MemberPathInput struct {
	WorkspaceID string `path:"workspaceId" doc:"Workspace ID" format:"uuid"`
	MemberID    string `path:"memberId" doc:"Member ID" format:"uuid"`
}

type OkResponse struct {
	OK bool `json:"ok"`
}

func NormalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
