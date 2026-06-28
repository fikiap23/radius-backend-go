package postgres

import (
	"github.com/radius/radius-backend/ent"
	enttask "github.com/radius/radius-backend/ent/task"
	"github.com/radius/radius-backend/internal/tasks/domain"
)

func toDomainTask(row *ent.Task) *domain.Task {
	labelIDs := row.LabelIds
	if labelIDs == nil {
		labelIDs = []string{}
	}
	return &domain.Task{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		WorkspaceID: row.WorkspaceID,
		Title:       row.Title,
		Description: row.Description,
		Status:      domain.TaskStatus(row.Status),
		ColumnID:    row.ColumnID,
		Priority:    domain.TaskPriority(row.Priority),
		DueAt:       row.DueAt,
		LabelIDs:    labelIDs,
		AssigneeID:  row.AssigneeID,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainTaskWithChildren(row *ent.Task) *domain.Task {
	t := toDomainTask(row)
	if edges := row.Edges.Subtasks; edges != nil {
		t.Subtasks = toDomainSubtasks(edges)
	} else {
		t.Subtasks = []*domain.TaskSubtask{}
	}
	if edges := row.Edges.ChecklistItems; edges != nil {
		t.Checklist = toDomainChecklistItems(edges)
	} else {
		t.Checklist = []*domain.TaskChecklistItem{}
	}
	if edges := row.Edges.Attachments; edges != nil {
		t.Attachments = toDomainAttachments(edges)
	} else {
		t.Attachments = []*domain.TaskAttachment{}
	}
	return t
}

func toDomainTasks(rows []*ent.Task) []*domain.Task {
	out := make([]*domain.Task, len(rows))
	for i, row := range rows {
		out[i] = toDomainTaskWithChildren(row)
	}
	return out
}

func toDomainSubtask(row *ent.TaskSubtask) *domain.TaskSubtask {
	return &domain.TaskSubtask{
		ID:     row.ID,
		TaskID: row.TaskID,
		Title:  row.Title,
		Done:   row.Done,
	}
}

func toDomainSubtasks(rows []*ent.TaskSubtask) []*domain.TaskSubtask {
	out := make([]*domain.TaskSubtask, len(rows))
	for i, row := range rows {
		out[i] = toDomainSubtask(row)
	}
	return out
}

func toDomainChecklistItem(row *ent.TaskChecklistItem) *domain.TaskChecklistItem {
	return &domain.TaskChecklistItem{
		ID:      row.ID,
		TaskID:  row.TaskID,
		Text:    row.Text,
		Checked: row.Checked,
	}
}

func toDomainChecklistItems(rows []*ent.TaskChecklistItem) []*domain.TaskChecklistItem {
	out := make([]*domain.TaskChecklistItem, len(rows))
	for i, row := range rows {
		out[i] = toDomainChecklistItem(row)
	}
	return out
}

func toDomainAttachment(row *ent.TaskAttachment) *domain.TaskAttachment {
	return &domain.TaskAttachment{
		ID:         row.ID,
		TaskID:     row.TaskID,
		Name:       row.Name,
		Size:       row.Size,
		MimeType:   row.MimeType,
		StorageKey: row.StorageKey,
		URL:        row.URL,
		UploadedAt: row.UploadedAt,
	}
}

func toDomainAttachments(rows []*ent.TaskAttachment) []*domain.TaskAttachment {
	out := make([]*domain.TaskAttachment, len(rows))
	for i, row := range rows {
		out[i] = toDomainAttachment(row)
	}
	return out
}

func toDomainActivityLog(row *ent.TaskActivityLog) *domain.TaskActivityLog {
	return &domain.TaskActivityLog{
		ID:          row.ID,
		TaskID:      row.TaskID,
		Title:       row.Title,
		Description: row.Description,
		Icon:        row.Icon,
		OccurredAt:  row.OccurredAt,
	}
}

func toDomainActivityLogs(rows []*ent.TaskActivityLog) []*domain.TaskActivityLog {
	out := make([]*domain.TaskActivityLog, len(rows))
	for i, row := range rows {
		out[i] = toDomainActivityLog(row)
	}
	return out
}

func toEntStatus(s domain.TaskStatus) enttask.Status {
	return enttask.Status(s)
}

func toEntPriority(p domain.TaskPriority) enttask.Priority {
	return enttask.Priority(p)
}
