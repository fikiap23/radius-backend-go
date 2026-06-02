package domain

import "time"

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
	MemberRoleViewer MemberRole = "viewer"
)

type MemberStatus string

const (
	MemberStatusActive  MemberStatus = "active"
	MemberStatusPending MemberStatus = "pending"
)

type WorkspaceMember struct {
	ID          string
	WorkspaceID string
	UserID      *string
	Name        string
	Email       string
	Role        MemberRole
	Status      MemberStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMemberUpdate struct {
	UserID *string
	Name   *string
	Email  *string
	Role   *MemberRole
	Status *MemberStatus
}
