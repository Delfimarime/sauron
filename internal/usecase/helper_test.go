package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Shared test-data constants reused across the package's use-case tests (kept
// here to satisfy goconst across files).
const (
	unknownField = "bogus"
	invalidOrder = "sideways"
	colFirst     = "first"
)

// TestSelectFields covers the empty-default, identity-first, dedupe, and
// unknown-field paths of the shared field selector.
func TestSelectFields(t *testing.T) {
	known := map[string]struct{}{
		fieldName:      {},
		fieldTransport: {},
		fieldURI:       {},
	}
	dflt := []string{fieldName, fieldTransport}

	t.Run("empty request yields default", func(t *testing.T) {
		got, err := selectFields(nil, known, dflt)
		require.NoError(t, err)
		assert.Equal(t, dflt, got)
	})

	t.Run("forces name present and first", func(t *testing.T) {
		got, err := selectFields([]string{fieldURI}, known, dflt)
		require.NoError(t, err)
		assert.Equal(t, []string{fieldName, fieldURI}, got)
	})

	t.Run("dedupes repeated and name", func(t *testing.T) {
		got, err := selectFields([]string{fieldName, fieldURI, fieldURI, fieldName}, known, dflt)
		require.NoError(t, err)
		assert.Equal(t, []string{fieldName, fieldURI}, got)
	})

	t.Run("unknown field yields usage error", func(t *testing.T) {
		got, err := selectFields([]string{unknownField}, known, dflt)
		assert.Nil(t, got)
		ucErr := asHelperError(t, err)
		assert.Equal(t, TypeUsage, ucErr.Type)
	})
}

// TestProjectRows asserts the generic projector honours column order and applies
// each column's projector to every item.
func TestProjectRows(t *testing.T) {
	type pair struct {
		first  string
		second string
	}
	items := []pair{{"a1", "b1"}, {"a2", "b2"}}
	projectors := map[string]func(pair) string{
		colFirst: func(p pair) string { return p.first },
		"second": func(p pair) string { return p.second },
	}

	t.Run("column ordering and projection", func(t *testing.T) {
		got := projectRows(items, []string{"second", colFirst}, projectors)
		assert.Equal(t, [][]string{{"b1", "a1"}, {"b2", "a2"}}, got)
	})

	t.Run("empty items yields empty rows", func(t *testing.T) {
		got := projectRows(nil, []string{colFirst}, projectors)
		assert.Empty(t, got)
	})
}

// TestDefaultSortOrder asserts the empty-value defaulting and pass-through.
func TestDefaultSortOrder(t *testing.T) {
	tests := []struct {
		name      string
		sort      string
		order     string
		wantSort  string
		wantOrder string
	}{
		{name: "both empty", sort: "", order: "", wantSort: fieldName, wantOrder: orderAsc},
		{name: "sort set", sort: fieldTransport, order: "", wantSort: fieldTransport, wantOrder: orderAsc},
		{name: "order set", sort: "", order: orderDesc, wantSort: fieldName, wantOrder: orderDesc},
		{name: "both set", sort: fieldURI, order: orderDesc, wantSort: fieldURI, wantOrder: orderDesc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSort, gotOrder := defaultSortOrder(tt.sort, tt.order)
			assert.Equal(t, tt.wantSort, gotSort)
			assert.Equal(t, tt.wantOrder, gotOrder)
		})
	}
}

// TestIsValidOrder asserts only the two accepted directions pass.
func TestIsValidOrder(t *testing.T) {
	tests := []struct {
		order string
		want  bool
	}{
		{order: orderAsc, want: true},
		{order: orderDesc, want: true},
		{order: "", want: false},
		{order: invalidOrder, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.order, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidOrder(tt.order))
		})
	}
}

// asHelperError asserts err is a *Error and returns it.
func asHelperError(t *testing.T, err error) *Error {
	t.Helper()
	require.Error(t, err)
	var ucErr *Error
	require.True(t, errors.As(err, &ucErr), "expected *Error, got %T", err)

	return ucErr
}
