package domain

import "time"

type Workspace struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WorkspaceUpdate struct {
	Name *string
	Slug *string
}
