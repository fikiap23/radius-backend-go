package domain

import "errors"

var (
	ErrWorkspaceNotFound            = errors.New("workspace not found")
	ErrWorkspaceSlugAlreadyExists   = errors.New("workspace slug already exists")
	ErrWorkspaceForbidden           = errors.New("workspace forbidden")
	ErrWorkspaceMemberNotFound      = errors.New("workspace member not found")
	ErrWorkspaceMemberAlreadyExists = errors.New("workspace member already exists")
	ErrInvalidMemberRole            = errors.New("invalid member role")
)
