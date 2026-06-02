package postgres

import (
	"errors"
	"strings"

	"github.com/lib/pq"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/users/domain"
)

func mapUserCreateError(err error) error {
	if !ent.IsConstraintError(err) {
		return err
	}
	if isUniqueOnEmail(err) {
		return domain.ErrEmailAlreadyExists
	}
	return err
}

func mapOAuthAccountCreateError(err error) error {
	if !ent.IsConstraintError(err) {
		return err
	}
	if isOAuthProviderAccountUnique(err) {
		return domain.ErrOAuthAccountAlreadyExists
	}
	return err
}

func isUniqueOnEmail(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if strings.Contains(pqErr.Constraint, "email") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "email") && strings.Contains(msg, "unique")
}

func isOAuthProviderAccountUnique(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if strings.Contains(pqErr.Constraint, "provider") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "provider") && strings.Contains(msg, "unique")
}
