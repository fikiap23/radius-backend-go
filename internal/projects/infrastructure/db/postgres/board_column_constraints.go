package postgres

import (
	"errors"
	"strings"

	"github.com/lib/pq"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/projects/domain"
)

func mapBoardColumnSaveError(err error) error {
	if !ent.IsConstraintError(err) {
		return err
	}
	if isUniqueBoardColumnStatus(err) {
		return domain.ErrBoardColumnStatusExists
	}
	return err
}

func isUniqueBoardColumnStatus(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if strings.Contains(pqErr.Constraint, "project_id_status") ||
			strings.Contains(pqErr.Constraint, "status") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "status") && strings.Contains(msg, "unique")
}
