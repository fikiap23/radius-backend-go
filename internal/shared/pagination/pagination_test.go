package pagination_test

import (
	"testing"

	"github.com/radius/radius-backend/internal/shared/pagination"
)

func TestParams_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   pagination.Params
		want pagination.Params
	}{
		{
			name: "defaults",
			in:   pagination.Params{},
			want: pagination.Params{Page: 1, PerPage: 20},
		},
		{
			name: "caps per page",
			in:   pagination.Params{Page: 2, PerPage: 500},
			want: pagination.Params{Page: 2, PerPage: 100},
		},
		{
			name: "invalid page",
			in:   pagination.Params{Page: 0, PerPage: 10},
			want: pagination.Params{Page: 1, PerPage: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.in.Normalize()
			if got != tt.want {
				t.Fatalf("Normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParams_OffsetLimit(t *testing.T) {
	t.Parallel()

	ol := pagination.Params{Page: 3, PerPage: 10}.OffsetLimit()
	if ol.Limit != 10 || ol.Offset != 20 {
		t.Fatalf("OffsetLimit() = %+v, want limit=10 offset=20", ol)
	}
}

func TestNewResult_meta(t *testing.T) {
	t.Parallel()

	r := pagination.NewResult([]string{"a", "b"}, 25, pagination.Params{Page: 2, PerPage: 10})

	if r.Meta.Page != 2 || r.Meta.PerPage != 10 || r.Meta.Total != 25 {
		t.Fatalf("meta = %+v", r.Meta)
	}
	if r.Meta.TotalPages != 3 {
		t.Fatalf("TotalPages = %d, want 3", r.Meta.TotalPages)
	}
	if !r.Meta.HasNext || !r.Meta.HasPrev {
		t.Fatalf("HasNext=%v HasPrev=%v", r.Meta.HasNext, r.Meta.HasPrev)
	}
}

func TestMap_preservesMeta(t *testing.T) {
	t.Parallel()

	src := pagination.NewResult([]int{1, 2}, 2, pagination.Params{Page: 1, PerPage: 20})
	dst := pagination.Map(src, func(n int) string {
		return "x"
	})

	if len(dst.Items) != 2 || dst.Meta.Total != 2 {
		t.Fatalf("Map() = %+v", dst)
	}
}

func TestHTTPQuery_Params(t *testing.T) {
	t.Parallel()

	p := pagination.HTTPQuery{Page: 2, PerPage: 15, SortDir: "asc"}.Params()
	if p.Page != 2 || p.PerPage != 15 || p.Sort.Dir != pagination.SortAsc {
		t.Fatalf("Params() = %+v", p)
	}
}
