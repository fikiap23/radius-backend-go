package domain

import "time"

type TaskStatus string

const (
	TaskStatusBacklog    TaskStatus = "backlog"
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusBacklog, TaskStatusTodo, TaskStatusInProgress, TaskStatusReview, TaskStatusDone:
		return true
	default:
		return false
	}
}

func (s TaskStatus) IsDone() bool {
	return s == TaskStatusDone
}

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

func (p TaskPriority) Valid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	default:
		return false
	}
}

type Task struct {
	ID          string
	ProjectID   string
	WorkspaceID string
	Title       string
	Description *string
	Status      TaskStatus
	ColumnID    *string
	Priority    TaskPriority
	DueAt       *time.Time
	LabelIDs    []string
	AssigneeID  *string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Subtasks    []*TaskSubtask
	Checklist   []*TaskChecklistItem
	Attachments []*TaskAttachment
}

type TaskSubtask struct {
	ID     string
	TaskID string
	Title  string
	Done   bool
}

type TaskChecklistItem struct {
	ID      string
	TaskID  string
	Text    string
	Checked bool
}

type TaskAttachment struct {
	ID         string
	TaskID     string
	Name       string
	Size       int64
	MimeType   string
	StorageKey string
	URL        string
	UploadedAt time.Time
}

type TaskActivityLog struct {
	ID          string
	TaskID      string
	Title       string
	Description *string
	Icon        string
	OccurredAt  time.Time
}
