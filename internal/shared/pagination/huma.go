package pagination

import "strings"

// HTTPQuery is embeddable in Huma list-operation input structs.
//
// Example:
//
//	type ListUsersInput struct {
//	    pagination.HTTPQuery
//	}
//
//	func (in ListUsersInput) Params() pagination.Params {
//	    return in.HTTPQuery.ParamsWithSort("createdAt", "createdAt", "updatedAt", "name", "email")
//	}
type HTTPQuery struct {
	Page    int    `query:"page" default:"1" minimum:"1" doc:"Page number (1-based)"`
	PerPage int    `query:"perPage" default:"20" minimum:"1" maximum:"100" doc:"Items per page"`
	Search  string `query:"search" doc:"Search filter (resource-specific fields)"`
	SortBy  string `query:"sortBy" default:"createdAt" doc:"Sort field"`
	SortDir string `query:"sortDir" default:"desc" enum:"asc,desc" doc:"Sort direction"`
}

// Params converts query params with pagination defaults only (no sort allowlist).
func (q HTTPQuery) Params() Params {
	return Params{
		Page:    q.Page,
		PerPage: q.PerPage,
		Search:  strings.TrimSpace(q.Search),
		Sort:    NormalizeSort(q.SortBy, ParseSortDir(q.SortDir), "createdAt", "createdAt"),
	}.Normalize()
}

// ParamsWithSort converts query params and validates sortBy against allowed fields.
func (q HTTPQuery) ParamsWithSort(defaultBy string, allowed ...string) Params {
	return Params{
		Page:    q.Page,
		PerPage: q.PerPage,
		Search:  strings.TrimSpace(q.Search),
		Sort:    NormalizeSort(q.SortBy, ParseSortDir(q.SortDir), defaultBy, allowed...),
	}.Normalize()
}
