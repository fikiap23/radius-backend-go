package pagination_test

import (
	"testing"

	"github.com/radius/radius-backend/internal/shared/pagination"
)

func TestNormalizeSort(t *testing.T) {
	t.Parallel()

	got := pagination.NormalizeSort("invalid", pagination.SortAsc, "createdAt", "createdAt", "name")
	if got.By != "createdAt" || got.Dir != pagination.SortAsc {
		t.Fatalf("NormalizeSort() = %+v", got)
	}
}

func TestParseSortDir(t *testing.T) {
	t.Parallel()

	if pagination.ParseSortDir("asc") != pagination.SortAsc {
		t.Fatal("expected asc")
	}
	if pagination.ParseSortDir("DESC") != pagination.SortDesc {
		t.Fatal("expected desc")
	}
}

func TestHTTPQuery_ParamsWithSort(t *testing.T) {
	t.Parallel()

	p := pagination.HTTPQuery{
		Page:    1,
		PerPage: 10,
		Search:  "  john  ",
		SortBy:  "name",
		SortDir: "asc",
	}.ParamsWithSort("createdAt", "createdAt", "name", "email")

	if p.Search != "john" || p.Sort.By != "name" || p.Sort.Dir != pagination.SortAsc {
		t.Fatalf("ParamsWithSort() = %+v", p)
	}
}
