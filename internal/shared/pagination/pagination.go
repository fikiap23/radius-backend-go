package pagination

import "strings"

// Defaults and limits for list endpoints.
const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Params holds normalized list query input (page, search, sort).
type Params struct {
	Page    int
	PerPage int
	Search  string
	Sort    Sort
}

// Normalize applies defaults and caps. Safe to call multiple times.
func (p Params) Normalize() Params {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PerPage <= 0 {
		p.PerPage = DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
	p.Search = strings.TrimSpace(p.Search)
	return p
}

// Limit returns the SQL/Ent LIMIT value.
func (p Params) Limit() int {
	return p.Normalize().PerPage
}

// Offset returns the SQL/Ent OFFSET value.
func (p Params) Offset() int {
	n := p.Normalize()
	return (n.Page - 1) * n.PerPage
}

// OffsetLimit is the limit/offset pair for repositories.
type OffsetLimit struct {
	Limit  int
	Offset int
}

// OffsetLimit returns limit and offset derived from page params.
func (p Params) OffsetLimit() OffsetLimit {
	n := p.Normalize()
	return OffsetLimit{
		Limit:  n.PerPage,
		Offset: (n.Page - 1) * n.PerPage,
	}
}

// Meta describes pagination state in API responses.
type Meta struct {
	Page       int     `json:"page"`
	PerPage    int     `json:"perPage"`
	Total      int64   `json:"total"`
	TotalPages int     `json:"totalPages"`
	HasNext    bool    `json:"hasNext"`
	HasPrev    bool    `json:"hasPrev"`
	Search     string  `json:"search,omitempty"`
	SortBy     string  `json:"sortBy,omitempty"`
	SortDir    SortDir `json:"sortDir,omitempty"`
}

// Result is a paginated list with metadata.
type Result[T any] struct {
	Items []T  `json:"items"`
	Meta  Meta `json:"meta"`
}

// TotalPages computes the number of pages for a total count.
func TotalPages(total int64, perPage int) int {
	if perPage <= 0 || total <= 0 {
		return 0
	}
	return int((total + int64(perPage) - 1) / int64(perPage))
}

// NewResult builds a paginated result from items, total count, and request params.
func NewResult[T any](items []T, total int64, params Params) Result[T] {
	if items == nil {
		items = []T{}
	}
	p := params.Normalize()
	totalPages := TotalPages(total, p.PerPage)
	return Result[T]{
		Items: items,
		Meta: Meta{
			Page:       p.Page,
			PerPage:    p.PerPage,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    totalPages > 0 && p.Page < totalPages,
			HasPrev:    p.Page > 1 && total > 0,
			Search:     p.Search,
			SortBy:     p.Sort.By,
			SortDir:    p.Sort.Dir,
		},
	}
}

// Empty returns an empty page with correct meta.
func Empty[T any](params Params) Result[T] {
	return NewResult([]T{}, 0, params)
}

// Map transforms item types while preserving meta (e.g. domain → DTO).
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	out := make([]U, len(r.Items))
	for i := range r.Items {
		out[i] = fn(r.Items[i])
	}
	return Result[U]{Items: out, Meta: r.Meta}
}

// MapSlice transforms a slice in one step (convenience for []*T → []DTO).
func MapSlice[T, U any](items []T, total int64, params Params, fn func(T) U) Result[U] {
	mapped := make([]U, len(items))
	for i := range items {
		mapped[i] = fn(items[i])
	}
	return NewResult(mapped, total, params)
}
