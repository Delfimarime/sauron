package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/delfimarime/sauron/pkg/telemetry"
)

// TestNewLoggerRoundTripper asserts the constructor retains its collaborators.
func TestNewLoggerRoundTripper(t *testing.T) {
	// Arrange.
	next := roundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, nil })

	// Act.
	rt := NewLoggerRoundTripper(zap.NewNop(), next)

	// Assert.
	require.NotNil(t, rt)
	require.NotNil(t, rt.next)
	require.NotNil(t, rt.logger)
}

// TestLoggerRoundTripper_RoundTrip covers the success and error logging paths.
func TestLoggerRoundTripper_RoundTrip(t *testing.T) {
	t.Run("logs the request and response on success", func(t *testing.T) {
		// Arrange.
		core, logs := observer.New(zap.DebugLevel)
		next := roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: http.NoBody}, nil
		})
		rt := NewLoggerRoundTripper(zap.New(core), next)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.test/path", strings.NewReader("{}"))
		req.Header.Set(telemetry.HeaderContentType, "application/json")

		// Act.
		resp, err := rt.RoundTrip(req)

		// Assert.
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, logs.Len())
	})

	t.Run("logs the error when the next round-tripper fails", func(t *testing.T) {
		// Arrange.
		core, logs := observer.New(zap.DebugLevel)
		sentinel := errors.New("dial failed")
		next := roundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, sentinel })
		rt := NewLoggerRoundTripper(zap.New(core), next)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.test", nil)

		// Act.
		resp, err := rt.RoundTrip(req)

		// Assert.
		require.ErrorIs(t, err, sentinel)
		assert.Nil(t, resp)
		if resp != nil {
			_ = resp.Body.Close()
		}
		assert.Equal(t, 2, logs.Len())
	})

	t.Run("tolerates a request without a URL", func(t *testing.T) {
		// Arrange.
		next := roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: http.NoBody}, nil
		})
		rt := NewLoggerRoundTripper(zap.NewNop(), next)
		req := &http.Request{Method: http.MethodGet, Header: http.Header{}}

		// Act.
		resp, err := rt.RoundTrip(req)

		// Assert.
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer func() { _ = resp.Body.Close() }()
	})
}

// TestWithLoggerRoundTripper asserts the option installs a logging round-tripper.
func TestWithLoggerRoundTripper(t *testing.T) {
	// Arrange.
	client := &http.Client{Transport: http.DefaultTransport}

	// Act.
	err := WithLoggerRoundTripper(zap.NewNop())(client)

	// Assert.
	require.NoError(t, err)
	_, ok := client.Transport.(*LoggerRoundTripper)
	assert.True(t, ok)
}
