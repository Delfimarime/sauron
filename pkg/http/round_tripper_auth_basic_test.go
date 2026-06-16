package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewBasicAuthRoundTrip asserts the constructor retains its collaborators.
func TestNewBasicAuthRoundTrip(t *testing.T) {
	// Arrange.
	base := roundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, nil })

	// Act.
	rt := NewBasicAuthRoundTrip(base, "alice", "s3cret")

	// Assert.
	require.NotNil(t, rt)
	require.NotNil(t, rt.base)
	assert.Equal(t, "alice", rt.username)
	assert.Equal(t, "s3cret", rt.password)
}

// TestBasicAuthRoundTrip_RoundTrip covers credential injection and preservation.
func TestBasicAuthRoundTrip_RoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		presetAuth string
		override   bool
	}{
		{name: "adds credentials when absent", override: true},
		{name: "preserves an existing Authorization header", presetAuth: "Bearer token", override: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			var gotUser, gotPass, gotAuth string
			var gotOK bool
			base := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				gotUser, gotPass, gotOK = r.BasicAuth()
				gotAuth = r.Header.Get("Authorization")
				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
			})
			rt := NewBasicAuthRoundTrip(base, "alice", "s3cret")
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.test", nil)
			if tc.presetAuth != "" {
				req.Header.Set("Authorization", tc.presetAuth)
			}

			// Act.
			resp, err := rt.RoundTrip(req)

			// Assert.
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer func() { _ = resp.Body.Close() }()
			if tc.override {
				require.True(t, gotOK)
				assert.Equal(t, "alice", gotUser)
				assert.Equal(t, "s3cret", gotPass)
				assert.Empty(t, req.Header.Get("Authorization"), "the original request must not be mutated")
			} else {
				assert.Equal(t, tc.presetAuth, gotAuth)
			}
		})
	}
}

// TestWithBasicAuth asserts the option installs a Basic-auth round-tripper.
func TestWithBasicAuth(t *testing.T) {
	// Arrange.
	client := &http.Client{Transport: http.DefaultTransport}

	// Act.
	err := WithBasicAuth("alice", "s3cret")(client)

	// Assert.
	require.NoError(t, err)
	_, ok := client.Transport.(*BasicAuthRoundTrip)
	assert.True(t, ok)
}
