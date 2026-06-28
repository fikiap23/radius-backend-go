package domain

import "errors"

var (
	ErrTaskNotFound           = errors.New("task not found")
	ErrTaskForbidden          = errors.New("task forbidden")
	ErrTaskInvalidColumn      = errors.New("task invalid column")
	ErrTaskInvalidAssignee    = errors.New("task invalid assignee")
	ErrTaskAttachmentNotFound = errors.New("task attachment not found")
)
