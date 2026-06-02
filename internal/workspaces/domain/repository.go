package domain

import (
	"context"

	"github.com/radius/radius-backend/internal/shared/pagination"
)

// WorkspaceFields represents a named column selection preset for workspace queries.
type WorkspaceFields int

const (
	WorkspaceFieldsAll     WorkspaceFields = iota // all columns
	WorkspaceFieldsProfile                       // id, name, slug, timestamps
	WorkspaceFieldsExists                        // id only
)

type WorkspaceFilter struct {
	ID     *string
	Slug   *string
	UserID *string // workspaces where user is an active member
	Search string  // partial match on name or slug (case-insensitive)
}

type WorkspaceQuery struct {
	Select WorkspaceFields
	Filter WorkspaceFilter
}

type WorkspaceUpdateData struct {
	Name *string
	Slug *string
}

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	FindByID(ctx context.Context, id string, fields ...WorkspaceFields) (*Workspace, error)
	FindBySlug(ctx context.Context, slug string, fields ...WorkspaceFields) (*Workspace, error)
	FindOne(ctx context.Context, q WorkspaceQuery) (*Workspace, error)
	FindMany(ctx context.Context, q WorkspaceQuery) ([]*Workspace, error)
	FindManyPaginate(ctx context.Context, q WorkspaceQuery, params pagination.Params) (*pagination.Result[*Workspace], error)
	UpdateByID(ctx context.Context, id string, data WorkspaceUpdateData) error
	DeleteByID(ctx context.Context, id string) error
}

// WorkspaceMemberFields represents a named column selection preset for member queries.
type WorkspaceMemberFields int

const (
	WorkspaceMemberFieldsAll     WorkspaceMemberFields = iota // all columns
	WorkspaceMemberFieldsProfile                             // list/detail fields
	WorkspaceMemberFieldsExists                              // id only
)

type WorkspaceMemberFilter struct {
	ID          *string
	WorkspaceID *string
	UserID      *string
	Email       *string
	Role        *MemberRole
	Status      *MemberStatus
	Search      string // partial match on name or email (case-insensitive)
}

type WorkspaceMemberQuery struct {
	Select WorkspaceMemberFields
	Filter WorkspaceMemberFilter
}

type WorkspaceMemberUpdateData struct {
	UserID *string
	Name   *string
	Email  *string
	Role   *MemberRole
	Status *MemberStatus
}

type WorkspaceMemberRepository interface {
	Create(ctx context.Context, member *WorkspaceMember) error
	FindByID(ctx context.Context, id string, fields ...WorkspaceMemberFields) (*WorkspaceMember, error)
	FindOne(ctx context.Context, q WorkspaceMemberQuery) (*WorkspaceMember, error)
	FindMany(ctx context.Context, q WorkspaceMemberQuery) ([]*WorkspaceMember, error)
	FindManyPaginate(ctx context.Context, q WorkspaceMemberQuery, params pagination.Params) (*pagination.Result[*WorkspaceMember], error)
	UpdateByID(ctx context.Context, id string, data WorkspaceMemberUpdateData) error
	DeleteByID(ctx context.Context, id string) error
}
