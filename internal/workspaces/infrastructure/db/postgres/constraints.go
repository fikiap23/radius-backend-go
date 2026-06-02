package postgres

import (
	"errors"
	"strings"

	"github.com/lib/pq"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/workspaces/domain"
)

func mapWorkspaceSaveError(err error) error {
	if !ent.IsConstraintError(err) {
		return err
	}
	if isUniqueOnWorkspaceSlug(err) {
		return domain.ErrWorkspaceSlugAlreadyExists
	}
	return err
}

func mapWorkspaceMemberSaveError(err error) error {
	if !ent.IsConstraintError(err) {
		return err
	}
	if isUniqueWorkspaceMemberEmail(err) {
		return domain.ErrWorkspaceMemberAlreadyExists
	}
	return err
}

func isUniqueOnWorkspaceSlug(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if strings.Contains(pqErr.Constraint, "slug") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "slug") && strings.Contains(msg, "unique")
}

func isUniqueWorkspaceMemberEmail(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if strings.Contains(pqErr.Constraint, "workspace_id_email") ||
			strings.Contains(pqErr.Constraint, "email") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "email") && strings.Contains(msg, "unique")
}
