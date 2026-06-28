package dto

import (
	"strings"
	"time"

	"github.com/radius/radius-backend/internal/tasks/domain"
)

type SubtaskResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type ChecklistItemResponse struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

type AttachmentResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mimeType"`
	URL        string    `json:"url"`
	UploadedAt time.Time `json:"uploadedAt"`
}

type TaskResponse struct {
	ID          string                  `json:"id"`
	ProjectID   string                  `json:"projectId"`
	WorkspaceID string                  `json:"workspaceId"`
	Title       string                  `json:"title"`
	Description *string                 `json:"description,omitempty"`
	Status      domain.TaskStatus       `json:"status"`
	ColumnID    *string                 `json:"columnId"`
	Priority    domain.TaskPriority     `json:"priority"`
	DueAt       *time.Time              `json:"dueAt"`
	LabelIDs    []string                `json:"labelIds"`
	AssigneeID  *string                 `json:"assigneeId"`
	Subtasks    []SubtaskResponse       `json:"subtasks"`
	Checklist   []ChecklistItemResponse `json:"checklist"`
	Attachments []AttachmentResponse    `json:"attachments"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
}

type TaskPatchResponse struct {
	ID        string                  `json:"id"`
	Status    *domain.TaskStatus      `json:"status,omitempty"`
	ColumnID  *string                 `json:"columnId,omitempty"`
	Subtasks  []SubtaskResponse       `json:"subtasks,omitempty"`
	Checklist []ChecklistItemResponse `json:"checklist,omitempty"`
}

type ActivityResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	OccurredAt  time.Time `json:"occurredAt"`
	Icon        string    `json:"icon"`
}

type OkResponse struct {
	OK bool `json:"ok"`
}

func MapTask(task *domain.Task) TaskResponse {
	labelIDs := task.LabelIDs
	if labelIDs == nil {
		labelIDs = []string{}
	}
	return TaskResponse{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		WorkspaceID: task.WorkspaceID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		ColumnID:    task.ColumnID,
		Priority:    task.Priority,
		DueAt:       task.DueAt,
		LabelIDs:    labelIDs,
		AssigneeID:  task.AssigneeID,
		Subtasks:    MapSubtasks(task.Subtasks),
		Checklist:   MapChecklistItems(task.Checklist),
		Attachments: MapAttachments(task.Attachments),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func MapTasks(tasks []*domain.Task) []TaskResponse {
	out := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		out[i] = MapTask(task)
	}
	return out
}

func MapSubtasks(items []*domain.TaskSubtask) []SubtaskResponse {
	if len(items) == 0 {
		return []SubtaskResponse{}
	}
	out := make([]SubtaskResponse, len(items))
	for i, item := range items {
		out[i] = SubtaskResponse{
			ID:    item.ID,
			Title: item.Title,
			Done:  item.Done,
		}
	}
	return out
}

func MapChecklistItems(items []*domain.TaskChecklistItem) []ChecklistItemResponse {
	if len(items) == 0 {
		return []ChecklistItemResponse{}
	}
	out := make([]ChecklistItemResponse, len(items))
	for i, item := range items {
		out[i] = ChecklistItemResponse{
			ID:      item.ID,
			Text:    item.Text,
			Checked: item.Checked,
		}
	}
	return out
}

func MapAttachments(items []*domain.TaskAttachment) []AttachmentResponse {
	if len(items) == 0 {
		return []AttachmentResponse{}
	}
	out := make([]AttachmentResponse, len(items))
	for i, item := range items {
		out[i] = MapAttachment(item)
	}
	return out
}

func MapAttachment(item *domain.TaskAttachment) AttachmentResponse {
	return AttachmentResponse{
		ID:         item.ID,
		Name:       item.Name,
		Size:       item.Size,
		MimeType:   item.MimeType,
		URL:        item.URL,
		UploadedAt: item.UploadedAt,
	}
}

func MapActivity(log *domain.TaskActivityLog) ActivityResponse {
	return ActivityResponse{
		ID:          log.ID,
		Title:       log.Title,
		Description: log.Description,
		OccurredAt:  log.OccurredAt,
		Icon:        log.Icon,
	}
}

func MapActivities(logs []*domain.TaskActivityLog) []ActivityResponse {
	out := make([]ActivityResponse, len(logs))
	for i, log := range logs {
		out[i] = MapActivity(log)
	}
	return out
}

// --- Huma input types ---

type ProjectTasksPathInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
}

type CreateTaskInput struct {
	ProjectID string `path:"projectId" doc:"Project ID" format:"uuid"`
	Body      struct {
		Title       string              `json:"title" minLength:"1" maxLength:"500"`
		Description *string             `json:"description,omitempty"`
		Status      domain.TaskStatus    `json:"status" enum:"backlog,todo,in_progress,review,done"`
		ColumnID    *string              `json:"columnId,omitempty" format:"uuid"`
		Priority    *domain.TaskPriority `json:"priority,omitempty" enum:"low,medium,high,urgent"`
		DueAt       *time.Time           `json:"dueAt,omitempty"`
		LabelIDs    []string             `json:"labelIds,omitempty"`
		AssigneeID  *string             `json:"assigneeId,omitempty" format:"uuid"`
	}
}

type TaskPathInput struct {
	TaskID string `path:"taskId" doc:"Task ID" format:"uuid"`
}

type UpdateTaskInput struct {
	TaskID string `path:"taskId" doc:"Task ID" format:"uuid"`
	Body   struct {
		Title       *string                `json:"title,omitempty" minLength:"1" maxLength:"500"`
		Description *string                `json:"description,omitempty"`
		Status      *domain.TaskStatus     `json:"status,omitempty" enum:"backlog,todo,in_progress,review,done"`
		ColumnID    *string                `json:"columnId,omitempty" format:"uuid"`
		Priority    *domain.TaskPriority   `json:"priority,omitempty" enum:"low,medium,high,urgent"`
		DueAt       *time.Time             `json:"dueAt"`
		LabelIDs    *[]string              `json:"labelIds,omitempty"`
		AssigneeID  *string                `json:"assigneeId,omitempty" format:"uuid"`
		Subtasks    *[]SubtaskPatchInput   `json:"subtasks,omitempty"`
		Checklist   *[]ChecklistPatchInput `json:"checklist,omitempty"`
	}
}

type SubtaskPatchInput struct {
	ID    string `json:"id,omitempty" format:"uuid"`
	Title string `json:"title" minLength:"1" maxLength:"500"`
	Done  bool   `json:"done"`
}

type ChecklistPatchInput struct {
	ID      string `json:"id,omitempty" format:"uuid"`
	Text    string `json:"text" minLength:"1" maxLength:"500"`
	Checked bool   `json:"checked"`
}

type ConfirmAttachmentInput struct {
	TaskID string `path:"taskId" doc:"Task ID" format:"uuid"`
	Body   struct {
		TempKey     string `json:"tempKey" minLength:"1"`
		FileName    string `json:"fileName" minLength:"1" maxLength:"255"`
		ContentType string `json:"contentType" minLength:"1" maxLength:"120"`
		Size        int64  `json:"size" minimum:"1"`
	}
}

type DeleteAttachmentPathInput struct {
	TaskID       string `path:"taskId" doc:"Task ID" format:"uuid"`
	AttachmentID string `path:"attachmentId" doc:"Attachment ID" format:"uuid"`
}

func (in CreateTaskInput) ToDomain(projectID, workspaceID string) domain.Task {
	status := in.Body.Status
	if !status.Valid() {
		status = domain.TaskStatusTodo
	}
	priority := domain.TaskPriorityMedium
	if in.Body.Priority != nil && in.Body.Priority.Valid() {
		priority = *in.Body.Priority
	}
	labelIDs := in.Body.LabelIDs
	if labelIDs == nil {
		labelIDs = []string{}
	}
	return domain.Task{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		Title:       strings.TrimSpace(in.Body.Title),
		Description: in.Body.Description,
		Status:      status,
		ColumnID:    in.Body.ColumnID,
		Priority:    priority,
		DueAt:       in.Body.DueAt,
		LabelIDs:    labelIDs,
		AssigneeID:  in.Body.AssigneeID,
	}
}
