package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCataloguePagingLine pins the populated-window and empty-page report
// formats.
func TestCataloguePagingLine(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// page, limit, count are the inputs.
		page  int64
		limit int64
		count int
		// want is the exact rendered line.
		want string
	}{
		{
			name:  "populated window",
			page:  1,
			limit: 20,
			count: 2,
			want:  "showing 1–2 (page 1, limit 20)",
		},
		{
			name:  "second page window",
			page:  2,
			limit: 1,
			count: 1,
			want:  "showing 2–2 (page 2, limit 1)",
		},
		{
			name:  "empty page reports zero results",
			page:  9,
			limit: 20,
			count: 0,
			want:  "showing 0 results (page 9, limit 20)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CataloguePagingLine(tt.page, tt.limit, tt.count))
		})
	}
}
