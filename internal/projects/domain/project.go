package domain

import "time"

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
)

type ProjectCover string

const (
	ProjectCoverEmerald ProjectCover = "emerald"
	ProjectCoverOcean   ProjectCover = "ocean"
	ProjectCoverSunset  ProjectCover = "sunset"
	ProjectCoverViolet  ProjectCover = "violet"
	ProjectCoverRose    ProjectCover = "rose"
	ProjectCoverSlate   ProjectCover = "slate"
)

type Project struct {
	ID            string
	WorkspaceID   string
	Name          string
	Description   *string
	Icon          string
	Cover         ProjectCover
	CoverImageURL *string
	Status        ProjectStatus
	IsFavorite    bool
	ArchivedAt    *time.Time
	OpenTasks     int
	Progress      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
