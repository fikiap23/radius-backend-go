package postgres

import (
	"github.com/radius/radius-backend/ent"
	entwm "github.com/radius/radius-backend/ent/workspacemember"
	"github.com/radius/radius-backend/internal/workspaces/domain"
)

func toDomainWorkspace(row *ent.Workspace) *domain.Workspace {
	return &domain.Workspace{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toDomainWorkspaces(rows []*ent.Workspace) []*domain.Workspace {
	out := make([]*domain.Workspace, len(rows))
	for i, row := range rows {
		out[i] = toDomainWorkspace(row)
	}
	return out
}

func toDomainWorkspaceMember(row *ent.WorkspaceMember) *domain.WorkspaceMember {
	return &domain.WorkspaceMember{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		UserID:      row.UserID,
		Name:        row.Name,
		Email:       row.Email,
		Role:        domain.MemberRole(row.Role),
		Status:      domain.MemberStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainWorkspaceMembers(rows []*ent.WorkspaceMember) []*domain.WorkspaceMember {
	out := make([]*domain.WorkspaceMember, len(rows))
	for i, row := range rows {
		out[i] = toDomainWorkspaceMember(row)
	}
	return out
}

func toEntMemberRole(role domain.MemberRole) entwm.Role {
	return entwm.Role(role)
}

func toEntMemberStatus(status domain.MemberStatus) entwm.Status {
	return entwm.Status(status)
}
