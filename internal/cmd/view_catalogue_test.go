package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
)

// TestRenderCatalogue covers the NAME/KIND table and the paging line for a
// populated window and an empty page.
func TestRenderCatalogue(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// result is the listing to render.
		result *usecase.ListCatalogueResult
		// wantContains are substrings the output must contain.
		wantContains []string
		// wantAbsent are substrings the output must never contain.
		wantAbsent []string
	}{
		{
			name: "populated window renders the table and the from-to line",
			result: &usecase.ListCatalogueResult{
				Kind:   usecase.CatalogueAgent,
				Items:  []string{"review", "doc"},
				Page:   1,
				Limit:  20,
				Offset: 0,
			},
			wantContains: []string{"NAME", "KIND", "review", "doc", "agent", "showing 1–2 (page 1, limit 20)"},
		},
		{
			name: "empty page renders no table and the zero-results line",
			result: &usecase.ListCatalogueResult{
				Kind:   usecase.CatalogueSkill,
				Items:  nil,
				Page:   9,
				Limit:  20,
				Offset: 160,
			},
			wantContains: []string{"showing 0 results (page 9, limit 20)"},
			wantAbsent:   []string{"NAME"},
		},
		{
			name: "single-row window reports the inclusive window",
			result: &usecase.ListCatalogueResult{
				Kind:   usecase.CatalogueSkill,
				Items:  []string{"b"},
				Page:   2,
				Limit:  1,
				Offset: 1,
			},
			wantContains: []string{"showing 2–2 (page 2, limit 1)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := renderCatalogue(&buf, tt.result)

			// Assert.
			require.NoError(t, err)
			out := buf.String()
			for _, want := range tt.wantContains {
				assert.Contains(t, out, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, out, absent)
			}
		})
	}
}

// TestRenderCatalogueWriteError surfaces a writer failure as an io error on both
// the table and the paging-line write.
func TestRenderCatalogueWriteError(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// result is the listing to render.
		result *usecase.ListCatalogueResult
	}{
		{
			name:   "table write fails",
			result: &usecase.ListCatalogueResult{Kind: usecase.CatalogueSkill, Items: []string{"a"}, Page: 1, Limit: 20},
		},
		{
			name:   "paging-line write fails on an empty page",
			result: &usecase.ListCatalogueResult{Kind: usecase.CatalogueSkill, Items: nil, Page: 1, Limit: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := renderCatalogue(&failingWriter{}, tt.result)

			// Assert.
			var ucErr *usecase.Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, usecase.TypeIO, ucErr.Type)
		})
	}
}

// TestPagingLineKindRendered confirms the kind is rendered in the KIND column.
func TestPagingLineKindRendered(t *testing.T) {
	// Arrange.
	var buf bytes.Buffer
	result := &usecase.ListCatalogueResult{Kind: usecase.CatalogueSkill, Items: []string{"x"}, Page: 1, Limit: 20}

	// Act.
	require.NoError(t, renderCatalogue(&buf, result))

	// Assert.
	rows := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	assert.Equal(t, "KIND", strings.Fields(rows[0])[1])
	assert.Contains(t, rows[1], "skill")
}
