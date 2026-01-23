package handlers

import (
	"net/http"
	"testing"

	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

// TestGetPaginationParams tests pagination parameter extraction
func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name         string
		queryString  string
		wantPage     int
		wantPageSize int
		wantOffset   int
	}{
		{
			name:         "default values",
			queryString:  "",
			wantPage:     1,
			wantPageSize: DefaultPageSize,
			wantOffset:   0,
		},
		{
			name:         "page only",
			queryString:  "?page=3",
			wantPage:     3,
			wantPageSize: DefaultPageSize,
			wantOffset:   (3 - 1) * DefaultPageSize,
		},
		{
			name:         "page_size only",
			queryString:  "?page_size=25",
			wantPage:     1,
			wantPageSize: 25,
			wantOffset:   0,
		},
		{
			name:         "both parameters",
			queryString:  "?page=2&page_size=10",
			wantPage:     2,
			wantPageSize: 10,
			wantOffset:   10, // (2-1) * 10
		},
		{
			name:         "invalid page (negative)",
			queryString:  "?page=-1",
			wantPage:     1, // should default
			wantPageSize: DefaultPageSize,
			wantOffset:   0,
		},
		{
			name:         "invalid page (zero)",
			queryString:  "?page=0",
			wantPage:     1, // should default
			wantPageSize: DefaultPageSize,
			wantOffset:   0,
		},
		{
			name:         "invalid page (string)",
			queryString:  "?page=abc",
			wantPage:     1, // should default
			wantPageSize: DefaultPageSize,
			wantOffset:   0,
		},
		{
			name:         "page_size exceeds max",
			queryString:  "?page_size=500",
			wantPage:     1,
			wantPageSize: MaxPageSize,
			wantOffset:   0,
		},
		{
			name:         "page_size at max",
			queryString:  "?page_size=100",
			wantPage:     1,
			wantPageSize: MaxPageSize,
			wantOffset:   0,
		},
		{
			name:         "large page number",
			queryString:  "?page=100&page_size=10",
			wantPage:     100,
			wantPageSize: 10,
			wantOffset:   990, // (100-1) * 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := testutil.NewContext(t, http.MethodGet, "/test"+tt.queryString, nil)
			params := GetPaginationParams(c)

			if params.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.wantPage)
			}
			if params.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", params.PageSize, tt.wantPageSize)
			}
			if params.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", params.Offset, tt.wantOffset)
			}
		})
	}
}

// TestIsPaginationRequested tests pagination detection
func TestIsPaginationRequested(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		want        bool
	}{
		{
			name:        "no pagination params",
			queryString: "",
			want:        false,
		},
		{
			name:        "page param",
			queryString: "?page=1",
			want:        true,
		},
		{
			name:        "page_size param",
			queryString: "?page_size=10",
			want:        true,
		},
		{
			name:        "both params",
			queryString: "?page=1&page_size=10",
			want:        true,
		},
		{
			name:        "other params only",
			queryString: "?status=active&type=loadBalancer",
			want:        false,
		},
		{
			name:        "mixed params",
			queryString: "?status=active&page=2",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := testutil.NewContext(t, http.MethodGet, "/test"+tt.queryString, nil)
			got := IsPaginationRequested(c)

			if got != tt.want {
				t.Errorf("IsPaginationRequested() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewPaginatedResponse tests paginated response creation
func TestNewPaginatedResponse(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		total          int
		params         PaginationParams
		wantTotalPages int
	}{
		{
			name:           "exact division",
			data:           []string{"a", "b"},
			total:          100,
			params:         PaginationParams{Page: 1, PageSize: 10, Offset: 0},
			wantTotalPages: 10,
		},
		{
			name:           "remainder",
			data:           []string{"a", "b"},
			total:          95,
			params:         PaginationParams{Page: 1, PageSize: 10, Offset: 0},
			wantTotalPages: 10, // ceil(95/10)
		},
		{
			name:           "single page",
			data:           []string{"a", "b", "c"},
			total:          3,
			params:         PaginationParams{Page: 1, PageSize: 10, Offset: 0},
			wantTotalPages: 1,
		},
		{
			name:           "empty data",
			data:           []string{},
			total:          0,
			params:         PaginationParams{Page: 1, PageSize: 10, Offset: 0},
			wantTotalPages: 0,
		},
		{
			name:           "one more than page size",
			data:           []string{"a"},
			total:          11,
			params:         PaginationParams{Page: 2, PageSize: 10, Offset: 10},
			wantTotalPages: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewPaginatedResponse(tt.data, tt.total, tt.params)

			if resp.Total != tt.total {
				t.Errorf("Total = %d, want %d", resp.Total, tt.total)
			}
			if resp.Page != tt.params.Page {
				t.Errorf("Page = %d, want %d", resp.Page, tt.params.Page)
			}
			if resp.PageSize != tt.params.PageSize {
				t.Errorf("PageSize = %d, want %d", resp.PageSize, tt.params.PageSize)
			}
			if resp.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", resp.TotalPages, tt.wantTotalPages)
			}
			if resp.Data == nil {
				t.Error("Data is nil")
			}
		})
	}
}

// TestPaginationConstants tests pagination constant values
func TestPaginationConstants(t *testing.T) {
	if DefaultPageSize <= 0 {
		t.Errorf("DefaultPageSize should be positive, got %d", DefaultPageSize)
	}
	if MaxPageSize <= 0 {
		t.Errorf("MaxPageSize should be positive, got %d", MaxPageSize)
	}
	if MaxPageSize < DefaultPageSize {
		t.Errorf("MaxPageSize (%d) should be >= DefaultPageSize (%d)", MaxPageSize, DefaultPageSize)
	}
}
