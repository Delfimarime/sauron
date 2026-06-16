package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripperFunc adapts a function to http.RoundTripper for tests.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestNew asserts options are applied in order and option errors abort construction.
func TestNew(t *testing.T) {
	t.Run("applies options in order", func(t *testing.T) {
		// Arrange.
		var order []int
		opt := func(n int) func(*http.Client) error {
			return func(*http.Client) error {
				order = append(order, n)
				return nil
			}
		}

		// Act.
		client, err := New(opt(1), opt(2))

		// Assert.
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, []int{1, 2}, order)
	})

	t.Run("propagates an option error", func(t *testing.T) {
		// Arrange.
		sentinel := errors.New("boom")

		// Act.
		client, err := New(func(*http.Client) error { return sentinel })

		// Assert.
		require.ErrorIs(t, err, sentinel)
		assert.Nil(t, client)
	})
}
