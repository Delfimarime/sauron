package source_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// TestOptions asserts each functional option allocates and sets the pointer
// field it documents (non-nil, correct dereferenced value), leaving the others
// nil. It is purely in-memory — no filesystem, no env.
func TestOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		option  source.Option
		field   string
		wantStr string
		wantInt int64
	}{
		{name: "WithSearch sets Search", option: source.WithSearch("go-style"), field: "search", wantStr: "go-style"},
		{name: "WithLimit sets Limit", option: source.WithLimit(25), field: "limit", wantInt: 25},
		{name: "WithOffset sets Offset", option: source.WithOffset(50), field: "offset", wantInt: 50},
		{name: "WithSort sets Sort", option: source.WithSort("name"), field: "sort", wantStr: "name"},
		{name: "WithOrder sets Order", option: source.WithOrder("desc"), field: "order", wantStr: "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var o source.Options

			// Act
			tt.option(&o)

			// Assert: only the targeted field is set, with the expected value.
			switch tt.field {
			case "search":
				require.NotNil(t, o.Search)
				assert.Equal(t, tt.wantStr, *o.Search)
				assert.Nil(t, o.Limit)
				assert.Nil(t, o.Offset)
				assert.Nil(t, o.Sort)
				assert.Nil(t, o.Order)
			case "limit":
				require.NotNil(t, o.Limit)
				assert.Equal(t, tt.wantInt, *o.Limit)
				assert.Nil(t, o.Search)
				assert.Nil(t, o.Offset)
				assert.Nil(t, o.Sort)
				assert.Nil(t, o.Order)
			case "offset":
				require.NotNil(t, o.Offset)
				assert.Equal(t, tt.wantInt, *o.Offset)
				assert.Nil(t, o.Search)
				assert.Nil(t, o.Limit)
				assert.Nil(t, o.Sort)
				assert.Nil(t, o.Order)
			case "sort":
				require.NotNil(t, o.Sort)
				assert.Equal(t, tt.wantStr, *o.Sort)
				assert.Nil(t, o.Search)
				assert.Nil(t, o.Limit)
				assert.Nil(t, o.Offset)
				assert.Nil(t, o.Order)
			case "order":
				require.NotNil(t, o.Order)
				assert.Equal(t, tt.wantStr, *o.Order)
				assert.Nil(t, o.Search)
				assert.Nil(t, o.Limit)
				assert.Nil(t, o.Offset)
				assert.Nil(t, o.Sort)
			}
		})
	}
}

// TestOptionsComposed asserts that applying every option together sets each
// pointer field independently (all non-nil with the expected values).
func TestOptionsComposed(t *testing.T) {
	t.Parallel()

	// Arrange
	var o source.Options
	options := []source.Option{
		source.WithSearch("go-style"),
		source.WithLimit(25),
		source.WithOffset(50),
		source.WithSort("name"),
		source.WithOrder("desc"),
	}

	// Act
	for _, opt := range options {
		opt(&o)
	}

	// Assert
	require.NotNil(t, o.Search)
	require.NotNil(t, o.Limit)
	require.NotNil(t, o.Offset)
	require.NotNil(t, o.Sort)
	require.NotNil(t, o.Order)
	assert.Equal(t, "go-style", *o.Search)
	assert.Equal(t, int64(25), *o.Limit)
	assert.Equal(t, int64(50), *o.Offset)
	assert.Equal(t, "name", *o.Sort)
	assert.Equal(t, "desc", *o.Order)
}

// TestErrNotImplemented asserts the sentinel error is a non-nil, identifiable
// value callers can match.
func TestErrNotImplemented(t *testing.T) {
	t.Parallel()

	require.Error(t, source.ErrNotImplemented)
	assert.Equal(t, "not implemented", source.ErrNotImplemented.Error())
}
