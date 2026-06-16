package http

import (
	"crypto/tls"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithSimpleRoundTripper asserts the option installs a configured transport.
func TestWithSimpleRoundTripper(t *testing.T) {
	t.Run("installs a configured transport", func(t *testing.T) {
		// Arrange.
		var got *http.Transport
		client := &http.Client{}

		// Act.
		err := WithSimpleRoundTripper(func(tr *http.Transport) error {
			got = tr
			return nil
		})(client)

		// Assert.
		require.NoError(t, err)
		tr, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.Same(t, got, tr)
		assert.True(t, tr.ForceAttemptHTTP2)
		assert.NotNil(t, tr.Proxy)
	})

	t.Run("propagates an option error", func(t *testing.T) {
		// Arrange.
		sentinel := errors.New("bad option")
		client := &http.Client{}

		// Act.
		err := WithSimpleRoundTripper(func(*http.Transport) error { return sentinel })(client)

		// Assert.
		require.ErrorIs(t, err, sentinel)
		assert.Nil(t, client.Transport)
	})
}

// TestNewSimpleRoundTripper asserts the transport is built and configured.
func TestNewSimpleRoundTripper(t *testing.T) {
	// Arrange.
	var applied bool

	// Act.
	rt, err := NewSimpleRoundTripper(func(*http.Transport) { applied = true })

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, rt)
	assert.True(t, applied)
	tr, ok := rt.(*http.Transport)
	require.True(t, ok)
	assert.True(t, tr.ForceAttemptHTTP2)
}

// TestWithTLS covers the no-truststore, missing-file, and invalid-PEM paths.
func TestWithTLS(t *testing.T) {
	t.Run("no truststore configures TLS from system roots", func(t *testing.T) {
		// Arrange + Act.
		apply, err := WithTLS("example.test", "", true)

		// Assert.
		require.NoError(t, err)
		require.NotNil(t, apply)
		transport := &http.Transport{}
		apply(transport)
		require.NotNil(t, transport.TLSClientConfig)
		assert.Equal(t, "example.test", transport.TLSClientConfig.ServerName)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
		assert.Equal(t, uint16(tls.VersionTLS12), transport.TLSClientConfig.MinVersion)
		assert.NotNil(t, transport.TLSClientConfig.RootCAs)
	})

	t.Run("missing truststore file returns a path error", func(t *testing.T) {
		// Arrange: t.TempDir yields a path only; the file is never created.
		absent := filepath.Join(t.TempDir(), "absent.pem")

		// Act.
		apply, err := WithTLS("example.test", absent, false)

		// Assert.
		require.Error(t, err)
		assert.Nil(t, apply)
		var pathErr *fs.PathError
		assert.ErrorAs(t, err, &pathErr)
	})

	t.Run("invalid PEM truststore returns errInvalidTrustStore", func(t *testing.T) {
		// Arrange + Act.
		apply, err := WithTLS("example.test", filepath.Join("testdata", "not-a-cert.pem"), false)

		// Assert.
		require.ErrorIs(t, err, errInvalidTrustStore)
		assert.Nil(t, apply)
	})
}

// TestWithConnectionPool asserts the configured limits are applied and that zero
// values leave the transport defaults untouched.
func TestWithConnectionPool(t *testing.T) {
	t.Run("applies the configured limits", func(t *testing.T) {
		// Arrange + Act.
		apply, err := WithConnectionPool(10, 5, 30*time.Second)
		require.NoError(t, err)
		transport := &http.Transport{}
		apply(transport)

		// Assert.
		assert.Equal(t, 10, transport.MaxIdleConns)
		assert.Equal(t, 5, transport.MaxIdleConnsPerHost)
		assert.Equal(t, 30*time.Second, transport.IdleConnTimeout)
	})

	t.Run("leaves defaults when given zero values", func(t *testing.T) {
		// Arrange + Act.
		apply, err := WithConnectionPool(0, 0, 0)
		require.NoError(t, err)
		transport := &http.Transport{}
		apply(transport)

		// Assert.
		assert.Zero(t, transport.MaxIdleConns)
		assert.Zero(t, transport.MaxIdleConnsPerHost)
		assert.Zero(t, transport.IdleConnTimeout)
	})
}

// TestWithTimeouts asserts the configured timeouts are applied and that zero
// values leave the transport defaults untouched.
func TestWithTimeouts(t *testing.T) {
	t.Run("applies the configured timeouts", func(t *testing.T) {
		// Arrange + Act.
		apply, err := WithTimeouts(time.Second, 2*time.Second, 3*time.Second)
		require.NoError(t, err)
		transport := &http.Transport{}
		apply(transport)

		// Assert.
		assert.Equal(t, time.Second, transport.TLSHandshakeTimeout)
		assert.Equal(t, 2*time.Second, transport.ResponseHeaderTimeout)
		assert.Equal(t, 3*time.Second, transport.ExpectContinueTimeout)
	})

	t.Run("leaves defaults when given zero values", func(t *testing.T) {
		// Arrange + Act.
		apply, err := WithTimeouts(0, 0, 0)
		require.NoError(t, err)
		transport := &http.Transport{}
		apply(transport)

		// Assert.
		assert.Zero(t, transport.TLSHandshakeTimeout)
		assert.Zero(t, transport.ResponseHeaderTimeout)
		assert.Zero(t, transport.ExpectContinueTimeout)
	})
}
