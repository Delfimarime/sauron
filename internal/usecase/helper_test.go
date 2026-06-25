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
	invalidOrder = "sideways"
)

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
		{name: "sort set", sort: "transport", order: "", wantSort: "transport", wantOrder: orderAsc},
		{name: "order set", sort: "", order: orderDesc, wantSort: fieldName, wantOrder: orderDesc},
		{name: "both set", sort: "uri", order: orderDesc, wantSort: "uri", wantOrder: orderDesc},
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

// TestFilterBy asserts the generic predicate keeps only the accepted items.
func TestFilterBy(t *testing.T) {
	got := filterBy([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
	assert.Equal(t, []int{2, 4}, got)
}

// asUseCaseError asserts err is a *Error with the expected Type and returns it.
func asUseCaseError(t *testing.T, err error, want Type) *Error {
	t.Helper()
	require.Error(t, err)

	var ucErr *Error
	require.True(t, errors.As(err, &ucErr), "want *Error, got %T", err)
	assert.Equal(t, want, ucErr.Type)

	return ucErr
}

// requireErrType asserts err is a *Error with the expected Type.
func requireErrType(t *testing.T, err error, want Type) {
	t.Helper()
	_ = asUseCaseError(t, err, want)
}
