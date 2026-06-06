package domain

import "errors"

var (
	ErrProjectNotFound    = errors.New("project not found")
	ErrProjectForbidden   = errors.New("project forbidden")
	ErrInvalidCover       = errors.New("invalid project cover")
	ErrInvalidStatus      = errors.New("invalid project status")
	ErrProjectArchived    = errors.New("project is already archived")
	ErrProjectNotArchived = errors.New("project is not archived")
)
