package pagination

import "strings"

// SortDir is the sort direction for list queries.
type SortDir string

const (
	SortAsc  SortDir = "asc"
	SortDesc SortDir = "desc"
)

// Sort holds field name and direction.
type Sort struct {
	By  string
	Dir SortDir
}

// ParseSortDir normalizes a query string to asc or desc.
func ParseSortDir(raw string) SortDir {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(SortAsc), "ascending":
		return SortAsc
	default:
		return SortDesc
	}
}

// NormalizeSort validates sort field against allowed keys and applies defaults.
func NormalizeSort(by string, dir SortDir, defaultBy string, allowed ...string) Sort {
	by = strings.TrimSpace(by)
	if by == "" {
		by = defaultBy
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, f := range allowed {
		allowedSet[f] = struct{}{}
	}
	if _, ok := allowedSet[by]; !ok {
		by = defaultBy
	}

	if dir != SortAsc {
		dir = SortDesc
	}

	return Sort{By: by, Dir: dir}
}

// IsAsc reports whether sort direction is ascending.
func (s Sort) IsAsc() bool {
	return s.Dir == SortAsc
}
